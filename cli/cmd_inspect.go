package cli

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

var inspectArtifactId string

func init() {

	inspectArtifactCmd.Flags().StringVarP(&inspectArtifactId, "id", "",
		inspectArtifactId, "inspect artifact with id")

	inspectCmd.AddCommand(inputCmd)
	inspectCmd.AddCommand(envCmd)
	inspectCmd.AddCommand(exportCmd)
	inspectArtifactCmd.AddCommand(inspectArtifactListCmd)
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
		tasks, err := getBuildTasks()
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

	task, ok := bobfile.BTasks[taskname]
	if !ok {
		fmt.Printf("%s\n", aurora.Red("Task does not exists"))
		exit(1)
	}

	// Build nix dependencies
	err = b.Nix().BuildNixDependencies(bobfile, []string{taskname}, []string{})
	errz.Fatal(err)

	task = bobfile.BTasks[taskname]

	taskEnv := task.Env()
	if len(task.StorePaths()) > 0 {
		taskEnv = nix.AddPATH(task.StorePaths(), task.Env())
	}
	for _, e := range taskEnv {
		println(e)
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
		tasks, err := getBuildTasks()
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

	task, ok := bobfile.BTasks[taskname]
	if !ok {
		fmt.Printf("%s\n", aurora.Red("Task does not exists"))
		exit(1)
	}

	for exportname, export := range task.GetExports() {
		fmt.Printf("%s (%s)\n", exportname, export)
	}
}

var inspectArtifactCmd = &cobra.Command{
	Use:   "artifact",
	Short: "Inspect artifacts by id",
	// Args:  cobra.ExactArgs(1),
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		if inspectArtifactId == "" {
			fmt.Printf("%s", aurora.Red("failed to set artifact id"))
			exit(1)
		}
		runInspectArtifact(inspectArtifactId)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		tasks, err := getBuildTasks()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		return tasks, cobra.ShellCompDirectiveDefault
	},
}

func runInspectArtifact(artifactID string) {
	b, err := bob.Bob()
	boblog.Log.Error(err, "Unable to initialize bob")

	info, err := b.ArtifactInspect(artifactID)
	if err != nil {
		if errors.As(err, &usererror.Err) {
			fmt.Printf("%s\n", errors.Unwrap(err).Error())
			exit(1)
		}
		errz.Log(err)
	}

	fmt.Printf("%s", info.String())
}

var inspectArtifactListCmd = &cobra.Command{
	Use:   "ls",
	Short: "List artifacts",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runInspectArtifactList()
	},
}

// runinspectArtifactList list artifacts in relation to tasks
func runInspectArtifactList() {
	b, err := bob.Bob()
	if err != nil {
		boblog.Log.Error(err, "Unable to initialize bob")
	}

	out, err := b.ArtifactList(context.TODO())
	if err != nil {
		boblog.Log.Error(err, "Unable to generate artifact list")
	}
	fmt.Println(out)
}

var inputCmd = &cobra.Command{
	Use:   "input",
	Short: "List inputs",
	Args:  cobra.ExactArgs(1),
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		taskname := args[0]
		runInspectInputs(taskname)
	},
}

// runInspectInputs list artifacts in relation to tasks
func runInspectInputs(taskname string) {
	b, err := bob.Bob()
	boblog.Log.Error(err, "Unable to initialise bob")

	bobfile, err := b.Aggregate()
	boblog.Log.Error(err, "Unable to aggregate bob file")

	task, ok := bobfile.BTasks[taskname]
	if !ok {
		fmt.Printf("%s\n", aurora.Red("Task does not exists"))
		exit(1)
	}

	task = bobfile.BTasks[taskname]

	inputs := task.Inputs()
	inputs = sort.StringSlice(inputs)
	for i, e := range inputs {
		fmt.Println(e)
		if i > 10 {
			break
		}
	}

	fmt.Printf("Task %s has %d inputs\n", taskname, len(inputs))
}
