# The Table Component: Column Management

This guide demonstrates how to give users dynamic control over the table's structure. You'll learn how to implement features for reordering, adding, removing, and resizing columns at runtime.

## What You'll Build

We will build an interactive table where the user can:
-   **Reorder columns** using `Ctrl+←`/`Ctrl+→`.
-   **Add and remove columns** from a predefined set using `+`/`-`.
-   **Adjust the width** of the currently active column using `W`/`w`.
-   **Cycle through text alignments** for the active column with `A`.

```text
// User can reorder columns, for example, moving 'Status' to the front.
┌────────────┬───────────────┬───────────────────┬────────────┐
│   Status   │ Employee Name │  Department       │     Salary │
├────────────┼───────────────┼───────────────────┼────────────┤
│ 🟢 Active  │ ► Alice       │  Engineering      │    $75,000 │
└────────────┴───────────────┴───────────────────┴────────────┘
```

## How It Works: Managing Column State

The key to dynamic column management is to maintain two separate lists in your application's state:

1.  `availableColumns`: A complete list of *all possible* columns the table could display.
2.  `visibleColumns`: A list of indices that maps to `availableColumns`, defining which columns are currently visible and in what order.

```go
type AppModel struct {
	table *table.Table
	// ...
	availableColumns []core.TableColumn // All possible columns
	visibleColumns   []int              // Indices of visible columns (e.g., [1, 0, 3])
	columnWidths     []int              // Current widths for each visible column
}
```

When the user reorders, adds, or removes a column, you manipulate the `visibleColumns` slice and then send a single command to VTable to update its layout.

## The `ColumnSetCmd` Command

All dynamic layout changes are powered by one core command:
-   `core.ColumnSetCmd(columns []core.TableColumn)`: Replaces the table's entire column configuration with the new set you provide.

## Step 1: Define All Available Columns

In your `main` function, create the master list of all columns that could potentially be displayed.

```go
// Define ALL possible columns, including ones hidden by default.
availableColumns := []core.TableColumn{
    {Title: "ID", Width: 8, Field: "id"},
    {Title: "Employee Name", Width: 25, Field: "name"},
    {Title: "Department", Width: 20, Field: "department"},
    {Title: "Status", Width: 15, Field: "status"},
    {Title: "Salary", Width: 12, Field: "salary"},
    {Title: "Email", Width: 30, Field: "email"},   // Hidden by default
    {Title: "Phone", Width: 18, Field: "phone"},   // Hidden by default
}

// Define the initial set of visible columns by their index.
visibleColumns := []int{0, 1, 2, 3, 4} // Shows ID, Name, Dept, Status, Salary
```

## Step 2: Implement Column Management Logic

In your `AppModel`, create helper methods to handle the logic for reordering, adding, and removing columns.

```go
// Rebuilds the slice of columns to be sent to the table.
func (m AppModel) buildCurrentColumns() []core.TableColumn {
	columns := make([]core.TableColumn, len(m.visibleColumns))
	for i, colIndex := range m.visibleColumns {
		col := m.availableColumns[colIndex]
		col.Width = m.columnWidths[i] // Apply the current dynamic width
		columns[i] = col
	}
	return columns
}

// Moves the currently active column one position to the left.
func (m AppModel) moveColumnLeft() (tea.Model, tea.Cmd) {
	_, _, currentColumn, _ := m.table.GetHorizontalScrollState()
	if currentColumn > 0 {
		// Swap the column's position in the visible list.
		m.visibleColumns[currentColumn], m.visibleColumns[currentColumn-1] =
			m.visibleColumns[currentColumn-1], m.visibleColumns[currentColumn]

		// Rebuild the column set and send the update command.
		newColumns := m.buildCurrentColumns()
		return m, core.ColumnSetCmd(newColumns)
	}
	return m, nil
}
// ... implement addColumn, removeColumn, adjustColumnWidth etc.
```

## Step 3: Add Keyboard Controls

In your app's `Update` method, map keys to your new column management functions.

```go
case tea.KeyMsg:
    switch msg.String() {
    case "ctrl+left":
        return m.moveColumnLeft()
    case "ctrl+right":
        return m.moveColumnRight()
    case "+", "=":
        return m.addColumn()
    case "-", "_":
        return m.removeColumn()
    case "W":
        return m.adjustColumnWidth(5) // Increase width
    case "w":
        return m.adjustColumnWidth(-5) // Decrease width
    case "A":
        return m.cycleColumnAlignment()
    case "R":
        return m.resetColumns()
    }
```

## What You'll Experience

-   **Column Reordering**: Use `Ctrl+←`/`Ctrl+→` to move the active column and watch the table layout update instantly.
-   **Dynamic Columns**: Press `+` to add hidden columns like "Email" and "Phone," and `-` to remove them.
-   **Live Width Adjustment**: Make columns wider or narrower with `W`/`w` to fit content perfectly.
-   **Reset to Default**: If the layout becomes messy, press `R` to instantly revert to the initial column configuration.

## Complete Example

See the full working code for this guide in the examples directory:
[`docs/05-table-component/examples/column-management/`](examples/column-management/)

To run it:
```bash
cd docs/05-table-component/examples/column-management
go run .
```

## What's Next?

You have now mastered the layout and structure of the VTable `Table` component. The next step is to add powerful data operations by implementing filtering and sorting in your `DataSource`.

**Next:** [Filtering and Sorting →](11-filtering-sorting.md) 