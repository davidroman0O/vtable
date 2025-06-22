# The Tree Component: Indentation

Indentation is what gives a tree its visual structure. This guide explains how to control the horizontal spacing that represents the depth of each node in the hierarchy, using VTable's `TreeIndentationComponent`.

## What You'll Build

You will learn to configure different indentation styles, transforming a simple list into a clearly structured tree.

**Standard 2-Space Indent:**
```
📁 Project
  📁 src
    📄 main.go
```

**Custom String Indent (`··`):**
```
📁 Project
··📁 src
····📄 main.go
```

**Styled String Indent (`│ `):**
```
📁 Project
│ 📁 src
│ │ 📄 main.go
```

## How It Works: The `TreeIndentationConfig`

Indentation is managed by the `TreeIndentationConfig` within your `TreeRenderConfig`.

```go
// Get the tree's render configuration
renderConfig := myTree.GetRenderConfig()

// Access the indentation configuration
indentConfig := &renderConfig.IndentationConfig

// Customize the indentation
indentConfig.Enabled = true
indentConfig.IndentSize = 4 // Use 4 spaces per level
indentConfig.IndentString = "" // An empty string means use IndentSize

// Apply styling to the indentation characters
indentConfig.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

// Apply the changes
myTree.SetRenderConfig(renderConfig)
```

### Key `TreeIndentationConfig` Properties
-   `Enabled`: Toggles the indentation component on or off.
-   `IndentSize`: The number of spaces to use for each level of depth. Only used if `IndentString` is empty.
-   `IndentString`: A custom string to repeat for each level of depth (e.g., `"  "`, `"··"`, `"— "`).
-   `Style`: A `lipgloss.Style` applied to the `IndentString`.
-   `UseConnectors`: A boolean to enable classic box-drawing connector lines (covered in the next guide).

## Step 1: Adjusting Indentation Size

The simplest way to change the layout is to adjust the `IndentSize`.

```go
renderConfig := myTree.GetRenderConfig()

// Use a wider, 4-space indent for more clarity
renderConfig.IndentationConfig.IndentSize = 4
renderConfig.IndentationConfig.IndentString = "" // Ensure we're using spaces

myTree.SetRenderConfig(renderConfig)
```

## Step 2: Using a Custom Indentation String

For a more distinct visual style, you can provide a custom string to be repeated.

```go
renderConfig := myTree.GetRenderConfig()

// Use a dotted line for indentation
renderConfig.IndentationConfig.IndentString = "··"
renderConfig.IndentationConfig.IndentSize = 0 // IndentSize is ignored

// Style the dots to be subtle
renderConfig.IndentationConfig.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

myTree.SetRenderConfig(renderConfig)
```

This will produce a tree where each level of depth is represented by `··`, creating a clear visual guide.

## What You'll Experience

-   **Layout Control**: Easily switch between compact (small `IndentSize`) and spacious (large `IndentSize`) layouts.
-   **Visual Theming**: Use custom strings and styles to match your application's aesthetic.
-   **Improved Readability**: Well-configured indentation makes complex hierarchies much easier to understand at a glance.

## Complete Example

See the full working code, which includes an interactive demo for cycling through different indentation themes.
[`docs/04-tree-component/examples/tree-indentation/`](examples/tree-indentation/)

To run it:
```bash
cd docs/04-tree-component/examples/tree-indentation
go run main.go
```
Press the `i` key in the running application to cycle through the different indentation styles.

## What's Next?

You now know how to control the spacing and style of your tree's hierarchy. To create an even clearer structure, the next guide will show you how to use box-drawing characters to create **connected lines** between parent and child nodes.

**Next:** [Connected Lines →](06-connected-lines.md) 