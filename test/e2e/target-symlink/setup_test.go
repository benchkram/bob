package targetsymlinktest

import (
	"os"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/errz"
)

func BobSetup() (_ *bob.B, err error) {
	defer errz.Recover(&err)

	return bob.Bob(
		bob.WithDir(dir),
		bob.WithCachingEnabled(true),
		bob.WithFilestore(artifactStore),
		bob.WithBuildinfoStore(buildInfoStore),
	)
}

// useBobfile sets the right bobfile to be used for test
func useBobfile(name string) error {
	return os.Rename(name+".yaml", "bob.yaml")
}

// releaseBobfile will revert changes done in useBobfile
func releaseBobfile(name string) error {
	return os.Rename("bob.yaml", name+".yaml")
}

// readDir returns the dir entries as string array
func readDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []string{}, err
	}

	var contents []string
	for _, e := range entries {
		contents = append(contents, e.Name())
	}
	return contents, nil
}
