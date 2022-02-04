package bob

import (
	"context"

	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/Benchkram/errz"
	"github.com/logrusorgru/aurora"
)

// InstallPackages defined in bobfiles
func (b *B) InstallPackages(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	boblog.Log.Info(aurora.Green("Installing packages...").String())

	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	err = aggregate.Validate()
	errz.Fatal(err)

	err = aggregate.Packages.Install(ctx)
	errz.Fatal(err)

	boblog.Log.Info(aurora.Green("All packages successfully installed").String())

	return nil
}
