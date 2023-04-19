package bobtask

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/bobtask/target"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/errz"
)

// ArtifactExtract extract an artifact from the localstore if it exists.
// Return true on a successful extract operation.
func (t *Task) ArtifactExtract(artifactName hash.In, invalidFiles map[string][]target.Reason) (success bool, err error) {
	defer errz.Recover(&err)

	boblog.Log.V(5).Info(fmt.Sprintf("Extracting artifact [artifact: %s, invalidFIles: %s]", artifactName, invalidFiles))

	homeDir, err := os.UserHomeDir()
	errz.Fatal(err)

	artifact, _, err := t.local.GetArtifact(context.TODO(), artifactName.String())
	if err != nil {
		_, ok := err.(*fs.PathError)
		if ok {
			return false, nil
		}
		errz.Fatal(err)
	}
	defer artifact.Close()

	// Assure task is cleaned up before extracting
	err = t.CleanTargetsWithReason(invalidFiles)
	errz.Fatal(err)

	archiveReader := newArchiveReader()
	err = archiveReader.Open(artifact, 0)
	errz.Fatal(err)
	defer archiveReader.Close()

	for {
		archiveFile, err := archiveReader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			errz.Fatal(err)
		}

		header, ok := archiveFile.Header.(*tar.Header)
		if !ok {
			return false, ErrInvalidTarHeaderType
		}

		// targets filesystem
		if strings.HasPrefix(header.Name, __targetsFilesystem) {
			filename := strings.TrimPrefix(header.Name, __targetsFilesystem+"/")

			// create directory structure
			dir := filepath.Dir(filename)
			if dir != "." && dir != "/" {
				err = os.MkdirAll(filepath.Join(t.dir, dir), 0775)
				errz.Fatal(err)
			}

			dst := filepath.Join(t.dir, filename)

			// symlink
			if archiveFile.FileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
				if dst == "/" || dst == homeDir {
					return false, fmt.Errorf("Cleanup of %s is not allowed", dst)
				}
				err = os.RemoveAll(dst)
				errz.Fatal(err)
				err = os.Symlink(header.Linkname, dst)
				errz.Fatal(err)
				continue
			}

			if shouldFetchFromCache(filename, invalidFiles) {
				// extract to destination
				f, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
				errz.Fatal(err)

				_, err = io.Copy(f, archiveFile)
				errz.Fatal(err)

				// closing the file right away to reduce the number of open files
				_ = f.Close()
			}
		}

		// targets docker
		if strings.HasPrefix(header.Name, __targetsDocker) {
			filename := strings.TrimPrefix(header.Name, __targetsDocker+"/")

			// create directory structure
			dir := filepath.Dir(filename)
			if dir != "." && dir != "/" {
				err = os.MkdirAll(filepath.Join(t.dir, dir), 0775)
				errz.Fatal(err)
			}

			// load the docker image from destination
			dst := filepath.Join(os.TempDir(), filename)

			// extract to destination
			f, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			errz.Fatal(err)
			_, err = io.Copy(f, archiveFile)
			// closing the file right away to reduce the number of open files
			_ = f.Close()
			errz.Fatal(err)

			boblog.Log.V(2).Info(fmt.Sprintf("[task:%s] loading docker image from %s", t.name, dst))
			err = t.dockerRegistryClient.ImageLoad(dst)
			errz.Fatal(err)

			// delete the extracted docker image archive
			// after `docker load`
			defer func() { _ = os.Remove(dst) }()
		}

	}

	return true, nil
}

// shouldFetchFromCache checks if a file should be brought back from cache inside the target
// A file will be brought back from cache if it's missing or was changed
func shouldFetchFromCache(filename string, invalidFiles map[string][]target.Reason) bool {
	// FIXME: accessing a hashmap twice is inefficent.
	if _, ok := invalidFiles[filename]; !ok {
		return true
	}
	for _, reason := range invalidFiles[filename] {
		if reason == target.ReasonSizeChanged || reason == target.ReasonHashChanged || reason == target.ReasonMissing {
			return true
		}
	}
	return false
}
