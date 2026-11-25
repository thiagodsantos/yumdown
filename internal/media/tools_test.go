package media

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thiagosantos/yumdown/internal/downloader"
)

func TestBinaryNamesByOperatingSystem(t *testing.T) {
	if yt, ffmpeg, ffprobe := binaryNames("windows"); yt != "yt-dlp.exe" || ffmpeg != "ffmpeg.exe" || ffprobe != "ffprobe.exe" {
		t.Fatalf("windows binary names = %q, %q, %q", yt, ffmpeg, ffprobe)
	}
	if yt, ffmpeg, ffprobe := binaryNames("linux"); yt != "yt-dlp" || ffmpeg != "ffmpeg" || ffprobe != "ffprobe" {
		t.Fatalf("linux binary names = %q, %q, %q", yt, ffmpeg, ffprobe)
	}
}

func TestResolverUsesExplicitDirectoryAndNeverFallsBackToPATH(t *testing.T) {
	root := t.TempDir()
	yt := filepath.Join(root, "yt-dlp")
	ffmpeg := filepath.Join(root, "ffmpeg")
	for _, path := range []string{yt, ffmpeg} {
		if err := os.WriteFile(path, []byte("fake"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	resolver := &Resolver{ExplicitDir: root, GOOS: "linux", GOARCH: "amd64"}
	resolved, err := resolver.Resolve()
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if resolved.YTDLP != yt || resolved.FFmpeg != ffmpeg || resolved.Dir != root {
		t.Fatalf("resolved tools = %#v", resolved)
	}

	missing := &Resolver{ConfigDir: t.TempDir(), GOOS: "linux", GOARCH: "amd64"}
	if _, err := missing.Resolve(); err == nil {
		t.Fatal("missing explicit tools should not fall back to PATH")
	}
}

func TestResolverPrefersPackagedThenExplicitThenAppConfigTools(t *testing.T) {
	writeTools := func(dir string) {
		t.Helper()
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"yt-dlp", "ffmpeg"} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte("fake"), 0o700); err != nil {
				t.Fatal(err)
			}
		}
	}

	appDir := t.TempDir()
	packagedDir := filepath.Join(appDir, "tools")
	configDir := t.TempDir()
	configTools := filepath.Join(configDir, "tools")
	explicitDir := t.TempDir()
	writeTools(packagedDir)
	writeTools(configTools)
	writeTools(explicitDir)
	resolver := &Resolver{
		ExecutablePath: filepath.Join(appDir, "yumdown"), ConfigDir: configDir,
		ExplicitDir: explicitDir, GOOS: "linux", GOARCH: "amd64",
	}

	resolved, err := resolver.Resolve()
	if err != nil || resolved.Dir != packagedDir {
		t.Fatalf("packaged tool priority: resolved=%#v error=%v", resolved, err)
	}
	for _, name := range []string{"yt-dlp", "ffmpeg"} {
		if err := os.Remove(filepath.Join(packagedDir, name)); err != nil {
			t.Fatal(err)
		}
	}
	resolved, err = resolver.Resolve()
	if err != nil || resolved.Dir != explicitDir {
		t.Fatalf("explicit tool priority: resolved=%#v error=%v", resolved, err)
	}
	for _, name := range []string{"yt-dlp", "ffmpeg"} {
		if err := os.Remove(filepath.Join(explicitDir, name)); err != nil {
			t.Fatal(err)
		}
	}
	resolved, err = resolver.Resolve()
	if err != nil || resolved.Dir != configTools {
		t.Fatalf("app-config tool fallback: resolved=%#v error=%v", resolved, err)
	}
}

func TestResolverRejectsUnsupportedPlatforms(t *testing.T) {
	resolver := &Resolver{GOOS: "plan9", GOARCH: "amd64"}
	if _, err := resolver.Resolve(); err == nil {
		t.Fatal("expected unsupported platform error")
	}
}

func TestSupportedFormatDetectionUsesEncoderNames(t *testing.T) {
	output := ` A..... libmp3lame MP3 (MPEG audio layer 3)
 A..... aac AAC (Advanced Audio Coding)
 A..... libopus Opus`
	for _, format := range []downloader.AudioFormat{downloader.FormatMP3, downloader.FormatM4A, downloader.FormatOpus} {
		if !supportsFormat(output, format) {
			t.Errorf("expected %s to be detected", format)
		}
	}
	if supportsFormat(output, downloader.FormatFLAC) {
		t.Fatal("FLAC encoder should not be inferred when absent")
	}
}
