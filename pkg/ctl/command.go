package ctl

import "io"

type Command interface {
	Name() string

	Start() error
	Stop() error
	Restart() error
	Init() error
	Running() bool

	Shutdown() error
	Done() <-chan struct{}

	Stdout() io.Reader
	Stderr() io.Reader
	Stdin() io.Writer
}
