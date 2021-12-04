package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/logrusorgru/aurora"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/errz"

	"github.com/spf13/cobra"
)

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
				os.Exit(1)
			}

			homedir := os.Getenv("HOME")
			if homedir != "" {
				if wd == homedir {
					fmt.Println("Can't delete home dir")
					os.Exit(1)
				}
			}

			reader := bufio.NewReader(os.Stdin)
			fmt.Printf("Delete content of %q: [y/n] ", wd)
			text, _ := reader.ReadString('\n')
			text = strings.ToLower(text)
			text = strings.TrimSuffix(text, "\n")
			if text != "y" {
				fmt.Printf("%s\n", aurora.Red("abort"))
				os.Exit(1)
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

	err = bob.CreatePlayground(wd)
	errz.Fatal(err)
}
