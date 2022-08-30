package target

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/bobtask/buildinfo"
	"github.com/benchkram/bob/pkg/dockermobyutil"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"

	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/bob/pkg/filehash"
)

// BuildInfo reads file info and computes the target hash
// for filesystem and docker targets.
func (t *T) BuildInfo() (bi *buildinfo.Targets, err error) {
	defer errz.Recover(&err)

	if t.current != nil {
		return t.current, nil
	}

	bi = buildinfo.NewTargets()

	// Filesystem
	buildInfoFilesystem, err := t.buildinfoFiles(t.filesystemEntries)
	errz.Fatal(err)
	bi.Filesystem = buildInfoFilesystem

	// Docker
	for _, image := range t.dockerImages {
		hash, err := t.dockerImageHash(image)
		errz.Fatal(err)

		bi.Docker[image] = buildinfo.BuildInfoDocker{Hash: hash}
	}

	t.current = bi

	return bi, nil
}

func (t *T) buildinfoFiles(paths []string) (bi buildinfo.BuildInfoFiles, _ error) {

	h := filehash.New()
	for _, path := range paths {
		path = filepath.Join(t.dir, path)

		if !file.Exists(path) {
			return buildinfo.BuildInfoFiles{}, usererror.Wrapm(fmt.Errorf("target does not exist %q", path), "failed to hash target")
		}
		targetInfo, err := os.Stat(path)
		if err != nil {
			return buildinfo.BuildInfoFiles{}, fmt.Errorf("failed to get file info %q: %w", path, err)
		}

		if targetInfo.IsDir() {
			if err := filepath.WalkDir(path, func(p string, f fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if f.IsDir() {
					return nil
				}

				err = h.AddFile(p)
				if err != nil {
					return fmt.Errorf("failed to hash target %q: %w", f, err)
				}

				info, err := f.Info()
				if err != nil {
					return fmt.Errorf("failed to get file info %q: %w", p, err)
				}
				bi.Files[p] = buildinfo.BuildInfoFile{Modified: info.ModTime(), Size: info.Size()}

				return nil
			}); err != nil {
				return buildinfo.BuildInfoFiles{}, fmt.Errorf("failed to walk dir %q: %w", path, err)
			}
			// TODO: what happens on a empty dir?
		} else {
			err = h.AddFile(path)
			if err != nil {
				return buildinfo.BuildInfoFiles{}, fmt.Errorf("failed to hash target %q: %w", path, err)
			}
			bi.Files[path] = buildinfo.BuildInfoFile{Modified: targetInfo.ModTime(), Size: targetInfo.Size()}
		}
	}

	bi.Hash = hex.EncodeToString(h.Sum())

	return bi, nil
}

func (t *T) dockerImageHash(image string) (string, error) {
	hash, err := t.dockerRegistryClient.ImageHash(image)
	if err != nil {
		if errors.Is(err, dockermobyutil.ErrImageNotFound) {
			return "", usererror.Wrapm(err, "failed to fetch docker image hash")
		} else {
			return "", fmt.Errorf("failed to get docker image hash info %q: %w", image, err)
		}
	}
	return hash, nil
}
