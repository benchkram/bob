package target

import (
	"path/filepath"

	"github.com/benchkram/bob/bobtask/buildinfo"
	"github.com/benchkram/bob/pkg/dockermobyutil"
)

type Target interface {
	BuildInfo() (*buildinfo.Targets, error)

	//Verify() bool
	VerifyShallow() bool
	Resolve() error

	FilesystemEntries() []string
	FilesystemEntriesPlain() []string
	FilesystemEntriesRaw() []string
	FilesystemEntriesRawPlain() []string
	// Exists() bool

	WithExpected(*buildinfo.Targets) *T

	// Paths() []string
	// PathsPlain() []string
	// Type() targettype.T
	DockerImages() []string
}

type T struct {
	// working dir of target
	dir string

	// expected is the last valid buildinfo of the target used to verify the targets integrity.
	// Loaded from the system and created on a previous run. Can be nil.
	expected *buildinfo.Targets

	// current is the currenlty created buildInfo during the run.
	// current avoids recomputations.
	// current *buildinfo.Targets

	// dockerRegistryClient utility functions to handle requests with local docker registry
	dockerRegistryClient dockermobyutil.RegistryClient

	// dockerImages an array of docker tags
	dockerImages []string
	// filesystemEntries is an array of files,
	// read from the filesystem.
	// resolve(filesystemEntriesRaw) = filesystemEntriesRaw.
	//
	// Usually the first required when IgnoreChildtargets() is called
	// on aggregate level.
	filesystemEntries *[]string
	// filesystemEntriesRaw is an array of files or directories,
	// as defined by the user.
	//
	// Used to verify that targets are created
	// without verifying against expected buildinfo.
	filesystemEntriesRaw []string

	// exposed due to yaml marshalling
	// PathsSerialize []string     `yaml:"Paths"`
	// TypeSerialize  targettype.T `yaml:"Type"`
}

func New(opts ...Option) *T {
	t := &T{
		dockerRegistryClient: dockermobyutil.NewRegistryClient(),
		// PathsSerialize:       []string{},
		// TypeSerialize:        DefaultType,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(t)
	}

	return t
}

// FilesystemEntries in relation to the umrella bobfile
func (t *T) FilesystemEntries() []string {

	if len(*t.filesystemEntries) == 0 {
		return []string{}
	}

	var pathsWithDir []string
	for _, v := range *t.filesystemEntries {
		pathsWithDir = append(pathsWithDir, filepath.Join(t.dir, v))
	}

	return pathsWithDir
}

func (t *T) FilesystemEntriesRaw() []string {
	var pathsWithDir []string
	for _, v := range t.filesystemEntriesRaw {
		pathsWithDir = append(pathsWithDir, filepath.Join(t.dir, v))
	}

	return pathsWithDir
}

// FilesystemEntriesPlain does return the pure path
// as given in the bobfile.
func (t *T) FilesystemEntriesPlain() []string {
	return *t.filesystemEntries
}

func (t *T) FilesystemEntriesRawPlain() []string {
	return t.filesystemEntriesRaw
}

func (t *T) WithExpected(expected *buildinfo.Targets) *T {
	t.expected = expected
	return t
}

func (t *T) DockerImages() []string {
	return t.dockerImages
}
