package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Benchkram/errz"
	"github.com/aquaproj/aqua/pkg/controller"
	"github.com/spf13/cobra"
)

const AQUA_ROOT = ".bob/.aqua"

var aquaCmd = &cobra.Command{
	Use:   "aqua",
	Short: "aqua test",
	Args:  cobra.MinimumNArgs(0),
	Long:  ``,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("I'm aqua")

		// Create local .aqua dir and reference to that
		os.MkdirAll(AQUA_ROOT, os.ModePerm)
		os.Setenv("AQUA_ROOT_DIR", AQUA_ROOT)

		aquabin, err := filepath.Abs(fmt.Sprintf("%s/bin", AQUA_ROOT))
		errz.Fatal(err)

		fmt.Println(os.Getenv("PATH"))

		// Add aqua bin to PATH
		os.Setenv("PATH", fmt.Sprintf("%s:%s", aquabin, os.Getenv("PATH")))

		fmt.Println(os.Getenv("PATH"))

		ctx := context.Background()
		param := &controller.Param{
			// TODO: checkout contents of config file

			ConfigFilePath: "aqua.yaml", // This could be nested somewhere inside .bob dir
			IsTest:         false,
			All:            true,
			AQUAVersion:    "v0.13.0",
		}

		// TODO: remove/reinstall packages

		ctrl, err := controller.New(ctx, param)
		errz.Fatal(err)

		err = ctrl.Install(context.Background(), param)
		errz.Fatal(err)

		command := exec.Command("whereis", "fzf")
		command.Stdout = os.Stdout
		err = command.Start()
		errz.Fatal(err)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveDefault
	},
}
