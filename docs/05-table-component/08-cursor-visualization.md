# The Table Component: Cursor Visualization

This guide covers the different ways you can visualize the cursor in your table. VTable provides several modes for highlighting the user's current position, from highlighting the entire row to indicating a single "active" cell.

## What You'll Build

You will learn how to configure and dynamically switch between three primary cursor styles, giving you fine-grained control over the table's user experience.

![VTable Cursor Visualization](examples/cursor-visualization/cursor-visualization.gif)

**1. Full Row Highlighting:**
The entire row under the cursor is highlighted.
```text
┌───────────────────┬───────────────┬────────────┐
│ ► Employee Name   │  Department   │   Status   │ (Full row has cursor style)
├───────────────────┼───────────────┼────────────┤
│   Alice Johnson   │  Engineering  │ 🟢 Active  │
└───────────────────┴───────────────┴────────────┘
```

**2. Active Cell Indication:**
Only the currently focused cell is highlighted.
```text
┌───────────────────┬───────────────┬────────────┐
│   Employee Name   │ [Department]  │   Status   │ (Cell has background)
├───────────────────┼───────────────┼────────────┤
│ ► Alice Johnson   │  Engineering  │ 🟢 Active  │
└───────────────────┴───────────────┴────────────┘
```

**3. Mixed Mode:**
Combines both styles for maximum visual feedback.
```text
┌───────────────────┬───────────────┬────────────┐
│ ► Employee Name   │ [Department]  │   Status   │ (Row style + Cell style)
├───────────────────┼───────────────┼────────────┤
│   Alice Johnson   │  Engineering  │ 🟢 Active  │
└───────────────────┴───────────────┴────────────┘
```

## How It Works: Configuration and Commands

Cursor visualization is controlled by three properties in your `core.TableConfig` and their corresponding commands.

-   `FullRowHighlighting` (`bool`): Enables or disables full row highlighting.
-   `ActiveCellIndicationEnabled` (`bool`): Enables or disables the single-cell background.
-   `ActiveCellBackgroundColor` (`string`): Sets the color for the active cell background.

You can change these settings at runtime by sending commands from your app.

### Core Commands
-   `core.FullRowHighlightEnableCmd(enabled bool)`: Turns full-row mode on or off.
-   `core.ActiveCellIndicationModeSetCmd(enabled bool)`: Turns active-cell mode on or off.
-   `core.ActiveCellBackgroundColorSetCmd(color string)`: Sets the background color for the active cell.

## Step 1: Configure Your Initial State

In your `main` function, set the initial cursor style in your `TableConfig`. Let's start with full row highlighting enabled and active cell disabled.

```go
config := core.TableConfig{
    // ...
    FullRowHighlighting:         true,  // Start with full row mode
    ActiveCellIndicationEnabled: false, // Start with cell mode off
    ActiveCellBackgroundColor:   "#3C3C3C", // A default color
    // ...
}
```

## Step 2: Implement Controls in Your App

In your app's `Update` method, add key mappings to send the appropriate commands to the table.

```go
// Add these fields to your application model to track the current state.
type AppModel struct {
	table *table.Table
	// ...
	fullRowHighlightEnabled bool
	activeCellEnabled       bool
}


func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// --- CURSOR VISUALIZATION CONTROLS ---
		case "r": // Toggle full row highlighting
			m.fullRowHighlightEnabled = !m.fullRowHighlightEnabled
			return m, core.FullRowHighlightEnableCmd(m.fullRowHighlightEnabled)

		case "c": // Toggle active cell indication
			m.activeCellEnabled = !m.activeCellEnabled
			return m, core.ActiveCellIndicationModeSetCmd(m.activeCellEnabled)

		case "C": // Cycle through different colors for the active cell
			newColor := getNextColor() // Your logic to pick a color
			return m, core.ActiveCellBackgroundColorSetCmd(newColor)

		case "m": // Toggle a "mixed mode"
			m.fullRowHighlightEnabled = !m.fullRowHighlightEnabled
			m.activeCellEnabled = !m.activeCellEnabled
			return m, tea.Batch(
				core.FullRowHighlightEnableCmd(m.fullRowHighlightEnabled),
				core.ActiveCellIndicationModeSetCmd(m.activeCellEnabled),
			)
		}
	}
	// ...
}
```

## What You'll Experience

-   **Full Row Mode (`r`):** Ideal for quickly scanning rows and focusing on a single record at a time.
-   **Active Cell Mode (`c`):** Perfect for spreadsheet-like applications where a single cell is the focus of an operation (e.g., editing). Use the left/right arrow keys to move the active cell between columns.
-   **Mixed Mode (`m`):** Provides the most information, showing both the selected row and the active cell. This is useful in complex data entry forms.

## Complete Example

See the full working code for this guide in the examples directory:
[`docs/05-table-component/examples/cursor-visualization/`](examples/cursor-visualization/)

To run it:
```bash
cd docs/05-table-component/examples/cursor-visualization
go run .
```

## What's Next?

You now have complete control over how the cursor is visualized in your table. The next logical step is to handle tables with more columns than can fit on the screen by implementing horizontal scrolling.

**Next:** [Horizontal Scrolling →](09-horizontal-scrolling.md) 