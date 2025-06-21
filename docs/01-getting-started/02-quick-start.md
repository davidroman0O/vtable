# Quick Start: Your First VTable List

Let's get your first VTable component running in 5 minutes.

## Run the Hello World Example

We have a complete working example ready for you:

```bash
cd docs/01-getting-started/examples/hello-world
go run main.go
```

You should see:
```
Hello World VTable List (press 'q' to quit)

► Item 1
  Item 2
  Item 3
  Item 4
  Item 5

Use ↑/↓ or j/k to navigate
```

**Try it:** Use arrow keys or j/k to navigate, then press `q` to quit.

## What Just Happened?

Let's break down the key parts from `main.go`:

### 1. **DataSource** - Where Your Data Comes From
```go
type SimpleDataSource struct {
    items []string  // Your actual data
}
```

The DataSource provides data to VTable in chunks. It implements these key methods:
- `GetTotal()` - Returns total number of items
- `LoadChunk()` - Loads a specific range of items

### 2. **Navigation Handling** - The Key to Movement
```go
case "up", "k":
    return app, core.CursorUpCmd()
case "down", "j":
    return app, core.CursorDownCmd()
```

This is crucial! You handle keyboard input and return the appropriate movement commands.

### 3. **List Creation** - Put It All Together
```go
listConfig := config.DefaultListConfig()
listConfig.ViewportConfig.Height = 5  // Show 5 items at a time
vtableList := list.NewList(listConfig, dataSource)
```

## Key Benefits You Just Got

Even in this simple example, you already have:

✅ **Virtual rendering** - Only 5 items rendered, regardless of data size  
✅ **Keyboard navigation** - Arrow keys and j/k work!  
✅ **Efficient memory** - Only visible items loaded  
✅ **Responsive scrolling** - Smooth navigation  

## What's Next?

This basic list is your foundation. In the next sections, you'll learn:
- How data virtualization works under the hood
- Creating custom DataSources  
- Viewport and navigation concepts
- Component rendering architecture

**Next:** [Data Virtualization →](../02-core-concepts/01-data-virtualization.md)

**Or jump ahead to:** [Basic List Usage →](../03-list-component/01-basic-usage.md) 