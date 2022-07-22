package cli

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
	"github.com/spf13/cobra"
)

var cmdSync = &cobra.Command{
	Use:   "sync",
	Short: "Sync (binary) test data via a bob-server.",
	Args:  cobra.MinimumNArgs(0),
	Long:  ``,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		// do nothing just show if the server can be contacted and maybe display status information
	},
}

var cmdSyncPush = &cobra.Command{
	Use:   "push",
	Short: "Make server collections exactly like local",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		allowInsecure, err := cmd.Flags().GetBool("insecure")
		errz.Fatal(err)

		runPush(allowInsecure)
	},
}

var cmdSyncPull = &cobra.Command{
	Use:   "pull",
	Short: "Make local collections exactly like server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		allowInsecure, err := cmd.Flags().GetBool("insecure")
		errz.Fatal(err)

		runPull(allowInsecure)
	},
}

var cmdSyncList = &cobra.Command{
	Use:   "ls",
	Short: "List files synced",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runList()
	},
}

var cmdSyncListRemote = &cobra.Command{
	Use:   "ls-remote",
	Short: "List collections on remote",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		allowInsecure, err := cmd.Flags().GetBool("insecure")
		errz.Fatal(err)
		runListRemote(allowInsecure)
	},
}

func runPush(allowInsecure bool) {
	var exitCode int
	defer func() {
		exit(exitCode)
	}()
	defer errz.Recover()

	b, err := bob.Bob(
		bob.WithInsecure(allowInsecure),
	)
	boblog.Log.Error(err, "Unable to initialize bob")

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

	err = b.SyncPush(ctx)
	if err != nil {
		exitCode = 1
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	}
}

func runPull(allowInsecure bool) {
	var exitCode int
	defer func() {
		exit(exitCode)
	}()
	defer errz.Recover()

	b, err := bob.Bob(
		bob.WithInsecure(allowInsecure),
	)
	boblog.Log.Error(err, "Unable to initialize bob")

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

	err = b.SyncPull(ctx)
	if err != nil {
		exitCode = 1
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	}
}

func runList() {
	var exitCode int
	defer func() {
		exit(exitCode)
	}()
	defer errz.Recover()

	b, err := bob.Bob()
	boblog.Log.Error(err, "Unable to initialize bob")

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

	err = b.SyncListLocal(ctx)
	if err != nil {
		exitCode = 1
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
			errz.Log(err)
		}
	}
}

func runListRemote(allowInsecure bool) {
	var exitCode int
	defer func() {
		exit(exitCode)
	}()
	defer errz.Recover()

	b, err := bob.Bob(bob.WithInsecure(allowInsecure))
	boblog.Log.Error(err, "Unable to initialize bob")

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

	err = b.SyncListRemote(ctx)
	if err != nil {
		exitCode = 1
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	}
}
