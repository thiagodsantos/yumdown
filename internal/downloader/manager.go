package downloader

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/thiagosantos/yumdown/internal/platform"
)

const (
	workerLimit = 4
	queueLimit  = 512
)

type ToolPaths struct {
	YTDLP   string
	FFmpeg  string
	FFprobe string
	Dir     string
}

type ToolResolver interface {
	Resolve() (ToolPaths, error)
}

type HistoryRepository interface {
	SaveHistory(context.Context, HistoryEntry) error
}

type FormatProvider func() []AudioFormat
type EventSink func(JobEvent)

type jobRecord struct {
	job    Job
	cancel context.CancelFunc
	ctx    context.Context
}

type Manager struct {
	ctx         context.Context
	cancel      context.CancelFunc
	resolver    ToolResolver
	history     HistoryRepository
	formats     FormatProvider
	events      EventSink
	queue       chan string
	mu          sync.Mutex
	cond        *sync.Cond
	jobs        map[string]*jobRecord
	order       []string
	concurrency int
	active      int
	closed      bool
	workers     sync.WaitGroup
}

func NewManager(parent context.Context, resolver ToolResolver, history HistoryRepository, formats FormatProvider, concurrency int, events EventSink) *Manager {
	if concurrency < 1 || concurrency > workerLimit {
		concurrency = 2
	}
	ctx, cancel := context.WithCancel(parent)
	m := &Manager{
		ctx: ctx, cancel: cancel, resolver: resolver, history: history, formats: formats,
		events: events, queue: make(chan string, queueLimit), jobs: make(map[string]*jobRecord),
		concurrency: concurrency,
	}
	m.cond = sync.NewCond(&m.mu)
	for i := 0; i < workerLimit; i++ {
		m.workers.Add(1)
		go m.worker()
	}
	return m
}

func (m *Manager) SetConcurrency(concurrency int) error {
	if concurrency < 1 || concurrency > workerLimit {
		return fmt.Errorf("concurrency must be between 1 and %d", workerLimit)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("download manager is closed")
	}
	m.concurrency = concurrency
	m.cond.Broadcast()
	return nil
}

func (m *Manager) StartBatch(req DownloadRequest) ([]Job, error) {
	if m.formats == nil {
		return nil, errors.New("audio format capabilities are unavailable")
	}
	validated, err := ValidateRequest(req, m.formats())
	if err != nil {
		return nil, err
	}
	if m.resolver == nil {
		return nil, errors.New("media tools are unavailable")
	}
	if _, err := m.resolver.Resolve(); err != nil {
		return nil, err
	}

	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil, errors.New("download manager is closed")
	}
	pending := 0
	for _, record := range m.jobs {
		if !record.job.Status.Terminal() {
			pending++
		}
	}
	if pending+len(validated.Videos) > queueLimit {
		m.mu.Unlock()
		return nil, errors.New("download queue is full; wait for current jobs to finish")
	}
	newRecords := make([]*jobRecord, 0, len(validated.Videos))
	for _, video := range validated.Videos {
		id, err := newID()
		if err != nil {
			m.mu.Unlock()
			return nil, fmt.Errorf("create download ID: %w", err)
		}
		name := video.Filename
		if name == "" {
			name = "yumdown-" + id[:8]
		}
		candidate := filepath.Join(validated.OutputDir, name+"."+string(validated.Format))
		for _, record := range newRecords {
			other := filepath.Join(record.job.OutputDir, record.job.OutputName+"."+string(record.job.Format))
			if samePath(candidate, other) {
				for _, created := range newRecords {
					created.cancel()
				}
				m.mu.Unlock()
				return nil, fmt.Errorf("multiple video rows use the filename %q", filepath.Base(candidate))
			}
		}
		if err := m.checkOutputNameLocked(validated.OutputDir, name, validated.Format); err != nil {
			for _, record := range newRecords {
				record.cancel()
			}
			m.mu.Unlock()
			return nil, err
		}
		job := newJob(id, video.URL, validated, name)
		jobCtx, cancel := context.WithCancel(m.ctx)
		newRecords = append(newRecords, &jobRecord{job: job, cancel: cancel, ctx: jobCtx})
	}
	for _, record := range newRecords {
		m.jobs[record.job.ID] = record
		m.order = append(m.order, record.job.ID)
		m.queue <- record.job.ID
	}
	created := make([]Job, 0, len(newRecords))
	for _, record := range newRecords {
		created = append(created, record.job)
	}
	m.mu.Unlock()
	return created, nil
}

