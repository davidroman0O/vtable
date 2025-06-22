package main

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/table"
)

// Employee represents our data structure
type Employee struct {
	ID         string
	Name       string
	Department string
	Status     string
	Salary     int
	Email      string
	Phone      string
}

// EmployeeDataSource implements filtering and sorting capabilities
type EmployeeDataSource struct {
	employees      []Employee
	filteredData   []Employee     // Cached filtered/sorted result
	filteredTotal  int            // Count after filtering
	activeFilters  map[string]any // Current filters
	sortFields     []string       // Current sort fields
	sortDirections []string       // Current sort directions
}

func NewEmployeeDataSource() *EmployeeDataSource {
	employees := []Employee{
		{"EMP001", "Alice Johnson", "Engineering", "Active", 85000, "alice@company.com", "(555) 123-4567"},
		{"EMP002", "Bob Smith", "Marketing", "Active", 65000, "bob@company.com", "(555) 234-5678"},
		{"EMP003", "Carol Davis", "Sales", "Remote", 70000, "carol@company.com", "(555) 345-6789"},
		{"EMP004", "David Wilson", "HR", "On Leave", 60000, "david@company.com", "(555) 456-7890"},
		{"EMP005", "Eve Brown", "Finance", "Active", 90000, "eve@company.com", "(555) 567-8901"},
		{"EMP006", "Frank Miller", "Engineering", "Active", 78000, "frank@company.com", "(555) 678-9012"},
		{"EMP007", "Grace Lee", "Marketing", "Part-time", 45000, "grace@company.com", "(555) 789-0123"},
		{"EMP008", "Henry Chen", "Sales", "Active", 72000, "henry@company.com", "(555) 890-1234"},
		{"EMP009", "Ivy Wang", "Engineering", "Active", 88000, "ivy@company.com", "(555) 901-2345"},
		{"EMP010", "Jack Brown", "Finance", "Remote", 75000, "jack@company.com", "(555) 012-3456"},
		{"EMP011", "Kate Wilson", "HR", "Active", 68000, "kate@company.com", "(555) 123-4567"},
		{"EMP012", "Leo Garcia", "Marketing", "Active", 62000, "leo@company.com", "(555) 234-5678"},
	}

	ds := &EmployeeDataSource{
		employees:      employees,
		activeFilters:  make(map[string]any),
		sortFields:     []string{},
		sortDirections: []string{},
	}

	// Initialize with all data
	ds.rebuildData()
	return ds
}

// Filter and sort management methods
func (ds *EmployeeDataSource) SetFilter(field string, value any) {
	ds.activeFilters[field] = value
	ds.rebuildData()
}

func (ds *EmployeeDataSource) ClearFilter(field string) {
	delete(ds.activeFilters, field)
	ds.rebuildData()
}

func (ds *EmployeeDataSource) ClearAllFilters() {
	ds.activeFilters = make(map[string]any)
	ds.rebuildData()
}

func (ds *EmployeeDataSource) SetSort(fields []string, directions []string) {
	ds.sortFields = fields
	ds.sortDirections = directions
	ds.rebuildData()
}

func (ds *EmployeeDataSource) ClearSort() {
	ds.sortFields = []string{}
	ds.sortDirections = []string{}
	ds.rebuildData()
}

// Data rebuilding logic
func (ds *EmployeeDataSource) rebuildData() {
	// Start with all data
	result := make([]Employee, 0, len(ds.employees))

	// Apply filters
	for _, emp := range ds.employees {
		if ds.matchesFilters(emp) {
			result = append(result, emp)
		}
	}

	// Apply sorting
	if len(ds.sortFields) > 0 {
		sort.Slice(result, func(i, j int) bool {
			return ds.compareEmployees(result[i], result[j])
		})
	}

	ds.filteredData = result
	ds.filteredTotal = len(result)
}

