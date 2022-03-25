package remotestore

type Option func(s *s)

func WithDir(dir string) Option {
	return func(s *s) {
		s.dir = dir
	}
}
