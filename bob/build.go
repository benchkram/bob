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

	fmt.Println("Building nix dependencies...")
	err = b.nix.BuildNixDependenciesInPipeline(ag, taskName)
	errz.Fatal(err)
	fmt.Println("Succeeded building nix dependencies")

	// Hint: Hash computation (playbook execution) can only start after
	// nix dependencies are resolved.
	// Nix dependencies are considered in the input hash of a task.
	p, err := ag.Playbook(
		taskName,
		playbook.WithCachingEnabled(b.enableCaching),
		playbook.WithPredictedNumOfTasks(len(ag.BTasks)),
		playbook.WithMaxParallel(b.maxParallel),
		playbook.WithRemoteStore(ag.Remotestore()),
		playbook.WithLocalStore(b.local),
	)
	errz.Fatal(err)

	err = p.Build(ctx)
	errz.Fatal(err)

	return nil
}
