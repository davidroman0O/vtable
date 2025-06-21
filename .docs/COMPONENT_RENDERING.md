# Component-Based List Rendering System

The vtable List component now features a completely redesigned rendering system that breaks down item rendering into distinct, optional, and reorderable components. This specialized system is designed specifically for lists (tables and trees have their own rendering systems).

## 🎯 **Key Benefits**

- **Modular**: Each part of rendering (cursor, enumerator, content, spacing, background) is a separate component
- **Flexible**: Components can be enabled/disabled, reordered, and customized independently
- **Powerful**: Fine-grained control over every aspect of list item appearance
- **Maintainable**: Clear separation of concerns makes the code easier to understand and extend

## 🧩 **Component Types**

### 1. **Cursor Component** (`ListComponentCursor`)
Handles cursor indicator rendering for the currently selected item.

```go
config.CursorConfig = ListCursorConfig{
    Enabled:         true,
    CursorIndicator: "► ",
    NormalSpacing:   "  ",
    Style:           lipgloss.NewStyle().Foreground(lipgloss.Color("205")),
}
```

### 2. **Pre-Spacing Component** (`ListComponentPreSpacing`)
Adds spacing before the main content (rarely used but available for special layouts).

```go
config.PreSpacingConfig = ListSpacingConfig{
    Enabled: true,
    Spacing: "  ",
    Style:   lipgloss.NewStyle(),
}
```

### 3. **Enumerator Component** (`ListComponentEnumerator`)
Handles enumeration (bullets, numbers, checkboxes, etc.) with alignment support.

```go
config.EnumeratorConfig = ListEnumeratorConfig{
    Enabled:    true,
    Enumerator: vtable.BulletEnumerator,
    Style:      lipgloss.NewStyle(),
    Alignment:  ListAlignmentRight,
    MaxWidth:   4,
}
```

### 4. **Content Component** (`ListComponentContent`)
Renders the main item content with optional formatting and text wrapping.

```go
config.ContentConfig = ListContentConfig{
    Enabled:   true,
    Formatter: myCustomFormatter,
    Style:     lipgloss.NewStyle(),
    WrapText:  true,
    MaxWidth:  80,
}
```

### 5. **Post-Spacing Component** (`ListComponentPostSpacing`)
Adds spacing after the main content.

```go
config.PostSpacingConfig = ListSpacingConfig{
    Enabled: true,
    Spacing: " ",
    Style:   lipgloss.NewStyle(),
}
```

### 6. **Background Component** (`ListComponentBackground`)
Applies background styling as a post-process with multiple modes.

```go
config.BackgroundConfig = ListBackgroundConfig{
    Enabled:           true,
    Style:             lipgloss.NewStyle().Background(lipgloss.Color("240")),
    ApplyToComponents: []ListComponentType{ListComponentContent},
    Mode:              ListBackgroundContentOnly,
}
```

## 🎨 **Component Order**

The order of components is fully customizable:

```go
// Standard order
config.ComponentOrder = []ListComponentType{
    ListComponentCursor,
    ListComponentEnumerator,
    ListComponentContent,
}

// Custom order: Content first, then enumerator
config.ComponentOrder = []ListComponentType{
    ListComponentContent,
    ListComponentEnumerator,
    ListComponentCursor,
}

// Minimal: Just content
config.ComponentOrder = []ListComponentType{
    ListComponentContent,
}
```

## 🚀 **Quick Start**

### Using Preset Configurations

```go
// Bullet list
config := vtable.BulletListConfig()

// Numbered list
config := vtable.NumberedListConfig()

// Checklist
config := vtable.ChecklistConfig()

// Minimal (content only)
config := vtable.MinimalListConfig()

// Custom order
config := vtable.CustomOrderListConfig([]vtable.ListComponentType{
    vtable.ListComponentContent,
    vtable.ListComponentEnumerator,
})

// Background styled
config := vtable.BackgroundStyledListConfig(
    lipgloss.NewStyle().Background(lipgloss.Color("240")),
    vtable.ListBackgroundContentOnly,
)
```

### Creating a List

```go
listConfig := vtable.ListConfig{
    ViewportConfig: vtable.DefaultViewportConfig(),
    RenderConfig:   vtable.BulletListConfig(), // Use any preset
    SelectionMode:  vtable.SelectionSingle,
    KeyMap:         vtable.DefaultNavigationKeyMap(),
    MaxWidth:       80,
}

list := vtable.NewList(listConfig, dataSource)
```

## 🎛️ **Advanced Customization**

### Custom Component Configuration

