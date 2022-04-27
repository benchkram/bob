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

	if len(allDeps) == 0 {
		fmt.Println("Nothing to install.")
	}

	fmt.Println("Installing following dependencies:")
	for _, v := range allDeps {
		fmt.Println(v)
	}
	fmt.Println()

	if len(allDeps) > 0 {
		_, err = nix.BuildPackages(nix.FilterPackageNames(allDeps))
		if err != nil {
			return err
		}
		_, err = nix.BuildFiles(nix.FilterNixFiles(allDeps))
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
