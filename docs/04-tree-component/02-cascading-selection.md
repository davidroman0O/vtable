# Cascading Selection

## What We're Adding

Taking our multi-project file tree from the previous example, we're adding **cascading selection** - when you select a folder, it automatically selects all the files and subfolders inside it. This is the behavior you'd expect in a file manager.

## Understanding Cascading Selection

**Basic selection**: Select one item at a time
**Cascading selection**: Select a parent → automatically selects all its children

```
Before:                After selecting "src":
► 📁 src              ► 📁 src (blue highlight)
    📄 main.go            📄 main.go (blue highlight)
    📄 app.go             📄 app.go (blue highlight)
```

This is perfect for operations like "select all files in this folder" or "delete this entire project branch."

## Step 1: Reuse the Multi-Project Structure

We'll use the same data structure from the basic tree example. If you followed that tutorial, you already have:

```go
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

And the same `FileTreeDataSource` structure:

```go
type FileTreeDataSource struct {
    rootNodes     []tree.TreeData[FileItem]
    selectedNodes map[string]bool
}
```

**Why reuse?** Cascading selection doesn't require any changes to your data structure - it's all handled by configuration.

## Step 2: Build Tree Data (Same as Before)

If you're starting fresh, here's a simplified version of the tree data:

```go
func NewFileTreeDataSource() *FileTreeDataSource {
    return &FileTreeDataSource{
        rootNodes: []tree.TreeData[FileItem]{
            // Project 1: Web Application
            {
                ID:   "webapp",
                Item: FileItem{Name: "Web Application", IsFolder: true},
                Children: []tree.TreeData[FileItem]{
                    {
                        ID:   "webapp_src",
                        Item: FileItem{Name: "src", IsFolder: true},
                        Children: []tree.TreeData[FileItem]{
                            {
                                ID:   "webapp_main",
                                Item: FileItem{Name: "main.go", IsFolder: false},
                            },
                            {
                                ID:   "webapp_app",
                                Item: FileItem{Name: "app.go", IsFolder: false},
                            },
                        },
                    },
                },
            },
        },
        selectedNodes: make(map[string]bool),
    }
}
```

**Pro tip**: Add more projects from the basic tree example to see cascading selection work across larger structures.

## Step 3: Add Visual Feedback Imports

For cascading selection to be useful, you need to **see** what's selected. Add these imports:

```go
import (
    "fmt"
    "log"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/davidroman0O/vtable/core"
    "github.com/davidroman0O/vtable/tree"
)
```

**Why lipgloss?** We'll use it to style selected items with blue backgrounds.

## Step 4: Create the Selection Formatter

This is the same formatter from the basic tree - it makes selected items visible:

```go
func fileTreeFormatter(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
    // Extract the FileItem from the data
    if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
        content := flatItem.Item.String() // Get the icon and name
        
        // Apply blue background if selected
        if item.Selected {
            return lipgloss.NewStyle().
                Background(lipgloss.Color("12")). // Blue background
                Foreground(lipgloss.Color("15")). // White text
                Render(content)
        }
        
        return content
    }
    
    return fmt.Sprintf("%v", item.Item)
}
```

**What makes cascading visible**: When you select a parent, all children get the blue background, making the cascading effect immediately clear.

## Step 5: Enable Cascading Selection

Here's where the magic happens. It's just one configuration change:

```go
func main() {
    // Same data source as basic tree
    dataSource := NewFileTreeDataSource()
    
    // Same list configuration
    listConfig := core.ListConfig{
        ViewportConfig: core.ViewportConfig{
            Height:    10,
            ChunkSize: 20,
        },
        SelectionMode: core.SelectionMultiple, // Required for cascading
        KeyMap:        core.DefaultNavigationKeyMap(),
    }
}
```

**Critical requirement**: `SelectionMode` must be `SelectionMultiple`. Cascading selection needs to select multiple items at once.

## Step 6: Configure the Tree for Cascading

```go
// Start with default tree config
treeConfig := tree.DefaultTreeConfig()

// Enable cascading selection
treeConfig.CascadingSelection = true // ← This is the key line!

// Add visual feedback
treeConfig.RenderConfig.ContentConfig.Formatter = fileTreeFormatter
```

**That's it!** `CascadingSelection = true` enables the entire cascading behavior. VTable handles all the logic.

## Step 7: Add Cursor Styling

While we're at it, let's make the current cursor position visible too:

```go
// Enable background styling for cursor items
treeConfig.RenderConfig.BackgroundConfig.Enabled = true
treeConfig.RenderConfig.BackgroundConfig.Style = lipgloss.NewStyle().
    Background(lipgloss.Color("240")). // Gray background for cursor
    Foreground(lipgloss.Color("15"))   // White text
```

**Visual hierarchy**: 
- Gray background = current cursor position
- Blue background = selected items

## Step 8: Create the Tree Component

```go
// Create the tree with our configurations
treeComponent := tree.NewTreeList(listConfig, treeConfig, dataSource)
```

Simple - just pass in all our configurations.

## Step 9: Build the App with Selection Feedback

Let's create an app that shows how many items are selected:

```go
type App struct {
    tree   *tree.TreeList[FileItem]
    status string
}

