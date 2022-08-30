package target

import (
	"github.com/benchkram/bob/bobtask/buildinfo"
	"github.com/benchkram/bob/bobtask/targettype"
	"github.com/benchkram/bob/pkg/dockermobyutil"
)

type Target interface {
	BuildInfo() (*buildinfo.Targets, error)

	Verify() bool
	// Exists() bool

	// WithExpectedHash(string) Target
	// WithDir(string) Target

	// Paths() []string
	// PathsPlain() []string
	// Type() targettype.T
}

type T struct {
	// working dir of target
	dir string

	// expected is the last valid buildinfo of the target used to verify the targets integrity.
	// Loaded from the system and created on a previous run. Can be nil.
	expected *buildinfo.Targets

	// current is the currenlty created buildInfo during the run.
	// current avoids recomputations.
	current *buildinfo.Targets

	// dockerRegistryClient utility functions to handle requests with local docker registry
	dockerRegistryClient dockermobyutil.RegistryClient

	// dockerImages an array of docker tags
	dockerImages []string
	// filesystemEntries an array of filesystem entries,
	// E.g. files or entire directories.
	filesystemEntries []string

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

const DefaultType = targettype.Path

func (t *T) clone() *T {
	target := New()
	target.dir = t.dir
	// target.PathsSerialize = t.PathsSerialize
	// target.TypeSerialize = t.TypeSerialize
	return target
}

func (t *T) WithDir(dir string) Target {
	t.dir = dir
	return t
}

// func (t *T) WithExpectedHash(expectedHash string) Target {
// 	t.expectedHash = expectedHash
// 	return t
// }

// // Paths in relation to the umrella bobfile
// func (t *T) Paths() []string {
// 	if len(t.PathsSerialize) == 0 {
// 		return []string{}
// 	}

// 	var pathsWithDir []string
// 	for _, v := range t.PathsSerialize {
// 		pathsWithDir = append(pathsWithDir, filepath.Join(t.dir, v))
// 	}

// 	return pathsWithDir
// }

// // PathsPlain does return the pure target path
// // as given in the bobfile.
// func (t *T) PathsPlain() []string {
// 	return t.PathsSerialize
// }

// func (t *T) Type() targettype.T {
// 	return t.TypeSerialize
// }
