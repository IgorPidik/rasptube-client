package ui

import "fmt"

func FormatMilliseconds(ms uint32) string {
	minutes := ms / (1000 * 60)
	seconds := ms / 1000
	if minutes > 0 {
		seconds = (ms % (minutes * 1000 * 60)) / 1000
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}
