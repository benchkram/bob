package version_test

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/benchkram/bob/bob"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	dir     string
	version string

	stdout *os.File
	stderr *os.File
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

	b, err = bob.Bob(bob.WithDir(dir), bob.WithCachingEnabled(false))
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := os.RemoveAll(dir)
	Expect(err).NotTo(HaveOccurred())

	bob.Version = version
})

func TestStatus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "version suite")
}

func capture() {
	stdout = os.Stdout
	stderr = os.Stderr

	var err error
	pr, pw, err = os.Pipe()
	Expect(err).NotTo(HaveOccurred())

	os.Stdout = pw
	os.Stderr = pw
}

func output() string {
	pw.Close()

	b, err := io.ReadAll(pr)
	Expect(err).NotTo(HaveOccurred())

	pr.Close()

	os.Stdout = stdout
	os.Stderr = stderr

	return string(b)
}