func (ds *EmployeeDataSource) matchesFilters(emp Employee) bool {
	for field, value := range ds.activeFilters {
		switch field {
		case "department":
			if emp.Department != value.(string) {
				return false
			}
		case "status":
			if emp.Status != value.(string) {
				return false
			}
		case "search":
			searchTerm := strings.ToLower(value.(string))
			if !strings.Contains(strings.ToLower(emp.Name), searchTerm) &&
				!strings.Contains(strings.ToLower(emp.Department), searchTerm) &&
				!strings.Contains(strings.ToLower(emp.Email), searchTerm) {
				return false
			}
		case "salary_min":
			if emp.Salary < value.(int) {
				return false
			}
		case "salary_max":
			if emp.Salary > value.(int) {
				return false
			}
		case "engineering":
			if emp.Department != "Engineering" {
				return false
			}
		case "marketing":
			if emp.Department != "Marketing" {
				return false
			}
		case "sales":
			if emp.Department != "Sales" {
				return false
			}
		case "finance":
			if emp.Department != "Finance" {
				return false
			}
		case "hr":
			if emp.Department != "HR" {
				return false
			}
		case "active_only":
			if emp.Status != "Active" {
				return false
			}
		case "remote_only":
			if emp.Status != "Remote" {
				return false
			}
		case "high_salary":
			if emp.Salary < 75000 {
				return false
			}
		case "low_salary":
			if emp.Salary >= 65000 {
				return false
			}
		}
	}
	return true
}

func (ds *EmployeeDataSource) compareEmployees(a, b Employee) bool {
	for i, field := range ds.sortFields {
		direction := "asc"
		if i < len(ds.sortDirections) {
			direction = ds.sortDirections[i]
		}

		var cmp int
		switch field {
		case "name":
			cmp = strings.Compare(a.Name, b.Name)
		case "department":
			cmp = strings.Compare(a.Department, b.Department)
		case "salary":
			if a.Salary < b.Salary {
				cmp = -1
			} else if a.Salary > b.Salary {
				cmp = 1
			}
		case "status":
			cmp = strings.Compare(a.Status, b.Status)
		case "id":
			cmp = strings.Compare(a.ID, b.ID)
		case "email":
			cmp = strings.Compare(a.Email, b.Email)
		case "phone":
			cmp = strings.Compare(a.Phone, b.Phone)
		}

		if cmp != 0 {
			if direction == "desc" {
				return cmp > 0
			}
			return cmp < 0
		}
	}
	return false
}

// Required DataSource interface methods
func (ds *EmployeeDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return core.DataTotalMsg{Total: ds.filteredTotal}
	}
}

func (ds *EmployeeDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {

		end := request.Start + request.Count
		if end > ds.filteredTotal {
			end = ds.filteredTotal
		}

		chunkItems := make([]core.Data[any], end-request.Start)
		for i := request.Start; i < end; i++ {
			emp := ds.filteredData[i]
			chunkItems[i-request.Start] = core.Data[any]{
				ID: emp.ID,
				Item: core.TableRow{
					ID:    emp.ID,
					Cells: []string{emp.ID, emp.Name, emp.Department, emp.Status, fmt.Sprintf("$%d", emp.Salary), emp.Email, emp.Phone},
				},
				Metadata: core.NewTypedMetadata(),
			}
		}

		return core.DataChunkLoadedMsg{
			StartIndex: request.Start,
			Items:      chunkItems,
			Request:    request,
		}
	}
}

// Simplified interface methods
func (ds *EmployeeDataSource) RefreshTotal() tea.Cmd                            { return ds.GetTotal() }
func (ds *EmployeeDataSource) SetSelected(index int, selected bool) tea.Cmd     { return nil }
func (ds *EmployeeDataSource) SetSelectedByID(id string, selected bool) tea.Cmd { return nil }
func (ds *EmployeeDataSource) SelectAll() tea.Cmd                               { return nil }
func (ds *EmployeeDataSource) ClearSelection() tea.Cmd                          { return nil }
func (ds *EmployeeDataSource) SelectRange(startIndex, endIndex int) tea.Cmd     { return nil }
func (ds *EmployeeDataSource) GetItemID(item any) string                        { return "" }

