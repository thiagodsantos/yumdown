package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/thiagosantos/yumdown/internal/downloader"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open local database: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("configure local database: %w", err)
	}
	store := &Store{db: db}
	if err := store.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) migrate(ctx context.Context) error {
	const schema = `
CREATE TABLE IF NOT EXISTS settings (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS history (
  id TEXT PRIMARY KEY,
  url TEXT NOT NULL,
  output_dir TEXT NOT NULL,
  output_path TEXT NOT NULL DEFAULT '',
  format TEXT NOT NULL,
  status TEXT NOT NULL,
  output_size INTEGER NOT NULL DEFAULT 0,
  error TEXT NOT NULL DEFAULT '',
  finished_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS history_finished_at_idx ON history(finished_at DESC);
`
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("migrate local database: %w", err)
	}
	return nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) GetSettings(ctx context.Context) (downloader.Settings, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE id = 1`).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return downloader.Settings{}, nil
	}
	if err != nil {
		return downloader.Settings{}, fmt.Errorf("read settings: %w", err)
	}
	var settings downloader.Settings
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		return downloader.Settings{}, fmt.Errorf("decode settings: %w", err)
	}
	return settings, nil
}

func (s *Store) SaveSettings(ctx context.Context, settings downloader.Settings) error {
	raw, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO settings(id, value) VALUES(1, ?)
ON CONFLICT(id) DO UPDATE SET value = excluded.value`, string(raw))
	if err != nil {
		return fmt.Errorf("save settings: %w", err)
	}
	return nil
}

func (s *Store) SaveHistory(ctx context.Context, entry downloader.HistoryEntry) error {
	finishedAt, err := time.Parse(time.RFC3339Nano, entry.FinishedAt)
	if err != nil {
		return fmt.Errorf("parse download history timestamp: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `INSERT OR REPLACE INTO history
(id, url, output_dir, output_path, format, status, output_size, error, finished_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`, entry.ID, entry.URL, entry.OutputDir, entry.OutputPath,
		string(entry.Format), string(entry.Status), entry.OutputSize, entry.Error, finishedAt.UTC().Format(downloader.TimestampLayout))
	if err != nil {
		return fmt.Errorf("save download history: %w", err)
	}
	return nil
}

func (s *Store) ListHistory(ctx context.Context, search, status string, limit int) ([]downloader.HistoryEntry, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	query := `SELECT id, url, output_dir, output_path, format, status, output_size, error, finished_at
FROM history WHERE (? = '' OR status = ?) AND (? = '' OR url LIKE ? ESCAPE char(92) OR output_path LIKE ? ESCAPE char(92) OR error LIKE ? ESCAPE char(92))
ORDER BY finished_at DESC LIMIT ?`
	needle := "%" + strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(search) + "%"
	rows, err := s.db.QueryContext(ctx, query, status, status, search, needle, needle, needle, limit)
	if err != nil {
		return nil, fmt.Errorf("query download history: %w", err)
	}
	defer rows.Close()
	entries := make([]downloader.HistoryEntry, 0)
	for rows.Next() {
		var entry downloader.HistoryEntry
		var format, state, finished string
		if err := rows.Scan(&entry.ID, &entry.URL, &entry.OutputDir, &entry.OutputPath, &format, &state,
			&entry.OutputSize, &entry.Error, &finished); err != nil {
			return nil, fmt.Errorf("read download history: %w", err)
		}
		entry.Format = downloader.AudioFormat(format)
		entry.Status = downloader.JobStatus(state)
		entry.FinishedAt = finished
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate download history: %w", err)
	}
	return entries, nil
}

func (s *Store) ClearHistory(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM history`); err != nil {
		return fmt.Errorf("clear download history: %w", err)
	}
	return nil
}
