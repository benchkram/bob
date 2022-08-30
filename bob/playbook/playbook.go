package playbook

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/bobtask/buildinfo"
	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/boberror"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/buildinfostore"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
	"github.com/logrusorgru/aurora"
)

// The playbook defines the order in which tasks are allowed to run.
// Also determines the possibility to run tasks in parallel.

var ErrDone = fmt.Errorf("playbook is done")
var ErrFailed = fmt.Errorf("playbook failed")
var ErrUnexpectedTaskState = fmt.Errorf("task state is unexpected")

type Playbook struct {
	// taskChannel is closed when the root
	// task completes.
	taskChannel chan *bobtask.Task

	// errorChannel to transport errors to the caller
	errorChannel chan error

	// root task
	root string

	Tasks StatusMap

	namePad int

	done bool
	// doneChannel is closed when the playbook is done.
	doneChannel chan struct{}

	// start is the point in time the playbook started
	start time.Time
	// end is the point in time the playbook ended
	end time.Time

	// enableCaching allows artifacts to be read & written to a store.
	// Default: true.
	enableCaching bool

	predictedNumOfTasks int

	// playMutex assures recomputation
	// can only be done sequentially.
	playMutex sync.Mutex

	// jobs is the maximum number of parallel jobs
	jobs int
}

func New(root string, opts ...Option) *Playbook {
	p := &Playbook{
		errorChannel:  make(chan error),
		Tasks:         make(StatusMap),
		doneChannel:   make(chan struct{}),
		enableCaching: true,
		root:          root,

		jobs: runtime.NumCPU(),

		predictedNumOfTasks: 100000,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(p)
	}

	// Try to make the task channel the same size as the number of tasks.
	// In the unlikely case all task would be ready to be executed at the same time.
	p.taskChannel = make(chan *bobtask.Task, p.predictedNumOfTasks)

	return p
}

type RebuildCause string

func (rc *RebuildCause) String() string {
	return string(*rc)
}

const (
	TaskInputChanged  RebuildCause = "input-changed"
	TaskForcedRebuild RebuildCause = "forced"
	DependencyChanged RebuildCause = "dependency-changed"
	TargetInvalid     RebuildCause = "target-invalid"
)

// TaskNeedsRebuild check if a tasks need a rebuild by looking at it's hash value
// and it's child tasks.
func (p *Playbook) TaskNeedsRebuild(taskname string, hashIn hash.In) (rebuildRequired bool, cause RebuildCause, err error) {
	ts, ok := p.Tasks[taskname]
	if !ok {
		return false, "", usererror.Wrap(boberror.ErrTaskDoesNotExistF(taskname))
	}
	task := ts.Task
	coloredName := task.ColoredName()

	// returns true if rebuild strategy set to `always`
	if task.Rebuild() == bobtask.RebuildAlways {
		boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tREBUILDING\t(rebuild set to always)", p.namePad, coloredName))
		return true, TaskForcedRebuild, nil
	}

	rebuildRequired, err = task.NeedsRebuild(&bobtask.RebuildOptions{HashIn: &hashIn})
	errz.Fatal(err)
	if rebuildRequired {
		boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(input changed)", p.namePad, coloredName))
		return true, TaskInputChanged, nil
	}

	var Done = fmt.Errorf("done")
	// Check if task needs a rebuild due to its dependencies changing
	err = p.Tasks.walk(task.Name(), func(tn string, t *Status, err error) error {
		if err != nil {
			return err
		}

		// TODO: In case the task does not exist check if a artifact can be used?
		//       Part of no-permission-workflow.

		// Ignore the task itself
		if task.Name() == tn {
			return nil
		}

		// Require a rebuild if the dependend task did require a rebuild
		if t.State() != StateNoRebuildRequired {
			boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(dependecy changed)", p.namePad, coloredName))
			rebuildRequired = true
			// Bail out early
			return Done
		}

		return nil
	})

	if errors.Is(err, Done) {
		return true, DependencyChanged, nil
	}

	if !rebuildRequired {
		// check rebuild due to invalidated targets
		target, err := task.Target()
		if err != nil {
			return true, "", err
		}
		if target != nil {
			// In case of a invalid traget a rebuild is required

			// TODO: simplify verify by check size + modification time of target.
			rebuildRequired = !target.Verify()

			// Try to load a target from the store when a rebuild is required.
			// If not assure the artifact exists in the store.
			if rebuildRequired {
				boblog.Log.V(2).Info(fmt.Sprintf("[task:%s] trying to get target from store", taskname))
				ok, err := task.ArtifactUnpack(hashIn)
				boblog.Log.Error(err, "Unable to get target from store")

				if ok {
					rebuildRequired = false
				} else {
					boblog.Log.V(3).Info(fmt.Sprintf("[task:%s] failed to get target from store", taskname))
				}
			} else {
				// Hint: Once there was a time when we created the target in the store
				// in case no rebuild was required and the target doesn't exist.
				// Though this should only be done after the target was really build..
				// If loaded from a remote.. it anyway is synced through the local store.
				if !task.ArtifactExists(hashIn) {
					boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(target not in local store)", p.namePad, coloredName))
					return true, TargetInvalid, err
				}
			}

			if rebuildRequired {
				boblog.Log.V(3).Info(fmt.Sprintf("%-*s\tNEEDS REBUILD\t(invalid targets)", p.namePad, coloredName))
			}
		}
	}

	return rebuildRequired, TargetInvalid, err
}

