package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/table"
)

// ================================
// EXAMPLE DATA SOURCE
// ================================

// ExampleTableDataSource provides sample data for the table
type ExampleTableDataSource struct {
	totalItems     int
	data           []core.TableRow
	selectedItems  map[string]bool // Actually store selection state!
	recentActivity []string        // Track recent selection activity
	// Add sorting and filtering state
	sortFields    []string
	sortDirs      []string
	filters       map[string]any
	filteredData  []core.TableRow // Cached filtered/sorted data
	filteredTotal int             // Total after filtering
}

// NewExampleTableDataSource creates a data source with sample table data
func NewExampleTableDataSource(totalItems int) *ExampleTableDataSource {
	data := make([]core.TableRow, totalItems)

	// Array of long descriptions to demonstrate text wrapping
	longDescriptions := []string{
		"This is a very long description that will definitely exceed the column width and demonstrate the text wrapping functionality using runewidth for proper Unicode handling",
		"Another extremely lengthy description with multiple words that should be intelligently truncated with ellipsis when the wrapping mode is enabled in the table formatter",
		"A comprehensive explanation of the item's purpose, functionality, and various attributes that would normally overflow the cell boundaries without proper text constraint handling",
		"Detailed documentation about this particular entry including its specifications, usage guidelines, and important notes for users who need complete information",
		"An extensive narrative describing the item's history, development process, and future roadmap that serves as a perfect example of text that needs wrapping or truncation",
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua ut enim ad minim veniam quis nostrud",
		"Professional description with technical specifications, performance metrics, compatibility information, and detailed instructions for optimal usage in various scenarios",
		"Complete product overview including features, benefits, limitations, system requirements, installation procedures, and troubleshooting guidelines for end users",
	}

	for i := 0; i < totalItems; i++ {
		// Use modulo to cycle through descriptions
		description := longDescriptions[i%len(longDescriptions)]

		data[i] = core.TableRow{
			ID: fmt.Sprintf("row-%d", i),
			Cells: []string{
				fmt.Sprintf("Item %d", i+1),
				fmt.Sprintf("Value %d", (i*37)%100),
				fmt.Sprintf("Status %d", i%3),
				fmt.Sprintf("Category %c", 'A'+(i%5)),
				description, // Add the long description as 5th column
			},
		}
	}

	return &ExampleTableDataSource{
		totalItems:     totalItems,
		data:           data,
		selectedItems:  make(map[string]bool), // Initialize selection state
		recentActivity: make([]string, 0),     // Initialize activity log
		sortFields:     []string{},
		sortDirs:       []string{},
		filters:        make(map[string]any),
		filteredData:   data, // Start with all data
		filteredTotal:  totalItems,
	}
}

// GetTotal returns the total number of items
func (ds *ExampleTableDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return core.DataTotalMsg{Total: ds.filteredTotal}
	}
}

