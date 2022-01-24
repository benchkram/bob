package bobgit

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Benchkram/bob/pkg/bobutil"
	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
	"github.com/cli/cli/git"
)

var ErrInsufficientConfig = fmt.Errorf("Insufficient Configeration.")
var ErrUptodateAllRepo = fmt.Errorf("All repositories up to date.")

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
	if errors.Is(err, ErrInsufficientConfig) {
		return usererror.Wrapm(ErrInsufficientConfig, "Git push failed")
	}
	errz.Fatal(err)

	if len(filteredRepo) == 0 {
		return usererror.Wrapm(ErrUptodateAllRepo, "Nothing to push")
	}

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
		} else {
			fmt.Println("WARNING!! No configured push destination for repository \"" + formatRepoNameForOutput(repo) + "\"")
			fmt.Printf("Are you sure want to continue with the rest of the repositories? (yes/no): ")
			resp := askForConfirmation()
			if !resp {
				return nil, ErrInsufficientConfig
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

// askForConfirmation uses Scanln to parse user input. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user. Typically, you should use fmt to print out a question
// before calling askForConfirmation. E.g. fmt.Println("WARNING: Are you sure? (yes/no)")
func askForConfirmation() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	response = strings.ToLower(response)

	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return askForConfirmation()
	}
}

// containsString returns true if slice contains element
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

// You might want to put the following two functions in a separate utility package.

// posString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}
