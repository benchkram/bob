package remotestore

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/pkg/store"
	storeclient "github.com/benchkram/bob/pkg/store-client"
)

type s struct {
	// client to call the remote store.
	client storeclient.I

	username string
	project  string

	wg  sync.WaitGroup
	err error
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
func (s *s) NewArtifact(ctx context.Context, artifactID string, size int64) (wc io.WriteCloser, err error) {
	s.wg.Add(1)
	reader, writer := io.Pipe()

	go func() {
		defer s.wg.Done()
		err := s.client.UploadArtifact(
			ctx,
			s.project,
			artifactID,
			reader,
			size,
		)
		if err != nil {
			// store the error, it will be returned from s.Done()
			s.err = err
		}
	}()

	return writer, nil
}

// GetArtifact opens a file
func (s *s) GetArtifact(ctx context.Context, id string) (rc io.ReadCloser, size int64, err error) {
	defer errz.Recover(&err)

	rc, size, err = s.client.GetArtifact(ctx, s.project, id)
	errz.Fatal(err)

	return rc, size, nil
}

func (s *s) Clean(_ context.Context) (err error) {
	defer errz.Recover(&err)
	return fmt.Errorf("not implemented")
}

// List the items id's in the store
func (s *s) List(ctx context.Context) (ids []string, err error) {
	defer errz.Recover(&err)

	ids, err = s.client.ListArtifacts(ctx, s.project)
	errz.Fatal(err)

	return ids, nil
}

// Done waits till all processing finished
func (s *s) Done() error {
	s.wg.Wait()

	return s.err
}

// ArtifactExists TODO: naive implementation.. implement one without downloading the artifact
func (s *s) ArtifactExists(ctx context.Context, id string) bool {
	_, _, err := s.client.GetArtifact(ctx, s.project, id)
	return err == nil
}
