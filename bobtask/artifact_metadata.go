package bobtask

import "time"

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
}

func NewArtifactMetadata() *ArtifactMetadata {
	am := &ArtifactMetadata{
		CreatedAt: time.Now(),
	}
	return am
}
