package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/thiagosantos/yumdown/internal/downloader"
	"github.com/thiagosantos/yumdown/internal/media"
	"github.com/thiagosantos/yumdown/internal/persistence"
	"github.com/thiagosantos/yumdown/internal/platform"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	mu           sync.RWMutex
	ctx          context.Context
	store        *persistence.Store
	manager      *downloader.Manager
	resolver     *media.Resolver
	settings     downloader.Settings
	capabilities downloader.Capabilities
	configDir    string
	initErr      error
}

func NewApp() *App { return &App{} }

func (a *App) startup(ctx context.Context) {
	a.mu.Lock()
	a.ctx = ctx
	a.mu.Unlock()

	configRoot, err := os.UserConfigDir()
	if err != nil {
		a.setInitError(fmt.Errorf("locate application data folder: %w", err))
		return
	}
	configDir := filepath.Join(configRoot, "Yumdown")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		a.setInitError(fmt.Errorf("create application data folder: %w", err))
		return
	}
	database, err := persistence.Open(filepath.Join(configDir, "yumdown.sqlite"))
	if err != nil {
		a.setInitError(err)
		return
	}
	settings, err := database.GetSettings(ctx)
	if err != nil {
		_ = database.Close()
		a.setInitError(err)
		return
	}
	if settings.OutputDir == "" {
		settings.OutputDir, err = defaultOutputDirectory()
		if err != nil {
			_ = database.Close()
			a.setInitError(err)
			return
		}
	}
	if !downloader.IsKnownFormat(settings.Format) {
		settings.Format = downloader.FormatMP3
	}
	if settings.Concurrency < 1 || settings.Concurrency > 4 {
		settings.Concurrency = 2
	}
	if !downloader.IsAudioQuality(settings.AudioQuality) {
		settings.AudioQuality = "192K"
	}
	if err := database.SaveSettings(ctx, settings); err != nil {
		_ = database.Close()
		a.setInitError(err)
		return
	}
	executable, _ := os.Executable()
	resolver := media.NewResolver(configDir, executable)

	a.mu.Lock()
	a.store = database
	a.settings = settings
	a.configDir = configDir
	a.resolver = resolver
	a.manager = downloader.NewManager(ctx, resolver, database, func() []downloader.AudioFormat {
		a.mu.RLock()
		defer a.mu.RUnlock()
		return append([]downloader.AudioFormat(nil), a.capabilities.Formats...)
	}, settings.Concurrency, func(event downloader.JobEvent) {
		wailsruntime.EventsEmit(ctx, "download:job", event)
	})
	a.mu.Unlock()
}

func (a *App) shutdown(context.Context) {
	a.mu.Lock()
	manager := a.manager
	database := a.store
	a.manager = nil
	a.store = nil
	a.mu.Unlock()
	if manager != nil {
		manager.Close()
	}
	if database != nil {
		_ = database.Close()
	}
}

func (a *App) StartDownloads(request downloader.DownloadRequest) ([]downloader.Job, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	capabilities := a.GetCapabilities()
	if !capabilities.ToolsAvailable {
		return nil, errors.New(capabilities.Message)
	}
	a.mu.RLock()
	settings := a.settings
	manager := a.manager
	a.mu.RUnlock()
	if request.OutputDir == "" {
		request.OutputDir = settings.OutputDir
	}
	if request.Format == "" {
		request.Format = settings.Format
	}
	if request.AudioQuality == "" {
		request.AudioQuality = settings.AudioQuality
	}
	return manager.StartBatch(request)
}

func (a *App) GetJobs() ([]downloader.Job, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	a.mu.RLock()
	manager := a.manager
	a.mu.RUnlock()
	return manager.Jobs(), nil
}

func (a *App) CancelDownload(id string) error {
	if err := a.ensureReady(); err != nil {
		return err
	}
	a.mu.RLock()
	manager := a.manager
	a.mu.RUnlock()
	return manager.Cancel(id)
}

func (a *App) GetHistory(search, status string) ([]downloader.HistoryEntry, error) {
	if err := a.ensureReady(); err != nil {
		return nil, err
	}
	if status == "all" {
		status = ""
	}
	switch downloader.JobStatus(status) {
	case "", downloader.StatusCompleted, downloader.StatusFailed, downloader.StatusCancelled:
	default:
		return nil, errors.New("invalid history status filter")
	}
	a.mu.RLock()
	database := a.store
	ctx := a.ctx
	a.mu.RUnlock()
	return database.ListHistory(ctx, strings.TrimSpace(search), status, 200)
}

