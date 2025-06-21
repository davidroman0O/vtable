package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/table"
)

// Simple employee data for demonstration
type EmployeeDataSource struct {
	employees []Employee
}

type Employee struct {
	ID         string
	Name       string
	Department string
	Status     string
	Salary     int
	Email      string
	Phone      string
}

func NewEmployeeDataSource() *EmployeeDataSource {
	employees := []Employee{
		{"EMP001", "Alice Johnson", "Engineering", "Active", 75000, "alice@company.com", "(555) 123-4567"},
		{"EMP002", "Bob Smith", "Marketing", "Active", 65000, "bob@company.com", "(555) 234-5678"},
		{"EMP003", "Carol Davis", "Sales", "Remote", 70000, "carol@company.com", "(555) 345-6789"},
		{"EMP004", "David Wilson", "HR", "On Leave", 60000, "david@company.com", "(555) 456-7890"},
		{"EMP005", "Eve Brown", "Finance", "Active", 80000, "eve@company.com", "(555) 567-8901"},
		{"EMP006", "Frank Miller", "Engineering", "Active", 78000, "frank@company.com", "(555) 678-9012"},
		{"EMP007", "Grace Lee", "Marketing", "Part-time", 45000, "grace@company.com", "(555) 789-0123"},
		{"EMP008", "Henry Chen", "Sales", "Active", 72000, "henry@company.com", "(555) 890-1234"},
	}

	return &EmployeeDataSource{employees: employees}
}

func (ds *EmployeeDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return core.DataTotalMsg{Total: len(ds.employees)}
	}
}

func (ds *EmployeeDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(10 * time.Millisecond)

		end := request.Start + request.Count
		if end > len(ds.employees) {
			end = len(ds.employees)
		}

		chunkItems := make([]core.Data[any], end-request.Start)
		for i := request.Start; i < end; i++ {
			emp := ds.employees[i]
			chunkItems[i-request.Start] = core.Data[any]{
				ID: emp.ID,
				Item: core.TableRow{
					ID:    emp.ID,
					Cells: []string{emp.ID, emp.Name, emp.Department, emp.Status, fmt.Sprintf("$%d", emp.Salary), emp.Email, emp.Phone},
				},
				Metadata: core.NewTypedMetadata(),
			}
		}

		return core.DataChunkLoadedMsg{
			StartIndex: request.Start,
			Items:      chunkItems,
			Request:    request,
		}
	}
}

// Required interface methods (simplified)
func (ds *EmployeeDataSource) RefreshTotal() tea.Cmd                            { return ds.GetTotal() }
func (ds *EmployeeDataSource) SetSelected(index int, selected bool) tea.Cmd     { return nil }
func (ds *EmployeeDataSource) SetSelectedByID(id string, selected bool) tea.Cmd { return nil }
func (ds *EmployeeDataSource) SelectAll() tea.Cmd                               { return nil }
func (ds *EmployeeDataSource) ClearSelection() tea.Cmd                          { return nil }
func (ds *EmployeeDataSource) SelectRange(startIndex, endIndex int) tea.Cmd     { return nil }
func (ds *EmployeeDataSource) GetItemID(item any) string                        { return "" }

type AppModel struct {
	table            *table.Table
	statusMessage    string
	availableColumns []core.TableColumn // All possible columns
	visibleColumns   []int              // Indices of visible columns
	columnWidths     []int              // Current widths for adjustment
}

