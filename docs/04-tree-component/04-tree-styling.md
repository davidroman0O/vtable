# Tree Styling

## What We're Adding

Taking our multi-project tree from previous examples, we're adding **visual styling** - colors, backgrounds, fonts, and themes that make your tree components look exactly how you want. Transform a basic functional tree into a polished, visually appealing interface.

## Understanding Tree Styling

Tree styling involves multiple visual layers that work together:

```
Basic tree:           Styled tree:           Themed tree:
📁 Folder            📁 Folder              📁 Folder
  📄 file.txt         📄 file.txt           📄 file.txt
                     (blue/green colors)    (dark theme + icons)
```

You can style:
- **Content colors** (folder vs file colors)
- **Background highlights** (selection, cursor, hover)
- **Font styles** (bold, italic, underline)
- **Symbol colors** (expand/collapse indicators)
- **Cursor appearance** (full-row vs content-only highlighting)
- **Border and spacing** (visual separation)

## Step 1: Start with the Working Tree

We'll use the same structure from previous examples, including custom symbols:

```go
// Same FileItem and data structure as before
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

**Building on previous work**: Styling enhances what you already have - it doesn't change your data or navigation logic.

## Step 2: Basic Content Styling

Let's start by coloring folders and files differently. We'll enhance our custom formatter:

```go
import "github.com/charmbracelet/lipgloss"

func styledTreeFormatter(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
    // Extract the FileItem from the data
    if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
        content := flatItem.Item.String()
        
        // Apply selection styling first (highest priority)
        if item.Selected {
            return lipgloss.NewStyle().
                Background(lipgloss.Color("12")). // Blue background
                Foreground(lipgloss.Color("15")). // White text
                Bold(true).
                Render(content)
        }
        
        // Apply content-based styling
        if flatItem.Item.IsFolder {
            // Style folders with blue color and bold text
            return lipgloss.NewStyle().
                Foreground(lipgloss.Color("12")). // Blue for folders
                Bold(true).
                Render(content)
        } else {
            // Style files with green color
            return lipgloss.NewStyle().
                Foreground(lipgloss.Color("10")). // Green for files
                Render(content)
        }
    }
    
    return fmt.Sprintf("%v", item.Item)
}
```

**Color hierarchy**: Selection styling takes priority, then content-specific styling applies.

## Step 3: Apply the Styled Formatter

```go
func main() {
    // Same data source and list config
    dataSource := NewFileTreeDataSource()
    listConfig := core.ListConfig{
        ViewportConfig: core.ViewportConfig{
            Height:    10,
            ChunkSize: 20,
        },
        SelectionMode: core.SelectionMultiple,
        KeyMap:        core.DefaultNavigationKeyMap(),
    }
    
    // Configure tree with styled formatter
    treeConfig := tree.DefaultTreeConfig()
    treeConfig.RenderConfig.ContentConfig.Formatter = styledTreeFormatter
}
```

Now your tree will show:
- **Blue bold folders** (📁 Web Application)
- **Green files** (📄 main.go)
- **Blue background with white text** for selected items

## Step 4: Style Tree Symbols

Let's add color to the expand/collapse symbols:

```go
// Style the tree symbols to match content
treeConfig.RenderConfig.TreeSymbolConfig.Style = lipgloss.NewStyle().
    Foreground(lipgloss.Color("240")). // Gray symbols
    Bold(true)

// Use custom symbols that look good with styling
treeConfig.RenderConfig.TreeSymbolConfig.ExpandedSymbol = "▼"
treeConfig.RenderConfig.TreeSymbolConfig.CollapsedSymbol = "▶"
treeConfig.RenderConfig.TreeSymbolConfig.LeafSymbol = "•"
treeConfig.RenderConfig.TreeSymbolConfig.SymbolSpacing = " "
```

**Visual consistency**: Gray symbols don't compete with the colorful content but still provide clear interaction cues.

## Step 5: Enhanced Cursor Styling

Let's explore both cursor styling approaches:

### Content-Only Cursor (Default)

This applies styling only to the text content:

```go
// Content-only cursor styling
treeConfig.RenderConfig.BackgroundConfig.Enabled = true
treeConfig.RenderConfig.BackgroundConfig.Style = lipgloss.NewStyle().
    Background(lipgloss.Color("236")). // Dark gray background
    Foreground(lipgloss.Color("15")).  // White text
    Bold(true)

