package remotesyncstore

import storeclient "github.com/benchkram/bob/pkg/store-client"

type Option func(s *S)

func WithClient(client storeclient.I) Option {
	return func(s *S) {
		s.client = client
	}
}
