package bob

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/Benchkram/bob/pkg/file"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"

	"github.com/logrusorgru/aurora"
)

var ErrNoValidURLToClone = fmt.Errorf("No valid URL to clone found.")

// Clone repos which are not yet in the workspace.
// Uses priority urls ssh >> https >> file.
func (b *B) Clone() (err error) {
	defer errz.Recover(&err)

	for _, repo := range b.Repositories {

		prioritylist, err := makeURLPriorityList(repo)
		errz.Fatal(err)

		// Check if repository is already checked out.
		if file.Exists(repo.Name) {
			fmt.Printf("%s\n", aurora.Yellow(fmt.Sprintf("Skipping %s as the directory `%s` already exists", repo.Name, repo.Name)))
			continue
		}

		// returns user error if no possible url found
		if len(prioritylist) == 0 {
			return usererror.Wrapm(ErrNoValidURLToClone, "Failed to clone repository")
		}

		var out []byte
		// starts cloning from the first item of the priority list,
		// break for successfull cloning and continue in case of failure
		for _, url := range prioritylist {
			out, err = cmdutil.RunGitWithOutput(b.dir, "clone", url, "--progress")
			// if err != nil keep iterating through the url lists
			if err == nil {
				break
			}
		}
		// log the last order if has err and block the execution
		errz.Fatal(err)

		if len(out) > 0 {
			buf := FprintCloneOutput(repo.Name, out, err == nil)
			fmt.Println(buf.String())
		}

		err = b.gitignoreAdd(repo.Name)
		errz.Fatal(err)
	}

	return b.write()
}

func (b *B) CloneRepo(repoURL string) (_ string, err error) {
	defer errz.Recover(&err)

	out, err := cmdutil.RunGitWithOutput(b.dir, "clone", repoURL, "--progress")
	errz.Fatal(err)

	buf := FprintCloneOutput(".", out, err == nil)
	fmt.Println(buf.String())

	repo, err := Parse(repoURL)
	errz.Fatal(err)

	// TODO: let repoName be handled by Parse().
	repoName := RepoName(repo.HTTPS.URL)

	absRepoPath, err := filepath.Abs(repoName)
	errz.Fatal(err)

	wd, err := os.Getwd()
	errz.Fatal(err)

	// change currenct directory to inside the repository
	err = os.Chdir(absRepoPath)
	errz.Fatal(err)

	// change revert back to current working directory
	defer func() { _ = os.Chdir(wd) }()

	bob, err := Bob(
		WithDir(absRepoPath),
		WithRequireBobConfig(),
	)
	errz.Fatal(err)

	if err := bob.Clone(); err != nil {
		if err := os.RemoveAll(absRepoPath); err != nil {
			log.Printf("failed to remove cloned repo: %v\n", err)
		}
		return "", err
	}

	return repoName, nil
}

// makeURLPriorityList returns list of urls from forwarded repo,
// sorted by the priority, ssh >> https >> file.
//
// It ignores ssh/http if any of them set to ""
//
// It als Checks if it is a valid git repo,
// as someone might changed it on disk.
func makeURLPriorityList(repo Repo) ([]string, error) {
	var ignorehttp bool = false
	var ignoressh bool = false

	var urls []string

	if repo.SSHUrl == "" {
		ignoressh = true
	}

	if repo.HTTPSUrl == "" {
		ignorehttp = true
	}

	if !ignoressh {
		repoFromSSH, err := Parse(repo.HTTPSUrl)
		if err != nil {
			return nil, err
		}
		urls = append(urls, repoFromSSH.SSH.String())
	}

	if !ignorehttp {
		repoFromHTTPS, err := Parse(repo.HTTPSUrl)
		if err != nil {
			return nil, err
		}
		urls = append(urls, repoFromHTTPS.HTTPS.String())
	}

	if repo.LocalUrl != "" {
		urls = append(urls, repo.LocalUrl)
	}

	return urls, nil
}

// FprintCloneOutput returns formatted output buffer with repository title
// from git clone output.
func FprintCloneOutput(reponame string, output []byte, success bool) *bytes.Buffer {
	buf := FprintRepoTitle(reponame, 20, success)
	if len(output) > 0 {
		for _, line := range ConvertToLines(output) {
			modified := fmt.Sprint(aurora.Gray(12, line))
			if !success {
				modified = fmt.Sprint(aurora.Red(line))
			}
			fmt.Fprintln(buf, modified)
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

// ConvertToLines converts bytes into a list of strings separeted by newline
func ConvertToLines(output []byte) []string {
	lines := strings.TrimSuffix(string(output), "\n")
	return strings.Split(lines, "\n")
}
