# The Table Component: Data Virtualization

Just like lists, VTable's `Table` component is fully virtualized, allowing it to handle massive datasets with ease. This guide demonstrates how to apply data virtualization to a table, enabling it to smoothly display 10,000 rows (or more) while only loading small chunks of data into memory.

## What You'll Build

We will create a table that appears to hold 10,000 employee records, but only loads and renders the visible rows, ensuring the application remains fast and responsive.

```text
// The UI will be just as fast with 10,000 rows as it was with 10.
┌───────────────────┬───────────────┬────────────┬────────────┐
│   Employee Name   │  Department   │   Status   │     Salary │
├───────────────────┼───────────────┼────────────┼────────────┤
│ ► Employee 5000   │  Engineering  │   Active   │    $82,000 │
│   Employee 5001   │   Marketing   │   Remote   │    $73,000 │
│   ...             │      ...      │     ...    │        ... │
└───────────────────┴───────────────┴────────────┴────────────┘
```

## Step 1: Create a `DataSource` for a Large Dataset

Instead of storing all 10,000 rows in a slice, we'll create a `DataSource` that *simulates* fetching data from a large database by generating it on demand.

```go
type LargeEmployeeDataSource struct {
	totalEmployees int
	// NOTE: We don't store the full dataset here to simulate a true
	// database-backed or API-driven data source.
}

func NewLargeEmployeeDataSource(totalCount int) *LargeEmployeeDataSource {
	return &LargeEmployeeDataSource{totalEmployees: totalCount}
}

// generateEmployees creates a slice of employees for a specific range.
// In a real application, this would be a database query.
func (ds *LargeEmployeeDataSource) generateEmployees(start, count int) []core.TableRow {
	// ... logic to generate 'count' employees starting from 'start' ...
}
```

## Step 2: Implement On-Demand Chunk Loading

The core of table virtualization is the `LoadChunk` method. It now generates data for the requested chunk instead of slicing an existing array.

```go
func (ds *LargeEmployeeDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		// Simulate a realistic database query delay.
		time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

		// Generate only the requested chunk of data.
		rows := ds.generateEmployees(request.Start, request.Count)

		var items []core.Data[any]
		for _, row := range rows {
			items = append(items, core.Data[any]{
				ID:   row.ID,
				Item: row,
			})
		}

		return core.DataChunkLoadedMsg{
			StartIndex: request.Start,
			Items:      items,
			Request:    request,
		}
	}
}
```
The `GetTotal()` method simply returns the total count, which it knows without loading the data.
```go
func (ds *LargeEmployeeDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return core.DataTotalMsg{Total: ds.totalEmployees}
	}
}
```

## Step 3: Configure the Table for Virtualization

To handle a large dataset smoothly, we'll adjust the `ViewportConfig`.

```go
func createLargeTableConfig() core.TableConfig {
	return core.TableConfig{
		// ... same columns as before ...
		ViewportConfig: core.ViewportConfig{
			Height:             10, // A comfortable number of rows to display
			TopThreshold:       3,  // Start loading data when 3 rows from the top
			BottomThreshold:    3,  // Start loading data when 3 rows from the bottom
			ChunkSize:          25, // Load data in chunks of 25
			BoundingAreaBefore: 50, // Keep 50 items before the viewport loaded
			BoundingAreaAfter:  50, // Keep 50 items after the viewport loaded
		},
		// ... other theme and keymap settings ...
	}
}
```
This configuration ensures a seamless user experience by keeping a buffer of 100 rows (50 before, 50 after) in memory, minimizing any loading delays during scrolling.

## Step 4: Add a "Jump to Index" Feature

Navigating a large table row-by-row can be slow. A "jump to index" feature is essential for large datasets.

```go
// In your App's Update method
case tea.KeyMsg:
    if app.showJumpForm {
        // ... handle numeric input for the jump form ...
        case "enter":
            // When user confirms, send the JumpToCmd
            if index, err := strconv.Atoi(app.jumpInput); err == nil {
                // ...
                return app, core.JumpToCmd(index - 1) // VTable uses 0-based index
            }
    } else {
        switch msg.String() {
        case "J": // Use 'J' to trigger the jump form
            app.showJumpForm = true
            app.jumpInput = ""
            return app, nil
        // ... other key handling ...
        }
    }
```

## What You'll Experience

-   **Fast Startup**: The application starts instantly, as it only fetches the total count, not the 10,000 rows.
-   **Smooth Scrolling**: As you scroll, new rows appear seamlessly. You'll see brief "Loading..." placeholders if you scroll very quickly, which is the virtualization engine fetching the next chunk.
-   **Low Memory Usage**: The application's memory footprint remains small and constant, no matter how far you scroll.
-   **Efficient Navigation**: You can jump from employee #1 to employee #9,500 instantly without loading all the rows in between.

## Complete Example

See the full working code for this guide in the examples directory:
[`docs/05-table-component/examples/virtualized-table/`](examples/virtualized-table/)

To run it:
```bash
cd docs/05-table-component/examples/virtualized-table
go run main.go
```

## What's Next?

Now that your table can handle large datasets, the next step is to add selection capabilities, allowing users to select one or more rows.

**Next:** [Table Selection →](03-table-selection.md) 