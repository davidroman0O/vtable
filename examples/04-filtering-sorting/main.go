package main

import (
	"fmt"
	"sort"
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

// Custom message to go back to menu
type BackToMenuMsg struct{}

// Main application model that manages different states
type AppModel struct {
	state      AppState
	menuModel  *MenuModel
	listModel  *ListFilterSortModel
	tableModel *TableFilterSortModel
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
			case 0: // List Filtering/Sorting Demo
				m.state = StateListDemo
				m.listModel = newListFilterSortModel()
				return m, m.listModel.Init()
			case 1: // Table Filtering/Sorting Demo
				m.state = StateTableDemo
				m.tableModel = newTableFilterSortModel()
				return m, m.tableModel.Init()
			}
		}
		return m, cmd

	case StateListDemo:
		newListModel, cmd := m.listModel.Update(msg)
		m.listModel = newListModel.(*ListFilterSortModel)
		return m, cmd

	case StateTableDemo:
		newTableModel, cmd := m.tableModel.Update(msg)
		m.tableModel = newTableModel.(*TableFilterSortModel)
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
			"List Demo - Filter tasks by priority, sort by status/priority",
			"Table Demo - Filter employees by department/salary, sort by multiple columns",
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
	s := "VTable Example 04: Filtering & Sorting\n\n"
	s += "Learn VTable filtering and sorting - search, filter, and sort your data!\n\n"
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

// ===== LIST FILTERING/SORTING DEMO =====

type Task struct {
	ID       int
	Title    string
	Priority string // High, Medium, Low
	Status   string // Todo, In Progress, Done
	Assignee string
}

// Task provider with filtering and sorting
type TaskProvider struct {
	tasks        []Task
	filteredData []Task
	filters      map[string]any
	sortField    string
	sortDir      string
	dirty        bool
}

func NewTaskProvider() *TaskProvider {
	return &TaskProvider{
		tasks: []Task{
			{1, "Implement user authentication", "High", "In Progress", "Alice"},
			{2, "Fix login bug", "High", "Todo", "Bob"},
			{3, "Write documentation", "Low", "Todo", "Carol"},
			{4, "Code review meeting", "Medium", "Done", "David"},
			{5, "Deploy to staging", "High", "Done", "Alice"},
			{6, "Update dependencies", "Low", "In Progress", "Eve"},
			{7, "Performance testing", "Medium", "Todo", "Frank"},
			{8, "Database migration", "High", "In Progress", "Bob"},
			{9, "UI improvements", "Low", "Done", "Carol"},
			{10, "Security audit", "High", "Todo", "David"},
		},
		filters: make(map[string]any),
		dirty:   true,
	}
}

func (p *TaskProvider) ensureFilteredData() {
	if !p.dirty && p.filteredData != nil {
		return
	}

	// Apply filters
	filtered := make([]Task, 0, len(p.tasks))
	for _, task := range p.tasks {
		if p.matchesFilters(task) {
			filtered = append(filtered, task)
		}
	}

	// Apply sorting
	if p.sortField != "" {
		p.sortTasks(filtered)
	}

	p.filteredData = filtered
	p.dirty = false
}

func (p *TaskProvider) matchesFilters(task Task) bool {
	for key, value := range p.filters {
		switch key {
		case "priority":
			if strVal, ok := value.(string); ok && !strings.EqualFold(task.Priority, strVal) {
				return false
			}
		case "status":
			if strVal, ok := value.(string); ok && !strings.EqualFold(task.Status, strVal) {
				return false
			}
		case "assignee":
			if strVal, ok := value.(string); ok && !strings.Contains(strings.ToLower(task.Assignee), strings.ToLower(strVal)) {
				return false
			}
		case "title":
			if strVal, ok := value.(string); ok && !strings.Contains(strings.ToLower(task.Title), strings.ToLower(strVal)) {
				return false
			}
		}
	}
	return true
}

func (p *TaskProvider) sortTasks(tasks []Task) {
	if p.sortField == "" {
		return
	}

	// Simple bubble sort
	for i := 0; i < len(tasks)-1; i++ {
		for j := 0; j < len(tasks)-i-1; j++ {
			if p.compareTasks(tasks[j], tasks[j+1]) > 0 {
				tasks[j], tasks[j+1] = tasks[j+1], tasks[j]
			}
		}
	}
}

