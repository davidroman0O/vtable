# DataSources: Your Data Provider

DataSources are how you connect your data to VTable components. Whether your data comes from a database, API, file, or memory, the DataSource interface provides a consistent way to feed data to VTable's virtualization system.

## The Async Philosophy

VTable's DataSource interface is built around **Bubble Tea's command pattern**. Instead of directly returning data, every method returns a `tea.Cmd` that will eventually produce a message with the result.

**Why async?** This prevents your UI from freezing while loading data from slow sources like databases or APIs. VTable can show loading indicators and handle errors gracefully while your data loads in the background.

## The DataSource Interface

```go
type DataSource[T any] interface {
    // Data loading
    LoadChunk(request DataRequest) tea.Cmd
    GetTotal() tea.Cmd
    RefreshTotal() tea.Cmd
    
    // Selection management
    SetSelected(index int, selected bool) tea.Cmd
    SetSelectedByID(id string, selected bool) tea.Cmd
    SelectAll() tea.Cmd
    ClearSelection() tea.Cmd
    SelectRange(startIndex, endIndex int) tea.Cmd
    
    // Pure function (synchronous)
    GetItemID(item T) string
}
```

**Notice:** Only `GetItemID` is synchronous. Everything else returns commands that produce messages.

## Core Data Methods

### GetTotal(): Getting the Dataset Size

VTable needs to know how many items exist to calculate scrollbars, page navigation, and viewport bounds.

```go
func (ds *MyDataSource) GetTotal() tea.Cmd {
    return func() tea.Msg {
        // Could be a database count, API call, array length, etc.
        count := len(ds.items)  // or ds.database.Count() or ds.api.GetCount()
        
        return core.DataTotalMsg{Total: count}
    }
}
```

**DataTotalMsg** tells VTable the total number of items. VTable uses this to:
- Position the scrollbar
- Calculate how many chunks exist
- Determine when the user reaches the end
- Set viewport boundaries

### LoadChunk(): The Heart of Virtualization

This is where the magic happens. VTable asks for a specific slice of your data, and you provide it asynchronously.

```go
func (ds *MyDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
    return func() tea.Msg {
        // Simulate database/API call
        time.Sleep(50 * time.Millisecond) // Real loading delay
        
        start := request.Start
        count := request.Count
        
        // Bounds checking
        if start >= len(ds.items) {
            return core.DataChunkLoadedMsg{
                StartIndex: start,
                Items:      []core.Data[any]{}, // Empty chunk
                Request:    request,
            }
        }
        
        end := start + count
        if end > len(ds.items) {
            end = len(ds.items)
        }
        
        // Build the chunk
        var chunkItems []core.Data[any]
        for i := start; i < end; i++ {
            item := ds.items[i]
            chunkItems = append(chunkItems, core.Data[any]{
                ID:       ds.GetItemID(item),
                Item:     item,
                Selected: ds.isSelected(i), // Include selection state
                Error:    ds.getItemError(i), // Any item-specific errors
                Loading:  ds.isItemLoading(i), // Loading state
                Disabled: ds.isItemDisabled(i), // Disabled state
            })
        }
        
        return core.DataChunkLoadedMsg{
            StartIndex: start,
            Items:      chunkItems,
            Request:    request, // Include original request for validation
        }
    }
}
```

**Key points:**
- **Bounds checking**: Always validate start/end indices
- **Error handling**: Return `DataChunkErrorMsg` if loading fails
- **Include request**: VTable validates responses against requests
- **Item state**: Each item can be selected, loading, disabled, or have errors

### DataRequest: What VTable Asks For

The `DataRequest` tells you exactly what data VTable needs:

```go
type DataRequest struct {
    Start          int               // First item index (e.g., 40)
    Count          int               // Number of items (e.g., 20)
    SortFields     []string          // Fields to sort by
    SortDirections []string          // "asc" or "desc" for each field
    Filters        map[string]any    // Active filters
}
```

**Example requests:**
- `{Start: 0, Count: 20}` = "Give me the first 20 items"
- `{Start: 100, Count: 50}` = "Give me items 100-149"
- `{Start: 40, Count: 20, SortFields: ["name"], SortDirections: ["asc"]}` = "Items 40-59, sorted by name ascending"

## Error Handling

When chunk loading fails, return `DataChunkErrorMsg`:

```go
func (ds *MyDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
    return func() tea.Msg {
        data, err := ds.database.QueryRange(request.Start, request.Count)
        if err != nil {
            return core.DataChunkErrorMsg{
                StartIndex: request.Start,
                Error:      err,
                Request:    request,
            }
        }
        
        // ... process successful data ...
        return core.DataChunkLoadedMsg{ /* ... */ }
    }
}
```

