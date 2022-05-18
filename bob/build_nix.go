package bob

import (
	"fmt"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
)

// BuildNixDependenciesInPipeline collects and builds nix-dependencies for a pipeline starting at taskName.
func BuildNixDependenciesInPipeline(ag *bobfile.Bobfile, taskName string) (err error) {
	defer errz.Recover(&err)

	if !nix.IsInstalled() {
		return usererror.Wrap(fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl()))
	}

	tasksInPipeline, err := ag.BTasks.CollectTasksInPipeline(taskName)
	errz.Fatal(err)

	return BuildNixDependencies(ag, tasksInPipeline)
}

// BuildNixDependencies builds nix dependencies and prepares the affected tasks
// by setting the store paths on each task in the given aggregate.
func BuildNixDependencies(ag *bobfile.Bobfile, tasksInPipeline []string) (err error) {
	defer errz.Recover(&err)

	if !nix.IsInstalled() {
		return usererror.Wrap(fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl()))
	}

	nixDependencies, err := ag.BTasks.CollectNixDependenciesForTasks(tasksInPipeline)
	errz.Fatal(err)

	if len(nixDependencies) == 0 {
		return nil
	}

	depStorePathMapping, err := nix.BuildDependencies(nix.UniqueDeps(append(nix.DefaultPackages(ag.Nixpkgs), nixDependencies...)))
	errz.Fatal(err)

	// Resolve nix storePaths from dependencies
	// and rewrite the affected tasks.
	for _, name := range tasksInPipeline {
		t := ag.BTasks[name]

		if !t.UseNix() {
			continue
		}

		// construct used dependencies for this task
		deps := nix.DefaultPackages(ag.Nixpkgs)
		deps = append(deps, t.Dependencies()...)
		deps = nix.UniqueDeps(deps)

		storePaths, err := nix.DependenciesToStorePaths(deps, depStorePathMapping)
		errz.Fatal(err)

		t.SetStorePaths(storePaths)
		ag.BTasks[name] = t
	}

	return nil
}
