package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
	"github.com/spf13/cobra"
)

var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Manage third party packages",
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

		prune, err := cmd.Flags().GetBool("prune")
		if err != nil {
			boblog.Log.Error(err, "parsing Flags")
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle prune call
		if prune {
			err = b.PrunePackages(ctx)
			if err != nil {
				if errors.As(err, &usererror.Err) {
					boblog.Log.UserError(err)
				} else {
					errz.Fatal(err)
				}
			}
			return
		}

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

var packageAddCmd = &cobra.Command{
	Use:   "add",
	Short: "package manager",
	Args:  cobra.MinimumNArgs(1),
	Long:  ``,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	Run: func(cmd *cobra.Command, args []string) {

		// No packages to add
		if len(args) < 1 {
			return
		}

		b, err := bob.Bob()
		if err != nil {
			fmt.Println("can't create bob")
			errz.Log(err)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = b.AddPackages(ctx, args...)
		if err != nil {
			fmt.Println("can't add packages to bobfile")
			errz.Log(err)
			return
		}

	},
	ValidArgsFunction: packageSearch,
}

// Use packageMangagers Search functionality to return available packages
func packageSearch(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	b, err := bob.Bob()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	aggregate, err := b.Aggregate()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res, err := aggregate.Packages.Search(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return res, cobra.ShellCompDirectiveDefault
}
