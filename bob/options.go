package bob

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
