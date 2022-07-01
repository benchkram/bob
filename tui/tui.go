package tui

import (
	"fmt"
	"os"
	"time"

	"github.com/benchkram/bob/pkg/ctl"
	"github.com/benchkram/errz"
	tea "github.com/charmbracelet/bubbletea"
)

type TUI struct {
	prog    *tea.Program
	events  chan interface{}
	stdout  *os.File
	stderr  *os.File
	output  *os.File
	buffer  *LineBuffer
	started bool
}

func New() (*TUI, error) {

	evts := make(chan interface{}, 1024)

	stdout := os.Stdout
	stderr := os.Stderr

	rout, wout, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	// Redirect of stderr and stdout to cache all logging until tui is either started or restored
	os.Stdout = wout
	os.Stderr = wout

	// FIXME: this is a hack to indicate a line to read to multiScanner().
	// If we don't do this, the scanner assumes theres is nothing to do and shuts down.
	// The TUI does not start and the program will exit.
	//
	// This happens only occasionaly on mid-size projects (i've no idea why).
	// Though, it works fine with the standard server-db example.
	fmt.Fprint(wout, "\n")

	buf, err := multiScanner(0, evts, rout)
	if err != nil {
		errz.Log(err)
	}

	return &TUI{
		prog:   nil,
		events: evts,
		stdout: stdout,
		stderr: stderr,
		output: wout,
		buffer: buf,
	}, nil
}

func (t *TUI) Start(cmder ctl.Commander) {

	t.started = true

	programEvts := make(chan interface{}, 1)

	// Create a bubletea program which takes control over stdout.
	t.prog = tea.NewProgram(
		newModel(cmder, t.events, programEvts, t.buffer),
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(t.stdout),
	)

	go func() {
		for e := range programEvts {
			switch e.(type) {
			case EnableScroll:
				t.prog.EnableMouseAllMotion()
			case DisableScroll:
				t.prog.DisableMouseAllMotion()
			}
		}
	}()

	if err := t.prog.Start(); err != nil {
		fmt.Printf("TUI runtime error: %v", err)
		os.Exit(1)
	}
}

func (t *TUI) Restore() {
	// wait for commander to finish printing
	time.Sleep(10 * time.Millisecond)

	os.Stdout = t.stdout
	os.Stderr = t.stderr

	t.output.Close()

	if !t.started {
		for _, l := range t.buffer.Lines(0, t.buffer.Len()) {
			fmt.Println(l)
		}
	}
}
