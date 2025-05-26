package main

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman0O/vtable"
)

// Application states
type AppState int

const (
	StateMenu AppState = iota
	StateListDemo
	StateTableDemo
)

// Input modes
type InputMode int

const (
	ModeNormal InputMode = iota
	ModeJump
)

// Custom message to go back to menu
type BackToMenuMsg struct{}

// Main application model that manages different states
type AppModel struct {
	state      AppState
	menuModel  *MenuModel
	listModel  *ListJumpModel
	tableModel *TableJumpModel
}

func newAppModel() *AppModel {
	return &AppModel{
		state:     StateMenu,
		menuModel: newMenuModel(),
	}
}

func (m *AppModel) Init() tea.Cmd {
	return m.menuModel.Init()
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case BackToMenuMsg:
		// Go back to menu
		m.state = StateMenu
		m.menuModel = newMenuModel()
		return m, m.menuModel.Init()
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.state {
	case StateMenu:
		newMenuModel, cmd := m.menuModel.Update(msg)
		m.menuModel = newMenuModel.(*MenuModel)

		// Check if user selected something
		if m.menuModel.selected != -1 {
			switch m.menuModel.selected {
			case 0: // List Jump Demo
				m.state = StateListDemo
				m.listModel = newListJumpModel()
				return m, m.listModel.Init()
			case 1: // Table Jump Demo
				m.state = StateTableDemo
				m.tableModel = newTableJumpModel()
				return m, m.tableModel.Init()
			}
		}
		return m, cmd

	case StateListDemo:
		newListModel, cmd := m.listModel.Update(msg)
		m.listModel = newListModel.(*ListJumpModel)
		return m, cmd

	case StateTableDemo:
		newTableModel, cmd := m.tableModel.Update(msg)
		m.tableModel = newTableModel.(*TableJumpModel)
		return m, cmd
	}

	return m, nil
}

