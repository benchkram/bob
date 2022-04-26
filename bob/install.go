package bob

import (
	"errors"
	"fmt"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/errz"
	"os/exec"
)

func (b B) Install() (err error) {
	defer errz.Recover(&err)

	ag, err := b.Aggregate()
	errz.Fatal(err)

	if !nix.IsInstalled() {
		return fmt.Errorf("nix is not installed on your system. Get it from %s", nix.DownloadURl())
	}

	if !ag.UseNix {
		return errors.New("`use-nix: true` is missing in the root bob.yaml file")
	}

	var allDeps []string

	for _, v := range ag.BTasks {
		for _, d := range v.Dependencies {
			if inSlice(d, allDeps) {
				continue
			}
			allDeps = append(allDeps, d)
		}
	}

	for _, v := range ag.Dependencies {
		if inSlice(v, allDeps) {
			continue
		}
		allDeps = append(allDeps, v)
	}

	fmt.Println("Installing following dependencies...")
	fmt.Println(allDeps)
	fmt.Println()

	if len(allDeps) > 0 {
		_, err = exec.LookPath("nix-build")
		errz.Fatal(err)
		_, err = nix.BuildPackages(nix.FilterPackageNames(allDeps))
		errz.Fatal(err)
		_, err := nix.BuildFiles(nix.FilterNixFiles(allDeps))
		errz.Fatal(err)
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
