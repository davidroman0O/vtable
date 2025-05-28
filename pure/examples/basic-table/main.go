package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	vtable "github.com/davidroman0O/vtable/pure"
)

// ================================
// DATA MODEL
// ================================

// Employee represents a simple employee record
type Employee struct {
	ID         int
	Name       string
	Department string
	Position   string
	Salary     int
	StartDate  time.Time
	Active     bool
}

// ================================
// DATA SOURCE IMPLEMENTATION
// ================================

// EmployeeDataSource implements TableDataSource for Employee data
type EmployeeDataSource struct {
	employees []vtable.Data[Employee]
	columns   []vtable.TableColumn
	sortBy    []string
	sortDirs  []string
	filters   map[string]any
}

// NewEmployeeDataSource creates a new employee data source with sample data
func NewEmployeeDataSource() *EmployeeDataSource {
	// Generate sample employee data
	employees := generateSampleEmployees(500) // Large dataset to test chunking

	// Convert to Data[Employee] format
	dataItems := make([]vtable.Data[Employee], len(employees))
	for i, emp := range employees {
		dataItems[i] = vtable.Data[Employee]{
			ID:       fmt.Sprintf("emp_%d", emp.ID),
			Item:     emp,
			Selected: false,
			Metadata: vtable.NewTypedMetadata(),
			Disabled: !emp.Active, // Inactive employees are disabled
			Hidden:   false,
			Error:    nil,
			Loading:  false,
		}
	}

	// Define table columns
	columns := []vtable.TableColumn{
		{Title: "ID", Field: "id", Width: 6, Alignment: vtable.AlignRight},
		{Title: "Name", Field: "name", Width: 20, Alignment: vtable.AlignLeft},
		{Title: "Department", Field: "department", Width: 15, Alignment: vtable.AlignLeft},
		{Title: "Position", Field: "position", Width: 18, Alignment: vtable.AlignLeft},
		{Title: "Salary", Field: "salary", Width: 10, Alignment: vtable.AlignRight},
		{Title: "Start Date", Field: "start_date", Width: 12, Alignment: vtable.AlignCenter},
		{Title: "Status", Field: "active", Width: 8, Alignment: vtable.AlignCenter},
	}

	return &EmployeeDataSource{
		employees: dataItems,
		columns:   columns,
		sortBy:    []string{},
		sortDirs:  []string{},
		filters:   make(map[string]any),
	}
}

// generateSampleEmployees creates sample employee data
func generateSampleEmployees(count int) []Employee {
	departments := []string{"Engineering", "Sales", "Marketing", "HR", "Finance", "Operations", "Support"}
	positions := []string{"Manager", "Senior", "Junior", "Lead", "Director", "Analyst", "Specialist"}
	names := []string{
		"Alice Johnson", "Bob Smith", "Carol Davis", "David Wilson", "Eva Brown",
		"Frank Miller", "Grace Lee", "Henry Taylor", "Ivy Chen", "Jack Anderson",
		"Kate Thompson", "Liam Garcia", "Mia Rodriguez", "Noah Martinez", "Olivia Lopez",
		"Paul Hernandez", "Quinn Gonzalez", "Ruby Perez", "Sam Wilson", "Tina Moore",
	}

	employees := make([]Employee, count)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < count; i++ {
		dept := departments[rand.Intn(len(departments))]
		pos := positions[rand.Intn(len(positions))]
		name := names[rand.Intn(len(names))]

		// Add some variety to names
		if rand.Float32() < 0.3 {
			name = fmt.Sprintf("%s %d", name, rand.Intn(100))
		}

		employees[i] = Employee{
			ID:         i + 1,
			Name:       name,
			Department: dept,
			Position:   fmt.Sprintf("%s %s", pos, dept),
			Salary:     40000 + rand.Intn(120000), // $40k - $160k
			StartDate:  time.Now().AddDate(-rand.Intn(10), -rand.Intn(12), -rand.Intn(30)),
			Active:     rand.Float32() > 0.1, // 90% active
		}
	}

	return employees
}

// ================================
// IMPLEMENT TableDataSource INTERFACE
// ================================

func (ds *EmployeeDataSource) LoadChunk(request vtable.DataRequest) tea.Cmd {
	return func() tea.Msg {
		return ds.LoadChunkImmediate(request)
	}
}

