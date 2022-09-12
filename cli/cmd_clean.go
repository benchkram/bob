package cli

import (
	"fmt"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean buildinfo and artifacts",
	//Args:  cobra.ExactArgs(1),
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		runClean()
	},
}

func runClean() {
	b, err := bob.Bob()
	boblog.Log.Error(err, "Unable to initialise bob")

	err = b.Clean()
	boblog.Log.Error(err, "Unable to clean buildinfo")

	fmt.Println("build info cleaned")
	fmt.Println("artifacts cleaned")
}

var cleanTargetsCmd = &cobra.Command{
	Use:   "targets",
	Short: "Clean targets",
	//Args:  cobra.ExactArgs(1),
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		runCleanTargets()
	},
}

func runCleanTargets() {
	b, err := bob.Bob()
	boblog.Log.Error(err, "Unable to initialise bob")

	ag, err := b.Aggregate()
	boblog.Log.Error(err, "Unable to aggregate bob file")

	for _, t := range ag.BTasks {
		err := t.Clean()
		if err != nil {
			boblog.Log.Error(err, "Unable to clean target")
		}
	}
}
