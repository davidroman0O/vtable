# The Table Component: Styling and Themes

With custom column formatting in place, let's focus on the overall visual appearance of the table. This guide covers how to use **Themes** to control colors, borders, and other stylistic elements of your table.

## What You'll Build

We will create several distinct visual themes for our employee table and add the ability to cycle through them dynamically. You will learn to control every aspect of the table's look and feel.

![VTable Table Styling Demo](examples/table-styling/table-styling.gif)

**Default Theme:**
```
┌───────────────────┬───────────────┬────────────┐
│   Employee Name   │  Department   │   Status   │
├───────────────────┼───────────────┼────────────┤
│ ► Alice Johnson   │  Engineering  │ 🟢 Active  │
└───────────────────┴───────────────┴────────────┘
```

**Heavy Border Theme:**
```
╔═══════════════════╦═══════════════╦════════════╗
║   Employee Name   ║  Department   ║   Status   ║
╠═══════════════════╬═══════════════╬════════════╣
║ ► Alice Johnson   ║  Engineering  ║ 🟢 Active  ║
╚═══════════════════╩═══════════════╩════════════╝
```

**Minimal (Borderless) Theme:**
```
  Employee Name     Department      Status
  ───────────     ──────────     ──────
  ► Alice Johnson    Engineering    🟢 Active
```

## How It Works: The `core.Theme` Struct

All table styling is controlled by the `core.Theme` struct, which is part of your `TableConfig`. A theme defines a collection of `lipgloss.Style` objects and border characters.

```go
type Theme struct {
	HeaderStyle        lipgloss.Style
	CellStyle          lipgloss.Style
	CursorStyle        lipgloss.Style
	SelectedStyle      lipgloss.Style
	FullRowCursorStyle lipgloss.Style
	BorderChars        core.BorderChars
	BorderColor        string
	// ... and other styles for different states
}
```

By creating different `Theme` objects, you can define completely different visual appearances for your table.

## Step 1: Define Your Custom Themes

It's a good practice to define your themes separately, so they can be easily managed and switched.

```go
type AppTheme struct {
	Name        string
	Description string
	VTableTheme core.Theme
}

var myThemes = []AppTheme{
	{
		Name:        "Default",
		Description: "Clean Unicode box drawing",
		VTableTheme: core.Theme{
			HeaderStyle: lipgloss.NewStyle().Foreground("#99").Bold(true),
			CellStyle:   lipgloss.NewStyle().Foreground("#FAFAFA"),
			BorderChars: core.DefaultBorderChars(),
			BorderColor: "240", // Gray
			// ... other styles
		},
	},
	{
		Name:        "Heavy",
		Description: "Double-line borders for emphasis",
		VTableTheme: core.Theme{
			HeaderStyle: lipgloss.NewStyle().Foreground("#FFFF00").Bold(true),
			BorderChars: core.BorderChars{ // Heavy borders
				Horizontal: "═", Vertical: "║", TopLeft: "╔", /* ... */
			},
			BorderColor: "33", // Blue
			// ... other styles
		},
	},
	// ... more themes
}
```

## Step 2: Apply a Theme to the Table

To apply a theme, you send a `core.TableThemeSetCmd` with the desired `core.Theme` object.

```go
// In your app's Update method:
case tea.KeyMsg:
    switch msg.String() {
    case "t": // 't' for theme
        // Cycle to the next theme
        app.currentTheme = (app.currentTheme + 1) % len(myThemes)
        selectedTheme := myThemes[app.currentTheme]

        // Send the command to update the table's theme
        return app, core.TableThemeSetCmd(selectedTheme.VTableTheme)
    }
```

## Step 3: Granular Border Control

Beyond full themes, VTable allows you to control individual border elements at runtime.

#### Border Visibility Commands
-   `core.BorderVisibilityCmd(bool)`: Toggle all borders on or off.
-   `core.TopBorderVisibilityCmd(bool)`: Toggle only the top border.
-   `core.BottomBorderVisibilityCmd(bool)`: Toggle only the bottom border.
-   `core.HeaderSeparatorVisibilityCmd(bool)`: Toggle the line between the header and the data.

#### Border Space Removal
For creating compact, truly borderless layouts, you can also remove the lines where the borders would be drawn.
-   `core.TopBorderSpaceRemovalCmd(bool)`
-   `core.BottomBorderSpaceRemovalCmd(bool)`

## What You'll Experience

-   **Thematic Control**: Switch between completely different visual styles with a single key press.
-   **Border Customization**: Fine-tune the exact appearance of your table's borders, from heavy lines to a completely borderless design.
-   **Layout Flexibility**: Combine themes and border controls to create the perfect look for your application.

## Complete Example

See the full working code, which includes an interactive demo for cycling through themes and toggling individual border settings.
[`docs/05-table-component/examples/table-styling/`](examples/table-styling/)

To run it:
```bash
cd docs/05-table-component/examples/table-styling
go run main.go
```
-   Press `t` to cycle through the visual themes.
-   Press `b`, `1`, `2`, `3` to toggle different border components.
-   Press `4`, `5` to see the compact layout with space removal.

## What's Next?

You have now mastered the styling of tables. Next, we'll look at different ways to visualize the cursor's position to enhance user feedback during navigation.

**Next:** [Cursor Visualization →](08-cursor-visualization.md) 