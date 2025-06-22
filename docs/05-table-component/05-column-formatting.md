# The Table Component: Column Formatting

Column formatters allow you to transform raw data from your `DataSource` into a rich, visually informative display. This guide will show you how to use `SimpleCellFormatter` functions to add icons, apply conditional styling, and format data for better readability.

## What You'll Build

We will take our table of employee data and apply custom formatters to each column, turning plain text into a dashboard with clear visual indicators.

**Before:**
```
│ Employee 1        │ Engineering │ Active     │ $75000     │
```

**After (with formatting):**
```
│ 👤 Employee 1       │ 🔧 Engineering │ 🟢 Active    │ 💰 $75,000 │
```

## How It Works: The `SimpleCellFormatter`

A `SimpleCellFormatter` is a function that receives the raw string value for a single cell and returns a new, formatted string for display. VTable handles the complexities of layout and truncation for you.

```go
// The function signature for a simple cell formatter.
type SimpleCellFormatter func(
	cellValue string,      // The raw string value for the cell.
	rowIndex int,          // The absolute index of the row.
	column core.TableColumn,  // The configuration for the current column.
	ctx core.RenderContext, // Global rendering context (themes, etc.).
	isCursor bool,         // Is the row under the cursor?
	isSelected bool,       // Is the row selected?
	isActiveCell bool,     // Is this the specific "active" cell?
) string
```
**The key benefit:** Your formatter function is simple and stateless. It receives all the context it needs to make rendering decisions.

## Step 1: Create Your Formatter Functions

Let's create a set of formatters, one for each column in our employee table.

#### Name Formatter
A simple formatter that adds a user icon.
```go
func nameFormatter(...) string {
    return "👤 " + cellValue
}
```

#### Department Formatter
Uses a map to return a department-specific icon.
```go
func deptFormatter(...) string {
    icons := map[string]string{
        "Engineering": "🔧", "Marketing": "📢", "Sales": "💼", /* ... */
    }
    if icon, exists := icons[cellValue]; exists {
        return icon + " " + cellValue
    }
    return "🏢 " + cellValue
}
```

#### Status Formatter
Uses a `switch` statement for conditional, color-coded icons.
```go
func statusFormatter(...) string {
    switch cellValue {
    case "Active": return "🟢 Active"
    case "On Leave": return "🟡 On Leave"
    // ...
    }
}
```

#### Salary Formatter
Parses the string value back to an integer to apply conditional formatting and adds a thousands separator.
```go
func salaryFormatter(...) string {
	if salary, err := strconv.Atoi(cellValue); err == nil {
		formatted := "$" + formatNumber(salary) // formatNumber adds commas
		if salary >= 100000 { return "💎 " + formatted }
		if salary >= 75000 { return "💰 " + formatted }
		// ...
	}
	return cellValue
}
```

## Step 2: Apply Formatters to the Table

The recommended way to apply formatters is by sending `core.CellFormatterSetCmd` commands during your application's initialization.

```go
// In your app's Init method:
func (app App) Init() tea.Cmd {
	return tea.Batch(
		app.table.Init(),
		app.table.Focus(),
		// Send a command for each column you want to format.
		core.CellFormatterSetCmd(0, nameFormatter),   // Column 0: Employee Name
		core.CellFormatterSetCmd(1, deptFormatter),   // Column 1: Department
		core.CellFormatterSetCmd(2, statusFormatter), // Column 2: Status
		core.CellFormatterSetCmd(3, salaryFormatter), // Column 3: Salary
		core.CellFormatterSetCmd(4, dateFormatter),   // Column 4: Hire Date
	)
}
```

## Step 3: Ensure `DataSource` Provides Raw Data

Your `DataSource` should provide the raw, unformatted data. For example, the salary should be a simple string of numbers (`"75000"`) so the `salaryFormatter` can parse and format it correctly.

```go
// In your DataSource:
func (ds *LargeEmployeeDataSource) employeeToTableRow(emp Employee) core.TableRow {
	return core.TableRow{
		ID: emp.ID,
		Cells: []string{
			emp.Name,
			emp.Department,
			emp.Status,
			fmt.Sprintf("%d", emp.Salary), // Provide raw number as a string
			emp.HireDate.Format("Jan 2006"),
		},
	}
}
```

## What You'll Experience

-   **Enhanced Readability**: Icons and visual cues make the table data much faster to scan and understand.
-   **Data Integrity**: The underlying data remains clean and unformatted, while the display is rich and informative.
-   **Clean Code**: Your formatting logic is encapsulated in small, stateless functions that are easy to test and maintain.

## Complete Example

See the full working code for this guide in the examples directory:
[`docs/05-table-component/examples/column-formatting/`](examples/column-formatting/)

To run it:
```bash
cd docs/05-table-component/examples/column-formatting
go run main.go
```

## What's Next?

Now that your table's content is beautifully formatted, the next step is to customize the table's overall appearance with themes, border styles, and color schemes.

**Next:** [Table Styling →](06-table-styling.md)