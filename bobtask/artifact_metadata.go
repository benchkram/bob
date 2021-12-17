package bobtask

import "time"

type ArtifactMetadata struct {

	// TODO: Project | Module
	Project string `yaml:"project,omitempty"`

	// Taskname this artifact was build for
	Taskname string `yaml:"taskname,omitempty"`

	// Sub path, set when created as part of
	// another parent project.
	Sub string `yaml:"sub,omitempty"`

	// CreatedAt timestamp the artifact was created
	CreatedAt time.Time `yaml:"created_at,omitempty"`
}

func NewArtifactMetadata() *ArtifactMetadata {
	am := &ArtifactMetadata{
		CreatedAt: time.Now(),
	}
	return am
}
