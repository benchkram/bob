package playbook

type StatusSlice []*Status

//var ErrWalkDone = fmt.Errorf("walking done")

// walk the task tree starting at root. Following dependend tasks.
func (tsm StatusSlice) walk(root int, fn func(taskID int, _ *Status, _ error) error) error {
	task := tsm[root]

	err := fn(root, task, nil)
	if err != nil {
		return err
	}
	for _, id := range task.Task.DependsOnIDs {
		err = tsm.walk(id, fn)
		if err != nil {
			return err
		}
	}

	return nil
}

// // walk the task tree starting at root. Tasks deeper in the tree are walked first.
// func (tsm StatusMap) walkBottomFirst(root string, fn func(taskname string, _ *Status, _ error) error) error {
// 	task, ok := tsm[root]
// 	if !ok {
// 		return usererror.Wrap(boberror.ErrTaskDoesNotExistF(root))
// 	}

// 	var err error
// 	for _, dependentTaskName := range task.Task.DependsOn {
// 		err = tsm.walk(dependentTaskName, fn)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return fn(root, task, err)
// }
