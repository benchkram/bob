package artifactstest

import (
	"context"
	"os"
	"path/filepath"

	"github.com/benchkram/bob/bob"
	"github.com/benchkram/bob/bob/playbook"
	"github.com/benchkram/errz"
)

// artifactRemove a artifact from the local artifact store
func artifactRemove(id string) error {
	fs, err := os.ReadDir(artifactDir)
	if err != nil {
		return err
	}
	for _, f := range fs {
		if f.Name() == id {
			err = os.Remove(filepath.Join(artifactDir, f.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// artifactExists checks if a artifact exists in the local artifact store
func artifactExists(id string) (exist bool, _ error) {
	fs, err := os.ReadDir(artifactDir)
	if err != nil {
		return false, err
	}

	for _, f := range fs {
		if f.Name() == id {
			exist = true
			break
		}
	}

	return exist, nil
}

// targetChanged appends a string to a target
func targetChange(dir string) error {
	f, err := os.OpenFile(dir, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	_, err = f.WriteString("change_the_target")
	if err != nil {
		return err
	}
	return f.Close()
}

// buildTask and returns it's state
func buildTask(b *bob.B, taskname string) (_ *playbook.Status, err error) {
	defer errz.Recover(&err)

	aggregate, err := b.Aggregate()
	errz.Fatal(err)
	pb, err := aggregate.Playbook(taskname)
	errz.Fatal(err)

	err = pb.Build(context.Background())
	errz.Fatal(err)

	return pb.TaskStatus(taskname)
}
