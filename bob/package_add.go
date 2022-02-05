package bob

import (
	"context"
	"fmt"
	"regexp"

	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/Benchkram/errz"
)

// AddPackages to bobfile
func (b *B) AddPackages(ctx context.Context, pkgs ...string) (err error) {
	defer errz.Recover(&err)

	// validate package strings
	pkgRegex := regexp.MustCompile(`(\w+\/?\w+)+@(v?(\d+.){2}\d+)`)
	for _, arg := range pkgs {
		if !pkgRegex.MatchString(arg) {
			fmt.Printf("\"%s\" is not a valid package.\n", arg)
			return
		}
	}

	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	err = aggregate.Validate()
	errz.Fatal(err)

	// make sure there are no duplicates
	fullList := aggregate.Packages.ListDirty

	for _, pkg := range pkgs {
		var found bool
		for _, bPkg := range fullList {
			if pkg == bPkg {
				found = true
				break
			}
		}

		if found {
			continue
		}

		boblog.Log.Info(fmt.Sprintf("Adding package %s", pkg))

		fullList = append(fullList, pkg)
	}

	aggregate.Packages.ListDirty = fullList

	// Write new bobfile context
	err = aggregate.BobfileSave(aggregate.Dir())
	errz.Fatal(err)

	return nil
}