func (app *App) Init() tea.Cmd {
    return app.tree.Init()
}
```

Standard Bubble Tea app structure.

## Step 10: Handle Input with Status Updates

```go
func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return app, tea.Quit
        case "enter":
            app.status = "Toggled expand/collapse"
            return app, app.tree.ToggleCurrentNode()
        case " ":
            app.status = "Selected item (cascading to children)"
            return app, core.SelectCurrentCmd()
        }
    }

    var cmd tea.Cmd
    _, cmd = app.tree.Update(msg)
    return app, cmd
}
```

**Key difference**: The status message mentions "cascading to children" to remind users about the behavior.

## Step 11: Add Clear Selection Command

With cascading selection, you'll often select many items at once. Add a way to clear them:

```go
// Add this case to the switch statement in Update():

case "c":
    app.status = "Cleared all selections"
    return app, core.SelectClearCmd()
```

**Why 'c' for clear?** It's intuitive and doesn't conflict with other common keys.

## Step 12: Add Navigation Commands

```go
// Add these cases to the switch statement:

case "up", "k":
    app.status = "Moved up"
    return app, core.CursorUpCmd()
case "down", "j":
    app.status = "Moved down"
    return app, core.CursorDownCmd()
case "pgup":
    app.status = "Page up"
    return app, core.PageUpCmd()
case "pgdn":
    app.status = "Page down"
    return app, core.PageDownCmd()
case "home", "g":
    app.status = "Jump to start"
    return app, core.JumpToStartCmd()
case "end", "G":
    app.status = "Jump to end"
    return app, core.JumpToEndCmd()
```

Same navigation as the basic tree example.

## Step 13: Display with Selection Count

```go
func (app *App) View() string {
    title := "🌳 Multi-Project Cascading Selection Demo"
    
    // Show how many items are selected
    selectionCount := app.tree.GetSelectionCount()
    selectionInfo := fmt.Sprintf("Selected: %d items", selectionCount)
    
    help := "Navigate: ↑/↓/j/k, Enter: expand/collapse, Space: select (cascades to children), c: clear, q: quit"
    status := fmt.Sprintf("Status: %s", app.status)

    return fmt.Sprintf("%s\n%s\n\n%s\n\n%s\n%s", 
        title, 
        selectionInfo,
        app.tree.View(), 
        status, 
        help)
}
```

**Selection count**: This number will jump when you select a folder with many children, clearly showing the cascading effect.

## Step 14: Create and Run the App

```go
// Create the app
app := &App{
    tree:   treeComponent,
    status: "Ready! Select a folder to see cascading selection with blue highlights",
}

// Run the application
p := tea.NewProgram(app)
if _, err := p.Run(); err != nil {
    log.Fatal(err)
}
```

## What You'll See

```
🌳 Multi-Project Cascading Selection Demo
Selected: 0 items

► 📁 Web Application
  📁 CLI Tool
  📁 API Service
  📁 Database Tools
  📁 Monitoring System

Status: Ready! Select a folder to see cascading selection with blue highlights
Navigate: ↑/↓/j/k, Enter: expand/collapse, Space: select (cascades to children), c: clear, q: quit
```

## Testing Cascading Selection

**Try this sequence**:

1. Navigate to "Web Application" and press Enter to expand it
2. Navigate to the "src" folder and press Space
3. Watch the selection count jump and see blue highlighting on all files in that folder

```
🌳 Multi-Project Cascading Selection Demo
Selected: 3 items

► 📁 Web Application
    📁 src (blue background)
      📄 main.go (blue background)
      📄 app.go (blue background)
  📁 CLI Tool

Status: Selected item (cascading to children)
```

**What happened**: Selecting the `src` folder automatically selected both `main.go` and `app.go`, jumping the count from 0 to 3.

## Visual Feedback

- **Current cursor**: Gray background highlighting  
- **Selected items**: Blue background with white text
- **Cascading effect**: When you select a parent, all children get blue highlighting
- **Selection count**: Shows total selected items in real-time

## Key Concepts

### 1. **Hierarchical Selection**
When you select a parent, all descendants are automatically selected. This follows the tree hierarchy.

### 2. **One Configuration Change**
Enabling cascading selection is just `CascadingSelection = true` in the tree config. VTable handles all the logic.

### 3. **Visual Feedback is Critical**
Without the blue highlighting, you wouldn't see what got selected. The formatter makes cascading selection useful.

### 4. **Selection Count Shows Impact**
The real-time count lets you see exactly how many items were affected by cascading selection.

## Try It Yourself

1. **Select entire projects**: Navigate to "Web Application" and press Space - see how many items get selected!
2. **Select subfolders**: Try selecting just smaller folders to see partial cascading
3. **Clear and try again**: Press 'c' to clear, then experiment with different folders
4. **Mix individual and cascading**: Select individual files, then select a folder to see them combine

## Advanced Behavior

When cascading selection is enabled:
- **Selecting a parent** selects all its children (shown with blue highlighting)
- **Deselecting a parent** deselects all its children  
- **Individual selection** still works for fine-grained control
- **Mixed selection** is supported (some children selected, some not)

## What's Next

You now understand cascading selection with proper visual feedback! Next, we'll explore customizing the tree symbols - changing those ▶ and ▼ arrows to +/- signs, custom icons, or anything you want.

The insight: **Cascading selection makes trees feel natural for bulk operations, and clear visual feedback shows exactly what's selected** - it respects the parent-child relationships in your data. 