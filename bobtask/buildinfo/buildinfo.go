package buildinfo

import (
	"time"

	"github.com/benchkram/bob/bobtask/hash"
)

// Targets maps in(put) hashes to target hashes
type Targets map[hash.In]string

type TargetBuildInfo struct {
	Checksum string
	Modified time.Time
	Size     int
}

// Creator information
type Creator struct{ Taskname string }

type I struct {
	// Info holds data about the creator of this object.
	Info Creator

	Target TargetBuildInfo
}

func New() *I {
	return &I{}
}
func Make() I {
	return I{}
}
