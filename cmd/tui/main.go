package main

import (
	"github.com/Benchkram/bob/cmd/tui/tui"
	"github.com/Benchkram/bob/pkg/execctl"
)

func main() {
	cmd1, err := execctl.NewCmd("app", "/bin/bash", "-c", "./script1.sh")
	if err != nil {
		panic(err)
	}

	cmd2, err := execctl.NewCmd("mongo", "/bin/bash", "-c", "./script2.sh")
	if err != nil {
		panic(err)
	}

	cmd3, err := execctl.NewCmd("redis", "/bin/bash", "-c", "./script3.sh")
	if err != nil {
		panic(err)
	}

	root := execctl.NewCmdTree(cmd1, cmd2, cmd3)

	t, err := tui.New(root)
	if err != nil {
		panic(err)
	}

	t.Start()
}
