package filestore

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/bob/pkg/store"
	"github.com/benchkram/errz"
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
func (s *s) NewArtifact(_ context.Context, artifactID string, _ int64) (io.WriteCloser, error) {
	return os.Create(filepath.Join(s.dir, artifactID))
}

// GetArtifact opens a file
func (s *s) GetArtifact(_ context.Context, id string) (empty io.ReadCloser, size int64, _ error) {
	f, err := os.Open(filepath.Join(s.dir, id))
	if err != nil {
		return nil, 0, err
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	return f, stat.Size(), nil
}

func (s *s) Clean(_ context.Context) (err error) {
	defer errz.Recover(&err)

	homeDir, err := os.UserHomeDir()
	errz.Fatal(err)
	if s.dir == "/" || s.dir == homeDir {
		return fmt.Errorf("Cleanup of %s is not allowed", s.dir)
	}

	entrys, err := os.ReadDir(s.dir)
	errz.Fatal(err)

	for _, entry := range entrys {
		if entry.IsDir() {
			continue
		}
		_ = os.Remove(filepath.Join(s.dir, entry.Name()))
	}

	return nil
}

// List the items id's in the store
func (s *s) List(_ context.Context) (items []string, err error) {
	defer errz.Recover(&err)
	entrys, err := os.ReadDir(s.dir)
	errz.Fatal(err)

	items = []string{}
	for _, e := range entrys {
		items = append(items, e.Name())
	}

	return items, nil
}

// Done does nothing
func (s *s) Done() error {
	return nil
}

func (s *s) ArtifactExists(ctx context.Context, id string) bool {
	return file.Exists(filepath.Join(s.dir, id))
}
