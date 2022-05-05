package nix_test

import (
	"github.com/benchkram/bob/bob/bobfile"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benchkram/bob/bob"

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

	// Initialize mock bob files from local directory
	bobFiles := []string{
		"with_use_nix_false",
		"with_bob_dependencies",
		"with_task_dependencies",
		"with_ambiguous_deps_in_root",
		"with_ambiguous_deps_in_task",
		"with_second_level",
		"with_second_level/second_level",
	}
	nameToBobfile := make(map[string]*bobfile.Bobfile)
	for _, name := range bobFiles {
		abs, err := filepath.Abs("./" + name)
		Expect(err).NotTo(HaveOccurred())
		bf, err := bobfile.BobfileRead(abs)
		Expect(err).NotTo(HaveOccurred())
		nameToBobfile[strings.ReplaceAll(name, "/", "_")] = bf
	}

	testDir, err := ioutil.TempDir("", "bob-test-nix-*")
	Expect(err).NotTo(HaveOccurred())
	dir = testDir
	err = os.Mkdir(dir+"/second_level", 0700)
	Expect(err).NotTo(HaveOccurred())

	err = os.Chdir(dir)
	Expect(err).NotTo(HaveOccurred())

	// Save bob files in dir to have them available in tests
	for name, bf := range nameToBobfile {
		err = bf.BobfileSave(dir, name+".yaml")
		Expect(err).NotTo(HaveOccurred())
	}

	b, err = bob.Bob(bob.WithDir(dir), bob.WithCachingEnabled(false))
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := os.RemoveAll(dir)
	Expect(err).NotTo(HaveOccurred())

	bob.Version = version
})

func TestBuild(t *testing.T) {
	_, err := exec.LookPath("nix")
	if err != nil {
		t.Skip("Test skipped because nix is not installed on your system")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "nix suite")
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
