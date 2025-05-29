package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	vtable "github.com/davidroman0O/vtable/pure"
)

// ================================
// EXAMPLE DATA SOURCE
// ================================

// ExampleTableDataSource provides sample data for the table
type ExampleTableDataSource struct {
	totalItems     int
	data           []vtable.TableRow
	selectedItems  map[string]bool // Actually store selection state!
	recentActivity []string        // Track recent selection activity
}

// NewExampleTableDataSource creates a data source with sample table data
func NewExampleTableDataSource(totalItems int) *ExampleTableDataSource {
	data := make([]vtable.TableRow, totalItems)
	for i := 0; i < totalItems; i++ {
		data[i] = vtable.TableRow{
			ID: fmt.Sprintf("row-%d", i),
			Cells: []string{
				fmt.Sprintf("Item %d", i+1),
				fmt.Sprintf("Value %d", (i*37)%100),
				fmt.Sprintf("Status %d", i%3),
				fmt.Sprintf("Category %c", 'A'+(i%5)),
			},
		}
	}

	return &ExampleTableDataSource{
		totalItems:     totalItems,
		data:           data,
		selectedItems:  make(map[string]bool), // Initialize selection state
		recentActivity: make([]string, 0),     // Initialize activity log
	}
}

// GetTotal returns the total number of items
func (ds *ExampleTableDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return vtable.DataTotalMsg{Total: ds.totalItems}
	}
}

// RefreshTotal refreshes the total count
func (ds *ExampleTableDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

// LoadChunk loads a chunk of data
func (ds *ExampleTableDataSource) LoadChunk(request vtable.DataRequest) tea.Cmd {
	return func() tea.Msg {
		// Simulate loading delay
		time.Sleep(10 * time.Millisecond)

		start := request.Start
		end := start + request.Count
		if end > ds.totalItems {
			end = ds.totalItems
		}

		var items []vtable.Data[any]
		for i := start; i < end; i++ {
			if i < len(ds.data) {
				items = append(items, vtable.Data[any]{
					ID:       ds.data[i].ID,
					Item:     ds.data[i],
					Selected: ds.selectedItems[ds.data[i].ID],
					Metadata: vtable.NewTypedMetadata(),
				})
			}
		}

		return vtable.DataChunkLoadedMsg{
			StartIndex: start,
			Items:      items,
			Request:    request,
		}
	}
}

// SetSelected sets the selection state of an item
func (ds *ExampleTableDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		if index >= 0 && index < len(ds.data) {
			id := ds.data[index].ID

			// Actually update selection state!
			if selected {
				ds.selectedItems[id] = true
				ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected: %s", ds.data[index].Cells[0]))
			} else {
				delete(ds.selectedItems, id)
				ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Deselected: %s", ds.data[index].Cells[0]))
			}

			// Keep only last 10 activities
			if len(ds.recentActivity) > 10 {
				ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
			}

			return vtable.SelectionResponseMsg{
				Success:   true,
				Index:     index,
				ID:        id,
				Selected:  selected,
				Operation: "toggle",
			}
		}

		return vtable.SelectionResponseMsg{
			Success:   false,
			Index:     index,
			ID:        "",
			Selected:  false,
			Operation: "toggle",
			Error:     fmt.Errorf("invalid index: %d", index),
		}
	}
}

// SetSelectedByID sets the selection state of an item by ID
func (ds *ExampleTableDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	return func() tea.Msg {
		// Find the item by ID
		for i, row := range ds.data {
			if row.ID == id {
				// Actually update selection state!
				if selected {
					ds.selectedItems[id] = true
					ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected: %s", row.Cells[0]))
				} else {
					delete(ds.selectedItems, id)
					ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Deselected: %s", row.Cells[0]))
				}

				// Keep only last 10 activities
				if len(ds.recentActivity) > 10 {
					ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
				}

				return vtable.SelectionResponseMsg{
					Success:   true,
					Index:     i,
					ID:        id,
					Selected:  selected,
					Operation: "toggle",
				}
			}
		}

		return vtable.SelectionResponseMsg{
			Success:   false,
			Index:     -1,
			ID:        id,
			Selected:  false,
			Operation: "toggle",
			Error:     fmt.Errorf("item not found: %s", id),
		}
	}
}

// ClearSelection clears all selections
func (ds *ExampleTableDataSource) ClearSelection() tea.Cmd {
	return func() tea.Msg {
		count := len(ds.selectedItems)
		ds.selectedItems = make(map[string]bool) // Clear all selections
		ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Cleared %d selections", count))

		// Keep only last 10 activities
		if len(ds.recentActivity) > 10 {
			ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
		}

		return vtable.SelectionResponseMsg{
			Success:   true,
			Index:     -1,
			ID:        "",
			Selected:  false,
			Operation: "clear",
		}
	}
}

// SelectAll selects all items
func (ds *ExampleTableDataSource) SelectAll() tea.Cmd {
	return func() tea.Msg {
		// Select all items
		for _, row := range ds.data {
			ds.selectedItems[row.ID] = true
		}

		ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected all %d items", len(ds.data)))

		// Keep only last 10 activities
		if len(ds.recentActivity) > 10 {
			ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
		}

		return vtable.SelectionResponseMsg{
			Success:   true,
			Index:     -1,
			ID:        "",
			Selected:  true,
			Operation: "selectAll",
		}
	}
}

