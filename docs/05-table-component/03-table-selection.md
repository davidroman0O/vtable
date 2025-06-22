# The Table Component: Selection

This guide demonstrates how to add row selection to your virtualized table. You'll learn how to configure single and multiple selection modes, manage selection state in your `DataSource`, and provide users with clear visual feedback.

## What You'll Build

We will enhance our 10,000-employee table with selection capabilities, allowing users to select rows individually or in bulk.

```text
// A table with several rows selected.
┌───────────────────┬───────────────┬────────────┬────────────┐
│   Employee Name   │  Department   │   Status   │     Salary │
├───────────────────┼───────────────┼────────────┼────────────┤
│   Employee 4999   │      ...      │     ...    │        ... │
│ ► Employee 5000   │  Engineering  │   Active   │    $82,000 │
│ ✓ Employee 5001   │   Marketing   │   Remote   │    $73,000 │ (Selected)
│ ✓ Employee 5002   │      ...      │     ...    │        ... │ (Selected)
│   ...             │      ...      │     ...    │        ... │
└───────────────────┴───────────────┴────────────┴────────────┘

Status: 2 items selected | Recent: Selected Employee 5002
```

## Step 1: Configure Selection Mode

In your `TableConfig`, set the `SelectionMode`.

```go
tableConfig := core.TableConfig{
    // ...
    SelectionMode: core.SelectionMultiple, // Or SelectionSingle, SelectionNone
    KeyMap: core.NavigationKeyMap{
        Select:    []string{"enter", " "}, // Use space or enter to select
        SelectAll: []string{"ctrl+a"},     // Use Ctrl+A to select all
        // ... other keymaps
    },
}
```

## Step 2: Track Selection State in the `DataSource`

The `DataSource` is the source of truth for selection. Update it to store the IDs of selected items.

```go
type LargeEmployeeDataSource struct {
	totalEmployees int
	data           []core.TableRow   // Store all data for reliable selection
	selectedItems  map[string]bool // Use a map for efficient lookups
	recentActivity []string        // For UI feedback
}

// In LoadChunk, you must now report the correct selection state for each item.
func (ds *LargeEmployeeDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		// ...
		for i := start; i < end; i++ {
			items = append(items, core.Data[any]{
				ID:       ds.data[i].ID,
				Item:     ds.data[i],
				Selected: ds.selectedItems[ds.data[i].ID], // Look up by ID
			})
		}
		// ...
	}
}
```
**Note:** For reliable selection across a large dataset that isn't fully in memory, your `DataSource` would typically query a database or other persistent store. For this example, we'll load all data into a slice to demonstrate the principle simply.

## Step 3: Implement Selection Methods

Implement the selection methods in your `DataSource` to modify the `selectedItems` map.

```go
// SetSelected toggles selection for a single item.
func (ds *LargeEmployeeDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		if index >= 0 && index < len(ds.data) {
			id := ds.data[index].ID
			if selected {
				ds.selectedItems[id] = true
			} else {
				delete(ds.selectedItems, id)
			}
			// ... logging and response message ...
		}
		// ...
	}
}

// SelectAll selects every item in the dataset.
func (ds *LargeEmployeeDataSource) SelectAll() tea.Cmd {
	return func() tea.Msg {
		for _, row := range ds.data {
			ds.selectedItems[row.ID] = true
		}
		// ... response message ...
	}
}

// ClearSelection removes all selections.
func (ds *LargeEmployeeDataSource) ClearSelection() tea.Cmd {
	return func() tea.Msg {
		ds.selectedItems = make(map[string]bool)
		// ... response message ...
	}
}
```

## Step 4: Handle Selection Commands in the App

Your app's `Update` method should send the appropriate commands when the user presses selection keys.

```go
func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// --- Selection Commands ---
		case " ", "enter":
			return app, core.SelectCurrentCmd()
		case "ctrl+a":
			return app, core.SelectAllCmd()
		case "c":
			return app, core.SelectClearCmd()
		case "s":
			app.showSelectionInfo() // A helper to display status
			return app, nil
		}
	// --- Handle Selection Responses ---
	case core.SelectionResponseMsg:
		app.updateStatus() // Update UI with new selection count
		// Pass the message on to the table.
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd
	// ...
	}
}
```

## What You'll Experience

-   **Interactive Selection**: Press `space` or `enter` to toggle the selection of the current row.
-   **Visual Feedback**: Selected rows are highlighted (the style depends on your theme). VTable provides a `✓` indicator by default.
-   **Bulk Operations**: Press `ctrl+a` to select all 10,000 rows instantly and `c` to clear them.
-   **State Persistence**: The selection state is maintained correctly even as you scroll through different data chunks.

## Complete Example

See the full working code for this guide in the examples directory:
[`docs/05-table-component/examples/selection-table/`](examples/selection-table/)

To run it:
```bash
cd docs/05-table-component/examples/selection-table
go run main.go
```

## What's Next?

With selection working, the next step is to gain fine-grained control over the table's layout by configuring cell constraints for width, alignment, and padding.

**Next:** [Cell Constraints →](04-cell-constraints.md) 