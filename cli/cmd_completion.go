package cli

import (
	"fmt"
	"os"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash, zsh completions",
	Long: `To create completion add
	source <(bob completion)	   // for bash
	source <(bob completion -z)    // for zsh
to your .bashrc / .zshrc
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if zsh {
				err := rootCmd.GenZshCompletion(os.Stdout)
				if err != nil {
					boblog.Log.Error(err, "Unable to generate Zsh completion")
					exit(1)
				}
			} else {
				err := rootCmd.GenBashCompletionV2(os.Stdout, true)
				if err != nil {
					boblog.Log.Error(err, "Unable to generate bash completion")
					exit(1)
				}
			}

			return
		}

		switch args[0] {
		case "install":
			if zsh {
				// TODO
				fmt.Println("TODO")
			} else {
				completionPath := "/etc/bash_completion.d/bob"

				err := rootCmd.GenBashCompletionFileV2(completionPath, true)
				if err != nil {
					boblog.Log.Error(err, "Unable to install bash completion")
					exit(1)
				}
			}
		default:
			break
		}

	},
	ValidArgs: []string{"install"},
}
