# Connected Lines

## What We're Adding

Taking our indented tree from previous examples, we're adding **connected lines** - box-drawing characters that create visual connections between tree nodes. Transform your tree from simple indentation to a classic file manager appearance with connecting lines that clearly show parent-child relationships.

## Understanding Connected Lines

Connected lines use Unicode box-drawing characters to visually connect tree nodes:

```
Basic indentation:       Connected lines:            Advanced connectors:
📁 Project               📁 Project                  📁 Project
  📁 src                 ├── 📁 src                  ├── 📁 src
    📄 main.go           │   ├── 📄 main.go          │   ├── 📄 main.go
    📄 app.go            │   └── 📄 app.go           │   └── 📄 app.go
  📁 tests               └── 📁 tests                └── 📁 tests
    📄 test.go               └── 📄 test.go              └── 📄 test.go
```

You can customize:
- **Connector style** (different box-drawing character sets)
- **Line thickness** (light, heavy, double lines)
- **Connection points** (T-junctions, corners, straight lines)
- **Visual styling** (colors, bold/normal weight)

## Box-Drawing Character Sets

### 1. Light Box Drawing (Default)
The most common and widely supported set:

```
📁 Project
├── 📁 src
│   ├── 📄 main.go
│   └── 📄 app.go
└── 📁 tests
    └── 📄 test.go
```

**Characters used**: `├` `│` `└` `─`

### 2. Heavy Box Drawing
Thicker lines for better visibility:

```
📁 Project
┣━━ 📁 src
┃   ┣━━ 📄 main.go
┃   ┗━━ 📄 app.go
┗━━ 📁 tests
    ┗━━ 📄 test.go
```

**Characters used**: `┣` `┃` `┗` `━`

### 3. Double Line Box Drawing
Professional appearance with double lines:

```
📁 Project
╠══ 📁 src
║   ╠══ 📄 main.go
║   ╚══ 📄 app.go
╚══ 📁 tests
    ╚══ 📄 test.go
```

**Characters used**: `╠` `║` `╚` `═`

## Step 1: Start with the Working Tree

We'll use the same multi-project structure and build on our indentation examples:

```go
// Same FileItem and data structure as before
type FileItem struct {
    Name     string
    IsFolder bool
}

func (f FileItem) String() string {
    if f.IsFolder {
        return "📁 " + f.Name
    }
    return "📄 " + f.Name
}
```

**Building on previous work**: Connected lines enhance the visual hierarchy established by indentation.

## Step 2: Enable Box-Drawing Connectors

VTable's TreeIndentationComponent has built-in support for box-drawing connectors:

```go
// Enable connected lines in the tree configuration
treeConfig := tree.DefaultTreeConfig()

// Enable connectors in the indentation config
treeConfig.RenderConfig.IndentationConfig.Enabled = true
treeConfig.RenderConfig.IndentationConfig.UseConnectors = true

// Configure connector styling
treeConfig.RenderConfig.IndentationConfig.ConnectorStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("240"))  // Gray connectors
```

**Result**: Your tree will now display with box-drawing connectors instead of simple indentation.

## Step 3: Customize Connector Appearance

Configure the visual appearance of the connecting lines:

```go
// Different connector styles
treeConfig.RenderConfig.IndentationConfig.ConnectorStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("244")).  // Light gray
    Bold(true)                          // Make lines more prominent

// You can also combine with indentation styling
treeConfig.RenderConfig.IndentationConfig.Style = lipgloss.NewStyle().
    Foreground(lipgloss.Color("242"))   // Slightly different gray for variety
```

**Visual coordination**: Connector colors should complement your tree symbols and content styling.

## Step 4: Multiple Connector Themes

Let's create different connector styles for different use cases:

```go
type ConnectorTheme struct {
    Name            string
    ConnectorStyle  lipgloss.Style
    Description     string
    UseConnectors   bool
}

var connectorThemes = []ConnectorTheme{
    {
        Name:           "None",
        ConnectorStyle: lipgloss.NewStyle(),
        Description:    "No connectors - simple indentation",
        UseConnectors:  false,
    },
    {
        Name:           "Light Gray",
        ConnectorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
        Description:    "Subtle gray connectors",
        UseConnectors:  true,
    },
    {
        Name:           "Bold Gray",
        ConnectorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Bold(true),
        Description:    "Bold gray connectors for clarity",
        UseConnectors:  true,
    },
    {
        Name:           "Blue Lines",
        ConnectorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("12")),
        Description:    "Blue connectors matching folder colors",
        UseConnectors:  true,
    },
    {
        Name:           "Dim Lines",
        ConnectorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
        Description:    "Very subtle dim connectors",
        UseConnectors:  true,
    },
    {
        Name:           "High Contrast",
        ConnectorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true),
        Description:    "High contrast white connectors",
        UseConnectors:  true,
    },
}
```

