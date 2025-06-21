package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/table"
)

// LargeEmployeeDataSource simulates a large employee database
type LargeEmployeeDataSource struct {
	totalEmployees int
	selectedItems  map[string]bool
}

func NewLargeEmployeeDataSource(totalCount int) *LargeEmployeeDataSource {
	return &LargeEmployeeDataSource{
		totalEmployees: totalCount,
		selectedItems:  make(map[string]bool),
	}
}

// Generate employee data on-demand (simulates database query)
func (ds *LargeEmployeeDataSource) generateEmployees(start, count int) []core.TableRow {
	var employees []core.TableRow

	departments := []string{"Engineering", "Marketing", "Sales", "HR", "Finance", "Operations"}
	statuses := []string{"Active", "On Leave", "Remote"}

	for i := 0; i < count && start+i < ds.totalEmployees; i++ {
		empID := start + i + 1

		employee := core.TableRow{
			ID: fmt.Sprintf("emp-%d", empID),
			Cells: []string{
				fmt.Sprintf("Employee %d", empID),
				departments[rand.Intn(len(departments))],
				statuses[rand.Intn(len(statuses))],
				fmt.Sprintf("$%d,000", 45+rand.Intn(100)), // $45k-$145k
			},
		}

		employees = append(employees, employee)
	}

	return employees
}

// GetTotal returns the total number of employees
func (ds *LargeEmployeeDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		// Simulate database count query delay
		time.Sleep(50 * time.Millisecond)
		return core.DataTotalMsg{Total: ds.totalEmployees}
	}
}

// LoadChunk loads a specific chunk of data with simulated delay
func (ds *LargeEmployeeDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		// Simulate realistic database query time (100-300ms)
		time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

		// Generate the requested chunk
		rows := ds.generateEmployees(request.Start, request.Count)

		var items []core.Data[any]
		for _, row := range rows {
			items = append(items, core.Data[any]{
				ID:       row.ID,
				Item:     row,
				Selected: ds.selectedItems[row.ID],
				Metadata: core.NewTypedMetadata(),
			})
		}

		return core.DataChunkLoadedMsg{
			StartIndex: request.Start,
			Items:      items,
			Request:    request,
		}
	}
}

// Implement remaining DataSource interface methods
func (ds *LargeEmployeeDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

func (ds *LargeEmployeeDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		// For this demo, we'll just return success - real implementation would update data
		return core.SelectionResponseMsg{Success: true, Index: index, Selected: selected}
	}
}

func (ds *LargeEmployeeDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	if selected {
		ds.selectedItems[id] = true
	} else {
		delete(ds.selectedItems, id)
	}
	return func() tea.Msg {
		return core.SelectionResponseMsg{Success: true, ID: id, Selected: selected}
	}
}

func (ds *LargeEmployeeDataSource) SelectAll() tea.Cmd {
	return func() tea.Msg {
		return core.SelectionResponseMsg{Success: true, Operation: "selectAll"}
	}
}

func (ds *LargeEmployeeDataSource) ClearSelection() tea.Cmd {
	ds.selectedItems = make(map[string]bool)
	return func() tea.Msg {
		return core.SelectionResponseMsg{Success: true, Operation: "clear"}
	}
}

func (ds *LargeEmployeeDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	return func() tea.Msg {
		return core.SelectionResponseMsg{Success: true, Operation: "range"}
	}
}

func (ds *LargeEmployeeDataSource) GetItemID(item any) string {
	if row, ok := item.(core.TableRow); ok {
		return row.ID
	}
	return ""
}

// Configuration functions
func createLargeTableConfig() core.TableConfig {
	return core.TableConfig{
		Columns:     createEmployeeColumns(),
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: core.ViewportConfig{
			Height:             10, // Show 10 rows (larger viewport for better UX)
			TopThreshold:       3,  // Load new chunks when 3 rows from top
			BottomThreshold:    3,  // Load new chunks when 3 rows from bottom
			ChunkSize:          25, // Load 25 rows per chunk (optimal balance)
			InitialIndex:       0,  // Start at first row
			BoundingAreaBefore: 50, // Keep 50 rows before viewport in memory
			BoundingAreaAfter:  50, // Keep 50 rows after viewport in memory
		},
		Theme: config.DefaultTheme(),
		KeyMap: core.NavigationKeyMap{
			Up:       []string{"up", "k"},
			Down:     []string{"down", "j"},
			PageUp:   []string{"pgup", "h"},
			PageDown: []string{"pgdown", "l"},
			Home:     []string{"home", "g"},
			End:      []string{"end", "G"},
			Select:   []string{"enter", " "},
			Quit:     []string{"q"},
		},
	}
}

