package cli

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/boblog"
)

var zsh bool

func init() {
	configInit()

	// completionCmd
	completionCmd.Flags().BoolVarP(&zsh, "zsh", "z",
		zsh, "Create zsh completion")
	rootCmd.AddCommand(completionCmd)

	rootCmd.Flags().Bool("version", false, "Show the CLI's version")

	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(installCmd)

	// clone
	CmdClone.Flags().Bool("fail-fast", false, "Fail on first error without user prompt")
	rootCmd.AddCommand(CmdClone)

	// workspace
	cmdAdd.Flags().Bool("plain", false, "Do not infer contrary protocol url")
	cmdWorkspace.AddCommand(cmdWorkspaceNew)
	cmdWorkspace.AddCommand(cmdAdd)
	rootCmd.AddCommand(cmdWorkspace)

	// runCmd
	runCmd.Flags().Bool("no-cache", false, "Set to true to not use cache")
	runCmd.Flags().Bool("insecure", false, "Set to true to use http instead of https when accessing a remote artifact store")
	runCmd.AddCommand(runListCmd)
	rootCmd.AddCommand(runCmd)

	// buildCmd
	buildCmd.Flags().Bool("dummy", false, "Create a dummy bobfile")
	buildCmd.Flags().Bool("no-cache", false, "Set to true to not use cache")
	buildCmd.Flags().Bool("insecure", false, "Set to true to use http instead of https when accessing a remote artifact store")
	buildCmd.AddCommand(buildListCmd)
	rootCmd.AddCommand(buildCmd)

	// gitCmd
	CmdGitCommit.Flags().StringP("message", "m", "", "Set the commit message for all repository")
	CmdGit.AddCommand(CmdGitAdd)
	CmdGit.AddCommand(CmdGitCommit)
	CmdGit.AddCommand(CmdGitStatus)
	rootCmd.AddCommand(CmdGit)

	// authCmd
	AuthCmd.AddCommand(AuthContextCreateCmd)
	AuthContextCreateCmd.Flags().StringP("token", "t", "", "The token used for authentication")
	AuthCmd.AddCommand(AuthContextUpdateCmd)
	AuthContextUpdateCmd.Flags().StringP("token", "t", "", "The new token value")
	AuthCmd.AddCommand(AuthContextDeleteCmd)
	AuthCmd.AddCommand(AuthContextSwitchCmd)
	AuthCmd.AddCommand(AuthContextListCmd)
	AuthCmd.Flags().StringP("token", "t", "", "The token used for authentication")
	rootCmd.AddCommand(AuthCmd)
}

var rootCmd = &cobra.Command{
	Use:   "bob",
	Short: "A build tool from space, down on earth.",
	Long: `A build tool from space, down on earth.

Commonly used cmds:
    bob build          execute the default build task
    bob workspace new  create a new workspace
    bob git status     git like status for multi repositories`,
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("version") != nil {
			showVersion, err := strconv.ParseBool(cmd.Flag("version").Value.String())
			if err == nil && showVersion {
				//TODO for go 1.18: check what we can use from runtime/debug: https://github.com/golang/go/issues/49168
				//bi, ok := debug.ReadBuildInfo()
				//if ok {
				//
				//}

				boblog.Log.Info(fmt.Sprintf("bob version %s %s/%s\n", bob.Version, runtime.GOOS, runtime.GOARCH))
				os.Exit(0)
			}
		}

		readGlobalConfig()
		logInit(GlobalConfig.Verbosity)
		_stopProfiling = profiling(
			GlobalConfig.CPUProfile,
			GlobalConfig.MEMProfile,
		)
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		boblog.Log.Error(err, "Unable to generate bash completion")
	},
}
