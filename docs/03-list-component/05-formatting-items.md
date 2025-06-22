# The List Component: Formatting Items

So far, our list has only displayed simple strings. Let's unlock the true power of VTable by working with structured data and creating a custom display for each item.

## What You'll Build

We'll transform our list from a simple string display into a rich, formatted list of people, showing multiple data fields in a clean, readable layout.

![Formatted Items Example](examples/formatted-items/formatted-items.gif)

## Step 1: Use a Rich Data Structure

Instead of a `[]string`, your `DataSource` will now manage a slice of `Person` structs.

```go
// Person represents our rich data structure
type Person struct {
	Name string
	Age  int
	City string
	Job  string
}

// Update the DataSource to use the new type
type PersonDataSource struct {
	people   []Person // NEW: Use a slice of structs
	selected map[int]bool
}
```

## Step 2: Create a Custom `ItemFormatter`

An `ItemFormatter` is a function that VTable calls for every visible item, giving you complete control over how it's displayed.

```go
// ItemFormatter is a function with a specific signature
func personFormatter(
    data core.Data[any],
    index int,
    ctx core.RenderContext,
    isCursor, isTopThreshold, isBottomThreshold bool,
) string {
	// 1. Type-assert the generic item back to your specific type
	person := data.Item.(Person)

	// 2. Add conditional logic based on the data
	var ageDisplay string
	if person.Age < 30 {
		ageDisplay = fmt.Sprintf("(%d) 🌟", person.Age) // Young professional
	} else if person.Age > 45 {
		ageDisplay = fmt.Sprintf("(%d) 👑", person.Age) // Senior professional
	} else {
		ageDisplay = fmt.Sprintf("(%d)", person.Age)
	}

	// 3. Return the final formatted string
	return fmt.Sprintf("%s %s - %s in %s",
		person.Name,
		ageDisplay,
		person.Job,
		person.City)
}
```

## Step 3: Configure the List to Use the Formatter

There are two main ways to set the formatter. The recommended approach is to set it in the configuration.

#### Recommended: Set Formatter in `ListRenderConfig`
This method integrates perfectly with VTable's component rendering system, which we'll cover later.

```go
// In your main function:
listConfig := config.DefaultListConfig()

// Set the formatter within the content component's configuration
listConfig.RenderConfig.ContentConfig.Formatter = personFormatter

// Create the list with the updated config
vtableList := list.NewList(listConfig, dataSource)
```

## Step 4: Update `DataSource` to Provide Structs

Finally, ensure your `LoadChunk` method provides `Person` structs in the `core.Data` wrapper.

```go
func (ds *PersonDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		// ...
		for i := request.Start; i < end; i++ {
			chunkItems = append(chunkItems, core.Data[any]{
				ID:   fmt.Sprintf("person-%d", i),
				Item: ds.people[i], // NEW: Pass the entire Person struct
				Selected: ds.selected[i],
			})
		}
		// ...
	}
}
```

## What You'll Experience

-   **Rich Data Display**: Each list item now shows multiple fields from your `Person` struct.
-   **Conditional Formatting**: The emoji next to the age changes based on the person's age, making the data more glanceable.
-   **Full Functionality**: All previous navigation and selection features work perfectly with the new formatted items.

## Complete Example

See the full working code for this guide in the examples directory:
[`docs/03-list-component/examples/formatted-items/`](examples/formatted-items/)

To run it:
```bash
cd docs/03-list-component/examples/formatted-items
go run main.go
```

## What's Next?

Our list now displays rich, formatted data. The next step is to make it visually appealing by adding colors and advanced styling.

**Next:** [Styling and Colors →](06-styling-and-colors.md) 