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

func (s *S) ReadFile(path string) (r io.ReadCloser, err error) {
	return os.Open(path)
}

func (s *S) DeleteFile(path string) (err error) {
	return os.Remove(path)
}

// WriteFile writes contents from read closer to a file
// if the file exists it is replaced
func (s *S) WriteFile(path string, reader io.ReadCloser) (err error) {
	defer errz.Recover(&err)
	absPath, err := filepath.Abs(path)
	errz.Fatal(err)
	if file.Exists(absPath) {
		err = s.DeleteFile(absPath)
		errz.Fatal(err)
	}
	outFile, err := os.Create(absPath)
	errz.Fatal(err)
	_, err = io.Copy(outFile, reader)
	err = reader.Close()
	errz.Fatal(err)
	return nil
}

func (s *S) MakeDir(path string) (err error) {
	defer errz.Recover(&err)
	// if path is a file: remove it
	fi, err := os.Stat(path)
	if os.IsExist(err) {
		if !fi.IsDir() {
			err = s.DeleteFile(path)
		} // else do nothing, dir exists
	} else if os.IsNotExist(err) {
		return os.MkdirAll(path, 0774)
	} else {
		errz.Fatal(err)
	}
	return nil
}

func New() *S {
	return &S{}
}
