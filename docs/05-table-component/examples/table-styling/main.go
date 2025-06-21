package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/table"
)

type Employee struct {
	ID          string
	Name        string
	Department  string
	Status      string
	Salary      int
	HireDate    time.Time
	Performance string
	Location    string
}

type LargeEmployeeDataSource struct {
	totalEmployees int
	data           []Employee
	selectedItems  map[string]bool
	recentActivity []string
}

func NewLargeEmployeeDataSource(totalCount int) *LargeEmployeeDataSource {
	data := make([]Employee, totalCount)
	departments := []string{"Engineering", "Marketing", "Sales", "HR", "Finance", "Operations"}
	statuses := []string{"Active", "On Leave", "Remote"}
	performances := []string{"Excellent", "Good", "Average", "Needs Improvement"}
	locations := []string{"New York", "San Francisco", "Austin", "Seattle", "Boston", "Denver"}
	firstNames := []string{"Alice", "Bob", "Carol", "David", "Eve", "Frank", "Grace", "Henry", "Ivy", "Jack"}
	lastNames := []string{"Johnson", "Smith", "Davis", "Wilson", "Brown", "Miller", "Lee", "Taylor", "Chen", "Roberts"}

	for i := 0; i < totalCount; i++ {
		daysAgo := rand.Intn(3650)
		hireDate := time.Now().AddDate(0, 0, -daysAgo)

		data[i] = Employee{
			ID:          fmt.Sprintf("emp-%d", i+1),
			Name:        fmt.Sprintf("%s %s", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))]),
			Department:  departments[rand.Intn(len(departments))],
			Status:      statuses[rand.Intn(len(statuses))],
			Salary:      45000 + rand.Intn(100000),
			HireDate:    hireDate,
			Performance: performances[rand.Intn(len(performances))],
			Location:    locations[rand.Intn(len(locations))],
		}
	}

	return &LargeEmployeeDataSource{
		totalEmployees: totalCount,
		data:           data,
		selectedItems:  make(map[string]bool),
		recentActivity: make([]string, 0),
	}
}

func (ds *LargeEmployeeDataSource) employeeToTableRow(emp Employee) core.TableRow {
	return core.TableRow{
		ID: emp.ID,
		Cells: []string{
			emp.Name,
			emp.Department,
			emp.Status,
			fmt.Sprintf("%d", emp.Salary),
			emp.HireDate.Format("Jan 2006"),
		},
	}
}

func (ds *LargeEmployeeDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return core.DataTotalMsg{Total: ds.totalEmployees}
	}
}

