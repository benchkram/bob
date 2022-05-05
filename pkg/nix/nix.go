package nix

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type NewPathKey struct{}

type Dependency string
type StorePath string
type DependenciesToStorePathMap map[Dependency]StorePath

// IsInstalled checks if nix is installed on the system
func IsInstalled() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}

func Build(dependencies []string, nixpkgs string) (DependenciesToStorePathMap, error) {
	pkgToStorePath := make(DependenciesToStorePathMap)
	for _, v := range dependencies {
		if strings.HasSuffix(v, ".nix") {
			storePath, err := buildFile(v, nixpkgs)
			if err != nil {
				return DependenciesToStorePathMap{}, err
			}
			pkgToStorePath[Dependency(v)] = StorePath(storePath)
		} else {
			storePath, err := buildPackage(v, nixpkgs)
			if err != nil {
				return DependenciesToStorePathMap{}, err
			}
			pkgToStorePath[Dependency(v)] = StorePath(storePath)
		}
	}

	return pkgToStorePath, nil
}

// buildPackage builds nix package: nix-build --no-out-link -E 'with import <nixpkgs> { }; pkg' and returns the store path
func buildPackage(pkgName string, nixpkgs string) (string, error) {
	nixExpression := fmt.Sprintf("with import %s { }; [%s]", source(nixpkgs), pkgName)
	cmd := exec.Command("nix-build", "--no-out-link", "-E", nixExpression)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			return "", errors.New(string(out))
		}
		return "", err
	}

	fmt.Print(string(out))
	for _, v := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(v, "/nix/store/") {
			return v, nil
		}
	}

	return "", nil
}

func buildFile(filePath string, nixpkgs string) (string, error) {
	nixExpression := fmt.Sprintf("with import %s { }; callPackage %s {}", source(nixpkgs), filePath)
	cmd := exec.Command("nix-build", "--no-out-link", "-E", nixExpression)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			return "", errors.New(string(out))
		}
		return "", err
	}
	fmt.Print(string(out))
	for _, v := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(v, "/nix/store/") {
			return v, nil
		}
	}

	return "", nil
}

// DependenciesToStorePaths resolves a dependency array to their
// associated nix storePath. The order of the output is guaranteed
// to match the order of the input.
func DependenciesToStorePaths(dependencies []string, m DependenciesToStorePathMap) ([]string, error) {
	storePaths := make([]string, len(dependencies))
	for i, d := range dependencies {
		storePath, ok := m[Dependency(d)]
		if !ok {
			return nil, fmt.Errorf("could not resolve store path for [%s]", d)
		}
		storePaths[i] = string(storePath)
	}

	return storePaths, nil
}

// StorePathBin adds the /bin dir to a array of storePaths
func StorePathsBin(storePaths []string) []string {
	out := make([]string, len(storePaths))
	for i, sp := range storePaths {
		out[i] = StorePathBin(sp)
	}
	return out
}

// StorePathBin adds the /bin dir to storePath
func StorePathBin(storePath string) string {
	return filepath.Join(storePath, "/bin")
}

func DownloadURl() string {
	url := "https://nixos.org/download.html"

	switch runtime.GOOS {
	case "windows":
		url = "https://nixos.org/download.html#nix-install-windows"
	case "darwin":
		url = "https://nixos.org/download.html#nix-install-macos"
	case "linux":
		url = "https://nixos.org/download.html#nix-install-linux"
	}

	return url
}

func source(nixpkgs string) string {
	if nixpkgs != "" {
		return fmt.Sprintf("(fetchTarball \"%s\")", nixpkgs)
	}
	return "<nixpkgs>"
}
