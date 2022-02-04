package packagemanager

import (
	"context"

	"github.com/blang/semver"
)

// Package contains information on a single package to be installed in a runtime environment
type Package struct {
	// Name (Location) of the package. Where to get the package from
	Name string
	// Version of the package. Which version is required
	Version semver.Version
}

// Packagemanager is used to install external packages and setup a runtime envrionment
type PackageManager interface {

	// Add Packages to be installed
	Add(...Package)

	// Insall all defined packages
	Install(context.Context) error

	// Setup run/build environment so that packages can be accessed
	SetEnvirionment() error
}
