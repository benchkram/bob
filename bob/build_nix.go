package bob

import (
	"fmt"
	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/sliceutil"
	"github.com/benchkram/bob/pkg/usererror"
	"strings"
)

// BuildNixForTask will collect and build dependencies for all tasks used in running of taskName
// adding the store paths to each task
func BuildNixForTask(ag *bobfile.Bobfile, taskName string) error {
	if !nix.IsInstalled() {
		return usererror.Wrap(fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl()))
	}

	tasksInPipeline := make([]string, 0)
	err := ag.BTasks.CollectTasksInPipeline(taskName, &tasksInPipeline)
	if err != nil {
		return err
	}

	nixDependencies := make([]nix.Dependency, 0)
	err = ag.BTasks.CollectNixDependencies(taskName, &nixDependencies)
	if err != nil {
		return err
	}

	if len(nixDependencies) == 0 {
		return nil
	}

	fmt.Println("Building nix dependencies...")
	storePaths, err := BuildNixDependencies(nix.UniqueDeps(append(nix.DefaultPackages(), nixDependencies...)))
	if err != nil {
		return err
	}

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

// BuildNixDependencies builds all nix dependencies inside nixDependencies
// and return the list of their store paths
// nixDependencies can be a package name or a .nix file path
func BuildNixDependencies(nixDependencies []nix.Dependency) ([]string, error) {
	storePaths := make([]string, len(nixDependencies))

	for k, v := range nixDependencies {
		if strings.HasSuffix(v.Name, ".nix") {
			storePath, err := nix.BuildFile(v.Name, v.Nixpkgs)
			if err != nil {
				return []string{}, err
			}
			storePaths[k] = storePath
		} else {
			storePath, err := nix.BuildPackage(v.Name, v.Nixpkgs)
			if err != nil {
				return []string{}, err
			}
			storePaths[k] = storePath
		}
	}
	return storePaths, nil
}