func (m *Manager) Jobs() []Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.snapshotsLocked()
}

func (m *Manager) Cancel(id string) error {
	m.mu.Lock()
	record, ok := m.jobs[id]
	if !ok {
		m.mu.Unlock()
		return errors.New("download job was not found")
	}
	if record.job.Status.Terminal() {
		m.mu.Unlock()
		return errors.New("download job is already finished")
	}
	if record.cancel != nil {
		record.cancel()
	}
	if record.job.Status == StatusQueued {
		record.job.Status = StatusCancelled
		record.job.Error = "Cancelled by user"
		record.job.UpdatedAt = timestampNow()
		job := record.job
		event, aggregate := m.eventLocked(record.job.ID)
		m.cond.Broadcast()
		m.mu.Unlock()
		m.persistTerminal(job)
		m.publishEvent(event, aggregate)
		return nil
	}
	m.cond.Broadcast()
	m.mu.Unlock()
	return nil
}

func (m *Manager) Close() {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	m.closed = true
	m.cancel()
	var cancelled []Job
	for _, record := range m.jobs {
		if !record.job.Status.Terminal() {
			if record.cancel != nil {
				record.cancel()
			}
			if record.job.Status == StatusQueued {
				record.job.Status = StatusCancelled
				record.job.Error = "Cancelled because the application is closing"
				record.job.UpdatedAt = timestampNow()
				cancelled = append(cancelled, record.job)
			}
		}
	}
	m.cond.Broadcast()
	m.mu.Unlock()
	for _, job := range cancelled {
		m.persistTerminal(job)
	}
	m.workers.Wait()
}

func (m *Manager) worker() {
	defer m.workers.Done()
	for {
		select {
		case <-m.ctx.Done():
			return
		case id := <-m.queue:
			m.runJob(id)
		}
	}
}

func (m *Manager) runJob(id string) {
	m.mu.Lock()
	record, ok := m.jobs[id]
	if !ok || record.job.Status.Terminal() {
		m.mu.Unlock()
		return
	}
	for m.active >= m.concurrency && !m.closed && record.ctx.Err() == nil {
		m.cond.Wait()
	}
	if record.job.Status.Terminal() {
		m.mu.Unlock()
		return
	}
	if record.ctx.Err() != nil || m.closed {
		job := record.job
		if !job.Status.Terminal() {
			job.Status = StatusCancelled
			job.Error = "Cancelled before the download started"
			job.UpdatedAt = timestampNow()
			record.job = job
		}
		event, aggregate := m.eventLocked(id)
		m.mu.Unlock()
		m.persistTerminal(job)
		m.publishEvent(event, aggregate)
		return
	}
	m.active++
	record.job.Status = StatusDownloading
	record.job.UpdatedAt = timestampNow()
	request := DownloadRequest{
		Videos:    []VideoRequest{{URL: record.job.URL, Filename: record.job.OutputName}},
		OutputDir: record.job.OutputDir, Format: record.job.Format,
		AudioQuality: record.job.AudioQuality,
	}
	jobCtx := record.ctx
	event, aggregate := m.eventLocked(id)
	m.mu.Unlock()
	m.publishEvent(event, aggregate)

	outputPath, err := m.download(jobCtx, id, request)

	m.mu.Lock()
	m.active--
	m.cond.Broadcast()
	current := m.jobs[id]
	if current == nil || current.job.Status.Terminal() {
		m.mu.Unlock()
		return
	}
	job := current.job
	if err != nil {
		if errors.Is(err, context.Canceled) || jobCtx.Err() != nil {
			job.Status = StatusCancelled
			job.Error = "Cancelled by user"
		} else {
			job.Status = StatusFailed
			job.Error = redactError(err.Error(), job.URL)
		}
	} else if info, statErr := os.Stat(outputPath); statErr != nil {
		job.Status = StatusFailed
		job.Error = "Download finished but the output file could not be verified: " + statErr.Error()
	} else {
		job.Status = StatusCompleted
		job.Progress.Percent = floatPointer(100)
		job.OutputPath = outputPath
		job.OutputSize = info.Size()
	}
	job.UpdatedAt = timestampNow()
	current.job = job
	event, aggregate = m.eventLocked(id)
	m.mu.Unlock()
	m.persistTerminal(job)
	m.publishEvent(event, aggregate)
}

