package buildinfostore

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/bobtask/buildinfo"
	"github.com/benchkram/bob/bobtask/buildinfo/protos"
	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/errz"

	"google.golang.org/protobuf/proto"
)

type ps struct {
	dir string
}

// NewProtoStore creates a proto store. The caller is responsible to pass an existing directory
func NewProtoStore(dir string) Store {
	return &ps{dir: dir}
}

// NewBuildInfo creates a new build info file.
func (ps *ps) NewBuildInfo(id string, info *buildinfo.I) (err error) {
	defer errz.Recover(&err)

	m := info.ToProto(id)

	b, err := proto.Marshal(m)
	errz.Fatal(err)

	err = ioutil.WriteFile(filepath.Join(ps.dir, id), b, 0666)
	errz.Fatal(err)

	return nil
}

// GetArtifact opens a file
func (ps *ps) GetBuildInfo(id string) (info *buildinfo.I, err error) {
	defer errz.Recover(&err)

	f, err := os.Open(filepath.Join(ps.dir, id))
	if err != nil {
		return nil, ErrBuildInfoDoesNotExist
	}
	errz.Fatal(err)
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	errz.Fatal(err)

	protoInfo := &protos.BuildInfo{}

	err = proto.Unmarshal(b, protoInfo)
	errz.Fatal(err)

	return buildinfo.FromProto(protoInfo), nil
}

func (ps *ps) GetBuildInfos() (_ []*buildinfo.I, err error) {
	defer errz.Recover(&err)

	entries, err := os.ReadDir(ps.dir)
	errz.Fatal(err)

	var buildInfos []*buildinfo.I
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(ps.dir, entry.Name()))
		errz.Fatal(err)

		bi := &protos.BuildInfo{}
		err = proto.Unmarshal(data, bi)
		errz.Fatal(err)

		buildInfos = append(buildInfos, buildinfo.FromProto(bi))
	}

	return buildInfos, nil
}

func (ps *ps) Clean() (err error) {
	defer errz.Recover(&err)

	homeDir, err := os.UserHomeDir()
	errz.Fatal(err)
	if ps.dir == "/" || ps.dir == homeDir {
		return fmt.Errorf("Cleanup of %s is not allowed", ps.dir)
	}

	entrys, err := os.ReadDir(ps.dir)
	errz.Fatal(err)

	for _, entry := range entrys {
		if entry.IsDir() {
			continue
		}
		_ = os.Remove(filepath.Join(ps.dir, entry.Name()))
	}

	return nil
}

func (ps *ps) BuildInfoExists(id string) bool {
	return file.Exists(filepath.Join(ps.dir, id))
}
