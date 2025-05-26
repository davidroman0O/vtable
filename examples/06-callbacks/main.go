package main

import (
	"fmt"
	"strings"
	"time"

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

// Custom message to go back to menu
type BackToMenuMsg struct{}

// Main application model that manages different states
type AppModel struct {
	state      AppState
	menuModel  *MenuModel
	listModel  *ListCallbackModel
	tableModel *TableCallbackModel
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
			case 0: // List Callback Demo
				m.state = StateListDemo
				m.listModel = newListCallbackModel()
				return m, m.listModel.Init()
			case 1: // Table Callback Demo
				m.state = StateTableDemo
				m.tableModel = newTableCallbackModel()
				return m, m.tableModel.Init()
			}
		}
		return m, cmd

	case StateListDemo:
		newListModel, cmd := m.listModel.Update(msg)
		m.listModel = newListModel.(*ListCallbackModel)
		return m, cmd

	case StateTableDemo:
		newTableModel, cmd := m.tableModel.Update(msg)
		m.tableModel = newTableModel.(*TableCallbackModel)
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
			"List Demo - Selection callbacks, item actions, navigation events",
			"Table Demo - Row callbacks, column events, selection handlers",
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
	s := "VTable Example 06: Callbacks\n\n"
	s += "Learn callback patterns - selection events, item actions, navigation handlers!\n\n"
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

type Task struct {
	ID          int
	Title       string
	Description string
	Status      string
	Priority    string
	Created     time.Time
}

// Task provider for list demo
type TaskProvider struct {
	tasks []Task
}

func NewTaskProvider() *TaskProvider {
	return &TaskProvider{
		tasks: []Task{
			{1, "Setup CI/CD", "Configure GitHub Actions workflow", "Todo", "High", time.Now().AddDate(0, 0, -2)},
			{2, "Write Tests", "Unit tests for core functionality", "In Progress", "Medium", time.Now().AddDate(0, 0, -1)},
			{3, "Code Review", "Review pull request #123", "Todo", "High", time.Now().AddDate(0, 0, -3)},
			{4, "Deploy Staging", "Deploy to staging environment", "Done", "Medium", time.Now().AddDate(0, 0, -5)},
			{5, "Update Docs", "Update API documentation", "Todo", "Low", time.Now().AddDate(0, 0, -1)},
			{6, "Fix Bug #456", "Login form validation error", "In Progress", "High", time.Now().AddDate(0, 0, -2)},
			{7, "Performance Test", "Load testing for new features", "Todo", "Medium", time.Now().AddDate(0, 0, -4)},
			{8, "Security Audit", "Review authentication flow", "Todo", "High", time.Now().AddDate(0, 0, -1)},
			{9, "Database Migration", "Update schema for v2.0", "In Progress", "High", time.Now().AddDate(0, 0, -3)},
			{10, "Client Meeting", "Discuss project requirements", "Done", "Medium", time.Now().AddDate(0, 0, -7)},
			{11, "Refactor API", "Clean up legacy endpoints", "Todo", "Medium", time.Now().AddDate(0, 0, -2)},
			{12, "Monitor Metrics", "Check application performance", "In Progress", "Low", time.Now().AddDate(0, 0, -1)},
			{13, "Update Dependencies", "Upgrade to latest versions", "Todo", "Low", time.Now().AddDate(0, 0, -5)},
			{14, "Backup Data", "Weekly data backup procedure", "Done", "Medium", time.Now().AddDate(0, 0, -6)},
			{15, "User Training", "Train new team members", "In Progress", "Medium", time.Now().AddDate(0, 0, -3)},
		},
	}
}

func (p *TaskProvider) GetTotal() int {
	return len(p.tasks)
}

