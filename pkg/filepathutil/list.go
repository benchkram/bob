package filepathutil

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/yargevad/filepathx"
)

// DefaultIgnores
var (
	DefaultIgnores = map[string]bool{
		"node_modules": true,
		".git":         true,
	}
)

//var listRecursiveMap = make(map[string][]string, 1024)

func ListRecursive(inp string) (all []string, err error) {

	// result, ok := listRecursiveMap[inp]
	// if ok {
	// 	return result, nil
	// }

	// TODO: new list recursive
	// * does input contain a glob? see https://pkg.go.dev/path/filepath#Match => read with filepathx.Glob
	// * check if input is a file => add file
	// * check if input is a dir => add files in dir recursively
	//
	// More input:
	// https://github.com/iriri/minimal/blob/9b2348d09c1ab2c25505f9933a3591ef9db6522a/gitignore/gitignore.go#L245
	// https://github.com/zabawaba99/go-gitignore/
	// https://github.com/gobwas/glob
	//
	// Thoughts: Is it possible to compile a ignoreList upfront?
	// Then check if the accessed file || dir can be skipped.
	// Maybe it's even possible to call skipdir on a walk func.

	// TODO: possibly ignore here too, before calling listDir
	if s, err := os.Stat(inp); err != nil || !s.IsDir() {
		// File
		// Use glob for unknowns (wildcard-paths) and existing files (non-dirs)
		matches, err := filepathx.Glob(inp)
		if err != nil {
			return nil, fmt.Errorf("failed to glob %q: %w", inp, err)
		}

		for _, m := range matches {
			s, err := os.Stat(m)
			if err == nil && !s.IsDir() {

				// Existing file
				all = append(all, m)
			} else {
				// Directory
				files, err := listDir(m)
				if err != nil {
					// TODO: handle error
					return nil, fmt.Errorf("failed to list dir: %w", err)
				}

				all = append(all, files...)
			}
		}
	} else {
		// Directory
		files, err := listDir(inp)
		if err != nil {
			return nil, fmt.Errorf("failed to list dir: %w", err)
		}

		all = append(all, files...)
	}

	// listRecursiveMap[inp] = all
	return all, nil
}

var WalkedDirs = map[string]int{}

func listDir(path string) ([]string, error) {
	times := WalkedDirs[path]
	WalkedDirs[path] = times + 1

	var all []string
	if err := filepath.WalkDir(path, func(p string, fi fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip default ignored
		if fi.IsDir() && ignored(fi.Name()) {
			return fs.SkipDir
		}

		// Skip dirs
		if !fi.IsDir() {
			all = append(all, p)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk dir %q: %w", path, err)
	}

	return all, nil
}

func ignored(fileName string) bool {
	return DefaultIgnores[fileName]
}
