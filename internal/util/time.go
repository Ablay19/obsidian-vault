package util

import (
	"fmt"
	"time"
)

// formatTime formats a time.Time object into a human-readable string.
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return "--"
	}
	diff := time.Since(t)

	if diff < time.Minute {
		return fmt.Sprintf("%ds ago", int(diff.Seconds()))
	}
	if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	}
	return t.Format("Jan 02, 2006 15:04 MST")
}
