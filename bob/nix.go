package bob

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"time"
)

// NixBuildPackages builds nix packages: nix-build --no-out-link -E 'with import <nixpkgs> { }; [pkg-1 pkg-2 pkg-3]'
// and returns the list of built store paths
func NixBuildPackages(packages []string) ([]string, error) {
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

// NixBuildFiles builds a nix expression from all .nix files defined in dependencies
//
// nix-build -E 'with import <nixpkgs> { };
//
//let
//  packages = rec {
//    randName = callPackage ./firstPackage.nix {};
//    randName = callPackage ./secondPackage.nix {};
//
//    inherit pkgs;
//  };
//in
//  packages
//'
func NixBuildFiles(files []string) ([]string, error) {
	fmt.Println("Building .nix files...")

	packagesListSection := ""
	for _, filePath := range files {
		packagesListSection += fmt.Sprintf("%s = callPackage %s {};", randPackageName(6), filePath)
	}

	nixExpression := fmt.Sprintf("with import <nixpkgs> { }; let packages = rec {%s inherit pkgs; }; in packages", packagesListSection)
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

func randPackageName(length int) string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
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
