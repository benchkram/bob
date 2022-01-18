package bobgit

import (
	"fmt"
	"os"
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

	for _, repo := range repoNames {
		// repoPath := filepath.Join(bobRoot, repo)

		// err = os.Chdir(repoPath)
		// errz.Fatal(err)

		// branch, err := git.CurrentBranch()
		// errz.Fatal(err)

		// fmt.Println(branch)

		// config := git.ReadBranchConfig(branch)
		// fmt.Println(config)

		// remote, err := git.Config("remote.origin.url")
		// if err != nil {
		// 	fmt.Println(repoPath)
		// }
		// errz.Fatal(err)
		// fmt.Println(remote)
		// if config.RemoteName != "" {
		output, err := cmdutil.GitUnpushedCommit(repo)
		if err != nil {
			fmt.Println(repo)
		}

		commits := parseCommitsOutput(output)
		fmt.Println(commits)
		// }

		// err = os.Chdir(bobRoot)
		// errz.Fatal(err)
	}

	return nil
}

func parseCommitsOutput(output []byte) []*git.Commit {
	commits := []*git.Commit{}
	sha := 0
	title := 1
	for _, line := range outputLines(output) {
		fmt.Println(line)
		split := strings.SplitN(line, " ", 2)
		if len(split) == 2 {
			fmt.Println(split[0])
			fmt.Println(split[1])
			commits = append(commits, &git.Commit{
				Sha:   split[sha],
				Title: split[title],
			})
		}
	}

	return commits
}

func outputLines(output []byte) []string {
	lines := strings.TrimSuffix(string(output), "\n")
	return strings.Split(lines, "\n")

}
