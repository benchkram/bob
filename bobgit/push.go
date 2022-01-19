package bobgit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Benchkram/bob/pkg/bobutil"
	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/Benchkram/errz"
	"github.com/cli/cli/git"
)

// Push run `git push` commands iteratively
// in all the git repositories under bob workspace
func Push() (err error) {
	defer errz.Recover(&err)

	bobRoot, err := bobutil.FindBobRoot()
	errz.Fatal(err)

	err = os.Chdir(bobRoot)
	errz.Fatal(err)

	// Assure toplevel is a git repo
	isGit, err := isGitRepo(bobRoot)
	errz.Fatal(err)
	if !isGit {
		return ErrCouldNotFindGitDir
	}

	repoNames, err := findRepos(bobRoot)
	errz.Fatal(err)

	filteredRepo, err := filterReadyToCommitRepos(bobRoot, repoNames)
	errz.Fatal(err)

	for _, repo := range filteredRepo {
		conf, err := getRepoConfig(bobRoot, repo)
		errz.Fatal(err)

		output, err := cmdutil.GitPushDry(repo, conf.RemoteName, conf.MergeRef)
		errz.Fatal(err)

		fmt.Println(string(output))
	}

	return nil
}

func filterReadyToCommitRepos(root string, repolist []string) ([]string, error) {
	filtered := []string{}
	for _, repo := range repolist {
		repoConfig, err := getRepoConfig(root, repo)
		errz.Fatal(err)

		if repoConfig.RemoteName != "" {
			output, err := cmdutil.GitUnpushedCommits(repo)
			errz.Fatal(err)

			commits := parseCommitsOutput(output)
			if len(commits) > 0 {
				filtered = append(filtered, repo)
			}
		}
	}

	return filtered, nil
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

// parseCommitsOutput parses `git cherry -v` output,
// returns list of git Commits with sha and title
//
// Example output parameter format:
//
// + f4698ccf831a7bce3577e126404e4bc2b9641438 added files
//
// + 2a5017272931ae3a668fde3646c309315bb137fc new info txt
func parseCommitsOutput(output []byte) []*git.Commit {
	commits := []*git.Commit{}
	sha := 1
	title := 2
	if len(output) > 0 {
		for _, line := range outputLines(output) {
			split := strings.SplitN(line, " ", 3)
			if len(split) == 3 {
				commits = append(commits, &git.Commit{
					Sha:   split[sha],
					Title: split[title],
				})
			}
		}
	}

	return commits
}

func outputLines(output []byte) []string {
	lines := strings.TrimSuffix(string(output), "\n")
	return strings.Split(lines, "\n")
}
