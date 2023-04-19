package target

import (
	"encoding/hex"
	"fmt"
	"github.com/benchkram/bob/bobtask/buildinfo"
	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/bob/pkg/filehash"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// BuildInfo reads file info and computes the target hash
// for filesystem and docker targets.
func (t *T) BuildInfo() (bi *buildinfo.Targets, err error) {
	defer errz.Recover(&err)

	bi = buildinfo.NewTargets()

	// Filesystem
	buildInfoFilesystem, err := t.buildinfoFiles(t.filesystemEntriesRaw)
	errz.Fatal(err)
	bi.Filesystem = buildInfoFilesystem

	// Docker
	for _, image := range t.dockerImages {
		hash, err := t.dockerImageHash(image)
		errz.Fatal(err)

		bi.Docker[image] = buildinfo.BuildInfoDocker{Hash: hash}
	}

	return bi, nil
}

func (t *T) buildinfoFiles(paths []string) (bi buildinfo.BuildInfoFiles, _ error) {
	bi = *buildinfo.NewBuildInfoFiles()

	h := filehash.New()

	// Use a sorted path array to assure the hash of all files
	// is computed in a consistent order.
	sort.Strings(paths)

	for _, path := range paths {
		path = filepath.Join(t.dir, path)

		if !file.Exists(path) {
			return buildinfo.BuildInfoFiles{}, usererror.Wrapm(ErrTargetDoesNotExist, fmt.Sprintf("[path: %q]", path))
		}
		targetInfo, err := os.Lstat(path)
		if err != nil {
			return buildinfo.BuildInfoFiles{}, fmt.Errorf("failed to get file info %q: %w", path, err)
		}

		if targetInfo.IsDir() {
			if err := filepath.WalkDir(path, func(p string, f fs.DirEntry, err error) error {
				if ShouldIgnore(p) {
					return nil
				}
				if err != nil {
					return err
				}
				if f.IsDir() {
					// we add directories to the list of files to later verify target, but no need for size and hash
					bi.Files[p] = buildinfo.BuildInfoFile{Size: -1, Hash: ""}
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

				contentHash, err := filehash.HashOfFile(p)
				if err != nil {
					return fmt.Errorf("failed to get file hash %q: %w", p, err)
				}

				bi.Files[p] = buildinfo.BuildInfoFile{Size: info.Size(), Hash: contentHash}

				return nil
			}); err != nil {
				return buildinfo.BuildInfoFiles{}, fmt.Errorf("failed to walk dir %q: %w", path, err)
			}
			// TODO: what happens on a empty dir?
		} else {
			if ShouldIgnore(path) {
				continue
			}
			err = h.AddFile(path)
			if err != nil {
				return buildinfo.BuildInfoFiles{}, fmt.Errorf("failed to hash target %q: %w", path, err)
			}
			contentHash, err := filehash.HashOfFile(path)
			if err != nil {
				return buildinfo.BuildInfoFiles{}, fmt.Errorf("failed to get file hash %q: %w", path, err)
			}
			bi.Files[path] = buildinfo.BuildInfoFile{Size: targetInfo.Size(), Hash: contentHash}
		}
	}

	bi.Hash = hex.EncodeToString(h.Sum())

	return bi, nil
}
