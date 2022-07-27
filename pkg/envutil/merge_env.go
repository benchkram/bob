package envutil

import "strings"

// MergeEnv will merge 2 environment lists
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

	// Add all from `b`
	newEnv = append(newEnv, b...)

	return newEnv
}
