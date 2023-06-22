package inputest

import (
	"github.com/benchkram/bob/bob"
	"github.com/benchkram/errz"
)

func BobSetup(env ...string) (_ *bob.B, err error) {
	defer errz.Recover(&err)

	return bob.Bob(
		bob.WithDir(dir),
		bob.WithCachingEnabled(false),
		bob.WithEnvVariables(env),
	)
}
