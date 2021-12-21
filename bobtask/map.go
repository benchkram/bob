package bobtask

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Benchkram/bob/pkg/boblog"
	"github.com/Benchkram/bob/pkg/multilinecmd"
	"github.com/Benchkram/errz"
	"github.com/logrusorgru/aurora"
)

type Map map[string]Task

// walk the task tree starting at root. Following dependend tasks.
// dependencies are expressed in local scope, level is used to resolve the taskname in global scope.
func (tm Map) Walk(root string, parentLevel string, fn func(taskname string, _ Task, _ error) error) error {
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
		err = tm.Walk(relTaskname, level, fn)
		if err != nil {
			return err
		}
	}

	return nil
}

// Sanitize task map and write filtered & sanitized
// propertys from dirty members to plain (e.g. dirtyInputs -> filter&sanitize -> inputs)
func (tm Map) Sanitize() {
	for key, task := range tm {

		sanitizedExports, err := task.sanitizeExports(task.Exports)
		errz.Fatal(err)
		task.Exports = sanitizedExports

		err = task.parseTargets()
		errz.Fatal(err)

		inputs, err := task.filteredInputs()
		if errors.Is(err, BackwardPathError) || errors.Is(err, OutsideDirError) {
			boblog.Log.V(1).Info(aurora.Red(err.Error()).String())
			os.Exit(1)
		}

		errz.Fatal(err)
		task.inputs = inputs

		task.cmds = multilinecmd.Split(task.CmdDirty)
		task.rebuild = task.sanitizeRebuild(task.RebuildDirty)

		tm[key] = task
	}
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
