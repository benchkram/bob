package bobtask

import (
	"time"

	"github.com/benchkram/bob/bobtask/target"
)

type ArtifactMetadata struct {

	// Project is the project in which the task is defined
	Project string `yaml:"project,omitempty"`

	// Taskname this artifact was build for
	Taskname string `yaml:"taskname,omitempty"`

	// InputHash and unique identifier
	InputHash string `yaml:"input_hash,omitempty"`

	// CreatedAt timestamp the artifact was created
	CreatedAt time.Time `yaml:"created_at,omitempty"`

	// TargetType sets the type of target, path/docker,
	// default sets to `path`
	TargetType target.Type `yaml:"target_type,omitempty"`
}

func NewArtifactMetadata() *ArtifactMetadata {
	am := &ArtifactMetadata{
		CreatedAt:  time.Now(),
		TargetType: target.Path,
	}
	return am
}
