package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable"
)

// Person represents a data record with various field types for formatting
type Person struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Age         int       `json:"age"`
	Salary      float64   `json:"salary"`
	Department  string    `json:"department"`
	Status      string    `json:"status"`
	JoinDate    time.Time `json:"join_date"`
	Performance float64   `json:"performance"` // 0.0 to 1.0
	Projects    int       `json:"projects"`
}

// PersonDataProvider provides person data with different formatting demonstrations
type PersonDataProvider struct {
	people    []Person
	selection map[int]bool
}

func NewPersonDataProvider() *PersonDataProvider {
	return &PersonDataProvider{
		people:    generatePeople(),
		selection: make(map[int]bool),
	}
}

func generatePeople() []Person {
	people := []Person{
		{ID: 1, Name: "Alice Johnson", Email: "alice@company.com", Age: 28, Salary: 75000, Department: "Engineering", Status: "Active", JoinDate: time.Date(2021, 3, 15, 0, 0, 0, 0, time.UTC), Performance: 0.92, Projects: 8},
		{ID: 2, Name: "Bob Smith", Email: "bob.smith@company.com", Age: 34, Salary: 95000, Department: "Engineering", Status: "Active", JoinDate: time.Date(2019, 7, 8, 0, 0, 0, 0, time.UTC), Performance: 0.88, Projects: 12},
		{ID: 3, Name: "Carol Davis", Email: "carol.d@company.com", Age: 29, Salary: 82000, Department: "Design", Status: "Active", JoinDate: time.Date(2020, 11, 22, 0, 0, 0, 0, time.UTC), Performance: 0.95, Projects: 6},
		{ID: 4, Name: "David Wilson", Email: "d.wilson@company.com", Age: 42, Salary: 110000, Department: "Management", Status: "Active", JoinDate: time.Date(2017, 1, 30, 0, 0, 0, 0, time.UTC), Performance: 0.78, Projects: 15},
		{ID: 5, Name: "Eve Brown", Email: "eve.brown@company.com", Age: 26, Salary: 68000, Department: "Marketing", Status: "Inactive", JoinDate: time.Date(2022, 5, 10, 0, 0, 0, 0, time.UTC), Performance: 0.85, Projects: 4},
		{ID: 6, Name: "Frank Miller", Email: "frank@company.com", Age: 31, Salary: 88000, Department: "Engineering", Status: "Active", JoinDate: time.Date(2020, 2, 14, 0, 0, 0, 0, time.UTC), Performance: 0.91, Projects: 9},
		{ID: 7, Name: "Grace Lee", Email: "grace.lee@company.com", Age: 27, Salary: 72000, Department: "Design", Status: "Active", JoinDate: time.Date(2021, 9, 5, 0, 0, 0, 0, time.UTC), Performance: 0.89, Projects: 7},
		{ID: 8, Name: "Henry Clark", Email: "h.clark@company.com", Age: 38, Salary: 105000, Department: "Management", Status: "Active", JoinDate: time.Date(2018, 12, 3, 0, 0, 0, 0, time.UTC), Performance: 0.82, Projects: 11},
		{ID: 9, Name: "Ivy Taylor", Email: "ivy.taylor@company.com", Age: 25, Salary: 65000, Department: "Marketing", Status: "Active", JoinDate: time.Date(2023, 1, 18, 0, 0, 0, 0, time.UTC), Performance: 0.87, Projects: 3},
		{ID: 10, Name: "Jack Anderson", Email: "jack.a@company.com", Age: 33, Salary: 92000, Department: "Engineering", Status: "Inactive", JoinDate: time.Date(2019, 10, 12, 0, 0, 0, 0, time.UTC), Performance: 0.76, Projects: 10},
	}
	return people
}

func (p *PersonDataProvider) GetTotal() int {
	return len(p.people)
}

func (p *PersonDataProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[Person], error) {
	start := request.Start
	count := request.Count

	if start >= len(p.people) {
		return []vtable.Data[Person]{}, nil
	}

	end := start + count
	if end > len(p.people) {
		end = len(p.people)
	}

	result := make([]vtable.Data[Person], end-start)
	for i := start; i < end; i++ {
		result[i-start] = vtable.Data[Person]{
			ID:       fmt.Sprintf("person-%d", p.people[i].ID),
			Item:     p.people[i],
			Selected: p.selection[i],
			Metadata: vtable.NewTypedMetadata(),
		}
	}

	return result, nil
}

// Implement required DataProvider methods
func (p *PersonDataProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionMultiple
}

func (p *PersonDataProvider) SetSelected(index int, selected bool) bool {
	if index < 0 || index >= len(p.people) {
		return false
	}
	if selected {
		p.selection[index] = true
	} else {
		delete(p.selection, index)
	}
	return true
}

