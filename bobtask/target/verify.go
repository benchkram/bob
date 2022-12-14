package target

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/filehash"
	"github.com/benchkram/bob/pkg/sliceutil"
)

// Hint: comparing the modification time is tricky as a artifact extraction
// from a tar archive changes the modification time of a file.

// VerifyShallow compare targets against an existing
// buildinfo. It will only check if the size of the files changed.
// Docker targets are verified similarly as in plain verify
// as there is no performance penalty.
// In case the expected buildinfo does not exist Verify checks against filesystemEntriesRaw.
func (t *T) VerifyShallow() *VerifyResult {
	r := NewVerifyResult()
	r.TargetIsValid = t.verifyFilesystemShallow(r) && t.verifyDocker()

	return r
}

// VerifyResult is the result of a target verify call
// it tells if the target is valid and if not InvalidFiles will contain the list of invalid files along with their reason
// a file can be invalid for multiple reasons. ex. a changed file is invalid because of size and content hash
type VerifyResult struct {
	// TargetIsValid shows if target is valid or not
	TargetIsValid bool
	// InvalidFiles maps filePath to reasons why it's invalid
	InvalidFiles map[string][]Reason
}

// NewVerifyResult initializes a new VerifyResult
func NewVerifyResult() *VerifyResult {
	var v VerifyResult
	v.InvalidFiles = make(map[string][]Reason)
	return &v
}

// AddInvalidReason adds a reason for invalidation to a certain filePath
func (v *VerifyResult) AddInvalidReason(filePath string, reason Reason) {
	if _, ok := v.InvalidFiles[filePath]; ok {
		v.InvalidFiles[filePath] = append(v.InvalidFiles[filePath], reason)
	} else {
		v.InvalidFiles[filePath] = []Reason{reason}
	}
}

// Reason contains the reason why a file/directory makes the target invalid
type Reason string

const (
	ReasonCreatedAfterBuild Reason = "CREATED-AFTER-BUILD"
	ReasonSizeChanged       Reason = "SIZE-CHANGED"
	ReasonHashChanged       Reason = "HASH-CHANGED"
	ReasonDeleted           Reason = "DELETED"
)

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

// verifyFilesystemShallow
// todo checking contents hash makes this method non-shallow anymore. rename this
func (t *T) verifyFilesystemShallow(v *VerifyResult) bool {
	if t.filesystemEntries == nil {
		return true
	}

	// check for deleted files
	for k, _ := range t.expected.Filesystem.Files {
		if !sliceutil.Contains(*t.filesystemEntries, k) {
			v.AddInvalidReason(k, ReasonDeleted)
		}
	}

	for _, path := range *t.filesystemEntries {
		fileInfo, err := os.Lstat(path)
		if err != nil {
			return false
		}

		// check for newly added files
		expectedFileInfo, ok := t.expected.Filesystem.Files[path]
		if !ok {
			v.AddInvalidReason(path, ReasonCreatedAfterBuild)
		}

		// directories are not checked for size/hash
		if fileInfo.IsDir() {
			continue
		}

		// check file size
		if fileInfo.Size() != expectedFileInfo.Size {
			v.AddInvalidReason(path, ReasonSizeChanged)
			boblog.Log.V(2).Info(fmt.Sprintf("failed to verify [%s], different sizes [current: %d != expected: %d]", path, fileInfo.Size(), expectedFileInfo.Size))
		}

		// checks the contents hash of the file with the ones from build info
		hashOfFile, err := filehash.HashOfFile(path)
		if err != nil {
			return false
		}
		if hashOfFile != expectedFileInfo.Hash {
			v.AddInvalidReason(path, ReasonHashChanged)
			boblog.Log.V(2).Info(fmt.Sprintf("failed to verify [%s], different hashes [current: %s != expected: %s]", path, hashOfFile, expectedFileInfo.Hash))
		}
	}

	return len(v.InvalidFiles) == 0
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