// This will only highlight the actual content text
```

**Result**: Only the "📄 main.go" text gets the background, not the entire row.

### Full-Row Cursor

To achieve full-row highlighting, we use VTable's component-based rendering system with the `TreeBackgroundComponent`:

```go
// Configure TreeBackgroundComponent for full-row cursor highlighting
treeConfig.BackgroundConfig.Enabled = true
treeConfig.BackgroundConfig.Mode = tree.TreeBackgroundEntireLine
treeConfig.BackgroundConfig.Style = lipgloss.NewStyle().
    Background(lipgloss.Color("240")).
    Foreground(lipgloss.Color("15"))

// Content formatter focuses only on content styling - no cursor handling
func createFullRowFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
    return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
        if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
            content := flatItem.Item.String()
            
            // Selection styling (highest priority)
            if item.Selected {
                return lipgloss.NewStyle().
                    Background(lipgloss.Color("12")).
                    Foreground(lipgloss.Color("15")).
                    Bold(true).
                    Render(content)
            }
            
            // Content-based styling only - TreeBackgroundComponent handles cursor
            if flatItem.Item.IsFolder {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("12")).
                    Bold(true).
                    Render(content)
            } else {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("10")).
                    Render(content)
            }
        }
        
        return fmt.Sprintf("%v", item.Item)
    }
}
```

**Key advantages of component-based approach**:
- **Separation of concerns** - Content formatting separate from cursor styling
- **Full rendering pipeline** - Cursor, indentation, symbols, content all coordinated
- **Proper width calculation** - TreeBackgroundComponent handles full row width automatically
- **Consistent behavior** - Works with all tree components (symbols, indentation, etc.)

### Background Modes

VTable's `TreeBackgroundComponent` supports different modes:

```go
// Content-only cursor (traditional)
treeConfig.BackgroundConfig.Mode = tree.TreeBackgroundContentOnly

// Full-row cursor (maximum visibility)
treeConfig.BackgroundConfig.Mode = tree.TreeBackgroundEntireLine

// Selective component highlighting
treeConfig.BackgroundConfig.Mode = tree.TreeBackgroundSelectiveComponents
treeConfig.BackgroundConfig.ApplyToComponents = []tree.TreeComponentType{
    tree.TreeComponentCursor,
    tree.TreeComponentContent,
}

