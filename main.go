package main

import (
	"os"

	"github.com/Benchkram/bob/cli"
	"github.com/Benchkram/errz"
)

func main() {
	var exitCode int
	defer func() { os.Exit(exitCode) }()

	if err := cli.Execute(); err != nil {
		errz.Log(err)
		exitCode = 1
	}
}
