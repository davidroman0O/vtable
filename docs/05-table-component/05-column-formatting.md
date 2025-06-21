# Column Formatting

## What We're Adding

Taking our cell constraints table and adding **custom cell formatters** that transform raw data into visually enhanced display text. We'll add icons, formatting, and conditional styling while keeping the implementation clean and simple.

## Why Column Formatting Matters

Column formatters let you:
- **Add visual indicators** with icons and symbols for better data recognition
- **Format numbers** with proper currency, percentages, and separators
- **Enhance readability** with consistent iconography and styling
- **Preserve data integrity** while improving visual presentation
- **Create professional displays** that are both functional and attractive

## Step 1: Simple Stateless Formatters

Replace basic column display with formatted versions using clean, pure functions:

```go
// Simple formatter - adds icon to employee names
func nameFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
    return "👤 " + cellValue
}

// Department formatter with conditional icons
func deptFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
    icons := map[string]string{
        "Engineering": "🔧",
        "Marketing":   "📢",
        "Sales":       "💼",
        "HR":          "👥",
        "Finance":     "💰",
        "Operations":  "⚙️",
    }
    if icon, exists := icons[cellValue]; exists {
        return icon + " " + cellValue
    }
    return "🏢 " + cellValue
}

// Status formatter with color-coded indicators
func statusFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
    switch cellValue {
    case "Active":
        return "🟢 " + cellValue
    case "On Leave":
        return "🟡 " + cellValue
    case "Remote":
        return "🔵 " + cellValue
    default:
        return "⚪ " + cellValue
    }
}

// Salary formatter with tier icons and number formatting
func salaryFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
    if salary, err := strconv.Atoi(cellValue); err == nil {
        formatted := "$" + formatNumber(salary)
        if salary >= 100000 {
            return "💎 " + formatted // Diamond tier
        } else if salary >= 75000 {
            return "💰 " + formatted // Gold tier
        } else if salary >= 50000 {
            return "💵 " + formatted // Silver tier
        } else {
            return "💳 " + formatted // Bronze tier
        }
    }
    return cellValue
}

// Date formatter with calendar icon
func dateFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
    return "📅 " + cellValue
}

// Helper function for number formatting
func formatNumber(n int) string {
    str := strconv.Itoa(n)
    if len(str) > 3 {
        return str[:len(str)-3] + "," + str[len(str)-3:]
    }
    return str
}
```

## Step 2: Apply Formatters to Table

Set up formatters during table initialization using the proper BubbleTea pattern:

```go
func (app App) Init() tea.Cmd {
    return tea.Batch(
        app.table.Init(),
        app.table.Focus(),
        // Apply formatters once during initialization
        core.CellFormatterSetCmd(0, nameFormatter),   // Employee column
        core.CellFormatterSetCmd(1, deptFormatter),   // Department column
        core.CellFormatterSetCmd(2, statusFormatter), // Status column
        core.CellFormatterSetCmd(3, salaryFormatter), // Salary column
        core.CellFormatterSetCmd(4, dateFormatter),   // Date column
    )
}
```

## Step 3: Enhanced Employee Data

Update the data source to include the fields needed for rich formatting:

```go
type Employee struct {
    ID          string
    Name        string
    Department  string
    Status      string
    Salary      int
    HireDate    time.Time
    Performance string
    Location    string
}

func NewLargeEmployeeDataSource(totalCount int) *LargeEmployeeDataSource {
    data := make([]Employee, totalCount)
    departments := []string{"Engineering", "Marketing", "Sales", "HR", "Finance", "Operations"}
    statuses := []string{"Active", "On Leave", "Remote"}
    performances := []string{"Excellent", "Good", "Average", "Needs Improvement"}
    locations := []string{"New York", "San Francisco", "Austin", "Seattle", "Boston", "Denver"}
    
    firstNames := []string{"Alice", "Bob", "Carol", "David", "Eve", "Frank", "Grace", "Henry", "Ivy", "Jack"}
    lastNames := []string{"Johnson", "Smith", "Davis", "Wilson", "Brown", "Miller", "Lee", "Taylor", "Chen", "Roberts"}

    for i := 0; i < totalCount; i++ {
        daysAgo := rand.Intn(3650) // Random hire date within last 10 years
        hireDate := time.Now().AddDate(0, 0, -daysAgo)

        data[i] = Employee{
            ID:          fmt.Sprintf("emp-%d", i+1),
            Name:        fmt.Sprintf("%s %s", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))]),
            Department:  departments[rand.Intn(len(departments))],
            Status:      statuses[rand.Intn(len(statuses))],
            Salary:      45000 + rand.Intn(100000), // $45k-$145k range
            HireDate:    hireDate,
            Performance: performances[rand.Intn(len(performances))],
            Location:    locations[rand.Intn(len(locations))],
        }
    }

    return &LargeEmployeeDataSource{
        totalEmployees: totalCount,
        data:           data,
        selectedItems:  make(map[string]bool),
        recentActivity: make([]string, 0),
    }
}

// Convert Employee to TableRow for display
func (ds *LargeEmployeeDataSource) employeeToTableRow(emp Employee) core.TableRow {
    return core.TableRow{
        ID: emp.ID,
        Cells: []string{
            emp.Name,
            emp.Department,
            emp.Status,
            fmt.Sprintf("%d", emp.Salary),              // Raw number for formatter
            emp.HireDate.Format("Jan 2006"),            // Basic date format
        },
    }
}
```

