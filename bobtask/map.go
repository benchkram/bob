package bobtask

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

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

// FilterInputs in parallel
func (tm Map) FilterInputs() (err error) {
	defer errz.Recover(&err)

	errors := []error{}
	errorsM := sync.Mutex{}
	mapM := sync.Mutex{}

	wd, err := filepath.Abs(".")
	errz.Fatal(err)

	wg := sync.WaitGroup{}
	mapM.Lock()
	for key, task := range tm {
		wg.Add(1)
		go func(k string, t Task) {

			errr := t.FilterInputs(wd)
			if errr != nil {
				errorsM.Lock()
				errors = append(errors, errr)
				errorsM.Unlock()
			}

			mapM.Lock()
			tm[k] = t
			mapM.Unlock()

			wg.Done()
		}(key, task)
	}
	mapM.Unlock()

	wg.Wait()
	if len(errors) > 0 {
		errz.Fatal(errors[0])
	}

	return nil
}

// FilterInputsSequential is the sequential version of FilterInputs.
// Can be handy for debugging input errors.
func (tm Map) FilterInputsSequential() (err error) {
	defer errz.Recover(&err)

	wd, err := filepath.Abs(".")
	errz.Fatal(err)

	for key, task := range tm {
		err = task.FilterInputs(wd)
		errz.Fatal(err)
		tm[key] = task
	}

	return nil
}

// Sanitize task map and write filtered & sanitized
// properties from dirty members to plain (e.g. dirtyInputs -> filter&sanitize -> inputs)
func (tm Map) Sanitize() (err error) {
	defer errz.Recover(&err)

	for key, task := range tm {

		err = task.parseTargets()
		errz.Fatal(err)

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
	var tasksInPipeline []string
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

// CollectNixDependenciesForTasks will collect all nix dependencies for task taskName
// in nixDependencies slice
func (tm Map) CollectNixDependenciesForTasks(whitelist []string) ([]nix.Dependency, error) {
	var nixDependencies []nix.Dependency
	for _, taskFromMap := range tm {
		// only add dependencies of whitelisted tasks.
		for _, taskName := range whitelist {
			if taskFromMap.Name() == taskName {
				nixDependencies = append(nixDependencies, taskFromMap.Dependencies()...)
			}
		}
	}

	return nixDependencies, nil
}

// IgnoreChildTargets fills the `InputAdditionalIgnores` field of each task
// with the targets of each tasks children.
func (tm Map) IgnoreChildTargets() (err error) {
	defer errz.Recover(&err)
	for name, umbrellaTask := range tm {

		err := tm.Walk(name, "", func(tn string, task Task, err error) error {
			if err != nil {
				return err
			}

			if task.target != nil {
				for _, p := range task.target.FilesystemEntriesRaw() {
					if umbrellaTask.Dir() == task.Dir() {
						// everything good.. use them as they are
						umbrellaTask.InputAdditionalIgnores = append(umbrellaTask.InputAdditionalIgnores, p)
					} else {

						//    List of cases to be covered.
						//
						//     umbrellaDIR                 currentTargetDIR      currentTargetPATH
						//
						//     .                           second-level          second-level/target
						//     .                           .                     aaa/bbb/target
						//     .                           second-level          aaa/second-level/target
						//
						//     second-level                second-level          second-level/target
						//     second-level                third-level           second-level/third-level/target
						//     second-level                third-level           second-level/third-level/aaa/bbb/target
						//     second-level                third-level           second-level/
						//
						//     second-level/third-level    third-level           second-level/third-level/target
						//
						//     third-level    fouth-level           second-level/third-level/fourth-level/target

						relP, err := filepath.Rel(umbrellaTask.Dir(), p)
						if err != nil {
							return err
						}
						umbrellaTask.InputAdditionalIgnores = append(umbrellaTask.InputAdditionalIgnores, relP)
					}
				}
			}

			return nil
		})
		errz.Fatal(err)

		tm[name] = umbrellaTask
	}

	return nil
}

// VerifyDuplicateTargets checks if multiple build tasks point to the same target.
func (tm Map) VerifyDuplicateTargets() error {
	// mapping [target][]taskname
	targetToTasks := make(map[string][]string)

	for taskName, v := range tm {
		if !v.TargetExists() {
			continue
		}
		for _, t := range v.target.DockerImages() {
			targetToTasks[t] = append(targetToTasks[t], taskName)

			if len(targetToTasks[t]) > 1 {
				return usererror.Wrap(CreateErrAmbiguousTargets(targetToTasks[t], t))
			}
		}
		for _, t := range v.target.FilesystemEntriesRaw() {
			targetToTasks[t] = append(targetToTasks[t], taskName)

			if len(targetToTasks[t]) > 1 {
				return usererror.Wrap(CreateErrAmbiguousTargets(targetToTasks[t], t))
			}
		}
	}

	// FIXME: A filesystem target can still point to a file inside
	// a directory target.
	//
	// Could be solved by being more strict with target definitions.
	// E.g. a directory must be defined as "dir/" instead of "dir".
	// This allows to catch that case without traversing the
	// actual filesystem.
	return nil
}

func CreateErrAmbiguousTargets(tasks []string, target string) error {
	sort.Strings(tasks)
	return fmt.Errorf("%w,\nmultiple tasks [%s] pointing to the same target `%s`", ErrAmbigousTargets, strings.Join(tasks, " "), target)
}
