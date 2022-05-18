package bob

import (
	"io/ioutil"
	"os"

	"github.com/benchkram/bob/pkg/usererror"

	"github.com/hashicorp/go-version"

	"gopkg.in/yaml.v3"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/pkg/buildinfostore"
	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/bob/pkg/store"
)

var Version = "0.0.0"

type B struct {
	// Repositories to track.
	Repositories []Repo `yaml:"repositories"`

	// dir is bob's working directory.
	dir string

	// local the place to store artifacts localy
	local store.Store
	// TODO: add a remote store
	// remotestore cas.Store

	buildInfoStore buildinfostore.Store

	// readConfig some commands need a fully initialised bob.
	// When this is true a `.bob.workspace` file must exist,
	// usually done by calling `bob init`
	readConfig bool

	// enableCaching allows to save and load artifacts
	// from the cache Default: true
	enableCaching bool
}

func newBob(opts ...Option) *B {
	wd, err := os.Getwd()
	if err != nil {
		errz.Fatal(err)
	}

	b := &B{
		dir:           wd,
		enableCaching: true,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(b)
	}

	return b
}

// BobWithBaseStoreDir initialises stores in the given directory
func BobWithBaseStoreDir(baseStoreDir string, opts ...Option) (*B, error) {
	bob := newBob(opts...)

	_, err := version.NewVersion(Version)
	if err != nil {
		return nil, ErrInvalidVersion
	}

	fs, err := Filestore(baseStoreDir)
	if err != nil {
		return nil, err
	}
	bob.local = fs

	bis, err := BuildinfoStore(baseStoreDir)
	if err != nil {
		return nil, err
	}
	bob.buildInfoStore = bis

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(bob)
	}

	return bob, nil
}

func Bob(opts ...Option) (*B, error) {
	bob := newBob(opts...)

	_, err := version.NewVersion(Version)
	if err != nil {
		return nil, ErrInvalidVersion
	}

	if bob.readConfig {
		err := bob.read()
		if err != nil {
			return nil, err
		}
	}

	if bob.local == nil {
		fs, err := DefaultFilestore()
		if err != nil {
			return nil, err
		}
		bob.local = fs
	}

	if bob.buildInfoStore == nil {
		bis, err := DefaultBuildinfoStore()
		if err != nil {
			return nil, err
		}
		bob.buildInfoStore = bis
	}

	return bob, nil
}

func (b *B) Dir() string {
	return b.dir
}

func (b *B) write() (err error) {
	defer errz.Recover(&err)

	bin, err := yaml.Marshal(b)
	errz.Fatal(err)

	const mode = 0644
	return ioutil.WriteFile(b.WorkspaceFilePath(), bin, mode)
}

func (b *B) read() (err error) {
	defer errz.Recover(&err)

	if !file.Exists(b.WorkspaceFilePath()) {
		// Initialise with default values if it does not exist.
		err := b.write()
		errz.Fatal(err)
	}

	bin, err := ioutil.ReadFile(b.WorkspaceFilePath())
	errz.Fatal(err)

	err = yaml.Unmarshal(bin, b)
	if err != nil {
		return usererror.Wrapm(err, "YAML unmarshal failed")
	}

	return nil
}