type AppModel struct {
	table         *table.Table
	dataSource    *EmployeeDataSource
	statusMessage string

	// Filter state
	currentFilter string
	activeFilters map[string]bool // Track which number filters are active

	// Sort state
	currentSort    string
	currentSortDir string

	// Search state
	searchMode   bool
	searchTerm   string
	searchActive bool
}

// `05-table-component/examples/filtering-sorting/main.go`
func main() {
	dataSource := NewEmployeeDataSource()

	columns := []core.TableColumn{
		{Title: "ID", Width: 8, Alignment: core.AlignCenter, Field: "id"},
		{Title: "Name", Width: 20, Alignment: core.AlignLeft, Field: "name"},
		{Title: "Department", Width: 15, Alignment: core.AlignCenter, Field: "department"},
		{Title: "Status", Width: 12, Alignment: core.AlignCenter, Field: "status"},
		{Title: "Salary", Width: 12, Alignment: core.AlignRight, Field: "salary"},
		{Title: "Email", Width: 25, Alignment: core.AlignLeft, Field: "email"},
		{Title: "Phone", Width: 15, Alignment: core.AlignCenter, Field: "phone"},
	}

	theme := core.Theme{
		HeaderStyle:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("57")),
		CellStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		CursorStyle:        lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("21")),
		SelectedStyle:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("57")),
		FullRowCursorStyle: lipgloss.NewStyle().Background(lipgloss.Color("21")).Foreground(lipgloss.Color("15")),
		BorderChars: core.BorderChars{
			Horizontal: "─", Vertical: "│", TopLeft: "┌", TopRight: "┐",
			BottomLeft: "└", BottomRight: "┘", TopT: "┬", BottomT: "┴",
			LeftT: "├", RightT: "┤", Cross: "┼",
		},
		BorderColor: "8",
	}

	config := core.TableConfig{
		Columns:     columns,
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: core.ViewportConfig{
			Height:    10,
			ChunkSize: 20,
		},
		Theme:                       theme,
		SelectionMode:               core.SelectionNone,
		ActiveCellIndicationEnabled: true,
	}

	tbl := table.NewTable(config, dataSource)
	tbl.Focus()

	model := AppModel{
		table:          tbl,
		dataSource:     dataSource,
		statusMessage:  "Enhanced Filtering & Sorting - Press s to sort active column, 1-9 for filters, / to search",
		currentFilter:  "",
		activeFilters:  make(map[string]bool),
		currentSort:    "",
		currentSortDir: "",
		searchMode:     false,
		searchTerm:     "",
		searchActive:   false,
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.table.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle search mode first
		if m.searchMode {
			switch msg.String() {
			case "enter":
				m.searchMode = false
				if m.searchTerm != "" {
					m.dataSource.SetFilter("search", m.searchTerm)
					m.searchActive = true
					m.statusMessage = fmt.Sprintf("Searching for: %s (%d results)", m.searchTerm, m.dataSource.filteredTotal)
					return m, core.DataRefreshCmd()
				} else {
					m.statusMessage = "Search cancelled"
					return m, nil
				}
			case "escape":
				m.searchMode = false
				m.searchTerm = ""
				m.statusMessage = "Search cancelled"
				return m, nil
			case "backspace":
				if len(m.searchTerm) > 0 {
					m.searchTerm = m.searchTerm[:len(m.searchTerm)-1]
				}
				return m, nil
			default:
				if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
					m.searchTerm += msg.String()
				}
				return m, nil
			}
		}

		// Normal key handling
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		// Column-based sorting
		case "s":
			return m.sortByActiveColumn()

		case "S":
			return m.clearSorting()

		// Number key filters (cumulative)
		case "1":
			return m.toggleFilter("engineering", "Engineering Dept")
		case "2":
			return m.toggleFilter("marketing", "Marketing Dept")
		case "3":
			return m.toggleFilter("sales", "Sales Dept")
		case "4":
			return m.toggleFilter("finance", "Finance Dept")
		case "5":
			return m.toggleFilter("hr", "HR Dept")
		case "6":
			return m.toggleFilter("active_only", "Active Status")
		case "7":
			return m.toggleFilter("remote_only", "Remote Status")
		case "8":
			return m.toggleFilter("high_salary", "High Salary (≥$75k)")
		case "9":
			return m.toggleFilter("low_salary", "Lower Salary (<$65k)")

		// Clear all filters
		case "0":
			return m.clearAllFilters()

		// Search control
		case "/":
			return m.enterSearchMode()

		// Navigation
		case "j", "down":
			return m, core.CursorDownCmd()

		case "k", "up":
			return m, core.CursorUpCmd()

		case "left":
			// Navigate to previous column (active cell)
			return m, core.PrevColumnCmd()

		case "right":
			// Navigate to next column (active cell)
			return m, core.NextColumnCmd()

		case ".", ",":
			// Allow column navigation with . and , as well
			var cmd tea.Cmd
			_, cmd = m.table.Update(msg)
			return m, cmd

		// Pass other keys to table
		default:
			var cmd tea.Cmd
			_, cmd = m.table.Update(msg)
			return m, cmd
		}

	default:
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd
	}
}

