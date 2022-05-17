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

type filesModel struct {
	source      components.FileListModel
	destination components.FileListModel
	ftpClient   *ftp.ServerConn
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
	serverList, err := components.InitFileListModelBuilder(server, "/", func(location string) ([]types.Entry, error) {
		files, err := c.List(location)
		if err != nil {
			return nil, err
		}
		return pkg.MapSlice(files, func(f *ftp.Entry) types.Entry {
			var tp int
			switch f.Type {
			case ftp.EntryTypeFile:
				tp = types.TypeFile
			case ftp.EntryTypeLink:
				tp = types.TypeLink
			case ftp.EntryTypeFolder:
				tp = types.TypeDirectory
			}
			return types.Entry{
				Name: f.Name,
				Type: tp,
				Size: f.Size,
			}
		}), nil
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
		return pkg.MapSlice(files, func(f os.DirEntry) types.Entry {
			var tp int
			switch {
			case f.IsDir():
				tp = types.TypeDirectory
			default:
				tp = types.TypeFile
			}

			info, _ := f.Info()
			return types.Entry{
				Name: f.Name(),
				Type: tp,
				Size: uint64(info.Size()),
			}
		}), nil
	}).
		WithTransferFn(pkg.PrepareUploadFn(c)).
		WithDeleteFn(pkg.OsDeleteFn).
		Build()

	return filesModel{
		source:      localList,
		destination: serverList,
		ftpClient:   c,
	}, nil
}

func (m filesModel) Init() tea.Cmd {
	return nil
}

func (m filesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, fKeys.Quit):
			_ = m.ftpClient.Quit()
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
				m,
				nil,
				func() error {
					err := m.source.Delete()
					if err != nil {
						return err
					}
					return nil
				},
				func() {
					_ = m.ftpClient.Quit()
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
			_ = m.ftpClient.Quit()
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
