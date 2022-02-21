package bob

import (
	"bufio"
	"bytes"
	"fmt"
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

// cloneURLItem used to map URL item from the `makeURLPriorityList`
// with it's protocol for logging purpuse
type cloneURLItem struct {
	url      string
	protocol string
}

// Clone repos which are not yet in the workspace.
// Uses priority urls ssh >> https >> file.
//
// failFast will not prompt the user in case of an error.
//
// TODO: it still happens that git prompts for user input
// in case of a missing password on https.
func (b *B) Clone(failFast bool) (err error) {
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
		// Starts cloning from the first item of the priority list,
		// break for successfull cloning and fallback to next item in
		// the map in case of failure
		for i, item := range prioritylist {
			out, err = cmdutil.RunGitWithOutput(b.dir, "clone", item.url, "--progress")
			if err == nil {
				break
			}

			fmt.Println(err.Error())
			err = nil

			// fail early, useful when used on ci.
			if failFast {
				return usererror.Wrap(fmt.Errorf("abort"))
			}

			// Get user feedback in case of a failure before trying the
			// next clone method.
			fmt.Printf("%s\n", aurora.Yellow(fmt.Sprintf("Failed to clone %s using %s", repo.Name, item.protocol)))
			if i < len(prioritylist)-1 {

				wd, _ := os.Getwd()
				target := filepath.Join(b.dir, repo.Name)
				target, _ = filepath.Rel(wd, target)

				fmt.Printf("[%s] is likely in an invalid state. Want to delete [%s] and clone using [%s]: (y/(a)bort/(i)ignore) ",
					aurora.Bold(target),
					aurora.Bold(target),
					aurora.Bold(prioritylist[i+1].protocol),
				)
				reader := bufio.NewReader(os.Stdin)
				text, _ := reader.ReadString('\n')
				text = strings.Replace(text, "\n", "", -1)
				text = strings.ToLower(text)

				if text == "y" {
					fmt.Println()
					if file.Exists(target) {
						_ = os.RemoveAll(target)
					}
				} else if text == "i" {
					fmt.Printf("ignoring %s\n\n", target)
					// Clear output as it's already been printed.
					out = []byte{}
					break
				} else {
					return usererror.Wrap(fmt.Errorf("abort"))
				}
			}
		}

		if len(out) > 0 {
			buf := FprintCloneOutput(repo.Name, out, err == nil)
			fmt.Println(buf.String())
		}

		err = b.gitignoreAdd(repo.Name)
		errz.Fatal(err)
	}

	return b.write()
}

// CloneRepo repo and sub repositories recursively.
// failFast will not prompt the user in case of an error.
func (b *B) CloneRepo(repoURL string, failFast bool) (_ string, err error) {
	defer errz.Recover(&err)

	out, err := cmdutil.RunGitWithOutput(b.dir, "clone", repoURL, "--progress")
	errz.Fatal(err)

	buf := FprintCloneOutput(".", out, err == nil)
	fmt.Println(buf.String())

	repo, err := Parse(repoURL)
	errz.Fatal(err)

	absRepoPath, err := filepath.Abs(repo.Name())
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

	if err := bob.Clone(failFast); err != nil {
		return "", err
	}

	return repo.Name(), nil
}

// makeURLPriorityList returns list of cloneURLItem from forwarded repo,
// ordered by the priority type, ssh >> https >> file.
//
// It ignores ssh/http if any of them set to ""
//
// It als checks if it is a valid git repo,
// as someone might changed it on disk.
func makeURLPriorityList(repo Repo) ([]cloneURLItem, error) {

	var urls []cloneURLItem

	if repo.SSHUrl != "" {
		repoFromSSH, err := Parse(repo.SSHUrl)
		if err != nil {
			return nil, err
		}
		urls = append(urls, cloneURLItem{
			url:      repoFromSSH.SSH.String(),
			protocol: "ssh",
		})
	}

	if repo.HTTPSUrl != "" {
		repoFromHTTPS, err := Parse(repo.HTTPSUrl)
		if err != nil {
			return nil, err
		}
		urls = append(urls, cloneURLItem{
			url:      repoFromHTTPS.HTTPS.String(),
			protocol: "https",
		})
	}

	if repo.LocalUrl != "" {
		urls = append(urls, cloneURLItem{
			url:      repo.LocalUrl,
			protocol: "local",
		})
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
	repopath := fmt.Sprintf(spacing, sanitizeReponame(reponame))
	title := fmt.Sprint(repopath, "\t", aurora.Green("success"))
	if !success {
		title = fmt.Sprint(repopath, "\t", aurora.Red("error"))
	}
	fmt.Fprint(buf, title)
	fmt.Fprintln(buf)

	return buf
}

// sanitizeReponame returns sanitized reponame.
//
// Example: "." => "/", "second-level" => "second-level/"
func sanitizeReponame(reponame string) string {
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
