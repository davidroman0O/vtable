# Table Styling

## What We're Adding

Taking our column-formatted table and adding **comprehensive styling capabilities**. We'll add theme switching, border controls, color schemes, and visual customization while keeping the same formatted data and functionality.

## Why Table Styling Matters

Table styling lets you:
- **Apply visual themes** that match your application's design language
- **Control border appearance** with granular visibility and styling options
- **Customize color schemes** for headers, cells, selections, and highlights
- **Create compact layouts** with space removal controls
- **Support different terminals** with adaptive color choices
- **Build professional interfaces** that enhance user experience

## Step 1: Theme System

Add multiple visual themes that can be switched dynamically:

```go
// Define visual themes for different aesthetics
type TableTheme struct {
    Name        string
    Description string
    
    // Theme configuration
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

// Predefined themes with distinct visual styles
var themes = []TableTheme{
    {
        Name:        "Default",
        Description: "Clean Unicode box drawing theme",
        BorderChars: core.BorderChars{
            Horizontal:  "─",
            Vertical:    "│",
            TopLeft:     "┌",
            TopRight:    "┐",
            BottomLeft:  "└",
            BottomRight: "┘",
            TopT:        "┬",
            BottomT:     "┴",
            LeftT:       "├",
            RightT:      "┤",
            Cross:       "┼",
        },
        Colors: ThemeColors{
            BorderColor:    "8",   // Gray
            HeaderColor:    "99",  // Purple
            CellColor:      "252", // Light gray
            CursorColor:    "205", // Pink
            SelectionColor: "57",  // Blue
        },
    },
    {
        Name:        "Heavy",
        Description: "Double-line borders for emphasis",
        BorderChars: core.BorderChars{
            Horizontal:  "═",
            Vertical:    "║",
            TopLeft:     "╔",
            TopRight:    "╗",
            BottomLeft:  "╚",
            BottomRight: "╝",
            TopT:        "╦",
            BottomT:     "╩",
            LeftT:       "╠",
            RightT:      "╣",
            Cross:       "╬",
        },
        Colors: ThemeColors{
            BorderColor:    "14",  // Cyan
            HeaderColor:    "11",  // Yellow
            CellColor:      "15",  // White
            CursorColor:    "22",  // Dark green
            SelectionColor: "235", // Dark gray
        },
    },
    {
        Name:        "Minimal",
        Description: "Clean borderless design",
        BorderChars: core.BorderChars{
            Horizontal:  " ",
            Vertical:    " ",
            TopLeft:     " ",
            TopRight:    " ",
            BottomLeft:  " ",
            BottomRight: " ",
            TopT:        " ",
            BottomT:     " ",
            LeftT:       " ",
            RightT:      " ",
            Cross:       " ",
        },
        Colors: ThemeColors{
            BorderColor:    "8",   // Gray (hidden)
            HeaderColor:    "0",   // Black
            CellColor:      "0",   // Black
            CursorColor:    "7",   // Light gray
            SelectionColor: "235", // Dark gray
        },
    },
    {
        Name:        "Retro",
        Description: "ASCII retro computing style",
        BorderChars: core.BorderChars{
            Horizontal:  "-",
            Vertical:    "|",
            TopLeft:     "+",
            TopRight:    "+",
            BottomLeft:  "+",
            BottomRight: "+",
            TopT:        "+",
            BottomT:     "+",
            LeftT:       "+",
            RightT:      "+",
            Cross:       "+",
        },
        Colors: ThemeColors{
            BorderColor:    "14",  // Cyan
            HeaderColor:    "13",  // Magenta
            CellColor:      "15",  // White
            CursorColor:    "201", // Bright magenta
            SelectionColor: "235", // Dark gray
        },
    },
}
```

## Step 2: Theme Application

Convert theme definitions to VTable's theme system:

```go
func convertToVTableTheme(theme TableTheme) core.Theme {
    return core.Theme{
        HeaderStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.HeaderColor)).Bold(true),
        CellStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.CellColor)),
        CursorStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.CursorColor)).Bold(true),
        SelectedStyle:      lipgloss.NewStyle().Background(lipgloss.Color(theme.Colors.SelectionColor)).Foreground(lipgloss.Color("230")),
        FullRowCursorStyle: lipgloss.NewStyle().Background(lipgloss.Color(theme.Colors.CursorColor)).Foreground(lipgloss.Color("15")).Bold(true),
        BorderChars:        theme.BorderChars,
        BorderColor:        theme.Colors.BorderColor,
        HeaderColor:        theme.Colors.HeaderColor,
        AlternateRowStyle:  lipgloss.NewStyle().Background(lipgloss.Color("235")),
        DisabledStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
        LoadingStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Italic(true),
        ErrorStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
    }
}

// Apply theme to table
func (app *App) applyTheme(themeIndex int) tea.Cmd {
    if themeIndex >= 0 && themeIndex < len(themes) {
        theme := themes[themeIndex]
        vtableTheme := convertToVTableTheme(theme)
        app.statusMessage = fmt.Sprintf("Theme: %s - %s", theme.Name, theme.Description)
        return core.TableThemeSetCmd(vtableTheme)
    }
    return nil
}
```

