package boblog

// rough logr.Logger implementation
// to be replaced with "logging"-branch

import (
	"errors"
	"fmt"
	"github.com/benchkram/bob/pkg/usererror"
	"unicode"

	"github.com/benchkram/errz"
	"github.com/logrusorgru/aurora"
)

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

func (l log) Error(err error, msg string, keysAndValues ...interface{}) {
	// Only log if there's actually an error
	if err == nil {
		return
	}

	// Error message will always be logged if exists
	fmt.Println(aurora.Red(msg))

	// Stack trace will only be logged if globalLogLevel >= 2
	if globalLogLevel >= 2 {
		errz.Log(err)
	} else {
		for {
			er := errors.Unwrap(err)
			if er == nil {
				break
			}

			err = er
		}

		fmt.Println(aurora.Red(err))
	}
}

// UserError is inteded to present errors to the user
// should go into a cli beatify package in the future..
func (l log) UserError(err error) {
	if err == nil {
		return
	}

	var uerr *usererror.E
	er := err
	for {
		if errors.As(er, &uerr) {
			err = uerr
			break
		}

		er = errors.Unwrap(er)
		if er == nil {
			break
		}
	}

	msg := err.Error()

	if msg != "" {
		tmp := []rune(msg)
		tmp[0] = unicode.ToUpper(tmp[0])
		msg = string(tmp)
	}

	fmt.Println(aurora.Red(msg))
}
