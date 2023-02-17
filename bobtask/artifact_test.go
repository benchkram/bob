package bobtask

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benchkram/bob/pkg/buildinfostore"
	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/bob/pkg/store/filestore"
	"github.com/benchkram/errz"
	"github.com/stretchr/testify/assert"
)

func TestArtifactCreateAndExtract(t *testing.T) {
	testdir, err := os.MkdirTemp("", "test-pack-and-unpack-artifact")
	assert.Nil(t, err)
	storage, err := os.MkdirTemp("", "test-pack-and-unpack-artifact-store")
	assert.Nil(t, err)
	buildinfoStorage, err := os.MkdirTemp("", "test-pack-and-unpack-buildinfo-store")
	assert.Nil(t, err)
	defer func() {
		os.RemoveAll(testdir)
		os.RemoveAll(storage)
		os.RemoveAll(buildinfoStorage)
	}()

	artifactStore := filestore.New(storage)
	buildinfoStore := buildinfostore.NewProtoStore(buildinfoStorage)

	assert.Nil(t, os.MkdirAll(filepath.Join(testdir, ".bbuild"), 0774))
	assert.Nil(t, os.MkdirAll(filepath.Join(testdir, ".bbuild/dirone"), 0774))
	assert.Nil(t, os.WriteFile(filepath.Join(testdir, ".bbuild/dirone/fileone"), []byte("fileone"), 0774))
	assert.Nil(t, os.WriteFile(filepath.Join(testdir, ".bbuild/dirone/filetwo"), []byte("filetwo"), 0774))
	assert.Nil(t, os.MkdirAll(filepath.Join(testdir, ".bbuild/dirtwo"), 0774))
	assert.Nil(t, os.WriteFile(filepath.Join(testdir, ".bbuild/dirtwo/fileone"), []byte("fileone"), 0774))
	assert.Nil(t, os.WriteFile(filepath.Join(testdir, ".bbuild/dirtwo/filetwo"), []byte("filetwo"), 0774))

	tsk := Make()
	tsk.dir = testdir
	tsk.local = artifactStore
	tsk.buildInfoStore = buildinfoStore
	tsk.name = "mytaskname"

	tsk.TargetDirty = ".bbuild/dirone/"
	err = tsk.parseTargets()
	assert.Nil(t, err)

	err = tsk.ArtifactCreate("aaa")
	errz.Log(err)
	assert.Nil(t, err)

	err = os.RemoveAll(filepath.Join(testdir, ".build/dirone"))
	assert.Nil(t, err)

	success, err := tsk.ArtifactExtract("aaa")
	assert.Nil(t, err)
	assert.True(t, success)

	assert.True(t, file.Exists(filepath.Join(testdir, ".bbuild/dirone/fileone")))
	assert.True(t, file.Exists(filepath.Join(testdir, ".bbuild/dirone/filetwo")))

	// assure artifact inspect returns without an error
	_, err = tsk.ArtifactInspect("aaa")
	assert.Nil(t, err)
}
