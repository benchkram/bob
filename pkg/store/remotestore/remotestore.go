package remotestore

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/benchkram/bob/pkg/store"
	storeclient "github.com/benchkram/bob/pkg/store-client"
	"github.com/benchkram/errz"
)

type s struct {
	// client to call the remote store.
	client storeclient.I

	username string
	project  string

	wg sync.WaitGroup
}

// New creates a remote store. The caller is responsible to pass a
// existing directory.
func New(username, project string, opts ...Option) store.Store {
	s := &s{
		username: username,
		project:  project,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(s)
	}

	if s.client == nil {
		panic(fmt.Errorf("no client"))
	}

	return s
}

// NewArtifact uploads an artifact. The caller is responsible to call Close().
// Existing artifacts are overwritten.
func (s *s) NewArtifact(ctx context.Context, artifactID string) (wc io.WriteCloser, err error) {
	s.wg.Add(1)
	reader, writer := io.Pipe()

	go func() {
		defer s.wg.Done()
		err := s.client.Upload(
			ctx,
			s.project,
			artifactID,
			reader,
		)
		errz.Log(err)

		//_ = writer.CloseWithError(err)
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

// Done waits till all processing finished
func (s *s) Done() {
	s.wg.Wait()
}
