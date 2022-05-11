package bob

import (
	"fmt"
	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
)

// BuildNixForTask will collect and build dependencies for all tasks used in running of taskName
// adding the store paths to each task
func BuildNixForTask(ag *bobfile.Bobfile, taskName string) (err error) {
	defer errz.Recover(&err)

	if !nix.IsInstalled() {
		return usererror.Wrap(fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl()))
	}

	tasksInPipeline := make([]string, 0)
	err = ag.BTasks.CollectTasksInPipeline(taskName, &tasksInPipeline)
	errz.Fatal(err)

	nixDependencies := make([]nix.Dependency, 0)
	err = ag.BTasks.CollectNixDependencies(taskName, &nixDependencies)
	errz.Fatal(err)

	if len(nixDependencies) == 0 {
		return nil
	}

	fmt.Println("Building nix dependencies...")
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
