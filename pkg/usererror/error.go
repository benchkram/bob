package usererror

import (
	"fmt"
)

// Err is a instatiation of E
// to be used for comparison with `errors.As(usererror.Err)`
// DO NOT WRITE TO Err!!
var Err = &E{}

type E struct {
	// err is the underlying error
	err error

	// msg contains a useful general message usualy shown to a user,
	// can be colored.
	msg string

	// summary of a underlying error.
	// Could be the last 5 lines of a build error.
	// For future use!
	summary string // nolint:structcheck, unused
}

func (e *E) Error() string {
	if e.msg == "" {
		return e.err.Error()
	}

	return fmt.Sprintf("%s: %s", e.msg, e.err)
}

func (e *E) Msg() string {
	return e.msg
}

func (e *E) Unwrap() error {
	return e.err
}

// Wrap an existing error and annotate it with the user error type
// which is intended to be shown to the user on a cli.
func Wrap(err error) *E {
	return &E{err: err}
}

func Wrapm(err error, msg string) *E {
	return &E{err: err, msg: msg}
}
