package downloader

import "time"

const TimestampLayout = "2006-01-02T15:04:05.000000000Z07:00"

func timestampNow() string {
	return time.Now().UTC().Format(TimestampLayout)
}

type AudioFormat string

const (
	FormatMP3  AudioFormat = "mp3"
	FormatM4A  AudioFormat = "m4a"
	FormatOpus AudioFormat = "opus"
	FormatWAV  AudioFormat = "wav"
	FormatFLAC AudioFormat = "flac"
)

var SupportedFormats = []AudioFormat{FormatMP3, FormatM4A, FormatOpus, FormatWAV, FormatFLAC}

type JobStatus string

const (
	StatusQueued      JobStatus = "queued"
	StatusDownloading JobStatus = "downloading"
	StatusConverting  JobStatus = "converting"
	StatusCompleted   JobStatus = "completed"
	StatusFailed      JobStatus = "failed"
	StatusCancelled   JobStatus = "cancelled"
)

func (s JobStatus) Terminal() bool {
	return s == StatusCompleted || s == StatusFailed || s == StatusCancelled
}

type VideoRequest struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

type DownloadRequest struct {
	Videos       []VideoRequest `json:"videos"`
	OutputDir    string         `json:"outputDir"`
	Format       AudioFormat    `json:"format"`
	AudioQuality string         `json:"audioQuality"`
}

type Progress struct {
	Percent             *float64 `json:"percent,omitempty"`
	DownloadedBytes     int64    `json:"downloadedBytes,omitempty"`
	TotalBytes          int64    `json:"totalBytes,omitempty"`
	SpeedBytesPerSecond *float64 `json:"speedBytesPerSecond,omitempty"`
	ETASeconds          *int64   `json:"etaSeconds,omitempty"`
}

type Job struct {
	ID           string      `json:"id"`
	URL          string      `json:"url"`
	OutputDir    string      `json:"outputDir"`
	OutputName   string      `json:"outputName"`
	Format       AudioFormat `json:"format"`
	AudioQuality string      `json:"audioQuality"`
	Status       JobStatus   `json:"status"`
	Progress     Progress    `json:"progress"`
	OutputPath   string      `json:"outputPath,omitempty"`
	OutputSize   int64       `json:"outputSize,omitempty"`
	Error        string      `json:"error,omitempty"`
	CreatedAt    string      `json:"createdAt"`
	UpdatedAt    string      `json:"updatedAt"`
}

type HistoryEntry struct {
	ID         string      `json:"id"`
	URL        string      `json:"url"`
	OutputDir  string      `json:"outputDir"`
	OutputPath string      `json:"outputPath,omitempty"`
	Format     AudioFormat `json:"format"`
	Status     JobStatus   `json:"status"`
	OutputSize int64       `json:"outputSize,omitempty"`
	Error      string      `json:"error,omitempty"`
	FinishedAt string      `json:"finishedAt"`
}

type Settings struct {
	OutputDir    string      `json:"outputDir"`
	Format       AudioFormat `json:"format"`
	Concurrency  int         `json:"concurrency"`
	AudioQuality string      `json:"audioQuality"`
}

type AggregateProgress struct {
	TotalItems    int      `json:"totalItems"`
	FinishedItems int      `json:"finishedItems"`
	Percent       *float64 `json:"percent,omitempty"`
	Indeterminate bool     `json:"indeterminate"`
}

type Capabilities struct {
	ToolsAvailable bool          `json:"toolsAvailable"`
	YTDLPVersion   string        `json:"ytDlpVersion,omitempty"`
	FFmpegVersion  string        `json:"ffmpegVersion,omitempty"`
	FFprobeFound   bool          `json:"ffprobeFound"`
	ToolsDirectory string        `json:"toolsDirectory,omitempty"`
	Formats        []AudioFormat `json:"formats"`
	Message        string        `json:"message,omitempty"`
}

type JobEvent struct {
	Job       Job               `json:"job"`
	Aggregate AggregateProgress `json:"aggregate"`
}
