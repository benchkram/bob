package targetnametest

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/errz"
	. "github.com/onsi/gomega"
)

func BobSetup() (_ *bob.B, err error) {
	return bobSetup()
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
func useBobfile(name string) {
	err := os.Rename(name+".yaml", "bob.yaml")
	Expect(err).ToNot(HaveOccurred())
}

// releaseBobfile will revert changes done in useBobfile
func releaseBobfile(name string) {
	err := os.Rename("bob.yaml", name+".yaml")
	Expect(err).ToNot(HaveOccurred())
}

func useSecondLevelBobfile(name string) {
	err := os.Rename(name+"_"+secondLevelDir+".yaml", filepath.Join(dir, secondLevelDir, "bob.yaml"))
	Expect(err).ToNot(HaveOccurred())
}

// releaseBobfile will revert changes done in useSecondLevelBobfile
func releaseSecondLevelBobfile(name string) {
	err := os.Rename(
		filepath.Join(dir, secondLevelDir, "bob.yaml"),
		name+"_"+secondLevelDir+".yaml")
	Expect(err).ToNot(HaveOccurred())
}