// RefreshTotal refreshes the total count
func (ds *ExampleTableDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

// LoadChunk loads a chunk of data
func (ds *ExampleTableDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		// Update sorting and filtering if changed in request
		if len(request.SortFields) > 0 || len(request.Filters) > 0 {
			// Apply new sort/filter settings
			if len(request.SortFields) > 0 {
				ds.sortFields = request.SortFields
				ds.sortDirs = request.SortDirections
			}
			if len(request.Filters) > 0 {
				ds.filters = request.Filters
			}
			ds.rebuildFilteredData()
		}

		// Simulate loading delay
		time.Sleep(10 * time.Millisecond)

		start := request.Start
		end := start + request.Count
		if end > ds.filteredTotal {
			end = ds.filteredTotal
		}

		var items []core.Data[any]
		for i := start; i < end; i++ {
			if i < len(ds.filteredData) {
				items = append(items, core.Data[any]{
					ID:       ds.filteredData[i].ID,
					Item:     ds.filteredData[i],
					Selected: ds.selectedItems[ds.filteredData[i].ID],
					Metadata: core.NewTypedMetadata(),
				})
			}
		}

		return core.DataChunkLoadedMsg{
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

			return core.SelectionResponseMsg{
				Success:   true,
				Index:     index,
				ID:        id,
				Selected:  selected,
				Operation: "toggle",
			}
		}

		return core.SelectionResponseMsg{
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

				return core.SelectionResponseMsg{
					Success:   true,
					Index:     i,
					ID:        id,
					Selected:  selected,
					Operation: "toggle",
				}
			}
		}

		return core.SelectionResponseMsg{
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

		return core.SelectionResponseMsg{
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

		return core.SelectionResponseMsg{
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

		return core.SelectionResponseMsg{
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
	if row, ok := item.(core.TableRow); ok {
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
// SORTING AND FILTERING METHODS
// ================================

// SetSort sets the sorting fields and directions and rebuilds the data cache
func (ds *ExampleTableDataSource) SetSort(fields []string, directions []string) {
	ds.sortFields = fields
	ds.sortDirs = directions
	ds.rebuildFilteredData()
}

// SetFilter sets a filter and rebuilds the data cache
func (ds *ExampleTableDataSource) SetFilter(field string, value any) {
	ds.filters[field] = value
	ds.rebuildFilteredData()
}

// ClearFilter removes a filter and rebuilds the data cache
func (ds *ExampleTableDataSource) ClearFilter(field string) {
	delete(ds.filters, field)
	ds.rebuildFilteredData()
}

// ClearAllFilters removes all filters and rebuilds the data cache
func (ds *ExampleTableDataSource) ClearAllFilters() {
	ds.filters = make(map[string]any)
	ds.rebuildFilteredData()
}

// rebuildFilteredData applies current filters and sorting to rebuild the data cache
func (ds *ExampleTableDataSource) rebuildFilteredData() {
	// Start with all data
	result := make([]core.TableRow, 0, len(ds.data))

	// Apply filters
	for _, row := range ds.data {
		include := true

		for field, filterValue := range ds.filters {
			switch field {
			case "category":
				if filterStr, ok := filterValue.(string); ok {
					if len(row.Cells) > 3 && row.Cells[3] != filterStr {
						include = false
						break
					}
				}
			case "value":
				if filterStr, ok := filterValue.(string); ok && filterStr == "high" {
					// Filter for high values (>50)
					if len(row.Cells) > 1 {
						if valueStr := strings.TrimPrefix(row.Cells[1], "Value "); valueStr != row.Cells[1] {
							if value, err := strconv.Atoi(valueStr); err == nil && value <= 50 {
								include = false
								break
							}
						}
					}
				}
			case "status":
				if filterStr, ok := filterValue.(string); ok && filterStr == "active" {
					// Filter for active status (Status 0)
					if len(row.Cells) > 2 && row.Cells[2] != "Status 0" {
						include = false
						break
					}
				}
			}
		}

		if include {
			result = append(result, row)
		}
	}

	// Apply sorting
	if len(ds.sortFields) > 0 {
		sort.Slice(result, func(i, j int) bool {
			for idx, field := range ds.sortFields {
				dir := "asc"
				if idx < len(ds.sortDirs) {
					dir = ds.sortDirs[idx]
				}

				var cellI, cellJ string
				switch field {
				case "name":
					if len(result[i].Cells) > 0 {
						cellI = result[i].Cells[0]
					}
					if len(result[j].Cells) > 0 {
						cellJ = result[j].Cells[0]
					}
				case "value":
					if len(result[i].Cells) > 1 {
						cellI = result[i].Cells[1]
					}
					if len(result[j].Cells) > 1 {
						cellJ = result[j].Cells[1]
					}
				case "status":
					if len(result[i].Cells) > 2 {
						cellI = result[i].Cells[2]
					}
					if len(result[j].Cells) > 2 {
						cellJ = result[j].Cells[2]
					}
				case "category":
					if len(result[i].Cells) > 3 {
						cellI = result[i].Cells[3]
					}
					if len(result[j].Cells) > 3 {
						cellJ = result[j].Cells[3]
					}
				}

				var cmp int
				// For value field, do numeric comparison
				if field == "value" {
					valueI := extractValueNumber(cellI)
					valueJ := extractValueNumber(cellJ)
					if valueI < valueJ {
						cmp = -1
					} else if valueI > valueJ {
						cmp = 1
					}
				} else {
					// String comparison for other fields
					if cellI < cellJ {
						cmp = -1
					} else if cellI > cellJ {
						cmp = 1
					}
				}

				if cmp != 0 {
					if dir == "desc" {
						return cmp > 0
					}
					return cmp < 0
				}
			}
			return false
		})
	}

	ds.filteredData = result
	ds.filteredTotal = len(result)
}

// extractValueNumber extracts the numeric value from "Value X" strings
func extractValueNumber(valueStr string) int {
	if strings.HasPrefix(valueStr, "Value ") {
		if numStr := strings.TrimPrefix(valueStr, "Value "); numStr != valueStr {
			if num, err := strconv.Atoi(numStr); err == nil {
				return num
			}
		}
	}
	return 0
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
		Description:   "Clean Unicode box drawing theme",
		CursorBg:      "12", // Blue
		SelectionBg:   "10", // Green
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
		Name:          "Heavy",
		Description:   "Heavy double-line borders theme",
		CursorBg:      "22",  // Dark green
		SelectionBg:   "235", // Dark gray
		HeaderBg:      "0",   // Black
		BorderColor:   "8",   // Gray
		PrimaryText:   "15",  // White
		SecondaryText: "7",   // Light gray
		AccentText:    "10",  // Green
		ErrorText:     "9",   // Red
		WarningText:   "11",  // Yellow
		SuccessText:   "10",  // Green
		ActiveIcon:    "●",
		WarningIcon:   "▲",
		ErrorIcon:     "■",
		UnknownIcon:   "○",
	}

	MinimalTheme = TableTheme{
		Name:          "Minimal",
		Description:   "Clean borderless minimalist theme",
		CursorBg:      "7",   // Light gray
		SelectionBg:   "235", // Dark gray
		HeaderBg:      "0",   // Black
		BorderColor:   "8",   // Gray
		PrimaryText:   "0",   // Black
		SecondaryText: "8",   // Gray
		AccentText:    "4",   // Blue
		ErrorText:     "1",   // Red
		WarningText:   "3",   // Yellow
		SuccessText:   "2",   // Green
		ActiveIcon:    "+",
		WarningIcon:   "!",
		ErrorIcon:     "x",
		UnknownIcon:   "?",
	}

	NeonTheme = TableTheme{
		Name:          "Retro",
		Description:   "ASCII retro computing theme",
		CursorBg:      "201", // Bright magenta cursor
		SelectionBg:   "235", // Dark gray for subtle selection
		HeaderBg:      "0",   // Black
		BorderColor:   "14",  // Cyan borders
		PrimaryText:   "15",  // White
		SecondaryText: "8",   // Muted gray
		AccentText:    "13",  // Magenta
		ErrorText:     "9",   // Standard red
		WarningText:   "11",  // Standard yellow
		SuccessText:   "10",  // Standard green
		ActiveIcon:    "*",
		WarningIcon:   "!",
		ErrorIcon:     "X",
		UnknownIcon:   "?",
	}
)

// Current active theme
var currentTheme = DefaultTheme

// SetTheme changes the active theme
func SetTheme(theme TableTheme) {
	currentTheme = theme
}

// convertToVTableTheme converts demo TableTheme to vtable.Theme
func convertToVTableTheme(theme TableTheme) core.Theme {
	var borderChars core.BorderChars

	// Create dramatically different visual styles based on theme
	switch theme.Name {
	case "Default":
		// Unicode box drawing - clean and modern
		borderChars = core.BorderChars{
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
		}
	case "Heavy":
		// Double/thick lines - dramatic and bold
		borderChars = core.BorderChars{
			Horizontal:  "═",
			Vertical:    "║",
			TopLeft:     "╔",
			TopRight:    "╗",
			BottomLeft:  "╚",
			BottomRight: "╝",
			TopT:        "╦",
			BottomT:     "╩",
			LeftT:       "╠",
			RightT:      "╣",
			Cross:       "╬",
		}
	case "Minimal":
		// Simple lines - subtle and clean
		borderChars = core.BorderChars{
			Horizontal:  " ", // No horizontal borders!
			Vertical:    " ", // No vertical borders!
			TopLeft:     " ",
			TopRight:    " ",
			BottomLeft:  " ",
			BottomRight: " ",
			TopT:        " ",
			BottomT:     " ",
			LeftT:       " ",
			RightT:      " ",
			Cross:       " ",
		}
	case "Retro":
		// ASCII retro computing style
		borderChars = core.BorderChars{
			Horizontal:  "-",
			Vertical:    "|",
			TopLeft:     "+",
			TopRight:    "+",
			BottomLeft:  "+",
			BottomRight: "+",
			TopT:        "+",
			BottomT:     "+",
			LeftT:       "+",
			RightT:      "+",
			Cross:       "+",
		}
	default:
		// Fallback to default
		borderChars = core.DefaultBorderChars()
	}

	return core.Theme{
		HeaderStyle:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.PrimaryText)).Background(lipgloss.Color(theme.HeaderBg)),
		CellStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color(theme.PrimaryText)),
		CursorStyle:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.PrimaryText)).Background(lipgloss.Color(theme.CursorBg)),
		SelectedStyle:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color(theme.SelectionBg)), // White text on selection
		FullRowCursorStyle: lipgloss.NewStyle().Background(lipgloss.Color(theme.CursorBg)).Foreground(lipgloss.Color(theme.PrimaryText)).Bold(true),
		BorderChars:        borderChars, // Use the custom border chars!
		BorderColor:        theme.BorderColor,
		HeaderColor:        theme.PrimaryText,
	}
}

