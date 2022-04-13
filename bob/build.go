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

	ag, err := b.Aggregate()
	errz.Fatal(err)

	b.PrintVersionCompatibility(ag)

	var storePaths []string
	ag.Dependencies = append(ag.BTasks[taskName].Dependencies, ag.Dependencies...)
	if ag.UseNix && len(ag.Dependencies) > 0 {
		_, err = exec.LookPath("nix-build")
		errz.Fatal(err)
		storePaths, err = NixBuild(ag.Dependencies)
		errz.Fatal(err)
	}

	playbook, err := ag.Playbook(
		taskName,
		playbook.WithCachingEnabled(b.enableCaching),
	)
	errz.Fatal(err)

	if ag.UseNix && len(storePaths) > 0 {
		fmt.Printf("Updating $PATH to: %s\n", StorePathsToPath(storePaths))
		err = os.Setenv("PATH", StorePathsToPath(storePaths))
		errz.Fatal(err)
	}

	err = playbook.Build(ctx)
	errz.Fatal(err)

	return err
}
