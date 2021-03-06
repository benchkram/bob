package target

import (
	"fmt"

	"github.com/benchkram/bob/pkg/dockermobyutil"
)

type Target interface {
	Hash() (string, error)
	Verify() bool
	Exists() bool

	WithHash(string) Target
	WithDir(string) Target
}

type T struct {
	// working dir of target
	dir string

	// last computed hash of target
	hash string

	// dockerRegistryClient utility functions to handle requests with local docker registry
	dockerRegistryClient dockermobyutil.RegistryClient

	Paths []string   `yaml:"Paths"`
	Type  TargetType `yaml:"Type"`
}

func Make() T {
	return T{
		dockerRegistryClient: dockermobyutil.NewRegistryClient(),
		Paths:                []string{},
		Type:                 Path,
	}
}

func New() *T {
	return new()
}

func new() *T {
	return &T{
		dockerRegistryClient: dockermobyutil.NewRegistryClient(),
		Paths:                []string{},
		Type:                 Path,
	}
}

type TargetType string

const (
	Path   TargetType = "path"
	Docker TargetType = "docker"
)

const DefaultType = Path

func (t *T) clone() *T {
	target := new()
	target.dir = t.dir
	target.Paths = t.Paths
	target.Type = t.Type
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

func ParseType(str string) (TargetType, error) {
	switch {
	case str == string(Path):
		return Path, nil
	case str == string(Docker):
		return Docker, nil
	default:
		return DefaultType, fmt.Errorf("Invalid Target type. Only supports 'path' and 'docker-image' as type.")
	}
}
