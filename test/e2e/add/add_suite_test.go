package addtest

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/Benchkram/bob/bob"
	"github.com/Benchkram/bob/test/repo/setup"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	dir string

	childs []string

	b *bob.B
)

var _ = BeforeSuite(func() {
	testDir, err := ioutil.TempDir("", "bob-test-add-*")
	Expect(err).NotTo(HaveOccurred())
	dir = testDir

	err = os.Chdir(dir)
	Expect(err).NotTo(HaveOccurred())

	top, cs, _, _, err := setup.BaseTestStructure(dir)
	Expect(err).NotTo(HaveOccurred())
	childs = cs

	b, err = bob.Bob(bob.WithDir(top))
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := os.RemoveAll(dir)
	Expect(err).NotTo(HaveOccurred())
})

func TestAdd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "add suite")
}
