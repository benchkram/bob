package bobgit

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Benchkram/bob/pkg/bobutil"
	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
	git "github.com/go-git/go-git/v5"
)

var ErrEmptyCommitMessage = fmt.Errorf("Bobgit does not allow empty message")

// Commit executes `git commit -m ${message}` in all repositories
// first level repositories found inside a .bob filtree.
func Commit(message string) (err error) {
	defer errz.Recover(&err)

	if message == "" {
		return usererror.Wrapm(ErrEmptyCommitMessage, "Aborting commit due to empty commit message")
	}

	bobRoot, err := bobutil.FindBobRoot()
	errz.Fatal(err)

	err = os.Chdir(bobRoot)
	errz.Fatal(err)

	// Assure toplevel is a git repo
	isGit, err := isGitRepo(bobRoot)
	errz.Fatal(err)
	if !isGit {
		return usererror.Wrap(ErrCouldNotFindGitDir)
	}

	// search for git repos inside bobRoot/.
	repoNames := []string{}
	err = filepath.WalkDir(bobRoot, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return nil
		}

		if d.Name() == ".git" {
			p, err := filepath.Rel(bobRoot, filepath.Dir(path))
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
	errz.Fatal(err)

	// repos with no changes, throws exit status 1 error
	// while executing `git commit --dry-run`
	// only repos with changes filtered out
	filteredRepo, err := filterModifiedRepos(repoNames)
	errz.Fatal(err)

	// execute dry-run first on all repositories
	// to check for errors
	for _, name := range filteredRepo {
		_, err := cmdutil.GitDryCommit(name, message)
		if err != nil {
			return usererror.Wrapm(err, "Failed to commit to git dry run")
		}
	}

	for _, name := range repoNames {
		_, err := cmdutil.GitCommit(name, message)
		if err != nil {
			return usererror.Wrapm(err, "Failed to commit to git")
		}
	}

	return nil
}

// filterModifiedRepos filters the repositories with changes
// by running `git status` command on each repository and look for
// modified files. returns a list of repository with changes.
func filterModifiedRepos(repolist []string) ([]string, error) {
	updatedRepo := []string{}

	for _, name := range repolist {
		output, err := cmdutil.GitStatus(name)
		if err != nil {
			return updatedRepo, err
		}

		status, err := parse(output)
		if err != nil {
			return updatedRepo, err
		}

		modified := false
		for _, filestatus := range status {
			if filestatus.Staging != git.Unmodified || filestatus.Worktree != git.Unmodified {
				modified = true
				break
			}
		}

		if modified {
			updatedRepo = append(updatedRepo, name)
		}
	}

	return updatedRepo, nil
}
