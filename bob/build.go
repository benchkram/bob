package bob

import (
	"context"
	"errors"
	"fmt"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bob/playbook"
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

	if b.nix != nil {
		fmt.Println("Building nix dependencies...")
		err = b.nix.BuildNixDependenciesInPipeline(ag, taskName)
		errz.Fatal(err)
		fmt.Println("Succeeded building nix dependencies")
	}

	playbook, err := ag.Playbook(
		taskName,
		playbook.WithCachingEnabled(b.enableCaching),
	)
	errz.Fatal(err)

	// HINT: It's not easily possible to parallelize hash computation
	// with building nix dependecies.. as the storePaths computed by
	// BuildNixDependenciesInPipeline are considered in the task input hash.
	err = playbook.PreComputeInputHashes()
	errz.Fatal(err)

	err = playbook.Build(ctx)
	errz.Fatal(err)

	return err
}
