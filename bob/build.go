package bob

import (
	"context"
	"errors"
	"fmt"
	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/errz"
	"os/exec"
	"strings"
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

	workingDirectoryBobFile, err := bobfile.BobfileRead(ag.Dir())
	errz.Fatal(err)

	workingFileDeps := make([]string, len(workingDirectoryBobFile.Dependencies))
	for k, v := range workingDirectoryBobFile.Dependencies {
		if strings.HasSuffix(v, ".nix") {
			workingFileDeps[k] = ag.Dir() + "/" + v
		} else {
			workingFileDeps[k] = v
		}
	}
	allDepsToInstall := append(ag.BTasks[taskName].Dependencies, workingFileDeps...)

	if ag.UseNix && !nix.IsInstalled() {
		return fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl())
	}

	if len(allDepsToInstall) > 0 && !ag.UseNix {
		fmt.Println("Found a list of dependencies, but use-nix is false")
	}

	var storePaths []string
	if ag.UseNix && len(allDepsToInstall) > 0 {
		_, err = exec.LookPath("nix-build")
		errz.Fatal(err)
		storePaths, err = nix.BuildPackages(nix.FilterPackageNames(allDepsToInstall))
		errz.Fatal(err)
		storePathsFromFiles, err := nix.BuildFiles(nix.FilterNixFiles(allDepsToInstall))
		errz.Fatal(err)
		storePaths = append(storePaths, storePathsFromFiles...)
	}

	playbook, err := ag.Playbook(
		taskName,
		playbook.WithCachingEnabled(b.enableCaching),
	)
	errz.Fatal(err)

	if ag.UseNix && len(storePaths) > 0 {
		ctx = context.WithValue(ctx, nix.NewPathKey{}, nix.StorePathsToPath(storePaths))
	}

	err = playbook.Build(ctx)
	errz.Fatal(err)

	return err
}
