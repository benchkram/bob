package inputest

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
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

	nameToBobfile map[string]*bobfile.Bobfile
)

var _ = BeforeSuite(func() {
	ctx = context.Background()

	version = bob.Version
	bob.Version = "1.0.0"

	// Initialize mock bob files from local directory
	bobFiles := []string{
		"with_one_level",
		"with_second_level",
		"with_second_level/second_level",
		"with_three_level",
		"with_three_level/second_level",
		"with_three_level/second_level/third_level",
		"with_same_input_and_target",
		"with_same_input_and_target_relative",
		"with_same_input_and_target_relative_target",
	}
	nameToBobfile = make(map[string]*bobfile.Bobfile)
	for _, name := range bobFiles {
		abs, err := filepath.Abs("./" + name)
		Expect(err).NotTo(HaveOccurred())
		bf, err := bobfile.BobfileRead(abs)
		Expect(err).NotTo(HaveOccurred())
		nameToBobfile[name] = bf
	}
})

var _ = AfterSuite(func() {
	for _, file := range tmpFiles {
		err := os.Remove(file)
		Expect(err).NotTo(HaveOccurred())
	}

	bob.Version = version
})

func TestInput(t *testing.T) {
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
	RunSpecs(t, "input suite")
}
