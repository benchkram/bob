package tasksemanticstest

import (
	"os"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/nix"
	. "github.com/onsi/gomega"
)

func Bob() (*bob.B, error) {
	nixBuilder, err := NixBuilder()
	if err != nil {
		return nil, err
	}
	return bob.Bob(
		bob.WithDir(dir),
		bob.WithCachingEnabled(true),
		bob.WithNixBuilder(nixBuilder),
	)
}

// tmpFiles tracks temporarily created files in these tests
// to be cleaned up at the end.
var tmpFiles []string

func NixBuilder() (*bob.NixBuilder, error) {
	file, err := os.CreateTemp("", ".nix_cache*")
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
func useBobfile(name string) {
	err := os.Rename(name+".yaml", "bob.yaml")
	Expect(err).NotTo(HaveOccurred())
}

// releaseBobfile will revert changes done in useBobfile
func releaseBobfile(name string) {
	err := os.Rename("bob.yaml", name+".yaml")
	Expect(err).NotTo(HaveOccurred())
}
