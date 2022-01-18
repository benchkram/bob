package filepathutil

import "strings"

// IsChild checks if path2 is a child of path1.
// assumes both path starts from same directory and
// always returns true if path1 is root(e.g. `.`)
func IsChild(path1 string, path2 string) bool {
	path1Depth := len(strings.Split(path1, "/"))
	path2Depth := len(strings.Split(path2, "/"))

	if path1 == "." || (path2Depth > path1Depth && strings.HasPrefix(path2, path1)) {
		return true
	}

	return false
}
