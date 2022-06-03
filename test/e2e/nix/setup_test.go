package nixtest

import (
	"io/ioutil"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/nix"

	. "github.com/onsi/gomega"
)

func Bob() (*bob.B, error) {
	return bob.Bob(
		bob.WithDir(dir),
		bob.WithCachingEnabled(false),
		bob.WithNixBuilder(MustNixBuilder()),
	)
}

func NixBuilder() (*bob.NixBuilder, error) {
	file, err := ioutil.TempFile("", ".nix_cache*")
	if err != nil {
		return nil, err
	}
	name := file.Name()
	file.Close()

	cache, err := nix.NewCacheStore(nix.WithPath(name))
	if err != nil {
		return nil, err
	}

	return bob.NewNixBuilder(bob.WithCache(cache)), nil
}

func MustNixBuilder() *bob.NixBuilder {
	n, err := NixBuilder()
	Expect(err).NotTo(HaveOccurred())
	return n
}
