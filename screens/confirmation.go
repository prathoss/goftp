package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type confirmation struct {
	message   string
	returnFn  returnFn
	onQuit    func()
	onConfirm onConfirmFn
	isOk      bool
}

type onConfirmFn func() error

func initConfirmation(messageText string, returnTo tea.Model, returnCmd tea.Cmd, onConfirm onConfirmFn, onQuit func()) (tea.Model, tea.Cmd) {
	rf := returnFn(func() (tea.Model, tea.Cmd) {
		return returnTo, returnCmd
	})
	if returnTo == nil {
		rf = func() (tea.Model, tea.Cmd) {
			return nil, tea.Quit
		}
	}
	return confirmation{
		message:   messageText,
		returnFn:  rf,
		onQuit:    onQuit,
		onConfirm: onConfirm,
		isOk:      true,
	}, nil
}

func (m confirmation) Init() tea.Cmd {
	return nil
}

func (m confirmation) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, cKeys.Switch):
			m.isOk = !m.isOk
			return m, nil
		case key.Matches(msg, cKeys.Ok):
			if m.isOk {
				if err := m.onConfirm(); err != nil {
					returnModel, returnCmd := m.returnFn()
					return initMessageWithOnQuit(
						fmt.Sprintf("Action failed: %s", err.Error()),
						returnModel,
						returnCmd,
						m.onQuit,
					)
				}
			}
			return m.returnFn()
		}
	}
	return m, nil
}

func (m confirmation) View() string {
	ok := "Ok"
	cancel := "Cancel"

	if m.isOk {
		ok = fmt.Sprintf(">%s", ok)
	} else {
		cancel = fmt.Sprintf(">%s", cancel)
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.
			NewStyle().
			Padding(2, 5).
			Border(lipgloss.RoundedBorder(), true).
			Render(m.message),
		ok,
		cancel,
		help.New().View(mKeys),
	)
}

var cKeys = cKeyMap{
	Ok: key.NewBinding(
		key.WithKeys(tea.KeyEnter.String()),
		key.WithHelp("enter", "return"),
	),
	Switch: key.NewBinding(
		key.WithKeys(tea.KeyUp.String(), "j", tea.KeyDown.String(), "k"),
		key.WithHelp("↓/j/↑/k", "switch answer"),
	),
}

type cKeyMap struct {
	Ok     key.Binding
	Switch key.Binding
}

func (m cKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{m.Ok, m.Switch}
}

func (m cKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{m.Ok, m.Switch}}
}