// SelectRange selects a range of items
func (ds *ExampleTableDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	return func() tea.Msg {
		var affectedIDs []string
		count := 0

		for i := startIndex; i <= endIndex && i < len(ds.data); i++ {
			ds.selectedItems[ds.data[i].ID] = true
			affectedIDs = append(affectedIDs, ds.data[i].ID)
			count++
		}

		ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected range: %d items", count))

		// Keep only last 10 activities
		if len(ds.recentActivity) > 10 {
			ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
		}

		return vtable.SelectionResponseMsg{
			Success:     true,
			Index:       startIndex,
			ID:          "",
			Selected:    true,
			Operation:   "range",
			AffectedIDs: affectedIDs,
		}
	}
}

// GetItemID returns the ID for a given item
func (ds *ExampleTableDataSource) GetItemID(item any) string {
	if row, ok := item.(vtable.TableRow); ok {
		return row.ID
	}
	return ""
}

// GetRecentActivity returns recent selection activity
func (ds *ExampleTableDataSource) GetRecentActivity() []string {
	return ds.recentActivity
}

// GetSelectionCount returns the number of selected items
func (ds *ExampleTableDataSource) GetSelectionCount() int {
	return len(ds.selectedItems)
}

// ================================
// EXAMPLE TABLE SETUP
// ================================

// CreateExampleTable creates a table with sample configuration
func CreateExampleTable() *vtable.Table {
	// Define columns
	columns := []vtable.TableColumn{
		{Title: "Name", Field: "name", Width: 20, Alignment: vtable.AlignLeft},
		{Title: "Value", Field: "value", Width: 10, Alignment: vtable.AlignRight},
		{Title: "Status", Field: "status", Width: 12, Alignment: vtable.AlignCenter},
		{Title: "Category", Field: "category", Width: 10, Alignment: vtable.AlignCenter},
	}

	// Create table configuration
	config := vtable.TableConfig{
		Columns:     columns,
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: vtable.ViewportConfig{
			Height:             10,
			TopThreshold:       2,
			BottomThreshold:    2,
			ChunkSize:          50,
			InitialIndex:       0,
			BoundingAreaBefore: 25,
			BoundingAreaAfter:  25,
		},
		Theme: vtable.Theme{
			HeaderStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("8")),
			CellStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
			CursorStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("12")),
			SelectedStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("10")),
			BorderChars: vtable.BorderChars{
				Horizontal:  "─",
				Vertical:    "│",
				TopLeft:     "┌",
				TopRight:    "┐",
				BottomLeft:  "└",
				BottomRight: "┘",
				TopT:        "┬",
				BottomT:     "┴",
				LeftT:       "├",
				RightT:      "┤",
				Cross:       "┼",
			},
			BorderColor: "8",
			HeaderColor: "15",
		},
		SelectionMode: vtable.SelectionMultiple,
		KeyMap: vtable.NavigationKeyMap{
			Up:        []string{"up", "k"},
			Down:      []string{"down", "j"},
			PageUp:    []string{"pgup", "ctrl+u"},
			PageDown:  []string{"pgdown", "ctrl+d"},
			Home:      []string{"home", "g"},
			End:       []string{"end", "G"},
			Select:    []string{"enter", " "},
			SelectAll: []string{"ctrl+a"},
			Filter:    []string{"/"},
			Sort:      []string{"s"},
			Quit:      []string{"q", "esc"},
		},
	}

	// Create data source with 1000 items
	dataSource := NewExampleTableDataSource(1000)

	// Create table
	table := vtable.NewTable(config, dataSource)

	return table
}

// ================================
// STYLING SYSTEM
// ================================

// TableTheme defines a complete visual theme for the table
type TableTheme struct {
	Name        string
	Description string

	// Colors
	CursorBg    string
	SelectionBg string
	HeaderBg    string
	BorderColor string

	// Text colors
	PrimaryText   string
	SecondaryText string
	AccentText    string
	ErrorText     string
	WarningText   string
	SuccessText   string

	// Status icons (still useful for status column)
	ActiveIcon  string
	WarningIcon string
	ErrorIcon   string
	UnknownIcon string
}

