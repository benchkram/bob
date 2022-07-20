package nix

import "strings"

// ReplacePATH replaces PATH from env with the list of all /nix/store paths
func ReplacePATH(storePaths []string, env []string) []string {
	for k, v := range env {
		pair := strings.SplitN(v, "=", 2)
		if pair[0] == "PATH" {
			env[k] = "PATH=" + strings.Join(StorePathsBin(storePaths), ":")
		}
	}
	return env
}
