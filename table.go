package vtable

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TableRow represents a row of data in the table.
type TableRow struct {
	// Cells contains the string values for each column in the row.
	Cells []string
}

// Table is a virtualized table component.
type Table struct {
	// config is the configuration for the table.
	config TableConfig

	// theme defines the visual appearance of the table.
	theme Theme

	// list is the underlying list component.
	list *List[TableRow]

	// totalWidth is the total width of the table.
	totalWidth int

	// horizontalBorderString is the horizontal border for the entire table.
	horizontalBorderTop    string
	horizontalBorderMiddle string
	horizontalBorderBottom string
}

// NewTable creates a new virtualized table component.
func NewTable(
	config TableConfig,
	provider DataProvider[TableRow],
	theme Theme,
) (*Table, error) {
	// Auto-fix any configuration issues instead of failing
	err := ValidateAndFixTableConfig(&config)
	if err != nil {
		return nil, err
	}

	// Get actual data size
	dataSize := provider.GetTotal()

	// Create a copy of the config to avoid modifying the original
	adjustedConfig := config

	// Adjust viewport height if it's larger than the dataset
	// This prevents empty rows from showing
	if dataSize < config.ViewportConfig.Height {
		adjustedConfig.ViewportConfig.Height = max(1, dataSize)
		// Recalculate thresholds for the new height
		ValidateAndFixViewportConfig(&adjustedConfig.ViewportConfig)
	}

	// Calculate the total width of the table
	totalWidth := 0
	for _, col := range adjustedConfig.Columns {
		totalWidth += col.Width
	}

	// Add space for borders if needed
	if adjustedConfig.ShowBorders {
		// Add vertical borders (one per column + one for the end)
		totalWidth += len(adjustedConfig.Columns) + 1
	}

	// Create horizontal border strings with the theme's border characters
	var horizontalBorderTop, horizontalBorderMiddle, horizontalBorderBottom string
	if adjustedConfig.ShowBorders {
		var topBuilder, middleBuilder, bottomBuilder strings.Builder

		for i, col := range adjustedConfig.Columns {
			// Top border
			if i == 0 {
				topBuilder.WriteString(theme.BorderChars.TopLeft)
			}
			for j := 0; j < col.Width; j++ {
				topBuilder.WriteString(theme.BorderChars.Horizontal)
			}
			if i == len(adjustedConfig.Columns)-1 {
				topBuilder.WriteString(theme.BorderChars.TopRight)
			} else {
				topBuilder.WriteString(theme.BorderChars.TopT)
			}

			// Middle border (separator)
			if i == 0 {
				middleBuilder.WriteString(theme.BorderChars.LeftT)
			}
			for j := 0; j < col.Width; j++ {
				middleBuilder.WriteString(theme.BorderChars.Horizontal)
			}
			if i == len(adjustedConfig.Columns)-1 {
				middleBuilder.WriteString(theme.BorderChars.RightT)
			} else {
				middleBuilder.WriteString(theme.BorderChars.Cross)
			}

			// Bottom border
			if i == 0 {
				bottomBuilder.WriteString(theme.BorderChars.BottomLeft)
			}
			for j := 0; j < col.Width; j++ {
				bottomBuilder.WriteString(theme.BorderChars.Horizontal)
			}
			if i == len(adjustedConfig.Columns)-1 {
				bottomBuilder.WriteString(theme.BorderChars.BottomRight)
			} else {
				bottomBuilder.WriteString(theme.BorderChars.BottomT)
			}
		}

		// Create proper styles for borders
		borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.BorderColor))
		horizontalBorderTop = borderStyle.Render(topBuilder.String())
		horizontalBorderMiddle = borderStyle.Render(middleBuilder.String())
		horizontalBorderBottom = borderStyle.Render(bottomBuilder.String())
	}

	// Create a formatter for the table rows
	formatter := func(data Data[TableRow], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		return formatTableRow(data, index, isCursor, isTopThreshold, isBottomThreshold, adjustedConfig, theme)
	}

	// Convert theme to styleConfig for backward compatibility with List
	styleConfig := ThemeToStyleConfig(&theme)

	// Create the underlying list
	list, err := NewList(adjustedConfig.ViewportConfig, provider, styleConfig, formatter)
	if err != nil {
		return nil, err
	}

	return &Table{
		config:                 adjustedConfig,
		theme:                  theme,
		list:                   list,
		totalWidth:             totalWidth,
		horizontalBorderTop:    horizontalBorderTop,
		horizontalBorderMiddle: horizontalBorderMiddle,
		horizontalBorderBottom: horizontalBorderBottom,
	}, nil
}

// NewSimpleTable creates a table with just columns and reasonable defaults.
// This is the easiest way to create a table - just provide columns and a data provider.
// Example: table, err := vtable.NewSimpleTable(columns, provider)
func NewSimpleTable(columns []TableColumn, provider DataProvider[TableRow]) (*Table, error) {
	config := NewSimpleTableConfig(columns)
	theme := *DefaultTheme()
	return NewTable(config, provider, theme)
}

// NewTableWithHeight creates a table with specified viewport height and auto-calculated thresholds.
// Example: table, err := vtable.NewTableWithHeight(columns, provider, 15)
func NewTableWithHeight(columns []TableColumn, provider DataProvider[TableRow], height int) (*Table, error) {
	config := NewTableConfig(columns, height)
	theme := *DefaultTheme()
	return NewTable(config, provider, theme)
}

// NewTableWithTheme creates a table with a custom theme.
// Example: table, err := vtable.NewTableWithTheme(columns, provider, vtable.DarkTheme())
func NewTableWithTheme(columns []TableColumn, provider DataProvider[TableRow], theme *Theme) (*Table, error) {
	config := NewSimpleTableConfig(columns)
	return NewTable(config, provider, *theme)
}

// NewTableWithHeightAndTheme creates a table with custom height and theme.
// Example: table, err := vtable.NewTableWithHeightAndTheme(columns, provider, 15, vtable.DarkTheme())
func NewTableWithHeightAndTheme(columns []TableColumn, provider DataProvider[TableRow], height int, theme *Theme) (*Table, error) {
	config := NewTableConfig(columns, height)
	return NewTable(config, provider, *theme)
}

// CellConstraint represents the constraints for a cell
type CellConstraint struct {
	Width     int
	Height    int // Currently only supports Height=1 (single line). Multi-line support TODO.
	Alignment int // Use vtable alignment constants
}

