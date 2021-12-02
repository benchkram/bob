package bob

import (
	"context"

	"github.com/Benchkram/errz"
)

func (b B) Clean() (err error) {
	defer errz.Recover(&err)

	err = b.CleanBuildInfoStore()
	errz.Fatal(err)
	err = b.CleanLocalStore()
	errz.Fatal(err)

	return nil
}

func (b B) CleanBuildInfoStore() error {
	return b.buildInfoStore.Clean()
}

func (b B) CleanLocalStore() error {
	return b.local.Clean(context.TODO())
}
