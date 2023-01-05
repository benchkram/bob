package bobtask

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/pkg/file"
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

func (t *Task) FilterInputs() (err error) {
	defer errz.Recover(&err)

	start := time.Now()

	inputs, err := t.FilteredInputs()
	errz.Fatal(err)
	t.inputs = inputs

	fmt.Printf("filtering input for task %s took %s\n", t.name, time.Since(start).String())

	return nil
}

// filteredInputs returns inputs filtered by ignores and file targets.
// Calls sanitize on the result.
func (t *Task) FilteredInputs() ([]string, error) {

	wd, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}

	inputDirty := fmt.Sprintf("%s\n%s", t.InputDirty, defaultIgnores)
	inputDirtyUnique := appendUnique([]string{}, split(inputDirty)...)
	inputDirtyRooted := inputDirtyUnique
	if t.dir != "." {
		inputDirtyRooted = make([]string, len(inputDirtyUnique))
		for i, input := range inputDirtyUnique {
			// keep ignored in tact
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

			ignores = appendUnique(ignores, list...)
			continue
		}

		list, err := filepathutil.ListRecursive(input)
		if err != nil {
			return nil, fmt.Errorf("failed to list input: %w", err)
		}

		inputs = appendUnique(inputs, list...)
	}

	// Ignore file & dir targets stored in the same directory
	if t.target != nil {
		rootedEntries := rooted(t.target.FilesystemEntriesRawPlain(), t.dir)
		for _, path := range rootedEntries {
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
					ignores = appendUnique(ignores, list...)
					continue
				}
				ignores = appendUnique(ignores, t.target.FilesystemEntriesRawPlain()...)
			}
		}
	}

	// Ignore additional items found during aggregation.
	// Usually the targets of child tasks which are already rooted.
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
		optimisationOptions{wd: wd},
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
