package bobtask

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/bob/pkg/filepathutil"
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

// filteredInputs returns inputs filtered by ignores and file targets.
// Calls sanitize on the result.
func (t *Task) filteredInputs() ([]string, error) {

	wd, err := filepath.Abs(t.dir)
	if err != nil {
		return nil, err
	}

	owd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}
	if err := os.Chdir(wd); err != nil {
		return nil, fmt.Errorf("failed to change current working directory to %s: %w", t.dir, err)
	}
	defer func() {
		if err := os.Chdir(owd); err != nil {
			log.Printf("failed to change current working directory back to %s: %v\n", owd, err)
		}
	}()

	inputDirty := fmt.Sprintf("%s\n%s", t.InputDirty, defaultIgnores)

	// Determine inputs and files to be ignored
	var inputs []string
	var ignores []string
	for _, input := range unique(split(inputDirty)) {
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

	// Also ignore file & dir targets stored in the same directory
	if t.target != nil {
		for _, path := range t.target.FilesystemEntriesRawPlain() {
			if file.Exists(path) {
				info, err := os.Stat(path)
				if err != nil {
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
	}

	// Also ignore additional ignores found during aggregation.
	// Usually the targets of child tasks.
	for _, path := range t.InputAdditionalIgnores {
		if file.Exists(path) {
			info, err := os.Stat(path)
			if err != nil {
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
			if input == ignore {
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
		optimisationOptions{wd: wd},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize inputs: %w", err)
	}

	sortedInputs := sanitizedInputs
	sort.Strings(sanitizedInputs)

	// fmt.Println("Inputs:", inputs)
	// fmt.Println("Ignores:", ignores)
	// fmt.Println("Filtered:", filteredInputs)
	// fmt.Println("Sanitized:", sanitizedInputs)
	// fmt.Println("Sorted:", sortedInputs)

	return sortedInputs, nil
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
