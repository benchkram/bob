package main

import (
	"github.com/benchkram/bob/pkg/execctl"
	"github.com/benchkram/bob/tui"
)

func main() {
	cmd1, err := execctl.NewCmd("app", "/bin/bash", execctl.WithArgs("-c", "./script1.sh"))
	if err != nil {
		panic(err)
	}

	cmd2, err := execctl.NewCmd("mongo", "/bin/bash", execctl.WithArgs("-c", "./script2.sh"))
	if err != nil {
		panic(err)
	}

	cmd3, err := execctl.NewCmd("redis", "/bin/bash", execctl.WithArgs("-c", "./script3.sh"))
	if err != nil {
		panic(err)
	}

	root := execctl.NewCmdTree(cmd1, cmd2, cmd3)

	t, err := tui.New()
	if err != nil {
		panic(err)
	}

	t.Start(root)
}
