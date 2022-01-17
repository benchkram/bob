package pathspec

import (
	"fmt"
	"strings"

	"github.com/Benchkram/bob/pkg/filepathutil"
)

// map the repo as key and pathspec as value
// where repository is relative to the bobroot
// and pathspec is relative to the repository.
type RepoPathspecMap map[string]string

var ErrRepoNotFound = fmt.Errorf("Repository name not found in target path repository list")

// P, stores the pathspec attribute for git
// intended to be used for the multi repo case.
type P struct {
	path          string
	possibleRepos RepoPathspecMap
}

func New(path string) *P {
	at := &P{
		path:          path,
		possibleRepos: ComputePossibleRepos(path),
	}

	return at
}

// SelectReposByPath filter the repos from all the repository list
// by the pathspec where git add command will be executed.
// Select all repos in case of target "." and set pathspecc
// to "." for all repos.
func (p *P) SelectReposByPath(repolist []string) []string {
	// return all the possible repos in case of
	// all `.`
	if p.path == "." || p.path == "./" {
		for _, repo := range repolist {
			if repo != "." {
				p.possibleRepos[repo] = "."
			}
		}
		return repolist
	}

	filterd := []string{}
	for _, repo := range repolist {
		for r := range p.possibleRepos {
			if repo == r {
				filterd = append(filterd, repo)
			}
		}
	}

	// get filtered repos from the target path
	filterd = removeParentRepos(filterd)

	// add all possible repositories that situated
	// in the target directory if path ends with `.`
	if strings.Contains(p.path, "/.") {
		for _, repo := range repolist {
			if !contains(filterd, repo) {
				temptarget := strings.Trim(p.path, ".")
				if strings.HasPrefix(repo, temptarget) {
					filterd = append(filterd, repo)
					p.possibleRepos[repo] = "."
				}
			}
		}
	}

	return filterd
}

// GetRelativePathspec returns the relative pathspec from the internal map.
func (p *P) GetRelativePathspec(reponame string) (string, error) {
	if val, ok := p.possibleRepos[reponame]; ok {
		return val, nil
	}
	return "", ErrRepoNotFound
}

// ComputePossibleRepos Compute all the possible repository path
// from the provided path starting from bobroot inside bob workspace
// and returns a map of string  where key is every repository and value is
// the relative path from that repository.
// repositories can be filtered later from the computed repository paths.
// e.g: 'bobroot/sample/path' computes items like ".": 'bobroot/sample/path',
// "bobroot": 'sample/path', "bobroot/sample": 'path' ..
// can be interpreted this way, if the selected repository path is `bobroot/sample`,
// then pathspec for that target path would be only `path`, and so on.
func ComputePossibleRepos(path string) RepoPathspecMap {

	possibleRepos := make(RepoPathspecMap)
	possibleRepos["."] = path

	if path == "." {
		return possibleRepos
	}

	splitted := strings.Split(path, "/")

	for i := 0; i < len(splitted)-1; i++ {
		repo := strings.Join(splitted[:i+1], "/")
		target := strings.Join(splitted[i+1:], "/")
		possibleRepos[repo] = target
	}
	possibleRepos[strings.Join(splitted, "/")] = "."

	return possibleRepos
}

// removeParentRepos removes the parent repository path from each
// repo in `repolist`.  Keeping only the child repository path.
func removeParentRepos(repolist []string) []string {
	filtered := []string{}
	for _, repo := range repolist {
		hasChild := false
		for _, r := range repolist {
			if r != repo {
				hasChild = filepathutil.IsChild(repo, r)
			}
		}
		if !hasChild {
			filtered = append(filtered, repo)
		}
	}
	return filtered
}

// contains returns true when e is contained in s
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
