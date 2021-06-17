package filepathutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yargevad/filepathx"
)

func ListRecursive(inp string) ([]string, error) {
	var all []string

	if s, err := os.Stat(inp); err != nil || !s.IsDir() {
		// Use glob for unknowns (wildcard-paths) and existing files (non-dirs)
		matches, err := filepathx.Glob(inp)
		if err != nil {
			return nil, fmt.Errorf("failed to glob %q: %w", inp, err)
		}

		for _, m := range matches {
			if s, err := os.Stat(m); err == nil && !s.IsDir() {
				// Existing file
				all = append(all, m)
			} else {
				// Directory
				files, err := listDir(m)
				if err != nil {
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

	return all, nil
}

func listDir(path string) ([]string, error) {
	var all []string
	if err := filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
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
