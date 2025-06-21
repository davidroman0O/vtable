# Advanced Background Styling in VTable

This document explains the sophisticated background styling system in VTable that gives you fine-grained control over **exactly what parts** of a list item get styled when it's the cursor line.

## The Problem

In list/tree UIs, developers often want different levels of control over cursor background styling:

- Some want to style **only the content** (not bullets, emojis, or indicators)
- Some want to style **everything** (the entire line including indicators)
- Some want to style **specific parts** (like only the cursor arrow)
- Some need **custom logic** for complex styling needs

The previous system only had a basic `CursorStyleContentOnly` boolean, which was limited.

## The Solution: BackgroundStylingMode

The new system provides 5 different modes for background styling:

### 1. BackgroundStyleEntireLine (Default)
Styles the entire rendered line including cursor indicator, enumerators, and content.

```go
config := vtable.FullLineBackgroundConfig()
// OR manually:
config.BackgroundStylingMode = vtable.BackgroundStyleEntireLine
```

**Visual Result:** `[STYLED: ► • Task content here]`

### 2. BackgroundStyleContentOnly
Styles only the main content, preserving indicators and enumerators unstyled.

```go
config := vtable.ContentOnlyBackgroundConfig()
// OR manually:
config.BackgroundStylingMode = vtable.BackgroundStyleContentOnly
```

**Visual Result:** `► • [STYLED: Task content here]`

### 3. BackgroundStyleWithEnumerator
Styles the enumerator and content, but not the cursor indicator.

```go
config := vtable.DefaultListRenderConfig()
config.EnableCursorBackground = true
config.BackgroundStylingMode = vtable.BackgroundStyleWithEnumerator
```

**Visual Result:** `► [STYLED: • Task content here]`

### 4. BackgroundStyleIndicatorOnly
Styles only the cursor indicator (if shown).

```go
config := vtable.IndicatorOnlyBackgroundConfig()
// OR manually:
config.BackgroundStylingMode = vtable.BackgroundStyleIndicatorOnly
```

**Visual Result:** `[STYLED: ►] • Task content here`

### 5. BackgroundStyleCustom
Allows you to provide a custom function for complete control.

```go
customStyler := func(cursorIndicator, enumerator, content string, isCursor bool, style lipgloss.Style) string {
    // Your custom logic here
    if len(content) > 25 {
        // Long content gets special styling
        return cursorIndicator + enumerator + style.Background(lipgloss.Color("214")).Render(content)
    }
    // Short content gets normal styling
    return cursorIndicator + enumerator + style.Render(content)
}

config := vtable.CustomBackgroundConfig(customStyler)
```

## Convenient Preset Functions

The library provides several preset configuration functions:

```go
// Full line styling (everything)
config := vtable.FullLineBackgroundConfig()

// Content only styling (most common use case)
config := vtable.ContentOnlyBackgroundConfig()

// Indicator only styling
config := vtable.IndicatorOnlyBackgroundConfig()

// No cursor indicator but with background
config := vtable.NoIndicatorBackgroundConfig()

// Custom styler
config := vtable.CustomBackgroundConfig(myCustomStyler)
```

## Migration from Old System

If you were using the old `CursorStyleContentOnly` boolean:

```go
// OLD WAY (deprecated)
config.CursorStyleContentOnly = true

// NEW WAY
config.BackgroundStylingMode = vtable.BackgroundStyleContentOnly
```

The new system is backward compatible - the old field still works but is deprecated.

## Example Usage

```go
package main

import (
    "github.com/charmbracelet/lipgloss"
    vtable "github.com/davidroman0O/vtable/pure"
)

func main() {
    // Create config with content-only styling
    config := vtable.ContentOnlyBackgroundConfig()
    
    // Customize the background style
    config.CursorBackgroundStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("33")).  // Blue background
        Foreground(lipgloss.Color("15")).  // White text
        Bold(true)
    
    // Create formatter
    formatter := vtable.EnhancedListFormatter(config)
    
    // Use in your list...
}
```

## Advanced Custom Stylers

For complex scenarios, you can create sophisticated custom stylers:

```go
// Example: Different styling based on item priority
priorityStyler := func(cursorIndicator, enumerator, content string, isCursor bool, style lipgloss.Style) string {
    if strings.Contains(content, "URGENT") {
        urgentStyle := lipgloss.NewStyle().
            Background(lipgloss.Color("196")). // Red
            Foreground(lipgloss.Color("15")).  // White
            Bold(true)
        return cursorIndicator + urgentStyle.Render(enumerator + content)
    } else if strings.Contains(content, "LOW") {
        lowStyle := lipgloss.NewStyle().
            Background(lipgloss.Color("240")). // Gray
            Foreground(lipgloss.Color("15"))
        return cursorIndicator + enumerator + lowStyle.Render(content)
    }
    
    // Default styling
    styledContent := style.Render(content)
    return cursorIndicator + enumerator + styledContent
}
```

## Integration with Tree Lists

This same system works seamlessly with TreeList components, providing the same level of control over tree item background styling.

## Performance Considerations

- The styling modes have minimal performance impact
- Custom stylers should avoid heavy computations
- Style objects are reused when possible

## Tips and Best Practices

1. **Start Simple**: Use the preset configs (`ContentOnlyBackgroundConfig()`, etc.)
2. **Content Only is Most Common**: Most UIs want to style content but preserve indicators
3. **Test Accessibility**: Ensure sufficient contrast in your background colors
4. **Custom Stylers**: Only use when the built-in modes aren't sufficient
5. **Migration**: Gradually migrate from `CursorStyleContentOnly` to the new system

## See Also

- `examples/background-styling-demo.go` - Complete working example
- Tree-list examples for tree-specific usage
- Styling documentation for general theming 