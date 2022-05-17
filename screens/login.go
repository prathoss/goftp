package screens

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/prathoss/goftp/pkg"
)

type loginModel struct {
	server         textinput.Model
	port           textinput.Model
	user           textinput.Model
	password       textinput.Model
	selectedCursor uint8
}

const (
	lmServerInput uint8 = iota
	lmPortInput
	lmUserInput
	lmPasswdInput
)

func InitLoginModel() (tea.Model, tea.Cmd) {
	server := textinput.New()
	server.Placeholder = "Server url"
	server.Focus()

	port := textinput.New()
	port.Placeholder = "Port"

	user := textinput.New()
	user.Placeholder = "User"

	passwd := textinput.New()
	passwd.Placeholder = "Password"
	passwd.EchoMode = textinput.EchoPassword
	passwd.EchoCharacter = '*'

	return loginModel{
		server:         server,
		port:           port,
		user:           user,
		password:       passwd,
		selectedCursor: 0,
	}, textinput.Blink
}

func InitLoginModelWithValues(server string, port int, user string) (tea.Model, tea.Cmd) {
	model, cmd := InitLoginModel()
	loginModel := model.(loginModel)
	loginModel.server.SetValue(server)
	loginModel.server.Blur()
	loginModel.port.SetValue(strconv.Itoa(port))
	loginModel.port.Blur()
	loginModel.user.SetValue(user)
	loginModel.user.Blur()
	loginModel.password.Focus()
	loginModel.selectedCursor = lmPasswdInput
	return loginModel, cmd
}

func (l loginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (l loginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return l, tea.Quit
		case tea.KeyEnter:
			// on enter login to server and move to next screen
			port, err := strconv.Atoi(l.port.Value())
			if err != nil {
				return initMessage("Port must be only numeric", l, textinput.Blink)
			}
			files, err := initFiles(l.server.Value(), port, l.user.Value(), l.password.Value())
			if err != nil {
				return initMessage(fmt.Sprintf("Could not login to server: %s", err.Error()), l, textinput.Blink)
			}
			if err := pkg.AddToConfig(pkg.ServerConf{
				Server: l.server.Value(),
				Port:   port,
				User:   l.user.Value(),
			}); err != nil {
				return initMessage(fmt.Sprintf("Could not save connection: %s", err.Error()), files, nil)
			}
			return files, nil
		case tea.KeyTab, tea.KeyDown:
			if l.selectedCursor < lmPasswdInput {
				l.selectedCursor++
				l.blurUnselected()
				cmd := l.focusInput()
				return l, cmd
			}
			return l, nil
		case tea.KeyShiftTab, tea.KeyUp:
			if l.selectedCursor > 0 {
				l.selectedCursor--
				l.blurUnselected()
				cmd := l.focusInput()
				return l, cmd
			}
			return l, nil
		}
	}

	cmd := l.updateInput(msg)
	return l, cmd
}

func (l *loginModel) updateInput(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if l.selectedCursor == lmPortInput && msg.Runes != nil && !pkg.IsMsgNumeric(msg) {
			return nil
		}
	}

	m, cmd := l.getSelectedInput().Update(msg)
	l.setSelectedInput(m)
	return cmd
}

func (l *loginModel) focusInput() tea.Cmd {
	return l.getSelectedInput().Focus()
}

func (l *loginModel) getSelectedInput() *textinput.Model {
	switch l.selectedCursor {
	case lmServerInput:
		return &l.server
	case lmPortInput:
		return &l.port
	case lmUserInput:
		return &l.user
	default:
		return &l.password
	}
}

func (l *loginModel) blurUnselected() {
	inputs := l.getUnselectedInputs()
	for i := 0; i < len(inputs); i++ {
		inputs[i].Blur()
	}
}

func (l *loginModel) getUnselectedInputs() []*textinput.Model {
	switch l.selectedCursor {
	case lmServerInput:
		return []*textinput.Model{&l.port, &l.user, &l.password}
	case lmPortInput:
		return []*textinput.Model{&l.server, &l.user, &l.password}
	case lmUserInput:
		return []*textinput.Model{&l.server, &l.port, &l.password}
	default:
		return []*textinput.Model{&l.server, &l.port, &l.user}
	}
}

func (l *loginModel) setSelectedInput(model textinput.Model) {
	switch l.selectedCursor {
	case lmServerInput:
		l.server = model
	case lmPortInput:
		l.port = model
	case lmUserInput:
		l.user = model
	default:
		l.password = model
	}
}

func (l loginModel) View() string {
	var b strings.Builder
	b.WriteString("Log in:\n")
	inputs := []textinput.Model{l.server, l.port, l.user, l.password}
	for i, input := range inputs {
		b.WriteString(input.View())
		if i < len(inputs)-1 {
			b.WriteRune('\n')
		}
	}
	return b.String()
}
