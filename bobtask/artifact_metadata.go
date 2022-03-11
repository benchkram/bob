package bobtask

import (
	"time"

	"github.com/Benchkram/bob/bobtask/target"
)

type ArtifactMetadata struct {
	// Builder project which triggered the build of the artifact
	Builder string `yaml:"builder,omitempty"`

	// TODO: Project | Module | for now it's the local abs dir
	// when using remote repos it rather be a unique name
	// like the repo remote url.
	//
	// Builder is the project in which the task is defined
	Project string `yaml:"project,omitempty"`

	// Taskname this artifact was build for
	Taskname string `yaml:"taskname,omitempty"`

	// InputHash and unique identifier
	InputHash string `yaml:"input_hash,omitempty"`

	// CreatedAt timestamp the artifact was created
	CreatedAt time.Time `yaml:"created_at,omitempty"`

	// TargetType sets the type of target, path/docker,
	// default sets to `path`
	TargetType target.TargetType `yaml:"target_type,omitempty"`
}

func NewArtifactMetadata() *ArtifactMetadata {
	am := &ArtifactMetadata{
		CreatedAt:  time.Now(),
		TargetType: target.Path,
	}
	return am
}
