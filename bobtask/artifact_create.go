package bobtask

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/benchkram/bob/bobtask/target"
	"github.com/benchkram/errz"
	"github.com/mholt/archiver/v3"
	"gopkg.in/yaml.v3"

	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/boblog"
)

const __targetsFilesystem = "targets/filesystem"
const __targetsDocker = "targets/docker"
const __metadata = "__metadata"

var ErrInvalidTarHeaderType = fmt.Errorf("invalid tar header type")

type archiveIO interface {
	archiver.Writer
	archiver.Reader
}

func newArchive() archiveIO             { return archiver.NewTarGz() } // TODO: use brotli
func newArchiveWriter() archiver.Writer { return newArchive() }
func newArchiveReader() archiver.Reader { return newArchive() }

// ArtifactCreate create an archive for one or multiple targets
func (t *Task) ArtifactCreate(artifactName hash.In) (err error) {
	defer errz.Recover(&err)

	if t.target == nil {
		return nil
	}

	boblog.Log.V(3).Info(fmt.Sprintf("[task:%s] creating artifact [%s] in localstore", t.name, artifactName))

	tt, err := t.Target()
	errz.Fatal(err)
	buildInfo, err := tt.BuildInfo()

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

	artifact, err := t.local.NewArtifact(context.TODO(), artifactName.String(), 0)
	errz.Fatal(err)
	defer artifact.Close()

	archiveWriter := newArchiveWriter()
	err = archiveWriter.Create(artifact)
	errz.Fatal(err)
	defer archiveWriter.Close()

	boblog.Log.V(3).Info(fmt.Sprintf("[task:%s] file in buildinfo %d", t.name, len(buildInfo.Filesystem.Files)))

	// targets filesystem
	for fname := range buildInfo.Filesystem.Files {
		if target.ShouldIgnore(fname) {
			continue
		}
		info, err := os.Lstat(fname)
		errz.Fatal(err)

		if info.IsDir() {
			continue
		}

		// trim the tasks directory from the internal name
		internalName := strings.TrimPrefix(fname, t.dir)
		// saved docker images are temporarily stored in the tmp dir,
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

// ArtifactExists return true when the artifact exists in localstore
func (t *Task) ArtifactExists(artifactName hash.In) bool {
	return t.local.ArtifactExists(context.TODO(), artifactName.String())
}

// GetArtifactMetadata creates a new artifact instance to retrive Metadata
// separately and returns ArtifactMetadata, close the artifacts before returning
func (t *Task) GetArtifactMetadata(artifactName string) (_ *ArtifactMetadata, err error) {
	artifact, _, err := t.local.GetArtifact(context.TODO(), artifactName)
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
