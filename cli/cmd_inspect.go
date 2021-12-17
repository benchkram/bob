package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

func init() {
	inspectCmd.AddCommand(envCmd)
	inspectCmd.AddCommand(exportCmd)
	inspectCmd.AddCommand(inspectArtifactCmd)
	rootCmd.AddCommand(inspectCmd)
}

var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect Tasks of a Bobfile",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "List environment for a task",
	Args:  cobra.ExactArgs(1),
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		taskname := args[0]
		runEnv(taskname)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		tasks, err := getTasks()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		return tasks, cobra.ShellCompDirectiveDefault
	},
}

func runEnv(taskname string) {
	b, err := bob.Bob()
	boblog.Log.Error(err, "Unable to initialise bob")

	bobfile, err := b.Aggregate()
	boblog.Log.Error(err, "Unable to aggregate bob file")

	task, ok := bobfile.Tasks[taskname]
	if !ok {
		fmt.Printf("%s\n", aurora.Red("Task does not exists"))
		os.Exit(1)
	}

	for _, env := range task.Env() {
		println(env)
	}
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "List exports for a task",
	Args:  cobra.ExactArgs(1),
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		taskname := args[0]
		runExport(taskname)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		tasks, err := getTasks()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		return tasks, cobra.ShellCompDirectiveDefault
	},
}

func runExport(taskname string) {
	b, err := bob.Bob()
	boblog.Log.Error(err, "Unable to initialize bob")

	bobfile, err := b.Aggregate()
	boblog.Log.Error(err, "Unable to aggregate bob file")

	task, ok := bobfile.Tasks[taskname]
	if !ok {
		fmt.Printf("%s\n", aurora.Red("Task does not exists"))
		os.Exit(1)
	}

	for exportname, export := range task.GetExports() {
		fmt.Printf("%s (%s)\n", exportname, export)
	}
}

var inspectArtifactCmd = &cobra.Command{
	Use:   "artifact",
	Short: "Inspect artifacts",
	Args:  cobra.ExactArgs(1),
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		taskname := args[0]
		runInspectArtifact(taskname)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		tasks, err := getTasks()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		return tasks, cobra.ShellCompDirectiveDefault
	},
}

func runInspectArtifact(taskname string) {
	b, err := bob.Bob()
	boblog.Log.Error(err, "Unable to initialize bob")

	bobfile, err := b.Aggregate()
	boblog.Log.Error(err, "Unable to aggregate bob file")

	task, ok := bobfile.Tasks[taskname]
	if !ok {
		fmt.Printf("%s\n", aurora.Red("Task does not exists"))
		os.Exit(1)
	}

	info, err := task.ArtifactInspect()
	if err != nil {
		if errors.As(err, &usererror.Err) {
			fmt.Printf("%s\n", errors.Unwrap(err).Error())
			os.Exit(1)
		}
		errz.Log(err)
	}

	fmt.Printf("%s", info.String())
}
