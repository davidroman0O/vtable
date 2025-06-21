# Basic Tree

## What We're Building

A hierarchical file explorer showing multiple projects with clean visual formatting. You'll see folders and files with proper icons and selection highlighting:

```
► 📁 Web Application
  📁 CLI Tool  
  📁 API Service
  📁 Database Tools
  📁 Monitoring System
```

## Understanding Trees vs Lists

A **List** shows items in a flat sequence: Item 1, Item 2, Item 3...

A **Tree** shows items in a **hierarchy**: Items can have children, which can have their own children, creating parent-child relationships.

VTable's tree component handles the complexity of flattening this hierarchy for efficient rendering while preserving the tree structure for navigation.

## Step 1: Define Your Data Structure

First, create a simple data type to represent files and folders:

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

**Why String()?** This provides the basic content for your items. We'll enhance the display with a custom formatter later.

## Step 2: Create the Data Source Structure

VTable trees need a `TreeDataSource` to provide the hierarchical data. Let's start with the basic structure:

```go
type FileTreeDataSource struct {
    rootNodes     []tree.TreeData[FileItem]
    selectedNodes map[string]bool
}
```

This structure holds:
- `rootNodes`: The top-level projects in our tree
- `selectedNodes`: A map tracking which items are selected

## Step 3: Build the Tree Data

Now let's create the actual tree structure. We'll start with one project to keep it simple:

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

**Understanding the structure**: Each `TreeData` has:
- `ID`: A unique string identifier
- `Item`: Your actual data (FileItem)
- `Children`: Optional nested TreeData objects

## Step 4: Add More Projects

Let's add a few more projects to demonstrate scrolling and navigation:

```go
// Add this after the Web Application project in the rootNodes slice:

// Project 2: CLI Tool
{
    ID:   "cli_tool",
    Item: FileItem{Name: "CLI Tool", IsFolder: true},
    Children: []tree.TreeData[FileItem]{
        {
            ID:   "cli_cmd",
            Item: FileItem{Name: "cmd", IsFolder: true},
            Children: []tree.TreeData[FileItem]{
                {
                    ID:   "cli_root",
                    Item: FileItem{Name: "root.go", IsFolder: false},
                },
            },
        },
    },
},

// Project 3: API Service
{
    ID:   "api_service",
    Item: FileItem{Name: "API Service", IsFolder: true},
    Children: []tree.TreeData[FileItem]{
        {
            ID:   "api_endpoints",
            Item: FileItem{Name: "endpoints", IsFolder: true},
            Children: []tree.TreeData[FileItem]{
                {
                    ID:   "api_users",
                    Item: FileItem{Name: "users.go", IsFolder: false},
                },
            },
        },
    },
},
```

**Why multiple projects?** This gives you enough content to test scrolling and navigation features.

## Step 5: Implement Core TreeDataSource Methods

The TreeDataSource interface requires several methods. Let's implement them one by one:

### Getting Root Nodes

```go
func (ds *FileTreeDataSource) GetRootNodes() []tree.TreeData[FileItem] {
    return ds.rootNodes
}
```

This is straightforward - just return the top-level projects.

### Finding Items by ID

```go
func (ds *FileTreeDataSource) GetItemByID(id string) (tree.TreeData[FileItem], bool) {
    return ds.findNodeByID(ds.rootNodes, id)
}

func (ds *FileTreeDataSource) findNodeByID(nodes []tree.TreeData[FileItem], id string) (tree.TreeData[FileItem], bool) {
    for _, node := range nodes {
        if node.ID == id {
            return node, true
        }
        if found, ok := ds.findNodeByID(node.Children, id); ok {
            return found, true
        }
    }
    return tree.TreeData[FileItem]{}, false
}
```

**How it works**: This recursively searches through the tree structure, checking each node and its children until it finds the matching ID.

## Step 6: Implement Selection Methods

Selection in trees works with Bubble Tea commands. Here are the key methods:

### Individual Selection

