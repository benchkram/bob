package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/ctl"
	"github.com/benchkram/bob/pkg/usererror"

	"github.com/benchkram/errz"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
	"github.com/xlab/treeprint"
)

func init() {
	treeprint.EdgeTypeLink = "│"
	treeprint.EdgeTypeMid = "├"
	treeprint.EdgeTypeEnd = "└"
	treeprint.IndentSize = 2
}

type keyMap struct {
	NextTab      key.Binding
	FollowOutput key.Binding
	Restart      key.Binding
	Quit         key.Binding
	SelectScroll key.Binding
	Up           key.Binding
	Down         key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Restart, k.NextTab, k.FollowOutput, k.SelectScroll, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Restart, k.NextTab, k.FollowOutput, k.SelectScroll, k.Quit},
	}
}

var keys = keyMap{
	NextTab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("[TAB]", "next tab"),
	),
	FollowOutput: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("[ESC]", "follow output"),
	),
	Restart: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("[^R]", "restart task"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("[^C]", "quit"),
	),
	SelectScroll: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("[^S]", "select text"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "pgup", "wheel up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "pgdown", "wheel down"),
	),
}

type model struct {
	keys          keyMap
	scroll        bool
	events        chan interface{}
	programEvents chan interface{}
	cmder         ctl.Commander
	tabs          []*tab
	currentTab    int
	starting      bool
	restarting    bool
	stopping      bool
	width         int
	height        int
	content       viewport.Model
	header        viewport.Model
	footer        help.Model
	follow        bool
	scrollOffset  int
	ready         bool
	error         error
}

type tab struct {
	name   string
	output *LineBuffer
}

func newModel(cmder ctl.Commander, evts, programEvts chan interface{}, buffer *LineBuffer) *model {
	tabs := []*tab{}

	tabs = append(tabs, &tab{
		name:   "status",
		output: buffer,
	})

	for i, cmd := range cmder.Subcommands() {
		buf, err := multiScanner(i+1, evts, cmd.Stdout(), cmd.Stderr())
		if err != nil {
			errz.Log(err)
		}

		tabs = append(tabs, &tab{
			name:   cmd.Name(),
			output: buf,
		})
	}

	return &model{
		cmder:         cmder,
		currentTab:    0,
		scroll:        true,
		tabs:          tabs,
		events:        evts,
		programEvents: programEvts,
		keys:          keys,
		follow:        true,
		footer: help.Model{
			ShowAll:        false,
			ShortSeparator: " · ",
			FullSeparator:  "",
			Ellipsis:       "...",
			Styles: help.Styles{
				ShortKey:  lipgloss.NewStyle().Foreground(lipgloss.Color("#bbb")),
				ShortDesc: lipgloss.NewStyle().Foreground(lipgloss.Color("#999")),
			},
		},
	}
}

