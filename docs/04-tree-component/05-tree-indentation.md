# Tree Indentation

## What We're Adding

Taking our styled tree from previous examples, we're adding **custom indentation control** - the spacing and alignment that visually represents the tree hierarchy. Transform how depth is displayed from basic spaces to sophisticated visual systems that make complex trees easy to navigate.

## Understanding Tree Indentation

Tree indentation is the horizontal spacing that shows hierarchy depth:

```
Basic indentation:       Custom indentation:        Advanced indentation:
📁 Root                  📁 Root                    📁 Root
  📁 Level 1               ···📁 Level 1              ╰── 📁 Level 1
    📄 Level 2               ·····📄 Level 2              ╰──── 📄 Level 2
      📄 Level 3               ·······📄 Level 3              ╰────── 📄 Level 3
```

You can customize:
- **Indentation size** (how much space per level)
- **Indentation style** (spaces, custom strings, visual markers)
- **Visual alignment** (consistent spacing, proportional depth)
- **Depth limits** (maximum nesting levels to display)

## Indentation Approaches

### 1. Space-Based Indentation
The simplest approach using repeated spaces:

```
📁 Project
  📁 src          (2 spaces)
    📄 main.go    (4 spaces)
    📄 app.go     (4 spaces)
  📁 tests        (2 spaces)
    📄 test.go    (4 spaces)
```

### 2. String-Based Indentation  
Custom strings repeated for each level:

```
📁 Project
··📁 src          (·· per level)
····📄 main.go    (···· for level 2)
····📄 app.go     
··📁 tests        
····📄 test.go    
```

### 3. Visual Marker Indentation
Special characters that show the tree structure:

```
📁 Project
├─📁 src          (tree-like connectors)
│ ├─📄 main.go    
│ └─📄 app.go     
└─📁 tests        
  └─📄 test.go    
```

## Step 1: Start with the Working Tree

We'll use the same multi-project structure from previous examples:

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

**Building on previous work**: Indentation enhances the visual hierarchy without changing your data or navigation logic.

## Step 2: Basic Indentation Control

Let's start by configuring the basic indentation settings:

```go
// Configure indentation using the TreeIndentationConfig
treeConfig := tree.DefaultTreeConfig()

// Basic space-based indentation
treeConfig.RenderConfig.IndentationConfig.Enabled = true
treeConfig.RenderConfig.IndentationConfig.IndentSize = 4  // 4 spaces per level
treeConfig.RenderConfig.IndentationConfig.IndentString = ""  // Use spaces
```

**Result**: Each tree level will be indented by 4 spaces instead of the default 2.

## Step 3: Custom String Indentation

Replace spaces with custom strings for each level:

```go
// Custom string indentation - dots for visual clarity
treeConfig.RenderConfig.IndentationConfig.IndentString = "··"  // Two dots per level
treeConfig.RenderConfig.IndentationConfig.IndentSize = 0       // Ignored when IndentString is set

// Apply styling to the indentation
treeConfig.RenderConfig.IndentationConfig.Style = lipgloss.NewStyle().
    Foreground(lipgloss.Color("240"))  // Gray dots
```

**Visual effect**: Each level gets `··` (gray dots) showing depth clearly.

## Step 4: Advanced Indentation Styling

Create visually distinctive indentation with styling:

```go
func createStyledIndentationTheme() {
    treeConfig := tree.DefaultTreeConfig()
    
    // Custom indentation with visual styling
    treeConfig.RenderConfig.IndentationConfig.Enabled = true
    treeConfig.RenderConfig.IndentationConfig.IndentString = "│ "  // Vertical bar + space
    treeConfig.RenderConfig.IndentationConfig.Style = lipgloss.NewStyle().
        Foreground(lipgloss.Color("244")).  // Light gray
        Bold(true)
    
    return treeConfig
}
```

**Advanced styling**: Creates a visual "guide line" showing the tree structure with styled vertical bars.

## Step 5: Multiple Indentation Themes

Let's create different indentation styles for different purposes:

