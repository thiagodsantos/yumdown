package downloader

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateURLAcceptsHTTPAndHTTPSOnly(t *testing.T) {
	for _, value := range []string{"https://video.example/watch?id=1", "http://video.example/clip"} {
		if err := ValidateURL(value); err != nil {
			t.Errorf("ValidateURL(%q) returned %v", value, err)
		}
	}
	for _, value := range []string{"", " video.example/watch", "file:///tmp/audio.mp3", "ftp://video.example/a", "https://user:pass@video.example/a", "https://video.example/a b"} {
		if err := ValidateURL(value); !errors.Is(err, ErrInvalidURL) {
			t.Errorf("ValidateURL(%q) error = %v, want ErrInvalidURL", value, err)
		}
	}
}

func TestValidateRequestNormalizesAndChecksOutputFolder(t *testing.T) {
	folder := t.TempDir()
	request, err := ValidateRequest(DownloadRequest{
		Videos:    []VideoRequest{{URL: "https://video.example/watch?id=1", Filename: "  sample audio  "}},
		OutputDir: folder, Format: FormatMP3,
	}, []AudioFormat{FormatMP3})
	if err != nil {
		t.Fatalf("ValidateRequest returned error: %v", err)
	}
	if request.Videos[0].Filename != "sample audio" || request.AudioQuality != "192K" {
		t.Fatalf("normalized request = %#v", request)
	}
	if request.OutputDir != folder {
		t.Fatalf("output directory = %q, want %q", request.OutputDir, folder)
	}
	entries, err := os.ReadDir(folder)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("write check left temporary entries: %v", entries)
	}
}

func TestValidateRequestStripsSelectedAudioExtension(t *testing.T) {
	request, err := ValidateRequest(DownloadRequest{
		Videos:    []VideoRequest{{URL: "https://video.example/watch", Filename: "recording.MP3"}},
		OutputDir: t.TempDir(), Format: FormatMP3,
	}, []AudioFormat{FormatMP3})
	if err != nil {
		t.Fatal(err)
	}
	if request.Videos[0].Filename != "recording" {
		t.Fatalf("filename = %q, want extension-free base name", request.Videos[0].Filename)
	}
}

func TestNormalizeFilenameRejectsPathsAndTemplateSyntax(t *testing.T) {
	for _, name := range []string{"../escape", "folder/name", `folder\name`, "CON", "name%", "bad:name"} {
		if _, err := NormalizeFilename(name); !errors.Is(err, ErrInvalidFilename) {
			t.Errorf("NormalizeFilename(%q) error = %v, want ErrInvalidFilename", name, err)
		}
	}
	if _, err := NormalizeFilename("A nice title — live"); err != nil {
		t.Fatalf("valid Unicode filename rejected: %v", err)
	}
}

func TestValidateRequestRejectsUnsupportedFormatAndMissingFolder(t *testing.T) {
	if _, err := ValidateRequest(DownloadRequest{
		Videos: []VideoRequest{{URL: "https://video.example/watch"}}, OutputDir: t.TempDir(), Format: FormatFLAC,
	}, []AudioFormat{FormatMP3}); !errors.Is(err, ErrInvalidFormat) {
		t.Fatalf("format error = %v, want ErrInvalidFormat", err)
	}
	missing := filepath.Join(t.TempDir(), "not-created")
	if _, err := ValidateRequest(DownloadRequest{
		Videos: []VideoRequest{{URL: "https://video.example/watch"}}, OutputDir: missing, Format: FormatMP3,
	}, []AudioFormat{FormatMP3}); err == nil {
		t.Fatal("expected missing destination to be rejected")
	}
}

func TestAggregateProgressIsItemWeightedAndCanBeIndeterminate(t *testing.T) {
	progress := 40.0
	jobs := []Job{
		{Status: StatusDownloading, Progress: Progress{Percent: &progress}},
		{Status: StatusQueued},
	}
	aggregate := aggregateProgress(jobs)
	if aggregate.Indeterminate || aggregate.Percent == nil || *aggregate.Percent != 20 {
		t.Fatalf("aggregate = %#v, want determinate 20%% item progress", aggregate)
	}
	jobs[0].Progress.Percent = nil
	aggregate = aggregateProgress(jobs)
	if !aggregate.Indeterminate || aggregate.Percent != nil {
		t.Fatalf("unknown active progress should be indeterminate: %#v", aggregate)
	}
}
