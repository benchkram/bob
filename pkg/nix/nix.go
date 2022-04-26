package nix

import (
	"fmt"
	"os/exec"
	"strings"
)

// IsInstalled checks if nix is installed on the system
func IsInstalled() bool {
	_, err := exec.LookPath("nix")
	if err != nil {
		return false
	}
	return true
}

// BuildPackages builds nix packages: nix-build --no-out-link -E 'with import <nixpkgs> { }; [pkg-1 pkg-2 pkg-3]'
// and returns the list of built store paths
func BuildPackages(packages []string) ([]string, error) {
	fmt.Println("Building nix dependencies...")

	for _, v := range defaultPackages() {
		if !inSlice(v, packages) {
			packages = append(packages, v)
		}
	}

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

func defaultPackages() []string {
	return []string{
		"bash",
		"coreutils",
		"gnused",
		"findutils",
	}
}

func BuildFiles(files []string) ([]string, error) {
	fmt.Println("Building .nix files...")

	var storePaths []string
	for _, pkg := range files {
		nixExpression := fmt.Sprintf("with import <nixpkgs> { }; callPackage %s {}", pkg)
		cmd := exec.Command("nix-build", "--no-out-link", "-E", nixExpression)
		out, err := cmd.CombinedOutput()
		if err != nil {
			if len(out) > 0 {
				fmt.Println(string(out))
			}
			return []string{}, err
		}
		fmt.Print(string(out))
		for _, v := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(v, "/nix/store/") {
				storePaths = append(storePaths, v)
			}
		}
	}

	return storePaths, nil
}

// StorePathsToPath creates a string ready to be added to $PATH appending /bin to each store path
func StorePathsToPath(storePaths []string) string {
	return strings.Join(storePaths, "/bin:") + "/bin"
}

func FilterPackageNames(dependencies []string) []string {
	var res []string
	for _, v := range dependencies {
		if strings.HasSuffix(v, ".nix") {
			continue
		}
		res = append(res, v)
	}
	return res
}

func FilterNixFiles(dependencies []string) []string {
	var res []string
	for _, v := range dependencies {
		if !strings.HasSuffix(v, ".nix") {
			continue
		}
		res = append(res, v)
	}
	return res
}

func inSlice(a string, s []string) bool {
	for _, v := range s {
		if v == a {
			return true
		}
	}
	return false
}
