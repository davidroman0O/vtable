# Advanced Tree Features

## What We're Adding

Taking our styled and connected tree from previous examples, we're adding **advanced tree operations** - expand all/collapse all commands, subtree manipulation, simple keyboard shortcuts, and cascading selection. Transform your tree from a basic browser into a power-user interface.

## Understanding Advanced Tree Operations

Advanced tree features provide efficient ways to work with large hierarchical datasets:

```
Basic tree:                   Advanced tree:
📁 Project                    📁 Project ✓ (auto-expanded)
├── 📁 src (collapsed)       ├── 📁 src ✓ (expanded)
├── 📁 tests (collapsed)     │   ├── 📄 main.go ✓ (selected via cascading)
└── 📁 docs (collapsed)      │   ├── 📄 app.go ✓ (selected via cascading)
                              │   └── 📁 handlers ✓ (selected via cascading)
                              │       ├── 📄 user.go ✓
                              │       └── 📄 auth.go ✓
                              ├── 📁 tests (collapsed)
                              └── 📁 docs (collapsed)
```

## New TreeList Methods Added

During implementation, we discovered the TreeList needed additional methods. These are now available:

### Current Node Access
```go
func (tl *TreeList[T]) GetCurrentNodeID() string
```
Returns the ID of the node currently under the cursor.

### Subtree Operations
```go
func (tl *TreeList[T]) ExpandSubtree(id string) tea.Cmd
func (tl *TreeList[T]) CollapseSubtree(id string) tea.Cmd
func (tl *TreeList[T]) ExpandCurrentSubtree() tea.Cmd  
func (tl *TreeList[T]) CollapseCurrentSubtree() tea.Cmd
```
Expand or collapse a node and ALL its descendants recursively.

### Bulk Operations
```go
func (tl *TreeList[T]) ExpandAll() tea.Cmd
func (tl *TreeList[T]) CollapseAll() tea.Cmd
```
Expand or collapse ALL nodes in the entire tree.

### Cascading Selection
```go
func (tl *TreeList[T]) SetCascadingSelection(enabled bool)
func (tl *TreeList[T]) GetCascadingSelection() bool
```
Enable/disable automatic selection of child nodes when parent is selected.

## Advanced Operation Types

### 1. Expand All / Collapse All
Bulk operations for the entire tree:

```
E: Expand Everything           C: Collapse Everything
📁 Project                     📁 Project
├── 📁 src                     ├── 📁 src (collapsed)
│   ├── 📄 main.go             ├── 📁 tests (collapsed)
│   └── 📄 app.go              └── 📁 docs (collapsed)
├── 📁 tests
│   └── 📄 test.go
└── 📁 docs
    └── 📄 README.md
```

### 2. Subtree Operations
Operations on current node's subtree:

```
e: Expand Current Subtree      c: Collapse Current Subtree
📁 src ← cursor here           📁 src ← cursor here
├── 📄 main.go                 (children hidden)
├── 📄 app.go
└── 📁 handlers
    ├── 📄 user.go
    └── 📄 auth.go
```

### 3. Cascading Selection
When enabled, selecting a parent automatically selects all children:

```
Select "src" folder:
📁 src ✓ (selected)
├── 📄 main.go ✓ (auto-selected)
├── 📄 app.go ✓ (auto-selected)
└── 📁 handlers ✓ (auto-selected)
    ├── 📄 user.go ✓ (auto-selected)
    └── 📄 auth.go ✓ (auto-selected)
```

## Step 1: Configure Advanced Features

Enable the features in your tree configuration:

