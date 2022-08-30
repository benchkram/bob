package bobtask

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/benchkram/errz"
	"github.com/mholt/archiver/v3"
	"gopkg.in/yaml.v3"

	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/bobtask/targettype"
	"github.com/benchkram/bob/pkg/boblog"
)

const __targetsFilesystem = "targets/filesystem"
const __targetsDocker = "targets/docker"
const __summary = "__summary"
const __metadata = "__metadata"

var ErrInvalidTarHeaderType = fmt.Errorf("invalid tar header type")

type archiveIO interface {
	archiver.Writer
	archiver.Reader
}

func newArchive() archiveIO             { return archiver.NewTarGz() } // TODO: use brotli
func newArchiveWriter() archiver.Writer { return newArchive() }
func newArchiveReader() archiver.Reader { return newArchive() }

// ArtifactPack creates an archive for a target
func (t *Task) ArtifactPack(artifactName hash.In) (err error) {
	defer errz.Recover(&err)

	if t.target == nil {
		return nil
	}

	boblog.Log.V(3).Info(fmt.Sprintf("[task:%s] creating artifact [%s] in localstore", t.name, artifactName))

	buildInfo, err := t.target.BuildInfo()
	errz.Fatal(err)

	dockerTargets := []string{}
	tempdir := ""

	// gather docker targets
	for dockerTarget := range buildInfo.Docker {
		targets, err := t.saveDockerImageTargets([]string{dockerTarget})
		errz.Fatal(err)
		dockerTargets = append(dockerTargets, targets...)

	}
	// in case of docker images, clear newly created targets by
	// images after archiving it in artifacts
	for _, target := range dockerTargets {
		defer func(dst string) { _ = os.Remove(dst) }(target)
	}

	artifact, err := t.local.NewArtifact(context.TODO(), artifactName.String())
	errz.Fatal(err)
	defer artifact.Close()

	archiveWriter := newArchiveWriter()
	err = archiveWriter.Create(artifact)
	errz.Fatal(err)
	defer archiveWriter.Close()

	// targets filesystem
	for fname := range buildInfo.Filesystem.Files {
		info, err := os.Lstat(fname)
		errz.Fatal(err)

		// trim the tasks directory from the internal name
		internalName := strings.TrimPrefix(fname, t.dir)
		// saved docker images are temporarly stored in the tmp dir,
		// this assures it's not added as prefix.
		internalName = strings.TrimPrefix(internalName, os.TempDir())
		internalName = strings.TrimPrefix(internalName, tempdir)
		internalName = strings.TrimPrefix(internalName, "/")

		// archiver needs the source path in case of a symlink,
		// so it can call `os.Readlink(source)`.
		var source string
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			abs, err := filepath.Abs(fname)
			errz.Fatal(err)
			source = abs
		}

		// open the file
		file, err := os.Open(fname)
		errz.Fatal(err)

		err = archiveWriter.Write(archiver.File{
			FileInfo: archiver.FileInfo{
				FileInfo:   info,
				CustomName: filepath.Join(__targetsFilesystem, internalName),
				SourcePath: source,
			},
			ReadCloser: file,
		})
		errz.Fatal(err)

		err = file.Close()
		errz.Fatal(err)
	}

	// targets docker
	for _, fname := range dockerTargets {
		info, err := os.Lstat(fname)
		errz.Fatal(err)

		// trim the tasks directory from the internal name
		internalName := strings.TrimPrefix(fname, t.dir)
		// saved docker images are temporarly stored in the tmp dir,
		// this assures it's not added as prefix.
		internalName = strings.TrimPrefix(internalName, os.TempDir())
		internalName = strings.TrimPrefix(internalName, tempdir)
		internalName = strings.TrimPrefix(internalName, "/")

		// archiver needs the source path in case of a symlink,
		// so it can call `os.Readlink(source)`.
		var source string
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			abs, err := filepath.Abs(fname)
			errz.Fatal(err)
			source = abs
		}

		// open the file
		file, err := os.Open(fname)
		errz.Fatal(err)

		err = archiveWriter.Write(archiver.File{
			FileInfo: archiver.FileInfo{
				FileInfo:   info,
				CustomName: filepath.Join(__targetsDocker, internalName),
				SourcePath: source,
			},
			ReadCloser: file,
		})
		errz.Fatal(err)

		err = file.Close()
		errz.Fatal(err)
	}

	metadata := NewArtifactMetadata()
	metadata.Taskname = t.name
	metadata.Project = t.Project()
	metadata.InputHash = artifactName.String()
	bin, err := yaml.Marshal(metadata)
	errz.Fatal(err)

	err = archiveWriter.Write(archiver.File{
		FileInfo: fileInfo{
			name: __metadata,
			data: bin,
		},
		ReadCloser: io.NopCloser(bytes.NewBuffer(bin)),
	})
	errz.Fatal(err)

	return nil
}

// func (t *Task) pathTargets() ([]string, error) {
// 	targets := []string{}
// 	for _, path := range t.target.PathsPlain() {
// 		stat, err := os.Stat(filepath.Join(t.dir, path))
// 		if err != nil {
// 			return targets, err
// 		}

// 		if stat.IsDir() {
// 			// TODO: Read all files from dir.
// 			root := filepath.Join(t.dir, path)
// 			_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
// 				if err != nil {
// 					return err
// 				}
// 				if d.IsDir() {
// 					return nil
// 				}

// 				targets = append(targets, path)
// 				return nil
// 			})
// 		} else {
// 			targets = append(targets, filepath.Join(t.dir, path))
// 		}

// 	}
// 	return targets, nil
// }

// saveDockerImageTargets calls `docker save` and returns a path to the tar archive.
func (t *Task) saveDockerImageTargets(in []string) ([]string, error) {
	targets := []string{}

	for _, image := range in {
		boblog.Log.V(2).Info(fmt.Sprintf("[image:%s] saving docker image", image))
		target, err := t.dockerRegistryClient.ImageSave(image)
		if err != nil {
			return targets, err
		}

		targets = append(targets, target)
	}

	return targets, nil
}

// ArtifactUnpack unpacks a artifact from the localstore if it exists.
// Return true on a succesful unpack operation.
func (t *Task) ArtifactUnpack(artifactName hash.In) (success bool, err error) {
	defer errz.Recover(&err)

	meta, err := t.GetArtifactMetadata(artifactName.String())
	errz.Fatal(err)

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

		// targets
		if strings.HasPrefix(header.Name, __targets) {

			filename := strings.TrimPrefix(header.Name, __targets+"/")

			// create directory structure
			dir := filepath.Dir(filename)
			if dir != "." && dir != "/" {
				err = os.MkdirAll(filepath.Join(t.dir, dir), 0775)
				errz.Fatal(err)
			}

			switch meta.TargetType {
			case targettype.Docker:
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
			case targettype.Path:
				fallthrough
			default:
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
		}
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

// GetArtifactMetadata creates a new artifact instance to retrive Metadata
// separately and returns ArtifactMetadata, close the artifacts before returning
func (t *Task) GetArtifactMetadata(artifactName string) (_ *ArtifactMetadata, err error) {
	artifact, err := t.local.GetArtifact(context.TODO(), artifactName)
	if err != nil {
		_, ok := err.(*fs.PathError)
		if ok {
			return nil, nil
		}
	}
	defer artifact.Close()

	artifactInfo, err := ArtifactInspectFromReader(artifact)
	if err != nil {
		return nil, err
	}

	return artifactInfo.Metadata(), nil
}

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
