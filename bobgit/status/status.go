package status

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/logrusorgru/aurora"
)

type MultiRepoStatus map[string]git.Status

type S struct {
	Staging   MultiRepoStatus
	Unstaged  MultiRepoStatus
	Untracked MultiRepoStatus
	Conflicts MultiRepoStatus

	Repos []string
}

func New() *S {
	s := &S{
		Staging:   make(MultiRepoStatus),
		Unstaged:  make(MultiRepoStatus),
		Untracked: make(MultiRepoStatus),
		Conflicts: make(MultiRepoStatus),
	}
	return s
}

// AddRepo iniitializes the state maps with a new repo
func (s *S) AddRepo(repoName string) {
	s.Staging[repoName] = make(git.Status)
	s.Unstaged[repoName] = make(git.Status)
	s.Untracked[repoName] = make(git.Status)
	s.Conflicts[repoName] = make(git.Status)
}

func (s *S) String() string {
	const spacing = "        "

	buf := bytes.NewBuffer(nil)

	{
		b := bytes.NewBuffer(nil)
		keys := sortedKeys(s.Conflicts)
		var conflictingRepos []string
		for _, repoName := range keys {
			repoStatus := s.Conflicts[repoName]
			for path, status := range repoStatus {
				fmt.Fprint(b, spacing)
				fprintChanges(b, repoName, path, status, aurora.Red)
				conflictingRepos = append(conflictingRepos, repoName)
			}
		}

		if len(conflictingRepos) > 0 {
			fmt.Fprintln(buf, "On at least one repository")
			fmt.Fprintln(buf, "You have unmerged paths.")
			fmt.Fprintln(buf, "  (fix conflicts using plain git)")
			fmt.Fprintln(buf, "\nUnmerged paths:")
			fmt.Fprint(buf, b)
		}
	}

	{
		fmt.Fprintln(buf, "Changes to be committed:")

		// keys := sortedKeys(s.Staging)
		// for _, repoName := range keys {
		// 	repoStatus := s.Staging[repoName]
		// 	for path, status := range repoStatus {
		// 		fmt.Fprint(buf, spacing)
		// 		//fmt.Fprintf(buf, "s%cw%c", status.Staging, status.Worktree) // debugging
		// 		//fmt.Fprintf(buf, " (%s) ", aurora.Green(repoName))
		// 		fprintChanges(buf, repoName, path, status, aurora.Green)
		// 	}
		// }
		FPrintMultirepoStatus(buf, spacing, s.Staging, aurora.Green)
		fmt.Fprintln(buf)
	}

	{
		fmt.Fprintln(buf, "Changes not staged for commit:")

		// keys := sortedKeys(s.Unstaged)
		// for _, repoName := range keys {
		// 	repoStatus := s.Unstaged[repoName]
		// 	for path, status := range repoStatus {
		// 		fmt.Fprint(buf, spacing)
		// 		//fmt.Fprintf(buf, "s%cw%c", status.Staging, status.Worktree) // debugging
		// 		//fmt.Fprintf(buf, " (%s) ", aurora.Red(repoName))
		// 		fprintChanges(buf, repoName, path, status, aurora.Red)
		// 	}
		// }
		FPrintMultirepoStatus(buf, spacing, s.Unstaged, aurora.Red)
		fmt.Fprintln(buf)
	}

	{
		fmt.Fprintln(buf, "Untracked files:")

		keys := sortedKeys(s.Untracked)
		for _, repoName := range keys {
			repoStatus := s.Untracked[repoName]
			repoStatusKeys := sortedKeysStatus(repoStatus)

			for _, path := range repoStatusKeys {
				fmt.Fprint(buf, spacing)
				fprintChanges(buf, repoName, path, &git.FileStatus{}, aurora.Red)
			}
		}
	}

	return buf.String()
}

