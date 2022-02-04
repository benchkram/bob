package bob

import (
	"context"

	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/Benchkram/errz"
	"github.com/logrusorgru/aurora"
)

// PrunePackages - remove all packages and package archives
func (b *B) PrunePackages(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	boblog.Log.Info(aurora.Green("Pune packages...").String())

	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	err = aggregate.Validate()
	errz.Fatal(err)

	err = aggregate.Packages.Prune(ctx)
	errz.Fatal(err)
	// TODO: do we want to also prune child repositories?

	boblog.Log.Info(aurora.Green("...done").String())

	return nil
}
