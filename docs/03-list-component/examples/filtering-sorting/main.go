package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/list"
)

// Person represents our data structure
type Person struct {
	Name string
	Age  int
	Job  string
	City string
}

// PersonDataSource implements a stateful DataSource for filtering and sorting
type PersonDataSource struct {
	people         []Person
	filteredData   []Person // The data after filters and sorts are applied
	activeFilters  map[string]any
	sortFields     []string
	sortDirections []string
}

func NewPersonDataSource() *PersonDataSource {
	ds := &PersonDataSource{
		people: []Person{
			{"Alice Johnson", 28, "UX Designer", "San Francisco"},
			{"Bob Chen", 34, "Software Engineer", "New York"},
			{"Carol Rodriguez", 45, "Product Manager", "Austin"},
			{"David Kim", 29, "DevOps Engineer", "Seattle"},
			{"Emma Thompson", 33, "Data Scientist", "Boston"},
			{"Frank Wilson", 41, "Tech Lead", "Denver"},
			{"Grace Lee", 26, "Frontend Developer", "Portland"},
			{"Henry Garcia", 38, "Backend Developer", "Chicago"},
			{"Ivy Martinez", 31, "QA Engineer", "Miami"},
			{"Jack Brown", 27, "Mobile Developer", "Los Angeles"},
			{"Kate Davis", 35, "Product Designer", "San Diego"},
			{"Liam Johnson", 42, "Engineering Manager", "Phoenix"},
			{"Maya Patel", 30, "Full Stack Developer", "Nashville"},
			{"Noah Williams", 25, "Junior Developer", "Atlanta"},
			{"Olivia Miller", 39, "Senior Developer", "Dallas"},
			{"Peter Anderson", 36, "DevOps Engineer", "San Jose"},
			{"Quinn Taylor", 32, "Product Manager", "San Antonio"},
			{"Rachel White", 29, "UX Designer", "Portland"},
			{"Sam Brooks", 43, "Engineering Manager", "Boston"},
			{"Tara Kim", 31, "Software Engineer", "San Francisco"},
		},
		activeFilters: make(map[string]any),
	}
	// Initial data state is the full, unsorted list
	ds.rebuildFilteredData()
	return ds
}

// rebuildFilteredData is the core logic that applies current filters and sorts.
// This should be called whenever filters or sorts change.
func (ds *PersonDataSource) rebuildFilteredData() {
	// 1. Apply filters
	if len(ds.activeFilters) > 0 {
		var filtered []Person
		for _, person := range ds.people {
			if ds.matchesAllFilters(person) {
				filtered = append(filtered, person)
			}
		}
		ds.filteredData = filtered
	} else {
		// No filters, use the original data
		ds.filteredData = ds.people
	}

	// 2. Apply sorting to the filtered data
	if len(ds.sortFields) > 0 {
		sorted := make([]Person, len(ds.filteredData))
		copy(sorted, ds.filteredData)

		sort.Slice(sorted, func(i, j int) bool {
			for idx, field := range ds.sortFields {
				direction := "asc"
				if idx < len(ds.sortDirections) {
					direction = ds.sortDirections[idx]
				}

				var comparison int
				switch field {
				case "name":
					comparison = strings.Compare(sorted[i].Name, sorted[j].Name)
				case "age":
					comparison = sorted[i].Age - sorted[j].Age
				case "job":
					comparison = strings.Compare(sorted[i].Job, sorted[j].Job)
				case "city":
					comparison = strings.Compare(sorted[i].City, sorted[j].City)
				}

				if comparison != 0 {
					if direction == "desc" {
						return comparison > 0
					}
					return comparison < 0
				}
			}
			return false
		})
		ds.filteredData = sorted
	}
}

func (ds *PersonDataSource) matchesAllFilters(person Person) bool {
	for field, value := range ds.activeFilters {
		switch field {
		case "job":
			if !strings.Contains(strings.ToLower(person.Job), strings.ToLower(value.(string))) {
				return false
			}
		case "city":
			if !strings.Contains(strings.ToLower(person.City), strings.ToLower(value.(string))) {
				return false
			}
		case "minAge":
			if person.Age < value.(int) {
				return false
			}
		case "search":
			searchTerm := strings.ToLower(value.(string))
			if !strings.Contains(strings.ToLower(person.Name), searchTerm) &&
				!strings.Contains(strings.ToLower(person.Job), searchTerm) &&
				!strings.Contains(strings.ToLower(person.City), searchTerm) {
				return false
			}
		}
	}
	return true
}

// --- Public methods for the App to manage state ---

func (ds *PersonDataSource) ToggleFilter(field, value string) {
	if _, ok := ds.activeFilters[field]; ok {
		delete(ds.activeFilters, field)
	} else {
		ds.activeFilters[field] = value
	}
	ds.rebuildFilteredData()
}

func (ds *PersonDataSource) ToggleNumericFilter(field string, value int) {
	if _, ok := ds.activeFilters[field]; ok {
		delete(ds.activeFilters, field)
	} else {
		ds.activeFilters[field] = value
	}
	ds.rebuildFilteredData()
}

func (ds *PersonDataSource) SetSearch(term string) {
	if term == "" {
		delete(ds.activeFilters, "search")
	} else {
		ds.activeFilters["search"] = term
	}
	ds.rebuildFilteredData()
}

func (ds *PersonDataSource) ClearAllFilters() {
	ds.activeFilters = make(map[string]any)
	ds.rebuildFilteredData()
}

