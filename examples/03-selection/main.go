package main

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman0O/vtable"
)

// Application states
type AppState int

const (
	StateMenu AppState = iota
	StateListSingle
	StateListMultiple
	StateTableSingle
	StateTableMultiple
)

// Custom message to go back to menu
type BackToMenuMsg struct{}

// Main application model that manages different states
type AppModel struct {
	state              AppState
	menuModel          *MenuModel
	listSingleModel    *ListSingleModel
	listMultipleModel  *ListMultipleModel
	tableSingleModel   *TableSingleModel
	tableMultipleModel *TableMultipleModel
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
			case 0: // List Single Demo
				m.state = StateListSingle
				m.listSingleModel = newListSingleModel()
				return m, m.listSingleModel.Init()
			case 1: // List Multiple Demo
				m.state = StateListMultiple
				m.listMultipleModel = newListMultipleModel()
				return m, m.listMultipleModel.Init()
			case 2: // Table Single Demo
				m.state = StateTableSingle
				m.tableSingleModel = newTableSingleModel()
				return m, m.tableSingleModel.Init()
			case 3: // Table Multiple Demo
				m.state = StateTableMultiple
				m.tableMultipleModel = newTableMultipleModel()
				return m, m.tableMultipleModel.Init()
			}
		}
		return m, cmd

	case StateListSingle:
		newListSingleModel, cmd := m.listSingleModel.Update(msg)
		m.listSingleModel = newListSingleModel.(*ListSingleModel)
		return m, cmd

	case StateListMultiple:
		newListMultipleModel, cmd := m.listMultipleModel.Update(msg)
		m.listMultipleModel = newListMultipleModel.(*ListMultipleModel)
		return m, cmd

	case StateTableSingle:
		newTableSingleModel, cmd := m.tableSingleModel.Update(msg)
		m.tableSingleModel = newTableSingleModel.(*TableSingleModel)
		return m, cmd

	case StateTableMultiple:
		newTableMultipleModel, cmd := m.tableMultipleModel.Update(msg)
		m.tableMultipleModel = newTableMultipleModel.(*TableMultipleModel)
		return m, cmd
	}

	return m, nil
}