func (p *TaskProvider) compareTasks(a, b Task) int {
	ascending := p.sortDir != "desc"
	var comparison int

	switch p.sortField {
	case "priority":
		priorityOrder := map[string]int{"High": 3, "Medium": 2, "Low": 1}
		comparison = priorityOrder[a.Priority] - priorityOrder[b.Priority]
	case "status":
		comparison = strings.Compare(a.Status, b.Status)
	case "title":
		comparison = strings.Compare(a.Title, b.Title)
	case "assignee":
		comparison = strings.Compare(a.Assignee, b.Assignee)
	default:
		comparison = 0
	}

	if !ascending {
		comparison = -comparison
	}
	return comparison
}

func (p *TaskProvider) GetTotal() int {
	p.ensureFilteredData()
	return len(p.filteredData)
}

func (p *TaskProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[string], error) {
	// Update filters from request to keep in sync
	p.filters = make(map[string]any)
	for k, v := range request.Filters {
		p.filters[k] = v
	}
	p.dirty = true

	p.ensureFilteredData()

	start := request.Start
	count := request.Count

	if start >= len(p.filteredData) {
		return []vtable.Data[string]{}, nil
	}

	if start+count > len(p.filteredData) {
		count = len(p.filteredData) - start
	}

	result := make([]vtable.Data[string], count)
	for i := 0; i < count; i++ {
		task := p.filteredData[start+i]
		display := fmt.Sprintf("[%s] %s - %s (%s)", task.Priority, task.Title, task.Status, task.Assignee)

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
	return vtable.SelectionNone
}

func (p *TaskProvider) SetSelected(index int, selected bool) bool {
	return false
}

func (p *TaskProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	return false
}

func (p *TaskProvider) SelectRange(startID, endID string) bool {
	return false
}

func (p *TaskProvider) SelectAll() bool {
	return false
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

// List model for filtering and sorting
type ListFilterSortModel struct {
	list          *vtable.TeaList[string]
	provider      *TaskProvider
	status        string
	activeFilters map[string]any
}

func newListFilterSortModel() *ListFilterSortModel {
	provider := NewTaskProvider()

	config := vtable.DefaultViewportConfig()
	config.Height = 10
	config.TopThresholdIndex = 0
	config.BottomThresholdIndex = 8

	style := vtable.DefaultStyleConfig()

	formatter := func(data vtable.Data[string], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return fmt.Sprintf("%s%s", prefix, data.Item)
	}

	list, err := vtable.NewTeaList(config, provider, style, formatter)
	if err != nil {
		panic(err)
	}

	return &ListFilterSortModel{
		list:          list,
		provider:      provider,
		status:        "Use 1-3 to filter, s/S to sort, r to reset",
		activeFilters: make(map[string]any),
	}
}

func (m *ListFilterSortModel) Init() tea.Cmd {
	return m.list.Init()
}

func (m *ListFilterSortModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }

		// Filtering
		case "1":
			// Filter by High priority
			m.toggleFilter("priority", "High")
			return m, nil
		case "2":
			// Filter by In Progress status
			m.toggleFilter("status", "In Progress")
			return m, nil
		case "3":
			// Filter by assignee containing "a"
			m.toggleFilter("assignee", "a")
			return m, nil

		// Sorting
		case "s":
			// Sort by priority
			m.toggleSort("priority")
			return m, nil
		case "S":
			// Sort by status
			m.toggleSort("status")
			return m, nil

		// Reset
		case "r":
			m.resetFiltersAndSort()
			return m, nil
		}
	}

	newList, cmd := m.list.Update(msg)
	m.list = newList.(*vtable.TeaList[string])
	return m, cmd
}

func (m *ListFilterSortModel) toggleFilter(field string, value any) {
	if existingValue, exists := m.activeFilters[field]; exists && existingValue == value {
		// Remove filter
		delete(m.activeFilters, field)
		m.status = fmt.Sprintf("Removed %s filter", field)
	} else {
		// Add/update filter
		m.activeFilters[field] = value
		m.status = fmt.Sprintf("Added %s filter: %v", field, value)
	}

	// Sync with list component - clear and re-add all filters
	m.list.ClearFilters()
	for k, v := range m.activeFilters {
		m.list.SetFilter(k, v)
	}

	// Force refresh
	m.list.SetDataProvider(m.provider)
}

func (m *ListFilterSortModel) toggleSort(field string) {
	if m.provider.sortField == field {
		// Toggle direction or remove
		if m.provider.sortDir == "asc" {
			m.provider.sortDir = "desc"
			m.status = fmt.Sprintf("Sorting by %s (descending)", field)
		} else {
			m.provider.sortField = ""
			m.provider.sortDir = ""
			m.status = fmt.Sprintf("Removed %s sort", field)
		}
	} else {
		// Set new sort
		m.provider.sortField = field
		m.provider.sortDir = "asc"
		m.status = fmt.Sprintf("Sorting by %s (ascending)", field)
	}
	m.provider.dirty = true

	// Force refresh
	m.list.SetDataProvider(m.provider)
}

