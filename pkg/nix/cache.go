package nix

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/bob/pkg/filehash"
)

type fileCacheStore struct {
	db map[string]string
	f  *os.File
}

// NewFileCacheStore initialize a Nix cache store inside dir
func NewFileCacheStore() (_ *fileCacheStore, err error) {
	defer errz.Recover(&err)

	c := &fileCacheStore{
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
func (c *fileCacheStore) Get(key string) (string, bool) {
	path, ok := c.db[key]

	// Assure path exists on the filesystem.
	if ok && !file.Exists(path) {
		return path, false
	}

	return path, ok
}

// Save dependency inside the cache with its corresponding store path
func (c *fileCacheStore) Save(key string, storePath string) (err error) {
	defer errz.Recover(&err)

	if _, err := c.f.Write([]byte(fmt.Sprintf("%s:%s\n", key, storePath))); err != nil {
		_ = c.f.Close() // ignore error; Write error takes precedence
		errz.Fatal(err)
	}
	c.db[key] = storePath

	return nil
}

// Close closes the file used in cache
func (c *fileCacheStore) Close() error {
	return c.f.Close()
}

// GenerateKey generates key for the cache for a Dependency
//
// if it's a .nix file it will hash the nixpkgs + file contents
// if it's a package name will hash the packageName:nixpkgs content
func GenerateKey(dependency Dependency) (_ string, err error) {
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
