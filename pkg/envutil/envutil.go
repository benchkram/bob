package envutil

import "strings"

// TODO: FIXME: create a global object to store task environments
// to avoid storing the same environment for each task multiple times
// This could reduce garbage collection.

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

	println("herererr")

	for _, v := range b {
		pair := strings.SplitN(v, "=", 2)
		envKeys[pair[0]] = v
	}

	//var newEnv []string
	// Add from `a` the keys which are not in `b`
	for _, v := range a {
		pair := strings.SplitN(v, "=", 2)
		if _, exists := envKeys[pair[0]]; exists {
			continue
		}
		envKeys[pair[0]] = v
		//newEnv = append(newEnv, v)
	}

	// Populate the result slice with the keys from the map
	result := make([]string, 0, len(envKeys))
	for _, v := range envKeys {
		result = append(result, v)
	}

	return result
}

func MergeChat(a []string, b []string) []string {
	// Use a map to store the keys in `b` to avoid duplicates
	envKeys := make(map[string]bool)

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

	// Populate the map with the keys in `b`
	for _, v := range b {
		pair := strings.SplitN(v, "=", 2)
		envKeys[pair[0]] = true
	}

	// Add from `a` the keys which are not in `b`
	for _, v := range a {
		pair := strings.SplitN(v, "=", 2)
		if _, exists := envKeys[pair[0]]; exists {
			continue
		}
		envKeys[pair[0]] = true
	}

	// Populate the result slice with the keys from the map
	result := make([]string, 0, len(envKeys))
	for key := range envKeys {
		result = append(result, key)
	}

	return result
}
