package cli

import (
	"fmt"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/errz"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var CmdClone = &cobra.Command{
	Use:   "clone",
	Short: "Clone a bob workspace and child repositorys",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var url string
		if len(args) > 0 {
			url = args[0]
		}

		runClone(url)
	},
}

func runClone(url string) {

	if len(url) == 0 {
		bob, err := bob.Bob(bob.WithRequireBobConfig())
		errz.Fatal(err)

		// Try to clone dependencies of current repo
		err = bob.Clone()
		errz.Fatal(err)

		fmt.Printf("%s\n", aurora.Green("Cloned"))
	} else {
		bob, err := bob.Bob()
		errz.Fatal(err)

		// Try to clone a new bob repo to the current working directory
		repoName, err := bob.CloneRepo(url)
		errz.Fatal(err)

		fmt.Printf("%s\n", aurora.Green(fmt.Sprintf("Cloned to %s", repoName)))
	}

}
