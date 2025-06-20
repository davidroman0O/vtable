package main

import (
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/list"
)

// Custom messages for our demo
type toggleBackgroundsMsg struct{}
type increaseSpacingMsg struct{}
type decreaseSpacingMsg struct{}
type cycleComponentMsg struct{}

// Commands
func toggleBackgroundsCmd() tea.Cmd {
	return func() tea.Msg {
		return toggleBackgroundsMsg{}
	}
}

func increaseSpacingCmd() tea.Cmd {
	return func() tea.Msg {
		return increaseSpacingMsg{}
	}
}

func decreaseSpacingCmd() tea.Cmd {
	return func() tea.Msg {
		return decreaseSpacingMsg{}
	}
}

func cycleComponentCmd() tea.Cmd {
	return func() tea.Msg {
		return cycleComponentMsg{}
	}
}

// Component types for width adjustment
type ComponentType int

const (
	ComponentSpacing ComponentType = iota
	ComponentCursor
	ComponentEnumerator
	ComponentContent
)

func (c ComponentType) String() string {
	switch c {
	case ComponentSpacing:
		return "Spacing"
	case ComponentCursor:
		return "Cursor"
	case ComponentEnumerator:
		return "Enumerator"
	case ComponentContent:
		return "Content"
	default:
		return "Unknown"
	}
}

// Person represents our data structure
type Person struct {
	Name string
	Age  int
	Job  string
	City string
}

// PersonDataSource implements core.DataSource for Person data
type PersonDataSource struct {
	people   []Person
	selected map[int]bool
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
		},
		selected: make(map[int]bool),
	}
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

// Formatter function for styled person display
func styledPersonFormatter(data core.Data[any], index int, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
	person := data.Item.(Person)

	// Name styling
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D4AA"))

	// Age styling
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

	// City styling
	cityStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#DDA0DD"))

	return fmt.Sprintf("%s %s - %s in %s",
		nameStyle.Render(person.Name),
		ageStyle.Render(fmt.Sprintf("(%d)", person.Age)),
		jobStyle.Render(person.Job),
		cityStyle.Render(person.City),
	)
}

// Application model
type App struct {
	list               *list.List
	currentLayout      int
	layoutNames        []string
	backgroundsEnabled bool
	selectedComponent  ComponentType
	// Individual component widths
	spacingWidth    int
	cursorWidth     int
	enumeratorWidth int
	contentWidth    int
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
		case "c":
			// Cycle through component layouts
			app.currentLayout = (app.currentLayout + 1) % len(app.layoutNames)
			app.setLayout()
			return app, nil
		case "w":
			// Cycle through components for width adjustment
			return app, cycleComponentCmd()
		case "b":
			return app, toggleBackgroundsCmd()
		case "+", "=":
			return app, increaseSpacingCmd()
		case "-", "_":
			return app, decreaseSpacingCmd()
		case "up", "k":
			return app, core.CursorUpCmd()
		case "down", "j":
			return app, core.CursorDownCmd()
		case "pgup", "h":
			return app, core.PageUpCmd()
		case "pgdn", "l":
			return app, core.PageDownCmd()
		case "home", "g":
			return app, core.JumpToStartCmd()
		case "end", "G":
			return app, core.JumpToEndCmd()
		case " ":
			return app, core.SelectCurrentCmd()
		case "ctrl+a":
			return app, core.SelectAllCmd()
		case "ctrl+d":
			return app, core.SelectClearCmd()
		}

	case toggleBackgroundsMsg:
		app.backgroundsEnabled = !app.backgroundsEnabled
		app.setLayout()
		return app, nil

	case cycleComponentMsg:
		app.selectedComponent = ComponentType((int(app.selectedComponent) + 1) % 4)
		return app, nil

	case increaseSpacingMsg:
		switch app.selectedComponent {
		case ComponentSpacing:
			if app.spacingWidth < 10 {
				app.spacingWidth++
				app.setLayout()
			}
		case ComponentCursor:
			if app.cursorWidth < 10 {
				app.cursorWidth++
				app.setLayout()
			}
		case ComponentEnumerator:
			if app.enumeratorWidth < 20 {
				app.enumeratorWidth++
				app.setLayout()
			}
		case ComponentContent:
			if app.contentWidth < 100 {
				app.contentWidth++
				app.setLayout()
			}
		}
		return app, nil

	case decreaseSpacingMsg:
		switch app.selectedComponent {
		case ComponentSpacing:
			if app.spacingWidth > 0 {
				app.spacingWidth--
				app.setLayout()
			}
		case ComponentCursor:
			if app.cursorWidth > 1 {
				app.cursorWidth--
				app.setLayout()
			}
		case ComponentEnumerator:
			if app.enumeratorWidth > 1 {
				app.enumeratorWidth--
				app.setLayout()
			}
		case ComponentContent:
			if app.contentWidth > 10 {
				app.contentWidth--
				app.setLayout()
			}
		}
		return app, nil
	}

	// Pass other messages to the list
	var cmd tea.Cmd
	_, cmd = app.list.Update(msg)
	return app, cmd
}

