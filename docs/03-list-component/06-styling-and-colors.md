# Styling and Colors: Beautiful Visual Design

Let's add beautiful colors and styling to our formatted list. Same rich data, now with gorgeous visual design!

## What We're Adding

Taking our formatted Person list and adding:
- **Color coding**: Different colors for names, ages, jobs, and cities
- **Cursor highlighting**: Bright highlighting for the current item
- **Selection styling**: Visual feedback for selected items
- **Conditional colors**: Colors based on data values (age ranges, job types)

## Key Changes

### 1. Import Lipgloss for Styling
```go
import (
	"github.com/charmbracelet/lipgloss"
	// ... other imports
)
```

### 2. Define Color Styles
```go
var (
	// Base styles
	nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	ageStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	jobStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#00CED1")).Italic(true)
	cityStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#98FB98"))
	
	// Cursor highlighting
	cursorStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#5A5A5A")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
	
	// Selection highlighting
	selectedStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#2E8B57")).
		Foreground(lipgloss.Color("#FFFFFF"))
)
```

### 3. Enhanced Styled Formatter
```go
func styledPersonFormatter(data core.Data[any], index int, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
	person := data.Item.(Person)
	
	// Choose age color based on value
	var ageColor lipgloss.Color
	if person.Age < 30 {
		ageColor = "#FFD700" // Gold for young
	} else if person.Age > 45 {
		ageColor = "#9370DB" // Purple for senior
	} else {
		ageColor = "#FFA500" // Orange for mid-career
	}
	
	// Format with colors
	styledName := nameStyle.Render(person.Name)
	styledAge := ageStyle.Foreground(ageColor).Render(fmt.Sprintf("(%d)", person.Age))
	styledJob := jobStyle.Render(person.Job)
	styledCity := cityStyle.Render(person.City)
	
	formatted := fmt.Sprintf("%s %s - %s in %s", styledName, styledAge, styledJob, styledCity)
	
	// Apply cursor or selection highlighting
	if isCursor {
		return cursorStyle.Render(formatted)
	} else if data.Selected {
		return selectedStyle.Render(formatted)
	}
	
	return formatted
}
```

### 4. Job-Based Color Coding
```go
func getJobColor(job string) lipgloss.Color {
	switch {
	case strings.Contains(job, "Engineer"):
		return "#00CED1" // Cyan for engineers
	case strings.Contains(job, "Manager"):
		return "#FF6347" // Tomato for managers
	case strings.Contains(job, "Designer"):
		return "#DA70D6" // Orchid for designers
	case strings.Contains(job, "Lead"):
		return "#32CD32" // Lime for leads
	default:
		return "#87CEEB" // Sky blue for others
	}
}
```

## Styling Concepts

**Lipgloss Styles**: Reusable style objects that define colors, backgrounds, formatting.

**Color Values**: Hex colors like `#FFFFFF` (white) or `#FF0000` (red).

**Conditional Styling**: Different styles based on data values or UI state.

**Cursor vs Selection**: Different highlighting for current position vs selected items.

## Advanced Styling

### Gradient Effects
```go
func gradientAgeStyle(age int) lipgloss.Style {
	// Gradient from green (young) to red (senior)
	intensity := float64(age-20) / float64(60-20) // Normalize 20-60 to 0-1
	
	if intensity < 0 { intensity = 0 }
	if intensity > 1 { intensity = 1 }
	
	// Interpolate between green and red
	red := int(255 * intensity)
	green := int(255 * (1 - intensity))
	
	color := lipgloss.Color(fmt.Sprintf("#%02X%02X00", red, green))
	return lipgloss.NewStyle().Foreground(color)
}
```

### Bordered Items
```go
func borderedFormatter(data core.Data[any], index int, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
	person := data.Item.(Person)
	
	content := fmt.Sprintf("%s (%d)\n%s in %s", 
		person.Name, person.Age, person.Job, person.City)
	
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5A5A5A")).
		Padding(0, 1)
	
	if isCursor {
		style = style.BorderForeground(lipgloss.Color("#FFFF00")) // Yellow border for cursor
	}
	
	return style.Render(content)
}
```

### Theme-Based Styling
```go
type Theme struct {
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color
	Text      lipgloss.Color
	Cursor    lipgloss.Color
}

var DarkTheme = Theme{
	Primary:   "#1E1E1E",
	Secondary: "#2D2D30",
	Accent:    "#007ACC",
	Text:      "#FFFFFF",
	Cursor:    "#FFFF00",
}
```

## What You'll Experience

1. **Colorful display**: Names in white bold, ages in gold/purple/orange, jobs in cyan, cities in green
2. **Cursor highlighting**: Current item gets bright background highlighting
3. **Selection feedback**: Selected items show with green background
4. **Visual hierarchy**: Different elements clearly distinguished by color
5. **Professional look**: Beautiful terminal UI that's easy to read and navigate

## Complete Example

See the styled list example: [`examples/styled-list/`](examples/styled-list/)

Run it:
```bash
cd docs/03-list-component/examples/styled-list
go run main.go
```

## Try It Yourself

1. **Navigate and see**: Move through the list and see cursor highlighting
2. **Select items**: Use spacebar and see green selection highlighting
3. **Observe colors**: Notice how different data fields have different colors
4. **Modify styles**: Change colors, add borders, try different themes
5. **Test combinations**: See how cursor and selection styles interact

## What's Next

Our list now has beautiful colors and styling! Next, we'll learn about performance optimization for very large datasets and advanced data loading techniques.

**Next:** [Performance and Large Datasets →](07-performance-large-datasets.md) 