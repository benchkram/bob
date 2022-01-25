package bobgit

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Benchkram/bob/pkg/bobutil"
	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/Benchkram/bob/pkg/strutil"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
	"github.com/cli/cli/git"
	"github.com/logrusorgru/aurora"
)

var ErrInsufficientConfig = fmt.Errorf("Repository Not configured Properly.")
var ErrUptodateAllRepo = fmt.Errorf("All repositories up to date.")

const configureInstruction string = "Either specify the URL from the command-line or configure a remote " +
	"repository and then push using the remote name."

// Push run `git push` commands iteratively
// in all the git repositories under bob workspace.
//
// Run through all the repositories with a confirm dialog in case of
// not configured remote.
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

	// pre-compute maximum repository name length for proper formatting
	maxlen := strutil.LongestStrLen(repoNames)

	filteredRepo, err := filterReadyToPushRepos(bobRoot, repoNames, maxlen)
	if errors.Is(err, ErrInsufficientConfig) {
		return usererror.Wrapm(ErrInsufficientConfig, "Git push failed")
	}
	errz.Fatal(err)

	if len(filteredRepo) == 0 {
		return usererror.Wrapm(ErrUptodateAllRepo, "Nothing to push")
	}

	// run git push --dry-run first
	// if error happens rejects the whole command with error message
	for _, repo := range filteredRepo {
		conf, err := getRepoConfig(bobRoot, repo)
		errz.Fatal(err)

		output, err := cmdutil.GitPushDry(repo, conf.RemoteName, conf.MergeRef)
		if err != nil {
			buf := FprintPushOutput(repo, output, maxlen, true)
			fmt.Println(buf.String())
		}
		errz.Fatal(err)
	}

	// run git push finally
	for _, repo := range filteredRepo {
		conf, err := getRepoConfig(bobRoot, repo)
		errz.Fatal(err)

		output, err := cmdutil.GitPush(repo, conf.RemoteName, conf.MergeRef)

		buf := FprintPushOutput(repo, output, maxlen, err == nil)
		fmt.Println(buf.String())
	}

	return nil
}

// filterReadyToPushRepos returns repositories with already preapred commits and ready to push.
//
// ask for confirmation in case of not configured remote. If confirmed with `no` rejects the whole
// command.
func filterReadyToPushRepos(root string, repolist []string, maxlen int) ([]string, error) {
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
			buf := FprintErrorPushDestination(repo, maxlen)
			fmt.Println(buf.String())
			resp := askForConfirmation("Sure want to continue with the rest of the repositories? (yes/no): ")
			if !resp {
				return nil, ErrInsufficientConfig
			}
			fmt.Println()
		}
	}

	return filtered, nil
}

// FprintErrorPushDestination created output buffer for not configured respository error
func FprintErrorPushDestination(reponame string, maxlen int) *bytes.Buffer {
	buf := FprintRepoTitle(reponame, maxlen, false)

	line1 := fmt.Sprintln("  ", aurora.Red("No configured push destination"))
	line2 := fmt.Sprintln("  ", aurora.Red(configureInstruction))
	fmt.Fprint(buf, line1, line2)
	return buf
}

// FprintPushOutput returns formatted output buffer with repository title
// from git push output.
func FprintPushOutput(reponame string, output []byte, maxlen int, success bool) *bytes.Buffer {
	buf := FprintRepoTitle(reponame, maxlen, success)

	if len(output) > 0 {
		for _, line := range strutil.ConvertToLines(output) {
			modified := fmt.Sprint(aurora.Gray(12, line))
			if !success {
				modified = fmt.Sprint(aurora.Red(line))
			}
			fmt.Fprintln(buf, "  ", modified)
		}
	}

	return buf
}

// FprintRepoTitle returns repo title buffer with success/error label
func FprintRepoTitle(reponame string, maxlen int, success bool) *bytes.Buffer {
	buf := bytes.NewBuffer(nil)
	spacing := "%-" + fmt.Sprint(maxlen) + "s"
	repopath := fmt.Sprintf(spacing, formatRepoNameForOutput(reponame))
	title := fmt.Sprint(repopath, "\t", aurora.Green("success"))
	if !success {
		title = fmt.Sprint(repopath, "\t", aurora.Red("error"))
	}
	fmt.Fprint(buf, title)
	fmt.Fprintln(buf)

	return buf
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
		for _, line := range strutil.ConvertToLines(output) {
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

// askForConfirmation uses Scanln to parse user input. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again.
//
// The function does not return until it gets a valid response from the user.
// prints the confirmation message before asking for the confirmation.
func askForConfirmation(confirmationMessage string) bool {

	fmt.Print(confirmationMessage)

	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	response = strings.ToLower(response)

	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if strutil.Contains(okayResponses, response) {
		return true
	} else if strutil.Contains(nokayResponses, response) {
		return false
	} else {
		return askForConfirmation("Please type yes or no and then press enter:")
	}
}
