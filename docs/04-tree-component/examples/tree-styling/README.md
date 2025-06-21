# Tree Styling Example

This example demonstrates how to apply visual styling and themes to VTable trees - colors, fonts, backgrounds, and complete visual makeovers that transform the appearance while keeping the same functionality. It also showcases different cursor highlighting approaches.

## What You'll Learn

- How to create custom formatters for different visual themes
- Color schemes for folders vs files
- Background styling for selection and cursor states  
- Font styling (bold, italic, underline)
- **Cursor styling options**: content-only vs full-row highlighting
- Building complete visual themes
- Dynamic theme switching at runtime

## Running the Example

```bash
cd docs/04-tree-component/examples/tree-styling
go run main.go
```

## Controls

- **↑/↓ or j/k**: Navigate up/down the tree
- **Enter**: Expand/collapse folders
- **Space**: Select/deselect items (theme-specific highlighting)
- **t**: Cycle through different visual themes and cursor styles
- **c**: Clear all selections
- **Page Up/Down**: Navigate faster through long lists
- **Home/End or g/G**: Jump to start/end
- **q or Ctrl+C**: Quit

## Visual Themes & Cursor Styles

Press **'t'** to cycle through these different themes:

### 1. Default Theme (Content-Only Cursor)
```
📁 Web Application    (blue, bold)
  📁 src              (blue, bold)
> 📄 main.go          (green, highlighted text only)
```
**Clean blue and green theme** - Content-only cursor highlighting for minimal appearance.

### 2. Full-Row Theme (Full-Row Cursor)
```
📁 Web Application    (blue, bold)
  📁 src              (blue, bold)
████████████████████████████████████
> 📄 main.go          (green)
████████████████████████████████████
```
**Full-row cursor highlighting** - Maximum visibility with entire row highlighted.

### 3. Dynamic Theme (Hybrid Cursor)
```
📁 Web Application    (blue, bold)
████████████████████████████████████ (full-row for folders)
> 📁 src              (blue, bold)
████████████████████████████████████
    📄 main.go        (green, content-only for files)
```
**Dynamic cursor styling** - Full-row highlighting for folders, content-only for files.

### 4. Dark Theme (Content-Only Cursor)
```
📁 Web Application    (bright cyan, bold)
  📁 src              (bright cyan, bold)  
> 📄 main.go          (bright green, highlighted text only)
```
**High contrast dark theme** - Bright colors on dark background.

### 5. Professional Theme (Content-Only Cursor)
```
📁 Web Application    (dark blue, bold)
  📁 src              (dark blue, bold)
> 📄 main.go          (dark gray, highlighted text only)
```
**Subdued business theme** - Muted colors perfect for corporate environments.

### 6. High Contrast Theme (Content-Only Cursor)
```
📁 Web Application    (black, bold)
  📁 src              (black, bold)
> 📄 main.go          (dark gray, highlighted text only)
```
**Maximum accessibility contrast** - Strong contrast for better accessibility.

### 7. Colorful Theme (Content-Only Cursor)
```
📁 Web Application    (orange, bold, italic)
  📁 src              (orange, bold, italic)
> 📄 main.go          (bright green, highlighted text only)
```
**Vibrant colors and styling** - Fun, colorful theme with multiple font styles.

## Cursor Styling Options

### Content-Only Cursor
- **Visual approach**: Only the text content gets background highlighting
- **Best for**: Clean, minimal interfaces; dense content; colorful themes
- **Implementation**: Uses the standard background configuration

### Full-Row Cursor  
- **Visual approach**: Entire row width is highlighted
- **Best for**: Maximum visibility; accessibility; professional applications
- **Implementation**: Handled in the formatter with width styling

### Dynamic Cursor
- **Visual approach**: Different cursor styles based on content type
- **Best for**: Complex hierarchies where different items need different treatment
- **Implementation**: Conditional logic in formatter (folders get full-row, files get content-only)

## Key Features Demonstrated

### Enhanced Theme Architecture
Each theme now includes cursor type and background mode information:

```go
type ThemeStyle struct {
    Name            string                    // Display name
    Formatter       func(...)string          // Content formatting function
    SymbolStyle     lipgloss.Style           // Tree symbol styling
    CursorStyle     lipgloss.Style           // Cursor background styling
    Description     string                   // Theme description
    CursorType      string                   // "content-only", "full-row", "dynamic"
    BackgroundMode  tree.TreeBackgroundMode  // How to apply cursor background
}
```

### Component-Based Cursor Implementation
The example demonstrates the proper way to implement cursor styling using VTable's component-based rendering system:

**Content-Only Cursor (Simple)**:
```go
// Uses TreeBackgroundComponent with content-only mode
treeConfig.BackgroundConfig.Enabled = true
treeConfig.BackgroundConfig.Mode = tree.TreeBackgroundContentOnly
treeConfig.BackgroundConfig.Style = cursorStyle
```

