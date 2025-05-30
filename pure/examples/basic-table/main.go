package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	// Add sorting and filtering state
	sortFields    []string
	sortDirs      []string
	filters       map[string]any
	filteredData  []vtable.TableRow // Cached filtered/sorted data
	filteredTotal int               // Total after filtering
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
		return vtable.DataTotalMsg{Total: ds.filteredTotal}
	}
}

// RefreshTotal refreshes the total count
func (ds *ExampleTableDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

// LoadChunk loads a chunk of data
func (ds *ExampleTableDataSource) LoadChunk(request vtable.DataRequest) tea.Cmd {
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

		var items []vtable.Data[any]
		for i := start; i < end; i++ {
			if i < len(ds.filteredData) {
				items = append(items, vtable.Data[any]{
					ID:       ds.filteredData[i].ID,
					Item:     ds.filteredData[i],
					Selected: ds.selectedItems[ds.filteredData[i].ID],
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
	result := make([]vtable.TableRow, 0, len(ds.data))

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
		Description:   "Clean blue and white theme",
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
		Description:   "Smooth cyberpunk neon theme",
		CursorBg:      "201", // Bright magenta cursor
		SelectionBg:   "235", // Dark gray for subtle selection (instead of jarring purple)
		HeaderBg:      "0",   // Black
		BorderColor:   "14",  // Cyan borders
		PrimaryText:   "15",  // White
		SecondaryText: "8",   // Muted gray (instead of bright cyan)
		AccentText:    "13",  // Magenta (less bright than 201)
		ErrorText:     "9",   // Standard red (instead of ultra-bright 196)
		WarningText:   "11",  // Standard yellow (instead of ultra-bright 226)
		SuccessText:   "10",  // Standard green (instead of ultra-bright 46)
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

// convertToVTableTheme converts demo TableTheme to vtable.Theme
func convertToVTableTheme(theme TableTheme) vtable.Theme {
	return vtable.Theme{
		HeaderStyle:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.PrimaryText)).Background(lipgloss.Color(theme.HeaderBg)),
		CellStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color(theme.PrimaryText)),
		CursorStyle:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.PrimaryText)).Background(lipgloss.Color(theme.CursorBg)),
		SelectedStyle:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color(theme.SelectionBg)), // White text on selection
		FullRowCursorStyle: lipgloss.NewStyle().Background(lipgloss.Color(theme.CursorBg)).Foreground(lipgloss.Color(theme.PrimaryText)).Bold(true),
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
		BorderColor: theme.BorderColor,
		HeaderColor: theme.PrimaryText,
	}
}

// ================================
// ENHANCED CELL FORMATTERS
// ================================

// NameCellFormatter formats the first column with proper selection/cursor handling
func NameCellFormatter(cellValue string, rowIndex int, column vtable.TableColumn, ctx vtable.RenderContext, isCursor, isSelected bool) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.PrimaryText))

	// Apply selection background if selected (and not overridden by cursor full row mode)
	if isSelected {
		style = style.Background(lipgloss.Color(currentTheme.SelectionBg)).Foreground(lipgloss.Color("15")) // White text on selection bg
	}

	return style.Render(cellValue)
}

// ValueCellFormatter formats value cells with colors and selection handling
func ValueCellFormatter(cellValue string, rowIndex int, column vtable.TableColumn, ctx vtable.RenderContext, isCursor, isSelected bool) string {
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

	// Apply selection background if selected
	if isSelected {
		style = style.Background(lipgloss.Color(currentTheme.SelectionBg)).Foreground(lipgloss.Color("15")) // White text on selection bg
	}

	return style.Render(cellValue)
}

// StatusCellFormatter formats status cells with icons, colors and selection handling
func StatusCellFormatter(cellValue string, rowIndex int, column vtable.TableColumn, ctx vtable.RenderContext, isCursor, isSelected bool) string {
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

	// Apply selection background if selected
	if isSelected {
		style = style.Background(lipgloss.Color(currentTheme.SelectionBg)).Foreground(lipgloss.Color("15")) // White text on selection bg
	}

	return style.Render(statusText)
}

