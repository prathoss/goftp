package screens

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type message struct {
	message  string
	returnFn returnFn
	onQuit   func()
}

type returnFn func() (tea.Model, tea.Cmd)

// initMessage creates a model with message, if returnTo is nil, the bubbletea will quit
func initMessage(messageText string, returnTo tea.Model, returnCmd tea.Cmd) (tea.Model, tea.Cmd) {
	return initMessageWithOnQuit(messageText, returnTo, returnCmd, nil)
}

func initMessageWithOnQuit(messageText string, returnTo tea.Model, returnCmd tea.Cmd, onQuit func()) (tea.Model, tea.Cmd) {
	rf := returnFn(func() (tea.Model, tea.Cmd) {
		return returnTo, returnCmd
	})
	if returnTo == nil {
		rf = func() (tea.Model, tea.Cmd) {
			return nil, tea.Quit
		}
	}
	return message{
		message:  messageText,
		returnFn: rf,
		onQuit:   onQuit,
	}, nil
}

func (m message) Init() tea.Cmd {
	return nil
}

func (m message) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, mKeys.Quit):
			if m.onQuit != nil {
				m.onQuit()
			}
			return m, tea.Quit
		case key.Matches(msg, mKeys.Ok):
			return m.returnFn()
		}
	}
	return m, nil
}

func (m message) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.
			NewStyle().
			Padding(2, 5).
			Border(lipgloss.RoundedBorder(), true).
			Render(m.message),
		help.New().View(mKeys),
	)
}

var mKeys = mKeyMap{
	Ok: key.NewBinding(
		key.WithKeys(tea.KeyEnter.String()),
		key.WithHelp("enter", "return"),
	),
	Quit: key.NewBinding(
		key.WithKeys(tea.KeyCtrlC.String(), "q"),
		key.WithHelp("ctrl+c/q", "quit"),
	),
}

type mKeyMap struct {
	Ok   key.Binding
	Quit key.Binding
}

func (m mKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{m.Ok, m.Quit}
}

func (m mKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{m.Ok, m.Quit}}
}
