# Tree Indentation Example

This example demonstrates how to customize tree indentation - the horizontal spacing that visually represents the hierarchical depth in tree structures. Experience different indentation approaches and see how they affect readability and user experience.

## What You'll Learn

- How to configure indentation size and style
- Different indentation approaches (spaces, custom strings, visual markers)
- Impact of indentation on tree readability
- Component-based indentation configuration
- Depth-aware formatting for deep trees
- Visual hierarchy through spacing

## Running the Example

```bash
cd docs/04-tree-component/examples/tree-indentation
go run main.go
```

## Controls

- **↑/↓ or j/k**: Navigate up/down the tree
- **Enter**: Expand/collapse folders
- **Space**: Select/deselect items
- **i**: Cycle through different indentation styles
- **c**: Clear all selections
- **Page Up/Down**: Navigate faster through long lists
- **Home/End or g/G**: Jump to start/end
- **q or Ctrl+C**: Quit

## Indentation Styles

Press **'i'** to cycle through these different indentation approaches:

### 1. Minimal (2 spaces)
```
📁 Web Application
  📁 src
    📄 main.go
    📄 app.go
    📁 handlers
      [L3] 📄 user_handler.go
      [L3] 📄 auth_handler.go
```
**Clean 2-space indentation** - Standard approach that balances clarity with space efficiency.

### 2. Spacious (4 spaces)
```
📁 Web Application
    📁 src
        📄 main.go
        📄 app.go
        📁 handlers
            [L3] 📄 user_handler.go
            [L3] 📄 auth_handler.go
```
**Wide 4-space indentation for clarity** - More generous spacing that makes deep hierarchies easier to follow.

### 3. Compact (1 space)
```
📁 Web Application
 📁 src
  📄 main.go
  📄 app.go
  📁 handlers
   [L3] 📄 user_handler.go
   [L3] 📄 auth_handler.go
```
**Minimal 1-space indentation for dense trees** - Conserves horizontal space for very wide trees.

### 4. Dotted
```
📁 Web Application
··📁 src
····📄 main.go
····📄 app.go
····📁 handlers
······[L3] 📄 user_handler.go
······[L3] 📄 auth_handler.go
```
**Gray dots show hierarchy clearly** - Visual dots make the depth structure obvious.

### 5. Dashed
```
📁 Web Application
- 📁 src
- - 📄 main.go
- - 📄 app.go
- - 📁 handlers
- - - [L3] 📄 user_handler.go
- - - [L3] 📄 auth_handler.go
```
**Dashes for distinctive hierarchy** - Dash characters create clear visual levels.

### 6. Arrows
```
📁 Web Application
→ 📁 src
→ → 📄 main.go
→ → 📄 app.go
→ → 📁 handlers
→ → → [L3] 📄 user_handler.go
→ → → [L3] 📄 auth_handler.go
```
**Arrow indicators pointing to content** - Arrows suggest navigation direction and depth.

### 7. Bullets
```
📁 Web Application
• 📁 src
• • 📄 main.go
• • 📄 app.go
• • 📁 handlers
• • • [L3] 📄 user_handler.go
• • • [L3] 📄 auth_handler.go
```
**Bullet points for each level** - Familiar bullet-style hierarchy from documents.

### 8. Boxed
```
📁 Web Application
│ 📁 src
│ │ 📄 main.go
│ │ 📄 app.go
│ │ 📁 handlers
│ │ │ [L3] 📄 user_handler.go
│ │ │ [L3] 📄 auth_handler.go
```
**Box-drawing characters for structure** - Creates visual "guide lines" showing tree structure.

## Key Features Demonstrated

### Component-Based Indentation Control

The example shows how to use VTable's `TreeIndentationComponent` system:

```go
// Configure indentation through TreeIndentationConfig
treeConfig.IndentationConfig.Enabled = true
treeConfig.IndentationConfig.IndentSize = 4        // Spaces per level
treeConfig.IndentationConfig.IndentString = "··"   // Custom string per level
treeConfig.IndentationConfig.Style = grayStyle     // Styling for indentation
```

### Dynamic Indentation Switching

Shows how to change indentation styles at runtime:

```go
func (app *App) applyIndentationTheme() {
    theme := indentationThemes[app.currentIndentation]
    treeConfig := app.tree.GetRenderConfig()
    
    // Apply new indentation settings
    treeConfig.IndentationConfig.IndentString = theme.IndentString
    treeConfig.IndentationConfig.IndentSize = theme.IndentSize
    treeConfig.IndentationConfig.Style = theme.Style
    
    app.tree.SetRenderConfig(treeConfig)
}
```

