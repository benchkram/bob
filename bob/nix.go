package bob

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// NixBuild builds nix packages: nix-build -E 'with import <nixpkgs> { }; [pkg-1 pkg-2 pkg-3]'
func NixBuild(packages []string) error {
	nixExpression := fmt.Sprintf("with import <nixpkgs> { }; [%s]", strings.Join(packages, " "))
	cmd := exec.Command("nix-build", "-E", nixExpression)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ClearNixBuildResults removes result files created after nix-build
func ClearNixBuildResults(packages []string) error {
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