// CategoryCellFormatter formats category cells with colors and selection handling
func CategoryCellFormatter(cellValue string, rowIndex int, column vtable.TableColumn, ctx vtable.RenderContext, isCursor, isSelected bool) string {
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

	// Apply selection background if selected
	if isSelected {
		style = style.Background(lipgloss.Color(currentTheme.SelectionBg)).Foreground(lipgloss.Color("15")) // White text on selection bg
	}

	return style.Render(cellValue)
}

// createCustomHeaderFormatter creates a custom header formatter with styling
func createCustomHeaderFormatter() map[int]vtable.SimpleHeaderFormatter {
	formatters := make(map[int]vtable.SimpleHeaderFormatter)

	// Simple header formatters like the working test
	formatters[0] = func(column vtable.TableColumn, ctx vtable.RenderContext) string {
		return "📝 " + column.Title
	}

	formatters[1] = func(column vtable.TableColumn, ctx vtable.RenderContext) string {
		return "💰 " + column.Title
	}

	formatters[2] = func(column vtable.TableColumn, ctx vtable.RenderContext) string {
		return "📊 " + column.Title
	}

	formatters[3] = func(column vtable.TableColumn, ctx vtable.RenderContext) string {
		return "🏷️ " + column.Title
	}

	return formatters
}

// ================================
// ADVANCED FORMATTER IMPLEMENTATIONS
// ================================

// createWrappingNameFormatter creates a name formatter with text wrapping support
func createWrappingNameFormatter() vtable.SimpleCellFormatter {
	return func(cellValue string, rowIndex int, column vtable.TableColumn, ctx vtable.RenderContext, isCursor, isSelected bool) string {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.PrimaryText))

		// For long content, show abbreviated version that fits better
		if len(cellValue) > 15 {
			// Break into words and create flowing text
			words := strings.Fields(cellValue)
			if len(words) > 1 {
				// Create a more compact representation
				abbreviated := words[0]
				if len(words) > 1 {
					abbreviated += "~" + words[len(words)-1] // First~Last pattern
				}
				return style.Render(abbreviated)
			}
		}

		return style.Render(cellValue)
	}
}

// createWrappingValueFormatter creates a value formatter with text wrapping support
func createWrappingValueFormatter() vtable.SimpleCellFormatter {
	return func(cellValue string, rowIndex int, column vtable.TableColumn, ctx vtable.RenderContext, isCursor, isSelected bool) string {
		// Parse value for color coding (same as original)
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

		// Add wrapping indicator that fits in single line
		if strings.HasPrefix(cellValue, "Value ") {
			valueStr := strings.TrimPrefix(cellValue, "Value ")
			enhanced := fmt.Sprintf("$%s⟲", valueStr) // Show currency symbol + wrap indicator
			return style.Render(enhanced)
		}

		return style.Render(cellValue)
	}
}

// createWrappingStatusFormatter creates a status formatter with text wrapping support
func createWrappingStatusFormatter() vtable.SimpleCellFormatter {
	return func(cellValue string, rowIndex int, column vtable.TableColumn, ctx vtable.RenderContext, isCursor, isSelected bool) string {
		// Convert status to visual representation with more compact info
		var statusText string
		var style lipgloss.Style

		switch cellValue {
		case "Status 0":
			style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.SuccessText))
			statusText = currentTheme.ActiveIcon + " Active✨" // Single line with indicator
		case "Status 1":
			style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.WarningText))
			statusText = currentTheme.WarningIcon + " Warn⚠️" // Single line with indicator
		case "Status 2":
			style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.ErrorText))
			statusText = currentTheme.ErrorIcon + " Error❌" // Single line with indicator
		default:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color(currentTheme.SecondaryText))
			statusText = currentTheme.UnknownIcon + " Unknown❓"
		}

		return style.Render(statusText)
	}
}

// createWrappingCategoryFormatter creates a category formatter with text wrapping support
func createWrappingCategoryFormatter() vtable.SimpleCellFormatter {
	return func(cellValue string, rowIndex int, column vtable.TableColumn, ctx vtable.RenderContext, isCursor, isSelected bool) string {
		// Color code categories (same as original)
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

		// Add compact wrapping indicator
		if strings.HasPrefix(cellValue, "Category ") {
			categoryLetter := strings.TrimPrefix(cellValue, "Category ")
			enhanced := fmt.Sprintf("📂%s-Type", categoryLetter) // More compact representation
			return style.Render(enhanced)
		}

		return style.Render(cellValue)
	}
}

