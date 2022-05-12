package nix

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/bob/pkg/filehash"
	"github.com/benchkram/errz"
)

type cacheStore struct {
	db map[string]string
	f  *os.File
}

// NewCacheStore initialize a Nix cache store inside dir
func NewCacheStore() (_ *cacheStore, err error) {
	defer errz.Recover(&err)

	c := &cacheStore{
		db: make(map[string]string),
	}

	home, err := os.UserHomeDir()
	errz.Fatal(err)

	nixCacheFile := filepath.Join(home, global.BobCacheNix)
	f, err := os.OpenFile(nixCacheFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	errz.Fatal(err)
	c.f = f

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		pair := strings.SplitN(scanner.Text(), ":", 2)

		c.db[pair[0]] = pair[1]
	}

	if err := scanner.Err(); err != nil {
		errz.Fatal(err)
	}

	return c, nil
}

// Get value from cache by its key
// Additionally also checks if path exists on the system
func (c *cacheStore) Get(key string) (string, bool) {
	path, ok := c.db[key]

	// Assure path exists on the filesystem.
	if ok && !file.Exists(path) {
		return path, false
	}

	return path, ok
}

// Save dependency inside the cache with its corresponding store path
func (c *cacheStore) Save(dependency Dependency, storePath string) (err error) {
	defer errz.Recover(&err)

	key, err := c.generateKey(dependency)

	if _, err := c.f.Write([]byte(fmt.Sprintf("%s:%s\n", key, storePath))); err != nil {
		_ = c.f.Close() // ignore error; Write error takes precedence
		errz.Fatal(err)
	}
	c.db[key] = storePath

	return nil
}

// FilterCachedDependencies will filter out dependencies which are already cached
func (c *cacheStore) FilterCachedDependencies(deps []Dependency) (_ []Dependency, err error) {
	defer errz.Recover()
	notCached := make([]Dependency, 0)
	for _, v := range deps {
		key, err := c.generateKey(v)
		errz.Fatal(err)

		if _, exists := c.Get(key); !exists {
			notCached = append(notCached, v)
		}
	}
	return notCached, nil
}

// Close closes the file used in cache
func (c *cacheStore) Close() error {
	return c.f.Close()
}

// generateKey generates key from the cache
// if it's a file then will hash the nixpkgs + file contents
// if it's a package name will hash the packageName:nixpkgs content
func (c *cacheStore) generateKey(dependency Dependency) (_ string, err error) {
	defer errz.Recover(&err)
	var h []byte

	if strings.HasSuffix(dependency.Name, ".nix") {
		aggregatedHashes := bytes.NewBuffer([]byte{})
		h, err = filehash.HashBytes(bytes.NewBufferString(dependency.Nixpkgs))
		errz.Fatal(err)
		aggregatedHashes.Write(h)

		fileHash, err := filehash.Hash(dependency.Name)
		errz.Fatal(err)
		aggregatedHashes.Write(fileHash)

		h, err = filehash.HashBytes(aggregatedHashes)
		errz.Fatal(err)
	} else {
		toHash := fmt.Sprintf("%s:%s", dependency.Name, dependency.Nixpkgs)
		h, err = filehash.HashBytes(bytes.NewBufferString(toHash))
		errz.Fatal(err)
	}

	return hex.EncodeToString(h), nil
}