// Predefined themes
var (
	DefaultTheme = TableTheme{
		Name:          "Default",
		Description:   "Clean blue and white theme",
		CursorBg:      "12", // Blue
		SelectionBg:   "57", // Purple-blue
		HeaderBg:      "8",  // Gray
		BorderColor:   "8",  // Gray
		PrimaryText:   "15", // White
		SecondaryText: "7",  // Light gray
		AccentText:    "14", // Cyan
		ErrorText:     "9",  // Red
		WarningText:   "11", // Yellow
		SuccessText:   "10", // Green
		ActiveIcon:    "✓",
		WarningIcon:   "⚠",
		ErrorIcon:     "✗",
		UnknownIcon:   "?",
	}

	DarkTheme = TableTheme{
		Name:          "Dark",
		Description:   "Dark theme with green accents",
		CursorBg:      "22", // Dark green
		SelectionBg:   "58", // Dark purple
		HeaderBg:      "0",  // Black
		BorderColor:   "8",  // Gray
		PrimaryText:   "15", // White
		SecondaryText: "7",  // Light gray
		AccentText:    "10", // Green
		ErrorText:     "9",  // Red
		WarningText:   "11", // Yellow
		SuccessText:   "10", // Green
		ActiveIcon:    "●",
		WarningIcon:   "▲",
		ErrorIcon:     "■",
		UnknownIcon:   "○",
	}

	MinimalTheme = TableTheme{
		Name:          "Minimal",
		Description:   "Clean minimal theme",
		CursorBg:      "7", // Light gray
		SelectionBg:   "8", // Gray
		HeaderBg:      "0", // Black
		BorderColor:   "8", // Gray
		PrimaryText:   "0", // Black
		SecondaryText: "8", // Gray
		AccentText:    "4", // Blue
		ErrorText:     "1", // Red
		WarningText:   "3", // Yellow
		SuccessText:   "2", // Green
		ActiveIcon:    "+",
		WarningIcon:   "!",
		ErrorIcon:     "x",
		UnknownIcon:   "?",
	}

	NeonTheme = TableTheme{
		Name:          "Neon",
		Description:   "Bright neon cyberpunk theme",
		CursorBg:      "201", // Bright magenta
		SelectionBg:   "57",  // Purple
		HeaderBg:      "0",   // Black
		BorderColor:   "14",  // Cyan
		PrimaryText:   "15",  // White
		SecondaryText: "14",  // Cyan
		AccentText:    "201", // Bright magenta
		ErrorText:     "196", // Bright red
		WarningText:   "226", // Bright yellow
		SuccessText:   "46",  // Bright green
		ActiveIcon:    "◆",
		WarningIcon:   "▲",
		ErrorIcon:     "◼",
		UnknownIcon:   "◯",
	}
)

// Current active theme
var currentTheme = DefaultTheme

// SetTheme changes the active theme
func SetTheme(theme TableTheme) {
	currentTheme = theme
}

// ================================
// ENHANCED CELL FORMATTERS
// ================================

// NameCellFormatter formats the first column with selection indicators (following old codebase concepts)
func NameCellFormatter(cellValue string, rowIndex, columnIndex int, column vtable.TableColumn, ctx vtable.RenderContext, isCursor, isSelected, isTopThreshold, isBottomThreshold bool) string {
	// Add selection indicators to the cell content (like the old codebase)
	value := cellValue
	if isSelected {
		if isCursor {
			value = "✓>" + value // Both selected and cursor
		} else {
			value = "✓ " + value // Just selected
		}
	}

	// Apply cell constraints to ensure content fits within column boundaries
	constraint := CellConstraint{
		Width:     column.Width,
		Height:    1,
		Alignment: vtable.AlignLeft, // Name column: left aligned
	}
	constrainedValue := enforceCellConstraints(value, constraint)

	// Apply row-level styling (not cell-level indicators)
	var style lipgloss.Style
	if isCursor && isSelected {
		// Both cursor and selected: bold selected style
		style = lipgloss.NewStyle().
			Background(lipgloss.Color(currentTheme.SelectionBg)).
			Foreground(lipgloss.Color(currentTheme.PrimaryText)).
			Bold(true)
	} else if isCursor {
		// Just cursor: cursor row style
		style = lipgloss.NewStyle().
			Background(lipgloss.Color(currentTheme.CursorBg)).
			Foreground(lipgloss.Color(currentTheme.PrimaryText)).
			Bold(true)
	} else if isSelected {
		// Just selected: selection style
		style = lipgloss.NewStyle().
			Background(lipgloss.Color(currentTheme.SelectionBg)).
			Foreground(lipgloss.Color(currentTheme.PrimaryText))
	} else {
		// Normal row
		style = lipgloss.NewStyle().
			Foreground(lipgloss.Color(currentTheme.PrimaryText))
	}

	return style.Render(constrainedValue)
}

// CellConstraint represents the constraints for a cell (from old codebase concepts)
type CellConstraint struct {
	Width     int
	Height    int // Currently only supports Height=1
	Alignment int // Use alignment constants
}

// properDisplayWidth calculates the correct display width of a string (from old codebase)
// This function combines lipgloss (for ANSI code handling) and go-runewidth (for proper Unicode width)
func properDisplayWidth(text string) int {
	// First, let lipgloss strip ANSI escape codes
	stripped := lipgloss.NewStyle().Render(text)
	// Then use go-runewidth for proper Unicode character width calculation
	return runewidth.StringWidth(stripped)
}

