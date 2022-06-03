package cli

import (
	"fmt"
	"os"

	"github.com/benchkram/errz"
	"github.com/spf13/cobra"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/boblog"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install all dependencies",
	// Args:  cobra.ExactArgs(1),
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		runInstall()
	},
}

func runInstall() {
	var exitCode int
	defer func() { os.Exit(exitCode) }()

	nix, err := bob.NewNixWithCache()
	errz.Fatal(err)

	b, err := bob.Bob(bob.WithNix(nix))
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
