package target

import (
	"path/filepath"

	"github.com/benchkram/bob/pkg/dockermobyutil"
)

type Target interface {
	Hash() (string, error)
	Verify() bool
	Exists() bool

	WithHash(string) Target
	WithDir(string) Target

	Paths() []string
	PathsPlain() []string
	Type() Type
}

type T struct {
	// working dir of target
	dir string

	// last computed hash of target
	hash string

	// dockerRegistryClient utility functions to handle requests with local docker registry
	dockerRegistryClient dockermobyutil.RegistryClient

	// exposed due to yaml marshalling
	PathsSerialize []string `yaml:"Paths"`
	TypeSerialize  Type     `yaml:"Type"`
}

func New(opts ...Option) *T {
	t := &T{
		dockerRegistryClient: dockermobyutil.NewRegistryClient(),
		PathsSerialize:       []string{},
		TypeSerialize:        Path,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(t)
	}

	return t
}

type Type string

const (
	Path   Type = "path"
	Docker Type = "docker"
)

const DefaultType = Path

func (t *T) clone() *T {
	target := New()
	target.dir = t.dir
	target.PathsSerialize = t.PathsSerialize
	target.TypeSerialize = t.TypeSerialize
	return target
}

func (t *T) WithDir(dir string) Target {
	target := t.clone()
	target.dir = dir
	return target
}
func (t *T) WithHash(hash string) Target {
	target := t.clone()
	target.hash = hash
	return target
}

// Paths in relation to the umrella bobfile
func (t *T) Paths() []string {
	if len(t.PathsSerialize) == 0 {
		return []string{}
	}

	var pathsWithDir []string
	for _, v := range t.PathsSerialize {
		pathsWithDir = append(pathsWithDir, filepath.Join(t.dir, v))
	}

	return pathsWithDir
}

// PathsPlain does return the pure target path
// as given in the bobfile.
func (t *T) PathsPlain() []string {
	return t.PathsSerialize
}

func (t *T) Type() Type {
	return t.TypeSerialize
}
