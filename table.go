package vtable

import (
	"fmt"
	"strconv"
	"strings"

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
	// Validate column configuration
	if len(config.Columns) == 0 {
		return nil, fmt.Errorf("table must have at least one column")
	}

	// Get actual data size
	dataSize := provider.GetTotal()

	// Create a copy of the config to avoid modifying the original
	adjustedConfig := config

	// Adjust viewport height if it's larger than the dataset
	// This prevents empty rows from showing
	if dataSize < config.ViewportConfig.Height {
		adjustedConfig.ViewportConfig.Height = max(1, dataSize)
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

		// Handle the last column properly for horizontal borders
		topBuilder.WriteString(theme.BorderChars.TopRight)
		middleBuilder.WriteString(theme.BorderChars.RightT)
		bottomBuilder.WriteString(theme.BorderChars.BottomRight)

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

// formatTableRow formats a single table row.
func formatTableRow(
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

		// Determine the style to use
		var style lipgloss.Style
		if isCursor && isSelected {
			// Both cursor and selected: Use a special combined style (bold selected style)
			style = theme.SelectedRowStyle.Copy().Width(config.Columns[i].Width).Bold(true)
		} else if isCursor {
			// Just cursor: Use selected style for cursor row
			style = theme.SelectedRowStyle.Copy().Width(config.Columns[i].Width)
		} else if isSelected {
			// Just selected: Use a modified style to show selection
			style = theme.RowEvenStyle.Copy().Width(config.Columns[i].Width).
				Background(lipgloss.Color("22")). // Dark green background for selected
				Foreground(lipgloss.Color("15"))  // White text
		} else if isTopThreshold {
			// Apply threshold styling if needed
			style = theme.RowStyle.Copy().Width(config.Columns[i].Width)
		} else if isBottomThreshold {
			// Apply threshold styling if needed
			style = theme.RowStyle.Copy().Width(config.Columns[i].Width)
		} else if index%2 == 0 {
			// Even rows
			style = theme.RowEvenStyle.Copy().Width(config.Columns[i].Width)
		} else {
			// Odd rows
			style = theme.RowOddStyle.Copy().Width(config.Columns[i].Width)
		}

		// Set alignment
		switch config.Columns[i].Alignment {
		case AlignCenter:
			style = style.Align(lipgloss.Center)
		case AlignRight:
			style = style.Align(lipgloss.Right)
		default:
			style = style.Align(lipgloss.Left)
		}

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

		// Truncate if needed (accounting for selection indicator)
		maxWidth := config.Columns[i].Width
		if len(value) > maxWidth {
			value = value[:maxWidth-3] + "..."
		}

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

		// Format the header cell
		style := t.theme.HeaderStyle.Copy().Width(col.Width)

		// Set alignment
		switch col.Alignment {
		case AlignCenter:
			style = style.Align(lipgloss.Center)
		case AlignRight:
			style = style.Align(lipgloss.Right)
		default:
			style = style.Align(lipgloss.Left)
		}

		// Add sort indicator to title if this column is sorted
		title := col.Title
		if sortDirection != "" {
			if sortDirection == "asc" {
				title = title + " ↑" // Up arrow for ascending
			} else {
				title = title + " ↓" // Down arrow for descending
			}
		}

		sb.WriteString(style.Render(title))

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

	return &TeaTable{
		table:   table,
		keyMap:  PlatformKeyMap(), // Use platform-specific key bindings
		focused: true,
	}, nil
}

// Init initializes the Tea model.
func (m *TeaTable) Init() tea.Cmd {
	return nil
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
	return m.table.Render()
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

// SetSort sets the sort field and direction, clearing any existing sorts.
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

	// Handle the last column properly for horizontal borders
	topBuilder.WriteString(t.theme.BorderChars.TopRight)
	middleBuilder.WriteString(t.theme.BorderChars.RightT)
	bottomBuilder.WriteString(t.theme.BorderChars.BottomRight)

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
			// Refresh the data to show selection changes
			m.RefreshData()
			return true
		}
	} else {
		// If not the current item, we need to determine current state differently
		// For now, just set as selected
		if m.table.list.DataProvider.SetSelected(index, true) {
			// Refresh the data to show selection changes
			m.RefreshData()
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
func (m *TeaTable) refreshCachedData() {
	// Clear chunks to force reload with updated selection state
	m.table.list.chunks = make(map[int]*chunk[TableRow])
	// Force update of visible items to reload fresh chunks
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
