# Tree Symbols

## What We're Adding

Taking our multi-project tree from previous examples, we're customizing the **tree symbols** - those little arrows that show whether folders are expanded or collapsed. Instead of the default ▶/▼ arrows, you can use +/- signs, custom icons, or even hide them completely.

## Understanding Tree Symbols

Tree symbols serve as visual indicators for node states:

```
Default style:        Plus/minus style:     Custom style:
▶ 📁 Folder          + 📁 Folder           [+] 📁 Folder
▼ 📁 Folder          - 📁 Folder           [-] 📁 Folder
  📄 file.txt           📄 file.txt             📄 file.txt
```

These symbols help users understand:
- **Which items can be expanded** (folders vs files)
- **Current expansion state** (opened vs closed)
- **How to interact** with the tree (click/press Enter to toggle)

## Step 1: Start with the Working Tree

We'll use the same structure from previous examples. If you followed those tutorials, you already have the working tree with multiple projects and selection styling.

```go
// Same FileItem and FileTreeDataSource from previous examples
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

**Why start here?** Tree symbols are purely visual - they don't change your data structure or logic, just how the tree looks.

## Step 2: Understanding Default Symbols

The default VTable tree uses these symbols:

```go
// This is what DefaultTreeConfig() sets up automatically
treeConfig := tree.DefaultTreeConfig()
// Internally this sets:
// - ExpandedSymbol: "▼"
// - CollapsedSymbol: "▶"  
// - LeafSymbol: "•"
```

You'll see this rendered as:

```
▼ 📁 Web Application
  ▶ 📁 src
  ▶ 📁 tests
▶ 📁 CLI Tool
• 📄 README.md
```

**Symbol meanings**:
- `▼` = expanded folder (showing children)
- `▶` = collapsed folder (children hidden)
- `•` = leaf item (files, no children)

## Step 3: Switch to Plus/Minus Symbols

Let's change to the classic +/- style that many file explorers use:

```go
func main() {
    // Same data source and list config as before
    dataSource := NewFileTreeDataSource()
    listConfig := core.ListConfig{
        ViewportConfig: core.ViewportConfig{
            Height:    10,
            ChunkSize: 20,
        },
        SelectionMode: core.SelectionMultiple,
        KeyMap:        core.DefaultNavigationKeyMap(),
    }
    
    // Start with default tree config
    treeConfig := tree.DefaultTreeConfig()
    treeConfig.RenderConfig.ContentConfig.Formatter = fileTreeFormatter
}
```

Now customize the symbols:

```go
// Customize tree symbols to use +/- style
treeConfig.RenderConfig.TreeSymbolConfig.ExpandedSymbol = "-"
treeConfig.RenderConfig.TreeSymbolConfig.CollapsedSymbol = "+"
treeConfig.RenderConfig.TreeSymbolConfig.LeafSymbol = " "
```

**What this does**:
- `"-"` for expanded folders (minus = "collapse this")
- `"+"` for collapsed folders (plus = "expand this") 
- `" "` (space) for files instead of bullet point

## Step 4: Add Symbol Spacing

The symbols might look cramped. Let's add some spacing:

```go
// Add spacing after symbols for better readability
treeConfig.RenderConfig.TreeSymbolConfig.SymbolSpacing = " "
```

Now your tree will look like:

```
- 📁 Web Application
  + 📁 src
  + 📁 tests
+ 📁 CLI Tool
  📄 README.md
```

**Better readability**: The space after symbols creates visual separation between the symbol and the folder icon.

## Step 5: Create Boxed Symbols

For a more distinct look, let's create boxed symbols:

```go
// Boxed symbol style
treeConfig.RenderConfig.TreeSymbolConfig.ExpandedSymbol = "[-]"
treeConfig.RenderConfig.TreeSymbolConfig.CollapsedSymbol = "[+]"
treeConfig.RenderConfig.TreeSymbolConfig.LeafSymbol = "   " // 3 spaces for alignment
treeConfig.RenderConfig.TreeSymbolConfig.SymbolSpacing = " "
```

This creates a more prominent visual style:

```
[-] 📁 Web Application
    [+] 📁 src
    [+] 📁 tests
[+] 📁 CLI Tool
    📄 README.md
```

**Why boxed?** The brackets make the interactive elements more obvious to users.

## Step 6: Use Unicode Symbols

For a modern look, try Unicode symbols:

```go
// Unicode arrow style
treeConfig.RenderConfig.TreeSymbolConfig.ExpandedSymbol = "◢"
treeConfig.RenderConfig.TreeSymbolConfig.CollapsedSymbol = "◤"
treeConfig.RenderConfig.TreeSymbolConfig.LeafSymbol = "◦"
treeConfig.RenderConfig.TreeSymbolConfig.SymbolSpacing = " "
```

Or try folder-like symbols:

```go
// Folder-style symbols
treeConfig.RenderConfig.TreeSymbolConfig.ExpandedSymbol = "📂"
treeConfig.RenderConfig.TreeSymbolConfig.CollapsedSymbol = "📁"
treeConfig.RenderConfig.TreeSymbolConfig.LeafSymbol = "📄"
treeConfig.RenderConfig.TreeSymbolConfig.SymbolSpacing = " "
```

**Creative options**: Unicode gives you many visual choices - arrows, geometric shapes, even emoji.

## Step 7: Style the Symbols

You can apply colors and styling to the symbols:

```go
import "github.com/charmbracelet/lipgloss"

// Add color styling to symbols
treeConfig.RenderConfig.TreeSymbolConfig.Style = lipgloss.NewStyle().
    Foreground(lipgloss.Color("12")). // Blue symbols
    Bold(true)
