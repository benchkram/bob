package envutil

import "strings"

// Merge two lists of environment variables in the "key=value" format.
// If variables are duplicated, the one from `b` is kept.
// See tests for details.
func Merge(a []string, b []string) []string {
	envKeys := make(map[string]string)

	if len(a) == 0 {
		r := make([]string, len(b))
		copy(r, b)
		return r
	}
	if len(b) == 0 {
		r := make([]string, len(a))
		copy(r, a)
		return r
	}

	for _, v := range b {
		pair := strings.SplitN(v, "=", 2)
		envKeys[pair[0]] = v
	}

	// Add from `a` the keys which are not in `b`
	for _, v := range a {
		pair := strings.SplitN(v, "=", 2)
		if _, exists := envKeys[pair[0]]; exists {
			continue
		}
		envKeys[pair[0]] = v
	}

	// Populate the result slice with the keys from the map
	result := make([]string, 0, len(envKeys))
	for _, v := range envKeys {
		result = append(result, v)
	}

	return result
}
