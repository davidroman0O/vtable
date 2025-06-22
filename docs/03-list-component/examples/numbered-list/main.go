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

// Same styled formatter as before - VTable adds enumerators automatically
func styledPersonFormatter(data core.Data[any], index int, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
	person := data.Item.(Person)

	// Same styling as before - no changes needed
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

	// Style components (same as before)
	styledName := lipgloss.NewStyle().Foreground(nameColor).Bold(true).Render(person.Name)
	styledAge := lipgloss.NewStyle().Foreground(ageColor).Render(getAgeText(person.Age))
	styledJob := lipgloss.NewStyle().Foreground(jobColor).Render(person.Job)
	styledCity := lipgloss.NewStyle().Foreground(cityColor).Render(person.City)

	// Return formatted content (VTable adds enumerators automatically)
	return fmt.Sprintf("%s %s - %s in %s",
		styledName, styledAge, styledJob, styledCity)
}

// Custom enumerator functions to demonstrate the system

// Custom bracket numbers: [1] [2] [3]
func customBracketEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	return fmt.Sprintf("[%d] ", index+1)
}

// Smart enumerator that changes based on selection
func smartEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	if item.Selected {
		return "✓ " // Checkmark for selected
	}
	return fmt.Sprintf("%d. ", index+1) // Numbers for unselected
}

// Job-aware enumerator that shows emojis based on job type
func jobAwareEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	person := item.Item.(Person)

	if strings.Contains(person.Job, "Manager") {
		return "👑 "
	} else if strings.Contains(person.Job, "Engineer") {
		return "⚙️ "
	} else if strings.Contains(person.Job, "Designer") {
		return "🎨 "
	}
	return fmt.Sprintf("%d. ", index+1)
}

// App demonstrating different enumerator approaches
type App struct {
	list              *list.List
	dataSource        *PersonDataSource
	selectionCount    int
	statusMessage     string
	currentEnumerator int
	enumeratorNames   []string
	fullRowSelection  bool // Track full row selection state
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

		// NEW: Cycle through enumerator styles
		case "e":
			app.currentEnumerator = (app.currentEnumerator + 1) % len(app.enumeratorNames)
			app.setEnumeratorStyle()

		// Toggle full row selection
		case "b":
			app.fullRowSelection = !app.fullRowSelection
			app.setEnumeratorStyle()
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

func (app *App) setEnumeratorStyle() {
	switch app.currentEnumerator {
	case 0: // Convenience method: Numbered
		app.list.SetNumberedStyle()
		// Fix MaxWidth for proper right alignment
		renderConfig := app.list.GetRenderConfig()
		renderConfig.EnumeratorConfig.MaxWidth = 4
		app.list.SetRenderConfig(renderConfig)

	case 1: // Convenience method: Bullets
		app.list.SetBulletStyle()

	case 2: // Convenience method: Checkboxes
		app.list.SetChecklistStyle()

	case 3: // Direct config: Custom brackets
		renderConfig := app.list.GetRenderConfig()
		renderConfig.EnumeratorConfig.Enumerator = customBracketEnumerator
		renderConfig.EnumeratorConfig.Alignment = core.ListAlignmentRight
		renderConfig.EnumeratorConfig.MaxWidth = 5
		app.list.SetRenderConfig(renderConfig)

	case 4: // Custom function: Smart conditional
		renderConfig := app.list.GetRenderConfig()
		renderConfig.EnumeratorConfig.Enumerator = smartEnumerator
		renderConfig.EnumeratorConfig.Alignment = core.ListAlignmentLeft
		renderConfig.EnumeratorConfig.MaxWidth = 3
		app.list.SetRenderConfig(renderConfig)

	case 5: // Custom function: Data-aware
		renderConfig := app.list.GetRenderConfig()
		renderConfig.EnumeratorConfig.Enumerator = jobAwareEnumerator
		renderConfig.EnumeratorConfig.Alignment = core.ListAlignmentLeft
		renderConfig.EnumeratorConfig.MaxWidth = 3
		app.list.SetRenderConfig(renderConfig)
	}

	// Apply full row selection styling based on toggle state
	if app.fullRowSelection {
		selectedBg := lipgloss.NewStyle().Background(lipgloss.Color("#444444"))
		app.list.SetFullRowSelection(true, selectedBg)
	} else {
		// Disable full row selection, only content gets background
		app.list.SetFullRowSelection(false, lipgloss.NewStyle())
		// Apply selection background only to content component
		selectedBg := lipgloss.NewStyle().Background(lipgloss.Color("#444444"))
		app.list.SetComponentBackgroundStyling(core.ListComponentContent, lipgloss.NewStyle(), selectedBg, lipgloss.NewStyle(), false, true, false)
	}
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

	enumStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00CED1")).
		Bold(true)

	backgroundStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B35")).
		Bold(true)

	title := titleStyle.Render("🔢 VTable Enumerator System Demo")
	currentEnum := enumStyle.Render(fmt.Sprintf("Current: %s (%d/%d)",
		app.enumeratorNames[app.currentEnumerator], app.currentEnumerator+1, len(app.enumeratorNames)))

	// Background mode status
	var backgroundMode string
	if app.fullRowSelection {
		backgroundMode = backgroundStyle.Render("Background: Full Row Selection")
	} else {
		backgroundMode = backgroundStyle.Render("Background: Content Only Selection")
	}

	help := helpStyle.Render("Navigate: ↑/↓ j/k • Page: h/l • Jump: g/G • Select: Space • Multi: Ctrl+A/D • Enumerator: e • Toggle: b • Quit: q")

	return fmt.Sprintf("%s\n%s\n%s\n\n%s\n\n%s\n%s",
		title,
		currentEnum,
		backgroundMode,
		app.list.View(),
		help,
		app.statusMessage,
	)
}

// `03-list-component/examples/numbered-list/main.go`
func main() {
	dataSource := NewPersonDataSource()

	listConfig := config.DefaultListConfig()
	listConfig.ViewportConfig.Height = 8
	listConfig.MaxWidth = 500 // Important: Allow width for enumerators and styling
	listConfig.SelectionMode = core.SelectionMultiple

	// Option 3: Formatter in config (clean and explicit)
	listConfig.RenderConfig.ContentConfig.Formatter = styledPersonFormatter

	// Create list with everything configured upfront
	vtableList := list.NewList(listConfig, dataSource)

	app := &App{
		list:       vtableList,
		dataSource: dataSource,
		enumeratorNames: []string{
			"Arabic Numbers (1. 2. 3.)",
			"Bullet Points (• • •)",
			"Checkboxes (☐ ☑)",
			"Custom Brackets ([1] [2])",
			"Smart Conditional (✓/numbers)",
			"Job-Aware Emojis (👑 ⚙️ 🎨)",
		},
		currentEnumerator: 0,
		fullRowSelection:  true, // Start with full row selection to show the fix
		selectionCount:    0,
		statusMessage:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true).Render("No people selected"),
	}

	// Start with numbered style
	app.setEnumeratorStyle()

	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
