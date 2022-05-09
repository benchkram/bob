package bob

import (
	"fmt"
	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/sliceutil"
	"github.com/benchkram/bob/pkg/usererror"
	"strings"
)

// BuildNix will collect and build dependencies for all tasks used in running of taskName
// adding the store paths to each task
func BuildNix(ag *bobfile.Bobfile, taskName string) error {
	if !nix.IsInstalled() {
		return usererror.Wrap(fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl()))
	}

	// Gather nix dependencies from tasks
	nixDependencies := make([]bobtask.Dependency, 0)
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

	storePaths := make([]string, len(nixDependencies))
	for _, v := range nixDependencies {
		if strings.HasSuffix(v.Name, ".nix") {
			storePath, err := nix.BuildFile(v.Name, v.Nixpkgs)
			if err != nil {
				return err
			}
			storePaths = append(storePaths, storePath)
		} else {
			storePath, err := nix.BuildPackage(v.Name, v.Nixpkgs)
			if err != nil {
				return err
			}
			storePaths = append(storePaths, storePath)
		}
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
