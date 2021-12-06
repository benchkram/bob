package cli

import (
	"fmt"
	"github.com/Benchkram/bob/bob"
	"github.com/spf13/cobra"
	"os"
	"runtime"
	"strconv"

	"github.com/Benchkram/errz"
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
		if cmd.Flag("version") != nil {
			showVersion, err := strconv.ParseBool(cmd.Flag("version").Value.String())
			if err == nil && showVersion {
				//TODO for go 1.18: check what we can use from runtime/debug: https://github.com/golang/go/issues/49168
				//bi, ok := debug.ReadBuildInfo()
				//if ok {
				//
				//}

				fmt.Printf("bob version v%s %s/%s\n", bob.Version, runtime.GOOS, runtime.GOARCH)
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
		errz.Fatal(err)
	},
}