func (ds *EmployeeDataSource) LoadChunkImmediate(request vtable.DataRequest) vtable.DataChunkLoadedMsg {
	// Apply filters and sorting
	filtered := ds.applyFiltersAndSort()

	// Calculate chunk bounds
	start := request.Start
	end := start + request.Count
	if end > len(filtered) {
		end = len(filtered)
	}

	// Extract chunk
	var items []vtable.Data[any]
	if start < len(filtered) {
		for i := start; i < end; i++ {
			items = append(items, vtable.Data[any]{
				ID:       filtered[i].ID,
				Item:     filtered[i].Item,
				Selected: filtered[i].Selected,
				Metadata: filtered[i].Metadata,
				Disabled: filtered[i].Disabled,
				Hidden:   filtered[i].Hidden,
				Error:    filtered[i].Error,
				Loading:  filtered[i].Loading,
			})
		}
	}

	return vtable.DataChunkLoadedMsg{
		StartIndex: start,
		Items:      items,
		Request:    request,
	}
}

func (ds *EmployeeDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		filtered := ds.applyFiltersAndSort()
		return vtable.DataTotalMsg{Total: len(filtered)}
	}
}

func (ds *EmployeeDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

func (ds *EmployeeDataSource) GetItemID(item any) string {
	if emp, ok := item.(Employee); ok {
		return fmt.Sprintf("emp_%d", emp.ID)
	}
	return ""
}

func (ds *EmployeeDataSource) SetSelected(index int, selected bool) tea.Cmd {
	filtered := ds.applyFiltersAndSort()
	if index >= 0 && index < len(filtered) {
		// Find the original item and update it
		id := filtered[index].ID
		for i := range ds.employees {
			if ds.employees[i].ID == id {
				ds.employees[i].Selected = selected
				break
			}
		}
	}
	return func() tea.Msg {
		return vtable.SelectionResponseMsg{Success: true}
	}
}

func (ds *EmployeeDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	for i := range ds.employees {
		if ds.employees[i].ID == id {
			ds.employees[i].Selected = selected
			break
		}
	}
	return func() tea.Msg {
		return vtable.SelectionResponseMsg{Success: true}
	}
}

func (ds *EmployeeDataSource) SelectAll() tea.Cmd {
	filtered := ds.applyFiltersAndSort()
	for _, item := range filtered {
		for i := range ds.employees {
			if ds.employees[i].ID == item.ID {
				ds.employees[i].Selected = true
				break
			}
		}
	}
	return func() tea.Msg {
		return vtable.SelectionResponseMsg{Success: true}
	}
}

func (ds *EmployeeDataSource) ClearSelection() tea.Cmd {
	for i := range ds.employees {
		ds.employees[i].Selected = false
	}
	return func() tea.Msg {
		return vtable.SelectionResponseMsg{Success: true}
	}
}

func (ds *EmployeeDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	filtered := ds.applyFiltersAndSort()
	for i := startIndex; i <= endIndex && i < len(filtered); i++ {
		id := filtered[i].ID
		for j := range ds.employees {
			if ds.employees[j].ID == id {
				ds.employees[j].Selected = true
				break
			}
		}
	}
	return func() tea.Msg {
		return vtable.SelectionResponseMsg{Success: true}
	}
}

func (ds *EmployeeDataSource) GetColumns() []vtable.TableColumn {
	return ds.columns
}

func (ds *EmployeeDataSource) GetCellValue(item Employee, columnField string) any {
	switch columnField {
	case "id":
		return item.ID
	case "name":
		return item.Name
	case "department":
		return item.Department
	case "position":
		return item.Position
	case "salary":
		return fmt.Sprintf("$%s", formatNumber(item.Salary))
	case "start_date":
		return item.StartDate.Format("2006-01-02")
	case "active":
		if item.Active {
			return "Active"
		}
		return "Inactive"
	default:
		return ""
	}
}

func (ds *EmployeeDataSource) SortBy(fields []string, directions []string) tea.Cmd {
	ds.sortBy = fields
	ds.sortDirs = directions
	return func() tea.Msg {
		return vtable.DataRefreshMsg{}
	}
}

func (ds *EmployeeDataSource) FilterBy(filters map[string]any) tea.Cmd {
	ds.filters = filters
	return func() tea.Msg {
		return vtable.DataRefreshMsg{}
	}
}

// ================================
// HELPER METHODS
// ================================

func (ds *EmployeeDataSource) applyFiltersAndSort() []vtable.Data[Employee] {
	// Start with all employees
	result := make([]vtable.Data[Employee], 0, len(ds.employees))

	// Apply filters
	for _, emp := range ds.employees {
		include := true

		// Apply each filter
		for field, value := range ds.filters {
			switch field {
			case "department":
				if dept, ok := value.(string); ok && dept != "" {
					if !strings.Contains(strings.ToLower(emp.Item.Department), strings.ToLower(dept)) {
						include = false
						break
					}
				}
			case "active":
				if active, ok := value.(bool); ok {
					if emp.Item.Active != active {
						include = false
						break
					}
				}
			case "min_salary":
				if minSal, ok := value.(int); ok {
					if emp.Item.Salary < minSal {
						include = false
						break
					}
				}
			}
		}

		if include {
			result = append(result, emp)
		}
	}

	// Apply sorting (simple implementation)
	if len(ds.sortBy) > 0 {
		field := ds.sortBy[0]
		ascending := len(ds.sortDirs) == 0 || ds.sortDirs[0] == "asc"

		// Simple bubble sort for demonstration
		for i := 0; i < len(result)-1; i++ {
			for j := 0; j < len(result)-i-1; j++ {
				shouldSwap := false

				switch field {
				case "id":
					if ascending {
						shouldSwap = result[j].Item.ID > result[j+1].Item.ID
					} else {
						shouldSwap = result[j].Item.ID < result[j+1].Item.ID
					}
				case "name":
					if ascending {
						shouldSwap = result[j].Item.Name > result[j+1].Item.Name
					} else {
						shouldSwap = result[j].Item.Name < result[j+1].Item.Name
					}
				case "salary":
					if ascending {
						shouldSwap = result[j].Item.Salary > result[j+1].Item.Salary
					} else {
						shouldSwap = result[j].Item.Salary < result[j+1].Item.Salary
					}
				}

				if shouldSwap {
					result[j], result[j+1] = result[j+1], result[j]
				}
			}
		}
	}

	return result
}

func formatNumber(n int) string {
	str := strconv.Itoa(n)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}
	return result.String()
}

