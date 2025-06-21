# Cursor Visualization

## What We're Adding

Building on our styled table, we're adding different cursor visualization modes:
- **Full Row Highlighting**: The entire row is highlighted when selected
- **Active Cell Indication**: Only the current cell is highlighted with a background color
- **Mixed Mode**: Combining row highlighting with active cell indication

This gives you fine-grained control over how users see their current position in the table.

## Code Changes

Starting from our table styling example, we need to add cursor visualization configuration:

```go
// Add these fields to your model
type AppModel struct {
	table                   *table.Table
	// Cursor visualization fields
	fullRowHighlightEnabled bool
	activeCellEnabled       bool
	activeCellColorIndex    int
	activeCellColors        []string
}

func main() {
	// ... existing setup code ...
	
	activeCellColors := []string{"#3C3C3C", "#1E3A8A", "#166534", "#7C2D12", "#581C87"}
	
	// Configure cursor visualization in table config
	config := core.TableConfig{
		Columns:     columns,
		ShowHeader:  true,
		ShowBorders: true,
		// Cursor visualization configuration
		FullRowHighlighting:         true,  // Enable full row highlighting
		ActiveCellIndicationEnabled: false, // Disable active cell indication initially
		ActiveCellBackgroundColor:   activeCellColors[0], // Set initial color
		// ... rest of config ...
	}
	
	model := AppModel{
		table:                   tbl,
		fullRowHighlightEnabled: true,
		activeCellEnabled:       false,
		activeCellColorIndex:    0,
		activeCellColors:        activeCellColors,
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			// Toggle full row highlighting
			m.fullRowHighlightEnabled = !m.fullRowHighlightEnabled
			if m.fullRowHighlightEnabled && m.activeCellEnabled {
				m.activeCellEnabled = false // Disable cell indication when enabling row
			}
			return m, core.FullRowHighlightEnableCmd(m.fullRowHighlightEnabled)
			
		case "c":
			// Toggle active cell indication
			m.activeCellEnabled = !m.activeCellEnabled
			if m.activeCellEnabled && m.fullRowHighlightEnabled {
				m.fullRowHighlightEnabled = false // Disable row highlighting when enabling cell
			}
			return m, core.ActiveCellIndicationModeSetCmd(m.activeCellEnabled)
			
		case "C":
			// Cycle active cell background colors
			m.activeCellColorIndex = (m.activeCellColorIndex + 1) % len(m.activeCellColors)
			newColor := m.activeCellColors[m.activeCellColorIndex]
			return m, core.ActiveCellBackgroundColorSetCmd(newColor)
			
		case "m":
			// Toggle mixed mode (both row highlighting and active cell)
			if m.fullRowHighlightEnabled && m.activeCellEnabled {
				// Both on, turn both off
				m.fullRowHighlightEnabled = false
				m.activeCellEnabled = false
				return m, tea.Batch(
					core.FullRowHighlightEnableCmd(false),
					core.ActiveCellIndicationModeSetCmd(false),
				)
			} else {
				// Turn both on
				m.fullRowHighlightEnabled = true
				m.activeCellEnabled = true
				return m, tea.Batch(
					core.FullRowHighlightEnableCmd(true),
					core.ActiveCellIndicationModeSetCmd(true),
				)
			}
		}
	}

	// Delegate to table for all other messages
	var cmd tea.Cmd
	_, cmd = m.table.Update(msg)
	return m, cmd
}

## Core Commands

The library provides these commands for controlling cursor visualization:

```go
// Enable/disable full row highlighting
core.FullRowHighlightEnableCmd(enabled bool) tea.Cmd

// Enable/disable active cell indication
core.ActiveCellIndicationModeSetCmd(enabled bool) tea.Cmd  

// Set active cell background color
core.ActiveCellBackgroundColorSetCmd(color string) tea.Cmd
```

## Key Features

### Full Row Highlighting (`r`)
- Highlights the entire row where the cursor is positioned
- Uses the theme's cursor/selection colors
- Good for emphasizing the current record
- Automatically disables active cell indication when enabled

### Active Cell Indication (`c`)
- Highlights only the cell that would be "active" for editing
- Uses a customizable background color
- Useful for precise cell-level operations
- Automatically disables full row highlighting when enabled

### Color Cycling (`C`)
- Cycles through different background colors for active cell indication
- Available colors: Dark Gray, Blue, Green, Brown, Purple
- Only affects active cell indication mode

### Mixed Mode (`m`)
- Combines both row highlighting and active cell indication
- Provides maximum visual feedback
- Useful for complex data entry scenarios where you need both row context and cell precision

## Controls

| Key | Action |
|-----|--------|
| `r` | Toggle full row highlighting |
| `c` | Toggle active cell indication |
| `C` | Cycle active cell background colors |
| `m` | Toggle mixed mode (both row and cell) |
| `↑` `↓` `j` `k` | Navigate rows |
| `h` `l` | Page up/down |
| `g` `G` | Jump to start/end |
| `q` | Quit |

## Try It Yourself

1. **Start with row highlighting**: Press `r` to see how full row highlighting affects the visual emphasis
2. **Switch to cell indication**: Press `c` to see how active cell highlighting provides more precise feedback
3. **Try mixed mode**: Press `m` to see both modes together and understand when you might need maximum visual feedback
4. **Change colors**: Press `C` to cycle through different active cell colors and find what works best for your use case
5. **Test navigation**: Use arrow keys to move around and observe how the cursor visualization follows your movement

## Progressive Enhancement

This example builds on:
- [Table Styling](06-table-styling.md) - Themes, borders, and visual customization
- [Column Formatting](05-column-formatting.md) - Custom cell formatters and styling
- [Cell Constraints](04-cell-constraints.md) - Layout and alignment control

## What's Next

In the next section, we'll explore [Horizontal Scrolling](09-horizontal-scrolling.md) to handle tables with many columns that don't fit on screen.

## Running the Example

```bash
cd docs/05-table-component/examples/cursor-visualization
go run .
```

The example demonstrates how different cursor visualization modes affect user experience and help users understand their current position in complex tables. 