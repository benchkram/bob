package bobgit

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"

	"github.com/Benchkram/bob/bobgit/status"
	"github.com/Benchkram/bob/pkg/bobutil"
	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/Benchkram/errz"
)

var ErrCouldNotFindGitDir = fmt.Errorf("Could not find a .git folder")

var dontFollow = []string{
	"node_modules",
	".git",
	".vscode",
}

// Status executes `git status -porcelain` in all repositories
// first level repositories found inside a .bob filtree.
// It parses the output of each call and creates a object
// containing status infos for all of them combined.
// The result is similar to what `git status` would print
// but visualy optimised for the multi repository case.
func Status() (s *status.S, err error) {
	defer errz.Recover(&err)

	bobRoot, err := bobutil.FindBobRoot()
	errz.Fatal(err)

	depth, err := wdDepth(bobRoot)
	errz.Fatal(err)

	err = os.Chdir(bobRoot)
	errz.Fatal(err)

	// Assure toplevel is a git repo
	isGit, err := isGitRepo(bobRoot)
	errz.Fatal(err)
	if !isGit {
		return nil, ErrCouldNotFindGitDir
	}

	// search for git repos inside bobRoot/.
	repoNames := []string{}
	err = filepath.WalkDir(bobRoot, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return nil
		}

		if d.Name() == ".git" {
			p, err := filepath.Rel(bobRoot, filepath.Dir(path))
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
	errz.Fatal(err)

	s = status.New()
	for _, name := range repoNames {

		prefix := strings.Repeat("../", depth)

		// repoPath is the path of the repo
		// relative to the top bob repo.
		repoPath := prefix + name
		if name == "." {
			repoPath = strings.TrimSuffix(repoPath, ".")
		}
		s.AddRepo(repoPath)

		output, err := cmdutil.GitStatus(name)
		errz.Fatal(err)

		status, err := parse(output)
		errz.Fatal(err)

		// localpath is the path as given by `git status`
		// in the respecting repo.
		//
		// TODO: compare and adapt with https://git-scm.com/docs/git-status#_short_format
		for localpath, status := range status {
			if status.Staging == git.Unmodified && status.Worktree == git.Unmodified {
				continue
			}

			// Conflicts
			// skip other checks if conflict happens for a file
			if status.Staging == git.UpdatedButUnmerged || status.Worktree == git.UpdatedButUnmerged {
				s.Conflicts[repoPath][localpath] = status
				continue
			}

			// if deleted in both, add to conflicts and skip others
			if status.Staging == git.Deleted && status.Worktree == git.Deleted {
				s.Conflicts[repoPath][localpath] = status
				continue
			}

			// Staging aka index
			if status.Staging == git.Renamed ||
				status.Staging == git.Added ||
				status.Staging == git.Deleted ||
				status.Staging == git.Copied ||
				status.Staging == git.Modified {
				s.Staging[repoPath][localpath] = status
			}

			// Unstaged aka worktree
			if status.Worktree == git.Modified ||
				status.Worktree == git.Deleted {
				s.Unstaged[repoPath][localpath] = status
			}

			// Untracked
			if status.Worktree == git.Untracked && status.Staging == git.Untracked {
				s.Untracked[repoPath][localpath] = status
			}
		}
	}

	s.Repos = append(s.Repos, repoNames...)

	return s, nil
}

// parse `git status --porcelaine=v1` output
// see https://git-scm.com/docs/git-status
func parse(buf []byte) (status git.Status, err error) {
	status = make(git.Status)

	scanner := bufio.NewScanner(bytes.NewBuffer(buf))
	for scanner.Scan() {
		line := scanner.Text()
		fileStatus := &git.FileStatus{}

		fileStatus.Staging = readX(line)  // aka index
		fileStatus.Worktree = readY(line) // worktree

		path := line[3:]
		if fileStatus.Staging == git.Renamed || fileStatus.Worktree == git.Renamed {
			parts := strings.Split(path, " ")
			fileStatus.Extra = parts[0] // name previous to rename
			// parts[1] // ->
			path = parts[2] // new name
		}

		status[path] = fileStatus
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return status, nil
}

func readX(line string) git.StatusCode {
	return git.StatusCode(line[0])
}

func readY(line string) git.StatusCode {
	return git.StatusCode(line[1])
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

// wdDepth returns the number of `../` traversals
// till reaching dir.
func wdDepth(dir string) (depth int, err error) {
	defer errz.Recover(&err)

	wd, err := os.Getwd()
	errz.Fatal(err)

	dir, err = filepath.Abs(dir)
	errz.Fatal(err)

	wdparts := len(strings.Split(wd, "/"))
	dirparts := len(strings.Split(dir, "/"))

	depth = wdparts - dirparts
	if depth < 0 {
		return 0, fmt.Errorf("wdDepth got a negative result")
	}

	return depth, nil
}
