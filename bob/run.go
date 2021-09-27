package bob

import (
	"context"
	"errors"

	"github.com/Benchkram/bob/bob/bobfile"
	"github.com/Benchkram/bob/bobrun"
	"github.com/Benchkram/bob/pkg/runctl"
	"github.com/Benchkram/errz"
	"github.com/sanity-io/litter"
)

// Examples of possible run usecase
//
// 1: executable requiring a database to run properly.
//    Database is setup in a docker-compose file.
//
// 2: [Done] plain docker-compose run with dependcies to build-cmds
//    containing instructions how to build the container image.
//
// 3: init script requiring a executable to run before
//    containing a health endpoint (REST?). So the init script can be
//    sure about the service to be functional.
//

// Run builds dependent tasks for a run cmd and starts it.
// A control and a stopped channel are returned to interact
// from a terminal with the run cmd.
func (b *B) Run(ctx context.Context, runName string) (_ runctl.Control, err error) {
	defer errz.Recover(&err)

	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	runTask, ok := aggregate.Runs[runName]
	if !ok {
		return nil, ErrRunDoesNotExist
	}

	// gather childs of runTask.
	// TODO: remove duplicates (can happen when multiple runTasks point to the same runTask)
	childRunTasks := b.runTasksInChain(runName, aggregate)

	// build dependencies of child run dependencies.
	for _, task := range childRunTasks {
		err = b.buildDependentTasks(ctx, task, aggregate)
		errz.Fatal(err)
	}

	// build dependencies of current runTask.
	err = b.buildDependentTasks(ctx, runName, aggregate)
	errz.Fatal(err)

	// TODO: Run dependent child tasks..
	//   * Listen for stop signal of top run and shutdown when it exists
	//     only return stoped when childs have stopped.
	//   * Shutdown on error of dependent run tasks
	//   * Forbid circular dependecys.
	//
	// for _, child := range run.DependsOn {
	// }
	// Iterate childRunTasks in reverse order to run
	// deepest child in the tree first.
	litter.Dump(childRunTasks)

	masterCtl := runctl.New()

	runCtls := []runctl.Control{}
	for i := len(childRunTasks) - 1; i >= 0; i-- {
		taskname := childRunTasks[i]
		println(taskname)
		task := aggregate.Runs[taskname]

		rc, err := task.Run(ctx)
		errz.Fatal(err)

		// Wait for child to return started.
		// Binarys send started imediately,
		// compose only when `up` is done.
		<-rc.Started()

		runCtls = append(runCtls, rc)
	}

	rc, err := runTask.Run(ctx)
	errz.Fatal(err)
	runCtls = append(runCtls, rc)

	// Controls run tasks and takes care of signal forwarding
	go func() {
		for {
			select {
			case <-ctx.Done():
				// Wait for all childs to be stopped.
				for _, v := range runCtls {
					<-v.Stopped()
				}
				masterCtl.EmitStop()
				return
			case s := <-masterCtl.Control():
				switch s {
				case bobrun.RestartSignal:
					// TODO: Send shutdown signal to all runTasks

					// In the meantime trigger a rebobfile.
					err = b.buildDependentTasks(ctx, runName, aggregate)
					errz.Fatal(err)

					// TODO: Wait for shutdown to complete

					// TODO: Restart
					for _, ctl := range runCtls {
						ctl.EmitSignal(bobrun.RestartSignal)
						// TODO: also listen for errors.
						<-ctl.Started()
					}
				}
			}
		}

	}()

	return masterCtl, nil
}

// runTasksInChain returns run tasks in the dependency chain.
// It will not error but return a empty error in case the runName
// does not exists.
func (b *B) runTasksInChain(runName string, aggregate *bobfile.Bobfile) []string {
	runTasks := []string{}

	run, ok := aggregate.Runs[runName]
	if !ok {
		return nil
	}

	for _, task := range run.DependsOn {
		if !isRunTask(task, aggregate) {
			continue
		}
		runTasks = append(runTasks, task)

		// assure all it's dependent runTasks are also added.
		childs := b.runTasksInChain(task, aggregate)
		runTasks = append(runTasks, childs...)
	}
	return runTasks
}

func isRunTask(name string, aggregate *bobfile.Bobfile) bool {
	_, ok := aggregate.Runs[name]
	return ok
}
func isBuildTask(name string, aggregate *bobfile.Bobfile) bool {
	_, ok := aggregate.Tasks[name]
	return ok
}

func (b *B) buildDependentTasks(ctx context.Context, runname string, aggregate *bobfile.Bobfile) (err error) {
	defer errz.Recover(&err)

	runTask, ok := aggregate.Runs[runname]
	if !ok {
		return ErrRunDoesNotExist
	}

	// Run dependent build tasks
	// before starting the run task
	for _, child := range runTask.DependsOn {
		if !isBuildTask(child, aggregate) {
			continue
		}

		playbook, err := aggregate.Playbook(child)
		if err != nil {
			if errors.Is(err, ErrTaskDoesNotExist) {
				continue
			}
			errz.Fatal(err)
		}

		err = b.BuildTask(ctx, child, playbook)
		errz.Fatal(err)
	}

	return nil
}