// enforceCellConstraints ensures text fits exactly within cell constraints
// This function handles padding, truncation, and alignment to produce exactly the required width
func enforceCellConstraints(text string, constraint CellConstraint) string {
	// Handle multi-line content by converting to single line for all cases
	// Convert line breaks to spaces
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
		case AlignCenter:
			leftPad := padding / 2
			rightPad := padding - leftPad
			text = strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
		case AlignRight:
			text = strings.Repeat(" ", padding) + text
		default: // AlignLeft
			text = text + strings.Repeat(" ", padding)
		}
	}

	return text
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FormatTableRow formats a table row for display with the given configuration.
// This is exported for use in custom animation formatters.
func FormatTableRow(
	data Data[TableRow],
	index int,
	isCursor bool,
	isTopThreshold bool,
	isBottomThreshold bool,
	config TableConfig,
	theme Theme,
) string {
	var sb strings.Builder

	row := data.Item
	isSelected := data.Selected

	// Ensure we don't iterate beyond the row's cell count
	cellCount := len(row.Cells)
	columnCount := len(config.Columns)

	// Format cells
	for i := 0; i < columnCount; i++ {
		// Add starting border if needed
		if config.ShowBorders && i == 0 {
			borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.BorderColor))
			if isCursor || isSelected {
				// Highlight border for cursor or selected row
				borderStyle = borderStyle.Bold(true)
			}
			sb.WriteString(borderStyle.Render(theme.BorderChars.Vertical))
		}

		// Determine the style to use (don't set width - enforceCellConstraints handles that)
		var style lipgloss.Style
		if isCursor && isSelected {
			// Both cursor and selected: Use a special combined style (bold selected style)
			style = theme.SelectedRowStyle.Copy().Bold(true)
		} else if isCursor {
			// Just cursor: Use selected style for cursor row
			style = theme.SelectedRowStyle.Copy()
		} else if isSelected {
			// Just selected: Use a modified style to show selection
			style = theme.RowEvenStyle.Copy().
				Background(lipgloss.Color("22")). // Dark green background for selected
				Foreground(lipgloss.Color("15"))  // White text
		} else if isTopThreshold {
			// Apply threshold styling if needed
			style = theme.RowStyle.Copy()
		} else if isBottomThreshold {
			// Apply threshold styling if needed
			style = theme.RowStyle.Copy()
		} else if index%2 == 0 {
			// Even rows
			style = theme.RowEvenStyle.Copy()
		} else {
			// Odd rows
			style = theme.RowOddStyle.Copy()
		}

		// Don't set alignment on lipgloss style - enforceCellConstraints handles that too

		// Get the cell value, or use empty string if this cell doesn't exist
		var value string
		if i < cellCount {
			value = row.Cells[i]
		}

		// Add selection indicator to the first column
		if i == 0 && isSelected {
			if isCursor {
				value = "✓>" + value // Both selected and cursor
			} else {
				value = "✓ " + value // Just selected
			}
		}

		// Apply cell constraints to ensure content fits within column boundaries
		constraint := CellConstraint{
			Width:     config.Columns[i].Width,
			Height:    1, // TODO: Support multi-line cells
			Alignment: config.Columns[i].Alignment,
		}
		value = enforceCellConstraints(value, constraint)

		// Render the cell
		sb.WriteString(style.Render(value))

		// Add border if needed
		if config.ShowBorders {
			borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.BorderColor))
			if isCursor || isSelected {
				// Highlight border for cursor or selected row
				borderStyle = borderStyle.Bold(true)
			}
			sb.WriteString(borderStyle.Render(theme.BorderChars.Vertical))
		}
	}

	return sb.String()
}

// formatTableRow is the internal function that calls FormatTableRow
func formatTableRow(
	data Data[TableRow],
	index int,
	isCursor bool,
	isTopThreshold bool,
	isBottomThreshold bool,
	config TableConfig,
	theme Theme,
) string {
	return FormatTableRow(data, index, isCursor, isTopThreshold, isBottomThreshold, config, theme)
}

// formatHeader formats the header row.
func (t *Table) formatHeader() string {
	var sb strings.Builder

	// Add top border if needed
	if t.config.ShowBorders {
		sb.WriteString(t.horizontalBorderTop)
		sb.WriteString("\n")
	}

	// Get current sort state from the list
	sortFields := t.list.dataRequest.SortFields
	sortDirections := t.list.dataRequest.SortDirections

	// Format header cells
	for i, col := range t.config.Columns {
		// Start with border if needed
		if t.config.ShowBorders && i == 0 {
			sb.WriteString(t.theme.HeaderBorderStyle.Render(t.theme.BorderChars.Vertical))
		}

		// Check if this column is being sorted
		// First try matching by Field ID if present
		sortDirection := ""
		columnFieldID := col.Field

		// If no Field ID is set, use the column index as a fallback
		if columnFieldID == "" {
			columnFieldID = strconv.Itoa(i)
		}

		// Find if this column is in the sort fields
		for j, field := range sortFields {
			if field == columnFieldID {
				sortDirection = sortDirections[j]
				break
			}
		}

		// Format the header cell using constraint-based approach instead of lipgloss .Width()
		style := t.theme.HeaderStyle.Copy()

		// Add sort indicator to title if this column is sorted
		title := col.Title
		if sortDirection != "" {
			if sortDirection == "asc" {
				title = title + " ↑" // Up arrow for ascending
			} else {
				title = title + " ↓" // Down arrow for descending
			}
		}

		// Apply cell constraints to header text (similar to data cells)
		constraint := CellConstraint{
			Width:     col.Width,
			Height:    1,
			Alignment: col.Alignment,
		}
		constrainedTitle := enforceCellConstraints(title, constraint)

		sb.WriteString(style.Render(constrainedTitle))

		// Add border if needed
		if t.config.ShowBorders {
			sb.WriteString(t.theme.HeaderBorderStyle.Render(t.theme.BorderChars.Vertical))
		}
	}

	// Add middle border if needed
	if t.config.ShowBorders {
		sb.WriteString("\n")
		sb.WriteString(t.horizontalBorderMiddle)
	}

	return sb.String()
}

// Render renders the table to a string.
func (t *Table) Render() string {
	return t.RenderWithAnimatedContent(nil, nil)
}

// RenderWithAnimatedContent renders the table with optional animated content and default formatter
func (t *Table) RenderWithAnimatedContent(animatedContent map[string]string, defaultFormatter ItemFormatter[TableRow]) string {
	var sb strings.Builder

	// Add header if needed
	if t.config.ShowHeader {
		header := t.formatHeader()
		sb.WriteString(header)
		sb.WriteString("\n")
	}

	// Count actual rows in the dataset
	actualRows := t.list.totalItems
	if actualRows <= 0 {
		// If there's no data, just render the header and bottom border if needed
		if t.config.ShowBorders {
			sb.WriteString(t.horizontalBorderBottom)
		}
		return sb.String()
	}

	// If no animated content provided, use the existing list rendering
	if animatedContent == nil || len(animatedContent) == 0 {
		// Render the list content - don't render beyond actual data
		list := t.list.Render()
		sb.WriteString(list)

		// Add bottom border if needed - directly after the last data row
		if t.config.ShowBorders {
			sb.WriteString("\n")
			sb.WriteString(t.horizontalBorderBottom)
		}

		return sb.String()
	}

	// Use animated content for rendering
	visibleItems := t.list.GetVisibleItems()
	state := t.list.GetState()

	// Render each visible row with animated content
	for i, item := range visibleItems {
		absoluteIndex := state.ViewportStartIndex + i

		// Skip if we've rendered all real data
		if absoluteIndex >= actualRows {
			break
		}

		isCursor := i == state.CursorViewportIndex
		isTopThreshold := i == t.config.ViewportConfig.TopThresholdIndex
		isBottomThreshold := i == t.config.ViewportConfig.BottomThresholdIndex

		var renderedRow string

		// Check if we have animated content for this row
		animKey := fmt.Sprintf("row-%d", absoluteIndex)
		if animatedContent[animKey] != "" {
			renderedRow = animatedContent[animKey]
		} else {
			// Use default formatter if provided, otherwise fallback to regular table row formatter
			if defaultFormatter != nil {
				ctx := DefaultRenderContext()
				ctx.CurrentTime = time.Now()
				ctx.MaxWidth = t.totalWidth
				ctx.Theme = &t.theme
				renderedRow = defaultFormatter(item, absoluteIndex, ctx, isCursor, isTopThreshold, isBottomThreshold)
			} else {
				// Fallback to regular table row formatter
				renderedRow = formatTableRow(item, absoluteIndex, isCursor, isTopThreshold, isBottomThreshold, t.config, t.theme)
			}
		}

		sb.WriteString(renderedRow)

		// Add a newline unless it's the last row
		if i < len(visibleItems)-1 && absoluteIndex < actualRows-1 {
			sb.WriteString("\n")
		}
	}

	// Add bottom border if needed - directly after the last data row
	if t.config.ShowBorders {
		sb.WriteString("\n")
		sb.WriteString(t.horizontalBorderBottom)
	}

	return sb.String()
}

// GetState returns the current viewport state.
func (t *Table) GetState() ViewportState {
	return t.list.GetState()
}

// GetCurrentRow returns the currently selected row.
func (t *Table) GetCurrentRow() (TableRow, bool) {
	data, ok := t.list.GetCurrentItem()
	if !ok {
		return TableRow{}, false
	}
	return data.Item, true
}

// GetConfig returns the current table configuration.
func (t *Table) GetConfig() TableConfig {
	return t.config
}

// GetTheme returns the current theme.
func (t *Table) GetTheme() Theme {
	return t.theme
}

// RenderDebugInfo renders debug information about the table.
func (t *Table) RenderDebugInfo() string {
	return t.list.RenderDebugInfo()
}

