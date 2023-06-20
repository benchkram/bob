package goinputdiscovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/benchkram/bob/pkg/inputdiscovery"
	"github.com/benchkram/bob/pkg/sliceutil"
	"github.com/benchkram/errz"
	"golang.org/x/tools/go/packages"
)

var Keyword = "go"

type goInputDiscovery struct {
	projectDir string

	closed map[string]bool
}

type Option func(discovery *goInputDiscovery)

func New(options ...Option) inputdiscovery.InputDiscovery {
	id := &goInputDiscovery{}

	id.closed = make(map[string]bool)
	for _, opt := range options {
		opt(id)
	}
	return id
}

// DiscoverInputs lists all directories which are used as input for the go package
// The path of the given package path has to be absolute.
// Returned paths are relative to the project dir.
// The function expects that there is a 'go.mod' and 'go.sum' file in the project dir.
func (id *goInputDiscovery) DiscoverInputs(packagePathAbs string) (_ []string, err error) {
	defer errz.Recover(&err)

	if !filepath.IsAbs(packagePathAbs) {
		return nil, fmt.Errorf("package path %s is not absolute", packagePathAbs)
	}

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

	paths := sliceutil.Unique(id.localDependencies(pkg.Imports, prefix))

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

	// add the go mod and go sum file (they have to exist)
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

	// add go.work and go.work.sum if any
	goWorkPath := filepath.Join(id.projectDir, "go.work")
	_, err = os.Stat(goWorkPath)
	if err == nil {
		resultAbs = append(resultAbs, goWorkPath)
	}
	goWorkSumPath := filepath.Join(id.projectDir, "go.work.sum")
	_, err = os.Stat(goWorkSumPath)
	if err == nil {
		resultAbs = append(resultAbs, goWorkSumPath)
	}

	// make paths relative to project dir
	var resultRel []string
	for _, p := range resultAbs {
		rel, err := filepath.Rel(id.projectDir, p)
		errz.Fatal(err)
		resultRel = append(resultRel, rel)
	}

	return resultRel, nil
}

// localDependencies is a recursive function that resolves the tree of imports provided by x/tools/packages
// it returns only packages with an ID-prefix same as the provided prefix
// it uses the closed variable to keep track of which packages were already handled to prevent an endless loop
func (id *goInputDiscovery) localDependencies(imports map[string]*packages.Package, prefix string) []string {
	var deps []string
	for _, pkg := range imports {
		// if the package is a local package add its whole dir
		if strings.HasPrefix(pkg.ID, prefix) {
			slug := strings.TrimPrefix(pkg.ID, prefix)
			slugParts := strings.Split(slug, "/")
			if len(slugParts) > 0 {
				deps = append(deps, slugParts[0]+"/")
			}

		} else if pkg.Module != nil && pkg.Module.Replace != nil && strings.HasPrefix(pkg.Module.Replace.Path, "./") {
			// go.mod allows to replace modules: https://go.dev/ref/mod#go-mod-file-replace
			// if the replacement is a local path starting with './' we add it to the local dependencies
			// go.mod also allows '../' paths, but we require that all content is inside the bob folder, so we ignore it
			deps = append(deps, strings.TrimPrefix(pkg.Module.Replace.Path, "./")+"/")
		} else if pkg.Module != nil && strings.HasPrefix(pkg.Module.Dir, id.projectDir) {
			// if go.work is used to extend the workspace we need to find the packages
			deps = append(deps, strings.TrimPrefix(pkg.Module.Dir, id.projectDir+"/"))
		}

		open := make(map[string]*packages.Package)
		for pkgId, secondLevelPkg := range pkg.Imports {
			if !id.closed[pkgId] {
				open[pkgId] = secondLevelPkg
				id.closed[pkgId] = true
			}
		}
		if len(open) > 0 {
			newDeps := id.localDependencies(open, prefix)
			deps = append(deps, newDeps...)
		}
	}
	return deps
}
