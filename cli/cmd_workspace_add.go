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
		httpsOnly, err := cmd.Flags().GetBool("https-only")
		errz.Fatal(err)

		sshOnly, err := cmd.Flags().GetBool("ssh-only")
		errz.Fatal(err)

		repoURL := args[0]
		runAdd(repoURL, httpsOnly, sshOnly)
	},
}

func runAdd(repoURL string, https bool, ssh bool) {
	err := add.Add(
		repoURL,
		add.WithHttpsOnly(https),
		add.WithSSHOnly(ssh),
	)
	errz.Fatal(err)

	fmt.Printf("%s\n", aurora.Green("Repo added"))
}