// SetFilter sets a filter for a specific field.
// Applying a filter will reload all data.
func (t *Table) SetFilter(field string, value any) {
	t.list.SetFilter(field, value)
}

// ClearFilters removes all filters.
// This will reload all data.
func (t *Table) ClearFilters() {
	t.list.ClearFilters()
}

// RemoveFilter removes a specific filter.
// This will reload all data.
func (t *Table) RemoveFilter(field string) {
	t.list.RemoveFilter(field)
}

// SetSort sets a sort field and direction.
// Direction should be "asc" or "desc".
// Applying a sort will reload all data.
func (t *Table) SetSort(field string, direction string) {
	t.list.SetSort(field, direction)
}

// AddSort adds a sort field and direction without clearing existing sorts.
// This allows for multi-column sorting.
func (t *Table) AddSort(field string, direction string) {
	t.list.AddSort(field, direction)
}

// RemoveSort removes a specific sort field.
func (t *Table) RemoveSort(field string) {
	t.list.RemoveSort(field)
}

// ClearSort removes any sorting criteria.
// This will reload all data.
func (t *Table) ClearSort() {
	t.list.ClearSort()
}

// GetDataRequest returns the current data request configuration.
func (t *Table) GetDataRequest() DataRequest {
	return t.list.GetDataRequest()
}

// All navigation methods delegate to the underlying list

func (t *Table) MoveUp() {
	t.list.MoveUp()
}

func (t *Table) MoveDown() {
	t.list.MoveDown()
}

func (t *Table) PageUp() {
	t.list.PageUp()
}

func (t *Table) PageDown() {
	t.list.PageDown()
}

func (t *Table) JumpToIndex(index int) {
	t.list.JumpToIndex(index)
}

func (t *Table) JumpToStart() {
	t.list.JumpToStart()
}

func (t *Table) JumpToEnd() {
	t.list.JumpToEnd()
}

func (t *Table) JumpToItem(key string, value any) bool {
	return t.list.JumpToItem(key, value)
}

// TeaTable is a Bubble Tea model wrapping a Table.
type TeaTable struct {
	// The underlying table model
	table *Table

	// Key mappings
	keyMap NavigationKeyMap

	// Whether the component is focused
	focused bool

	// Optional callbacks
	onSelectRow      func(row TableRow, index int)
	onScroll         func(state ViewportState)
	onHighlight      func(row TableRow, index int)
	onFiltersChanged func(filters map[string]any)
	onSortChanged    func(field, direction string)

	// Animation support
	animationEngine   *AnimationEngine
	animatedFormatter ItemFormatterAnimated[TableRow]
	defaultFormatter  ItemFormatter[TableRow] // Optional formatter for when animations are disabled/not active
	animationConfig   AnimationConfig
	lastAnimationTime time.Time

	// Animation behavior controls
	animateOnlyCursorRow bool // If true, only animate the cursor row; if false, animate all visible rows

	// Real-time data updates
	realTimeUpdates    bool
	realTimeInterval   time.Duration
	lastRealTimeUpdate time.Time

	// Animation content cache
	cachedAnimationContent map[string]string

	// Track cursor position for cache invalidation
	lastCursorIndex int
}

// NewTeaTable creates a new Bubble Tea model for a virtualized table.
func NewTeaTable(
	config TableConfig,
	provider DataProvider[TableRow],
	theme Theme,
) (*TeaTable, error) {
	// Create the underlying table
	table, err := NewTable(config, provider, theme)
	if err != nil {
		return nil, err
	}

	// Initialize animation system
	animConfig := DefaultAnimationConfig()
	animEngine := NewAnimationEngine(animConfig)

	return &TeaTable{
		table:                  table,
		keyMap:                 PlatformKeyMap(), // Use platform-specific key bindings
		focused:                true,
		animationEngine:        animEngine,
		animationConfig:        animConfig,
		animateOnlyCursorRow:   true, // By default, only animate cursor row
		cachedAnimationContent: make(map[string]string),
		lastCursorIndex:        0, // Initialize cursor tracking
	}, nil
}

// NewSimpleTeaTable creates a Bubble Tea table with just columns and reasonable defaults.
// This is the easiest way to create a Bubble Tea table - just provide columns and a data provider.
// Example: table, err := vtable.NewSimpleTeaTable(columns, provider)
func NewSimpleTeaTable(columns []TableColumn, provider DataProvider[TableRow]) (*TeaTable, error) {
	config := NewSimpleTableConfig(columns)
	theme := *DefaultTheme()
	return NewTeaTable(config, provider, theme)
}

// NewTeaTableWithHeight creates a Bubble Tea table with specified viewport height.
// Example: table, err := vtable.NewTeaTableWithHeight(columns, provider, 15)
func NewTeaTableWithHeight(columns []TableColumn, provider DataProvider[TableRow], height int) (*TeaTable, error) {
	config := NewTableConfig(columns, height)
	theme := *DefaultTheme()
	return NewTeaTable(config, provider, theme)
}

// NewTeaTableWithTheme creates a Bubble Tea table with a custom theme.
// Example: table, err := vtable.NewTeaTableWithTheme(columns, provider, vtable.DarkTheme())
func NewTeaTableWithTheme(columns []TableColumn, provider DataProvider[TableRow], theme *Theme) (*TeaTable, error) {
	config := NewSimpleTableConfig(columns)
	return NewTeaTable(config, provider, *theme)
}

// NewTeaTableWithHeightAndTheme creates a Bubble Tea table with custom height and theme.
// Example: table, err := vtable.NewTeaTableWithHeightAndTheme(columns, provider, 15, vtable.DarkTheme())
func NewTeaTableWithHeightAndTheme(columns []TableColumn, provider DataProvider[TableRow], height int, theme *Theme) (*TeaTable, error) {
	config := NewTableConfig(columns, height)
	return NewTeaTable(config, provider, *theme)
}

// Init initializes the Tea model.
func (m *TeaTable) Init() tea.Cmd {
	// Start the global animation loop
	return StartGlobalAnimationLoop()
}

