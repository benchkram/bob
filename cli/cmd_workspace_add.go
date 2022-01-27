package cli

import (
	"fmt"

	"github.com/Benchkram/bob/pkg/add"
	"github.com/Benchkram/errz"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var cmdAdd = &cobra.Command{
	Use:   "add",
	Short: "Add a git repository to a workspace",
	Args:  cobra.ExactArgs(1),
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		explicit, err := cmd.Flags().GetBool("explicit-protocol")
		errz.Fatal(err)

		repoURL := args[0]
		runAdd(repoURL, explicit)
	},
}

func runAdd(repoURL string, explcitprotcl bool) {
	err := add.Add(
		repoURL,
		add.WithExplicitProtocol(explcitprotcl),
	)
	errz.Fatal(err)

	fmt.Printf("%s\n", aurora.Green("Repo added"))
}