```go
func (ds *FileTreeDataSource) SetSelected(id string, selected bool) tea.Cmd {
    if selected {
        ds.selectedNodes[id] = true
    } else {
        delete(ds.selectedNodes, id)
    }
    return core.SelectionResponseCmd(true, -1, id, selected, "toggle", nil, nil)
}

func (ds *FileTreeDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
    return ds.SetSelected(id, selected)
}
```

**Key insight**: These methods return `tea.Cmd` instead of directly modifying state. This keeps the interface async and lets VTable handle the updates properly.

### Bulk Selection

```go
func (ds *FileTreeDataSource) SelectAll() tea.Cmd {
    ds.selectAllNodes(ds.rootNodes)
    return core.SelectionResponseCmd(true, -1, "", true, "selectAll", nil, nil)
}

func (ds *FileTreeDataSource) selectAllNodes(nodes []tree.TreeData[FileItem]) {
    for _, node := range nodes {
        ds.selectedNodes[node.ID] = true
        ds.selectAllNodes(node.Children)
    }
}

func (ds *FileTreeDataSource) ClearSelection() tea.Cmd {
    ds.selectedNodes = make(map[string]bool)
    return core.SelectionResponseCmd(true, -1, "", false, "clear", nil, nil)
}

func (ds *FileTreeDataSource) SelectRange(startID, endID string) tea.Cmd {
    ds.selectedNodes[startID] = true
    ds.selectedNodes[endID] = true
    return core.SelectionResponseCmd(true, -1, "", true, "range", nil, []string{startID, endID})
}
```

**Why separate methods?** Each handles a different selection pattern - individual, all, clear, and range selection.

## Step 7: Add Visual Feedback with Custom Formatting

To show clean file names and highlight selected items, we need a custom formatter:

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

First, add the lipgloss import - we'll use it for styling.

### The Custom Formatter Function

```go
func fileTreeFormatter(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
    // Extract the FileItem from the data
    if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
        content := flatItem.Item.String() // Use our FileItem's String() method
        
        // Apply selection styling if item is selected
        if item.Selected {
            return lipgloss.NewStyle().
                Background(lipgloss.Color("12")). // Blue background for selected
                Foreground(lipgloss.Color("15")). // White text
                Render(content)
        }
        
        return content
    }
    
    // Fallback to default formatting
    return fmt.Sprintf("%v", item.Item)
}
```

**What this does**:
- Extracts your FileItem from VTable's internal data structure
- Applies blue background styling to selected items
- Falls back to basic formatting if something goes wrong

**Why do we need this?** Without a custom formatter, you'd see raw struct data like `{main.go false}` instead of clean icons like `📄 main.go`.

## Step 8: Create the App Structure

Now let's build the Bubble Tea app that uses our tree:

```go
type App struct {
    tree   *tree.TreeList[FileItem]
    status string
}

func (app *App) Init() tea.Cmd {
    return app.tree.Init()
}
```

Simple structure - just the tree component and a status message.

### Handle User Input

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
            app.status = "Toggled selection"
            return app, core.SelectCurrentCmd()
        }
    }
    
    // Pass other messages to the tree
    var cmd tea.Cmd
    _, cmd = app.tree.Update(msg)
    return app, cmd
}
```

**Key actions**:
- `Enter`: Expand/collapse folders
- `Space`: Select/deselect items
- Everything else gets passed to the tree for navigation

### Add Navigation Commands

```go
// Add these cases to the switch statement in Update():

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

**Navigation patterns**: Standard arrow keys, Vim-style j/k, page navigation, and home/end jumping.

### Display the Tree

```go
func (app *App) View() string {
    title := "🌳 Multi-Project File Tree Demo"
    help := "Navigate: ↑/↓/j/k, Page: PgUp/PgDn, Jump: Home/End/g/G, Enter: expand/collapse, Space: select, q: quit"
    status := fmt.Sprintf("Status: %s", app.status)

    return fmt.Sprintf("%s\n\n%s\n\n%s\n%s", title, app.tree.View(), status, help)
}
```

