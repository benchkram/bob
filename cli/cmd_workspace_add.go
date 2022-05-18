package cli

import (
	"errors"
	"fmt"

	"github.com/benchkram/bob/pkg/add"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var cmdAdd = &cobra.Command{
	Use:   "add",
	Short: "Add a git repository to a workspace",
	Args:  cobra.ExactArgs(1),
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		plain, err := cmd.Flags().GetBool("plain")
		errz.Fatal(err)

		repoURL := args[0]
		runAdd(repoURL, plain)
	},
}

func runAdd(repoURL string, plain bool) {
	err := add.Add(
		repoURL,
		add.WithPlainProtocol(plain),
	)

	if errors.As(err, &usererror.Err) {
		boblog.Log.UserError(err)
		exit(1)
	} else {
		errz.Fatal(err)
	}

	fmt.Printf("%s\n", aurora.Green("Repo added"))
}
