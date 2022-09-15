package taskdecorationtest

import (
	"io/ioutil"
	"os"

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

// useBobfile sets the right bobfile to be used for test
func useBobfile(name string) error {
	return os.Rename(name+".yaml", "bob.yaml")
}

// releaseBobfile will revert changes done in useBobfile
func releaseBobfile(name string) error {
	return os.Rename("bob.yaml", name+".yaml")
}

// readDir returns the dir entries as string array
func readDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []string{}, err
	}

	var contents []string
	for _, e := range entries {
		contents = append(contents, e.Name())
	}
	return contents, nil
}