**Full-Row Cursor (Proper)**:
```go
// Uses TreeBackgroundComponent with entire-line mode  
treeConfig.BackgroundConfig.Enabled = true
treeConfig.BackgroundConfig.Mode = tree.TreeBackgroundEntireLine
treeConfig.BackgroundConfig.Style = cursorStyle

// Content formatter focuses only on content - no cursor handling
func createContentOnlyFormatter() func(...) string {
    return func(...) string {
        // Only handle content styling - TreeBackgroundComponent handles cursor
        if flatItem.Item.IsFolder {
            return folderStyle.Render(content)
        } else {
            return fileStyle.Render(content)
        }
    }
}
```

**Dynamic Cursor (Context-Aware)**:
```go
// Uses selective component highlighting
treeConfig.BackgroundConfig.Mode = tree.TreeBackgroundSelectiveComponents
treeConfig.BackgroundConfig.ApplyToComponents = []tree.TreeComponentType{
    tree.TreeComponentCursor,
    tree.TreeComponentIndentation,
    tree.TreeComponentContent,
}

// Formatter handles different behavior for different content types
func createDynamicFormatter() func(...) string {
    return func(...) string {
        if isCursor {
            if flatItem.Item.IsFolder {
                // Return content without background - TreeBackgroundComponent applies full-row
                return folderCursorStyle.Render(content)
            } else {
                // Apply background directly for files (content-only)
                return fileCursorStyle.Background(color).Render(content)
            }
        }
        // Regular content styling...
    }
}
```

### Dynamic Theme Switching
Shows how to properly switch between different cursor approaches using the component system:

```go
func (app *App) applyTheme() {
    theme := themes[app.currentTheme]
    treeConfig := app.tree.GetRenderConfig()
    
    // Apply content formatter (focused on content only)
    treeConfig.ContentConfig.Formatter = theme.Formatter
    
    // Apply cursor styling using TreeBackgroundComponent
    treeConfig.BackgroundConfig.Enabled = true
    treeConfig.BackgroundConfig.Style = theme.CursorStyle
    treeConfig.BackgroundConfig.Mode = theme.BackgroundMode
    
    // For dynamic cursors, specify which components get background
    if theme.CursorType == "dynamic" {
        treeConfig.BackgroundConfig.ApplyToComponents = []tree.TreeComponentType{
            tree.TreeComponentCursor,
            tree.TreeComponentIndentation,
            tree.TreeComponentTreeSymbol,
            tree.TreeComponentContent,
        }
    }
    
    app.tree.SetRenderConfig(treeConfig)
}
```

## Visual Elements

- **Content Colors**: Different colors for folders vs files
- **Font Styles**: Bold, italic, underline combinations  
- **Cursor Highlighting**: Content-only, full-row, or dynamic approaches
- **Background Highlights**: Selection and cursor visibility
- **Symbol Coordination**: Tree symbols that match each theme
- **Contrast Management**: Readable combinations in all themes

## Cursor Styling Best Practices

### Choosing the Right Cursor Style

**Content-Only Cursors**:
- ✅ **Use when**: Clean, minimal interfaces
- ✅ **Good for**: Dense content, colorful themes, experienced users
- ✅ **Benefits**: Less visual noise, doesn't compete with content styling

**Full-Row Cursors**:
- ✅ **Use when**: Maximum visibility is crucial
- ✅ **Good for**: Professional tools, accessibility needs, file managers
- ✅ **Benefits**: Impossible to miss, works well in all lighting conditions

**Dynamic Cursors**:
- ✅ **Use when**: Different content types have different importance levels
- ✅ **Good for**: Complex hierarchies, advanced interfaces
- ✅ **Benefits**: Optimized visibility per content type

### Implementation Guidelines

1. **Visual Hierarchy**: Selection > Cursor > Content Type > Default
2. **Accessibility**: Ensure sufficient contrast in all cursor styles
3. **Consistency**: Match cursor style to your application's overall design
4. **User Choice**: Consider allowing users to select their preferred cursor style
5. **Context Matters**: Test cursor visibility with your actual content

## What to Try

1. **Compare cursor styles** - Notice how different approaches affect navigation
2. **Test with different content** - See how cursor visibility changes with various tree structures
3. **Accessibility testing** - Try the high contrast theme and full-row cursors
4. **Selection behavior** - Observe how selection highlighting works with different cursor styles
5. **Performance impact** - Notice any difference between simple and complex cursor implementations

## Learning Path

This example builds on:
- [Basic Tree](../basic-tree/) - Tree structure and navigation
- [Cascading Selection](../cascading-selection/) - Selection behavior
- [Tree Symbols](../tree-symbols/) - Symbol customization

Next examples in the tree series:
- Tree Indentation - Custom spacing and alignment
- Connected Lines - Box-drawing tree connectors
- Tree Enumerators - Bullet points and numbering

## Key Insights

### Component-Based vs Formatter-Based Approach