## Step 4: Complete App Structure

Clean, simple app structure following BubbleTea patterns:

```go
type App struct {
    table         *table.Table
    dataSource    *LargeEmployeeDataSource
    statusMessage string
}

func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return app, tea.Quit
        case " ", "enter":
            return app, core.SelectCurrentCmd()
        case "ctrl+a":
            return app, core.SelectAllCmd()
        case "c":
            return app, core.SelectClearCmd()
        default:
            var cmd tea.Cmd
            _, cmd = app.table.Update(msg)
            return app, cmd
        }
    default:
        var cmd tea.Cmd
        _, cmd = app.table.Update(msg)
        return app, cmd
    }
}

func (app App) View() string {
    var sections []string
    sections = append(sections, "Column Formatting Demo - Simple & Working")
    sections = append(sections, "")
    sections = append(sections, app.table.View())
    sections = append(sections, "")
    sections = append(sections, "Controls: ↑↓/jk=move, Space=select, ctrl+a=select all, c=clear, q=quit")
    sections = append(sections, "Formatting: Simple emoji icons for each column type")
    return strings.Join(sections, "\n")
}
```

## Formatter Details

### Employee Column (👤)
- Consistent person icon for all employees
- Simple, clean presentation

### Department Column  
- **🔧 Engineering**: Technical teams
- **📢 Marketing**: Marketing and communications  
- **💼 Sales**: Sales and business development
- **👥 HR**: Human resources
- **💰 Finance**: Finance and accounting
- **⚙️ Operations**: Operations and logistics
- **🏢 Default**: Fallback for unknown departments

### Status Column
- **🟢 Active**: Currently working
- **🟡 On Leave**: Temporarily away
- **🔵 Remote**: Working remotely  
- **⚪ Unknown**: Default for unrecognized status

### Salary Column
- **💎 $100K+**: Diamond tier (top earners)
- **💰 $75K-$100K**: Gold tier (high earners)
- **💵 $50K-$75K**: Silver tier (mid-range)
- **💳 Under $50K**: Bronze tier (entry level)
- Includes comma formatting for readability

### Date Column
- **📅**: Calendar icon for hire dates
- Simple, consistent presentation

## Adding Custom Formatters

To create your own formatter:

```go
// 1. Create the formatter function
func phoneFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
    if len(cellValue) == 10 {
        return fmt.Sprintf("📞 (%s) %s-%s", 
            cellValue[:3], 
            cellValue[3:6], 
            cellValue[6:])
    }
    return "📞 " + cellValue
}

// 2. Add column to table config
{Title: "Phone", Field: "phone", Width: 18, Alignment: core.AlignCenter}

// 3. Apply formatter in Init()
core.CellFormatterSetCmd(5, phoneFormatter)
```

## Key Principles

**Stateless Design**: Formatters are pure functions with no external dependencies. They receive all needed data through parameters.

**Simple Logic**: Each formatter handles one specific transformation. Complex formatting should be broken into multiple simple formatters.

**Performance**: Use simple string operations and avoid complex computations in the render loop.

**Consistency**: Apply formatting patterns consistently across similar data types.

## Try It Yourself

1. Run the example: `cd docs/05-table-component/examples/column-formatting && go run main.go`
2. Navigate with arrow keys or j/k
3. Select employees with space/enter
4. Try bulk operations with ctrl+a and c
5. Modify formatters to experiment with different icons and styles

## What's Next

In the next section, we'll explore [Table Styling](06-table-styling.md) where we'll add comprehensive theming, color schemes, and visual customization to our formatted table.

## Key Takeaways

- **Formatters transform display without changing data** - original values remain intact
- **Apply formatters once during initialization** - no complex state management needed
- **Keep formatters simple and stateless** - easier to test, debug, and maintain
- **Use consistent iconography** - improves user recognition and experience
- **Follow BubbleTea patterns** - value receivers, proper command handling, clean separation