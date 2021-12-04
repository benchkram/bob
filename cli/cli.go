package cli

var _stopProfiling func()

func stopProfiling() {
	if _stopProfiling != nil {
		_stopProfiling()
	}
}

func Execute() error {
	defer stopProfiling()
	return rootCmd.Execute()
}