func (m *AppModel) View() string {
	switch m.state {
	case StateMenu:
		return m.menuModel.View()
	case StateListSingle:
		return m.listSingleModel.View()
	case StateListMultiple:
		return m.listMultipleModel.View()
	case StateTableSingle:
		return m.tableSingleModel.View()
	case StateTableMultiple:
		return m.tableMultipleModel.View()
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
			"List Single Demo - Single selection with space bar",
			"List Multiple Demo - Multi-selection with space and Ctrl+A",
			"Table Single Demo - Single selection with space bar",
			"Table Multiple Demo - Multi-selection with space and Ctrl+A",
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
	s := "VTable Example 03: Selection\n\n"
	s += "Learn VTable selection modes - single, multi, and range selection!\n\n"
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

// ===== LIST SINGLE DEMO =====

// Selectable task provider for SINGLE selection
type TaskSingleProvider struct {
	tasks    []string
	selected map[int]bool
}

func NewTaskSingleProvider() *TaskSingleProvider {
	return &TaskSingleProvider{
		tasks: []string{
			"Review pull request #42",
			"Fix bug in authentication",
			"Write documentation",
			"Update dependencies",
			"Run performance tests",
			"Deploy to staging",
			"Code review meeting",
			"Plan next sprint",
		},
		selected: make(map[int]bool),
	}
}

func (p *TaskSingleProvider) GetTotal() int {
	return len(p.tasks)
}

func (p *TaskSingleProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[string], error) {
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
		index := start + i
		result[i] = vtable.Data[string]{
			ID:       fmt.Sprintf("task-%d", index),
			Item:     p.tasks[index],
			Selected: p.selected[index],
			Metadata: vtable.NewTypedMetadata(),
		}
	}
	return result, nil
}

// SINGLE selection methods
func (p *TaskSingleProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionSingle
}

func (p *TaskSingleProvider) SetSelected(index int, selected bool) bool {
	if index < 0 || index >= len(p.tasks) {
		return false
	}
	if selected {
		// Single selection - clear others
		for k := range p.selected {
			p.selected[k] = false
		}
		p.selected[index] = true
	} else {
		p.selected[index] = false
	}
	return true
}

func (p *TaskSingleProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	for _, id := range ids {
		if index, err := strconv.Atoi(id[5:]); err == nil { // Remove "task-" prefix
			p.SetSelected(index, selected)
		}
	}
	return true
}

func (p *TaskSingleProvider) SelectRange(startID, endID string) bool { return false }
func (p *TaskSingleProvider) SelectAll() bool                        { return false }
func (p *TaskSingleProvider) ClearSelection() {
	for k := range p.selected {
		p.selected[k] = false
	}
}

func (p *TaskSingleProvider) GetSelectedIndices() []int {
	var indices []int
	for i, sel := range p.selected {
		if sel {
			indices = append(indices, i)
		}
	}
	return indices
}

func (p *TaskSingleProvider) GetSelectedIDs() []string {
	var ids []string
	for i, sel := range p.selected {
		if sel {
			ids = append(ids, fmt.Sprintf("task-%d", i))
		}
	}
	return ids
}

func (p *TaskSingleProvider) GetItemID(item *string) string {
	for i, task := range p.tasks {
		if task == *item {
			return fmt.Sprintf("task-%d", i)
		}
	}
	return ""
}

// List model for SINGLE selection
type ListSingleModel struct {
	list     *vtable.TeaList[string]
	provider *TaskSingleProvider
}

func newListSingleModel() *ListSingleModel {
	provider := NewTaskSingleProvider()

	formatter := func(data vtable.Data[string], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		var prefix string
		if data.Selected && isCursor {
			prefix = "✓>"
		} else if data.Selected {
			prefix = "✓ "
		} else if isCursor {
			prefix = "> "
		} else {
			prefix = "  "
		}
		return fmt.Sprintf("%s %s", prefix, data.Item)
	}

	list, err := vtable.NewTeaListWithHeight(provider, formatter, 8)
	if err != nil {
		panic(err)
	}

	return &ListSingleModel{list: list, provider: provider}
}

func (m *ListSingleModel) Init() tea.Cmd {
	return m.list.Init()
}

func (m *ListSingleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }
		case " ", "space":
			// Handle selection BEFORE component update to prevent PageDown behavior
			m.list.ToggleCurrentSelection()
			// Return early to prevent component from processing space as PageDown
			return m, nil
		}
	}

	newList, cmd := m.list.Update(msg)
	m.list = newList.(*vtable.TeaList[string])
	return m, cmd
}

func (m *ListSingleModel) View() string {
	selectedCount := len(m.provider.GetSelectedIndices())
	return fmt.Sprintf("VTable Example 03: Selection - List SINGLE Selection Demo\n\n%s\n\nSelected: %d task(s) | Space to toggle (single only), ↑/↓ to navigate, q/ESC to go back",
		m.list.View(), selectedCount)
}

// ===== LIST MULTIPLE DEMO =====

// Selectable task provider for MULTIPLE selection
type TaskMultipleProvider struct {
	tasks    []string
	selected map[int]bool
}

func NewTaskMultipleProvider() *TaskMultipleProvider {
	return &TaskMultipleProvider{
		tasks: []string{
			"Review pull request #42",
			"Fix bug in authentication",
			"Write documentation",
			"Update dependencies",
			"Run performance tests",
			"Deploy to staging",
			"Code review meeting",
			"Plan next sprint",
		},
		selected: make(map[int]bool),
	}
}

func (p *TaskMultipleProvider) GetTotal() int {
	return len(p.tasks)
}

func (p *TaskMultipleProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[string], error) {
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
		index := start + i
		result[i] = vtable.Data[string]{
			ID:       fmt.Sprintf("task-%d", index),
			Item:     p.tasks[index],
			Selected: p.selected[index],
			Metadata: vtable.NewTypedMetadata(),
		}
	}
	return result, nil
}

// MULTIPLE selection methods
func (p *TaskMultipleProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionMultiple
}

func (p *TaskMultipleProvider) SetSelected(index int, selected bool) bool {
	if index < 0 || index >= len(p.tasks) {
		return false
	}
	p.selected[index] = selected
	return true
}

func (p *TaskMultipleProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	for _, id := range ids {
		if index, err := strconv.Atoi(id[5:]); err == nil { // Remove "task-" prefix
			p.SetSelected(index, selected)
		}
	}
	return true
}

