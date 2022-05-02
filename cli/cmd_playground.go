//go:build dev
// +build dev

package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/errz"

	"github.com/spf13/cobra"
)

func init() {
	playgroundCmd.Flags().Bool("clean", false, "Delete directory content before creating the playground")
	rootCmd.AddCommand(playgroundCmd)
}

var playgroundCmd = &cobra.Command{
	Use:   "playground",
	Short: "Create a temporary playground in the current dir",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		clean, err := strconv.ParseBool(cmd.Flag("clean").Value.String())
		errz.Fatal(err)

		if clean {

			wd, err := os.Getwd()
			errz.Fatal(err)
			if wd == "/" {
				fmt.Println("Can't delete root '/'")
				exit(1)
			}

			homedir := os.Getenv("HOME")
			if homedir != "" {
				if wd == homedir {
					fmt.Println("Can't delete home dir")
					exit(1)
				}
			}

			reader := bufio.NewReader(os.Stdin)
			fmt.Printf("Delete content of %q: [y/n] ", wd)
			text, _ := reader.ReadString('\n')
			text = strings.ToLower(text)
			text = strings.TrimSuffix(text, "\n")
			if text != "y" {
				fmt.Printf("%s\n", aurora.Red("abort"))
				exit(1)
			}

			files, err := os.ReadDir(wd)
			errz.Fatal(err)
			for _, file := range files {
				err = os.RemoveAll(file.Name())
				errz.Fatal(err)
			}
		}

		runPlayground()
	},
}

func runPlayground() {
	wd, err := os.Getwd()
	errz.Fatal(err)

	err = bob.CreatePlayground(bob.PlaygroundOptions{Dir: wd, ProjectName: "bob-playground", ProjectNameSecondLevel: "bob-playground-second-level"})
	errz.Fatal(err)
}
