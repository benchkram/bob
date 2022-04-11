package bob

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// NixBuild builds nix packages: nix-build -E 'with import <nixpkgs> { }; [pkg-1 pkg-2 pkg-3]'
// and returns the list of built store paths
func NixBuild(packages []string) ([]string, error) {
	nixExpression := fmt.Sprintf("with import <nixpkgs> { }; [%s]", strings.Join(packages, " "))

	cmd := exec.Command("nix-build", "-E", nixExpression)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}, err
	}

	return strings.Split(string(out), "\n"), nil
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

// StorePathsToPath creates a string ready to be added to $PATH appending /bin to each store path
func StorePathsToPath(storePaths []string) string {
	return strings.TrimSuffix(strings.Join(storePaths, "/bin:"), ":")
}
