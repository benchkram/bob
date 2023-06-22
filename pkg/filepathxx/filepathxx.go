// From https://github.com/klauspost/filepathx/blob/master/filepathx.go
//
// Package filepathx adds double-star globbing support to the Glob function from the core path/filepath package.
// You might recognize "**" recursive globs from things like your .gitignore file, and zsh.
// The "**" glob represents a recursive wildcard matching zero-or-more directory levels deep.
package filepathxx

import (
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
)

// Globs represents one filepath glob, with its elements joined by "**".
type Globs []string

// Glob adds double-star support to the core path/filepath Glob function.
// It's useful when your globs might have double-stars, but you're not sure.
func Glob(pattern string) ([]string, error) {
	if !strings.Contains(pattern, "**") {
		// passthru to core package if no double-star
		return filepath.Glob(pattern)
	}
	return Globs(strings.Split(pattern, "**")).Expand()
}

const replaceIfAny = "[]*"

var replacements [][2]string

func init() {
	// Escape `filepath.Match` syntax.
	// On Unix escaping works with `\\`,
	// on Windows it is disabled, therefore
	// replace it by '?' := any character.
	if runtime.GOOS == "windows" {
		replacements = [][2]string{
			// Windows cannot have * in file names.
			{"[", "?"},
			{"]", "?"},
		}
	} else {
		replacements = [][2]string{
			{"*", "\\*"},
			{"[", "\\["},
			{"]", "\\]"},
		}
	}
}

// Expand finds matches for the provided Globs.
func (globs Globs) Expand() ([]string, error) {
	var matches = []string{""} // accumulate here
	for _, glob := range globs {
		if glob == "" {
			// If the glob is empty string that means it was **
			// By setting this to . patterns like **/*.txt are supported
			glob = "."
		}
		var hits []string
		//var hitMap = map[string]bool{}
		for _, match := range matches {
			if strings.ContainsAny(match, replaceIfAny) {
				for _, sr := range replacements {
					match = strings.ReplaceAll(match, sr[0], sr[1])
				}
			}

			paths, err := filepath.Glob(filepath.Join(match, glob))
			if err != nil {
				return nil, err
			}

			for _, path := range paths {
				err = filepath.WalkDir(path, func(path string, _ fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					// save deduped match from current iteration
					// if _, ok := hitMap[path]; !ok {
					// 	hits = append(hits, path)
					// 	hitMap[path] = true
					// }
					hits = appendUnique(hits, path)
					// if !contains(path, hits) {
					// 	hits = append(hits, path)
					// }
					return nil
				})
				if err != nil {
					return nil, err
				}
			}
		}
		matches = hits
	}

	// fix up return value for nil input
	if globs == nil && len(matches) > 0 && matches[0] == "" {
		matches = matches[1:]
	}

	return matches, nil
}

func appendUnique(a []string, x string) []string {
	for _, y := range a {
		if x == y {
			return a
		}
	}
	return append(a, x)
}
