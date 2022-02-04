package cli

import (
	"context"

	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/bob/tui"
	"github.com/pkg/errors"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/bob/global"
	"github.com/Benchkram/errz"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run interactive tasks",
	Args:  cobra.MinimumNArgs(0),
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		taskname := global.DefaultBuildTask
		if len(args) > 0 {
			taskname = args[0]
		}

		noCache, err := cmd.Flags().GetBool("no-cache")
		errz.Fatal(err)

		run(taskname, noCache)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		tasks, err := getRuns()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		return tasks, cobra.ShellCompDirectiveDefault
	},
}

func run(taskname string, noCache bool) {
	b, err := bob.Bob(
		bob.WithCachingEnabled(!noCache),
	)
	errz.Fatal(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t, err := tui.New()
	if err != nil {
		errz.Log(err)

		return
	}

	// make sure packages are installed
	err = b.InstallPackages(ctx)
	if err != nil {
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	}

	commander, err := b.Run(ctx, taskname)
	if err != nil {
		switch err {
		case bob.ErrNoRebuildRequired:
		default:
			if errors.As(err, &usererror.Err) {
				boblog.Log.UserError(err)
			} else {
				errz.Fatal(err)
			}
		}
	}

	if commander != nil {
		t.Start(commander)
	}

	cancel()

	if commander != nil {
		<-commander.Done()
	}

	t.Restore()
}

func getRuns() ([]string, error) {
	b, err := bob.Bob()
	if err != nil {
		return nil, err
	}
	return b.GetRunList()
}
