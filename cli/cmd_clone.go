package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"

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
		https, err := cmd.Flags().GetBool("https")
		errz.Fatal(err)
		ssh, err := cmd.Flags().GetBool("ssh")
		errz.Fatal(err)

		if ssh && https {
			fmt.Printf("%s\n", aurora.Red("You can only use one of --ssh or --https"))
			os.Exit(0)
		}

		var protocol string
		if ssh {
			protocol = "ssh"
		} else if https {
			protocol = "https"
		}

		var url string
		if len(args) > 0 {
			url = args[0]
		}
		runClone(url, failFast, protocol)
	},
}

func runClone(url string, failFast bool, protocol string) {

	if len(url) == 0 {
		bob, err := bob.Bob(bob.WithRequireBobConfig())
		errz.Fatal(err)

		// Try to clone dependencies of current repo
		err = bob.Clone(failFast, protocol)
		errz.Fatal(err)

		fmt.Printf("%s\n", aurora.Green("Cloned"))
	} else {
		bob, err := bob.Bob()
		errz.Fatal(err)

		// Try to clone a new bob repo to the current working directory
		repoName, err := bob.CloneRepo(url, failFast)
		if errors.As(err, &usererror.Err) {
			fmt.Printf("%s\n", aurora.Red(errors.Unwrap(err).Error()))
			exit(1)
		} else {
			errz.Fatal(err)
		}

		fmt.Printf("%s\n", aurora.Green(fmt.Sprintf("Cloned to %s", repoName)))
	}

}
