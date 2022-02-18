package bob

import (
	"context"

	"github.com/Benchkram/errz"
)

// InstallPackages defined in bobfiles
func (b *B) InstallPackages(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	err = aggregate.Validate()
	errz.Fatal(err)

	err = aggregate.Packages.Install(ctx)
	errz.Fatal(err)

	return nil
}