// ================================
// ENHANCED CELL FORMATTERS
// ================================

// NameCellFormatter formats the first column with proper selection/cursor handling
func NameCellFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.PrimaryText))

	// Apply selection background if this row is selected
	if isSelected {
		style = style.Background(lipgloss.Color(currentTheme.SelectionBg)).Foreground(lipgloss.Color("15")) // White text on selection background
	}

	return style.Render(cellValue)
}

// ValueCellFormatter formats value cells with colors and selection handling
func ValueCellFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
	// Parse value for color coding
	var style lipgloss.Style
	if strings.HasPrefix(cellValue, "Value ") {
		valueStr := strings.TrimPrefix(cellValue, "Value ")
		if value, err := strconv.Atoi(valueStr); err == nil {
			switch {
			case value < 30:
				style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.ErrorText))
			case value < 70:
				style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.WarningText))
			default:
				style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.SuccessText))
			}
		}
	} else {
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.PrimaryText))
	}

	// Apply selection background if this row is selected (overrides color coding)
	if isSelected {
		style = style.Background(lipgloss.Color(currentTheme.SelectionBg)).Foreground(lipgloss.Color("15")) // White text on selection background
	}

	return style.Render(cellValue)
}

// StatusCellFormatter formats status cells with icons, colors and selection handling
func StatusCellFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
	// Convert status to visual representation
	var statusText string
	var style lipgloss.Style

	switch cellValue {
	case "Status 0":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.SuccessText))
		statusText = currentTheme.ActiveIcon + " Active"
	case "Status 1":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.WarningText))
		statusText = currentTheme.WarningIcon + " Warning"
	case "Status 2":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.ErrorText))
		statusText = currentTheme.ErrorIcon + " Error"
	default:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.SecondaryText))
		statusText = currentTheme.UnknownIcon + " Unknown"
	}

	// Apply selection background if this row is selected (overrides color coding)
	if isSelected {
		style = style.Background(lipgloss.Color(currentTheme.SelectionBg)).Foreground(lipgloss.Color("15")) // White text on selection background
	}

	return style.Render(statusText)
}

// CategoryCellFormatter formats category cells with colors and selection handling
func CategoryCellFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
	// Color code categories
	var style lipgloss.Style
	switch cellValue {
	case "Category A":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.AccentText))
	case "Category B":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.SuccessText))
	case "Category C":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.WarningText))
	case "Category D":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.ErrorText))
	case "Category E":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.PrimaryText))
	default:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.SecondaryText))
	}

	// Apply selection background if this row is selected (overrides color coding)
	if isSelected {
		style = style.Background(lipgloss.Color(currentTheme.SelectionBg)).Foreground(lipgloss.Color("15")) // White text on selection background
	}

	return style.Render(cellValue)
}

// DescriptionCellFormatter formats the description column with proper text handling
func DescriptionCellFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.SecondaryText)).Italic(true)

	// Apply selection background if this row is selected
	if isSelected {
		style = style.Background(lipgloss.Color(currentTheme.SelectionBg)).Foreground(lipgloss.Color("15")) // White text on selection background
	}

	return style.Render(cellValue)
}

// createCustomHeaderFormatter creates a custom header formatter with styling
func createCustomHeaderFormatter() map[int]core.SimpleHeaderFormatter {
	formatters := make(map[int]core.SimpleHeaderFormatter)

	// Simple header formatters like the working test
	formatters[0] = func(column core.TableColumn, ctx core.RenderContext) string {
		return "📝 " + column.Title
	}

	formatters[1] = func(column core.TableColumn, ctx core.RenderContext) string {
		return "💰 " + column.Title
	}

	formatters[2] = func(column core.TableColumn, ctx core.RenderContext) string {
		return "📊 " + column.Title
	}

	formatters[3] = func(column core.TableColumn, ctx core.RenderContext) string {
		return "🏷️ " + column.Title
	}

	formatters[4] = func(column core.TableColumn, ctx core.RenderContext) string {
		return "📄 " + column.Title
	}

	return formatters
}

// ================================
// ADVANCED FORMATTER IMPLEMENTATIONS
// ================================

// createFullRowCursorFormatter creates a formatter that highlights the entire row when cursor is on it
func createFullRowCursorFormatter(baseFormatter core.SimpleCellFormatter) core.SimpleCellFormatter {
	return func(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
		// If this is the cursor row, apply enhanced highlighting to all cells in the row
		if isCursor {
			// Full row highlighting - bold with distinctive background
			// IMPORTANT: No padding here - let the table's applyCellConstraints handle padding extension
			fullRowStyle := lipgloss.NewStyle().
				Background(lipgloss.Color(currentTheme.CursorBg)).
				Foreground(lipgloss.Color("15")). // White text
				Bold(true).
				Underline(true) // Extra emphasis for full row mode

			// Don't add row indicator - just style the content strongly
			return fullRowStyle.Render(cellValue)
		}

		// For non-cursor rows, use the base formatter without extra styling
		return baseFormatter(cellValue, rowIndex, column, ctx, false, isSelected, isActiveCell)
	}
}

// createSelectionAwareFormatter creates a formatter that ensures selection highlighting is always applied
func createSelectionAwareFormatter(baseFormatter core.SimpleCellFormatter) core.SimpleCellFormatter {
	return func(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
		// Use the base formatter - DON'T force background styling
		// Let the table's built-in selection system handle backgrounds properly
		return baseFormatter(cellValue, rowIndex, column, ctx, isCursor, isSelected, isActiveCell)
	}
}

// ================================
// FORMATTER MANAGEMENT HELPERS
// ================================

// resetAllFormattersToBase resets all cell formatters to their base implementations
func resetAllFormattersToBase() tea.Cmd {
	return tea.Batch(
		core.CellFormatterSetCmd(0, NameCellFormatter),
		core.CellFormatterSetCmd(1, ValueCellFormatter),
		core.CellFormatterSetCmd(2, StatusCellFormatter),
		core.CellFormatterSetCmd(3, CategoryCellFormatter),
		core.CellFormatterSetCmd(4, DescriptionCellFormatter),
	)
}

// ================================
// MAIN INTERACTIVE APPLICATION
// ================================

// AppModel represents the state of our application
type AppModel struct {
	table                *table.Table
	dataSource           *ExampleTableDataSource
	statusMessage        string
	showHelp             bool
	showDebug            bool
	inputMode            bool
	indexInput           string
	fullRowCursorEnabled bool // Track if full row cursor highlighting is enabled
	selectionModeEnabled bool // Track if selection highlighting is enabled
	scrollResetEnabled   bool // Track if scroll reset on navigation is enabled
	loadingChunks        map[int]bool
	chunkHistory         []string

	// Theme cycling
	themeIndex int

	// Column ordering
	columnOrderIndex int

	// Sorting states
	sortingEnabled   bool
	currentSortField string
	currentSortDir   string

	// Filtering states
	filteringEnabled   bool
	currentFilter      string
	currentFilterField string

	// Border states
	showTopBorder       bool
	showBottomBorder    bool
	showHeaderSeparator bool
	removeTopSpace      bool
	removeBottomSpace   bool

	// Active cell indication demo state
	activeCellEnabled bool // Simple on/off toggle for active cell indication
}

