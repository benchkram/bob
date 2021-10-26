package tui

import (
	"fmt"
	"os"

	"github.com/Benchkram/bob/pkg/ctl"
	tea "github.com/charmbracelet/bubbletea"
)

type TUI struct {
	prog   *tea.Program
	events chan interface{}
	stdout *os.File
	stderr *os.File
}

func New(cmder ctl.Commander) (*TUI, error) {
	evts := make(chan interface{}, 1)

	stdout := os.Stdout
	stderr := os.Stderr

	rout, wout, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	os.Stdout = wout
	os.Stderr = wout

	prog := tea.NewProgram(
		newModel(cmder, evts, rout, stdout),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(stdout),
	)

	return &TUI{
		prog:   prog,
		events: evts,
		stdout: stdout,
		stderr: stderr,
	}, nil
}

func (t *TUI) Start() {
	defer func() {
		// restore outputs
		os.Stdout = t.stdout
		os.Stderr = t.stderr
	}()

	if err := t.prog.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