// enforceCellConstraints ensures text fits exactly within cell constraints (from old codebase)
func enforceCellConstraints(text string, constraint CellConstraint) string {
	// Handle multi-line content by converting to single line
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	// Collapse multiple spaces
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}
	text = strings.TrimSpace(text)

	// Get actual display width using proper Unicode calculation
	actualWidth := properDisplayWidth(text)
	targetWidth := constraint.Width

	// If text is too long, truncate it
	if actualWidth > targetWidth {
		if targetWidth <= 0 {
			return ""
		}

		if targetWidth <= 3 {
			// For very small widths, just return dots
			return strings.Repeat(".", targetWidth)
		}

		// Calculate how much we need to remove
		excessWidth := actualWidth - targetWidth

		// For small overflows (1-2 characters) or short target widths, use simple truncation
		// For longer text that would benefit from ellipsis indication, use ellipsis
		useEllipsis := targetWidth >= 6 && excessWidth >= 3

		if useEllipsis {
			// Try to fit with ellipsis
			truncated := text
			for properDisplayWidth(truncated+"...") > targetWidth && len(truncated) > 0 {
				runes := []rune(truncated)
				if len(runes) > 0 {
					truncated = string(runes[:len(runes)-1])
				} else {
					break
				}
			}

			if len(truncated) > 0 && properDisplayWidth(truncated+"...") <= targetWidth {
				text = truncated + "..."
			} else {
				// Fallback to simple truncation
				text = text
				for properDisplayWidth(text) > targetWidth && len(text) > 0 {
					runes := []rune(text)
					if len(runes) > 0 {
						text = string(runes[:len(runes)-1])
					} else {
						break
					}
				}
			}
		} else {
			// Simple truncation - just remove characters until we fit
			for properDisplayWidth(text) > targetWidth && len(text) > 0 {
				runes := []rune(text)
				if len(runes) > 0 {
					text = string(runes[:len(runes)-1])
				} else {
					break
				}
			}
		}

		actualWidth = properDisplayWidth(text)
	}

	// If text is shorter than target, add padding based on alignment
	if actualWidth < targetWidth {
		padding := targetWidth - actualWidth

		switch constraint.Alignment {
		case vtable.AlignCenter:
			leftPad := padding / 2
			rightPad := padding - leftPad
			text = strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
		case vtable.AlignRight:
			text = strings.Repeat(" ", padding) + text
		default: // AlignLeft
			text = text + strings.Repeat(" ", padding)
		}
	}

	return text
}

// ValueCellFormatter formats value cells with colors and row-level styling
func ValueCellFormatter(cellValue string, rowIndex, columnIndex int, column vtable.TableColumn, ctx vtable.RenderContext, isCursor, isSelected, isTopThreshold, isBottomThreshold bool) string {
	// Parse value for color coding but will be overridden by row styling if needed
	var baseStyle lipgloss.Style
	if strings.HasPrefix(cellValue, "Value ") {
		valueStr := strings.TrimPrefix(cellValue, "Value ")
		if value, err := strconv.Atoi(valueStr); err == nil {
			switch {
			case value < 30:
				baseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.ErrorText))
			case value < 70:
				baseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.WarningText))
			default:
				baseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.SuccessText))
			}
		}
	} else {
		baseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.PrimaryText))
	}

	// Apply row-level styling (overrides base styling for selections)
	var style lipgloss.Style
	if isCursor && isSelected {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color(currentTheme.SelectionBg)).
			Foreground(lipgloss.Color(currentTheme.PrimaryText)).
			Bold(true)
	} else if isCursor {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color(currentTheme.CursorBg)).
			Foreground(lipgloss.Color(currentTheme.PrimaryText)).
			Bold(true)
	} else if isSelected {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color(currentTheme.SelectionBg)).
			Foreground(lipgloss.Color(currentTheme.PrimaryText))
	} else {
		// Use base color coding for normal rows
		style = baseStyle
	}

	// Apply cell constraints
	constraint := CellConstraint{
		Width:     column.Width,
		Height:    1,
		Alignment: vtable.AlignRight, // Value column: right aligned
	}
	constrainedValue := enforceCellConstraints(cellValue, constraint)

	return style.Render(constrainedValue)
}

// StatusCellFormatter formats status cells with icons and row-level styling
func StatusCellFormatter(cellValue string, rowIndex, columnIndex int, column vtable.TableColumn, ctx vtable.RenderContext, isCursor, isSelected, isTopThreshold, isBottomThreshold bool) string {
	// Convert status to visual representation
	var statusText string
	var baseStyle lipgloss.Style

	switch cellValue {
	case "Status 0":
		baseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.SuccessText))
		statusText = currentTheme.ActiveIcon + " Active"
	case "Status 1":
		baseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.WarningText))
		statusText = currentTheme.WarningIcon + " Warning"
	case "Status 2":
		baseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.ErrorText))
		statusText = currentTheme.ErrorIcon + " Error"
	default:
		baseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.SecondaryText))
		statusText = currentTheme.UnknownIcon + " Unknown"
	}

	// Apply row-level styling (overrides base styling for selections)
	var style lipgloss.Style
	if isCursor && isSelected {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color(currentTheme.SelectionBg)).
			Foreground(lipgloss.Color(currentTheme.PrimaryText)).
			Bold(true)
	} else if isCursor {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color(currentTheme.CursorBg)).
			Foreground(lipgloss.Color(currentTheme.PrimaryText)).
			Bold(true)
	} else if isSelected {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color(currentTheme.SelectionBg)).
			Foreground(lipgloss.Color(currentTheme.PrimaryText))
	} else {
		// Use base color coding for normal rows
		style = baseStyle
	}

	// Apply cell constraints
	constraint := CellConstraint{
		Width:     column.Width,
		Height:    1,
		Alignment: vtable.AlignCenter, // Status column: center aligned
	}
	constrainedValue := enforceCellConstraints(statusText, constraint)

	return style.Render(constrainedValue)
}

