package downloader

import (
	"context"
	"encoding/json"

	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type testResolver struct{ paths ToolPaths }

func (r testResolver) Resolve() (ToolPaths, error) { return r.paths, nil }

type memoryHistory struct{ entries chan HistoryEntry }

func (h *memoryHistory) SaveHistory(_ context.Context, entry HistoryEntry) error {
	h.entries <- entry
	return nil
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("YUMDOWN_TEST_HELPER") != "1" {
		return
	}
	args := os.Args
	for index, arg := range args {
		if arg == "--" {
			args = args[index+1:]
			break
		}
	}
	argsJSON, _ := json.Marshal(args)
	_ = os.WriteFile(os.Getenv("YUMDOWN_HELPER_ARGS"), argsJSON, 0o600)
	if os.Getenv("YUMDOWN_HELPER_MODE") == "fail" {
		fmt.Fprintln(os.Stderr, "HELPER_ERROR: simulated failure")
		os.Exit(9)
	}
	if os.Getenv("YUMDOWN_HELPER_MODE") == "long-line" {
		fmt.Print(strings.Repeat("x", 2*1024*1024))
		os.Exit(0)
	}
	if os.Getenv("YUMDOWN_HELPER_MODE") == "sleep" {
		_ = os.WriteFile(os.Getenv("YUMDOWN_HELPER_STARTED"), []byte("started"), 0o600)
		time.Sleep(30 * time.Second)
		os.Exit(0)
	}
	var outputTemplate string
	for index, arg := range args {
		if arg == "--output" && index+1 < len(args) {
			outputTemplate = args[index+1]
			break
		}
	}
	output := strings.ReplaceAll(outputTemplate, "%(ext)s", "mp3")
	if output == "" {
		fmt.Fprintln(os.Stderr, "missing output argument")
		os.Exit(10)
	}
	if err := os.WriteFile(output, []byte("audio fixture"), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(11)
	}
	fmt.Println("__YUMDOWN_PROGRESS__12|12|12|100|2|0")
	fmt.Println("__YUMDOWN_OUTPUT__" + output)
	os.Exit(0)
}

func fakeYTDLP(t *testing.T) string {
	t.Helper()
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		t.Skip("the local fake executable uses a POSIX test wrapper")
	}
	wrapper := filepath.Join(t.TempDir(), "fake-yt-dlp")
	body := fmt.Sprintf("#!/bin/sh\nexec %s -test.run=TestHelperProcess -- \"$@\"\n", shellQuote(os.Args[0]))
	if err := os.WriteFile(wrapper, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	return wrapper
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func TestStartBatchInvokesFakeToolWithDistinctArgumentsAndPersistsHistory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the local fake executable uses a POSIX test wrapper")
	}
	wrapper := fakeYTDLP(t)
	outputDir := t.TempDir()
	argsPath := filepath.Join(t.TempDir(), "args.json")
	t.Setenv("YUMDOWN_TEST_HELPER", "1")
	t.Setenv("YUMDOWN_HELPER_ARGS", argsPath)
	t.Setenv("YUMDOWN_HELPER_MODE", "success")
	history := &memoryHistory{entries: make(chan HistoryEntry, 1)}
	manager := NewManager(context.Background(), testResolver{paths: ToolPaths{YTDLP: wrapper, FFmpeg: "/fake/ffmpeg"}}, history,
		func() []AudioFormat { return []AudioFormat{FormatMP3} }, 1, nil)
	defer manager.Close()

	jobs, err := manager.StartBatch(DownloadRequest{
		Videos:    []VideoRequest{{URL: "https://video.example/watch?id=1&title=a", Filename: "safe-name"}},
		OutputDir: outputDir, Format: FormatMP3, AudioQuality: "128K",
	})
	if err != nil {
		t.Fatalf("StartBatch returned error: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("created %d jobs, want one", len(jobs))
	}
	completed := waitForJob(t, manager, jobs[0].ID, StatusCompleted)
	if completed.OutputSize != int64(len("audio fixture")) {
		t.Fatalf("output size = %d", completed.OutputSize)
	}
	if _, err := os.Stat(completed.OutputPath); err != nil {
		t.Fatalf("output file was not created: %v", err)
	}
	var args []string
	data, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &args); err != nil {
		t.Fatal(err)
	}
	if !containsArgument(args, "https://video.example/watch?id=1&title=a") {
		t.Fatalf("URL was not passed as one distinct argument: %#v", args)
	}
	if !containsArgument(args, "128K") || !containsArgument(args, "--ignore-config") {
		t.Fatalf("expected safe arguments not present: %#v", args)
	}
	select {
	case entry := <-history.entries:
		if entry.Status != StatusCompleted || entry.OutputSize != completed.OutputSize {
			t.Fatalf("history record = %#v", entry)
		}
	case <-time.After(time.Second):
		t.Fatal("terminal history was not persisted")
	}
}