## Step 5: Dynamic Connector Switching

Build an app that can toggle connectors on/off and switch styles:

```go
type App struct {
    tree             *tree.TreeList[FileItem]
    status           string
    currentConnector int
    dataSource       *FileTreeDataSource
}

func (app *App) applyConnectorTheme() {
    theme := connectorThemes[app.currentConnector]
    
    // Get current config
    treeConfig := app.tree.GetRenderConfig()
    
    // Apply connector theme
    treeConfig.IndentationConfig.UseConnectors = theme.UseConnectors
    treeConfig.IndentationConfig.ConnectorStyle = theme.ConnectorStyle
    
    // Apply the updated config
    app.tree.SetRenderConfig(treeConfig)
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "l":
            // Cycle through connector themes ('l' for lines)
            app.currentConnector = (app.currentConnector + 1) % len(connectorThemes)
            app.applyConnectorTheme()
            theme := connectorThemes[app.currentConnector]
            app.status = fmt.Sprintf("Connectors: %s - %s", theme.Name, theme.Description)
            return app, nil
        }
    }
    
    // ... rest of update logic
}
```

## Step 6: Advanced Connector Features

### Responsive Connector Visibility

Automatically hide connectors when space is limited:

```go
func createResponsiveConnectorConfig(maxWidth int) tree.TreeRenderConfig {
    config := tree.DefaultTreeConfig()
    
    // Enable connectors only if we have enough width
    if maxWidth > 60 {
        config.RenderConfig.IndentationConfig.UseConnectors = true
        config.RenderConfig.IndentationConfig.ConnectorStyle = lipgloss.NewStyle().
            Foreground(lipgloss.Color("240"))
    } else {
        // Fall back to simple indentation for narrow displays
        config.RenderConfig.IndentationConfig.UseConnectors = false
        config.RenderConfig.IndentationConfig.IndentSize = 2
    }
    
    return config
}
```

### Themed Connector Coordination

Coordinate connector colors with your overall theme:

```go
func createThemedConnectorStyle(baseTheme string) lipgloss.Style {
    switch baseTheme {
    case "dark":
        return lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Bold(true)
    case "light":
        return lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
    case "high-contrast":
        return lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
    case "professional":
        return lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
    default:
        return lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
    }
}
```

## Step 7: Content Formatting with Connectors

Enhance your content formatter to work well with connected lines:

```go
func createConnectorAwareFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
    return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
        if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
            content := flatItem.Item.String()
            
            // Apply selection styling (highest priority)
            if item.Selected {
                return lipgloss.NewStyle().
                    Background(lipgloss.Color("12")).
                    Foreground(lipgloss.Color("15")).
                    Bold(true).
                    Render(content)
            }
            
            // Content styling that works well with connectors
            if flatItem.Item.IsFolder {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("12")).  // Blue folders
                    Bold(true).
                    Render(content)
            } else {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("10")).  // Green files
                    Render(content)
            }
        }
        
        return fmt.Sprintf("%v", item.Item)
    }
}
```

## Step 8: Complete Connected Lines Example

```go
func main() {
    // Create the data source
    dataSource := NewFileTreeDataSource()
    
    // Configure the list component
    listConfig := core.ListConfig{
        ViewportConfig: core.ViewportConfig{
            Height:    14,
            ChunkSize: 20,
        },
        SelectionMode: core.SelectionMultiple,
        KeyMap:        core.DefaultNavigationKeyMap(),
    }
    
    // Start with default tree configuration
    treeConfig := tree.DefaultTreeConfig()
    
    // Enable connected lines
    treeConfig.RenderConfig.IndentationConfig.Enabled = true
    treeConfig.RenderConfig.IndentationConfig.UseConnectors = true
    treeConfig.RenderConfig.IndentationConfig.ConnectorStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("240"))
    
    // Configure content formatting
    treeConfig.RenderConfig.ContentConfig.Formatter = createConnectorAwareFormatter()
    
    // Enable background styling for cursor items
    treeConfig.RenderConfig.BackgroundConfig.Enabled = true
    treeConfig.RenderConfig.BackgroundConfig.Style = lipgloss.NewStyle().
        Background(lipgloss.Color("240")).
        Foreground(lipgloss.Color("15"))
    
    // Create the tree
    treeComponent := tree.NewTreeList(listConfig, treeConfig, dataSource)
    
    // Create the app
    app := &App{
        tree:             treeComponent,
        status:           "Ready! Press 'l' to cycle through connector styles",
        currentConnector: 1, // Start with light gray connectors
        dataSource:       dataSource,
    }
    
    // Apply initial connector theme
    app.applyConnectorTheme()
    
    // Run
    p := tea.NewProgram(app)
    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## What You'll See

### No Connectors (Simple Indentation)
```
🌳 Connected Lines Demo