// CategoryCellFormatter formats category cells with row-level styling
func CategoryCellFormatter(cellValue string, rowIndex, columnIndex int, column vtable.TableColumn, ctx vtable.RenderContext, isCursor, isSelected, isTopThreshold, isBottomThreshold bool) string {
	// Color code categories for normal rows
	var baseStyle lipgloss.Style
	switch cellValue {
	case "Category A":
		baseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.AccentText))
	case "Category B":
		baseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.SuccessText))
	case "Category C":
		baseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.WarningText))
	case "Category D":
		baseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.ErrorText))
	case "Category E":
		baseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.PrimaryText))
	default:
		baseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.SecondaryText))
	}

	// Apply row-level styling (overrides base styling for selections)
	var style lipgloss.Style
	if isCursor && isSelected {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color(currentTheme.SelectionBg)).
			Foreground(lipgloss.Color(currentTheme.PrimaryText)).
			Bold(true)
	} else if isCursor {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color(currentTheme.CursorBg)).
			Foreground(lipgloss.Color(currentTheme.PrimaryText)).
			Bold(true)
	} else if isSelected {
		style = lipgloss.NewStyle().
			Background(lipgloss.Color(currentTheme.SelectionBg)).
			Foreground(lipgloss.Color(currentTheme.PrimaryText))
	} else {
		// Use base color coding for normal rows
		style = baseStyle
	}

	// Apply cell constraints
	constraint := CellConstraint{
		Width:     column.Width,
		Height:    1,
		Alignment: vtable.AlignLeft, // Category column: left aligned
	}
	constrainedValue := enforceCellConstraints(cellValue, constraint)

	return style.Render(constrainedValue)
}

// createCustomLoadingFormatter creates a custom loading row formatter
func createCustomLoadingFormatter() vtable.LoadingRowFormatter {
	return func(index int, columns []vtable.TableColumn, ctx vtable.RenderContext, isCursor bool) string {
		var parts []string

		// Create custom loading cells for each column
		for i, col := range columns {
			var loadingText string
			var style lipgloss.Style

			// Different loading text for each column
			switch i {
			case 0: // Name column
				loadingText = "⏳ Loading..."
			case 1: // Value column
				loadingText = "..."
			case 2: // Status column
				loadingText = "⚡ Loading"
			case 3: // Category column
				loadingText = "•••"
			default:
				loadingText = "..."
			}

			// Apply cell constraints
			constraint := CellConstraint{
				Width:     col.Width,
				Height:    1,
				Alignment: col.Alignment,
			}
			constrainedText := enforceCellConstraints(loadingText, constraint)

			// Apply cursor or loading styling
			if isCursor {
				style = lipgloss.NewStyle().
					Background(lipgloss.Color(currentTheme.CursorBg)).
					Foreground(lipgloss.Color(currentTheme.PrimaryText)).
					Bold(true)
			} else {
				style = lipgloss.NewStyle().
					Foreground(lipgloss.Color(currentTheme.SecondaryText)).
					Italic(true)
			}

			parts = append(parts, style.Render(constrainedText))
		}

		// Join with borders if enabled
		result := strings.Join(parts, "│")
		result = "│" + result + "│"

		return result
	}
}

// createCustomHeaderCellFormatter creates a custom header cell formatter
func createCustomHeaderCellFormatter() vtable.HeaderCellFormatter {
	return func(column vtable.TableColumn, columnIndex int, ctx vtable.RenderContext) string {
		var headerText string
		var style lipgloss.Style

		// Different styling for each column
		switch columnIndex {
		case 0: // Name column
			headerText = "📝 " + column.Title
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("14")). // Cyan
				Bold(true).
				Background(lipgloss.Color("8"))

		case 1: // Value column
			headerText = "💰 " + column.Title
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("11")). // Yellow
				Bold(true).
				Background(lipgloss.Color("8"))

		case 2: // Status column
			headerText = "📊 " + column.Title
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")). // Green
				Bold(true).
				Background(lipgloss.Color("8"))

		case 3: // Category column
			headerText = "🏷️ " + column.Title
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("13")). // Magenta
				Bold(true).
				Background(lipgloss.Color("8"))

		default:
			headerText = column.Title
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Bold(true).
				Background(lipgloss.Color("8"))
		}

		// Apply constraint - use the column's HeaderConstraint if specified
		var constraint CellConstraint
		if column.HeaderConstraint.Width > 0 {
			// Convert vtable.CellConstraint to local CellConstraint
			constraint = CellConstraint{
				Width:     column.HeaderConstraint.Width,
				Alignment: column.HeaderConstraint.Alignment,
				Height:    column.HeaderConstraint.Height,
			}
		} else {
			// Use column alignment and width as fallback
			alignment := column.HeaderAlignment
			if alignment == 0 {
				alignment = column.Alignment
			}
			constraint = CellConstraint{
				Width:     column.Width,
				Alignment: alignment,
				Height:    1,
			}
		}

		constrainedText := enforceCellConstraints(headerText, constraint)
		return style.Render(constrainedText)
	}
}

// ================================
// MAIN INTERACTIVE APPLICATION
// ================================

// AppModel wraps our table for the Tea application
type AppModel struct {
	table             *vtable.Table
	dataSource        *ExampleTableDataSource
	showDebug         bool
	showHelp          bool
	statusMessage     string
	indexInput        string
	inputMode         bool // true when entering a number for JumpToIndex
	currentThemeIndex int  // Track current theme
	// Add chunk loading observability
	loadingChunks map[int]bool
	chunkHistory  []string
}

// Available themes list
var availableThemes = []TableTheme{
	DefaultTheme,
	DarkTheme,
	MinimalTheme,
	NeonTheme,
}

