package main

import (
	"fmt"
	"math/rand"

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
			"List Demo - 1 million generated items (virtualized)",
			"Table Demo - 500k user records (on-demand generation)",
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
	s := "VTable Example 02: Large Datasets\n\n"
	s += "VTable's virtualization power - only loads what you see!\n\n"
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

// Large dataset provider - generates data on demand
type LargeListProvider struct {
	totalSize int
}

func NewLargeListProvider(size int) *LargeListProvider {
	return &LargeListProvider{totalSize: size}
}

func (p *LargeListProvider) GetTotal() int {
	return p.totalSize
}

// This is the key method - only generate data that's actually requested
func (p *LargeListProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[string], error) {
	start := request.Start
	count := request.Count

	// Don't generate beyond what exists
	if start >= p.totalSize {
		return []vtable.Data[string]{}, nil
	}

	// Adjust count if it would exceed total
	if start+count > p.totalSize {
		count = p.totalSize - start
	}

	result := make([]vtable.Data[string], count)

	// Generate random data for this chunk
	rand.Seed(int64(start)) // Consistent data for same position

	for i := 0; i < count; i++ {
		actualIndex := start + i
		itemText := fmt.Sprintf("Generated Item #%d (Random: %d)",
			actualIndex,
			rand.Intn(100000))

		result[i] = vtable.Data[string]{
			ID:       fmt.Sprintf("item-%d", actualIndex),
			Item:     itemText,
			Metadata: vtable.NewTypedMetadata(),
		}
	}

	return result, nil
}

// Required methods
func (p *LargeListProvider) GetSelectionMode() vtable.SelectionMode            { return vtable.SelectionNone }
func (p *LargeListProvider) SetSelected(index int, selected bool) bool         { return false }
func (p *LargeListProvider) SetSelectedByIDs(ids []string, selected bool) bool { return false }
func (p *LargeListProvider) SelectRange(startID, endID string) bool            { return false }
func (p *LargeListProvider) SelectAll() bool                                   { return false }
func (p *LargeListProvider) ClearSelection()                                   {}
func (p *LargeListProvider) GetSelectedIndices() []int                         { return nil }
func (p *LargeListProvider) GetSelectedIDs() []string                          { return nil }
func (p *LargeListProvider) GetItemID(item *string) string                     { return *item }

// List model
type ListModel struct {
	list *vtable.TeaList[string]
}

func newListModel() ListModel {
	// 1 million items!
	provider := NewLargeListProvider(1000000)

	formatter := func(data vtable.Data[string], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return fmt.Sprintf("%s[%d] %s", prefix, index, data.Item)
	}

	list, err := vtable.NewTeaListWithHeight(provider, formatter, 10)
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
		case "j", "down":
			m.list.MoveDown()
			return m, nil
		case "k", "up":
			m.list.MoveUp()
			return m, nil
		case "g":
			m.list.JumpToStart()
			return m, nil
		case "G":
			m.list.JumpToEnd()
			return m, nil
		case "d":
			m.list.PageDown()
			return m, nil
		case "u":
			m.list.PageUp()
			return m, nil
		case " ":
			m.list.PageDown()
			return m, nil
		}
	}

	newList, cmd := m.list.Update(msg)
	m.list = newList.(*vtable.TeaList[string])
	return m, cmd
}

func (m ListModel) View() string {
	return "VTable Example 02: Large Datasets - List Demo (1M items)\n\n" +
		m.list.View() +
		"\n\nPress q/ESC to go back, j/k or ↑/↓ to navigate, d/u for page, g/G for start/end, Space for page down"
}

// ===== TABLE DEMO =====

type User struct {
	ID       int
	Username string
	Email    string
	Score    int
	Country  string
}

var countries = []string{
	"USA", "Canada", "UK", "France", "Germany", "Japan", "Australia",
	"Brazil", "Mexico", "India", "China", "Russia", "Italy", "Spain",
}

// Large table data provider
type LargeTableProvider struct {
	totalSize int
}

func NewLargeTableProvider(size int) *LargeTableProvider {
	return &LargeTableProvider{totalSize: size}
}

func (p *LargeTableProvider) GetTotal() int {
	return p.totalSize
}

func (p *LargeTableProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[vtable.TableRow], error) {
	start := request.Start
	count := request.Count

	if start >= p.totalSize {
		return []vtable.Data[vtable.TableRow]{}, nil
	}

	if start+count > p.totalSize {
		count = p.totalSize - start
	}

	result := make([]vtable.Data[vtable.TableRow], count)

	// Use deterministic randomization
	rand.Seed(int64(start))

	for i := 0; i < count; i++ {
		actualIndex := start + i
		user := User{
			ID:       actualIndex + 1,
			Username: fmt.Sprintf("user_%d", actualIndex+1),
			Email:    fmt.Sprintf("user_%d@example.com", actualIndex+1),
			Score:    rand.Intn(10000),
			Country:  countries[rand.Intn(len(countries))],
		}

		row := vtable.TableRow{
			Cells: []string{
				fmt.Sprintf("%d", user.ID),
				user.Username,
				user.Email,
				fmt.Sprintf("%d", user.Score),
				user.Country,
			},
		}

		result[i] = vtable.Data[vtable.TableRow]{
			ID:       fmt.Sprintf("user-%d", user.ID),
			Item:     row,
			Metadata: vtable.NewTypedMetadata(),
		}
	}

	return result, nil
}

// Required methods
func (p *LargeTableProvider) GetSelectionMode() vtable.SelectionMode            { return vtable.SelectionNone }
func (p *LargeTableProvider) SetSelected(index int, selected bool) bool         { return false }
func (p *LargeTableProvider) SetSelectedByIDs(ids []string, selected bool) bool { return false }
func (p *LargeTableProvider) SelectRange(startID, endID string) bool            { return false }
func (p *LargeTableProvider) SelectAll() bool                                   { return false }
func (p *LargeTableProvider) ClearSelection()                                   {}
func (p *LargeTableProvider) GetSelectedIndices() []int                         { return nil }
func (p *LargeTableProvider) GetSelectedIDs() []string                          { return nil }
func (p *LargeTableProvider) GetItemID(item *vtable.TableRow) string            { return item.Cells[0] }

// Table model
type TableModel struct {
	table *vtable.TeaTable
}

func newTableModel() TableModel {
	// 500k user records!
	provider := NewLargeTableProvider(500000)

	columns := []vtable.TableColumn{
		vtable.NewRightColumn("ID", 8),
		vtable.NewColumn("Username", 15),
		vtable.NewColumn("Email", 25),
		vtable.NewRightColumn("Score", 8),
		vtable.NewColumn("Country", 12),
	}

	table, err := vtable.NewTeaTableWithHeight(columns, provider, 12)
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
		case "j", "down":
			m.table.MoveDown()
			return m, nil
		case "k", "up":
			m.table.MoveUp()
			return m, nil
		case "g":
			m.table.JumpToStart()
			return m, nil
		case "G":
			m.table.JumpToEnd()
			return m, nil
		case "d":
			m.table.PageDown()
			return m, nil
		case "u":
			m.table.PageUp()
			return m, nil
		case " ":
			m.table.PageDown()
			return m, nil
		}
	}

	newTable, cmd := m.table.Update(msg)
	m.table = newTable.(*vtable.TeaTable)
	return m, cmd
}

func (m TableModel) View() string {
	return "VTable Example 02: Large Datasets - Table Demo (500K users)\n\n" +
		m.table.View() +
		"\n\nPress q/ESC to go back, j/k or ↑/↓ to navigate, d/u for page, g/G for start/end, Space for page down"
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
