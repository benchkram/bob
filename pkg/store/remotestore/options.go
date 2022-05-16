package remotestore

import (
	storeclient "github.com/benchkram/bob/pkg/store-client"
)

type Option func(s *s)

func WithClient(client storeclient.I) Option {
	return func(s *s) {
		s.client = client
	}
}
