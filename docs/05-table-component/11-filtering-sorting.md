# Filtering and Sorting

## What We're Adding

Building on our column management example, we're adding powerful data manipulation capabilities:
- **Column-based sorting**: Sort by any column using active cell navigation
- **Cumulative filtering**: Apply multiple filters simultaneously using number keys
- **Text search**: Filter data across multiple fields with live search
- **Active cell navigation**: Move between columns with arrow keys to control sorting
- **Smart UI**: Real-time display of active column, filters, and sort state

This transforms a static table into an interactive data exploration tool with intuitive controls.

## New Concepts

### Active Cell Sorting
Instead of cycling through predefined sorts, you navigate to any column and sort it:

```go
// Navigate to column, then sort it
case "left", "right":   // Move between columns
case "s":               // Sort the active column (asc→desc→clear)
```

### Cumulative Filtering  
Filters can be combined rather than replaced, allowing complex data views:

```go
// Multiple filters active simultaneously
activeFilters := map[string]bool{
    "engineering": true,  // Department filter
    "active_only": true,  // Status filter  
    "high_salary": true,  // Salary filter
}
// Result: Shows only "Active Engineering employees with high salaries"
```

### Number Key Filter System
Pre-configured filters mapped to number keys for quick access:

```go
// Department filters (1-5)
case "1": toggleFilter("engineering")
case "2": toggleFilter("marketing") 
case "3": toggleFilter("sales")

// Status filters (6-7)  
case "6": toggleFilter("active_only")
case "7": toggleFilter("remote_only")

// Salary filters (8-9)
case "8": toggleFilter("high_salary")  // ≥$75k
case "9": toggleFilter("low_salary")   // <$65k
```

## Code Changes

### 1. Enhanced Data Source with Cumulative Filtering
Extend filtering to support multiple simultaneous filters:

```go
func (ds *EmployeeDataSource) matchesFilters(emp Employee) bool {
    for field, value := range ds.activeFilters {
        switch field {
        case "engineering":
            if emp.Department != "Engineering" { return false }
        case "marketing":
            if emp.Department != "Marketing" { return false }
        case "active_only":
            if emp.Status != "Active" { return false }
        case "high_salary":
            if emp.Salary < 75000 { return false }
        // ... more filter types
        }
    }
    return true // All active filters must pass
}
```

### 2. Column-Based Sorting System
Replace predefined sort cycles with active column sorting:

```go
func (m AppModel) sortByActiveColumn() (tea.Model, tea.Cmd) {
    // Get current active column
    _, _, currentColumn, _ := m.table.GetHorizontalScrollState()
    
    // Map column index to field name
    columnFields := []string{"id", "name", "department", "status", "salary", "email", "phone"}
    
    if currentColumn < len(columnFields) {
        field := columnFields[currentColumn]
        
        // Toggle sort direction: asc → desc → clear
        if m.currentSort == field {
            if m.currentSortDir == "asc" {
                m.currentSortDir = "desc"
            } else {
                m.currentSort = ""  // Clear sorting
                m.dataSource.ClearSort()
                return m, core.DataRefreshCmd()
            }
        } else {
            m.currentSort = field
            m.currentSortDir = "asc"
        }
        
        m.dataSource.SetSort([]string{field}, []string{m.currentSortDir})
        return m, core.DataRefreshCmd()
    }
}
```

### 3. Number Key Filter Controls
Add cumulative filtering with number keys:

```go
// In Update method
case "1": return m.toggleFilter("engineering", "Engineering Dept")
case "2": return m.toggleFilter("marketing", "Marketing Dept") 
case "3": return m.toggleFilter("sales", "Sales Dept")
case "4": return m.toggleFilter("finance", "Finance Dept")
case "5": return m.toggleFilter("hr", "HR Dept")
case "6": return m.toggleFilter("active_only", "Active Status")
case "7": return m.toggleFilter("remote_only", "Remote Status")
case "8": return m.toggleFilter("high_salary", "High Salary (≥$75k)")
case "9": return m.toggleFilter("low_salary", "Lower Salary (<$65k)")
case "0": return m.clearAllFilters()

func (m AppModel) toggleFilter(filterKey, filterName string) (tea.Model, tea.Cmd) {
    if m.activeFilters[filterKey] {
        // Remove filter
        m.activeFilters[filterKey] = false
        m.dataSource.ClearFilter(filterKey)
    } else {
        // Add filter  
        m.activeFilters[filterKey] = true
        m.dataSource.SetFilter(filterKey, true)
    }
    return m, core.DataRefreshCmd()
}
```

### 4. Active Cell Navigation
Add arrow key navigation for column movement:

```go
case "left":
    // Navigate to previous column (active cell)
    return m, core.PrevColumnCmd()

case "right":
    // Navigate to next column (active cell)  
    return m, core.NextColumnCmd()

case "s":
    // Sort by current active column
    return m.sortByActiveColumn()
```

### 5. Enhanced Status Display
Show real-time state of active column, filters, and sorting:

