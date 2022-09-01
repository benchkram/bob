package buildinfostore

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/bobtask/buildinfo"
	"github.com/benchkram/errz"
	"go.mongodb.org/mongo-driver/bson"
)

// var ErrBuildInfoDoesNotExist = fmt.Errorf("build info does not exist")

type b struct {
	dir string
}

// NewBsonStore creates a bsonStore. The caller is responsible to pass an
// existing directory.
func NewBsonStore(dir string) Store {
	bs := &b{
		dir: dir,
	}

	return bs
}

// NewBuildInfo creates a new build info file.
func (s *b) NewBuildInfo(id string, info *buildinfo.I) (err error) {
	defer errz.Recover(&err)

	data, err := bson.Marshal(info)
	errz.Fatal(err)

	err = ioutil.WriteFile(filepath.Join(s.dir, id), data, 0666)
	errz.Fatal(err)

	return nil
}

func (s *b) GetBuildInfo(id string) (info *buildinfo.I, err error) {
	defer errz.Recover(&err)

	info = &buildinfo.I{}

	f, err := os.Open(filepath.Join(s.dir, id))
	if err != nil {
		return nil, ErrBuildInfoDoesNotExist
	}
	errz.Fatal(err)
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	errz.Fatal(err)

	err = bson.Unmarshal(data, info)
	errz.Fatal(err)

	return info, nil
}

func (s *b) GetBuildInfos() (_ []*buildinfo.I, err error) {
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
		err = bson.Unmarshal(b, bi)
		errz.Fatal(err)

		buildinfos = append(buildinfos, bi)
	}

	return buildinfos, nil
}

func (s *b) Clean() (err error) {
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
