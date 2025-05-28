# Enhanced List Rendering

The vtable List component now includes advanced rendering capabilities inspired by lipgloss/list but fully integrated with our pure Tea model architecture.

## Features

### 🎯 **Enumerators**
Multiple built-in enumerator styles for list items:

- **Bullet** (`•`) - Clean bullet points
- **Numbered** (`1. 2. 3.`) - Sequential numbering with alignment
- **Alphabetical** (`a. b. c.`) - Alphabetical enumeration
- **Checkbox** (`☐ ☑`) - Interactive checkbox style
- **Dash** (`-`) - Simple dash points
- **Arrow** (`→`) - Arrow-style points
- **Custom** - Define your own patterns with placeholders

### 🎨 **Conditional Formatting**
Smart enumerators that change based on item state:

```go
// Automatically shows checkboxes for selected items, 
// error symbols for errors, loading spinners for loading items
list.SetConditionalStyle()
```

### 📐 **Text Wrapping & Alignment**
- Automatic text wrapping with proper indentation
- Enumerator alignment for consistent spacing
- Multi-line content support with continuation indentation

### ⚙️ **Configuration Options**
Full control over rendering behavior:

```go
config := vtable.ListRenderConfig{
    Enumerator:      vtable.BulletEnumerator,
    ShowEnumerator:  true,
    IndentSize:      2,
    ItemSpacing:     0,
    MaxWidth:        80,
    WrapText:        true,
    AlignEnumerator: true,
}
```

## Quick Start

### Basic Usage

```go
// Create a list with default bullet points
list := vtable.NewList(config, dataSource)

// Switch to numbered list
list.SetNumberedStyle()

// Switch to checklist style
list.SetChecklistStyle()
```

### Custom Enumerators

```go
// Use a custom pattern with placeholders
list.SetCustomEnumerator("[{index1}] ")  // [1] [2] [3]
list.SetCustomEnumerator("{id}: ")       // item-1: item-2:

// Create a completely custom enumerator
customEnum := func(item vtable.Data[any], index int, ctx vtable.RenderContext) string {
    if item.Selected {
        return "★ "
    }
    return "☆ "
}
list.SetEnumerator(customEnum)
```

### Conditional Formatting

```go
// Create conditional enumerator
conditionalEnum := vtable.NewConditionalEnumerator(vtable.BulletEnumerator).
    When(vtable.IsSelected, vtable.CheckboxEnumerator).
    When(vtable.IsError, func(item vtable.Data[any], index int, ctx vtable.RenderContext) string {
        return "❌ "
    }).
    When(vtable.IsLoading, func(item vtable.Data[any], index int, ctx vtable.RenderContext) string {
        return "⏳ "
    })

list.SetEnumerator(conditionalEnum.Enumerate)
```

### Advanced Configuration

```go
// Configure detailed rendering options
renderConfig := vtable.ListRenderConfig{
    Enumerator:      vtable.ArabicEnumerator,
    ShowEnumerator:  true,
    IndentSize:      4,
    ItemSpacing:     1,
    MaxWidth:        100,
    WrapText:        true,
    AlignEnumerator: true,
}

list.SetRenderConfig(renderConfig)

// Or configure individual options
list.SetEnumeratorAlignment(true)
list.SetTextWrapping(true)
list.SetIndentSize(4)
```

## Built-in Enumerators

| Enumerator | Example Output | Use Case |
|------------|----------------|----------|
| `BulletEnumerator` | `• Item 1`<br>`• Item 2` | General lists |
| `ArabicEnumerator` | `1. Item 1`<br>`2. Item 2` | Ordered lists |
| `AlphabetEnumerator` | `a. Item 1`<br>`b. Item 2` | Alphabetical lists |
| `CheckboxEnumerator` | `☐ Item 1`<br>`☑ Item 2` | Todo lists |
| `DashEnumerator` | `- Item 1`<br>`- Item 2` | Simple lists |
| `ArrowEnumerator` | `→ Item 1`<br>`→ Item 2` | Navigation lists |

## Specialized Formatters

Pre-built formatters for common use cases:

```go
// Ready-to-use formatters
checklistFormatter := vtable.ChecklistFormatter()
numberedFormatter := vtable.NumberedListFormatter()
bulletFormatter := vtable.BulletListFormatter()
alphabeticalFormatter := vtable.AlphabeticalListFormatter()
conditionalFormatter := vtable.ConditionalListFormatter()

// Apply to list
list.SetFormatter(checklistFormatter)
```

## Multi-line Support

The enhanced rendering system properly handles multi-line content:

```go
// Multi-line items are automatically indented
items := []vtable.Data[any]{
    {ID: "1", Item: "Short item"},
    {ID: "2", Item: "This is a very long item that will wrap to multiple lines and be properly indented"},
}

// Configure wrapping
list.SetTextWrapping(true)
list.SetIndentSize(2)
```

Output:
```
• Short item
• This is a very long item that will wrap to
  multiple lines and be properly indented
```

## Integration with Existing Features

The enhanced rendering system works seamlessly with all existing vtable features:

- ✅ **Selection** - Checkboxes automatically reflect selection state
- ✅ **Error States** - Error items can have special enumerators
- ✅ **Loading States** - Loading items can show spinners
- ✅ **Styling** - Full lipgloss styling integration
- ✅ **Animations** - Compatible with animated formatters
- ✅ **Chunked Loading** - Works with large datasets
- ✅ **Filtering & Sorting** - Enumerators update with data changes

## Performance

The enhanced rendering system is designed for performance:

- **Lazy Evaluation** - Enumerators only calculated when needed
- **Caching** - Width calculations cached for alignment
- **Memory Efficient** - No additional memory overhead
- **Fast Rendering** - Optimized for large lists

## Migration from Basic Rendering

Existing lists automatically get enhanced rendering with bullet points. To migrate:

```go
// Before (still works)
list := vtable.NewList(config, dataSource)

// After (enhanced features)
list.SetNumberedStyle()           // Add numbering
list.SetEnumeratorAlignment(true) // Align numbers
list.SetTextWrapping(true)        // Enable wrapping
```

## Examples

See `examples/enhanced_list_example.go` for a complete working example that demonstrates:

- Switching between different enumerator styles
- Selection integration with checkboxes
- Error and loading state visualization
- Text wrapping and alignment
- Interactive style switching

Run the example:
```bash
cd pure/examples
go run enhanced_list_example.go
```

## Best Practices

1. **Choose the Right Enumerator**
   - Use bullets for general lists
   - Use numbers for ordered/sequential content
   - Use checkboxes for selectable items
   - Use conditional for mixed content

2. **Configure Alignment**
   - Enable alignment for numbered lists
   - Disable for simple bullets to save space

3. **Handle Long Content**
   - Enable text wrapping for variable-length content
   - Set appropriate indent size for readability

4. **Performance Considerations**
   - Use simple enumerators for very large lists
   - Cache custom enumerator results if expensive to calculate

## Future Enhancements

Planned features for future releases:

- **Nested Lists** - Support for hierarchical list structures
- **Custom Styling** - Per-enumerator styling options
- **Animation Integration** - Animated enumerator transitions
- **Accessibility** - Screen reader optimizations
- **Themes** - Pre-built enumerator themes 