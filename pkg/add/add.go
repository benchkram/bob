package add

import (
	"github.com/benchkram/bob/bob"
	"github.com/benchkram/errz"
)

type AddParams struct {
	repoUrl string
	plain   bool
}

type Option func(a *AddParams)

func WithPlainProtocol(explicit bool) Option {
	return func(a *AddParams) {
		a.plain = explicit
	}
}

func Add(repoURL string, opts ...Option) (err error) {
	defer errz.Recover(&err)

	bob, err := bob.Bob(bob.WithRequireBobConfig())
	errz.Fatal(err)

	params := &AddParams{
		repoUrl: repoURL,
		plain:   false,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(params)
	}

	err = bob.Add(params.repoUrl, params.plain)
	errz.Fatal(err)

	return nil
}
