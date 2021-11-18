package bobtask

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/Benchkram/bob/bob/global"
	"github.com/Benchkram/bob/bobtask/target"
	"github.com/Benchkram/errz"
	"github.com/mholt/archiver/v3"
)

// Target takes care of populating the targets members correctly.
// It returns a nil in case of a not existing target and a nil error.
func (t *Task) Target() (empty target.Target, _ error) {
	if t.target == nil {
		return empty, nil
	}

	hash, err := t.ReadHash()
	if err != nil {
		if errors.Is(err, ErrHashesFileDoesNotExist) || errors.Is(err, ErrTaskHashDoesNotExist) {
			return t.target.WithDir(t.dir), nil
		}
		return empty, err
	}

	targetHash, ok := hash.Targets[t.name]
	if !ok {
		return t.target.WithDir(t.dir), nil
	}

	return t.target.WithDir(t.dir).WithHash(targetHash), nil
}

// Clean the tragets defined by this task.
// This assures that we can be sure a target was correctly created
// and has not been there before the task ran.
func (t *Task) Clean() error {
	if t.target != nil {
		for _, f := range t.target.Paths {
			if t.dir == "" {
				panic("Task dir not set")
			}
			p := filepath.Join(t.dir, f)
			if p == "/" {
				panic("Root cleanup is permitted")
			}

			//fmt.Printf("Cleaning %s ", p)
			err := os.RemoveAll(p)
			if err != nil {
				//fmt.Printf("%s\n", aurora.Red("failed"))
				return err
			}
			//fmt.Printf("%s\n", aurora.Green("done"))
		}
	}

	return nil
}

// Pack creates a archive for a target.
// Does nothing and returns nil is traget is undefined.
func (t *Task) Pack(hash string) (err error) {
	defer errz.Recover(&err)

	if t.target == nil {
		return nil
	}

	paths := []string{}
	for _, path := range t.target.Paths {
		paths = append(paths, filepath.Join(t.dir, path))
	}

	archive := filepath.Join(t.dir, global.BobCacheDir, hash+".tar.br")
	err = os.RemoveAll(archive)
	errz.Fatal(err)
	err = archiver.Archive(paths, archive)
	errz.Fatal(err)

	return nil
}

// Unpack target from archive
func (t *Task) Unpack(hash string) (err error) {
	archive := filepath.Join(t.dir, global.BobCacheDir, hash+".tar.br")
	return archiver.Unarchive(archive, t.dir)
}