This example demonstrates the **proper component-based approach** for cursor styling, which is superior to handling cursor styling in formatters:

**Component-Based Approach (Recommended)**:
✅ **Separation of concerns** - Content formatters focus only on content styling  
✅ **Automatic coordination** - TreeBackgroundComponent handles width calculation and component coordination  
✅ **Full rendering pipeline** - Works seamlessly with all tree components (cursor, indentation, symbols, content)  
✅ **Consistent behavior** - Maintains expected VTable rendering patterns  
✅ **Easier maintenance** - Configuration-based rather than code-based  

**Formatter-Based Approach (Legacy)**:  
❌ **Mixed concerns** - Formatters handle both content and cursor styling  
❌ **Manual width handling** - Must calculate and manage full-row width manually  
❌ **Component conflicts** - Can interfere with other tree components  
❌ **Complex implementation** - Requires custom logic in every formatter  
❌ **Inconsistent behavior** - May not work properly with all tree features  

### Component-Based Architecture

VTable's tree rendering uses a **component pipeline**:

```
TreeComponentOrder: [Cursor] → [Indentation] → [TreeSymbol] → [Content] → [Background]
```

Each component:
1. **Renders its part** of the tree item
2. **Passes data** to the next component in the pipeline  
3. **Coordinates with others** through the TreeComponentContext

The `TreeBackgroundComponent` operates as a **post-processing step** that can:
- Apply background to the entire line (`TreeBackgroundEntireLine`)
- Apply background to specific components (`TreeBackgroundSelectiveComponents`)
- Apply background to content only (`TreeBackgroundContentOnly`)

This architecture ensures that cursor highlighting **works correctly** with all tree features like indentation, symbols, and custom spacing.

### Cursor Visibility Balance
The best cursor approach balances **visibility** with **aesthetics**:
- Content-only cursors provide subtle navigation feedback
- Full-row cursors ensure maximum visibility
- Dynamic cursors optimize for content importance

### Implementation Complexity
- **Content-only**: Simple configuration-based approach
- **Full-row**: Uses TreeBackgroundComponent with EntireLine mode  
- **Dynamic**: Combines component configuration with conditional formatter logic

### User Experience Impact
Different cursor styles create different user experiences:
- **Minimal cursors**: Feel lightweight and unobtrusive
- **Full-row cursors**: Feel solid and professional
- **Dynamic cursors**: Feel smart and context-aware

**Choose the cursor style that matches your application's personality and your users' needs!**

## Theme Details

### Default Theme
- **Folders**: Blue (#12), bold
- **Files**: Green (#10)
- **Selection**: Blue background (#12), white text
- **Symbols**: Gray (#240)
- **Cursor**: Gray background (#240)

### Dark Theme  
- **Folders**: Bright cyan (#39), bold
- **Files**: Bright green (#46)
- **Selection**: Bright blue background (#33), black text
- **Symbols**: Light gray (#244), bold
- **Cursor**: Very dark gray background (#235)

### Professional Theme
- **Folders**: Dark blue (#24), bold
- **Files**: Dark gray (#240)
- **Selection**: Soft blue background (#153), black text
- **Symbols**: Gray (#240)
- **Cursor**: Light gray background (#250)

### High Contrast Theme
- **Folders**: Black (#0), bold
- **Files**: Dark gray (#8)
- **Selection**: Black background (#0), white text, underlined
- **Symbols**: Black (#0), bold
- **Cursor**: Light gray background (#7)

### Colorful Theme
- **Folders**: Orange (#208), bold, italic
- **Files**: Bright green (#82)
- **Selection**: Bright magenta background (#201), white text
- **Symbols**: Purple (#129), bold
- **Cursor**: Purple background (#93)

## Code Structure

The example demonstrates:

1. **Theme Definitions**: Array of complete theme configurations
2. **Formatter Functions**: One per theme, handling all styling logic
3. **Coordination**: Symbols and cursor styling that match content
4. **Runtime Switching**: Dynamic theme application without restart

## Styling Best Practices

### Color Guidelines
- **Maintain contrast**: Ensure text is readable in all themes
- **Be consistent**: Use the same color for the same content type
- **Consider accessibility**: Test themes with colorblind users
- **Match platform conventions**: Use familiar color schemes

### Theme Design
- **Start simple**: Begin with 2-3 colors, add complexity gradually
- **Test in context**: How themes look depends on terminal settings
- **Provide options**: Different users prefer different visual styles
- **Document choices**: Explain why certain colors were chosen

## Key Insights

### Visual Hierarchy
Good themes establish clear priority:
1. Selection (most important)
2. Cursor position  
3. Content type (folders vs files)
4. Symbols and decoration

### Cohesive Design
All elements should work together - content colors, symbol styling, and cursor highlights should feel like they belong to the same design system.

### Flexibility
The theme system allows complete visual transformation without changing any business logic or navigation behavior.

**Choose themes that match your application's purpose and your users' needs!** 