func (p *PersonDataProvider) SelectAll() bool {
	for i := 0; i < len(p.people); i++ {
		p.selection[i] = true
	}
	return true
}

func (p *PersonDataProvider) ClearSelection() {
	p.selection = make(map[int]bool)
}

func (p *PersonDataProvider) GetSelectedIndices() []int {
	indices := make([]int, 0, len(p.selection))
	for idx := range p.selection {
		indices = append(indices, idx)
	}
	return indices
}

func (p *PersonDataProvider) GetItemID(item *Person) string {
	return fmt.Sprintf("%d", item.ID)
}

func (p *PersonDataProvider) GetSelectedIDs() []string {
	ids := make([]string, 0, len(p.selection))
	for idx := range p.selection {
		if idx < len(p.people) {
			ids = append(ids, fmt.Sprintf("%d", p.people[idx].ID))
		}
	}
	return ids
}

func (p *PersonDataProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	return true
}

func (p *PersonDataProvider) SelectRange(startID, endID string) bool {
	return true
}

// FormatterType represents different formatter styles
type FormatterType int

const (
	FormatterBasic FormatterType = iota
	FormatterDetailed
	FormatterCompact
	FormatterColorful
	FormatterPerformance
)

// Main application model
type CustomFormatterModel struct {
	personList         *vtable.TeaList[Person]
	provider           *PersonDataProvider
	currentFormat      FormatterType
	formatNames        []string
	formatDescriptions []string
}

func newCustomFormatterDemo() *CustomFormatterModel {
	provider := NewPersonDataProvider()

	// Configure viewport - consistent with other examples
	viewportConfig := vtable.ViewportConfig{
		Height:               12,
		TopThresholdIndex:    2,
		BottomThresholdIndex: 9,
		ChunkSize:            20,
		InitialIndex:         0,
		Debug:                false,
	}

	// Create style config
	styleConfig := vtable.StyleConfig{
		BorderStyle:      "245",
		HeaderStyle:      "bold 252 on 238",
		RowStyle:         "252",
		SelectedRowStyle: "bold 252 on 63",
	}

	// Start with basic formatter
	basicFormatter := createBasicFormatter()

	// Create the list
	list, err := vtable.NewTeaList(viewportConfig, provider, styleConfig, basicFormatter)
	if err != nil {
		log.Fatal(err)
	}

	return &CustomFormatterModel{
		personList:    list,
		provider:      provider,
		currentFormat: FormatterBasic,
		formatNames: []string{
			"Basic",
			"Detailed",
			"Compact",
			"Colorful",
			"Performance",
		},
		formatDescriptions: []string{
			"Simple name and email display",
			"Full information with labels",
			"Minimal space-efficient layout",
			"Color-coded by department and status",
			"Performance metrics with visual indicators",
		},
	}
}

// Basic formatter - simple name and email
func createBasicFormatter() func(vtable.Data[Person], int, vtable.RenderContext, bool, bool, bool) string {
	return func(data vtable.Data[Person], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		person := data.Item

		// ASCII-only selection indicator (exactly 2 chars)
		prefix := "  "
		if data.Selected && isCursor {
			prefix = "*>"
		} else if data.Selected {
			prefix = "* "
		} else if isCursor {
			prefix = "> "
		}

		return fmt.Sprintf("%s%s <%s>", prefix, person.Name, person.Email)
	}
}

// Detailed formatter - comprehensive information with labels
func createDetailedFormatter() func(vtable.Data[Person], int, vtable.RenderContext, bool, bool, bool) string {
	return func(data vtable.Data[Person], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		person := data.Item

		// ASCII-only selection indicator (exactly 2 chars)
		prefix := "  "
		if data.Selected && isCursor {
			prefix = "#>"
		} else if data.Selected {
			prefix = "# "
		} else if isCursor {
			prefix = "> "
		}

		// Format salary with commas
		salary := formatCurrency(person.Salary)

		// Calculate years since joining
		years := time.Since(person.JoinDate).Hours() / (24 * 365.25)

		return fmt.Sprintf("%s%s | Age: %d | %s | %s | Salary: %s | Years: %.1f | Projects: %d",
			prefix,
			person.Name,
			person.Age,
			person.Department,
			person.Status,
			salary,
			years,
			person.Projects,
		)
	}
}

