package bobtask

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Benchkram/bob/bob/global"
	"github.com/Benchkram/bob/bobtask/hash"
	"github.com/Benchkram/bob/pkg/file"
	"github.com/Benchkram/errz"
)

func (t *Task) ReadHash() (taskHashs hash.Task, err error) {
	defer errz.Recover(&err)

	hashes, err := t.ReadHashes()
	errz.Fatal(err)

	hash, ok := hashes[t.name]
	if !ok {
		return taskHashs, ErrTaskHashDoesNotExist
	}

	return hash, nil
}

func (t *Task) ReadHashes() (taskHashs hash.Hashes, _ error) {
	hashesFile := filepath.Join(t.dir, global.TaskHashesFileName)

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

func (t *Task) StoreHash(hashTask *hash.Task) (err error) {
	defer errz.Recover(&err)

	hashes, err := t.ReadHashes()
	if err != nil {
		if !errors.Is(err, ErrHashesFileDoesNotExist) {
			errz.Fatal(err)
		}
	}

	if hashes == nil {
		hashes = make(hash.Hashes)
	}

	hashes[t.name] = *hashTask

	return WriteHashes(t, hashes)
}

func WriteHashes(t *Task, taskhashes hash.Hashes) error {
	hashesFile := filepath.Join(t.dir, global.TaskHashesFileName)

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

func ReadFileHashes(t *Task) ([]hash.H, error) {
	hashesFile := filepath.Join(t.dir, global.FileHashesFileName)

	if !file.Exists(hashesFile) {
		return []hash.H{}, ErrHashesFileDoesNotExist
	}

	c, err := ioutil.ReadFile(hashesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %w", hashesFile, err)
	}

	var fhs []hash.H
	if err := json.Unmarshal(c, &fhs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return fhs, nil
}

func WriteFileHashes(t *Task, fh []hash.H) error {
	hashesFile := filepath.Join(t.dir, global.FileHashesFileName)

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
