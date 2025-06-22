# The Table Component: Basic Usage

Let's build your first VTable **Table**. The Table component is designed for displaying structured, columnar data. Like all VTable components, it's fully virtualized, allowing it to handle thousands or millions of rows with high performance.

## What You'll Build

A clean, bordered table that displays a list of employees with multiple columns, proper alignment, and full keyboard navigation.

```text
┌───────────────────┬───────────────┬────────────┬────────────┐
│   Employee Name   │  Department   │   Status   │     Salary │
├───────────────────┼───────────────┼────────────┼────────────┤
│ ► Alice Johnson   │  Engineering  │   Active   │    $75,000 │
│   Bob Smith       │   Marketing   │   Active   │    $65,000 │
│   Carol Davis     │  Engineering  │  On Leave  │    $80,000 │
│   ...             │      ...      │     ...    │        ... │
└───────────────────┴───────────────┴────────────┴────────────┘
```

## How Tables Differ from Lists

-   **Structure**: Tables are defined by a set of **columns**, each with its own properties like title, width, and alignment.
-   **Data**: The `DataSource` for a table provides data as `core.TableRow` objects, where each row contains an array of strings (`Cells`) that correspond to the defined columns.
-   **Rendering**: Tables use a more complex rendering system that draws headers, borders, and ensures cell content fits within column boundaries.

## Step 1: Define Your Table's Columns

The first step is to define the structure of your table by creating a slice of `core.TableColumn`.

```go
func createEmployeeColumns() []core.TableColumn {
	return []core.TableColumn{
		{
			Title:     "Employee Name",
			Field:     "name", // Used for sorting/filtering later
			Width:     25,
			Alignment: core.AlignLeft,
		},
		{
			Title:     "Department",
			Field:     "department",
			Width:     15,
			Alignment: core.AlignCenter,
		},
		{
			Title:     "Status",
			Field:     "status",
			Width:     12,
			Alignment: core.AlignCenter,
		},
		{
			Title:     "Salary",
			Field:     "salary",
			Width:     12,
			Alignment: core.AlignRight, // Right-align for numbers
		},
	}
}
```

## Step 2: Implement the `DataSource` for Table Data

Your `DataSource` will provide `core.TableRow` objects to the table.

```go
// EmployeeDataSource implements the DataSource for our table.
type EmployeeDataSource struct {
	data []core.TableRow
}

// createEmployeeData returns our sample dataset.
func createEmployeeData() []core.TableRow {
	return []core.TableRow{
		{ID: "emp-1", Cells: []string{"Alice Johnson", "Engineering", "Active", "$75,000"}},
		{ID: "emp-2", Cells: []string{"Bob Smith", "Marketing", "Active", "$65,000"}},
		// ... more rows ...
	}
}

// LoadChunk provides the data to the table.
func (ds *EmployeeDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		// ... (implementation is very similar to the List's DataSource)

		// The key difference is that the `Item` is a `core.TableRow`
		items = append(items, core.Data[any]{
			ID:   ds.data[i].ID,
			Item: ds.data[i], // Pass the TableRow struct here
		})

		// ...
	}
}
```

## Step 3: Create and Configure the Table

Now, assemble the columns and `DataSource` into a `Table` component.

```go
import "github.com/davidroman0O/vtable/table"

func createTable() *table.Table {
    dataSource := NewEmployeeDataSource()

    tableConfig := core.TableConfig{
        Columns:     createEmployeeColumns(),
        ShowHeader:  true,
        ShowBorders: true,
        ViewportConfig: core.ViewportConfig{
            Height: 8, // Show 8 rows
        },
        Theme:         config.DefaultTheme(), // Use the default look
        KeyMap:        core.DefaultNavigationKeyMap(),
    }

    return table.NewTable(tableConfig, dataSource)
}
```

## Step 4: Integrate with Bubble Tea

The integration is identical to the `List` and `Tree` components. You create the component, pass messages to its `Update` method, and call its `View` method.

```go
type App struct {
	table *table.Table
}

func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// ... handle quit and navigation keys ...

	// Pass all other messages to the table.
	var cmd tea.Cmd
	_, cmd = app.table.Update(msg)
	return app, cmd
}

func (app App) View() string {
	return app.table.View()
}
```

## What You'll Experience

-   **A Structured Table**: A properly rendered table with headers, borders, and aligned columns.
-   **Keyboard Navigation**: Full navigation capabilities (up/down, page, jump) work out of the box.
-   **Automatic Truncation**: Text that is too long for a column will be automatically truncated with an ellipsis (...).

## Complete Example

See the full working code for this guide in the examples directory:
[`docs/05-table-component/examples/basic-table/`](examples/basic-table/)

To run it:
```bash
cd docs/05-table-component/examples/basic-table
go run main.go
```

## What's Next?

You've successfully built a basic table. The next step is to see how VTable's data virtualization handles a much larger dataset, demonstrating the component's true power and performance.

**Next:** [Data Virtualization in Tables →](02-data-virtualization.md) 