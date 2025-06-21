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

// LargeEmployeeDataSource simulates a large employee database with selection tracking
type LargeEmployeeDataSource struct {
	totalEmployees int
	data           []core.TableRow // Store all data like the full featured example
	selectedItems  map[string]bool // Selection state
	recentActivity []string        // Track recent selection activity
}

func NewLargeEmployeeDataSource(totalCount int) *LargeEmployeeDataSource {
	// Generate ALL data upfront like the full featured example
	data := make([]core.TableRow, totalCount)

	departments := []string{"Engineering", "Marketing", "Sales", "HR", "Finance", "Operations"}
	statuses := []string{"Active", "On Leave", "Remote"}

	for i := 0; i < totalCount; i++ {
		data[i] = core.TableRow{
			ID: fmt.Sprintf("emp-%d", i+1),
			Cells: []string{
				fmt.Sprintf("Employee %d", i+1),
				departments[rand.Intn(len(departments))],
				statuses[rand.Intn(len(statuses))],
				fmt.Sprintf("$%d,000", 45+rand.Intn(100)), // $45k-$145k
			},
		}
	}

	return &LargeEmployeeDataSource{
		totalEmployees: totalCount,
		data:           data,
		selectedItems:  make(map[string]bool),
		recentActivity: make([]string, 0),
	}
}

// GetTotal returns the total number of employees
func (ds *LargeEmployeeDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		// Simulate database count query delay
		time.Sleep(10 * time.Millisecond)
		return core.DataTotalMsg{Total: ds.totalEmployees}
	}
}

// LoadChunk loads a specific chunk of data with selection state - EXACTLY like full featured example
func (ds *LargeEmployeeDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		// Simulate realistic database query time
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

		start := request.Start
		end := start + request.Count
		if end > ds.totalEmployees {
			end = ds.totalEmployees
		}

		var items []core.Data[any]
		for i := start; i < end; i++ {
			if i < len(ds.data) {
				items = append(items, core.Data[any]{
					ID:       ds.data[i].ID,
					Item:     ds.data[i],
					Selected: ds.selectedItems[ds.data[i].ID], // Apply selection state EXACTLY like full featured example
					Metadata: core.NewTypedMetadata(),
				})
			}
		}

		return core.DataChunkLoadedMsg{
			StartIndex: start,
			Items:      items,
			Request:    request,
		}
	}
}

// Selection interface methods - EXACTLY like full featured example
func (ds *LargeEmployeeDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

func (ds *LargeEmployeeDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		if index >= 0 && index < len(ds.data) {
			id := ds.data[index].ID

			// Actually update selection state EXACTLY like full featured example
			if selected {
				ds.selectedItems[id] = true
				ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected: %s", ds.data[index].Cells[0]))
			} else {
				delete(ds.selectedItems, id)
				ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Deselected: %s", ds.data[index].Cells[0]))
			}

			// Keep only last 10 activities
			if len(ds.recentActivity) > 10 {
				ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
			}

			return core.SelectionResponseMsg{
				Success:   true,
				Index:     index,
				ID:        id,
				Selected:  selected,
				Operation: "toggle",
			}
		}

		return core.SelectionResponseMsg{
			Success:   false,
			Index:     index,
			ID:        "",
			Selected:  false,
			Operation: "toggle",
			Error:     fmt.Errorf("invalid index: %d", index),
		}
	}
}

func (ds *LargeEmployeeDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	return func() tea.Msg {
		// Find the item by ID - EXACTLY like full featured example
		for i, row := range ds.data {
			if row.ID == id {
				// Actually update selection state!
				if selected {
					ds.selectedItems[id] = true
					ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected: %s", row.Cells[0]))
				} else {
					delete(ds.selectedItems, id)
					ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Deselected: %s", row.Cells[0]))
				}

				// Keep only last 10 activities
				if len(ds.recentActivity) > 10 {
					ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
				}

				return core.SelectionResponseMsg{
					Success:   true,
					Index:     i,
					ID:        id,
					Selected:  selected,
					Operation: "toggle",
				}
			}
		}

		return core.SelectionResponseMsg{
			Success:   false,
			Index:     -1,
			ID:        id,
			Selected:  false,
			Operation: "toggle",
			Error:     fmt.Errorf("item not found: %s", id),
		}
	}
}

func (ds *LargeEmployeeDataSource) ClearSelection() tea.Cmd {
	return func() tea.Msg {
		count := len(ds.selectedItems)
		ds.selectedItems = make(map[string]bool) // Clear all selections
		ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Cleared %d selections", count))

		// Keep only last 10 activities
		if len(ds.recentActivity) > 10 {
			ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
		}

		return core.SelectionResponseMsg{
			Success:   true,
			Index:     -1,
			ID:        "",
			Selected:  false,
			Operation: "clear",
		}
	}
}

func (ds *LargeEmployeeDataSource) SelectAll() tea.Cmd {
	return func() tea.Msg {
		// Select all items - EXACTLY like full featured example
		for _, row := range ds.data {
			ds.selectedItems[row.ID] = true
		}

		ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected all %d items", len(ds.data)))

		// Keep only last 10 activities
		if len(ds.recentActivity) > 10 {
			ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
		}

		return core.SelectionResponseMsg{
			Success:   true,
			Index:     -1,
			ID:        "",
			Selected:  true,
			Operation: "selectAll",
		}
	}
}

