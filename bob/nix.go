package bob

import (
	"fmt"
	"os/exec"
	"strings"
)

// NixBuild builds nix packages: nix-build --no-out-link -E 'with import <nixpkgs> { }; [pkg-1 pkg-2 pkg-3]'
// and returns the list of built store paths
func NixBuild(packages []string) ([]string, error) {
	fmt.Println("Building nix dependencies...")

	nixExpression := fmt.Sprintf("with import <nixpkgs> { }; [%s]", strings.Join(packages, " "))
	cmd := exec.Command("nix-build", "--no-out-link", "-E", nixExpression)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			fmt.Println(string(out))
		}
		return []string{}, err
	}

	fmt.Println(string(out))
	var storePaths []string
	for _, v := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(v, "/nix/store/") {
			storePaths = append(storePaths, v)
		}
	}

	return storePaths, nil
}

// StorePathsToPath creates a string ready to be added to $PATH appending /bin to each store path
func StorePathsToPath(storePaths []string) string {
	return strings.Join(storePaths, "/bin:") + "/bin"
}
