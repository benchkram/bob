package artifactstest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/buildinfostore"
	"github.com/benchkram/bob/pkg/store"
	"github.com/benchkram/bob/test/setup"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Test case overview for target invalidation with artifacts in the local store.
// Input change not included, should not change for those tests.
//
// dne = does not exist
//
//    BUILDINFO          TARGET       LOCAL_ARTIFACT

// Following cases are tested in nobuildinfo_test.go
// 1  dne                unchanged    dne                  | 0 0 0 |     =>   rebuild (buildinfostore cleaned?)
// 2  dne                unchanged    exists               | 0 0 1 |     =>   no-rebuild-required (update target from artifact)
// 3  dne                changed/dne  dne                  | 0 1 0 |     =>   rebuild (buildinfostore cleaned?)
// 4  dne                changed/dne  exists               | 0 1 1 |     =>   no-rebuild-required (update target from artifact)
//
// Following cases are tested in artifact_test.go
// 5  exists             unchanged    dne                  | 1 0 0 |     =>   rebuild-required (to assure the target is correctly pushed to the local store)
// 6  exists             unchanged    exists               | 1 0 1 |     =>   no-rebuild-required
// 7  exists             changed      dne                  | 1 1 0 |     =>   rebuild
// 8  exists             changed      exists               | 1 1 1 |     =>   no-rebuild-required (update target from artifact)
//

var (
	dir         string
	artifactDir string

	artifactStore  store.Store
	buildinfoStore buildinfostore.Store

	cleanup func() error

	b *bob.B

	bNoCache *bob.B
)

// reset base test dir to it's
// initial state.
func reset() error {
	err := os.RemoveAll(dir)
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0777)
}

var _ = BeforeSuite(func() {
	boblog.SetLogLevel(10)

	var err error
	var storageDir string
	dir, storageDir, cleanup, err = setup.TestDirs("artifacts")
	Expect(err).NotTo(HaveOccurred())
	artifactDir = filepath.Join(storageDir, global.BobCacheArtifactsDir)

	err = os.Chdir(dir)
	Expect(err).NotTo(HaveOccurred())

	// objects using the local filesystem for caching and persistent.
	// Initialized on a temporary location to avoid interference with
	// a local bob installation.
	artifactStore, err = bob.Filestore(storageDir)
	Expect(err).NotTo(HaveOccurred())
	buildinfoStore, err = bob.BuildinfoStore(storageDir)
	Expect(err).NotTo(HaveOccurred())
	nixBuilder, err := bob.NixBuilder(storageDir)
	Expect(err).NotTo(HaveOccurred())

	b, err = bob.Bob(
		bob.WithDir(dir),
		bob.WithFilestore(artifactStore),
		bob.WithBuildinfoStore(buildinfoStore),
		bob.WithNixBuilder(nixBuilder),
	)
	Expect(err).NotTo(HaveOccurred())

	bNoCache, err = bob.BobWithBaseStoreDir(
		storageDir,
		bob.WithDir(dir),
		bob.WithCachingEnabled(false),
		bob.WithNixBuilder(nixBuilder),
	)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := cleanup()
	Expect(err).NotTo(HaveOccurred())
})

func TestArtifact(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "artifacts suite")
}