// Update updates the Tea model based on messages.
func (m *TeaTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If not focused, don't handle messages
	if !m.focused {
		return m, nil
	}

	var cmds []tea.Cmd
	previousState := m.table.GetState()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Up):
			m.table.MoveUp()
		case key.Matches(msg, m.keyMap.Down):
			m.table.MoveDown()
		case key.Matches(msg, m.keyMap.PageUp):
			m.table.PageUp()
		case key.Matches(msg, m.keyMap.PageDown):
			m.table.PageDown()
		case key.Matches(msg, m.keyMap.Home):
			m.table.JumpToStart()
		case key.Matches(msg, m.keyMap.End):
			m.table.JumpToEnd()
		case key.Matches(msg, m.keyMap.Select):
			if m.onSelectRow != nil {
				if row, ok := m.GetCurrentRow(); ok {
					m.onSelectRow(row, m.GetState().CursorIndex)
				}
			}
		}
	case FilterMsg:
		previousFilters := make(map[string]any)
		for k, v := range m.table.list.dataRequest.Filters {
			previousFilters[k] = v
		}

		// Handle the filter message
		if msg.Clear {
			m.table.ClearFilters()
		} else if msg.Remove {
			m.table.RemoveFilter(msg.Field)
		} else {
			m.table.SetFilter(msg.Field, msg.Value)
		}

		// After filter changes, ensure the visual state is properly updated
		// If the number of items changes dramatically, we may need to adjust cursor position
		if m.table.list.totalItems == 0 {
			// No matching items after filter
			cmds = append(cmds, func() tea.Msg {
				return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("home")}
			})
		} else if m.table.list.totalItems <= m.table.list.Config.Height {
			// Small enough dataset to show everything, jump to start
			cmds = append(cmds, func() tea.Msg {
				return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("home")}
			})
		}

		// Call the filter changed callback if filters changed
		if m.onFiltersChanged != nil {
			hasChanged := len(previousFilters) != len(m.table.list.dataRequest.Filters)
			if !hasChanged {
				// Check if any values changed
				for k, v := range previousFilters {
					if newV, ok := m.table.list.dataRequest.Filters[k]; !ok || newV != v {
						hasChanged = true
						break
					}
				}
			}

			if hasChanged {
				m.onFiltersChanged(m.table.list.dataRequest.Filters)
			}
		}
	case SortMsg:
		// Store previous sorts for callback comparison
		previousSortFields := make([]string, len(m.table.list.dataRequest.SortFields))
		previousSortDirections := make([]string, len(m.table.list.dataRequest.SortDirections))
		copy(previousSortFields, m.table.list.dataRequest.SortFields)
		copy(previousSortDirections, m.table.list.dataRequest.SortDirections)

		// Handle the sort message
		if msg.Clear {
			m.table.ClearSort()
		} else if msg.Remove {
			m.table.RemoveSort(msg.Field)
		} else if msg.Add {
			m.table.AddSort(msg.Field, msg.Direction)
		} else {
			m.table.SetSort(msg.Field, msg.Direction)
		}

		// Call the sort changed callback if sort changed
		if m.onSortChanged != nil {
			// Check if sorting has changed
			changed := len(previousSortFields) != len(m.table.list.dataRequest.SortFields)
			if !changed {
				for i, field := range previousSortFields {
					if i >= len(m.table.list.dataRequest.SortFields) ||
						field != m.table.list.dataRequest.SortFields[i] ||
						previousSortDirections[i] != m.table.list.dataRequest.SortDirections[i] {
						changed = true
						break
					}
				}
			}

			if changed && len(m.table.list.dataRequest.SortFields) > 0 {
				m.onSortChanged(
					strings.Join(m.table.list.dataRequest.SortFields, ","),
					strings.Join(m.table.list.dataRequest.SortDirections, ","),
				)
			} else if changed {
				m.onSortChanged("", "")
			}
		}
	case GlobalAnimationTickMsg:
		// Handle global animation tick - this runs continuously while animations are active
		if cmd := m.animationEngine.ProcessGlobalTick(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Handle real-time data updates if enabled
		if m.realTimeUpdates {
			now := msg.Timestamp
			if now.Sub(m.lastRealTimeUpdate) >= m.realTimeInterval {
				// Time for a real-time data refresh
				m.lastRealTimeUpdate = now
				m.ForceDataRefresh()
			}
		}
	case AnimationUpdateMsg:
		// Animations have been updated - trigger re-render by doing nothing
		// The View() method will automatically pick up the changes

		// CRITICAL FIX: Only update animation content when we receive animation update messages
		// This decouples animation updates from cursor movements
		if m.animatedFormatter != nil {
			m.updateAnimationContent()
		}
	}

	// Check if we need to trigger callbacks based on state changes
	currentState := m.table.GetState()

	// Call onScroll if viewport changed
	if m.onScroll != nil && (previousState.ViewportStartIndex != currentState.ViewportStartIndex) {
		m.onScroll(currentState)
	}

	// Call onHighlight if highlighted item changed
	if m.onHighlight != nil && previousState.CursorIndex != currentState.CursorIndex {
		if row, ok := m.GetCurrentRow(); ok {
			m.onHighlight(row, currentState.CursorIndex)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the Tea model.
func (m *TeaTable) View() string {
	// Handle animations if we have an animated formatter
	if m.animatedFormatter != nil {
		// Check if cursor position has changed
		currentCursorIndex := m.table.GetState().CursorIndex
		if currentCursorIndex != m.lastCursorIndex {
			// Cursor moved - update cache immediately for smooth movement
			m.updateCursorInCache(m.lastCursorIndex, currentCursorIndex)
			m.lastCursorIndex = currentCursorIndex
		}

		// Ensure cache is populated for all visible items
		m.ensureAnimationCachePopulated()

		// Use cached animation content - only updated when animation update messages are received
		return m.table.RenderWithAnimatedContent(m.cachedAnimationContent, m.defaultFormatter)
	}

	return m.table.Render()
}

// processAnimations handles the animation lifecycle for visible items
func (m *TeaTable) processAnimations() {
	// Don't process animations if they're disabled or no formatter is set
	if !m.animationConfig.Enabled || m.animatedFormatter == nil {
		return
	}

	visibleItems := m.table.list.GetVisibleItems()
	state := m.table.GetState()

	// Track which animations should be active
	activeAnimationKeys := make(map[string]bool)

	// Calculate delta time
	now := time.Now()
	deltaTime := time.Duration(0)
	if !m.lastAnimationTime.IsZero() {
		deltaTime = now.Sub(m.lastAnimationTime)
	}
	m.lastAnimationTime = now

	// Process each visible item
	for i, dataItem := range visibleItems {
		absoluteIndex := state.ViewportStartIndex + i
		animationKey := fmt.Sprintf("row-%d", absoluteIndex)
		activeAnimationKeys[animationKey] = true

		// Create render context with table-specific information and delta time
		ctx := DefaultRenderContext()
		ctx.CurrentTime = now
		ctx.DeltaTime = deltaTime
		ctx.MaxWidth = m.table.totalWidth
		ctx.Theme = &m.table.theme

		// Get animation state
		animState := m.animationEngine.GetAnimationState(animationKey)

		// Determine cursor state
		isCursor := i == state.CursorViewportIndex
		isTopThreshold := i == m.table.list.Config.TopThresholdIndex
		isBottomThreshold := i == m.table.list.Config.BottomThresholdIndex

		// Call animated formatter
		result := m.animatedFormatter(dataItem, absoluteIndex, ctx, animState, isCursor, isTopThreshold, isBottomThreshold)

		// CRITICAL FIX: Only register animations ONCE when they first become visible
		// Do NOT re-register on every view render (this causes acceleration)
		if len(result.RefreshTriggers) > 0 && !m.animationEngine.IsVisible(animationKey) {
			// Register animation only if it doesn't exist yet
			if cmd := m.animationEngine.RegisterAnimation(animationKey, result.RefreshTriggers, result.AnimationState); cmd != nil {
				// Animation loop started - this only happens once
				_ = cmd
			}
		}

		// Update animation state ONLY if it actually changed
		// Don't update state on every render to prevent animation reset
		if len(result.AnimationState) > 0 {
			currentState := m.animationEngine.GetAnimationState(animationKey)
			hasChanges := false

			// Check if state actually changed
			for k, newValue := range result.AnimationState {
				if currentValue, exists := currentState[k]; !exists || !deepEqual(currentValue, newValue) {
					hasChanges = true
					break
				}
			}

			// Only update if there are actual changes
			if hasChanges {
				m.animationEngine.UpdateAnimationState(animationKey, result.AnimationState)
			}
		}

		// Make sure the animation is visible (this is safe to call repeatedly)
		m.animationEngine.SetVisible(animationKey, true)
	}

	// Clean up animations for items that are no longer visible
	activeAnimations := m.animationEngine.GetActiveAnimations()
	for _, animKey := range activeAnimations {
		if !activeAnimationKeys[animKey] {
			m.animationEngine.SetVisible(animKey, false)
		}
	}
}

// Focus sets the focus state of the component.
func (m *TeaTable) Focus() {
	m.focused = true
}

// Blur removes focus from the component.
func (m *TeaTable) Blur() {
	m.focused = false
}

// IsFocused returns whether the component is focused.
func (m *TeaTable) IsFocused() bool {
	return m.focused
}

// GetState returns the current viewport state.
func (m *TeaTable) GetState() ViewportState {
	return m.table.GetState()
}

// GetCurrentRow returns the currently selected row.
func (m *TeaTable) GetCurrentRow() (TableRow, bool) {
	return m.table.GetCurrentRow()
}

// SetKeyMap sets the key mappings for the component.
func (m *TeaTable) SetKeyMap(keyMap NavigationKeyMap) {
	m.keyMap = keyMap
}

// GetKeyMap returns the current key mappings for the component.
func (m *TeaTable) GetKeyMap() NavigationKeyMap {
	return m.keyMap
}

// SetFilter sets a filter for a specific field.
func (m *TeaTable) SetFilter(field string, value any) {
	m.table.SetFilter(field, value)
}

// ClearFilters removes all filters.
func (m *TeaTable) ClearFilters() {
	m.table.ClearFilters()
}

// RemoveFilter removes a specific filter.
func (m *TeaTable) RemoveFilter(field string) {
	m.table.RemoveFilter(field)
}

// SetSort sets a sort field and direction, clearing any existing sorts.
func (m *TeaTable) SetSort(field, direction string) {
	m.table.SetSort(field, direction)
}

// AddSort adds a sort field and direction without clearing existing sorts.
// This allows for multi-column sorting.
func (m *TeaTable) AddSort(field, direction string) {
	m.table.AddSort(field, direction)
}

// RemoveSort removes a specific sort.
func (m *TeaTable) RemoveSort(field string) {
	m.table.RemoveSort(field)
}

// ClearSort removes any sorting criteria.
func (m *TeaTable) ClearSort() {
	m.table.ClearSort()
}

// GetDataRequest returns the current data request configuration.
func (m *TeaTable) GetDataRequest() DataRequest {
	return m.table.GetDataRequest()
}

// OnFiltersChanged sets a callback function that will be called when filters change.
func (m *TeaTable) OnFiltersChanged(callback func(filters map[string]any)) {
	m.onFiltersChanged = callback
}

// OnSortChanged sets a callback function that will be called when sorting changes.
func (m *TeaTable) OnSortChanged(callback func(field, direction string)) {
	m.onSortChanged = callback
}

// MoveUp moves the cursor up one position.
func (m *TeaTable) MoveUp() {
	m.table.MoveUp()
}

// MoveDown moves the cursor down one position.
func (m *TeaTable) MoveDown() {
	m.table.MoveDown()
}

// PageUp moves the cursor up by a page.
func (m *TeaTable) PageUp() {
	m.table.PageUp()
}

// PageDown moves the cursor down by a page.
func (m *TeaTable) PageDown() {
	m.table.PageDown()
}

// JumpToItem jumps to a row with the specified key-value pair.
// Returns true if the row was found and jumped to, false otherwise.
func (m *TeaTable) JumpToItem(key string, value any) bool {
	return m.table.JumpToItem(key, value)
}

// JumpToIndex jumps to the specified index in the dataset.
func (m *TeaTable) JumpToIndex(index int) {
	m.table.JumpToIndex(index)
}

// JumpToStart jumps to the start of the dataset.
func (m *TeaTable) JumpToStart() {
	m.table.JumpToStart()
}

// JumpToEnd jumps to the end of the dataset.
func (m *TeaTable) JumpToEnd() {
	m.table.JumpToEnd()
}

// GetHelpView returns a string describing the key bindings.
func (m *TeaTable) GetHelpView() string {
	return GetKeyMapDescription(m.keyMap)
}

// RenderDebugInfo renders debug information about the table.
func (m *TeaTable) RenderDebugInfo() string {
	return m.table.RenderDebugInfo()
}

// SetTheme updates the theme without recreating the table.
// This is much better than creating a new table when only the theme changes.
func (m *TeaTable) SetTheme(theme Theme) {
	// Update theme
	m.table.theme = theme

	// Re-calculate borders with new theme
	m.table.recalculateBorders()
}

// SetAnimatedFormatter sets a formatter that supports animations for the entire row.
func (m *TeaTable) SetAnimatedFormatter(formatter ItemFormatterAnimated[TableRow]) {
	m.animatedFormatter = formatter
	// Initialize animation content cache if needed
	if m.cachedAnimationContent == nil {
		m.cachedAnimationContent = make(map[string]string)
	}
}

// SetDefaultFormatter sets a formatter to use for rows when animations are not active.
// This gives you control over how non-animated rows are rendered.
// If no default formatter is set, rows fall back to the standard table formatter.
func (m *TeaTable) SetDefaultFormatter(formatter ItemFormatter[TableRow]) {
	m.defaultFormatter = formatter
}

// ClearDefaultFormatter removes the default formatter, causing non-animated rows
// to use the standard table formatter.
func (m *TeaTable) ClearDefaultFormatter() {
	m.defaultFormatter = nil
}

// SetCellAnimatedFormatter sets a formatter that supports animations for individual cells.
// This enables true cell-level animations like horizontal text scrolling, smooth transitions, etc.
// Note: This is a placeholder for future cell-level animation implementation
func (m *TeaTable) SetCellAnimatedFormatter(formatter CellFormatterAnimated) {
	// TODO: Implement cell-level animation support
	// For now, this is a placeholder that demonstrates the API design
	// Full implementation would require extending the table rendering system
}

// ClearAnimatedFormatter removes any animated formatter and falls back to the default formatter.
func (m *TeaTable) ClearAnimatedFormatter() {
	m.animatedFormatter = nil
	m.animationEngine.Cleanup()
	// Explicitly stop the loop since we're no longer using animations
	m.animationEngine.StopLoop()
}

// SetAnimationConfig updates the animation configuration.
func (m *TeaTable) SetAnimationConfig(config AnimationConfig) tea.Cmd {
	m.animationConfig = config
	return m.animationEngine.UpdateConfig(config)
}

// EnableAnimations enables the animation system and starts the loop if needed.
func (m *TeaTable) EnableAnimations() tea.Cmd {
	m.animationConfig.Enabled = true

	// Clear cached content to force fresh rendering when animations are re-enabled
	if m.animatedFormatter != nil {
		m.cachedAnimationContent = make(map[string]string)
		// Update animation content to ensure proper state
		m.updateAnimationContent()
	}

	// Update config - the animation engine will handle reactivation and loop restart
	return m.animationEngine.UpdateConfig(m.animationConfig)
}

// DisableAnimations disables the animation system and stops the loop.
func (m *TeaTable) DisableAnimations() {
	m.animationConfig.Enabled = false
	m.animationEngine.UpdateConfig(m.animationConfig)

	// Clear cached animation content to ensure fallback to default formatting
	m.cachedAnimationContent = make(map[string]string)
}

// SetAnimateOnlyCursorRow sets whether to animate only the cursor row or all visible rows.
// By default, only the cursor row is animated for better performance.
func (m *TeaTable) SetAnimateOnlyCursorRow(cursorOnly bool) {
	if m.animateOnlyCursorRow == cursorOnly {
		return // No change
	}

	m.animateOnlyCursorRow = cursorOnly

	// Clear cache and update animations with new behavior
	if m.animatedFormatter != nil {
		m.cachedAnimationContent = make(map[string]string)
		m.updateAnimationContent()
	}
}

// GetAnimateOnlyCursorRow returns whether only the cursor row is animated.
func (m *TeaTable) GetAnimateOnlyCursorRow() bool {
	return m.animateOnlyCursorRow
}

// IsAnimationEnabled returns whether animations are currently enabled.
func (m *TeaTable) IsAnimationEnabled() bool {
	return m.animationConfig.Enabled
}

// IsAnimationLoopRunning returns whether the animation loop is currently running.
func (m *TeaTable) IsAnimationLoopRunning() bool {
	return m.animationEngine.IsRunning()
}

// GetAnimationConfig returns the current animation configuration.
func (m *TeaTable) GetAnimationConfig() AnimationConfig {
	return m.animationConfig
}

// SetTickInterval sets the animation tick interval for smoother or more efficient animations.
func (m *TeaTable) SetTickInterval(interval time.Duration) tea.Cmd {
	m.animationConfig.TickInterval = interval
	return m.animationEngine.UpdateConfig(m.animationConfig)
}

// GetTickInterval returns the current animation tick interval.
func (m *TeaTable) GetTickInterval() time.Duration {
	return m.animationConfig.TickInterval
}

// recalculateBorders recalculates the border strings using the current theme.
func (t *Table) recalculateBorders() {
	// Only create border strings if borders are enabled
	if !t.config.ShowBorders {
		return
	}

	// Build border strings with proper junction characters
	var topBuilder, middleBuilder, bottomBuilder strings.Builder

	for i, col := range t.config.Columns {
		// Top border
		if i == 0 {
			topBuilder.WriteString(t.theme.BorderChars.TopLeft)
		}
		for j := 0; j < col.Width; j++ {
			topBuilder.WriteString(t.theme.BorderChars.Horizontal)
		}
		if i == len(t.config.Columns)-1 {
			topBuilder.WriteString(t.theme.BorderChars.TopRight)
		} else {
			topBuilder.WriteString(t.theme.BorderChars.TopT)
		}

		// Middle border (separator)
		if i == 0 {
			middleBuilder.WriteString(t.theme.BorderChars.LeftT)
		}
		for j := 0; j < col.Width; j++ {
			middleBuilder.WriteString(t.theme.BorderChars.Horizontal)
		}
		if i == len(t.config.Columns)-1 {
			middleBuilder.WriteString(t.theme.BorderChars.RightT)
		} else {
			middleBuilder.WriteString(t.theme.BorderChars.Cross)
		}

		// Bottom border
		if i == 0 {
			bottomBuilder.WriteString(t.theme.BorderChars.BottomLeft)
		}
		for j := 0; j < col.Width; j++ {
			bottomBuilder.WriteString(t.theme.BorderChars.Horizontal)
		}
		if i == len(t.config.Columns)-1 {
			bottomBuilder.WriteString(t.theme.BorderChars.BottomRight)
		} else {
			bottomBuilder.WriteString(t.theme.BorderChars.BottomT)
		}
	}

	// Create proper styles for borders
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.theme.BorderColor))
	t.horizontalBorderTop = borderStyle.Render(topBuilder.String())
	t.horizontalBorderMiddle = borderStyle.Render(middleBuilder.String())
	t.horizontalBorderBottom = borderStyle.Render(bottomBuilder.String())
}

