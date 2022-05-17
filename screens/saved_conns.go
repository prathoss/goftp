package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prathoss/goftp/pkg"
)

type savedConnections struct {
	selected int
	confs    []pkg.ServerConf
}

func InitSavedConnections(confs []pkg.ServerConf) tea.Model {
	return savedConnections{
		selected: 0,
		confs:    confs,
	}
}

func (s savedConnections) Init() tea.Cmd {
	return nil
}

func (s savedConnections) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, scKeys.Down):
			if s.selected < len(s.confs)-1 {
				s.selected++
			}
			return s, nil
		case key.Matches(msg, scKeys.Up):
			if s.selected > 0 {
				s.selected--
			}
			return s, nil
		case key.Matches(msg, scKeys.Select):
			selectedServer := s.confs[s.selected]
			return InitLoginModelWithValues(selectedServer.Server, selectedServer.Port, selectedServer.User)
		case key.Matches(msg, scKeys.Skip):
			return InitLoginModel()
		case key.Matches(msg, scKeys.Quit):
			return s, tea.Quit
		}
	}
	return s, nil
}

func (s savedConnections) View() string {
	lines := make([]string, len(s.confs)+1)
	for i, conf := range s.confs {
		selector := " "
		if i == s.selected {
			selector = ">"
		}
		lines[i] = fmt.Sprintf("%s%s@%s", selector, conf.User, conf.Server)
	}
	lines = append(lines, help.New().View(scKeys))
	return strings.Join(lines, "\n")
}

var scKeys = scKeyMap{
	Up: key.NewBinding(
		key.WithKeys(tea.KeyUp.String(), "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys(tea.KeyDown.String(), "j"),
		key.WithHelp("↓/j", "down"),
	),
	Select: key.NewBinding(
		key.WithKeys(tea.KeyEnter.String()),
		key.WithHelp("enter", "select"),
	),
	Skip: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "skip"),
	),
	Quit: key.NewBinding(
		key.WithKeys(tea.KeyCtrlC.String(), "q"),
		key.WithHelp("ctrl+c/q", "quit"),
	),
}

type scKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Skip   key.Binding
	Quit   key.Binding
}

func (s scKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{s.Up, s.Down, s.Select, s.Skip, s.Quit}
}

func (s scKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{s.Up, s.Down, s.Select, s.Skip, s.Quit}}
}