func (ds *LargeEmployeeDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		start := request.Start
		end := start + request.Count
		if end > ds.totalEmployees {
			end = ds.totalEmployees
		}

		var items []core.Data[any]
		for i := start; i < end; i++ {
			if i < len(ds.data) {
				tableRow := ds.employeeToTableRow(ds.data[i])
				items = append(items, core.Data[any]{
					ID:       ds.data[i].ID,
					Item:     tableRow,
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

func (ds *LargeEmployeeDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	return func() tea.Msg {
		for i, emp := range ds.data {
			if emp.ID == id {
				if selected {
					ds.selectedItems[id] = true
				} else {
					delete(ds.selectedItems, id)
				}
				return core.SelectionResponseMsg{
					Success:  true,
					Index:    i,
					ID:       id,
					Selected: selected,
				}
			}
		}
		return core.SelectionResponseMsg{Success: false}
	}
}

func (ds *LargeEmployeeDataSource) ClearSelection() tea.Cmd {
	return func() tea.Msg {
		ds.selectedItems = make(map[string]bool)
		return core.SelectionResponseMsg{Success: true}
	}
}

func (ds *LargeEmployeeDataSource) SelectAll() tea.Cmd {
	return func() tea.Msg {
		for _, emp := range ds.data {
			ds.selectedItems[emp.ID] = true
		}
		return core.SelectionResponseMsg{Success: true}
	}
}

func (ds *LargeEmployeeDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	return func() tea.Msg {
		for i := startIndex; i <= endIndex && i < len(ds.data); i++ {
			ds.selectedItems[ds.data[i].ID] = true
		}
		return core.SelectionResponseMsg{Success: true}
	}
}

func (ds *LargeEmployeeDataSource) GetItemID(item any) string {
	if emp, ok := item.(Employee); ok {
		return emp.ID
	}
	return ""
}

func (ds *LargeEmployeeDataSource) GetRecentActivity() []string {
	return ds.recentActivity
}

func (ds *LargeEmployeeDataSource) GetSelectionCount() int {
	return len(ds.selectedItems)
}

// Theme definitions
type TableTheme struct {
	Name        string
	Description string
	BorderChars core.BorderChars
	Colors      ThemeColors
}

type ThemeColors struct {
	BorderColor    string
	HeaderColor    string
	CellColor      string
	CursorColor    string
	SelectionColor string
}

var themes = []TableTheme{
	{
		Name:        "Default",
		Description: "Clean Unicode box drawing theme",
		BorderChars: core.BorderChars{
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
		Colors: ThemeColors{
			BorderColor:    "8",   // Gray
			HeaderColor:    "99",  // Purple
			CellColor:      "252", // Light gray
			CursorColor:    "205", // Pink
			SelectionColor: "57",  // Blue
		},
	},
	{
		Name:        "Heavy",
		Description: "Double-line borders for emphasis",
		BorderChars: core.BorderChars{
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
		},
		Colors: ThemeColors{
			BorderColor:    "14",  // Cyan
			HeaderColor:    "11",  // Yellow
			CellColor:      "15",  // White
			CursorColor:    "22",  // Dark green
			SelectionColor: "235", // Dark gray
		},
	},
	{
		Name:        "Minimal",
		Description: "Clean borderless design",
		BorderChars: core.BorderChars{
			Horizontal:  " ",
			Vertical:    " ",
			TopLeft:     " ",
			TopRight:    " ",
			BottomLeft:  " ",
			BottomRight: " ",
			TopT:        " ",
			BottomT:     " ",
			LeftT:       " ",
			RightT:      " ",
			Cross:       " ",
		},
		Colors: ThemeColors{
			BorderColor:    "8",   // Gray (hidden)
			HeaderColor:    "0",   // Black
			CellColor:      "0",   // Black
			CursorColor:    "7",   // Light gray
			SelectionColor: "235", // Dark gray
		},
	},
	{
		Name:        "Retro",
		Description: "ASCII retro computing style",
		BorderChars: core.BorderChars{
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
		},
		Colors: ThemeColors{
			BorderColor:    "14",  // Cyan
			HeaderColor:    "13",  // Magenta
			CellColor:      "15",  // White
			CursorColor:    "201", // Bright magenta
			SelectionColor: "235", // Dark gray
		},
	},
}

func convertToVTableTheme(theme TableTheme) core.Theme {
	return core.Theme{
		HeaderStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.HeaderColor)).Bold(true),
		CellStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.CellColor)),
		CursorStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Colors.CursorColor)).Bold(true),
		SelectedStyle:      lipgloss.NewStyle().Background(lipgloss.Color(theme.Colors.SelectionColor)).Foreground(lipgloss.Color("230")),
		FullRowCursorStyle: lipgloss.NewStyle().Background(lipgloss.Color(theme.Colors.CursorColor)).Foreground(lipgloss.Color("15")).Bold(true),
		BorderChars:        theme.BorderChars,
		BorderColor:        theme.Colors.BorderColor,
		HeaderColor:        theme.Colors.HeaderColor,
		AlternateRowStyle:  lipgloss.NewStyle().Background(lipgloss.Color("235")),
		DisabledStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
		LoadingStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Italic(true),
		ErrorStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
	}
}

type App struct {
	table               *table.Table
	dataSource          *LargeEmployeeDataSource
	statusMessage       string
	currentTheme        int
	showBorders         bool
	showTopBorder       bool
	showBottomBorder    bool
	showHeaderSeparator bool
	removeTopSpace      bool
	removeBottomSpace   bool
}

