package add

import (
	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/errz"
)

type AddParams struct {
	repoUrl  string
	explicit bool
}

type Option func(a *AddParams)

func WithExplicitProtocol(explicit bool) Option {
	return func(a *AddParams) {
		a.explicit = explicit
	}
}

func Add(repoURL string, opts ...Option) (err error) {
	defer errz.Recover(&err)

	bob, err := bob.Bob(bob.WithRequireBobConfig())
	errz.Fatal(err)

	params := &AddParams{
		repoUrl:  repoURL,
		explicit: false,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(params)
	}

	err = bob.Add(params.repoUrl, params.explicit)
	errz.Fatal(err)

	return nil
}
