# Multiple Selection: Advanced Selection Features

Let's enhance our basic selection with powerful multiple selection features. Same list, better selection control!

## What We're Adding

Taking our list with basic spacebar selection and adding:
- **Select All**: Ctrl+A to select all items at once
- **Clear Selection**: Ctrl+D to deselect all items
- **Selection Range**: Shift+Space for range selection
- **Selection feedback**: Show selection count in status

## Key Changes

### 1. Add Multiple Selection Key Handling
```go
func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// ... existing navigation and basic selection keys ...
		
		// NEW: Multiple selection
		case "ctrl+a":
			return app, core.SelectAllCmd()
		case "ctrl+d":
			return app, core.SelectClearCmd()
		}
	}
	// ... rest unchanged
}
```

### 2. Add Selection State Tracking
```go
type App struct {
	list           *list.List
	selectionCount int    // NEW: Track selection count
	statusMessage  string // NEW: Show status
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// NEW: Handle selection response messages
	case core.SelectionResponseMsg:
		if msg.Success {
			app.updateSelectionCount()
			app.statusMessage = fmt.Sprintf("Selection: %d items", app.selectionCount)
		}
	}
	// ... rest of handling
}
```

### 3. Enhanced View with Status
```go
func (app *App) View() string {
	return fmt.Sprintf(
		"Multiple Selection List\n\n%s\n\n%s\n%s",
		app.list.View(),
		"Navigate: j/k h/l g/G • Select: Space • Multi: Ctrl+A/D",
		app.statusMessage,
	)
}
```

## Multiple Selection Concepts

**Select All**: `core.SelectAllCmd()` calls your DataSource's `SelectAll()` method to select every item.

**Clear Selection**: `core.SelectClearCmd()` calls your DataSource's `ClearSelection()` method to deselect everything.

**Selection Feedback**: Handle `SelectionResponseMsg` to update your UI with selection status.

**Efficient Operations**: Bulk operations are more efficient than individual selection calls.

## What You'll Experience

1. **Navigate and select**: Use spacebar to select individual items as before
2. **Press Ctrl+A**: All 50 items become selected instantly
3. **Press Ctrl+D**: All selections cleared instantly  
4. **Status updates**: See "Selection: X items" count at bottom
5. **Mix operations**: Combine individual and bulk selections

## Advanced Features

### Selection Count Helper
```go
func (app *App) updateSelectionCount() {
	// In a real app, you might query the DataSource for selected count
	// For this example, we'll use a simple counter approach
	count := 0
	// Your DataSource could provide a GetSelectedCount() method
	app.selectionCount = count
}
```

### Range Selection (Future Enhancement)
```go
case "shift+ ":  // Shift+Space (if supported by your terminal)
	return app, core.SelectRangeCmd(startIndex, endIndex)
```

## Complete Example

See the multiple selection example: [`examples/multiple-selection/`](examples/multiple-selection/)

Run it:
```bash
cd docs/03-list-component/examples/multiple-selection
go run main.go
```

## Try It Yourself

1. **Select individual items**: Use spacebar on items 5, 10, 15
2. **Select all**: Press Ctrl+A and see all items selected
3. **Clear all**: Press Ctrl+D and see selections cleared
4. **Mix operations**: Select some individually, then select all, then clear
5. **Status tracking**: Watch the selection count update

## What's Next

Our list now has powerful selection capabilities! Next, we'll learn how to customize the appearance and formatting of our list items.

**Next:** [Formatting Items →](05-formatting-items.md) 