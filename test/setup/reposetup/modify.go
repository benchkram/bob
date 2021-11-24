package reposetup

import (
	"fmt"

	git "github.com/go-git/go-git/v5"
)

func ModifyRepo(path string) error {
	basePath := path

	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open repo: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	if err := modifyFile(basePath, "file1", "new contents"); err != nil {
		return err
	}

	if err := createNewFile(basePath, "untrackedfile", fileContents); err != nil {
		return err
	}

	const newFileName = "addedfile"
	if err := createNewFile(basePath, newFileName, fileContents); err != nil {
		return err
	}
	if _, err := wt.Add(newFileName); err != nil {
		return fmt.Errorf("failed to add file %q: %w", newFileName, err)
	}

	return nil
}

func modifyFile(basePath, name, contents string) error {
	return createNewFile(basePath, name, contents)
}
