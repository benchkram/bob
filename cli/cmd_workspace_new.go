package cli

import (
	"errors"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/spf13/cobra"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/errz"
)

var cmdWorkspaceNew = &cobra.Command{
	Use:   "new",
	Short: "Create a new bob workspace",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runWorkspaceNew()
	},
}

func runWorkspaceNew() {
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