func (ds *PersonDataSource) ToggleSort(field string) {
	if len(ds.sortFields) > 0 && ds.sortFields[0] == field {
		if ds.sortDirections[0] == "asc" {
			ds.sortDirections[0] = "desc"
		} else {
			// Third toggle clears the sort
			ds.sortFields = []string{}
			ds.sortDirections = []string{}
		}
	} else {
		// New sort field
		ds.sortFields = []string{field}
		ds.sortDirections = []string{"asc"}
	}
	ds.rebuildFilteredData()
}

func (ds *PersonDataSource) ClearAllSorts() {
	ds.sortFields = []string{}
	ds.sortDirections = []string{}
	ds.rebuildFilteredData()
}

// --- VTable DataSource Interface Implementation ---

// GetTotal now returns the count of the pre-filtered/sorted data
func (ds *PersonDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return core.DataTotalMsg{Total: len(ds.filteredData)}
	}
}

// LoadChunk now serves data from the pre-filtered/sorted slice
func (ds *PersonDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		start := request.Start
		count := request.Count
		var items []core.Data[any]

		if start < len(ds.filteredData) {
			end := start + count
			if end > len(ds.filteredData) {
				end = len(ds.filteredData)
			}
			for i := start; i < end; i++ {
				items = append(items, core.Data[any]{
					ID:       fmt.Sprintf("person-%d", i),
					Item:     ds.filteredData[i],
					Selected: false, // Selection not implemented in this example
				})
			}
		}

		return core.DataChunkLoadedMsg{
			StartIndex: request.Start,
			Items:      items,
			Request:    request,
		}
	}
}

func (ds *PersonDataSource) RefreshTotal() tea.Cmd                            { return ds.GetTotal() }
func (ds *PersonDataSource) SetSelected(index int, selected bool) tea.Cmd     { return nil }
func (ds *PersonDataSource) SetSelectedByID(id string, selected bool) tea.Cmd { return nil }
func (ds *PersonDataSource) SelectAll() tea.Cmd                               { return nil }
func (ds *PersonDataSource) ClearSelection() tea.Cmd                          { return nil }
func (ds *PersonDataSource) SelectRange(startIndex, endIndex int) tea.Cmd     { return nil }
func (ds *PersonDataSource) GetItemID(item any) string                        { return fmt.Sprintf("%v", item) }

// --- Formatter ---
func styledPersonFormatter(data core.Data[any], index int, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
	person, ok := data.Item.(Person)
	if !ok {
		return fmt.Sprintf("Item %d: %v", index+1, data.Item)
	}

	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D4AA"))
	ageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	jobStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF"))
	cityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#DDA0DD"))

	return fmt.Sprintf("%s %s - %s in %s",
		nameStyle.Render(person.Name),
		ageStyle.Render(fmt.Sprintf("(%d)", person.Age)),
		jobStyle.Render(person.Job),
		cityStyle.Render(person.City),
	)
}

// --- Application Model ---
type App struct {
	list       *list.List
	dataSource *PersonDataSource // Hold a reference to the DataSource
	status     string
}

func (app *App) Init() tea.Cmd {
	return app.list.Init()
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return app, tea.Quit

		// --- Filter Controls ---
		case "1":
			app.status = "Toggled filter: Engineers"
			app.dataSource.ToggleFilter("job", "Engineer")
			return app, core.DataRefreshCmd()
		case "2":
			app.status = "Toggled filter: Managers"
			app.dataSource.ToggleFilter("job", "Manager")
			return app, core.DataRefreshCmd()
		case "3":
			app.status = "Toggled filter: Age 30+"
			app.dataSource.ToggleNumericFilter("minAge", 30)
			return app, core.DataRefreshCmd()

		// --- Sort Controls ---
		// Note: '!' is Shift+1, '@' is Shift+2, etc. on many keyboards
		case "!":
			app.status = "Toggling sort by Name"
			app.dataSource.ToggleSort("name")
			return app, core.DataRefreshCmd()
		case "@":
			app.status = "Toggling sort by Age"
			app.dataSource.ToggleSort("age")
			return app, core.DataRefreshCmd()
		case "#":
			app.status = "Toggling sort by Job"
			app.dataSource.ToggleSort("job")
			return app, core.DataRefreshCmd()

		// --- Clear Controls ---
		case "0":
			app.status = "Cleared all filters"
			app.dataSource.ClearAllFilters()
			return app, core.DataRefreshCmd()
		case "S":
			app.status = "Cleared all sorts"
			app.dataSource.ClearAllSorts()
			return app, core.DataRefreshCmd()

		// --- Navigation ---
		default:
			var cmd tea.Cmd
			_, cmd = app.list.Update(msg)
			return app, cmd
		}
	}

	// Pass all other messages to the list
	var cmd tea.Cmd
	_, cmd = app.list.Update(msg)
	return app, cmd
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

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true)

	title := titleStyle.Render("🔍 VTable Filtering & Sorting Demo")

	help := helpStyle.Render("Filters: 1=Eng 2=Mgr 3=30+ 0=Clear • Sorts (Shift+Num): !=Name @=Age #=Job S=Clear • Quit: q")

	return fmt.Sprintf("%s\n\n%s\n\n%s\n%s",
		title,
		app.list.View(),
		statusStyle.Render(app.status),
		help,
	)
}

// `03-list-component/examples/filtering-sorting/main.go`
func main() {
	dataSource := NewPersonDataSource()

	listConfig := config.DefaultListConfig()
	listConfig.ViewportConfig.Height = 10
	listConfig.MaxWidth = 800
	listConfig.SelectionMode = core.SelectionMultiple
	listConfig.RenderConfig.ContentConfig.Formatter = styledPersonFormatter

	vtableList := list.NewList(listConfig, dataSource)

	app := &App{
		list:       vtableList,
		dataSource: dataSource,
		status:     "Ready! Press number keys to filter, Shift+Number to sort.",
	}

	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
