package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Benchkram/errz"
	"github.com/spf13/cobra"

	"github.com/Benchkram/bob/pkg/composectl"
	"github.com/Benchkram/bob/pkg/composeutil"
)

var dockerCmd = &cobra.Command{
	Use:   "compose",
	Short: "Manage docker-compose environment",
	Args:  cobra.MinimumNArgs(1),
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		defer errz.Recover(&err)

		switch args[0] { // guaranteed by minimumArgs == 1
		case "up":
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

			path := "docker-compose.yml"
			p, err := composeutil.ProjectFromConfig(path)
			if err != nil {
				errz.Fatal(err)
			}

			configs := composeutil.PortConfigs(p)

			hasPortConflict := composeutil.HasPortConflicts(configs)

			mappings := ""
			conflicts := ""
			if hasPortConflict {
				conflicts = composeutil.GetPortConflicts(configs)

				resolved, err := composeutil.ResolvePortConflicts(p, configs)
				if err != nil {
					errz.Fatal(err)
				}

				mappings = composeutil.GetNewPortMappings(resolved)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ctl, err := composectl.New(p, conflicts, mappings)
			if err != nil {
				errz.Fatal(err)
			}

			fmt.Println()
			err = ctl.Up(ctx)
			if err != nil {
				errz.Fatal(err)
			}

			defer func() {
				fmt.Print("\n\n")
				err := ctl.Down(context.Background())
				if err != nil {
					errz.Log(err)
				}
				fmt.Println("\nEnvironment down.")
			}()

			fmt.Print("\nEnvironment up.\n\n")

			<-stop
		}
	},
	ValidArgs: []string{"up"},
}
