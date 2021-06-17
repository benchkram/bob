package add

import (
	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/errz"
)

func Add(repoURL string) (err error) {
	defer errz.Recover(&err)

	bob, err := bob.Bob(bob.WithRequireBobConfig())
	errz.Fatal(err)

	err = bob.Add(repoURL)
	errz.Fatal(err)

	return nil
}
