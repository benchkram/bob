package bobgit

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Benchkram/bob/bobgit/add"
	"github.com/Benchkram/bob/pkg/bobutil"
	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
)

var ErrOutsideBobWorkspace = fmt.Errorf("Not allowed, path pointing outside of Bob workspace.")

// Status executes `git add` in all repositories
// first level repositories found inside a .bob filtree.
// It parses the output of each call and creates a object
// containing status infos for all of them combined.
// The result is similar to what `git add` would print
// but visualy optimised for the multi repository case.
func Add(target string) (err error) {
	defer errz.Recover(&err)

	if strings.HasPrefix(target, "/") {
		return usererror.Wrap(ErrOutsideBobWorkspace)
	}

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

	at := add.NewTarget(target)

	// search for git repos inside bobRoot/.
	allRepos, err := getAllRepos(bobRoot)
	errz.Fatal(err)

	filteredRepos := at.PopulateAndFilterRepos(allRepos)

	for _, name := range filteredRepos {
		thistarget, err := at.GetRelativeTarget(name)
		errz.Fatal(err)

		if name == "." {
			name = strings.TrimSuffix(name, ".")
		}

		output, err := cmdutil.GitAddDry(name, thistarget)
		if err != nil {
			return usererror.Wrapm(err, "Failed to Add files to git.")
		}

		filenames := parseAddDryOutput(output)

		if len(filenames) > 0 {
			err = cmdutil.GitAdd(name, target)
			if err != nil {
				return usererror.Wrapm(err, "Failed to Add files to git.")
			}
		}
	}

	return nil
}

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

func convertTargetPathRelativeToRoot(root string, target string) (string, error) {
	dir, err := filepath.Abs(target)
	errz.Fatal(err)

	if !strings.HasPrefix(dir, root) {
		return target, ErrOutsideBobWorkspace
	}

	relativepath := dir[len(root)+1:]
	return relativepath, nil
}

// isDirectory determines if a file represented
// by `path` is a directory or not
// func IsDirectory(path string) (bool, error) {
// 	fileInfo, err := os.Stat(path)

// 	// returns isDirectory false if file does not exist
// 	// to process the directory further in case of regex
// 	// in case of a sure directory it should not be processed
// 	// further
// 	if err != nil && errors.Is(err, os.ErrNotExist) {
// 		return false, nil
// 	} else if err != nil {
// 		return false, err
// 	}

// 	return fileInfo.IsDir(), err
// }
