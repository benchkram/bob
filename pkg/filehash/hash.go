package filehash

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
)

var (
	// Using sha1 shouldn't be a problem as hash collisions are very unlikely.
	// sha1 is about twice as fast as sha256 on my — Leon's — machine.
	hashFunc = sha1.New
)

func Hash(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open: %w", err)
	}
	defer f.Close()

	return HashBytes(io.Reader(f))
}

func HashBytes(r io.Reader) ([]byte, error) {
	h := hashFunc()
	if _, err := io.Copy(h, r); err != nil {
		return nil, fmt.Errorf("failed to copy: %w", err)
	}

	return h.Sum(nil), nil
}
