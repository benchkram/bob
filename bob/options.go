package bob

import (
	"github.com/benchkram/bob/pkg/auth"
	"github.com/benchkram/bob/pkg/buildinfostore"
	"github.com/benchkram/bob/pkg/envutil"
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

func WithAuthStore(store *auth.Store) Option {
	return func(b *B) {
		b.authStore = store
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

func WithEnvStore(store envutil.Store) Option {
	return func(b *B) {
		b.envStore = store
	}
}

func WithCachingEnabled(enabled bool) Option {
	return func(b *B) {
		b.enableCaching = enabled
	}
}

func WithPushEnabled(enabled bool) Option {
	return func(b *B) {
		b.enablePush = enabled
	}
}

func WithPullEnabled(enabled bool) Option {
	return func(b *B) {
		b.enablePull = enabled
	}
}

func WithInsecure(allow bool) Option {
	return func(b *B) {
		b.allowInsecure = allow
	}
}

func WithNixBuilder(nix *NixBuilder) Option {
	return func(b *B) {
		b.nix = nix
	}
}

func WithEnvVariables(env []string) Option {
	return func(b *B) {
		b.env = env
	}
}

func WithMaxParallel(maxParallel int) Option {
	return func(b *B) {
		b.maxParallel = maxParallel
	}
}
