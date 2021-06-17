package bob

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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

		// Try to clone from ssh, if it fails try https
		err = cmdutil.RunGit(b.dir, "clone", repoFromSSH.SSH.String())
		if err != nil {
			fmt.Printf("%s\n", aurora.Yellow(fmt.Sprintf("Failed to clone %s using ssh", repo.Name)))

			// Let's try https
			err := cmdutil.RunGit(b.dir, "clone", repoFromHTTPS.HTTPS.String())
			if err != nil {
				fmt.Printf("%s\n", aurora.Yellow(fmt.Sprintf("Failed to clone %s using https", repo.Name)))

				err := cmdutil.RunGit(b.dir, "clone", localUrl)
				errz.Fatal(err)
			}
		}

		err = b.gitignoreAdd(repo.Name)
		errz.Fatal(err)
	}

	return b.write()
}

func (b *B) CloneRepo(repoURL string) (_ string, err error) {
	defer errz.Recover(&err)

	err = cmdutil.RunGit(b.dir, "clone", repoURL)
	errz.Fatal(err)

	repo, err := Parse(repoURL)
	errz.Fatal(err)
	// TODO: let repoName be handled by Parse().
	repoName := RepoName(repo.HTTPS.URL)

	absRepoPath, err := filepath.Abs(repoName)
	errz.Fatal(err)

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