// Compact formatter - minimal space usage
func createCompactFormatter() func(vtable.Data[Person], int, vtable.RenderContext, bool, bool, bool) string {
	return func(data vtable.Data[Person], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		person := data.Item

		// ASCII-only selection indicator (exactly 2 chars)
		prefix := "  "
		if data.Selected && isCursor {
			prefix = "o>"
		} else if data.Selected {
			prefix = "o "
		} else if isCursor {
			prefix = "> "
		}

		// Truncate name if too long
		name := person.Name
		if len(name) > 15 {
			name = name[:12] + "..."
		}

		// Department abbreviation
		dept := abbreviateDepartment(person.Department)

		// Status indicator (different from selection)
		status := "+"
		if person.Status == "Inactive" {
			status = "-"
		}

		return fmt.Sprintf("%s%-15s %s %2d %s $%dK P%d",
			prefix,
			name,
			status,
			person.Age,
			dept,
			int(person.Salary/1000),
			person.Projects,
		)
	}
}

// Colorful formatter - color-coded by department and status
func createColorfulFormatter() func(vtable.Data[Person], int, vtable.RenderContext, bool, bool, bool) string {
	return func(data vtable.Data[Person], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		person := data.Item

		// ASCII-only selection indicator (exactly 2 chars) - but colored
		prefix := "  "
		if data.Selected && isCursor {
			prefix = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render("*>")
		} else if data.Selected {
			prefix = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("* ")
		} else if isCursor {
			prefix = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render("> ")
		}

		// Color-code by department
		nameStyle := getDepartmentStyle(person.Department)

		// Status styling
		statusStyle := getStatusStyle(person.Status)

		// Salary styling by range
		salaryStyle := getSalaryStyle(person.Salary)

		coloredName := nameStyle.Render(person.Name)
		coloredStatus := statusStyle.Render(person.Status)
		coloredSalary := salaryStyle.Render(formatCurrency(person.Salary))

		return fmt.Sprintf("%s%s | %s | %s | %s | Age %d",
			prefix,
			coloredName,
			person.Department,
			coloredStatus,
			coloredSalary,
			person.Age,
		)
	}
}

// Performance formatter - focus on performance metrics with visual indicators
func createPerformanceFormatter() func(vtable.Data[Person], int, vtable.RenderContext, bool, bool, bool) string {
	return func(data vtable.Data[Person], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		person := data.Item

		// ASCII-only selection indicator (exactly 2 chars)
		prefix := "  "
		if data.Selected && isCursor {
			prefix = "^>"
		} else if data.Selected {
			prefix = "^ "
		} else if isCursor {
			prefix = "> "
		}

		// Performance bar (10 characters wide)
		perfBar := createPerformanceBar(person.Performance)

		// Performance rating
		rating := getPerformanceRating(person.Performance)

		// Projects indicator
		projectsIndicator := createProjectsIndicator(person.Projects)

		// Tenure indicator
		years := time.Since(person.JoinDate).Hours() / (24 * 365.25)
		tenureIndicator := createTenureIndicator(years)

		return fmt.Sprintf("%s%-20s %s %s %s %s P:%d Y:%.1f",
			prefix,
			person.Name,
			perfBar,
			rating,
			projectsIndicator,
			tenureIndicator,
			person.Projects,
			years,
		)
	}
}

// Helper functions for formatting

func formatCurrency(amount float64) string {
	return fmt.Sprintf("$%.0f", amount)
}

func abbreviateDepartment(dept string) string {
	switch dept {
	case "Engineering":
		return "ENG"
	case "Design":
		return "DES"
	case "Marketing":
		return "MKT"
	case "Management":
		return "MGT"
	default:
		return "OTH"
	}
}

func getDepartmentStyle(dept string) lipgloss.Style {
	switch dept {
	case "Engineering":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("39")) // Blue
	case "Design":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("213")) // Pink
	case "Marketing":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("46")) // Green
	case "Management":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Orange
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // Default
	}
}

func getStatusStyle(status string) lipgloss.Style {
	switch status {
	case "Active":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true) // Green
	case "Inactive":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
	}
}

func getSalaryStyle(salary float64) lipgloss.Style {
	if salary >= 100000 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true) // Green - high
	} else if salary >= 80000 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow - medium
	} else {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray - entry level
	}
}

func createPerformanceBar(performance float64) string {
	barWidth := 10
	filled := int(performance * float64(barWidth))

	bar := ""
	for i := 0; i < barWidth; i++ {
		if i < filled {
			if performance >= 0.9 {
				bar += "█" // Solid for excellent
			} else if performance >= 0.8 {
				bar += "▓" // Medium for good
			} else {
				bar += "▒" // Light for fair
			}
		} else {
			bar += "░" // Empty
		}
	}

	// Color the bar based on performance
	if performance >= 0.9 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render(bar) // Green
	} else if performance >= 0.8 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(bar) // Yellow
	} else {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(bar) // Red
	}
}

