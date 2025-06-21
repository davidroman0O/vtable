# Table Styling Example

A comprehensive demonstration of VTable's styling and theming capabilities, building on the column-formatting example with theme switching, border controls, and visual customization.

## Features

- **Theme Switching**: 4 distinct visual themes (Default, Heavy, Minimal, Retro)
- **Border Controls**: Granular control over table borders and separators
- **Space Removal**: Compact layout options for dense displays
- **Color Schemes**: Different color palettes for headers, cells, and highlights
- **Column Formatting**: Same emoji formatters from the previous example
- **Interactive Selection**: Multi-select with formatted data display

## Themes

### Default Theme
- **Style**: Clean Unicode box drawing (`┌─┐ │ └─┘`)
- **Colors**: Gray borders, purple headers, clean text
- **Best For**: Professional applications, documentation

### Heavy Theme  
- **Style**: Double-line borders (`╔═╗ ║ ╚═╝`)
- **Colors**: Cyan borders, yellow headers, bright text
- **Best For**: Important data, emphasis tables

### Minimal Theme
- **Style**: Borderless design (invisible borders)
- **Colors**: Black text, subtle highlighting
- **Best For**: Content-focused displays, embedded tables

### Retro Theme
- **Style**: ASCII computing style (`+-+ | +-+`)
- **Colors**: Cyan/magenta retro colors
- **Best For**: Terminal applications, debugging tools

## Controls

### Navigation
- **↑↓ or j/k**: Navigate up/down
- **Space or Enter**: Toggle selection
- **Ctrl+A**: Select all
- **c**: Clear selection

### Styling Controls
- **T**: Cycle through themes
- **B**: Toggle all borders on/off
- **1**: Toggle top border
- **2**: Toggle bottom border
- **3**: Toggle header separator
- **4**: Toggle top space removal (compact)
- **5**: Toggle bottom space removal (compact)
- **q**: Quit

## How It Works

### Theme System

```go
type TableTheme struct {
    Name        string
    Description string
    BorderChars core.BorderChars
    Colors      ThemeColors
}

type ThemeColors struct {
    BorderColor    string
    HeaderColor    string
    CellColor      string
    CursorColor    string
    SelectionColor string
}
```

### Theme Application

Themes are converted to VTable's theme system and applied dynamically:

```go
func convertToVTableTheme(theme TableTheme) core.Theme {
    return core.Theme{
        HeaderStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.HeaderColor)).Bold(true),
        CellStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.CellColor)),
        CursorStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.CursorColor)).Bold(true),
        SelectedStyle:      lipgloss.NewStyle().Background(lipgloss.Color(theme.Colors.SelectionColor)).Foreground(lipgloss.Color("230")),
        BorderChars:        theme.BorderChars,
        BorderColor:        theme.Colors.BorderColor,
        // ... other styling options
    }
}
```

### Border Controls

VTable provides granular border control through specific commands:

```go
// Border visibility
case "b":
    app.showBorders = !app.showBorders
    return app, core.BorderVisibilityCmd(app.showBorders)

case "1":
    app.showTopBorder = !app.showTopBorder
    return app, core.TopBorderVisibilityCmd(app.showTopBorder)

// Space removal for compact layouts
case "4":
    app.removeTopSpace = !app.removeTopSpace
    return app, core.TopBorderSpaceRemovalCmd(app.removeTopSpace)
```

### Column Formatting

Uses the same formatters from the column-formatting example:

- **👤** Employee names with person icons
- **🔧 📢 💼 👥 💰 ⚙️** Department-specific icons
- **🟢 🟡 🔵** Status indicators
- **💎 💰 💵 💳** Salary tier icons
- **📅** Date formatting with calendar icons

## Running the Example

```bash
cd docs/05-table-component/examples/table-styling
go run main.go
```

## Try This

1. **Press T** repeatedly to cycle through all four themes
2. **Press B** to see the table without any borders
3. **Use 1, 2, 3** to control individual border sections
4. **Try 4 and 5** to see compact mode with space removal
5. **Navigate and select** to see how themes affect interaction

## Design Principles

- **Consistent formatting**: Same emoji formatters work across all themes
- **Progressive enhancement**: Builds cleanly on column-formatting
- **Terminal compatibility**: Themes adapt to different color capabilities
- **User control**: Interactive switching allows preference testing
- **Visual hierarchy**: Each theme emphasizes different aspects of the data

This example demonstrates how visual styling can enhance data presentation without affecting core functionality, making tables both beautiful and highly usable. 