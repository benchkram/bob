package nix

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/bob/pkg/filehash"
	"github.com/benchkram/errz"
)

// ShellCache caches the output of nix-shell command
type ShellCache struct {
	// dir is the root directory where the cache files are stored
	dir string
	// dataCache caches the file contents, so we don't need to call os.ReadFile on second read of the same file
	dataCache map[string][]byte
}

// NewShellCache creates a new instance of ShellCache
func NewShellCache(dir string) *ShellCache {
	dataCache := make(map[string][]byte)

	return &ShellCache{dir, dataCache}
}

// Save caches the output inside a file named by the key cache
func (c *ShellCache) Save(key string, output []byte) (err error) {
	defer errz.Recover(&err)

	err = os.MkdirAll(c.dir, 0775)
	errz.Fatal(err)

	err = os.WriteFile(filepath.Join(c.dir, key), output, 0644)
	errz.Fatal(err)

	return nil
}

// Get the data by cache key
// If Reading the file returns an error, empty data is returned
func (c *ShellCache) Get(key string) ([]byte, bool) {
	if i, ok := c.dataCache[filepath.Join(c.dir, key)]; ok {
		return i, true
	}

	if !file.Exists(filepath.Join(c.dir, key)) {
		return []byte{}, false
	}
	data, err := os.ReadFile(filepath.Join(c.dir, key))
	if err != nil {
		return []byte{}, false
	}

	c.dataCache[filepath.Join(c.dir, key)] = data
	return data, true
}

// GenerateKey generates key for the cache based on a list of Dependency and nix-shell command
//
// if Dependency it's a .nix file it will hash the nixpkgs + file contents
// if Dependency it's a package name will hash the packageName:nixpkgs content
func (c *ShellCache) GenerateKey(deps []Dependency, nixShellCmd string) (_ string, err error) {
	defer errz.Recover(&err)
	h := filehash.New()

	for _, dependency := range deps {
		if strings.HasSuffix(dependency.Name, ".nix") {
			err = h.AddBytes(bytes.NewBufferString(dependency.Nixpkgs))
			errz.Fatal(err)

			err = h.AddFile(dependency.Name)
			errz.Fatal(err)
		} else {
			toHash := fmt.Sprintf("%s:%s", dependency.Name, dependency.Nixpkgs)
			err = h.AddBytes(bytes.NewBufferString(toHash))
			errz.Fatal(err)
		}
	}

	err = h.AddBytes(bytes.NewBufferString(nixShellCmd))
	errz.Fatal(err)

	hashIn := hash.In(hex.EncodeToString(h.Sum()))

	return hashIn.String(), nil
}

// Clean will clean entire shell env cache
func (c *ShellCache) Clean() (err error) {
	defer errz.Recover(&err)

	if !file.Exists(c.dir) {
		return nil
	}

	entries, err := os.ReadDir(c.dir)
	errz.Fatal(err)

	for _, entry := range entries {
		err = os.Remove(filepath.Join(c.dir, entry.Name()))
		errz.Fatal(err)
	}
	return nil
}