func TestStartBatchCreatesSeparateFilesForMultipleURLs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the local fake executable uses a POSIX test wrapper")
	}
	wrapper := fakeYTDLP(t)
	outputDir := t.TempDir()
	t.Setenv("YUMDOWN_TEST_HELPER", "1")
	t.Setenv("YUMDOWN_HELPER_ARGS", filepath.Join(t.TempDir(), "args.json"))
	t.Setenv("YUMDOWN_HELPER_MODE", "success")
	history := &memoryHistory{entries: make(chan HistoryEntry, 2)}
	manager := NewManager(context.Background(), testResolver{paths: ToolPaths{YTDLP: wrapper, FFmpeg: "/fake/ffmpeg"}}, history,
		func() []AudioFormat { return []AudioFormat{FormatMP3} }, 2, nil)
	defer manager.Close()

	jobs, err := manager.StartBatch(DownloadRequest{
		Videos: []VideoRequest{
			{URL: "https://video.example/watch?id=1", Filename: "batch-01"},
			{URL: "https://video.example/watch?id=2", Filename: "batch-02"},
		},
		OutputDir: outputDir, Format: FormatMP3,
	})
	if err != nil {
		t.Fatalf("StartBatch returned error: %v", err)
	}
	if len(jobs) != 2 || jobs[0].OutputName != "batch-01" || jobs[1].OutputName != "batch-02" {
		t.Fatalf("batch jobs = %#v", jobs)
	}
	first := waitForJob(t, manager, jobs[0].ID, StatusCompleted)
	second := waitForJob(t, manager, jobs[1].ID, StatusCompleted)
	if first.OutputPath == second.OutputPath {
		t.Fatalf("batch items wrote to the same output path %q", first.OutputPath)
	}
	for range 2 {
		select {
		case entry := <-history.entries:
			if entry.Status != StatusCompleted {
				t.Fatalf("history status = %s", entry.Status)
			}
		case <-time.After(time.Second):
			t.Fatal("batch history was not persisted")
		}
	}
}

func TestDownloadSurfacesProcessFailures(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the local fake executable uses a POSIX test wrapper")
	}
	wrapper := fakeYTDLP(t)
	t.Setenv("YUMDOWN_TEST_HELPER", "1")
	t.Setenv("YUMDOWN_HELPER_ARGS", filepath.Join(t.TempDir(), "args.json"))
	t.Setenv("YUMDOWN_HELPER_MODE", "fail")
	manager := NewManager(context.Background(), testResolver{paths: ToolPaths{YTDLP: wrapper, FFmpeg: "/fake/ffmpeg"}}, nil,
		func() []AudioFormat { return []AudioFormat{FormatMP3} }, 1, nil)
	defer manager.Close()

	id := "failure-test"
	manager.jobs[id] = &jobRecord{job: Job{ID: id, OutputName: "fixture"}}
	_, err := manager.download(context.Background(), id, DownloadRequest{
		Videos: []VideoRequest{{URL: "https://video.example/watch"}}, OutputDir: t.TempDir(), Format: FormatMP3,
	})
	if err == nil || !strings.Contains(err.Error(), "HELPER_ERROR") {
		t.Fatalf("download error = %v, want fake process failure", err)
	}
}

func TestDownloadTerminatesToolWhenOutputLineExceedsParserLimit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the local fake executable uses a POSIX test wrapper")
	}
	wrapper := fakeYTDLP(t)
	t.Setenv("YUMDOWN_TEST_HELPER", "1")
	t.Setenv("YUMDOWN_HELPER_ARGS", filepath.Join(t.TempDir(), "args.json"))
	t.Setenv("YUMDOWN_HELPER_MODE", "long-line")
	manager := NewManager(context.Background(), testResolver{paths: ToolPaths{YTDLP: wrapper, FFmpeg: "/fake/ffmpeg"}}, nil,
		func() []AudioFormat { return []AudioFormat{FormatMP3} }, 1, nil)
	defer manager.Close()
	manager.jobs["long-line"] = &jobRecord{job: Job{ID: "long-line", OutputName: "fixture"}}

	_, err := manager.download(context.Background(), "long-line", DownloadRequest{
		Videos: []VideoRequest{{URL: "https://video.example/watch"}}, OutputDir: t.TempDir(), Format: FormatMP3,
	})
	if err == nil || !strings.Contains(err.Error(), "token too long") {
		t.Fatalf("download error = %v, want a scanner limit error", err)
	}
}

