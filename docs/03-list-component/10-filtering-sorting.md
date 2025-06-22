# The List Component: Filtering and Sorting

This guide explains how to add powerful filtering and sorting capabilities to your list. The key insight is that **VTable has built-in support for these features; you just need to implement the data manipulation logic in your `DataSource`.**

## How It Works: The `DataRequest`

Every time VTable needs data, it sends a `DataRequest` to your `DataSource`. This request object has always contained fields for filters and sorts.

```go
type DataRequest struct {
    Start          int
    Count          int
    SortFields     []string        // VTable automatically sends active sort fields.
    SortDirections []string        // e.g., "asc" or "desc".
    Filters        map[string]any  // VTable automatically sends active filters.
}
```
Your `DataSource` is responsible for using these parameters to return the correct slice of data.

## The Filtering and Sorting Flow

1.  **User Action**: The user presses a key (e.g., `'1'`) to toggle a filter.
2.  **VTable Command**: Your app sends a command like `core.FilterSetCmd("job", "Engineer")`.
3.  **VTable State Update**: VTable updates its internal state to know that this filter is active. It then triggers a data refresh.
4.  **DataSource Request**: VTable calls your `DataSource.LoadChunk` method with a `DataRequest` that now includes `Filters: {"job": "Engineer"}`.
5.  **Your Logic**: Your `DataSource` applies this filter to your data, returns the filtered and sorted results, and provides the new total count.
6.  **UI Update**: VTable renders the new, filtered data.

VTable manages the UI state; your `DataSource` handles the data logic.

## Step 1: Implement a Stateful `DataSource`

For filtering and sorting to work correctly, your `DataSource` must be stateful. It needs to keep track of the active filters and sorts so it can provide an accurate total count of the filtered data.

```go
type PersonDataSource struct {
	people         []Person
	filteredData   []Person       // A cached slice of the data after filtering/sorting.
	activeFilters  map[string]any // The currently active filters.
	sortFields     []string
	sortDirections []string
}

// This central method re-applies filters and sorts whenever they change.
func (ds *PersonDataSource) rebuildFilteredData() {
	// 1. Apply filters to the original dataset.
	// 2. Apply sorting to the filtered results.
	// 3. Cache the final result in ds.filteredData.
    // (See the full example for the implementation.)
}
```

## Step 2: Implement Filter and Sort Logic in the `DataSource`

Create public methods on your `DataSource` that your application can call to change the state.

```go
func (ds *PersonDataSource) ToggleFilter(field, value string) {
	if _, ok := ds.activeFilters[field]; ok {
		delete(ds.activeFilters, field)
	} else {
		ds.activeFilters[field] = value
	}
	ds.rebuildFilteredData() // Re-run the filter and sort logic.
}

func (ds *PersonDataSource) ToggleSort(field string) {
    // ... logic to cycle through asc, desc, and none ...
	ds.rebuildFilteredData()
}
```

## Step 3: Update the App to Manage `DataSource` State

Your application is now responsible for orchestrating the state changes.

```go
func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1": // Filter by Engineers
            // 1. Tell the DataSource to update its state.
			app.dataSource.ToggleFilter("job", "Engineer")
            // 2. Tell the List to refresh its data.
			return app, core.DataRefreshCmd()

		case "!": // Sort by Name
			app.dataSource.ToggleSort("name")
			return app, core.DataRefreshCmd()
		}
	}
	// ...
}
```
This pattern ensures that when the `List` asks the `DataSource` for the new total count via `GetTotal()`, the `DataSource` has already been updated and can provide the correct, filtered total.

## What You'll Experience

-   **Interactive Filtering**: Press keys to instantly and reliably filter the list.
-   **Dynamic Sorting**: Sort the list by different criteria on the fly.
-   **Correct UI**: The scrollbar and item counts will now always be accurate, even when filters result in zero items.

## Complete Example

See the full, corrected working code in the examples directory.
[`docs/03-list-component/examples/filtering-sorting/`](examples/filtering-sorting/)

To run it:
```bash
cd docs/03-list-component/examples/filtering-sorting
go run main.go
```
Press number keys (`1`, `2`, `3`...) to apply filters and symbol keys (`!`, `@`, `#`...) to toggle sorting.

## What's Next?

This concludes the guides for the List component. You now have the tools to build highly functional, visually appealing, and performant lists. The same core concepts apply to the Table and Tree components.

**Next:** [The Tree Component: Basic Usage →](../04-tree-component/01-basic-tree.md)
