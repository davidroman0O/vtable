# Data Virtualization

## What We're Adding

Taking our basic employee table and making it handle **large datasets** by loading data in chunks instead of all at once. Instead of 12 employees, we'll work with 10,000 employees that load as you scroll.

## Why This Matters

Loading 10,000 rows at startup would be slow. Instead, VTable loads small chunks (like 25 rows) and gets more as you scroll. This keeps the interface responsive.

## Step 1: Create a Large Dataset

Instead of a fixed array, create a data source that generates employees on demand:

```go
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

// Generate employees when requested (simulates database)
func (ds *LargeEmployeeDataSource) generateEmployees(start, count int) []core.TableRow {
	var employees []core.TableRow
	
	departments := []string{"Engineering", "Marketing", "Sales", "HR", "Finance"}
	statuses := []string{"Active", "On Leave", "Remote"}
	
	for i := 0; i < count && start+i < ds.totalEmployees; i++ {
		empID := start + i + 1
		employee := core.TableRow{
			ID: fmt.Sprintf("emp-%d", empID),
			Cells: []string{
				fmt.Sprintf("Employee %d", empID),
				departments[rand.Intn(len(departments))],
				statuses[rand.Intn(len(statuses))],
				fmt.Sprintf("$%d,000", 50+rand.Intn(80)),
			},
		}
		employees = append(employees, employee)
	}
	return employees
}
```

## Step 2: Implement Chunk Loading

The key difference is `LoadChunk` - it loads pieces of data instead of everything:

```go
// GetTotal returns the total number of employees  
func (ds *LargeEmployeeDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		// Simulate database count query delay
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

// Implement remaining DataSource interface methods...
func (ds *LargeEmployeeDataSource) RefreshTotal() tea.Cmd { return ds.GetTotal() }
func (ds *LargeEmployeeDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
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
	return func() tea.Msg { return core.SelectionResponseMsg{Success: true, Operation: "selectAll"} }
}
func (ds *LargeEmployeeDataSource) ClearSelection() tea.Cmd {
	ds.selectedItems = make(map[string]bool)
	return func() tea.Msg { return core.SelectionResponseMsg{Success: true, Operation: "clear"} }
}
func (ds *LargeEmployeeDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	return func() tea.Msg { return core.SelectionResponseMsg{Success: true, Operation: "range"} }
}
func (ds *LargeEmployeeDataSource) GetItemID(item any) string {
	if row, ok := item.(core.TableRow); ok {
		return row.ID
	}
	return ""
}
```

## Step 3: Configure for Large Data

Use a bigger viewport and configure chunk loading:

```go
func createTableConfig() core.TableConfig {
	return core.TableConfig{
		Columns:     createEmployeeColumns(), // Same as basic table
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: core.ViewportConfig{
			Height:             10, // Show more rows
			ChunkSize:          25, // Load 25 rows at a time
			TopThreshold:       3,  // Load when near top
			BottomThreshold:    3,  // Load when near bottom
			BoundingAreaBefore: 50, // Keep some rows in memory
			BoundingAreaAfter:  50,
		},
		Theme: config.DefaultTheme(),
		KeyMap: core.NavigationKeyMap{
			Up:       []string{"up", "k"},
			Down:     []string{"down", "j"},
			PageUp:   []string{"pgup", "h"},
			PageDown: []string{"pgdown", "l"},
			Home:     []string{"home", "g"},
			End:      []string{"end", "G"},
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
```

## Step 4: Add Loading Feedback

Show when data is loading:

```go
type App struct {
	table         *table.Table
	dataSource    *LargeEmployeeDataSource
	statusMessage string
	totalEmployees int
	showJumpForm  bool
	jumpInput     string
}

func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return app, tea.Quit
		default:
			var cmd tea.Cmd
			_, cmd = app.table.Update(msg)
			app.updateStatus()
			return app, cmd
		}

	case core.DataTotalMsg:
		app.totalEmployees = msg.Total
		app.updateStatus()
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd

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

func (app *App) updateStatus() {
	state := app.table.GetState()
	
	if app.showJumpForm {
		app.statusMessage = fmt.Sprintf("Enter employee number (1-%d), Enter to jump, Esc to cancel", app.totalEmployees)
	} else {
		app.statusMessage = fmt.Sprintf("Employee %d of %d | Use j/k ↑↓ h/l g/G J (jump), q to quit",
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

	// Join all sections
	return fmt.Sprintf("%s", strings.Join(sections, "\n"))
}
```

## Step 5: Complete Program

```go
func main() {
	// Create large dataset
	dataSource := NewLargeEmployeeDataSource(10000)
	tableConfig := createTableConfig()
	
	employeeTable := table.NewTable(tableConfig, dataSource)
	
	app := App{
		table:         employeeTable,
		dataSource:    dataSource,
		statusMessage: "Loading employees...",
	}
	
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
```

## What You'll See

```
Employee 1 of 10000 | Use j/k ↑↓ h/l g/G J (jump), q to quit

│ ●  │Employee Name       │  Department   │   Status   │      Salary│
│ ►  │Employee 1          │  Engineering  │   Active   │     $67,000│
│    │Employee 2          │   Marketing   │   Remote   │     $58,000│  
│    │Employee 3          │      Sales    │   Active   │     $73,000│
│    │Employee 4          │        HR     │  On Leave  │     $51,000│
│    │Employee 5          │  Engineering  │   Active   │     $89,000│
│    │Employee 6          │   Marketing   │   Remote   │     $64,000│
│    │Employee 7          │     Finance   │   Active   │     $76,000│
│    │Employee 8          │  Operations   │   Active   │     $52,000│
│    │Employee 9          │      Sales    │   Active   │     $68,000│
│    │Employee 10         │  Engineering  │  On Leave  │     $91,000│
```

**Navigation:**
- Use `j/k` or arrow keys to scroll one row
- Use `g/G` to jump to start/end
- Use `h/l` or `PgUp/PgDn` to jump by pages
- Use `J` to open jump-to-index form (type employee number, press Enter)
- Notice smooth scrolling even with 10,000 rows

**Jump-to-index example:**
```
Jump to employee (1-10000): 5000_

Enter employee number (1-10000), Enter to jump, Esc to cancel

│ ●  │Employee Name       │  Department   │   Status   │      Salary│
│ ►  │Employee 5000       │  Engineering  │   Active   │     $67,000│
...
```

## Key Changes from Basic Table

1. **Data source generates on demand** instead of storing everything
2. **LoadChunk method** loads pieces as needed  
3. **Bigger viewport** (10 rows instead of 5)
4. **Chunk configuration** in ViewportConfig
5. **Status shows position** in large dataset

## Try It

1. **Change dataset size**: Try `10000`, `50000`, or `100000` employees
2. **Change chunk size**: Try `ChunkSize: 10` or `ChunkSize: 50`
3. **Scroll around**: Use `l` or Page Down to jump quickly through thousands of rows

## What's Next

The [table selection](03-table-selection.md) section shows how to add selection to large datasets.

## Key Point

**The same table code works for 10 rows or 10,000 rows** - VTable handles the complexity of chunk loading automatically. You just configure the chunk size and thresholds. 