func multiScanner(tabId int, events chan interface{}, rs ...io.Reader) (*LineBuffer, error) {
	buf := NewLineBuffer(120) // use some default width

	for _, r := range rs {
		s := bufio.NewScanner(r)
		s.Split(bufio.ScanLines)

		go func() {
			i := 0
			for s.Scan() {
				err := s.Err()
				if err != nil {
					return
				}

				_, _ = buf.Write(s.Bytes())

				i++

				events <- Update{tab: tabId}
			}
		}()
	}

	return buf, nil
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		start(m),
		nextEvent(m.events),
		tick(),
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	updateHeader := false

	switch msg := msg.(type) {

	case tea.KeyMsg:
		// for _, r := range msg.Runes {
		//	print(fmt.Sprintf("%s\n", strconv.QuoteRuneToASCII(r)))
		// }

		switch {
		// case string(msg.Runes[0]) == "[":
		//	errz.Log(fmt.Errorf("%#v", msg.Runes[0]))

		case key.Matches(msg, m.keys.NextTab):
			m.currentTab = (m.currentTab + 1) % len(m.tabs)

			m.setOffset(m.tabs[m.currentTab].output.Len())
			m.updateContent()
			updateHeader = true

		case key.Matches(msg, m.keys.FollowOutput):
			m.follow = true
			m.setOffset(m.tabs[m.currentTab].output.Len())
			m.updateContent()

		case key.Matches(msg, m.keys.Restart):
			status := fmt.Sprintf("\n%-*s\n", 10, "restarting")
			status = aurora.Colorize(status, aurora.CyanFg|aurora.BoldFm).String()

			for i, t := range m.tabs {
				_, err := t.output.Write([]byte(status))
				errz.Log(err)

				m.events <- Update{tab: i}
			}

			m.follow = true
			m.setOffset(m.tabs[m.currentTab].output.Len())
			m.updateContent()
			updateHeader = true

			cmds = append(cmds, restart(m))

		case key.Matches(msg, m.keys.Quit):
			status := fmt.Sprintf("\n%-*s\n", 10, "stopping")
			status = aurora.Colorize(status, aurora.RedFg|aurora.BoldFm).String()

			for i, t := range m.tabs {
				_, err := t.output.Write([]byte(status))
				errz.Log(err)

				m.events <- Update{tab: i}
			}

			m.follow = true
			m.setOffset(m.tabs[m.currentTab].output.Len())
			m.updateContent()
			updateHeader = true

			cmds = append(cmds, stop(m))

		case key.Matches(msg, m.keys.SelectScroll):
			scroll := !m.scroll
			if scroll {
				m.programEvents <- EnableScroll{}
				m.keys.SelectScroll.SetHelp("[^S]", "select text")
			} else {
				m.programEvents <- DisableScroll{}
				m.keys.SelectScroll.SetHelp("[^S]", "scroll text")
			}

			m.scroll = scroll

		case key.Matches(msg, m.keys.Up):
			m.follow = false
			m.updateOffset(-1)
			m.updateContent()

		case key.Matches(msg, m.keys.Down):
			m.updateOffset(1)
			m.updateContent()
		}

	case tea.MouseMsg:
		switch {
		case msg.Type == tea.MouseWheelUp:
			m.follow = false
			m.updateOffset(-1)
			m.updateContent()

		case msg.Type == tea.MouseWheelDown:
			m.updateOffset(1)
			m.updateContent()
		}

	case time.Time:
		cmds = append(cmds, tick())

	case tea.WindowSizeMsg:
		if !m.ready {
			// initialize viewports
			m.header.SetContent("\n")
			m.updateOffset(0)
			if m.follow {
				m.updateOffset(m.tabs[m.currentTab].output.Len())
			}
			m.updateContent()
			updateHeader = true
			m.ready = true
		}

		m.width = msg.Width
		m.height = msg.Height

		m.header.Width = m.width
		m.header.Height = 2

		m.content.Width = m.width
		m.content.Height = m.height - 4

		m.footer.Width = m.width

		for _, t := range m.tabs {
			// update all lines in the buffers so that soft wrapping works nicely
			t.output.SetWidth(m.width)
		}

		if m.follow {
			m.updateOffset(m.tabs[m.currentTab].output.Len())
		}
		// re-render content after resize
		m.updateContent()

	case Quit:
		cmds = append(cmds, tea.Quit)

	case Started:
		m.starting = false
		updateHeader = true

	case Restarted:
		m.restarting = false
		updateHeader = true

	case Update:
		// ignore updates for tabs that are not currently in view
		if msg.tab == m.currentTab {
			// scroll to end for the current tab if following output
			if m.follow {
				m.updateOffset(m.tabs[m.currentTab].output.Len())
			}

			// always re-render content if a new message is received
			m.updateContent()
		}

		// listen for the next update event
		cmds = append(cmds, nextEvent(m.events))
	}

	// only re-render the header if necessary
	if updateHeader {
		// calculate header status
		var status string
		if m.starting {
			status = fmt.Sprintf("%-*s", 10, "starting")
			status = aurora.Colorize(status, aurora.BlueFg|aurora.BoldFm).String()
		} else if m.restarting {
			status = fmt.Sprintf("%-*s", 10, "restarting")
			status = aurora.Colorize(status, aurora.CyanFg|aurora.BoldFm).String()
		} else if m.stopping {
			status = fmt.Sprintf("%-*s", 10, "stopping")
			status = aurora.Colorize(status, aurora.RedFg|aurora.BoldFm).String()
		} else {
			status = fmt.Sprintf("%-*s", 10, "running")
			status = aurora.Colorize(status, aurora.GreenFg|aurora.BoldFm).String()
		}

		// create tabs
		tabs := make([]string, len(m.tabs))
		for i, tab := range m.tabs {
			var name string
			if i == m.currentTab {
				name = aurora.Colorize(tab.name, aurora.BoldFm).String()
			} else {
				name = aurora.Colorize(tab.name, aurora.WhiteFg).String()
			}

			tabs[i] = fmt.Sprintf("[%s]", name)
		}

		tabsView := strings.Join(tabs, " ")

		m.header.SetContent(fmt.Sprintf("%s  %s", status, tabsView))
		m.header, _ = m.header.Update(msg)
	}

	var updateCmd tea.Cmd
	m.content, updateCmd = m.content.Update(msg)
	cmds = append(cmds, updateCmd)

	// if m.error != nil {
	//	//cmds = append(cmds, stop(m))
	// }

	return m, tea.Batch(cmds...)
}

