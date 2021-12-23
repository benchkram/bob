package bob

import (
	"github.com/Benchkram/bob/pkg/buildinfostore"
	"github.com/Benchkram/bob/pkg/store"
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

func WithBuildinfoStore(store buildinfostore.Store) Option {
	return func(b *B) {
		b.buildInfoStore = store
	}
}

func WithDisableCache(cache bool) Option {
	return func(b *B) {
		b.disableCache = cache
	}
}