func main() {
	// Create data source
	dataSource := NewExampleTableDataSource(1000)

	// Create table
	table := CreateExampleTableWithDataSource(dataSource)

	// Create app model
	app := AppModel{
		table:             table,
		dataSource:        dataSource,
		showDebug:         true,
		showHelp:          true,
		statusMessage:     "Welcome! Use arrow keys to navigate, space to select, ? to toggle help",
		indexInput:        "",
		inputMode:         false,
		currentThemeIndex: 0, // Start with DefaultTheme
		loadingChunks:     make(map[int]bool),
		chunkHistory:      make([]string, 0),
	}

	// Run the interactive program
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.table.Init(),
		m.table.Focus(),
		// Set cell formatters through the Tea model loop
		vtable.CellFormatterSetCmd(0, NameCellFormatter),     // Name column with selection indicators
		vtable.CellFormatterSetCmd(1, ValueCellFormatter),    // Value column
		vtable.CellFormatterSetCmd(2, StatusCellFormatter),   // Status column
		vtable.CellFormatterSetCmd(3, CategoryCellFormatter), // Category column
		// Set custom loading formatter
		vtable.LoadingFormatterSetCmd(createCustomLoadingFormatter()),
		// Set custom header cell formatter
		vtable.HeaderCellFormatterSetCmd(createCustomHeaderCellFormatter()),
	)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle input mode for JumpToIndex
		if m.inputMode {
			switch msg.String() {
			case "enter":
				// Parse the input and jump to index
				if index, err := strconv.Atoi(m.indexInput); err == nil && index >= 0 && index < 1000 {
					m.inputMode = false
					m.indexInput = ""
					m.statusMessage = fmt.Sprintf("Jumping to index %d", index)
					return m, vtable.JumpToCmd(index)
				} else {
					m.statusMessage = "Invalid index! Please enter a number between 0-999"
					m.inputMode = false
					m.indexInput = ""
					return m, nil
				}
			case "escape":
				m.inputMode = false
				m.indexInput = ""
				m.statusMessage = "Jump cancelled"
				return m, nil
			case "backspace":
				if len(m.indexInput) > 0 {
					m.indexInput = m.indexInput[:len(m.indexInput)-1]
				}
				return m, nil
			default:
				// Only allow digits
				if len(msg.String()) == 1 && msg.String() >= "0" && msg.String() <= "9" {
					if len(m.indexInput) < 4 { // Limit to 4 digits (0-999)
						m.indexInput += msg.String()
					}
				}
				return m, nil
			}
		}

		// Normal key handling
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "?":
			m.showHelp = !m.showHelp
			if m.showHelp {
				m.statusMessage = "Help visible - press ? to hide"
			} else {
				m.statusMessage = "Help hidden - press ? to show"
			}
			return m, nil

		case "d":
			m.showDebug = !m.showDebug
			if m.showDebug {
				m.statusMessage = "Debug mode ON"
			} else {
				m.statusMessage = "Debug mode OFF"
			}
			return m, nil

		case "t":
			// Cycle through themes
			m.currentThemeIndex = (m.currentThemeIndex + 1) % len(availableThemes)
			newTheme := availableThemes[m.currentThemeIndex]
			SetTheme(newTheme)
			m.statusMessage = fmt.Sprintf("Theme changed to: %s - %s", newTheme.Name, newTheme.Description)
			return m, nil

		case "r":
			m.statusMessage = "Refreshing data..."
			return m, vtable.DataRefreshCmd()

		// === NAVIGATION KEYS ===
		case "g":
			// Jump to start (like vim)
			m.statusMessage = "Jumping to start"
			return m, vtable.JumpToStartCmd()

		case "G":
			// Jump to end (like vim)
			m.statusMessage = "Jumping to end"
			return m, vtable.JumpToEndCmd()

		case "J":
			// Enter jump-to-index mode (uppercase J)
			m.inputMode = true
			m.indexInput = ""
			m.statusMessage = "Enter index to jump to (0-999): "
			return m, nil

		case "h":
			// Page up using proper command
			m.statusMessage = "Page up"
			return m, vtable.PageUpCmd()

		case "l":
			// Page down using proper command
			m.statusMessage = "Page down"
			return m, vtable.PageDownCmd()

		case "j", "up":
			// Move up using proper command
			return m, vtable.CursorUpCmd()

		case "k", "down":
			// Move down using proper command
			return m, vtable.CursorDownCmd()

		// === SELECTION KEYS ===
		case " ":
			return m, vtable.SelectCurrentCmd()

		case "a":
			return m, vtable.SelectAllCmd()

		case "c":
			return m, vtable.SelectClearCmd()

		case "s":
			selectionCount := m.dataSource.GetSelectionCount()
			if selectionCount > 0 {
				m.statusMessage = fmt.Sprintf("SELECTED: %d items total (look for background highlighting)", selectionCount)
			} else {
				m.statusMessage = "No items selected - use Space to select items"
			}
			return m, nil

		// === QUICK JUMP SHORTCUTS ===
		case "1":
			m.statusMessage = "Quick jump to index 100"
			return m, vtable.JumpToCmd(100)

		case "2":
			m.statusMessage = "Quick jump to index 250"
			return m, vtable.JumpToCmd(250)

		case "3":
			m.statusMessage = "Quick jump to index 500"
			return m, vtable.JumpToCmd(500)

		case "4":
			m.statusMessage = "Quick jump to index 750"
			return m, vtable.JumpToCmd(750)

		case "5":
			m.statusMessage = "Quick jump to index 900"
			return m, vtable.JumpToCmd(900)

		default:
			// Pass other keys to table
			var cmd tea.Cmd
			_, cmd = m.table.Update(msg)

			// Update status with current position
			state := m.table.GetState()
			m.statusMessage = fmt.Sprintf("Position: %d/%d (Viewport: %d-%d)",
				state.CursorIndex+1, m.table.GetTotalItems(),
				state.ViewportStartIndex,
				state.ViewportStartIndex+9) // viewport height is 10

			return m, cmd
		}

	case vtable.SelectionResponseMsg:
		// Update status based on selection
		selectionCount := m.dataSource.GetSelectionCount()
		state := m.table.GetState()
		if msg.Success {
			switch msg.Operation {
			case "toggle":
				m.statusMessage = fmt.Sprintf("Selected item at index %d - %d total selected (look for background highlighting)", state.CursorIndex, selectionCount)
			case "selectAll":
				m.statusMessage = fmt.Sprintf("Selected ALL %d items in datasource (look for background highlighting!)", selectionCount)
			case "clear":
				m.statusMessage = "All selections cleared - indicators removed"
			}
		} else {
			m.statusMessage = fmt.Sprintf("Selection failed: %v", msg.Error)
		}

		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd

	// Handle navigation messages to update status
	case vtable.PageUpMsg:
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		state := m.table.GetState()
		m.statusMessage = fmt.Sprintf("Page up - now at index %d", state.CursorIndex)
		return m, cmd

	case vtable.PageDownMsg:
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		state := m.table.GetState()
		m.statusMessage = fmt.Sprintf("Page down - now at index %d", state.CursorIndex)
		return m, cmd

	case vtable.JumpToMsg:
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		state := m.table.GetState()
		m.statusMessage = fmt.Sprintf("Jumped to index %d", state.CursorIndex)
		return m, cmd

	case vtable.JumpToStartMsg:
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		m.statusMessage = "Jumped to start"
		return m, cmd

	case vtable.JumpToEndMsg:
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		m.statusMessage = "Jumped to end"
		return m, cmd

	case vtable.CursorUpMsg, vtable.CursorDownMsg:
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		state := m.table.GetState()
		m.statusMessage = fmt.Sprintf("Position: %d/%d (Viewport: %d-%d)",
			state.CursorIndex+1, m.table.GetTotalItems(),
			state.ViewportStartIndex,
			state.ViewportStartIndex+9)
		return m, cmd

	// Handle chunk loading observability messages
	case vtable.ChunkLoadingStartedMsg:
		m.loadingChunks[msg.ChunkStart] = true
		historyEntry := fmt.Sprintf("Started loading chunk %d (size: %d)", msg.ChunkStart, msg.Request.Count)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		// Also pass to table
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd

	case vtable.ChunkLoadingCompletedMsg:
		delete(m.loadingChunks, msg.ChunkStart)
		historyEntry := fmt.Sprintf("Completed chunk %d (%d items)", msg.ChunkStart, msg.ItemCount)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		// Also pass to table
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd

	case vtable.ChunkUnloadedMsg:
		historyEntry := fmt.Sprintf("Unloaded chunk %d", msg.ChunkStart)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		// Also pass to table
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd

	case vtable.DataChunkLoadedMsg:
		// This is the actual chunk data loading completion
		historyEntry := fmt.Sprintf("Loaded chunk %d (%d items)", msg.StartIndex, len(msg.Items))
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		// Also pass to table
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd

	case vtable.DataTotalMsg:
		historyEntry := fmt.Sprintf("Total items: %d", msg.Total)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		// Also pass to table
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd

	case vtable.DataRefreshMsg:
		historyEntry := "Data refresh triggered"
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		// Also pass to table
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd

	case vtable.DataLoadErrorMsg:
		historyEntry := fmt.Sprintf("Data load error: %v", msg.Error)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		// Also pass to table
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd

	default:
		// Pass all other messages to table
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd
	}
}

