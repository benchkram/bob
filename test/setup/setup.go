package setup

import (
	"io/ioutil"
	"os"
)

// TestDirs setup a general test dir and a "out-of-tree" storage dir used in tests.
// Call cleanup() to delete all dirs at the end of the test.
func TestDirs(testname string) (testDir, storageDir string, cleanup func() error, _ error) {
	plain := func() error { return nil }

	testDir, err := ioutil.TempDir("", "bob-test-"+testname+"-*")
	if err != nil {
		return testDir, storageDir, plain, err
	}

	storageDir, err = ioutil.TempDir("", "bob-test-"+testname+"-storage-*")
	if err != nil {
		return testDir, storageDir, plain, err
	}

	cleanup = func() (err error) {
		err = os.RemoveAll(testDir)
		if err != nil {
			return err
		}
		err = os.RemoveAll(storageDir)
		if err != nil {
			return err
		}
		return nil
	}
	return testDir, storageDir, cleanup, nil
}
