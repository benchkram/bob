package buildinfo

import (
	"time"
)

// Targets maps in(put) hashes to target hashes
//type Targets map[hash.In][]TargetBuildInfo
type Targets struct {
	File   map[string]BuildInfoFile   `yaml:"file"`
	Docker map[string]BuildInfoDocker `yaml:"docker"`
}

type BuildInfoFile struct {
	Hash     string    `yaml:"hash"`
	Modified time.Time `yaml:"modified"`
	Size     int       `yaml:"size"`
}

type BuildInfoDocker struct {
	Hash string `yaml:"hash"`
}

// Creator information
type Creator struct {
	Taskname  string
	InputHash string
}

type I struct {
	// Info holds data about the creator of this object.
	Info Creator

	Target Targets
}

func New() *I {
	return &I{}
}

func Make() I {
	return I{}
}
