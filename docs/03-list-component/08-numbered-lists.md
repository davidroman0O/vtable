# Numbered Lists: Using VTable's Enumerator System

Let's add numbers to our styled list using VTable's enumerator system. We'll learn how to use the built-in numbered style and create custom enumerators!

## What We're Adding

Taking our beautifully styled Person list and adding:
- **Built-in numbering**: Using VTable's numbered enumerator
- **Custom enumerators**: Creating your own enumeration functions

## VTable's Enumerator System

VTable uses a **component-based rendering pipeline**. Each list item is rendered by combining different components:

1. **Cursor** - Shows current position (► or spaces)
2. **Enumerator** - Shows numbers, bullets, checkboxes, etc.
3. **Content** - Your formatted item data

## Setup

First, set up your list with the formatter in the config:

```go
func main() {
	dataSource := NewPersonDataSource()

	listConfig := config.DefaultListConfig()
	listConfig.ViewportConfig.Height = 8
	listConfig.MaxWidth = 500
	listConfig.SelectionMode = core.SelectionMultiple
	
	// Set formatter in config
	listConfig.RenderConfig.ContentConfig.Formatter = styledPersonFormatter

	// Create list
	vtableList := list.NewList(listConfig, dataSource)

	// Add numbered enumerator
	vtableList.SetNumberedStyle()
}
```

## Formatter Setup

**Note**: To use enumerators, set your formatter in the config rather than as a parameter to `NewList()`:

```go
// Bypasses component system
vtableList := list.NewList(listConfig, dataSource, styledPersonFormatter)

// Recommended - Formatter in config
listConfig.RenderConfig.ContentConfig.Formatter = styledPersonFormatter
vtableList := list.NewList(listConfig, dataSource)
```

When you pass a formatter to `NewList()`, VTable uses it directly and bypasses the component-based rendering system that includes enumerators.

## Built-in Numbered Style

The simplest way to add numbers:

```go
vtableList.SetNumberedStyle()
```

This gives you:
```
►  1. Alice Johnson (28) 🌟 - UX Designer in San Francisco
   2. Bob Chen (34) - Software Engineer in New York
   3. Carol Rodriguez (45) - Product Manager in Austin
```

## Understanding Enumerator Functions

### Enumerator Function Signature
```go
type ListEnumerator func(item core.Data[any], index int, ctx core.RenderContext) string
```

**Parameters:**
- `item` - The data item being rendered
- `index` - Zero-based position in the list  
- `ctx` - Rendering context (cursor state, etc.)

**Returns:** String to display before the content

### Built-in Enumerators
```go
// Arabic numbers: "1. 2. 3."
list.ArabicEnumerator(item, index, ctx) // Returns "1. ", "2. ", etc.

// Bullet points: "• • •"
list.BulletEnumerator(item, index, ctx)  // Returns "• "

// Checkboxes based on selection: "☐ ☑"
list.CheckboxEnumerator(item, index, ctx) // Returns "☐ " or "☑ "
```

## Custom Enumerators

Create your own enumerator function for specialized numbering:

```go
// Custom bracket numbers: [1] [2] [3]
func customBracketEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	return fmt.Sprintf("[%d] ", index+1)
}

// Smart enumerator that changes based on selection
func smartEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	if item.Selected {
		return "✓ " // Checkmark for selected
	}
	return fmt.Sprintf("%d. ", index+1) // Numbers for unselected
}

// Job-aware enumerator with emojis
func jobAwareEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	person := item.Item.(Person)
	
	if strings.Contains(person.Job, "Manager") {
		return "👑 "
	} else if strings.Contains(person.Job, "Engineer") {
		return "⚙️ "
	} else if strings.Contains(person.Job, "Designer") {
		return "🎨 "
	}
	return fmt.Sprintf("%d. ", index+1)
}
```

### Using Custom Enumerators

```go
// Set a custom enumerator
renderConfig := vtableList.GetRenderConfig()
renderConfig.EnumeratorConfig.Enumerator = customBracketEnumerator
renderConfig.EnumeratorConfig.Alignment = core.ListAlignmentRight
renderConfig.EnumeratorConfig.MaxWidth = 5
vtableList.SetRenderConfig(renderConfig)
```

## Complete Example

See the numbered list example: [`examples/numbered-list/`](examples/numbered-list/)

Run it:
```bash
cd docs/03-list-component/examples/numbered-list
go run main.go
```

## Try It Yourself

1. **Start with SetNumberedStyle()**: Use the built-in numbered style
2. **Create custom functions**: Write your own enumerator that returns different strings
3. **Make it smart**: Use item data to create dynamic enumerators
4. **Test with data**: Try enumerators that respond to selection state or item content

## Key Concepts

**Component Pipeline**: Enumerators are one component in VTable's rendering pipeline.

**Function-Based**: Enumerators are functions that return strings for each item.

**Context Aware**: Access to item data, index, and render context for smart enumeration.

## What's Next

Now you understand VTable's enumerator system! You can create numbered lists and build custom enumerators that respond to your data.

**Next:** [Advanced Features →](09-advanced-features.md)