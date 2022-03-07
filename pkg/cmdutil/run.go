package cmdutil

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cli/cli/git"
)

type Runnable interface {
	Run() error
	Output() ([]byte, error)
	OutputCombined() ([]byte, error)
	RunPipe() error
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

func (run *run) RunPipe() error {
	stdin, err := run.cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	_, err = run.cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stderr, err := run.cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	// Wait for prompt
	promptDetected := func(bytes []byte) bool {
		frags := strings.Split(string(bytes), "\n")
		if len(frags) == 0 {
			return false
		}

		// fmt.Println(frags)

		last := frags[len(frags)-1]
		return strings.HasPrefix(last, "Are you sure you want to continue connecting (yes/no/[fingerprint])?")
	}

	prompt := make(chan bool, 1)

	go func(ch chan<- bool) {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanBytes)

		buff := []byte{}
		for scanner.Scan() {
			bytes := scanner.Bytes()
			buff = append(buff, bytes...)
			// fmt.Println(string(buff))
			if promptDetected(buff) {
				ch <- true
			}
		}
		ch <- true
	}(prompt)

	if err := run.cmd.Start(); err != nil {
		panic(err)
	}

	defer run.cmd.Wait()

	<-prompt

	// Send input to the prompt
	io.WriteString(stdin, "yes")
	io.WriteString(stdin, "\n")

	return nil
}

func RunAddToKnownHost(host string, port int) ([]byte, error) {
	cmd := exec.Command("ssh-keyscan", "-p", fmt.Sprint(port), "-H", host, ">>", "~/.ssh/known_hosts")

	cmd.Env = os.Environ()

	r := &run{cmd}

	fmt.Println(cmd)

	return r.OutputCombined()
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

func RunGitPipe(root string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Env = os.Environ()
	cmd.Dir = root

	r := &run{cmd}

	return r.RunPipe()
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
