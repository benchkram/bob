package bob

import (
	"context"
	"errors"
	"fmt"

	"github.com/benchkram/errz"

	"github.com/benchkram/bob/bob/bobfile"
	"github.com/benchkram/bob/pkg/boberror"
	"github.com/benchkram/bob/pkg/ctl"
	"github.com/benchkram/bob/pkg/sliceutil"
)

// Examples of possible interactive usecase
//
// 1: [Done] executable requiring a database to run properly.
//    Database is setup in a docker-compose file.
//
// 2: [Done] plain docker-compose run with dependcies to build-cmds
//    containing instructions how to build the container image.
//
// 3: [Done] init script requiring an executable to run before
//    containing a health endpoint (REST?). So the init script can be
//    sure about the service to be functional.
//

// Run builds dependent tasks for a run cmd and starts it.
// A control is returned to interact with the run cmd.
//
// Canceling the cmd from the outside must be done through the context.
//
// FIXME: Forbid circular dependecys.
func (b *B) Run(ctx context.Context, runTaskName string) (_ ctl.Commander, err error) {
	defer errz.Recover(&err)

	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	b.PrintVersionCompatibility(aggregate)

	runTask, ok := aggregate.RTasks[runTaskName]
	if !ok {
		return nil, ErrRunDoesNotExist
	}

	if aggregate.UseNix && b.nix != nil {
		for i, task := range aggregate.RTasks {
			task.AddEnvironment(b.env)
			aggregate.RTasks[i] = task
		}
		for i, task := range aggregate.BTasks {
			task.AddEnvironment(b.env)
			aggregate.BTasks[i] = task
		}
	}

	// gather interactive tasks
	childInteractiveTasks := runTasksInPipeline(runTaskName, aggregate)
	interactiveTasks := []string{runTask.Name()}
	interactiveTasks = append(interactiveTasks, childInteractiveTasks...)

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

	builder := NewBuilder(runTaskName, aggregate, executeBuildTasksInPipeline, b.nix)
	commander := ctl.NewCommander(ctx, builder, runCommands...)

	return commander, nil
}

// runTasksInPipeline returns run tasks in the pipeline.
// Task on a higher level in the tree appear at the front of the slice..
//
// It will not error but return a empty error in case the runName
// does not exists.
func runTasksInPipeline(runTaskName string, aggregate *bobfile.Bobfile) []string {
	runTasks := []string{}

	run, ok := aggregate.RTasks[runTaskName]
	if !ok {
		return nil
	}

	for _, task := range run.DependsOn {
		if !isRunTask(task, aggregate) {
			continue
		}
		runTasks = append(runTasks, task)

		// assure all it's dependent runTasks are also added.
		childs := runTasksInPipeline(task, aggregate)
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

// executeBuildTasksInPipeline takes a run task and starts the required builds.
func executeBuildTasksInPipeline(
	ctx context.Context,
	runTaskName string,
	aggregate *bobfile.Bobfile,
	nix *NixBuilder,
) (err error) {
	defer errz.Recover(&err)

	_, ok := aggregate.RTasks[runTaskName]
	if !ok {
		return ErrRunDoesNotExist
	}
	runTasksInPipeline := runTasksInPipeline(runTaskName, aggregate)

	// Gather build tasks from run task dependencies.
	// This is required to get the top most build tasks and start a build for each.
	// Each run task could have could have distinct build pipeline beneth it.
	// This implies that multiple unrelated builds could be started
	// on a run invocation.

	// umbrella run task
	buildTasks, err := gatherBuildTasks(runTaskName, aggregate)
	errz.Fatal(err)
	// child run tasks
	for _, runTaskName := range runTasksInPipeline {
		childBuildTasks, err := gatherBuildTasks(runTaskName, aggregate)
		errz.Fatal(err)
		buildTasks = append(buildTasks, childBuildTasks...)
	}
	buildTasks = sliceutil.Unique(buildTasks)

	// Build nix dependencies
	if aggregate.UseNix && nix != nil {
		fmt.Println("Building nix dependencies...")
		err = nix.BuildNixDependencies(aggregate, buildTasks, append(runTasksInPipeline, runTaskName))
		errz.Fatal(err)
		fmt.Println("Succeeded building nix dependencies")
	}

	// Initiate each build
	for _, buildTask := range buildTasks {
		playbook, err := aggregate.Playbook(buildTask)
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

// gatherBuildTasks returns all direct build tasks in the pipeline of a run task.
func gatherBuildTasks(runTaskName string, aggregate *bobfile.Bobfile) ([]string, error) {
	runTask, ok := aggregate.RTasks[runTaskName]
	if !ok {
		return nil, ErrRunDoesNotExist
	}

	buildTasks := []string{}
	for _, child := range runTask.DependsOn {
		if isRunTask(child, aggregate) {
			continue
		}
		buildTasks = append(buildTasks, child)
	}

	return buildTasks, nil
}
