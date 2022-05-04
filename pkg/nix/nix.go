package nix

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type NewPathKey struct{}

// IsInstalled checks if nix is installed on the system
func IsInstalled() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}

func Build(dependencies []string, nixpkgs string) (map[string]string, error) {
	for _, v := range defaultPackages() {
		if !inSlice(v, dependencies) {
			dependencies = append(dependencies, v)
		}
	}
	pkgToStorePath := make(map[string]string)
	for _, v := range dependencies {
		if strings.HasSuffix(v, ".nix") {
			storePath, err := buildFile(v, nixpkgs)
			if err != nil {
				return map[string]string{}, err
			}
			pkgToStorePath[v] = storePath
		} else {
			storePath, err := buildPackage(v, nixpkgs)
			if err != nil {
				return map[string]string{}, err
			}
			pkgToStorePath[v] = storePath
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

func defaultPackages() []string {
	return []string{
		"bash",
		"coreutils",
		"gnused",
		"findutils",
	}
}

// StorePathsToPath creates a string ready to be added to $PATH appending /bin to each store path
func StorePathsToPath(storePaths []string) string {
	return strings.Join(storePaths, "/bin:") + "/bin"
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

func inSlice(a string, s []string) bool {
	for _, v := range s {
		if v == a {
			return true
		}
	}
	return false
}