// Same formatters as column-formatting example
func nameFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
	return "👤 " + cellValue
}

func deptFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
	icons := map[string]string{
		"Engineering": "🔧",
		"Marketing":   "📢",
		"Sales":       "💼",
		"HR":          "👥",
		"Finance":     "💰",
		"Operations":  "⚙️",
	}
	if icon, exists := icons[cellValue]; exists {
		return icon + " " + cellValue
	}
	return "🏢 " + cellValue
}

func statusFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
	switch cellValue {
	case "Active":
		return "🟢 " + cellValue
	case "On Leave":
		return "🟡 " + cellValue
	case "Remote":
		return "🔵 " + cellValue
	default:
		return "⚪ " + cellValue
	}
}

func salaryFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
	if salary, err := strconv.Atoi(cellValue); err == nil {
		formatted := "$" + formatNumber(salary)
		if salary >= 100000 {
			return "💎 " + formatted
		} else if salary >= 75000 {
			return "💰 " + formatted
		} else if salary >= 50000 {
			return "💵 " + formatted
		} else {
			return "💳 " + formatted
		}
	}
	return cellValue
}

func dateFormatter(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor, isSelected, isActiveCell bool) string {
	return "📅 " + cellValue
}

func formatNumber(n int) string {
	str := strconv.Itoa(n)
	if len(str) > 3 {
		return str[:len(str)-3] + "," + str[len(str)-3:]
	}
	return str
}

func createStyledTableConfig() core.TableConfig {
	return core.TableConfig{
		Columns: []core.TableColumn{
			{Title: "Employee", Field: "name", Width: 25, Alignment: core.AlignLeft},
			{Title: "Department", Field: "department", Width: 20, Alignment: core.AlignCenter},
			{Title: "Status", Field: "status", Width: 15, Alignment: core.AlignCenter},
			{Title: "Salary", Field: "salary", Width: 18, Alignment: core.AlignRight},
			{Title: "Hire Date", Field: "hire_date", Width: 15, Alignment: core.AlignCenter},
		},
		ShowHeader:              true,
		ShowBorders:             true,
		ShowTopBorder:           true,
		ShowBottomBorder:        true,
		ShowHeaderSeparator:     true,
		RemoveTopBorderSpace:    false,
		RemoveBottomBorderSpace: false,
		ViewportConfig: core.ViewportConfig{
			Height:             10,
			ChunkSize:          25,
			TopThreshold:       3,
			BottomThreshold:    3,
			BoundingAreaBefore: 50,
			BoundingAreaAfter:  50,
		},
		Theme:         convertToVTableTheme(themes[0]),
		SelectionMode: core.SelectionMultiple,
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

func (app App) Init() tea.Cmd {
	return tea.Batch(
		app.table.Init(),
		app.table.Focus(),
		// Keep the same formatters from column-formatting example
		core.CellFormatterSetCmd(0, nameFormatter),
		core.CellFormatterSetCmd(1, deptFormatter),
		core.CellFormatterSetCmd(2, statusFormatter),
		core.CellFormatterSetCmd(3, salaryFormatter),
		core.CellFormatterSetCmd(4, dateFormatter),
	)
}

func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return app, tea.Quit
		case " ", "enter":
			return app, core.SelectCurrentCmd()
		case "ctrl+a":
			return app, core.SelectAllCmd()
		case "c":
			return app, core.SelectClearCmd()

		// Theme controls
		case "t":
			app.currentTheme = (app.currentTheme + 1) % len(themes)
			newTheme := themes[app.currentTheme]
			vtableTheme := convertToVTableTheme(newTheme)
			app.statusMessage = fmt.Sprintf("Theme: %s - %s", newTheme.Name, newTheme.Description)

			// CRITICAL FIX: Call table.SetTheme directly like the working example!
			return app, app.table.SetTheme(vtableTheme)

		// Border visibility controls
		case "b":
			app.showBorders = !app.showBorders
			if app.showBorders {
				app.statusMessage = "All borders enabled"
			} else {
				app.statusMessage = "All borders disabled"
			}
			return app, core.BorderVisibilityCmd(app.showBorders)

		case "1":
			app.showTopBorder = !app.showTopBorder
			if app.showTopBorder {
				app.statusMessage = "Top border enabled"
			} else {
				app.statusMessage = "Top border disabled"
			}
			return app, core.TopBorderVisibilityCmd(app.showTopBorder)

		case "2":
			app.showBottomBorder = !app.showBottomBorder
			if app.showBottomBorder {
				app.statusMessage = "Bottom border enabled"
			} else {
				app.statusMessage = "Bottom border disabled"
			}
			return app, core.BottomBorderVisibilityCmd(app.showBottomBorder)

		case "3":
			app.showHeaderSeparator = !app.showHeaderSeparator
			if app.showHeaderSeparator {
				app.statusMessage = "Header separator enabled"
			} else {
				app.statusMessage = "Header separator disabled"
			}
			return app, core.HeaderSeparatorVisibilityCmd(app.showHeaderSeparator)

		// Space removal controls
		case "4":
			app.removeTopSpace = !app.removeTopSpace
			if app.removeTopSpace {
				app.statusMessage = "Top border space removed - compact layout"
			} else {
				app.statusMessage = "Top border space preserved"
			}
			return app, core.TopBorderSpaceRemovalCmd(app.removeTopSpace)

		case "5":
			app.removeBottomSpace = !app.removeBottomSpace
			if app.removeBottomSpace {
				app.statusMessage = "Bottom border space removed - compact layout"
			} else {
				app.statusMessage = "Bottom border space preserved"
			}
			return app, core.BottomBorderSpaceRemovalCmd(app.removeBottomSpace)

		default:
			var cmd tea.Cmd
			_, cmd = app.table.Update(msg)
			return app, cmd
		}

	// Forward border messages to the table
	case core.BorderVisibilityMsg,
		core.TopBorderVisibilityMsg,
		core.BottomBorderVisibilityMsg,
		core.HeaderSeparatorVisibilityMsg,
		core.TopBorderSpaceRemovalMsg,
		core.BottomBorderSpaceRemovalMsg:
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd

	default:
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd
	}
}

