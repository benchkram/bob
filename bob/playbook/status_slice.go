package playbook

import "fmt"

type StatusSlice []*Status

var ErrWalkDone = fmt.Errorf("walking done")

// // walk the task tree starting at root. Following dependend tasks.
// func (tsm StatusSlice) walk(root int, fn func(taskID int, _ *Status, _ error) error) error {
// 	task := tsm[root]

// 	err := fn(root, task, nil)
// 	if err != nil {
// 		return err
// 	}
// 	for _, id := range task.Task.DependsOnIDs {
// 		err = tsm.walk(id, fn)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// walk the task tree starting at root. Tasks deeper in the tree are walked first.
func (tsm StatusSlice) walkBottomFirst(root int, fn func(taskID int, _ *Status, _ error) error) error {
	task := tsm[root]

	var err error
	for _, id := range task.Task.DependsOnIDs {
		err = tsm.walkBottomFirst(id, fn)
		if err != nil {
			return err
		}
	}

	return fn(root, task, err)
}
