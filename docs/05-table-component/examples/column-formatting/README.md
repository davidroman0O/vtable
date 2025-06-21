# Column Formatting Example

A clean, simple demonstration of column formatting in VTable using stateless formatters.

## Features

- **Employee Icons** (👤): Simple person icon for all employee names
- **Department Icons**: Specific emoji for each department type
  - 🔧 Engineering
  - 📢 Marketing  
  - 💼 Sales
  - 👥 HR
  - 💰 Finance
  - ⚙️ Operations
- **Status Indicators**: Color-coded status with emoji
  - 🟢 Active
  - 🟡 On Leave
  - 🔵 Remote
- **Salary Tiers**: Icon-based salary ranges
  - 💎 $100K+ (Diamond tier)
  - 💰 $75K+ (Gold tier)
  - 💵 $50K+ (Silver tier)
  - 💳 Under $50K (Bronze tier)
- **Date Formatting**: Calendar icon with formatted dates

## How It Works

### Simple, Stateless Formatters

```go
func nameFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
    return "👤 " + cellValue
}

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
```

### Applying Formatters

Formatters are applied once during initialization:

```go
func (app App) Init() tea.Cmd {
    return tea.Batch(
        app.table.Init(),
        app.table.Focus(),
        // Set formatters once
        core.CellFormatterSetCmd(0, nameFormatter),
        core.CellFormatterSetCmd(1, deptFormatter),
        core.CellFormatterSetCmd(2, statusFormatter),
        core.CellFormatterSetCmd(3, salaryFormatter),
        core.CellFormatterSetCmd(4, dateFormatter),
    )
}
```

## Running the Example

```bash
cd docs/05-table-component/examples/column-formatting
go run main.go
```

## Controls

- **↑↓ or j/k**: Navigate up/down
- **Space or Enter**: Toggle selection
- **Ctrl+A**: Select all
- **c**: Clear selection  
- **q**: Quit

## Design Principles

1. **Stateless formatters**: No external dependencies or state capture
2. **Simple logic**: Each formatter does one thing well
3. **BubbleTea compliance**: Proper value receivers and immutable updates
4. **Clean separation**: Formatters only handle visual presentation
5. **Performance**: No complex computations or state lookups in render loop

This example demonstrates the **right way** to implement column formatting - simple, clean, and following established patterns. 