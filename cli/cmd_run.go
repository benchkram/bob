package cli

import (
	"context"
	"os"

	"github.com/Benchkram/bob/tui"

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
		run(taskname)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		tasks, err := getRuns()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}
		return tasks, cobra.ShellCompDirectiveDefault
	},
}

func run(taskname string) {

	b, err := bob.Bob()
	errz.Fatal(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t, err := tui.New()
	if err != nil {
		errz.Log(err)

		return
	}

	commander, err := b.Run(ctx, taskname)
	if err != nil {
		switch err {
		case bob.ErrNoRebuildRequired:
		default:
			errz.Log(err)
			os.Exit(1)
		}
	}

	t.Start(commander)

	cancel()
	<-commander.Done()
}

func getRuns() ([]string, error) {
	b, err := bob.Bob()
	if err != nil {
		return nil, err
	}
	return b.GetRunList()
}
