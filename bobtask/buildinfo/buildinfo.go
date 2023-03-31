package buildinfo

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/benchkram/bob/bobtask/buildinfo/protos"
)

type I struct {
	// Meta holds data about the creator/origin of this object.
	Meta Meta

	// Target aggregates buildinfos of multiple files or docker images
	Target Targets
}

func New() *I {
	return &I{
		Target: MakeTargets(),
	}
}

func (i *I) Describe() string {
	buf := bytes.NewBufferString("")

	fmt.Fprintln(buf, "Meta:")
	fmt.Fprintln(buf, "\ttask:", i.Meta.Task)
	fmt.Fprintln(buf, "\tinput hash", i.Meta.InputHash)

	fmt.Fprintln(buf, "Filesystem:")
	fmt.Fprintln(buf, "\thash of all files", i.Target.Filesystem.Hash)
	fmt.Fprintln(buf, "\t# of files", len(i.Target.Filesystem.Files))
	fmt.Fprintln(buf, "\tfiles:")

	sortedFiles := []string{}
	for filename := range i.Target.Filesystem.Files {
		sortedFiles = append(sortedFiles, filename)
	}
	sort.Strings(sortedFiles)

	for _, filename := range sortedFiles {
		v := i.Target.Filesystem.Files[filename]
		fmt.Fprintln(buf, "\t", filename, v.Size)
	}

	return buf.String()
}

type Targets struct {
	Filesystem BuildInfoFiles             `yaml:"file"`
	Docker     map[string]BuildInfoDocker `yaml:"docker"`
}

func NewTargets() *Targets {
	return &Targets{
		Filesystem: MakeBuildInfoFiles(),
		Docker:     make(map[string]BuildInfoDocker),
	}
}

func MakeTargets() Targets {
	return *NewTargets()
}

type BuildInfoFiles struct {
	// Hash contains the hash of all files
	Hash string `yaml:"hash"`
	// Files contains modtime & size of each file
	Files map[string]BuildInfoFile `yaml:"file"`
}

func NewBuildInfoFiles() *BuildInfoFiles {
	return &BuildInfoFiles{
		Files: make(map[string]BuildInfoFile),
	}
}
func MakeBuildInfoFiles() BuildInfoFiles {
	return *NewBuildInfoFiles()
}

type BuildInfoFile struct {
	// Size of a file
	Size int64 `yaml:"size"`
}

type BuildInfoDocker struct {
	Hash string `yaml:"hash"`
}

// Creator information
type Meta struct {
	// Task usually the taskname
	Task string `yaml:"task"`

	// InputHash used for target creation
	InputHash string `yaml:"input_hash"`
}

func (i *I) ToProto(inputHash string) *protos.BuildInfo {

	filesystem := &protos.BuildInfoFiles{
		Targets: make(map[string]*protos.BuildInfoFile, len(i.Target.Filesystem.Files)),
	}
	filesystem.Hash = i.Target.Filesystem.Hash
	for k, v := range i.Target.Filesystem.Files {
		filesystem.Targets[k] = &protos.BuildInfoFile{Size: v.Size}
	}

	docker := make(map[string]*protos.BuildInfoDocker)
	for k, v := range i.Target.Docker {
		docker[k] = &protos.BuildInfoDocker{Hash: v.Hash}
	}

	return &protos.BuildInfo{
		Meta: &protos.Meta{
			Task:      i.Meta.Task,
			InputHash: inputHash,
		},
		Target: &protos.Targets{
			Filesystem: filesystem,
			Docker:     docker,
		},
	}
}

func FromProto(p *protos.BuildInfo) *I {
	if p == nil {
		return nil
	}

	bi := New()

	if p.Meta != nil {
		bi.Meta.Task = p.Meta.Task
		bi.Meta.InputHash = p.Meta.InputHash
	}

	if p.Target != nil {
		if p.Target.Filesystem != nil {
			bi.Target.Filesystem.Hash = p.Target.Filesystem.Hash
			for k, v := range p.Target.Filesystem.Targets {
				bi.Target.Filesystem.Files[k] = BuildInfoFile{Size: v.Size}
			}
		}

		for k, v := range p.Target.Docker {
			bi.Target.Docker[k] = BuildInfoDocker{Hash: v.Hash}
		}
	}

	return bi
}
