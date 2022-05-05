package bob

import (
	"fmt"
	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/sliceutil"
	"github.com/benchkram/bob/pkg/usererror"
)

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
		if !task.UseNix() {
			return nil
		}
		tasksInPipeline = append(tasksInPipeline, task.Name())
		nixDependencies = append(nixDependencies, task.Dependencies()...)
		return nil
	})

	if err != nil {
		return err
	}

	if len(nixDependencies) == 0 {
		return nil
	}

	fmt.Println("Building nix dependencies...")
	depStorePathMapping, err := nix.Build(sliceutil.Unique(append(nix.DefaultPackages(), nixDependencies...)), ag.Nixpkgs)
	if err != nil {
		return err
	}

	// Resolve nix storePaths from dependencies
	// and rewrite the affected tasks.
	for _, name := range tasksInPipeline {
		t := ag.BTasks[name]

		// construct used dependencies for this task
		deps := nix.DefaultPackages()
		deps = append(deps, t.Dependencies()...)
		deps = sliceutil.Unique(deps)

		storePaths, err := nix.DependenciesToStorePaths(deps, depStorePathMapping)
		if err != nil {
			return err
		}
		t.SetStorePaths(storePaths)
		ag.BTasks[name] = t
	}

	return nil
}
