package store

import (
	"context"
	"fmt"
	"io"
)

// get inspiration from https://github.com/tus/tusd/blob/48ffebec56fcf3221461b3f8cbe000e5367e2d48/pkg/handler/datastore.go#L50

type Artifact interface {
	io.ReadWriteCloser
}

type Store interface {
	NewArtifact(_ context.Context, artifactID string, size int64) (io.WriteCloser, error)
	GetArtifact(_ context.Context, id string) (io.ReadCloser, int64, error)

	List(context.Context) ([]string, error)

	Clean(context.Context) error

	ArtifactExists(ctx context.Context, id string) bool

	Done() error
}

var (
	ErrArtifactNotFoundinSrc = fmt.Errorf("artifact not found in src")
	ErrArtifactAlreadyExists = fmt.Errorf("artifact already exists")
)