func (p *Playbook) Play() (err error) {
	return p.play()
}

func (p *Playbook) play() error {

	if p.done {
		return ErrDone
	}

	p.playMutex.Lock()
	defer p.playMutex.Unlock()

	if p.start.IsZero() {
		p.start = time.Now()
	}

	// Walk the task chain and determine the next build task. Send it to the task channel.
	// Returns `taskQueued` when a task has been send to the taskChannel.
	// Returns `taskFailed` when a task has failed.
	// Once it returns `nil` the playbook is done with it's work.
	var taskQueued = fmt.Errorf("task queued")
	var taskFailed = fmt.Errorf("task failed")
	err := p.Tasks.walk(p.root, func(taskname string, task *Status, err error) error {
		if err != nil {
			return err
		}

		//boblog.Log.V(3).Info(fmt.Sprintf("%-*s\t walking", p.namePad, taskname))

		switch task.State() {
		case StatePending:
			// Check if all dependent tasks are completed
			for _, dependentTaskName := range task.Task.DependsOn {
				t, ok := p.Tasks[dependentTaskName]
				if !ok {
					//fmt.Printf("Task %s does not exist", dependentTaskName)
					return usererror.Wrap(boberror.ErrTaskDoesNotExistF(dependentTaskName))
				}

				state := t.State()
				if state != StateCompleted && state != StateNoRebuildRequired {
					// A dependent task is not completed.
					// So this task is not yet ready to run.
					return nil
				}
			}
		case StateFailed:
			return taskFailed
		case StateCanceled:
			return nil
		case StateNoRebuildRequired:
			return nil
		case StateCompleted:
			return nil
		case StateRunning:
			return nil
		default:
		}

		// fmt.Printf("sending task %s to channel\n", task.Task.Name())
		// setting the task start time before passing it to channel
		task.Start = time.Now()
		// TODO: for async assure to handle send to a closed channel.
		_ = p.setTaskState(task.Name(), StateRunning, nil)
		p.taskChannel <- task.Task
		return taskQueued
	})

	// taskQueued => return nil (happy path)
	// taskFailed => return PlaybookFailed (ErrFailed)
	// default    => return err
	if err != nil {
		if errors.Is(err, taskQueued) {
			return nil
		}
		if errors.Is(err, taskFailed) {
			return ErrFailed
		}
		return err
	}

	// Avoid finishing the playbook before all task are done running
	if p.numRunningTasks() > 0 {
		return nil
	}

	// no work done, usually happens when
	// no task needs a rebuild.
	p.Done()

	return nil
}

func (p *Playbook) numRunningTasks() int {
	var parallel int
	_ = p.Tasks.walk(p.root, func(taskname string, task *Status, err error) error {
		if err != nil {
			return err
		}

		if task.State() == StateRunning {
			parallel++
		}

		return nil
	})
	return parallel
}

func (p *Playbook) Done() {
	if !p.done {
		p.done = true
		p.end = time.Now()
		close(p.taskChannel)
		close(p.doneChannel)
	}
}
func (p *Playbook) DoneChan() chan struct{} {
	return p.doneChannel
}

// TaskChannel returns the next task
func (p *Playbook) TaskChannel() <-chan *bobtask.Task {
	return p.taskChannel
}

func (p *Playbook) ErrorChannel() <-chan error {
	return p.errorChannel
}

func (p *Playbook) setTaskState(taskname string, state State, taskError error) error {
	task, ok := p.Tasks[taskname]
	if !ok {
		return boberror.ErrTaskDoesNotExistF(taskname)
	}

	task.SetState(state, taskError)
	switch state {
	case StateCompleted, StateCanceled, StateNoRebuildRequired:
		task.End = time.Now()
	}

	//p.Tasks[taskname] = task
	return nil
}

func (p *Playbook) pack(taskname string, hash hash.In) error {
	task, ok := p.Tasks[taskname]
	if !ok {
		return usererror.Wrap(boberror.ErrTaskDoesNotExistF(taskname))
	}
	return task.Task.ArtifactPack(hash)
}

func (p *Playbook) storeHash(taskname string, buildinfo *buildinfo.I) error {
	task, ok := p.Tasks[taskname]
	if !ok {
		return usererror.Wrap(boberror.ErrTaskDoesNotExistF(taskname))
	}

	return task.Task.WriteBuildinfo(buildinfo)
}

