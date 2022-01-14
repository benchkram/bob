package bobgit

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/Benchkram/bob/bobgit/pathspec"
	"github.com/Benchkram/bob/pkg/bobutil"
	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
)

// pathspecItem stores a git pathspec with the repository
// path relative to the bob root
type pathspecItem struct {
	pathspec string
	repo     string
}

// Add executes `git add` in all repositories
// first level repositories found inside a .bob filtree.
// run git add commands by travsersing all the repositories
// inside the bob workspace. if target is provided "."
// it runs `git add .` in all repos, else run `git add ${relativeTargetPath}`
// only on the selected repos depending on the target path
func Add(targets ...string) (err error) {
	pathlist := []pathspecItem{}

	for _, target := range targets {
		defer errz.Recover(&err)

		bobRoot, err := bobutil.FindBobRoot()
		errz.Fatal(err)

		target, err = convertTargetPathRelativeToRoot(bobRoot, target)
		if err != nil {
			return usererror.Wrap(err)
		}

		err = os.Chdir(bobRoot)
		errz.Fatal(err)

		// Assure toplevel is a git repo
		isGit, err := isGitRepo(bobRoot)
		errz.Fatal(err)
		if !isGit {
			return usererror.Wrap(ErrCouldNotFindGitDir)
		}

		ps := pathspec.New(target)

		// search for git repos inside bobRoot/.
		allRepos, err := findRepos(bobRoot)
		errz.Fatal(err)

		filteredRepos := ps.SelectReposByPath(allRepos)

		for _, name := range filteredRepos {
			thistarget, err := ps.GetRelativePathspec(name)
			errz.Fatal(err)

			if name == "." {
				name = strings.TrimSuffix(name, ".")
			}

			output, err := cmdutil.GitAddDry(name, thistarget)
			if err != nil {
				return usererror.Wrapm(err, "Failed to Add files to git.")
			}

			filenames := parseAddDryOutput(output)

			for _, f := range filenames {
				pathspecItem := pathspecItem{
					repo:     name,
					pathspec: f,
				}
				pathlist = append(pathlist, pathspecItem)
			}
		}
	}

	for _, pi := range pathlist {
		err = cmdutil.GitAdd(pi.repo, pi.pathspec)
		if err != nil {
			return usererror.Wrapm(err, "Failed to Add files to git.")
		}
	}

	return nil
}

// parseAddDryOutput parse the output from `git add --dry-run`
// and returns a list of filenames
//
// Example output:
//   $ git add . --dry-run
//   add 'bobgit/add.go'
//   add 'bobgit/bobgit.go'
//   add 'qq'
//
func parseAddDryOutput(buf []byte) (_ []string) {
	fileNames := []string{}

	scanner := bufio.NewScanner(bytes.NewBuffer(buf))
	for scanner.Scan() {
		line := scanner.Text()
		nameStart := strings.Index(line, "'") + 1
		nameEnd := strings.Index(line[nameStart:], "'") + nameStart
		fileName := line[nameStart:nameEnd]
		fileNames = append(fileNames, fileName)
	}

	return fileNames
}

// convertTargetPathRelativeToRoot returns the relative targetpath from
// the provided root. e.g. converts `../sample/path` to `bobroot/sample/path`
func convertTargetPathRelativeToRoot(root string, target string) (string, error) {
	dir, err := filepath.Abs(target)
	errz.Fatal(err)

	if dir == root {
		return ".", nil
	}

	if !strings.HasPrefix(dir, root) {
		return target, ErrOutsideBobWorkspace
	}

	relativepath := dir[len(root)+1:]

	if target == "." || target == "" || strings.HasSuffix(target, "/.") {
		relativepath = relativepath + "/."
	}

	return relativepath, nil
}
