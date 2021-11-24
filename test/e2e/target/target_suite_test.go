package targettest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/bob/global"
	"github.com/Benchkram/bob/pkg/buildinfostore"
	"github.com/Benchkram/bob/test/setup"

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
	var err error
	var storageDir string
	dir, storageDir, cleanup, err = setup.TestDirs("target")
	Expect(err).NotTo(HaveOccurred())

	err = os.Chdir(dir)
	Expect(err).NotTo(HaveOccurred())

	buildinfoStore = buildinfostore.New(filepath.Join(storageDir, global.BobCacheBuildinfoDir))
	b, err = bob.BobWithBaseStoreDir(
		storageDir,
		bob.WithBuildinfoStore(buildinfoStore),
		bob.WithDir(dir),
	)

	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := cleanup()
	Expect(err).NotTo(HaveOccurred())
})

func TestStatus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "target-build suite")
}
