package nix

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bob/global"
	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/bob/pkg/filehash"
)

type Cache struct {
	db   map[string]string
	f    *os.File
	path string
}

type CacheOption func(f *Cache)

// WithPath adds a custom file path which is used
// to store cache content on the filesystem.
func WithPath(path string) CacheOption {
	return func(n *Cache) {
		n.path = path
	}
}

// NewCacheStore initialize a Nix cache store inside dir
func NewCacheStore(opts ...CacheOption) (_ *Cache, err error) {
	defer errz.Recover(&err)

	c := Cache{
		db:   make(map[string]string),
		path: global.BobNixCacheFile,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(&c)
	}

	f, err := os.OpenFile(c.path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
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

	return &c, nil
}

// Get value from cache by its key
// Additionally also checks if path exists on the system
func (c *Cache) Get(key string) (string, bool) {
	path, ok := c.db[key]

	// Assure path exists on the filesystem.
	if ok && !file.Exists(path) {
		return path, false
	}

	return path, ok
}

// Save dependency inside the cache with its corresponding store path
func (c *Cache) Save(key string, storePath string) (err error) {
	defer errz.Recover(&err)

	if _, err := c.f.Write([]byte(fmt.Sprintf("%s:%s\n", key, storePath))); err != nil {
		_ = c.f.Close() // ignore error; Write error takes precedence
		errz.Fatal(err)
	}
	c.db[key] = storePath

	return nil
}

// Close closes the file used in cache
func (c *Cache) Close() error {
	return c.f.Close()
}

func (c *Cache) Clean() (err error) {
	defer errz.Recover(&err)
	err = c.f.Truncate(0)
	errz.Fatal(err)

	c.db = make(map[string]string)
	return nil
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