func (p *Playbook) ExecutionTime() time.Duration {
	return p.end.Sub(p.start)
}

// TaskStatus returns the current state of a task
func (p *Playbook) TaskStatus(taskname string) (ts *Status, _ error) {
	status, ok := p.Tasks[taskname]
	if !ok {
		return ts, usererror.Wrap(boberror.ErrTaskDoesNotExistF(taskname))
	}
	return status, nil
}

// TaskCompleted sets a task to completed
func (p *Playbook) TaskCompleted(taskname string, hashIn hash.In) (err error) {
	defer errz.Recover(&err)

	task, ok := p.Tasks[taskname]
	if !ok {
		return usererror.Wrap(boberror.ErrTaskDoesNotExistF(taskname))
	}

	buildInfo, err := task.ReadBuildInfo()
	if err != nil {
		if errors.Is(err, buildinfostore.ErrBuildInfoDoesNotExist) {
			// assure buildinfo is initialized correctly
			buildInfo = buildinfo.New()
		} else {
			errz.Fatal(err)
		}
	}
	buildInfo.Meta.Task = task.Name()
	buildInfo.Meta.InputHash = hashIn.String()

	// Compute buildinfo for the target
	target, err := task.Task.Target()
	errz.Fatal(err)
	if target != nil {
		bi, err := target.BuildInfo()
		errz.Fatal(err)
		buildInfo.Target = bi
	}

	// Store buildinfo
	err = p.storeHash(taskname, buildInfo)
	errz.Fatal(err)

	// Store targets in the artifact store
	if p.enableCaching {
		err = p.pack(taskname, hashIn)
		errz.Fatal(err)
	}

	// update task state and trigger another playbook run
	err = p.setTaskState(taskname, StateCompleted, nil)
	errz.Fatal(err)
	err = p.play()
	if err != nil {
		if !errors.Is(err, ErrDone) {
			errz.Fatal(err)
		}
	}

	return nil
}

// TaskNoRebuildRequired sets a task's state to indicate that no rebuild is required
func (p *Playbook) TaskNoRebuildRequired(taskname string) (err error) {
	defer errz.Recover(&err)

	err = p.setTaskState(taskname, StateNoRebuildRequired, nil)
	errz.Fatal(err)

	err = p.play()
	if err != nil {
		if !errors.Is(err, ErrDone) {
			errz.Fatal(err)
		}
	}

	return nil
}

// TaskFailed sets a task to failed
func (p *Playbook) TaskFailed(taskname string, taskErr error) (err error) {
	defer errz.Recover(&err)

	err = p.setTaskState(taskname, StateFailed, taskErr)
	errz.Fatal(err)

	// p.errorChannel <- fmt.Errorf("Task %s failed", taskname)

	// give the playbook the chance to set
	// the state to done.
	_ = p.play()

	return nil
}

// TaskCanceled sets a task to canceled
func (p *Playbook) TaskCanceled(taskname string) (err error) {
	defer errz.Recover(&err)

	err = p.setTaskState(taskname, StateCanceled, nil)
	errz.Fatal(err)

	// p.errorChannel <- fmt.Errorf("Task %s cancelled", taskname)

	return nil
}

func (p *Playbook) List() (err error) {
	defer errz.Recover(&err)

	keys := make([]string, 0, len(p.Tasks))
	for k := range p.Tasks {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Println(k)
	}

	return nil
}

func (p *Playbook) String() string {
	description := bytes.NewBufferString("")

	fmt.Fprint(description, "Playbook:\n")

	keys := make([]string, 0, len(p.Tasks))
	for k := range p.Tasks {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		task := p.Tasks[k]
		fmt.Fprintf(description, "  %s(%s): %s\n", k, task.Task.Name(), task.State())
	}

	return description.String()
}

type State string

// Summary state indicators.
// The nbsp are intended to align on the cli.
func (s *State) Summary() string {
	switch *s {
	case StatePending:
		return "⌛       "
	case StateCompleted:
		return aurora.Green("✔").Bold().String() + "       "
	case StateNoRebuildRequired:
		return aurora.Green("cached").String() + "  "
	case StateFailed:
		return aurora.Red("failed").String() + "  "
	case StateCanceled:
		return aurora.Faint("canceled").String()
	default:
		return ""
	}
}

func (s *State) Short() string {
	switch *s {
	case StatePending:
		return "pending"
	case StateCompleted:
		return "done"
	case StateNoRebuildRequired:
		return "cached"
	case StateFailed:
		return "failed"
	case StateCanceled:
		return "canceled"
	default:
		return ""
	}
}

const (
	StatePending           State = "PENDING"
	StateCompleted         State = "COMPLETED"
	StateNoRebuildRequired State = "CACHED"
	StateFailed            State = "FAILED"
	StateRunning           State = "RUNNING"
	StateCanceled          State = "CANCELED"
)
