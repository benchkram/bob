package bobgit

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/Benchkram/bob/bobgit/add"
	"github.com/Benchkram/bob/pkg/bobutil"
	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
)

// Add executes `git add` in all repositories
// first level repositories found inside a .bob filtree.
// run git add commands by travsersing all the repositories
// inside the bob workspace. if target is provided "." or  ""
// it runs `git add .` in all repos, else run `git add ${relativeTargetPath}`
// only on the selected repos depending on the target path
func Add(targets ...string) (err error) {
	for _, target := range targets {
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

		filteredRepos := at.SelectReposByTarget(allRepos)

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
				err = cmdutil.GitAdd(name, thistarget)
				if err != nil {
					return usererror.Wrapm(err, "Failed to Add files to git.")
				}
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

// func MatchFiles(pattern string) ([]string, error) {
// 	abspath, err := filepath.Abs(pattern)
// 	if err != nil {
// 		return nil, err
// 	}

// 	root := filepath.Dir(abspath)
// 	filepattern := filepath.Base(pattern)
// 	if filepattern == "." {
// 		filepattern = "*"
// 	}

// 	var matches []string
// 	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}

// 		if add.Contains(dontFollow, info.Name()) {
// 			return filepath.SkipDir
// 		}

// 		if info.IsDir() {
// 			return nil
// 		}

// 		if matched, err := filepath.Match(filepattern, filepath.Base(path)); err != nil {
// 			return err
// 		} else if matched {
// 			temppath := path[len(root)+1:]
// 			matches = append(matches, temppath)
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return matches, nil
// }