### Depth-Aware Formatting

For deep trees (level 3+), the example adds depth indicators:

```go
// Add depth indicators for very deep items
var depthIndicator string
if depth > 2 {
    depthIndicator = fmt.Sprintf("[L%d] ", depth)  // Shows level number
}
```

**Result**: Deep items show `[L3]`, `[L4]`, etc. to help users understand their depth.

### Indentation Types

**Space-Based Indentation**:
- Uses `IndentSize` to control spaces per level
- Set `IndentString = ""` to use spaces
- Most familiar and compatible approach

**String-Based Indentation**:
- Uses `IndentString` to define custom characters
- Set `IndentSize = 0` when using custom strings
- Allows any string pattern (dots, dashes, arrows, etc.)

## Visual Hierarchy Impact

### Minimal vs Spacious
- **Minimal (2 spaces)**: Good balance for most applications
- **Spacious (4 spaces)**: Better for complex trees or accessibility
- **Compact (1 space)**: Conserves space but may be harder to follow

### Visual Markers vs Spaces
- **Dots/Dashes**: Make hierarchy structure more obvious
- **Arrows**: Suggest navigation direction
- **Box characters**: Create visual "guide lines"
- **Spaces**: Clean and familiar, don't compete with content

### Deep Tree Considerations
- **Depth indicators**: Help users understand their position in deep hierarchies
- **Responsive indentation**: Could adjust based on available width
- **Performance**: Simple spaces are fastest, styled strings are more expensive

## Implementation Details

### Theme Configuration Structure
```go
type IndentationTheme struct {
    Name         string         // Display name
    IndentString string         // Custom string per level ("", "··", "- ", etc.)
    IndentSize   int           // Spaces per level (when IndentString is "")
    Style        lipgloss.Style // Styling for indentation
    Description  string        // User-friendly description
}
```

### Real-Time Configuration Updates
The example demonstrates live configuration updates without restarting the component:

1. User presses 'i' to cycle themes
2. `applyIndentationTheme()` modifies the tree config
3. `app.tree.SetRenderConfig()` applies changes immediately
4. Next render uses new indentation style

## Choosing the Right Indentation

### Consider Your Use Case

**File Managers**: Boxed or minimal - clear hierarchy without visual clutter
**Code Editors**: Spacious or minimal - familiar to developers
**Documentation**: Dotted or dashed - clear structure indication
**Accessibility**: Spacious with high contrast - easier to follow

### User Preferences

**Experienced Users**: May prefer minimal indentation to save space
**New Users**: May benefit from visual markers that make structure obvious
**Accessibility Needs**: May require larger indentation or high contrast markers

### Content Characteristics

**Wide Content**: Use minimal indentation to preserve horizontal space
**Deep Trees**: Consider depth indicators or responsive indentation
**Dense Information**: Visual markers may help users navigate
**Simple Hierarchies**: Minimal spacing is often sufficient

## Learning Path

This example builds on:
- [Basic Tree](../basic-tree/) - Tree structure and navigation
- [Cascading Selection](../cascading-selection/) - Selection behavior
- [Tree Symbols](../tree-symbols/) - Symbol customization
- [Tree Styling](../tree-styling/) - Colors and visual themes

Next examples in the tree series:
- Connected Lines - Box-drawing tree connectors
- Tree Enumerators - Bullet points and numbering
- Advanced Features - Auto-expand, expand all, etc.

## Performance Notes

### Indentation Performance Impact

**Space-Based** (Fastest):
- Minimal string operations
- No styling overhead
- Best for large trees

**String-Based** (Moderate):
- String repetition per level
- Styling operations if styled
- Still efficient for most uses

**Complex Styling** (Slower):
- Multiple lipgloss operations
- Consider for smaller trees or when performance isn't critical

### Optimization Tips
- **Cache styled strings** for repeated patterns
- **Use simple approaches** for very large datasets
- **Test with real data** to ensure performance meets needs

## Key Insights

### Visual Hierarchy Principles
**Indentation is about communication** - it tells users how items relate to each other in the hierarchy. The right indentation approach makes tree navigation intuitive and efficient.

### Balance Clarity and Space
**Too little indentation** makes hierarchy hard to follow  
**Too much indentation** wastes horizontal space  
**Just right** depends on your users and use case

### Consistency Matters
**Pick one approach** and use it throughout your application. Mixing indentation styles within the same interface confuses users.

### Context-Aware Design
**Different applications need different approaches** - a file manager has different needs than a documentation browser or code editor.

**Choose indentation that matches your users' mental model and supports their workflow!** 