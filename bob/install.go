package bob

import (
	"errors"
	"fmt"
	"strings"

	"github.com/benchkram/bob/pkg/nix"
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
		fmt.Println("Building nix dependencies...")
		for _, v := range allDeps {
			if strings.HasSuffix(v.Name, ".nix") {
				_, err := nix.BuildFile(v.Name, v.Nixpkgs)
				if err != nil {
					return err
				}
			} else {
				_, err := nix.BuildPackage(v.Name, v.Nixpkgs)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