func (m AppModel) sortByActiveColumn() (tea.Model, tea.Cmd) {
	// Get current active column
	_, _, currentColumn, _ := m.table.GetHorizontalScrollState()

	// Map column index to field name
	columnFields := []string{"id", "name", "department", "status", "salary", "email", "phone"}
	columnNames := []string{"ID", "Name", "Department", "Status", "Salary", "Email", "Phone"}

	if currentColumn < len(columnFields) {
		field := columnFields[currentColumn]
		columnName := columnNames[currentColumn]

		// Toggle sort direction if same column, otherwise start with ascending
		if m.currentSort == field {
			if m.currentSortDir == "asc" {
				m.currentSortDir = "desc"
				m.statusMessage = fmt.Sprintf("Sort: %s (Z→A / High→Low)", columnName)
			} else {
				// Clear sorting
				m.currentSort = ""
				m.currentSortDir = ""
				m.dataSource.ClearSort()
				m.statusMessage = "Sorting cleared"
				return m, core.DataRefreshCmd()
			}
		} else {
			m.currentSort = field
			m.currentSortDir = "asc"
			m.statusMessage = fmt.Sprintf("Sort: %s (A→Z / Low→High)", columnName)
		}

		m.dataSource.SetSort([]string{field}, []string{m.currentSortDir})
		return m, core.DataRefreshCmd()
	}

	return m, nil
}

func (m AppModel) toggleFilter(filterKey, filterName string) (tea.Model, tea.Cmd) {
	if m.activeFilters[filterKey] {
		// Remove filter
		m.activeFilters[filterKey] = false
		m.dataSource.ClearFilter(filterKey)
		m.statusMessage = fmt.Sprintf("Removed filter: %s (%d results)", filterName, m.dataSource.filteredTotal)
	} else {
		// Add filter
		m.activeFilters[filterKey] = true
		m.dataSource.SetFilter(filterKey, true)
		m.statusMessage = fmt.Sprintf("Added filter: %s (%d results)", filterName, m.dataSource.filteredTotal)
	}

	return m, core.DataRefreshCmd()
}

func (m AppModel) clearAllFilters() (tea.Model, tea.Cmd) {
	m.activeFilters = make(map[string]bool)
	m.searchActive = false
	m.dataSource.ClearAllFilters()
	m.statusMessage = fmt.Sprintf("All filters cleared (%d results)", m.dataSource.filteredTotal)
	return m, core.DataRefreshCmd()
}