func (app *App) setLayout() {
	renderConfig := app.list.GetRenderConfig()

	// Create spacing string based on current width
	spacingStr := strings.Repeat(" ", app.spacingWidth)

	switch app.currentLayout {
	case 0: // Default: [Cursor][Enumerator][Content]
		app.list.SetNumberedStyle()
		renderConfig = app.list.GetRenderConfig()

	case 1: // Numbers at end: [Cursor][Content][Enumerator]
		renderConfig.EnumeratorConfig.Enumerator = list.ArabicEnumerator
		renderConfig.EnumeratorConfig.Alignment = core.ListAlignmentRight
		renderConfig.ComponentOrder = []core.ListComponentType{
			core.ListComponentCursor,
			core.ListComponentContent,
			core.ListComponentEnumerator,
		}

	case 2: // Content only: [Content]
		renderConfig.ComponentOrder = []core.ListComponentType{
			core.ListComponentContent,
		}

	case 3: // With spacing: [PreSpacing][Cursor][Enumerator][Content][PostSpacing]
		renderConfig.EnumeratorConfig.Enumerator = list.ArabicEnumerator
		renderConfig.PreSpacingConfig.Enabled = true
		renderConfig.PreSpacingConfig.Spacing = spacingStr
		renderConfig.PostSpacingConfig.Enabled = true
		renderConfig.PostSpacingConfig.Spacing = spacingStr
		renderConfig.ComponentOrder = []core.ListComponentType{
			core.ListComponentPreSpacing,
			core.ListComponentCursor,
			core.ListComponentEnumerator,
			core.ListComponentContent,
			core.ListComponentPostSpacing,
		}
	}

	// Apply component widths
	renderConfig.CursorConfig.MaxWidth = app.cursorWidth
	renderConfig.EnumeratorConfig.MaxWidth = app.enumeratorWidth
	renderConfig.ContentConfig.MaxWidth = app.contentWidth
	renderConfig.ContentConfig.WrapText = false

	// Apply or clear background styling
	if app.backgroundsEnabled {
		// Style the cursor component
		renderConfig.CursorConfig.NormalBackground = lipgloss.NewStyle().Background(lipgloss.Color("#404040"))
		renderConfig.CursorConfig.SelectedBackground = lipgloss.NewStyle().Background(lipgloss.Color("#606060"))
		renderConfig.CursorConfig.ApplyNormalBg = true
		renderConfig.CursorConfig.ApplySelectedBg = true

		// Style the enumerator component
		renderConfig.EnumeratorConfig.Style = lipgloss.NewStyle().
			Background(lipgloss.Color("#2D5016")).
			Foreground(lipgloss.Color("#FFFFFF"))

		// Style the content component
		renderConfig.ContentConfig.NormalBackground = lipgloss.NewStyle().Background(lipgloss.Color("#1A1A2E"))
		renderConfig.ContentConfig.SelectedBackground = lipgloss.NewStyle().Background(lipgloss.Color("#2A2A4E"))
		renderConfig.ContentConfig.ApplyNormalBg = true
		renderConfig.ContentConfig.ApplySelectedBg = true
	} else {
		// Clear all background styling
		renderConfig.CursorConfig.NormalBackground = lipgloss.NewStyle()
		renderConfig.CursorConfig.SelectedBackground = lipgloss.NewStyle()
		renderConfig.CursorConfig.ApplyNormalBg = false
		renderConfig.CursorConfig.ApplySelectedBg = false

		// Reset enumerator to default style (no background)
		renderConfig.EnumeratorConfig.Style = lipgloss.NewStyle()

		// Clear content backgrounds
		renderConfig.ContentConfig.NormalBackground = lipgloss.NewStyle()
		renderConfig.ContentConfig.SelectedBackground = lipgloss.NewStyle()
		renderConfig.ContentConfig.ApplyNormalBg = false
		renderConfig.ContentConfig.ApplySelectedBg = false
	}

	app.list.SetRenderConfig(renderConfig)
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

	layoutStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00CED1")).
		Bold(true)

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B35")).
		Bold(true)

	title := titleStyle.Render("🔧 VTable Component Rendering Demo")
	currentLayout := layoutStyle.Render(fmt.Sprintf("Layout %d/%d: %s",
		app.currentLayout+1, len(app.layoutNames), app.layoutNames[app.currentLayout]))

	// Status information
	var status strings.Builder
	status.WriteString(statusStyle.Render("Settings: "))
	if app.backgroundsEnabled {
		status.WriteString("Backgrounds ON • ")
	} else {
		status.WriteString("Backgrounds OFF • ")
	}

	// Show selected component and its current width
	status.WriteString(selectedStyle.Render(fmt.Sprintf("Adjusting: %s", app.selectedComponent.String())))
	status.WriteString(" • ")

	// Show all component widths
	switch app.selectedComponent {
	case ComponentSpacing:
		status.WriteString(selectedStyle.Render(fmt.Sprintf("Spacing: %d", app.spacingWidth)))
	case ComponentCursor:
		status.WriteString(selectedStyle.Render(fmt.Sprintf("Cursor: %d", app.cursorWidth)))
	case ComponentEnumerator:
		status.WriteString(selectedStyle.Render(fmt.Sprintf("Enumerator: %d", app.enumeratorWidth)))
	case ComponentContent:
		status.WriteString(selectedStyle.Render(fmt.Sprintf("Content: %d", app.contentWidth)))
	}

	help := helpStyle.Render("Navigate: ↑/↓ j/k • Select: Space • Layout: c • Component: w • Backgrounds: b • Width: +/- • Quit: q")

	return fmt.Sprintf("%s\n%s\n%s\n\n%s\n\n%s",
		title,
		currentLayout,
		status.String(),
		app.list.View(),
		help,
	)
}

func main() {
	// Create data source
	dataSource := NewPersonDataSource()

	// Configure list
	listConfig := config.DefaultListConfig()
	listConfig.ViewportConfig.Height = 5
	listConfig.MaxWidth = 600
	listConfig.SelectionMode = core.SelectionMultiple

	// Set formatter in config
	listConfig.RenderConfig.ContentConfig.Formatter = styledPersonFormatter

	// Create list
	vtableList := list.NewList(listConfig, dataSource)

	// Create application
	app := &App{
		list:               vtableList,
		currentLayout:      0,
		backgroundsEnabled: false,
		selectedComponent:  ComponentSpacing,
		// Initialize component widths
		spacingWidth:    2,
		cursorWidth:     2,
		enumeratorWidth: 4,
		contentWidth:    50,
		layoutNames: []string{
			"Default: [Cursor][Enumerator][Content]",
			"Numbers at End: [Cursor][Content][Enumerator]",
			"Content Only: [Content]",
			"With Spacing: [Spacing][Cursor][Enumerator][Content][Spacing]",
		},
	}

	// Set initial layout
	app.setLayout()

	// Run the application (removed WithAltScreen)
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
