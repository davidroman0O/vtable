package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/list"
)

// Person represents our rich data structure
type Person struct {
	Name string
	Age  int
	City string
	Job  string
}

// PersonDataSource manages our Person data with selection
type PersonDataSource struct {
	people   []Person
	selected map[int]bool
}

func NewPersonDataSource() *PersonDataSource {
	// Sample data with diverse people
	people := []Person{
		{"Alice Johnson", 28, "San Francisco", "UX Designer"},
		{"Bob Chen", 34, "New York", "Software Engineer"},
		{"Carol Rodriguez", 45, "Austin", "Product Manager"},
		{"David Kim", 29, "Seattle", "DevOps Engineer"},
		{"Emma Wilson", 52, "Portland", "Tech Lead"},
		{"Frank Taylor", 26, "Denver", "Frontend Developer"},
		{"Grace Patel", 38, "Boston", "Data Scientist"},
		{"Henry Martinez", 41, "Chicago", "Backend Engineer"},
		{"Iris Thompson", 33, "Miami", "QA Engineer"},
		{"Jack Brown", 47, "Phoenix", "Software Architect"},
		{"Kate Davis", 25, "Atlanta", "Junior Developer"},
		{"Luis Garcia", 36, "Dallas", "DevOps Engineer"},
		{"Maya Singh", 42, "San Diego", "Senior Engineer"},
		{"Noah Clark", 30, "Las Vegas", "Full Stack Developer"},
		{"Olivia Lee", 39, "Orlando", "Technical Writer"},
		{"Paul Anderson", 44, "Nashville", "Engineering Manager"},
		{"Quinn Murphy", 27, "Sacramento", "Mobile Developer"},
		{"Rachel Green", 35, "Salt Lake City", "Platform Engineer"},
		{"Sam White", 31, "Kansas City", "Cloud Engineer"},
		{"Tina Lopez", 48, "Tampa", "Principal Engineer"},
	}

	return &PersonDataSource{
		people:   people,
		selected: make(map[int]bool),
	}
}

func (ds *PersonDataSource) GetSelectedCount() int {
	return len(ds.selected)
}

func (ds *PersonDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return core.DataTotalMsg{Total: len(ds.people)}
	}
}

func (ds *PersonDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

func (ds *PersonDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		var items []core.Data[any]

		for i := request.Start; i < request.Start+request.Count && i < len(ds.people); i++ {
			items = append(items, core.Data[any]{
				ID:       fmt.Sprintf("person-%d", i),
				Item:     ds.people[i],
				Selected: ds.selected[i],
			})
		}

		return core.DataChunkLoadedMsg{
			StartIndex: request.Start,
			Items:      items,
			Request:    request,
		}
	}
}

func (ds *PersonDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		if index >= 0 && index < len(ds.people) {
			if selected {
				ds.selected[index] = true
			} else {
				delete(ds.selected, index)
			}
			return core.SelectionResponseMsg{
				Success:  true,
				Index:    index,
				ID:       fmt.Sprintf("person-%d", index),
				Selected: selected,
			}
		}
		return core.SelectionResponseMsg{Success: false}
	}
}

func (ds *PersonDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	var index int
	if _, err := fmt.Sscanf(id, "person-%d", &index); err == nil {
		return ds.SetSelected(index, selected)
	}
	return func() tea.Msg {
		return core.SelectionResponseMsg{Success: false}
	}
}

func (ds *PersonDataSource) SelectAll() tea.Cmd {
	return func() tea.Msg {
		for i := 0; i < len(ds.people); i++ {
			ds.selected[i] = true
		}
		return core.SelectionResponseMsg{Success: true}
	}
}

func (ds *PersonDataSource) ClearSelection() tea.Cmd {
	return func() tea.Msg {
		ds.selected = make(map[int]bool)
		return core.SelectionResponseMsg{Success: true}
	}
}

func (ds *PersonDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	return func() tea.Msg {
		if startIndex > endIndex {
			startIndex, endIndex = endIndex, startIndex
		}
		for i := startIndex; i <= endIndex && i < len(ds.people); i++ {
			ds.selected[i] = true
		}
		return core.SelectionResponseMsg{Success: true}
	}
}

func (ds *PersonDataSource) GetItemID(item any) string {
	return fmt.Sprintf("%v", item)
}

// Get job-specific color based on job title
func getJobColor(job string) lipgloss.Color {
	switch {
	case strings.Contains(job, "Engineer"):
		return "#00CED1" // Dark Turquoise for engineers
	case strings.Contains(job, "Manager"):
		return "#FF6347" // Tomato for managers
	case strings.Contains(job, "Designer"):
		return "#DA70D6" // Orchid for designers
	case strings.Contains(job, "Lead") || strings.Contains(job, "Principal"):
		return "#32CD32" // Lime Green for leads
	case strings.Contains(job, "Developer"):
		return "#87CEEB" // Sky Blue for developers
	case strings.Contains(job, "Architect"):
		return "#DDA0DD" // Plum for architects
	default:
		return "#F0E68C" // Khaki for others
	}
}

