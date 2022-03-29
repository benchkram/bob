package bob

import (
	"context"

	"github.com/benchkram/errz"
)

func (b *B) Verify(ctx context.Context) (err error) {
	defer errz.Recover(&err)

	// Aggregate()  calls VerifyBefore() internaly
	_, err = b.Aggregate()
	errz.Fatal(err)

	return err
}
