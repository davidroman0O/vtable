# Connected Lines Example

This example demonstrates how to use **connected lines** - box-drawing characters that create visual connections between tree nodes. Experience the classic file manager appearance with connecting lines that clearly show parent-child relationships.

## What You'll Learn

- How to enable and configure box-drawing connectors
- Different connector styles and their visual impact
- Real-time connector theme switching
- Component-based connector configuration
- Terminal compatibility considerations
- Performance implications of different connector styles

## Running the Example

```bash
cd docs/04-tree-component/examples/connected-lines
go run main.go
```

## Controls

- **↑/↓ or j/k**: Navigate up/down the tree
- **Enter**: Expand/collapse folders
- **Space**: Select/deselect items
- **l**: Cycle through different connector styles ('l' for lines)
- **c**: Clear all selections
- **Page Up/Down**: Navigate faster through long lists
- **Home/End or g/G**: Jump to start/end
- **q or Ctrl+C**: Quit

## Connector Styles

Press **'l'** to cycle through these different connector approaches:

### 1. None (Simple Indentation)
```
📁 Web Application
  📁 src
    📄 main.go
    📄 app.go
    📁 handlers
      📄 user_handler.go
      📄 auth_handler.go
```
**No connectors - simple indentation** - Clean baseline for comparison.

### 2. Light Gray Connectors
```
📁 Web Application
├── 📁 src
│   ├── 📄 main.go
│   ├── 📄 app.go
│   └── 📁 handlers
│       ├── 📄 user_handler.go
│       └── 📄 auth_handler.go
├── 📁 tests
│   └── 📄 unit_test.go
└── 📁 config
    ├── 📄 .env
    └── 📄 config.yaml
```
**Subtle gray connectors** - Most widely used approach, balances clarity with visual noise.

### 3. Bold Gray Connectors
```
📁 Web Application
├── 📁 src
│   ├── 📄 main.go
│   ├── 📄 app.go
│   └── 📁 handlers
│       ├── 📄 user_handler.go
│       └── 📄 auth_handler.go
├── 📁 tests
│   └── 📄 unit_test.go
└── 📁 config
    ├── 📄 .env
    └── 📄 config.yaml
```
**Bold gray connectors for clarity** - Enhanced visibility for complex hierarchies.

### 4. Blue Lines (Themed)
```
📁 Web Application
├── 📁 src            (blue connectors)
│   ├── 📄 main.go
│   ├── 📄 app.go
│   └── 📁 handlers
│       ├── 📄 user_handler.go
│       └── 📄 auth_handler.go
├── 📁 tests
│   └── 📄 unit_test.go
└── 📁 config
    ├── 📄 .env
    └── 📄 config.yaml
```
**Blue connectors matching folder colors** - Coordinated with content theming.

### 5. Green Lines (File-Themed)
```
📁 Web Application
├── 📁 src            (green connectors)
│   ├── 📄 main.go
│   ├── 📄 app.go
│   └── 📁 handlers
│       ├── 📄 user_handler.go
│       └── 📄 auth_handler.go
├── 📁 tests
│   └── 📄 unit_test.go
└── 📁 config
    ├── 📄 .env
    └── 📄 config.yaml
```
**Green connectors matching file colors** - Alternative coordinated theming.

### 6. Dim Lines (Minimal)
```
📁 Web Application
├── 📁 src            (very dim connectors)
│   ├── 📄 main.go
│   ├── 📄 app.go
│   └── 📁 handlers
│       ├── 📄 user_handler.go
│       └── 📄 auth_handler.go
├── 📁 tests
│   └── 📄 unit_test.go
└── 📁 config
    ├── 📄 .env
    └── 📄 config.yaml
```
**Very subtle dim connectors** - Maximum content focus with minimal structure hints.

### 7. High Contrast (Accessibility)
```
📁 Web Application
├── 📁 src            (bright white connectors)
│   ├── 📄 main.go
│   ├── 📄 app.go
│   └── 📁 handlers
│       ├── 📄 user_handler.go
│       └── 📄 auth_handler.go
├── 📁 tests
│   └── 📄 unit_test.go
└── 📁 config
    ├── 📄 .env
    └── 📄 config.yaml
```
**High contrast white connectors** - Maximum visibility for accessibility needs.

### 8. Cyan Accent (Decorative)
```
📁 Web Application
├── 📁 src            (bright cyan connectors)
│   ├── 📄 main.go
│   ├── 📄 app.go
│   └── 📁 handlers
│       ├── 📄 user_handler.go
│       └── 📄 auth_handler.go
├── 📁 tests
│   └── 📄 unit_test.go
└── 📁 config
    ├── 📄 .env
    └── 📄 config.yaml
```
**Bright cyan accent connectors** - Eye-catching design for special interfaces.

## Key Features Demonstrated

### Component-Based Connector Control

The example shows how to use VTable's `TreeIndentationConfig` system for connectors:

```go
// Configure connectors through TreeIndentationConfig
treeConfig.IndentationConfig.UseConnectors = true
treeConfig.IndentationConfig.ConnectorStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("240"))  // Gray connectors

// Connectors require indentation to be enabled
treeConfig.IndentationConfig.Enabled = true
```

### Dynamic Connector Switching

Shows how to change connector styles at runtime:

```go
func (app *App) applyConnectorTheme() {
    theme := connectorThemes[app.currentConnector]
    treeConfig := app.tree.GetRenderConfig()
    
    // Apply connector settings
    treeConfig.IndentationConfig.UseConnectors = theme.UseConnectors
    treeConfig.IndentationConfig.ConnectorStyle = theme.ConnectorStyle
    
    app.tree.SetRenderConfig(treeConfig)
}
```

