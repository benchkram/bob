package bob

import (
	"context"
	"errors"
	"fmt"

	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/sliceutil"
	"github.com/benchkram/errz"
)

var (
	ErrNoRebuildRequired = errors.New("no rebuild required")
)

// Build a task and it's dependencies.
func (b *B) Build(ctx context.Context, taskName string) (err error) {
	defer errz.Recover(&err)

	ag, err := b.Aggregate()
	errz.Fatal(err)

	b.PrintVersionCompatibility(ag)

	// TODO: Don't check on the aggregate if UseNix is true...
	// check if for each bobfile/ task
	//
	// only build/use the dependencies on task which
	// have UseNix set to true

	if ag.UseNix {
		// Gather nix dependencies from tasks
		nixDependencies := []string{}
		nixDependencies = append(nixDependencies, ag.Dependencies...)
		tasksInPipeline := []string{}
		err = ag.BTasks.Walk(taskName, "", func(tn string, task bobtask.Task, err error) error {
			if err != nil {
				return err
			}
			tasksInPipeline = append(tasksInPipeline, task.Name())
			nixDependencies = append(nixDependencies, task.Dependencies()...)
			return err
		})
		errz.Fatal(err)

		// Build nix dependencies & resolve nix store paths
		fmt.Println("Building nix dependencies...", ag.Nixpkgs)
		nixDependencies = sliceutil.Unique(
			append(nix.DefaultPackages(), nixDependencies...),
		)
		depStorePathMapping, err := nix.Build(nixDependencies, ag.Nixpkgs)
		errz.Fatal(err)

		// Resolve nix storePaths from dependencies``
		// and rewrite the affected tasks.
		for _, tn := range tasksInPipeline {
			t := ag.BTasks[tn]

			// construct used dependencies for this task
			deps := nix.DefaultPackages()
			deps = append(deps, t.Dependencies()...)
			deps = sliceutil.Unique(deps)

			// resolve nix store paths from dependencies
			storePaths, err := nix.DependenciesToStorePaths(deps, depStorePathMapping)
			errz.Fatal(err)

			fmt.Printf("Setting dependencies for task %s\n", t.Name())
			t.SetStorePaths(storePaths)
			ag.BTasks[tn] = t
		}
	}

	playbook, err := ag.Playbook(
		taskName,
		playbook.WithCachingEnabled(b.enableCaching),
	)
	errz.Fatal(err)

	err = playbook.Build(ctx)
	errz.Fatal(err)

	return err
}
