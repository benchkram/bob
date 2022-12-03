package playbook

import (
	"fmt"

	"github.com/benchkram/bob/pkg/boberror"
	"github.com/benchkram/bob/pkg/usererror"
)

type StatusMap map[string]*Status

var ErrWalkDone = fmt.Errorf("walking done")

// walk the task tree starting at root. Following dependend tasks.
func (tsm StatusMap) walk(root string, fn func(taskname string, _ *Status, _ error) error) error {
	task, ok := tsm[root]
	if !ok {
		return usererror.Wrap(boberror.ErrTaskDoesNotExistF(root))
	}

	err := fn(root, task, nil)
	if err != nil {
		return err
	}
	for _, dependentTaskName := range task.Task.DependsOn {
		err = tsm.walk(dependentTaskName, fn)
		if err != nil {
			return err
		}
	}

	return nil
}
