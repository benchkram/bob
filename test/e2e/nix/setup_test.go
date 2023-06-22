package nixtest

import (
	"github.com/benchkram/bob/bob"
)

func Bob() (*bob.B, error) {
	return bob.Bob(
		bob.WithDir(dir),
		bob.WithCachingEnabled(false),
	)
}
