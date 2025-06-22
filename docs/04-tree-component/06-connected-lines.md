# The Tree Component: Connected Lines

Let's enhance our tree's visual structure by adding **connected lines**. This feature uses box-drawing characters to create clear visual connections between parent and child nodes, giving your tree a classic, professional file-explorer look.

## What You'll Build

We will transform our indented tree into one with clear, connecting lines that visually represent the hierarchy.

![VTable Connected Lines Example](examples/connected-lines/connected-lines.gif)

**Before (Simple Indentation):**
```📁 Project
  📁 src
    📄 main.go
```

**After (Connected Lines):**
```
📁 Project
├── 📁 src
│   └── 📄 main.go
└── 📁 tests
```

## How It Works: The `UseConnectors` Flag

This feature is part of VTable's `TreeIndentationComponent`. You can enable it with a single boolean flag.

```go
// Get the tree's render configuration
renderConfig := myTree.GetRenderConfig()

// Access the indentation configuration
indentConfig := &renderConfig.IndentationConfig

// Enable connectors
indentConfig.UseConnectors = true

// Optionally, style the connector lines
indentConfig.ConnectorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

// Apply the changes
myTree.SetRenderConfig(renderConfig)
```

When `UseConnectors` is `true`, the `IndentationComponent` will automatically render the appropriate box-drawing characters (`├`, `│`, `└`, `─`) based on the node's position in the tree (e.g., if it's the last child, it gets a `└` instead of a `├`).

## Step 1: Enable Connectors

Starting with the code from the previous guide, enable the `UseConnectors` flag in your `TreeConfig`.

```go
// In your main function:
treeConfig := tree.DefaultTreeConfig()

// Enable the indentation component and tell it to use connectors.
treeConfig.RenderConfig.IndentationConfig.Enabled = true
treeConfig.RenderConfig.IndentationConfig.UseConnectors = true
```

## Step 2: Style the Connectors

You can style the color and font weight of the connector lines using the `ConnectorStyle` property.

```go
// Style the connector lines to be a subtle gray.
treeConfig.RenderConfig.IndentationConfig.ConnectorStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("240"))

// Or make them bolder for higher contrast.
treeConfig.RenderConfig.IndentationConfig.ConnectorStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("244")).
    Bold(true)
```

## What You'll Experience

-   **Clear Hierarchy**: The connecting lines make the parent-child relationships instantly obvious.
-   **Professional Look**: Gives your TUI the polished feel of a native file explorer or IDE.
-   **Easy Configuration**: A single boolean flag toggles this powerful visual feature.

## Complete Example

See the full working code, which includes an interactive demo for cycling through different connector styles.
[`docs/04-tree-component/examples/connected-lines/`](examples/connected-lines/)

To run it:
```bash
cd docs/04-tree-component/examples/connected-lines
go run main.go
```
Press the `l` key in the running application to cycle through different connector line styles.

## What's Next?

You have now mastered the visual layout of the tree component. The final guide in this series will cover advanced tree operations, such as expanding and collapsing entire subtrees and other power-user features.

**Next:** [Advanced Tree Features →](07-advanced-features.md) 