func (m *model) updateContent() {
	buf := m.tabs[m.currentTab].output
	viewportHeight := m.height - 4
	bufLen := buf.Len()

	offset := m.scrollOffset

	maxOffset := bufLen

	from := min(offset, maxOffset)
	to := max(offset+viewportHeight, 0)

	lines := buf.Lines(from, to)

	m.content.SetContent(strings.Join(lines, "\n"))
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func (m *model) View() string {
	var view strings.Builder

	view.WriteString(m.header.View())
	view.WriteString("\n")
	view.WriteString(m.content.View())
	view.WriteString("\n\n")
	view.WriteString(m.footer.View(m.keys))

	return view.String()
}

func (m *model) updateOffset(delta int) {
	m.setOffset(m.scrollOffset + delta*3)
}

func (m *model) setOffset(offset int) {
	viewportHeight := m.height - 4
	buf := m.tabs[m.currentTab].output
	bufLen := buf.Len()

	maxOffset := bufLen

	if maxOffset > viewportHeight {
		maxOffset -= viewportHeight

		if offset == bufLen-viewportHeight {
			m.follow = true
		}

	} else {
		maxOffset = 0 // do not allow scrolling if there is nothing else to see
		m.follow = true
	}

	offset = max(min(offset, maxOffset), 0)

	m.scrollOffset = offset
}

func tick() tea.Cmd {
	return tea.Tick(
		1000*time.Millisecond, func(t time.Time) tea.Msg {
			return t
		},
	)
}

func start(m *model) tea.Cmd {
	m.starting = true

	return func() tea.Msg {
		err := m.cmder.Start()
		if errors.As(err, &usererror.Err) {
			boblog.Log.UserError(err)
		} else if err != nil && err != context.Canceled {
			boblog.Log.Error(err, "Error during commander execution")
		}

		m.error = err

		return Started{}
	}
}

func restart(m *model) tea.Cmd {
	m.restarting = true

	return func() tea.Msg {
		err := m.cmder.Restart()
		errz.Log(err)

		return Restarted{}
	}
}

func stop(m *model) tea.Cmd {
	m.stopping = true
	m.programEvents <- DisableScroll{}

	return func() tea.Msg {
		err := m.cmder.Stop()
		errz.Log(err)

		return Quit{}
	}
}

func nextEvent(evts chan interface{}) tea.Cmd {
	return func() tea.Msg {
		return <-evts
	}
}
