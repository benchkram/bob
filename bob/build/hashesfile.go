package build

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/Benchkram/bob/pkg/file"
	"github.com/Benchkram/bob/pkg/filehash"
	"github.com/Benchkram/errz"
)

var (
	// TODO: Use bob.BuildToolDir instead of ".bob"
	BobCacheDir        = ".bobcache"
	FileHashesFileName = filepath.Join(BobCacheDir, "filehashes")
	TaskHashesFileName = filepath.Join(BobCacheDir, "hashes")
)

type TaskHashes map[string]string

type FileHash struct {
	Path string
	Hash string
}

func HashFiles(files []string) []FileHash {
	fhs := make([]FileHash, 0, len(files))
	for _, f := range files {
		h, err := filehash.Hash(f)
		if err != nil {
			log.Printf("failed to hash file %q: %v\n", f, err)
			continue
		}

		fhs = append(fhs, FileHash{
			Path: f,
			Hash: hex.EncodeToString(h),
		})
	}
	return fhs
}

func ReadHashes(t *Task) (taskHashs TaskHashes, _ error) {
	hashesFile := filepath.Join(t.dir, TaskHashesFileName)

	if !file.Exists(hashesFile) {
		return nil, ErrHashesFileDoesNotExist
	}

	c, err := ioutil.ReadFile(hashesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %w", hashesFile, err)
	}

	if err := json.Unmarshal(c, &taskHashs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return taskHashs, nil
}

func StoreHash(t *Task, hash string) (err error) {
	defer errz.Recover(&err)

	hashes, err := ReadHashes(t)
	if err != nil {
		if !errors.Is(err, ErrHashesFileDoesNotExist) {
			errz.Fatal(err)
		}
	}

	if hashes == nil {
		hashes = TaskHashes{}
	}

	hashes[t.name] = hash

	return WriteHashes(t, hashes)
}

func WriteHashes(t *Task, taskhashes TaskHashes) error {
	hashesFile := filepath.Join(t.dir, TaskHashesFileName)

	data, err := json.Marshal(taskhashes)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(hashesFile), 0755); err != nil {
		return fmt.Errorf("failed to create .bobcache dir: %w", err)
	}

	if err := ioutil.WriteFile(hashesFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func ReadFileHashes(t *Task) ([]FileHash, error) {
	hashesFile := filepath.Join(t.dir, FileHashesFileName)

	if !file.Exists(hashesFile) {
		return []FileHash{}, ErrHashesFileDoesNotExist
	}

	c, err := ioutil.ReadFile(hashesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %w", hashesFile, err)
	}

	var fhs []FileHash
	if err := json.Unmarshal(c, &fhs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return fhs, nil
}

func WriteFileHashes(t *Task, fh []FileHash) error {
	hashesFile := filepath.Join(t.dir, FileHashesFileName)

	data, err := json.Marshal(fh)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	const mode = 0644
	if err := ioutil.WriteFile(hashesFile, data, mode); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func FileHashesDiffer(fhs1, fhs2 []FileHash) error {
	for _, fh1 := range fhs1 {
		var found bool
		for _, fh2 := range fhs2 {
			if fh1.Path == fh2.Path {
				found = true
				if fh1.Hash != fh2.Hash {
					return fmt.Errorf("hashes of file %q differ: %q != %q", fh1.Path, fh1.Hash, fh2.Hash)
				}
				break
			}
		}

		if !found {
			return fmt.Errorf("did not find file %q in second slice", fh1.Path)
		}
	}

	for _, fh2 := range fhs2 {
		var found bool
		for _, fh1 := range fhs1 {
			if fh1.Path == fh2.Path {
				found = true
				if fh1.Hash != fh2.Hash {
					return fmt.Errorf("hashes of file %q differ: %q != %q", fh2.Path, fh2.Hash, fh1.Hash)
				}
				break
			}
		}

		if !found {
			return fmt.Errorf("did not find file %q in first slice", fh2.Path)
		}
	}

	return nil
}
