package bobtask

import (
	"bytes"
	"fmt"
	"time"
)

type ArtifactInfo interface {
	String() string
}

// artifactInfo is a helper to debug artifacts
// during testing or from the cli.
type artifactInfo struct {
	createdAt time.Time

	taskname string
	id       string
	targets  []string

	metadata *ArtifactMetadata
}

func newArtifactInfo() *artifactInfo {
	ai := &artifactInfo{
		targets: []string{},
	}
	return ai
}

func (ai *artifactInfo) String() string {
	buf := bytes.NewBufferString("")

	indent := "  "
	fmt.Fprintf(buf, "%s\n", "Artifact Info")

	fmt.Fprintf(buf, "%s%s%s\n", indent, "id:       ", ai.id)

	fmt.Fprintf(buf, "%s%s\n", indent, "targets:")
	i := indent + "  "
	for _, t := range ai.targets {
		fmt.Fprintf(buf, "%s%s\n", i, t)
	}

	fmt.Fprintf(buf, "%s%s\n", indent, "metadata:")
	i = indent + "  "
	if ai.metadata != nil {
		fmt.Fprintf(buf, "%s%s%s\n", i, "taskname: ", ai.metadata.Taskname)
		fmt.Fprintf(buf, "%s%s%s\n", i, "createdAt: ", ai.metadata.CreatedAt.String())
	}

	return buf.String()
}
