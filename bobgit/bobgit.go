package bobgit

import (
	"io/fs"
	"path/filepath"
)

// search for git repos inside provided root directory
func getAllRepos(root string) ([]string, error) {
	repoNames := []string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return nil
		}

		if d.Name() == ".git" {
			p, err := filepath.Rel(root, filepath.Dir(path))
			if err != nil {
				return err
			}
			repoNames = append(repoNames, p)
		}

		for _, dir := range dontFollow {
			if d.Name() == dir {
				return fs.SkipDir
			}
		}

		return nil
	})

	if err != nil {
		return repoNames, err
	}

	return repoNames, nil
}
