# Core Concepts: DataSources

The `DataSource` is the bridge between your data and VTable. It's a Go interface that you implement to tell VTable how to fetch, count, and manage your data. Whether your data lives in memory, a database, or a remote API, the `DataSource` provides a consistent, asynchronous way for VTable to interact with it.

## The Asynchronous Philosophy

VTable's `DataSource` interface is built around **Bubble Tea's command pattern**. Instead of methods that directly return data (which would block the UI), every method returns a `tea.Cmd`. This command, when run by Bubble Tea's runtime, will eventually produce a message with the requested data or an error.

**Why is this so important?**
This async pattern prevents your UI from freezing. While your `DataSource` is fetching data from a slow database or a network API, VTable can continue to render, show loading indicators, and remain responsive to user input.

## The `DataSource` Interface

Here is the contract your data source must fulfill:

```go
type DataSource[T any] interface {
    // Required for data loading
    LoadChunk(request DataRequest) tea.Cmd
    GetTotal() tea.Cmd

    // Required for selection management
    SetSelected(index int, selected bool) tea.Cmd
    // ... other selection methods ...

    // Required for item identification (the only synchronous method)
    GetItemID(item T) string
}
```

Let's break down the essential methods.

### `GetTotal() tea.Cmd`

Before it can do anything, VTable needs to know the total number of items in your dataset. This is crucial for calculating scrollbar positions, page boundaries, and when to stop requesting more data.

Your implementation should return a command that produces a `core.DataTotalMsg`.

```go
func (ds *MyDataSource) GetTotal() tea.Cmd {
    return func() tea.Msg {
        // This could be from a database, an API call, or just a slice length.
        count, err := ds.database.Count("users")
        if err != nil {
            return core.DataLoadErrorMsg{Error: err}
        }
        return core.DataTotalMsg{Total: count}
    }
}
```

### `LoadChunk(request DataRequest) tea.Cmd`

This is the heart of data virtualization. VTable will call this method whenever it needs a "chunk" of data to display in the viewport or buffer in the bounding area.

Your implementation receives a `DataRequest` and must return a command that produces either a `core.DataChunkLoadedMsg` on success or a `core.DataChunkErrorMsg` on failure.

```go
// DataRequest tells you what VTable needs.
type DataRequest struct {
    Start          int            // The starting index of the requested chunk.
    Count          int            // The number of items to load.
    SortFields     []string       // Optional: Fields to sort by.
    SortDirections []string       // Optional: "asc" or "desc".
    Filters        map[string]any // Optional: Filters to apply.
}

func (ds *MyDataSource) LoadChunk(request DataRequest) tea.Cmd {
    return func() tea.Msg {
        // Simulate a database query with a delay.
        time.Sleep(50 * time.Millisecond)

        // 1. Fetch your raw data (e.g., from a database or API).
        rawData, err := ds.fetchItems(request.Start, request.Count)
        if err != nil {
            return core.DataChunkErrorMsg{Error: err, Request: request}
        }

        // 2. Convert your raw data into core.Data[any] items.
        var chunkItems []core.Data[any]
        for _, item := range rawData {
            chunkItems = append(chunkItems, core.Data[any]{
                ID:       ds.GetItemID(item), // A stable, unique ID
                Item:     item,               // Your original data item
                Selected: ds.isSelected(item),// Include selection state
            })
        }

        // 3. Return the data in a DataChunkLoadedMsg.
        return core.DataChunkLoadedMsg{
            StartIndex: request.Start,
            Items:      chunkItems,
            Request:    request, // Important: Include the original request.
        }
    }
}
```

### `GetItemID(item T) string`

This is the only **synchronous** method in the interface. VTable calls it to get a stable, unique string identifier for each of your data items. This ID is crucial for managing state across data reloads, especially for features like selection and animations.

**Good ID examples:**
-   Database primary keys: `"user-12345"`
-   UUIDs: `"550e8400-e29b-41d4-a716-446655440000"`
-   Composite keys: `"product-ABC-2024"`

```go
func (ds *MyDataSource) GetItemID(item any) string {
    // Assuming 'item' is a struct with a unique 'ID' field.
    if person, ok := item.(Person); ok {
        return fmt.Sprintf("person-%d", person.ID)
    }
    // Fallback for other types.
    return fmt.Sprintf("%v", item)
}
```

## Selection Management

Your `DataSource` is the single source of truth for which items are selected. VTable sends commands to update the selection state, and your implementation should handle the logic.

```go
// Your DataSource needs a way to store selection state.
type MyDataSource struct {
    items         []MyItem
    selectedItems map[string]bool // Use a map for efficient lookups by ID.
}

func (ds *MyDataSource) SetSelected(index int, selected bool) tea.Cmd {
    return func() tea.Msg {
        // 1. Find the item's unique ID.
        item := ds.items[index]
        itemID := ds.GetItemID(item)

        // 2. Update your internal selection state.
        if selected {
            ds.selectedItems[itemID] = true
        } else {
            delete(ds.selectedItems, itemID)
        }

        // 3. Return a response message.
        return core.SelectionResponseMsg{
            Success:  true,
            ID:       itemID,
            Selected: selected,
        }
    }
}
```
When VTable receives the `SelectionResponseMsg`, it knows the operation was successful and will automatically send a `DataChunksRefreshCmd` to reload the visible chunks, ensuring the UI reflects the new selection state.

## What's Next?

You've learned how the `DataSource` connects your data to VTable's virtualization engine. Now, let's explore the viewport system in more detail to understand how navigation and scrolling are calculated.

**Next:** [The Viewport System →](03-viewport-system.md) 