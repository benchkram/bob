package bobtask

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"strings"

	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
	"github.com/mholt/archiver/v3"
	"gopkg.in/yaml.v3"
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
		} else if strings.HasPrefix(header.Name, __metadata) {
			bin, err := ioutil.ReadAll(archiveFile)
			errz.Fatal(err)

			metadata := NewArtifactMetadata()
			err = yaml.Unmarshal(bin, metadata)
			errz.Fatal(err)

			info.metadata = metadata
		}

	}

	return info, nil
}
