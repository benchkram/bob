package tui

import (
	"fmt"
	"github.com/Benchkram/bob/pkg/ctl"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

type TUI struct {
	prog   *tea.Program
	events chan interface{}
	stdout *os.File
	stderr *os.File
	output *os.File
}

func New() (*TUI, error) {
	evts := make(chan interface{}, 1)

	stdout := os.Stdout
	stderr := os.Stderr

	rout, wout, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	os.Stdout = wout
	os.Stderr = wout

	return &TUI{
		prog:   nil,
		events: evts,
		stdout: stdout,
		stderr: stderr,
		output: rout,
	}, nil
}

func (t *TUI) Start(cmder ctl.Commander) {
	programEvts := make(chan interface{}, 1)

	t.prog = tea.NewProgram(
		newModel(cmder, t.events, programEvts, t.output),
		tea.WithAltScreen(),
		tea.WithoutCatchPanics(),
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

	defer func() {
		// restore outputs on exit of the TUI
		os.Stdout = t.stdout
		os.Stderr = t.stderr
	}()

	if err := t.prog.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
