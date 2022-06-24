package buildinfostore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/bobtask/buildinfo"
	"github.com/benchkram/errz"
)

// var ErrBuildInfoDoesNotExist = fmt.Errorf("build info does not exist")

type s struct {
	dir string
}

// New creates a filestore. The caller is responsible to pass an
// existing directory.
func New(dir string, opts ...Option) Store {
	s := &s{
		dir: dir,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(s)
	}

	return s
}

// NewBuildInfo creates a new build info file.
func (s *s) NewBuildInfo(id string, info *buildinfo.I) (err error) {
	defer errz.Recover(&err)

	b, err := json.Marshal(info)
	errz.Fatal(err)

	err = ioutil.WriteFile(filepath.Join(s.dir, id), b, 0666)
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
	errz.Fatal(err)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
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

func (s *s) Clean(projectName string) (err error) {
	defer errz.Recover(&err)

	homeDir, err := os.UserHomeDir()
	errz.Fatal(err)
	if s.dir == "/" || s.dir == homeDir {
		return fmt.Errorf("Cleanup of %s is not allowed", s.dir)
	}

	entries, err := os.ReadDir(s.dir)
	errz.Fatal(err)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if projectName == "" {
			err = os.Remove(filepath.Join(s.dir, entry.Name()))
			errz.Fatal(err)
			continue
		}

		bi, err := s.GetBuildInfo(entry.Name())
		errz.Fatal(err)

		if bi.Info.Project == projectName {
			err = os.Remove(filepath.Join(s.dir, entry.Name()))
			errz.Fatal(err)
		}
	}
	return nil
}
