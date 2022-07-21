package cli

import (
	"bytes"
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
				var buf bytes.Buffer
				err := rootCmd.GenZshCompletion(&buf)
				if err != nil {
					boblog.Log.Error(err, "Unable to generate Zsh completion")
					exit(1)
				}
				_, err = buf.WriteString("\ncompdef _bob bob\n")
				if err != nil {
					boblog.Log.Error(err, "Unable to generate Zsh completion")
					exit(1)
				}
				fmt.Println(buf.String())
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
