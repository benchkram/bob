package bobtask

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"errors"

	"github.com/Benchkram/bob/bobtask/export"
)

var (
	BackwardPathError = fmt.Errorf("'../' not allowed in file path")
	OutsideDirError   = fmt.Errorf("filepath outside of project")
)

func (t *Task) sanitizeInputs(inputs []string) ([]string, error) {
	projectRoot, err := resolve(t.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project root %q: %w", t.dir, err)
	}

	sanitized := make([]string, 0, len(inputs))
	resolved := make(map[string]struct{})
	for _, f := range inputs {
		if strings.Contains(f, "../") {
			return nil, fmt.Errorf("%w %q", BackwardPathError, f)
		}

		resolvedPath, err := resolve(f)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				log.Printf("failed to resolve %q: %v, ignoring\n", f, err)
				continue
			}
			return nil, fmt.Errorf("failed to resolve %q: %w", f, err)
		}

		if _, ok := resolved[resolvedPath]; !ok {
			if isOutsideOfProject(projectRoot, resolvedPath) {
				return nil, fmt.Errorf("%q: %w", resolvedPath, OutsideDirError)
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
			return nil, fmt.Errorf("%w %q", BackwardPathError, string(export))
		}
		sanitizedExports[name] = export
	}
	return sanitizedExports, nil
}

// Adapted from https://github.com/moby/moby/blob/7b9275c0da707b030e62c96b679a976f31f929d3/pkg/containerfs/containerfs.go#L73-L79 et al.
// TODO: This is just a very basic implementation only preventing the inclusion of files outside of the project.
// It is very likely still possible to include other files with malicious intention.
func resolve(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to determine abs path: %w", err)
	}

	sym, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("failed to follow symlink of %q: %w", abs, err)
	}

	return sym, nil
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
