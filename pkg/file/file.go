package file

import (
	"io/ioutil"
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
func Copy(dst, src string) error {
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

func LastModTime(filePath string) (time.Time, error) {
	file, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, err
	}
	return file.ModTime(), nil
}