// SetDataProvider updates the data provider for the table.
// Note: This will reset the viewport position to maintain integrity.
func (t *TeaTable) SetDataProvider(provider DataProvider[TableRow]) {
	// Update the config to match the actual dataset size
	// This ensures we never try to show more rows than exist
	actualSize := provider.GetTotal()
	if actualSize < t.table.config.ViewportConfig.Height {
		// Create a temporary copy of the config
		newConfig := t.table.config
		newConfig.ViewportConfig.Height = max(1, actualSize)
		t.table.config = newConfig
	}

	// Store current position
	currentPos := t.table.list.State.CursorIndex

	// Update provider on the underlying list
	t.table.list.DataProvider = provider

	// Invalidate cache since we're changing the data source
	t.table.list.InvalidateTotalItemsCache()
	t.table.list.totalItems = provider.GetTotal()

	// Clear chunks and reload data
	t.table.list.chunks = make(map[int]*chunk[TableRow])

	// Try to restore position or adjust if needed
	if currentPos >= t.table.list.totalItems {
		currentPos = t.table.list.totalItems - 1
	}
	if currentPos < 0 {
		currentPos = 0
	}

	t.JumpToIndex(currentPos)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// SetColumns updates the column configuration for the table.
// This will recalculate border widths and refresh the display.
func (m *TeaTable) SetColumns(columns []TableColumn) {
	// Store current position
	currentPos := m.table.list.State.CursorIndex

	// Update columns
	m.table.config.Columns = columns

	// Recalculate total width
	totalWidth := 0
	for _, col := range columns {
		totalWidth += col.Width
	}

	// Add space for borders if needed
	if m.table.config.ShowBorders {
		// Add vertical borders (one per column + one for the end)
		totalWidth += len(columns) + 1
	}

	m.table.totalWidth = totalWidth

	// Recalculate borders
	m.table.recalculateBorders()

	// Restore position
	m.JumpToIndex(currentPos)
}

// SetHeaderVisibility sets whether the header is visible.
func (m *TeaTable) SetHeaderVisibility(visible bool) {
	m.table.config.ShowHeader = visible
}

// SetBorderVisibility sets whether borders are visible.
func (m *TeaTable) SetBorderVisibility(visible bool) {
	if m.table.config.ShowBorders == visible {
		return // No change
	}

	m.table.config.ShowBorders = visible

	// Recalculate borders
	if visible {
		m.table.recalculateBorders()
	} else {
		// Clear borders
		m.table.horizontalBorderTop = ""
		m.table.horizontalBorderMiddle = ""
		m.table.horizontalBorderBottom = ""
	}

	// Recalculate total width
	totalWidth := 0
	for _, col := range m.table.config.Columns {
		totalWidth += col.Width
	}

	// Add space for borders if needed
	if visible {
		// Add vertical borders (one per column + one for the end)
		totalWidth += len(m.table.config.Columns) + 1
	}

	m.table.totalWidth = totalWidth
}

// OnSelect sets a callback function that will be called when a row is selected.
func (m *TeaTable) OnSelect(callback func(row TableRow, index int)) {
	m.onSelectRow = callback
}

// OnHighlight sets a callback function that will be called when the highlighted row changes.
func (m *TeaTable) OnHighlight(callback func(row TableRow, index int)) {
	m.onHighlight = callback
}

// OnScroll sets a callback function that will be called when the viewport scrolls.
func (m *TeaTable) OnScroll(callback func(state ViewportState)) {
	m.onScroll = callback
}

// HandleKeypress programmatically simulates pressing a key.
func (m *TeaTable) HandleKeypress(keyStr string) {
	// Create a key message and update
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyStr)}
	m.Update(keyMsg)
}

