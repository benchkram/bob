package format

import (
	"fmt"
	"time"
)

// DisplayDuration used to display durations output to users
func DisplayDuration(d time.Duration) string {
	if d.Minutes() > 1 {
		return fmt.Sprintf("%.1fm", float64(d)/float64(time.Minute))
	}
	if d.Seconds() > 1 {
		return fmt.Sprintf("%.1fs", float64(d)/float64(time.Second))
	}
	return fmt.Sprintf("%.1fms", float64(d)/float64(time.Millisecond)+0.1) // add .1ms so that it never returns 0.0ms
}
