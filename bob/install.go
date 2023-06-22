package bob

import (
	"fmt"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/pkg/nix"
)

func (b B) Install() (err error) {
	defer errz.Recover(&err)

	ag, err := b.Aggregate()
	errz.Fatal(err)

	if !nix.IsInstalled() {
		return fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl())
	}

	var allDeps []nix.Dependency
	for _, v := range ag.BTasks {
		allDeps = append(allDeps, v.Dependencies()...)
	}
	allDeps = nix.UniqueDeps(allDeps)

	if len(allDeps) == 0 {
		fmt.Println("Nothing to install.")
	}

	fmt.Println("Installing following dependencies:")
	for _, v := range allDeps {
		fmt.Println(v.Name)
	}
	fmt.Println()

	if len(allDeps) > 0 {
		err = b.nix.BuildDependencies(allDeps)
		if err != nil {
			return err
		}
	}

	return nil
}
