package bobtask

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/Benchkram/bob/pkg/filepathutil"
)

type Input struct {
	Inputs []string
	Ignore []string
}

func MakeInput() Input {
	return Input{
		Inputs: []string{},
		Ignore: []string{},
	}
}

func (t *Task) Inputs() []string {
	return t.inputs
}

// filteredInputs returns inputs filtered by ignores.
// Calls sanitize on the result.
func (t *Task) filteredInputs() ([]string, error) {
	owd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}
	if err := os.Chdir(t.dir); err != nil {
		return nil, fmt.Errorf("failed to change current working directory to %s: %w", t.dir, err)
	}
	defer func() {
		if err := os.Chdir(owd); err != nil {
			log.Printf("failed to change current working directory back to %s: %v\n", owd, err)
		}
	}()

	// Determine inputs
	var inputs []string
	for _, input := range unique(t.InputDirty.Inputs) {
		list, err := filepathutil.ListRecursive(input)
		if err != nil {
			return nil, fmt.Errorf("failed to list input: %w", err)
		}

		inputs = append(inputs, list...)
	}
	inputs = unique(inputs)

	// Determine files to be ignored
	var ignores []string
	for _, ignore := range unique(t.InputDirty.Ignore) {
		list, err := filepathutil.ListRecursive(ignore)
		if err != nil {
			return nil, fmt.Errorf("failed to list ignore: %w", err)
		}

		ignores = append(ignores, list...)
	}
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

	sanitizedInputs, err := t.sanitizeInputs(filteredInputs)
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
