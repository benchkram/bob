package filepathutil

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/benchkram/bob/pkg/filepathxx"
	"github.com/logrusorgru/aurora"
)

// DefaultIgnores
var (
	DefaultIgnores = map[string]bool{
		"node_modules": true,
		".git":         true,
	}
)

//var listRecursiveCache = make(map[string][]string, 1024)

func ClearListRecursiveCache() {
	// listRecursiveCache = make(map[string][]string, 1024)
}

// ListRecursive lists all files relative to input. It ignores symbolic links
// which are not inside the projectRoot.
func ListRecursive(inp string, projectRoot string) (all []string, err error) {
	// if result, ok := listRecursiveCache[inp]; ok {
	// 	return result, nil
	// }

	// FIXME: when "*" is passed as input it's likely to hit the cache
	// as there is no further information. Think how to handle the cache correctly
	// in those cases. For now the cache is disabled!

	// FIXME: new list recursive
	// * does input contain a glob? see https://pkg.go.dev/path/filepath#Match => read with filepathx.Glob
	// * check if input is a file => add file
	// * check if input is a dir => add files in dir recursively
	//
	// More input:
	// https://github.com/iriri/minimal/blob/9b2348d09c1ab2c25505f9933a3591ef9db6522a/gitignore/gitignore.go#L245
	// https://github.com/zabawaba99/go-gitignore/
	// https://github.com/gobwas/glob
	//
	// Thoughts: Is it possible to compile a ignoreList upfront?
	// Then check if the accessed file || dir can be skipped.
	// Maybe it's even possible to call skipdir on a walk func.

	// symLinkError are gathered here and printed at the end of
	// the function to stdout.
	symlinkErrors := []error{}

	// FIXME: possibly ignore here too, before calling listDir
	if s, err := os.Lstat(inp); err != nil || !s.IsDir() {
		// File

		// Use glob for unknowns (wildcard-paths) and existing files (non-dirs)
		matches, err := filepathxx.Glob(inp)
		if err != nil {
			return nil, fmt.Errorf("failed to glob %q: %w", inp, err)
		}

		for _, m := range matches {
			s, err := os.Lstat(m)
			if err == nil && !s.IsDir() {
				isValid, err := isValidFile(s.Name(), s, projectRoot)
				if err != nil {
					symlinkErrors = append(symlinkErrors, err)
				}

				if !isValid {
					continue
				}

				// Existing file
				all = append(all, m)
			} else {
				// Directory
				files, symErrors, err := listDir(m, projectRoot)
				if err != nil {
					return nil, fmt.Errorf("failed to list dir: %w", err)
				}
				symlinkErrors = append(symlinkErrors, symErrors...)
				all = append(all, files...)
			}
		}
	} else {
		// Directory
		files, symErrors, err := listDir(inp, projectRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to list dir: %w", err)
		}
		symlinkErrors = append(symlinkErrors, symErrors...)
		all = append(all, files...)
	}

	for i, sErr := range symlinkErrors {
		fmt.Println(fmt.Sprintf("%s", aurora.Red("Warning: ")) + sErr.Error())
		if i > 10 {
			break
		}
	}

	// listRecursiveMap[inp] = all
	return all, nil
}

func listDir(path string, projectRoot string) (all []string, symlinkErrors []error, _ error) {

	symlinkErrors = []error{}
	all = []string{}
	if err := filepath.WalkDir(path, func(p string, fi fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip default ignored
		if fi.IsDir() && ignored(fi.Name()) {
			return fs.SkipDir
		}

		// Append file
		if fi.IsDir() {
			return nil

		}

		fileInfo, err := fi.Info()
		if err != nil {
			return err
		}

		isValid, err := isValidFile(p, fileInfo, projectRoot)
		if err != nil {
			symlinkErrors = append(symlinkErrors, err)
		}

		if isValid {
			all = append(all, p)
		}

		return nil
	}); err != nil {
		return nil, nil, fmt.Errorf("failed to walk dir %q: %w", path, err)
	}

	return all, symlinkErrors, nil
}

// isValidFile returns true if a symlink resolves succesfully into a path relative to projectRoot.
// It also returns true if the file is a regular file or directory.
//
// The returned error contains a "failed to follow symlink" hint and should
// be presented to the user.
//
// filepathHint is used to get around the
func isValidFile(path string, info fs.FileInfo, projectRoot string) (bool, error) {
	if info.Mode()&os.ModeSymlink != 0 {

		sym, err := filepath.EvalSymlinks(path)
		if err != nil {
			return false, fmt.Errorf("failed to follow symlink %q: %w", path, err)
		}

		if strings.HasPrefix(sym, "/") {
			return false, fmt.Errorf("symbolic link [%s] points to a location [%s] outside of the project [%s]", path, sym, projectRoot)
		}

		absSym, err := filepath.Abs(sym)
		if err != nil {
			return false, err
		}
		if !strings.HasPrefix(absSym, projectRoot) {
			return false, fmt.Errorf("symbolic link [%s] points to a location [%s] outside of the project [%s]", path, sym, projectRoot)
		}
	}

	return true, nil
}

func ignored(fileName string) bool {
	return DefaultIgnores[fileName]
}
