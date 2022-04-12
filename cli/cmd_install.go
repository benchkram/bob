package cli

import (
	"fmt"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install all dependencies",
	//Args:  cobra.ExactArgs(1),
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		runInstall()
	},
}

func runInstall() {
	b, err := bob.Bob()
	if err != nil {
		boblog.Log.Error(err, "Unable to initialise bob")
		return
	}

	if err = b.Install(); err != nil {
		boblog.Log.Error(err, "Unable to install dependencies")
		return
	}

	fmt.Println("All dependencies installed!")
}
