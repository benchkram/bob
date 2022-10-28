package clonetest

import (
	"io/ioutil"
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
	testDir, err := ioutil.TempDir("", "bob-test-clone-*")
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
	t.Skip("Fails with fatal: unable to access 'https://github.com/pkg/errors.git/': error:16000069:STORE routines::unregistered scheme")
	RegisterFailHandler(Fail)
	RunSpecs(t, "clone suite")
}
