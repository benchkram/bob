package targetcleanuptest

import (
	"io/ioutil"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/errz"
)

func BobSetup() (_ *bob.B, err error) {
	return bobSetup()
}

func BobSetupNoCache() (_ *bob.B, err error) {
	return bobSetup(bob.WithCachingEnabled(false))
}

func bobSetup(opts ...bob.Option) (_ *bob.B, err error) {
	defer errz.Recover(&err)

	nixBuilder, err := NixBuilder()
	errz.Fatal(err)

	static := []bob.Option{
		bob.WithDir(dir),
		bob.WithNixBuilder(nixBuilder),
		bob.WithFilestore(artifactStore),
		bob.WithBuildinfoStore(buildInfoStore),
	}
	static = append(static, opts...)
	return bob.Bob(
		static...,
	)
}

// tmpFiles tracks temporarily created files in these tests
// to be cleaned up at the end.
var tmpFiles []string

func NixBuilder() (*bob.NixBuilder, error) {
	file, err := ioutil.TempFile("", ".nix_cache*")
	if err != nil {
		return nil, err
	}
	name := file.Name()
	file.Close()

	tmpFiles = append(tmpFiles, name)

	cache, err := nix.NewCacheStore(nix.WithPath(name))
	if err != nil {
		return nil, err
	}

	return bob.NewNixBuilder(bob.WithCache(cache)), nil
}