func (p *TaskProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[string], error) {
	start := request.Start
	count := request.Count

	if start >= len(p.tasks) {
		return []vtable.Data[string]{}, nil
	}

	if start+count > len(p.tasks) {
		count = len(p.tasks) - start
	}

	result := make([]vtable.Data[string], count)
	for i := 0; i < count; i++ {
		task := p.tasks[start+i]
		display := fmt.Sprintf("[%s] %s - %s (%s)", task.Status, task.Title, task.Priority, task.Description)

		result[i] = vtable.Data[string]{
			ID:       fmt.Sprintf("task-%d", task.ID),
			Item:     display,
			Selected: false,
			Metadata: vtable.NewTypedMetadata(),
		}
	}
	return result, nil
}

// Selection methods for DataProvider interface
func (p *TaskProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionMultiple
}

func (p *TaskProvider) SetSelected(index int, selected bool) bool {
	return true
}

func (p *TaskProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	return true
}

func (p *TaskProvider) SelectRange(startID, endID string) bool {
	return true
}

func (p *TaskProvider) SelectAll() bool {
	return true
}

func (p *TaskProvider) ClearSelection() {
}

func (p *TaskProvider) GetSelectedIndices() []int {
	return []int{}
}

func (p *TaskProvider) GetSelectedIDs() []string {
	return []string{}
}

func (p *TaskProvider) GetItemID(item *string) string {
	return ""
}

// List model with callbacks
type ListCallbackModel struct {
	list        *vtable.TeaList[string]
	provider    *TaskProvider
	status      string
	eventLog    []string
	maxLogItems int
}

func newListCallbackModel() *ListCallbackModel {
	provider := NewTaskProvider()

	formatter := func(data vtable.Data[string], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return fmt.Sprintf("%s%s", prefix, data.Item)
	}

	list, err := vtable.NewTeaListWithHeight(provider, formatter, 10)
	if err != nil {
		panic(err)
	}

	model := &ListCallbackModel{
		list:        list,
		provider:    provider,
		status:      "List with callbacks - navigate and select items to see events",
		eventLog:    []string{},
		maxLogItems: 8,
	}

	return model
}

func (m *ListCallbackModel) Init() tea.Cmd {
	return m.list.Init()
}

func (m *ListCallbackModel) addEvent(event string) {
	timestamp := time.Now().Format("15:04:05")
	m.eventLog = append(m.eventLog, fmt.Sprintf("[%s] %s", timestamp, event))

	// Keep only the last maxLogItems events
	if len(m.eventLog) > m.maxLogItems {
		m.eventLog = m.eventLog[len(m.eventLog)-m.maxLogItems:]
	}
}

