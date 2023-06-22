package buildinfostore

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/bobtask/buildinfo"
	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/errz"
)

// var ErrBuildInfoDoesNotExist = fmt.Errorf("build info does not exist")

type s struct {
	// dir is the base directory of the store
	dir string
}

// New creates a filestore. The caller is responsible to pass a
// existing directory.
func New(dir string) Store {
	s := &s{
		dir: dir,
	}

	// for _, opt := range opts {
	// 	if opt == nil {
	// 		continue
	// 	}
	// 	opt(s)
	// }

	return s
}

// NewBuildInfo creates a new build info file.
func (s *s) NewBuildInfo(id string, info *buildinfo.I) (err error) {
	defer errz.Recover(&err)

	b, err := json.Marshal(info)
	errz.Fatal(err)

	err = os.WriteFile(filepath.Join(s.dir, id), b, 0666)
	errz.Fatal(err)

	return nil
}

// GetArtifact opens a file
func (s *s) GetBuildInfo(id string) (info *buildinfo.I, err error) {
	defer errz.Recover(&err)

	info = &buildinfo.I{}

	f, err := os.Open(filepath.Join(s.dir, id))
	if err != nil {
		return nil, ErrBuildInfoDoesNotExist
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	errz.Fatal(err)

	err = json.Unmarshal(b, info)
	errz.Fatal(err)

	return info, nil
}

func (s *s) GetBuildInfos() (_ []*buildinfo.I, err error) {
	defer errz.Recover(&err)

	buildinfos := []*buildinfo.I{}

	entrys, err := os.ReadDir(s.dir)
	errz.Fatal(err)

	for _, entry := range entrys {
		if entry.IsDir() {
			continue
		}

		b, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
		errz.Fatal(err)

		bi := &buildinfo.I{}
		err = json.Unmarshal(b, bi)
		errz.Fatal(err)

		buildinfos = append(buildinfos, bi)
	}

	return buildinfos, nil
}

func (s *s) Clean() (err error) {
	defer errz.Recover(&err)

	homeDir, err := os.UserHomeDir()
	errz.Fatal(err)
	if s.dir == "/" || s.dir == homeDir {
		return fmt.Errorf("Cleanup of %s is not allowed", s.dir)
	}

	entrys, err := os.ReadDir(s.dir)
	errz.Fatal(err)

	for _, entry := range entrys {
		if entry.IsDir() {
			continue
		}
		_ = os.Remove(filepath.Join(s.dir, entry.Name()))
	}

	return nil
}

func (s *s) BuildInfoExists(id string) bool {
	return file.Exists(filepath.Join(s.dir, id))
}