📁 Web Application
  📁 src
    📄 main.go
    📄 app.go
  📁 tests
    📄 unit_test.go
```

### Light Gray Connectors
```
🌳 Connected Lines Demo

📁 Web Application
├── 📁 src
│   ├── 📄 main.go
│   ├── 📄 app.go
│   └── 📁 handlers
│       ├── 📄 user_handler.go
│       └── 📄 auth_handler.go
├── 📁 tests
│   └── 📄 unit_test.go
└── 📁 CLI Tool
    └── 📁 cmd
        ├── 📄 root.go
        └── 📄 version.go
```

### Bold Gray Connectors
```
🌳 Connected Lines Demo

📁 Web Application
├── 📁 src
│   ├── 📄 main.go
│   ├── 📄 app.go
│   └── 📁 handlers
│       ├── 📄 user_handler.go
│       └── 📄 auth_handler.go
├── 📁 tests
│   └── 📄 unit_test.go
└── 📁 CLI Tool
    └── 📁 cmd
        ├── 📄 root.go
        └── 📄 version.go
```

### Blue Connectors (Themed)
```
🌳 Connected Lines Demo

📁 Web Application
├── 📁 src            (blue connectors)
│   ├── 📄 main.go
│   ├── 📄 app.go
│   └── 📁 handlers
│       ├── 📄 user_handler.go
│       └── 📄 auth_handler.go
├── 📁 tests
│   └── 📄 unit_test.go
└── 📁 CLI Tool
    └── 📁 cmd
        ├── 📄 root.go
        └── 📄 version.go
```

## Connected Lines Best Practices

### Visual Clarity Guidelines

**Do:**
- **Use subtle colors** - connectors should guide, not dominate
- **Maintain consistency** - same connector style throughout the tree
- **Test in different terminals** - box-drawing support varies
- **Consider your content** - connectors should complement, not compete

**Don't:**
- **Use overly bright connectors** - they distract from content
- **Mix connector styles** - inconsistency confuses users
- **Ignore narrow displays** - connectors need horizontal space
- **Forget accessibility** - ensure connectors are visible but not essential

### Choosing Connector Style

**Light Gray (Subtle)**:
- **Use when**: Clean, minimal interfaces
- **Good for**: Focus on content, professional applications
- **Works with**: Any color scheme

**Bold/Dark (Prominent)**:
- **Use when**: Complex hierarchies need clear structure
- **Good for**: File managers, code editors
- **Works with**: Light backgrounds, high contrast needs

**Themed (Coordinated)**:
- **Use when**: Connectors should match your design system
- **Good for**: Branded applications, consistent styling
- **Works with**: Specific color palettes

### Terminal Compatibility

**High Compatibility**: Light box-drawing characters (├ │ └ ─)
**Medium Compatibility**: Heavy box-drawing characters (┣ ┃ ┗ ━)
**Lower Compatibility**: Double-line characters (╠ ║ ╚ ═)

**Best Practice**: Start with light box-drawing and provide fallback options.

### Performance Considerations

**Connector Impact**:
- **Minimal overhead** - box-drawing characters are single Unicode points
- **Styling cost** - color/bold styling has normal lipgloss overhead
- **Terminal rendering** - some terminals render box-drawing slower

**Optimization Tips**:
- **Cache styled connector strings** for repeated patterns
- **Provide connector-free fallback** for performance-critical scenarios
- **Test with large trees** to ensure acceptable rendering speed

## Key Concepts

### 1. **Visual Hierarchy Enhancement**
Connected lines make parent-child relationships immediately obvious.

### 2. **Progressive Enhancement**
Connectors enhance indentation - they don't replace it.

### 3. **Terminal Compatibility**
Box-drawing characters are widely supported but not universal.

### 4. **Subtle Visual Guidance**
Best connectors guide attention without stealing it.

### 5. **Responsive Design**
Consider disabling connectors on very narrow displays.

## Try It Yourself

1. **Experiment with connector colors** - try different grays, blues, or themed colors
2. **Test different connector weights** - compare normal vs bold styling
3. **Try mixed approaches** - connectors for some levels, simple indentation for others
4. **Test terminal compatibility** - verify connectors display correctly in your target terminals
5. **Consider your users** - do connectors help or hinder navigation in your specific use case?

## What's Next

You now understand how to create connected tree lines! Next, we'll explore tree enumerators - adding bullet points, numbers, and custom prefixes to tree items for additional visual organization.

The insight: **Connected lines transform trees from simple lists to visual hierarchies** - they make the structure immediately clear and help users understand complex relationships at a glance. Use them when structure clarity is more important than minimal visual noise. 