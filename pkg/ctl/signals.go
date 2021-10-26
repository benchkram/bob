package ctl

import "syscall"

// signal summary for internal use
type Signal syscall.Signal

const (
	Start    = Signal(0x71)
	Stop     = Signal(syscall.SIGSTOP)
	Restart  = Signal(0x70)
	Shutdown = Signal(syscall.SIGINT)
)
