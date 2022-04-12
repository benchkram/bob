package bob

import (
	"context"
	"errors"
	"fmt"
	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/errz"
	"os"
	"os/exec"
)

var (
	ErrNoRebuildRequired = errors.New("no rebuild required")
)

// Build a task and it's dependencies.
func (b *B) Build(ctx context.Context, taskName string) (err error) {
	defer errz.Recover(&err)

	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	b.PrintVersionCompatibility(aggregate)

	var storePaths []string
	aggregate.Dependencies = append(aggregate.BTasks[taskName].Dependencies, aggregate.Dependencies...)
	if len(aggregate.Dependencies) > 0 {
		_, err = exec.LookPath("nix-build")
		errz.Fatal(err)
		storePaths, err = NixBuild(aggregate.Dependencies)
		errz.Fatal(err)
	}

	playbook, err := aggregate.Playbook(
		taskName,
		playbook.WithCachingEnabled(b.enableCaching),
	)
	errz.Fatal(err)

	if len(storePaths) > 0 {
		fmt.Printf("Updating $PATH to: %s\n", StorePathsToPath(storePaths))
		err = os.Setenv("PATH", StorePathsToPath(storePaths))
		errz.Fatal(err)
	}

	err = playbook.Build(ctx)
	errz.Fatal(err)

	return err
}
