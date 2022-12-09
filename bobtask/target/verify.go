package target

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/filehash"
)

// Hint: comparing the modification time is tricky as a artifact extraction
// from a tar archive changes the modification time of a file.

// VerifyShallow compare targets against a existing
// buildinfo. It will only check if the size of any of the files changed.
// Docker targets are verified similarly as in plain verify
// as there is no performance penality.
// In case the expect buildinfo does not exist Verify checks against filesystemEntriesRaw.
func (t *T) VerifyShallow() bool {
	return t.verifyFilesystemShallow() && t.verifyDocker()
}

// Verify existence and integrity of targets against an expected buildinfo.
// In case the expected buildinfo does not exist Verify checks against filesystemEntriesRaw.
//
// Verify returns true when no targets are defined.
// Verify returns when there is nothing to compare against.
func (t *T) Verify() bool {
	return t.verifyFilesystem() && t.verifyDocker()
}

func (t *T) preConditionsFilesystem() bool {
	if len(*t.filesystemEntries) == 0 && len(t.filesystemEntriesRaw) == 0 {
		return true
	}

	// In case there was no previous local build
	// verify returns false indicating that there can't
	// exist a valid target from a previous build.
	// Loading from the cache must be handled by the calling function.
	if t.expected == nil {
		return false
	}

	// This usually indicates a file was added/removed manually
	// from a target directory.
	if len(t.expected.Filesystem.Files) != len(*t.filesystemEntries) {
		return false
	}

	return true
}

func (t *T) verifyFilesystemShallow() bool {

	if !t.preConditionsFilesystem() {
		return false
	}

	for _, path := range *t.filesystemEntries {

		fileInfo, err := os.Lstat(path)
		if err != nil {
			return false
		}

		expectedFileInfo, ok := t.expected.Filesystem.Files[path]
		if !ok {
			return false
		}

		// A shallow verify compares the size of the target
		if fileInfo.Size() != expectedFileInfo.Size {
			boblog.Log.V(2).Info(fmt.Sprintf("failed to verify [%s], different sizes [current: %d != expected: %d]",
				path, fileInfo.Size(), expectedFileInfo.Size))
			return false
		}
	}

	return true
}

func (t *T) verifyFilesystem() bool {

	if !t.preConditionsFilesystem() {
		return false
	}

	h := filehash.New()

	for _, path := range *t.filesystemEntries {

		fileInfo, err := os.Stat(path)
		if err != nil {
			return false
		}

		expectedFileInfo, ok := t.expected.Filesystem.Files[path]
		if !ok {

			return false
		}

		// Compare size of the target
		if fileInfo.Size() != expectedFileInfo.Size {
			return false
		}

		err = h.AddFile(path)
		if err != nil {
			return false
		}
	}

	ret := hex.EncodeToString(h.Sum()) == t.expected.Filesystem.Hash

	return ret
}

func (t *T) verifyDocker() bool {
	if len(t.dockerImages) == 0 {
		return true
	}

	// In case there was no previous local build
	// verify return false indicating that there can't
	// exist a valid target from a previous build.
	// Loading from the cash must be handled by the calling function.
	if t.expected == nil {
		return false
	}

	// This usually indicates an image was added/removed manually
	// from a target directory.
	if len(t.expected.Docker) != len(t.dockerImages) {
		return false
	}

	for _, image := range t.dockerImages {
		expectedImageInfo, ok := t.expected.Docker[image]
		if !ok {
			return false
		}

		exists, err := t.dockerRegistryClient.ImageExists(image)
		if err != nil {
			return false
		}
		if !exists {
			return false
		}

		imageHash, err := t.dockerImageHash(image)
		if err != nil {
			boblog.Log.Error(err, fmt.Sprintf("Unable to verify docker image hash [%s]", image))
			return false
		}
		if imageHash != expectedImageInfo.Hash {
			return false
		}
	}

	return true
}
