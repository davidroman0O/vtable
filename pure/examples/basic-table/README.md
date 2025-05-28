# Basic Table Example

This example demonstrates the core functionality of the VTable `Table` component with a comprehensive employee management interface.

## Features Demonstrated

### 🏗️ **Table Architecture**
- **Clean separation**: Data source, configuration, and UI logic
- **Type-safe implementation**: Generic `Table[Employee]` with proper type constraints
- **Reusable patterns**: Following List/TreeList architecture for consistency

### 📊 **Data Management**
- **Large dataset**: 500 employee records to demonstrate chunking
- **Chunked loading**: 50 items per chunk for efficient memory usage
- **Smart caching**: Automatic chunk loading/unloading based on viewport
- **Real-time updates**: Immediate response to sorting and filtering

### 🎯 **Table Operations**
- **Sorting**: Multiple sort options (Name, Salary, ID) with ascending/descending
- **Filtering**: Department, active status, and salary range filters
- **Selection**: Multiple selection mode with visual feedback
- **Navigation**: Full keyboard navigation with page up/down, home/end

### 🎨 **Professional Rendering**
- **Bordered table**: Clean borders with proper alignment
- **Column formatting**: Right-aligned numbers, centered dates, left-aligned text
- **Status indicators**: Visual distinction for active/inactive employees
- **Responsive layout**: Proper column width management

## Data Model

```go
type Employee struct {
    ID         int
    Name       string
    Department string
    Position   string
    Salary     int
    StartDate  time.Time
    Active     bool
}
```

## Key Components

### 1. **EmployeeDataSource**
Implements `TableDataSource[Employee]` interface:
- Chunk loading with filtering and sorting
- Selection management
- Cell value formatting
- Real-time data operations

### 2. **Table Configuration**
```go
// List configuration (viewport, chunking, selection)
listConfig := vtable.NewListConfigBuilder().
    WithViewportHeight(15).
    WithChunkSize(50).
    WithSelectionMode(vtable.SelectionMultiple).
    WithMaxWidth(120).
    Build()

// Table configuration (columns, borders, styling)
tableConfig := vtable.DefaultTableConfig()
tableConfig.Columns = dataSource.GetColumns()
tableConfig.ShowHeader = true
tableConfig.ShowBorders = true
```

### 3. **Interactive Controls**
- **Navigation**: Arrow keys, Page Up/Down, Home/End
- **Selection**: Space/Enter to select, Ctrl+A for select all
- **Sorting**: 'S' key cycles through sort options
- **Filtering**: 'F' key cycles through filter options
- **Refresh**: 'R' key refreshes data
- **Help**: 'H' key toggles help display

## Column Configuration

| Column | Width | Alignment | Field | Format |
|--------|-------|-----------|-------|---------|
| ID | 6 | Right | `id` | Integer |
| Name | 20 | Left | `name` | String |
| Department | 15 | Left | `department` | String |
| Position | 18 | Left | `position` | String |
| Salary | 10 | Right | `salary` | `$123,456` |
| Start Date | 12 | Center | `start_date` | `2006-01-02` |
| Status | 8 | Center | `active` | `Active`/`Inactive` |

## Sorting Options

Press 'S' to cycle through:
1. **None** - Original order
2. **Name A-Z** - Alphabetical ascending
3. **Name Z-A** - Alphabetical descending  
4. **Salary Low-High** - Salary ascending
5. **Salary High-Low** - Salary descending
6. **ID 1-999** - ID ascending

## Filtering Options

Press 'F' to cycle through:
1. **None** - All employees
2. **Engineering** - Engineering department only
3. **Active** - Active employees only
4. **High Salary** - Salary > $80,000

## Performance Features

### Chunked Loading
- **Viewport-based**: Only loads visible data plus buffer
- **Smart boundaries**: Configurable bounding areas before/after viewport
- **Automatic unloading**: Removes chunks outside bounding area
- **Loading indicators**: Visual feedback during chunk operations

### Memory Efficiency
- **50 items per chunk**: Optimal balance of performance and memory
- **Lazy loading**: Data loaded on-demand as user navigates
- **Cursor preservation**: Maintains position during sort/filter operations
- **Efficient updates**: Minimal re-rendering on data changes

## Running the Example

```bash
cd pure/examples/basic-table
go run main.go
```

### Navigation Keys (Actually Working in this Example)

**Navigation:**
- **j/↑** - Move up one row
- **k/↓** - Move down one row  
- **h** - Page up
- **l** - Page down
- **g** - Jump to first row
- **G** - Jump to last row
- **Space/Enter** - Select current row
- **a** - Select all rows
- **c** - Clear selection

**Actions:**
- **s** - Cycle sort options
- **f** - Cycle filter options
- **r** - Refresh data
- **?** - Toggle help
- **q/Ctrl+C** - Quit

> **Note**: This example uses vim-style navigation keys: `j/k` for up/down movement and `h/l` for page up/down navigation. These keys are handled directly in the app's Update method and send commands to the table component.

## Key Learnings

### 1. **Data Source Implementation**
- Implement all `TableDataSource[T]` methods
- Handle filtering and sorting in `applyFiltersAndSort()`
- Convert between typed and `any` data for chunk loading
- Maintain selection state across operations

### 2. **Configuration Management**
- Use builder pattern for clean configuration
- Separate list config (viewport) from table config (display)
- Define columns with proper alignment and formatting
- Configure chunking parameters for your dataset size

### 3. **Message Handling**
- Forward table messages to table component
- Handle custom key bindings in main model
- Use status messages for user feedback
- Preserve table state across updates

### 4. **Professional UI**
- Consistent styling with lipgloss
- Clear help text and status indicators
- Responsive layout with proper spacing
- Visual feedback for all operations

This example serves as a comprehensive template for building professional table interfaces with the VTable library, demonstrating all core features in a real-world scenario. 