// Get age-specific color
func getAgeColor(age int) lipgloss.Color {
	if age < 30 {
		return "#FFD700" // Gold for young
	} else if age > 45 {
		return "#9370DB" // Purple for senior
	} else {
		return "#FFA500" // Orange for mid-career
	}
}

// Get age text with emoji
func getAgeText(age int) string {
	if age < 30 {
		return fmt.Sprintf("(%d) 🌟", age)
	} else if age > 45 {
		return fmt.Sprintf("(%d) 👑", age)
	} else {
		return fmt.Sprintf("(%d)", age)
	}
}

// Checkbox formatter - same as styled formatter but with checkbox prefix
func checkboxPersonFormatter(data core.Data[any], index int, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
	person := data.Item.(Person)

	// Add checkbox indicator
	var checkbox string
	if data.Selected {
		checkbox = "[x]" // Selected
	} else {
		checkbox = "[ ]" // Unselected
	}

	// Same styling as before
	var nameColor, ageColor, jobColor, cityColor lipgloss.Color

	if isCursor {
		nameColor, ageColor, jobColor, cityColor = "#FFFF00", "#FFD700", "#00FFFF", "#00FF00"
	} else if data.Selected {
		nameColor, ageColor, jobColor, cityColor = "#FF69B4", "#FFA500", "#87CEEB", "#98FB98"
	} else {
		nameColor = "#FFFFFF"
		ageColor = getAgeColor(person.Age)
		jobColor = getJobColor(person.Job)
		cityColor = "#98FB98"
	}

	// Style components
	styledCheckbox := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render(checkbox)
	styledName := lipgloss.NewStyle().Foreground(nameColor).Bold(true).Render(person.Name)
	styledAge := lipgloss.NewStyle().Foreground(ageColor).Render(getAgeText(person.Age))
	styledJob := lipgloss.NewStyle().Foreground(jobColor).Render(person.Job)
	styledCity := lipgloss.NewStyle().Foreground(cityColor).Render(person.City)

	// Format with checkbox prefix
	return fmt.Sprintf("%s %s %s - %s in %s",
		styledCheckbox, styledName, styledAge, styledJob, styledCity)
}

// App with checkbox styling
type App struct {
	list           *list.List
	dataSource     *PersonDataSource
	selectionCount int
	statusMessage  string
}

func (app *App) Init() tea.Cmd {
	return app.list.Init()
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return app, tea.Quit

		// Navigation
		case "up", "k":
			return app, core.CursorUpCmd()
		case "down", "j":
			return app, core.CursorDownCmd()
		case "pgup", "h":
			return app, core.PageUpCmd()
		case "pgdown", "l":
			return app, core.PageDownCmd()
		case "home", "g":
			return app, core.JumpToStartCmd()
		case "end", "G":
			return app, core.JumpToEndCmd()

		// Selection
		case " ": // Spacebar
			return app, core.SelectCurrentCmd()
		case "ctrl+a":
			return app, core.SelectAllCmd()
		case "ctrl+d":
			return app, core.SelectClearCmd()
		}

	case core.SelectionResponseMsg:
		if msg.Success {
			app.updateSelectionCount()
			app.updateStatusMessage()
		}
	}

	var cmd tea.Cmd
	_, cmd = app.list.Update(msg)
	return app, cmd
}

func (app *App) updateSelectionCount() {
	app.selectionCount = app.dataSource.GetSelectedCount()
}

func (app *App) updateStatusMessage() {
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)

	if app.selectionCount == 0 {
		app.statusMessage = statusStyle.Render("No people selected")
	} else if app.selectionCount == 1 {
		app.statusMessage = statusStyle.Render("1 person selected")
	} else {
		app.statusMessage = statusStyle.Render(fmt.Sprintf("%d people selected", app.selectionCount))
	}
}

func (app *App) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#5A5A5A")).
		Bold(true).
		Padding(0, 1).
		MarginBottom(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true)

	title := titleStyle.Render("☑ Checkbox List with Visual Selection")
	help := helpStyle.Render("Navigate: ↑/↓ j/k • Page: h/l • Jump: g/G • Select: Space • Multi: Ctrl+A/D • Quit: q")

	return fmt.Sprintf("%s\n\n%s\n\n%s\n%s",
		title,
		app.list.View(),
		help,
		app.statusMessage,
	)
}

// `03-list-component/examples/checkbox-list/main.go`
func main() {
	dataSource := NewPersonDataSource()

	listConfig := config.DefaultListConfig()
	listConfig.ViewportConfig.Height = 8
	listConfig.MaxWidth = 500 // Important: Allow width for checkboxes and styling
	listConfig.SelectionMode = core.SelectionMultiple

	// Set the formatter in the config (Option 3 approach)
	listConfig.RenderConfig.ContentConfig.Formatter = checkboxPersonFormatter

	// Create list with checkbox formatter
	vtableList := list.NewList(listConfig, dataSource)

	app := &App{
		list:           vtableList,
		dataSource:     dataSource,
		selectionCount: 0,
		statusMessage:  lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true).Render("No people selected"),
	}

	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