func (m AppModel) View() string {
	var view strings.Builder

	// Show help if enabled
	if m.showHelp {
		view.WriteString("=== TABLE DEMO ===\n")
		view.WriteString("Visual Indicators: row highlighting = cursor • background color = selected\n")
		view.WriteString("Header Features: independent alignment per column • custom cell constraints • emoji icons\n")
		view.WriteString("Navigation: j/k or ↑/↓ move • h/l page up/down • g=start • G=end • J=jump to index • 1-5=quick jumps\n")
		view.WriteString("Selection: Space=toggle • a=select all • c=clear • s=show count\n")
		view.WriteString("Themes: t=cycle themes (Default→Dark→Minimal→Neon)\n")
		view.WriteString("Debug: d=toggle debug (shows chunk loading activity)\n")
		view.WriteString("Other: r=refresh • ?=help • q=quit\n")
		view.WriteString(fmt.Sprintf("Current Theme: %s - %s\n\n", currentTheme.Name, currentTheme.Description))
	}

	// Show status message or input prompt
	if m.inputMode {
		view.WriteString(fmt.Sprintf("%s%s_", m.statusMessage, m.indexInput))
	} else {
		view.WriteString(m.statusMessage)
	}
	view.WriteString("\n\n")

	// Show table
	view.WriteString(m.table.View())

	// Show selection info
	selectionCount := m.dataSource.GetSelectionCount() // Use DataSource count!
	if selectionCount > 0 {
		view.WriteString(fmt.Sprintf("\n\nSelected: %d items", selectionCount))
	}

	// Show recent activity
	recentActivity := m.dataSource.GetRecentActivity()
	if len(recentActivity) > 0 {
		view.WriteString("\n\nRecent Activity:")
		for i := len(recentActivity) - 1; i >= 0 && i >= len(recentActivity)-5; i-- {
			view.WriteString(fmt.Sprintf("\n  • %s", recentActivity[i]))
		}
	}

	// Show debug info if enabled
	if m.showDebug {
		state := m.table.GetState()
		view.WriteString(fmt.Sprintf("\n\nDEBUG: Cursor=%d, Viewport=%d-%d, Total=%d",
			state.CursorIndex,
			state.ViewportStartIndex,
			state.ViewportStartIndex+9, // viewport height is 10
			m.table.GetTotalItems()))

		// Show threshold flags
		view.WriteString(fmt.Sprintf("\nThresholds: top=%v, bottom=%v",
			state.IsAtTopThreshold, state.IsAtBottomThreshold))

		// Show currently loading chunks
		if len(m.loadingChunks) > 0 {
			view.WriteString("\nLoading chunks: ")
			var chunks []string
			for chunk := range m.loadingChunks {
				chunks = append(chunks, fmt.Sprintf("%d", chunk))
			}
			view.WriteString(strings.Join(chunks, ", "))
		}

		// Show recent chunk history
		if len(m.chunkHistory) > 0 {
			view.WriteString("\nRecent chunk activity:")
			for _, entry := range m.chunkHistory {
				view.WriteString(fmt.Sprintf("\n  • %s", entry))
			}
		}

		if len(m.loadingChunks) == 0 && len(m.chunkHistory) == 0 {
			view.WriteString("\nNo chunk activity yet - navigate around to see chunking!")
		}
	}

	return view.String()
}

