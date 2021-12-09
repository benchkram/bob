package main

import (
	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/cli"
	"github.com/Benchkram/bob/pkg/boblog"
)

var Version = "0.0.0"

func main() {
	bob.Version = Version

	if err := cli.Execute(); err != nil {
		boblog.Log.Error(err, "Error on execution of bob command")
	}
}
