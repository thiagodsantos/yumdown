package downloader

import (
	"fmt"
	"strconv"
	"strings"
)

const progressPrefix = "__YUMDOWN_PROGRESS__"
const outputPrefix = "__YUMDOWN_OUTPUT__"

// ParseProgress parses the stable, app-specific record emitted by yt-dlp's
// --progress-template. Unknown fields stay unknown rather than being guessed.
func ParseProgress(line string) (Progress, bool) {
	if !strings.HasPrefix(line, progressPrefix) {
		return Progress{}, false
	}
	fields := strings.Split(strings.TrimPrefix(line, progressPrefix), "|")
	if len(fields) != 6 {
		return Progress{}, false
	}

	progress := Progress{}
	if value, ok := parseInt(fields[0]); ok {
		progress.DownloadedBytes = value
	}
	total, totalOK := parseInt(fields[1])
	if !totalOK || total <= 0 {
		total, totalOK = parseInt(fields[2])
	}
	if totalOK && total > 0 {
		progress.TotalBytes = total
	}
	if value, ok := parseFloat(fields[3]); ok && value >= 0 && value <= 100 {
		progress.Percent = &value
	} else if progress.TotalBytes > 0 {
		value := float64(progress.DownloadedBytes) / float64(progress.TotalBytes) * 100
		if value > 100 {
			value = 100
		}
		progress.Percent = &value
	}
	if value, ok := parseFloat(fields[4]); ok && value >= 0 {
		progress.SpeedBytesPerSecond = &value
	}
	if value, ok := parseInt(fields[5]); ok && value >= 0 {
		progress.ETASeconds = &value
	}
	return progress, true
}

func parseInt(value string) (int64, bool) {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "NA") || strings.EqualFold(value, "none") {
		return 0, false
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	return parsed, err == nil
}

func parseFloat(value string) (float64, bool) {
	value = strings.TrimSpace(strings.TrimSuffix(value, "%"))
	if value == "" || strings.EqualFold(value, "NA") || strings.EqualFold(value, "none") {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	return parsed, err == nil
}

func progressTemplate() string {
	return fmt.Sprintf("download:%s%%(progress.downloaded_bytes)s|%%(progress.total_bytes)s|%%(progress.total_bytes_estimate)s|%%(progress._percent_str)s|%%(progress.speed)s|%%(progress.eta)s", progressPrefix)
}