### Connector Theme Configuration

Demonstrates structured theme management:

```go
type ConnectorTheme struct {
    Name           string         // Display name for user
    ConnectorStyle lipgloss.Style // Styling for connector lines
    Description    string         // User-friendly description
    UseConnectors  bool          // Whether to show connectors
}
```

## Box-Drawing Characters Used

### Standard Light Box-Drawing
- **├** - T-junction (branch)
- **│** - Vertical line (continuation)
- **└** - L-corner (last item)
- **─** - Horizontal line (connector)

These characters are part of the Unicode Box Drawing block (U+2500-U+257F) and are widely supported across terminals.

### Visual Structure

```
📁 Root Project
├── 📁 First Child        (├── shows "has siblings below")
│   ├── 📄 Nested Item     (│   continues parent line)
│   └── 📄 Last Nested     (└── shows "last at this level")
├── 📁 Second Child        (├── shows "has siblings below")
│   └── 📄 Only Child      (└── shows "last at this level")
└── 📁 Last Child          (└── shows "last at root level")
    └── 📄 Final Item       (└── no continuation needed)
```

## Implementation Details

### Connector Requirements
Connectors require the indentation system to be enabled:

```go
// Enable indentation (required for connectors)
treeConfig.IndentationConfig.Enabled = true

// Then enable connectors
treeConfig.IndentationConfig.UseConnectors = true
```

### Content Coordination
The example shows connector-aware content formatting:

```go
func createConnectorAwareFormatter() {
    // Content styling that works well with connectors
    if flatItem.Item.IsFolder {
        return lipgloss.NewStyle().
            Foreground(lipgloss.Color("12")).  // Blue folders
            Bold(true).
            Render(content)
    } else {
        return lipgloss.NewStyle().
            Foreground(lipgloss.Color("10")).  // Green files
            Render(content)
    }
}
```

### Real-Time Configuration Updates
Demonstrates live configuration changes without component restart:

1. User presses 'l' to cycle connector styles
2. `applyConnectorTheme()` modifies the tree config
3. `app.tree.SetRenderConfig()` applies changes immediately
4. Next render uses new connector style

## Visual Impact Analysis

### Structure Clarity
**With Connectors**: Parent-child relationships are immediately obvious
**Without Connectors**: Users must infer relationships from indentation

### Horizontal Space Usage
**Light Gray**: Minimal impact on content width
**Bold Connectors**: Slightly more prominent but same width
**No Connectors**: Saves 2-3 characters per level

### Terminal Compatibility
**High Compatibility**: Light box-drawing characters
**Medium Compatibility**: Bold/styled connectors
**Universal**: No connectors (fallback option)

## Choosing Connector Style

### Use Cases by Style

**Light Gray (Default)**:
- General-purpose applications
- Professional interfaces
- Good balance of clarity and subtlety

**Bold Gray (Enhanced)**:
- Complex file hierarchies
- Code project explorers
- When structure is critical

**Themed Colors (Coordinated)**:
- Branded applications
- Consistent design systems
- Visual harmony with content

**High Contrast (Accessibility)**:
- Accessibility requirements
- High-contrast terminals
- Users with vision needs

**No Connectors (Minimal)**:
- Space-constrained interfaces
- Simple hierarchies
- Maximum content focus

### Performance Considerations

**Connector Overhead**: Minimal - single Unicode characters
**Styling Cost**: Normal lipgloss rendering overhead
**Terminal Impact**: Some terminals render box-drawing slower

### Best Practices

**Visual Hierarchy**:
- Connectors should guide, not dominate
- Maintain consistency throughout the tree
- Test in target terminal environments

**Responsive Design**:
- Consider disabling on very narrow displays
- Provide fallback for unsupported terminals
- Allow user preference settings

## Learning Path

This example builds on:
- [Basic Tree](../basic-tree/) - Tree structure and navigation
- [Cascading Selection](../cascading-selection/) - Selection behavior
- [Tree Symbols](../tree-symbols/) - Symbol customization
- [Tree Styling](../tree-styling/) - Colors and visual themes
- [Tree Indentation](../tree-indentation/) - Custom indentation

Next examples in the tree series:
- Tree Enumerators - Bullet points, numbers, custom prefixes
- Advanced Features - Auto-expand, expand all, collapse all

## Terminal Compatibility Notes

### High Compatibility
Most modern terminals support basic box-drawing characters:
- Terminal.app (macOS)
- iTerm2
- Windows Terminal
- GNOME Terminal
- VS Code Terminal

### Potential Issues
- Some older terminals may not support box-drawing
- Font rendering can affect appearance
- Copy/paste may not preserve characters correctly

### Fallback Strategy
Always provide a no-connector option for maximum compatibility:

```go
{
    Name:           "None",
    UseConnectors:  false,
    Description:    "No connectors - simple indentation",
}
```

## Key Insights

### Visual Communication
**Connected lines transform indentation into visual hierarchy** - they make the tree structure immediately apparent and help users understand complex relationships.

### Progressive Enhancement
**Connectors enhance rather than replace indentation** - they build on the spatial relationships established by indentation.

### User Choice
**Different users prefer different visual approaches** - providing options respects user preferences and accessibility needs.

### Context Matters
**The right connector style depends on your application** - file managers benefit from clear structure, while content-focused apps may prefer subtlety.

**Choose connectors that serve your users' primary task!** 