func createEmployeeColumns() []core.TableColumn {
	return []core.TableColumn{
		{Title: "Employee Name", Field: "name", Width: 20, Alignment: core.AlignLeft},
		{Title: "Department", Field: "department", Width: 15, Alignment: core.AlignCenter},
		{Title: "Status", Field: "status", Width: 12, Alignment: core.AlignCenter},
		{Title: "Salary", Field: "salary", Width: 12, Alignment: core.AlignRight},
	}
}

// Application structure with loading state tracking and jump-to-index form
type App struct {
	table          *table.Table
	dataSource     *LargeEmployeeDataSource
	statusMessage  string
	totalEmployees int

	// Jump-to-index form
	showJumpForm bool
	jumpInput    string
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
		// Handle jump form if it's open
		if app.showJumpForm {
			switch msg.String() {
			case "enter":
				// Jump to the entered index
				if index, err := strconv.Atoi(app.jumpInput); err == nil && index > 0 {
					// Convert to 0-based index and jump
					targetIndex := index - 1
					if targetIndex < app.totalEmployees {
						app.showJumpForm = false
						app.jumpInput = ""
						app.updateStatus()
						return app, core.JumpToCmd(targetIndex)
					}
				}
				// Invalid input, close form
				app.showJumpForm = false
				app.jumpInput = ""
				app.updateStatus()
				return app, nil
			case "esc":
				// Cancel jump form
				app.showJumpForm = false
				app.jumpInput = ""
				app.updateStatus()
				return app, nil
			case "backspace":
				// Remove last character
				if len(app.jumpInput) > 0 {
					app.jumpInput = app.jumpInput[:len(app.jumpInput)-1]
				}
				return app, nil
			default:
				// Add typed numbers to input
				if len(msg.String()) == 1 && msg.String() >= "0" && msg.String() <= "9" {
					app.jumpInput += msg.String()
				}
				return app, nil
			}
		}

		// Normal key handling when form is not open
		switch msg.String() {
		case "q", "ctrl+c":
			return app, tea.Quit
		case "J":
			// Open jump form (uppercase J like in full example)
			app.showJumpForm = true
			app.jumpInput = ""
			return app, nil
		default:
			var cmd tea.Cmd
			_, cmd = app.table.Update(msg)
			app.updateStatus()
			return app, cmd
		}

	// Handle total count received
	case core.DataTotalMsg:
		app.totalEmployees = msg.Total
		app.updateStatus()
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd

	// Handle chunk loading completed
	case core.DataChunkLoadedMsg:
		app.updateStatus()
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd

	// Handle navigation with status updates
	case core.CursorUpMsg, core.CursorDownMsg:
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		app.updateStatus()
		return app, cmd

	default:
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd
	}
}

func (app *App) updateStatus() {
	state := app.table.GetState()

	if app.showJumpForm {
		app.statusMessage = fmt.Sprintf("Enter employee number (1-%d), Enter to jump, Esc to cancel", app.totalEmployees)
	} else {
		app.statusMessage = fmt.Sprintf("Employee %d of %d | Press q to quit",
			state.CursorIndex+1, app.totalEmployees)
	}
}

func (app App) View() string {
	var sections []string

	// Show jump form if active
	if app.showJumpForm {
		sections = append(sections, fmt.Sprintf("Jump to employee (1-%d): %s_", app.totalEmployees, app.jumpInput))
		sections = append(sections, "")
	}

	// Status message
	sections = append(sections, app.statusMessage)
	sections = append(sections, "")

	// Table
	sections = append(sections, app.table.View())

	// Always show controls
	sections = append(sections, "")
	sections = append(sections, "Controls: ↑↓/jk=move, PageUp/Down=fast, Home/End=start/end, J=jump, q=quit")
	sections = append(sections, "Data virtualization: Only visible rows loaded for smooth 10k employee scrolling")

	// Join all sections
	return strings.Join(sections, "\n")
}

func main() {
	// Create large dataset (10,000 employees)
	totalEmployees := 10000
	dataSource := NewLargeEmployeeDataSource(totalEmployees)
	tableConfig := createLargeTableConfig()

	// Create table with large dataset
	employeeTable := table.NewTable(tableConfig, dataSource)

	// Create app with loading state tracking
	app := App{
		table:         employeeTable,
		dataSource:    dataSource,
		statusMessage: "Loading 10,000 employees... | Press q to quit",
	}

	// Run the program
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
