# Checkbox Lists: Visual Selection Indicators

Let's add checkboxes to our styled list. Same colorful person data, now with "[ ]" and "[x]" selection indicators!

## What We're Adding

Taking our beautifully styled Person list and adding:
- **Checkbox indicators**: Visual "[ ]" for unselected and "[x]" for selected items
- **Clear visual feedback**: Instantly see selection state at a glance

## Key Changes

### 1. Add Checkbox to Formatter
```go
func checkboxPersonFormatter(data core.Data[any], index int, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
	person := data.Item.(Person)

	// NEW: Add checkbox indicator
	var checkbox string
	if data.Selected {
		checkbox = "[x]" // Selected
	} else {
		checkbox = "[ ]" // Unselected
	}

	// Same styling as before
	var nameColor, ageColor, jobColor, cityColor lipgloss.Color
	
	if isCursor {
		nameColor, ageColor, jobColor, cityColor = "#FFFF00", "#FFD700", "#00FFFF", "#00FF00"
	} else if data.Selected {
		nameColor, ageColor, jobColor, cityColor = "#FF69B4", "#FFA500", "#87CEEB", "#98FB98"
	} else {
		nameColor = "#FFFFFF"
		ageColor = getAgeColor(person.Age)
		jobColor = getJobColor(person.Job)
		cityColor = "#98FB98"
	}

	// Style components
	styledCheckbox := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(checkbox)
	styledName := lipgloss.NewStyle().Foreground(nameColor).Bold(true).Render(person.Name)
	styledAge := lipgloss.NewStyle().Foreground(ageColor).Render(getAgeText(person.Age))
	styledJob := lipgloss.NewStyle().Foreground(jobColor).Render(person.Job)
	styledCity := lipgloss.NewStyle().Foreground(cityColor).Render(person.City)

	// NEW: Format with checkbox prefix
	return fmt.Sprintf("%s %s %s - %s in %s",
		styledCheckbox, styledName, styledAge, styledJob, styledCity)
}
```

### 2. Use Checkbox Formatter
```go
func main() {
	dataSource := NewPersonDataSource()

	listConfig := config.DefaultListConfig()
	listConfig.ViewportConfig.Height = 8
	listConfig.MaxWidth = 500
	listConfig.SelectionMode = core.SelectionMultiple

	// Set the formatter in config (Option 3 approach)
	listConfig.RenderConfig.ContentConfig.Formatter = checkboxPersonFormatter

	// Create list with checkbox formatter
	vtableList := list.NewList(listConfig, dataSource)
}
```

## Checkbox Concepts

**Visual State**: [ ] shows unselected, [x] shows selected items.

**Instant Feedback**: Spacebar toggles show immediate checkbox changes.

**Familiar Pattern**: Everyone recognizes checkbox UI.

## What You'll Experience

1. **Clear selection**: See [ ] and [x] indicators for each item
2. **All features work**: Navigation, selection, and styling work as before
3. **Better usability**: Easier to see what's selected in long lists

## Complete Example

See the checkbox list example: [`examples/checkbox-list/`](examples/checkbox-list/)

Run it:
```bash
cd docs/03-list-component/examples/checkbox-list
go run main.go
```

## Try It Yourself

1. **Navigate and select**: Use spacebar to toggle checkboxes
2. **Select all**: Press Ctrl+A and see all become [x]
3. **Clear selection**: Press Ctrl+D and see all become [ ]

## What's Next

Our list now has professional checkbox selection! Next, we'll learn how to add numbered lists with "1. 2. 3." prefixes.

**Next:** [Numbered Lists →](08-numbered-lists.md) 