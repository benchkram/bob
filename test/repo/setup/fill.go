package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/Benchkram/errz"
)

const (
	fileContents = "hello world\n"
)

func fillRepo(repo git.Repository, basePath string) error {
	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	if err := createAndCommitNewFile(wt, basePath, "initial", fileContents); err != nil {
		return err
	}

	if err := createAndSwitchToBranch(wt, "branch1"); err != nil {
		return err
	}
	if err := createAndCommitNewFile(wt, basePath, "file1", fileContents); err != nil {
		return err
	}

	if err := createAndSwitchToBranch(wt, "branch2"); err != nil {
		return err
	}
	if err := createAndCommitNewFile(wt, basePath, "file2", fileContents); err != nil {
		return err
	}

	if err := createAndSwitchToBranch(wt, "branch3"); err != nil {
		return err
	}
	if err := createAndCommitNewFile(wt, basePath, "file3", fileContents); err != nil {
		return err
	}

	return nil
}

func createNewFile(basePath, name, contents string) error {
	file, err := os.Create(filepath.Join(basePath, name))
	if err != nil {
		return fmt.Errorf("failed to create file %q: %w", name, err)
	}
	defer file.Close()
	if _, err := file.WriteString(contents); err != nil {
		return fmt.Errorf("failed to write to new file %q: %w", name, err)
	}

	return nil
}

func createAndCommitNewFile(wt *git.Worktree, basePath, name, contents string) error {
	if err := createNewFile(basePath, name, contents); err != nil {
		return err
	}

	if _, err := wt.Add(name); err != nil {
		return fmt.Errorf("failed to add file %q: %w", name, err)
	}

	msg := fmt.Sprintf("Add %s", name)
	err := commit(wt, msg)
	errz.Fatal(err)

	return nil
}

func commit(wt *git.Worktree, msg string) error {
	author := &object.Signature{
		Name:  "Bob The Builder",
		Email: "bob@thebuilder.com",
		When:  time.Now(),
	}

	if _, err := wt.Commit(msg, &git.CommitOptions{Author: author}); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

func createAndSwitchToBranch(wt *git.Worktree, name string) error {
	if err := wt.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", name)),
	}); err != nil {
		return fmt.Errorf("failed to create and checkout branch %q: %w", name, err)
	}

	return nil
}
