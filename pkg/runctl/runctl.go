package runctl

import "os"

type Control interface {
	Control() <-chan os.Signal
	Stopped() <-chan struct{}
	Started() <-chan struct{}
	Error() <-chan error

	EmitSignal(os.Signal)
	EmitStop()
	EmitStarted()
	EmitError(error)
}

type runControl struct {
	ctl   chan os.Signal
	stop  chan struct{}
	start chan struct{}
	err   chan error
}

func New() *runControl {
	return &runControl{
		ctl:   make(chan os.Signal, 1),
		stop:  make(chan struct{}, 1),
		start: make(chan struct{}, 1),
		err:   make(chan error, 1),
	}
}

func (rc *runControl) Control() <-chan os.Signal {
	return rc.ctl
}
func (rc *runControl) Stopped() <-chan struct{} {
	return rc.stop
}
func (rc *runControl) Started() <-chan struct{} {
	return rc.start
}
func (rc *runControl) Error() <-chan error {
	return rc.err
}

func (rc *runControl) EmitSignal(s os.Signal) {
	rc.ctl <- s
}
func (rc *runControl) EmitStop() {
	close(rc.stop)
}
func (rc *runControl) EmitStarted() {
	rc.start <- struct{}{}
}
func (rc *runControl) EmitError(err error) {
	rc.err <- err
}
