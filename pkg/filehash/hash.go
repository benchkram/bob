package filehash

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/cespare/xxhash/v2"
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

func HashAsString(file string) (string, error) {
	b, err := Hash(file)
	if err != nil {
		return "", err
	}
	encryptedHash := hex.EncodeToString(b)
	return encryptedHash, nil
}
