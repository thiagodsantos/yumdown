package persistence

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/thiagosantos/yumdown/internal/downloader"
)

func TestStorePersistsSettingsAndSearchableHistory(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer store.Close()
	ctx := context.Background()

	settings := downloader.Settings{OutputDir: t.TempDir(), Format: downloader.FormatMP3, Concurrency: 2, AudioQuality: "192K"}
	if err := store.SaveSettings(ctx, settings); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.GetSettings(ctx)
	if err != nil || loaded != settings {
		t.Fatalf("settings = %#v, error = %v", loaded, err)
	}

	entry := downloader.HistoryEntry{
		ID: "job-1", URL: "https://video.example/watch?id=secret", OutputDir: settings.OutputDir,
		OutputPath: filepath.Join(settings.OutputDir, "track.mp3"), Format: downloader.FormatMP3,
		Status: downloader.StatusCompleted, OutputSize: 4096, FinishedAt: time.Now().UTC().Format(downloader.TimestampLayout),
	}
	if err := store.SaveHistory(ctx, entry); err != nil {
		t.Fatal(err)
	}
	entries, err := store.ListHistory(ctx, "secret", "completed", 20)
	if err != nil || len(entries) != 1 {
		t.Fatalf("history = %#v, error = %v", entries, err)
	}
	if entries[0].ID != entry.ID || entries[0].OutputSize != entry.OutputSize || entries[0].FinishedAt != entry.FinishedAt {
		t.Fatalf("persisted entry = %#v", entries[0])
	}
	entries, err = store.ListHistory(ctx, "%", "", 20)
	if err != nil || len(entries) != 0 {
		t.Fatalf("literal wildcard search = %#v, error = %v", entries, err)
	}
	if err := store.ClearHistory(ctx); err != nil {
		t.Fatal(err)
	}
	entries, err = store.ListHistory(ctx, "", "", 20)
	if err != nil || len(entries) != 0 {
		t.Fatalf("history after clear = %#v, error = %v", entries, err)
	}
}
