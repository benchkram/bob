package bob

import (
	"errors"
	"fmt"

	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/sliceutil"
	"github.com/benchkram/errz"
)

func (b B) Install() (err error) {
	defer errz.Recover(&err)

	ag, err := b.Aggregate()
	errz.Fatal(err)

	if !ag.UseNix {
		return errors.New("`use-nix: true` is missing in the root bob.yaml file")
	}

	if !nix.IsInstalled() {
		return fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl())
	}

	var allDeps []string
	for _, v := range ag.BTasks {
		allDeps = append(allDeps, v.DependenciesDirty...)
	}
	allDeps = sliceutil.Unique(allDeps)

	if len(allDeps) == 0 {
		fmt.Println("Nothing to install.")
	}

	fmt.Println("Installing following dependencies:")
	for _, v := range allDeps {
		fmt.Println(v)
	}
	fmt.Println()

	if len(allDeps) > 0 {
		fmt.Println("Building nix dependencies...")
		_, err = nix.Build(allDeps, ag.Nixpkgs)
	}

	return nil
}
