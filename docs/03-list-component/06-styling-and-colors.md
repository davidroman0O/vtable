# The List Component: Styling and Colors

With formatted data in place, let's make our list visually appealing by adding colors and text styles. We'll use the `lipgloss` library to create a beautiful, professional-looking list.

## What You'll Build

We will transform our formatted text list into a styled interface with distinct colors for different data fields and visual feedback for cursor and selection states.

![Styled List Example](examples/styled-list/styled-list.gif)

## Step 1: Define Your `lipgloss` Styles

The best practice is to define your styles as global variables. This makes them easy to reuse and manage.

```go
import "github.com/charmbracelet/lipgloss"

// Color styles for different elements
var (
	// Base styles for different data fields
	nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	ageStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")) // Orange
	jobStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00CED1")).Italic(true) // Cyan
	cityStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#98FB98")) // Light Green

	// Special styles for cursor and selection states
	cursorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#5A5A5A")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2E8B57")). // Sea Green
			Foreground(lipgloss.Color("#FFFFFF"))
)
```

## Step 2: Enhance the Custom Formatter with Styles

Now, we'll update our `personFormatter` to apply these `lipgloss` styles. The key is to render each part of the string with its corresponding style.

```go
func styledPersonFormatter(
    data core.Data[any],
    index int,
    ctx core.RenderContext,
    isCursor, isTopThreshold, isBottomThreshold bool,
) string {
	person := data.Item.(Person)

	// Apply base styles to each part of the data
	styledName := nameStyle.Render(person.Name)
	styledAge := ageStyle.Render(fmt.Sprintf("(%d)", person.Age))
	styledJob := jobStyle.Render(person.Job)
	styledCity := cityStyle.Render(person.City)

	// Combine the styled parts into the final string
	formattedLine := fmt.Sprintf("%s %s - %s in %s",
		styledName, styledAge, styledJob, styledCity)

	// Apply a full-row background style for cursor or selection states
	if isCursor {
		return cursorStyle.Render(formattedLine)
	}
	if data.Selected {
		return selectedStyle.Render(formattedLine)
	}

	return formattedLine
}
```

## Step 3: Use the Styled Formatter

Ensure your list is configured to use this new `styledPersonFormatter`.

```go
// In your main function:
listConfig := config.DefaultListConfig()
// ... other configurations ...

// Set the new styled formatter
listConfig.RenderConfig.ContentConfig.Formatter = styledPersonFormatter

vtableList := list.NewList(listConfig, dataSource)
```

## Step 4: Add Conditional Styling Based on Data

You can make your styling even more dynamic by changing colors based on the data itself.

```go
// Add these helper functions
func getJobColor(job string) lipgloss.Color {
	if strings.Contains(job, "Engineer") {
		return "#00CED1" // Cyan
	}
	// ... other job-based colors
	return "#87CEEB" // Default Sky Blue
}

func getAgeStyle(age int) (lipgloss.Style, string) {
	if age < 30 {
		return lipgloss.NewStyle().Foreground("#FFD700"), fmt.Sprintf("(%d) 🌟", age)
	}
	// ... other age-based styles
	return lipgloss.NewStyle().Foreground("#FFA500"), fmt.Sprintf("(%d)", age)
}

// Update the formatter to use these helpers
func styledPersonFormatter(...) string {
    // ...
    ageStyler, ageText := getAgeStyle(person.Age)
    jobStyler := jobStyle.Copy().Foreground(getJobColor(person.Job))

    styledAge := ageStyler.Render(ageText)
    styledJob := jobStyler.Render(person.Job)
    // ...
}
```

## What You'll Experience

-   **Colorful Display**: Each part of the item (name, age, job) has its own distinct style.
-   **Clear Cursor**: The currently focused item has a prominent background highlight, making it easy to see.
-   **Obvious Selection**: Selected items have a different background color, clearly distinguishing them from the cursor.
-   **Data-Driven Styles**: Colors and icons change based on the item's data (e.g., age or job title).

## Complete Example

See the full working code for this guide in the examples directory:
[`docs/03-list-component/examples/styled-list/`](examples/styled-list/)

To run it:
```bash
cd docs/03-list-component/examples/styled-list
go run main.go
```

## What's Next?

Your list is now not only functional but also visually appealing. Next, we'll explore how to use VTable's built-in enumerator system to easily add prefixes like checkboxes to our list items.

**Next:** [Checkbox Lists →](07-checkbox-lists.md) 