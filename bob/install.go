package bob

import (
	"errors"
	"fmt"
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

	var allDeps []string
	for _, v := range ag.BTasks {
		for _, d := range v.AllDependencies {
			if inSlice(d, allDeps) {
				continue
			}
			allDeps = append(allDeps, d)
		}
	}

	if len(allDeps) == 0 {
		fmt.Println("Nothing to install.")
	}

	fmt.Println("Installing following dependencies:")
	for _, v := range allDeps {
		fmt.Println(v)
	}
	fmt.Println()

	if len(allDeps) > 0 {
		_, err = nix.BuildPackages(nix.FilterPackageNames(allDeps), ag.Nixpkgs)
		if err != nil {
			return err
		}
		_, err = nix.BuildFiles(nix.FilterNixFiles(allDeps), ag.Nixpkgs)
		if err != nil {
			return err
		}
	}

	return nil
}

func inSlice(a string, s []string) bool {
	for _, v := range s {
		if v == a {
			return true
		}
	}
	return false
}