// RefreshData forces a reload of data from the provider.
// Useful when the underlying data has changed.
func (t *TeaTable) RefreshData() {
	// Update the config to match the actual dataset size
	actualSize := t.table.list.DataProvider.GetTotal()
	if actualSize < t.table.config.ViewportConfig.Height {
		// Create a temporary copy of the config
		newConfig := t.table.config
		newConfig.ViewportConfig.Height = max(1, actualSize)
		t.table.config = newConfig
	}

	// Store current position
	currentPos := t.table.list.State.CursorIndex

	// Invalidate cache since we know data has changed externally
	t.table.list.InvalidateTotalItemsCache()

	// Update total items count
	t.table.list.totalItems = t.table.list.DataProvider.GetTotal()

	// Clear chunks and reload data
	t.table.list.chunks = make(map[int]*chunk[TableRow])

	// Restore position or adjust if needed
	if currentPos >= t.table.list.totalItems {
		currentPos = t.table.list.totalItems - 1
	}
	if currentPos < 0 {
		currentPos = 0
	}

	t.JumpToIndex(currentPos)
}

// Selection methods - delegate to the underlying DataProvider

// ToggleSelection toggles the selection state of the row at the given index.
func (m *TeaTable) ToggleSelection(index int) bool {
	// Get current selection state
	if data, ok := m.table.list.GetCurrentItem(); ok && m.table.list.State.CursorIndex == index {
		newSelected := !data.Selected
		if m.table.list.DataProvider.SetSelected(index, newSelected) {
			// Use efficient cache refresh instead of full data reload
			m.refreshCachedData()
			return true
		}
	} else {
		// If not the current item, we need to determine current state differently
		// For now, just set as selected
		if m.table.list.DataProvider.SetSelected(index, true) {
			// Use efficient cache refresh instead of full data reload
			m.refreshCachedData()
			return true
		}
	}
	return false
}

// ToggleCurrentSelection toggles the selection state of the currently highlighted row.
func (m *TeaTable) ToggleCurrentSelection() bool {
	currentIndex := m.table.list.State.CursorIndex
	if data, ok := m.table.list.GetCurrentItem(); ok {
		newSelected := !data.Selected
		if m.table.list.DataProvider.SetSelected(currentIndex, newSelected) {
			// Just invalidate cache - let normal render cycle fetch fresh data
			m.refreshCachedData()
			return true
		}
	}
	return false
}

// SelectAll selects all rows.
func (m *TeaTable) SelectAll() bool {
	if m.table.list.DataProvider.SelectAll() {
		// Just invalidate cache - let normal render cycle fetch fresh data
		m.refreshCachedData()
		return true
	}
	return false
}

// ClearSelection clears all selections.
func (m *TeaTable) ClearSelection() {
	m.table.list.DataProvider.ClearSelection()
	// Just invalidate cache - let normal render cycle fetch fresh data
	m.refreshCachedData()
}

