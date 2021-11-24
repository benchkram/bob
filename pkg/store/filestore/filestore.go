package filestore

import (
	"context"
	"os"
	"path/filepath"

	"github.com/Benchkram/bob/pkg/store"
)

type s struct {
	dir string
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
func (s *s) NewArtifact(_ context.Context, id string) (store.Artifact, error) {
	return os.Create(filepath.Join(s.dir, id))
}

// GetArtifact opens a file
func (s *s) GetArtifact(_ context.Context, id string) (empty store.Artifact, _ error) {
	return os.Open(filepath.Join(s.dir, id))
}
