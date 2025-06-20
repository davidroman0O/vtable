# Component Rendering: Changing How Your List Looks

Let's learn how VTable builds each list item and how you can rearrange the pieces to create different layouts.

## What We're Adding

Taking our styled list and learning:
- **How components work**: Understanding the building blocks of each list item
- **Changing order**: Moving pieces around to create new layouts
- **Adding spacing**: Putting space before and after content

## Understanding Components

VTable builds each list item using **components** - individual pieces that get combined together:

```
►  1. Alice Johnson (28) - UX Designer
^  ^  ^
|  |  └─ Content (your formatted data)
|  └──── Enumerator (numbers, bullets, checkboxes)
└─────── Cursor (shows current position)
```

Every list item is: `[Cursor] + [Enumerator] + [Content]`

## Basic Setup

Start with our familiar styled list:

```go
func main() {
	dataSource := NewPersonDataSource()

	listConfig := config.DefaultListConfig()
	listConfig.ViewportConfig.Height = 8
	listConfig.MaxWidth = 500
	listConfig.SelectionMode = core.SelectionMultiple
	
	// Set formatter in config
	listConfig.RenderConfig.ContentConfig.Formatter = styledPersonFormatter

	// Create list with numbered style
	vtableList := list.NewList(listConfig, dataSource)
	vtableList.SetNumberedStyle()
}
```

This gives us the default order: `[Cursor][Enumerator][Content]`

## Changing Component Order

You can rearrange the components. Let's put numbers at the end:

```go
// Get the render configuration
renderConfig := vtableList.GetRenderConfig()

// Change the order - put enumerator after content
renderConfig.ComponentOrder = []core.ListComponentType{
	core.ListComponentCursor,
	core.ListComponentContent,
	core.ListComponentEnumerator,
}

// Apply the changes
vtableList.SetRenderConfig(renderConfig)
```

This changes your list from:
```
►  1. Alice Johnson (28) - UX Designer
```

To:
```
► Alice Johnson (28) - UX Designer  1.
```

The components are the same, just in a different order: `[Cursor][Content][Enumerator]`

## Adding Spacing Components

VTable has two special spacing components you can enable:

```go
renderConfig := vtableList.GetRenderConfig()

// Enable spacing before and after
renderConfig.PreSpacingConfig.Enabled = true
renderConfig.PreSpacingConfig.Spacing = "  "  // 2 spaces before
renderConfig.PostSpacingConfig.Enabled = true
renderConfig.PostSpacingConfig.Spacing = " "   // 1 space after

// Include spacing in the component order
renderConfig.ComponentOrder = []core.ListComponentType{
	core.ListComponentPreSpacing,
	core.ListComponentCursor,
	core.ListComponentEnumerator,
	core.ListComponentContent,
	core.ListComponentPostSpacing,
}

vtableList.SetRenderConfig(renderConfig)
```

Now your list looks like:
```
  ►  1. Alice Johnson (28) - UX Designer 
```

The order is: `[PreSpacing][Cursor][Enumerator][Content][PostSpacing]`

## Practical Layouts

### Checklist Style
```go
renderConfig := vtableList.GetRenderConfig()
renderConfig.PreSpacingConfig.Enabled = true
renderConfig.PreSpacingConfig.Spacing = " "
renderConfig.PostSpacingConfig.Enabled = true
renderConfig.PostSpacingConfig.Spacing = " "
renderConfig.ComponentOrder = []core.ListComponentType{
	core.ListComponentPreSpacing,
	core.ListComponentEnumerator,
	core.ListComponentContent,
	core.ListComponentPostSpacing,
}
vtableList.SetChecklistStyle()
```

Result: ` ☐ Alice Johnson (28) - UX Designer `

### Content Only
```go
renderConfig := vtableList.GetRenderConfig()
renderConfig.ComponentOrder = []core.ListComponentType{
	core.ListComponentContent,
}
vtableList.SetRenderConfig(renderConfig)
```

Result: `Alice Johnson (28) - UX Designer`

### Custom Symbols
```go
renderConfig := vtableList.GetRenderConfig()
renderConfig.CursorConfig.CursorIndicator = "🔥"
renderConfig.CursorConfig.NormalSpacing = "  "
renderConfig.ComponentOrder = []core.ListComponentType{
	core.ListComponentEnumerator,
	core.ListComponentCursor,
	core.ListComponentContent,
}
vtableList.SetRenderConfig(renderConfig)
```

Result: `1.🔥Alice Johnson (28) - UX Designer`

## All Available Components

```go
core.ListComponentCursor      // ► or spaces
core.ListComponentEnumerator  // 1. or • or ☐
core.ListComponentContent     // Your formatted data
core.ListComponentPreSpacing  // Spaces before everything
core.ListComponentPostSpacing // Spaces after everything
```

## Complete Example

See the component rendering example: [`examples/component-rendering/`](examples/component-rendering/)

Run it:
```bash
cd docs/03-list-component/examples/component-rendering
go run main.go
```

Press 'c' to cycle through different component layouts and see how they work!

## Try It Yourself

1. **Change order**: Move cursor, enumerator, and content around
2. **Add spacing**: Put spaces before and after content
3. **Remove components**: Create minimal layouts with just content
4. **Mix and match**: Create your own custom arrangements

## Key Concepts

**Components**: Each list item is built from individual pieces you can rearrange.

**Component Order**: You control which pieces appear and in what sequence.

**Spacing**: Special spacing components let you add space before and after.

## What's Next

Now you understand how VTable builds list items! You can create any layout by rearranging the component pieces.

**Next:** [Filtering and Sorting →](10-filtering-sorting.md) 