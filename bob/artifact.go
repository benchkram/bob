package bob

import (
	"bytes"
	"context"
	"fmt"

	"github.com/Benchkram/bob/bobtask"
	"github.com/Benchkram/errz"
)

type TaskArtifactsMap map[string]*bobtask.ArtifactMetadata

func (b *B) ArtifactList() (description string, err error) {
	defer errz.Recover(&err)

	bobfile, err := b.Aggregate()
	errz.Fatal(err)

	items, err := b.Localstore().List(context.TODO())
	errz.Fatal(err)

	// projectTasknameMap helper map to the artifactname
	// in relation to a project_taskname identifier
	projectTasknameMap := make(TaskArtifactsMap)

	// prepare projectTasknameMap once from artifact store
	for _, item := range items {
		artifact, err := b.Localstore().GetArtifact(context.TODO(), item)
		errz.Fatal(err)
		defer artifact.Close()

		artifactInfo, err := bobtask.ArtifactInspectFromReader(artifact)
		errz.Fatal(err)

		m := artifactInfo.Metadata()
		if m == nil {
			continue
		}

		projectTasknameMap[taskArtifactsMapKey(m.Project, m.Taskname)] = m
	}

	// List artifacts in relation to tasknames in alphabetcal order
	buf := bytes.NewBufferString("")
	sortedKeys := bobfile.Tasks.KeysSortedAlpabethically()
	for _, key := range sortedKeys {
		task := bobfile.Tasks[key]

		fmt.Fprintln(buf, task.Name())
		metadata, ok := projectTasknameMap[taskArtifactsMapKey(task.Dir(), task.Name())]
		if ok {
			fmt.Fprintln(buf, "  "+metadata.InputHash)
		}
	}

	return buf.String(), nil
}

func taskArtifactsMapKey(projectName, taskname string) string {
	return projectName + "_" + taskname
}
