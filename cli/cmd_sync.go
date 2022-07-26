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
	Use:   "sync [--force] [--insecure]",
	Short: "Pull the sync collections defined in bobfile",
	Long:  ``,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		allowInsecure, err := cmd.PersistentFlags().GetBool("insecure")
		errz.Fatal(err)
		force, err := cmd.Flags().GetBool("force")
		errz.Fatal(err)

		runPull(allowInsecure, force)
	},
}

var cmdSyncCreate = &cobra.Command{
	Use:        "create collection_name path/to/dir",
	Short:      "Create a new collection or collection version",
	Long:       ``,
	Args:       cobra.ExactArgs(2),
	ArgAliases: []string{"collectionName", "path"},
	Run: func(cmd *cobra.Command, args []string) {
		allowInsecure, err := cmd.Flags().GetBool("insecure")
		errz.Fatal(err)
		dry, err := cmd.Flags().GetBool("dry")
		errz.Fatal(err)
		version, err := cmd.Flags().GetString("set-version")
		errz.Fatal(err)

		// collection_name can be anything but not empty
		collectionName := args[0]
		// path is validated in createPush
		path := args[1]

		runCreatePush(collectionName, version, path, dry, allowInsecure)
	},
}

var cmdSyncList = &cobra.Command{
	Use:   "ls-local",
	Short: "List files tracked by sync",
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

func runCreatePush(collectionName, version, path string, dry, allowInsecure bool) {
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

	err = b.SyncCreatePush(ctx, collectionName, version, path, dry)
	if err != nil {
		exitCode = 1
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	}
}

func runPull(allowInsecure bool, force bool) {
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

	err = b.SyncPull(ctx, force)
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
			errz.Log(err)
			errz.Fatal(err)
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
