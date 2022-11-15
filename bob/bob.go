package bob

import (
	"io/ioutil"
	"os"
	"runtime"

	"github.com/benchkram/bob/pkg/auth"
	"github.com/benchkram/bob/pkg/dockermobyutil"
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

	// remotestore the place to store artifacts remotly
	remote store.Store

	// buildInfoStore stores build infos for tasks.
	buildInfoStore buildinfostore.Store

	// readConfig some commands need a fully initialised bob.
	// When this is true a `.bob.workspace` file must exist,
	// usually done by calling `bob init`
	readConfig bool

	// enableCaching allows to save and load artifacts
	// from the cache Default: true
	enableCaching bool

	// allowInsecure uses http protocol for accessing the remote artifact store, if any
	allowInsecure bool

	// enablePush enables upload artifacts to remote store
	enablePush bool

	// enablePull enables the artifacts download from remote store
	enablePull bool

	// nix builds dependencies for tasks
	nix *NixBuilder

	// authStore is used to store authentication credentials for remote store
	authStore *auth.Store

	// env is a list of strings representing the environment in the form "key=value"
	env []string

	// maxParallel is the maximum number of parallel executed tasks
	maxParallel int

	// dockerRegistryClient is used to access the local docker registry
	dockerRegistryClient dockermobyutil.RegistryClient
}

func newBob(opts ...Option) *B {
	wd, err := os.Getwd()
	errz.Fatal(err)

	b := &B{
		dir:           wd,
		enableCaching: true,
		allowInsecure: false,
		maxParallel:   runtime.NumCPU(),

		dockerRegistryClient: dockermobyutil.NewRegistryClient(),
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

	authStore, err := AuthStore(baseStoreDir)
	if err != nil {
		return nil, err
	}
	bob.authStore = authStore

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

	if bob.nix == nil {
		nix, err := DefaultNix()
		if err != nil {
			return nil, err
		}
		bob.nix = nix
	}

	if bob.authStore == nil {
		fs, err := DefaultAuthStore()
		if err != nil {
			return nil, err
		}
		bob.authStore = fs
	}

	return bob, nil
}

func (b *B) Dir() string {
	return b.dir
}

func (b *B) Nix() *NixBuilder {
	return b.nix
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
