package bob

import (
	"context"
	"errors"
	"fmt"
	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/pkg/ctl"
	"github.com/benchkram/bob/pkg/nix"
	"github.com/benchkram/bob/pkg/sliceutil"
	"github.com/benchkram/errz"
)

// Examples of possible interactive usecase
//
// 1: [Done] executable requiring a database to run properly.
//    Database is setup in a docker-compose file.
//
// 2: [Done] plain docker-compose run with dependcies to build-cmds
//    containing instructions how to build the container image.
//
// TODO:
// 3: init script requiring a executable to run before
//    containing a health endpoint (REST?). So the init script can be
//    sure about the service to be functional.
//

// Run builds dependent tasks for a run cmd and starts it.
// A control is returned to interact with the run cmd.
//
// Canceling the cmd from the outside must be done through the context.
//
// TODO: Forbid circular dependecys.
func (b *B) Run(ctx context.Context, runName string) (_ ctl.Commander, err error) {
	defer errz.Recover(&err)

	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	// TODO: build nix depencies here

	b.PrintVersionCompatibility(aggregate)

	runTask, ok := aggregate.RTasks[runName]
	if !ok {
		return nil, ErrRunDoesNotExist
	}

	// gather interactive tasks
	childInteractiveTasks := b.interactiveTasksInChain(runName, aggregate)
	interactiveTasks := []string{runTask.Name()}
	interactiveTasks = append(interactiveTasks, childInteractiveTasks...)

	//

	// build dependencies & main runTask
	for _, task := range interactiveTasks {

		err = buildNonInteractive(ctx, task, aggregate)
		errz.Fatal(err)
	}

	// generate run controls to steer the run cmd.
	runCtls := []ctl.Command{}
	for _, name := range interactiveTasks {
		interactiveTask := aggregate.RTasks[name]

		rc, err := interactiveTask.Run(ctx)
		errz.Fatal(err)

		runCtls = append(runCtls, rc)
	}

	builder := NewBuilder(b, runName, aggregate, buildNonInteractive)
	commander := ctl.NewCommander(ctx, builder, runCtls...)

	return commander, nil
}

// interactiveTasksInChain returns run tasks in the dependency chain.
// Task on a higher level in the tree appear at the front of the slice..
//
// It will not error but return a empty error in case the runName
// does not exists.
func (b *B) interactiveTasksInChain(runName string, aggregate *bobfile.Bobfile) []string {
	runTasks := []string{}

	run, ok := aggregate.RTasks[runName]
	if !ok {
		return nil
	}

	for _, task := range run.DependsOn {
		if !isInteractive(task, aggregate) {
			continue
		}
		runTasks = append(runTasks, task)

		// assure all it's dependent runTasks are also added.
		childs := b.interactiveTasksInChain(task, aggregate)
		runTasks = append(runTasks, childs...)
	}

	return normalize(runTasks)
}

// normalize removes duplicated entrys from the run task list.
// The duplicate closest to the top of the chain is removed
// so that child tasks are started first.
func normalize(tasks []string) []string {
	sanitized := []string{}

	for i, task := range tasks {
		keep := true

		// last element can always be added safely
		if i < len(tasks) {
			for _, jtask := range tasks[i+1:] {
				if task == jtask {
					keep = false
					break
				}
			}
		}

		if keep {
			sanitized = append(sanitized, task)
		}
	}

	return sanitized
}

func isInteractive(name string, aggregate *bobfile.Bobfile) bool {
	_, ok := aggregate.RTasks[name]
	return ok
}

// func isNonInteractive(name string, aggregate *bobfile.Bobfile) bool {
// 	_, ok := aggregate.Tasks[name]
// 	return ok
// }

// buildNonInteractive takes a interactive task to build it's non-interactive children.
func buildNonInteractive(ctx context.Context, runname string, aggregate *bobfile.Bobfile) (err error) {
	defer errz.Recover(&err)

	interactive, ok := aggregate.RTasks[runname]
	if !ok {
		return ErrRunDoesNotExist
	}

	var tasksInPipeline []string
	nixDependencies := make([]nix.Dependency, 0)
	for _, child := range interactive.DependsOn {
		if isInteractive(child, aggregate) {
			continue
		}
		err = aggregate.BTasks.CollectTasksInPipeline(child, &tasksInPipeline)
		if err != nil {
			return err
		}
		err = aggregate.BTasks.CollectNixDependencies(child, &nixDependencies)
		if err != nil {
			return err
		}
	}

	if len(nixDependencies) > 0 {
		fmt.Println("Building nix dependencies...")

		storePaths, err := BuildNixDependencies(nix.UniqueDeps(append(nix.DefaultPackages(), nixDependencies...)))
		if err != nil {
			return err
		}

		for _, name := range tasksInPipeline {
			t := aggregate.BTasks[name]
			t.SetStorePaths(sliceutil.Unique(storePaths))
			aggregate.BTasks[name] = t
		}
	}

	// Run dependent build tasks
	// before starting the run task
	for _, child := range interactive.DependsOn {
		if isInteractive(child, aggregate) {
			continue
		}

		playbook, err := aggregate.Playbook(child)
		if err != nil {
			if errors.Is(err, ErrTaskDoesNotExist) {
				continue
			}
			errz.Fatal(err)
		}

		err = playbook.Build(ctx)
		errz.Fatal(err)
	}

	return nil
}
