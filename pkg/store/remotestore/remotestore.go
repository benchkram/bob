package remotestore

import (
	"context"
	"fmt"
	"io"

	"github.com/Benchkram/bob/pkg/store"
	storeclient "github.com/Benchkram/bob/pkg/store-client"
	"github.com/Benchkram/errz"
)

type s struct {
	dir string

	// client to call the remote store.
	client storeclient.I
}

// New creates a remote store. The caller is responsible to pass a
// existing directory.
func New(dir string, opts ...Option) store.Store {
	s := &s{
		dir: dir,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(s)
	}

	return s
}

// NewArtifact uploads an artifact. The caller is responsible to call Close().
// Existing artifacts are overwritten.
func (s *s) NewArtifact(ctx context.Context, id string) (io.WriteCloser, error) {

	reader, writer := io.Pipe()
	go func() {
		err := s.client.Upload(ctx,
			"projectID:TODO:XXX",
			id,
			"contentType:TODO",
			id, /*TODO:*/
			reader,
		)
		_ = writer.CloseWithError(err)
	}()

	return writer, nil
}

// GetArtifact opens a file
func (s *s) GetArtifact(_ context.Context, id string) (empty io.ReadCloser, _ error) {
	return empty, fmt.Errorf("not implemented")
}

func (s *s) Clean(_ context.Context) (err error) {
	defer errz.Recover(&err)
	return fmt.Errorf("not implemented")
}

// List the items id's in the store
func (s *s) List(_ context.Context) (items []string, err error) {
	defer errz.Recover(&err)
	return items, fmt.Errorf("not implemented")
}
