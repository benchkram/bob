package playbook

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/errz"
	"github.com/logrusorgru/aurora"
)

var colorPool = []aurora.Color{
	1,
	aurora.BlueFg,
	aurora.GreenFg,
	aurora.CyanFg,
	aurora.MagentaFg,
	aurora.YellowFg,
	aurora.RedFg,
}
var round = 10 * time.Millisecond

// pickTaskColors picks a display color for each task in the playbook.
func (p *Playbook) pickTaskColors() {
	tasks := []string{}
	for _, t := range p.Tasks {
		tasks = append(tasks, t.Name())
	}
	sort.Strings(tasks)

	// Adjust padding of first column based on the taskname length.
	// Also assign fixed color to the tasks.
	p.namePad = 0
	for i, name := range tasks {
		if len(name) > p.namePad {
			p.namePad = len(name)
		}

		color := colorPool[i%len(colorPool)]
		p.Tasks[name].Task.SetColor(color)
	}
	p.namePad += 14

	dependencies := len(tasks) - 1
	rootName := p.Tasks[p.root].ColoredName()
	boblog.Log.V(1).Info(fmt.Sprintf("Running task %s with %d dependencies", rootName, dependencies))
}

// Build the playbook starting at root.
func (p *Playbook) Build(ctx context.Context) (err error) {
	done := make(chan error)

	processedTasks := []*bobtask.Task{}

	p.pickTaskColors()

	runningJobs := 0
	const numJobs = 5
	jobs := make(chan *bobtask.Task, numJobs)

	go func() {
		for j := range jobs {
			boblog.Log.V(5).Info(fmt.Sprintf("RUNNING JOBS  %d ", runningJobs))
			go func(t *bobtask.Task) {
				runningJobs++
				err := p.build(ctx, t)
				if err != nil {
					// cancel all running jobs
					ctx.Done()
					done <- err
				}
				runningJobs--
			}(j)
		}
	}()

	go func() {
		// TODO: Run a worker pool so that multiple tasks can run in parallel.

		c := p.TaskChannel()
		for t := range c {
			processedTasks = append(processedTasks, t)
			jobs <- t
			err = p.Play()
			if err != nil {
				if errors.Is(err, ErrDone) {
					break
				} else {
					errz.Fatal(err)
				}
			}
		}

		close(done)
	}()

	err = p.Play()
	errz.Fatal(err)

	err = <-done
	if err != nil {
		p.Done()
	}
	errz.Fatal(err)

	close(jobs)

	// iterate through tasks and log
	// skipped input files.
	var skippedInputs int
	for _, task := range processedTasks {
		skippedInputs = logSkippedInputs(
			skippedInputs,
			task.ColoredName(),
			task.LogSkippedInput(),
		)
	}

	p.summary(processedTasks)

	return err
}

// didWriteBuildOutput assures that a new line is added
// before writing state or logs of a task to stdout.
var didWriteBuildOutput bool

// build a single task and update the playbook state after completion.
func (p *Playbook) build(ctx context.Context, task *bobtask.Task) (err error) {
	defer errz.Recover(&err)

	var taskSuccessFul bool
	var taskErr error
	defer func() {
		if !taskSuccessFul {
			errr := p.TaskFailed(task.Name(), taskErr)
			if errr != nil {
				boblog.Log.Error(errr, "Setting the task state to failed, failed.")
			}
		}
	}()

	coloredName := task.ColoredName()

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-done:
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				boblog.Log.V(1).Info(fmt.Sprintf("%-*s\t%s", p.namePad, coloredName, StateCanceled))
				_ = p.TaskCanceled(task.Name())
			}
		}
	}()

	// Tasks without predecessors don't need recomputation.
	// If a task has a predecssor it might happen that the
	// task altered the filesystem. For this case hashes must be recomputated.
	var recomputeInputs bool
	if len(task.DependsOn) > 0 {
		recomputeInputs = true
	}
	hashIn, err := task.HashIn(recomputeInputs)
	errz.Fatal(err)

	rebuildRequired, rebuildCause, err := p.TaskNeedsRebuild(task.Name(), hashIn)
	errz.Fatal(err)

	// task might need a rebuild due to a input change.
	// but could still be possible to load the targets from the artifact store.
	// If a task needs a rebuild due to a dependency change => rebuild.
	//if rebuildRequired && rebuildCause != DependencyChanged && rebuildCause != TaskForcedRebuild {
	if rebuildRequired {
		switch rebuildCause {
		case TaskInputChanged:
			fallthrough
		case TargetInvalid:
			boblog.Log.V(2).Info(fmt.Sprintf("%-*s\t%s, unpacking artifact", p.namePad, coloredName, rebuildCause))
			// TODO: validate if unpack is required?
			success, err := task.ArtifactUnpack(hashIn)
			errz.Fatal(err)
			if success {
				rebuildRequired = false
			}
		case TaskForcedRebuild:
		case DependencyChanged:
		default:
		}
	}

	if !rebuildRequired {
		status := StateNoRebuildRequired
		boblog.Log.V(2).Info(fmt.Sprintf("%-*s\t%s", p.namePad, coloredName, status.Short()))
		taskSuccessFul = true
		return p.TaskNoRebuildRequired(task.Name())
	}

	if !didWriteBuildOutput {
		boblog.Log.V(1).Info("")
		didWriteBuildOutput = true
	}
	boblog.Log.V(1).Info(fmt.Sprintf("%-*s\trunning task...", p.namePad, coloredName))

	err = task.Clean()
	errz.Fatal(err)

	err = task.Run(ctx, p.namePad)
	if err != nil {
		taskSuccessFul = false
		taskErr = err
	}
	errz.Fatal(err)

	taskSuccessFul = true

	err = task.VerifyAfter()
	errz.Fatal(err)

	target, err := task.Target()
	if err != nil {
		errz.Fatal(err)
	}

	// Check targets are created correctly.
	// On success the target hash is computed
	// inside TaskCompleted().
	if target != nil {
		if !target.Exists() {
			boblog.Log.V(1).Info(fmt.Sprintf("%-*s\t%s\t(invalid targets)", p.namePad, coloredName, StateFailed))
			err = p.TaskFailed(task.Name(), fmt.Errorf("targets not created"))
			if err != nil {
				if errors.Is(err, ErrFailed) {
					return err
				}
			}
		}
	}

	err = p.TaskCompleted(task.Name(), hashIn)
	errz.Fatal(err)

	taskStatus, err := p.TaskStatus(task.Name())
	errz.Fatal(err)

	state := taskStatus.State()
	boblog.Log.V(1).Info(fmt.Sprintf("%-*s\t%s", p.namePad, coloredName, "..."+state.Short()))

	return nil
}

const maxSkippedInputs = 5

// logSkippedInputs until max is reached
func logSkippedInputs(count int, taskname string, skippedInputs []string) int {
	if len(skippedInputs) == 0 {
		return count
	}
	if count >= maxSkippedInputs {
		return maxSkippedInputs
	}

	for _, f := range skippedInputs {
		count = count + 1
		boblog.Log.V(1).Info(fmt.Sprintf("skipped %s '%s' %s", taskname, f, os.ErrPermission))

		if count >= maxSkippedInputs {
			boblog.Log.V(1).Info(fmt.Sprintf("skipped %s %s", taskname, "& more..."))
			break
		}
	}

	return count
}