```go
type IndentationTheme struct {
    Name         string
    IndentString string
    IndentSize   int
    Style        lipgloss.Style
    Description  string
}

var indentationThemes = []IndentationTheme{
    {
        Name:         "Minimal",
        IndentString: "",
        IndentSize:   2,
        Style:        lipgloss.NewStyle(),
        Description:  "Clean 2-space indentation",
    },
    {
        Name:         "Spacious", 
        IndentString: "",
        IndentSize:   4,
        Style:        lipgloss.NewStyle(),
        Description:  "Wide 4-space indentation for clarity",
    },
    {
        Name:         "Dotted",
        IndentString: "··",
        IndentSize:   0,
        Style:        lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
        Description:  "Gray dots show hierarchy clearly",
    },
    {
        Name:         "Dashed",
        IndentString: "- ",
        IndentSize:   0,
        Style:        lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
        Description:  "Dashes for distinctive hierarchy",
    },
    {
        Name:         "Boxed",
        IndentString: "│ ",
        IndentSize:   0,
        Style:        lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Bold(true),
        Description:  "Box-drawing characters for structure",
    },
}
```

## Step 6: Dynamic Indentation Switching

Build an app that can switch between different indentation styles:

```go
type App struct {
    tree                *tree.TreeList[FileItem]
    status              string
    currentIndentation  int
    dataSource          *FileTreeDataSource
}

func (app *App) applyIndentationTheme() {
    theme := indentationThemes[app.currentIndentation]
    
    // Get current config
    treeConfig := app.tree.GetRenderConfig()
    
    // Apply indentation theme
    treeConfig.IndentationConfig.IndentString = theme.IndentString
    treeConfig.IndentationConfig.IndentSize = theme.IndentSize
    treeConfig.IndentationConfig.Style = theme.Style
    
    // Apply the updated config
    app.tree.SetRenderConfig(treeConfig)
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "i":
            // Cycle through indentation themes
            app.currentIndentation = (app.currentIndentation + 1) % len(indentationThemes)
            app.applyIndentationTheme()
            theme := indentationThemes[app.currentIndentation]
            app.status = fmt.Sprintf("Indentation: %s - %s", theme.Name, theme.Description)
            return app, nil
        }
    }
    
    // ... rest of update logic
}
```

## Step 7: Advanced Indentation Features

### Depth-Aware Indentation

Create indentation that adapts based on tree depth:

```go
func createDepthAwareIndentationFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
    return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
        if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
            content := flatItem.Item.String()
            
            // Add depth indicators for very deep items
            var depthIndicator string
            if depth > 3 {
                depthIndicator = fmt.Sprintf("[%d] ", depth)  // Show depth number for deep items
            }
            
            // Apply selection and cursor styling
            if item.Selected {
                return lipgloss.NewStyle().
                    Background(lipgloss.Color("12")).
                    Foreground(lipgloss.Color("15")).
                    Bold(true).
                    Render(depthIndicator + content)
            }
            
            // Regular content with depth indicator
            if flatItem.Item.IsFolder {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("12")).
                    Bold(true).
                    Render(depthIndicator + content)
            } else {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("10")).
                    Render(depthIndicator + content)
            }
        }
        
        return fmt.Sprintf("%v", item.Item)
    }
}
```

### Responsive Indentation

Adjust indentation based on available width:

```go
func createResponsiveIndentationConfig(maxWidth int) tree.TreeRenderConfig {
    config := tree.DefaultTreeConfig()
    
    // Adjust indentation size based on available width
    if maxWidth < 40 {
        // Narrow width - minimal indentation
        config.RenderConfig.IndentationConfig.IndentSize = 1
    } else if maxWidth < 80 {
        // Medium width - standard indentation  
        config.RenderConfig.IndentationConfig.IndentSize = 2
    } else {
        // Wide width - spacious indentation
        config.RenderConfig.IndentationConfig.IndentSize = 3
    }
    
    return config
}
```

## Step 8: Complete Indentation Example

```go
func main() {
    // Create the data source
    dataSource := NewFileTreeDataSource()
    
    // Configure the list component
    listConfig := core.ListConfig{
        ViewportConfig: core.ViewportConfig{
            Height:    12,
            ChunkSize: 20,
        },
        SelectionMode: core.SelectionMultiple,
        KeyMap:        core.DefaultNavigationKeyMap(),
    }
    
    // Start with default tree configuration
    treeConfig := tree.DefaultTreeConfig()
    
    // Enable and configure indentation
    treeConfig.RenderConfig.IndentationConfig.Enabled = true
    treeConfig.RenderConfig.IndentationConfig.IndentSize = 2
    treeConfig.RenderConfig.IndentationConfig.Style = lipgloss.NewStyle()
    
    // Create the tree
    treeComponent := tree.NewTreeList(listConfig, treeConfig, dataSource)
    
    // Create the app
    app := &App{
        tree:               treeComponent,
        status:             "Ready! Press 'i' to cycle through indentation styles",
        currentIndentation: 0,
        dataSource:         dataSource,
    }
    
    // Apply initial indentation theme
    app.applyIndentationTheme()
    
    // Run
    p := tea.NewProgram(app)
    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## What You'll See

### Minimal Indentation (2 spaces)
```
🌳 Tree Indentation Demo

