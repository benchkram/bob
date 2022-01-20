package cli

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/Benchkram/bob/pkg/boblog"

	"github.com/Benchkram/bob/bob"
	"github.com/spf13/cobra"
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
	rootCmd.AddCommand(CmdClone)
	rootCmd.AddCommand(cleanCmd)

	// workspace
	cmdWorkspace.AddCommand(cmdWorkspaceNew)
	cmdWorkspace.AddCommand(cmdAdd)
	rootCmd.AddCommand(cmdWorkspace)

	// runCmd
	runCmd.Flags().Bool("no-cache", false, "Set to true to not use cache")
	runCmd.AddCommand(runListCmd)
	rootCmd.AddCommand(runCmd)

	// buildCmd
	buildCmd.Flags().Bool("dummy", false, "Create a dummy bobfile")
	buildCmd.Flags().Bool("no-cache", false, "Set to true to not use cache")
	buildCmd.AddCommand(buildListCmd)
	rootCmd.AddCommand(buildCmd)

	// gitCmd
	CmdGit.AddCommand(CmdGitStatus)
	rootCmd.AddCommand(CmdGit)

	// aqua test
	rootCmd.AddCommand(aquaCmd)
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
