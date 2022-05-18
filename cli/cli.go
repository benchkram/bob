package cli

import (
	"os"
	"sync"
)

var _stopProfiling func()

var once sync.Once

// stopProfiling triggeres _stopProfiling.
// It's save to be called multiple times.
func stopProfiling() {
	if _stopProfiling != nil {
		once.Do(_stopProfiling)
	}
}

func exit(code int) {
	stopProfiling()
	os.Exit(code)
}

func Execute() error {
	defer stopProfiling()
	return rootCmd.Execute()
}
