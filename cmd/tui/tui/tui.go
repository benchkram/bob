package tui

import (
	"fmt"
	"github.com/Benchkram/bob/pkg/execctl"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

type TUI struct {
	prog   *tea.Program
	events chan interface{}
}

func New(tree execctl.CommandTree) (*TUI, error) {
	evts := make(chan interface{}, 1)

	prog := tea.NewProgram(
		newModel(tree, evts),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	return &TUI{
		prog:   prog,
		events: evts,
	}, nil
}

func (t *TUI) Start() {
	if err := t.prog.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