```

This makes symbols stand out with blue color and bold text.

## Step 8: Hide Symbols for Minimal Style

For a clean, minimal look, you can hide symbols entirely:

```go
// Minimal style with no symbols
treeConfig.RenderConfig.TreeSymbolConfig.Enabled = false
```

Your tree becomes:

```
📁 Web Application
  📁 src
  📁 tests
📁 CLI Tool
📄 README.md
```

**When to use**: When you want a clean list-like appearance but still need tree structure for navigation.

## Step 9: Control Symbol Visibility

You can show symbols only for folders, not files:

```go
// Show symbols only for items with children
treeConfig.RenderConfig.TreeSymbolConfig.ShowForLeaves = false
```

This hides the symbol for files:

```
▼ 📁 Web Application
  ▶ 📁 src
  ▶ 📁 tests
▶ 📁 CLI Tool
📄 README.md        // No symbol here
```

**Cleaner look**: Reduces visual noise by only showing symbols where they're meaningful.

## Step 10: Create the Complete Example

Let's put together a complete example with custom symbols:

```go
func main() {
    // Create data source (same as previous examples)
    dataSource := NewFileTreeDataSource()
    
    // Configure list settings
    listConfig := core.ListConfig{
        ViewportConfig: core.ViewportConfig{
            Height:    10,
            ChunkSize: 20,
        },
        SelectionMode: core.SelectionMultiple,
        KeyMap:        core.DefaultNavigationKeyMap(),
    }
    
    // Start with default tree config
    treeConfig := tree.DefaultTreeConfig()
    treeConfig.RenderConfig.ContentConfig.Formatter = fileTreeFormatter
    
    // Apply custom symbols
    treeConfig.RenderConfig.TreeSymbolConfig.ExpandedSymbol = "[-]"
    treeConfig.RenderConfig.TreeSymbolConfig.CollapsedSymbol = "[+]"
    treeConfig.RenderConfig.TreeSymbolConfig.LeafSymbol = "   "
    treeConfig.RenderConfig.TreeSymbolConfig.SymbolSpacing = " "
    treeConfig.RenderConfig.TreeSymbolConfig.ShowForLeaves = false
    
    // Add symbol styling
    treeConfig.RenderConfig.TreeSymbolConfig.Style = lipgloss.NewStyle().
        Foreground(lipgloss.Color("12")).
        Bold(true)
}
```

## Step 11: Build the App

```go
// Create tree component
treeComponent := tree.NewTreeList(listConfig, treeConfig, dataSource)

// Create app with updated title
app := &App{
    tree:   treeComponent,
    status: "Ready! Custom symbols make the tree style unique",
}

// Run the application
p := tea.NewProgram(app)
if _, err := p.Run(); err != nil {
    log.Fatal(err)
}
```

## What You'll See

```
🌳 Custom Symbol Tree Demo

[-] 📁 Web Application
    [+] 📁 src
    [+] 📁 tests
[+] 📁 CLI Tool
[+] 📁 API Service
📄 README.md

Status: Ready! Custom symbols make the tree style unique
Navigate: ↑/↓/j/k, Enter: expand/collapse, Space: select, q: quit
```

## Symbol Style Examples

### Classic Plus/Minus
```go
ExpandedSymbol: "-"
CollapsedSymbol: "+"
LeafSymbol: " "
```
```
- 📁 Projects
+ 📁 Archive
  📄 file.txt
```

### Boxed Style
```go
ExpandedSymbol: "[-]"
CollapsedSymbol: "[+]"
LeafSymbol: "   "
```
```
[-] 📁 Projects
[+] 📁 Archive
    📄 file.txt
```

### Arrow Style
```go
ExpandedSymbol: "↓"
CollapsedSymbol: "→"
LeafSymbol: "·"
```
```
↓ 📁 Projects
→ 📁 Archive
· 📄 file.txt
```

### Minimal Style
```go
Enabled: false
```
```
📁 Projects
📁 Archive
📄 file.txt
```

## Key Concepts

### 1. **Visual Hierarchy**
Symbols help users understand the tree structure at a glance - what can be expanded, what's currently open, and what are leaf items.

### 2. **Interaction Cues**
Good symbols make it obvious which items are interactive (can be expanded/collapsed) versus static (files).

### 3. **Style Consistency**
Choose symbols that match your application's overall design - minimal, detailed, colorful, or monochrome.

### 4. **Configuration Flexibility**
Every aspect of tree symbols can be customized - the characters, colors, spacing, and even visibility.

## Try It Yourself

1. **Test different symbols**: Try `"⮟"/"⮞"`, `"⯆"/"⯈"`, or `"🔽"/"▶️"`
2. **Experiment with colors**: Make symbols green for open, gray for closed
3. **Try no leaf symbols**: Set `ShowForLeaves: false` for cleaner files
4. **Create themed symbols**: Use `"📂"/"📁"` for a folder theme

## Symbol Best Practices

### Do:
- **Keep symbols simple** - they're seen frequently
- **Make them visually distinct** - users need to quickly differentiate states
- **Consider color accessibility** - don't rely only on color for meaning
- **Test with your data** - some symbols work better with certain content

### Don't:
- **Use too many different symbols** - stick to 2-3 maximum
- **Make symbols too large** - they shouldn't overwhelm the content
- **Forget about spacing** - cramped symbols are hard to read
- **Change symbols mid-application** - consistency helps usability

## What's Next

You now understand how to customize tree symbols! Next, we'll explore tree styling - adding colors, backgrounds, and visual themes to make your tree components look exactly how you want.

The insight: **Tree symbols are small details that make a big difference in usability** - they guide user understanding and interaction with hierarchical data. 