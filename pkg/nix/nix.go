package nix

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/benchkram/bob/pkg/usererror"
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
		} else {
			unsatisfiedDeps = append(unsatisfiedDeps, v)
		}
	}

	if len(unsatisfiedDeps) > 0 {
		fmt.Println("Building nix dependencies. This may take a while...")
	}

	for _, v := range unsatisfiedDeps {
		var br buildResult

		if strings.HasSuffix(v.Name, ".nix") {
			br, err = buildFile(v.Name, v.Nixpkgs)
			if err != nil {
				return DependenciesToStorePathMap{}, err
			}
			pkgToStorePath[v] = StorePath(br.storePath)
		} else {
			br, err = buildPackage(v.Name, v.Nixpkgs)
			if err != nil {
				return DependenciesToStorePathMap{}, err
			}
			pkgToStorePath[v] = StorePath(br.storePath)
		}

		fmt.Println()
		fmt.Printf("%s: %s took %s\n", v.Name, pkgToStorePath[v], displayDuration(br.duration))

		if cache != nil {
			key, err := GenerateKey(v)
			errz.Fatal(err)

			err = cache.Save(key, string(pkgToStorePath[v]))
			errz.Fatal(err)
		}
	}
	if len(unsatisfiedDeps) > 0 {
		fmt.Println("Succeeded building nix dependencies")
	}

	return pkgToStorePath, nil
}

type buildResult struct {
	storePath string
	duration  time.Duration
}

// buildPackage builds a nix package: nix-build --no-out-link -E 'with import <nixpkgs> { }; pkg' and returns the store path
func buildPackage(pkgName string, nixpkgs string) (buildResult, error) {
	nixExpression := fmt.Sprintf("with import %s { }; [%s]", source(nixpkgs), pkgName)
	cmd := exec.Command("nix-build", "--no-out-link", "-E", nixExpression)

	fmt.Printf("%s: ", pkgName)
	ticker := time.NewTicker(1 * time.Second)
	done := make(chan bool)

	start := time.Now()

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				fmt.Print(".")
			}
		}
	}()

	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf

	err := cmd.Run()
	if err != nil {
		ticker.Stop()
		done <- true
		return buildResult{}, usererror.Wrap(errors.New("could not build package"))
	}

	for _, v := range strings.Split(stdoutBuf.String(), "\n") {
		if strings.HasPrefix(v, "/nix/store/") {
			ticker.Stop()
			done <- true
			return buildResult{
				storePath: v,
				duration:  time.Since(start),
			}, nil
		}
	}

	return buildResult{}, nil
}

// buildFile builds a .nix expression file
// `nix-build --no-out-link -E 'with import <nixpkgs> { }; callPackage filepath.nix {}'`
func buildFile(filePath string, nixpkgs string) (buildResult, error) {
	nixExpression := fmt.Sprintf("with import %s { }; callPackage %s {}", source(nixpkgs), filePath)
	cmd := exec.Command("nix-build", "--no-out-link", "-E", nixExpression)

	fmt.Printf("%s: ", filePath)
	ticker := time.NewTicker(1 * time.Second)
	done := make(chan bool)
	start := time.Now()

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				fmt.Print(".")
			}
		}
	}()

	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf

	err := cmd.Run()
	if err != nil {
		ticker.Stop()
		done <- true
		return buildResult{}, usererror.Wrap(fmt.Errorf("could not build file `%s`", filePath))
	}

	for _, v := range strings.Split(stdoutBuf.String(), "\n") {
		ticker.Stop()
		done <- true
		if strings.HasPrefix(v, "/nix/store/") {
			return buildResult{
				storePath: v,
				duration:  time.Since(start),
			}, nil
		}
	}

	return buildResult{}, nil
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

// BuildEnvironment is running nix-shell for a list of dependencies and fetch its whole environment
//
// nix-shell --pure --keep NIX_SSL_CERT_FILE --keep SSL_CERT_FILE -p --command 'env' -E nixExpressionFromDeps
//
// nix shell can be started with empty list of packages so this method works with empty deps as well
func BuildEnvironment(deps []Dependency, nixpkgs string, cache *Cache) (_ []string, err error) {
	defer errz.Recover(&err)

	// building dependencies with nix-build to display store paths to output
	_, err = BuildDependencies(deps, cache)
	errz.Fatal(err)

	expression := nixExpression(deps, nixpkgs)

	var arguments []string
	arguments = append(arguments, []string{"--keep", "NIX_SSL_CERT_FILE", "--keep", "SSL_CERT_FILE"}...)
	arguments = append(arguments, []string{"--command", "env"}...)
	arguments = append(arguments, []string{"-E", expression}...)

	cmd := exec.Command("nix-shell", "--pure")
	cmd.Args = append(cmd.Args, arguments...)

	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	errz.Fatal(err)

	env := strings.Split(out.String(), "\n")

	// if NIX_SSL_CERT_FILE && SSL_CERT_FILE are set to /no-cert-file.crt unset them
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

// nixExpression computes the Nix expression which is passed to nix-shell via -E flag
// Example of a Nix expression containing go_1_18 and a custom oapicodegen_v1.6.0.nix file:
// { pkgs ? import <nixpkgs> {} }:
//
//	pkgs.mkShell {
//	 buildInputs = [
//	    pkgs.go_1_18
//	    (pkgs.callPackage ./oapicodegen_v1.6.0.nix { } )
//	 ];
//	}
func nixExpression(deps []Dependency, nixpkgs string) string {
	var buildInputs []string
	for _, v := range deps {
		if strings.HasSuffix(v.Name, ".nix") {
			buildInputs = append(buildInputs, fmt.Sprintf("(pkgs.callPackage %s{ } )", v.Name))
		} else {
			buildInputs = append(buildInputs, "pkgs."+v.Name)
		}
	}

	exp := `
{ pkgs ? import %s {} }:
pkgs.mkShell {
  buildInputs = [
	 %s
  ];
}
`
	return fmt.Sprintf(exp, source(nixpkgs), strings.Join(buildInputs, "\n"))
}

func displayDuration(d time.Duration) string {
	if d.Minutes() > 1 {
		return fmt.Sprintf("%.1fm", float64(d)/float64(time.Minute))
	}
	if d.Seconds() > 1 {
		return fmt.Sprintf("%.1fs", float64(d)/float64(time.Second))
	}
	return fmt.Sprintf("%.1fms", float64(d)/float64(time.Millisecond)+0.1) // add .1ms so that it never returns 0.0ms
}