func (m *ListFilterSortModel) resetFiltersAndSort() {
	m.activeFilters = make(map[string]any)
	m.provider.filters = make(map[string]any)
	m.provider.sortField = ""
	m.provider.sortDir = ""
	m.provider.dirty = true

	// Clear list state
	m.list.ClearFilters()
	m.list.SetDataProvider(m.provider)

	m.status = "Reset all filters and sorting"
}

// getSortedFilterDisplay returns a consistent display of active filters
func (m *ListFilterSortModel) getSortedFilterDisplay() string {
	if len(m.activeFilters) == 0 {
		return ""
	}

	// Sort keys for consistent display
	keys := make([]string, 0, len(m.activeFilters))
	for k := range m.activeFilters {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, m.activeFilters[k]))
	}

	return fmt.Sprintf(" | Filters: %s", strings.Join(parts, ", "))
}

func (m *ListFilterSortModel) View() string {
	filterInfo := m.getSortedFilterDisplay()

	sortInfo := ""
	if m.provider.sortField != "" {
		sortInfo = fmt.Sprintf(" | Sort: %s (%s)", m.provider.sortField, m.provider.sortDir)
	}

	return fmt.Sprintf("VTable Example 04: Filtering & Sorting - List Demo\n\n%s\n\n%s%s%s\n\n1=High priority, 2=In Progress, 3=Has 'a', s=Sort priority, S=Sort status, r=Reset, q/ESC=Back",
		m.list.View(), m.status, filterInfo, sortInfo)
}

// ===== TABLE FILTERING/SORTING DEMO =====

type Employee struct {
	ID         int
	Name       string
	Department string
	Salary     int
	Experience int
}

// Employee provider with filtering and sorting
type EmployeeProvider struct {
	employees    []Employee
	filteredData []Employee
	filters      map[string]any
	sortFields   []string
	sortDirs     []string
	dirty        bool
}

func NewEmployeeProvider() *EmployeeProvider {
	return &EmployeeProvider{
		employees: []Employee{
			{1, "Alice Johnson", "Engineering", 95000, 5},
			{2, "Bob Smith", "Marketing", 65000, 3},
			{3, "Carol Davis", "Engineering", 87000, 4},
			{4, "David Wilson", "Sales", 72000, 6},
			{5, "Eve Brown", "HR", 58000, 2},
			{6, "Frank Miller", "Engineering", 102000, 8},
			{7, "Grace Lee", "Marketing", 69000, 3},
			{8, "Henry Taylor", "Sales", 76000, 5},
			{9, "Ivy Chen", "Engineering", 93000, 6},
			{10, "Jack Adams", "HR", 61000, 4},
			{11, "Kate Wilson", "Engineering", 89000, 4},
			{12, "Liam Johnson", "Sales", 78000, 7},
			{13, "Mia Brown", "Marketing", 71000, 5},
			{14, "Noah Davis", "HR", 63000, 3},
			{15, "Olivia Miller", "Engineering", 98000, 7},
		},
		filters:    make(map[string]any),
		sortFields: []string{},
		sortDirs:   []string{},
		dirty:      true,
	}
}

func (p *EmployeeProvider) ensureFilteredData() {
	if !p.dirty && p.filteredData != nil {
		return
	}

	// Apply filters
	filtered := make([]Employee, 0, len(p.employees))
	for _, emp := range p.employees {
		if p.matchesFilters(emp) {
			filtered = append(filtered, emp)
		}
	}

	// Apply sorting
	if len(p.sortFields) > 0 {
		p.sortEmployees(filtered)
	}

	p.filteredData = filtered
	p.dirty = false
}

func (p *EmployeeProvider) matchesFilters(emp Employee) bool {
	for key, value := range p.filters {
		switch key {
		case "department":
			if strVal, ok := value.(string); ok && !strings.EqualFold(emp.Department, strVal) {
				return false
			}
		case "minSalary":
			if intVal, ok := value.(int); ok && emp.Salary < intVal {
				return false
			}
		case "minExperience":
			if intVal, ok := value.(int); ok && emp.Experience < intVal {
				return false
			}
		case "name":
			if strVal, ok := value.(string); ok && !strings.Contains(strings.ToLower(emp.Name), strings.ToLower(strVal)) {
				return false
			}
		}
	}
	return true
}

