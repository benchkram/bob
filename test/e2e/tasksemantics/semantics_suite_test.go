package tasksemanticstest

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benchkram/bob/bob/bobfile"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	dir string

	stdout *os.File
	stderr *os.File
	pr     *os.File
	pw     *os.File
)

var _ = BeforeSuite(func() {
	// Initialize mock bob files from local directory
	bobFiles := []string{
		"rebuild_on_input_change",
		"no_input_with_target",
		"no_input_unknown_target",
		"no_input_no_target",
		"with_input_no_target",
		"no_input_no_target_rebuild_always",
		"compound_task",
		"rebuild_always_with_target",
	}
	nameToBobfile := make(map[string]*bobfile.Bobfile)
	for _, name := range bobFiles {
		abs, err := filepath.Abs("./" + name)
		Expect(err).NotTo(HaveOccurred())
		bf, err := bobfile.BobfileRead(abs)
		Expect(err).NotTo(HaveOccurred())
		nameToBobfile[strings.ReplaceAll(name, "/", "_")] = bf
	}

	testDir, err := ioutil.TempDir("", "bob-test-task-semantics-*")
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
	RunSpecs(t, "task semantics suite")
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