## Step 3: Border Controls

Add granular border control with keyboard shortcuts:

```go
type App struct {
    table         *table.Table
    dataSource    *LargeEmployeeDataSource
    statusMessage string
    
    // Styling controls
    currentTheme    int
    showBorders     bool
    showTopBorder   bool
    showBottomBorder bool
    showHeaderSeparator bool
    removeTopSpace  bool
    removeBottomSpace bool
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
            
        // Theme controls
        case "t":
            app.currentTheme = (app.currentTheme + 1) % len(themes)
            return app, app.applyTheme(app.currentTheme)
            
        // Border visibility controls
        case "b":
            app.showBorders = !app.showBorders
            if app.showBorders {
                app.statusMessage = "All borders enabled"
            } else {
                app.statusMessage = "All borders disabled"
            }
            return app, core.BorderVisibilityCmd(app.showBorders)
            
        case "1":
            app.showTopBorder = !app.showTopBorder
            if app.showTopBorder {
                app.statusMessage = "Top border enabled"
            } else {
                app.statusMessage = "Top border disabled"
            }
            return app, core.TopBorderVisibilityCmd(app.showTopBorder)
            
        case "2":
            app.showBottomBorder = !app.showBottomBorder
            if app.showBottomBorder {
                app.statusMessage = "Bottom border enabled"
            } else {
                app.statusMessage = "Bottom border disabled"
            }
            return app, core.BottomBorderVisibilityCmd(app.showBottomBorder)
            
        case "3":
            app.showHeaderSeparator = !app.showHeaderSeparator
            if app.showHeaderSeparator {
                app.statusMessage = "Header separator enabled"
            } else {
                app.statusMessage = "Header separator disabled"
            }
            return app, core.HeaderSeparatorVisibilityCmd(app.showHeaderSeparator)
            
        // Space removal controls
        case "4":
            app.removeTopSpace = !app.removeTopSpace
            if app.removeTopSpace {
                app.statusMessage = "Top border space removed - compact layout"
            } else {
                app.statusMessage = "Top border space preserved"
            }
            return app, core.TopBorderSpaceRemovalCmd(app.removeTopSpace)
            
        case "5":
            app.removeBottomSpace = !app.removeBottomSpace
            if app.removeBottomSpace {
                app.statusMessage = "Bottom border space removed - compact layout"
            } else {
                app.statusMessage = "Bottom border space preserved"
            }
            return app, core.BottomBorderSpaceRemovalCmd(app.removeBottomSpace)
            
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
```

## Step 4: Enhanced Table Configuration

Set up the table with theming support and proper configuration:

```go
func createStyledTableConfig() core.TableConfig {
    return core.TableConfig{
        Columns: []core.TableColumn{
            {Title: "Employee", Field: "name", Width: 25, Alignment: core.AlignLeft},
            {Title: "Department", Field: "department", Width: 20, Alignment: core.AlignCenter},
            {Title: "Status", Field: "status", Width: 15, Alignment: core.AlignCenter},
            {Title: "Salary", Field: "salary", Width: 18, Alignment: core.AlignRight},
            {Title: "Hire Date", Field: "hire_date", Width: 15, Alignment: core.AlignCenter},
        },
        ShowHeader:              true,
        ShowBorders:             true,
        ShowTopBorder:           true,
        ShowBottomBorder:        true,
        ShowHeaderSeparator:     true,
        RemoveTopBorderSpace:    false,
        RemoveBottomBorderSpace: false,
        ViewportConfig: core.ViewportConfig{
            Height:             10,
            ChunkSize:          25,
            TopThreshold:       3,
            BottomThreshold:    3,
            BoundingAreaBefore: 50,
            BoundingAreaAfter:  50,
        },
        Theme:         convertToVTableTheme(themes[0]), // Start with default theme
        SelectionMode: core.SelectionMultiple,
        KeyMap: core.NavigationKeyMap{
            Up:        []string{"up", "k"},
            Down:      []string{"down", "j"},
            PageUp:    []string{"pgup", "h"},
            PageDown:  []string{"pgdown", "l"},
            Home:      []string{"home", "g"},
            End:       []string{"end", "G"},
            Select:    []string{"enter", " "},
            SelectAll: []string{"ctrl+a"},
            Quit:      []string{"q"},
        },
    }
}

func (app App) Init() tea.Cmd {
    return tea.Batch(
        app.table.Init(),
        app.table.Focus(),
        // Keep the same formatters from column-formatting example
        core.CellFormatterSetCmd(0, nameFormatter),
        core.CellFormatterSetCmd(1, deptFormatter),
        core.CellFormatterSetCmd(2, statusFormatter),
        core.CellFormatterSetCmd(3, salaryFormatter),
        core.CellFormatterSetCmd(4, dateFormatter),
        // Apply initial theme
        app.applyTheme(app.currentTheme),
    )
}
```

