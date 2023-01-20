package bobtask

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/pkg/filepathutil"
	"github.com/benchkram/errz"
)

func (t *Task) Inputs() []string {
	return t.inputs
}

func (t *Task) SetInputs(inputs []string) {
	t.inputs = inputs
}

var (
	defaultIgnores = fmt.Sprintf("!%s\n!%s",
		global.BobWorkspaceFile,
		filepath.Join(global.BobCacheDir, "*"),
	)
)

func (t *Task) FilterInputs(wd string) (err error) {
	defer errz.Recover(&err)

	inputs, err := t.FilteredInputs(wd)
	errz.Fatal(err)
	t.inputs = inputs

	return nil
}

// filteredInputs returns inputs filtered by ignores and file targets.
// Calls sanitize on the result.
func (t *Task) FilteredInputs(wd string) ([]string, error) {

	inputDirty := split(fmt.Sprintf("%s\n%s", t.InputDirty, defaultIgnores))
	inputDirtyRooted := inputDirty
	if t.dir != "." {
		inputDirtyRooted = make([]string, len(inputDirty))
		for i, input := range inputDirty {
			// keep ignored in inputDirty
			if strings.HasPrefix(input, "!") {
				inputDirtyRooted[i] = "!" + filepath.Join(t.dir, strings.TrimPrefix(input, "!"))
				continue
			}
			inputDirtyRooted[i] = filepath.Join(t.dir, input)
		}
	}

	// Determine inputs and files to be ignored
	var inputs []string
	var ignores []string
	for _, input := range inputDirtyRooted {
		// Ignore starts with !
		if strings.HasPrefix(input, "!") {
			input = strings.TrimPrefix(input, "!")
			list, err := filepathutil.ListRecursive(input)
			if err != nil {
				return nil, fmt.Errorf("failed to list input: %w", err)
			}

			ignores = append(ignores, list...)
			continue
		}

		list, err := filepathutil.ListRecursive(input)
		if err != nil {
			return nil, fmt.Errorf("failed to list input: %w", err)
		}

		inputs = append(inputs, list...)
	}

	// Ignore file & dir targets stored in the same directory
	if t.target != nil {
		for _, path := range rooted(t.target.FilesystemEntriesRawPlain(), t.dir) {
			info, err := os.Stat(path)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
				return nil, fmt.Errorf("failed to stat %s: %w", path, err)
			}

			if info.IsDir() {
				list, err := filepathutil.ListRecursive(path)
				if err != nil {
					return nil, fmt.Errorf("failed to list input: %w", err)
				}
				ignores = append(ignores, list...)
				continue
			}
			ignores = append(ignores, t.target.FilesystemEntriesRawPlain()...)
		}
	}

	// Ignore additional items found during aggregation.
	// Usually the targets of child tasks which are already rooted.
	for _, path := range t.InputAdditionalIgnores {
		info, err := os.Stat(path)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("failed to stat %s: %w", path, err)
		}

		if info.IsDir() {
			list, err := filepathutil.ListRecursive(path)
			if err != nil {
				return nil, fmt.Errorf("failed to list input: %w", err)
			}
			ignores = append(ignores, list...)
			continue
		}
		ignores = append(ignores, path)
	}

	inputs = unique(inputs)
	ignores = unique(ignores)

	// Filter
	filteredInputs := make([]string, 0, len(inputs))
	for _, input := range inputs {
		var isIgnored bool
		for _, ignore := range ignores {
			if strings.TrimPrefix(input, "./") == ignore {
				isIgnored = true
				break
			}
		}

		if !isIgnored {
			filteredInputs = append(filteredInputs, input)
		}
	}

	sanitizedInputs, err := t.sanitizeInputs(
		filteredInputs,
		optimisationOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize inputs: %w", err)
	}

	sort.Strings(sanitizedInputs)

	// fmt.Println(t.name)
	// fmt.Println("Inputs:", inputs)
	// fmt.Println("Ignores:", ignores)
	// fmt.Println("Filtered:", filteredInputs)
	// fmt.Println("Sanitized:", sanitizedInputs)
	// fmt.Println("Sorted:", sortedInputs)

	return sanitizedInputs, nil
}

func rooted(ss []string, prefix string) []string {
	if prefix == "." {
		return ss
	}
	for i, s := range ss {
		ss[i] = filepath.Join(prefix, s)
	}
	return ss
}

func unique(ss []string) []string {
	unique := make([]string, 0, len(ss))

	um := make(map[string]struct{})
	for _, s := range ss {
		if _, ok := um[s]; !ok {
			um[s] = struct{}{}
			unique = append(unique, s)
		}
	}

	return unique
}

func appendUnique(a []string, xx ...string) []string {
	for _, x := range xx {
		add := true
		for _, y := range a {
			if x == y {
				add = false
				break
			}
		}

		if add {
			a = append(a, x)
		}
	}
	return a
}

// Split splits a single-line "input" to a slice of inputs.
//
// It currently supports the following syntaxes:
//
//	Input: |-
//	  main1.go
//	  someotherfile
//	Output:
//	  [ "./main1.go", "!someotherfile" ]
func split(inputDirty string) []string {

	// Replace leading and trailing spaces for clarity.
	inputDirty = strings.TrimSpace(inputDirty)

	lines := strings.Split(inputDirty, "\n")
	if len(lines) == 1 && len(lines[0]) == 0 {
		return []string{}
	}

	inputs := []string{}

	// Remove possible trailing spaces
	for _, line := range lines {
		// Remove commented and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		inputs = append(inputs, strings.TrimSpace(line))
	}

	return inputs
}
