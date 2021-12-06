package main

import (
	"github.com/Benchkram/bob/bob"
	"os"

	"github.com/Benchkram/bob/cli"
	"github.com/Benchkram/errz"
)

var Version = "0.0.0"

func main() {
	var exitCode int
	defer func() { os.Exit(exitCode) }()

	bob.Version = Version

	if err := cli.Execute(); err != nil {
		errz.Log(err)
		exitCode = 1
	}
}
