# Component Rendering: Making It Look Good

VTable components render themselves using **Lipgloss** for styling and a **component-based system** for flexible appearance control. This section teaches you how VTable handles visual output and how to customize it.

## The View() Method

Every VTable component has a `View()` method that returns a styled string ready for terminal display:

```go
// In your Bubble Tea app
func (m MyApp) View() string {
    // VTable components render themselves
    return m.list.View()  // Returns the complete styled list
}
```

**Key point**: You don't manually style or format VTable output - components handle all rendering internally using their configurations.

## Styling with Lipgloss

VTable uses **Lipgloss** (from Charm) for all styling. Lipgloss styles are configured once and applied automatically:

```go
import "github.com/charmbracelet/lipgloss"

// Create a style
cursorStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("205")).
    Bold(true)

// VTable applies it automatically when rendering cursor items
```

**What you can style:**
- **Colors**: Foreground, background, borders
- **Text effects**: Bold, italic, underline
- **Layout**: Padding, margins, alignment
- **Borders**: Different border styles and characters

## Default Themes

VTable provides ready-to-use themes from the `config` package:

```go
import "github.com/davidroman0O/vtable/config"

// For lists
styleConfig := config.DefaultStyleConfig()
// Includes: CursorStyle, SelectedStyle, DefaultStyle, etc.

// For tables  
theme := config.DefaultTheme()
// Includes: HeaderStyle, CellStyle, BorderChars, etc.
```

**What's included:**
- **List styles**: Cursor, selected, normal, loading, error states
- **Table themes**: Headers, cells, borders, alternating rows
- **Colors**: Sensible defaults that work in most terminals
- **Border characters**: Unicode box-drawing characters

## Customizing Styles

### List Component Styling

```go
import (
    "github.com/davidroman0O/vtable/config"
    "github.com/charmbracelet/lipgloss"
)

// Start with defaults and customize
styleConfig := config.DefaultStyleConfig()

// Custom cursor style
styleConfig.CursorStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#FF6B35")).
    Background(lipgloss.Color("#1A1A1A")).
    Bold(true)

// Custom selection style
styleConfig.SelectedStyle = lipgloss.NewStyle().
    Background(lipgloss.Color("#2D3748")).
    Foreground(lipgloss.Color("#F7FAFC"))

// Apply to your list
listConfig := config.DefaultListConfig()
listConfig.StyleConfig = styleConfig
```

### Table Component Theming

```go
// Start with default theme and customize
theme := config.DefaultTheme()

// Custom header styling
theme.HeaderStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#4A90E2")).
    Background(lipgloss.Color("#2C3E50")).
    Bold(true).
    Padding(0, 1)

// Custom border characters (rounded corners)
theme.BorderChars = core.BorderChars{
    Horizontal: "─",
    Vertical:   "│", 
    TopLeft:    "╭",
    TopRight:   "╮",
    BottomLeft: "╰", 
    BottomRight:"╯",
    // ... other border chars
}

// Apply to your table
tableConfig := config.DefaultTableConfig() 
tableConfig.Theme = theme
```

## Component-Based Rendering

VTable uses a **component system** where different parts of each row are rendered by separate components:

### List Components
- **Cursor**: Shows `► ` or spaces for cursor position
- **Enumerator**: Numbers, bullets, or custom markers  
- **Content**: The actual item text
- **Background**: Optional background styling

### Table Components  
- **Cursor**: Row selection indicator
- **Cells**: The actual table data
- **Borders**: Table borders and separators
- **Background**: Row highlighting

### Enabling Advanced Rendering

```go
// For lists - use component-based rendering
import "github.com/davidroman0O/vtable/list"

renderConfig := list.BulletListConfig()  // Pre-configured bullet list
// or
renderConfig := config.DefaultListRenderConfig()  // Build your own

// For tables - component rendering is built-in
tableConfig := config.DefaultTableConfig()
tableConfig.ShowBorders = true
tableConfig.FullRowHighlighting = true
```

## Color Support

VTable automatically detects your terminal's color capabilities:

### ANSI Colors (16 colors)
```go
lipgloss.Color("1")   // Red
lipgloss.Color("10")  // Bright green
```

### 256 Colors
```go  
lipgloss.Color("196") // Bright red
lipgloss.Color("39")  // Blue
```

### True Color (24-bit)
```go
lipgloss.Color("#FF6B35") // Orange
lipgloss.Color("#4A90E2") // Blue
```

### Adaptive Colors
```go
lipgloss.AdaptiveColor{
    Light: "#333333",  // Dark text on light background
    Dark:  "#FFFFFF",  // Light text on dark background  
}
```

**VTable automatically chooses the best color based on your terminal's capabilities.**

## Practical Examples

### Dark Theme List
```go
darkStyle := config.DefaultStyleConfig()
darkStyle.CursorStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#000000")).
    Background(lipgloss.Color("#00FF7F")).
    Bold(true)
darkStyle.SelectedStyle = lipgloss.NewStyle().
    Background(lipgloss.Color("#404040")).
    Foreground(lipgloss.Color("#FFFFFF"))
darkStyle.DefaultStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#CCCCCC"))
```

### Minimal Table
```go
minimalTheme := config.DefaultTheme()
minimalTheme.BorderChars = core.BorderChars{
    Horizontal: " ",
    Vertical:   " ",  
    // All border chars set to spaces = no visible borders
}
minimalTheme.HeaderStyle = lipgloss.NewStyle().
    Bold(true).
    Underline(true)
```

### Status-Colored Table
```go
statusTheme := config.DefaultTheme()
// You'd customize this based on data content
statusTheme.CellStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("252"))
statusTheme.ErrorStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#FF4444")).
    Bold(true)
statusTheme.LoadingStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#FFA500")).
    Italic(true)
```

## What You Need to Know

1. **Import `config` package** for default themes and styles
2. **Use Lipgloss styles** - VTable integrates directly with Lipgloss
3. **Start with defaults** and customize what you need
4. **Colors auto-adapt** to terminal capabilities  
5. **Component system** handles complex rendering automatically
6. **View() method** returns the final styled output

VTable handles all the complex rendering logic. You just configure styles and themes, then call `View()` to get beautiful terminal output.

**Next:** [Complete Core Concepts →](../03-list-component/README.md) 