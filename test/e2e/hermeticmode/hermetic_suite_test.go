package hermeticmodetest

import (
	"bufio"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/pkg/file"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	dir     string
	version string
)

type project struct {
	name  string
	gomod string
	// .gox extension to avoid `main redeclared in this block`
	main string
}

var _ = BeforeSuite(func() {
	version = bob.Version
	bob.Version = "1.0.0"

	// Initialize mock bob files from local directory
	bobFiles := []string{
		"build_with_use_nix_false",
		"build_with_use_nix_true",
		"init_with_use_nix_false",
		"init_with_use_nix_true",
		"init_once_with_use_nix_false",
		"init_once_with_use_nix_true",
		"binary_with_use_nix_false",
		"binary_with_use_nix_true",
	}
	nameToBobfile := make(map[string]*bobfile.Bobfile)
	for _, name := range bobFiles {
		abs, err := filepath.Abs("./" + name)
		Expect(err).NotTo(HaveOccurred())
		bf, err := bobfile.BobfileRead(abs)
		Expect(err).NotTo(HaveOccurred())
		nameToBobfile[strings.ReplaceAll(name, "/", "_")] = bf
	}

	testDir, err := ioutil.TempDir("", "bob-test-hermetic-mode-*")
	Expect(err).NotTo(HaveOccurred())

	projects := []project{
		{
			name:  "server",
			gomod: "./server/go.mod",
			main:  "./server/main.go",
		},
		{
			name:  "server-with-env",
			gomod: "./server-with-env/go.mod",
			main:  "./server-with-env/main.go",
		},
	}

	// copy projects in the test dir
	for _, p := range projects {
		err = file.Copy(p.gomod, testDir+"/"+p.name+"_go.mod")
		Expect(err).NotTo(HaveOccurred())

		err = file.Copy(p.main, testDir+"/"+p.name+"_main.gox")
		Expect(err).NotTo(HaveOccurred())
	}

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

func TestRun(t *testing.T) {
	_, err := exec.LookPath("nix")
	if err != nil {
		// Allow skipping tests only locally.
		// CI is always set to true on GitHub actions.
		// https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables
		if os.Getenv("CI") != "true" {
			t.Skip("Test skipped because nix is not installed on your system")
		}
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "hermetic mode suite")
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

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// useProject set up the project to be used
// make sure to call releaseProject to revert changes
func useProject(name string) {
	err := os.Rename(name+"_go.mod", "go.mod")
	Expect(err).NotTo(HaveOccurred())

	err = os.Rename(name+"_main.gox", "main.go")
	Expect(err).NotTo(HaveOccurred())
}

// releaseProject will revert the changes done in useProject
func releaseProject(name string) {
	err := os.Rename("go.mod", name+"_go.mod")
	Expect(err).NotTo(HaveOccurred())

	err = os.Rename("main.go", name+"_main.gox")
	Expect(err).NotTo(HaveOccurred())
}

func assertKeyHasValue(key, value string, env []string) {
	for _, v := range env {
		pair := strings.SplitN(v, "=", 2)
		if pair[0] == key {
			Expect(pair[1]).To(Equal(value))
		}
	}
}
