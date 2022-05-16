package bob

import (
	"github.com/benchkram/bob/pkg/buildinfostore"
	"github.com/benchkram/bob/pkg/store"
)

type Option func(b *B)

func WithDir(dir string) Option {
	return func(b *B) {
		b.dir = dir
	}
}

// WithRequireBobConfig forces bob to read the configuration
// from `.bob/config`. Currently only used by `bob clone` and `bob git ...
func WithRequireBobConfig() Option {
	return func(b *B) {
		b.readConfig = true
	}
}

func WithFilestore(store store.Store) Option {
	return func(b *B) {
		b.local = store
	}
}

func WithRemotestore(store store.Store) Option {
	return func(b *B) {
		b.remote = store
	}
}

func WithBuildinfoStore(store buildinfostore.Store) Option {
	return func(b *B) {
		b.buildInfoStore = store
	}
}

func WithCachingEnabled(enabled bool) Option {
	return func(b *B) {
		b.enableCaching = enabled
	}
}

func WithInsecure(allow bool) Option {
	return func(b *B) {
		b.allowInsecure = allow
	}
}