📁 Web Application
  📁 src
    📄 main.go
    📄 app.go
  📁 tests
    📄 unit_test.go
```

### Spacious Indentation (4 spaces) 
```
🌳 Tree Indentation Demo

📁 Web Application
    📁 src
        📄 main.go
        📄 app.go
    📁 tests
        📄 unit_test.go
```

### Dotted Indentation
```
🌳 Tree Indentation Demo

📁 Web Application
··📁 src
····📄 main.go
····📄 app.go
··📁 tests
····📄 unit_test.go
```

### Dashed Indentation
```
🌳 Tree Indentation Demo

📁 Web Application
- 📁 src
- - 📄 main.go
- - 📄 app.go
- 📁 tests
- - 📄 unit_test.go
```

### Boxed Indentation
```
🌳 Tree Indentation Demo

📁 Web Application
│ 📁 src
│ │ 📄 main.go
│ │ 📄 app.go
│ 📁 tests
│ │ 📄 unit_test.go
```

## Indentation Best Practices

### Visual Hierarchy Guidelines

**Do:**
- **Maintain consistency** - same indentation approach throughout the tree
- **Consider depth limits** - very deep trees can become hard to read
- **Test with real data** - ensure indentation works with your actual tree structure
- **Balance clarity and space** - too much indentation wastes horizontal space

**Don't:**
- **Mix indentation styles** - use one approach per tree
- **Make indentation too subtle** - users need to see the hierarchy clearly
- **Ignore narrow screens** - test indentation on smaller widths
- **Use complex indentation for simple trees** - match complexity to need

### Choosing Indentation Style

**Space-Based (Default)**:
- **Use when**: Simple, clean interfaces
- **Good for**: Most applications, familiar to users
- **Configurable**: IndentSize controls spacing

**String-Based (Custom)**:
- **Use when**: Need visual distinctiveness
- **Good for**: Complex hierarchies, debugging tree structure
- **Flexible**: Any string can be repeated per level

**Visual Markers**: 
- **Use when**: Maximum hierarchy clarity needed
- **Good for**: File managers, code editors
- **Advanced**: Can combine with box-drawing characters

### Performance Considerations

**Indentation Impact**:
- **Space-based**: Minimal performance impact
- **String-based**: Slightly more expensive (string operations)
- **Complex styling**: More expensive (style rendering)

**Optimization Tips**:
- **Cache styled strings** for repeated indentation patterns
- **Use simple strings** for very large trees
- **Consider responsive indentation** that adapts to tree size

## Key Concepts

### 1. **Separation of Concerns**
Indentation is handled by the TreeIndentationComponent, separate from content formatting.

### 2. **Consistent Visual Language**
Indentation should work harmoniously with your tree symbols and content styling.

### 3. **Depth Representation**
Different indentation approaches communicate hierarchy in different ways - choose what works for your data.

### 4. **Responsive Design**
Consider how indentation looks across different screen sizes and terminal widths.

### 5. **User Experience**
Good indentation makes trees easier to navigate and understand at a glance.

## Try It Yourself

1. **Experiment with indentation sizes** - try 1, 2, 3, 4, and 6 spaces per level
2. **Create custom indentation strings** - try "→ ", "● ", "▸ " or other symbols
3. **Test with deep trees** - see how indentation looks with 5+ levels of nesting
4. **Combine with styling** - add colors and fonts to your indentation
5. **Consider your users** - what indentation style works best for your application?

## What's Next

You now understand how to control tree indentation! Next, we'll explore connected lines - using box-drawing characters to create visual connections between tree nodes for the ultimate in hierarchy clarity.

The insight: **Good indentation makes tree hierarchy intuitive** - users should be able to understand the structure at a glance, and navigation should feel natural and predictable. 