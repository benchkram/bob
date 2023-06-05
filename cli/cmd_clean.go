package cli

import (
	"fmt"
	"os"

	"github.com/benchkram/bob/bob"
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

var cleanAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Execute bob clean system & bob clean target together",
	Long:  `Execute bob clean system & bob clean target together`,
	Run: func(cmd *cobra.Command, args []string) {
		runCleanTargets()
		runCleanSystem()
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
	boblog.Log.Error(err, "Unable to clean [oneOf buildinfo, environement-cache, artifacts or .nix_cache ] ")

	fmt.Println("build info cleaned")
	fmt.Println("artifacts cleaned")
	fmt.Println("env cache cleaned")
	fmt.Println(".nix_cache cleaned")
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

		err = t.Clean(true)
		errz.Fatal(err)
	}
}
