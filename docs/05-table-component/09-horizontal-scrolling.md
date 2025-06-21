# Horizontal Scrolling

## What We're Adding

Building on our cursor visualization example, we're adding horizontal scrolling capabilities to handle tables with many columns that don't fit on screen:
- **Character-based scrolling**: Smooth scrolling one character at a time
- **Word-based scrolling**: Jump by word boundaries for faster navigation
- **Smart scrolling**: Automatically choose the best scrolling method
- **Scope control**: Control which columns participate in scrolling
- **Reset behavior**: Return to the beginning or specific positions

This is essential for tables with many columns or wide content that exceeds the terminal width.

## Code Changes

Starting from our cursor visualization example, we need to add horizontal scrolling configuration:

```go
type AppModel struct {
	table                   *table.Table
	// Cursor visualization fields
	fullRowHighlightEnabled bool
	activeCellEnabled       bool
	activeCellColorIndex    int
	activeCellColors        []string
	// Horizontal scrolling fields (now managed by table)
	// These are just for display purposes
	scrollModeLabels        []string
}

func main() {
	// Create columns with fewer columns but keep description for horizontal scrolling demo
	columns := []core.TableColumn{
		{Title: "ID", Width: 8, Alignment: core.AlignCenter},
		{Title: "Employee Name", Width: 25, Alignment: core.AlignLeft},
		{Title: "Department", Width: 20, Alignment: core.AlignCenter},
		{Title: "Status", Width: 15, Alignment: core.AlignCenter},
		{Title: "Salary", Width: 12, Alignment: core.AlignRight},
		{Title: "Description", Width: 50, Alignment: core.AlignLeft}, // Wide column to demonstrate horizontal scrolling
	}

	activeCellColors := []string{"#3C3C3C", "#1E3A8A", "#166534", "#7C2D12", "#581C87"}
	scrollModeLabels := []string{"character", "word", "smart"}

	// Configure table
	config := core.TableConfig{
		Columns:     columns,
		ShowHeader:  true,
		ShowBorders: true,
		// Cursor visualization configuration
		FullRowHighlighting:         false, // Disable full row highlighting by default
		ActiveCellIndicationEnabled: true,  // Enable active cell indication by default
		ActiveCellBackgroundColor:   activeCellColors[0],
		ViewportConfig: core.ViewportConfig{
			Height:             15,
			ChunkSize:          25,
			TopThreshold:       3,
			BottomThreshold:    3,
			BoundingAreaBefore: 50,
			BoundingAreaAfter:  50,
		},
		Theme:         theme,
		SelectionMode: core.SelectionNone,
	}

	model := AppModel{
		table:                   tbl,
		fullRowHighlightEnabled: false, // Start with full row highlighting disabled
		activeCellEnabled:       true,  // Start with active cell indication enabled
		activeCellColorIndex:    0,
		activeCellColors:        activeCellColors,
		scrollModeLabels:        scrollModeLabels,
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Existing cursor visualization controls...
		case "r":
			m.fullRowHighlightEnabled = !m.fullRowHighlightEnabled
			if m.fullRowHighlightEnabled && m.activeCellEnabled {
				m.activeCellEnabled = false
			}
			return m, core.FullRowHighlightEnableCmd(m.fullRowHighlightEnabled)

		// === HORIZONTAL SCROLLING CONTROLS ===
		case "left", "<":
			// Scroll horizontally left within the current column
			m.statusMessage = "Scrolling left"
			return m, core.HorizontalScrollLeftCmd()

		case "right", ">":
			// Scroll horizontally right within the current column
			m.statusMessage = "Scrolling right"
			return m, core.HorizontalScrollRightCmd()

		case "shift+left", "H":
			// Fast scroll left using page-based horizontal scrolling
			m.statusMessage = "Fast scrolling left"
			return m, core.HorizontalScrollPageLeftCmd()

		case "shift+right", "L":
			// Fast scroll right using page-based horizontal scrolling
			m.statusMessage = "Fast scrolling right"
			return m, core.HorizontalScrollPageRightCmd()

		case "[":
			// Word-based scrolling left
			m.statusMessage = "Word scrolling left"
			return m, core.HorizontalScrollWordLeftCmd()

		case "]":
			// Word-based scrolling right
			m.statusMessage = "Word scrolling right"
			return m, core.HorizontalScrollWordRightCmd()

		case "{":
			// Smart scrolling left
			m.statusMessage = "Smart scrolling left"
			return m, core.HorizontalScrollSmartLeftCmd()

		case "}":
			// Smart scrolling right
			m.statusMessage = "Smart scrolling right"
			return m, core.HorizontalScrollSmartRightCmd()

		case ".":
			// Navigate to next column
			m.statusMessage = "Next column"
			return m, core.NextColumnCmd()

		case ",":
			// Navigate to previous column
			m.statusMessage = "Previous column"
			return m, core.PrevColumnCmd()

		case "backspace", "delete":
			// Reset horizontal scrolling
			m.statusMessage = "Resetting horizontal scroll"
			return m, core.HorizontalScrollResetCmd()

		case "s":
			// Toggle scroll mode (character/word/smart)
			m.statusMessage = "Toggling scroll mode"
			return m, core.HorizontalScrollModeToggleCmd()

		case "S":
			// Toggle scroll scope (current row/all rows)
			m.statusMessage = "Toggling scroll scope"
			return m, core.HorizontalScrollScopeToggleCmd()
		}
	}

	// Delegate to table for all other messages
	var cmd tea.Cmd
	_, cmd = m.table.Update(msg)
	return m, cmd
}

func (m AppModel) View() string {
	// Get horizontal scrolling state from table
	scrollMode, scrollAllRows, currentColumn, offsets := m.table.GetHorizontalScrollState()
	
	// Determine if any horizontal scrolling is active
	hasActiveScrolling := false
	for _, offset := range offsets {
		if offset > 0 {
			hasActiveScrolling = true
			break
		}
	}
	
	scrollStatus := "OFF"
	if hasActiveScrolling {
		scrollStatus = "ON"
	}
	
	scopeStatus := "current"
	if scrollAllRows {
		scopeStatus = "all"
	}

	status := fmt.Sprintf("Employee %d/%d | HScroll: %s (%s) | Col: %d | Scope: %s",
		state.CursorIndex+1,
		m.table.GetTotalItems(),
		scrollStatus,
		scrollMode,
		currentColumn,
		scopeStatus,
	)

	return status + "\n" + controls + "\n" + m.statusMessage + "\n\n" + m.table.View()
}

## Core Commands

The library provides two types of horizontal navigation commands:

### Horizontal Scrolling Commands (Content within Columns)
```go
// Basic horizontal scrolling within columns
core.HorizontalScrollLeftCmd() tea.Cmd
core.HorizontalScrollRightCmd() tea.Cmd

