package cmdutil

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/cli/cli/git"
)

type Runnable interface {
	Run() error
	Output() ([]byte, error)
	OutputCombined() ([]byte, error)
}

type run struct {
	cmd *exec.Cmd
}

func (run *run) Run() error {
	var stdout, stderr bytes.Buffer
	run.cmd.Stdout = &stdout
	run.cmd.Stderr = &stderr

	err := run.cmd.Run()
	if err != nil {
		return CmdError{
			Stderr: &stderr,
			Args:   run.cmd.Args,
			Err:    err,
		}
	}

	return nil
}

func (run *run) Output() ([]byte, error) {
	var stderr bytes.Buffer
	run.cmd.Stderr = &stderr

	output, err := run.cmd.Output()

	if err != nil {
		return nil, CmdError{
			Stderr: &stderr,
			Args:   run.cmd.Args,
			Err:    err,
		}
	}

	return output, nil
}

func (run *run) OutputCombined() ([]byte, error) {
	var b bytes.Buffer
	run.cmd.Stdout = &b
	run.cmd.Stderr = &b

	err := run.cmd.Run()
	if err != nil {
		return nil, CmdError{
			Stderr: &b,
			Args:   run.cmd.Args,
			Err:    err,
		}
	}

	return b.Bytes(), nil
}

// gitprepare inits git cmd with `root` as the working dir.
func gitprepare(root string, args ...string) (r Runnable, _ error) {
	cmd, err := git.GitCommand(args...)
	if err != nil {
		return r, fmt.Errorf("failed to make git command: %w", err)
	}

	cmd.Dir = root

	return &run{cmd}, nil
}

func RunGit(root string, args ...string) error {
	r, err := gitprepare(root, args...)
	if err != nil {
		return err
	}
	return r.Run()
}

func RunGitWithOutput(root string, args ...string) ([]byte, error) {
	r, err := gitprepare(root, args...)
	if err != nil {
		return nil, err
	}
	return r.OutputCombined()
}

func GitStatus(root string) ([]byte, error) {
	r, err := gitprepare(root, "status", "--porcelain")
	if err != nil {
		return nil, err
	}
	return r.Output()
}

func GitDryCommit(root string, message string) ([]byte, error) {
	r, err := gitprepare(root, "commit", "-m", strconv.Quote(message), "--dry-run", "--porcelain")
	if err != nil {
		return nil, err
	}
	return r.Output()
}

func GitCommit(root string, message string) ([]byte, error) {
	r, err := gitprepare(root, "commit", "-m", strconv.Quote(message))
	if err != nil {
		return nil, err
	}
	return r.OutputCombined()
}

func GitAddDry(root string, targetDir string) ([]byte, error) {
	r, err := gitprepare(root, "add", targetDir, "--dry-run")
	if err != nil {
		return nil, err
	}

	return r.Output()
}

func GitAdd(root string, targetDir string) error {
	r, err := gitprepare(root, "add", targetDir)
	if err != nil {
		return err
	}

	return r.Run()
}