func main() {
	dataSource := NewEmployeeDataSource()

	// Define ALL possible columns (more than we'll show initially)
	availableColumns := []core.TableColumn{
		{Title: "ID", Width: 8, Alignment: core.AlignCenter, Field: "id"},
		{Title: "Employee Name", Width: 20, Alignment: core.AlignLeft, Field: "name"},
		{Title: "Department", Width: 15, Alignment: core.AlignCenter, Field: "department"},
		{Title: "Status", Width: 12, Alignment: core.AlignCenter, Field: "status"},
		{Title: "Salary", Width: 10, Alignment: core.AlignRight, Field: "salary"},
		{Title: "Email", Width: 25, Alignment: core.AlignLeft, Field: "email"},   // Hidden initially
		{Title: "Phone", Width: 15, Alignment: core.AlignCenter, Field: "phone"}, // Hidden initially
	}

	// Start with only some columns visible
	visibleColumns := []int{0, 1, 2, 3, 4} // ID, Name, Department, Status, Salary

	// Build initial column set
	initialColumns := make([]core.TableColumn, len(visibleColumns))
	columnWidths := make([]int, len(visibleColumns))
	for i, colIndex := range visibleColumns {
		initialColumns[i] = availableColumns[colIndex]
		columnWidths[i] = availableColumns[colIndex].Width
	}

	theme := core.Theme{
		HeaderStyle:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")),
		CellStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		CursorStyle:        lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("57")),
		SelectedStyle:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("57")),
		FullRowCursorStyle: lipgloss.NewStyle().Background(lipgloss.Color("57")).Foreground(lipgloss.Color("15")),
		BorderChars: core.BorderChars{
			Horizontal: "─", Vertical: "│", TopLeft: "┌", TopRight: "┐",
			BottomLeft: "└", BottomRight: "┘", TopT: "┬", BottomT: "┴",
			LeftT: "├", RightT: "┤", Cross: "┼",
		},
		BorderColor: "8",
	}

	config := core.TableConfig{
		Columns:     initialColumns,
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: core.ViewportConfig{
			Height:    10,
			ChunkSize: 20,
		},
		Theme:         theme,
		SelectionMode: core.SelectionNone,
	}

	tbl := table.NewTable(config, dataSource)
	tbl.Focus()

	model := AppModel{
		table:            tbl,
		statusMessage:    "Column Management Demo - Use Ctrl+←→ to reorder, +/- to add/remove, W/w to adjust widths",
		availableColumns: availableColumns,
		visibleColumns:   visibleColumns,
		columnWidths:     columnWidths,
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.table.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		// === COLUMN MANAGEMENT CONTROLS ===
		case "ctrl+left":
			return m.moveColumnLeft()

		case "ctrl+right":
			return m.moveColumnRight()

		case "+", "=":
			return m.addColumn()

		case "-", "_":
			return m.removeColumn()

		case "W":
			return m.adjustColumnWidth(5)

		case "w":
			return m.adjustColumnWidth(-5)

		case "A":
			return m.cycleColumnAlignment()

		case "R":
			return m.resetColumns()

		// Basic navigation
		case "j", "down":
			return m, core.CursorDownCmd()

		case "k", "up":
			return m, core.CursorUpCmd()

		case ".":
			return m, core.NextColumnCmd()

		case ",":
			return m, core.PrevColumnCmd()

		// Pass other keys to table
		default:
			var cmd tea.Cmd
			_, cmd = m.table.Update(msg)
			return m, cmd
		}

	default:
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd
	}
}

// Move current column left in the display order
func (m AppModel) moveColumnLeft() (tea.Model, tea.Cmd) {
	_, _, currentColumn, _ := m.table.GetHorizontalScrollState()

	if currentColumn > 0 {
		// Swap positions in visible columns list
		m.visibleColumns[currentColumn], m.visibleColumns[currentColumn-1] =
			m.visibleColumns[currentColumn-1], m.visibleColumns[currentColumn]

		// Swap column widths
		m.columnWidths[currentColumn], m.columnWidths[currentColumn-1] =
			m.columnWidths[currentColumn-1], m.columnWidths[currentColumn]

		// Update table with new order
		newColumns := m.buildCurrentColumns()
		m.statusMessage = "Moved column left"
		return m, tea.Batch(
			core.ColumnSetCmd(newColumns),
			core.PrevColumnCmd(), // Move focus to follow the column
		)
	}

	m.statusMessage = "Cannot move column further left"
	return m, nil
}

// Move current column right in the display order
func (m AppModel) moveColumnRight() (tea.Model, tea.Cmd) {
	_, _, currentColumn, _ := m.table.GetHorizontalScrollState()

	if currentColumn < len(m.visibleColumns)-1 {
		// Swap positions in visible columns list
		m.visibleColumns[currentColumn], m.visibleColumns[currentColumn+1] =
			m.visibleColumns[currentColumn+1], m.visibleColumns[currentColumn]

		// Swap column widths
		m.columnWidths[currentColumn], m.columnWidths[currentColumn+1] =
			m.columnWidths[currentColumn+1], m.columnWidths[currentColumn]

		// Update table with new order
		newColumns := m.buildCurrentColumns()
		m.statusMessage = "Moved column right"
		return m, tea.Batch(
			core.ColumnSetCmd(newColumns),
			core.NextColumnCmd(), // Move focus to follow the column
		)
	}

	m.statusMessage = "Cannot move column further right"
	return m, nil
}

// Add the next available column
func (m AppModel) addColumn() (tea.Model, tea.Cmd) {
	// Find first column not currently visible
	visibleSet := make(map[int]bool)
	for _, colIndex := range m.visibleColumns {
		visibleSet[colIndex] = true
	}

	for i, col := range m.availableColumns {
		if !visibleSet[i] {
			// Add this column
			m.visibleColumns = append(m.visibleColumns, i)
			m.columnWidths = append(m.columnWidths, col.Width)

			newColumns := m.buildCurrentColumns()
			m.statusMessage = fmt.Sprintf("Added column: %s", col.Title)
			return m, core.ColumnSetCmd(newColumns)
		}
	}

	m.statusMessage = "All available columns are already visible"
	return m, nil
}

