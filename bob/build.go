package bob

import (
	"context"
	"errors"
	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/errz"
	"os"
	"os/exec"
)

var (
	ErrNoRebuildRequired = errors.New("no rebuild required")
)

// Build a task and it's dependencies.
func (b *B) Build(ctx context.Context, taskname string) (err error) {
	defer errz.Recover(&err)

	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	b.PrintVersionCompatibility(aggregate)

	var storePaths []string
	aggregate.Dependencies = append(aggregate.Dependencies, aggregate.BTasks[taskname].Dependencies...)
	if len(aggregate.Dependencies) > 0 {
		_, err = exec.LookPath("nix-build")
		errz.Fatal(err)
		storePaths, err = NixBuild(aggregate.Dependencies)
		errz.Fatal(err)
		err = ClearNixBuildResults(aggregate.Dependencies)
		errz.Fatal(err)
	}

	playbook, err := aggregate.Playbook(
		taskname,
		playbook.WithCachingEnabled(b.enableCaching),
	)
	errz.Fatal(err)

	if len(storePaths) > 0 {
		err = os.Setenv("PATH", StorePathsToPath(storePaths))
		errz.Fatal(err)
	}

	err = playbook.Build(ctx)
	errz.Fatal(err)

	return err
}
