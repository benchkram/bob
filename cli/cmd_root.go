package cli

import (
	"github.com/spf13/cobra"

	"github.com/Benchkram/errz"
)

var zsh bool

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
