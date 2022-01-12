package pathspec

import (
	"fmt"
	"strings"

	"github.com/Benchkram/bob/pkg/filepathutil"
)

// map the pathspec and repo name
// where pathspec is relative to the repository
// and repository is relative to the bobroot
type RepoPathspecMap map[string]string

var ErrRepoNotFound = fmt.Errorf("Repository name not found in target path repository list")

// P, stores the pathspec attribute for git.
// also applies some computations and manipulation
// to the path attributes necessary for bob multirepo structure
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

// SelectReposByPath, filter the repos from all the repository list
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
			if !Contains(filterd, repo) {
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

// GetRelativePathspec, Returns the relative pathspec by the reponame key
// created from the provided path earlier, for a relative repository from
// the bob root.
func (p *P) GetRelativePathspec(reponame string) (string, error) {
	if val, ok := p.possibleRepos[reponame]; ok {
		return val, nil
	}

	return "", ErrRepoNotFound
}

// ComputePossibleRepos, Compute all the possible repository path from
// bob root and relative pathspec from a provided path inside bob workspace
// and returns a map of string  where key is every repository and value is
// the relative path from that repository.
// repositories can be filtered later from the computed repository paths
// e.g: 'bobroot/sample/path' computes items like ".": 'bobroot/sample/path',
// "bobroot": 'sample/path', "bobroot/sample": 'path' ..
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

// removeParentRepos, filters out the parent repositories from the
// selected repos and only keep the child repository
// to execute the git add command later
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

// Contains returns if a slice of string
// contains a specific string or not
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
