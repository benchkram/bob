package inittest

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/benchkram/bob/bob"
)

var (
	dir string

	b *bob.B
)

var _ = BeforeSuite(func() {
	testDir, err := os.MkdirTemp("", "bob-test-init-*")
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

func TestInit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "init suite")
}
