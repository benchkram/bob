package playbook

import "github.com/benchkram/bob/pkg/boberror"

type StatusMap map[string]*Status

// walk the task tree starting at root. Following dependend tasks.
func (tsm StatusMap) walk(root string, fn func(taskname string, _ *Status, _ error) error) error {
	task, ok := tsm[root]
	if !ok {
		return boberror.ErrTaskDoesNotExistF(root)
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
