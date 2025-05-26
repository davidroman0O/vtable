package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman0O/vtable"
)

// Application states
type AppState int

const (
	StateMenu AppState = iota
	StateList
	StateTable
)

// Custom message to go back to menu
type BackToMenuMsg struct{}

// Main application model that manages different states
type AppModel struct {
	state      AppState
	menuModel  MenuModel
	listModel  ListModel
	tableModel TableModel
}

func newAppModel() AppModel {
	return AppModel{
		state:     StateMenu,
		menuModel: newMenuModel(),
	}
}

func (m AppModel) Init() tea.Cmd {
	switch m.state {
	case StateMenu:
		return m.menuModel.Init()
	case StateList:
		return m.listModel.Init()
	case StateTable:
		return m.tableModel.Init()
	default:
		return nil
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case BackToMenuMsg:
		// Go back to menu
		m.state = StateMenu
		m.menuModel = newMenuModel()
		return m, m.menuModel.Init()
	default:
		// Handle other messages based on current state
		switch m.state {
		case StateMenu:
			newMenuModel, cmd := m.menuModel.Update(msg)
			m.menuModel = newMenuModel.(MenuModel)

			// Check if user selected something
			if m.menuModel.selected != -1 {
				switch m.menuModel.selected {
				case 0: // List Demo
					m.state = StateList
					m.listModel = newListModel()
					return m, m.listModel.Init()
				case 1: // Table Demo
					m.state = StateTable
					m.tableModel = newTableModel()
					return m, m.tableModel.Init()
				}
			}
			return m, cmd

		case StateList:
			newListModel, cmd := m.listModel.Update(msg)
			m.listModel = newListModel.(ListModel)
			return m, cmd

		case StateTable:
			newTableModel, cmd := m.tableModel.Update(msg)
			m.tableModel = newTableModel.(TableModel)
			return m, cmd
		}
	}

	return m, nil
}

func (m AppModel) View() string {
	switch m.state {
	case StateMenu:
		return m.menuModel.View()
	case StateList:
		return m.listModel.View()
	case StateTable:
		return m.tableModel.View()
	default:
		return "Unknown state"
	}
}

// Menu model for choosing between list and table
type MenuModel struct {
	choices  []string
	cursor   int
	selected int
}

func newMenuModel() MenuModel {
	return MenuModel{
		choices: []string{
			"List Demo - Simple string list",
			"Table Demo - Basic table with Person data",
		},
		cursor:   0,
		selected: -1,
	}
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = m.cursor
		}
	}
	return m, nil
}

func (m MenuModel) View() string {
	s := "VTable Example 01: Hello World\n\n"
	s += "Choose a demo to run:\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\nPress j/k or ↑/↓ to navigate, Enter to select, q to quit.\n"
	return s
}

// ===== LIST DEMO =====

// String data provider for list
type StringProvider struct {
	items []string
}

func (p *StringProvider) GetTotal() int {
	return len(p.items)
}

func (p *StringProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[string], error) {
	result := make([]vtable.Data[string], len(p.items))
	for i, item := range p.items {
		result[i] = vtable.Data[string]{
			ID:       fmt.Sprintf("%d", i),
			Item:     item,
			Metadata: vtable.NewTypedMetadata(),
		}
	}
	return result, nil
}

// Required methods
func (p *StringProvider) GetSelectionMode() vtable.SelectionMode            { return vtable.SelectionNone }
func (p *StringProvider) SetSelected(index int, selected bool) bool         { return false }
func (p *StringProvider) SetSelectedByIDs(ids []string, selected bool) bool { return false }
func (p *StringProvider) SelectRange(startID, endID string) bool            { return false }
func (p *StringProvider) SelectAll() bool                                   { return false }
func (p *StringProvider) ClearSelection()                                   {}
func (p *StringProvider) GetSelectedIndices() []int                         { return nil }
func (p *StringProvider) GetSelectedIDs() []string                          { return nil }
func (p *StringProvider) GetItemID(item *string) string                     { return *item }

