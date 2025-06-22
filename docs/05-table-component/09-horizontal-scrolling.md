# The Table Component: Horizontal Scrolling

This guide covers how to manage tables with more columns than can fit on the screen by implementing **horizontal scrolling**. You will learn how to navigate between columns and scroll through wide cell content.

## What You'll Build

We will create a table with a wide "Description" column that requires horizontal scrolling to view its full content. You will implement controls for:
-   **Column Navigation**: Moving the "active cell" focus between different columns.
-   **Content Scrolling**: Scrolling the text *within* a wide cell.
-   **Scroll Modes**: Switching between character, word, and smart scrolling.

```text
// A table where the 'Description' column is horizontally scrolled.
┌────┬──────────────┬───────────────────────────...
│ ID │ Employee Name│ Description
├────┼──────────────┼───────────────────────────...
│ ►  │ Alice Johnson│ lizing in various aspects ...
└────┴──────────────┴───────────────────────────...
```

## How It Works: Active Cell and Scroll Offsets

Horizontal navigation in VTable is managed by two key concepts:

1.  **The Active Column**: This is the column that currently has focus for horizontal scrolling. VTable highlights this column (using the active cell indication style) and directs all horizontal scroll commands to it.
2.  **Scroll Offsets**: VTable maintains a separate horizontal scroll offset for each column. This allows you to scroll a wide "Description" column without affecting the other columns.

You control these states by sending commands.

## Core Horizontal Navigation Commands

#### Column Navigation (Moving Focus)
-   `core.NextColumnCmd()`: Moves the active column focus to the right.
-   `core.PrevColumnCmd()`: Moves the active column focus to the left.

#### Content Scrolling (Within a Column)
-   `core.HorizontalScrollLeftCmd()`: Scrolls the content of the active column to the left.
-   `core.HorizontalScrollRightCmd()`: Scrolls the content to the right.
-   **Word/Smart Scrolling**: `core.HorizontalScrollWordLeftCmd()`, `core.HorizontalScrollSmartRightCmd()`, etc., provide more advanced scrolling behaviors.
-   `core.HorizontalScrollResetCmd()`: Resets the scroll offset for the active column back to the beginning.

## Step 1: Create a Table with Wide Content

First, ensure your table has at least one column with content that is wider than its configured `Width`.

```go
columns := []core.TableColumn{
    // ... other columns ...
    {
        Title: "Description",
        Width: 35, // A fixed width for the column
        // The content in the DataSource will be much longer than 35 chars.
    },
}
```

## Step 2: Implement Keyboard Controls

In your app's `Update` method, map keys to the horizontal navigation and scrolling commands.

```go
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// --- COLUMN NAVIGATION ---
		case ".":
			return m, core.NextColumnCmd()
		case ",":
			return m, core.PrevColumnCmd()

		// --- CONTENT SCROLLING ---
		case "left", "<":
			return m, core.HorizontalScrollLeftCmd()
		case "right", ">":
			return m, core.HorizontalScrollRightCmd()
		case "H", "shift+left": // Fast scroll
			return m, core.HorizontalScrollPageLeftCmd()
		case "L", "shift+right":
			return m, core.HorizontalScrollPageRightCmd()

		// --- SCROLL MODE & RESET ---
		case "s": // Toggle scroll mode (character, word, smart)
			return m, core.HorizontalScrollModeToggleCmd()
		case "backspace", "delete":
			return m, core.HorizontalScrollResetCmd()
		}
	}
	// ...
}
```

## Step 3: Display Scrolling State (Optional)

To provide feedback to the user, you can get the current scrolling state from the table and display it in your `View`.

```go
func (m AppModel) View() string {
    // Get the full horizontal scroll state from the table.
	scrollMode, scrollAllRows, currentColumn, offsets := m.table.GetHorizontalScrollState()

    // Check if any scrolling is active.
	hasActiveScrolling := false
	for _, offset := range offsets {
		if offset > 0 {
			hasActiveScrolling = true
			break
		}
	}
	scrollStatus := "OFF"
	if hasActiveScrolling {
		scrollStatus = fmt.Sprintf("ON (Offset: %d)", offsets[currentColumn])
	}

    // Display the status.
    status := fmt.Sprintf("HScroll: %s (%s) | Active Column: %d",
        scrollStatus,
        scrollMode,
        currentColumn,
	)
    // ...
}
```

## What You'll Experience

-   **Column Navigation**: Pressing `.` and `,` will move the "active cell" highlight between columns.
-   **Content Scrolling**: When the "Description" column is active, pressing `left` and `right` will scroll the text within that column's cells, while other columns remain static.
-   **Scroll Modes**: Pressing `s` will cycle between `character`, `word`, and `smart` scrolling, changing how far the content moves with each key press.
-   **Reset**: Pressing `backspace` will snap the content of the active column back to the beginning.

## Complete Example

See the full working code for this guide in the examples directory:
[`docs/05-table-component/examples/horizontal-scrolling/`](examples/horizontal-scrolling/)

To run it:
```bash
cd docs/05-table-component/examples/horizontal-scrolling
go run .
```

## What's Next?

You now have a fully navigable table that can handle both vertical and horizontal overflow. The next step is to give the user even more control by allowing them to dynamically manage the columns themselves—reordering, adding, and removing them at runtime.

**Next:** [Column Management →](10-column-management.md) 