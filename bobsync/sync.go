package bobsync

import "context"

type Sync struct {
	name string

	Path string `yaml:"path"`

	Version string `yaml:"version"`
}

func (s *Sync) Push(ctx context.Context) (err error) {
	return nil
}

func (s *Sync) Pull(ctx context.Context) (err error) {
	return nil
}

func (s *Sync) ListLocal(ctx context.Context) (err error) {
	return nil
}

func (s *Sync) ListRemote(ctx context.Context) (err error) {
	return nil
}