func FPrintMultirepoStatus(buf *bytes.Buffer, spacing string, repos MultiRepoStatus, color func(interface{}) aurora.Value) {
	keys := sortedKeys(repos)
	for _, repoName := range keys {
		repoStatus := repos[repoName]
		repoStatusKeys := sortedKeysStatus(repoStatus)
		for _, path := range repoStatusKeys {
			fmt.Fprint(buf, spacing)
			fprintChanges(buf, repoName, path, repoStatus[path], color)
		}
	}
}

// fprintChanges helper to color & highlight the output based on status
func fprintChanges(
	buf *bytes.Buffer,
	repoPath, localPath string,
	status *git.FileStatus,
	color func(interface{}) aurora.Value,
) {
	if repoPath == "." {
		repoPath = ""
	}

	dir, basename := splitDirAndBasename(repoPath)
	// fmt.Printf("prefix: [%s] path: [%s]\n", dir, basename)

	// check for the conflicts first
	if status.Staging == git.UpdatedButUnmerged || status.Worktree == git.UpdatedButUnmerged {
		conflictText := getConflictText(status)
		fmt.Fprint(buf, color(conflictText+aurora.Bold(withSlash(basename)).String()).String())
		fmt.Fprintln(buf, color(localPath).String())
	} else if status.Staging == git.Renamed {
		//fmt.Fprint(buf, color("renamed:   "+aurora.Bold(withSlash(repoName)).String()).String())
		fmt.Fprint(buf, color("renamed:    "))
		fmt.Fprint(buf, color(dir).String())
		fmt.Fprint(buf, color(aurora.Bold(withSlash(basename))).String())
		fmt.Fprint(buf, color(status.Extra).String())
		fmt.Fprint(buf, color(" -> ").String())
		fmt.Fprint(buf, color(dir).String())
		fmt.Fprint(buf, color(aurora.Bold(withSlash(basename))).String())
		fmt.Fprintln(buf, color(localPath).String())
	} else if status.Staging == git.Modified || status.Worktree == git.Modified {
		fmt.Fprint(buf, color("modified:   "+dir+aurora.Bold(withSlash(basename)).String()).String())
		fmt.Fprintln(buf, color(localPath).String())
	} else if status.Staging == git.Added || status.Worktree == git.Added {
		fmt.Fprint(buf, color("new file:   "+dir+aurora.Bold(withSlash(basename)).String()).String())
		fmt.Fprintln(buf, color(localPath).String())
	} else if status.Staging == git.Deleted || status.Worktree == git.Deleted {
		fmt.Fprint(buf, color("deleted:    "+dir+aurora.Bold(withSlash(basename)).String()).String())
		fmt.Fprintln(buf, color(localPath).String())
	} else {
		fmt.Fprint(buf, color(dir).String())
		fmt.Fprint(buf, color(aurora.Bold(withSlash(basename)).String()).String())
		fmt.Fprintln(buf, color(localPath).String())
	}
}

// sortedKeys returns the keys of `MultiRepoStatus` sorted by `sort.Strings`
func sortedKeys(m MultiRepoStatus) (keys []string) {
	keys = make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeysStatus(m git.Status) (keys []string) {
	keys = make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func withSlash(s string) string {
	if strings.HasSuffix(s, "/") || s == "" {
		return s
	}
	return s + "/"
}

// split repoPath in dir and basename
// with some additional sanitising.
func splitDirAndBasename(path string) (prefix, name string) {
	if path != "" {
		prefix = filepath.Dir(path) + "/"
		if prefix == "./" {
			prefix = ""
		}

		name = filepath.Base(path)
		if name == ".." {
			name = ""
		}
	}
	return prefix, name
}

// return the merge conflict text for the file
// depending on the conflicting status
// on merge, delete, etc
func getConflictText(status *git.FileStatus) string {
	conflictText := "both modified: \t"

	// fmt.Println(string(status.Staging) + " " + string(status.Worktree))

	if status.Worktree == git.UpdatedButUnmerged && status.Staging == git.Deleted {
		conflictText = "deleted by us: \t"
	}
	if status.Staging == git.UpdatedButUnmerged && status.Worktree == git.Deleted {
		conflictText = "deleted by them: \t"
	}

	return conflictText
}
