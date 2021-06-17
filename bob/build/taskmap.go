package build

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
)

type TaskMap map[string]Task

// walk the task tree starting at root. Following dependend tasks.
// dependencies are expressed in local scope, level is used to resolve the taskname in global scope.
func (tm TaskMap) walk(root string, parentLevel string, fn func(taskname string, _ Task, _ error) error) error {
	taskname := root //filepath.Join(parentLevel, root)
	//fmt.Printf("Walk started on root %s with parentLevel: %s using taskname:%s\n", root, parentLevel, taskname)

	task, ok := tm[taskname]
	if !ok {
		return ErrTaskDoesNotExist
	}

	err := fn(taskname, task, nil)
	if err != nil {
		return err
	}

	level := filepath.Dir(task.name)
	if level == "." {
		level = ""
	}
	for _, relTaskname := range task.DependsOn {
		err = tm.walk(relTaskname, level, fn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tm TaskMap) String() string {
	description := bytes.NewBufferString("")

	fmt.Fprint(description, "TaskMap:\n")

	keys := make([]string, 0, len(tm))
	for k := range tm {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		task := tm[k]
		fmt.Fprintf(description, "  %s(%s): -\n", k, task.name)
	}

	return description.String()
}
