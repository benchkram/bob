package localsyncstore

import (
	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/errz"
	"io"
	"os"
	"path/filepath"
)

type S struct {
}

func (s *S) ReadFile(bobDir, collectionPath, localPath string) (r io.ReadCloser, err error) {
	return os.Open(filepath.Join(bobDir, collectionPath, localPath))
}

func (s *S) DeleteFile(bobDir, collectionPath, localPath string) (err error) {
	return os.Remove(filepath.Join(bobDir, collectionPath, localPath))
}

// WriteFile writes contents from read closer to a file
// if the file exists it is replaced
func (s *S) WriteFile(bobDir, collectionPath, localPath string, reader io.ReadCloser) (err error) {
	defer errz.Recover(&err)
	absPath := filepath.Join(bobDir, collectionPath, localPath)
	if file.Exists(absPath) {
		err = s.DeleteFile(bobDir, collectionPath, localPath)
		errz.Fatal(err)
	}
	outFile, err := os.Create(absPath)
	errz.Fatal(err)
	_, err = io.Copy(outFile, reader)
	err = reader.Close()
	errz.Fatal(err)
	return nil
}

func New() *S {
	return &S{}
}