## Step 5: Enhanced View with Styling Info

Display current styling state and controls:

```go
func (app App) View() string {
    var sections []string
    
    sections = append(sections, "Table Styling Demo - Themes & Border Controls")
    sections = append(sections, "")
    sections = append(sections, app.table.View())
    sections = append(sections, "")
    
    // Controls
    sections = append(sections, "Controls: ↑↓/jk=move, Space=select, ctrl+a=select all, c=clear, q=quit")
    sections = append(sections, "Styling: T=theme, B=borders, 1=top, 2=bottom, 3=separator, 4=top-space, 5=bottom-space")
    sections = append(sections, "")
    
    // Current styling state
    currentTheme := themes[app.currentTheme]
    sections = append(sections, fmt.Sprintf("Theme: %s (%s)", currentTheme.Name, currentTheme.Description))
    
    borderStatus := "Off"
    if app.showBorders {
        borderStatus = "On"
    }
    sections = append(sections, fmt.Sprintf("Borders: %s | Top: %t | Bottom: %t | Header: %t", 
        borderStatus, app.showTopBorder, app.showBottomBorder, app.showHeaderSeparator))
    
    if app.removeTopSpace || app.removeBottomSpace {
        sections = append(sections, fmt.Sprintf("Space Removal: Top=%t | Bottom=%t (Compact mode)", 
            app.removeTopSpace, app.removeBottomSpace))
    }
    
    return strings.Join(sections, "\n")
}
```

## Theme Details

### Default Theme
- **Style**: Clean Unicode box drawing
- **Borders**: Standard `┌─┐ │ └─┘` characters
- **Colors**: Gray borders, purple headers, clean text
- **Use Case**: Professional applications, documentation

### Heavy Theme  
- **Style**: Double-line borders for emphasis
- **Borders**: Bold `╔═╗ ║ ╚═╝` characters
- **Colors**: Cyan borders, yellow headers, bright text
- **Use Case**: Important data, emphasis tables

### Minimal Theme
- **Style**: Borderless clean design
- **Borders**: All spaces (invisible borders)
- **Colors**: Black text, subtle highlighting
- **Use Case**: Content-focused displays, embedded tables

### Retro Theme
- **Style**: ASCII computing aesthetic
- **Borders**: Simple `+-+ | +-+` characters  
- **Colors**: Cyan/magenta retro colors
- **Use Case**: Terminal applications, debugging tools

## Border Control Features

### Granular Visibility
- **All Borders** (B): Toggle entire border system
- **Top Border** (1): Control top table frame
- **Bottom Border** (2): Control bottom table frame
- **Header Separator** (3): Control line between header and data

### Space Removal
- **Top Space** (4): Remove empty line above table
- **Bottom Space** (5): Remove empty line below table
- **Compact Mode**: Both space removals for minimal layout

## Color Support

VTable automatically adapts to your terminal's color capabilities:

### Terminal Types
- **16 Colors**: Basic ANSI colors (1-16)
- **256 Colors**: Extended palette (1-255)
- **True Color**: 24-bit RGB support (#RRGGBB)

### Adaptive Design
Themes use color codes that work across different terminal types, ensuring consistent appearance whether you're using a basic terminal or a modern one with full color support.

## Try It Yourself

1. Run the example: `cd docs/05-table-component/examples/table-styling && go run main.go`
2. Press **T** to cycle through themes and see different visual styles
3. Press **B** to toggle all borders on/off
4. Use **1, 2, 3** to control individual border sections
5. Try **4, 5** to see compact space removal modes
6. Navigate and select data to see how themes affect interaction

## What's Next

In the next section, we'll explore [Table Headers](07-table-headers.md) where we'll add custom header formatting, multi-line headers, and advanced header styling.

## Key Takeaways

- **Themes provide consistent visual identity** across your entire table interface
- **Granular border controls** let you customize exactly which borders appear
- **Space removal creates compact layouts** for dense information display
- **Color adaptation works across terminals** for broad compatibility
- **Visual styling enhances usability** without affecting functionality
- **Theme switching enables user preferences** and different display modes 