func (m *ListCallbackModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }

		case "enter":
			state := m.list.GetState()
			m.addEvent(fmt.Sprintf("Item activated: index %d", state.CursorIndex))
			m.status = fmt.Sprintf("Activated item at index %d", state.CursorIndex)
			return m, nil

		case " ":
			oldCount := m.list.GetSelectionCount()
			if m.list.ToggleCurrentSelection() {
				state := m.list.GetState()
				newCount := m.list.GetSelectionCount()
				if newCount > oldCount {
					m.addEvent(fmt.Sprintf("Item selected: index %d", state.CursorIndex))
					m.status = fmt.Sprintf("Selected item %d (total: %d)", state.CursorIndex, newCount)
				} else {
					m.addEvent(fmt.Sprintf("Item deselected: index %d", state.CursorIndex))
					m.status = fmt.Sprintf("Deselected item %d (total: %d)", state.CursorIndex, newCount)
				}
			}
			return m, nil

		case "ctrl+a":
			m.list.SelectAll()
			newCount := m.list.GetSelectionCount()
			m.addEvent(fmt.Sprintf("Select all: %d items", newCount))
			m.status = fmt.Sprintf("Selected all %d items", newCount)
			return m, nil

		case "ctrl+d":
			oldCount := m.list.GetSelectionCount()
			m.list.ClearSelection()
			m.addEvent(fmt.Sprintf("Clear selection: %d items deselected", oldCount))
			m.status = "Cleared all selections"
			return m, nil

		case "d":
			state := m.list.GetState()
			selectedCount := m.list.GetSelectionCount()
			if selectedCount > 0 {
				m.addEvent(fmt.Sprintf("Delete action: %d selected items", selectedCount))
				m.status = fmt.Sprintf("Delete %d selected items", selectedCount)
			} else {
				m.addEvent(fmt.Sprintf("Delete action: item at index %d", state.CursorIndex))
				m.status = fmt.Sprintf("Delete item at index %d", state.CursorIndex)
			}
			return m, nil

		case "e":
			state := m.list.GetState()
			m.addEvent(fmt.Sprintf("Edit action: index %d", state.CursorIndex))
			m.status = fmt.Sprintf("Edit item at index %d", state.CursorIndex)
			return m, nil
		}

		// Track navigation events
		oldState := m.list.GetState()
		newList, cmd := m.list.Update(msg)
		m.list = newList.(*vtable.TeaList[string])
		newState := m.list.GetState()

		// Check if cursor moved
		if oldState.CursorIndex != newState.CursorIndex {
			m.addEvent(fmt.Sprintf("Navigation: %d → %d", oldState.CursorIndex, newState.CursorIndex))
			m.status = fmt.Sprintf("Moved to item %d", newState.CursorIndex)
		}

		return m, cmd
	}

	// Update the list
	newList, cmd := m.list.Update(msg)
	m.list = newList.(*vtable.TeaList[string])
	return m, cmd
}

func (m *ListCallbackModel) View() string {
	var sb strings.Builder

	sb.WriteString("VTable Example 06: Callbacks - List Demo\n\n")

	// List
	sb.WriteString(m.list.View())
	sb.WriteString("\n\n")

	// Status
	sb.WriteString(m.status)
	sb.WriteString("\n\n")

	// Event log
	sb.WriteString("Event Log:\n")
	if len(m.eventLog) == 0 {
		sb.WriteString("  (no events yet)")
	} else {
		for _, event := range m.eventLog {
			sb.WriteString(fmt.Sprintf("  %s\n", event))
		}
	}
	sb.WriteString("\n")

	// Help
	sb.WriteString("Actions: space=select enter=activate d=delete e=edit ctrl+a=select_all ctrl+d=clear q=quit")

	return sb.String()
}

// ===== TABLE DEMO =====

type Employee struct {
	ID         int
	Name       string
	Department string
	Role       string
	Salary     int
	StartDate  string
	Status     string
}

// Employee provider for table demo
type EmployeeProvider struct {
	employees []Employee
}

func NewEmployeeProvider() *EmployeeProvider {
	return &EmployeeProvider{
		employees: []Employee{
			{1, "Alice Johnson", "Engineering", "Senior Dev", 95000, "2022-01-15", "Active"},
			{2, "Bob Smith", "Marketing", "Manager", 75000, "2021-03-20", "Active"},
			{3, "Carol Davis", "Engineering", "Tech Lead", 110000, "2020-08-10", "Active"},
			{4, "David Wilson", "Sales", "Rep", 60000, "2023-01-05", "Active"},
			{5, "Eve Brown", "HR", "Specialist", 55000, "2022-11-12", "Active"},
			{6, "Frank Miller", "Engineering", "Junior Dev", 70000, "2023-06-01", "Active"},
			{7, "Grace Lee", "Finance", "Analyst", 65000, "2021-09-15", "Active"},
			{8, "Henry Garcia", "Engineering", "Senior Dev", 92000, "2022-04-20", "On Leave"},
			{9, "Ivy Chen", "Marketing", "Designer", 58000, "2023-02-10", "Active"},
			{10, "Jack Taylor", "Sales", "Manager", 85000, "2020-12-01", "Active"},
			{11, "Kelly White", "Engineering", "DevOps", 88000, "2022-07-08", "Active"},
			{12, "Liam Anderson", "HR", "Manager", 78000, "2021-05-25", "Active"},
			{13, "Mia Thompson", "Finance", "Controller", 95000, "2020-10-30", "Active"},
			{14, "Noah Martinez", "Engineering", "Intern", 35000, "2023-09-01", "Active"},
			{15, "Olivia Rodriguez", "Marketing", "Manager", 82000, "2021-01-18", "Active"},
		},
	}
}

