package bobtask

import (
	"time"
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

	// Files of the target
	Files []File `yaml:"files"`
}

type File struct {
	Path string `yaml:"path"`
	Hash string `yaml:"hash"`
}

func NewArtifactMetadata() *ArtifactMetadata {
	am := &ArtifactMetadata{
		CreatedAt: time.Now(),
		Files:     make([]File, 0),
	}
	return am
}