// refreshCachedData invalidates cached chunks to force refresh of visible data
// This should ONLY affect visual representation, NOT trigger data provider calls
func (m *TeaTable) refreshCachedData() {
	// DON'T clear chunks - that would trigger unnecessary data fetching
	// Instead, we should only update the selection state in existing chunks
	m.updateSelectionInVisibleChunks()
}

// updateSelectionInVisibleChunks updates selection state in currently loaded chunks
// without triggering any data provider calls
func (m *TeaTable) updateSelectionInVisibleChunks() {
	selectedIndices := m.table.list.DataProvider.GetSelectedIndices()
	selectedSet := make(map[int]bool)
	for _, idx := range selectedIndices {
		selectedSet[idx] = true
	}

	// Update selection state in all loaded chunks
	for _, chunk := range m.table.list.chunks {
		for i := range chunk.Items {
			absoluteIndex := chunk.StartIndex + i
			chunk.Items[i].Selected = selectedSet[absoluteIndex]
		}
	}

	// Update visible items to reflect selection changes
	m.table.list.updateVisibleItems()
}

// GetSelectedIndices returns the indices of all selected rows.
func (m *TeaTable) GetSelectedIndices() []int {
	return m.table.list.DataProvider.GetSelectedIndices()
}

// GetSelectionCount returns the number of selected rows.
func (m *TeaTable) GetSelectionCount() int {
	return len(m.table.list.DataProvider.GetSelectedIndices())
}

// EnableRealTimeUpdates enables periodic data refreshing for dynamic data sources
func (m *TeaTable) EnableRealTimeUpdates(interval time.Duration) {
	m.realTimeUpdates = true
	m.realTimeInterval = interval
	m.lastRealTimeUpdate = time.Now()
}

// DisableRealTimeUpdates disables periodic data refreshing
func (m *TeaTable) DisableRealTimeUpdates() {
	m.realTimeUpdates = false
}

// IsRealTimeUpdatesEnabled returns whether real-time updates are enabled
func (m *TeaTable) IsRealTimeUpdatesEnabled() bool {
	return m.realTimeUpdates
}

// ForceDataRefresh forces a complete data reload - use sparingly!
// This should only be called when you know the data structure has changed
func (m *TeaTable) ForceDataRefresh() {
	// Invalidate cache since we know data has changed externally
	m.table.list.InvalidateTotalItemsCache()
	m.table.list.refreshData()
}

// updateAnimationContent updates the animation content cache
func (m *TeaTable) updateAnimationContent() {
	visibleItems := m.table.list.GetVisibleItems()
	state := m.table.GetState()

	// Track which animations should be active
	activeAnimationKeys := make(map[string]bool)

	// Calculate delta time
	now := time.Now()
	deltaTime := time.Duration(0)
	if !m.lastAnimationTime.IsZero() {
		deltaTime = now.Sub(m.lastAnimationTime)
	}
	m.lastAnimationTime = now

	// Process each visible item
	for i, dataItem := range visibleItems {
		absoluteIndex := state.ViewportStartIndex + i
		animationKey := fmt.Sprintf("row-%d", absoluteIndex)

		// Determine if this row should be animated
		isCursor := i == state.CursorViewportIndex
		shouldAnimate := !m.animateOnlyCursorRow || isCursor

		if !shouldAnimate {
			// This row should NOT be animated - remove from cache and animation engine
			m.animationEngine.SetVisible(animationKey, false)
			delete(m.cachedAnimationContent, animationKey)
			// Row will fall back to regular table rendering
			continue
		}

		activeAnimationKeys[animationKey] = true

		// Create render context with table-specific information and delta time
		ctx := DefaultRenderContext()
		ctx.CurrentTime = now
		ctx.DeltaTime = deltaTime
		ctx.MaxWidth = m.table.totalWidth
		ctx.Theme = &m.table.theme

		// Get animation state
		animState := m.animationEngine.GetAnimationState(animationKey)

		// Determine threshold states
		isTopThreshold := i == m.table.list.Config.TopThresholdIndex
		isBottomThreshold := i == m.table.list.Config.BottomThresholdIndex

		// Call animated formatter
		result := m.animatedFormatter(dataItem, absoluteIndex, ctx, animState, isCursor, isTopThreshold, isBottomThreshold)

		// CRITICAL FIX: Actually update the cached content!
		m.cachedAnimationContent[animationKey] = result.Content

		// CRITICAL FIX: Only register animations ONCE when they first become visible
		// Do NOT re-register on every view render (this causes acceleration)
		if len(result.RefreshTriggers) > 0 && !m.animationEngine.IsVisible(animationKey) {
			// Register animation only if it doesn't exist yet
			if cmd := m.animationEngine.RegisterAnimation(animationKey, result.RefreshTriggers, result.AnimationState); cmd != nil {
				// Animation loop started - this only happens once
				_ = cmd
			}
		}

		// Update animation state ONLY if it actually changed
		// Don't update state on every render to prevent animation reset
		if len(result.AnimationState) > 0 {
			currentState := m.animationEngine.GetAnimationState(animationKey)
			hasChanges := false

			// Check if state actually changed
			for k, newValue := range result.AnimationState {
				if currentValue, exists := currentState[k]; !exists || !deepEqual(currentValue, newValue) {
					hasChanges = true
					break
				}
			}

			// Only update if there are actual changes
			if hasChanges {
				m.animationEngine.UpdateAnimationState(animationKey, result.AnimationState)
			}
		}

		// Make sure the animation is visible (this is safe to call repeatedly)
		m.animationEngine.SetVisible(animationKey, true)
	}

	// Clean up animations for items that are no longer visible or should not be animated
	activeAnimations := m.animationEngine.GetActiveAnimations()
	for _, animKey := range activeAnimations {
		if !activeAnimationKeys[animKey] {
			m.animationEngine.SetVisible(animKey, false)
			// Remove cached content for invisible items
			delete(m.cachedAnimationContent, animKey)
		}
	}
}

// updateCursorInCache updates the cache for the cursor position
func (m *TeaTable) updateCursorInCache(oldIndex, newIndex int) {
	visibleItems := m.table.list.GetVisibleItems()
	state := m.table.GetState()

	// Find which viewport positions the old and new cursor positions map to
	oldViewportIndex := -1
	newViewportIndex := -1

	for i := range visibleItems {
		absoluteIndex := state.ViewportStartIndex + i
		if absoluteIndex == oldIndex {
			oldViewportIndex = i
		}
		if absoluteIndex == newIndex {
			newViewportIndex = i
		}
	}

	// Update cache for old cursor position (remove cursor)
	if oldViewportIndex >= 0 && oldViewportIndex < len(visibleItems) {
		m.updateSingleRowCache(oldViewportIndex, false) // Not cursor anymore
	}

	// Update cache for new cursor position (add cursor)
	if newViewportIndex >= 0 && newViewportIndex < len(visibleItems) {
		m.updateSingleRowCache(newViewportIndex, true) // Now cursor
	}
}

