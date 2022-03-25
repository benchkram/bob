package remotestore

import (
	"context"
	"fmt"

	"github.com/Benchkram/bob/pkg/store"
	storeclient "github.com/Benchkram/bob/pkg/store-client"
	"github.com/Benchkram/errz"
)

type s struct {
	dir string

	// client to call the remote store.
	client storeclient.I
}

// New creates a filestore. The caller is responsible to pass a
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

// NewArtifact creates a new file. The caller is responsible to call Close().
// Existing artifacts are overwritten.
func (s *s) NewArtifact(_ context.Context, id string) (empty store.Artifact, _ error) {
	return empty, fmt.Errorf("not implemented")
}

// GetArtifact opens a file
func (s *s) GetArtifact(_ context.Context, id string) (empty store.Artifact, _ error) {
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
