package packagemanager

import (
	"context"

	"github.com/blang/semver"
)

type Package struct {
	Name    string
	Version semver.Version
}

type PackageManager interface {
	Add(...Package)
	Install(context.Context) error
}
