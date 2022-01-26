package add

import (
	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/errz"
)

type AddParams struct {
	repoUrl   string
	httpsOnly bool
	sshOnly   bool
}

type Option func(a *AddParams)

func WithHttpsOnly(https bool) Option {
	return func(a *AddParams) {
		a.httpsOnly = https
	}
}

func WithSSHOnly(ssh bool) Option {
	return func(a *AddParams) {
		a.sshOnly = ssh
	}
}

func Add(repoURL string, opts ...Option) (err error) {
	defer errz.Recover(&err)

	bob, err := bob.Bob(bob.WithRequireBobConfig())
	errz.Fatal(err)

	params := &AddParams{
		repoUrl:   repoURL,
		httpsOnly: false,
		sshOnly:   false,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(params)
	}

	err = bob.Add(params.repoUrl, params.httpsOnly, params.sshOnly)
	errz.Fatal(err)

	return nil
}
