package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/table"
)

// EmployeeDataSource provides employee data for the table
type EmployeeDataSource struct {
	employees     []Employee
	selectedItems map[string]bool
}

type Employee struct {
	ID          string
	Name        string
	Department  string
	Position    string
	Status      string
	Salary      int
	Email       string
	Phone       string
	Description string
}

func NewEmployeeDataSource() *EmployeeDataSource {
	employees := make([]Employee, 1000)
	departments := []string{"Engineering", "Marketing", "Sales", "HR", "Finance", "Operations"}
	positions := []string{"Manager", "Senior Developer", "Analyst", "Coordinator", "Specialist", "Director"}
	statuses := []string{"Active", "On Leave", "Remote", "Part-time"}

	for i := 0; i < 1000; i++ {
		employees[i] = Employee{
			ID:          fmt.Sprintf("EMP%04d", i+1),
			Name:        fmt.Sprintf("Employee %d", i+1),
			Department:  departments[i%len(departments)],
			Position:    fmt.Sprintf("%s %s", positions[i%len(positions)], departments[i%len(departments)]),
			Status:      statuses[i%len(statuses)],
			Salary:      50000 + (i%50)*1000,
			Email:       fmt.Sprintf("employee%d@company.com", i+1),
			Phone:       fmt.Sprintf("(555) %03d-%04d", (i%900)+100, i%10000),
			Description: fmt.Sprintf("Experienced %s professional specializing in various aspects of %s operations with %d years of experience in the field.", departments[i%len(departments)], departments[i%len(departments)], (i%15)+1),
		}
	}

	return &EmployeeDataSource{
		employees:     employees,
		selectedItems: make(map[string]bool),
	}
}

func (ds *EmployeeDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return core.DataTotalMsg{Total: len(ds.employees)}
	}
}

func (ds *EmployeeDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(50 * time.Millisecond)

		end := request.Start + request.Count
		if end > len(ds.employees) {
			end = len(ds.employees)
		}
		if request.Start >= end {
			return core.DataChunkLoadedMsg{Items: []core.Data[any]{}}
		}

		chunkItems := make([]core.Data[any], end-request.Start)
		for i := request.Start; i < end; i++ {
			emp := ds.employees[i]
			chunkItems[i-request.Start] = core.Data[any]{
				ID: emp.ID,
				Item: core.TableRow{
					ID: emp.ID,
					Cells: []string{
						emp.ID,
						emp.Name,
						emp.Department,
						emp.Status,
						fmt.Sprintf("$%s", formatNumber(emp.Salary)),
						emp.Description,
					},
				},
				Selected: ds.selectedItems[emp.ID],
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

func (ds *EmployeeDataSource) RefreshTotal() tea.Cmd { return ds.GetTotal() }
func (ds *EmployeeDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		if index >= 0 && index < len(ds.employees) {
			id := ds.employees[index].ID
			if selected {
				ds.selectedItems[id] = true
			} else {
				delete(ds.selectedItems, id)
			}
			return core.SelectionResponseMsg{
				Success:  true,
				Index:    index,
				ID:       id,
				Selected: selected,
			}
		}
		return core.SelectionResponseMsg{Success: false}
	}
}
func (ds *EmployeeDataSource) SetSelectedByID(id string, selected bool) tea.Cmd { return nil }
func (ds *EmployeeDataSource) SelectAll() tea.Cmd                               { return nil }
func (ds *EmployeeDataSource) ClearSelection() tea.Cmd                          { return nil }
func (ds *EmployeeDataSource) SelectRange(startIndex, endIndex int) tea.Cmd     { return nil }
func (ds *EmployeeDataSource) GetItemID(item any) string {
	if row, ok := item.(core.TableRow); ok {
		return row.ID
	}
	return ""
}

func formatNumber(n int) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	result := ""
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}
	return result
}

type AppModel struct {
	table         *table.Table
	dataSource    *EmployeeDataSource
	statusMessage string
	// Cursor visualization fields
	fullRowHighlightEnabled bool
	activeCellEnabled       bool
	activeCellColorIndex    int
	activeCellColors        []string
	// Horizontal scrolling fields (now managed by table)
	// These are just for display purposes
	scrollModeLabels []string
}

