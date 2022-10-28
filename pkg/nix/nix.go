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

	var unsatisfiedDeps []Dependency
	pkgToStorePath := make(DependenciesToStorePathMap)

	for _, v := range deps {
		if cache != nil {
			key, err := GenerateKey(v)
			errz.Fatal(err)

			if storePath, ok := cache.Get(key); ok {
				pkgToStorePath[v] = StorePath(storePath)
				continue
			}
			unsatisfiedDeps = append(unsatisfiedDeps, v)
		}
	}

	if len(unsatisfiedDeps) > 0 {
		fmt.Println("Building nix dependencies...")
		defer fmt.Println("Succeeded building nix dependencies")
	}

	for _, v := range unsatisfiedDeps {
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
			key, err := GenerateKey(v)
			errz.Fatal(err)

			err = cache.Save(key, string(pkgToStorePath[v]))
			errz.Fatal(err)
		}
	}
	return pkgToStorePath, nil
}

// buildPackage builds a nix package: nix-build --no-out-link -E 'with import <nixpkgs> { }; pkg' and returns the store path
func buildPackage(pkgName string, nixpkgs string) (string, error) {
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

// BuildEnvironment is running nix-shell for a list of dependencies and fetch its whole environment
//
// nix-shell --pure -p package1 package2 --command 'env' -I nixpkgs=tarballURL
//
// The -I nixpkgs=tarballURL is added only if deps have Nixpkgs URL set
func BuildEnvironment(deps []Dependency) (_ []string, err error) {
	defer errz.Recover(&err)

	var listOfPackages []string
	var nixpkgs string
	for _, v := range deps {
		if strings.HasSuffix(v.Name, ".nix") {
			continue
		}
		listOfPackages = append(listOfPackages, v.Name)
		if v.Nixpkgs != "" {
			nixpkgs = v.Nixpkgs
		}
	}

	arguments := append([]string{"--pure", "-p", "--command", "'env'", "--keep", "NIX_SSL_CERT_FILE SSL_CERT_FILE"}, listOfPackages...)

	if nixpkgs != "" {
		arguments = append(arguments, "-I", "nixpkgs="+nixpkgs)
	}

	cmd := exec.Command("nix-shell", arguments...)
	fmt.Println("CMD:", cmd.String())

	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		errz.Fatal(err)
	}

	env := strings.Split(out.String(), "\n")

	// build .nix files
	var fileStorePaths []string
	for _, v := range deps {
		if !strings.HasSuffix(v.Name, ".nix") {
			continue
		}
		storePath, err := buildFile(v.Name, v.Nixpkgs)
		if err != nil {
			return []string{}, err
		}
		fileStorePaths = append(fileStorePaths, storePath)
	}

	// add file store paths to existing env PATH
	for k, e := range env {
		pair := strings.SplitN(e, "=", 2)
		if pair[0] == "PATH" {
			updatedPath := "PATH=" + pair[1] + ":" + strings.Join(StorePathsBin(fileStorePaths), ":")
			env[k] = updatedPath
		}
	}

	// fix NIX_SSL_CERT_FILE && SSL_CERT_FILE
	var clearedEnv []string
	for _, e := range env {
		pair := strings.SplitN(e, "=", 2)
		if pair[0] == "NIX_SSL_CERT_FILE" && pair[1] == "/no-cert-file.crt" {
			continue
		}
		if pair[0] == "SSL_CERT_FILE" && pair[1] == "/no-cert-file.crt" {
			continue
		}
		clearedEnv = append(clearedEnv, e)
	}

	return clearedEnv, nil
}
