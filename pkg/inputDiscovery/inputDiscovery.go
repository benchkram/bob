package inputDiscovery

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/benchkram/errz"
	"golang.org/x/mod/modfile"
)

type InputDiscovery interface {
	GetInputs(string) ([]string, error)
}

type goInputDiscovery struct {
}

func NewGoInputDiscovery() InputDiscovery {
	return &goInputDiscovery{}
}

// GetInputs lists all directories which are used as input for the main go file
// The path of the given mainFile has to be absolute.
// Returned paths are absolute.
// The function expects that there is a 'go.mod' file next to the main file.
func (id *goInputDiscovery) GetInputs(mainFileAbs string) (_ []string, err error) {
	defer errz.Recover(&err)

	// TODO: check if file is a main file
	// TODO: check if 'go list' is available

	mainFileDir := filepath.Dir(mainFileAbs)
	mainFile := filepath.Base(mainFileAbs)

	modFilePath := fmt.Sprintf("%s/go.mod", mainFileDir)

	modFileContent, err := os.ReadFile(modFilePath)
	errz.Fatal(err)

	modFile, err := modfile.Parse(modFilePath, modFileContent, nil)
	errz.Fatal(err)
	packageName := modFile.Module.Mod.Path

	cmd := exec.Command("go", "list", "-f", "'{{ join .Deps \"\\n\" }}'", mainFile)
	cmd.Dir = mainFileDir
	out, err := cmd.Output()
	errz.Fatal(err)

	paths := make(map[string]bool)
	lines := strings.Split(string(out), "\n")
	for _, l := range lines {
		prefix := packageName + "/"
		if strings.HasPrefix(l, prefix) {
			slug := strings.TrimPrefix(l, prefix)
			slugParts := strings.Split(slug, "/")
			if len(slugParts) > 0 {
				paths[slugParts[0]] = true
			}
		}
	}
	var result []string
	for p, _ := range paths {
		result = append(result, filepath.Join(mainFileDir, p))
	}

	// add the main file itself
	result = append(result, mainFileAbs)

	// add the go mod and go sum file
	result = append(result, modFilePath)
	sumFilePath := fmt.Sprintf("%s/go.sum", mainFileDir)
	result = append(result, sumFilePath)
	return result, nil
}
