package downloader

import (
	"strings"
	"testing"
)

func TestProgressTemplateUsesStructuredYTDLPFields(t *testing.T) {
	template := progressTemplate()
	for _, field := range []string{
		progressPrefix,
		"%(progress.downloaded_bytes)s",
		"%(progress.total_bytes_estimate)s",
		"%(progress._percent_str)s",
		"%(progress.speed)s",
		"%(progress.eta)s",
	} {
		if !strings.Contains(template, field) {
			t.Errorf("progress template missing %q: %s", field, template)
		}
	}
}

func TestParseProgressUsesExactPrefixAndNormalizesFields(t *testing.T) {
	progress, ok := ParseProgress("__YUMDOWN_PROGRESS__50|100|120|41.6|12.5|4")
	if !ok {
		t.Fatal("expected progress record to be parsed")
	}
	if progress.Percent == nil || *progress.Percent != 41.6 || progress.DownloadedBytes != 50 || progress.TotalBytes != 100 {
		t.Fatalf("unexpected progress values: %#v", progress)
	}
	if progress.SpeedBytesPerSecond == nil || *progress.SpeedBytesPerSecond != 12.5 || progress.ETASeconds == nil || *progress.ETASeconds != 4 {
		t.Fatalf("unexpected transfer stats: %#v", progress)
	}
}

func TestParseProgressUsesEstimateAndLeavesUnknownValues(t *testing.T) {
	progress, ok := ParseProgress("__YUMDOWN_PROGRESS__25|NA|50|NA|NA|NA")
	if !ok {
		t.Fatal("expected progress record to be parsed")
	}
	if progress.TotalBytes != 50 || progress.Percent == nil || *progress.Percent != 50 {
		t.Fatalf("estimated progress = %#v", progress)
	}
	if _, ok := ParseProgress("ordinary yt-dlp output"); ok {
		t.Fatal("human-readable output should not be parsed as a progress record")
	}
	if _, ok := ParseProgress("__YUMDOWN_PROGRESS__1|2"); ok {
		t.Fatal("malformed progress record should be ignored")
	}
}
