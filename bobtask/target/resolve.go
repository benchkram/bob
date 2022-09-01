package target

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/pkg/boblog"
)

// resolve filesystem entries based on filesystemEntriesRaw.
// Becomes a noop after the first call.
func (t *T) Resolve() error {
	resolved := []string{}

	boblog.Log.V(2).Info(fmt.Sprintf("resolving 111"))
	if t.filesystemEntries != nil {
		return nil
	}
	boblog.Log.V(2).Info(fmt.Sprintf("resolving 222"))

	for _, path := range t.FilesystemEntriesRaw() {
		boblog.Log.V(2).Info(fmt.Sprintf("resolving %s", path))
		fileInfo, err := os.Lstat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}

		boblog.Log.V(2).Info(fmt.Sprintf("resolving 3333"))
		if fileInfo.IsDir() {
			boblog.Log.V(2).Info(fmt.Sprintf("resolving 4444"))
			if err := filepath.WalkDir(path, func(p string, fi fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				// Skip dirs
				if !fi.IsDir() {
					resolved = append(resolved, p)
				}

				return nil
			}); err != nil {
				return fmt.Errorf("failed to walk dir %q: %w", path, err)
			}
		} else {
			resolved = append(resolved, path)
		}
	}

	t.filesystemEntries = &resolved

	//litter.Dump()

	return nil
}