// Cursor indicator only
treeConfig.BackgroundConfig.Mode = tree.TreeBackgroundIndicatorOnly
```

### Hybrid Approach: Selective Components

For fine-grained control, use selective component highlighting:

```go
// Apply background to specific components only
treeConfig.BackgroundConfig.Mode = tree.TreeBackgroundSelectiveComponents
treeConfig.BackgroundConfig.ApplyToComponents = []tree.TreeComponentType{
    tree.TreeComponentCursor,      // Highlight cursor indicator
    tree.TreeComponentIndentation, // Highlight indentation
    tree.TreeComponentContent,     // Highlight content
    // tree.TreeComponentTreeSymbol - excluded, so symbols remain unaffected
}
```

### Advanced Cursor Styling

You can create context-aware cursor behavior by combining different approaches:

```go
func createDynamicCursorFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
    return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
        if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
            content := flatItem.Item.String()
            
            // Selection always gets full treatment
            if item.Selected {
                return lipgloss.NewStyle().
                    Background(lipgloss.Color("12")).
                    Foreground(lipgloss.Color("15")).
                    Bold(true).
                    Render(content)
            }
            
            // For dynamic cursor behavior:
            // - Folders: Let TreeBackgroundComponent handle full-row highlighting
            // - Files: Apply content-only highlighting directly
            if isCursor {
                if flatItem.Item.IsFolder {
                    // Return content without background - TreeBackgroundComponent will apply full-row
                    return lipgloss.NewStyle().
                        Foreground(lipgloss.Color("15")).
                        Bold(true).
                        Render(content)
                } else {
                    // Apply content-only background directly for files
                    return lipgloss.NewStyle().
                        Background(lipgloss.Color("240")).
                        Foreground(lipgloss.Color("15")).
                        Bold(true).
                        Render(content)
                }
            }
            
            // Regular content styling
            if flatItem.Item.IsFolder {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("12")).
                    Bold(true).
                    Render(content)
            } else {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("10")).
                    Render(content)
            }
        }
        
        return fmt.Sprintf("%v", item.Item)
    }
}
```

### Cursor Style Guidelines

**Component-Based Approach (Recommended)**:
- **Use TreeBackgroundComponent** for cursor highlighting
- **Keep content formatters focused** on content styling only
- **Let the rendering pipeline** handle cursor coordination

**Legacy Formatter Approach**:
- **Only when necessary** for very specific custom behavior
- **More complex** and harder to maintain
- **Can conflict** with other tree components

### Component-Based vs Formatter-Based

**Component-Based (Best Practice)**:
```go
// ✅ Clean separation - content formatter handles content only
treeConfig.ContentConfig.Formatter = contentOnlyFormatter
treeConfig.BackgroundConfig.Enabled = true
treeConfig.BackgroundConfig.Mode = tree.TreeBackgroundEntireLine
treeConfig.BackgroundConfig.Style = cursorStyle
```

**Formatter-Based (Legacy)**:
```go
// ❌ Mixed concerns - formatter handles both content and cursor
if isCursor {
    return lipgloss.NewStyle().Width(ctx.MaxWidth).Render(content)
}
```

## Step 6: Create a Dark Theme

Let's create a complete dark theme for the tree:

```go
func createDarkThemeFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
    return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
        if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
            content := flatItem.Item.String()
            
            // Selection styling (bright highlight)
            if item.Selected {
                return lipgloss.NewStyle().
                    Background(lipgloss.Color("33")). // Bright blue
                    Foreground(lipgloss.Color("0")).  // Black text
                    Bold(true).
                    Render(content)
            }
            
            // Dark theme colors
            if flatItem.Item.IsFolder {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("39")). // Bright cyan for folders
                    Bold(true).
                    Render(content)
            } else {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("46")). // Bright green for files
                    Render(content)
            }
        }
        
        return fmt.Sprintf("%v", item.Item)
    }
}
```

Apply the dark theme:

```go
// Apply dark theme
treeConfig.RenderConfig.ContentConfig.Formatter = createDarkThemeFormatter()

// Dark theme symbols
treeConfig.RenderConfig.TreeSymbolConfig.Style = lipgloss.NewStyle().
    Foreground(lipgloss.Color("244")). // Light gray symbols
    Bold(true)

// Dark theme cursor
treeConfig.RenderConfig.BackgroundConfig.Style = lipgloss.NewStyle().
    Background(lipgloss.Color("235")). // Very dark gray
    Foreground(lipgloss.Color("15")).
    Bold(true)
```

## Step 7: Create a Professional Theme

For business applications, create a more subdued, professional theme:

```go
func createProfessionalThemeFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
    return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
        if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
            content := flatItem.Item.String()
            
            // Professional selection styling
            if item.Selected {
                return lipgloss.NewStyle().
                    Background(lipgloss.Color("153")). // Soft blue
                    Foreground(lipgloss.Color("0")).   // Black text
                    Render(content)
            }
            
            // Professional colors
            if flatItem.Item.IsFolder {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("24")). // Dark blue for folders
                    Bold(true).
                    Render(content)
            } else {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("240")). // Dark gray for files
                    Render(content)
            }
        }
        
        return fmt.Sprintf("%v", item.Item)
    }
}
```

**Professional look**: Muted colors that work well in business interfaces without being distracting.

## Step 8: Add Font Styling Variations

Experiment with different font styles:

```go
func createFontStyledFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
    return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
        if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
            content := flatItem.Item.String()
            
            if item.Selected {
                return lipgloss.NewStyle().
                    Background(lipgloss.Color("12")).
                    Foreground(lipgloss.Color("15")).
                    Bold(true).
                    Underline(true). // Add underline for emphasis
                    Render(content)
            }
            
            if flatItem.Item.IsFolder {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("12")).
                    Bold(true).
                    Italic(true). // Italic folders
                    Render(content)
            } else {
                return lipgloss.NewStyle().
                    Foreground(lipgloss.Color("10")).
                    Render(content)
            }
        }
        
        return fmt.Sprintf("%v", item.Item)
    }
}
```

**Typography variety**: Bold, italic, and underline styles add visual hierarchy.

## Step 9: Create Theme Switching

Let's build an app that can switch between different themes and cursor styles:

```go
type ThemeStyle struct {
    Name        string
    Formatter   func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string
    SymbolStyle lipgloss.Style
    CursorStyle lipgloss.Style
    Description string
    CursorType  string // "content-only", "full-row", "dynamic"
}

