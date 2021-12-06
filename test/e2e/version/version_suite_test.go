package version_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/Benchkram/bob/bob"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	dir     string
	version string

	stdout *os.File
	pr     *os.File
	pw     *os.File

	b *bob.B
)

var _ = BeforeSuite(func() {
	version = bob.Version
	bob.Version = "1.0.0"

	testDir, err := ioutil.TempDir("", "bob-test-version-*")
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

	bob.Version = version
})

func TestStatus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "export-build suite")
}

func capture() {
	stdout = os.Stdout

	var err error
	pr, pw, err = os.Pipe()
	Expect(err).NotTo(HaveOccurred())

	os.Stdout = pw
}

func output() string {
	pw.Close()

	b, err := ioutil.ReadAll(pr)
	Expect(err).NotTo(HaveOccurred())

	os.Stdout = stdout

	return string(b)
}