func (m *Manager) download(ctx context.Context, id string, req DownloadRequest) (string, error) {
	if len(req.Videos) != 1 {
		return "", errors.New("a download process requires exactly one video")
	}
	tools, err := m.resolver.Resolve()
	if err != nil {
		return "", err
	}
	if tools.YTDLP == "" || tools.FFmpeg == "" {
		return "", errors.New("yt-dlp and ffmpeg are required")
	}
	m.mu.Lock()
	record := m.jobs[id]
	if record == nil {
		m.mu.Unlock()
		return "", errors.New("download job was not found")
	}
	name := record.job.OutputName
	m.mu.Unlock()
	outputTemplate := filepath.Join(req.OutputDir, name+".%(ext)s")
	args := []string{
		"--ignore-config",
		"--no-playlist",
		"--no-overwrites",
		"--newline",
		"--progress",
		"--progress-template", progressTemplate(),
		"--print", "after_move:" + outputPrefix + "%(filepath)s",
		"--extract-audio",
		"--audio-format", string(req.Format),
		"--audio-quality", req.AudioQuality,
		"--ffmpeg-location", tools.FFmpeg,
		"--output", outputTemplate,
		"--",
		req.Videos[0].URL,
	}
	cmd := exec.CommandContext(ctx, tools.YTDLP, args...)
	platform.ConfigureProcessTree(cmd)
	cmd.Cancel = func() error { return platform.TerminateProcessTree(cmd) }
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("prepare yt-dlp output: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("prepare yt-dlp errors: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("start yt-dlp: %w", err)
	}

	var readers sync.WaitGroup
	var outputMu sync.Mutex
	var readerErrorMu sync.Mutex
	var readerError error
	var outputPath string
	var errorText = &tailBuffer{limit: 8192}
	recordReaderError := func(err error) {
		if err == nil {
			return
		}
		readerErrorMu.Lock()
		firstError := readerError == nil
		if firstError {
			readerError = err
		}
		readerErrorMu.Unlock()
		if firstError {
			_ = platform.TerminateProcessTree(cmd)
		}
	}
	readers.Add(2)
	go func() {
		defer readers.Done()
		if err := m.readProcessLines(stdout, func(line string) {
			if strings.HasPrefix(line, outputPrefix) {
				outputMu.Lock()
				outputPath = strings.TrimPrefix(line, outputPrefix)
				outputMu.Unlock()
				return
			}
			if progress, ok := ParseProgress(line); ok {
				m.updateProgress(id, progress)
				return
			}
			if strings.Contains(line, "[ExtractAudio]") || strings.Contains(strings.ToLower(line), "postprocess") {
				m.updateStatus(id, StatusConverting)
			}
		}); err != nil {
			recordReaderError(err)
		}
	}()
	go func() {
		defer readers.Done()
		if err := m.readProcessLines(stderr, func(line string) {
			errorText.Write([]byte(line + "\n"))
			if strings.HasPrefix(line, outputPrefix) {
				outputMu.Lock()
				outputPath = strings.TrimPrefix(line, outputPrefix)
				outputMu.Unlock()
				return
			}
			if progress, ok := ParseProgress(line); ok {
				m.updateProgress(id, progress)
				return
			}
			if strings.Contains(line, "[ExtractAudio]") || strings.Contains(strings.ToLower(line), "postprocess") {
				m.updateStatus(id, StatusConverting)
			}
		}); err != nil {
			recordReaderError(err)
		}
	}()
	waitErr := cmd.Wait()
	readers.Wait()
	readerErrorMu.Lock()
	readErr := readerError
	readerErrorMu.Unlock()
	if readErr != nil {
		return "", fmt.Errorf("read yt-dlp output: %w", readErr)
	}
	if waitErr != nil {
		message := strings.TrimSpace(errorText.String())
		if message == "" {
			return "", waitErr
		}
		return "", fmt.Errorf("yt-dlp failed: %s", message)
	}
	outputMu.Lock()
	resolvedOutput := outputPath
	outputMu.Unlock()
	if resolvedOutput == "" {
		return "", errors.New("yt-dlp completed without reporting the output file")
	}
	absOutput, err := filepath.Abs(resolvedOutput)
	if err != nil {
		return "", fmt.Errorf("resolve output file: %w", err)
	}
	absDir, err := filepath.Abs(req.OutputDir)
	if err != nil {
		return "", fmt.Errorf("resolve output folder: %w", err)
	}
	expectedOutput := filepath.Join(absDir, name+"."+string(req.Format))
	if !samePath(absOutput, expectedOutput) {
		return "", errors.New("yt-dlp reported an unexpected output path")
	}
	info, err := os.Stat(absOutput)
	if err != nil {
		return "", fmt.Errorf("download finished but output file is missing: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("download output is not a regular file")
	}
	return absOutput, nil
}

func (m *Manager) checkOutputNameLocked(outputDir, name string, format AudioFormat) error {
	candidate := filepath.Join(outputDir, name+"."+string(format))
	if _, err := os.Lstat(candidate); err == nil {
		return fmt.Errorf("output file %q already exists; choose another filename", filepath.Base(candidate))
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check output filename: %w", err)
	}
	for _, record := range m.jobs {
		if record.job.Status.Terminal() {
			continue
		}
		activePath := filepath.Join(record.job.OutputDir, record.job.OutputName+"."+string(record.job.Format))
		if samePath(candidate, activePath) {
			return fmt.Errorf("another queued download is using the filename %q", filepath.Base(candidate))
		}
	}
	return nil
}

func samePath(left, right string) bool {
	left = filepath.Clean(left)
	right = filepath.Clean(right)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}

func (m *Manager) readProcessLines(reader io.Reader, consume func(string)) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		consume(scanner.Text())
	}
	return scanner.Err()
}

