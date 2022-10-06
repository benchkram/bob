package store

import (
	"context"
	"fmt"
	"io"

	"github.com/benchkram/errz"
)

// Sync an item from the src store to the dst store.
// In case the item exists in dst Sync does nothing and returns nil.
func Sync(ctx context.Context, src, dst Store, id string, msgOnSync string) (err error) {
	defer errz.Recover(&err)

	found, err := exists(ctx, src, id)
	if !found {
		return ErrArtifactNotFoundinSrc
	}
	errz.Fatal(err)

	found, err = exists(ctx, dst, id)
	errz.Fatal(err)
	if found {
		return ErrArtifactAlreadyExists
	}

	if msgOnSync != "" {
		fmt.Println(msgOnSync)
	}

	srcReader, size, err := src.GetArtifact(ctx, id)
	errz.Fatal(err)

	dstWriter, err := dst.NewArtifact(ctx, id, size)
	errz.Fatal(err)

	tr := io.TeeReader(srcReader, dstWriter)
	buf := make([]byte, 256)
	for {
		_, err := tr.Read(buf)
		if err == io.EOF {
			_ = dstWriter.Close()
			break
		}
		errz.Fatal(err)
	}

	return src.Done()
}

func exists(ctx context.Context, store Store, id string) (found bool, err error) {
	defer errz.Recover(&err)

	artifactIds, err := store.List(ctx)
	errz.Fatal(err)

	for _, i := range artifactIds {
		if i == id {
			found = true
			break
		}
	}

	return found, nil
}
