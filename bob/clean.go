package bob

import (
	"context"

	"github.com/benchkram/errz"
)

func (b B) Clean() (err error) {
	defer errz.Recover(&err)

	err = b.CleanBuildInfoStore()
	errz.Fatal(err)
	err = b.CleanLocalStore()
	errz.Fatal(err)
	err = b.CleanNixCache()
	errz.Fatal(err)

	return nil
}

func (b B) CleanBuildInfoStore() error {
	return b.buildInfoStore.Clean()
}

func (b B) CleanLocalStore() error {
	return b.local.Clean(context.TODO())
}

func (b B) CleanNixCache() error {
	return b.Nix().Clean()
}
