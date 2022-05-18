package bobtask

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"errors"

	"github.com/benchkram/bob/bobtask/export"
)

type optimisationOptions struct {
	// wd is the current working directory
	// to avoid calls to os.Getwd.
	wd string
}

// sanitizeInputs assures that inputs are only cosidered when they are inside the project dir.
// Needs to be called with the current working directory set to the tasks working directory.
func (t *Task) sanitizeInputs(inputs []string, opts optimisationOptions) ([]string, error) {

	projectRoot, err := resolve(".", optimisationOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project root %q: %w", t.dir, err)
	}

	sanitized := make([]string, 0, len(inputs))
	resolved := make(map[string]struct{})
	for _, f := range inputs {
		if strings.Contains(f, "../") {
			return nil, fmt.Errorf("'../' not allowed in file path %q", f)
		}

		resolvedPath, err := resolve(f, opts)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				log.Printf("failed to resolve %q: %v, ignoring\n", f, err)
				continue
			}
			return nil, fmt.Errorf("failed to resolve %q: %w", f, err)
		}

		if _, ok := resolved[resolvedPath]; !ok {
			if isOutsideOfProject(projectRoot, resolvedPath) {
				return nil, fmt.Errorf("file %q is outside of project [pr: %s]", resolvedPath, projectRoot)
			}

			resolved[resolvedPath] = struct{}{}
			sanitized = append(sanitized, resolvedPath)
		}
	}

	return sanitized, nil
}

func (t *Task) sanitizeExports(exports export.Map) (export.Map, error) {
	sanitizedExports := make(export.Map)
	for name, export := range exports {
		if strings.Contains(string(export), "..") {
			return nil, fmt.Errorf("'..' not allowed in file path %q", string(export))
		}
		sanitizedExports[name] = export
	}
	return sanitizedExports, nil
}

type absolutePathOrError struct {
	abs string
	err error
}

// absPathMap caches already resolved absolute paths.
// FIXME: in case of asynchronous calls this should be a sync map
// or use a lock.
var absPathMap = make(map[string]absolutePathOrError, 10000)

// resolve is a very basic implementation only preventing the inclusion of files outside of the project.
// It is very likely still possible to include other files with malicious intention.
func resolve(path string, opts optimisationOptions) (_ string, err error) {

	var abs string
	if filepath.IsAbs(path) {
		abs = filepath.Clean(path)
	} else {
		if opts.wd != "" {
			// use given wd to avoid calling os.Getwd() for each path.
			abs = filepath.Join(opts.wd, path)
		} else {
			abs, err = filepath.Abs(path)
			if err != nil {
				return "", err
			}
		}

	}

	aoe, ok := absPathMap[abs]
	if ok {
		return aoe.abs, aoe.err
	}

	lstat, err := os.Lstat(abs)
	if err != nil {
		return "", fmt.Errorf("lstat failed %q: %w", abs, err)
	}

	// follow symlinks
	if lstat.Mode()&os.ModeSymlink != 0 {
		sym, err := filepath.EvalSymlinks(abs)
		if err != nil {
			a := absolutePathOrError{"", fmt.Errorf("failed to follow symlink of %q: %w", abs, err)}
			absPathMap[abs] = a
			return a.abs, a.err
		}
		absPathMap[abs] = absolutePathOrError{abs: sym, err: nil}
		return sym, nil
	}

	absPathMap[abs] = absolutePathOrError{abs: abs, err: nil}
	return abs, nil
}

// sanitizeRebuild used to transform from dirty member to internal member
func (t *Task) sanitizeRebuild(s string) RebuildType {
	switch strings.ToLower(s) {
	case string(RebuildAlways):
		return RebuildAlways
	case string(RebuildOnChange):
		return RebuildOnChange
	default:
		return RebuildOnChange
	}
}

func isOutsideOfProject(root, f string) bool {
	return !strings.HasPrefix(f, root)
}
