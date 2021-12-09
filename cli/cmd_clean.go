package cli

import (
	"fmt"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/pkg/boblog"
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
