package addtest

import (
	"os"
	"testing"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/test/setup"
	"github.com/benchkram/bob/test/setup/reposetup"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	dir    string
	childs []string

	b *bob.B

	cleanup func() error
)

var _ = BeforeSuite(func() {
	var err error
	var storageDir string

	dir, storageDir, cleanup, err = setup.TestDirs("add")
	Expect(err).NotTo(HaveOccurred())

	err = os.Chdir(dir)
	Expect(err).NotTo(HaveOccurred())

	top, cs, _, _, err := reposetup.BaseTestStructure(dir)
	Expect(err).NotTo(HaveOccurred())
	childs = cs

	b, err = bob.BobWithBaseStoreDir(storageDir, bob.WithDir(top))
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := cleanup()
	Expect(err).NotTo(HaveOccurred())
})

func TestAdd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "add suite")
}
