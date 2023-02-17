package playbook

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/benchkram/bob/bobtask"
	"github.com/benchkram/bob/bobtask/buildinfo"
	"github.com/benchkram/bob/bobtask/hash"
	"github.com/benchkram/bob/pkg/boberror"
	"github.com/benchkram/bob/pkg/store"
	"github.com/benchkram/bob/pkg/usererror"
	"github.com/benchkram/errz"
)

// The playbook defines the order in which tasks are allowed to run.
// Also determines the possibility to run tasks in parallel.

var ErrDone = fmt.Errorf("playbook is done")
var ErrFailed = fmt.Errorf("playbook failed")

type Playbook struct {
	// taskChannel is closed when the root
	// task completes.
	taskChannel chan *bobtask.Task

	// errorChannel to transport errors to the caller
	errorChannel chan error

	// root task
	root string
	// rootID for optimized access
	rootID int

	Tasks StatusMap
	// TasksOptimized uses a array instead of an map
	TasksOptimized StatusSlice

	namePad int

	done bool
	// doneChannel is closed when the playbook is done.
	doneChannel chan struct{}

	// start is the point in time the playbook started
	start time.Time

	// enableCaching allows artifacts to be read & written to a store.
	// Default: true.
	enableCaching bool

	// predictedNumOfTasks is used to pick
	// an appropriate channel size for the task queue.
	predictedNumOfTasks int

	// maxParallel is the maximum number of parallel executed tasks
	maxParallel int

	// remoteStore is the artifacts remote store
	remoteStore store.Store

	// localStore is the artifacts local store
	localStore store.Store

	// enablePush allows pushing artifacts to remote store
	enablePush bool

	// enablePull allows pulling artifacts from remote store
	enablePull bool

	// oncePrepareOptimizedAccess is used to initalize the optimized
	// slice to access tasks.
	oncePrepareOptimizedAccess sync.Once
}

func New(root string, rootID int, opts ...Option) *Playbook {
	p := &Playbook{
		errorChannel:   make(chan error),
		Tasks:          make(StatusMap),
		TasksOptimized: make(StatusSlice, 0),
		doneChannel:    make(chan struct{}),
		enableCaching:  true,
		root:           root,
		rootID:         rootID,

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
	InputNotFoundInBuildInfo RebuildCause = "input-not-in-build-info"
	TaskForcedRebuild        RebuildCause = "forced"
	DependencyChanged        RebuildCause = "dependency-changed"
	TargetInvalid            RebuildCause = "target-invalid"
	TargetNotInLocalStore    RebuildCause = "target-not-in-localstore"
)

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

func (p *Playbook) ExecutionTime() time.Duration {
	return time.Since(p.start)
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
func (p *Playbook) TaskCompleted(taskID int) (err error) {
	defer errz.Recover(&err)

	task := p.TasksOptimized[taskID]

	buildInfo, err := p.computeBuildinfo(task.Name())
	errz.Fatal(err)

	// Store buildinfo
	err = p.storeBuildInfo(task.Name(), buildInfo)
	errz.Fatal(err)

	// Store targets in the artifact store
	if p.enableCaching {
		hashIn, err := task.HashIn()
		errz.Fatal(err)
		err = p.artifactCreate(task.Name(), hashIn)
		errz.Fatal(err)
	}

	// update task state and trigger another playbook run
	err = p.setTaskState(taskID, StateCompleted, nil)
	errz.Fatal(err)

	return nil
}

// TaskNoRebuildRequired sets a task's state to indicate that no rebuild is required
func (p *Playbook) TaskNoRebuildRequired(taskID int) (err error) {
	defer errz.Recover(&err)

	err = p.setTaskState(taskID, StateNoRebuildRequired, nil)
	errz.Fatal(err)

	return nil
}

// TaskFailed sets a task to failed
func (p *Playbook) TaskFailed(taskID int, taskErr error) (err error) {
	defer errz.Recover(&err)

	err = p.setTaskState(taskID, StateFailed, taskErr)
	errz.Fatal(err)

	return nil
}

// TaskCanceled sets a task to canceled
func (p *Playbook) TaskCanceled(taskID int) (err error) {

	defer errz.Recover(&err)

	err = p.setTaskState(taskID, StateCanceled, nil)
	errz.Fatal(err)

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

func (p *Playbook) setTaskState(taskID int, state State, taskError error) error {
	task := p.TasksOptimized[taskID]

	task.SetState(state, taskError)
	switch state {
	case StateCompleted, StateCanceled, StateNoRebuildRequired, StateFailed:
		task.SetEnd(time.Now())
	}

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
