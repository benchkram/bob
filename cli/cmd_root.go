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
var flagEnvVars []string

func init() {
	configInit()

	// completionCmd
	completionCmd.Flags().BoolVarP(&zsh, "zsh", "z",
		zsh, "Create zsh completion")
	rootCmd.AddCommand(completionCmd)

	rootCmd.Flags().Bool("version", false, "Show the CLI's version")

	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(initCmd)

	// clone
	CmdClone.Flags().Bool("fail-fast", false, "Fail on first error without user prompt")
	CmdClone.Flags().Bool("ssh", false, "Prefer ssh for cloning")
	CmdClone.Flags().Bool("https", false, "Prefer https for cloning")
	rootCmd.AddCommand(CmdClone)

	// workspace
	cmdAdd.Flags().Bool("plain", false, "Do not infer contrary protocol url")
	cmdWorkspace.AddCommand(cmdWorkspaceNew)
	cmdWorkspace.AddCommand(cmdAdd)
	rootCmd.AddCommand(cmdWorkspace)

	// runCmd
	runCmd.Flags().Bool("no-cache", false, "Set to true to not use cache")
	runCmd.Flags().Bool("insecure", false, "Set to true to use http instead of https when accessing a remote artifact store")
	runCmd.Flags().StringSliceVar(&flagEnvVars, "env", []string{}, "Set environment variables to run task")
	runCmd.AddCommand(runListCmd)
	rootCmd.AddCommand(runCmd)

	// buildCmd
	buildCmd.Flags().Bool("dummy", false, "Create a dummy bobfile")
	buildCmd.Flags().Bool("no-cache", false, "Set to true to not use cache")
	buildCmd.Flags().Bool("push", false, "Set to true to push artifacts to remote store")
	buildCmd.Flags().Bool("no-pull", false, "Set to true to disable artifacts download from remote store")
	buildCmd.Flags().Bool("insecure", false, "Set to true to use http instead of https when accessing a remote artifact store")
	buildCmd.Flags().IntP("jobs", "j", runtime.NumCPU(), "Maximum number of parallel started jobs")
	buildCmd.Flags().StringSliceVar(&flagEnvVars, "env", []string{}, "Set environment variables to build task")
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
	rootCmd.AddCommand(AuthCmd)

	// cleanCmd
	cleanCmd.AddCommand(cleanTargetsCmd)
	cleanCmd.AddCommand(cleanSystemCmd)
	rootCmd.AddCommand(cleanCmd)
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
				// TODO for go 1.18: check what we can use from runtime/debug: https://github.com/golang/go/issues/49168
				// bi, ok := debug.ReadBuildInfo()
				// if ok {
				//
				// }

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
