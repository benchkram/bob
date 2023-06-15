package goinputdiscovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/benchkram/bob/pkg/inputdiscovery"
	"github.com/benchkram/errz"
	"golang.org/x/tools/go/packages"
)

var Keyword = "gopackage"

type goInputDiscovery struct {
	projectDir string
}

type Option func(discovery *goInputDiscovery)

func NewGoInputDiscovery(options ...Option) inputdiscovery.InputDiscovery {
	id := &goInputDiscovery{}
	for _, opt := range options {
		opt(id)
	}
	return id
}

// GetInputs lists all directories which are used as input for the go package
// The path of the given package path has to be absolute.
// Returned paths are relative to the project dir.
// The function expects that there is a 'go.mod' and 'go.sum' file in the project dir.
func (id *goInputDiscovery) GetInputs(packagePathAbs string) (_ []string, err error) {
	defer errz.Recover(&err)

	cfg := &packages.Config{
		Dir:  id.projectDir,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedDeps | packages.NeedImports | packages.NeedModule | packages.NeedEmbedFiles,
	}

	pkgs, err := packages.Load(cfg, packagePathAbs)
	errz.Fatal(err)

	if len(pkgs) < 1 {
		return nil, fmt.Errorf("did not find a go package at %s", packagePathAbs)
	} else if len(pkgs) > 1 {
		return nil, fmt.Errorf("found more than one go package at %s", packagePathAbs)
	}

	pkg := pkgs[0]

	if len(pkg.Errors) == 1 {
		return nil, fmt.Errorf("load package failed: %w", pkg.Errors[0])
	} else if len(pkg.Errors) > 1 {
		var errs []string
		for _, e := range pkg.Errors {
			errs = append(errs, e.Error())
		}
		return nil, fmt.Errorf("load package failed with multiple errors: %s", strings.Join(errs, ";"))
	}

	if pkg.Module == nil {
		return nil, fmt.Errorf("package has no go.mod file: expected it to be in the bob root dir")
	}
	modFilePath := pkg.Module.GoMod
	packageName := pkg.Module.Path

	prefix := packageName + "/"

	dr := newDepResolver()

	paths := unique(dr.localDependencies(pkg.Imports, prefix))

	var resultAbs []string
	for _, p := range paths {
		resultAbs = append(resultAbs, filepath.Join(id.projectDir, p))
	}

	// add go files in package
	resultAbs = append(resultAbs, pkg.GoFiles...)

	// add embedded files
	resultAbs = append(resultAbs, pkg.EmbedFiles...)

	// add all other files
	resultAbs = append(resultAbs, pkg.OtherFiles...)

	// add the go mod and go sum file if they exist
	_, err = os.Stat(modFilePath)
	if err != nil {
		return nil, fmt.Errorf("can not find 'go.mod' file at %s", modFilePath)
	}
	resultAbs = append(resultAbs, modFilePath)
	sumFilePath := filepath.Join(id.projectDir, "go.sum")
	_, err = os.Stat(sumFilePath)
	if err != nil {
		return nil, fmt.Errorf("can not find 'go.sum' file at %s", sumFilePath)
	}
	resultAbs = append(resultAbs, sumFilePath)

	// make paths relative to project dir
	var resultRel []string
	for _, p := range resultAbs {
		rel, err := filepath.Rel(id.projectDir, p)
		errz.Fatal(err)
		resultRel = append(resultRel, rel)
	}

	return resultRel, nil
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
