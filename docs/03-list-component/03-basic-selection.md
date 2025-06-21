# Basic Selection: Choose Your Items

Let's add item selection to our navigation-enhanced list. Same list, now you can select items!

## What We're Adding

Taking our "Item 1, Item 2, Item 3..." list with enhanced navigation and adding:
- **Selection toggle**: Spacebar to select/deselect current item
- **Visual feedback**: Selected items show differently
- **Working DataSource**: Actually implement the selection methods

## Key Changes

### 1. Add Selection Key Handling
```go
func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// ... existing navigation keys ...
		
		// NEW: Selection
		case " ":  // Spacebar
			return app, core.SelectCurrentCmd()
		}
	}
	// ... rest unchanged
}
```

### 2. Enable Selection in List Config
```go
listConfig := config.DefaultListConfig()
listConfig.SelectionMode = core.SelectionMultiple  // Enable selection
```

### 3. Implement DataSource Selection Methods
```go
type SimpleDataSource struct {
	items    []string
	selected map[int]bool  // NEW: Track selected items
}

func (ds *SimpleDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		if index >= 0 && index < len(ds.items) {
			if selected {
				ds.selected[index] = true
			} else {
				delete(ds.selected, index)
			}
			return core.SelectionResponseMsg{
				Success:  true,
				Index:    index,
				Selected: selected,
			}
		}
		return core.SelectionResponseMsg{Success: false}
	}
}
```

### 4. Return Selected Items in LoadChunk
```go
func (ds *SimpleDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		var items []core.Data[any]

		for i := request.Start; i < request.Start+request.Count && i < len(ds.items); i++ {
			items = append(items, core.Data[any]{
				ID:       fmt.Sprintf("item-%d", i),
				Item:     ds.items[i],
				Selected: ds.selected[i],  // NEW: Include selection state
			})
		}
		// ... rest unchanged
	}
}
```

## Selection Concepts

**Selection State**: The DataSource owns and manages which items are selected using a `map[int]bool`.

**Selection Commands**: `core.SelectCurrentCmd()` tells VTable to select the item under the cursor.

**Response Messages**: VTable sends `SelectionResponseMsg` back to confirm selection changes.

**Visual Feedback**: Selected items automatically render differently (VTable handles this).

## What You'll Experience

1. **Navigate normally**: Use j/k, h/l, g/G as before
2. **Press spacebar**: Current item toggles selected state
3. **Visual change**: Selected items show differently (styling varies by theme)
4. **Multiple selection**: Can select multiple items, navigate between them

## Complete Example

See the basic selection example: [`examples/basic-selection/`](examples/basic-selection/)

Run it:
```bash
cd docs/03-list-component/examples/basic-selection
go run main.go
```

## Try It Yourself

1. **Select some items**: Navigate and press spacebar on different items
2. **Mixed selection**: Select items 2, 5, and 8, then navigate around
3. **Toggle off**: Press spacebar again on selected items to deselect
4. **Selection persistence**: Notice selected items stay selected as you scroll

## What's Next

Our list now has basic selection! Next, we'll enhance this with multiple selection features like "select all" and range selection.

**Next:** [Multiple Selection →](04-multiple-selection.md) 