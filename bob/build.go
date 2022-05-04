package bob

import (
	"context"
	"errors"
	"fmt"
	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/pkg/nix"
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

	pkgToStorePath := make(map[string]string)
	err = ag.BTasks.Walk(taskName, "", func(tn string, task bobtask.Task, err error) error {
		if err != nil {
			return err
		}
		if len(task.AllDependencies) == 0 {
			return nil
		}
		fmt.Println("Building nix dependencies...", ag.Nixpkgs)
		pkgToStorePath, err = nix.Build(task.AllDependencies, ag.Nixpkgs)
		return err
	})
	errz.Fatal(err)

	playbook, err := ag.Playbook(
		taskName,
		playbook.WithCachingEnabled(b.enableCaching),
		playbook.WithPkgToStorePath(pkgToStorePath),
	)
	errz.Fatal(err)
	err = playbook.Build(ctx)
	errz.Fatal(err)

	return err
}
