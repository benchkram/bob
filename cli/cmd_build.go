package cli

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Benchkram/bob/pkg/usererror"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/bob/bobfile"
	"github.com/Benchkram/bob/bob/global"
	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/Benchkram/errz"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Run tasks",
	Args:  cobra.MinimumNArgs(0),
	Long:  ``,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		dummy, err := strconv.ParseBool(cmd.Flag("dummy").Value.String())
		errz.Fatal(err)

		taskname := global.DefaultBuildTask
		if len(args) > 0 {
			taskname = args[0]
		}

		runBuild(dummy, taskname)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		tasks, err := getTasks()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		return tasks, cobra.ShellCompDirectiveDefault
	},
}

var buildListCmd = &cobra.Command{
	Use:   "ls",
	Short: "ls",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runBuildList()
	},
}

func runBuild(dummy bool, taskname string) {
	if dummy {
		wd, err := os.Getwd()
		errz.Fatal(err)
		err = bobfile.CreateDummyBobfile(wd, false)
		errz.Fatal(err)
		return
	}

	b, err := bob.Bob(
		bob.WithDisableCache(false),
	)
	errz.Fatal(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

		<-stop
		cancel()
	}()

	err = b.Build(ctx, taskname)
	if errors.As(err, &usererror.Err) {
		boblog.Log.UserError(err)
	} else {
		errz.Fatal(err)
	}
}

func runBuildList() {
	b, err := bob.Bob()
	boblog.Log.Error(err, "Unable to initialize bob")

	err = b.List()
	boblog.Log.Error(err, "Unable to aggregate bob file")
}

func getTasks() ([]string, error) {
	b, err := bob.Bob()
	if err != nil {
		return nil, err
	}
	return b.GetList()
}
