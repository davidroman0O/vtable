# The Table Component: Cell Constraints

Cell constraints give you precise control over the layout of your table's columns and headers. This guide will show you how to manage column width, text alignment, and cell padding to create a polished, readable, and professional-looking table.

## What You'll Build

We will take our selectable employee table and add dynamic controls to adjust its layout in real-time. You'll learn how to:
-   Change column widths.
-   Set different text alignments for data cells and headers.
-   Configure padding within cells for better spacing.
-   Observe how VTable automatically truncates text that overflows its container.

![VTable Cell Constraints Demo](examples/cell-constraints/cell-constraints.gif)

```text
// Example of a table with custom constraints
┌─── Employee Name ───┬─── Department ───┬─── Status ───┐
│ Alice Johnson        │   Engineering    │    Active    │
│ Bob Smith            │     Marketing    │    Remote    │
│ ...                  │        ...       │       ...    │
└──────────────────────┴──────────────────┴──────────────┘
```

## How It Works: The `TableColumn` and Constraints

The `core.TableColumn` struct is the center of layout control. You can define separate constraints for header cells and data cells.

```go
column := core.TableColumn{
    Title:     "Employee Name",
    Width:     25, // The default width for this column

    // --- Constraints for DATA cells ---
    Alignment: core.AlignLeft,  // Data will be left-aligned.

    // --- Constraints for the HEADER cell ---
    HeaderAlignment:  core.AlignCenter, // Header text will be center-aligned.
    HeaderConstraint: core.CellConstraint{
        // You can provide even more specific overrides for the header.
        Padding: core.PaddingConfig{Left: 2, Right: 2},
    },
}
```

-   `Width`: The total width of the column in characters. Text overflowing this width will be truncated.
-   `Alignment`: Sets the horizontal alignment (`AlignLeft`, `AlignCenter`, `AlignRight`) for the data cells in this column.
-   `HeaderAlignment`: *Independently* sets the alignment for the header cell.
-   `HeaderConstraint`: An optional, more specific set of constraints that applies *only* to the header.

## Step 1: Add More Data for Demonstration

To better demonstrate text truncation and layout changes, our `DataSource` will now provide a longer description for each employee.

```go
// In your DataSource's data generation:
longDescriptions := []string{
    "Experienced software engineer specializing in backend systems...",
    "Creative marketing professional focused on digital campaigns...",
    // ... more long strings
}

// Add the description to the Cells slice
Cells: []string{
    // ... other cells
    longDescriptions[i%len(longDescriptions)],
},
```

## Step 2: Create a Dynamic Column Builder

Instead of a static column definition, create a function that builds the column configuration based on the application's current state. This allows for real-time updates.

```go
// In your App model:
type App struct {
    // ...
    widthMode     int // 0=narrow, 1=normal, 2=wide
    alignmentMode int // 0=mixed, 1=left, 2=center, 3=right
    paddingMode   int // 0=none, 1=normal, 2=extra
}

// This function builds the column slice based on the app's state.
func (app *App) buildColumnsWithConstraints() []core.TableColumn {
    // ... logic to set nameWidth, deptWidth, etc. based on app.widthMode ...
    // ... logic to set alignment based on app.alignmentMode ...
    // ... logic to set padding based on app.paddingMode ...

    columns := []core.TableColumn{
        {
            Title:           "Employee Name",
            Width:           nameWidth,
            Alignment:       dataAlignment,
            HeaderAlignment: headerAlignment,
        },
        // ... other columns ...
    }
    return columns
}
```

## Step 3: Add Keyboard Controls to Change Constraints

In your `Update` method, add keys to cycle through the different layout modes. When a key is pressed, update the state and send a `core.ColumnSetCmd` to the table.

```go
// In your app's Update method:
case tea.KeyMsg:
    switch msg.String() {
    case "w": // Cycle column widths
        app.cycleColumnWidths()
        return app, app.updateTableColumns()
    case "a": // Cycle data alignment
        app.cycleAlignment()
        return app, app.updateTableColumns()
    case "A": // Cycle header alignment (Shift+A)
        app.cycleHeaderAlignment()
        return app, app.updateTableColumns()
    case "p": // Cycle padding
        app.cyclePadding()
        return app, app.updateTableColumns()
    }

// This helper function applies the changes.
func (app *App) updateTableColumns() tea.Cmd {
	columns := app.buildColumnsWithConstraints()
	return core.ColumnSetCmd(columns)
}
```

## What You'll Experience

-   **Dynamic Layouts**: Press `w`, `a`, `p`, etc., to see the table layout change instantly.
-   **Clear Separation**: Notice how you can change the header alignment (`A`) independently of the data alignment (`a`).
-   **Automatic Truncation**: As you make the "Description" column narrower (`t`), the text will be automatically truncated with an ellipsis (`...`).
-   **Padding Control**: Cycle through padding modes (`p`) to see how extra spacing improves readability.

## Complete Example

See the full working code, which includes an interactive demo for cycling through different constraint settings.
[`docs/05-table-component/examples/cell-constraints/`](examples/cell-constraints/)

To run it:
```bash
cd docs/05-table-component/examples/cell-constraints
go run main.go
```

## What's Next?

You now have full control over the layout and structure of your table. The next step is to customize the *content* within the cells by using custom column formatters to add icons and data-driven styling.

**Next:** [Column Formatting →](05-column-formatting.md) 