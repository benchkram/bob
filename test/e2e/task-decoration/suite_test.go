package taskdecorationtest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/pkg/buildinfostore"
	"github.com/benchkram/bob/pkg/store"
	"github.com/benchkram/bob/test/setup"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	// dir is the basic test directory
	// in which the test is executed.
	dir string

	// artifactStore temporary store to
	// avoid interfering with the users cache.
	artifactStore store.Store
	// buildInfoStore temporary store
	// to avoid interfering with the users cache.
	buildInfoStore buildinfostore.Store

	// cleanup is called at the end to remove all test files from the system.
	cleanup func() error

	// tmpFiles tracks temporarily created files
	// to be cleaned up at the end.
	tmpFiles []string
)

var _ = BeforeSuite(func() {

	// Initialize mock bob files from local directory
	bobFiles := []string{
		// TODO: add files for test setup
	}
	nameToBobfile := make(map[string]*bobfile.Bobfile)
	for _, name := range bobFiles {
		abs, err := filepath.Abs("./" + name)
		Expect(err).NotTo(HaveOccurred())
		bf, err := bobfile.BobfileRead(abs)
		Expect(err).NotTo(HaveOccurred())
		nameToBobfile[strings.ReplaceAll(name, "/", "_")] = bf
	}

	var err error
	var storageDir string
	dir, storageDir, cleanup, err = setup.TestDirs("task-decoration")
	Expect(err).NotTo(HaveOccurred())

	artifactStore, err = bob.Filestore(storageDir)
	Expect(err).NotTo(HaveOccurred())
	buildInfoStore, err = bob.BuildinfoStore(storageDir)
	Expect(err).NotTo(HaveOccurred())

	err = os.Chdir(dir)
	Expect(err).NotTo(HaveOccurred())

	// Save bob files in dir to have them available in tests
	for name, bf := range nameToBobfile {
		err = bf.BobfileSave(dir, name+".yaml")
		Expect(err).NotTo(HaveOccurred())
	}
})

var _ = AfterSuite(func() {
	err := os.RemoveAll(dir)
	Expect(err).NotTo(HaveOccurred())

	for _, file := range tmpFiles {
		err = os.Remove(file)
		Expect(err).NotTo(HaveOccurred())
	}

	err = cleanup()
	Expect(err).NotTo(HaveOccurred())
})

func TestOverwrite(t *testing.T) {
	_, err := exec.LookPath("nix")
	if err != nil {
		// Allow to skip tests only locally.
		// CI is always set to true on GitHub actions.
		// https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables
		if os.Getenv("CI") != "true" {
			t.Skip("Test skipped because nix is not installed on your system")
		}
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "target cleanup suite")
}