var themes = []ThemeStyle{
    {
        Name:        "Default",
        Formatter:   createDefaultFormatter(),
        SymbolStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
        CursorStyle: lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("15")),
        Description: "Clean blue and green theme",
        CursorType:  "content-only",
    },
    {
        Name:        "Full-Row",
        Formatter:   createFullRowFormatter(),
        SymbolStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
        CursorStyle: lipgloss.NewStyle(), // Handled in formatter
        Description: "Full-row cursor highlighting",
        CursorType:  "full-row",
    },
    {
        Name:        "Dynamic",
        Formatter:   createDynamicCursorFormatter(),
        SymbolStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
        CursorStyle: lipgloss.NewStyle(), // Handled in formatter
        Description: "Different cursor styles by content type",
        CursorType:  "dynamic",
    },
    {
        Name:        "Dark",
        Formatter:   createDarkThemeFormatter(),
        SymbolStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Bold(true),
        CursorStyle: lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color("15")).Bold(true),
        Description: "High contrast dark theme",
        CursorType:  "content-only",
    },
    {
        Name:        "Professional",
        Formatter:   createProfessionalThemeFormatter(),
        SymbolStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
        CursorStyle: lipgloss.NewStyle().Background(lipgloss.Color("250")).Foreground(lipgloss.Color("0")),
        Description: "Subdued business theme",
        CursorType:  "content-only",
    },
}
```

## Step 10: Enhanced Theme Switching Logic

```go
func (app *App) applyTheme() {
    theme := themes[app.currentTheme]
    
    // Get current config
    treeConfig := app.tree.GetRenderConfig()
    
    // Apply theme
    treeConfig.ContentConfig.Formatter = theme.Formatter
    treeConfig.TreeSymbolConfig.Style = theme.SymbolStyle
    
    // Apply cursor styling based on type
    if theme.CursorType == "content-only" {
        treeConfig.BackgroundConfig.Enabled = true
        treeConfig.BackgroundConfig.Style = theme.CursorStyle
    } else {
        // For full-row and dynamic cursors, disable background config
        // since the formatter handles cursor styling
        treeConfig.BackgroundConfig.Enabled = false
    }
    
    // Apply the updated config
    app.tree.SetRenderConfig(treeConfig)
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "t":
            // Cycle through themes
            app.currentTheme = (app.currentTheme + 1) % len(themes)
            app.applyTheme()
            theme := themes[app.currentTheme]
            app.status = fmt.Sprintf("Theme: %s (%s) - %s", 
                theme.Name, theme.CursorType, theme.Description)
            return app, nil
        }
    }
    
    // ... rest of update logic
}
```

## Step 11: Complete Styled Example

```go
func main() {
    // Create data source
    dataSource := NewFileTreeDataSource()
    
    // Configure list
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
    
    // Apply initial styling
    treeConfig.RenderConfig.ContentConfig.Formatter = createDefaultFormatter()
    treeConfig.RenderConfig.TreeSymbolConfig.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
    treeConfig.RenderConfig.BackgroundConfig.Enabled = true
    treeConfig.RenderConfig.BackgroundConfig.Style = lipgloss.NewStyle().
        Background(lipgloss.Color("240")).
        Foreground(lipgloss.Color("15"))
    
    // Create tree
    treeComponent := tree.NewTreeList(listConfig, treeConfig, dataSource)
    
    // Create app
    app := &App{
        tree:         treeComponent,
        status:       "Ready! Press 't' to cycle through visual themes",
        currentTheme: 0,
        dataSource:   dataSource,
    }
    
    // Apply initial theme
    app.applyTheme()
    
    // Run
    p := tea.NewProgram(app)
    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## What You'll See

### Content-Only Cursor
```
🌳 Styled Tree Demo - Default Theme

► 📁 Web Application    (blue, bold)
  📁 CLI Tool           (blue, bold)
> 📄 main.go            (highlighted content only)
  📁 API Service        (blue, bold)
```

### Full-Row Cursor
```
🌳 Styled Tree Demo - Full-Row Theme

► 📁 Web Application    (blue, bold)
  📁 CLI Tool           (blue, bold)
██████████████████████████████████████ (entire row highlighted)
> 📄 main.go            
██████████████████████████████████████
  📁 API Service        (blue, bold)
```

### Dynamic Cursor
```
🌳 Styled Tree Demo - Dynamic Theme

► 📁 Web Application    (blue, bold)
██████████████████████████████████████ (full-row for folders)
> 📁 CLI Tool           
██████████████████████████████████████
    📄 main.go          (content-only for files)
  📁 API Service        (blue, bold)
```

## Cursor Styling Best Practices

### Visual Clarity
**Do:**
- **Ensure cursor is always visible** - users must know where they are
- **Make cursor distinct from selection** - different colors/styles
- **Test with your content** - some styles work better with certain data
- **Consider terminal variety** - test in different terminal applications

**Don't:**
- **Make cursor too subtle** - visibility is more important than aesthetics
- **Use same color for cursor and selection** - confusing for users
- **Forget about accessibility** - some users need high contrast
- **Ignore content width** - full-row cursors need proper width handling

### Choosing Cursor Style

**Use Content-Only When:**
- Interface is clean and minimal
- Content has lots of colors/styling
- Users are experienced with the interface
- Dense information display

**Use Full-Row When:**
- Maximum visibility is crucial
- Users may have accessibility needs
- Professional/business applications
- File manager or IDE-like interfaces

**Use Dynamic When:**
- Different content types have different importance
- Complex hierarchies
- Advanced user interfaces
- You want to optimize for both cases

### Technical Implementation

**Content-Only**: 
```go
// Simple - use BackgroundConfig
treeConfig.BackgroundConfig.Enabled = true
treeConfig.BackgroundConfig.Style = cursorStyle
```

**Full-Row**:
```go
// Complex - handle in formatter
if isCursor {
    return style.Width(ctx.Width).Align(lipgloss.Left).Render(content)
}
```

**Best Practice**: Always provide both options and let users choose their preference.

## Key Concepts

### 1. **Layered Styling**
Styles are applied in priority order - selection, cursor, content type, default.

### 2. **Consistent Visual Language**
All elements should work together to create a cohesive appearance.

### 3. **Context-Aware Styling**
Different item types (folders vs files) and states (selected vs unselected) should have appropriate visual treatment.

### 4. **Cursor Visibility**
The cursor style should match your application's needs - content-only for minimal interfaces, full-row for maximum visibility.

### 5. **Theme Flexibility**
Good styling systems allow easy theme switching without code changes.

## Try It Yourself

1. **Create a custom theme** with your favorite colors
2. **Experiment with cursor styles** - try content-only vs full-row highlighting
3. **Test different combinations** - mix cursor styles with different content themes
4. **Consider your users** - what cursor style works best for your application?
5. **Test accessibility** - ensure your cursors are visible to users with different needs

## What's Next

You now understand how to style tree components with different cursor approaches! Next, we'll explore tree indentation - controlling the spacing and alignment that shows the hierarchical structure.

The insight: **Good cursor styling balances visibility with aesthetics** - your users need to know where they are, but the cursor shouldn't dominate the interface. Choose the approach that matches your application's purpose and user needs. 