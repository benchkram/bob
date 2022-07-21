package nixruntest

import (
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

	stdout *os.File
	stderr *os.File
	pr     *os.File
	pw     *os.File
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
		"with_use_nix_false",
		"init_with_bob_dependencies",
		"init_with_task_dependencies",
		"init_once_with_bob_dependencies",
		"init_once_with_task_dependencies",
		"binary_with_bob_dependencies",
		"binary_with_task_dependencies",
	}
	nameToBobfile := make(map[string]*bobfile.Bobfile)
	for _, name := range bobFiles {
		abs, err := filepath.Abs("./" + name)
		Expect(err).NotTo(HaveOccurred())
		bf, err := bobfile.BobfileRead(abs)
		Expect(err).NotTo(HaveOccurred())
		nameToBobfile[strings.ReplaceAll(name, "/", "_")] = bf
	}

	testDir, err := ioutil.TempDir("", "bob-test-nix-run-*")
	Expect(err).NotTo(HaveOccurred())

	projects := []project{
		{
			name:  "server",
			gomod: "./server/go.mod",
			main:  "./server/main.go",
		},
		{
			name:  "server-with-dep-inside",
			gomod: "./server-with-dep-inside/go.mod",
			main:  "./server-with-dep-inside/main.go",
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
	err = os.Mkdir(dir+"/second_level", 0700)
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
	RunSpecs(t, "nix run suite")
}

func startScan() {
	stdout = os.Stdout
	stderr = os.Stderr

	var err error
	pr, pw, err = os.Pipe()
	Expect(err).NotTo(HaveOccurred())

	os.Stdout = pw
	os.Stderr = pw
}

func stopScan() {
	os.Stdout = stdout
	os.Stderr = stderr

	pw.Close()
	pr.Close()
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