// createFullRowCursorFormatter creates a formatter that highlights the entire row when cursor is on it
func createFullRowCursorFormatter(baseFormatter vtable.SimpleCellFormatter) vtable.SimpleCellFormatter {
	return func(cellValue string, rowIndex int, column vtable.TableColumn, ctx vtable.RenderContext, isCursor, isSelected bool) string {
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
		return baseFormatter(cellValue, rowIndex, column, ctx, false, isSelected)
	}
}

// createSelectionAwareFormatter creates a formatter that ensures selection highlighting is always applied
func createSelectionAwareFormatter(baseFormatter vtable.SimpleCellFormatter) vtable.SimpleCellFormatter {
	return func(cellValue string, rowIndex int, column vtable.TableColumn, ctx vtable.RenderContext, isCursor, isSelected bool) string {
		// Always apply the base formatter first (which should handle selection highlighting)
		result := baseFormatter(cellValue, rowIndex, column, ctx, isCursor, isSelected)

		// If selected but the result doesn't seem to have selection styling, force it
		if isSelected && !strings.Contains(result, currentTheme.SelectionBg) {
			// Force selection background
			selectionStyle := lipgloss.NewStyle().
				Background(lipgloss.Color(currentTheme.SelectionBg)).
				Foreground(lipgloss.Color("15")) // White text on selection bg
			return selectionStyle.Render(cellValue)
		}

		return result
	}
}

// ================================
// FORMATTER MANAGEMENT HELPERS
// ================================

// resetAllFormattersToBase resets all cell formatters to their base implementations
func resetAllFormattersToBase() tea.Cmd {
	return tea.Batch(
		vtable.CellFormatterSetCmd(0, NameCellFormatter),
		vtable.CellFormatterSetCmd(1, ValueCellFormatter),
		vtable.CellFormatterSetCmd(2, StatusCellFormatter),
		vtable.CellFormatterSetCmd(3, CategoryCellFormatter),
	)
}

// ================================
// MAIN INTERACTIVE APPLICATION
// ================================

// AppModel represents the state of our application
type AppModel struct {
	table                *vtable.Table
	dataSource           *ExampleTableDataSource
	statusMessage        string
	showHelp             bool
	showDebug            bool
	inputMode            bool
	indexInput           string
	wrappingEnabled      bool
	fullRowCursorEnabled bool // Track if full row cursor highlighting is enabled
	selectionModeEnabled bool // Track if selection highlighting is enabled
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
}

// Available themes list
var availableThemes = []TableTheme{
	DefaultTheme,
	DarkTheme,
	MinimalTheme,
	NeonTheme,
}

