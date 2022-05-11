package bob

import (
	"context"
	"errors"
	"fmt"

	"github.com/benchkram/bob/bob/playbook"
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

	fmt.Println("Building nix dependencies...")
	err = BuildNixDependenciesInPipeline(ag, taskName)
	errz.Fatal(err)
	fmt.Println("Succeded building nix dependencies")

	playbook, err := ag.Playbook(
		taskName,
		playbook.WithCachingEnabled(b.enableCaching),
	)
	errz.Fatal(err)

	err = playbook.Build(ctx)
	errz.Fatal(err)

	return err
}