// Remove the current column
func (m AppModel) removeColumn() (tea.Model, tea.Cmd) {
	_, _, currentColumn, _ := m.table.GetHorizontalScrollState()

	if len(m.visibleColumns) <= 1 {
		m.statusMessage = "Cannot remove last column"
		return m, nil
	}

	removedColumn := m.availableColumns[m.visibleColumns[currentColumn]]

	// Remove from visible columns and widths
	m.visibleColumns = append(m.visibleColumns[:currentColumn], m.visibleColumns[currentColumn+1:]...)
	m.columnWidths = append(m.columnWidths[:currentColumn], m.columnWidths[currentColumn+1:]...)

	newColumns := m.buildCurrentColumns()
	m.statusMessage = fmt.Sprintf("Removed column: %s", removedColumn.Title)

	// Adjust focus if we removed the last column
	var cmd tea.Cmd
	if currentColumn >= len(m.visibleColumns) && len(m.visibleColumns) > 0 {
		cmd = tea.Batch(
			core.ColumnSetCmd(newColumns),
			core.PrevColumnCmd(),
		)
	} else {
		cmd = core.ColumnSetCmd(newColumns)
	}

	return m, cmd
}

// Adjust width of current column
func (m AppModel) adjustColumnWidth(delta int) (tea.Model, tea.Cmd) {
	_, _, currentColumn, _ := m.table.GetHorizontalScrollState()

	if currentColumn < len(m.columnWidths) {
		newWidth := m.columnWidths[currentColumn] + delta
		if newWidth < 5 {
			newWidth = 5 // Minimum width
		}
		if newWidth > 50 {
			newWidth = 50 // Maximum width
		}

		m.columnWidths[currentColumn] = newWidth
		newColumns := m.buildCurrentColumns()

		action := "increased"
		if delta < 0 {
			action = "decreased"
		}

		m.statusMessage = fmt.Sprintf("Column width %s to %d", action, newWidth)
		return m, core.ColumnSetCmd(newColumns)
	}

	return m, nil
}

// Cycle column alignment for current column
func (m AppModel) cycleColumnAlignment() (tea.Model, tea.Cmd) {
	_, _, currentColumn, _ := m.table.GetHorizontalScrollState()

	if currentColumn < len(m.visibleColumns) {
		colIndex := m.visibleColumns[currentColumn]
		col := m.availableColumns[colIndex]

		// Cycle through alignments
		switch col.Alignment {
		case core.AlignLeft:
			col.Alignment = core.AlignCenter
		case core.AlignCenter:
			col.Alignment = core.AlignRight
		case core.AlignRight:
			col.Alignment = core.AlignLeft
		}

		m.availableColumns[colIndex] = col
		newColumns := m.buildCurrentColumns()

		alignmentNames := map[int]string{
			core.AlignLeft:   "left",
			core.AlignCenter: "center",
			core.AlignRight:  "right",
		}

		m.statusMessage = fmt.Sprintf("Column alignment: %s", alignmentNames[col.Alignment])
		return m, core.ColumnSetCmd(newColumns)
	}

	return m, nil
}

// Reset to default column configuration
func (m AppModel) resetColumns() (tea.Model, tea.Cmd) {
	// Reset to initial configuration
	m.visibleColumns = []int{0, 1, 2, 3, 4} // ID, Name, Department, Status, Salary
	m.columnWidths = []int{8, 20, 15, 12, 10}

	// Reset alignments to defaults
	m.availableColumns[0].Alignment = core.AlignCenter
	m.availableColumns[1].Alignment = core.AlignLeft
	m.availableColumns[2].Alignment = core.AlignCenter
	m.availableColumns[3].Alignment = core.AlignCenter
	m.availableColumns[4].Alignment = core.AlignRight

	newColumns := m.buildCurrentColumns()
	m.statusMessage = "Reset to default column configuration"
	return m, core.ColumnSetCmd(newColumns)
}

// Build current column configuration from visible columns
func (m AppModel) buildCurrentColumns() []core.TableColumn {
	columns := make([]core.TableColumn, len(m.visibleColumns))
	for i, colIndex := range m.visibleColumns {
		col := m.availableColumns[colIndex]
		col.Width = m.columnWidths[i] // Use current width
		columns[i] = col
	}
	return columns
}

func (m AppModel) View() string {
	// Get current column info for display
	_, _, currentColumn, _ := m.table.GetHorizontalScrollState()

	currentColumnName := "N/A"
	currentColumnWidth := 0
	if currentColumn < len(m.visibleColumns) {
		colIndex := m.visibleColumns[currentColumn]
		currentColumnName = m.availableColumns[colIndex].Title
		currentColumnWidth = m.columnWidths[currentColumn]
	}

	status := fmt.Sprintf("Columns: %d/%d | Current: %s (%d chars)",
		len(m.visibleColumns),
		len(m.availableColumns),
		currentColumnName,
		currentColumnWidth,
	)

	controls := "Ctrl+←→=reorder | +-=add/remove | Ww=width | A=alignment | R=reset | .,=navigate | ↑↓jk=rows | q=quit"

	return status + "\n" + controls + "\n" + m.statusMessage + "\n\n" + m.table.View()
}
