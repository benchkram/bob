package add

import (
	"fmt"
	"strings"

	"github.com/Benchkram/bob/pkg/filepathutil"
)

type RepoTargetMap map[string]string

var ErrRepoNotFound = fmt.Errorf("Repository name not found in target path repository list")

type Target struct {
	target        string
	possibleRepos RepoTargetMap
}

func NewTarget(target string) *Target {
	at := &Target{
		target:        target,
		possibleRepos: ComputePossibleRepos(target),
	}

	return at
}

// select the repos by the target path where git add command
// will be executed.
// Select all repos in case of target "." and set target path
// to "." for all repos.
func (at *Target) SelectReposByTarget(repolist []string) []string {
	if at.target == "." || at.target == "./" {
		for _, repo := range repolist {
			if repo != "." {
				at.possibleRepos[repo] = "."
			}
		}
		return repolist
	}

	filterd := []string{}
	for _, repo := range repolist {
		for r := range at.possibleRepos {
			if repo == r {
				filterd = append(filterd, repo)
			}
		}
	}

	filterd = removeParentRepos(filterd)
	return filterd
}

// Returns the relative target path created
// from the provided target for a repo
// inside the bob workspace
func (at *Target) GetRelativeTarget(reponame string) (string, error) {
	if val, ok := at.possibleRepos[reponame]; ok {
		return val, nil
	}

	return "", ErrRepoNotFound
}

// Compute all the possible repository path and relative target path
// from the provided target path inside bob workspace
// repositories can be filtered later from the computed repository paths
func ComputePossibleRepos(target string) RepoTargetMap {

	possibleRepos := make(RepoTargetMap)
	possibleRepos["."] = target

	if target == "." {
		return possibleRepos
	}

	splitted := strings.Split(target, "/")

	for i := 0; i < len(splitted)-1; i++ {
		repo := strings.Join(splitted[:i+1], "/")
		target := strings.Join(splitted[i+1:], "/")
		possibleRepos[repo] = target
	}
	possibleRepos[strings.Join(splitted, "/")] = "."

	return possibleRepos
}

// filter out the parent repositories from the
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
