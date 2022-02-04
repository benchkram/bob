package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/errz"
	"github.com/spf13/cobra"
)

var aquaCmd = &cobra.Command{
	Use:   "aqua",
	Short: "aqua test",
	Args:  cobra.MinimumNArgs(0),
	Long:  ``,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		defer errz.Recover(&err)

		b, err := bob.Bob()
		if err != nil {
			// TODO: usererror
			fmt.Println("can't create bob")
			errz.Log(err)
			return
		}

		ctx := context.Background()
		err = b.InstallPackages(ctx)
		if err != nil {
			// TODO: usererror
			fmt.Println("can't install packages")
			errz.Log(err)
			return
		}

		fmt.Println()

		for _, cmd := range args {
			fmt.Printf("Test if comamnd \"%s\" can be executed\n", cmd)
			ex := exec.Command(cmd, "--version")
			ex.Stdout = os.Stdout
			err = ex.Start()
			if err != nil {
				// TODO: usererror
				fmt.Println("can't run command")
				errz.Log(err)
				return
			}
			err = ex.Wait()
			if err != nil {
				// TODO: usererror
				fmt.Println("command did not exit correctly")
				errz.Log(err)
				return
			}
			fmt.Println()
		}

	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveDefault
	},
}
