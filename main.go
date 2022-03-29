package main

import (
	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/cli"
	"github.com/benchkram/bob/pkg/boblog"
)

var Version = "0.0.0"

func main() {
	bob.Version = Version

	if err := cli.Execute(); err != nil {
		boblog.Log.Error(err, "Error on execution of bob command")
	}
}
