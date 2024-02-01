package filehash

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/cespare/xxhash/v2"
	"io"
	"os"
)

var (
	hashFunc = xxhash.New
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

func HashToString(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

func HashString(inp string) ([]byte, error) {
	hash := New()
	var buf bytes.Buffer
	buf.WriteString(inp)
	err := hash.AddBytes(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, err
	}
	return hash.Sum(), nil
}
