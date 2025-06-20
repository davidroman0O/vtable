# Formatting Items: Custom Display

Let's customize how our list items are displayed. Same list with selection, now with rich formatting!

## What We're Adding

Taking our list with multiple selection features and adding:
- **Rich data**: Replace simple strings with structured data (Person records)
- **Custom formatting**: Show name, age, and city in a formatted layout
- **Visual styling**: Add colors and emphasis to different data fields
- **Conditional formatting**: Different styles based on data values

## Key Changes

### 1. Enhanced Data Structure
```go
type Person struct {
	Name string
	Age  int
	City string
	Job  string
}

type PersonDataSource struct {
	people   []Person      // NEW: Rich structured data
	selected map[int]bool
}
```

### 2. Custom Formatter Function
```go
func personFormatter(data core.Data[any], index int, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
	person := data.Item.(Person)  // Cast to our Person type
	
	// Format: "John Doe (32) - Software Engineer in New York"
	formatted := fmt.Sprintf("%s (%d) - %s in %s", 
		person.Name, 
		person.Age, 
		person.Job, 
		person.City)
	
	return formatted
}
```

### 3. Apply Formatter to List
```go
func main() {
	// ... create dataSource with Person data ...
	
	listConfig := config.DefaultListConfig()
	listConfig.ViewportConfig.Height = 8
	listConfig.SelectionMode = core.SelectionMultiple
	
	// NEW: Create list with custom formatter
	vtableList := list.NewList(listConfig, dataSource, personFormatter)
}
```

### 4. Enhanced Data Loading
```go
func (ds *PersonDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		var items []core.Data[any]

		for i := request.Start; i < request.Start+request.Count && i < len(ds.people); i++ {
			items = append(items, core.Data[any]{
				ID:       fmt.Sprintf("person-%d", i),
				Item:     ds.people[i],  // NEW: Pass Person struct
				Selected: ds.selected[i],
			})
		}
		// ... rest unchanged
	}
}
```

## Formatting Concepts

**Item Formatter**: A function that takes item data and returns a formatted string for display.

**Type Assertion**: Cast `data.Item.(Person)` to access your structured data fields.

**Rich Display**: Show multiple fields in a single list item with custom layout.

**Context Awareness**: Formatter receives cursor/threshold info for conditional styling.

## Advanced Formatting

### Conditional Styling
```go
func styledPersonFormatter(data core.Data[any], index int, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
	person := data.Item.(Person)
	
	// Different formatting based on age
	var ageDisplay string
	if person.Age < 30 {
		ageDisplay = fmt.Sprintf("(%d) 🌟", person.Age)  // Young
	} else if person.Age > 50 {
		ageDisplay = fmt.Sprintf("(%d) 👑", person.Age)  // Senior
	} else {
		ageDisplay = fmt.Sprintf("(%d)", person.Age)     // Normal
	}
	
	return fmt.Sprintf("%s %s - %s in %s", 
		person.Name, ageDisplay, person.Job, person.City)
}
```

### Multi-line Formatting
```go
func detailedPersonFormatter(data core.Data[any], index int, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
	person := data.Item.(Person)
	
	// Create a detailed two-line display
	line1 := fmt.Sprintf("%s, %d years old", person.Name, person.Age)
	line2 := fmt.Sprintf("  %s in %s", person.Job, person.City)
	
	return line1 + "\n" + line2
}
```

## What You'll Experience

1. **Rich data display**: See "Alice Johnson (28) - Designer in San Francisco" instead of "Item 1"
2. **Selection works**: All navigation and selection features work with formatted items
3. **Custom layouts**: See how the same data can be displayed differently
4. **Visual variety**: Icons, emojis, and structured text make the list more engaging

## Complete Example

See the formatted items example: [`examples/formatted-items/`](examples/formatted-items/)

Run it:
```bash
cd docs/03-list-component/examples/formatted-items
go run main.go
```

## Try It Yourself

1. **Explore the data**: Navigate through the list and see the rich person information
2. **Test selection**: Select people and see the selection count with formatted names
3. **Modify the formatter**: Change the layout, add emojis, or show different fields
4. **Add conditions**: Format differently based on age, city, or job

## What's Next

Our list now displays rich, formatted data! Next, we'll learn how to apply colors and advanced styling to make our list even more visually appealing.

**Next:** [Styling and Colors →](06-styling-and-colors.md) 