// Available themes list
var availableThemes = []TableTheme{
	DefaultTheme,
	DarkTheme,
	MinimalTheme,
	NeonTheme,
}

// Available column orderings
var availableColumnOrders = [][]core.TableColumn{
	// Default order: Name, Value, Status, Category, Description
	{
		{
			Title:           "Name",
			Field:           "name",
			Width:           25,
			Alignment:       core.AlignLeft,
			HeaderAlignment: core.AlignCenter,
			HeaderConstraint: core.CellConstraint{
				Width:     25,
				Alignment: core.AlignCenter,
				Padding:   core.PaddingConfig{Left: 1, Right: 1},
			},
		},
		{
			Title:           "Value",
			Field:           "value",
			Width:           15,
			Alignment:       core.AlignRight,
			HeaderAlignment: core.AlignLeft,
			HeaderConstraint: core.CellConstraint{
				Width:     15,
				Alignment: core.AlignLeft,
			},
		},
		{
			Title:           "Status",
			Field:           "status",
			Width:           18,
			Alignment:       core.AlignCenter,
			HeaderAlignment: core.AlignRight,
			HeaderConstraint: core.CellConstraint{
				Width:     18,
				Alignment: core.AlignRight,
			},
		},
		{
			Title:           "Category",
			Field:           "category",
			Width:           20,
			Alignment:       core.AlignLeft,
			HeaderAlignment: core.AlignCenter,
			HeaderConstraint: core.CellConstraint{
				Width:     20,
				Alignment: core.AlignCenter,
				Padding:   core.PaddingConfig{Left: 2, Right: 2},
			},
		},
		{
			Title:           "Description",
			Field:           "description",
			Width:           22, // Medium width to show moderate truncation
			Alignment:       core.AlignLeft,
			HeaderAlignment: core.AlignLeft,
		},
	},
	// Value-first order: Value, Name, Status, Category, Description
	{
		{
			Title:           "Value",
			Field:           "value",
			Width:           15,
			Alignment:       core.AlignRight,
			HeaderAlignment: core.AlignCenter,
		},
		{
			Title:           "Name",
			Field:           "name",
			Width:           25,
			Alignment:       core.AlignLeft,
			HeaderAlignment: core.AlignLeft,
		},
		{
			Title:           "Status",
			Field:           "status",
			Width:           18,
			Alignment:       core.AlignCenter,
			HeaderAlignment: core.AlignCenter,
		},
		{
			Title:           "Category",
			Field:           "category",
			Width:           20,
			Alignment:       core.AlignLeft,
			HeaderAlignment: core.AlignRight,
		},
		{
			Title:           "Description",
			Field:           "description",
			Width:           35, // Slightly narrower to show more truncation
			Alignment:       core.AlignLeft,
			HeaderAlignment: core.AlignLeft,
		},
	},
	// Status-first order: Status, Category, Name, Value, Description
	{
		{
			Title:           "Status",
			Field:           "status",
			Width:           18,
			Alignment:       core.AlignCenter,
			HeaderAlignment: core.AlignCenter,
		},
		{
			Title:           "Category",
			Field:           "category",
			Width:           20,
			Alignment:       core.AlignLeft,
			HeaderAlignment: core.AlignLeft,
		},
		{
			Title:           "Name",
			Field:           "name",
			Width:           25,
			Alignment:       core.AlignLeft,
			HeaderAlignment: core.AlignRight,
		},
		{
			Title:           "Value",
			Field:           "value",
			Width:           15,
			Alignment:       core.AlignRight,
			HeaderAlignment: core.AlignCenter,
		},
		{
			Title:           "Description",
			Field:           "description",
			Width:           20, // Narrow width to show aggressive truncation
			Alignment:       core.AlignLeft,
			HeaderAlignment: core.AlignLeft,
		},
	},
}