func (a *App) ClearHistory() error {
	if err := a.ensureReady(); err != nil {
		return err
	}
	a.mu.RLock()
	database := a.store
	ctx := a.ctx
	a.mu.RUnlock()
	return database.ClearHistory(ctx)
}

func (a *App) GetSettings() (downloader.Settings, error) {
	if err := a.ensureReady(); err != nil {
		return downloader.Settings{}, err
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.settings, nil
}

func (a *App) SaveSettings(settings downloader.Settings) (downloader.Settings, error) {
	if err := a.ensureReady(); err != nil {
		return downloader.Settings{}, err
	}
	if !downloader.IsKnownFormat(settings.Format) {
		return downloader.Settings{}, downloader.ErrInvalidFormat
	}
	if settings.Concurrency < 1 || settings.Concurrency > 4 {
		return downloader.Settings{}, errors.New("concurrency must be between 1 and 4")
	}
	if !downloader.IsAudioQuality(settings.AudioQuality) {
		return downloader.Settings{}, downloader.ErrInvalidQuality
	}
	if strings.TrimSpace(settings.OutputDir) == "" {
		return downloader.Settings{}, errors.New("choose an output folder")
	}
	outputDir, err := filepath.Abs(settings.OutputDir)
	if err != nil {
		return downloader.Settings{}, fmt.Errorf("resolve output folder: %w", err)
	}
	info, err := os.Stat(outputDir)
	if err != nil || !info.IsDir() {
		return downloader.Settings{}, errors.New("output folder does not exist")
	}
	settings.OutputDir = outputDir
	a.mu.RLock()
	database := a.store
	manager := a.manager
	ctx := a.ctx
	a.mu.RUnlock()
	if err := database.SaveSettings(ctx, settings); err != nil {
		return downloader.Settings{}, err
	}
	if err := manager.SetConcurrency(settings.Concurrency); err != nil {
		return downloader.Settings{}, err
	}
	a.mu.Lock()
	a.settings = settings
	a.mu.Unlock()
	return settings, nil
}

func (a *App) GetCapabilities() downloader.Capabilities {
	a.mu.RLock()
	resolver := a.resolver
	ctx := a.ctx
	a.mu.RUnlock()
	if resolver == nil || ctx == nil {
		return downloader.Capabilities{Formats: []downloader.AudioFormat{}, Message: "The application is still starting."}
	}
	capabilities := resolver.Capabilities(ctx)
	a.mu.Lock()
	a.capabilities = capabilities
	a.mu.Unlock()
	return capabilities
}

func (a *App) SelectOutputFolder() (string, error) {
	if err := a.ensureReady(); err != nil {
		return "", err
	}
	a.mu.RLock()
	ctx := a.ctx
	defaultDirectory := a.settings.OutputDir
	a.mu.RUnlock()
	return wailsruntime.OpenDirectoryDialog(ctx, wailsruntime.OpenDialogOptions{
		Title: "Choose an audio output folder", DefaultDirectory: defaultDirectory,
	})
}

func (a *App) OpenFolder(path string) error {
	if err := a.ensureReady(); err != nil {
		return err
	}
	return platform.OpenFolder(path)
}

func (a *App) OpenOutputFolder() error {
	if err := a.ensureReady(); err != nil {
		return err
	}
	a.mu.RLock()
	path := a.settings.OutputDir
	a.mu.RUnlock()
	return platform.OpenFolder(path)
}

func (a *App) StartupError() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.initErr == nil {
		return ""
	}
	return a.initErr.Error()
}

func (a *App) ensureReady() error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.initErr != nil {
		return a.initErr
	}
	if a.store == nil || a.manager == nil {
		return errors.New("application services are not ready")
	}
	return nil
}

func (a *App) setInitError(err error) {
	a.mu.Lock()
	a.initErr = err
	a.mu.Unlock()
}

func defaultOutputDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home folder: %w", err)
	}
	path := filepath.Join(home, "Downloads", "Yumdown")
	if err := os.MkdirAll(path, 0o755); err != nil {
		return "", fmt.Errorf("create default output folder: %w", err)
	}
	return path, nil
}
