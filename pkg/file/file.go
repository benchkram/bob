package file

import (
	"os"
	"time"
)

// Exists return true when a file exists, false otherwise.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Copy file to destination
// TODO: Use io.Copy instead of completely reading, then writing the file to use Go optimizations.
func Copy(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

func LastModTime(filePath string) (time.Time, error) {
	file, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, err
	}
	return file.ModTime(), nil
}

// IsSymlink checks if the file is symbolic link
// If there is an error, it will be of type *os.PathError.
func IsSymlink(name string) (is bool, err error) {
	fileInfo, err := os.Lstat(name)
	if err != nil {
		return false, err
	}
	return fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink, nil
}
