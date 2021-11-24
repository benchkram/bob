package reposetup

import (
	"fmt"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
)

func createAndFillRepo(basePath, name string) error {
	const isBare = false

	repoPath := filepath.Join(basePath, name)
	repo, err := git.PlainInit(repoPath, isBare)
	if err != nil {
		return fmt.Errorf("failed to create repo %q: %w", repoPath, err)
	}

	if err := fillRepo(*repo, repoPath); err != nil {
		return fmt.Errorf("failed to fill repo %q: %w", repoPath, err)
	}

	return nil
}