// List model
type ListModel struct {
	list *vtable.TeaList[string]
}

func newListModel() ListModel {
	provider := &StringProvider{
		items: []string{"Apple", "Banana", "Cherry"},
	}

	formatter := func(data vtable.Data[string], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return fmt.Sprintf("%s%d: %s", prefix, index, data.Item)
	}

	list, err := vtable.NewTeaListWithHeight(provider, formatter, 3)
	if err != nil {
		panic(err)
	}

	return ListModel{list: list}
}

func (m ListModel) Init() tea.Cmd {
	return m.list.Init()
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}

	newList, cmd := m.list.Update(msg)
	m.list = newList.(*vtable.TeaList[string])
	return m, cmd
}

func (m ListModel) View() string {
	return "VTable Example 01: Hello World - List Demo\n\n" +
		m.list.View() +
		"\n\nPress q or ESC to go back to menu, ↑/↓ to navigate"
}

// ===== TABLE DEMO =====

type Person struct {
	Name string
	Age  int
	City string
}

// Table data provider
type PeopleProvider struct {
	people []Person
}

func (p *PeopleProvider) GetTotal() int {
	return len(p.people)
}

func (p *PeopleProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[vtable.TableRow], error) {
	result := make([]vtable.Data[vtable.TableRow], len(p.people))

	for i, person := range p.people {
		row := vtable.TableRow{
			Cells: []string{
				person.Name,
				fmt.Sprintf("%d", person.Age),
				person.City,
			},
		}

		result[i] = vtable.Data[vtable.TableRow]{
			ID:       fmt.Sprintf("person-%d", i),
			Item:     row,
			Metadata: vtable.NewTypedMetadata(),
		}
	}
	return result, nil
}

// Required methods
func (p *PeopleProvider) GetSelectionMode() vtable.SelectionMode            { return vtable.SelectionNone }
func (p *PeopleProvider) SetSelected(index int, selected bool) bool         { return false }
func (p *PeopleProvider) SetSelectedByIDs(ids []string, selected bool) bool { return false }
func (p *PeopleProvider) SelectRange(startID, endID string) bool            { return false }
func (p *PeopleProvider) SelectAll() bool                                   { return false }
func (p *PeopleProvider) ClearSelection()                                   {}
func (p *PeopleProvider) GetSelectedIndices() []int                         { return nil }
func (p *PeopleProvider) GetSelectedIDs() []string                          { return nil }
func (p *PeopleProvider) GetItemID(item *vtable.TableRow) string            { return item.Cells[0] }

// Table model
type TableModel struct {
	table *vtable.TeaTable
}

func newTableModel() TableModel {
	provider := &PeopleProvider{
		people: []Person{
			{"Alice", 28, "New York"},
			{"Bob", 34, "San Francisco"},
			{"Carol", 22, "Chicago"},
		},
	}

	columns := []vtable.TableColumn{
		vtable.NewColumn("Name", 15),
		vtable.NewRightColumn("Age", 5),
		vtable.NewColumn("City", 15),
	}

	table, err := vtable.NewTeaTableWithHeight(columns, provider, 4)
	if err != nil {
		panic(err)
	}

	return TableModel{table: table}
}

func (m TableModel) Init() tea.Cmd {
	return m.table.Init()
}

func (m TableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }
		}
	}

	newTable, cmd := m.table.Update(msg)
	m.table = newTable.(*vtable.TeaTable)
	return m, cmd
}

func (m TableModel) View() string {
	return "VTable Example 01: Hello World - Table Demo\n\n" +
		m.table.View() +
		"\n\nPress q or ESC to go back to menu, ↑/↓ to navigate"
}

// ===== MAIN =====

func main() {
	app := newAppModel()
	p := tea.NewProgram(app)

	if _, err := p.Run(); err != nil {
		panic(err)
	}

	// Clean exit
	fmt.Print("\033[H\033[2J")
	fmt.Print("\033[?25h")
	fmt.Print("\n\n")
}