func (p *TaskMultipleProvider) SelectRange(startID, endID string) bool {
	startIndex, err1 := strconv.Atoi(startID[5:])
	endIndex, err2 := strconv.Atoi(endID[5:])

	if err1 != nil || err2 != nil {
		return false
	}

	if startIndex > endIndex {
		startIndex, endIndex = endIndex, startIndex
	}

	for i := startIndex; i <= endIndex; i++ {
		p.SetSelected(i, true)
	}
	return true
}

func (p *TaskMultipleProvider) SelectAll() bool {
	for i := 0; i < len(p.tasks); i++ {
		p.selected[i] = true
	}
	return true
}

func (p *TaskMultipleProvider) ClearSelection() {
	for k := range p.selected {
		p.selected[k] = false
	}
}

func (p *TaskMultipleProvider) GetSelectedIndices() []int {
	var indices []int
	for i, sel := range p.selected {
		if sel {
			indices = append(indices, i)
		}
	}
	return indices
}

func (p *TaskMultipleProvider) GetSelectedIDs() []string {
	var ids []string
	for i, sel := range p.selected {
		if sel {
			ids = append(ids, fmt.Sprintf("task-%d", i))
		}
	}
	return ids
}

func (p *TaskMultipleProvider) GetItemID(item *string) string {
	for i, task := range p.tasks {
		if task == *item {
			return fmt.Sprintf("task-%d", i)
		}
	}
	return ""
}

// List model for MULTIPLE selection
type ListMultipleModel struct {
	list     *vtable.TeaList[string]
	provider *TaskMultipleProvider
}

func newListMultipleModel() *ListMultipleModel {
	provider := NewTaskMultipleProvider()

	formatter := func(data vtable.Data[string], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		prefix := "  "
		if isCursor {
			prefix = "> "
		}

		return fmt.Sprintf("%s%s", prefix, data.Item)
	}

	list, err := vtable.NewTeaListWithHeight(provider, formatter, 8)
	if err != nil {
		panic(err)
	}

	return &ListMultipleModel{list: list, provider: provider}
}

func (m *ListMultipleModel) Init() tea.Cmd {
	return m.list.Init()
}

func (m *ListMultipleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }
		case " ", "space":
			// Handle selection BEFORE component update to prevent PageDown behavior
			m.list.ToggleCurrentSelection()
			// Return early to prevent component from processing space as PageDown
			return m, nil
		case "ctrl+a":
			// Handle select all BEFORE component update
			m.list.SelectAll()
			return m, nil
		}
	}

	newList, cmd := m.list.Update(msg)
	m.list = newList.(*vtable.TeaList[string])
	return m, cmd
}

func (m *ListMultipleModel) View() string {
	selectedCount := len(m.provider.GetSelectedIndices())
	return fmt.Sprintf("VTable Example 03: Selection - List MULTIPLE Selection Demo\n\n%s\n\nSelected: %d task(s) | Space to toggle, Ctrl+A for all, ↑/↓ to navigate, q/ESC to go back",
		m.list.View(), selectedCount)
}

// ===== TABLE SINGLE DEMO =====

type Employee struct {
	ID         int
	Name       string
	Department string
	Salary     int
}

// Single-selectable employee provider
type EmployeeSingleProvider struct {
	employees []Employee
	selected  map[int]bool
}

func NewEmployeeSingleProvider() *EmployeeSingleProvider {
	return &EmployeeSingleProvider{
		employees: []Employee{
			{1, "Alice Johnson", "Engineering", 95000},
			{2, "Bob Smith", "Marketing", 65000},
			{3, "Carol Davis", "Engineering", 87000},
			{4, "David Wilson", "Sales", 72000},
			{5, "Eve Brown", "HR", 58000},
			{6, "Frank Miller", "Engineering", 102000},
			{7, "Grace Lee", "Marketing", 69000},
			{8, "Henry Taylor", "Sales", 76000},
			{9, "Ivy Chen", "Engineering", 93000},
			{10, "Jack Adams", "HR", 61000},
		},
		selected: make(map[int]bool),
	}
}

func (p *EmployeeSingleProvider) GetTotal() int {
	return len(p.employees)
}

