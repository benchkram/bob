package add

import (
	"fmt"
	"strings"

	"github.com/Benchkram/bob/pkg/filepathutil"
)

type RepoTargetMap map[string]string

var ErrRepoNotFound = fmt.Errorf("Repository name not found in target path repository list")

type AddTarget struct {
	target        string
	possibleRepos RepoTargetMap
}

func NewTarget(target string) *AddTarget {
	at := &AddTarget{
		target:        target,
		possibleRepos: ComputePossibleRepos(target),
	}

	return at
}

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

func (at *AddTarget) PopulateAndFilterRepos(repolist []string) []string {
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

func (at *AddTarget) GetRelativeTarget(reponame string) (string, error) {
	if val, ok := at.possibleRepos[reponame]; ok {
		return val, nil
	}

	return "", ErrRepoNotFound
}

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
