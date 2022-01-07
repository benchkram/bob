package bob

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"

	"github.com/Benchkram/bob/bobtask"
	"github.com/Benchkram/bob/pkg/usererror"
	"github.com/Benchkram/errz"
)

// ArtifactList list artifacts belonging to each tasks.
// Artifacts are matched by project & taskname as well as their input hash stored
// in the artifacts metadata if required.
func (b *B) ArtifactList(ctx context.Context) (description string, err error) {
	defer errz.Recover(&err)

	bobfile, err := b.Aggregate()
	errz.Fatal(err)

	items, err := b.Localstore().List(ctx)
	errz.Fatal(err)

	metadataAll := []*bobtask.ArtifactMetadata{}
	// prepare projectTasknameMap once from artifact store
	for _, item := range items {
		artifact, err := b.Localstore().GetArtifact(ctx, item)
		errz.Fatal(err)
		defer artifact.Close()

		artifactInfo, err := bobtask.ArtifactInspectFromReader(artifact)
		errz.Fatal(err)

		m := artifactInfo.Metadata()
		if m == nil {
			continue
		}
		metadataAll = append(metadataAll, m)
	}

	// List artifacts in relation to tasknames in alphabetical order
	buf := bytes.NewBufferString("")
	sortedKeys := bobfile.BTasks.KeysSortedAlpabethically()
	for _, key := range sortedKeys {
		task := bobfile.BTasks[key]

		fmt.Fprintln(buf, task.Name())

		// additionaly check if there is a artifact match by inputHash
		for _, m := range metadataAll {
			var match bool

			if m.Project == task.Dir() && m.Taskname == task.Name() {
				match = true
			}

			// check input hash match in case we have no match yet.
			if !match {
				inputHash, err := task.HashIn()
				errz.Fatal(err)
				if m.InputHash == inputHash.String() {
					match = true
				}
			}

			if match {
				fmt.Fprintln(buf, "  "+m.InputHash)
			}
		}
	}

	return buf.String(), nil
}

func (b *B) ArtifactInspect(artifactID string) (ai bobtask.ArtifactInfo, err error) {

	artifact, err := b.local.GetArtifact(context.TODO(), artifactID)
	if err != nil {
		_, ok := err.(*fs.PathError)
		if ok {
			return ai, usererror.Wrap(bobtask.ErrArtifactDoesNotExist)
		}
		errz.Fatal(err)
	}
	defer artifact.Close()

	return bobtask.ArtifactInspectFromReader(artifact)
}