func (m *Manager) updateProgress(id string, progress Progress) {
	m.mu.Lock()
	record := m.jobs[id]
	if record == nil || record.job.Status.Terminal() {
		m.mu.Unlock()
		return
	}
	record.job.Progress = progress
	record.job.Status = StatusDownloading
	record.job.UpdatedAt = timestampNow()
	event, aggregate := m.eventLocked(id)
	m.mu.Unlock()
	m.publishEvent(event, aggregate)
}

func (m *Manager) updateStatus(id string, status JobStatus) {
	m.mu.Lock()
	record := m.jobs[id]
	if record == nil || record.job.Status.Terminal() {
		m.mu.Unlock()
		return
	}
	record.job.Status = status
	if status == StatusConverting {
		record.job.Progress = Progress{}
	}
	record.job.UpdatedAt = timestampNow()
	event, aggregate := m.eventLocked(id)
	m.mu.Unlock()
	m.publishEvent(event, aggregate)
}

func (m *Manager) eventLocked(id string) (JobEvent, AggregateProgress) {
	jobs := m.snapshotsLocked()
	aggregate := aggregateProgress(jobs)
	for _, job := range jobs {
		if job.ID == id {
			return JobEvent{Job: job, Aggregate: aggregate}, aggregate
		}
	}
	return JobEvent{Aggregate: aggregate}, aggregate
}

func (m *Manager) snapshotsLocked() []Job {
	jobs := make([]Job, 0, len(m.order))
	for _, id := range m.order {
		if record := m.jobs[id]; record != nil {
			jobs = append(jobs, record.job)
		}
	}
	return jobs
}

func (m *Manager) publishEvent(event JobEvent, _ AggregateProgress) {
	if m.events != nil {
		m.events(event)
	}
}

func (m *Manager) persistTerminal(job Job) {
	if m.history == nil || !job.Status.Terminal() {
		return
	}
	entry := HistoryEntry{
		ID: job.ID, URL: job.URL, OutputDir: job.OutputDir, OutputPath: job.OutputPath,
		Format: job.Format, Status: job.Status, OutputSize: job.OutputSize,
		Error: job.Error, FinishedAt: job.UpdatedAt,
	}
	_ = m.history.SaveHistory(context.Background(), entry)
}

func newID() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

func floatPointer(value float64) *float64 { return &value }

var urlInError = regexp.MustCompile(`(?i)https?://[^ "'<>]+`)

func redactedURL(value string) string {
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "[video URL]"
	}
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	parsed.RawFragment = ""
	return parsed.String()
}

func redactError(message, rawURL string) string {
	message = strings.ReplaceAll(message, rawURL, redactedURL(rawURL))
	message = urlInError.ReplaceAllStringFunc(message, func(value string) string {
		trimmed := strings.TrimRight(value, ".,;:)]}")
		return redactedURL(trimmed) + value[len(trimmed):]
	})
	message = strings.TrimSpace(message)
	if len(message) > 1000 {
		message = message[:1000] + "…"
	}
	if message == "" {
		return "Download failed. Check the URL, media-tool setup, and destination folder."
	}
	return message
}

type tailBuffer struct {
	mu    sync.Mutex
	data  []byte
	limit int
}

func (b *tailBuffer) Write(value []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = append(b.data, value...)
	if len(b.data) > b.limit {
		b.data = append([]byte(nil), b.data[len(b.data)-b.limit:]...)
	}
	return len(value), nil
}

func (b *tailBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.data)
}