func (p *EmployeeProvider) GetTotal() int {
	return len(p.employees)
}

func (p *EmployeeProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[vtable.TableRow], error) {
	start := request.Start
	count := request.Count

	if start >= len(p.employees) {
		return []vtable.Data[vtable.TableRow]{}, nil
	}

	if start+count > len(p.employees) {
		count = len(p.employees) - start
	}

	result := make([]vtable.Data[vtable.TableRow], count)
	for i := 0; i < count; i++ {
		emp := p.employees[start+i]

		row := vtable.TableRow{
			Cells: []string{
				fmt.Sprintf("%d", emp.ID),
				emp.Name,
				emp.Department,
				emp.Role,
				fmt.Sprintf("$%d", emp.Salary),
				emp.StartDate,
				emp.Status,
			},
		}

		result[i] = vtable.Data[vtable.TableRow]{
			ID:       fmt.Sprintf("emp-%d", emp.ID),
			Item:     row,
			Selected: false,
			Metadata: vtable.NewTypedMetadata(),
		}
	}
	return result, nil
}

// Selection methods for DataProvider interface
func (p *EmployeeProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionMultiple
}

func (p *EmployeeProvider) SetSelected(index int, selected bool) bool {
	return true
}

func (p *EmployeeProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	return true
}

func (p *EmployeeProvider) SelectRange(startID, endID string) bool {
	return true
}

func (p *EmployeeProvider) SelectAll() bool {
	return true
}

func (p *EmployeeProvider) ClearSelection() {
}

func (p *EmployeeProvider) GetSelectedIndices() []int {
	return []int{}
}

func (p *EmployeeProvider) GetSelectedIDs() []string {
	return []string{}
}

func (p *EmployeeProvider) GetItemID(item *vtable.TableRow) string {
	return ""
}

// Table model with callbacks
type TableCallbackModel struct {
	table       *vtable.TeaTable
	provider    *EmployeeProvider
	status      string
	eventLog    []string
	maxLogItems int
}

func newTableCallbackModel() *TableCallbackModel {
	provider := NewEmployeeProvider()

	columns := []vtable.TableColumn{
		vtable.NewRightColumn("ID", 4),
		vtable.NewColumn("Name", 15),
		vtable.NewColumn("Department", 12),
		vtable.NewColumn("Role", 12),
		vtable.NewRightColumn("Salary", 8),
		vtable.NewColumn("Start Date", 12),
		vtable.NewColumn("Status", 10),
	}

	table, err := vtable.NewTeaTableWithHeight(columns, provider, 10)
	if err != nil {
		panic(err)
	}

	model := &TableCallbackModel{
		table:       table,
		provider:    provider,
		status:      "Table with callbacks - navigate and interact with rows to see events",
		eventLog:    []string{},
		maxLogItems: 8,
	}

	return model
}

func (m *TableCallbackModel) Init() tea.Cmd {
	return m.table.Init()
}

func (m *TableCallbackModel) addEvent(event string) {
	timestamp := time.Now().Format("15:04:05")
	m.eventLog = append(m.eventLog, fmt.Sprintf("[%s] %s", timestamp, event))

	// Keep only the last maxLogItems events
	if len(m.eventLog) > m.maxLogItems {
		m.eventLog = m.eventLog[len(m.eventLog)-m.maxLogItems:]
	}
}

