package filehash

import (
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
)

type H struct {
	hash   hash.Hash
	buffer []byte
}

func New() *H {
	return &H{
		hash:   hashFunc(),
		buffer: make([]byte, 32*1024),
	}
}

func (h *H) AddFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open: %w", err)
	}

	err = h.AddBytes(f)
	f.Close() // avoiding defer for performance

	return err
}

func (h *H) AddBytes(r io.Reader) error {
	if _, err := io.CopyBuffer(h.hash, r, h.buffer); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}
	return nil
}

func (h *H) Sum() []byte {
	return h.hash.Sum(nil)
}

// HashOfFile gives hash of a file content
func HashOfFile(path string) (string, error) {
	h := New()
	err := h.AddFile(path)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum()), nil
}

// HashOfFiles gives hash of multiple files content
func HashOfFiles(paths ...string) (string, error) {
	h := New()

	for _, path := range paths {
		err := h.AddFile(path)
		if err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(h.Sum()), nil
}
