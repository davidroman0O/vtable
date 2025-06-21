# Column Management

## What We're Adding

Building on our horizontal scrolling example, we're adding the ability to dynamically manage table columns:
- **Column reordering**: Move columns left/right to reorganize the table
- **Add/remove columns**: Show more data or hide unneeded columns
- **Width adjustment**: Make columns wider or narrower in real-time
- **Alignment changes**: Switch between left/center/right alignment

This lets users customize their table layout without restarting the application.

## New Concepts

### Available vs Visible Columns
Instead of hardcoding all columns, we separate:
- **Available columns**: All possible columns the table could show
- **Visible columns**: Which ones are currently displayed

```go
// All possible columns
availableColumns := []core.TableColumn{
    {Title: "ID", Width: 8, Alignment: core.AlignCenter},
    {Title: "Name", Width: 25, Alignment: core.AlignLeft},
    {Title: "Email", Width: 30, Alignment: core.AlignLeft},    // Hidden initially
    {Title: "Phone", Width: 18, Alignment: core.AlignCenter}, // Hidden initially
    // ... more columns
}

// Which ones to show (by index into availableColumns)
visibleColumns := []int{0, 1, 2, 3} // Show ID, Name, Dept, Status
```

### Dynamic Column State
We track which columns are visible and their current configuration:

```go
type AppModel struct {
    // ... existing fields ...
    availableColumns []core.TableColumn // All possible columns
    visibleColumns   []int              // Indices of visible columns  
    columnWidths     []int              // Current widths for adjustment
}
```

## Code Changes

### 1. Setup Available Columns
Replace the hardcoded columns with a larger set:

```go
func main() {
    // Define ALL possible columns (more than we'll show)
    availableColumns := []core.TableColumn{
        {Title: "ID", Width: 8, Alignment: core.AlignCenter, Field: "id"},
        {Title: "Employee Name", Width: 25, Alignment: core.AlignLeft, Field: "name"},
        {Title: "Department", Width: 20, Alignment: core.AlignCenter, Field: "department"},
        {Title: "Status", Width: 15, Alignment: core.AlignCenter, Field: "status"},
        {Title: "Salary", Width: 12, Alignment: core.AlignRight, Field: "salary"},
        {Title: "Email", Width: 30, Alignment: core.AlignLeft, Field: "email"},      // NEW
        {Title: "Phone", Width: 18, Alignment: core.AlignCenter, Field: "phone"},   // NEW
        {Title: "Description", Width: 50, Alignment: core.AlignLeft, Field: "description"},
    }

    // Start with only some columns visible
    visibleColumns := []int{0, 1, 2, 3, 4, 7} // ID, Name, Dept, Status, Salary, Description
    
    // Build initial column set
    initialColumns := make([]core.TableColumn, len(visibleColumns))
    for i, colIndex := range visibleColumns {
        initialColumns[i] = availableColumns[colIndex]
    }
}
```

### 2. Add Column Management Controls
Add new key handlers to the Update function:

```go
// === COLUMN MANAGEMENT CONTROLS ===
case "ctrl+left":
    return m.moveColumnLeft()

case "ctrl+right": 
    return m.moveColumnRight()

case "+", "=":
    return m.addColumn()

case "-", "_":
    return m.removeColumn()

case "W":
    return m.adjustColumnWidth(5)  // Increase width

case "w":
    return m.adjustColumnWidth(-5) // Decrease width

case "A":
    return m.cycleColumnAlignment()

case "R":
    return m.resetColumns()
```

### 3. Implement Column Operations
Add these methods to handle column changes:

