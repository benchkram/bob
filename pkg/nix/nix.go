package nix

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/benchkram/errz"
)

type Dependency struct {
	// Name of the dependency
	Name string
	// Nixpkgs can be empty or a link to desired revision
	// ex. https://github.com/NixOS/nixpkgs/archive/eeefd01d4f630fcbab6588fe3e7fffe0690fbb20.tar.gz
	Nixpkgs string
}

type StorePath string
type DependenciesToStorePathMap map[Dependency]StorePath

// IsInstalled checks if nix is installed on the system
func IsInstalled() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}

// BuildDependencies build nix dependencies and returns a <package>-<nix store path> map
//
// dependencies can be either a package name ex. php or a path to .nix file
// nixpkgs can be empty which means it will use local nixpkgs channel
// or a link to desired revision ex. https://github.com/NixOS/nixpkgs/archive/eeefd01d4f630fcbab6588fe3e7fffe0690fbb20.tar.gz
func BuildDependencies(deps []Dependency, cache *Cache) (_ DependenciesToStorePathMap, err error) {
	defer errz.Recover(&err)

	pkgToStorePath := make(DependenciesToStorePathMap)

	for _, v := range deps {
		var key string

		if cache != nil {
			key, err = GenerateKey(v)
			errz.Fatal(err)
			if storePath, ok := cache.Get(key); ok {
				pkgToStorePath[v] = StorePath(storePath)
				continue
			}
		}

		if strings.HasSuffix(v.Name, ".nix") {
			storePath, err := buildFile(v.Name, v.Nixpkgs)
			if err != nil {
				return DependenciesToStorePathMap{}, err
			}
			pkgToStorePath[v] = StorePath(storePath)
		} else {
			storePath, err := buildPackage(v.Name, v.Nixpkgs)
			if err != nil {
				return DependenciesToStorePathMap{}, err
			}
			pkgToStorePath[v] = StorePath(storePath)
		}

		if cache != nil {
			err = cache.Save(key, string(pkgToStorePath[v]))
			errz.Fatal(err)
		}

	}
	return pkgToStorePath, nil
}

// buildPackage builds a nix package: nix-build --no-out-link -E 'with import <nixpkgs> { }; pkg' and returns the store path
func buildPackage(pkgName string, nixpkgs string) (string, error) {
	fmt.Println("DEBUG ENV")
	fmt.Println(os.Environ())

	nixExpression := fmt.Sprintf("with import %s { }; [%s]", source(nixpkgs), pkgName)
	cmd := exec.Command("nix-build", "--no-out-link", "-E", nixExpression)

	var stdoutBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	for _, v := range strings.Split(stdoutBuf.String(), "\n") {
		if strings.HasPrefix(v, "/nix/store/") {
			return v, nil
		}
	}

	return "", nil
}

// buildFile builds a .nix expression file
// `nix-build --no-out-link -E 'with import <nixpkgs> { }; callPackage filepath.nix {}'`
func buildFile(filePath string, nixpkgs string) (string, error) {
	nixExpression := fmt.Sprintf("with import %s { }; callPackage %s {}", source(nixpkgs), filePath)
	cmd := exec.Command("nix-build", "--no-out-link", "-E", nixExpression)

	var stdoutBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	for _, v := range strings.Split(stdoutBuf.String(), "\n") {
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
		out[i] = storePathBin(sp)
	}
	return out
}

// storePathBin adds the /bin dir to storePath
func storePathBin(storePath string) string {
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

// DependenciesToStorePaths resolves a dependency array to their
// associated nix storePath. The order of the output is guaranteed
// to match the order of the input.
func DependenciesToStorePaths(dependencies []Dependency, m DependenciesToStorePathMap) ([]string, error) {
	storePaths := make([]string, len(dependencies))

	for i, d := range dependencies {
		storePath, ok := m[d]
		if !ok {
			return nil, fmt.Errorf("could not resolve store path for [%s]", d)
		}
		storePaths[i] = string(storePath)
	}

	return storePaths, nil
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
