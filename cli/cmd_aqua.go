package cli

import (
	"context"
	"fmt"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/Benchkram/errz"
	"github.com/logrusorgru/aurora"
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

		boblog.Log.Info(aurora.Green("Installing packages...").String())
		fmt.Println()

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
		boblog.Log.Info(aurora.Green("All packages successfully installed").String())

	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveDefault
	},
}
