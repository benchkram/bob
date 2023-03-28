package bobutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/bob/global"
)

var ErrCouldNotFindBobWorkspace = fmt.Errorf("Could not find a .bob.workspace file")

// Hint on how git finds it top repo dir.
// https://stackoverflow.com/questions/65499497/how-does-git-know-its-in-a-git-repo

// FindBobRoot returns the absolute path to the bob root dir,
// starts directory traversal from working directory till
// `/$HOME` or `/` is reached.
func FindBobRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	var bobRoot bool
	if bobRoot, err = isBobRoot(dir); err != nil {
		return "", err
	}

	for !bobRoot {
		dir, err = filepath.Abs(filepath.Join(dir, "../"))
		if err != nil {
			return "", err
		}

		if dir == os.Getenv("HOME") || dir == "/" {
			return "", ErrCouldNotFindBobWorkspace
		}

		if bobRoot, err = isBobRoot(dir); err != nil {
			return "", err
		}
	}

	return dir, nil

}

// isBobRoot checks if a ".bob" folder is present in this directory
func isBobRoot(dir string) (bool, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	for _, f := range files {
		if f.Name() == global.BobWorkspaceFile {
			return true, nil
		}
	}

	return false, nil
}
