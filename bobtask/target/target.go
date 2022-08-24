package target

import (
	"fmt"

	"github.com/benchkram/bob/pkg/dockermobyutil"
)

type Target interface {
	Hash() (string, error)
	Verify() bool
	Exists() bool

	WithExpectedHash(string) Target
	WithDir(string) Target
}

type T struct {
	Paths []string   `yaml:"Paths"`
	Type  TargetType `yaml:"Type"`

	// working dir of target
	dir string

	// expectedHash is the last computed hash of the target used to verify the targets integrity.
	// Loaded from the system and created on a previous run
	expectedHash string

	// currentHash is the currenlty created hash during the run.
	// Reused to avoid multiple computations.
	currentHash string

	// dockerRegistryClient utility functions to handle requests with local docker registry
	dockerRegistryClient dockermobyutil.RegistryClient
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

func (t *T) WithDir(dir string) Target {
	t.dir = dir
	return t
}

func (t *T) WithExpectedHash(expectedHash string) Target {
	t.expectedHash = expectedHash
	return t
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
