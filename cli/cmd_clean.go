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
	// Args:  cobra.ExactArgs(1),
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		runClean(cleanGlobal)
	},
}

func runClean(isGlobal bool) {
	b, err := bob.Bob()
	boblog.Log.Error(err, "Unable to initialise bob")

	if isGlobal {
		err = b.Clean()
	} else {
		fmt.Println("todo clean locally")
	}

	boblog.Log.Error(err, "Unable to clean buildinfo")

	fmt.Println("build info cleaned")
	fmt.Println("artifacts cleaned")
}