func (m AppModel) clearSorting() (tea.Model, tea.Cmd) {
	m.currentSort = ""
	m.currentSortDir = ""
	m.dataSource.ClearSort()
	m.statusMessage = "Sorting cleared - original order"
	return m, core.DataRefreshCmd()
}

func (m AppModel) enterSearchMode() (tea.Model, tea.Cmd) {
	m.searchMode = true
	m.searchTerm = ""
	m.statusMessage = "Search mode: Type to filter data, Enter to apply, Esc to cancel"
	return m, nil
}

func (m AppModel) View() string {
	var view strings.Builder

	// Show current state
	view.WriteString("=== ENHANCED FILTERING & SORTING ===\n")

	// Get current active column info
	_, _, currentColumn, _ := m.table.GetHorizontalScrollState()
	columnNames := []string{"ID", "Name", "Department", "Status", "Salary", "Email", "Phone"}
	currentColumnName := "Unknown"
	if currentColumn < len(columnNames) {
		currentColumnName = columnNames[currentColumn]
	}

	view.WriteString(fmt.Sprintf("Data: %d/%d | Active Column: %s | Sort: %s | Filters: %s\n",
		m.dataSource.filteredTotal,
		len(m.dataSource.employees),
		currentColumnName,
		m.getSortDescription(),
		m.getActiveFiltersDescription(),
	))

	// Show controls
	view.WriteString("Controls: s=sort-active-column S=clear-sort | 1-9=toggle-filters 0=clear-filters | /=search | .,=change-column | ↑↓jk=navigate | q=quit\n")

	// Show status or search prompt
	if m.searchMode {
		view.WriteString(fmt.Sprintf("🔍 Search: %s_\n", m.searchTerm))
	} else {
		view.WriteString(fmt.Sprintf("Status: %s\n", m.statusMessage))
	}

	view.WriteString("\n")

	// Show table
	view.WriteString(m.table.View())

	// Show help
	if !m.searchMode {
		view.WriteString("\n\nNUMBER FILTERS (cumulative): 1=Engineering 2=Marketing 3=Sales 4=Finance 5=HR | 6=Active 7=Remote | 8=High$ 9=Low$ | 0=Clear")
		view.WriteString("\nCOLUMN SORTING: Navigate to column with . or ,  then press 's' to sort by that column (asc→desc→clear)")
		view.WriteString("\nSEARCH (/): Type to search across Name, Department, Email")
	}

	return view.String()
}

func (m AppModel) getSortDescription() string {
	if m.currentSort == "" {
		return "None"
	}

	direction := "↑"
	if m.currentSortDir == "desc" {
		direction = "↓"
	}

	fieldNames := map[string]string{
		"id":         "ID",
		"name":       "Name",
		"department": "Dept",
		"status":     "Status",
		"salary":     "Salary",
		"email":      "Email",
		"phone":      "Phone",
	}

	if name, exists := fieldNames[m.currentSort]; exists {
		return fmt.Sprintf("%s%s", name, direction)
	}

	return "Custom"
}

func (m AppModel) getActiveFiltersDescription() string {
	if len(m.activeFilters) == 0 && !m.searchActive {
		return "None"
	}

	var filters []string

	filterNames := map[string]string{
		"engineering": "Eng",
		"marketing":   "Mkt",
		"sales":       "Sales",
		"finance":     "Fin",
		"hr":          "HR",
		"active_only": "Active",
		"remote_only": "Remote",
		"high_salary": "High$",
		"low_salary":  "Low$",
	}

	for key, active := range m.activeFilters {
		if active {
			if name, exists := filterNames[key]; exists {
				filters = append(filters, name)
			}
		}
	}

	if m.searchActive {
		filters = append(filters, "Search")
	}

	if len(filters) == 0 {
		return "None"
	}

	return strings.Join(filters, "+")
}
