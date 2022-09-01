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
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/errz"
)

// ArtifactUnpack unpacks a artifact from the localstore if it exists.
// Return true on a succesful unpack operation.
func (t *Task) ArtifactUnpack(artifactName hash.In) (success bool, err error) {
	defer errz.Recover(&err)

	//	meta, err := t.GetArtifactMetadata(artifactName.String())
	//	errz.Fatal(err)

	artifact, err := t.local.GetArtifact(context.TODO(), artifactName.String())
	if err != nil {
		_, ok := err.(*fs.PathError)
		if ok {
			return false, nil
		}
		errz.Fatal(err)
	}
	defer artifact.Close()

	// Assure tasks is cleaned up before unpacking
	err = t.Clean()
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
				err = os.Symlink(header.Linkname, dst)
				errz.Fatal(err)
				continue
			}

			// extract to destination
			f, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			errz.Fatal(err)
			_, err = io.Copy(f, archiveFile)
			// closing the file right away to reduce the number of open files
			_ = f.Close()
			errz.Fatal(err)
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

			// delete the unpacked docker image archive
			// after `docker load`
			defer func() { _ = os.Remove(dst) }()
		}

	}

	return true, nil
}