// Available column orderings
var availableColumnOrders = [][]vtable.TableColumn{
	// Default order: Name, Value, Status, Category
	{
		{
			Title:           "Name",
			Field:           "name",
			Width:           25,
			Alignment:       vtable.AlignLeft,
			HeaderAlignment: vtable.AlignCenter,
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
			Alignment:       vtable.AlignRight,
			HeaderAlignment: vtable.AlignLeft,
			HeaderConstraint: vtable.CellConstraint{
				Width:     15,
				Alignment: vtable.AlignLeft,
			},
		},
		{
			Title:           "Status",
			Field:           "status",
			Width:           18,
			Alignment:       vtable.AlignCenter,
			HeaderAlignment: vtable.AlignRight,
			HeaderConstraint: vtable.CellConstraint{
				Width:     18,
				Alignment: vtable.AlignRight,
			},
		},
		{
			Title:           "Category",
			Field:           "category",
			Width:           20,
			Alignment:       vtable.AlignLeft,
			HeaderAlignment: vtable.AlignCenter,
			HeaderConstraint: vtable.CellConstraint{
				Width:     20,
				Alignment: vtable.AlignCenter,
				Padding:   vtable.PaddingConfig{Left: 2, Right: 2},
			},
		},
	},
	// Value-first order: Value, Name, Status, Category
	{
		{
			Title:           "Value",
			Field:           "value",
			Width:           15,
			Alignment:       vtable.AlignRight,
			HeaderAlignment: vtable.AlignCenter,
		},
		{
			Title:           "Name",
			Field:           "name",
			Width:           25,
			Alignment:       vtable.AlignLeft,
			HeaderAlignment: vtable.AlignLeft,
		},
		{
			Title:           "Status",
			Field:           "status",
			Width:           18,
			Alignment:       vtable.AlignCenter,
			HeaderAlignment: vtable.AlignCenter,
		},
		{
			Title:           "Category",
			Field:           "category",
			Width:           20,
			Alignment:       vtable.AlignLeft,
			HeaderAlignment: vtable.AlignRight,
		},
	},
	// Status-first order: Status, Category, Name, Value
	{
		{
			Title:           "Status",
			Field:           "status",
			Width:           18,
			Alignment:       vtable.AlignCenter,
			HeaderAlignment: vtable.AlignCenter,
		},
		{
			Title:           "Category",
			Field:           "category",
			Width:           20,
			Alignment:       vtable.AlignLeft,
			HeaderAlignment: vtable.AlignLeft,
		},
		{
			Title:           "Name",
			Field:           "name",
			Width:           25,
			Alignment:       vtable.AlignLeft,
			HeaderAlignment: vtable.AlignRight,
		},
		{
			Title:           "Value",
			Field:           "value",
			Width:           15,
			Alignment:       vtable.AlignRight,
			HeaderAlignment: vtable.AlignCenter,
		},
	},
}

