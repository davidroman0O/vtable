package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/table"
)

// LargeEmployeeDataSource with enhanced data for constraint demonstration
type LargeEmployeeDataSource struct {
	totalEmployees int
	data           []core.TableRow
	selectedItems  map[string]bool
	recentActivity []string
}

func NewLargeEmployeeDataSource(totalCount int) *LargeEmployeeDataSource {
	// Generate ALL data upfront
	data := make([]core.TableRow, totalCount)

	departments := []string{"Engineering", "Marketing", "Sales", "HR", "Finance", "Operations"}
	statuses := []string{"Active", "On Leave", "Remote"}

	// Long descriptions to demonstrate text constraints and truncation
	longDescriptions := []string{
		"Experienced software engineer specializing in backend systems and database optimization with 5+ years",
		"Creative marketing professional focused on digital campaigns and brand management across multiple channels",
		"Senior sales representative with expertise in B2B relationships and territory management nationwide",
		"Human resources specialist handling recruitment, employee relations, and organizational development programs",
		"Financial analyst responsible for budget planning, forecasting, and quarterly reporting to executive team",
		"Operations manager overseeing logistics, supply chain optimization, and process improvement initiatives",
		"Product manager driving feature development and cross-functional collaboration with engineering teams",
		"Customer success specialist ensuring client satisfaction and managing long-term partnership relationships",
	}

	for i := 0; i < totalCount; i++ {
		data[i] = core.TableRow{
			ID: fmt.Sprintf("emp-%d", i+1),
			Cells: []string{
				fmt.Sprintf("Employee %d", i+1),
				departments[rand.Intn(len(departments))],
				statuses[rand.Intn(len(statuses))],
				fmt.Sprintf("$%d,000", 45+rand.Intn(100)),
				longDescriptions[i%len(longDescriptions)], // Long description for truncation demo
			},
		}
	}

	return &LargeEmployeeDataSource{
		totalEmployees: totalCount,
		data:           data,
		selectedItems:  make(map[string]bool),
		recentActivity: make([]string, 0),
	}
}

// Same selection methods as the working selection example
func (ds *LargeEmployeeDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(10 * time.Millisecond)
		return core.DataTotalMsg{Total: ds.totalEmployees}
	}
}