func (app App) View() string {
	var sections []string

	sections = append(sections, "Table Styling Demo - Themes & Border Controls")
	sections = append(sections, "")
	sections = append(sections, app.table.View())
	sections = append(sections, "")

	// Controls
	sections = append(sections, "Controls: ↑↓/jk=move, Space=select, ctrl+a=select all, c=clear, q=quit")
	sections = append(sections, "Styling: T=theme, B=borders, 1=top, 2=bottom, 3=separator, 4=top-space, 5=bottom-space")
	sections = append(sections, "")

	// Current styling state
	if app.currentTheme < len(themes) {
		currentTheme := themes[app.currentTheme]
		sections = append(sections, fmt.Sprintf("Theme: %s (%s)", currentTheme.Name, currentTheme.Description))
	}

	borderStatus := "Off"
	if app.showBorders {
		borderStatus = "On"
	}
	sections = append(sections, fmt.Sprintf("Borders: %s | Top: %t | Bottom: %t | Header: %t",
		borderStatus, app.showTopBorder, app.showBottomBorder, app.showHeaderSeparator))

	if app.removeTopSpace || app.removeBottomSpace {
		sections = append(sections, fmt.Sprintf("Space Removal: Top=%t | Bottom=%t (Compact mode)",
			app.removeTopSpace, app.removeBottomSpace))
	}

	return strings.Join(sections, "\n")
}

func main() {
	dataSource := NewLargeEmployeeDataSource(1000)
	tableConfig := createStyledTableConfig()
	employeeTable := table.NewTable(tableConfig, dataSource)

	app := App{
		table:               employeeTable,
		dataSource:          dataSource,
		statusMessage:       "Table Styling Demo | Press T to cycle themes",
		currentTheme:        0,
		showBorders:         true,
		showTopBorder:       true,
		showBottomBorder:    true,
		showHeaderSeparator: true,
		removeTopSpace:      false,
		removeBottomSpace:   false,
	}

	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
