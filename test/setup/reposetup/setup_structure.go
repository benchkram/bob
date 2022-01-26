package reposetup

import (
	"fmt"
	"path/filepath"

	"github.com/Benchkram/bob/bob"

	git "github.com/go-git/go-git/v5"

	"github.com/Benchkram/errz"
)

const (
	// TopRepoDirName is the name of the top-repo
	// ususally the first repo containing a .bob directory
	TopRepoDirName = "top-repo"

	// ChildReposDirName syntetically created repos for testing purpose
	ChildReposDirName = "repos"
)

// Names for the used child repos.
var (
	ChildOne   = "child1"
	ChildTwo   = "child2"
	ChildThree = "child3"

	Childs = []string{
		ChildOne,
		ChildTwo,
		ChildThree,
	}

	ChildRecursive  = "childrecursive"
	ChildPlayground = "childplayground"
)

func TopRepo(basePath string) (string, error) {
	const isBare = false

	repoPath := filepath.Join(basePath, TopRepoDirName)
	_, err := git.PlainInit(repoPath, isBare)
	if err != nil {
		return "", fmt.Errorf("failed to create repo %q: %w", repoPath, err)
	}

	return repoPath, nil
}

// ChildRepos creates child repos as defined and returns the path
// to the wrapping directory
func ChildRepos(basePath string) (childs []string, err error) {
	defer errz.Recover(&err)

	childs = []string{}

	for _, child := range Childs {
		err = createAndFillRepo(basePath, child)
		errz.Fatal(err)

		childs = append(childs, filepath.Join(basePath, child))
	}

	return childs, nil
}

func RecursiveRepo(basePath string) (_ string, err error) {
	defer errz.Recover(&err)

	err = createAndFillRepo(basePath, ChildRecursive)
	errz.Fatal(err)

	path := filepath.Join(basePath, ChildRecursive)
	b, err := bob.Bob(bob.WithDir(path))
	errz.Fatal(err)

	err = b.Init()
	errz.Fatal(err)

	err = b.Add("https://github.com/pkg/errors.git", false, false)
	errz.Fatal(err)

	repo, err := git.PlainOpen(path)
	if err != nil {
		return "", fmt.Errorf("failed to open repo: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	err = wt.AddGlob(".")
	errz.Fatal(err)

	err = commit(wt, "Add bob files")
	errz.Fatal(err)

	return path, nil
}

func PlaygroundRepo(basePath string) (_ string, err error) {
	defer errz.Recover(&err)

	err = createAndFillRepo(basePath, ChildPlayground)
	errz.Fatal(err)

	path := filepath.Join(basePath, ChildPlayground)
	err = bob.CreatePlayground(path)
	errz.Fatal(err)

	repo, err := git.PlainOpen(path)
	if err != nil {
		return "", fmt.Errorf("failed to open repo: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	err = wt.AddGlob(".")
	errz.Fatal(err)

	err = commit(wt, "Add playground files")
	errz.Fatal(err)

	return path, nil
}

// BaseTestStructure fill a existing directory with basic testing blueprints.
// The generated structure looks like:
//
// basePath/
// basePath/top-repo
// basePath/top-repo/.git
// basePath/repos/child1/.git
// basePath/repos/child2/.git
// basePath/repos/child3/.git
//
// TODO: @leonklingele update this comment with the playground stuff.
//
// The top-repo is intended to call `bob init`.
// The child repos can be add to the top repo using `bob add`
//
func BaseTestStructure(basePath string) (topRepo string, childs []string, recursive, playgroundRepo string, err error) {
	defer errz.Recover(&err)

	if !filepath.IsAbs(basePath) {
		return "", nil, "", "", fmt.Errorf("basePath must be absolut")
	}

	topRepo, err = TopRepo(basePath)
	errz.Fatal(err)

	childs, err = ChildRepos(filepath.Join(basePath, ChildReposDirName))
	errz.Fatal(err)

	recursive, err = RecursiveRepo(basePath)
	errz.Fatal(err)

	playgroundRepo, err = PlaygroundRepo(basePath)
	errz.Fatal(err)

	return topRepo, childs, recursive, playgroundRepo, nil
}
