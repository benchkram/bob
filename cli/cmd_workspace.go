package cli

import (
	"errors"

	"github.com/benchkram/bob/pkg/usererror"
	"github.com/spf13/cobra"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/errz"
)

var cmdWorkspace = &cobra.Command{
	Use:   "workspace",
	Short: "Manage a bob workspace",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runInitWorkspace()
	},
}

func runInitWorkspace() {
	b, err := bob.Bob()
	errz.Fatal(err)

	err = b.Init()
	if err != nil {
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
	}
}