```go
// Configure tree with advanced features
treeConfig := tree.DefaultTreeConfig()

// Enable cascading selection (parent selects all children)
treeConfig.CascadingSelection = true

// Enable connected lines for better visual hierarchy
    treeConfig.RenderConfig.IndentationConfig.Enabled = true
    treeConfig.RenderConfig.IndentationConfig.UseConnectors = true
    treeConfig.RenderConfig.IndentationConfig.ConnectorStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("240"))
    
// Enhanced content formatting with expansion indicators
    treeConfig.RenderConfig.ContentConfig.Formatter = createAdvancedFormatter()
    
    // Background styling for cursor items
    treeConfig.RenderConfig.BackgroundConfig.Enabled = true
    treeConfig.RenderConfig.BackgroundConfig.Style = lipgloss.NewStyle().
        Background(lipgloss.Color("240")).
        Foreground(lipgloss.Color("15"))
```

## Step 2: Enhanced Content Formatter

Create a formatter that shows expansion state:

```go
func createAdvancedFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
    return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
        if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
            content := flatItem.Item.String()
            
            // Add visual indicator for folders with children
            if flatItem.Item.IsFolder && hasChildren {
                if isExpanded {
                    content = content + " (expanded)"
                } else {
                    content = content + " (...)"
                }
            }
            
            // Apply selection styling (highest priority)
            if item.Selected {
                return lipgloss.NewStyle().
                    Background(lipgloss.Color("12")).
                    Foreground(lipgloss.Color("15")).
                    Bold(true).
                    Render(content)
            }
            
            // Content styling
            if flatItem.Item.IsFolder {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("12")).
                    Bold(true).
                    Render(content)
            } else {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("10")).
                    Render(content)
            }
        }
        
        return fmt.Sprintf("%v", item.Item)
    }
}
```

## Step 3: Simple Keyboard Shortcuts

Use simple letter keys instead of modifier combinations:

```go
func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return app, tea.Quit

        // Basic tree operations
        case "enter":
            app.status = "Toggled expand/collapse"
            return app, app.tree.ToggleCurrentNode()
        case " ":
            app.status = "Toggled selection" 
            return app, core.SelectCurrentCmd()

        // Advanced operations - SIMPLE KEYS!
        case "E":
            app.status = "Expanded entire tree"
            return app, app.tree.ExpandAll()
        case "C":
            app.status = "Collapsed entire tree"
            return app, app.tree.CollapseAll()
        case "e":
            app.status = "Expanded current subtree"
            return app, app.tree.ExpandCurrentSubtree()
        case "c":
            app.status = "Collapsed current subtree"
            return app, app.tree.CollapseCurrentSubtree()

        // Selection operations
        case "a":
            app.status = "Selected all items"
            return app, core.SelectAllCmd()
        case "x":
            app.status = "Cleared all selections"
            return app, core.SelectClearCmd()

        // Navigation shortcuts
        case "h", "left":
            app.status = "Navigate up"
            return app, core.CursorUpCmd()
        case "l", "right":
            app.status = "Expand/toggle current node"
            return app, app.tree.ToggleCurrentNode()
        case "j", "down":
            app.status = "Navigate down"
            return app, core.CursorDownCmd()
        case "k", "up":
            app.status = "Navigate up"
            return app, core.CursorUpCmd()
        }
    }

    // Pass messages to tree
    var cmd tea.Cmd
    _, cmd = app.tree.Update(msg)
    return app, cmd
}
```

## Step 4: Auto-Expand on Startup

Set up initial expansion when the app starts:

```go
func main() {
    // Create data source and tree...
    treeComponent := tree.NewTreeList(listConfig, treeConfig, dataSource)

    // Auto-expand some nodes on startup using TreeList methods directly
    var autoExpandCommands []tea.Cmd
    if len(dataSource.rootNodes) > 0 {
        // Expand the first root node
        autoExpandCommands = append(autoExpandCommands, 
            treeComponent.ExpandNode(dataSource.rootNodes[0].ID))
    }

    app := &App{
        tree:   treeComponent,
        status: "Advanced tree ready! Try E/C to expand/collapse all",
    }

    p := tea.NewProgram(app, tea.WithoutSignalHandler())

    // Apply auto-expand after starting
    go func() {
        for _, cmd := range autoExpandCommands {
            if cmd != nil {
                p.Send(cmd())
            }
        }
    }()

    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## What You'll See

### Auto-Expanded Tree on Startup
```
🌳 Advanced Tree Features Demo
E/C: expand/collapse all | e/c: subtree | a/x: select all/clear

