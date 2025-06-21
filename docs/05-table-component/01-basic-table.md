# Basic Table

## What We're Adding

A **functional data table** with multiple columns, keyboard navigation, and clean visual indicators. This builds the foundation for all advanced table features - proper DataSource implementation, column configuration, and component-based rendering.

## Understanding Tables in VTable

Tables in VTable are built on three core concepts:

- **DataSource**: Provides the actual data rows through chunk loading
- **Columns**: Define structure, width, alignment, and field mapping  
- **Navigation**: Keyboard controls with cursor positioning

Unlike simple lists, tables have:
- **Multiple columns** with different data types
- **Column-specific formatting** (alignment, width)
- **Structured data access** through field mapping
- **Component rendering** for clean visual indicators

## Step 1: Create the Table Data Structure

First, let's define our data structure. We'll create a simple employee table:

```go
// TableRow represents a single row of table data
type TableRow struct {
    ID    string   // Unique identifier
    Cells []string // Column values: [Name, Department, Status, Salary]
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
    }
}
```

## Step 2: Implement the DataSource

Tables require a DataSource that implements the `TableDataSource` interface:

```go
// EmployeeDataSource implements TableDataSource for employee data
type EmployeeDataSource struct {
    data []core.TableRow
}

func NewEmployeeDataSource() *EmployeeDataSource {
    return &EmployeeDataSource{
        data: createEmployeeData(),
    }
}

// GetTotal returns total number of employees
func (ds *EmployeeDataSource) GetTotal() tea.Cmd {
    return func() tea.Msg {
        return core.DataTotalMsg{Total: len(ds.data)}
    }
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

// GetItemID returns the ID for a table row
func (ds *EmployeeDataSource) GetItemID(item any) string {
    if row, ok := item.(core.TableRow); ok {
        return row.ID
    }
    return ""
}
```

## Step 3: Define Table Columns

Configure columns with specific widths, alignments, and field mappings:

```go
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
```

## Step 4: Create the Table Configuration

Configure the table with viewport settings and navigation:

```go
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
        Theme: core.DefaultTheme(),
        KeyMap: core.NavigationKeyMap{
            Up:    []string{"up", "k"},
            Down:  []string{"down", "j"},
            Home:  []string{"home", "g"},
            End:   []string{"end", "G"},
            Quit:  []string{"q", "esc"},
        },
    }
}
```

## Step 5: Create the Main Application

Build the Bubble Tea application with table integration:

```go
package main

import (
    "fmt"
    "log"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/davidroman0O/vtable/core"
    "github.com/davidroman0O/vtable/table"
)

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
        statusMessage: "Employee Directory - Use ↑↓ or j/k to navigate, q to quit",
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
    return fmt.Sprintf("%s\n\n%s", 
        app.statusMessage,
        app.table.View())
}
```

## Step 6: Component Rendering System

VTable uses component-based rendering for all tables automatically:

```go
func (app App) Init() tea.Cmd {
    return tea.Batch(
        app.table.Init(),
        app.table.Focus(),
    )
}
```

Component rendering provides clean visual indicators (►) that are rendered separately from your data content, ensuring no contamination of actual cell values. This system is always enabled and provides the clean cursor indicators you see in the output.

## What You'll See

```
Employee Directory - Use ↑↓ or j/k to navigate, q to quit

┌─────────────────────────┬───────────────┬────────────┬────────────┐
│ Employee Name           │  Department   │   Status   │     Salary │
├─────────────────────────┼───────────────┼────────────┼────────────┤
│ ► Alice Johnson         │ Engineering   │   Active   │    $75,000 │
│   Bob Smith             │   Marketing   │   Active   │    $65,000 │
│   Carol Davis           │ Engineering   │ On Leave   │    $80,000 │
│   David Wilson          │     Sales     │   Active   │    $70,000 │
│   Eve Brown             │      HR       │   Active   │    $60,000 │
│   Frank Miller          │ Engineering   │   Active   │    $85,000 │
│   Grace Lee             │   Marketing   │   Active   │    $62,000 │
│   Henry Taylor          │     Sales     │ On Leave   │    $68,000 │
└─────────────────────────┴───────────────┴────────────┴────────────┘
```

**Key Visual Elements:**
- **Clean cursor indicator** (►) shows current row
- **Column alignment** reflects configuration (left/center/right)
- **Proper spacing** with Unicode box drawing
- **Header separation** distinguishes data from headers

## Try It Yourself

1. **Navigation**: Use `j/k` or arrow keys to move up/down
2. **Quick jumps**: Press `g` for start, `G` for end  
3. **Different data**: Add more employees to see scrolling
4. **Column widths**: Change `Width` values to see text adjustment

## Key Navigation Controls

| Key | Action |
|-----|--------|
| `↑` `k` | Move up one row |
| `↓` `j` | Move down one row |
| `g` | Jump to first row |
| `G` | Jump to last row |
| `q` | Quit application |

## What You've Built

✅ **Structured data table** with proper column definitions  
✅ **DataSource implementation** following VTable patterns  
✅ **Keyboard navigation** with configurable key mappings  
✅ **Component rendering** for clean visual indicators  
✅ **Responsive layout** with column alignment and spacing  

## What's Next

The [data virtualization](02-data-virtualization.md) section shows how to handle large datasets efficiently with chunk loading and performance optimization.

## Key Insights

**Tables vs Lists**: Tables require structured column definitions and field mapping, making them more complex but far more powerful for tabular data.

**Component rendering** provides clean visual feedback without cluttering cell content - the cursor indicator (►) is rendered separately from your data.

**DataSource pattern** enables VTable to work with any data backend - you control how data is loaded and structured. 