func main() {
	// Create data source
	dataSource := NewExampleTableDataSource(1000)

	// Create table
	table := CreateExampleTableWithDataSource(dataSource)

	// Create app model
	app := AppModel{
		table:         table,
		dataSource:    dataSource,
		showDebug:     true,
		showHelp:      true,
		statusMessage: "Welcome! Use arrow keys to navigate, space to select, ? to toggle help",
		indexInput:    "",
		inputMode:     false,
		themeIndex:    0, // Start with DefaultTheme
		loadingChunks: make(map[int]bool),
		chunkHistory:  make([]string, 0),
		// Initialize new feature flags
		columnOrderIndex:     0, // Start with default column order
		sortingEnabled:       false,
		filteringEnabled:     false,
		currentSortField:     "",
		currentSortDir:       "",
		currentFilter:        "",
		currentFilterField:   "",
		wrappingEnabled:      false,
		fullRowCursorEnabled: false,
		selectionModeEnabled: false,
	}

	// Run the interactive program
	p := tea.NewProgram(app, tea.WithAltScreen())
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
	cmds = append(cmds, vtable.CellFormatterSetCmd(0, NameCellFormatter))     // Name column
	cmds = append(cmds, vtable.CellFormatterSetCmd(1, ValueCellFormatter))    // Value column
	cmds = append(cmds, vtable.CellFormatterSetCmd(2, StatusCellFormatter))   // Status column
	cmds = append(cmds, vtable.CellFormatterSetCmd(3, CategoryCellFormatter)) // Category column

	// Set header formatters for each column
	for columnIndex, formatter := range headerFormatters {
		cmds = append(cmds, vtable.HeaderFormatterSetCmd(columnIndex, formatter))
	}

	// Enable component renderer - now FIXED to not contaminate cell content
	cmds = append(cmds, m.table.EnableComponentRenderer())

	return tea.Batch(cmds...)
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
			m.themeIndex = (m.themeIndex + 1) % len(availableThemes)
			newTheme := availableThemes[m.themeIndex]
			SetTheme(newTheme)
			m.statusMessage = fmt.Sprintf("Theme changed to: %s - %s", newTheme.Name, newTheme.Description)

			// CRITICAL FIX: Update the table's built-in theme too!
			vtableTheme := convertToVTableTheme(newTheme)
			return m, m.table.SetTheme(vtableTheme)

		case "r":
			m.statusMessage = "Refreshing data..."
			return m, vtable.DataRefreshCmd()

		// === NEW ADVANCED FEATURES ===
		case "o":
			// Cycle through column orderings
			m.columnOrderIndex = (m.columnOrderIndex + 1) % len(availableColumnOrders)
			newColumns := availableColumnOrders[m.columnOrderIndex]
			orderNames := []string{"Default (Name→Value→Status→Category)", "Value-first", "Status-first"}
			m.statusMessage = fmt.Sprintf("Column order changed to: %s", orderNames[m.columnOrderIndex])
			return m, vtable.ColumnSetCmd(newColumns)

		case "T":
			// Toggle sorting (uppercase T for table sorting)
			if !m.sortingEnabled {
				// Enable sorting on Name field ascending
				m.sortingEnabled = true
				m.currentSortField = "name"
				m.currentSortDir = "asc"
				m.statusMessage = "Sorting enabled: Name (ascending) - press T again to cycle"
				m.dataSource.SetSort([]string{"name"}, []string{"asc"})
				return m, vtable.DataRefreshCmd()
			} else {
				// Cycle through different sorts
				switch m.currentSortField + "_" + m.currentSortDir {
				case "name_asc":
					m.currentSortField = "name"
					m.currentSortDir = "desc"
					m.statusMessage = "Sorting: Name (descending)"
					m.dataSource.SetSort([]string{"name"}, []string{"desc"})
					return m, vtable.DataRefreshCmd()
				case "name_desc":
					m.currentSortField = "value"
					m.currentSortDir = "asc"
					m.statusMessage = "Sorting: Value (ascending)"
					m.dataSource.SetSort([]string{"value"}, []string{"asc"})
					return m, vtable.DataRefreshCmd()
				case "value_asc":
					m.currentSortField = "value"
					m.currentSortDir = "desc"
					m.statusMessage = "Sorting: Value (descending)"
					m.dataSource.SetSort([]string{"value"}, []string{"desc"})
					return m, vtable.DataRefreshCmd()
				case "value_desc":
					m.currentSortField = "status"
					m.currentSortDir = "asc"
					m.statusMessage = "Sorting: Status (ascending)"
					m.dataSource.SetSort([]string{"status"}, []string{"asc"})
					return m, vtable.DataRefreshCmd()
				case "status_asc":
					// Add multi-sort: Status + Name
					m.statusMessage = "Multi-sort: Status (asc) + Name (asc)"
					m.dataSource.SetSort([]string{"status", "name"}, []string{"asc", "asc"})
					return m, vtable.DataRefreshCmd()
				default:
					// Clear sorting
					m.sortingEnabled = false
					m.currentSortField = ""
					m.currentSortDir = ""
					m.statusMessage = "Sorting disabled"
					m.dataSource.SetSort([]string{}, []string{})
					return m, vtable.DataRefreshCmd()
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
						vtable.CellFormatterSetCmd(0, createSelectionAwareFormatter(NameCellFormatter)),
						vtable.CellFormatterSetCmd(1, createSelectionAwareFormatter(ValueCellFormatter)),
						vtable.CellFormatterSetCmd(2, createSelectionAwareFormatter(StatusCellFormatter)),
						vtable.CellFormatterSetCmd(3, createSelectionAwareFormatter(CategoryCellFormatter)),
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
				return m, vtable.DataRefreshCmd()
			} else {
				// Cycle through different filters
				switch m.currentFilter {
				case "Category A":
					m.currentFilter = "Category B"
					m.statusMessage = "Filtering: Category B only"
					m.dataSource.ClearAllFilters()
					m.dataSource.SetFilter("category", "Category B")
					return m, vtable.DataRefreshCmd()
				case "Category B":
					// Switch to value filter
					m.currentFilterField = "value"
					m.currentFilter = "high" // Filter for high values
					m.statusMessage = "Filtering: High values only (>50)"
					m.dataSource.ClearAllFilters()
					m.dataSource.SetFilter("value", "high")
					return m, vtable.DataRefreshCmd()
				case "high":
					// Switch to status filter
					m.currentFilterField = "status"
					m.currentFilter = "active"
					m.statusMessage = "Filtering: Active status only"
					m.dataSource.ClearAllFilters()
					m.dataSource.SetFilter("status", "active")
					return m, vtable.DataRefreshCmd()
				default:
					// Clear filtering
					m.filteringEnabled = false
					m.currentFilterField = ""
					m.currentFilter = ""
					m.statusMessage = "Filtering disabled"
					m.dataSource.ClearAllFilters()
					return m, vtable.DataRefreshCmd()
				}
			}

		case "w":
			// Toggle text wrapping
			m.wrappingEnabled = !m.wrappingEnabled
			if m.wrappingEnabled {
				m.statusMessage = "Text wrapping enabled - long content will wrap within cells"
				// Reset first, then apply wrapping-enabled formatters
				return m, tea.Batch(
					resetAllFormattersToBase(),
					tea.Sequence(
						vtable.CellFormatterSetCmd(0, createWrappingNameFormatter()),
						vtable.CellFormatterSetCmd(1, createWrappingValueFormatter()),
						vtable.CellFormatterSetCmd(2, createWrappingStatusFormatter()),
						vtable.CellFormatterSetCmd(3, createWrappingCategoryFormatter()),
					),
				)
			} else {
				m.statusMessage = "Text wrapping disabled - long content will be truncated"
				// Restore normal formatters
				return m, resetAllFormattersToBase()
			}

		case "R":
			// Toggle full row cursor highlighting (uppercase R)
			if m.fullRowCursorEnabled {
				// Disable full row highlighting using the BUILT-IN system
				m.fullRowCursorEnabled = false
				m.statusMessage = "Full row cursor highlighting disabled"
				return m, vtable.FullRowHighlightEnableCmd(false)
			} else {
				// Enable full row highlighting using the BUILT-IN system
				m.fullRowCursorEnabled = true
				// Disable selection mode if enabled
				if m.selectionModeEnabled {
					m.selectionModeEnabled = false
				}
				m.statusMessage = "Full row cursor highlighting enabled - entire row highlighted"
				return m, vtable.FullRowHighlightEnableCmd(true)
			}

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
		view.WriteString("=== ADVANCED TABLE FORMATTER DEMO ===\n")
		view.WriteString("✨ NEW FEATURES: All advanced table features showcased!\n\n")

		view.WriteString("🎨 VISUAL FEATURES:\n")
		view.WriteString("• Cell Formatters: Each column styled with colors + icons + auto-truncation\n")
		view.WriteString("• Header Formatters: Emoji headers (📝💰📊🏷️) with perfect Unicode alignment\n")
		view.WriteString("• Component Renderer: Clean cursor (►) + selection (✓) indicators\n")
		view.WriteString("• Multi-Themes: t=cycle (Default→Dark→Minimal→Neon)\n\n")

		view.WriteString("📊 DATA FEATURES:\n")
		view.WriteString("• Column Ordering: o=cycle column arrangements\n")
		view.WriteString("• Sort/Multi-Sort: T=cycle sorts (Name↑↓→Value↑↓→Status↑→Multi-Sort→Clear)\n")
		view.WriteString("• Filter/Multi-Filter: F=cycle filters (CategoryA→CategoryB→HighValues→ActiveStatus→Clear)\n\n")

		view.WriteString("🖱️  INTERACTION:\n")
		view.WriteString("• Navigation: j/k↑↓=move • h/l=page • g/G=start/end • J=jump • 1-5=quick jumps\n")
		view.WriteString("• Selection: Space=toggle • a=select all • c=clear • s=show count\n")
		view.WriteString("• Text Wrapping: w=toggle (wrap vs truncate long content)\n")
		view.WriteString("• Selection Highlighting: S=toggle selection background highlighting\n")
		view.WriteString("• Full Row Highlighting: R=toggle full row highlighting\n\n")

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

		view.WriteString(fmt.Sprintf("• Text Wrapping: %s • Selection Mode: %s • Full Row: %s\n\n",
			map[bool]string{true: "Enabled", false: "Disabled"}[m.wrappingEnabled],
			map[bool]string{true: "Enabled", false: "Disabled"}[m.selectionModeEnabled],
			map[bool]string{true: "Enabled", false: "Disabled"}[m.fullRowCursorEnabled]))
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
		Theme:         convertToVTableTheme(currentTheme),
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
