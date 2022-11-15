package cli

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/benchkram/bob/bob/bobfile/project"
	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Init a bob project",
	Args:  cobra.MaximumNArgs(1),
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var project string
		if len(args) > 0 {
			project = strings.TrimSpace(args[0])
		}
		runInit(project)
	},
}

var withoutProject = `nixpkgs: https://github.com/NixOS/nixpkgs/archive/refs/tags/22.05.tar.gz
build:
  build:
    input: .
    cmd: touch hello-world
    target: hello-world
`

var withProject = `project: %s
nixpkgs: https://github.com/NixOS/nixpkgs/archive/refs/tags/22.05.tar.gz
build:
  build:
    input: .
    cmd: touch hello-world
    target: hello-world
`

func runInit(projectName string) {
	if _, err := os.Stat(global.BobFileName); err == nil {
		boblog.Log.UserError(fmt.Errorf("there is already a %s in your project", global.BobFileName))
		os.Exit(1)
	}

	wd, _ := os.Getwd()

	var err error
	if projectName != "" {
		_, err = project.Parse(projectName)
		if err != nil {
			boblog.Log.UserError(err)
			os.Exit(1)
		}
		err = createBobfile(fmt.Sprintf(withProject, projectName))
		fmt.Printf("Initialized basic %s in %s\n", global.BobFileName, wd)
	} else {
		err = createBobfile(withoutProject)
		fmt.Printf("Initialized basic %s in %s\n", global.BobFileName, wd)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func createBobfile(content string) error {
	return os.WriteFile(global.BobFileName, []byte(content), 0664)
}
