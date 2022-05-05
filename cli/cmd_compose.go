//go:build dev
// +build dev

package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/benchkram/bob/pkg/composectl"
	"github.com/benchkram/bob/pkg/composeutil"
	"github.com/benchkram/errz"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(composeCmd)
}

var composeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Manage docker-compose environment",
	Args:  cobra.MinimumNArgs(1),
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		defer errz.Recover()

		switch args[0] { // guaranteed by minimumArgs == 1
		case "up":
			path := "docker-compose.yml"
			p, err := composeutil.ProjectFromConfig(path)
			if err != nil {
				errz.Fatal(err)
			}

			cfgs := composeutil.ProjectPortConfigs(p)

			portConflicts := ""
			portMapping := ""
			if composeutil.HasPortConflicts(cfgs) {
				conflicts := composeutil.PortConflicts(cfgs)

				portConflicts = conflicts.String()

				fmt.Println("Conflicting ports detected:")
				fmt.Println(portConflicts)

				// TODO: disable once we also resolve binaries' ports
				errz.Fatal(fmt.Errorf(fmt.Sprint("conflicting ports detected:\n", conflicts)))

				resolved, err := composeutil.ResolvePortConflicts(conflicts)
				errz.Fatal(err)

				portMapping = resolved.String()

				fmt.Println("Resolved port mapping:")
				fmt.Println(portMapping)

				// update project's ports
				composeutil.ApplyPortMapping(p, resolved)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ctl, err := composectl.New()
			if err != nil {
				errz.Fatal(err)
			}

			fmt.Println()
			err = ctl.Up(ctx, p)
			if err != nil {
				errz.Fatal(err)
			}

			defer func() {
				fmt.Print("\n\n")
				err := ctl.Down(context.Background(), p)
				if err != nil {
					errz.Log(err)
				}
				fmt.Println("\nEnvironment down.")
			}()

			fmt.Print("\nEnvironment up.")

			stop := make(chan os.Signal, 1)
			signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

			<-stop
		}
	},
	ValidArgs: []string{"up"},
}
