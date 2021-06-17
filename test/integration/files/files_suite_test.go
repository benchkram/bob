package filestest

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/Benchkram/bob/test/repo/setup"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	dir string

	files []string
)

var _ = BeforeSuite(func() {
	testDir, err := ioutil.TempDir("", "bob-test-files-*")
	Expect(err).NotTo(HaveOccurred())
	dir = testDir

	files = setup.SetupBaseFileStructure(dir)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := os.RemoveAll(dir)
	Expect(err).NotTo(HaveOccurred())
})

func TestFiles(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "files suite")
}