func getPerformanceRating(performance float64) string {
	if performance >= 0.95 {
		return "⭐⭐⭐" // Excellent
	} else if performance >= 0.9 {
		return "⭐⭐" // Very Good
	} else if performance >= 0.8 {
		return "⭐" // Good
	} else {
		return "📈" // Needs Improvement
	}
}

func createProjectsIndicator(projects int) string {
	if projects >= 15 {
		return "🔥" // High
	} else if projects >= 10 {
		return "💪" // Medium-High
	} else if projects >= 5 {
		return "👍" // Medium
	} else {
		return "🌱" // New/Low
	}
}

func createTenureIndicator(years float64) string {
	if years >= 5 {
		return "🏆" // Veteran
	} else if years >= 3 {
		return "🎯" // Experienced
	} else if years >= 1 {
		return "🚀" // Growing
	} else {
		return "🆕" // New
	}
}

func (m *CustomFormatterModel) Init() tea.Cmd {
	return m.personList.Init()
}

func (m *CustomFormatterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "tab", "right", "l":
			// Cycle to next formatter
			m.currentFormat = FormatterType((int(m.currentFormat) + 1) % len(m.formatNames))
			m.updateFormatter()
			return m, nil

		case "left", "h":
			// Cycle to previous formatter
			current := int(m.currentFormat)
			if current == 0 {
				current = len(m.formatNames)
			}
			m.currentFormat = FormatterType(current - 1)
			m.updateFormatter()
			return m, nil

		case " ", "space":
			// Toggle selection
			m.personList.ToggleCurrentSelection()
			return m, nil

		// Direct formatter selection
		case "1":
			m.currentFormat = FormatterBasic
			m.updateFormatter()
			return m, nil
		case "2":
			m.currentFormat = FormatterDetailed
			m.updateFormatter()
			return m, nil
		case "3":
			m.currentFormat = FormatterCompact
			m.updateFormatter()
			return m, nil
		case "4":
			m.currentFormat = FormatterColorful
			m.updateFormatter()
			return m, nil
		case "5":
			m.currentFormat = FormatterPerformance
			m.updateFormatter()
			return m, nil
		}
	}

	// Update the list
	updatedList, cmd := m.personList.Update(msg)
	m.personList = updatedList.(*vtable.TeaList[Person])
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *CustomFormatterModel) updateFormatter() {
	// Clear any existing animated formatter
	m.personList.ClearAnimatedFormatter()

	// Apply the new formatter
	switch m.currentFormat {
	case FormatterBasic:
		m.personList.SetFormatter(createBasicFormatter())
	case FormatterDetailed:
		m.personList.SetFormatter(createDetailedFormatter())
	case FormatterCompact:
		m.personList.SetFormatter(createCompactFormatter())
	case FormatterColorful:
		m.personList.SetFormatter(createColorfulFormatter())
	case FormatterPerformance:
		m.personList.SetFormatter(createPerformanceFormatter())
	}
}

func (m *CustomFormatterModel) View() string {
	var sb strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Render("VTable Example 09: Custom Formatters")

	sb.WriteString(title + "\n\n")

	// Current formatter info
	currentStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("15")).
		Padding(0, 1)

	sb.WriteString(fmt.Sprintf("Current Formatter: %s - %s\n",
		currentStyle.Render(m.formatNames[m.currentFormat]),
		m.formatDescriptions[m.currentFormat],
	))

	// Formatter selector
	sb.WriteString("Available Formatters: ")
	for i, name := range m.formatNames {
		if FormatterType(i) == m.currentFormat {
			style := lipgloss.NewStyle().
				Background(lipgloss.Color("39")).
				Foreground(lipgloss.Color("15")).
				Padding(0, 1)
			sb.WriteString(style.Render(fmt.Sprintf("%d:%s", i+1, name)))
		} else {
			sb.WriteString(fmt.Sprintf("%d:%s", i+1, name))
		}
		if i < len(m.formatNames)-1 {
			sb.WriteString(" | ")
		}
	}
	sb.WriteString("\n\n")

	// The main list
	sb.WriteString(m.personList.View())

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		MarginTop(1)

	help := helpStyle.Render(
		"Controls: ↑/↓ navigate • SPACE select • TAB/→/← cycle formatters • 1-5 direct select • q quit\n" +
			"Formatters: Basic=simple • Detailed=checkboxes • Compact=circles • Colorful=colored • Performance=stars\n" +
			"Features: Custom selection indicators • Color coding • Visual indicators • Progress bars • Icons")

	sb.WriteString("\n" + help)

	return sb.String()
}

func main() {
	model := newCustomFormatterDemo()

	// Configure the Tea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	// Clean exit
	fmt.Print("\033[H\033[2J")
	fmt.Print("\033[?25h")
	fmt.Print("\n\n")
}
