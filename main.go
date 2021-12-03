package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Benchkram/errz"
)

var zsh bool

var _stopProfiling func()

func stopProfiling() {
	if _stopProfiling != nil {
		_stopProfiling()
	}
}

func init() {
	configInit()

	// completionCmd
	completionCmd.Flags().BoolVarP(&zsh, "zsh", "z",
		zsh, "Create zsh completion")
	rootCmd.AddCommand(completionCmd)

	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(CmdClone)
	rootCmd.AddCommand(cleanCmd)
	// used for debugging and not part of the cli.
	// rootCmd.AddCommand(composeCmd)

	// workspace
	cmdWorkspace.AddCommand(cmdWorkspaceNew)
	cmdWorkspace.AddCommand(cmdAdd)
	rootCmd.AddCommand(cmdWorkspace)

	// runCmd
	runCmd.AddCommand(runListCmd)
	rootCmd.AddCommand(runCmd)

	// playgroundCmd
	playgroundCmd.Flags().Bool("clean", false, "Delete directory content before creating the playground")
	rootCmd.AddCommand(playgroundCmd)

	// inspectCmd
	inspectCmd.AddCommand(envCmd)
	inspectCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(inspectCmd)

	// buildCmd
	buildCmd.Flags().Bool("dummy", false, "Create a dummy bobfile")
	buildCmd.AddCommand(buildListCmd)
	rootCmd.AddCommand(buildCmd)

	// gitCmd
	CmdGit.AddCommand(CmdGitStatus)
	rootCmd.AddCommand(CmdGit)
}

var rootCmd = &cobra.Command{
	Use:   "bob",
	Short: "cli to run bob - the build tool",
	Long:  `TODO`,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		readGlobalConfig()
		logInit(GlobalConfig.Verbosity)
		_stopProfiling = profiling(
			GlobalConfig.CPUProfile,
			GlobalConfig.MEMProfile,
		)
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		errz.Fatal(err)
	},
}

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash, zsh completions",
	Long: `To create completion add
	source <(bob completion)	   // for bash
	source <(bob completion -z)    // for zsh
# ~/.bashrc or ~/.profile ~/.zsh???
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if zsh {
				err := rootCmd.GenZshCompletion(os.Stdout)
				if err != nil {
					errz.Log(err)
					os.Exit(1)
				}
			} else {
				err := rootCmd.GenBashCompletionV2(os.Stdout, true)
				if err != nil {
					errz.Log(err)
					os.Exit(1)
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
					errz.Log(err)
					os.Exit(1)
				}
			}
		default:
			break
		}

	},
	ValidArgs: []string{"install"},
}

func main() {
	var exitCode int
	defer func() { os.Exit(exitCode) }()
	defer stopProfiling()

	if err := rootCmd.Execute(); err != nil {
		errz.Log(err)
		exitCode = 1
	}
}
