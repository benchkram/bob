package bob

import (
	"fmt"
	"github.com/benchkram/errz"
	"os/exec"
)

func (b B) Install() (err error) {
	defer errz.Recover(&err)

	ag, err := b.Aggregate()
	errz.Fatal(err)

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
		_, err = NixBuild(allDeps)
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