func (m *AppModel) View() string {
	switch m.state {
	case StateMenu:
		return m.menuModel.View()
	case StateListDemo:
		return m.listModel.View()
	case StateTableDemo:
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

func newMenuModel() *MenuModel {
	return &MenuModel{
		choices: []string{
			"List Demo - Jump through 500 items instantly",
			"Table Demo - Jump through 1000 records quickly",
		},
		cursor:   0,
		selected: -1,
	}
}

func (m *MenuModel) Init() tea.Cmd {
	return nil
}

func (m *MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *MenuModel) View() string {
	s := "VTable Example 07: Search & Jump\n\n"
	s += "Jump navigation through hundreds of items - instant access!\n\n"
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

type Item struct {
	ID          int
	Name        string
	Description string
	Value       int
}

// Item provider for list demo
type ItemProvider struct {
	items     []Item
	selection map[int]bool
}

func NewItemProvider() *ItemProvider {
	// Generate 500 items
	items := make([]Item, 500)
	categories := []string{"Widget", "Gadget", "Tool", "Device", "Component", "Module", "System", "Interface"}
	adjectives := []string{"Smart", "Advanced", "Digital", "Premium", "Standard", "Custom", "Universal", "Portable"}

	for i := 0; i < 500; i++ {
		category := categories[i%len(categories)]
		adjective := adjectives[(i/len(categories))%len(adjectives)]

		items[i] = Item{
			ID:          i + 1,
			Name:        fmt.Sprintf("%s %s %03d", adjective, category, i+1),
			Description: fmt.Sprintf("This is item number %d with special features", i+1),
			Value:       (i * 7 % 1000) + 100, // Some pseudo-random values
		}
	}

	return &ItemProvider{
		items:     items,
		selection: make(map[int]bool),
	}
}

func (p *ItemProvider) GetTotal() int {
	return len(p.items)
}

func (p *ItemProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[string], error) {
	start := request.Start
	count := request.Count

	if start >= len(p.items) {
		return []vtable.Data[string]{}, nil
	}

	if start+count > len(p.items) {
		count = len(p.items) - start
	}

	result := make([]vtable.Data[string], count)
	for i := 0; i < count; i++ {
		item := p.items[start+i]
		display := fmt.Sprintf("#%03d: %-25s | %s | Value: $%d",
			item.ID, item.Name, item.Description, item.Value)

		result[i] = vtable.Data[string]{
			ID:       fmt.Sprintf("item-%d", item.ID),
			Item:     display,
			Selected: p.selection[start+i],
			Metadata: vtable.NewTypedMetadata(),
		}
	}
	return result, nil
}

// Selection methods for DataProvider interface
func (p *ItemProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionSingle
}

func (p *ItemProvider) SetSelected(index int, selected bool) bool {
	if index < 0 || index >= len(p.items) {
		return false
	}
	if selected {
		p.selection[index] = true
	} else {
		delete(p.selection, index)
	}
	return true
}

func (p *ItemProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	return true
}

func (p *ItemProvider) SelectRange(startID, endID string) bool {
	return true
}

func (p *ItemProvider) SelectAll() bool {
	for i := 0; i < len(p.items); i++ {
		p.selection[i] = true
	}
	return true
}

func (p *ItemProvider) ClearSelection() {
	p.selection = make(map[int]bool)
}

func (p *ItemProvider) GetSelectedIndices() []int {
	indices := make([]int, 0, len(p.selection))
	for idx := range p.selection {
		indices = append(indices, idx)
	}
	return indices
}

func (p *ItemProvider) GetSelectedIDs() []string {
	return []string{}
}

func (p *ItemProvider) GetItemID(item *string) string {
	return ""
}

// List model with jump functionality
type ListJumpModel struct {
	list      *vtable.TeaList[string]
	provider  *ItemProvider
	mode      InputMode
	jumpInput string
	status    string
}

func newListJumpModel() *ListJumpModel {
	provider := NewItemProvider()

	formatter := func(data vtable.Data[string], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return fmt.Sprintf("%s%s", prefix, data.Item)
	}

	list, err := vtable.NewTeaListWithHeight(provider, formatter, 12)
	if err != nil {
		panic(err)
	}

	model := &ListJumpModel{
		list:      list,
		provider:  provider,
		mode:      ModeNormal,
		jumpInput: "",
		status:    fmt.Sprintf("500 items loaded. Press : to jump to any item (1-500), g/G for first/last"),
	}

	return model
}

func (m *ListJumpModel) Init() tea.Cmd {
	return m.list.Init()
}

func (m *ListJumpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle mode-specific keys first
		switch m.mode {
		case ModeJump:
			return m.handleJumpMode(msg)
		case ModeNormal:
			return m.handleNormalMode(msg)
		}
	}

	// Update the list
	newList, cmd := m.list.Update(msg)
	m.list = newList.(*vtable.TeaList[string])
	return m, cmd
}

func (m *ListJumpModel) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m, func() tea.Msg { return BackToMenuMsg{} }

	case ":":
		m.mode = ModeJump
		m.jumpInput = ""
		m.status = "Jump to item (1-500): "
		return m, nil

	case "g":
		// Jump to first
		m.list.JumpToStart()
		m.status = "Jumped to first item (1/500)"
		return m, nil

	case "G":
		// Jump to last
		m.list.JumpToEnd()
		m.status = "Jumped to last item (500/500)"
		return m, nil

	case "ctrl+d":
		// Jump down 25 items
		state := m.list.GetState()
		newIndex := state.CursorIndex + 25
		if newIndex >= m.provider.GetTotal() {
			newIndex = m.provider.GetTotal() - 1
		}
		m.list.JumpToIndex(newIndex)
		m.status = fmt.Sprintf("Jumped down to item %d/500", newIndex+1)
		return m, nil

	case "ctrl+u":
		// Jump up 25 items
		state := m.list.GetState()
		newIndex := state.CursorIndex - 25
		if newIndex < 0 {
			newIndex = 0
		}
		m.list.JumpToIndex(newIndex)
		m.status = fmt.Sprintf("Jumped up to item %d/500", newIndex+1)
		return m, nil

	case "ctrl+f":
		// Jump down 50 items
		state := m.list.GetState()
		newIndex := state.CursorIndex + 50
		if newIndex >= m.provider.GetTotal() {
			newIndex = m.provider.GetTotal() - 1
		}
		m.list.JumpToIndex(newIndex)
		m.status = fmt.Sprintf("Jumped page down to item %d/500", newIndex+1)
		return m, nil

	case "ctrl+b":
		// Jump up 50 items
		state := m.list.GetState()
		newIndex := state.CursorIndex - 50
		if newIndex < 0 {
			newIndex = 0
		}
		m.list.JumpToIndex(newIndex)
		m.status = fmt.Sprintf("Jumped page up to item %d/500", newIndex+1)
		return m, nil
	}

	// Pass to list for normal navigation
	newList, cmd := m.list.Update(msg)
	m.list = newList.(*vtable.TeaList[string])

	// Update status with current position
	state := m.list.GetState()
	m.status = fmt.Sprintf("Position: %d/500 - Press : to jump to any item", state.CursorIndex+1)

	return m, cmd
}

