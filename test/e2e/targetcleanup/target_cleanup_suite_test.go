package targetcleanuptest

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/bob/bobfile"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	dir     string
	version string

	ctx context.Context
)

var _ = BeforeSuite(func() {
	ctx = context.Background()

	version = bob.Version
	bob.Version = "1.0.0"

	// Initialize mock bob files from local directory
	bobFiles := []string{
		"with_dir_target",
	}
	nameToBobfile := make(map[string]*bobfile.Bobfile)
	for _, name := range bobFiles {
		abs, err := filepath.Abs("./" + name)
		Expect(err).NotTo(HaveOccurred())
		bf, err := bobfile.BobfileRead(abs)
		Expect(err).NotTo(HaveOccurred())
		nameToBobfile[strings.ReplaceAll(name, "/", "_")] = bf
	}

	testDir, err := ioutil.TempDir("", "bob-test-target-cleanup-*")
	Expect(err).NotTo(HaveOccurred())
	dir = testDir

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

	bob.Version = version
})

func TestBuild(t *testing.T) {
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
	RunSpecs(t, "nix suite")
}

// useBobfile sets the right bobfile to be used for test
func useBobfile(name string) {
	err := os.Rename(name+".yaml", "bob.yaml")
	Expect(err).NotTo(HaveOccurred())
}

// releaseBobfile will revert changes done in useBobfile
func releaseBobfile(name string) {
	err := os.Rename("bob.yaml", name+".yaml")
	Expect(err).NotTo(HaveOccurred())
}

// readDir returns the dir entries as string array
func readDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []string{}, err
	}

	var contents []string
	for _, e := range entries {
		contents = append(contents, e.Name())
	}
	return contents, nil
}
