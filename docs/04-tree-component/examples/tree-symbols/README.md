# Tree Symbols Example

This example demonstrates how to customize tree symbols in VTable - the little indicators that show whether folders are expanded, collapsed, or are leaf items.

## What You'll Learn

- How to change tree symbols from default ▶/▼ to custom styles
- Different symbol style options (plus/minus, boxed, Unicode, emoji)
- How to hide symbols for a minimal look
- How to dynamically change symbol styles at runtime
- Styling symbols with colors and formatting

## Running the Example

```bash
cd docs/04-tree-component/examples/tree-symbols
go run main.go
```

## Controls

- **↑/↓ or j/k**: Navigate up/down the tree
- **Enter**: Expand/collapse folders
- **Space**: Select/deselect items (blue highlighting)
- **s**: Cycle through different symbol styles
- **c**: Clear all selections
- **Page Up/Down**: Navigate faster through long lists
- **Home/End or g/G**: Jump to start/end
- **q or Ctrl+C**: Quit

## Symbol Styles

Press **'s'** to cycle through these different styles:

### 1. Default
```
▼ 📁 Web Application
  ▶ 📁 src
  • 📄 main.go
```
VTable's default arrow style with bullet points for files.

### 2. Plus/Minus
```
- 📁 Web Application
  + 📁 src
  📄 main.go
```
Classic file explorer style - familiar to most users.

### 3. Boxed
```
[-] 📁 Web Application
    [+] 📁 src
    📄 main.go
```
Boxed symbols that stand out clearly from content.

### 4. Unicode
```
◢ 📁 Web Application
  ◤ 📁 src
  ◦ 📄 main.go
```
Modern Unicode symbols for a contemporary look.

### 5. Emoji
```
📂 📁 Web Application
  📁 📁 src
  📄 📄 main.go
```
Emoji-based symbols that match the folder/file theme.

### 6. Minimal
```
📁 Web Application
  📁 src
  📄 main.go
```
No symbols at all - clean, list-like appearance.

## Key Features Demonstrated

### Dynamic Style Switching
The example shows how to change symbol styles at runtime using `GetRenderConfig()` and `SetRenderConfig()`:

```go
func (app *App) applySymbolStyle() {
    style := symbolStyles[app.currentStyle]
    treeConfig := app.tree.GetRenderConfig()
    
    // Apply new symbol configuration
    treeConfig.TreeSymbolConfig.ExpandedSymbol = style.ExpandedSymbol
    treeConfig.TreeSymbolConfig.CollapsedSymbol = style.CollapsedSymbol
    // ... more config
    
    app.tree.SetRenderConfig(treeConfig)
}
```

### Symbol Configuration Options
Each style demonstrates different configuration aspects:

- **Symbol characters**: What text/Unicode to display
- **Spacing**: How much space after symbols
- **Leaf visibility**: Whether to show symbols for files
- **Colors and styling**: Using lipgloss for visual enhancement
- **Enable/disable**: Turning symbols on or off entirely

### Visual Feedback
The example maintains:
- **Selection highlighting**: Blue background for selected items
- **Cursor highlighting**: Gray background for current position
- **Status updates**: Shows which style is currently active
- **Clean formatting**: Custom formatter for readable file names

## Code Structure

The example is organized to be educational:

1. **Symbol Style Definitions**: Array of different style configurations
2. **Dynamic Application**: Method to apply styles to the tree
3. **Interactive Controls**: Key handling for style cycling
4. **Visual Feedback**: Status and help text showing current state

## What to Try

1. **Expand some folders** first, then cycle through styles to see how different symbols look
2. **Select some items** to see how selection highlighting works with different symbols
3. **Try the minimal style** to see how trees look without symbols
4. **Notice the color differences** between styles (blue, green, purple symbols)

## Learning Path

This example builds on:
- [Basic Tree](../basic-tree/) - Tree structure and navigation
- [Cascading Selection](../cascading-selection/) - Selection behavior

Next examples in the tree series:
- Tree Styling - Colors and visual themes
- Tree Indentation - Custom spacing and alignment
- Connected Lines - Box-drawing tree connectors

## Key Insight

Tree symbols are small visual elements that have a big impact on usability. They help users understand:
- What can be interacted with (folders vs files)
- Current state (expanded vs collapsed)
- Visual hierarchy and structure

Choose symbols that match your application's design and your users' expectations! 