package cli

import (
	"fmt"
	"os"

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
	var exitCode int
	defer func() { os.Exit(exitCode) }()

	b, err := bob.Bob()
	if err != nil {
		exitCode = 1
		boblog.Log.UserError(err)
		return
	}

	if err = b.Install(); err != nil {
		exitCode = 1
		boblog.Log.UserError(err)
		return
	}

	fmt.Printf("\nAll dependencies installed!\n")
}
