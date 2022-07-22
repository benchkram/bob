package cli

import (
	"context"

	"github.com/benchkram/errz"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/bob/tui"
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

		allowInsecure, err := cmd.Flags().GetBool("insecure")
		errz.Fatal(err)

		run(taskname, noCache, allowInsecure)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		tasks, err := getRunTasks()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		return tasks, cobra.ShellCompDirectiveDefault
	},
}

func run(taskname string, noCache bool, allowInsecure bool) {
	var exitCode int
	defer func() {
		exit(exitCode)
	}()
	defer errz.Recover()

	b, err := bob.Bob(
		bob.WithCachingEnabled(!noCache),
		bob.WithInsecure(allowInsecure),
		bob.WithEnvVariables(parseEnvVarsFlag(flagEnvVars)),
	)
	if err != nil {
		exitCode = 1
		errz.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t, err := tui.New()
	defer t.Restore()
	if err != nil {
		exitCode = 1
		errz.Fatal(err)
	}

	commander, err := b.Run(ctx, taskname)
	if err != nil {
		exitCode = 1
		switch err {
		case bob.ErrNoRebuildRequired:
		default:
			if errors.As(err, &usererror.Err) {
				boblog.Log.UserError(err)
				return
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
}

func getRunTasks() ([]string, error) {
	b, err := bob.Bob()
	if err != nil {
		return nil, err
	}
	return b.GetRunTasks()
}
