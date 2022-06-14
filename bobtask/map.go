package bobtask

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/pkg/boberror"
	"github.com/benchkram/bob/pkg/multilinecmd"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/usererror"
)

type Map map[string]Task

// walk the task tree starting at root. Following dependend tasks.
// dependencies are expressed in local scope, level is used to resolve the taskname in global scope.
func (tm Map) Walk(root string, parentLevel string, fn func(taskname string, _ Task, _ error) error) error {
	taskname := root // filepath.Join(parentLevel, root)
	// fmt.Printf("Walk started on root %s with parentLevel: %s using taskname:%s\n", root, parentLevel, taskname)

	task, ok := tm[taskname]
	if !ok {
		return usererror.Wrap(boberror.ErrTaskDoesNotExistF(taskname))
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
		err = tm.Walk(relTaskname, level, fn)
		if err != nil {
			return err
		}
	}

	return nil
}

// Sanitize task map and write filtered & sanitized
// propertys from dirty members to plain (e.g. dirtyInputs -> filter&sanitize -> inputs)
func (tm Map) Sanitize() (err error) {
	defer errz.Recover(&err)

	for key, task := range tm {

		sanitizedExports, err := task.sanitizeExports(task.Exports)
		errz.Fatal(err)
		task.Exports = sanitizedExports

		err = task.parseTargets()
		errz.Fatal(err)

		inputs, err := task.filteredInputs()
		errz.Fatal(err)
		task.inputs = inputs

		task.cmds = multilinecmd.Split(task.CmdDirty)
		task.rebuild = task.sanitizeRebuild(task.RebuildDirty)

		tm[key] = task
	}

	return nil
}

func (tm Map) String() string {
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

func (tm Map) KeysSortedAlpabethically() (keys []string) {
	keys = make([]string, 0, len(tm))
	for key := range tm {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// CollectTasksInPipeline will collect all task names in the pipeline for task taskName
// in the tasksInPipeline slice
func (tm Map) CollectTasksInPipeline(taskName string) ([]string, error) {
	tasksInPipeline := []string{}
	err := tm.Walk(taskName, "", func(tn string, task Task, err error) error {
		if err != nil {
			return err
		}
		tasksInPipeline = append(tasksInPipeline, task.Name())
		return nil
	})

	if err != nil {
		return nil, err
	}
	return tasksInPipeline, nil
}

// CollectNixDependencies will collect all nix dependencies for task taskName
// in nixDependencies slice
func (tm Map) CollectNixDependenciesForTasks(whitelist []string) ([]nix.Dependency, error) {
	nixDependecies := []nix.Dependency{}
	for _, taskFromMap := range tm {
		if !taskFromMap.UseNix() {
			continue
		}

		// only add dependecies of whitelisted tasks.
		for _, taskName := range whitelist {
			if taskFromMap.Name() == taskName {
				nixDependecies = append(nixDependecies, taskFromMap.Dependencies()...)
			}
		}
	}

	return nixDependecies, nil
}