func (m *ListJumpModel) handleJumpMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.mode = ModeNormal
		if m.jumpInput == "" {
			m.status = "Jump cancelled - no input"
		} else {
			if index, err := strconv.Atoi(m.jumpInput); err == nil {
				// Convert to 0-based index
				index--
				if index >= 0 && index < m.provider.GetTotal() {
					m.list.JumpToIndex(index)
					m.status = fmt.Sprintf("Jumped to item %d/500", index+1)
				} else {
					m.status = fmt.Sprintf("Invalid item number: %d (valid range: 1-500)", index+1)
				}
			} else {
				m.status = fmt.Sprintf("Invalid number: %s", m.jumpInput)
			}
		}
		return m, nil

	case "escape":
		m.mode = ModeNormal
		state := m.list.GetState()
		m.status = fmt.Sprintf("Position: %d/500 - Jump cancelled", state.CursorIndex+1)
		return m, nil

	case "backspace":
		if len(m.jumpInput) > 0 {
			m.jumpInput = m.jumpInput[:len(m.jumpInput)-1]
		}
		return m, nil

	default:
		// Add digit to jump input
		if len(msg.String()) == 1 && msg.String() >= "0" && msg.String() <= "9" {
			m.jumpInput += msg.String()
		}
		return m, nil
	}
}

func (m *ListJumpModel) View() string {
	var sb strings.Builder

	sb.WriteString("VTable Example 07: Jump Navigation - List Demo\n\n")

	// List
	sb.WriteString(m.list.View())
	sb.WriteString("\n\n")

	// Status with current input
	switch m.mode {
	case ModeJump:
		sb.WriteString(fmt.Sprintf("Jump to item (1-500): %s", m.jumpInput))
	default:
		sb.WriteString(m.status)
	}
	sb.WriteString("\n\n")

	// Help based on mode
	switch m.mode {
	case ModeJump:
		sb.WriteString("Type item number (1-500), Enter to jump, Esc to cancel")
	default:
		sb.WriteString("Navigation: :=jump g=first G=last ctrl+d/u=±25 ctrl+f/b=±50 j/k=±1 q=quit")
	}

	return sb.String()
}

// ===== TABLE DEMO =====

type Record struct {
	ID          int
	Code        string
	Name        string
	Category    string
	Status      string
	Priority    int
	Value       float64
	LastUpdated string
}

// Record provider for table demo
type RecordProvider struct {
	records   []Record
	selection map[int]bool
}

func NewRecordProvider() *RecordProvider {
	// Generate 1000 records
	records := make([]Record, 1000)
	categories := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta", "Theta"}
	statuses := []string{"Active", "Pending", "Complete", "Review", "Archived"}

	for i := 0; i < 1000; i++ {
		category := categories[i%len(categories)]
		status := statuses[i%len(statuses)]

		records[i] = Record{
			ID:          i + 1,
			Code:        fmt.Sprintf("REC-%04d", i+1),
			Name:        fmt.Sprintf("Record Item %04d", i+1),
			Category:    category,
			Status:      status,
			Priority:    (i % 5) + 1,
			Value:       float64((i*13%10000)+100) / 100.0,
			LastUpdated: fmt.Sprintf("2024-01-%02d", (i%28)+1),
		}
	}

	return &RecordProvider{
		records:   records,
		selection: make(map[int]bool),
	}
}

