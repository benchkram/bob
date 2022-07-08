package bobtask

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/pkg/store"
	"github.com/benchkram/errz"
)

type s struct {
	dir string
}

// NewArtifactStore creates a filestore. The caller is responsible to pass a
// existing directory.
func NewArtifactStore(dir string) store.Store {
	s := &s{
		dir: dir,
	}

	return s
}

// NewArtifact creates a new file. The caller is responsible to call Close().
// Existing artifacts are overwritten.
func (s *s) NewArtifact(_ context.Context, artifactID string) (io.WriteCloser, error) {
	return os.Create(filepath.Join(s.dir, artifactID))
}

// GetArtifact opens a file
func (s *s) GetArtifact(_ context.Context, id string) (empty io.ReadCloser, _ error) {
	return os.Open(filepath.Join(s.dir, id))
}

// Clean deletes all artifacts from store for project. If project is empty then
// it deletes only artifacts belonging to that project
func (s *s) Clean(ctx context.Context, project string) (err error) {
	defer errz.Recover(&err)

	homeDir, err := os.UserHomeDir()
	errz.Fatal(err)
	if s.dir == "/" || s.dir == homeDir {
		return fmt.Errorf("Cleanup of %s is not allowed", s.dir)
	}

	entries, err := os.ReadDir(s.dir)
	errz.Fatal(err)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if project == "" {
			_ = os.Remove(filepath.Join(s.dir, entry.Name()))
			continue
		}

		artifact, err := s.GetArtifact(ctx, entry.Name())
		errz.Fatal(err)
		defer artifact.Close()

		ai, err := ArtifactInspectFromReader(artifact)
		errz.Fatal(err)

		m := ai.Metadata()
		if m == nil {
			continue
		}

		if m.Project == project {
			err = os.Remove(filepath.Join(s.dir, entry.Name()))
			errz.Fatal(err)
		}
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
