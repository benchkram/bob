package filehash

import (
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
	defer f.Close()

	return h.AddBytes(f)
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
