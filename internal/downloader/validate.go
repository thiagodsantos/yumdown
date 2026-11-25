package downloader

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrNoURLs          = errors.New("enter at least one video URL")
	ErrInvalidURL      = errors.New("only valid http and https video URLs are supported")
	ErrInvalidFormat   = errors.New("unsupported audio format")
	ErrInvalidQuality  = errors.New("unsupported audio quality")
	ErrInvalidFilename = errors.New("filename contains unsupported characters")
)

var qualityValues = map[string]struct{}{
	"64K": {}, "96K": {}, "128K": {}, "160K": {}, "192K": {}, "256K": {}, "320K": {},
}

var windowsReservedName = regexp.MustCompile(`(?i)^(con|prn|aux|nul|com[1-9]|lpt[1-9])(\..*)?$`)

func ValidateURL(raw string) error {
	if len(raw) > 4096 || strings.TrimSpace(raw) != raw || raw == "" {
		return ErrInvalidURL
	}
	for _, r := range raw {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return ErrInvalidURL
		}
	}
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed == nil || parsed.Hostname() == "" || parsed.User != nil || parsed.Opaque != "" {
		return ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrInvalidURL
	}
	return nil
}

func ValidateRequest(req DownloadRequest, supportedFormats []AudioFormat) (DownloadRequest, error) {
	if len(req.Videos) == 0 {
		return req, ErrNoURLs
	}
	if len(req.Videos) > 100 {
		return req, errors.New("a batch can contain at most 100 videos")
	}
	for i := range req.Videos {
		video := &req.Videos[i]
		if err := ValidateURL(video.URL); err != nil {
			return req, fmt.Errorf("video row %d: %w", i+1, err)
		}
		name, err := NormalizeFilename(video.Filename)
		if err != nil {
			return req, fmt.Errorf("video row %d: %w", i+1, err)
		}
		if extension := filepath.Ext(name); extension != "" {
			for _, format := range SupportedFormats {
				if strings.EqualFold(extension, "."+string(format)) {
					name = strings.TrimSuffix(name, extension)
					if name == "" {
						return req, fmt.Errorf("video row %d: %w", i+1, ErrInvalidFilename)
					}
					break
				}
			}
		}
		video.Filename = name
	}
	if !IsSupportedFormat(supportedFormats, req.Format) {
		return req, ErrInvalidFormat
	}
	if req.AudioQuality == "" {
		req.AudioQuality = "192K"
	}
	if !IsAudioQuality(req.AudioQuality) {
		return req, ErrInvalidQuality
	}

	if strings.TrimSpace(req.OutputDir) == "" {
		return req, errors.New("choose an output folder")
	}
	abs, err := filepath.Abs(req.OutputDir)
	if err != nil {
		return req, fmt.Errorf("resolve output folder: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return req, fmt.Errorf("output folder is unavailable: %w", err)
	}
	if !info.IsDir() {
		return req, errors.New("output location must be a folder")
	}
	probe, err := os.CreateTemp(abs, ".yumdown-write-check-*")
	if err != nil {
		return req, fmt.Errorf("output folder is not writable: %w", err)
	}
	probePath := probe.Name()
	if err := probe.Close(); err != nil {
		_ = os.Remove(probePath)
		return req, fmt.Errorf("check output folder: %w", err)
	}
	if err := os.Remove(probePath); err != nil {
		return req, fmt.Errorf("check output folder: %w", err)
	}
	req.OutputDir = abs
	return req, nil
}

func NormalizeFilename(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil
	}
	if strings.ContainsAny(name, `/\\<>:"|?*%`) || name == "." || name == ".." || strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") || windowsReservedName.MatchString(name) {
		return "", ErrInvalidFilename
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return "", ErrInvalidFilename
		}
	}
	return name, nil
}

func IsAudioQuality(value string) bool {
	_, ok := qualityValues[value]
	return ok
}

func IsKnownFormat(format AudioFormat) bool {
	return IsSupportedFormat(SupportedFormats, format)
}

func IsSupportedFormat(formats []AudioFormat, format AudioFormat) bool {
	for _, candidate := range formats {
		if candidate == format {
			return true
		}
	}
	return false
}

func aggregateProgress(jobs []Job) AggregateProgress {
	result := AggregateProgress{TotalItems: len(jobs), Indeterminate: len(jobs) > 0}
	if len(jobs) == 0 {
		result.Indeterminate = false
		zero := 0.0
		result.Percent = &zero
		return result
	}
	var completedUnits float64
	for _, job := range jobs {
		if job.Status.Terminal() {
			result.FinishedItems++
			completedUnits++
			continue
		}
		if job.Status == StatusQueued {
			continue
		}
		if job.Progress.Percent == nil {
			continue
		}
		pct := *job.Progress.Percent
		if pct < 0 {
			pct = 0
		}
		if pct > 100 {
			pct = 100
		}
		completedUnits += pct / 100
	}
	// This is an item-count estimate, not a byte-weighted measure.
	percent := completedUnits / float64(len(jobs)) * 100
	result.Percent = &percent
	result.Indeterminate = false
	for _, job := range jobs {
		if (job.Status == StatusDownloading || job.Status == StatusConverting) && job.Progress.Percent == nil {
			result.Indeterminate = true
			result.Percent = nil
			break
		}
	}
	return result
}

func newJob(id, rawURL string, req DownloadRequest, outputName string) Job {
	now := timestampNow()
	return Job{
		ID: id, URL: rawURL, OutputDir: req.OutputDir, OutputName: outputName,
		Format: req.Format, AudioQuality: req.AudioQuality, Status: StatusQueued, CreatedAt: now, UpdatedAt: now,
	}
}
