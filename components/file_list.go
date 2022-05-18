package components

import (
	"errors"
	"fmt"
	"path"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/prathoss/goftp/pkg"
	"github.com/prathoss/goftp/types"
)

type FileListModelBuilder struct {
	name       string
	location   string
	listFn     listFn
	transferFn transferFn
	deleteFn   deleteFn
}

func InitFileListModelBuilder(name, location string, listFn listFn) *FileListModelBuilder {
	return &FileListModelBuilder{
		name:     name,
		location: location,
		listFn:   listFn,
	}
}

func (f FileListModelBuilder) WithTransferFn(fn transferFn) FileListModelBuilder {
	f.transferFn = fn
	return f
}

func (f FileListModelBuilder) WithDeleteFn(fn deleteFn) FileListModelBuilder {
	f.deleteFn = fn
	return f
}

func (f FileListModelBuilder) Build() (FileListModel, error) {
	return InitFileListModel(f.name, f.location, f.listFn, f.transferFn, f.deleteFn)
}

type FileListModel struct {
	name         string
	location     string
	entries      []types.Entry
	cursor       int
	selected     map[int]string
	listFn       listFn
	topItemIndex int
	itemsInVew   int
	transferFn   transferFn
	deleteFn     deleteFn
}

type listFn func(location string) ([]types.Entry, error)

type transferFn func(string, []types.Entry, string) error

type deleteFn func(location string, entries []types.Entry) error

var ErrNotSet = errors.New("function not set")

func InitFileListModel(name, location string, listFn listFn, transferFn transferFn, deleteFn deleteFn) (FileListModel, error) {
	flm := FileListModel{
		name:         name,
		location:     location,
		entries:      []types.Entry{},
		cursor:       0,
		selected:     map[int]string{},
		listFn:       listFn,
		topItemIndex: 0,
		itemsInVew:   10,
		transferFn:   transferFn,
		deleteFn:     deleteFn,
	}
	if err := flm.Refresh(); err != nil {
		return FileListModel{}, err
	}
	return flm, nil
}

func (m *FileListModel) Up() {
	if m.cursor <= 0 {
		return
	}
	m.cursor--
	if m.cursor < m.topItemIndex {
		m.topItemIndex = m.cursor
	}
}

func (m *FileListModel) Down() {
	if m.cursor >= len(m.entries)-1 {
		return
	}
	m.cursor++
	if m.cursor > m.topItemIndex+m.itemsInVew-1 {
		m.topItemIndex = m.cursor - m.itemsInVew + 1
	}
}

func (m *FileListModel) ToggleSelection() {
	if len(m.entries) == 0 {
		return
	}
	if _, exists := m.selected[m.cursor]; exists {
		delete(m.selected, m.cursor)
	} else {
		m.selected[m.cursor] = m.entries[m.cursor].Name
	}
}

func (m *FileListModel) Enter() error {
	if len(m.entries) == 0 {
		return nil
	}
	selectedEntry := m.entries[m.cursor].Name
	newLocation := path.Join(m.location, selectedEntry)
	return m.move(newLocation)
}

func (m *FileListModel) Return() error {
	newLocation := path.Dir(m.location)
	return m.move(newLocation)
}

func (m *FileListModel) Refresh() error {
	return m.move(m.location)
}

func (m *FileListModel) move(newLocation string) error {
	if m.listFn == nil {
		return nil
	}
	newEntries, err := m.listFn(newLocation)
	if err != nil {
		return err
	}
	sort.Slice(newEntries, func(i, j int) bool {
		ei := newEntries[i]
		ej := newEntries[j]
		if ei.Type != ej.Type {
			return ei.Type < ej.Type
		}
		return ei.Name < ej.Name
	})
	m.reset()
	m.entries = newEntries
	m.location = newLocation
	return nil
}

func (m *FileListModel) reset() {
	m.cursor = 0
	m.topItemIndex = 0
	m.DeselectAll()
}

func (m *FileListModel) DeselectAll() {
	for i := range m.selected {
		delete(m.selected, i)
	}
}

func (m FileListModel) GetAllSelectedAbsolute() []string {
	result := make([]string, 0, len(m.selected))
	for i := range m.selected {
		result = append(result, path.Join(m.location, m.entries[i].Name))
	}
	return result
}

func (m FileListModel) Transfer(destination string) error {
	if m.transferFn == nil {
		return fmt.Errorf("transfer %w", ErrNotSet)
	}
	selected := m.getAllSelected()
	if len(selected) == 0 {
		return nil
	}
	return m.transferFn(m.location, selected, destination)
}

func (m *FileListModel) Delete() error {
	if m.deleteFn == nil {
		return fmt.Errorf("delete %w", ErrNotSet)
	}
	if len(m.entries) == 0 {
		return nil
	}
	if err := m.deleteFn(m.location, m.getAllSelected()); err != nil {
		return err
	}
	return m.move(m.location)
}

func (m FileListModel) getAllSelected() []types.Entry {
	selected := make([]types.Entry, 0, len(m.selected))
	for i := range m.selected {
		selected = append(selected, m.entries[i])
	}
	return selected
}

func (m FileListModel) GetSelectedCount() int {
	return len(m.selected)
}

func (m FileListModel) GetLocation() string {
	return m.location
}

func (m FileListModel) View(showCursorSelected bool) string {
	lines := make([]string, 0, m.itemsInVew)
	for i := m.topItemIndex; i < pkg.Min(m.topItemIndex+m.itemsInVew, len(m.entries)); i++ {
		entry := m.entries[i]
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		selected := "[ ]"
		if _, ok := m.selected[i]; ok {
			selected = "[x]"
		}

		sizeView := lipgloss.
			NewStyle().
			Align(lipgloss.Right).
			Width(12).
			Render(pkg.PrettyPrintSize(entry.Size))

		nameView := lipgloss.NewStyle().
			Padding(0, 0, 0, 2).
			Render(entry.Name)

		item := lipgloss.JoinHorizontal(lipgloss.Bottom, entry.TypeString(), sizeView, nameView)
		if showCursorSelected {
			typeView := lipgloss.NewStyle().
				Padding(0, 0, 0, 2).
				Render(entry.TypeString())

			item = lipgloss.JoinHorizontal(lipgloss.Bottom, cursor, selected, typeView, sizeView, nameView)
		}

		lines = append(lines, item)
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		fmt.Sprintf("%s:%s", m.name, m.location),
		lipgloss.NewStyle().
			Height(m.itemsInVew).
			Border(lipgloss.NormalBorder(), true).
			Padding(0, 1).
			Render(lipgloss.JoinVertical(lipgloss.Left, lines...)),
		fmt.Sprintf("[%d-%d]/%d", m.topItemIndex+1, m.topItemIndex+m.itemsInVew, len(m.entries)),
	)
}
