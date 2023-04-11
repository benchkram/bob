package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/filehash"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
	"github.com/logrusorgru/aurora"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
)

var inspectArtifactId string

func init() {

	inspectArtifactCmd.Flags().StringVarP(&inspectArtifactId, "id", "",
		inspectArtifactId, "inspect artifact with id")

	inspectCmd.AddCommand(inputCmd)
	inspectCmd.AddCommand(envCmd)
	// artifact
	inspectArtifactCmd.AddCommand(inspectArtifactListCmd)
	inspectCmd.AddCommand(inspectArtifactCmd)
	//diff
	inspectBuildInfoCmd.AddCommand(inspectBuildInfoDiffCmd)
	inspectCmd.AddCommand(inspectBuildInfoCmd)
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
	task = bobfile.BTasks[taskname]

	taskEnv, err := b.Nix().BuildEnvironment(task.Dependencies(), task.Nixpkgs())
	errz.Fatal(err)

	for _, e := range taskEnv {
		println(e)
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

	bobfile, err := b.AggregateWithNixDeps(taskname)
	boblog.Log.Error(err, "Unable to aggregate bob file")

	task, ok := bobfile.BTasks[taskname]
	if !ok {
		fmt.Printf("%s\n", aurora.Red("Task does not exists"))
		exit(1)
	}

	task = bobfile.BTasks[taskname]

	inputs := task.Inputs()
	sort.Strings(inputs)
	for _, e := range inputs {
		contentHash, err := filehash.HashOfFile(e)
		if err != nil {
			boblog.Log.Error(err, "unable to compute hash of file")
			exit(1)
		}

		info, err := os.Stat(e)
		if err != nil {
			boblog.Log.Error(err, "unable to stat ffile hash of file")
			exit(1)
		}

		fmt.Printf("\t%s %d %s\n", e, info.Size(), contentHash)
	}

	hash, err := task.HashIn()
	if err != nil {
		boblog.Log.Error(err, "unable to compute hash")
		exit(1)
	}

	fmt.Println()
	fmt.Println()
	fmt.Printf("Summary:\n")
	fmt.Printf("\ttask-name:           %s\n", taskname)
	fmt.Printf("\t# of inputs:         %d\n", len(inputs))
	fmt.Printf("\tinput hash:          %s\n", hash)
}

var inspectBuildInfoCmd = &cobra.Command{
	Use:   "buildinfo",
	Short: "Inspect build info",
	Args:  cobra.ExactArgs(1),
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		inspectBuildInfo(args[0])
	},
}

func inspectBuildInfo(hash string) {
	var exitCode int
	defer func() { os.Exit(exitCode) }()

	bs, err := bob.DefaultBuildinfoStore()
	if err != nil {
		panic(err)
	}

	bi, err := bs.GetBuildInfo(hash)
	if err != nil {
		panic(err)
	}

	fmt.Println(bi.Describe())
}

var inspectBuildInfoDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff build infos",
	Args:  cobra.ExactArgs(2),
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		diffBuildInfo(args[0], args[1])
	},
}

func diffBuildInfo(hashA, hashB string) {
	var exitCode int
	defer func() { os.Exit(exitCode) }()

	bs, err := bob.DefaultBuildinfoStore()
	if err != nil {
		panic(err)
	}

	biA, err := bs.GetBuildInfo(hashA)
	if err != nil {
		panic(err)
	}

	biB, err := bs.GetBuildInfo(hashB)
	if err != nil {
		panic(err)
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(biA.Describe(), biB.Describe(), false)
	fmt.Println(dmp.DiffPrettyText(diffs))
}
