package bobtask

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Benchkram/bob/bobtask/hash"
	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/Benchkram/errz"
	"github.com/mholt/archiver/v3"
)

const __targets = "targets"
const __exports = "exports"
const __summary = "__summary"

var ErrInvalidTarHeaderType = fmt.Errorf("invalid tar header type")

type archiveIO interface {
	archiver.Writer
	archiver.Reader
}

func newArchive() archiveIO             { return archiver.NewTarGz() } // TODO: use brotli
func newArchiveWriter() archiver.Writer { return newArchive() }
func newArchiveReader() archiver.Reader { return newArchive() }

// ArtifactPack creates a archive for a target & exports.
func (t *Task) ArtifactPack(artifactName hash.In) (err error) {
	defer errz.Recover(&err)

	if t.target == nil && len(t.Exports) == 0 {
		return nil
	}

	boblog.Log.V(3).Info(fmt.Sprintf("[task:%s] creating artifact [%s] in localstore", t.name, artifactName))

	targets := []string{}
	exports := []string{}
	if t.target != nil {
		for _, path := range t.target.Paths {
			stat, err := os.Stat(filepath.Join(t.dir, path))
			errz.Fatal(err)
			if stat.IsDir() {
				// TODO: Read all files from dir.
				root := filepath.Join(t.dir, path)
				_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if d.IsDir() {
						return nil
					}

					targets = append(targets, path)
					return nil
				})
			} else {
				targets = append(targets, filepath.Join(t.dir, path))
			}

		}
	}
	for _, path := range t.Exports {
		exports = append(exports, filepath.Join(t.dir, path.String()))
	}

	artifact, err := t.local.NewArtifact(context.TODO(), artifactName.String())
	errz.Fatal(err)
	defer artifact.Close()

	archiveWriter := newArchiveWriter()
	err = archiveWriter.Create(artifact)
	errz.Fatal(err)
	defer archiveWriter.Close()

	// targets
	for _, fname := range targets {
		info, err := os.Stat(fname)
		errz.Fatal(err)

		// trim the tasks directory from the internal name
		internalName := strings.TrimPrefix(fname, t.dir)
		internalName = strings.TrimPrefix(internalName, "/")

		// open the file
		file, err := os.Open(fname)
		errz.Fatal(err)

		err = archiveWriter.Write(archiver.File{
			FileInfo: archiver.FileInfo{
				FileInfo:   info,
				CustomName: filepath.Join(__targets, internalName),
			},
			ReadCloser: file,
		})
		errz.Fatal(err)

		err = file.Close()
		errz.Fatal(err)
	}

	// exports
	exportSummary, err := json.Marshal(t.Exports)
	errz.Fatal(err)
	err = archiveWriter.Write(archiver.File{
		FileInfo: archiver.FileInfo{
			FileInfo: fileInfo{
				name: __summary,
				data: exportSummary,
			},
			CustomName: filepath.Join(__exports, __summary),
		},
		ReadCloser: io.NopCloser(bytes.NewBuffer(exportSummary)),
	})
	for _, fname := range exports {
		info, err := os.Stat(fname)
		errz.Fatal(err)

		// get file's name for the inside of the archive
		internalName, err := archiver.NameInArchive(info, fname, fname)
		errz.Fatal(err)

		// open the file
		file, err := os.Open(fname)
		errz.Fatal(err)

		err = archiveWriter.Write(archiver.File{
			FileInfo: archiver.FileInfo{
				FileInfo:   info,
				CustomName: filepath.Join(__exports, internalName),
			},
			ReadCloser: file,
		})
		errz.Fatal(err)

		err = file.Close()
		errz.Fatal(err)
	}

	return nil
}

// ArtifactUnpack unpacks a artifact from the localstore if it exists.
// Return true on a succesful unpack operation.
func (t *Task) ArtifactUnpack(artifactName hash.In) (success bool, err error) {
	defer errz.Recover(&err)

	artifact, err := t.local.GetArtifact(context.TODO(), artifactName.String())
	if err != nil {
		_, ok := err.(*fs.PathError)
		if ok {
			return false, nil
		}
		errz.Fatal(err)
	}
	defer artifact.Close()

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

		// targets
		if strings.HasPrefix(header.Name, __targets) {
			filename := strings.TrimPrefix(header.Name, __targets+"/")

			// create directory structure
			dir := filepath.Dir(filename)
			if dir != "." && dir != "/" {
				err = os.MkdirAll(filepath.Join(t.dir, dir), 0775)
				errz.Fatal(err)
			}

			// create dst
			dst := filepath.Join(t.dir, filename)
			f, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
			errz.Fatal(err)
			defer f.Close()

			// copy
			_, err = io.Copy(f, archiveFile)
			errz.Fatal(err)
		}

		// exports
		// TODO: handle exports
	}

	return true, nil
}

// ArtifactExists return true when the artifact exists in localstore
func (t *Task) ArtifactExists(artifactName hash.In) bool {
	artifact, err := t.local.GetArtifact(context.TODO(), artifactName.String())
	if err != nil {
		return false
	}
	artifact.Close()
	return true
}

// ArtifactInspectFromPath opens a artifact from path and returns
// a string containing compact information about a target.
// func ArtifactInspectFromPath(path string) (description string, err error) {
// 	defer errz.Recover(&err)

// 	buf := bytes.NewBufferString(description)

// 	f, err := os.Open(path)
// 	errz.Fatal(err)

// 	archiveReader := newArchiveReader()
// 	err = archiveReader.Open(f, 0)
// 	errz.Fatal(err)
// 	defer archiveReader.Close()

// 	for {
// 		archiveFile, err := archiveReader.Read()
// 		if err != nil {
// 			if errors.Is(err, io.EOF) {
// 				break
// 			}
// 			errz.Fatal(err)
// 		}

// 		header, ok := archiveFile.Header.(*tar.Header)
// 		if !ok {
// 			return "", ErrInvalidTarHeaderType
// 		}

// 		// targets
// 		indent := "  "
// 		fmt.Fprint(buf, "Targets:\n")
// 		if strings.HasPrefix(header.Name, __targets) {

// 			fmt.Fprintf(buf, "%s%s\n", indent, header.Name)
// 			// filename := strings.TrimPrefix(header.Name, __targets+"/")

// 			// // create directory structure
// 			// dir := filepath.Dir(filename)
// 			// if dir != "." && dir != "/" {
// 			// 	err = os.MkdirAll(filepath.Join(t.dir, dir), 0775)
// 			// 	errz.Fatal(err)
// 			// }

// 			// // create dst
// 			// dst := filepath.Join(t.dir, filename)
// 			// f, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
// 			// errz.Fatal(err)
// 			// defer f.Close()

// 			// // copy
// 			// _, err = io.Copy(f, archiveFile)
// 			// errz.Fatal(err)
// 		}

// 		// exports
// 		// TODO: handle exports
// 	}
// 	return buf.String(), nil
// }

// // TODO: implement me
// func artifactWalk() {

// }

type fileInfo struct {
	name string
	data []byte
}

func (mif fileInfo) Name() string       { return mif.name }
func (mif fileInfo) Size() int64        { return int64(len(mif.data)) }
func (mif fileInfo) Mode() os.FileMode  { return 0444 }
func (mif fileInfo) ModTime() time.Time { return time.Now() }
func (mif fileInfo) IsDir() bool        { return false }
func (mif fileInfo) Sys() interface{}   { return nil }
