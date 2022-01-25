package bobgit

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Benchkram/errz"
	"github.com/cli/cli/git"
)

var (
	ErrOutsideBobWorkspace = fmt.Errorf("Not allowed, path pointing outside of Bob workspace.")
	ErrCouldNotFindGitDir  = fmt.Errorf("Could not find a .git folder")
)

var dontFollow = []string{
	"node_modules",
	".git",
	".vscode",
}

// isGitRepo return true is the directory contains a `.git` directory
func isGitRepo(dir string) (isGit bool, err error) {
	defer errz.Recover(&err)
	entrys, err := os.ReadDir(dir)
	errz.Fatal(err)

	for _, entry := range entrys {
		if !entry.IsDir() {
			continue
		}

		if entry.Name() == ".git" {
			isGit = true
			break
		}
	}

	return isGit, nil
}

// findRepos searches for git repos inside provided root directory
func findRepos(root string) ([]string, error) {
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

// getRepoConfig detects repository current branch and
// returns the Branch Config with remote name, url and merge ref
func getRepoConfig(root string, repo string) (_ *git.BranchConfig, err error) {
	repoPath := filepath.Join(root, repo)

	err = os.Chdir(repoPath)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = os.Chdir(root)
		errz.Fatal(err)
	}()

	branch, err := git.CurrentBranch()
	if err != nil {
		return nil, err
	}

	config := git.ReadBranchConfig(branch)
	return &config, nil
}

// formatRepoNameForOutput returns formatted reponame for output.
//
// Example: "." => "/", "second-level" => "second-level/"
func formatRepoNameForOutput(reponame string) string {
	repopath := reponame
	if reponame == "." {
		repopath = "/"
	} else if repopath[len(repopath)-1:] != "/" {
		repopath = repopath + "/"
	}
	return repopath
}
