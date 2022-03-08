package cmdutil

import (
	"bytes"
	"fmt"
	"os"
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

func RemoveFromKnownHost(host string, port int) ([]byte, error) {
	hostkey := fmt.Sprintf("[%s]:%d", host, port)
	cmd := exec.Command("ssh-keygen", "-R", hostkey)

	cmd.Env = os.Environ()

	r := &run{cmd}

	return r.OutputCombined()
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
	r, err := gitprepare(root, "add", targetDir, "--dry-run", "--verbose")
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

func GitUnpushedCommits(root string) ([]byte, error) {
	r, err := gitprepare(root, "cherry", "-v")
	if err != nil {
		return nil, err
	}

	return r.Output()
}

func GitPushDry(root string, remote string, ref string) ([]byte, error) {
	r, err := gitprepare(root, "push", "--set-upstream", remote, ref, "--dry-run")
	if err != nil {
		return nil, fmt.Errorf("failed to make git command: %w", err)
	}

	return r.OutputCombined()
}

func GitPush(root string, remote string, ref string) ([]byte, error) {
	r, err := gitprepare(root, "push", "--set-upstream", remote, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to make git command: %w", err)
	}

	return r.OutputCombined()
}

// GitPushFirstTime, runs git push command for the first time with remote and branch name and
// disable ssh checking if ssh set to true
func GitPushFirstTime(root string, remote string, branch string, ssh bool) error {
	if ssh {
		err := DisableSSHChecking(root)
		if err != nil {
			return err
		}
	}

	r, err := gitprepare(root, "push", "-u", remote, branch)
	if err != nil {
		return fmt.Errorf("failed to make git command: %w", err)
	}

	return r.Run()
}

func DisableSSHChecking(root string) error {
	sshcommand := "ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no"

	r, err := gitprepare(root, "config", "core.sshCommand", sshcommand)
	if err != nil {
		return fmt.Errorf("failed to config git command: %w", err)
	}

	return r.Run()
}
