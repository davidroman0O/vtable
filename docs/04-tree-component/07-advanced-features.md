# The Tree Component: Advanced Features

This guide covers the advanced, power-user features of VTable's `TreeList`. You'll learn how to implement bulk operations like "Expand/Collapse All," manipulate entire subtrees, and enable smart navigation shortcuts to create a highly efficient and productive user experience.

## What You'll Build

We will add a suite of advanced commands to our file tree, turning it from a simple browser into a powerful organizational tool.

-   **Bulk Operations**: `E` to Expand All, `C` to Collapse All.
-   **Subtree Manipulation**: `e`/`c` to expand/collapse the current folder and all its children.
-   **Cascading Selection**: Selecting a folder will automatically select its entire contents.
-   **Auto-Expansion**: Automatically expand specific nodes on startup.

## How It Works: New `TreeList` Methods

To support these features, the `TreeList` component provides several powerful methods that you can call from your application.

#### Subtree Operations
-   `tree.ExpandSubtree(id)`: Expands a node and all of its descendants recursively.
-   `tree.CollapseSubtree(id)`: Collapses a node and all its descendants.
-   `tree.ExpandCurrentSubtree()`: A convenience method that operates on the currently focused node.
-   `tree.CollapseCurrentSubtree()`: A convenience method for the current node.

#### Bulk Operations
-   `tree.ExpandAll()`: Expands every single node in the entire tree.
-   `tree.CollapseAll()`: Collapses every node, showing only the top-level roots.

#### Cascading Selection
-   `tree.SetCascadingSelection(enabled bool)`: A configuration method to enable or disable this feature.

## Step 1: Enable Advanced Features in `TreeConfig`

In your `main` function, enable the advanced features you want to use.

```go
// In your main function:
treeConfig := tree.DefaultTreeConfig()

// Enable cascading selection for intuitive folder selection.
treeConfig.CascadingSelection = true

// Use connected lines for a clear visual hierarchy.
treeConfig.RenderConfig.IndentationConfig.UseConnectors = true

// Use a formatter that provides visual feedback for these features.
treeConfig.RenderConfig.ContentConfig.Formatter = createAdvancedFormatter()
```

## Step 2: Implement Advanced Key Mappings

In your app's `Update` method, map simple, memorable keys to these powerful commands.

```go
func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// --- Basic tree operations ---
		case "enter": // Toggle a single node
			return app, app.tree.ToggleCurrentNode()
		case " ": // Select a node (and its children, if cascading is on)
			return app, core.SelectCurrentCmd()

		// --- ADVANCED OPERATIONS ---
		case "E": // Expand All (Shift+e)
			return app, app.tree.ExpandAll()
		case "C": // Collapse All (Shift+c)
			return app, app.tree.CollapseAll()
		case "e": // Expand current subtree
			return app, app.tree.ExpandCurrentSubtree()
		case "c": // Collapse current subtree
			return app, app.tree.CollapseCurrentSubtree()

		// --- Selection ---
		case "a": // Select All (standard shortcut)
			return app, core.SelectAllCmd()
		case "x": // Clear selection (eXclude)
			return app, core.SelectClearCmd()
		}
	}
	// ... rest of update logic ...
}
```

## Step 3: Provide Visual Feedback

An advanced formatter can show the state of folders, making the UI more informative.

```go
func createAdvancedFormatter() func(...) string {
	return func(item core.Data[any], ...) string {
		if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
			content := flatItem.Item.String()

			// NEW: Add a visual hint about the folder's state.
			if flatItem.Item.IsFolder && hasChildren {
				if isExpanded {
					content += " (expanded)"
				} else {
					content += " (...)"
				}
			}

			// Apply selection and other styling...
			// ...
			return styledContent
		}
		return fmt.Sprintf("%v", item.Item)
	}
}
```

## What You'll Experience

-   **Total Control**: With a single keystroke (`E`), you can instantly see every file in your entire project structure. Another key (`C`) collapses it back to a clean overview.
-   **Focused Workflow**: Navigate to a specific project folder, press `e`, and instantly see its entire contents without affecting other parts of the tree.
-   **Efficient Selection**: Select an entire project directory for a bulk operation by simply navigating to it and pressing the spacebar.

## Complete Example

See the full working code, which demonstrates all of these advanced features in an interactive application.
[`docs/04-tree-component/examples/advanced-features/`](examples/advanced-features/)

To run it:
```bash
cd docs/04-tree-component/examples/advanced-features
go run .
```

## What's Next?

This guide concludes the series on the Tree component. You have now learned everything from basic tree creation to implementing sophisticated, power-user features. The same core principles of `DataSources`, `ViewModels`, and `Component Rendering` apply to all VTable components.

**Next:** [The Table Component: Basic Usage →](../05-table-component/01-basic-table.md) 