```go
// Get current active column info
_, _, currentColumn, _ := m.table.GetHorizontalScrollState()
columnNames := []string{"ID", "Name", "Department", "Status", "Salary", "Email", "Phone"}
currentColumnName := columnNames[currentColumn]

// Display comprehensive state
fmt.Sprintf("Data: %d/%d | Active Column: %s | Sort: %s | Filters: %s",
    m.dataSource.filteredTotal,
    len(m.dataSource.employees), 
    currentColumnName,
    m.getSortDescription(),      // e.g., "Name↑" or "Salary↓"
    m.getActiveFiltersDescription(), // e.g., "Eng+Active+High$"
)
```

## Key Features Explained

### Column-Based Sorting Workflow
1. **Navigate**: Use `←→` arrows to move to any column
2. **Sort**: Press `s` to sort that column (asc→desc→clear)
3. **Visual feedback**: Active cell highlights current column
4. **Status display**: Shows which column is active and sort direction

### Cumulative Filtering System
- **Combine filters**: Press multiple number keys to layer filters
- **Example workflow**: Press `1` + `6` + `8` = "Engineering + Active + High Salary"
- **Toggle on/off**: Same key removes filter if already active
- **Clear all**: Press `0` to remove all filters instantly

### Smart Filter Categories
- **Departments (1-5)**: Engineering, Marketing, Sales, Finance, HR
- **Status (6-7)**: Active employees, Remote workers
- **Salary (8-9)**: High earners (≥$75k), Lower salaries (<$65k)
- **Search (/)**: Text search across Name, Department, Email

### Enhanced Navigation
- **Row navigation**: `↑↓` or `jk` to move between employees
- **Column navigation**: `←→` or `.,` to move between fields
- **Active cell**: Visual highlight shows exactly where you are
- **Sort control**: Current column determines what gets sorted

## Core Commands Used

```go
// Active cell navigation (NEW)
core.PrevColumnCmd() tea.Cmd  // Move to previous column
core.NextColumnCmd() tea.Cmd  // Move to next column

// Data manipulation (enhanced)
core.DataRefreshCmd() tea.Cmd // Refresh after filter/sort changes

// Standard navigation (inherited)
core.CursorUpCmd() tea.Cmd
core.CursorDownCmd() tea.Cmd
```

## Enhanced Status Display
Real-time feedback shows complete table state:

```go
// Shows: "Data: 8/12 | Active Column: Salary | Sort: Salary↓ | Filters: Eng+Active+High$"
status := fmt.Sprintf("Data: %d/%d | Active Column: %s | Sort: %s | Filters: %s", 
    m.dataSource.filteredTotal,    // Filtered count
    m.dataSource.totalEmployees,   // Total count  
    currentColumnName,             // Which column is active
    m.getSortDescription(),        // Current sort with direction
    m.getActiveFiltersDescription(), // Compact filter list
)
```

## Controls

| Key | Action |
|-----|--------|
| **NAVIGATION** ||
| `←` `→` | Move between columns (active cell) |
| `↑` `↓` or `j` `k` | Move between rows |
| `.` `,` | Alternative column navigation |
| **SORTING** ||
| `s` | Sort active column (asc→desc→clear) |
| `S` | Clear all sorting |
| **FILTERING (Cumulative)** ||
| `1` | Toggle Engineering department |
| `2` | Toggle Marketing department |
| `3` | Toggle Sales department |
| `4` | Toggle Finance department |
| `5` | Toggle HR department |
| `6` | Toggle Active status only |
| `7` | Toggle Remote status only |
| `8` | Toggle High salary (≥$75k) |
| `9` | Toggle Low salary (<$65k) |
| `0` | Clear all filters |
| **SEARCH** ||
| `/` | Enter text search mode |
| `Enter` | Apply search (in search mode) |
| `Esc` | Cancel search (in search mode) |

## Try It Yourself

### Basic Workflow
1. **Navigate columns**: Use `←→` to move to the Salary column
2. **Sort by salary**: Press `s` to sort high→low→clear
3. **Add filters**: Press `1` + `6` for "Engineering + Active" 
4. **Combine more**: Press `8` to add "High Salary"
5. **Search within results**: Press `/` then type "alice"

### Advanced Filtering
1. **Complex filter**: `1` + `6` + `8` = Engineering + Active + High Salary
2. **Modify filter**: Press `1` again to remove Engineering 
3. **Different view**: Press `2` + `7` = Marketing + Remote workers
4. **Reset**: Press `0` to clear all filters

### Column Sorting  
1. **Any column**: Navigate to Email column with `→→→→→`
2. **Sort emails**: Press `s` to sort alphabetically
3. **Department sort**: Navigate to Department, press `s`
4. **Status awareness**: UI shows "Active Column: Department"

## What's Next

In the next section, we'll explore [Debug and Observability](12-debug-observability.md) to add monitoring and troubleshooting features to our interactive data table.

## Running the Example

```bash
cd docs/05-table-component/examples/filtering-sorting
go run .
```

This example demonstrates how to build sophisticated data exploration tools with:
- **Intuitive navigation**: Arrow keys for both rows and columns
- **Smart sorting**: Sort any column by navigating to it
- **Layered filtering**: Combine multiple criteria simultaneously  
- **Real-time feedback**: Always know what filters and sorting are active
- **Efficient workflow**: Number keys for instant filter toggling 