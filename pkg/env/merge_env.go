package env

import "strings"

// MergeEnv will merge to environment lists
// See tests for usage
func MergeEnv(a []string, b []string) []string {
	envKeys := make(map[string]bool)

	for _, v := range b {
		pair := strings.SplitN(v, "=", 2)
		envKeys[pair[0]] = true
	}

	var newEnv []string
	// Add from `a` the keys which are not in `b`
	for _, v := range a {
		pair := strings.SplitN(v, "=", 2)
		if _, exists := envKeys[pair[0]]; exists {
			continue
		}
		newEnv = append(newEnv, v)
	}

	// Add from `b` remaining ones
	for _, v := range b {
		newEnv = append(newEnv, v)
	}

	return newEnv
}
