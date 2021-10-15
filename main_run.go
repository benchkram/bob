package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/bob/global"
	"github.com/Benchkram/errz"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a development environment",
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

	ctl, err := b.Run(ctx, taskname)
	if err != nil {
		switch err {
		case bob.ErrNoRebuildRequired:
		default:
			errz.Log(err)
			os.Exit(1)
		}
	}

	// go func() {
	// 	time.Sleep(5 * time.Second)
	// 	for {
	// 		// println("Sending compose restart signal")
	// 		ctl.Restart()
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }()

	// Wait for Ctrl-C or the cmd's stop channel to close.
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	select {
	case <-signalChannel:
	case <-ctl.Done():
	}

	cancel()
	<-ctl.Done()
	fmt.Printf("%s stopped\n", ctl.Name())
}

func getRuns() ([]string, error) {
	b, err := bob.Bob()
	if err != nil {
		return nil, err
	}
	return b.GetRunList()
}
