package nix

import "strings"

// AddPATH adds to PATH the list of all /nix/store paths
func AddPATH(storePaths []string, env []string) []string {
	env = append(env, "PATH="+strings.Join(StorePathsBin(storePaths), ":"))
	return env
}
