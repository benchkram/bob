package cli

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/benchkram/bob/pkg/auth"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
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
    cmd: echo "Hello bob!"
`

var withProject = `project: %s
nixpkgs: https://github.com/NixOS/nixpkgs/archive/refs/tags/22.05.tar.gz
build:
  build:
    input: .
    cmd: touch hello-world
    target: hello-world
`

func runInit(project string) {
	if _, err := os.Stat("bob.yaml"); err == nil {
		boblog.Log.UserError(errors.New("there is already a bob.yaml in your project"))
		os.Exit(1)
	}

	wd, _ := os.Getwd()

	var err error
	if project != "" {
		var token string
		fmt.Print("Enter your access token:")
		fmt.Scanf("%s", &token)

		if token == "" {
			boblog.Log.UserError(errors.New("missing token"))
			os.Exit(1)
		}

		err = runAuthContextCreate(defaultContext, token)

		var name string
		if errors.As(err, &usererror.Err) {
			if errors.Is(err, auth.ErrAlreadyExists) {
				fmt.Printf("Enter your name for auth context[%s]:", path.Base(project))
				fmt.Scanf("%s", &name)

				if name == "" {
					name = path.Base(project)
				}
				_ = runAuthContextCreate(name, token)
			} else {
				boblog.Log.UserError(err)
				os.Exit(1)
			}
		} else {
			errz.Fatal(err)
		}

		err = runAuthContextSwitch(name)
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
			os.Exit(1)
		} else {
			errz.Fatal(err)
		}
		err = createBobfile(fmt.Sprintf(withProject, project))
		fmt.Printf("Initialized bob project in %s\n", wd)
		fmt.Println("Run your first build: bob build --push")
	} else {
		err = createBobfile(withoutProject)
		fmt.Printf("Initialized bob project in %s\n", wd)
		fmt.Println("Run your first build: bob build")
	}

	if err != nil {
		log.Fatal(err)
	}
}

func createBobfile(content string) error {
	return os.WriteFile("bob.yaml", []byte(content), 0664)
}
