package screens

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jlaffaye/ftp"
	"github.com/prathoss/goftp/components"
	"github.com/prathoss/goftp/pkg"
	"github.com/prathoss/goftp/types"
)

type ftpModel struct {
	client             *ftp.ServerConn
	keepAliveQuitChan  chan<- struct{}
	keepAliveErrorChan <-chan error
	ftpAliveError      error
}

type filesModel struct {
	source      components.FileListModel
	destination components.FileListModel
	ftpModel    *ftpModel
}

func initFiles(server string, port int, user, passwd string) (tea.Model, error) {
	// server
	c, err := ftp.DialTimeout(fmt.Sprintf("%s:%d", server, port), 5*time.Second)
	if err != nil {
		return nil, err
	}
	err = c.Login(user, passwd)
	if err != nil {
		return nil, err
	}
	ftpKeepAliveQuit, ftpKeepAliveError := pkg.KeepFtpAlive(c)
	ftpModel := &ftpModel{
		client:             c,
		keepAliveQuitChan:  ftpKeepAliveQuit,
		keepAliveErrorChan: ftpKeepAliveError,
	}
	// listen for possible error from keeping ftp alive
	go func() {
		ftpModel.ftpAliveError = <-ftpKeepAliveError
	}()
	serverList, err := components.InitFileListModelBuilder(server, "/", func(location string) ([]types.Entry, error) {
		files, err := c.List(location)
		if err != nil {
			return nil, err
		}
		return pkg.MapSlice(files, pkg.FtpToEntry), nil
	}).
		WithTransferFn(pkg.PrepareDownloadFn(c)).
		WithDeleteFn(pkg.PrepareFtpDeleteFn(c)).
		Build()

	// local
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	localList, err := components.InitFileListModelBuilder("Local", dir, func(location string) ([]types.Entry, error) {
		files, err := os.ReadDir(location)
		if err != nil {
			return nil, err
		}
		return pkg.MapSlice(files, pkg.OsToEntry), nil
	}).
		WithTransferFn(pkg.PrepareUploadFn(c)).
		WithDeleteFn(pkg.OsDeleteFn).
		Build()

	return filesModel{
		source:      localList,
		destination: serverList,
		ftpModel:    ftpModel,
	}, nil
}

func (m filesModel) Init() tea.Cmd {
	return nil
}

func (m filesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if err := m.ftpModel.ftpAliveError; err != nil {
		cfg, _ := pkg.GetConfig()

		return initMessage(
			fmt.Sprintf("Connection with server lost: %s", err.Error()),
			InitSavedConnections(cfg.Servers),
			nil,
		)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, fKeys.Quit):
			_ = m.Close()
			return m, tea.Quit
		case key.Matches(msg, fKeys.Down):
			m.source.Down()
		case key.Matches(msg, fKeys.Up):
			m.source.Up()
		case key.Matches(msg, fKeys.Enter):
			err := m.source.Enter()
			if err != nil {
				return m.sendMessage(fmt.Sprintf("Could not open dir: %s", err.Error()))
			}
			return m, nil
		case key.Matches(msg, fKeys.Return):
			err := m.source.Return()
			if err != nil {
				return m.sendMessage(fmt.Sprintf("Could not open dir: %s", err.Error()))
			}
			return m, nil
		case key.Matches(msg, fKeys.Transfer):
			if err := m.source.Transfer(m.destination.GetLocation()); err != nil {
				return m.sendMessage(fmt.Sprintf("Could not transfer files: %s", err.Error()))
			}
			if err := m.destination.Refresh(); err != nil {
				return m.sendMessage(fmt.Sprintf("Could not refresh files: %s", err.Error()))
			}
			m.source.DeselectAll()
		case key.Matches(msg, fKeys.Switch):
			m.source, m.destination = m.destination, m.source
		case key.Matches(msg, fKeys.ToggleSelection):
			m.source.ToggleSelection()
		case key.Matches(msg, fKeys.Delete):
			return initConfirmation(fmt.Sprintf("Realy want to delete %d files", m.source.GetSelectedCount()),
				&m,
				nil,
				m.source.Delete,
				func() {
					_ = m.Close()
				},
			)
		case key.Matches(msg, fKeys.Help):
			// TODO: implement
		}
	}
	return m, nil
}

func (m filesModel) sendMessage(message string) (tea.Model, tea.Cmd) {
	return initMessageWithOnQuit(
		message,
		m,
		nil,
		func() {
			_ = m.Close()
		},
	)
}

func (m filesModel) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().
			Margin(0, 0, 1).
			Render(fmt.Sprintf("Number of selected files: %d", m.source.GetSelectedCount())),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.source.View(true),
			lipgloss.NewStyle().
				Margin(0, 2).
				Render(""),
			m.destination.View(false),
		),
		help.New().View(fKeys),
	)
}

func (m filesModel) Close() error {
	m.ftpModel.keepAliveQuitChan <- struct{}{}
	return m.ftpModel.client.Quit()
}

var fKeys = fKeyMap{
	Up: key.NewBinding(
		key.WithKeys(tea.KeyUp.String(), "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys(tea.KeyDown.String(), "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys(tea.KeyEnter.String()),
		key.WithHelp("enter", "enter"),
	),
	Return: key.NewBinding(
		key.WithKeys(tea.KeyBackspace.String()),
		key.WithHelp("backspace", "return"),
	),
	Quit: key.NewBinding(
		key.WithKeys(tea.KeyCtrlC.String(), "q"),
		key.WithHelp("ctrl+c/q", "quit"),
	),
	Transfer: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "transfer"),
	),
	Switch: key.NewBinding(
		key.WithKeys(tea.KeyTab.String()),
		key.WithHelp("tab", "switch"),
	),
	ToggleSelection: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle selection"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

type fKeyMap struct {
	Up              key.Binding
	Down            key.Binding
	Enter           key.Binding
	Return          key.Binding
	Quit            key.Binding
	Transfer        key.Binding
	Switch          key.Binding
	ToggleSelection key.Binding
	Delete          key.Binding
	Help            key.Binding
}

func (f fKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{f.Up, f.Down, f.Enter, f.Return, f.Quit, f.Help}
}

func (f fKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{f.Up, f.Down, f.Enter, f.Return, f.Quit},
		{f.ToggleSelection, f.Transfer, f.Switch, f.Delete},
	}
}