// `05-table-component/examples/horizontal-scrolling/main.go`
func main() {
	dataSource := NewEmployeeDataSource()

	// Create columns with varying widths - fewer columns but keep description for horizontal scrolling demo
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

	// Simple default theme
	theme := core.Theme{
		HeaderStyle:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")),
		CellStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		CursorStyle:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")),
		SelectedStyle:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("57")),
		FullRowCursorStyle: lipgloss.NewStyle().Background(lipgloss.Color("205")).Foreground(lipgloss.Color("15")).Bold(true),
		BorderChars: core.BorderChars{
			Horizontal: "─", Vertical: "│", TopLeft: "┌", TopRight: "┐",
			BottomLeft: "└", BottomRight: "┘", TopT: "┬", BottomT: "┴",
			LeftT: "├", RightT: "┤", Cross: "┼",
		},
		BorderColor: "8",
		HeaderColor: "99",
	}

	config := core.TableConfig{
		Columns:                     columns,
		ShowHeader:                  true,
		ShowBorders:                 true,
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

	tbl := table.NewTable(config, dataSource)
	tbl.Focus()

	model := AppModel{
		table:                   tbl,
		dataSource:              dataSource,
		statusMessage:           "Horizontal Scrolling Demo - Use ← → for scrolling, .,/colnavs, s/S for modes",
		fullRowHighlightEnabled: false, // Start with full row highlighting disabled
		activeCellEnabled:       true,  // Start with active cell indication enabled
		activeCellColorIndex:    0,
		activeCellColors:        activeCellColors,
		scrollModeLabels:        scrollModeLabels,
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

		// === CURSOR VISUALIZATION CONTROLS ===
		case "r":
			// Toggle full row highlighting
			if m.fullRowHighlightEnabled {
				m.fullRowHighlightEnabled = false
				m.statusMessage = "Full row highlighting: DISABLED"
				return m, core.FullRowHighlightEnableCmd(false)
			} else {
				m.fullRowHighlightEnabled = true
				if m.activeCellEnabled {
					m.activeCellEnabled = false
				}
				m.statusMessage = "Full row highlighting: ENABLED"
				return m, core.FullRowHighlightEnableCmd(true)
			}

		case "c":
			// Toggle active cell indication
			if m.activeCellEnabled {
				m.activeCellEnabled = false
				m.statusMessage = "Active cell indication: DISABLED"
				return m, core.ActiveCellIndicationModeSetCmd(false)
			} else {
				m.activeCellEnabled = true
				if m.fullRowHighlightEnabled {
					m.fullRowHighlightEnabled = false
				}
				m.statusMessage = "Active cell indication: ENABLED"
				return m, core.ActiveCellIndicationModeSetCmd(true)
			}

		case "C":
			// Cycle active cell background colors
			m.activeCellColorIndex = (m.activeCellColorIndex + 1) % len(m.activeCellColors)
			newColor := m.activeCellColors[m.activeCellColorIndex]
			m.statusMessage = fmt.Sprintf("Active cell color: %s", newColor)
			return m, core.ActiveCellBackgroundColorSetCmd(newColor)

		case "m":
			// Toggle mixed mode (both row highlighting and active cell)
			if m.fullRowHighlightEnabled && m.activeCellEnabled {
				// Both on, turn both off
				m.fullRowHighlightEnabled = false
				m.activeCellEnabled = false
				m.statusMessage = "Mixed mode: DISABLED"
			} else {
				// Turn both on
				m.fullRowHighlightEnabled = true
				m.activeCellEnabled = true
				m.statusMessage = "Mixed mode: ENABLED"
			}
			return m, tea.Batch(
				core.FullRowHighlightEnableCmd(m.fullRowHighlightEnabled),
				core.ActiveCellIndicationModeSetCmd(m.activeCellEnabled),
			)

		// === HORIZONTAL SCROLLING CONTROLS ===
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
			// Horizontal scroll left
			m.statusMessage = "Horizontal scroll left"
			return m, core.HorizontalScrollLeftCmd()
		case ",":
			// Horizontal scroll right
			m.statusMessage = "Horizontal scroll right"
			return m, core.HorizontalScrollRightCmd()
		case "backspace", "delete":
			// Reset horizontal scrolling
			m.statusMessage = "Resetting horizontal scroll"
			return m, core.HorizontalScrollResetCmd()
		case "home":
			// Jump to start
			m.statusMessage = "Jumping to start"
			return m, core.JumpToStartCmd()
		case "end":
			// Jump to end
			m.statusMessage = "Jumping to end"
			return m, core.JumpToEndCmd()
		case "s":
			// Toggle scroll mode (for display purposes)
			m.statusMessage = "Toggling scroll mode"
			return m, core.HorizontalScrollModeToggleCmd()
		case "S":
			// Toggle scroll scope (for display purposes)
			m.statusMessage = "Toggling scroll scope"
			return m.table.Update(core.HorizontalScrollScopeToggleMsg{})

		// === NAVIGATION KEYS ===
		case "j", "down":
			return m, core.CursorDownCmd()
		case "k", "up":
			return m, core.CursorUpCmd()
		case "left":
			return m, core.CursorLeftCmd()
		case "right":
			return m, core.CursorRightCmd()
		case "h":
			m.statusMessage = "Page up"
			return m, core.PageUpCmd()
		case "l":
			m.statusMessage = "Page down"
			return m, core.PageDownCmd()
		case "g":
			m.statusMessage = "Jumping to start"
			return m, core.JumpToStartCmd()
		case "G":
			m.statusMessage = "Jumping to end"
			return m, core.JumpToEndCmd()

		// Pass all other keys to table
		default:
			var cmd tea.Cmd
			_, cmd = m.table.Update(msg)
			return m, cmd
		}

	// Pass ALL other messages to table
	default:
		var cmd tea.Cmd
		_, cmd = m.table.Update(msg)
		return m, cmd
	}
}

func (m AppModel) View() string {
	// Status line showing current modes
	var cursorMode string
	if m.fullRowHighlightEnabled && m.activeCellEnabled {
		cursorMode = "mixed"
	} else if m.fullRowHighlightEnabled {
		cursorMode = "row"
	} else if m.activeCellEnabled {
		cursorMode = "cell"
	} else {
		cursorMode = "none"
	}

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

	// Get table state for position info
	state := m.table.GetState()
	currentColor := m.activeCellColors[m.activeCellColorIndex]

	status := fmt.Sprintf("Employee %d/%d | Cursor: %s | Color: %s | HScroll: %s (%s) | Col: %d | Scope: %s",
		state.CursorIndex+1,
		m.table.GetTotalItems(),
		cursorMode,
		currentColor,
		scrollStatus,
		scrollMode,
		currentColumn,
		scopeStatus,
	)

	controls := "←/→=column | </>/[]/{}HL=scroll | Del=reset | s=mode | S=scope | rcCm=cursor | ↑↓jk=move | gG=start/end | q=quit"

	return status + "\n" + controls + "\n" + m.statusMessage + "\n\n" + m.table.View()
}
