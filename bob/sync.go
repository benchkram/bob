package bob

import (
	"context"
	"os"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/errz"
)

func (b *B) SyncPush(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	wd, _ := os.Getwd()
	aggregate, err := bobfile.BobfileRead(wd)

	for _, sync := range aggregate.SyncCollections {
		err = sync.Push(ctx)
		errz.Fatal(err)
	}

	return nil
}

func (b *B) SyncPull(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	//aggregate, err := b.Aggregate()

	return nil
}

func (b *B) SyncListLocal(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	//aggregate, err := b.Aggregate()

	return nil
}

func (b *B) SyncListRemote(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	//aggregate, err := b.Aggregate()

	return nil
}
