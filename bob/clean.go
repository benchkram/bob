package bob

import (
	"context"

	"github.com/benchkram/errz"
)

// Clean will clear build info store and local store for all projects
func (b B) Clean() (err error) {
	defer errz.Recover(&err)

	err = b.CleanBuildInfoStore("")
	errz.Fatal(err)
	err = b.CleanLocalStore("")
	errz.Fatal(err)

	return nil
}

// CleanProject will clear build info store and local store for project with name project
func (b B) CleanProject(project string) (err error) {
	defer errz.Recover(&err)

	err = b.CleanBuildInfoStore(project)
	errz.Fatal(err)
	err = b.CleanLocalStore(project)
	errz.Fatal(err)

	return nil
}

// CleanBuildInfoStore will clear build info store for project. If project is empty
// it clears for all projects
func (b B) CleanBuildInfoStore(project string) error {
	return b.buildInfoStore.Clean(project)
}

// CleanLocalStore will clear artifacts for project. If project is empty
// it clears for all projects
func (b B) CleanLocalStore(project string) error {
	return b.local.Clean(context.TODO(), project)
}
