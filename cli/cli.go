package cli

import "os"

var _stopProfiling func()

func stopProfiling() {
	if _stopProfiling != nil {
		_stopProfiling()
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
