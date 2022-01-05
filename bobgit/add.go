package bobgit

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Benchkram/bob/pkg/bobutil"
	"github.com/Benchkram/bob/pkg/cmdutil"
	"github.com/Benchkram/errz"
)

var ErrOutsideCurrentDir = fmt.Errorf("Not allowed, path pointing to outside of repository.")

// Status executes `git add` in all repositories
// first level repositories found inside a .bob filtree.
// It parses the output of each call and creates a object
// containing status infos for all of them combined.
// The result is similar to what `git add` would print
// but visualy optimised for the multi repository case.
func Add(target string) (err error) {
	defer errz.Recover(&err)

	if strings.HasPrefix(target, "../") || strings.HasPrefix(target, "/") {
		return ErrOutsideCurrentDir
	}

	bobRoot, err := bobutil.FindBobRoot()
	errz.Fatal(err)

	depth, err := wdDepth(bobRoot)
	errz.Fatal(err)

	target = trimTarget(target)
	if depth > 0 {
		target = addCurrentDirectoryInFront(bobRoot, target)
	}

	err = os.Chdir(bobRoot)
	errz.Fatal(err)

	// Assure toplevel is a git repo
	isGit, err := isGitRepo(bobRoot)
	errz.Fatal(err)
	if !isGit {
		return ErrCouldNotFindGitDir
	}

	// search for git repos inside bobRoot/.
	repoNames, err := getAllRepos(bobRoot)
	errz.Fatal(err)

	filterRepoNames, err := compareTargetWithRepos(repoNames, target)
	errz.Fatal(err)

	for _, name := range filterRepoNames {

		if name == "." {
			name = strings.TrimSuffix(name, ".")
		}
		target = prepareAccordingTarget(target, name)

		fmt.Println(name)
		fmt.Println(target)

		output, err := cmdutil.GitAddDry(name, target)
		errz.Fatal(err)

		filenames, err := parseAddDryOutput(output)
		errz.Fatal(err)

		// fmt.Println(filenames)
		// fmt.Println("=========================")

		if len(filenames) > 0 {
			err = cmdutil.GitAdd(name, target)
			errz.Fatal(err)
		}
	}

	return nil
}

func parseAddDryOutput(buf []byte) (_ []string, err error) {
	fileNames := []string{}

	scanner := bufio.NewScanner(bytes.NewBuffer(buf))
	for scanner.Scan() {
		line := scanner.Text()
		nameStart := strings.Index(line, "'") + 1
		nameEnd := strings.Index(line[nameStart:], "'") + nameStart
		fileName := line[nameStart:nameEnd]
		fileNames = append(fileNames, fileName)
	}

	return fileNames, nil
}

func addCurrentDirectoryInFront(bobroot string, target string) string {
	wd, err := os.Getwd()
	errz.Fatal(err)

	dir, err := filepath.Abs(bobroot)
	errz.Fatal(err)

	if strings.HasPrefix(wd, dir) {
		trimmed := wd[len(dir):]
		updated := filepath.Join(".", trimmed, target)
		return updated
	}

	return target
}

func trimTarget(target string) string {
	tempTarget := strings.Trim(target, " ")

	tempTarget = strings.TrimPrefix(tempTarget, "./")
	tempTarget = strings.TrimSuffix(tempTarget, "/")

	return tempTarget
}

func prepareAccordingTarget(target string, repo string) string {
	if repo == "" {
		return target
	}

	if target == repo {
		return "."
	}

	tempTarget := target
	if strings.HasPrefix(target, repo) {
		tempTarget = target[len(repo)+1:]
	}

	return tempTarget
}

func compareTargetWithRepos(repos []string, target string) ([]string, error) {

	selected := []string{}

	if target == "." || target == "./" {
		return repos, nil
	}

	isDir, err := isDirectory(target)
	if err != nil {
		return selected, err
	}
	targetPath := target
	if !isDir {
		targetPath = filepath.Dir(target)
	}

	for _, r := range repos {
		if r == targetPath {
			selected = append(selected, r)
		}
	}

	if len(selected) == 0 {
		for _, r := range repos {
			if strings.HasPrefix(target, r) {
				selected = append(selected, r)
			}
		}
	}

	if len(selected) == 0 {
		selected = append(selected, ".")
	}

	return selected, nil
}

// isDirectory determines if a file represented
// by `path` is a directory or not
func isDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)

	// returns isDirectory false if file does not exist
	// to process the directory further in case of regex
	// in case of a sure directory it should not be processed
	// further
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}