func (ds *LargeEmployeeDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

		start := request.Start
		end := start + request.Count
		if end > ds.totalEmployees {
			end = ds.totalEmployees
		}

		var items []core.Data[any]
		for i := start; i < end; i++ {
			if i < len(ds.data) {
				items = append(items, core.Data[any]{
					ID:       ds.data[i].ID,
					Item:     ds.data[i],
					Selected: ds.selectedItems[ds.data[i].ID],
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

func (ds *LargeEmployeeDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

func (ds *LargeEmployeeDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		if index >= 0 && index < len(ds.data) {
			id := ds.data[index].ID

			if selected {
				ds.selectedItems[id] = true
				ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected: %s", ds.data[index].Cells[0]))
			} else {
				delete(ds.selectedItems, id)
				ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Deselected: %s", ds.data[index].Cells[0]))
			}

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

func (ds *LargeEmployeeDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	return func() tea.Msg {
		for i, row := range ds.data {
			if row.ID == id {
				if selected {
					ds.selectedItems[id] = true
					ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected: %s", row.Cells[0]))
				} else {
					delete(ds.selectedItems, id)
					ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Deselected: %s", row.Cells[0]))
				}

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

func (ds *LargeEmployeeDataSource) ClearSelection() tea.Cmd {
	return func() tea.Msg {
		count := len(ds.selectedItems)
		ds.selectedItems = make(map[string]bool)
		ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Cleared %d selections", count))

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

func (ds *LargeEmployeeDataSource) SelectAll() tea.Cmd {
	return func() tea.Msg {
		for _, row := range ds.data {
			ds.selectedItems[row.ID] = true
		}

		ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected all %d items", len(ds.data)))

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

func (ds *LargeEmployeeDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	return func() tea.Msg {
		var affectedIDs []string
		count := 0

		for i := startIndex; i <= endIndex && i < len(ds.data); i++ {
			ds.selectedItems[ds.data[i].ID] = true
			affectedIDs = append(affectedIDs, ds.data[i].ID)
			count++
		}

		ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected range: %d items", count))

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

func (ds *LargeEmployeeDataSource) GetItemID(item any) string {
	if row, ok := item.(core.TableRow); ok {
		return row.ID
	}
	return ""
}

func (ds *LargeEmployeeDataSource) GetRecentActivity() []string {
	return ds.recentActivity
}

func (ds *LargeEmployeeDataSource) GetSelectionCount() int {
	return len(ds.selectedItems)
}

// Application structure with constraint state tracking
type App struct {
	table          *table.Table
	dataSource     *LargeEmployeeDataSource
	statusMessage  string
	totalEmployees int

	// Jump-to-index form
	showJumpForm bool
	jumpInput    string

	// NEW: Constraint state tracking
	widthMode        int // 0=narrow, 1=normal, 2=wide
	alignmentMode    int // 0=mixed, 1=all-left, 2=all-center, 3=all-right
	headerAlignMode  int // 0=mixed, 1=all-left, 2=all-center, 3=all-right
	paddingMode      int // 0=none, 1=normal, 2=extra
	descriptionWidth int // Cycle through different widths
}

func (app *App) cycleColumnWidths() {
	app.widthMode = (app.widthMode + 1) % 3
	switch app.widthMode {
	case 0:
		app.statusMessage = "Column widths: NARROW (compact layout)"
	case 1:
		app.statusMessage = "Column widths: NORMAL (balanced layout)"
	case 2:
		app.statusMessage = "Column widths: WIDE (spacious layout)"
	}
}

func (app *App) cycleAlignment() {
	app.alignmentMode = (app.alignmentMode + 1) % 4
	switch app.alignmentMode {
	case 0:
		app.statusMessage = "Data alignment: MIXED (name=left, dept=center, status=center, salary=right)"
	case 1:
		app.statusMessage = "Data alignment: ALL LEFT"
	case 2:
		app.statusMessage = "Data alignment: ALL CENTER"
	case 3:
		app.statusMessage = "Data alignment: ALL RIGHT"
	}
}

func (app *App) cycleHeaderAlignment() {
	app.headerAlignMode = (app.headerAlignMode + 1) % 4
	switch app.headerAlignMode {
	case 0:
		app.statusMessage = "Header alignment: MIXED (different from data alignment)"
	case 1:
		app.statusMessage = "Header alignment: ALL LEFT"
	case 2:
		app.statusMessage = "Header alignment: ALL CENTER"
	case 3:
		app.statusMessage = "Header alignment: ALL RIGHT"
	}
}

func (app *App) cyclePadding() {
	app.paddingMode = (app.paddingMode + 1) % 3
	switch app.paddingMode {
	case 0:
		app.statusMessage = "Padding: NONE (tight layout)"
	case 1:
		app.statusMessage = "Padding: NORMAL (comfortable spacing)"
	case 2:
		app.statusMessage = "Padding: EXTRA (spacious layout)"
	}
}

func (app *App) cycleDescriptionWidth() {
	widths := []int{20, 30, 40, 50}
	app.descriptionWidth = (app.descriptionWidth + 1) % len(widths)
	app.statusMessage = fmt.Sprintf("Description width: %d characters (see truncation effect)", widths[app.descriptionWidth])
}

func (app *App) updateTableColumns() tea.Cmd {
	columns := app.buildColumnsWithConstraints()
	return core.ColumnSetCmd(columns)
}

func (app *App) buildColumnsWithConstraints() []core.TableColumn {
	// Base widths for each mode
	var nameWidth, deptWidth, statusWidth, salaryWidth, descWidth int

	switch app.widthMode {
	case 0: // Narrow
		nameWidth, deptWidth, statusWidth, salaryWidth = 15, 12, 10, 10
	case 1: // Normal
		nameWidth, deptWidth, statusWidth, salaryWidth = 20, 15, 12, 12
	case 2: // Wide
		nameWidth, deptWidth, statusWidth, salaryWidth = 25, 20, 15, 15
	}

	// Description width based on cycle
	descWidths := []int{20, 30, 40, 50}
	descWidth = descWidths[app.descriptionWidth]

	// Padding based on mode
	var leftPad, rightPad int
	switch app.paddingMode {
	case 0: // None
		leftPad, rightPad = 0, 0
	case 1: // Normal
		leftPad, rightPad = 1, 1
	case 2: // Extra
		leftPad, rightPad = 2, 2
	}

	columns := []core.TableColumn{
		{
			Title:           "Employee Name",
			Field:           "name",
			Width:           nameWidth,
			Alignment:       getColumnAlignment(app.alignmentMode, core.AlignLeft),
			HeaderAlignment: getColumnAlignment(app.headerAlignMode, core.AlignCenter),
			HeaderConstraint: core.CellConstraint{
				Width:     nameWidth,
				Alignment: getColumnAlignment(app.headerAlignMode, core.AlignCenter),
				Padding:   core.PaddingConfig{Left: leftPad + 1, Right: rightPad + 1},
			},
		},
		{
			Title:           "Department",
			Field:           "department",
			Width:           deptWidth,
			Alignment:       getColumnAlignment(app.alignmentMode, core.AlignCenter),
			HeaderAlignment: getColumnAlignment(app.headerAlignMode, core.AlignLeft),
			HeaderConstraint: core.CellConstraint{
				Width:     deptWidth,
				Alignment: getColumnAlignment(app.headerAlignMode, core.AlignLeft),
				Padding:   core.PaddingConfig{Left: leftPad, Right: rightPad},
			},
		},
		{
			Title:           "Status",
			Field:           "status",
			Width:           statusWidth,
			Alignment:       getColumnAlignment(app.alignmentMode, core.AlignCenter),
			HeaderAlignment: getColumnAlignment(app.headerAlignMode, core.AlignRight),
			HeaderConstraint: core.CellConstraint{
				Width:     statusWidth,
				Alignment: getColumnAlignment(app.headerAlignMode, core.AlignRight),
				Padding:   core.PaddingConfig{Left: leftPad, Right: rightPad},
			},
		},
		{
			Title:           "Salary",
			Field:           "salary",
			Width:           salaryWidth,
			Alignment:       getColumnAlignment(app.alignmentMode, core.AlignRight),
			HeaderAlignment: getColumnAlignment(app.headerAlignMode, core.AlignCenter),
			HeaderConstraint: core.CellConstraint{
				Width:     salaryWidth,
				Alignment: getColumnAlignment(app.headerAlignMode, core.AlignCenter),
				Padding:   core.PaddingConfig{Left: leftPad, Right: rightPad},
			},
		},
		{
			Title:           "Description",
			Field:           "description",
			Width:           descWidth,
			Alignment:       getColumnAlignment(app.alignmentMode, core.AlignLeft),
			HeaderAlignment: getColumnAlignment(app.headerAlignMode, core.AlignLeft),
			HeaderConstraint: core.CellConstraint{
				Width:     descWidth,
				Alignment: getColumnAlignment(app.headerAlignMode, core.AlignLeft),
				Padding:   core.PaddingConfig{Left: leftPad, Right: rightPad},
			},
		},
	}

	return columns
}

func getColumnAlignment(mode int, defaultAlign int) int {
	switch mode {
	case 0: // Mixed - use default
		return defaultAlign
	case 1: // All left
		return core.AlignLeft
	case 2: // All center
		return core.AlignCenter
	case 3: // All right
		return core.AlignRight
	default:
		return defaultAlign
	}
}

func createTableConfig() core.TableConfig {
	return core.TableConfig{
		Columns:       createInitialEmployeeColumns(),
		ShowHeader:    true,
		ShowBorders:   true,
		SelectionMode: core.SelectionMultiple,
		ViewportConfig: core.ViewportConfig{
			Height:             10,
			ChunkSize:          25,
			TopThreshold:       3,
			BottomThreshold:    3,
			BoundingAreaBefore: 50,
			BoundingAreaAfter:  50,
		},
		Theme: config.DefaultTheme(),
		KeyMap: core.NavigationKeyMap{
			Up:        []string{"up", "k"},
			Down:      []string{"down", "j"},
			PageUp:    []string{"pgup", "h"},
			PageDown:  []string{"pgdown", "l"},
			Home:      []string{"home", "g"},
			End:       []string{"end", "G"},
			Select:    []string{"enter", " "},
			SelectAll: []string{"ctrl+a"},
			Quit:      []string{"q"},
		},
	}
}

func createInitialEmployeeColumns() []core.TableColumn {
	return []core.TableColumn{
		{
			Title:           "Employee Name",
			Field:           "name",
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
			Title:           "Department",
			Field:           "department",
			Width:           15,
			Alignment:       core.AlignCenter,
			HeaderAlignment: core.AlignLeft,
			HeaderConstraint: core.CellConstraint{
				Width:     15,
				Alignment: core.AlignLeft,
				Padding:   core.PaddingConfig{Left: 1, Right: 0},
			},
		},
		{
			Title:           "Status",
			Field:           "status",
			Width:           12,
			Alignment:       core.AlignCenter,
			HeaderAlignment: core.AlignRight,
			HeaderConstraint: core.CellConstraint{
				Width:     12,
				Alignment: core.AlignRight,
				Padding:   core.PaddingConfig{Left: 0, Right: 1},
			},
		},
		{
			Title:           "Salary",
			Field:           "salary",
			Width:           12,
			Alignment:       core.AlignRight,
			HeaderAlignment: core.AlignCenter,
			HeaderConstraint: core.CellConstraint{
				Width:     12,
				Alignment: core.AlignCenter,
				Padding:   core.PaddingConfig{Left: 1, Right: 1},
			},
		},
		{
			Title:           "Description",
			Field:           "description",
			Width:           30,
			Alignment:       core.AlignLeft,
			HeaderAlignment: core.AlignLeft,
			HeaderConstraint: core.CellConstraint{
				Width:     30,
				Alignment: core.AlignLeft,
				Padding:   core.PaddingConfig{Left: 1, Right: 1},
			},
		},
	}
}

func (app App) Init() tea.Cmd {
	return tea.Batch(
		app.table.Init(),
		app.table.Focus(),
	)
}

func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle jump form if it's open
		if app.showJumpForm {
			switch msg.String() {
			case "enter":
				if index, err := strconv.Atoi(app.jumpInput); err == nil && index > 0 {
					targetIndex := index - 1
					if targetIndex < app.totalEmployees {
						app.showJumpForm = false
						app.jumpInput = ""
						app.updateStatus()
						return app, core.JumpToCmd(targetIndex)
					}
				}
				app.showJumpForm = false
				app.jumpInput = ""
				app.updateStatus()
				return app, nil
			case "esc":
				app.showJumpForm = false
				app.jumpInput = ""
				app.updateStatus()
				return app, nil
			case "backspace":
				if len(app.jumpInput) > 0 {
					app.jumpInput = app.jumpInput[:len(app.jumpInput)-1]
				}
				return app, nil
			default:
				if len(msg.String()) == 1 && msg.String() >= "0" && msg.String() <= "9" {
					app.jumpInput += msg.String()
				}
				return app, nil
			}
		}

		// Normal key handling when form is not open
		switch msg.String() {
		case "q", "ctrl+c":
			return app, tea.Quit
		case "J":
			app.showJumpForm = true
			app.jumpInput = ""
			return app, nil

		// Selection commands (same as before)
		case " ", "enter":
			return app, core.SelectCurrentCmd()
		case "ctrl+a":
			return app, core.SelectAllCmd()
		case "c":
			return app, core.SelectClearCmd()
		case "s":
			app.showSelectionInfo()
			return app, nil

		// NEW: Cell constraint controls
		case "w":
			app.cycleColumnWidths()
			return app, app.updateTableColumns()
		case "a":
			app.cycleAlignment()
			return app, app.updateTableColumns()
		case "A":
			app.cycleHeaderAlignment()
			return app, app.updateTableColumns()
		case "p":
			app.cyclePadding()
			return app, app.updateTableColumns()
		case "t":
			app.cycleDescriptionWidth()
			return app, app.updateTableColumns()

		default:
			var cmd tea.Cmd
			_, cmd = app.table.Update(msg)
			app.updateStatus()
			return app, cmd
		}

	// Handle selection responses
	case core.SelectionResponseMsg:
		app.updateStatus()
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd

	// Handle total count received
	case core.DataTotalMsg:
		app.totalEmployees = msg.Total
		app.updateStatus()
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd

	// Handle chunk loading completed
	case core.DataChunkLoadedMsg:
		app.updateStatus()
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd

	default:
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd
	}
}

func (app *App) showSelectionInfo() {
	count := app.dataSource.GetSelectionCount()
	if count > 0 {
		app.statusMessage = fmt.Sprintf("✓ %d employees selected | Use c to clear, space to toggle", count)
	} else {
		app.statusMessage = "No employees selected | Use space to select, ctrl+a for all"
	}
}

func (app *App) updateStatus() {
	state := app.table.GetState()
	selectionCount := app.dataSource.GetSelectionCount()

	if app.showJumpForm {
		app.statusMessage = fmt.Sprintf("Enter employee number (1-%d), Enter to jump, Esc to cancel", app.totalEmployees)
	} else {
		app.statusMessage = fmt.Sprintf("Employee %d/%d | Selected: %d",
			state.CursorIndex+1, app.totalEmployees, selectionCount)
	}
}

func (app App) View() string {
	var sections []string

	// Show jump form if active
	if app.showJumpForm {
		sections = append(sections, fmt.Sprintf("Jump to employee (1-%d): %s_", app.totalEmployees, app.jumpInput))
		sections = append(sections, "")
	}

	// Status message
	sections = append(sections, app.statusMessage)
	sections = append(sections, "")

	// Table
	sections = append(sections, app.table.View())

	// Always show controls
	sections = append(sections, "")
	sections = append(sections, "Controls: ↑↓/jk=move, Space=select, W=width, A=align, Shift+A=header-align, P=pad, T=text-width, q=quit")
	sections = append(sections, "Constraints: Adjust column widths, alignments, padding, and text truncation in real-time")

	// Show constraint info
	sections = append(sections, "")
	sections = append(sections, fmt.Sprintf("Constraints: Width=%s | Data=%s | Header=%s | Padding=%s | Desc=%dch",
		[]string{"narrow", "normal", "wide"}[app.widthMode],
		[]string{"mixed", "left", "center", "right"}[app.alignmentMode],
		[]string{"mixed", "left", "center", "right"}[app.headerAlignMode],
		[]string{"none", "normal", "extra"}[app.paddingMode],
		[]int{20, 30, 40, 50}[app.descriptionWidth]))

	// Show selection info
	selectionCount := app.dataSource.GetSelectionCount()
	if selectionCount > 0 {
		sections = append(sections, "")
		sections = append(sections, fmt.Sprintf("Selected: %d items", selectionCount))
	}

	// Show recent activity
	recentActivity := app.dataSource.GetRecentActivity()
	if len(recentActivity) > 0 {
		sections = append(sections, "")
		sections = append(sections, "Recent Activity:")
		for i := len(recentActivity) - 1; i >= 0 && i >= len(recentActivity)-3; i-- {
			sections = append(sections, fmt.Sprintf("  • %s", recentActivity[i]))
		}
	}

	return strings.Join(sections, "\n")
}

// `05-table-component/examples/cell-constraints/main.go`
func main() {
	// Create large dataset with selection tracking
	dataSource := NewLargeEmployeeDataSource(1000)
	tableConfig := createTableConfig()

	// Create table with constraints enabled
	employeeTable := table.NewTable(tableConfig, dataSource)

	// Create app with constraint tracking
	app := App{
		table:            employeeTable,
		dataSource:       dataSource,
		statusMessage:    "Cell Constraints Demo | All controls shown below",
		widthMode:        1, // Start with normal width
		alignmentMode:    0, // Start with mixed alignment
		headerAlignMode:  0, // Start with mixed header alignment
		paddingMode:      1, // Start with normal padding
		descriptionWidth: 1, // Start with 30 char description
	}

	// Run the program
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
