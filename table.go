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
	// Config is the configuration for the table.
	Config TableConfig

	// Theme defines the visual appearance of the table.
	Theme Theme

	// List is the underlying list component.
	List *List[TableRow]

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
		Config:                 config,
		Theme:                  theme,
		List:                   list,
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
	if t.Config.ShowBorders {
		sb.WriteString(t.horizontalBorderTop)
		sb.WriteString("\n")
	}

	// Format header cells
	for i, col := range t.Config.Columns {
		// Start with border if needed
		if t.Config.ShowBorders && i == 0 {
			sb.WriteString(t.Theme.HeaderBorderStyle.Render(t.Theme.BorderChars.Vertical))
		}

		// Format the header cell
		style := t.Theme.HeaderStyle.Copy().Width(col.Width)

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
		if t.Config.ShowBorders {
			sb.WriteString(t.Theme.HeaderBorderStyle.Render(t.Theme.BorderChars.Vertical))
		}
	}

	// Add middle border if needed
	if t.Config.ShowBorders {
		sb.WriteString("\n")
		sb.WriteString(t.horizontalBorderMiddle)
	}

	return sb.String()
}

// Render renders the table to a string.
func (t *Table) Render() string {
	var sb strings.Builder

	// Add header if needed
	if t.Config.ShowHeader {
		header := t.formatHeader()
		sb.WriteString(header)
		sb.WriteString("\n")
	}

	// Render the list
	sb.WriteString(t.List.Render())

	// Add bottom border if needed
	if t.Config.ShowBorders {
		sb.WriteString("\n")
		sb.WriteString(t.horizontalBorderBottom)
	}

	return sb.String()
}

// GetState returns the current viewport state.
func (t *Table) GetState() ViewportState {
	return t.List.GetState()
}

// GetCurrentRow returns the currently selected row.
func (t *Table) GetCurrentRow() (TableRow, bool) {
	return t.List.GetCurrentItem()
}

// RenderDebugInfo renders debug information about the table.
func (t *Table) RenderDebugInfo() string {
	return t.List.RenderDebugInfo()
}

// All navigation methods delegate to the underlying list

func (t *Table) MoveUp() {
	t.List.MoveUp()
}

func (t *Table) MoveDown() {
	t.List.MoveDown()
}

func (t *Table) PageUp() {
	t.List.PageUp()
}

func (t *Table) PageDown() {
	t.List.PageDown()
}

func (t *Table) JumpToIndex(index int) {
	t.List.JumpToIndex(index)
}

func (t *Table) JumpToStart() {
	t.List.JumpToStart()
}

func (t *Table) JumpToEnd() {
	t.List.JumpToEnd()
}

func (t *Table) JumpToItem(key string, value any) bool {
	return t.List.JumpToItem(key, value)
}

// TeaTable is a Bubble Tea model wrapping a Table.
type TeaTable struct {
	// The underlying table model
	Table *Table

	// Key mappings
	KeyMap NavigationKeyMap

	// Whether the component is focused
	Focused bool
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
		Table:   table,
		KeyMap:  PlatformKeyMap(), // Use platform-specific key bindings
		Focused: true,
	}, nil
}

// Init initializes the Tea model.
func (m *TeaTable) Init() tea.Cmd {
	return nil
}

// Update updates the Tea model based on messages.
func (m *TeaTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If not focused, don't handle messages
	if !m.Focused {
		return m, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Up):
			m.Table.MoveUp()
		case key.Matches(msg, m.KeyMap.Down):
			m.Table.MoveDown()
		case key.Matches(msg, m.KeyMap.PageUp):
			m.Table.PageUp()
		case key.Matches(msg, m.KeyMap.PageDown):
			m.Table.PageDown()
		case key.Matches(msg, m.KeyMap.Home):
			m.Table.JumpToStart()
		case key.Matches(msg, m.KeyMap.End):
			m.Table.JumpToEnd()
			// Note: Search and Select are handled by the parent application
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the Tea model.
func (m *TeaTable) View() string {
	return m.Table.Render()
}

// Focus sets the focus state of the component.
func (m *TeaTable) Focus() {
	m.Focused = true
}

// Blur removes focus from the component.
func (m *TeaTable) Blur() {
	m.Focused = false
}

// IsFocused returns whether the component is focused.
func (m *TeaTable) IsFocused() bool {
	return m.Focused
}

// GetState returns the current viewport state.
func (m *TeaTable) GetState() ViewportState {
	return m.Table.GetState()
}

// GetCurrentRow returns the currently selected row.
func (m *TeaTable) GetCurrentRow() (TableRow, bool) {
	return m.Table.GetCurrentRow()
}

// SetKeyMap sets the key mappings for the component.
func (m *TeaTable) SetKeyMap(keyMap NavigationKeyMap) {
	m.KeyMap = keyMap
}

// JumpToItem jumps to a row with the specified key-value pair.
// Returns true if the row was found and jumped to, false otherwise.
func (m *TeaTable) JumpToItem(key string, value any) bool {
	return m.Table.JumpToItem(key, value)
}

// JumpToIndex jumps to the specified index in the dataset.
func (m *TeaTable) JumpToIndex(index int) {
	m.Table.JumpToIndex(index)
}

// GetHelpView returns a string describing the key bindings.
func (m *TeaTable) GetHelpView() string {
	return GetKeyMapDescription(m.KeyMap)
}
