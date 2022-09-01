package target

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/filehash"
)

// Hint: comparing the modification time is tricky as a unpack
// form a tar archive changes the modification time of a file.

// VerifyShallow compare targets against a existing
// buildinfo. It will only check if the size of any of the files changed.
// Docker targets are verified similarly as in plain verify
// as there is no performance penality.
// In case the expect buildinfo does not exist Verify checks against filesystemEntriesRaw.
func (t *T) VerifyShallow() bool {
	return t.verifyFilesystemShallow() && t.verifyDocker()
}

// Verify existence and integrity of targets against a expected buildinfo.
// In case the expect buildinfo does not exist Verify checks against filesystemEntriesRaw.
//
// Verify returns true when no targets are defined.
// Verify returns when there is nothing to compare against.
func (t *T) Verify() bool {
	return t.verifyFilesystem() && t.verifyDocker()
}

func (t *T) preConditionsFilesystem() bool {
	if len(*t.filesystemEntries) == 0 && len(t.filesystemEntriesRaw) == 0 {
		boblog.Log.V(2).Info("BBBBBBB")
		return true
	}

	// In case there was no previous local build
	// verify returns false indicating that there can't
	// exist a valid target from a previous build.
	// Loading from the cash must be handled by the calling function.
	if t.expected == nil {
		boblog.Log.V(2).Info("CCCCCCCCC")
		return false
	}

	_ = t.Resolve()
	// This usually indicates a file was added/removed manually
	// from a target directory.
	if len(t.expected.Filesystem.Files) != len(*t.filesystemEntries) {
		boblog.Log.V(2).Info(fmt.Sprintf("DDDDDDDD [%d/%d]", len(t.expected.Filesystem.Files), len(*t.filesystemEntries)))
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

	boblog.Log.V(2).Info("11111111111")

	if !t.preConditionsFilesystem() {
		return false
	}

	boblog.Log.V(2).Info("222222222222")
	h := filehash.New()

	for _, path := range *t.filesystemEntries {

		fileInfo, err := os.Stat(path)
		if err != nil {
			return false
		}

		expectedFileInfo, ok := t.expected.Filesystem.Files[path]
		if !ok {
			boblog.Log.V(2).Info("33333333333")

			return false
		}

		// Compare size of the target
		if fileInfo.Size() != expectedFileInfo.Size {
			boblog.Log.V(2).Info("777777")
			return false
		}

		err = h.AddFile(path)
		if err != nil {
			boblog.Log.V(2).Info("55555")
			return false
		}
	}

	boblog.Log.V(2).Info("6666666")

	if t.expected == nil {
		boblog.Log.V(2).Info("nilnilnil")
	}

	ret := hex.EncodeToString(h.Sum()) == t.expected.Filesystem.Hash

	boblog.Log.V(2).Info("7777")

	return ret
}

// func (t *T) verifyFile(groundTruth string) bool {
// 	if len(t.PathsSerialize) == 0 {
// 		return true
// 	}

// 	if t.expectedHash == "" {
// 		return true
// 	}

// 	// check plain existence
// 	if !t.existsFile() {
// 		return false
// 	}

// 	// check integrity by comparing hash
// 	hash, err := t.Hash()
// 	if err != nil {
// 		boblog.Log.Error(err, "Unable to create target hash")
// 		return false
// 	}

// 	return hash == groundTruth
// }

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

	// This usually indicates a image was added/removed manually
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
