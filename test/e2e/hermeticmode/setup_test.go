package hermeticmodetest

import (
	"io/ioutil"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/errz"
)

func BobSetup(env ...string) (_ *bob.B, err error) {
	defer errz.Recover(&err)

	nixBuilder, err := NixBuilder()
	errz.Fatal(err)

	return bob.Bob(
		bob.WithDir(dir),
		bob.WithCachingEnabled(false),
		bob.WithNixBuilder(nixBuilder),
		bob.WithEnvVariables(env),
	)
}

// tmpFiles tracks temporarily created files in these tests
// to be cleaned up at the end.
var tmpFiles []string

func NixBuilder() (_ *bob.NixBuilder, err error) {
	defer errz.Recover(&err)

	file, err := ioutil.TempFile("", ".nix_cache*")
	errz.Fatal(err)
	name := file.Name()
	file.Close()

	tmpFiles = append(tmpFiles, name)

	cache, err := nix.NewCacheStore(nix.WithPath(name))
	errz.Fatal(err)

	return bob.NewNixBuilder(bob.WithCache(cache)), nil
}
