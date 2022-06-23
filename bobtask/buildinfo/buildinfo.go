package buildinfo

import (
	"github.com/benchkram/bob/bobtask/hash"
)

// Targets maps in(put) hashes to target hashes
type Targets map[hash.In]string

// Creator information
type Creator struct {
	Taskname string
	Project  string
}

type I struct {
	// Info holds data about the creator of this object.
	Info Creator

	// Targets hold hash values on all related
	// targets in the build chain.
	Targets Targets
}

func New() *I {
	return &I{
		Targets: make(Targets),
	}
}
