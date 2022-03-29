package cli

import "github.com/benchkram/bob/pkg/boblog"

func logInit(level int) {
	// Log levels
	// 0 - no logs, only errors
	// 1 - info logs..
	// 2 - debug logs..
	// 3 - debug logs with hints why a task is beeing rebuild
	// 5 - debug logs TBD

	boblog.SetLogLevel(level)
}
