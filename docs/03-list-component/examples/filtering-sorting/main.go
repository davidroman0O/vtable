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

// PersonDataSource implements core.DataSource
// The key insight: VTable sends filters/sorts in every DataRequest
type PersonDataSource struct {
	people   []Person
	selected map[int]bool

	// Cache filtered data to get accurate counts
	lastFilters  map[string]any
	lastSorts    []string
	lastSortDirs []string
	filteredData []Person
}

func NewPersonDataSource() *PersonDataSource {
	return &PersonDataSource{
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
		selected:     make(map[int]bool),
		lastFilters:  make(map[string]any),
		filteredData: []Person{},
	}
}

// LoadChunk receives filter/sort parameters from VTable automatically
func (ds *PersonDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		// Update cached filtered data if filters/sorts changed
		ds.updateFilteredData(request.Filters, request.SortFields, request.SortDirections)

		// Return the requested chunk from filtered/sorted data
		start := request.Start
		count := request.Count

		var items []core.Data[any]
		for i := start; i < start+count && i < len(ds.filteredData); i++ {
			items = append(items, core.Data[any]{
				ID:       fmt.Sprintf("person-%d", i),
				Item:     ds.filteredData[i],
				Selected: ds.selected[i], // Selection by filtered index
			})
		}

		return core.DataChunkLoadedMsg{
			StartIndex: request.Start,
			Items:      items,
			Request:    request,
		}
	}
}

func (ds *PersonDataSource) updateFilteredData(filters map[string]any, sortFields []string, sortDirs []string) {
	// Check if we need to recalculate
	filtersChanged := !mapsEqual(ds.lastFilters, filters)
	sortsChanged := !slicesEqual(ds.lastSorts, sortFields) || !slicesEqual(ds.lastSortDirs, sortDirs)

	if filtersChanged || sortsChanged || len(ds.filteredData) == 0 {
		// Apply filters and sorts
		filtered := ds.applyFilters(ds.people, filters)
		sorted := ds.applySorts(filtered, sortFields, sortDirs)

		// Update cache
		ds.filteredData = sorted
		ds.lastFilters = copyMap(filters)
		ds.lastSorts = copySlice(sortFields)
		ds.lastSortDirs = copySlice(sortDirs)

		// Reset selection when data changes
		ds.selected = make(map[int]bool)
	}
}

// Helper functions for comparison
func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func copyMap(m map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range m {
		result[k] = v
	}
	return result
}

func copySlice(s []string) []string {
	result := make([]string, len(s))
	copy(result, s)
	return result
}

// Apply filtering logic based on VTable's filter parameters
func (ds *PersonDataSource) applyFilters(people []Person, filters map[string]any) []Person {
	if len(filters) == 0 {
		return people // No filters active
	}

	var filtered []Person
	for _, person := range people {
		if ds.matchesAllFilters(person, filters) {
			filtered = append(filtered, person)
		}
	}
	return filtered
}

func (ds *PersonDataSource) matchesAllFilters(person Person, filters map[string]any) bool {
	for field, value := range filters {
		switch field {
		case "job":
			if jobFilter, ok := value.(string); ok {
				if !strings.Contains(strings.ToLower(person.Job), strings.ToLower(jobFilter)) {
					return false
				}
			}
		case "city":
			if cityFilter, ok := value.(string); ok {
				if !strings.Contains(strings.ToLower(person.City), strings.ToLower(cityFilter)) {
					return false
				}
			}
		case "minAge":
			if ageFilter, ok := value.(int); ok {
				if person.Age < ageFilter {
					return false
				}
			}
		case "search":
			if searchTerm, ok := value.(string); ok {
				searchLower := strings.ToLower(searchTerm)
				if !strings.Contains(strings.ToLower(person.Name), searchLower) &&
					!strings.Contains(strings.ToLower(person.Job), searchLower) &&
					!strings.Contains(strings.ToLower(person.City), searchLower) {
					return false
				}
			}
		}
	}
	return true // Passes all filters
}

