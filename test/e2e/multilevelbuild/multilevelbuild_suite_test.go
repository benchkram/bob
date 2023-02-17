package multilevelbuildtest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/test/setup"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	dir         string
	artifactDir string

	cleanup func() error

	b *bob.B
)

var _ = BeforeSuite(func() {
	var err error
	var storageDir string
	dir, storageDir, cleanup, err = setup.TestDirs("multilevel-build")
	Expect(err).NotTo(HaveOccurred())
	artifactDir = filepath.Join(storageDir, global.BobCacheArtifactsDir)

	err = os.Chdir(dir)
	Expect(err).NotTo(HaveOccurred())

	b, err = bob.BobWithBaseStoreDir(storageDir, bob.WithDir(dir))
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := cleanup()
	Expect(err).NotTo(HaveOccurred())
})

func TestStatus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "multilevel-build suite")
}

// artifactsClean deletes all artifacts from the store
func artifactsClean() error {
	fs, err := os.ReadDir(artifactDir)
	if err != nil {
		return err
	}
	for _, f := range fs {
		err = os.Remove(filepath.Join(artifactDir, f.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}
