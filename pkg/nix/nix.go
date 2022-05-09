package nix

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// IsInstalled checks if nix is installed on the system
func IsInstalled() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}

// BuildPackage builds a nix package: nix-build --no-out-link -E 'with import <nixpkgs> { }; pkg' and returns the store path
func BuildPackage(pkgName string, nixpkgs string) (string, error) {
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

// BuildFile builds a .nix expression file
// `nix-build --no-out-link -E 'with import <nixpkgs> { }; callPackage filepath.nix {}'`
func BuildFile(filePath string, nixpkgs string) (string, error) {
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

// StorePathsBin adds the /bin dir to each of storePaths
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

// DownloadURl give nix download URL based on OS
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

// AddDir add the dir path to .nix files specified in dependencies
func AddDir(dir string, dependencies []string) []string {
	for k, v := range dependencies {
		if strings.HasSuffix(v, ".nix") {
			dependencies[k] = dir + "/" + v
		}
	}
	return dependencies
}

// Source of nixpkgs from where dependencies are built. If empty will use local <nixpkgs>
// or a specific tarball can be used ex. https://github.com/NixOS/nixpkgs/archive/eeefd01d4f630fcbab6588fe3e7fffe0690fbb20.tar.gz
func source(nixpkgs string) string {
	if nixpkgs != "" {
		return fmt.Sprintf("(fetchTarball \"%s\")", nixpkgs)
	}
	return "<nixpkgs>"
}