VTable handles errors by:
- Showing error indicators in place of data
- Retrying failed chunks when the user scrolls back
- Allowing the application to display error messages

## Selection Management

DataSources manage their own selection state. VTable sends commands, and you update your internal state accordingly.

### SetSelected(): Toggle Individual Items

```go
func (ds *MyDataSource) SetSelected(index int, selected bool) tea.Cmd {
    return func() tea.Msg {
        // Bounds checking
        if index < 0 || index >= len(ds.items) {
            return core.SelectionResponseMsg{
                Success: false,
                Index:   index,
                Error:   fmt.Errorf("index out of bounds"),
            }
        }
        
        // Update selection state
        itemID := ds.GetItemID(ds.items[index])
        if selected {
            ds.selectedItems[itemID] = true
        } else {
            delete(ds.selectedItems, itemID)
        }
        
        return core.SelectionResponseMsg{
            Success:   true,
            Index:     index,
            ID:        itemID,
            Selected:  selected,
            Operation: "toggle",
        }
    }
}
```

### SelectAll(): Bulk Selection

```go
func (ds *MyDataSource) SelectAll() tea.Cmd {
    return func() tea.Msg {
        // Select all items in your dataset
        selectedIDs := make([]string, 0, len(ds.items))
        for i, item := range ds.items {
            itemID := ds.GetItemID(item)
            ds.selectedItems[itemID] = true
            selectedIDs = append(selectedIDs, itemID)
        }
        
        return core.SelectionResponseMsg{
            Success:     true,
            Index:       -1, // No specific index
            Selected:    true,
            Operation:   "selectAll",
            AffectedIDs: selectedIDs,
        }
    }
}
```

### GetItemID(): The Synchronous Method

This is the only method that returns data directly. It extracts a stable, unique identifier from your data:

```go
func (ds *MyDataSource) GetItemID(item any) string {
    if person, ok := item.(Person); ok {
        return fmt.Sprintf("person-%s-%d", person.Name, person.ID)
    }
    return fmt.Sprintf("%v", item) // Fallback
}
```

**Requirements for item IDs:**
- **Stable**: Same item always produces the same ID
- **Unique**: No two different items have the same ID  
- **String**: Must be a string value
- **Consistent**: ID shouldn't change when the item is updated

**Good ID examples:**
- Database primary keys: `"user-12345"`
- UUIDs: `"550e8400-e29b-41d4-a716-446655440000"`
- Composite keys: `"product-ABC-2024"`

## Filtering and Sorting

VTable passes filter and sort criteria in the `DataRequest`. Your DataSource should apply these before returning data:

```go
func (ds *MyDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
    return func() tea.Msg {
        // Apply filters first
        filteredItems := ds.applyFilters(ds.items, request.Filters)
        
        // Then apply sorting
        sortedItems := ds.applySorting(filteredItems, request.SortFields, request.SortDirections)
        
        // Finally, slice the requested chunk
        start := request.Start
        end := start + request.Count
        if end > len(sortedItems) {
            end = len(sortedItems)
        }
        
        var chunkItems []core.Data[any]
        for i := start; i < end; i++ {
            // ... build chunk items ...
        }
        
        return core.DataChunkLoadedMsg{
            StartIndex: start,
            Items:      chunkItems,
            Request:    request,
        }
    }
}

func (ds *MyDataSource) applyFilters(items []MyItem, filters map[string]any) []MyItem {
    if len(filters) == 0 {
        return items
    }
    
    var result []MyItem
    for _, item := range items {
        include := true
        
        // Check each filter
        for field, value := range filters {
            switch field {
            case "status":
                if item.Status != value.(string) {
                    include = false
                    break
                }
            case "category":
                if item.Category != value.(string) {
                    include = false
                    break
                }
            // Add more filter conditions as needed
            }
        }
        
        if include {
            result = append(result, item)
        }
    }
    
    return result
}
```

## Key Takeaways

1. **Everything is async**: Return `tea.Cmd`, not direct data
2. **Messages matter**: Use the correct message types for responses
3. **Selection is yours**: DataSources manage their own selection state
4. **IDs are critical**: Implement `GetItemID` carefully for stable identification
5. **Handle errors**: Return appropriate error messages when things fail
6. **Validate requests**: Check bounds and handle edge cases
7. **Think chunked**: Your data will be requested in pieces, not all at once

The DataSource interface decouples your data from VTable's UI, allowing you to connect any data source while VTable handles all the virtualization complexity.

**Next:** [Viewport System →](03-viewport-system.md) 