// Page-based horizontal scrolling
core.HorizontalScrollPageLeftCmd() tea.Cmd
core.HorizontalScrollPageRightCmd() tea.Cmd

// Word-boundary scrolling
core.HorizontalScrollWordLeftCmd() tea.Cmd
core.HorizontalScrollWordRightCmd() tea.Cmd

// Smart-boundary scrolling
core.HorizontalScrollSmartLeftCmd() tea.Cmd
core.HorizontalScrollSmartRightCmd() tea.Cmd

// Mode and scope controls
core.HorizontalScrollModeToggleCmd() tea.Cmd        // character/word/smart
core.HorizontalScrollScopeToggleCmd() tea.Cmd       // current row/all rows
core.HorizontalScrollResetCmd() tea.Cmd             // reset all scroll offsets
```

### Column Navigation Commands (Moving Between Columns)
```go
// Column navigation (moves focus between columns)
core.CursorLeftCmd() tea.Cmd   // Previous column
core.CursorRightCmd() tea.Cmd  // Next column
core.NextColumnCmd() tea.Cmd   // Next column (alternative)
core.PrevColumnCmd() tea.Cmd   // Previous column (alternative)

// Page-based column navigation
core.PageLeftCmd() tea.Cmd     // Page left through columns
core.PageRightCmd() tea.Cmd    // Page right through columns
```

### Vertical Navigation (Inherited)
```go
core.CursorUpCmd() tea.Cmd
core.CursorDownCmd() tea.Cmd
core.PageUpCmd() tea.Cmd
core.PageDownCmd() tea.Cmd

