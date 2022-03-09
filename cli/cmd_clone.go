package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var CmdClone = &cobra.Command{
	Use:   "clone",
	Short: "Clone a bob workspace and child repositorys",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		failFast, err := cmd.Flags().GetBool("fail-fast")
		errz.Fatal(err)

		var url string
		if len(args) > 0 {
			url = args[0]
		}
		runClone(url, failFast)
	},
}

func runClone(url string, failFast bool) {

	if len(url) == 0 {
		bob, err := bob.Bob(bob.WithRequireBobConfig())
		errz.Fatal(err)

		// Try to clone dependencies of current repo
		err = bob.Clone(failFast)
		errz.Fatal(err)

		fmt.Printf("%s\n", aurora.Green("Cloned"))
	} else {
		bob, err := bob.Bob()
		errz.Fatal(err)

		// Try to clone a new bob repo to the current working directory
		repoName, err := bob.CloneRepo(url, failFast)
		if errors.As(err, &usererror.Err) {
			fmt.Printf("%s\n", aurora.Red(errors.Unwrap(err).Error()))
			os.Exit(1)
		} else {
			errz.Fatal(err)
		}

		fmt.Printf("%s\n", aurora.Green(fmt.Sprintf("Cloned to %s", repoName)))
	}

}