func (p *EmployeeProvider) sortEmployees(employees []Employee) {
	if len(p.sortFields) == 0 {
		return
	}

	// Simple bubble sort with multi-field support
	for i := 0; i < len(employees)-1; i++ {
		for j := 0; j < len(employees)-i-1; j++ {
			if p.compareEmployees(employees[j], employees[j+1]) > 0 {
				employees[j], employees[j+1] = employees[j+1], employees[j]
			}
		}
	}
}

func (p *EmployeeProvider) compareEmployees(a, b Employee) int {
	for i, field := range p.sortFields {
		ascending := p.sortDirs[i] != "desc"
		var comparison int

		switch field {
		case "name":
			comparison = strings.Compare(a.Name, b.Name)
		case "department":
			comparison = strings.Compare(a.Department, b.Department)
		case "salary":
			if a.Salary < b.Salary {
				comparison = -1
			} else if a.Salary > b.Salary {
				comparison = 1
			}
		case "experience":
			if a.Experience < b.Experience {
				comparison = -1
			} else if a.Experience > b.Experience {
				comparison = 1
			}
		}

		if comparison != 0 {
			if !ascending {
				comparison = -comparison
			}
			return comparison
		}
	}
	return 0
}

func (p *EmployeeProvider) GetTotal() int {
	p.ensureFilteredData()
	return len(p.filteredData)
}

func (p *EmployeeProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[vtable.TableRow], error) {
	// Update filters from request to keep in sync
	p.filters = make(map[string]any)
	for k, v := range request.Filters {
		p.filters[k] = v
	}

	// Update sorts from request to keep in sync
	p.sortFields = make([]string, len(request.SortFields))
	copy(p.sortFields, request.SortFields)
	p.sortDirs = make([]string, len(request.SortDirections))
	copy(p.sortDirs, request.SortDirections)

	p.dirty = true
	p.ensureFilteredData()

	start := request.Start
	count := request.Count

	if start >= len(p.filteredData) {
		return []vtable.Data[vtable.TableRow]{}, nil
	}

	if start+count > len(p.filteredData) {
		count = len(p.filteredData) - start
	}

	result := make([]vtable.Data[vtable.TableRow], count)
	for i := 0; i < count; i++ {
		emp := p.filteredData[start+i]

		row := vtable.TableRow{
			Cells: []string{
				fmt.Sprintf("%d", emp.ID),
				emp.Name,
				emp.Department,
				fmt.Sprintf("$%d", emp.Salary),
				fmt.Sprintf("%d yrs", emp.Experience),
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
	return vtable.SelectionNone
}

func (p *EmployeeProvider) SetSelected(index int, selected bool) bool {
	return false
}

func (p *EmployeeProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	return false
}

func (p *EmployeeProvider) SelectRange(startID, endID string) bool {
	return false
}

func (p *EmployeeProvider) SelectAll() bool {
	return false
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

// Table model for filtering and sorting
type TableFilterSortModel struct {
	table         *vtable.TeaTable
	provider      *EmployeeProvider
	status        string
	activeFilters map[string]any
	activeSorts   map[string]string // field -> direction
}

func newTableFilterSortModel() *TableFilterSortModel {
	provider := NewEmployeeProvider()

	config := vtable.TableConfig{
		Columns: []vtable.TableColumn{
			{Title: "ID", Width: 8, Alignment: vtable.AlignRight},
			{Title: "Name", Width: 15, Alignment: vtable.AlignLeft},
			{Title: "Department", Width: 12, Alignment: vtable.AlignLeft},
			{Title: "Salary", Width: 10, Alignment: vtable.AlignRight},
			{Title: "Experience", Width: 12, Alignment: vtable.AlignRight},
		},
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: vtable.ViewportConfig{
			Height:               10,
			TopThresholdIndex:    0,
			BottomThresholdIndex: 8,
			ChunkSize:            20,
			InitialIndex:         0,
		},
	}

	theme := vtable.DefaultTheme()
	table, err := vtable.NewTeaTable(config, provider, *theme)
	if err != nil {
		panic(err)
	}

	return &TableFilterSortModel{
		table:         table,
		provider:      provider,
		status:        "Use 1-4 to filter, Shift+1-5 to sort, r to reset",
		activeFilters: make(map[string]any),
		activeSorts:   make(map[string]string),
	}
}

func (m *TableFilterSortModel) Init() tea.Cmd {
	return m.table.Init()
}

func (m *TableFilterSortModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }

		// Filtering
		case "1":
			// Filter by Engineering department
			m.toggleFilter("department", "Engineering")
			return m, nil
		case "2":
			// Filter by salary >= 80000
			m.toggleFilter("minSalary", 80000)
			return m, nil
		case "3":
			// Filter by experience >= 5 years
			m.toggleFilter("minExperience", 5)
			return m, nil
		case "4":
			// Filter by name containing "a"
			m.toggleFilter("name", "a")
			return m, nil

		// Sorting (Shift + number)
		case "!": // Shift+1 - Sort by name
			m.toggleSort("name")
			return m, nil
		case "@": // Shift+2 - Sort by department
			m.toggleSort("department")
			return m, nil
		case "#": // Shift+3 - Sort by salary
			m.toggleSort("salary")
			return m, nil
		case "$": // Shift+4 - Sort by experience
			m.toggleSort("experience")
			return m, nil

		// Reset
		case "r":
			m.resetFiltersAndSort()
			return m, nil
		}
	}

	newTable, cmd := m.table.Update(msg)
	m.table = newTable.(*vtable.TeaTable)
	return m, cmd
}

func (m *TableFilterSortModel) toggleFilter(field string, value any) {
	if existingValue, exists := m.activeFilters[field]; exists && existingValue == value {
		// Remove filter
		delete(m.activeFilters, field)
		m.status = fmt.Sprintf("Removed %s filter", field)
	} else {
		// Add/update filter
		m.activeFilters[field] = value
		m.status = fmt.Sprintf("Added %s filter: %v", field, value)
	}

	// Sync with table component - clear and re-add all filters
	m.table.ClearFilters()
	for k, v := range m.activeFilters {
		m.table.SetFilter(k, v)
	}

	// Force refresh
	m.table.SetDataProvider(m.provider)
}

func (m *TableFilterSortModel) toggleSort(field string) {
	if direction, exists := m.activeSorts[field]; exists {
		// Toggle direction or remove
		if direction == "asc" {
			m.activeSorts[field] = "desc"
			m.status = fmt.Sprintf("Sorting by %s (descending)", field)
		} else {
			delete(m.activeSorts, field)
			m.status = fmt.Sprintf("Removed %s sort", field)
		}
	} else {
		// Add new sort field
		m.activeSorts[field] = "asc"
		m.status = fmt.Sprintf("Sorting by %s (ascending)", field)
	}

	// Sync with table component - clear and re-add all sorts
	m.table.ClearSort()

	// Sort keys for consistent ordering
	keys := make([]string, 0, len(m.activeSorts))
	for k := range m.activeSorts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		m.table.AddSort(k, m.activeSorts[k])
	}

	// Force refresh
	m.table.SetDataProvider(m.provider)
}

