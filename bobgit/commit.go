package bobgit

import (
	"fmt"
	"os"

	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/Benchkram/bob/pkg/bobutil"
	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
	git "github.com/go-git/go-git/v5"
)

var ErrEmptyCommitMessage = fmt.Errorf("Bobgit does not allow empty message")

// Commit executes `git commit -m ${message}` in all repositories.
//
// indifferent of the subdirectories and subrepositories,
// it walks through all the repositories starting from bobroot
// and run `git commit -m {message}` command.
//
// Only shows user messages in case of nothing to commit.
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
	repoNames, err := findRepos(bobRoot)
	errz.Fatal(err)

	// repos with no changes, throws exit status 1 error
	// while executing `git commit --dry-run`
	// only repos with changes filtered out
	filteredRepo, untrackedRepo, err := filterModifiedRepos(repoNames)
	errz.Fatal(err)

	// throw some user message for untracked repositories
	if len(filteredRepo) == 0 {
		s := "nothing to commit, working trees are clean."
		if len(untrackedRepo) > 0 {
			s = "nothing added to commit but untracked files present."
		}
		boblog.Log.V(1).Info(s)
		return nil
	}

	// execute dry-run first on all repositories
	// to check for errors
	for _, name := range filteredRepo {
		_, err := cmdutil.GitDryCommit(name, message)
		if err != nil {
			return usererror.Wrapm(err, "Failed to commit to git in repo "+name)
		}
	}

	for _, name := range filteredRepo {
		_, err := cmdutil.GitCommit(name, message)
		if err != nil {
			return usererror.Wrapm(err, "Failed to commit to git in repo "+name)
		}
	}

	return nil
}

// filterModifiedRepos filters the repositories with changes
// by running `git status` command on each repository and look for
// tracked files in staging.
//
// returns a list of repository which consist tracked files,
// also returns a list of untracked but modified repositories.
func filterModifiedRepos(repolist []string) ([]string, []string, error) {
	updatedRepo := []string{}
	untrackedRepo := []string{}

	for _, name := range repolist {
		output, err := cmdutil.GitStatus(name)
		if err != nil {
			return updatedRepo, untrackedRepo, err
		}

		status, err := parse(output)
		if err != nil {
			return updatedRepo, untrackedRepo, err
		}

		tracked := false
		untracked := false
		for _, filestatus := range status {
			if filestatus.Staging != git.Untracked {
				tracked = true
				break
			}

			if !untracked && filestatus.Worktree != git.Unmodified {
				untracked = true
			}
		}

		if tracked {
			updatedRepo = append(updatedRepo, name)
		} else if untracked {
			untrackedRepo = append(untrackedRepo, name)
		}
	}

	return updatedRepo, untrackedRepo, nil
}