```go
// Move current column left in the display order
func (m AppModel) moveColumnLeft() (tea.Model, tea.Cmd) {
    _, _, currentColumn, _ := m.table.GetHorizontalScrollState()
    
    if currentColumn > 0 {
        // Swap positions in visible columns list
        m.visibleColumns[currentColumn], m.visibleColumns[currentColumn-1] = 
            m.visibleColumns[currentColumn-1], m.visibleColumns[currentColumn]
        
        // Update table with new order
        newColumns := m.buildCurrentColumns()
        return m, tea.Batch(
            core.ColumnSetCmd(newColumns),
            core.PrevColumnCmd(), // Move focus to follow the column
        )
    }
    return m, nil
}

// Add the next available column
func (m AppModel) addColumn() (tea.Model, tea.Cmd) {
    // Find first column not currently visible
    visibleSet := make(map[int]bool)
    for _, colIndex := range m.visibleColumns {
        visibleSet[colIndex] = true
    }
    
    for i, col := range m.availableColumns {
        if !visibleSet[i] {
            // Add this column
            m.visibleColumns = append(m.visibleColumns, i)
            m.columnWidths = append(m.columnWidths, col.Width)
            
            newColumns := m.buildCurrentColumns()
            return m, core.ColumnSetCmd(newColumns)
        }
    }
    return m, nil // All columns already visible
}

// Adjust width of current column
func (m AppModel) adjustColumnWidth(delta int) (tea.Model, tea.Cmd) {
    _, _, currentColumn, _ := m.table.GetHorizontalScrollState()
    
    newWidth := m.columnWidths[currentColumn] + delta
    if newWidth < 5 { newWidth = 5 }     // Minimum width
    if newWidth > 100 { newWidth = 100 } // Maximum width
    
    m.columnWidths[currentColumn] = newWidth
    newColumns := m.buildCurrentColumns()
    return m, core.ColumnSetCmd(newColumns)
}

// Build current column configuration from visible columns
func (m AppModel) buildCurrentColumns() []core.TableColumn {
    columns := make([]core.TableColumn, len(m.visibleColumns))
    for i, colIndex := range m.visibleColumns {
        col := m.availableColumns[colIndex]
        col.Width = m.columnWidths[i] // Use current width
        columns[i] = col
    }
    return columns
}
```

## Key Features Explained

### Column Reordering
- `Ctrl+←` `Ctrl+→` swap the current column with its neighbor
- The focus follows the moved column automatically
- Changes the visual order without affecting data

### Dynamic Add/Remove
- `+` finds the first hidden column and adds it to the table
- `-` removes the current column (but keeps it in available list)
- Prevents removing the last column

### Width Adjustment  
- `W` `w` increase/decrease current column width by 5 characters
- Enforces minimum (5) and maximum (100) width limits
- Changes apply immediately

### Reset Functionality
- `R` returns to the default column configuration
- Restores original widths and column order
- Useful escape hatch when layout gets messy

## Core Commands Used

```go
// Set all columns at once (main command for column management)
core.ColumnSetCmd(columns []TableColumn) tea.Cmd

// Navigate between columns (inherited from horizontal scrolling)
core.NextColumnCmd() tea.Cmd
core.PrevColumnCmd() tea.Cmd
```

## Status Display
Update the View to show column management state:

```go
// Show visible/total columns and current column info
status := fmt.Sprintf("Cols: %d/%d | Current: %s (%d)", 
    len(m.visibleColumns),
    len(m.availableColumns), 
    currentColumnName,
    currentColumnWidth,
)
```

## Controls

| Key | Action |
|-----|--------|
| `Ctrl+←` `Ctrl+→` | Move current column left/right |
| `+` `=` | Add next available column |
| `-` `_` | Remove current column |
| `W` `w` | Increase/decrease column width |
| `A` | Cycle column alignment |
| `R` | Reset to default columns |

## Try It Yourself

1. **Reorder columns**: Use `Ctrl+←` `Ctrl+→` to move the ID column around
2. **Add hidden columns**: Press `+` to add Email and Phone columns
3. **Remove columns**: Use `-` to remove columns you don't need
4. **Adjust widths**: Make the Description column narrower with `w`
5. **Reset when confused**: Press `R` to go back to defaults

## What's Next

In the next section, we'll explore [Filtering and Sorting](11-filtering-sorting.md) to add data manipulation to our configurable table.

## Running the Example

```bash
cd docs/05-table-component/examples/column-management
go run .
```

This example shows how to build user-configurable tables that adapt to different workflows and screen sizes. 