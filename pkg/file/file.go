package file

import (
	"io/ioutil"
	"os"

	"github.com/benchkram/errz"
)

// Exists return true when a file exists, false otherwise.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Copy file to destination
// TODO: Use io.Copy instead of completely reading, then writing the file to use Go optimizations.
func Copy(src, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dst, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

// IsSymlink checks if the file is symbolic link
// If there is an error, it will be of type *os.PathError.
func IsSymlink(name string) (is bool, err error) {
	defer errz.Recover(&err)

	fileInfo, err := os.Lstat(name)
	errz.Fatal(err)

	return fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink, nil
}
