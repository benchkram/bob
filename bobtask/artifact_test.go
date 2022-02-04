package bobtask

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Benchkram/bob/pkg/file"
	"github.com/Benchkram/bob/pkg/store/filestore"
	"github.com/Benchkram/errz"
	"github.com/stretchr/testify/assert"
)

func TestPackAndUnpackArtifacts(t *testing.T) {

	testdir, err := ioutil.TempDir("", "test-pack-and-unpack-artifact")
	assert.Nil(t, err)
	storage, err := ioutil.TempDir("", "test-pack-and-unpack-artifact-store")
	assert.Nil(t, err)
	defer func() {
		os.RemoveAll(testdir)
		os.RemoveAll(storage)
	}()

	artifactStore := filestore.New(storage)

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
	tsk.name = "mytaskname"

	tsk.TargetDirty = ".bbuild/dirone/"
	err = tsk.parseTargets()
	assert.Nil(t, err)

	err = tsk.ArtifactPack("aaa")
	errz.Log(err)
	assert.Nil(t, err)

	err = os.RemoveAll(filepath.Join(testdir, ".build/dirone"))
	assert.Nil(t, err)

	success, err := tsk.ArtifactUnpack("aaa")
	assert.Nil(t, err)
	assert.True(t, success)

	assert.True(t, file.Exists(filepath.Join(testdir, ".bbuild/dirone/fileone")))
	assert.True(t, file.Exists(filepath.Join(testdir, ".bbuild/dirone/filetwo")))

	// assure artifact inspect returns without an error
	_, err = tsk.ArtifactInspect("aaa")
	assert.Nil(t, err)
}
