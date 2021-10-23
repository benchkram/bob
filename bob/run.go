package bob

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/Benchkram/bob/bob/bobfile"
	"github.com/Benchkram/bob/pkg/ctl"
	"github.com/Benchkram/bob/pkg/runctl"
	"github.com/Benchkram/errz"
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
// A control is returned to interact with the run cmd.
//
// Canceling the cmd from the outside must be done through the context.
//
// TODO: Forbid circular dependecys.
func (b *B) Run(ctx context.Context, runName string) (_ runctl.Control, err error) {
	defer errz.Recover(&err)

	aggregate, err := b.Aggregate()
	errz.Fatal(err)

	runTask, ok := aggregate.Runs[runName]
	if !ok {
		return nil, ErrRunDoesNotExist
	}

	// gather child run Tasks
	childRunTasks := b.runTasksInChain(runName, aggregate)
	runTasks := []string{runTask.Name()}
	runTasks = append(runTasks, childRunTasks...)

	// build dependencies & main runTask
	for _, task := range runTasks {
		err = b.buildDependentTasks(ctx, task, aggregate)
		errz.Fatal(err)
	}

	// generate run controls to steer the run cmd.
	runCtls := []ctl.Command{}
	for _, name := range runTasks {
		task := aggregate.Runs[name]

		rc, err := task.Run(ctx)
		errz.Fatal(err)

		go func() { _, _ = io.Copy(os.Stdout, rc.Stdout()) }()
		go func() { _, _ = io.Copy(os.Stderr, rc.Stderr()) }()

		runCtls = append(runCtls, rc)
	}

	commander := runctl.NewCommander(ctx, runCtls...)

	// Control run tasks and take care of signal forwarding
	mainCtl := runctl.New("main", 0)
	go func() {
		restarting := runctl.Flag{}

		for {
			select {
			case <-ctx.Done():
				// wait till all cmds are done
				<-commander.Done()
				mainCtl.EmitDone()
				return
			case s := <-mainCtl.Control():
				switch s {
				case runctl.Restart:
					// prevent a restart to happen multiple times.
					// Blocks till the first restart request is finished.
					done, err := restarting.InProgress()
					if err != nil {
						continue
					}

					go func() {
						defer done()

						err := commander.Stop()
						errz.Log(err)

						// Trigger a rebuild.
						err = b.buildDependentTasks(ctx, runName, aggregate)
						errz.Fatal(err)

						err = commander.Start()
						errz.Log(err)

						mainCtl.EmitRestarted()
					}()
				}
			}
		}
	}()

	go func() {
		_ = commander.Start()
	}()

	return mainCtl, nil
}

// runTasksInChain returns run tasks in the dependency chain.
// Task on a higher level in the tree appear at the front of the slice..
//
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

		err = playbook.BuildTask(ctx, child)
		errz.Fatal(err)
	}

	return nil
}
