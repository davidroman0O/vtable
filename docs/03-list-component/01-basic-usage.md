# Basic Usage: Your First List

The List component displays a scrollable list of items with data virtualization. Let's start with the simplest possible list.

## What We're Building

A basic list that shows:
```
► Item 1
  Item 2  
  Item 3
  Item 4
  Item 5
```

Navigate with arrow keys, smooth scrolling, handles any dataset size.

## Essential Components

### 1. DataSource
Your data provider that implements `core.DataSource`:

```go
type SimpleDataSource struct {
	items []string
}
```

**Key methods:**
- `GetTotal()` - Returns total item count
- `LoadChunk()` - Loads a range of items for the viewport

### 2. List Creation
```go
import "github.com/davidroman0O/vtable/list"

listConfig := config.DefaultListConfig()
listConfig.ViewportConfig.Height = 5  // Show 5 items

vtableList := list.NewList(listConfig, dataSource)
```

### 3. Navigation Handling
```go
case "up", "k":
    return app, core.CursorUpCmd()
case "down", "j":  
    return app, core.CursorDownCmd()
```

Map keys to VTable commands, then pass other messages to the list.

## Key Concepts

**Data Virtualization**: Only visible items are loaded and rendered. The list handles 10 items or 10 million items the same way.

**Viewport**: The visible area (Height=5 means 5 items shown). The viewport slides over your data as you scroll.

**Chunking**: Data loads in chunks (default 100 items) as needed, not all at once.

## What You Get

**Efficient rendering** - Only 5 items rendered regardless of data size  
**Smooth navigation** - Arrow keys and j/k work immediately  
**Memory efficient** - Constant memory usage  
**Responsive scrolling** - No lag with large datasets  

## Complete Example

See the working example: [`examples/basic-list/`](examples/basic-list/)

Run it:
```bash
cd docs/03-list-component/examples/basic-list
go run main.go
```

## Try It Yourself

1. **Change viewport size**: Modify `Height` to see more/fewer items
2. **Bigger dataset**: Generate 1000 items instead of 20
3. **Different data**: Use your own structs instead of strings

## What's Next

This basic list is your foundation. Next, we'll add more navigation options like page up/down and home/end keys.

**Next:** [Navigation and Keys →](02-navigation-and-keys.md) 