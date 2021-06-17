package cmdutil

// copied from https://github.com/cli/cli/blob/trunk/internal/run/run.go

import (
	"bytes"
	"fmt"
	"strings"
)

type CmdError struct {
	Stderr *bytes.Buffer
	Args   []string
	Err    error
}

func (e CmdError) Error() string {
	msg := e.Stderr.String()
	if msg != "" && !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	return fmt.Sprintf("%s%s: %s", msg, e.Args[0], e.Err)
}
