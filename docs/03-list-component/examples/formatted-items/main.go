package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
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
		{"Henry Martinez", 41, "Chicago", "Backend Developer"},
		{"Iris Thompson", 33, "Miami", "QA Engineer"},
		{"Jack Brown", 47, "Phoenix", "Architect"},
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
				Item:     ds.people[i], // Pass the Person struct
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

// Custom formatter function for Person data
func personFormatter(data core.Data[any], index int, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
	person := data.Item.(Person) // Type assertion to access Person fields

	// Add age styling based on value
	var ageDisplay string
	if person.Age < 30 {
		ageDisplay = fmt.Sprintf("(%d) 🌟", person.Age) // Young professional
	} else if person.Age > 45 {
		ageDisplay = fmt.Sprintf("(%d) 👑", person.Age) // Senior professional
	} else {
		ageDisplay = fmt.Sprintf("(%d)", person.Age) // Mid-career
	}

	// Format: "Name (Age) - Job in City"
	return fmt.Sprintf("%s %s - %s in %s",
		person.Name,
		ageDisplay,
		person.Job,
		person.City)
}

// App with formatting capabilities
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
	if app.selectionCount == 0 {
		app.statusMessage = "No people selected"
	} else if app.selectionCount == 1 {
		app.statusMessage = "1 person selected"
	} else {
		app.statusMessage = fmt.Sprintf("%d people selected", app.selectionCount)
	}
}

func (app *App) View() string {
	return fmt.Sprintf(
		"Formatted Items List (press 'q' to quit)\n\n%s\n\n%s\n%s",
		app.list.View(),
		"Navigate: ↑/↓ j/k • Page: h/l • Jump: g/G • Select: Space • Multi: Ctrl+A/D",
		app.statusMessage,
	)
}

func main() {
	dataSource := NewPersonDataSource()

	listConfig := config.DefaultListConfig()
	listConfig.ViewportConfig.Height = 8
	listConfig.SelectionMode = core.SelectionMultiple

	// Create list with custom person formatter
	vtableList := list.NewList(listConfig, dataSource, personFormatter)

	app := &App{
		list:           vtableList,
		dataSource:     dataSource,
		selectionCount: 0,
		statusMessage:  "No people selected",
	}

	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