func (m *TableCallbackModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }

		case "enter":
			state := m.table.GetState()
			m.addEvent(fmt.Sprintf("Row activated: index %d", state.CursorIndex))
			m.status = fmt.Sprintf("Activated row %d", state.CursorIndex)
			return m, nil

		case " ":
			oldCount := m.table.GetSelectionCount()
			if m.table.ToggleCurrentSelection() {
				state := m.table.GetState()
				newCount := m.table.GetSelectionCount()
				if newCount > oldCount {
					m.addEvent(fmt.Sprintf("Row selected: index %d", state.CursorIndex))
					m.status = fmt.Sprintf("Selected row %d (total: %d)", state.CursorIndex, newCount)
				} else {
					m.addEvent(fmt.Sprintf("Row deselected: index %d", state.CursorIndex))
					m.status = fmt.Sprintf("Deselected row %d (total: %d)", state.CursorIndex, newCount)
				}
			}
			return m, nil

		case "ctrl+a":
			m.table.SelectAll()
			newCount := m.table.GetSelectionCount()
			m.addEvent(fmt.Sprintf("Select all: %d rows", newCount))
			m.status = fmt.Sprintf("Selected all %d rows", newCount)
			return m, nil

		case "ctrl+d":
			oldCount := m.table.GetSelectionCount()
			m.table.ClearSelection()
			m.addEvent(fmt.Sprintf("Clear selection: %d rows deselected", oldCount))
			m.status = "Cleared all selections"
			return m, nil

		case "delete":
			state := m.table.GetState()
			selectedCount := m.table.GetSelectionCount()
			if selectedCount > 0 {
				m.addEvent(fmt.Sprintf("Delete action: %d selected rows", selectedCount))
				m.status = fmt.Sprintf("Delete %d selected employees", selectedCount)
			} else {
				m.addEvent(fmt.Sprintf("Delete action: row at index %d", state.CursorIndex))
				m.status = fmt.Sprintf("Delete employee at row %d", state.CursorIndex)
			}
			return m, nil

		case "e":
			state := m.table.GetState()
			m.addEvent(fmt.Sprintf("Edit action: row %d", state.CursorIndex))
			m.status = fmt.Sprintf("Edit employee at row %d", state.CursorIndex)
			return m, nil

		case "v":
			state := m.table.GetState()
			m.addEvent(fmt.Sprintf("View action: row %d", state.CursorIndex))
			m.status = fmt.Sprintf("View details for employee at row %d", state.CursorIndex)
			return m, nil
		}

		// Track navigation events
		oldState := m.table.GetState()
		newTable, cmd := m.table.Update(msg)
		m.table = newTable.(*vtable.TeaTable)
		newState := m.table.GetState()

		// Check if cursor moved
		if oldState.CursorIndex != newState.CursorIndex {
			m.addEvent(fmt.Sprintf("Navigation: row %d → %d", oldState.CursorIndex, newState.CursorIndex))
			m.status = fmt.Sprintf("Moved to row %d", newState.CursorIndex)
		}

		return m, cmd
	}

	// Update the table
	newTable, cmd := m.table.Update(msg)
	m.table = newTable.(*vtable.TeaTable)
	return m, cmd
}

func (m *TableCallbackModel) View() string {
	var sb strings.Builder

	sb.WriteString("VTable Example 06: Callbacks - Table Demo\n\n")

	// Table
	sb.WriteString(m.table.View())
	sb.WriteString("\n\n")

	// Status
	sb.WriteString(m.status)
	sb.WriteString("\n\n")

	// Event log
	sb.WriteString("Event Log:\n")
	if len(m.eventLog) == 0 {
		sb.WriteString("  (no events yet)")
	} else {
		for _, event := range m.eventLog {
			sb.WriteString(fmt.Sprintf("  %s\n", event))
		}
	}
	sb.WriteString("\n")

	// Help
	sb.WriteString("Actions: space=select enter=activate e=edit v=view delete=delete ctrl+a=select_all ctrl+d=clear q=quit")

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
