# The Tree Component: Basic Usage

Let's build your first VTable **Tree**. The Tree component is designed to display hierarchical data, such as file systems, organizational charts, or nested categories, while still benefiting from VTable's high-performance data virtualization.

## What You'll Build

A basic, navigable file explorer that shows a nested structure of projects, folders, and files. You'll be able to expand and collapse folders to explore the hierarchy.

![Basic Tree Navigation](examples/basic-tree/basic-tree.gif)

## How Trees Differ from Lists

-   **Lists** display a flat sequence of items.
-   **Trees** display a **hierarchy**, where items can have parent-child relationships.

VTable's Tree component handles the complexity of flattening this hierarchy for efficient rendering while allowing you to navigate the structure intuitively.

## Step 1: Define Your Hierarchical Data

First, define the data structure for your tree nodes. We'll represent files and folders.

```go
// FileItem represents our data for a single node.
type FileItem struct {
	Name     string
	IsFolder bool
}

// The String() method provides a default text representation.
func (f FileItem) String() string {
	if f.IsFolder {
		return "📁 " + f.Name
	}
	return "📄 " + f.Name
}
```
Next, you need to structure this data hierarchically using `tree.TreeData`.

```go
import "github.com/davidroman0O/vtable/tree"

// The TreeData struct wraps your item and its children.
var myTree = []tree.TreeData[FileItem]{
    {
        ID:   "webapp", // A stable, unique ID for this node
        Item: FileItem{Name: "Web Application", IsFolder: true},
        Children: []tree.TreeData[FileItem]{ // Nested children
            {
                ID:   "webapp_src",
                Item: FileItem{Name: "src", IsFolder: true},
                Children: []tree.TreeData[FileItem]{
                    {ID: "webapp_main", Item: FileItem{Name: "main.go"}},
                },
            },
        },
    },
    // ... more root nodes
}
```

## Step 2: Implement the `TreeDataSource`

The `TreeDataSource` is similar to the `List`'s `DataSource`, but it's designed for hierarchical data.

```go
type FileTreeDataSource struct {
	rootNodes     []tree.TreeData[FileItem]
	selectedNodes map[string]bool
}

// GetRootNodes returns the top-level nodes of the tree.
func (ds *FileTreeDataSource) GetRootNodes() []tree.TreeData[FileItem] {
	return ds.rootNodes
}

// GetItemByID recursively finds a node by its ID.
func (ds *FileTreeDataSource) GetItemByID(id string) (tree.TreeData[FileItem], bool) {
	// ... recursive search implementation ...
}

// The TreeList component does NOT use LoadChunk. It builds its view
// by traversing the root nodes you provide. The selection methods
// remain the same as the List's DataSource.
```
Unlike a `List`, the `TreeList` doesn't use `LoadChunk`. It builds its flattened, visible view by recursively traversing the `rootNodes` you provide, respecting the expansion state of each node.

## Step 3: Create the Tree Component

Create the `TreeList` component, passing in a `ListConfig` (for viewport settings) and a `TreeConfig` (for tree-specific appearance).

```go
import "github.com/davidroman0O/vtable/tree"

func createTree() *tree.TreeList[FileItem] {
    dataSource := NewFileTreeDataSource()

    // Trees reuse the standard ListConfig for viewport management.
    listConfig := core.ListConfig{
        ViewportConfig: core.ViewportConfig{Height: 10, ChunkSize: 20},
        SelectionMode:  core.SelectionMultiple,
    }

    // TreeConfig controls tree-specific visuals.
    treeConfig := tree.DefaultTreeConfig()

    // Create the TreeList component.
    return tree.NewTreeList(listConfig, treeConfig, dataSource)
}
```

## Step 4: Integrate with Bubble Tea

The integration is nearly identical to the `List` component.

```go
type App struct {
	tree *tree.TreeList[FileItem]
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// --- Navigation ---
		case "up", "k":
			return app, core.CursorUpCmd()
		case "down", "j":
			return app, core.CursorDownCmd()

		// --- Tree-Specific Actions ---
		case "enter":
			// Toggle expand/collapse on the current node.
			return app, app.tree.ToggleCurrentNode()
		case " ":
			// Select the current node.
			return app, core.SelectCurrentCmd()
		}
	}

	// Pass all other messages to the tree component.
	var cmd tea.Cmd
	_, cmd = app.tree.Update(msg)
	return app, cmd
}
```

## What You'll Experience

-   **Hierarchical View**: The application displays your nested data structure.
-   **Expand/Collapse**: Press `Enter` on a folder to expand or collapse it, revealing or hiding its children.
-   **Smooth Navigation**: Use arrow keys to navigate the flattened list view seamlessly.
-   **Selection**: Press `spacebar` to select items. The visual feedback will depend on your formatter and theme.

## Complete Example

See the full working code for this guide in the examples directory:
[`docs/04-tree-component/examples/basic-tree/`](examples/basic-tree/)

To run it:
```bash
cd docs/04-tree-component/examples/basic-tree
go run main.go
```

## What's Next?

You've built a functional tree view. Next, we'll enhance it by enabling **cascading selection**, a powerful feature where selecting a parent node automatically selects all of its descendants.

**Next:** [Cascading Selection →](02-cascading-selection.md)