// Jump navigation
core.JumpToStartCmd() tea.Cmd
core.JumpToEndCmd() tea.Cmd
```

## Key Features

### Scroll Modes

**Character-based scrolling (`s` to cycle)**
- Scrolls one character at a time
- Provides smooth, precise control
- Best for fine-tuned positioning

**Word-based scrolling**
- Jumps by word boundaries
- Faster navigation through content
- Good for text-heavy columns

**Smart scrolling**
- Automatically chooses character or word based on content
- Adapts to column content type
- Balances speed and precision

### Navigation Controls

**Horizontal Scrolling (within columns)**
- `←` `→` `<` `>`: Character-by-character scrolling within focused column
- `Shift+←` `Shift+→` `H` `L`: Page-based horizontal scrolling
- `[` `]`: Word-boundary scrolling
- `{` `}`: Smart-boundary scrolling
- `Backspace` `Delete`: Reset all horizontal scroll offsets

**Column Navigation (between columns)**
- `.`: Move focus to next column
- `,`: Move focus to previous column

**Configuration**
- `s`: Toggle scroll mode (character → word → smart)
- `S`: Toggle scroll scope (current row ↔ all rows)

## Controls

| Key | Action |
|-----|--------|
| `←` `→` `<` `>` | Scroll horizontally within focused column |
| `Shift+←` `Shift+→` `H` `L` | Fast horizontal scrolling (page-based) |
| `[` `]` | Word-boundary horizontal scrolling |
| `{` `}` | Smart-boundary horizontal scrolling |
| `.` `,` | Navigate between columns (focus) |
| `Backspace` `Delete` | Reset horizontal scroll offsets |
| `Home` `End` | Jump to start/end of data |
| `s` | Toggle scroll mode (character/word/smart) |
| `S` | Toggle scroll scope (current row/all rows) |

### Inherited Controls
| Key | Action |
|-----|--------|
| `r` | Toggle full row highlighting |
| `c` | Toggle active cell indication |
| `C` | Cycle active cell colors |
| `m` | Toggle mixed cursor mode |
| `↑` `↓` `j` `k` | Navigate rows |
| `h` `l` | Page up/down |
| `g` `G` | Jump to start/end of data |
| `q` | Quit |

## Try It Yourself

1. **Test horizontal scrolling**: Use `←` `→` to scroll within the wide Description column
2. **Try different scroll modes**: Press `s` to cycle through character, word, and smart scrolling
3. **Test word boundaries**: Use `[` `]` to jump by word boundaries
4. **Navigate columns**: Press `.` `,` to move focus between columns
5. **Toggle scope**: Press `S` to switch between current-row and all-rows scrolling
6. **Reset scrolling**: Press `Backspace` to reset all scroll offsets

## Progressive Enhancement

This example builds on:
- [Cursor Visualization](08-cursor-visualization.md) - Row highlighting and active cell indication
- [Table Styling](06-table-styling.md) - Themes, borders, and visual customization
- [Column Formatting](05-column-formatting.md) - Custom cell formatters and styling

## What's Next

In the next section, we'll explore [Column Management](10-column-management.md) to dynamically control column ordering and configuration.

## Running the Example

```bash
cd docs/05-table-component/examples/horizontal-scrolling
go run .
```

The example demonstrates how horizontal scrolling makes it possible to work with tables that have more columns than can fit on screen, with different scrolling behaviors for different use cases. 