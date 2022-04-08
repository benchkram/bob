package bob

import (
	"context"
	"errors"
	"fmt"
	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/errz"
	"os"
	"os/exec"
	"strings"
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

	playbook, err := aggregate.Playbook(
		taskname,
		playbook.WithCachingEnabled(b.enableCaching),
	)
	errz.Fatal(err)

	err = playbook.Build(ctx)
	errz.Fatal(err)

	err = nixBuild(aggregate.Dependencies)
	errz.Fatal(err)
	err = clearNixBuildResults(aggregate.Dependencies)
	errz.Fatal(err)

	return err
}

// nix-build -E 'with import <nixpkgs> { }; [pkg-1 pkg-2 pkg-3]'
func nixBuild(packages []string) error {
	nixExpression := fmt.Sprintf("with import <nixpkgs> { }; [%s]", strings.Join(packages, " "))
	cmd := exec.Command("nix-build", "-E", nixExpression)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

//Remove result files created after nix-build
func clearNixBuildResults(packages []string) error {
	for k := range packages {
		var fileName string
		if k == 0 {
			fileName = "result"
		} else {
			fileName = fmt.Sprintf("result-%d", k+1)
		}
		err := os.Remove(fileName)
		if err != nil {
			return err
		}
	}
	return nil
}
