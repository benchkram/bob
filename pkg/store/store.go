package store

import (
	"context"
	"io"
)

// get inspiration from https://github.com/tus/tusd/blob/48ffebec56fcf3221461b3f8cbe000e5367e2d48/pkg/handler/datastore.go#L50

type Artifact interface {
	io.ReadWriteCloser
}

type Store interface {
	NewArtifact(_ context.Context, id string) (Artifact, error)
	GetArtifact(_ context.Context, id string) (Artifact, error)

	List(context.Context) ([]string, error)

	Clean(context.Context) error
}