```go
config := vtable.DefaultListRenderConfig()

// Customize cursor
config.CursorConfig.CursorIndicator = "→ "
config.CursorConfig.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))

// Customize enumerator
config.EnumeratorConfig.Enumerator = vtable.ArabicEnumerator
config.EnumeratorConfig.Alignment = vtable.ListAlignmentRight
config.EnumeratorConfig.MaxWidth = 4

// Customize content
config.ContentConfig.WrapText = true
config.ContentConfig.MaxWidth = 60

// Customize background
config.BackgroundConfig.Enabled = true
config.BackgroundConfig.Mode = vtable.ListBackgroundContentOnly
config.BackgroundConfig.Style = lipgloss.NewStyle().
    Background(lipgloss.Color("240")).
    Foreground(lipgloss.Color("15"))

// Set custom order
config.ComponentOrder = []vtable.ListComponentType{
    vtable.ListComponentCursor,
    vtable.ListComponentEnumerator,
    vtable.ListComponentContent,
}
```

### Runtime Configuration Changes

```go
// Get current config
config := list.GetRenderConfig()

// Modify it
config.EnumeratorConfig.Enumerator = vtable.CheckboxEnumerator
config.BackgroundConfig.Enabled = true

// Apply changes
list.SetRenderConfig(config)
```

## 🎯 **Background Styling Modes**

The background component supports multiple styling modes:

### `ListBackgroundEntireLine`
Applies background to the entire rendered line (cursor + enumerator + content).

### `ListBackgroundContentOnly`
Applies background only to the main content, preserving indicators/enumerators unstyled.

### `ListBackgroundIndicatorOnly`
Applies background only to the cursor indicator.

### `ListBackgroundSelectiveComponents`
Applies background only to specified components in `ApplyToComponents`.

```go
config.BackgroundConfig = ListBackgroundConfig{
    Enabled:           true,
    Style:             myBackgroundStyle,
    ApplyToComponents: []ListComponentType{
        ListComponentEnumerator,
        ListComponentContent,
    },
    Mode: ListBackgroundSelectiveComponents,
}
```

## 🔧 **Integration with Existing API**

The new system maintains backward compatibility through the existing API:

```go
// These still work and use the component system under the hood
list.SetBulletStyle()
list.SetNumberedStyle()
list.SetChecklistStyle()
list.SetEnumerator(myCustomEnumerator)
```

## 📊 **Performance**

The component system is designed for performance:

- **Lazy Evaluation**: Components only render when enabled
- **Efficient Caching**: Component data is cached during rendering
- **Memory Efficient**: No additional memory overhead
- **Fast Rendering**: Optimized for large lists

## 🎨 **Examples**

### Basic Usage

```go
// Create a numbered list with background styling
config := vtable.NumberedListConfig()
config.BackgroundConfig.Enabled = true
config.BackgroundConfig.Mode = vtable.ListBackgroundContentOnly
config.BackgroundConfig.Style = lipgloss.NewStyle().Background(lipgloss.Color("240"))

list := vtable.NewList(vtable.ListConfig{
    RenderConfig: config,
    // ... other config
}, dataSource)
```

### Custom Order Example

```go
// Content first, then enumerator, then cursor
config := vtable.CustomOrderListConfig([]vtable.ListComponentType{
    vtable.ListComponentContent,
    vtable.ListComponentEnumerator,
    vtable.ListComponentCursor,
})

// This would render as: "My content • ► " instead of "► • My content"
```

### Minimal List

```go
// Just content, no indicators
config := vtable.MinimalListConfig()
// Renders as: "My content" (no cursor, no enumerator)
```

## 🔄 **Migration from Old System**

The old `ListRenderConfig` structure has been completely replaced. If you were using the old system:

**Old:**
```go
config := vtable.ListRenderConfig{
    Enumerator:      vtable.BulletEnumerator,
    ShowEnumerator:  true,
    CursorIndicator: "► ",
    ShowCursor:      true,
    // ...
}
```

**New:**
```go
config := vtable.BulletListConfig() // Or build manually:

config := vtable.DefaultListRenderConfig()
config.EnumeratorConfig.Enumerator = vtable.BulletEnumerator
config.EnumeratorConfig.Enabled = true
config.CursorConfig.CursorIndicator = "► "
config.CursorConfig.Enabled = true
```

## 🎯 **Best Practices**

1. **Use Presets**: Start with preset configurations and customize as needed
2. **Component Order**: Think about the visual flow when ordering components
3. **Background Styling**: Use selective background modes for better visual hierarchy
4. **Performance**: Disable unused components for better performance
5. **Consistency**: Maintain consistent styling across your application

## 🔮 **Future Extensions**

The component system is designed to be extensible. Future components might include:

- **Prefix/Suffix Components**: For additional decorations
- **Status Components**: For showing item states
- **Custom Components**: User-defined rendering components
- **Animation Components**: For animated transitions

The component-based system provides the foundation for these future enhancements while maintaining clean separation of concerns. 