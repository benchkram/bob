package cli

import (
	"fmt"
	"os"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/bobtask/target"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean buildinfo and artifacts",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var cleanSystemCmd = &cobra.Command{
	Use:   "system",
	Short: "Remove buildinfo and local cache",
	Long: `Remove all entries from 
  ~/.bobcache/buildinfo 
  ~/.bobcache/artifacts`,
	Run: func(cmd *cobra.Command, args []string) {
		runCleanSystem()
	},
}

func runCleanSystem() {
	b, err := bob.Bob()
	boblog.Log.Error(err, "Unable to initialise bob")

	err = b.Clean()
	boblog.Log.Error(err, "Unable to clean buildinfo")

	fmt.Println("build info cleaned")
	fmt.Println("artifacts cleaned")
}

var cleanTargetsCmd = &cobra.Command{
	Use:   "targets",
	Short: "Remove targets declared by the current project",
	Long:  `Remove filesystem targets declared by the current project, docker targets are not removed`,
	Run: func(cmd *cobra.Command, args []string) {
		runCleanTargets()
	},
}

func runCleanTargets() {
	b, err := bob.Bob()
	boblog.Log.Error(err, "Unable to initialise bob")

	ag, err := b.Aggregate()
	if err != nil {
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else {
			errz.Fatal(err)
		}
		os.Exit(1)
	}

	for _, t := range ag.BTasks {
		if !t.TargetExists() {
			continue
		}

		taskTarget, err := t.Target()
		errz.Fatal(err)

		invalidFiles := make(map[string][]target.Reason)
		for _, v := range taskTarget.FilesystemEntriesRawPlain() {
			invalidFiles[v] = append(invalidFiles[v], target.ReasonCreatedAfterBuild)
		}

		err = t.Clean(invalidFiles, true)
		if err != nil {
			boblog.Log.Error(err, "Unable to clean target")
		}
	}
}
