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
	"github.com/Benchkram/errz"

	"github.com/logrusorgru/aurora"
)

// Clone repos which are not yet in the workspace.
// Uses ssh with a fallback on https.
func (b *B) Clone() (err error) {
	defer errz.Recover(&err)

	for _, repo := range b.Repositories {

		// Check if it is a valid git repo,
		// as someone might changed it on disk.
		repoFromHTTPS, err := Parse(repo.HTTPSUrl)
		errz.Fatal(err)
		repoFromSSH, err := Parse(repo.SSHUrl)
		errz.Fatal(err)
		localUrl := repo.LocalUrl

		// Check if repository is already checked out.
		if file.Exists(repo.Name) {
			fmt.Printf("%s\n", aurora.Yellow(fmt.Sprintf("Skipping %s as the directory `%s` already exists", repo.Name, repo.Name)))
			continue
		}

		var output *[]byte
		// Try to clone from ssh, if it fails try https
		out, err := cmdutil.RunGitWithOutput(b.dir, "clone", repoFromSSH.SSH.String(), "--progress")
		output = &out
		if err != nil {
			fmt.Printf("%s\n", aurora.Yellow(fmt.Sprintf("Failed to clone %s using ssh", repo.Name)))

			// Let's try https
			out, err := cmdutil.RunGitWithOutput(b.dir, "clone", repoFromHTTPS.HTTPS.String(), "--progress")
			output = &out
			if err != nil {
				fmt.Printf("%s\n", aurora.Yellow(fmt.Sprintf("Failed to clone %s using https", repo.Name)))

				out, err := cmdutil.RunGitWithOutput(b.dir, "clone", localUrl, "--progress")
				output = &out
				errz.Fatal(err)
			}
		}

		if output != nil && len(*output) > 0 {
			buf := FprintCloneOutput(repo.Name, *output, err == nil)
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
