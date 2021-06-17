package multilinecmd

import (
	"strings"
)

// Split splits a single-line "command" to a slice of commands.
//
// It currently supports the following syntaxes:
//
/*
  Input:
    echo Hello
  Output:
    [ "echo Hello", ]

  Input:
    echo Hello
    some long \
      command \
      with multiple line-breaks
  Output:
    [
      "echo Hello",
      "some long command with multiple line-breaks",
    ]
*/
func Split(cmd string) []string {
	// Remove backslash-newlines with a space.
	// This adds support to use single commands spanning across multiple lines.
	cmd = strings.ReplaceAll(cmd, "\\\n", " ")

	// Replace multiple spaces with a single one.
	// TODO: Reenable? echo 'Hello  World' will then no longer echo with two spaces.
	// cmd = regexp.MustCompile(` +`).ReplaceAllString(cmd, " ")

	// Replace leading and trailing spaces for clarity.
	cmd = strings.TrimSpace(cmd)

	cmds := strings.Split(cmd, "\n")
	if len(cmds) == 1 && len(cmds[0]) == 0 {
		return []string{}
	}

	return cmds
}
