package bob

import (
	"context"
	"errors"
	"fmt"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/pkg/boberror"
	"github.com/benchkram/bob/pkg/ctl"
)

// Examples of possible interactive usecase
//
// 1: [Done] executable requiring a database to run properly.
//    Database is setup in a docker-compose file.
//
// 2: [Done] plain docker-compose run with dependencies to build-cmds
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

	b.PrintVersionCompatibility(aggregate)

	runTask, ok := aggregate.RTasks[runName]
	if !ok {
		return nil, ErrRunDoesNotExist
	}

	// gather interactive tasks
	childInteractiveTasks := b.runTasksInPipeline(runName, aggregate)
	interactiveTasks := []string{runTask.Name()}
	interactiveTasks = append(interactiveTasks, childInteractiveTasks...)

	// build dependencies & main runTask
	for _, task := range interactiveTasks {
		err = executeBuildTasksInPipeline(ctx, task, aggregate, b.nix)
		errz.Fatal(err)
	}

	// generate run controls to steer the run cmd.
	runCommands := []ctl.Command{}
	for _, name := range interactiveTasks {
		runTask := aggregate.RTasks[name]

		command, err := runTask.Command(ctx)
		errz.Fatal(err)

		runCommands = append(runCommands, command)
	}

	builder := NewBuilder(b, runName, aggregate, executeBuildTasksInPipeline)
	commander := ctl.NewCommander(ctx, builder, runCommands...)

	return commander, nil
}

// runTasksInPipeline returns run tasks in the pipeline.
// Task on a higher level in the tree appear at the front of the slice..
//
// It will not error but return a empty error in case the runName
// does not exists.
func (b *B) runTasksInPipeline(runName string, aggregate *bobfile.Bobfile) []string {
	runTasks := []string{}

	run, ok := aggregate.RTasks[runName]
	if !ok {
		return nil
	}

	for _, task := range run.DependsOn {
		if !isRunTask(task, aggregate) {
			continue
		}
		runTasks = append(runTasks, task)

		// assure all it's dependent runTasks are also added.
		childs := b.runTasksInPipeline(task, aggregate)
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

func isRunTask(name string, aggregate *bobfile.Bobfile) bool {
	_, ok := aggregate.RTasks[name]
	return ok
}

// func isBuildTask(name string, aggregate *bobfile.Bobfile) bool {
// 	_, ok := aggregate.Tasks[name]
// 	return ok
// }

// executeBuildTasksInPipeline takes a run task but only executes the dependent build tasks
func executeBuildTasksInPipeline(ctx context.Context, runname string, aggregate *bobfile.Bobfile, nix *NixBuilder) (err error) {
	defer errz.Recover(&err)

	interactive, ok := aggregate.RTasks[runname]
	if !ok {
		return ErrRunDoesNotExist
	}

	// Gather build tasks
	buildTasks := []string{}
	for _, child := range interactive.DependsOn {
		if isRunTask(child, aggregate) {
			continue
		}
		buildTasks = append(buildTasks, child)
	}

	// Build nix dependencies
	if nix != nil {
		fmt.Println("Building nix dependencies...")
		err = nix.BuildNixDependencies(aggregate, buildTasks)
		errz.Fatal(err)
		fmt.Println("Succeded building nix dependencies")
	}

	// Run dependent build tasks
	// before starting the run task
	for _, child := range interactive.DependsOn {
		if isRunTask(child, aggregate) {
			continue
		}

		playbook, err := aggregate.Playbook(child)
		if err != nil {
			if errors.Is(err, boberror.ErrTaskDoesNotExist) {
				continue
			}
			errz.Fatal(err)
		}

		err = playbook.Build(ctx)
		errz.Fatal(err)
	}

	return nil
}
