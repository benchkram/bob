package buildinfo

import (
	"time"
)

type Targets struct {
	Filesystem BuildInfoFiles             `yaml:"file"`
	Docker     map[string]BuildInfoDocker `yaml:"docker"`
}

func NewTargets() *Targets {
	return &Targets{}
}

func MakeTargets() Targets {
	return Targets{}
}

type BuildInfoFiles struct {
	// Hash contains the hash of all files
	Hash string `yaml:"hash"`
	// Files contains modtime & size of each file
	Files map[string]BuildInfoFile `yaml:"file"`
}

type BuildInfoFile struct {
	Modified time.Time `yaml:"modified"`
	Size     int64     `yaml:"size"`
}

type BuildInfoDocker struct {
	Hash string `yaml:"hash"`
}

// Creator information
type Meta struct {
	Task      string `yaml:"task"`
	InputHash string `yaml:"input_hash"`
}

type I struct {
	// Target aggregates buildinfos of multiple files or docker images
	Target Targets

	// Meta holds data about the creator of this object.
	Meta Meta
}

func New() *I {
	return &I{}
}

func Make() I {
	return I{}
}