func (m *TableFilterSortModel) resetFiltersAndSort() {
	m.activeFilters = make(map[string]any)
	m.activeSorts = make(map[string]string)

	// Clear table state
	m.table.ClearFilters()
	m.table.ClearSort()
	m.table.SetDataProvider(m.provider)

	m.status = "Reset all filters and sorting"
}

// getSortedFilterDisplay returns a consistent display of active filters
func (m *TableFilterSortModel) getSortedFilterDisplay() string {
	if len(m.activeFilters) == 0 {
		return ""
	}

	// Sort keys for consistent display
	keys := make([]string, 0, len(m.activeFilters))
	for k := range m.activeFilters {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, m.activeFilters[k]))
	}

	return fmt.Sprintf(" | Filters: %s", strings.Join(parts, ", "))
}

// getSortedSortDisplay returns a consistent display of active sorts
func (m *TableFilterSortModel) getSortedSortDisplay() string {
	if len(m.activeSorts) == 0 {
		return ""
	}

	// Sort keys for consistent display
	keys := make([]string, 0, len(m.activeSorts))
	for k := range m.activeSorts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s (%s)", k, m.activeSorts[k]))
	}

	return fmt.Sprintf(" | Sort: %s", strings.Join(parts, ", "))
}

func (m *TableFilterSortModel) View() string {
	filterInfo := m.getSortedFilterDisplay()
	sortInfo := m.getSortedSortDisplay()

	return fmt.Sprintf("VTable Example 04: Filtering & Sorting - Table Demo\n\n%s\n\n%s%s%s\n\n1=Engineering, 2=Salary≥80k, 3=Exp≥5yrs, 4=Has 'a' | Shift+1-4=Sort | r=Reset, q/ESC=Back",
		m.table.View(), m.status, filterInfo, sortInfo)
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