func (p *EmployeeSingleProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[vtable.TableRow], error) {
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
		index := start + i
		emp := p.employees[index]

		row := vtable.TableRow{
			Cells: []string{
				fmt.Sprintf("%d", emp.ID),
				emp.Name,
				emp.Department,
				fmt.Sprintf("$%d", emp.Salary),
			},
		}

		result[i] = vtable.Data[vtable.TableRow]{
			ID:       fmt.Sprintf("emp-%d", index),
			Item:     row,
			Selected: p.selected[index],
			Metadata: vtable.NewTypedMetadata(),
		}
	}
	return result, nil
}

// SINGLE selection methods
func (p *EmployeeSingleProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionSingle
}

func (p *EmployeeSingleProvider) SetSelected(index int, selected bool) bool {
	if index < 0 || index >= len(p.employees) {
		return false
	}
	if selected {
		// Single selection - clear others
		for k := range p.selected {
			p.selected[k] = false
		}
		p.selected[index] = true
	} else {
		p.selected[index] = false
	}
	return true
}

func (p *EmployeeSingleProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	for _, id := range ids {
		if index, err := strconv.Atoi(id[4:]); err == nil { // Remove "emp-" prefix
			p.SetSelected(index, selected)
		}
	}
	return true
}

func (p *EmployeeSingleProvider) SelectRange(startID, endID string) bool { return false }
func (p *EmployeeSingleProvider) SelectAll() bool                        { return false }
func (p *EmployeeSingleProvider) ClearSelection() {
	for k := range p.selected {
		p.selected[k] = false
	}
}

func (p *EmployeeSingleProvider) GetSelectedIndices() []int {
	var indices []int
	for i, sel := range p.selected {
		if sel {
			indices = append(indices, i)
		}
	}
	return indices
}

func (p *EmployeeSingleProvider) GetSelectedIDs() []string {
	var ids []string
	for i, sel := range p.selected {
		if sel {
			ids = append(ids, fmt.Sprintf("emp-%d", i))
		}
	}
	return ids
}

func (p *EmployeeSingleProvider) GetItemID(item *vtable.TableRow) string {
	return item.Cells[0] // Use ID column
}

// Table model for SINGLE selection
type TableSingleModel struct {
	table    *vtable.TeaTable
	provider *EmployeeSingleProvider
}

func newTableSingleModel() *TableSingleModel {
	provider := NewEmployeeSingleProvider()

	columns := []vtable.TableColumn{
		vtable.NewRightColumn("ID", 4),
		vtable.NewColumn("Name", 15),
		vtable.NewColumn("Department", 12),
		vtable.NewRightColumn("Salary", 10),
	}

	table, err := vtable.NewTeaTableWithHeight(columns, provider, 8)
	if err != nil {
		panic(err)
	}

	return &TableSingleModel{table: table, provider: provider}
}

func (m *TableSingleModel) Init() tea.Cmd {
	return m.table.Init()
}

func (m *TableSingleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }
		case " ", "space":
			// Handle selection BEFORE component update to prevent PageDown behavior
			m.table.ToggleCurrentSelection()
			// Return early to prevent component from processing space as PageDown
			return m, nil
		}
	}

	newTable, cmd := m.table.Update(msg)
	m.table = newTable.(*vtable.TeaTable)
	return m, cmd
}

func (m *TableSingleModel) View() string {
	selectedCount := len(m.provider.GetSelectedIndices())
	return fmt.Sprintf("VTable Example 03: Selection - Table SINGLE Selection Demo\n\n%s\n\nSelected: %d employee(s) | Space to toggle (single only), ↑/↓ to navigate, q/ESC to go back",
		m.table.View(), selectedCount)
}

// ===== TABLE MULTIPLE DEMO =====

// Multi-selectable employee provider
type EmployeeMultipleProvider struct {
	employees []Employee
	selected  map[int]bool
}

func NewEmployeeMultipleProvider() *EmployeeMultipleProvider {
	return &EmployeeMultipleProvider{
		employees: []Employee{
			{1, "Alice Johnson", "Engineering", 95000},
			{2, "Bob Smith", "Marketing", 65000},
			{3, "Carol Davis", "Engineering", 87000},
			{4, "David Wilson", "Sales", 72000},
			{5, "Eve Brown", "HR", 58000},
			{6, "Frank Miller", "Engineering", 102000},
			{7, "Grace Lee", "Marketing", 69000},
			{8, "Henry Taylor", "Sales", 76000},
			{9, "Ivy Chen", "Engineering", 93000},
			{10, "Jack Adams", "HR", 61000},
		},
		selected: make(map[int]bool),
	}
}

