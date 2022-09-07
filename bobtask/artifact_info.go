package bobtask

import (
	"bytes"
	"fmt"
	"time"
)

type ArtifactInfo interface {
	Metadata() *ArtifactMetadata
	String() string
}

// artifactInfo is a helper to debug artifacts
// during testing or from the cli.
type artifactInfo struct {
	// targetsFilesystem are stored locally on disk
	targetsFilesystem []string

	// targetsDocker stored in a local docker registry
	targetsDocker []string

	metadata *ArtifactMetadata
}

func newArtifactInfo() *artifactInfo {
	ai := &artifactInfo{
		targetsFilesystem: []string{},
		targetsDocker:     []string{},
	}
	return ai
}

func (ai *artifactInfo) Metadata() *ArtifactMetadata {
	return ai.metadata
}

func (ai *artifactInfo) String() string {
	buf := bytes.NewBufferString("")

	indent := "  "
	fmt.Fprintf(buf, "%s\n", "Artifact Info")

	fmt.Fprintf(buf, "%s%s\n", indent, "filesystem targets:")
	i := indent + "  "
	for _, t := range ai.targetsFilesystem {
		fmt.Fprintf(buf, "%s%s\n", i, t)
	}

	fmt.Fprintf(buf, "%s%s\n", indent, "docker targets:")
	i = indent + "  "
	for _, t := range ai.targetsDocker {
		fmt.Fprintf(buf, "%s%s\n", i, t)
	}

	fmt.Fprintf(buf, "%s%s\n", indent, "metadata:")
	i = indent + "  "
	if ai.metadata != nil {
		fmt.Fprintf(buf, "%s%s%s\n", i, "taskname: ", ai.metadata.Taskname)
		fmt.Fprintf(buf, "%s%s%s\n", i, "inputHash: ", ai.metadata.InputHash)
		fmt.Fprintf(buf, "%s%s%s\n", i, "project: ", ai.metadata.Project)
		fmt.Fprintf(buf, "%s%s%s\n", i, "createdAt: ", ai.metadata.CreatedAt.Format(time.RFC822Z))
	}

	return buf.String()
}
