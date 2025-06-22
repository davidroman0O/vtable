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

type App struct {
	table         *table.Table
	dataSource    *LargeEmployeeDataSource
	statusMessage string
}

// Simple, stateless formatters that work with any data
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
		if salary >= 100000 {
			return "💎 $" + formatNumber(salary)
		} else if salary >= 75000 {
			return "💰 $" + formatNumber(salary)
		} else if salary >= 50000 {
			return "💵 $" + formatNumber(salary)
		} else {
			return "💳 $" + formatNumber(salary)
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

func createTableConfig() core.TableConfig {
	return core.TableConfig{
		Columns: []core.TableColumn{
			{Title: "Employee", Field: "name", Width: 25, Alignment: core.AlignLeft},
			{Title: "Department", Field: "department", Width: 20, Alignment: core.AlignCenter},
			{Title: "Status", Field: "status", Width: 15, Alignment: core.AlignCenter},
			{Title: "Salary", Field: "salary", Width: 18, Alignment: core.AlignRight},
			{Title: "Hire Date", Field: "hire_date", Width: 15, Alignment: core.AlignCenter},
		},
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

func (app App) Init() tea.Cmd {
	return tea.Batch(
		app.table.Init(),
		app.table.Focus(),
		// Set formatters once
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
		default:
			var cmd tea.Cmd
			_, cmd = app.table.Update(msg)
			return app, cmd
		}
	default:
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd
	}
}

func (app App) View() string {
	var sections []string
	sections = append(sections, "Column Formatting Demo - Simple & Working")
	sections = append(sections, "")
	sections = append(sections, app.table.View())
	sections = append(sections, "")
	sections = append(sections, "Controls: ↑↓/jk=move, Space=select, ctrl+a=select all, c=clear, q=quit")
	sections = append(sections, "Formatting: Simple emoji icons for each column type")
	return strings.Join(sections, "\n")
}

// `05-table-component/examples/column-formatting/main.go`
func main() {
	dataSource := NewLargeEmployeeDataSource(1000)
	tableConfig := createTableConfig()
	employeeTable := table.NewTable(tableConfig, dataSource)

	app := App{
		table:         employeeTable,
		dataSource:    dataSource,
		statusMessage: "Simple Column Formatting Demo",
	}

	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