func (p *EmployeeMultipleProvider) GetTotal() int {
	return len(p.employees)
}

func (p *EmployeeMultipleProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[vtable.TableRow], error) {
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
		index := start + i
		emp := p.employees[index]

		row := vtable.TableRow{
			Cells: []string{
				fmt.Sprintf("%d", emp.ID),
				emp.Name,
				emp.Department,
				fmt.Sprintf("$%d", emp.Salary),
			},
		}

		result[i] = vtable.Data[vtable.TableRow]{
			ID:       fmt.Sprintf("emp-%d", index),
			Item:     row,
			Selected: p.selected[index],
			Metadata: vtable.NewTypedMetadata(),
		}
	}
	return result, nil
}

// MULTIPLE selection methods
func (p *EmployeeMultipleProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionMultiple
}

func (p *EmployeeMultipleProvider) SetSelected(index int, selected bool) bool {
	if index < 0 || index >= len(p.employees) {
		return false
	}
	p.selected[index] = selected
	return true
}

func (p *EmployeeMultipleProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	for _, id := range ids {
		if index, err := strconv.Atoi(id[4:]); err == nil { // Remove "emp-" prefix
			p.SetSelected(index, selected)
		}
	}
	return true
}

func (p *EmployeeMultipleProvider) SelectRange(startID, endID string) bool {
	startIndex, err1 := strconv.Atoi(startID[4:])
	endIndex, err2 := strconv.Atoi(endID[4:])

	if err1 != nil || err2 != nil {
		return false
	}

	if startIndex > endIndex {
		startIndex, endIndex = endIndex, startIndex
	}

	for i := startIndex; i <= endIndex; i++ {
		p.SetSelected(i, true)
	}
	return true
}

func (p *EmployeeMultipleProvider) SelectAll() bool {
	for i := 0; i < len(p.employees); i++ {
		p.selected[i] = true
	}
	return true
}

func (p *EmployeeMultipleProvider) ClearSelection() {
	for k := range p.selected {
		p.selected[k] = false
	}
}

func (p *EmployeeMultipleProvider) GetSelectedIndices() []int {
	var indices []int
	for i, sel := range p.selected {
		if sel {
			indices = append(indices, i)
		}
	}
	return indices
}

func (p *EmployeeMultipleProvider) GetSelectedIDs() []string {
	var ids []string
	for i, sel := range p.selected {
		if sel {
			ids = append(ids, fmt.Sprintf("emp-%d", i))
		}
	}
	return ids
}

func (p *EmployeeMultipleProvider) GetItemID(item *vtable.TableRow) string {
	return item.Cells[0] // Use ID column
}

// Table model for MULTIPLE selection
type TableMultipleModel struct {
	table    *vtable.TeaTable
	provider *EmployeeMultipleProvider
}

func newTableMultipleModel() *TableMultipleModel {
	provider := NewEmployeeMultipleProvider()

	columns := []vtable.TableColumn{
		vtable.NewRightColumn("ID", 8),
		vtable.NewColumn("Name", 15),
		vtable.NewColumn("Department", 12),
		vtable.NewRightColumn("Salary", 10),
	}

	table, err := vtable.NewTeaTableWithHeight(columns, provider, 8)
	if err != nil {
		panic(err)
	}

	return &TableMultipleModel{table: table, provider: provider}
}

func (m *TableMultipleModel) Init() tea.Cmd {
	return m.table.Init()
}

func (m *TableMultipleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }
		case " ", "space":
			// Handle selection BEFORE component update to prevent PageDown behavior
			m.table.ToggleCurrentSelection()
			// Return early to prevent component from processing space as PageDown
			return m, nil
		case "ctrl+a":
			// Handle select all BEFORE component update
			m.table.SelectAll()
			return m, nil
		}
	}

	newTable, cmd := m.table.Update(msg)
	m.table = newTable.(*vtable.TeaTable)
	return m, cmd
}

func (m *TableMultipleModel) View() string {
	selectedCount := len(m.provider.GetSelectedIndices())
	return fmt.Sprintf("VTable Example 03: Selection - Table MULTIPLE Selection Demo\n\n%s\n\nSelected: %d employee(s) | Space to toggle, Ctrl+A for all, ↑/↓ to navigate, q/ESC to go back",
		m.table.View(), selectedCount)
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
