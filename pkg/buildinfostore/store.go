package buildinfostore

import (
	"fmt"

	"github.com/Benchkram/bob/bobtask/buildinfo"
)

// get inspiration from https://github.com/tus/tusd/blob/48ffebec56fcf3221461b3f8cbe000e5367e2d48/pkg/handler/datastore.go#L50

var ErrBuildInfoDoesNotExist = fmt.Errorf("Build info does not exist")

type Store interface {
	NewBuildInfo(id string, _ *buildinfo.I) error

	GetBuildInfo(id string) (*buildinfo.I, error)
	GetBuildInfos() ([]*buildinfo.I, error)

	Clean() error
}