// Apply sorting logic based on VTable's sort parameters
func (ds *PersonDataSource) applySorts(people []Person, fields []string, directions []string) []Person {
	if len(fields) == 0 {
		return people // No sorting requested
	}

	// Make a copy to avoid modifying the original
	sorted := make([]Person, len(people))
	copy(sorted, people)

	sort.Slice(sorted, func(i, j int) bool {
		for idx, field := range fields {
			direction := "asc"
			if idx < len(directions) {
				direction = directions[idx]
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

	return sorted
}

// GetTotal must return count after filtering
func (ds *PersonDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		// Ensure filtered data is initialized with no filters/sorts
		if len(ds.filteredData) == 0 {
			ds.updateFilteredData(make(map[string]any), []string{}, []string{})
		}
		// Return the count of filtered data
		return core.DataTotalMsg{Total: len(ds.filteredData)}
	}
}

func (ds *PersonDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

// Required DataSource interface methods
func (ds *PersonDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		if index >= 0 && index < len(ds.filteredData) {
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
		for i := 0; i < len(ds.filteredData); i++ {
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
		for i := startIndex; i <= endIndex && i < len(ds.filteredData); i++ {
			ds.selected[i] = true
		}
		return core.SelectionResponseMsg{Success: true}
	}
}

func (ds *PersonDataSource) GetItemID(item any) string {
	return fmt.Sprintf("%v", item)
}

// Formatter function for styled person display
func styledPersonFormatter(data core.Data[any], index int, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
	person, ok := data.Item.(Person)
	if !ok {
		return fmt.Sprintf("Item %d: %v", index+1, data.Item)
	}

	// Styled components
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D4AA"))
	ageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	// Job styling with role-specific colors
	var jobStyle lipgloss.Style
	if strings.Contains(person.Job, "Manager") {
		jobStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")) // Gold
	} else if strings.Contains(person.Job, "Engineer") || strings.Contains(person.Job, "Developer") {
		jobStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF")) // Blue
	} else if strings.Contains(person.Job, "Designer") {
		jobStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF69B4")) // Pink
	} else {
		jobStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#98FB98")) // Light green
	}

	cityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#DDA0DD"))

	return fmt.Sprintf("%s %s - %s in %s",
		nameStyle.Render(person.Name),
		ageStyle.Render(fmt.Sprintf("(%d)", person.Age)),
		jobStyle.Render(person.Job),
		cityStyle.Render(person.City),
	)
}

// Application model - much simpler!
type App struct {
	list   *list.List
	status string
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

		// Filter controls - use VTable's built-in commands
		case "1":
			app.status = "Filtering by Engineers"
			return app, core.FilterSetCmd("job", "Engineer")
		case "2":
			app.status = "Filtering by Managers"
			return app, core.FilterSetCmd("job", "Manager")
		case "3":
			app.status = "Filtering by age 30+"
			return app, core.FilterSetCmd("minAge", 30)
		case "4":
			app.status = "Filtering by San (cities)"
			return app, core.FilterSetCmd("city", "San")
		case "5":
			app.status = "Searching for 'Developer'"
			return app, core.FilterSetCmd("search", "Developer")

		// Sort controls - use VTable's built-in commands
		case "!":
			app.status = "Toggling sort by Name"
			return app, core.SortToggleCmd("name")
		case "@":
			app.status = "Toggling sort by Age"
			return app, core.SortToggleCmd("age")
		case "#":
			app.status = "Toggling sort by Job"
			return app, core.SortToggleCmd("job")
		case "$":
			app.status = "Toggling sort by City"
			return app, core.SortToggleCmd("city")

		// Clear controls
		case "r":
			app.status = "Cleared all filters and sorts"
			return app, tea.Batch(
				core.FiltersClearAllCmd(),
				core.SortsClearAllCmd(),
			)

		// Navigation - pass through to VTable
		case "up", "k":
			return app, core.CursorUpCmd()
		case "down", "j":
			return app, core.CursorDownCmd()
		case "pgup":
			return app, core.PageUpCmd()
		case "pgdn":
			return app, core.PageDownCmd()
		case "home", "g":
			return app, core.JumpToStartCmd()
		case "end", "G":
			return app, core.JumpToEndCmd()
		case " ":
			return app, core.SelectCurrentCmd()
		}
	}

	// VTable handles all filter/sort state automatically
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

	help := helpStyle.Render("Filters: 1=Engineer 2=Manager 3=30+ 4=San 5=Developer • Sorts: !=Name @=Age #=Job $=City • Clear: r=All • Quit: q")

	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
		title,
		app.list.View(),
		statusStyle.Render(app.status),
		help,
	)
}

func main() {
	// Create data source that implements the filtering interface
	dataSource := NewPersonDataSource()

	// Configure list
	listConfig := config.DefaultListConfig()
	listConfig.ViewportConfig.Height = 10
	listConfig.MaxWidth = 800
	listConfig.SelectionMode = core.SelectionMultiple
	listConfig.RenderConfig.ContentConfig.Formatter = styledPersonFormatter

	// Create list - filtering/sorting is built-in!
	vtableList := list.NewList(listConfig, dataSource)

	// Create application - much simpler now
	app := &App{
		list:   vtableList,
		status: "Ready! VTable's filtering/sorting is built-in. Try the number keys and symbols.",
	}

	// Run the application
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
