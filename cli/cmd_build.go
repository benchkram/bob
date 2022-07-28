package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/benchkram/errz"
	"github.com/spf13/cobra"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
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

		noCache, err := cmd.Flags().GetBool("no-cache")
		errz.Fatal(err)

		allowInsecure, err := cmd.Flags().GetBool("insecure")
		errz.Fatal(err)

		taskname := global.DefaultBuildTask
		if len(args) > 0 {
			taskname = args[0]
		}

		runBuild(dummy, taskname, noCache, allowInsecure, flagEnvVars)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		tasks, err := getBuildTasks()
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

func runBuild(dummy bool, taskname string, noCache, allowInsecure bool, flagEnvVars []string) {
	var exitCode int
	defer func() {
		exit(exitCode)
	}()
	defer errz.Recover()

	if dummy {
		wd, err := os.Getwd()
		errz.Fatal(err)
		err = bobfile.CreateDummyBobfile(wd, false)
		errz.Fatal(err)
		return
	}

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

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

		<-stop
		cancel()
	}()

	err = b.Build(ctx, taskname)
	if err != nil {
		exitCode = 1
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	}
}

func runBuildList() {
	b, err := bob.Bob()
	boblog.Log.Error(err, "Unable to initialize bob")

	tasks, err := b.GetBuildTasks()
	boblog.Log.Error(err, "Unable to aggregate bob file")

	for _, t := range tasks {
		fmt.Println(t)
	}
}

func getBuildTasks() ([]string, error) {
	b, err := bob.Bob()
	if err != nil {
		return nil, err
	}
	return b.GetBuildTasks()
}

// parseEnvVarsFlag will parse flagEnvVars and return the environment variables
// based on:
//   --env VAR_ONE         uses VAR_ONE from host environment variable
//   --env VAR_ONE=value   overwrites value from host with given `value`
func parseEnvVarsFlag(flagEnvVars []string) []string {
	var result []string
	for _, v := range flagEnvVars {
		if strings.Contains(v, "=") {
			result = append(result, v)
		} else {
			// get it from host
			value := os.Getenv(v)
			result = append(result, v+"="+value)
		}
	}

	return result
}