// CreateExampleTableWithDataSource creates a table with the given data source
func CreateExampleTableWithDataSource(dataSource *ExampleTableDataSource) *vtable.Table {
	// Create table configuration with wider columns and custom header alignments
	columns := []vtable.TableColumn{
		{
			Title:           "Name",
			Field:           "name",
			Width:           25,
			Alignment:       vtable.AlignLeft,   // Cell content: left aligned
			HeaderAlignment: vtable.AlignCenter, // Header: center aligned
			HeaderConstraint: vtable.CellConstraint{
				Width:     25,
				Alignment: vtable.AlignCenter,
				Padding:   vtable.PaddingConfig{Left: 1, Right: 1},
			},
		},
		{
			Title:           "Value",
			Field:           "value",
			Width:           15,
			Alignment:       vtable.AlignRight, // Cell content: right aligned
			HeaderAlignment: vtable.AlignLeft,  // Header: left aligned (different!)
			HeaderConstraint: vtable.CellConstraint{
				Width:     15,
				Alignment: vtable.AlignLeft,
			},
		},
		{
			Title:           "Status",
			Field:           "status",
			Width:           18,
			Alignment:       vtable.AlignCenter, // Cell content: center aligned
			HeaderAlignment: vtable.AlignRight,  // Header: right aligned (different!)
			HeaderConstraint: vtable.CellConstraint{
				Width:     18,
				Alignment: vtable.AlignRight,
			},
		},
		{
			Title:           "Category",
			Field:           "category",
			Width:           20,
			Alignment:       vtable.AlignLeft,   // Cell content: left aligned
			HeaderAlignment: vtable.AlignCenter, // Header: center aligned
			HeaderConstraint: vtable.CellConstraint{
				Width:     20,
				Alignment: vtable.AlignCenter,
				Padding:   vtable.PaddingConfig{Left: 2, Right: 2},
			},
		},
	}

	config := vtable.TableConfig{
		Columns:     columns,
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: vtable.ViewportConfig{
			Height:             10,
			TopThreshold:       2,
			BottomThreshold:    2,
			ChunkSize:          50,
			InitialIndex:       0,
			BoundingAreaBefore: 25,
			BoundingAreaAfter:  25,
		},
		Theme: vtable.Theme{
			HeaderStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("8")),
			CellStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("15")),
			CursorStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("12")),
			SelectedStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("10")),
			BorderChars: vtable.BorderChars{
				Horizontal:  "─",
				Vertical:    "│",
				TopLeft:     "┌",
				TopRight:    "┐",
				BottomLeft:  "└",
				BottomRight: "┘",
				TopT:        "┬",
				BottomT:     "┴",
				LeftT:       "├",
				RightT:      "┤",
				Cross:       "┼",
			},
			BorderColor: "8",
			HeaderColor: "15",
		},
		SelectionMode: vtable.SelectionMultiple,
		KeyMap: vtable.NavigationKeyMap{
			Up:        []string{"up", "k"},
			Down:      []string{"down", "j"},
			PageUp:    []string{"pgup", "ctrl+u"},
			PageDown:  []string{"pgdown", "ctrl+d"},
			Home:      []string{"home", "g"},
			End:       []string{"end", "G"},
			Select:    []string{"enter", " "},
			SelectAll: []string{"ctrl+a"},
			Filter:    []string{"/"},
			Sort:      []string{"s"},
			Quit:      []string{"q", "esc"},
		},
	}

	// Create table with data source
	table := vtable.NewTable(config, dataSource)

	return table
}
