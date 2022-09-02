package targettest

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/buildinfostore"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/test/setup"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	dir string

	cleanup func() error

	buildinfoStore buildinfostore.Store

	b *bob.B
)

var _ = BeforeSuite(func() {
	boblog.SetLogLevel(10)
	var err error
	var storageDir string
	dir, storageDir, cleanup, err = setup.TestDirs("target")
	Expect(err).NotTo(HaveOccurred())

	err = os.Chdir(dir)
	Expect(err).NotTo(HaveOccurred())

	buildinfoStore = buildinfostore.New(filepath.Join(storageDir, global.BobCacheBuildinfoDir))

	nixBuilder, err := NixBuilder()
	Expect(err).NotTo(HaveOccurred())

	b, err = bob.BobWithBaseStoreDir(
		storageDir,
		bob.WithBuildinfoStore(buildinfoStore),
		bob.WithDir(dir),
		bob.WithNixBuilder(nixBuilder),
	)

	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	for _, file := range tmpFiles {
		err := os.Remove(file)
		Expect(err).NotTo(HaveOccurred())
	}
	err := cleanup()
	Expect(err).NotTo(HaveOccurred())
})

func TestStatus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "target-build suite")
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
