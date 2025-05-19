package vtable

import (
	"fmt"
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

	// Calculate the total width of the table
	totalWidth := 0
	for _, col := range config.Columns {
		totalWidth += col.Width
	}

	// Add space for borders if needed
	if config.ShowBorders {
		// Add vertical borders (one per column + one for the end)
		totalWidth += len(config.Columns) + 1
	}

	// Create horizontal border strings with the theme's border characters
	var horizontalBorderTop string
	var horizontalBorderMiddle string
	var horizontalBorderBottom string

	if config.ShowBorders {
		// Build border strings with proper junction characters
		var topBuilder strings.Builder
		var middleBuilder strings.Builder
		var bottomBuilder strings.Builder

		// Start with corner characters
		topBuilder.WriteString(theme.BorderChars.TopLeft)
		middleBuilder.WriteString(theme.BorderChars.LeftT)
		bottomBuilder.WriteString(theme.BorderChars.BottomLeft)

		for i, col := range config.Columns {
			// Add horizontal line for each column width
			topBuilder.WriteString(strings.Repeat(theme.BorderChars.Horizontal, col.Width))
			middleBuilder.WriteString(strings.Repeat(theme.BorderChars.Horizontal, col.Width))
			bottomBuilder.WriteString(strings.Repeat(theme.BorderChars.Horizontal, col.Width))

			// Add junction if not the last column
			if i < len(config.Columns)-1 {
				topBuilder.WriteString(theme.BorderChars.TopT)
				middleBuilder.WriteString(theme.BorderChars.Cross)
				bottomBuilder.WriteString(theme.BorderChars.BottomT)
			}
		}

		// Add right corner characters
		topBuilder.WriteString(theme.BorderChars.TopRight)
		middleBuilder.WriteString(theme.BorderChars.RightT)
		bottomBuilder.WriteString(theme.BorderChars.BottomRight)

		horizontalBorderTop = theme.BorderStyle.Render(topBuilder.String())
		horizontalBorderMiddle = theme.BorderStyle.Render(middleBuilder.String())
		horizontalBorderBottom = theme.BorderStyle.Render(bottomBuilder.String())
	}

	// Create a formatter for the table rows
	formatter := func(row TableRow, index int, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		return formatTableRow(row, index, isCursor, isTopThreshold, isBottomThreshold, config, theme)
	}

	// Convert theme to styleConfig for backward compatibility with List
	styleConfig := ThemeToStyleConfig(theme)

	// Create the underlying list
	list, err := NewList(config.ViewportConfig, provider, styleConfig, formatter)
	if err != nil {
		return nil, err
	}

	return &Table{
		config:                 config,
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
	row TableRow,
	index int,
	isCursor bool,
	isTopThreshold bool,
	isBottomThreshold bool,
	config TableConfig,
	theme Theme,
) string {
	var sb strings.Builder

	// Format cells
	for i, value := range row.Cells {
		// Add starting border if needed
		if config.ShowBorders && i == 0 {
			borderStyle := theme.BorderStyle
			if isCursor {
				// Highlight border for cursor row
				borderStyle = borderStyle.Copy().Bold(true)
			}
			sb.WriteString(borderStyle.Render(theme.BorderChars.Vertical))
		}

		// Determine the style to use
		var style lipgloss.Style
		if isCursor {
			// Use selected style for cursor row
			style = theme.SelectedRowStyle.Copy().Width(config.Columns[i].Width)
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

		// Truncate if needed
		if len(value) > config.Columns[i].Width {
			value = value[:config.Columns[i].Width-3] + "..."
		}

		// Render the cell
		sb.WriteString(style.Render(value))

		// Add border if needed
		if config.ShowBorders {
			borderStyle := theme.BorderStyle
			if isCursor {
				// Highlight border for cursor row
				borderStyle = borderStyle.Copy().Bold(true)
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

	// Format header cells
	for i, col := range t.config.Columns {
		// Start with border if needed
		if t.config.ShowBorders && i == 0 {
			sb.WriteString(t.theme.HeaderBorderStyle.Render(t.theme.BorderChars.Vertical))
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

		sb.WriteString(style.Render(col.Title))

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

	// Render the list
	sb.WriteString(t.list.Render())

	// Add bottom border if needed
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
	return t.list.GetCurrentItem()
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
	onSelectRow func(row TableRow, index int)
	onScroll    func(state ViewportState)
	onHighlight func(row TableRow, index int)
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
	var topBuilder strings.Builder
	var middleBuilder strings.Builder
	var bottomBuilder strings.Builder

	// Start with corner characters
	topBuilder.WriteString(t.theme.BorderChars.TopLeft)
	middleBuilder.WriteString(t.theme.BorderChars.LeftT)
	bottomBuilder.WriteString(t.theme.BorderChars.BottomLeft)

	for i, col := range t.config.Columns {
		// Add horizontal line for each column width
		topBuilder.WriteString(strings.Repeat(t.theme.BorderChars.Horizontal, col.Width))
		middleBuilder.WriteString(strings.Repeat(t.theme.BorderChars.Horizontal, col.Width))
		bottomBuilder.WriteString(strings.Repeat(t.theme.BorderChars.Horizontal, col.Width))

		// Add junction if not the last column
		if i < len(t.config.Columns)-1 {
			topBuilder.WriteString(t.theme.BorderChars.TopT)
			middleBuilder.WriteString(t.theme.BorderChars.Cross)
			bottomBuilder.WriteString(t.theme.BorderChars.BottomT)
		}
	}

	// Add right corner characters
	topBuilder.WriteString(t.theme.BorderChars.TopRight)
	middleBuilder.WriteString(t.theme.BorderChars.RightT)
	bottomBuilder.WriteString(t.theme.BorderChars.BottomRight)

	t.horizontalBorderTop = t.theme.BorderStyle.Render(topBuilder.String())
	t.horizontalBorderMiddle = t.theme.BorderStyle.Render(middleBuilder.String())
	t.horizontalBorderBottom = t.theme.BorderStyle.Render(bottomBuilder.String())
}

// SetDataProvider updates the data provider for the table.
// Note: This will reset the viewport position to maintain integrity.
func (m *TeaTable) SetDataProvider(provider DataProvider[TableRow]) {
	// Store current position
	currentPos := m.table.list.State.CursorIndex

	// Update provider on the underlying list
	m.table.list.DataProvider = provider
	m.table.list.totalItems = provider.GetTotal()

	// Clear chunks and reload data
	m.table.list.chunks = make(map[int]*chunk[TableRow])

	// Try to restore position or adjust if needed
	if currentPos >= m.table.list.totalItems {
		currentPos = m.table.list.totalItems - 1
	}
	if currentPos < 0 {
		currentPos = 0
	}

	m.JumpToIndex(currentPos)
}

// RefreshData forces a reload of data from the provider.
// Useful when the underlying data has changed.
func (m *TeaTable) RefreshData() {
	// Store current position
	currentPos := m.table.list.State.CursorIndex

	// Update total items count
	m.table.list.totalItems = m.table.list.DataProvider.GetTotal()

	// Clear chunks and reload data
	m.table.list.chunks = make(map[int]*chunk[TableRow])

	// Restore position or adjust if needed
	if currentPos >= m.table.list.totalItems {
		currentPos = m.table.list.totalItems - 1
	}
	if currentPos < 0 {
		currentPos = 0
	}

	m.JumpToIndex(currentPos)
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
