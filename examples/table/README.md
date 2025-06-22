# Simplified Table Formatter Demo

This example demonstrates the new **simplified table formatter system** that addresses the frustrations with the previous complex table implementation.

## Key Improvements

### 🎯 **Simplified Formatter Interfaces**

**Before (Complex):**
```go
func OldCellFormatter(cellValue string, rowIndex, columnIndex int, column TableColumn, ctx RenderContext, isCursor, isSelected, isTopThreshold, isBottomThreshold bool) string {
    // Manual constraint handling
    constraint := CellConstraint{Width: column.Width, Alignment: column.Alignment}
    constrainedValue := enforceCellConstraints(cellValue, constraint)
    // Manual styling...
    return style.Render(constrainedValue)
}
```

**After (Simplified):**
```go
func NewCellFormatter(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor, isSelected bool) string {
    // Just focus on styling - truncation is automatic!
    return style.Render(cellValue)
}
```

### ✨ **Automatic Truncation**

- **No manual constraint handling** - the table automatically truncates content with `...` if it exceeds column width
- **Unicode-aware** - proper handling of emoji and international characters
- **Consistent behavior** - all cells follow the same truncation rules

### 🎨 **Two Simple Formatter Types**

1. **`SimpleCellFormatter`** - For column cell formatting
2. **`SimpleHeaderFormatter`** - For header cell formatting

Both automatically handle:
- Width constraints and truncation
- Proper Unicode width calculation
- Consistent padding and alignment

## Example Usage

```go
// Simple cell formatter with automatic truncation
func ValueFormatter(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor, isSelected bool) string {
    // Parse and style the value
    var style lipgloss.Style
    if value, err := strconv.Atoi(cellValue); err == nil {
        if value < 50 {
            style = lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
        } else {
            style = lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
        }
    }
    
    // Apply row-level styling for cursor/selection
    if isCursor {
        style = style.Background(lipgloss.Color("blue"))
    }
    
    return style.Render(cellValue) // Automatic truncation!
}

// Simple header formatter with automatic truncation
func HeaderFormatter(column TableColumn, ctx RenderContext) string {
    headerText := "📊 " + column.Title
    style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("cyan"))
    return style.Render(headerText) // Automatic truncation!
}

// Set formatters easily
table.SetCellFormatter(1, ValueFormatter)
table.SetHeaderFormatter(1, HeaderFormatter)
```

## Running the Demo

```bash
go run main.go
```

### Demo Features

- **4 columns** with different formatter styles
- **Emoji headers** with automatic truncation
- **Color-coded values** based on content
- **Selection indicators** in the name column
- **Theme switching** (press `t`)
- **Automatic truncation** when content is too wide

### Key Bindings

- `j/k` or `↑/↓` - Navigate
- `Space` - Toggle selection
- `a` - Select all
- `c` - Clear selection
- `t` - Cycle themes
- `?` - Toggle help
- `q` - Quit

## Benefits

1. **Less Code** - No manual constraint handling
2. **More Consistent** - Automatic truncation everywhere
3. **Easier to Use** - Simple formatter signatures
4. **Better UX** - Proper Unicode handling
5. **Maintainable** - Clear separation of concerns

The table handles all the complex width calculations, truncation, and alignment automatically, so you can focus on just the styling and content formatting! 