Clean layout with title, tree content, status, and help text.

## Step 9: Configure the Tree Component

Now let's put it all together with proper configuration:

### List Configuration

```go
func main() {
    // Create the data source
    dataSource := NewFileTreeDataSource()
    
    // Configure basic list settings
    listConfig := core.ListConfig{
        ViewportConfig: core.ViewportConfig{
            Height:    10,  // Show 10 lines at once
            ChunkSize: 20,  // Load 20 items per chunk
        },
        SelectionMode: core.SelectionMultiple, // Allow multiple selections
        KeyMap:        core.DefaultNavigationKeyMap(),
    }
}
```

**Configuration explained**:
- `Height`: How many lines to show in the viewport
- `ChunkSize`: How many items to load at once (for performance)
- `SelectionMode`: Allow selecting multiple items
- `KeyMap`: Use VTable's default navigation keys

### Tree Configuration

```go
// Configure tree with custom formatter and styling
treeConfig := tree.DefaultTreeConfig()
treeConfig.RenderConfig.ContentConfig.Formatter = fileTreeFormatter

// Enable background styling for cursor items
treeConfig.RenderConfig.BackgroundConfig.Enabled = true
treeConfig.RenderConfig.BackgroundConfig.Style = lipgloss.NewStyle().
    Background(lipgloss.Color("240")). // Gray background for cursor
    Foreground(lipgloss.Color("15"))   // White text
```

**Styling setup**:
- Apply our custom formatter to show clean file names
- Enable gray background highlighting for the current cursor position

### Create and Run the App

```go
// Create the tree component
treeComponent := tree.NewTreeList(listConfig, treeConfig, dataSource)

// Create the app
app := &App{
    tree:   treeComponent,
    status: "Ready! Navigate with arrows, Enter to expand/collapse, Space to select (blue highlight)",
}

// Run the application
p := tea.NewProgram(app)
if _, err := p.Run(); err != nil {
    log.Fatal(err)
}
```

**Final assembly**: Connect all the pieces and start the Bubble Tea program.

## What You'll See

```
🌳 Multi-Project File Tree Demo

► 📁 Web Application
  📁 CLI Tool
  📁 API Service
  📁 Database Tools
  📁 Monitoring System

Status: Ready! Navigate with arrows, Enter to expand/collapse, Space to select (blue highlight)
Navigate: ↑/↓/j/k, Page: PgUp/PgDn, Jump: Home/End/g/G, Enter: expand/collapse, Space: select, q: quit
```

## Visual Feedback

- **Current cursor**: Gray background highlighting
- **Selected items**: Blue background with white text  
- **Tree symbols**: ▶ for collapsed folders, ▼ for expanded folders, • for files
- **Clean icons**: 📁 for folders, 📄 for files

## Key Concepts

### 1. **Hierarchical Data Structure**
Trees are built using nested `TreeData[T]` objects, each with an ID, your data, and optional children.

### 2. **Automatic Flattening**
VTable automatically converts your tree structure into a flat list for efficient rendering and navigation.

### 3. **Custom Formatting**
Custom formatters provide clean display and visual feedback for selection and cursor states.

### 4. **Command Pattern**
Selection methods return `tea.Cmd` instead of directly modifying state, keeping the interface responsive.

## Try It Yourself

1. **Navigate between projects**: Use arrow keys to move between different root projects
2. **Expand folders**: Press Enter on folders to see their contents
3. **Select items**: Press Space to select files and folders - see the blue highlighting
4. **Test scrolling**: With multiple projects, you can test Page Up/Down navigation

## What's Next

You now have a working tree with proper visual feedback! Next, we'll add cascading selection - when you select a folder, it automatically selects all files inside it.

The core insight: **Trees are just lists with parent-child relationships and proper visual feedback makes the interaction clear**. VTable handles the complexity while giving you full control over the structure and appearance.