📁 Web Application (expanded)
├── 📁 src (expanded)
│   ├── 📄 main.go
│   ├── 📄 app.go
│   ├── 📁 handlers (...)
│   └── 📁 models (...)
├── 📁 tests (...)
└── 📁 config (...)
📁 CLI Tool (...)
📁 Documentation (...)

Status: Advanced tree ready! Try E/C to expand/collapse all
↑/↓/j/k: navigate | Enter: toggle | Space: select | q: quit
```

### After Pressing 'E' (Expand All)
```
📁 Web Application (expanded)
├── 📁 src (expanded)
│   ├── 📄 main.go
│   ├── 📄 app.go
│   ├── 📁 handlers (expanded)
│   │   ├── 📄 user_handler.go
│   │   ├── 📄 auth_handler.go
│   │   └── 📄 middleware.go
│   └── 📁 models (expanded)
│       ├── 📄 user.go
│       └── 📄 product.go
├── 📁 tests (expanded)
│   ├── 📄 unit_test.go
│   └── 📄 integration_test.go
└── 📁 config (expanded)
    ├── 📄 .env
    └── 📄 config.yaml
[... all nodes expanded ...]
```

### Cascading Selection in Action
```
📁 src ✅ (selected)
├── 📄 main.go ✅ (auto-selected)
├── 📄 app.go ✅ (auto-selected)
├── 📁 handlers ✅ (auto-selected)
│   ├── 📄 user_handler.go ✅ (auto-selected)
│   ├── 📄 auth_handler.go ✅ (auto-selected)
│   └── 📄 middleware.go ✅ (auto-selected)
└── 📁 models ✅ (auto-selected)
    ├── 📄 user.go ✅ (auto-selected)
    └── 📄 product.go ✅ (auto-selected)
```

## Key Implementation Insights

### 1. **Library Enhancement Required**
The TreeList needed additional methods for advanced features:
- `GetCurrentNodeID()` - get node under cursor
- `ExpandSubtree(id)` / `CollapseSubtree(id)` - recursive operations
- `ExpandAll()` / `CollapseAll()` - bulk operations
- `ExpandCurrentSubtree()` / `CollapseCurrentSubtree()` - convenience methods

### 2. **Simple Keyboard Design**
No modifier keys required:
- **E/C**: Global expand/collapse all
- **e/c**: Current subtree operations  
- **a/x**: Selection operations
- **h/j/k/l**: Vi-style navigation

### 3. **Direct API Usage**
Call TreeList methods directly instead of complex data source intermediaries:
```go
// ✅ Clean and direct
return app, app.tree.ExpandAll()

// ❌ Complex and indirect  
return app, app.dataSource.ExpandAllNodes()
```

### 4. **Component-Based Rendering**
Use the enhanced tree formatter:
- Shows expansion state indicators
- Proper selection highlighting with cascading
- Clean visual hierarchy with connectors


## Try It Yourself

1. **Test bulk operations** - press 'E' to expand all, 'C' to collapse all
2. **Try subtree operations** - navigate to a folder, press 'e' to expand just that subtree
3. **Test cascading selection** - select a parent folder and see children auto-select  
4. **Navigate efficiently** - use h/j/k/l for quick tree navigation
5. **Clear and select** - use 'a' to select all, 'x' to clear all selections

## What's Next

You now understand how to implement comprehensive advanced tree features! This completes the tree component documentation series, taking you from basic tree structure to sophisticated tree manipulation.

The insight: **Advanced tree features require both library enhancement and thoughtful UX design** - the TreeList needed new methods to support these operations, and simple keyboard shortcuts work better than complex modifier combinations. 