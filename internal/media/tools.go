package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/thiagosantos/yumdown/internal/downloader"
)

type Resolver struct {
	ExecutablePath string
	ConfigDir      string
	ExplicitDir    string
	GOOS           string
	GOARCH         string
}

func NewResolver(configDir, executablePath string) *Resolver {
	return &Resolver{
		ExecutablePath: executablePath,
		ConfigDir:      configDir,
		ExplicitDir:    os.Getenv("YUMDOWN_TOOLS_DIR"),
		GOOS:           runtime.GOOS,
		GOARCH:         runtime.GOARCH,
	}
}

func (r *Resolver) Resolve() (downloader.ToolPaths, error) {
	if !supportedPlatform(r.GOOS, r.GOARCH) {
		return downloader.ToolPaths{}, fmt.Errorf("media tools are not configured for %s/%s", r.GOOS, r.GOARCH)
	}
	ytName, ffmpegName, ffprobeName := binaryNames(r.GOOS)
	var candidates []string
	if r.ExecutablePath != "" {
		candidates = append(candidates, filepath.Join(filepath.Dir(r.ExecutablePath), "tools"))
	}
	var explicitDirErr error
	if r.ExplicitDir != "" {
		if filepath.IsAbs(r.ExplicitDir) {
			candidates = append(candidates, filepath.Clean(r.ExplicitDir))
		} else {
			explicitDirErr = errors.New("YUMDOWN_TOOLS_DIR must be an absolute path")
		}
	}
	if r.ConfigDir != "" {
		candidates = append(candidates, filepath.Join(r.ConfigDir, "tools"))
	}
	for _, dir := range candidates {
		yt := filepath.Join(dir, ytName)
		ffmpeg := filepath.Join(dir, ffmpegName)
		if isExecutable(yt) && isExecutable(ffmpeg) {
			ffprobe := filepath.Join(dir, ffprobeName)
			if !isExecutable(ffprobe) {
				ffprobe = ""
			}
			return downloader.ToolPaths{YTDLP: yt, FFmpeg: ffmpeg, FFprobe: ffprobe, Dir: dir}, nil
		}
	}
	if explicitDirErr != nil {
		return downloader.ToolPaths{}, explicitDirErr
	}
	return downloader.ToolPaths{}, fmt.Errorf("yt-dlp and ffmpeg were not found; place both executables in %s/tools or set YUMDOWN_TOOLS_DIR to an absolute tools directory", r.ConfigDir)
}

func (r *Resolver) Capabilities(ctx context.Context) downloader.Capabilities {
	result := downloader.Capabilities{Formats: []downloader.AudioFormat{}}
	binaries, err := r.Resolve()
	if err != nil {
		result.Message = err.Error()
		return result
	}
	result.ToolsDirectory = binaries.Dir
	version, err := runVersion(ctx, binaries.YTDLP, "--version")
	if err != nil {
		result.Message = "yt-dlp could not be started: " + err.Error()
		return result
	}
	result.YTDLPVersion = version
	ffmpegVersion, err := runVersion(ctx, binaries.FFmpeg, "-version")
	if err != nil {
		result.Message = "ffmpeg could not be started: " + err.Error()
		return result
	}
	result.FFmpegVersion = ffmpegVersion
	result.FFprobeFound = binaries.FFprobe != ""
	formatsOutput, err := runOutput(ctx, binaries.FFmpeg, "-hide_banner", "-encoders")
	if err != nil {
		result.Message = "ffmpeg encoder capabilities could not be read: " + err.Error()
		return result
	}
	for _, format := range downloader.SupportedFormats {
		if supportsFormat(formatsOutput, format) {
			result.Formats = append(result.Formats, format)
		}
	}
	result.ToolsAvailable = true
	if len(result.Formats) == 0 {
		result.ToolsAvailable = false
		result.Message = "ffmpeg is available but no supported audio encoders were detected"
	}
	return result
}

func supportsFormat(output string, format downloader.AudioFormat) bool {
	checks := map[downloader.AudioFormat][]string{
		downloader.FormatMP3:  {"libmp3lame", " mp3 "},
		downloader.FormatM4A:  {" aac "},
		downloader.FormatOpus: {"libopus", " opus "},
		downloader.FormatWAV:  {"pcm_s16le"},
		downloader.FormatFLAC: {" flac "},
	}
	lower := strings.ToLower(output)
	for _, marker := range checks[format] {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func supportedPlatform(goos, goarch string) bool {
	if goarch != "amd64" && goarch != "arm64" {
		return false
	}
	return goos == "windows" || goos == "darwin" || goos == "linux"
}

func binaryNames(goos string) (string, string, string) {
	if goos == "windows" {
		return "yt-dlp.exe", "ffmpeg.exe", "ffprobe.exe"
	}
	return "yt-dlp", "ffmpeg", "ffprobe"
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode().Perm()&0o111 != 0
}

func runVersion(ctx context.Context, binary, arg string) (string, error) {
	output, err := runOutput(ctx, binary, arg)
	if err != nil {
		return "", err
	}
	line := strings.TrimSpace(strings.SplitN(output, "\n", 2)[0])
	if line == "" {
		return "", errors.New("tool returned an empty version")
	}
	return line, nil
}

func runOutput(parent context.Context, binary string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, args...)
	var output synchronizedBuffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(output.String()))
	}
	return output.String(), nil
}

type synchronizedBuffer struct {
	mu sync.Mutex
	bytes.Buffer
}

func (b *synchronizedBuffer) Write(value []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.Write(value)
}

func (b *synchronizedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Buffer.String()
}
