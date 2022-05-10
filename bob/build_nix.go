package bob

import (
	"fmt"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/sliceutil"
	"github.com/benchkram/bob/pkg/usererror"
)

// BuildNix will collect and build dependencies for all tasks used in running of taskName
// adding the store paths to each task
func BuildNix(ag *bobfile.Bobfile, taskName string) error {
	if !ag.UseNix {
		return nil
	}

	if !nix.IsInstalled() {
		return usererror.Wrap(fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl()))
	}

	// Gather nix dependencies from tasks
	nixDependencies := make([]string, 0)
	var tasksInPipeline []string
	err := ag.BTasks.Walk(taskName, "", func(tn string, task bobtask.Task, err error) error {
		if err != nil {
			return err
		}
		tasksInPipeline = append(tasksInPipeline, task.Name())
		if task.UseNix() {
			nixDependencies = append(nixDependencies, task.Dependencies()...)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if len(nixDependencies) == 0 {
		return nil
	}
	fmt.Println("Building nix dependencies...")

	// FIXME: warn or abort build when there are Bobfiles with different Nixpkgs.
	// Curently only nixpkgs of aggregate is considered.
	storePaths, err := nix.Build(
		sliceutil.Unique(append(nix.DefaultPackages(), nixDependencies...)),
		ag.Nixpkgs,
	)
	if err != nil {
		return err
	}

	for _, name := range tasksInPipeline {
		t := ag.BTasks[name]
		t.SetStorePaths(sliceutil.Unique(storePaths))
		ag.BTasks[name] = t
	}

	return nil
}
