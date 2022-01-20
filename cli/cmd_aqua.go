package cli

import (
	"context"
	"fmt"

	"github.com/Benchkram/errz"
	"github.com/aquaproj/aqua/pkg/controller"
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
		fmt.Println("I'm aqua")

		ctx := context.Background()
		param := &controller.Param{
			// TODO: checkout contents of config file
			// ConfigFilePath: "config",<
			IsTest:      false,
			All:         true,
			AQUAVersion: "v0.13.0",
		}

		// TODO: find allocation of .aqua path inside aqua

		// TODO: remove/reinstall packages

		ctrl, err := controller.New(ctx, param)
		errz.Fatal(err)

		err = ctrl.Install(context.Background(), param)
		errz.Fatal(err)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveDefault
	},
}
