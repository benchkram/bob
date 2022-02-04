package bobpackages

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Benchkram/bob/pkg/aqua"
	"github.com/Benchkram/bob/pkg/packagemanager"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/blang/semver"
)

var (
	ErrInvalidPackageDefinition = errors.New("Invalid package definition")
)

type Packages struct {
	ListDirty []string `yaml:"packages"`

	// Packages managed in a map to eliminate duplicates
	Packages map[string]packagemanager.Package `yaml:"-"`
	manager  packagemanager.PackageManager
}

// Sanitize dirty inputs to package definitions
// Returns usererror on bad package definition
func (p *Packages) Sanitize() error {

	for _, pkg := range p.ListDirty {
		splits := strings.Split(pkg, "@")
		if len(splits) != 2 {
			return usererror.Wrap(ErrInvalidPackageDefinition)
		}
		name, versionStr := splits[0], splits[1]
		version, err := semver.Parse(versionStr)
		if err != nil {
			return usererror.Wrapm(err, fmt.Sprintf("bad package version format: %s", versionStr))
		}

		p.Packages[pkg] = packagemanager.Package{
			Name:    name,
			Version: version,
		}
	}

	// get list of packages to add them to packagemanager
	pkgs := make([]packagemanager.Package, 0, len(p.Packages))
	for _, pkg := range p.Packages {
		pkgs = append(pkgs, pkg)
	}

	p.manager = aqua.New()
	p.manager.Add(pkgs...)

	return nil
}

// Install passes down install call to internal packagemanager
func (p *Packages) Install(ctx context.Context) error {
	return p.manager.Install(ctx)
}