func (p *RecordProvider) GetTotal() int {
	return len(p.records)
}

func (p *RecordProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[vtable.TableRow], error) {
	start := request.Start
	count := request.Count

	if start >= len(p.records) {
		return []vtable.Data[vtable.TableRow]{}, nil
	}

	if start+count > len(p.records) {
		count = len(p.records) - start
	}

	result := make([]vtable.Data[vtable.TableRow], count)
	for i := 0; i < count; i++ {
		record := p.records[start+i]

		cells := []string{
			fmt.Sprintf("%d", record.ID),
			record.Code,
			record.Name,
			record.Category,
			record.Status,
			fmt.Sprintf("%d", record.Priority),
			fmt.Sprintf("$%.2f", record.Value),
			record.LastUpdated,
		}

		row := vtable.TableRow{
			Cells: cells,
		}

		result[i] = vtable.Data[vtable.TableRow]{
			ID:       fmt.Sprintf("record-%d", record.ID),
			Item:     row,
			Selected: p.selection[start+i],
			Metadata: vtable.NewTypedMetadata(),
		}
	}
	return result, nil
}

// Selection methods for DataProvider interface
func (p *RecordProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionSingle
}

func (p *RecordProvider) SetSelected(index int, selected bool) bool {
	if index < 0 || index >= len(p.records) {
		return false
	}
	if selected {
		p.selection[index] = true
	} else {
		delete(p.selection, index)
	}
	return true
}

func (p *RecordProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	return true
}

func (p *RecordProvider) SelectRange(startID, endID string) bool {
	return true
}

func (p *RecordProvider) SelectAll() bool {
	for i := 0; i < len(p.records); i++ {
		p.selection[i] = true
	}
	return true
}

func (p *RecordProvider) ClearSelection() {
	p.selection = make(map[int]bool)
}

func (p *RecordProvider) GetSelectedIndices() []int {
	indices := make([]int, 0, len(p.selection))
	for idx := range p.selection {
		indices = append(indices, idx)
	}
	return indices
}

func (p *RecordProvider) GetSelectedIDs() []string {
	return []string{}
}

func (p *RecordProvider) GetItemID(item *vtable.TableRow) string {
	return ""
}

// Table model with jump functionality
type TableJumpModel struct {
	table     *vtable.TeaTable
	provider  *RecordProvider
	mode      InputMode
	jumpInput string
	status    string
}

func newTableJumpModel() *TableJumpModel {
	provider := NewRecordProvider()

	columns := []vtable.TableColumn{
		vtable.NewColumn("ID", 8),
		vtable.NewColumn("Code", 12),
		vtable.NewColumn("Name", 20),
		vtable.NewColumn("Category", 15),
		vtable.NewColumn("Status", 10),
		vtable.NewRightColumn("Priority", 8),
		vtable.NewRightColumn("Value", 12),
		vtable.NewColumn("Updated", 12),
	}

	table, err := vtable.NewTeaTableWithHeight(columns, provider, 12)
	if err != nil {
		panic(err)
	}

	model := &TableJumpModel{
		table:     table,
		provider:  provider,
		mode:      ModeNormal,
		jumpInput: "",
		status:    fmt.Sprintf("1000 records loaded. Press : to jump to any record (1-1000), g/G for first/last"),
	}

	return model
}

func (m *TableJumpModel) Init() tea.Cmd {
	return m.table.Init()
}

func (m *TableJumpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle mode-specific keys first
		switch m.mode {
		case ModeJump:
			return m.handleJumpMode(msg)
		case ModeNormal:
			return m.handleNormalMode(msg)
		}
	}

	// Update the table
	newTable, cmd := m.table.Update(msg)
	m.table = newTable.(*vtable.TeaTable)
	return m, cmd
}

