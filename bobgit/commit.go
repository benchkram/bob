package bobgit

import (
	"os"

	"github.com/Benchkram/bob/pkg/bobutil"
	"github.com/Benchkram/errz"
)

// Commit executes `git commit -m ${message}` in all repositories
// first level repositories found inside a .bob filtree.
// The result is similar to what `git commit` would print
// but visualy optimised for the multi repository case.
func Commit(message string) (err error) {
	defer errz.Recover(&err)

	bobRoot, err := bobutil.FindBobRoot()
	errz.Fatal(err)

	// depth, err := wdDepth(bobRoot)
	// errz.Fatal(err)

	err = os.Chdir(bobRoot)
	errz.Fatal(err)

	// Assure toplevel is a git repo
	isGit, err := isGitRepo(bobRoot)
	errz.Fatal(err)
	if !isGit {
		return ErrCouldNotFindGitDir
	}

	return nil
}