// ================================
// MAIN APPLICATION
// ================================

type model struct {
	table         *vtable.Table[Employee]
	dataSource    *EmployeeDataSource
	currentSort   string
	currentFilter string
	showHelp      bool
	statusMsg     string
}

func initialModel() model {
	// Create data source
	dataSource := NewEmployeeDataSource()

	// Create list config
	listConfig := vtable.NewListConfigBuilder().
		WithViewportHeight(15).
		WithChunkSize(50).
		WithSelectionMode(vtable.SelectionMultiple).
		WithMaxWidth(120).
		Build()

	// Create table config
	tableConfig := vtable.DefaultTableConfig()
	tableConfig.Columns = dataSource.GetColumns()
	tableConfig.ShowHeader = true
	tableConfig.ShowBorders = true

	// Create table
	table := vtable.NewTable(listConfig, tableConfig, dataSource)

	// Focus the table so navigation works
	table.Focus()

	return model{
		table:         table,
		dataSource:    dataSource,
		currentSort:   "none",
		currentFilter: "none",
		showHelp:      true,
		statusMsg:     "Basic Table Example - Press '?' to toggle help",
	}
}

func (m model) Init() tea.Cmd {
	return m.table.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		switch key {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		// === NAVIGATION KEYS (vim-style) ===
		case "j", "up":
			// Move up
			m.statusMsg = "Moving up"
			return m, vtable.CursorUpCmd()

		case "k", "down":
			// Move down
			m.statusMsg = "Moving down"
			return m, vtable.CursorDownCmd()

		case "h":
			// Page up
			m.statusMsg = "Page up"
			return m, vtable.PageUpCmd()

		case "l":
			// Page down
			m.statusMsg = "Page down"
			return m, vtable.PageDownCmd()

		case "g":
			// Jump to start
			m.statusMsg = "Jump to start"
			return m, vtable.JumpToStartCmd()

		case "G":
			// Jump to end
			m.statusMsg = "Jump to end"
			return m, vtable.JumpToEndCmd()

		case " ", "enter":
			// Select current item
			m.statusMsg = "Toggling selection"
			return m, vtable.SelectCurrentCmd()

		case "a":
			// Select all
			m.statusMsg = "Selecting all"
			return m, vtable.SelectAllCmd()

		case "c":
			// Clear selection
			m.statusMsg = "Clearing selection"
			return m, vtable.SelectClearCmd()

		case "s":
			// Cycle through sort options
			switch m.currentSort {
			case "none":
				m.currentSort = "name_asc"
				m.statusMsg = "Sorted by Name (A-Z)"
				return m, m.table.SortByColumn("name", "asc")
			case "name_asc":
				m.currentSort = "name_desc"
				m.statusMsg = "Sorted by Name (Z-A)"
				return m, m.table.SortByColumn("name", "desc")
			case "name_desc":
				m.currentSort = "salary_asc"
				m.statusMsg = "Sorted by Salary (Low-High)"
				return m, m.table.SortByColumn("salary", "asc")
			case "salary_asc":
				m.currentSort = "salary_desc"
				m.statusMsg = "Sorted by Salary (High-Low)"
				return m, m.table.SortByColumn("salary", "desc")
			case "salary_desc":
				m.currentSort = "id_asc"
				m.statusMsg = "Sorted by ID (1-999)"
				return m, m.table.SortByColumn("id", "asc")
			default:
				m.currentSort = "none"
				m.statusMsg = "Sort cleared"
				return m, m.table.ClearSort()
			}

		case "f":
			// Cycle through filter options
			switch m.currentFilter {
			case "none":
				m.currentFilter = "engineering"
				m.statusMsg = "Filtered: Engineering department only"
				return m, m.table.FilterByColumn("department", "Engineering")
			case "engineering":
				m.currentFilter = "active"
				m.statusMsg = "Filtered: Active employees only"
				return m, m.table.FilterByColumn("active", true)
			case "active":
				m.currentFilter = "high_salary"
				m.statusMsg = "Filtered: Salary > $80,000"
				return m, m.table.FilterByColumn("min_salary", 80000)
			default:
				m.currentFilter = "none"
				m.statusMsg = "All filters cleared"
				return m, m.table.ClearAllFilters()
			}

		case "r":
			m.statusMsg = "Data refreshed"
			return m, vtable.DataRefreshCmd()

		default:
			// Pass other keys to table (like arrow keys)
			tableModel, cmd := m.table.Update(msg)
			m.table = tableModel.(*vtable.Table[Employee])
			return m, cmd
		}

	default:
		// Pass non-keyboard messages to table
		tableModel, cmd := m.table.Update(msg)
		m.table = tableModel.(*vtable.Table[Employee])
		return m, cmd
	}
}

func (m model) View() string {
	var b strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		Render("📊 Basic Table Example")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Status
	status := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(m.statusMsg)
	b.WriteString(status)
	b.WriteString("\n\n")

	// Table
	b.WriteString(m.table.View())
	b.WriteString("\n\n")

	// Help
	if m.showHelp {
		help := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render(getHelpText())
		b.WriteString(help)
	}

	// Footer
	state := m.table.GetState()
	selectionCount := m.table.GetSelectionCount()
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(fmt.Sprintf(
			"Position: %d | Selected: %d | Sort: %s | Filter: %s | Press '?' for help",
			state.CursorIndex+1,
			selectionCount,
			m.currentSort,
			m.currentFilter,
		))
	b.WriteString("\n")
	b.WriteString(footer)

	return b.String()
}

func getHelpText() string {
	return `Navigation (Working Keys):
  j/↑         Move up      k/↓         Move down
  h           Page up      l           Page down  
  g           Go to start    G           Go to end
  Space/Enter Select item    a           Select all
  c           Clear selection

Actions:
  s           Cycle sort     f           Cycle filter
  r           Refresh data   ?           Toggle help
  q/Ctrl+C    Quit

Current dataset: 500 employees with chunked loading (50 items per chunk)`
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
