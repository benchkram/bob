package boblog

// rough logr.Logger implementation
// to be replaced with "logging"-branch

import "fmt"

var Log = log{level: 0}

var globalLogLevel = 1

func SetLogLevel(level int) {
	if level < 0 {
		level = 0
	}
	globalLogLevel = level
}

type log struct {
	level int
}

func (l log) V(level int) log {
	if level < 0 {
		return l
	}

	l.level = l.level + level
	return l
}

func (l log) Info(msg string) {
	if l.level > globalLogLevel {
		return
	}
	fmt.Println(msg)
}
