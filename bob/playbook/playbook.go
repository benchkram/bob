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
	"github.com/benchkram/bob/bobtask/target"
	"github.com/benchkram/bob/pkg/boberror"
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

	// predictedNumOfTasks is used to pick
	// an appropriate channel size for the task queue.
	predictedNumOfTasks int

	// playMutex assures recomputation
	// can only be done sequentially.
	playMutex sync.Mutex

	// maxParallel is the maximum number of parallel executed tasks
	maxParallel int
}

func New(root string, opts ...Option) *Playbook {
	p := &Playbook{
		errorChannel:  make(chan error),
		Tasks:         make(StatusMap),
		doneChannel:   make(chan struct{}),
		enableCaching: true,
		root:          root,

		maxParallel: runtime.NumCPU(),

		predictedNumOfTasks: 100000,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(p)
	}

	// Try to make the task channel the same size as the number of tasks.
	// (Matthias) There was a reason why this was neccessary, probably it's related
	// to beeing able to shutdown the playbook correctly? Unsure!
	p.taskChannel = make(chan *bobtask.Task, p.predictedNumOfTasks)

	return p
}

type RebuildCause string

func (rc *RebuildCause) String() string {
	return string(*rc)
}

const (
	TaskInputChanged      RebuildCause = "input-changed"
	TaskForcedRebuild     RebuildCause = "forced"
	DependencyChanged     RebuildCause = "dependency-changed"
	TargetInvalid         RebuildCause = "target-invalid"
	TargetNotInLocalStore RebuildCause = "target-not-in-localstore"
)

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

func (p *Playbook) artifactCreate(taskname string, hash hash.In) error {
	task, ok := p.Tasks[taskname]
	if !ok {
		return usererror.Wrap(boberror.ErrTaskDoesNotExistF(taskname))
	}
	return task.Task.ArtifactCreate(hash)
}

func (p *Playbook) storeBuildInfo(taskname string, buildinfo *buildinfo.I) error {
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
func (p *Playbook) TaskCompleted(taskname string) (err error) {
	defer errz.Recover(&err)

	task, ok := p.Tasks[taskname]
	if !ok {
		return usererror.Wrap(boberror.ErrTaskDoesNotExistF(taskname))
	}

	buildInfo, err := p.computeBuildinfo(taskname)
	errz.Fatal(err)

	// Store buildinfo
	err = p.storeBuildInfo(taskname, buildInfo)
	errz.Fatal(err)

	// Store targets in the artifact store
	if p.enableCaching {
		hashIn, err := task.HashIn()
		errz.Fatal(err)
		err = p.artifactCreate(taskname, hashIn)
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

// computeBuildinfo for a task.
// Should only be called after processing is done.
func (p *Playbook) computeBuildinfo(taskname string) (_ *buildinfo.I, err error) {
	defer errz.Recover(&err)

	task, ok := p.Tasks[taskname]
	if !ok {
		return nil, usererror.Wrap(boberror.ErrTaskDoesNotExistF(taskname))
	}

	hashIn, err := task.HashIn()
	errz.Fatal(err)

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
	trgt, err := task.Task.Target()
	errz.Fatal(err)
	if trgt != nil {
		bi, err := trgt.BuildInfo()
		if err != nil {
			if errors.Is(err, target.ErrTargetDoesNotExist) {
				return nil, usererror.Wrapm(err,
					fmt.Sprintf("Target does not exist for task [%s].\nDid you define the wrong target?\nDid you forget to create the target at all? \n\n", taskname))
			} else {
				errz.Fatal(err)
			}
		}

		buildInfo.Target = *bi
	}

	return buildInfo, nil
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
