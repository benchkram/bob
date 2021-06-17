package buildtest

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/Benchkram/bob/bob"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	dir string

	b *bob.B
)

var _ = BeforeSuite(func() {
	testDir, err := ioutil.TempDir("", "bob-test-build-*")
	Expect(err).NotTo(HaveOccurred())
	dir = testDir

	err = os.Chdir(dir)
	Expect(err).NotTo(HaveOccurred())

	b, err = bob.Bob(bob.WithDir(dir))
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := os.RemoveAll(dir)
	Expect(err).NotTo(HaveOccurred())
})

func TestStatus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "build suite")
}