func main() {
	// Create data source
	dataSource := NewExampleTableDataSource(1000)

	// Create table
	table := CreateExampleTableWithDataSource(dataSource)

	// CRITICAL: Focus the table so it can receive key events
	table.Focus()

	initialTheme := availableThemes[0]
	SetTheme(initialTheme)

	// Create app model
	model := AppModel{
		table:                table,
		dataSource:           dataSource,
		statusMessage:        fmt.Sprintf("Ready! %d items loaded with theme: %s", dataSource.totalItems, initialTheme.Name),
		showHelp:             true,
		showDebug:            false,
		inputMode:            false,
		indexInput:           "",
		fullRowCursorEnabled: true,
		selectionModeEnabled: false,
		scrollResetEnabled:   true, // Match the config setting

		loadingChunks:       make(map[int]bool),
		chunkHistory:        make([]string, 0),
		themeIndex:          0,
		columnOrderIndex:    0,
		sortingEnabled:      false,
		currentSortField:    "",
		currentSortDir:      "",
		filteringEnabled:    false,
		currentFilter:       "",
		currentFilterField:  "",
		showTopBorder:       true,
		showBottomBorder:    true,
		showHeaderSeparator: true,
		removeTopSpace:      false,
		removeBottomSpace:   false,
		activeCellEnabled:   false,
	}

	// Run the interactive program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}

func (m AppModel) Init() tea.Cmd {
	// Get the header formatters
	headerFormatters := createCustomHeaderFormatter()

	var cmds []tea.Cmd
	cmds = append(cmds, m.table.Init())
	cmds = append(cmds, m.table.Focus())

	// Set cell formatters through the Tea model loop
	cmds = append(cmds, core.CellFormatterSetCmd(0, NameCellFormatter))        // Name column
	cmds = append(cmds, core.CellFormatterSetCmd(1, ValueCellFormatter))       // Value column
	cmds = append(cmds, core.CellFormatterSetCmd(2, StatusCellFormatter))      // Status column
	cmds = append(cmds, core.CellFormatterSetCmd(3, CategoryCellFormatter))    // Category column
	cmds = append(cmds, core.CellFormatterSetCmd(4, DescriptionCellFormatter)) // Description column

	// Set header formatters for each column
	for columnIndex, formatter := range headerFormatters {
		cmds = append(cmds, core.HeaderFormatterSetCmd(columnIndex, formatter))
	}

	return tea.Batch(cmds...)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Handle chunk loading messages for debug tracking
	case core.ChunkLoadingStartedMsg:
		m.loadingChunks[msg.ChunkStart] = true
		historyEntry := fmt.Sprintf("Started loading chunk %d (size: %d)", msg.ChunkStart, msg.Request.Count)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[len(m.chunkHistory)-10:]
		}
		return m, nil

	case core.ChunkLoadingCompletedMsg:
		delete(m.loadingChunks, msg.ChunkStart)
		historyEntry := fmt.Sprintf("Completed chunk %d (%d items)", msg.ChunkStart, msg.ItemCount)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[len(m.chunkHistory)-10:]
		}
		return m, nil

	case core.ChunkUnloadedMsg:
		historyEntry := fmt.Sprintf("Unloaded chunk %d", msg.ChunkStart)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[len(m.chunkHistory)-10:]
		}
		return m, nil

	case core.SelectionResponseMsg:
		// Update status based on selection
		selectionCount := m.dataSource.GetSelectionCount()
		state := m.table.GetState()
		if msg.Success {
			switch msg.Operation {
			case "toggle":
				m.statusMessage = fmt.Sprintf("Selected item at index %d - %d total selected (look for background highlighting)", state.CursorIndex, selectionCount)
			case "selectAll":
				m.statusMessage = fmt.Sprintf("Selected ALL %d items in datasource (look for background highlighting)", selectionCount)
			case "clear":
				m.statusMessage = "All selections cleared - indicators removed"
			}
		} else {
			m.statusMessage = fmt.Sprintf("Selection failed: %v", msg.Error)
		}
		// Also pass to table
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd

	case core.PageUpMsg:
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		state := m.table.GetState()
		m.statusMessage = fmt.Sprintf("Page up - now at index %d", state.CursorIndex)
		return m, cmd

	case core.PageDownMsg:
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		state := m.table.GetState()
		m.statusMessage = fmt.Sprintf("Page down - now at index %d", state.CursorIndex)
		return m, cmd

	case core.JumpToMsg:
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		state := m.table.GetState()
		m.statusMessage = fmt.Sprintf("Jumped to index %d", state.CursorIndex)
		return m, cmd

	case core.JumpToStartMsg:
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		m.statusMessage = "Jumped to start"
		return m, cmd

	case core.JumpToEndMsg:
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		m.statusMessage = "Jumped to end"
		return m, cmd

	case core.CursorUpMsg, core.CursorDownMsg:
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		state := m.table.GetState()
		m.statusMessage = fmt.Sprintf("Position: %d/%d (Viewport: %d-%d)",
			state.CursorIndex, m.table.GetTotalItems(),
			state.ViewportStartIndex,
			state.ViewportStartIndex+9) // viewport height is 10
		return m, cmd

	case core.DataTotalMsg:
		historyEntry := fmt.Sprintf("Total items: %d", msg.Total)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[len(m.chunkHistory)-10:]
		}
		// Also pass to table
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd

	case core.DataRefreshMsg:
		historyEntry := "Data refresh triggered"
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[len(m.chunkHistory)-10:]
		}
		// Also pass to table
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd

	case core.DataLoadErrorMsg:
		historyEntry := fmt.Sprintf("Data load error: %v", msg.Error)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[len(m.chunkHistory)-10:]
		}
		// Also pass to table
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd

	case core.DataChunkLoadedMsg:
		// This is the actual chunk data loading completion
		historyEntry := fmt.Sprintf("Loaded chunk %d (%d items)", msg.StartIndex, len(msg.Items))
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[len(m.chunkHistory)-10:]
		}
		// Also pass to table
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd

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
					return m, core.JumpToCmd(index)
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
			m.themeIndex = (m.themeIndex + 1) % len(availableThemes)
			newTheme := availableThemes[m.themeIndex]
			SetTheme(newTheme)
			m.statusMessage = fmt.Sprintf("Theme changed to: %s - %s", newTheme.Name, newTheme.Description)

			// CRITICAL FIX: Update the table's built-in theme too!
			vtableTheme := convertToVTableTheme(newTheme)
			return m, m.table.SetTheme(vtableTheme)

		case "r":
			// Refresh data
			m.statusMessage = "Refreshing data..."
			return m, core.DataRefreshCmd()

		case "R":
			// Toggle full row cursor highlighting (uppercase R)
			if m.fullRowCursorEnabled {
				// Disable full row highlighting using the BUILT-IN system
				m.fullRowCursorEnabled = false
				m.statusMessage = "Full row cursor highlighting disabled"
				return m, core.FullRowHighlightEnableCmd(false)
			} else {
				// Enable full row highlighting using the BUILT-IN system
				m.fullRowCursorEnabled = true
				// Disable selection mode if enabled
				if m.selectionModeEnabled {
					m.selectionModeEnabled = false
				}
				m.statusMessage = "Full row cursor highlighting enabled - entire row highlighted"
				return m, core.FullRowHighlightEnableCmd(true)
			}

		// === BORDER CONTROL KEYS ===
		case "B":
			// Cycle through different border combinations
			// Cycle through border combinations: All → Top+Bottom → Header Only → None → All
			if m.showTopBorder && m.showBottomBorder && m.showHeaderSeparator {
				// All borders → Top and Bottom only
				m.showTopBorder = true
				m.showBottomBorder = true
				m.showHeaderSeparator = false
				m.removeTopSpace = false    // Reset space removal
				m.removeBottomSpace = false // Reset space removal
				m.statusMessage = "Borders: Top and Bottom only"
				return m, tea.Batch(
					core.TopBorderVisibilityCmd(true),
					core.BottomBorderVisibilityCmd(true),
					core.HeaderSeparatorVisibilityCmd(false),
					core.TopBorderSpaceRemovalCmd(false),
					core.BottomBorderSpaceRemovalCmd(false),
				)
			} else if m.showTopBorder && m.showBottomBorder && !m.showHeaderSeparator {
				// Top+Bottom → Header separator only
				m.showTopBorder = false
				m.showBottomBorder = false
				m.showHeaderSeparator = true
				m.removeTopSpace = false    // Reset space removal
				m.removeBottomSpace = false // Reset space removal
				m.statusMessage = "Borders: Header separator only"
				return m, tea.Batch(
					core.TopBorderVisibilityCmd(false),
					core.BottomBorderVisibilityCmd(false),
					core.HeaderSeparatorVisibilityCmd(true),
					core.TopBorderSpaceRemovalCmd(false),
					core.BottomBorderSpaceRemovalCmd(false),
				)
			} else if !m.showTopBorder && !m.showBottomBorder && m.showHeaderSeparator {
				// Header only → No borders
				m.showTopBorder = false
				m.showBottomBorder = false
				m.showHeaderSeparator = false
				m.removeTopSpace = false    // Reset space removal
				m.removeBottomSpace = false // Reset space removal
				m.statusMessage = "Borders: None (clean minimal look)"
				return m, tea.Batch(
					core.TopBorderVisibilityCmd(false),
					core.BottomBorderVisibilityCmd(false),
					core.HeaderSeparatorVisibilityCmd(false),
					core.TopBorderSpaceRemovalCmd(false),
					core.BottomBorderSpaceRemovalCmd(false),
				)
			} else {
				// None or other → All borders
				m.showTopBorder = true
				m.showBottomBorder = true
				m.showHeaderSeparator = true
				m.removeTopSpace = false    // Reset space removal
				m.removeBottomSpace = false // Reset space removal
				m.statusMessage = "Borders: All enabled (full table frame)"
				return m, tea.Batch(
					core.TopBorderVisibilityCmd(true),
					core.BottomBorderVisibilityCmd(true),
					core.HeaderSeparatorVisibilityCmd(true),
					core.TopBorderSpaceRemovalCmd(false),
					core.BottomBorderSpaceRemovalCmd(false),
				)
			}

		case "6":
			// Toggle top border specifically
			m.showTopBorder = !m.showTopBorder
			if m.showTopBorder {
				m.statusMessage = "Top border enabled"
			} else {
				m.statusMessage = "Top border disabled"
			}
			return m, core.TopBorderVisibilityCmd(m.showTopBorder)

		case "7":
			// Toggle bottom border specifically
			m.showBottomBorder = !m.showBottomBorder
			if m.showBottomBorder {
				m.statusMessage = "Bottom border enabled"
			} else {
				m.statusMessage = "Bottom border disabled"
			}
			return m, core.BottomBorderVisibilityCmd(m.showBottomBorder)

		case "8":
			// Toggle header separator specifically
			m.showHeaderSeparator = !m.showHeaderSeparator
			if m.showHeaderSeparator {
				m.statusMessage = "Header separator enabled"
			} else {
				m.statusMessage = "Header separator disabled"
			}
			return m, core.HeaderSeparatorVisibilityCmd(m.showHeaderSeparator)

		case "9":
			// Toggle top border space removal
			m.removeTopSpace = !m.removeTopSpace
			if m.removeTopSpace {
				m.statusMessage = "Top border space removed - table flows to top"
			} else {
				m.statusMessage = "Top border space preserved"
			}
			return m, core.TopBorderSpaceRemovalCmd(m.removeTopSpace)

		case "0":
			// Toggle bottom border space removal
			m.removeBottomSpace = !m.removeBottomSpace
			if m.removeBottomSpace {
				m.statusMessage = "Bottom border space removed - table flows to bottom"
			} else {
				m.statusMessage = "Bottom border space preserved"
			}
			return m, core.BottomBorderSpaceRemovalCmd(m.removeBottomSpace)

		case "o":
			// Cycle through column orderings
			m.columnOrderIndex = (m.columnOrderIndex + 1) % len(availableColumnOrders)
			newColumns := availableColumnOrders[m.columnOrderIndex]
			orderNames := []string{"Default (Name→Value→Status→Category)", "Value-first", "Status-first"}
			m.statusMessage = fmt.Sprintf("Column order changed to: %s", orderNames[m.columnOrderIndex])
			return m, core.ColumnSetCmd(newColumns)

		case "T":
			// Toggle sorting (uppercase T for table sorting)
			if !m.sortingEnabled {
				// Enable sorting on Name field ascending
				m.sortingEnabled = true
				m.currentSortField = "name"
				m.currentSortDir = "asc"
				m.statusMessage = "Sorting enabled: Name (ascending) - press T again to cycle"
				m.dataSource.SetSort([]string{"name"}, []string{"asc"})
				return m, core.DataRefreshCmd()
			} else {
				// Cycle through different sorts
				switch m.currentSortField + "_" + m.currentSortDir {
				case "name_asc":
					m.currentSortField = "name"
					m.currentSortDir = "desc"
					m.statusMessage = "Sorting: Name (descending)"
					m.dataSource.SetSort([]string{"name"}, []string{"desc"})
					return m, core.DataRefreshCmd()
				case "name_desc":
					m.currentSortField = "value"
					m.currentSortDir = "asc"
					m.statusMessage = "Sorting: Value (ascending)"
					m.dataSource.SetSort([]string{"value"}, []string{"asc"})
					return m, core.DataRefreshCmd()
				case "value_asc":
					m.currentSortField = "value"
					m.currentSortDir = "desc"
					m.statusMessage = "Sorting: Value (descending)"
					m.dataSource.SetSort([]string{"value"}, []string{"desc"})
					return m, core.DataRefreshCmd()
				case "value_desc":
					m.currentSortField = "status"
					m.currentSortDir = "asc"
					m.statusMessage = "Sorting: Status (ascending)"
					m.dataSource.SetSort([]string{"status"}, []string{"asc"})
					return m, core.DataRefreshCmd()
				case "status_asc":
					// Add multi-sort: Status + Name
					m.statusMessage = "Multi-sort: Status (asc) + Name (asc)"
					m.dataSource.SetSort([]string{"status", "name"}, []string{"asc", "asc"})
					return m, core.DataRefreshCmd()
				default:
					// Clear sorting
					m.sortingEnabled = false
					m.currentSortField = ""
					m.currentSortDir = ""
					m.statusMessage = "Sorting disabled"
					m.dataSource.SetSort([]string{}, []string{})
					return m, core.DataRefreshCmd()
				}
			}

		case "S":
			// Toggle selection highlighting mode (uppercase S - different from lowercase 's' which shows selection count)
			if m.selectionModeEnabled {
				// Disable selection highlighting
				m.selectionModeEnabled = false
				m.statusMessage = "Selection highlighting disabled"
				return m, resetAllFormattersToBase()
			} else {
				// Enable selection highlighting
				m.selectionModeEnabled = true
				// Disable full row mode if enabled
				if m.fullRowCursorEnabled {
					m.fullRowCursorEnabled = false
				}
				m.statusMessage = "Selection highlighting enabled - selected rows highlighted"
				// Use special selection-aware formatters
				return m, tea.Batch(
					resetAllFormattersToBase(),
					tea.Sequence(
						core.CellFormatterSetCmd(0, createSelectionAwareFormatter(NameCellFormatter)),
						core.CellFormatterSetCmd(1, createSelectionAwareFormatter(ValueCellFormatter)),
						core.CellFormatterSetCmd(2, createSelectionAwareFormatter(StatusCellFormatter)),
						core.CellFormatterSetCmd(3, createSelectionAwareFormatter(CategoryCellFormatter)),
						core.CellFormatterSetCmd(4, createSelectionAwareFormatter(DescriptionCellFormatter)),
					),
				)
			}

		case "F":
			// Toggle filtering (uppercase F for filter setup)
			if !m.filteringEnabled {
				// Enable filtering - show only Category A items
				m.filteringEnabled = true
				m.currentFilterField = "category"
				m.currentFilter = "Category A"
				m.statusMessage = "Filtering enabled: Category A only - press F to cycle filters"
				m.dataSource.SetFilter("category", "Category A")
				return m, core.DataRefreshCmd()
			} else {
				// Cycle through different filters
				switch m.currentFilter {
				case "Category A":
					m.currentFilter = "Category B"
					m.statusMessage = "Filtering: Category B only"
					m.dataSource.ClearAllFilters()
					m.dataSource.SetFilter("category", "Category B")
					return m, core.DataRefreshCmd()
				case "Category B":
					// Switch to value filter
					m.currentFilterField = "value"
					m.currentFilter = "high" // Filter for high values
					m.statusMessage = "Filtering: High values only (>50)"
					m.dataSource.ClearAllFilters()
					m.dataSource.SetFilter("value", "high")
					return m, core.DataRefreshCmd()
				case "high":
					// Switch to status filter
					m.currentFilterField = "status"
					m.currentFilter = "active"
					m.statusMessage = "Filtering: Active status only"
					m.dataSource.ClearAllFilters()
					m.dataSource.SetFilter("status", "active")
					return m, core.DataRefreshCmd()
				default:
					// Clear filtering
					m.filteringEnabled = false
					m.currentFilterField = ""
					m.currentFilter = ""
					m.statusMessage = "Filtering disabled"
					m.dataSource.ClearAllFilters()
					return m, core.DataRefreshCmd()
				}
			}

		// === NAVIGATION KEYS ===
		case "g":
			// Jump to start (like vim)
			m.statusMessage = "Jumping to start"
			return m, core.JumpToStartCmd()

		case "G":
			// Jump to end (like vim)
			m.statusMessage = "Jumping to end"
			return m, core.JumpToEndCmd()

		case "J":
			// Enter jump-to-index mode (uppercase J)
			m.inputMode = true
			m.indexInput = ""
			m.statusMessage = "Enter index to jump to (0-999): "
			return m, nil

		case "h":
			// Page up using proper command
			m.statusMessage = "Page up"
			return m, core.PageUpCmd()

		case "l":
			// Page down using proper command
			m.statusMessage = "Page down"
			return m, core.PageDownCmd()

		case "j", "up":
			// Move up using proper command
			return m, core.CursorUpCmd()

		case "k", "down":
			// Move down using proper command
			return m, core.CursorDownCmd()

		// === SELECTION KEYS ===
		case " ":
			return m, core.SelectCurrentCmd()

		case "a":
			return m, core.SelectAllCmd()

		case "c":
			return m, core.SelectClearCmd()

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
			return m, core.JumpToCmd(100)

		case "2":
			m.statusMessage = "Quick jump to index 250"
			return m, core.JumpToCmd(250)

		case "3":
			m.statusMessage = "Quick jump to index 500"
			return m, core.JumpToCmd(500)

		case "4":
			m.statusMessage = "Quick jump to index 750"
			return m, core.JumpToCmd(750)

		case "5":
			m.statusMessage = "Quick jump to index 900"
			return m, core.JumpToCmd(900)

		case "m", "M":
			// Toggle scroll mode and show the new state
			var cmd tea.Cmd
			_, cmd = m.table.Update(msg)

			// Get the new state to show in status
			newMode, _, _, _ := m.table.GetHorizontalScrollState()
			switch newMode {
			case "character":
				m.statusMessage = "Horizontal scroll mode: CHARACTER (letter-by-letter, press M to change to word)"
			case "word":
				m.statusMessage = "Horizontal scroll mode: WORD (word-by-word, press M to change to smart)"
			case "smart":
				m.statusMessage = "Horizontal scroll mode: SMART (intelligent boundaries, press M to change to character)"
			}
			return m, cmd

		case "v", "V":
			// Toggle scroll scope and show the new state
			var cmd tea.Cmd
			_, cmd = m.table.Update(msg)

			// Get the new state to show in status
			_, scrollAllRows, _, _ := m.table.GetHorizontalScrollState()
			if scrollAllRows {
				m.statusMessage = "Horizontal scroll scope: ALL ROWS move together (press V to change to current row only)"
			} else {
				m.statusMessage = "Horizontal scroll scope: CURRENT ROW ONLY moves (press V to change to all rows)"
			}
			return m, cmd

		case "D":
			// Run scroll debug test (uppercase D for Debug)
			m.statusMessage = "Scroll debug test - see scroll_test.go file and build separately"
			return m, nil

		case "z", "Z":
			// Toggle scroll reset on navigation
			m.scrollResetEnabled = !m.scrollResetEnabled
			if m.scrollResetEnabled {
				m.statusMessage = "Scroll reset ON - horizontal scroll resets when navigating between rows (press Z to disable)"
			} else {
				m.statusMessage = "Scroll reset OFF - horizontal scroll persists when navigating between rows (press Z to enable)"
			}

			// Update the table configuration
			m.table.SetResetScrollOnNavigation(m.scrollResetEnabled)
			return m, nil

		case "A":
			// Toggle active cell indication mode (uppercase A for Active cell)
			// Simple toggle: off -> on -> off

			if m.activeCellEnabled {
				// Disable active cell indication
				m.activeCellEnabled = false
				m.statusMessage = "Active cell indication: DISABLED - use C/arrow keys to change column"
				return m, core.ActiveCellIndicationModeSetCmd(false)
			} else {
				// Enable active cell indication with background mode
				m.activeCellEnabled = true
				m.statusMessage = "Active cell indication: ENABLED (background highlighting) - use C/arrow keys to change column"
				return m, core.ActiveCellIndicationModeSetCmd(true)
			}

		case "C":
			// Cycle active column for testing (uppercase C for Column)
			// Get current horizontal scroll state
			_, _, currentCol, _ := m.table.GetHorizontalScrollState()
			newCol := (currentCol + 1) % 5 // Cycle through 5 columns (0-4)

			m.statusMessage = fmt.Sprintf("Active column changed to: %d (%s) - use arrow keys to scroll horizontally",
				newCol, []string{"Name", "Value", "Status", "Category", "Description"}[newCol])

			// Use the existing column navigation - simulate pressing "." to move to next column
			var cmd tea.Cmd
			_, cmd = m.table.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'.'}})
			return m, cmd

		default:
			// Pass other keys to table
			var cmd tea.Cmd
			_, cmd = m.table.Update(msg)
			return m, cmd
		}

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
		view.WriteString("=== ADVANCED TABLE FORMATTER DEMO ===\n")
		view.WriteString("✨ NEW FEATURES: All advanced table features showcased!\n\n")

		view.WriteString("🎨 VISUAL FEATURES:\n")
		view.WriteString("• Cell Formatters: Each column styled with colors + icons + auto-truncation\n")
		view.WriteString("• Header Formatters: Emoji headers (📝💰📊🏷️) with perfect Unicode alignment\n")
		view.WriteString("• Component Renderer: Clean cursor (►) + selection (✓) indicators\n")
		view.WriteString("• Visual Themes: t=cycle (Default→Heavy→Minimal→Retro) - different borders!\n")
		view.WriteString("• Border Controls: B=cycle borders • 6=top • 7=bottom • 8=header separator • 9=remove top space • 0=remove bottom space\n\n")

		view.WriteString("📊 DATA FEATURES:\n")
		view.WriteString("• Column Ordering: o=cycle column arrangements\n")
		view.WriteString("• Sort/Multi-Sort: T=cycle sorts (Name↑↓→Value↑↓→Status↑→Multi-Sort→Clear)\n")
		view.WriteString("• Filter/Multi-Filter: F=cycle filters (CategoryA→CategoryB→HighValues→ActiveStatus→Clear)\n\n")

		view.WriteString("🖱️  INTERACTION:\n")
		view.WriteString("• Navigation: j/k↑↓=move • h/l=page • g/G=start/end • J=jump • 1-5=quick jumps\n")
		view.WriteString("• Selection: Space=toggle • a=select all • c=clear • s=show count\n")
		view.WriteString("• Selection Highlighting: S=toggle selection background highlighting\n")
		view.WriteString("• Full Row Highlighting: R=toggle full row highlighting\n")
		view.WriteString("• Horizontal Scrolling: ←→=character • []=word • {}=smart • .,=change column • M=toggle mode • V=toggle scope • Backspace=reset (ANSI-aware)\n")
		view.WriteString("• Scroll Reset: Z=toggle (reset scroll when navigating between rows)\n")
		view.WriteString("• Active Cell Indication: A=toggle on/off (background highlighting) • C=cycle active column\n\n")

		view.WriteString("🔧 SYSTEM:\n")
		view.WriteString("• Debug: d=toggle (chunk loading activity) • r=refresh • ?=help • q=quit\n\n")

		// Show current feature states
		view.WriteString("📈 CURRENT STATE:\n")
		view.WriteString(fmt.Sprintf("• Theme: %s • Order: %s\n",
			currentTheme.Name,
			[]string{"Default", "Value-first", "Status-first"}[m.columnOrderIndex]))

		if m.sortingEnabled {
			if m.currentSortField != "" {
				view.WriteString(fmt.Sprintf("• Sorting: %s (%s)\n", m.currentSortField, m.currentSortDir))
			} else {
				view.WriteString("• Sorting: Multi-sort active\n")
			}
		} else {
			view.WriteString("• Sorting: Disabled\n")
		}

		if m.filteringEnabled {
			view.WriteString(fmt.Sprintf("• Filtering: %s=%s\n", m.currentFilterField, m.currentFilter))
		} else {
			view.WriteString("• Filtering: Disabled\n")
		}

		view.WriteString(fmt.Sprintf("• Selection Mode: %s • Full Row: %s • Scroll Reset: %s\n",
			map[bool]string{true: "Enabled", false: "Disabled"}[m.selectionModeEnabled],
			map[bool]string{true: "Enabled", false: "Disabled"}[m.fullRowCursorEnabled],
			map[bool]string{true: "Enabled", false: "Disabled"}[m.scrollResetEnabled]))

		// Get horizontal scrolling state from table
		scrollMode, scrollAllRows, currentCol, scrollOffsets := m.table.GetHorizontalScrollState()

		// Make scope description clearer
		scopeDesc := "current row only"
		if scrollAllRows {
			scopeDesc = "all rows move together"
		}

		view.WriteString(fmt.Sprintf("• Horizontal Scroll: Mode=%s • Scope=%s • Column=%d • Offsets=%v\n",
			scrollMode, scopeDesc, currentCol, scrollOffsets))

		view.WriteString(fmt.Sprintf("• Borders: Top=%s • Bottom=%s • Header=%s • TopSpace=%s • BottomSpace=%s\n\n",
			map[bool]string{true: "On", false: "Off"}[m.showTopBorder],
			map[bool]string{true: "On", false: "Off"}[m.showBottomBorder],
			map[bool]string{true: "On", false: "Off"}[m.showHeaderSeparator],
			map[bool]string{true: "Removed", false: "Preserved"}[m.removeTopSpace],
			map[bool]string{true: "Removed", false: "Preserved"}[m.removeBottomSpace]))
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
	selectionCount := m.dataSource.GetSelectionCount()
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
func CreateExampleTableWithDataSource(dataSource *ExampleTableDataSource) *table.Table {
	// Create table configuration with wider columns and custom header alignments
	columns := []core.TableColumn{
		{
			Title:           "Name",
			Field:           "name",
			Width:           25,
			Alignment:       core.AlignLeft,   // Cell content: left aligned
			HeaderAlignment: core.AlignCenter, // Header: center aligned
			HeaderConstraint: core.CellConstraint{
				Width:     25,
				Alignment: core.AlignCenter,
				Padding:   core.PaddingConfig{Left: 1, Right: 1},
			},
		},
		{
			Title:           "Value",
			Field:           "value",
			Width:           15,
			Alignment:       core.AlignRight, // Cell content: right aligned
			HeaderAlignment: core.AlignLeft,  // Header: left aligned (different!)
			HeaderConstraint: core.CellConstraint{
				Width:     15,
				Alignment: core.AlignLeft,
			},
		},
		{
			Title:           "Status",
			Field:           "status",
			Width:           18,
			Alignment:       core.AlignCenter, // Cell content: center aligned
			HeaderAlignment: core.AlignRight,  // Header: right aligned (different!)
			HeaderConstraint: core.CellConstraint{
				Width:     18,
				Alignment: core.AlignRight,
			},
		},
		{
			Title:           "Category",
			Field:           "category",
			Width:           20,
			Alignment:       core.AlignLeft,   // Cell content: left aligned
			HeaderAlignment: core.AlignCenter, // Header: center aligned
			HeaderConstraint: core.CellConstraint{
				Width:     20,
				Alignment: core.AlignCenter,
				Padding:   core.PaddingConfig{Left: 2, Right: 2},
			},
		},
		{
			Title:           "Description",
			Field:           "description",
			Width:           22, // Medium width to show moderate truncation
			Alignment:       core.AlignLeft,
			HeaderAlignment: core.AlignLeft,
		},
	}

	config := core.TableConfig{
		Columns:     columns,
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: core.ViewportConfig{
			Height:             10,
			TopThreshold:       2,
			BottomThreshold:    2,
			ChunkSize:          50,
			InitialIndex:       0,
			BoundingAreaBefore: 25,
			BoundingAreaAfter:  25,
		},
		Theme:                   convertToVTableTheme(currentTheme),
		SelectionMode:           core.SelectionMultiple,
		FullRowHighlighting:     true,
		ResetScrollOnNavigation: true, // Enable scroll reset for better UX
		// Enable active cell indication with background color mode
		ActiveCellIndicationEnabled: false, // Use boolean instead of string
		ActiveCellBackgroundColor:   "226", // Bright yellow background for active cell
		KeyMap: core.NavigationKeyMap{
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
	table := table.NewTable(config, dataSource)

	return table
}
