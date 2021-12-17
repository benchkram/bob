package bobtask

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/Benchkram/bob/bobtask/hash"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
	"github.com/mholt/archiver/v3"
)

var ErrArtifactDoesNotExist = fmt.Errorf("artifact does not exist")

func (t *Task) ArtifactInspect(artifactNames ...hash.In) (ai ArtifactInfo, err error) {

	var artifactName hash.In
	if len(artifactNames) > 0 {
		artifactName = artifactNames[0]
	} else {
		artifactName, err = t.HashIn()
		errz.Fatal(err)
	}

	artifact, err := t.local.GetArtifact(context.TODO(), artifactName.String())
	if err != nil {
		_, ok := err.(*fs.PathError)
		if ok {
			return ai, usererror.Wrap(ErrArtifactDoesNotExist)
		}
		errz.Fatal(err)
	}
	defer artifact.Close()

	archiveReader := newArchiveReader()
	err = archiveReader.Open(artifact, 0)
	errz.Fatal(err)
	defer archiveReader.Close()

	info, err := artifactInspect(archiveReader)
	errz.Fatal(err)
	info.id = artifactName.String()
	info.taskname = t.name // TODO: get from artifact

	return info, nil
}

// ArtifactInspectFromPath opens a artifact from a io reader and returns
// a string containing compact information about a target.
func ArtifactInspectFromReader(reader io.ReadCloser) (_ ArtifactInfo, err error) {
	defer errz.Recover(&err)

	archiveReader := newArchiveReader()
	err = archiveReader.Open(reader, 0)
	errz.Fatal(err)
	defer archiveReader.Close()

	return artifactInspect(archiveReader)
}

func artifactInspect(archiveReader archiver.Reader) (_ *artifactInfo, err error) {
	defer errz.Recover(&err)

	info := newArtifactInfo()
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
			return nil, ErrInvalidTarHeaderType
		}

		if strings.HasPrefix(header.Name, __targets) {
			info.targets = append(info.targets, header.Name)
			// filename := strings.TrimPrefix(header.Name, __targets+"/")

		}
	}

	return info, nil
}