func (ds *LargeEmployeeDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	return func() tea.Msg {
		var affectedIDs []string
		count := 0

		for i := startIndex; i <= endIndex && i < len(ds.data); i++ {
			ds.selectedItems[ds.data[i].ID] = true
			affectedIDs = append(affectedIDs, ds.data[i].ID)
			count++
		}

		ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected range: %d items", count))

		// Keep only last 10 activities
		if len(ds.recentActivity) > 10 {
			ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
		}

		return core.SelectionResponseMsg{
			Success:     true,
			Index:       startIndex,
			ID:          "",
			Selected:    true,
			Operation:   "range",
			AffectedIDs: affectedIDs,
		}
	}
}

func (ds *LargeEmployeeDataSource) GetItemID(item any) string {
	if row, ok := item.(core.TableRow); ok {
		return row.ID
	}
	return ""
}

func (ds *LargeEmployeeDataSource) GetRecentActivity() []string {
	return ds.recentActivity
}

func (ds *LargeEmployeeDataSource) GetSelectionCount() int {
	return len(ds.selectedItems)
}

// Configuration functions
func createTableConfig() core.TableConfig {
	return core.TableConfig{
		Columns:       createEmployeeColumns(),
		ShowHeader:    true,
		ShowBorders:   true,
		SelectionMode: core.SelectionMultiple, // Enable multiple selection
		ViewportConfig: core.ViewportConfig{
			Height:             10,
			ChunkSize:          25,
			TopThreshold:       3,
			BottomThreshold:    3,
			BoundingAreaBefore: 50,
			BoundingAreaAfter:  50,
		},
		Theme: config.DefaultTheme(),
		KeyMap: core.NavigationKeyMap{
			Up:        []string{"up", "k"},
			Down:      []string{"down", "j"},
			PageUp:    []string{"pgup", "h"},
			PageDown:  []string{"pgdown", "l"},
			Home:      []string{"home", "g"},
			End:       []string{"end", "G"},
			Select:    []string{"enter", " "}, // Space and Enter to select
			SelectAll: []string{"ctrl+a"},     // Ctrl+A to select all
			Quit:      []string{"q"},
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

// Application structure with selection tracking
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
			// Open jump form
			app.showJumpForm = true
			app.jumpInput = ""
			return app, nil

		// Selection commands
		case " ", "enter":
			// Toggle selection of current item
			return app, core.SelectCurrentCmd()
		case "ctrl+a":
			// Select all items
			return app, core.SelectAllCmd()
		case "c":
			// Clear all selections
			return app, core.SelectClearCmd()
		case "s":
			// Show selection info
			app.showSelectionInfo()
			return app, nil

		default:
			var cmd tea.Cmd
			_, cmd = app.table.Update(msg)
			app.updateStatus()
			return app, cmd
		}

	// Handle selection responses - EXACTLY like full featured example
	case core.SelectionResponseMsg:
		app.updateStatus()
		// Pass to table without extra chunk refresh
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd

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

	default:
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd
	}
}

func (app *App) showSelectionInfo() {
	count := app.dataSource.GetSelectionCount()
	if count > 0 {
		app.statusMessage = fmt.Sprintf("✓ %d employees selected | Use c to clear, space to toggle", count)
	} else {
		app.statusMessage = "No employees selected | Use space to select, ctrl+a for all"
	}
}

func (app *App) updateStatus() {
	state := app.table.GetState()
	selectionCount := app.dataSource.GetSelectionCount()

	if app.showJumpForm {
		app.statusMessage = fmt.Sprintf("Enter employee number (1-%d), Enter to jump, Esc to cancel", app.totalEmployees)
	} else {
		app.statusMessage = fmt.Sprintf("Employee %d/%d | Selected: %d | Use space/enter ctrl+a c s J, q to quit",
			state.CursorIndex+1, app.totalEmployees, selectionCount)
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

	// Show selection info
	selectionCount := app.dataSource.GetSelectionCount()
	if selectionCount > 0 {
		sections = append(sections, "")
		sections = append(sections, fmt.Sprintf("Selected: %d items", selectionCount))
	}

	// Show recent activity
	recentActivity := app.dataSource.GetRecentActivity()
	if len(recentActivity) > 0 {
		sections = append(sections, "")
		sections = append(sections, "Recent Activity:")
		for i := len(recentActivity) - 1; i >= 0 && i >= len(recentActivity)-3; i-- {
			sections = append(sections, fmt.Sprintf("  • %s", recentActivity[i]))
		}
	}

	// Join all sections
	return strings.Join(sections, "\n")
}

func main() {
	// Create large dataset with selection tracking
	dataSource := NewLargeEmployeeDataSource(10000)
	tableConfig := createTableConfig()

	// Create table with selection enabled
	employeeTable := table.NewTable(tableConfig, dataSource)

	// Create app with selection tracking
	app := App{
		table:         employeeTable,
		dataSource:    dataSource,
		statusMessage: "Loading employees...",
	}

	// Run the program
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
