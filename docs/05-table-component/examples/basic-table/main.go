package main

import (
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/table"
)

// EmployeeDataSource implements TableDataSource for employee data
type EmployeeDataSource struct {
	data []core.TableRow
}

func NewEmployeeDataSource() *EmployeeDataSource {
	return &EmployeeDataSource{
		data: createEmployeeData(),
	}
}

// Employee data for our table
func createEmployeeData() []core.TableRow {
	return []core.TableRow{
		{ID: "emp-1", Cells: []string{"Alice Johnson", "Engineering", "Active", "$75,000"}},
		{ID: "emp-2", Cells: []string{"Bob Smith", "Marketing", "Active", "$65,000"}},
		{ID: "emp-3", Cells: []string{"Carol Davis", "Engineering", "On Leave", "$80,000"}},
		{ID: "emp-4", Cells: []string{"David Wilson", "Sales", "Active", "$70,000"}},
		{ID: "emp-5", Cells: []string{"Eve Brown", "HR", "Active", "$60,000"}},
		{ID: "emp-6", Cells: []string{"Frank Miller", "Engineering", "Active", "$85,000"}},
		{ID: "emp-7", Cells: []string{"Grace Lee", "Marketing", "Active", "$62,000"}},
		{ID: "emp-8", Cells: []string{"Henry Taylor", "Sales", "On Leave", "$68,000"}},
		{ID: "emp-9", Cells: []string{"Ivy Chen", "Engineering", "Active", "$78,000"}},
		{ID: "emp-10", Cells: []string{"Jack Roberts", "Marketing", "Active", "$63,000"}},
		{ID: "emp-11", Cells: []string{"Kate Williams", "Sales", "Active", "$72,000"}},
		{ID: "emp-12", Cells: []string{"Leo Martinez", "HR", "On Leave", "$61,000"}},
	}
}

// GetTotal returns total number of employees
func (ds *EmployeeDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return core.DataTotalMsg{Total: len(ds.data)}
	}
}

// RefreshTotal refreshes the total count
func (ds *EmployeeDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

// LoadChunk loads a chunk of employee data
func (ds *EmployeeDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		start := request.Start
		end := start + request.Count
		if end > len(ds.data) {
			end = len(ds.data)
		}

		var items []core.Data[any]
		for i := start; i < end; i++ {
			items = append(items, core.Data[any]{
				ID:       ds.data[i].ID,
				Item:     ds.data[i],
				Selected: false, // No selection yet
				Metadata: core.NewTypedMetadata(),
			})
		}

		return core.DataChunkLoadedMsg{
			StartIndex: start,
			Items:      items,
			Request:    request,
		}
	}
}

// SetSelected sets the selection state of an item by index
func (ds *EmployeeDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		return core.SelectionResponseMsg{
			Success:  true,
			Index:    index,
			Selected: selected,
		}
	}
}

// SetSelectedByID sets the selection state of an item by ID
func (ds *EmployeeDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	return func() tea.Msg {
		return core.SelectionResponseMsg{
			Success:  true,
			ID:       id,
			Selected: selected,
		}
	}
}

// SelectAll selects all items
func (ds *EmployeeDataSource) SelectAll() tea.Cmd {
	return func() tea.Msg {
		return core.SelectionResponseMsg{
			Success:   true,
			Operation: "selectAll",
		}
	}
}

// ClearSelection clears all selections
func (ds *EmployeeDataSource) ClearSelection() tea.Cmd {
	return func() tea.Msg {
		return core.SelectionResponseMsg{
			Success:   true,
			Operation: "clear",
		}
	}
}

// SelectRange selects a range of items
func (ds *EmployeeDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	return func() tea.Msg {
		return core.SelectionResponseMsg{
			Success:   true,
			Operation: "range",
		}
	}
}

// GetItemID returns the ID for a table row
func (ds *EmployeeDataSource) GetItemID(item any) string {
	if row, ok := item.(core.TableRow); ok {
		return row.ID
	}
	return ""
}

func createEmployeeColumns() []core.TableColumn {
	return []core.TableColumn{
		{
			Title:           "Employee Name",
			Field:           "name",
			Width:           25,
			Alignment:       core.AlignLeft,
			HeaderAlignment: core.AlignLeft,
		},
		{
			Title:           "Department",
			Field:           "department",
			Width:           15,
			Alignment:       core.AlignCenter,
			HeaderAlignment: core.AlignCenter,
		},
		{
			Title:           "Status",
			Field:           "status",
			Width:           12,
			Alignment:       core.AlignCenter,
			HeaderAlignment: core.AlignCenter,
		},
		{
			Title:           "Salary",
			Field:           "salary",
			Width:           12,
			Alignment:       core.AlignRight,
			HeaderAlignment: core.AlignRight,
		},
	}
}

func createTableConfig() core.TableConfig {
	return core.TableConfig{
		Columns:     createEmployeeColumns(),
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: core.ViewportConfig{
			Height:             8,  // Show 8 rows at once
			TopThreshold:       2,  // Load new data when 2 rows from top
			BottomThreshold:    2,  // Load new data when 2 rows from bottom
			ChunkSize:          20, // Load 20 rows per chunk
			InitialIndex:       0,  // Start at first row
			BoundingAreaBefore: 10, // Keep 10 rows before viewport
			BoundingAreaAfter:  10, // Keep 10 rows after viewport
		},
		Theme: config.DefaultTheme(),
		KeyMap: core.NavigationKeyMap{
			Up:   []string{"up", "k"},
			Down: []string{"down", "j"},
			Home: []string{"home", "g"},
			End:  []string{"end", "G"},
			Quit: []string{"q", "esc"},
		},
	}
}

// App represents our application state
type App struct {
	table         *table.Table
	dataSource    *EmployeeDataSource
	statusMessage string
}

func main() {
	// Create data source and table
	dataSource := NewEmployeeDataSource()
	tableConfig := createTableConfig()

	// Create table with data source
	employeeTable := table.NewTable(tableConfig, dataSource)

	// Focus the table for keyboard input
	employeeTable.Focus()

	// Create app
	app := App{
		table:         employeeTable,
		dataSource:    dataSource,
		statusMessage: "Employee Directory - Navigate with keys shown below",
	}

	// Run the program
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (app App) Init() tea.Cmd {
	return tea.Batch(
		app.table.Init(),
		app.table.Focus(),
	)
}

func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return app, tea.Quit
		default:
			// Pass all other keys to table
			var cmd tea.Cmd
			_, cmd = app.table.Update(msg)
			return app, cmd
		}

	case core.CursorUpMsg, core.CursorDownMsg:
		// Handle navigation with status updates
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		state := app.table.GetState()
		app.statusMessage = fmt.Sprintf("Position: %d/%d",
			state.CursorIndex+1, app.table.GetTotalItems())
		return app, cmd

	case core.JumpToStartMsg:
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		app.statusMessage = "Jumped to first employee"
		return app, cmd

	case core.JumpToEndMsg:
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		app.statusMessage = "Jumped to last employee"
		return app, cmd

	default:
		// Pass all other messages to table
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd
	}
}

func (app App) View() string {
	var sections []string

	// Status message
	sections = append(sections, app.statusMessage)
	sections = append(sections, "")

	// Table
	sections = append(sections, app.table.View())

	// Always show controls
	sections = append(sections, "")
	sections = append(sections, "Controls: ↑↓/jk=move, Home/End=start/end, g/G=start/end, q=quit")

	return strings.Join(sections, "\n")
}