func TestCancelStopsActiveProcessAndPersistsCancelledState(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the local fake executable uses a POSIX test wrapper")
	}
	wrapper := fakeYTDLP(t)
	outputDir := t.TempDir()
	startedPath := filepath.Join(t.TempDir(), "started")
	history := &memoryHistory{entries: make(chan HistoryEntry, 1)}
	t.Setenv("YUMDOWN_TEST_HELPER", "1")
	t.Setenv("YUMDOWN_HELPER_ARGS", filepath.Join(t.TempDir(), "args.json"))
	t.Setenv("YUMDOWN_HELPER_MODE", "sleep")
	t.Setenv("YUMDOWN_HELPER_STARTED", startedPath)
	manager := NewManager(context.Background(), testResolver{paths: ToolPaths{YTDLP: wrapper, FFmpeg: "/fake/ffmpeg"}}, history,
		func() []AudioFormat { return []AudioFormat{FormatMP3} }, 1, nil)
	defer manager.Close()

	jobs, err := manager.StartBatch(DownloadRequest{
		Videos: []VideoRequest{{URL: "https://video.example/watch"}}, OutputDir: outputDir, Format: FormatMP3,
	})
	if err != nil {
		t.Fatalf("StartBatch returned error: %v", err)
	}
	waitForFile(t, startedPath)
	if err := manager.Cancel(jobs[0].ID); err != nil {
		t.Fatalf("Cancel returned error: %v", err)
	}
	cancelled := waitForJob(t, manager, jobs[0].ID, StatusCancelled)
	if cancelled.Error == "" {
		t.Fatal("cancelled job should explain why it stopped")
	}
	select {
	case entry := <-history.entries:
		if entry.Status != StatusCancelled {
			t.Fatalf("history status = %s, want cancelled", entry.Status)
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled job was not persisted")
	}
}

func waitForFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("fake downloader did not start in time")
}

func TestStartBatchRejectsExistingOutput(t *testing.T) {
	outputDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outputDir, "existing.mp3"), []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	manager := NewManager(context.Background(), testResolver{}, nil,
		func() []AudioFormat { return []AudioFormat{FormatMP3} }, 1, nil)
	defer manager.Close()
	_, err := manager.StartBatch(DownloadRequest{
		Videos:    []VideoRequest{{URL: "https://video.example/watch", Filename: "existing.mp3"}},
		OutputDir: outputDir, Format: FormatMP3,
	})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("StartBatch error = %v, want existing-output error", err)
	}
	if len(manager.Jobs()) != 0 {
		t.Fatal("job was added despite an output collision")
	}
}

func TestRedactErrorRemovesURLQuerySecrets(t *testing.T) {
	message := redactError("source rejected https://video.example/watch?id=secret&token=abc", "https://video.example/watch?id=secret&token=abc")
	if strings.Contains(message, "secret") || strings.Contains(message, "token=abc") {
		t.Fatalf("sensitive query data remained in error: %q", message)
	}
	if !strings.Contains(message, "https://video.example/watch") {
		t.Fatalf("redacted error lost useful source context: %q", message)
	}
}

func TestCancelUnknownJobReturnsError(t *testing.T) {
	manager := NewManager(context.Background(), nil, nil, nil, 1, nil)
	defer manager.Close()
	if err := manager.Cancel("missing"); err == nil {
		t.Fatal("expected an unknown job error")
	}
}

func waitForJob(t *testing.T, manager *Manager, id string, status JobStatus) Job {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		for _, job := range manager.Jobs() {
			if job.ID == id && job.Status == status {
				return job
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	for _, job := range manager.Jobs() {
		if job.ID == id {
			t.Fatalf("job did not reach %s; current job is %#v", status, job)
		}
	}
	t.Fatal("job disappeared before reaching terminal state")
	return Job{}
}

func containsArgument(args []string, expected string) bool {
	for _, arg := range args {
		if arg == expected {
			return true
		}
	}
	return false
}
