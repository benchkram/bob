package goinputdiscovery

import (
	"strings"

	"golang.org/x/tools/go/packages"
)

type depResolver struct {
	closed map[string]bool
}

func newDepResolver() *depResolver {
	dr := &depResolver{}
	dr.closed = make(map[string]bool)
	return dr
}

// localDependencies is a recursive function that resolves the tree of imports provided by x/tools/packages
// it returns only packages with an ID-prefix same as the provided prefix
// it uses the closed variable to keep track of which packages were already handled to prevent an endless loop
func (dr *depResolver) localDependencies(imports map[string]*packages.Package, prefix string) []string {
	var deps []string
	for _, pkg := range imports {
		// if the package is a local package add its whole dir
		if strings.HasPrefix(pkg.ID, prefix) {
			slug := strings.TrimPrefix(pkg.ID, prefix)
			slugParts := strings.Split(slug, "/")
			if len(slugParts) > 0 {
				deps = append(deps, slugParts[0]+"/")
			}
		}
		open := make(map[string]*packages.Package)
		for pkgId, secondLevelPkg := range pkg.Imports {
			if !dr.closed[pkgId] {
				open[pkgId] = secondLevelPkg
				dr.closed[pkgId] = true
			}
		}
		if len(open) > 0 {
			newDeps := dr.localDependencies(open, prefix)
			deps = append(deps, newDeps...)
		}
	}
	return deps
}