// updateSingleRowCache updates the cache for a single row with cursor state
func (m *TeaTable) updateSingleRowCache(viewportIndex int, isCursor bool) {
	visibleItems := m.table.list.GetVisibleItems()
	state := m.table.GetState()

	if viewportIndex >= len(visibleItems) {
		return
	}

	absoluteIndex := state.ViewportStartIndex + viewportIndex
	animationKey := fmt.Sprintf("row-%d", absoluteIndex)
	dataItem := visibleItems[viewportIndex]

	// Determine if this row should be animated
	shouldAnimate := !m.animateOnlyCursorRow || isCursor

	if !shouldAnimate {
		// This row should NOT be animated - remove from cache and animation engine
		m.animationEngine.SetVisible(animationKey, false)
		delete(m.cachedAnimationContent, animationKey)
		// Row will fall back to regular table rendering
		return
	}

	// Row should be animated - proceed with animation logic

	// Create render context
	ctx := DefaultRenderContext()
	ctx.CurrentTime = time.Now()
	ctx.DeltaTime = 0 // No delta time for cursor updates
	ctx.MaxWidth = m.table.totalWidth
	ctx.Theme = &m.table.theme

	// Get existing animation state (preserve it!)
	existingAnimState := m.animationEngine.GetAnimationState(animationKey)

	// Determine threshold states
	isTopThreshold := viewportIndex == m.table.list.Config.TopThresholdIndex
	isBottomThreshold := viewportIndex == m.table.list.Config.BottomThresholdIndex

	// Call animated formatter with updated cursor state
	result := m.animatedFormatter(dataItem, absoluteIndex, ctx, existingAnimState, isCursor, isTopThreshold, isBottomThreshold)

	// Update only the content in cache
	m.cachedAnimationContent[animationKey] = result.Content

	// IMPROVED: Allow animation updates when cursor state changes, but prevent acceleration
	// Only update if there are actual changes to refresh triggers or animation state
	if len(result.RefreshTriggers) > 0 {
		// Check if this animation is already active and visible
		animationExists := m.animationEngine.IsVisible(animationKey)

		// Only re-register if no animation exists
		// This allows cursor state changes to trigger animations when needed
		if !animationExists {
			if cmd := m.animationEngine.RegisterAnimation(animationKey, result.RefreshTriggers, result.AnimationState); cmd != nil {
				_ = cmd
			}
		}
	}

	// Update animation state only if it has actually changed
	if len(result.AnimationState) > 0 {
		hasChanges := false

		// Check if state actually changed
		for k, newValue := range result.AnimationState {
			if currentValue, exists := existingAnimState[k]; !exists || !deepEqual(currentValue, newValue) {
				hasChanges = true
				break
			}
		}

		// Only update if there are actual changes
		if hasChanges {
			m.animationEngine.UpdateAnimationState(animationKey, result.AnimationState)
		}
	}

	// Make sure the animation is visible
	m.animationEngine.SetVisible(animationKey, true)
}

// ensureAnimationCachePopulated ensures all visible items have cached animation content
func (m *TeaTable) ensureAnimationCachePopulated() {
	visibleItems := m.table.list.GetVisibleItems()
	state := m.table.GetState()

	for i := range visibleItems {
		absoluteIndex := state.ViewportStartIndex + i
		animationKey := fmt.Sprintf("row-%d", absoluteIndex)

		// Check if this row should be animated
		isCursor := i == state.CursorViewportIndex
		shouldAnimate := !m.animateOnlyCursorRow || isCursor

		// Only populate cache for rows that should be animated
		if !shouldAnimate {
			continue
		}

		// If we don't have cached content for this item, it means it's newly visible
		// We need to populate the cache without triggering animation acceleration
		if _, exists := m.cachedAnimationContent[animationKey]; !exists {
			// This is a new visible item - we need to populate its cache
			// But we should only call the formatter once for initial setup
			m.populateInitialAnimationContent(i, absoluteIndex, animationKey)
		}
	}
}

// populateInitialAnimationContent populates cache for a newly visible item
func (m *TeaTable) populateInitialAnimationContent(viewportIndex, absoluteIndex int, animationKey string) {
	visibleItems := m.table.list.GetVisibleItems()
	if viewportIndex >= len(visibleItems) {
		return
	}

	dataItem := visibleItems[viewportIndex]
	state := m.table.GetState()

	// Create render context
	ctx := DefaultRenderContext()
	ctx.CurrentTime = time.Now()
	ctx.DeltaTime = 0 // No delta time for initial render
	ctx.MaxWidth = m.table.totalWidth
	ctx.Theme = &m.table.theme

	// Get animation state (will be empty for new animations)
	animState := m.animationEngine.GetAnimationState(animationKey)

	// Determine cursor state
	isCursor := viewportIndex == state.CursorViewportIndex
	isTopThreshold := viewportIndex == m.table.list.Config.TopThresholdIndex
	isBottomThreshold := viewportIndex == m.table.list.Config.BottomThresholdIndex

	// Call animated formatter ONCE for initial setup
	result := m.animatedFormatter(dataItem, absoluteIndex, ctx, animState, isCursor, isTopThreshold, isBottomThreshold)

	// Cache the content
	m.cachedAnimationContent[animationKey] = result.Content

	// Register animation if needed (this should only happen once)
	if len(result.RefreshTriggers) > 0 && !m.animationEngine.IsVisible(animationKey) {
		if cmd := m.animationEngine.RegisterAnimation(animationKey, result.RefreshTriggers, result.AnimationState); cmd != nil {
			_ = cmd
		}
	}

	// Set initial animation state
	if len(result.AnimationState) > 0 {
		m.animationEngine.UpdateAnimationState(animationKey, result.AnimationState)
	}

	// Make animation visible
	m.animationEngine.SetVisible(animationKey, true)
}

// GetCachedTotal returns the cached total items count without triggering any data provider calls.
// This is useful for UI elements that need to display the total count efficiently.
// Returns the last known total from the cache, which may be stale if InvalidateTotalItemsCache() was called.
func (m *TeaTable) GetCachedTotal() int {
	return m.table.list.GetCachedTotal()
}

// CreateAnimatedTableRow is a helper function that applies animations to individual cells
// while maintaining proper table formatting, borders, and styling.
// This prevents alignment issues that can occur when manually building table content.
func CreateAnimatedTableRow(
	data Data[TableRow],
	index int,
	isCursor bool,
	isTopThreshold bool,
	isBottomThreshold bool,
	config TableConfig,
	theme Theme,
	cellAnimations map[int]string, // map of column index to animated content
) string {
	var sb strings.Builder

	row := data.Item
	isSelected := data.Selected

	// Ensure we don't iterate beyond the row's cell count
	cellCount := len(row.Cells)
	columnCount := len(config.Columns)

	// Format cells
	for i := 0; i < columnCount; i++ {
		// Add starting border if needed
		if config.ShowBorders && i == 0 {
			borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.BorderColor))
			if isCursor || isSelected {
				// Highlight border for cursor or selected row
				borderStyle = borderStyle.Bold(true)
			}
			sb.WriteString(borderStyle.Render(theme.BorderChars.Vertical))
		}

		// Determine the style to use (don't set width - enforceCellConstraints handles that)
		var style lipgloss.Style
		if isCursor && isSelected {
			// Both cursor and selected: Use a special combined style (bold selected style)
			style = theme.SelectedRowStyle.Copy().Bold(true)
		} else if isCursor {
			// Just cursor: Use selected style for cursor row
			style = theme.SelectedRowStyle.Copy()
		} else if isSelected {
			// Just selected: Use a modified style to show selection
			style = theme.RowEvenStyle.Copy().
				Background(lipgloss.Color("22")). // Dark green background for selected
				Foreground(lipgloss.Color("15"))  // White text
		} else if isTopThreshold {
			// Apply threshold styling if needed
			style = theme.RowStyle.Copy()
		} else if isBottomThreshold {
			// Apply threshold styling if needed
			style = theme.RowStyle.Copy()
		} else if index%2 == 0 {
			// Even rows
			style = theme.RowEvenStyle.Copy()
		} else {
			// Odd rows
			style = theme.RowOddStyle.Copy()
		}

		// Don't set alignment on lipgloss style - enforceCellConstraints handles that too

		// Get the cell value, using animated content if available
		var value string
		if animatedContent, hasAnimation := cellAnimations[i]; hasAnimation {
			// Use animated content for this cell
			value = animatedContent
		} else if i < cellCount {
			// Use regular cell content
			value = row.Cells[i]
		}

		// Add selection indicator to the first column
		if i == 0 && isSelected {
			if isCursor {
				value = "✓>" + value // Both selected and cursor
			} else {
				value = "✓ " + value // Just selected
			}
		}

		// Apply cell constraints to ensure content fits within column boundaries
		constraint := CellConstraint{
			Width:     config.Columns[i].Width,
			Height:    1, // TODO: Support multi-line cells
			Alignment: config.Columns[i].Alignment,
		}
		value = enforceCellConstraints(value, constraint)

		// Render the cell
		sb.WriteString(style.Render(value))

		// Add border if needed
		if config.ShowBorders {
			borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.BorderColor))
			if isCursor || isSelected {
				// Highlight border for cursor or selected row
				borderStyle = borderStyle.Bold(true)
			}
			sb.WriteString(borderStyle.Render(theme.BorderChars.Vertical))
		}
	}

	return sb.String()
}