func (m *TableJumpModel) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m, func() tea.Msg { return BackToMenuMsg{} }

	case ":":
		m.mode = ModeJump
		m.jumpInput = ""
		m.status = "Jump to record (1-1000): "
		return m, nil

	case "g":
		// Jump to first
		m.table.JumpToStart()
		m.status = "Jumped to first record (1/1000)"
		return m, nil

	case "G":
		// Jump to last
		m.table.JumpToEnd()
		m.status = "Jumped to last record (1000/1000)"
		return m, nil

	case "ctrl+d":
		// Jump down 50 records
		state := m.table.GetState()
		newIndex := state.CursorIndex + 50
		if newIndex >= m.provider.GetTotal() {
			newIndex = m.provider.GetTotal() - 1
		}
		m.table.JumpToIndex(newIndex)
		m.status = fmt.Sprintf("Jumped down to record %d/1000", newIndex+1)
		return m, nil

	case "ctrl+u":
		// Jump up 50 records
		state := m.table.GetState()
		newIndex := state.CursorIndex - 50
		if newIndex < 0 {
			newIndex = 0
		}
		m.table.JumpToIndex(newIndex)
		m.status = fmt.Sprintf("Jumped up to record %d/1000", newIndex+1)
		return m, nil

	case "ctrl+f":
		// Jump down 100 records
		state := m.table.GetState()
		newIndex := state.CursorIndex + 100
		if newIndex >= m.provider.GetTotal() {
			newIndex = m.provider.GetTotal() - 1
		}
		m.table.JumpToIndex(newIndex)
		m.status = fmt.Sprintf("Jumped page down to record %d/1000", newIndex+1)
		return m, nil

	case "ctrl+b":
		// Jump up 100 records
		state := m.table.GetState()
		newIndex := state.CursorIndex - 100
		if newIndex < 0 {
			newIndex = 0
		}
		m.table.JumpToIndex(newIndex)
		m.status = fmt.Sprintf("Jumped page up to record %d/1000", newIndex+1)
		return m, nil
	}

	// Pass to table for normal navigation
	newTable, cmd := m.table.Update(msg)
	m.table = newTable.(*vtable.TeaTable)

	// Update status with current position
	state := m.table.GetState()
	m.status = fmt.Sprintf("Position: %d/1000 - Press : to jump to any record", state.CursorIndex+1)

	return m, cmd
}

func (m *TableJumpModel) handleJumpMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.mode = ModeNormal
		if m.jumpInput == "" {
			m.status = "Jump cancelled - no input"
		} else {
			if index, err := strconv.Atoi(m.jumpInput); err == nil {
				// Convert to 0-based index
				index--
				if index >= 0 && index < m.provider.GetTotal() {
					m.table.JumpToIndex(index)
					m.status = fmt.Sprintf("Jumped to record %d/1000", index+1)
				} else {
					m.status = fmt.Sprintf("Invalid record number: %d (valid range: 1-1000)", index+1)
				}
			} else {
				m.status = fmt.Sprintf("Invalid number: %s", m.jumpInput)
			}
		}
		return m, nil

	case "escape":
		m.mode = ModeNormal
		state := m.table.GetState()
		m.status = fmt.Sprintf("Position: %d/1000 - Jump cancelled", state.CursorIndex+1)
		return m, nil

	case "backspace":
		if len(m.jumpInput) > 0 {
			m.jumpInput = m.jumpInput[:len(m.jumpInput)-1]
		}
		return m, nil

	default:
		// Add digit to jump input
		if len(msg.String()) == 1 && msg.String() >= "0" && msg.String() <= "9" {
			m.jumpInput += msg.String()
		}
		return m, nil
	}
}

func (m *TableJumpModel) View() string {
	var sb strings.Builder

	sb.WriteString("VTable Example 07: Jump Navigation - Table Demo\n\n")

	// Table
	sb.WriteString(m.table.View())
	sb.WriteString("\n\n")

	// Status with current input
	switch m.mode {
	case ModeJump:
		sb.WriteString(fmt.Sprintf("Jump to record (1-1000): %s", m.jumpInput))
	default:
		sb.WriteString(m.status)
	}
	sb.WriteString("\n\n")

	// Help based on mode
	switch m.mode {
	case ModeJump:
		sb.WriteString("Type record number (1-1000), Enter to jump, Esc to cancel")
	default:
		sb.WriteString("Navigation: :=jump g=first G=last ctrl+d/u=±50 ctrl+f/b=±100 j/k=±1 q=quit")
	}

	return sb.String()
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
