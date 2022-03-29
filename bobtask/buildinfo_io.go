package bobtask

import (
	"github.com/benchkram/bob/bobtask/buildinfo"
	"github.com/benchkram/errz"
)

// ReadBuildinfo indexed by the input hash from the store
func (t *Task) ReadBuildinfo() (bi *buildinfo.I, err error) {
	defer errz.Recover(&err)

	hashIn, err := t.HashIn()
	errz.Fatal(err)

	bi, err = t.buildInfoStore.GetBuildInfo(hashIn.String())
	errz.Fatal(err)

	return bi, nil
}

// WriteBuildinfo indexed by the input hash to the store
func (t *Task) WriteBuildinfo(buildinfo *buildinfo.I) (err error) {
	defer errz.Recover(&err)

	hashIn, err := t.HashIn()
	errz.Fatal(err)

	err = t.buildInfoStore.NewBuildInfo(hashIn.String(), buildinfo)
	errz.Fatal(err)

	return nil
}
