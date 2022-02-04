package bobgit

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/Benchkram/bob/pkg/bobutil"
	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
	git "github.com/go-git/go-git/v5"
	"github.com/logrusorgru/aurora"
)

var ErrEmptyCommitMessage = fmt.Errorf("aaaaaa")

var CleanWorkingDirMessage = "nothing to commit, working trees are clean."

func UntrackedRepoMessage(repolist []string) string {
	formattedRepos := []string{}

	for _, repo := range repolist {
		formattedRepos = append(formattedRepos, formatRepoNameForOutput(repo))
	}

	return fmt.Sprint("nothing to commit but untracked files present in repositories [", strings.Join(formattedRepos, ", "), "]")
}

// Commit executes `git commit -m ${message}` in all repositories.
//
// indifferent of the subdirectories and subrepositories,
// it walks through all the repositories starting from bobroot
// and run `git commit -m {message}` command.
//
// Only returns user messages in case of nothing to commit.
func Commit(message string) (s string, err error) {
	defer errz.Recover(&err)

	if message == "" {
		return "", ErrEmptyCommitMessage
	}

	bobRoot, err := bobutil.FindBobRoot()
	errz.Fatal(err)

	err = os.Chdir(bobRoot)
	errz.Fatal(err)

	// Assure toplevel is a git repo
	isGit, err := isGitRepo(bobRoot)
	errz.Fatal(err)
	if !isGit {
		return "", usererror.Wrap(ErrCouldNotFindGitDir)
	}

	// search for git repos inside bobRoot/.
	repoNames, err := findRepos(bobRoot)
	errz.Fatal(err)

	// repos with no changes, throws exit status 1 error
	// while executing `git commit --dry-run`.
	//
	// only repos with changes filtered out
	filteredRepo, untrackedRepo, err := filterModifiedRepos(repoNames)
	errz.Fatal(err)

	// throw some user message for untracked repositories
	if len(filteredRepo) == 0 {
		s := CleanWorkingDirMessage
		if len(untrackedRepo) > 0 {
			s = UntrackedRepoMessage(untrackedRepo)
		}
		return s, nil
	}

	maxRepoLen := 0
	// execute dry-run first on all repositories
	// to check for errors
	for _, name := range filteredRepo {
		_, err := cmdutil.GitDryCommit(name, message)

		// set maximum repository name length for indentation
		if len(name) > maxRepoLen {
			maxRepoLen = len(name)
		}

		if err != nil {
			return "", usererror.Wrapm(err, "Failed to commit to git in repo \""+name+"\"")
		}
	}

	for _, name := range filteredRepo {
		output, err := cmdutil.GitCommit(name, message)
		buf := FprintCommitOutput(name, output, maxRepoLen, err == nil)
		// instead of returnting the git output prints it here
		if len(output) > 0 {
			fmt.Println(buf.String())
		}
	}

	return "", nil
}

// filterModifiedRepos filters the repositories with changes
// by running `git status --porcelain` command on each repository and
// look for tracked files in staging.
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

		var tracked bool = false
		var untracked bool = false
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

// FprintCommitOutput formats output buffer from every repository commit output
func FprintCommitOutput(reponame string, output []byte, maxlen int, success bool) *bytes.Buffer {
	buf := bytes.NewBuffer(nil)

	// format the reponame for output
	repopath := reponame
	if reponame == "." {
		repopath = "/"
	} else if repopath[len(repopath)-1:] != "/" {
		repopath = repopath + "/"
	}

	spacing := "%-" + fmt.Sprint(maxlen) + "s"
	repopath = fmt.Sprintf(spacing, repopath)
	title := fmt.Sprint(repopath, "\t", aurora.Green("success"))
	if !success {
		title = fmt.Sprint(repopath, "\t", aurora.Red("failed!!"))
	}
	fmt.Fprint(buf, title)
	fmt.Fprintln(buf)

	if len(output) > 0 {
		for _, line := range outputLines(output) {
			modified := fmt.Sprint("  ", aurora.Gray(12, line))
			fmt.Fprintln(buf, modified)
		}
	}
	return buf
}

func outputLines(output []byte) []string {
	lines := strings.TrimSuffix(string(output), "\n")
	return strings.Split(lines, "\n")
}
