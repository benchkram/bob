package clonetest

import (
	"os"
	"testing"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/test/setup/reposetup"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	dir string

	top            string
	childs         []string
	recursiveRepo  string
	playgroundRepo string

	b *bob.B
)

var _ = BeforeSuite(func() {
	testDir, err := os.MkdirTemp("", "bob-test-clone-*")
	Expect(err).NotTo(HaveOccurred())
	dir = testDir

	t, cs, recursive, playground, err := reposetup.BaseTestStructure(dir)
	Expect(err).NotTo(HaveOccurred())
	top = t
	childs = cs
	recursiveRepo = recursive
	playgroundRepo = playground

	err = os.Chdir(top)
	Expect(err).NotTo(HaveOccurred())

	b, err = bob.Bob(bob.WithDir(top))
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := os.RemoveAll(dir)
	Expect(err).NotTo(HaveOccurred())
})

func TestClone(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "clone suite")
}
