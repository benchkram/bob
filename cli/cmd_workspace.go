package cli

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/Benchkram/errz"
)

var cmdWorkspace = &cobra.Command{
	Use:   "workspace",
	Short: "Manage a bob workspace",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runInit()
	},
}

func runInit() {
	b, err := bob.Bob()
	errz.Fatal(err)

	err = b.Init()
	if err != nil {
		if errors.Is(err, bob.ErrWorkspaceAlreadyInitialised) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	}
}
