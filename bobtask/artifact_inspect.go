package bobtask

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/Benchkram/bob/bobtask/hash"
	"github.com/Benchkram/errz"
	"github.com/mholt/archiver/v3"
)

var ErrArtifactDoesNotExist = fmt.Errorf("artifact does not exist")

func (t *Task) ArtifactInspect(artifactName hash.In) (_ string, err error) {
	artifact, err := t.local.GetArtifact(context.TODO(), artifactName.String())
	if err != nil {
		_, ok := err.(*fs.PathError)
		if ok {
			return "", ErrArtifactDoesNotExist
		}
		errz.Fatal(err)
	}
	defer artifact.Close()

	return ArtifactInspectFromReader(artifact)
}

// ArtifactInspectFromPath opens a artifact from a io reader and returns
// a string containing compact information about a target.
func ArtifactInspectFromReader(reader io.ReadCloser) (_ string, err error) {
	defer errz.Recover(&err)

	archiveReader := newArchiveReader()
	err = archiveReader.Open(reader, 0)
	errz.Fatal(err)
	defer archiveReader.Close()

	return artifactInspect(archiveReader)
}

func artifactInspect(archiveReader archiver.Reader) (description string, err error) {
	defer errz.Recover(&err)

	buf := bytes.NewBufferString(description)

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
			return "", ErrInvalidTarHeaderType
		}

		// targets
		indent := "  "

		//fmt.Fprint(buf, "Targets:\n")
		if strings.HasPrefix(header.Name, __targets) {
			fmt.Fprintf(buf, "%s%s\n", indent, header.Name)
			// filename := strings.TrimPrefix(header.Name, __targets+"/")

		}

	}

	return buf.String(), nil
}
