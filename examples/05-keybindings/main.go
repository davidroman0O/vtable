package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman0O/vtable"
)

// Application states
type AppState int

const (
	StateMenu AppState = iota
	StateListDemo
	StateTableDemo
)

// Custom message to go back to menu
type BackToMenuMsg struct{}

// Main application model that manages different states
type AppModel struct {
	state      AppState
	menuModel  *MenuModel
	listModel  *ListKeybindingModel
	tableModel *TableKeybindingModel
}

func newAppModel() *AppModel {
	return &AppModel{
		state:     StateMenu,
		menuModel: newMenuModel(),
	}
}

func (m *AppModel) Init() tea.Cmd {
	return m.menuModel.Init()
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case BackToMenuMsg:
		// Go back to menu
		m.state = StateMenu
		m.menuModel = newMenuModel()
		return m, m.menuModel.Init()
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.state {
	case StateMenu:
		newMenuModel, cmd := m.menuModel.Update(msg)
		m.menuModel = newMenuModel.(*MenuModel)

		// Check if user selected something
		if m.menuModel.selected != -1 {
			switch m.menuModel.selected {
			case 0: // List Keybinding Demo
				m.state = StateListDemo
				m.listModel = newListKeybindingModel()
				return m, m.listModel.Init()
			case 1: // Table Keybinding Demo
				m.state = StateTableDemo
				m.tableModel = newTableKeybindingModel()
				return m, m.tableModel.Init()
			}
		}
		return m, cmd

	case StateListDemo:
		newListModel, cmd := m.listModel.Update(msg)
		m.listModel = newListModel.(*ListKeybindingModel)
		return m, cmd

	case StateTableDemo:
		newTableModel, cmd := m.tableModel.Update(msg)
		m.tableModel = newTableModel.(*TableKeybindingModel)
		return m, cmd
	}

	return m, nil
}

func (m *AppModel) View() string {
	switch m.state {
	case StateMenu:
		return m.menuModel.View()
	case StateListDemo:
		return m.listModel.View()
	case StateTableDemo:
		return m.tableModel.View()
	default:
		return "Unknown state"
	}
}

// Menu model for choosing between list and table
type MenuModel struct {
	choices  []string
	cursor   int
	selected int
}

func newMenuModel() *MenuModel {
	return &MenuModel{
		choices: []string{
			"List Demo - Vim hjkl navigation (h/l jump 5 items, j/k move 1)",
			"Table Demo - Vim hjkl navigation (h/l columns, j/k rows)",
		},
		cursor:   0,
		selected: -1,
	}
}

func (m *MenuModel) Init() tea.Cmd {
	return nil
}

func (m *MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = m.cursor
		}
	}
	return m, nil
}

func (m *MenuModel) View() string {
	s := "VTable Example 05: Basic Vim Navigation\n\n"
	s += "Simple hjkl keybindings for navigation!\n\n"
	s += "Choose a demo to run:\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\nPress j/k or ↑/↓ to navigate, Enter to select, q to quit.\n"
	return s
}

// ===== LIST DEMO =====

type Note struct {
	ID      int
	Title   string
	Content string
	Tags    []string
	Created time.Time
}

// Note provider for list demo
type NoteProvider struct {
	notes []Note
}

func NewNoteProvider() *NoteProvider {
	return &NoteProvider{
		notes: []Note{
			{1, "Meeting Notes", "Discuss project timeline and milestones", []string{"work", "meeting"}, time.Now().AddDate(0, 0, -2)},
			{2, "Shopping List", "Milk, bread, eggs, coffee, fruits", []string{"personal", "shopping"}, time.Now().AddDate(0, 0, -1)},
			{3, "Book Ideas", "Science fiction novel about time travel", []string{"creative", "writing"}, time.Now().AddDate(0, 0, -3)},
			{4, "Code Review", "Check the new authentication module", []string{"work", "code"}, time.Now().AddDate(0, 0, -1)},
			{5, "Vacation Plan", "Visit Japan in spring for cherry blossoms", []string{"personal", "travel"}, time.Now().AddDate(0, 0, -5)},
			{6, "Learning Goals", "Master Go programming and system design", []string{"education", "programming"}, time.Now().AddDate(0, 0, -4)},
			{7, "Recipe Ideas", "Try making homemade pasta and sauce", []string{"cooking", "recipes"}, time.Now().AddDate(0, 0, -2)},
			{8, "Bug Report", "Login form validation not working properly", []string{"work", "bug"}, time.Now().AddDate(0, 0, -1)},
			{9, "Gift Ideas", "Birthday present for mom - jewelry or books", []string{"personal", "gifts"}, time.Now().AddDate(0, 0, -3)},
			{10, "Health Goals", "Exercise 3x week, eat more vegetables", []string{"health", "fitness"}, time.Now().AddDate(0, 0, -6)},
			{11, "Research Paper", "Machine learning applications in healthcare", []string{"work", "research"}, time.Now().AddDate(0, 0, -7)},
			{12, "Movie List", "Must watch sci-fi movies this year", []string{"entertainment", "movies"}, time.Now().AddDate(0, 0, -2)},
			{13, "Budget Planning", "Monthly expenses and savings goals", []string{"personal", "finance"}, time.Now().AddDate(0, 0, -4)},
			{14, "Exercise Routine", "Morning workout plan and schedule", []string{"health", "fitness"}, time.Now().AddDate(0, 0, -1)},
			{15, "Project Ideas", "Side projects to work on weekends", []string{"programming", "side-projects"}, time.Now().AddDate(0, 0, -8)},
			{16, "Travel Itinerary", "Europe trip destinations and activities", []string{"travel", "planning"}, time.Now().AddDate(0, 0, -10)},
			{17, "Book Reading List", "Technical books to read this quarter", []string{"education", "books"}, time.Now().AddDate(0, 0, -3)},
			{18, "Garden Planning", "Vegetables to plant in spring", []string{"gardening", "hobby"}, time.Now().AddDate(0, 0, -15)},
			{19, "Car Maintenance", "Oil change and tire rotation schedule", []string{"maintenance", "car"}, time.Now().AddDate(0, 0, -5)},
			{20, "Client Notes", "Meeting summary with potential client", []string{"work", "client"}, time.Now().AddDate(0, 0, -1)},
			{21, "Password List", "Update all security passwords", []string{"security", "personal"}, time.Now().AddDate(0, 0, -12)},
			{22, "Grocery Planning", "Weekly meal prep and shopping", []string{"food", "planning"}, time.Now().AddDate(0, 0, -1)},
			{23, "Course Notes", "Database design principles", []string{"education", "database"}, time.Now().AddDate(0, 0, -6)},
			{24, "Music Playlist", "Workout songs and motivation tracks", []string{"music", "fitness"}, time.Now().AddDate(0, 0, -4)},
			{25, "Home Repairs", "Kitchen faucet and bathroom tiles", []string{"home", "maintenance"}, time.Now().AddDate(0, 0, -9)},
			{26, "Investment Ideas", "Stock market research and analysis", []string{"finance", "investment"}, time.Now().AddDate(0, 0, -7)},
			{27, "Game Development", "Unity tutorial progress and notes", []string{"programming", "gamedev"}, time.Now().AddDate(0, 0, -11)},
			{28, "Language Learning", "Spanish vocabulary and grammar", []string{"education", "language"}, time.Now().AddDate(0, 0, -5)},
			{29, "Photography Tips", "Camera settings for landscape shots", []string{"photography", "hobby"}, time.Now().AddDate(0, 0, -8)},
			{30, "Networking Events", "Tech meetups and conferences", []string{"work", "networking"}, time.Now().AddDate(0, 0, -3)},
			{31, "Art Projects", "Digital painting techniques to practice", []string{"art", "creative"}, time.Now().AddDate(0, 0, -14)},
			{32, "Podcast List", "Technology and business podcasts", []string{"education", "podcasts"}, time.Now().AddDate(0, 0, -2)},
			{33, "Volunteer Work", "Local community service opportunities", []string{"community", "volunteer"}, time.Now().AddDate(0, 0, -16)},
			{34, "Pet Care", "Vet appointments and grooming schedule", []string{"pets", "care"}, time.Now().AddDate(0, 0, -6)},
			{35, "Gift Wrapping", "Supplies needed for holiday season", []string{"holidays", "gifts"}, time.Now().AddDate(0, 0, -20)},
			{36, "Skill Assessment", "Programming skills to improve", []string{"programming", "skills"}, time.Now().AddDate(0, 0, -4)},
			{37, "Daily Habits", "Morning routine optimization", []string{"personal", "habits"}, time.Now().AddDate(0, 0, -8)},
			{38, "Team Building", "Activities for remote team bonding", []string{"work", "team"}, time.Now().AddDate(0, 0, -5)},
			{39, "Hardware Upgrade", "Computer parts and specifications", []string{"tech", "hardware"}, time.Now().AddDate(0, 0, -12)},
			{40, "Meditation Practice", "Mindfulness exercises and techniques", []string{"wellness", "meditation"}, time.Now().AddDate(0, 0, -9)},
			{41, "Social Media", "Content calendar and posting schedule", []string{"marketing", "social"}, time.Now().AddDate(0, 0, -3)},
			{42, "Time Management", "Productivity tools and methods", []string{"productivity", "tools"}, time.Now().AddDate(0, 0, -7)},
			{43, "Backup Strategy", "Data backup and recovery plan", []string{"security", "backup"}, time.Now().AddDate(0, 0, -18)},
			{44, "Library Visits", "Books to borrow and return dates", []string{"books", "library"}, time.Now().AddDate(0, 0, -4)},
			{45, "Weather Tracking", "Local climate patterns and trends", []string{"weather", "data"}, time.Now().AddDate(0, 0, -11)},
			{46, "Career Goals", "Professional development objectives", []string{"career", "goals"}, time.Now().AddDate(0, 0, -13)},
			{47, "Appliance Manual", "Warranty info and user guides", []string{"home", "reference"}, time.Now().AddDate(0, 0, -25)},
			{48, "Emergency Contacts", "Important phone numbers list", []string{"emergency", "contacts"}, time.Now().AddDate(0, 0, -30)},
			{49, "Insurance Review", "Policy updates and coverage check", []string{"insurance", "finance"}, time.Now().AddDate(0, 0, -22)},
			{50, "Birthday Calendar", "Friends and family important dates", []string{"personal", "calendar"}, time.Now().AddDate(0, 0, -6)},
			{51, "Software Updates", "Applications needing updates", []string{"tech", "maintenance"}, time.Now().AddDate(0, 0, -3)},
			{52, "Conference Notes", "DevOps summit key takeaways", []string{"work", "conference"}, time.Now().AddDate(0, 0, -8)},
			{53, "Recipe Collection", "Family recipes and cooking notes", []string{"cooking", "family"}, time.Now().AddDate(0, 0, -45)},
			{54, "Plant Care", "Watering schedule and fertilizer notes", []string{"gardening", "plants"}, time.Now().AddDate(0, 0, -5)},
			{55, "YouTube Channels", "Educational and entertainment subscriptions", []string{"entertainment", "education"}, time.Now().AddDate(0, 0, -12)},
			{56, "Server Monitoring", "System alerts and performance logs", []string{"work", "devops"}, time.Now().AddDate(0, 0, -2)},
			{57, "Charity Donations", "Annual giving and tax deduction tracking", []string{"charity", "finance"}, time.Now().AddDate(0, 0, -90)},
			{58, "Sleep Tracking", "Sleep patterns and quality metrics", []string{"health", "sleep"}, time.Now().AddDate(0, 0, -7)},
			{59, "Code Snippets", "Useful programming code examples", []string{"programming", "reference"}, time.Now().AddDate(0, 0, -15)},
			{60, "Wine Tasting", "Notes from vineyard visits", []string{"hobby", "wine"}, time.Now().AddDate(0, 0, -28)},
			{61, "Furniture Ideas", "Home decoration and layout plans", []string{"home", "decoration"}, time.Now().AddDate(0, 0, -19)},
			{62, "API Documentation", "REST endpoint specifications", []string{"work", "api"}, time.Now().AddDate(0, 0, -4)},
			{63, "Hiking Trails", "Local outdoor recreation spots", []string{"outdoor", "hiking"}, time.Now().AddDate(0, 0, -21)},
			{64, "Email Templates", "Standard responses for common queries", []string{"work", "templates"}, time.Now().AddDate(0, 0, -33)},
			{65, "Tax Documents", "Important financial records location", []string{"finance", "taxes"}, time.Now().AddDate(0, 0, -120)},
			{66, "Browser Bookmarks", "Important websites and resources", []string{"reference", "web"}, time.Now().AddDate(0, 0, -14)},
			{67, "Fitness Goals", "Strength training progression tracking", []string{"fitness", "goals"}, time.Now().AddDate(0, 0, -9)},
			{68, "Cloud Storage", "File organization and sharing setup", []string{"tech", "storage"}, time.Now().AddDate(0, 0, -26)},
			{69, "Database Schema", "Table relationships and constraints", []string{"work", "database"}, time.Now().AddDate(0, 0, -6)},
			{70, "Meeting Agenda", "Weekly team sync topics", []string{"work", "meeting"}, time.Now().AddDate(0, 0, -1)},
		},
	}
}

func (p *NoteProvider) GetTotal() int {
	return len(p.notes)
}

func (p *NoteProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[string], error) {
	start := request.Start
	count := request.Count

	if start >= len(p.notes) {
		return []vtable.Data[string]{}, nil
	}

	if start+count > len(p.notes) {
		count = len(p.notes) - start
	}

	result := make([]vtable.Data[string], count)
	for i := 0; i < count; i++ {
		note := p.notes[start+i]
		tags := strings.Join(note.Tags, ", ")
		display := fmt.Sprintf("[%d] %s (%s)", note.ID, note.Title, tags)

		result[i] = vtable.Data[string]{
			ID:       fmt.Sprintf("note-%d", note.ID),
			Item:     display,
			Selected: false,
			Metadata: vtable.NewTypedMetadata(),
		}
	}
	return result, nil
}

// Selection methods for DataProvider interface
func (p *NoteProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionSingle
}

func (p *NoteProvider) SetSelected(index int, selected bool) bool {
	return true
}

func (p *NoteProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	return true
}

func (p *NoteProvider) SelectRange(startID, endID string) bool {
	return true
}

func (p *NoteProvider) SelectAll() bool {
	return true
}

func (p *NoteProvider) ClearSelection() {
}

func (p *NoteProvider) GetSelectedIndices() []int {
	return []int{}
}

func (p *NoteProvider) GetSelectedIDs() []string {
	return []string{}
}

func (p *NoteProvider) GetItemID(item *string) string {
	return ""
}

// List model with vim navigation
type ListKeybindingModel struct {
	list     *vtable.TeaList[string]
	provider *NoteProvider
	status   string
}

func newListKeybindingModel() *ListKeybindingModel {
	provider := NewNoteProvider()

	formatter := func(data vtable.Data[string], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return fmt.Sprintf("%s%s", prefix, data.Item)
	}

	list, err := vtable.NewTeaListWithHeight(provider, formatter, 12)
	if err != nil {
		panic(err)
	}

	return &ListKeybindingModel{
		list:     list,
		provider: provider,
		status:   "Vim navigation: h/l jump 5 items, j/k move 1 item, space to select, q to quit",
	}
}

func (m *ListKeybindingModel) Init() tea.Cmd {
	return m.list.Init()
}

func (m *ListKeybindingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }

		// Vim navigation
		case "h":
			// Jump up by 5 items
			for i := 0; i < 5; i++ {
				m.list.MoveUp()
			}
			m.status = "h - jumped up 5 items"
			return m, nil

		case "j":
			m.list.MoveDown()
			m.status = "j - moved down 1 item"
			return m, nil

		case "k":
			m.list.MoveUp()
			m.status = "k - moved up 1 item"
			return m, nil

		case "l":
			// Jump down by 5 items
			for i := 0; i < 5; i++ {
				m.list.MoveDown()
			}
			m.status = "l - jumped down 5 items"
			return m, nil

		case " ", "space":
			if m.list.ToggleCurrentSelection() {
				state := m.list.GetState()
				m.status = fmt.Sprintf("Toggled selection for item %d", state.CursorIndex)
			}
			return m, nil
		}
	}

	// Let the list handle other navigation
	newList, cmd := m.list.Update(msg)
	m.list = newList.(*vtable.TeaList[string])
	return m, cmd
}

func (m *ListKeybindingModel) View() string {
	var sb strings.Builder

	sb.WriteString("VTable Example 05: Vim Navigation - List Demo\n\n")

	// List
	sb.WriteString(m.list.View())
	sb.WriteString("\n\n")

	// Status
	sb.WriteString(m.status)
	sb.WriteString("\n\n")

	// Help
	sb.WriteString("Vim Keys: h=up5 j=down k=up l=down5 • space=select • q=quit")

	return sb.String()
}

// ===== TABLE DEMO =====

type Project struct {
	ID         int
	Name       string
	Status     string
	Priority   string
	Assignee   string
	Deadline   string
	Completion int
}

// Project provider for table demo
type ProjectProvider struct {
	projects []Project
}

func NewProjectProvider() *ProjectProvider {
	return &ProjectProvider{
		projects: []Project{
			{1, "Website Redesign", "In Progress", "High", "Alice", "2024-02-15", 75},
			{2, "Mobile App", "Planning", "Medium", "Bob", "2024-03-01", 15},
			{3, "Database Migration", "In Progress", "High", "Carol", "2024-01-30", 90},
			{4, "API Documentation", "On Hold", "Low", "David", "2024-04-01", 30},
			{5, "Security Audit", "Not Started", "High", "Eve", "2024-02-10", 0},
			{6, "Performance Testing", "In Progress", "Medium", "Frank", "2024-02-20", 60},
			{7, "User Training", "Completed", "Low", "Grace", "2024-01-15", 100},
			{8, "Code Review", "In Progress", "Medium", "Henry", "2024-02-05", 80},
			{9, "Deployment Pipeline", "Planning", "High", "Ivy", "2024-03-15", 25},
			{10, "Monitoring Setup", "Not Started", "Medium", "Jack", "2024-02-25", 0},
			{11, "Frontend Refactor", "In Progress", "High", "Alice", "2024-03-10", 45},
			{12, "Payment Gateway", "Planning", "High", "Bob", "2024-04-15", 10},
			{13, "User Analytics", "In Progress", "Medium", "Carol", "2024-02-28", 65},
			{14, "Cloud Migration", "On Hold", "High", "David", "2024-05-01", 20},
			{15, "Load Testing", "Not Started", "Medium", "Eve", "2024-03-20", 0},
			{16, "Auth System", "In Progress", "High", "Frank", "2024-02-12", 85},
			{17, "Admin Panel", "Completed", "Medium", "Grace", "2024-01-25", 100},
			{18, "Email Service", "In Progress", "Low", "Henry", "2024-03-05", 70},
			{19, "Search Feature", "Planning", "Medium", "Ivy", "2024-04-10", 5},
			{20, "Data Backup", "Not Started", "High", "Jack", "2024-03-30", 0},
			{21, "Cache Optimization", "In Progress", "Medium", "Alice", "2024-02-22", 55},
			{22, "Error Handling", "Planning", "Low", "Bob", "2024-04-20", 15},
			{23, "Logging System", "In Progress", "Medium", "Carol", "2024-02-18", 40},
			{24, "A/B Testing", "On Hold", "Low", "David", "2024-05-15", 25},
			{25, "Content Management", "Not Started", "Medium", "Eve", "2024-04-05", 0},
			{26, "Image Processing", "In Progress", "High", "Frank", "2024-02-26", 60},
			{27, "Notification System", "Completed", "Medium", "Grace", "2024-01-20", 100},
			{28, "Report Generation", "In Progress", "Low", "Henry", "2024-03-12", 35},
			{29, "File Upload", "Planning", "Medium", "Ivy", "2024-04-25", 8},
			{30, "Rate Limiting", "Not Started", "High", "Jack", "2024-03-25", 0},
			{31, "Social Login", "In Progress", "Medium", "Alice", "2024-02-29", 50},
			{32, "Multi-tenant", "Planning", "High", "Bob", "2024-05-10", 12},
			{33, "Chat Feature", "In Progress", "Low", "Carol", "2024-03-08", 75},
			{34, "Subscription Model", "On Hold", "Medium", "David", "2024-06-01", 30},
			{35, "Inventory System", "Not Started", "High", "Eve", "2024-04-18", 0},
			{36, "Order Processing", "In Progress", "High", "Frank", "2024-02-14", 80},
			{37, "Customer Support", "Completed", "Medium", "Grace", "2024-01-28", 100},
			{38, "Shipping Integration", "In Progress", "Medium", "Henry", "2024-03-14", 45},
			{39, "Tax Calculator", "Planning", "Low", "Ivy", "2024-05-05", 18},
			{40, "Fraud Detection", "Not Started", "High", "Jack", "2024-04-12", 0},
			{41, "Webhook System", "In Progress", "Medium", "Alice", "2024-02-24", 65},
			{42, "GraphQL API", "Planning", "Medium", "Bob", "2024-04-30", 22},
			{43, "Redis Cache", "In Progress", "High", "Carol", "2024-02-16", 90},
			{44, "Queue System", "On Hold", "Medium", "David", "2024-05-20", 35},
			{45, "Health Checks", "Not Started", "Low", "Eve", "2024-04-22", 0},
			{46, "Backup Restore", "In Progress", "High", "Frank", "2024-02-20", 55},
			{47, "Data Export", "Completed", "Low", "Grace", "2024-01-30", 100},
			{48, "Audit Logging", "In Progress", "Medium", "Henry", "2024-03-16", 40},
			{49, "Feature Flags", "Planning", "Medium", "Ivy", "2024-05-12", 28},
			{50, "Config Management", "Not Started", "High", "Jack", "2024-04-08", 0},
			{51, "Docker Setup", "In Progress", "Medium", "Alice", "2024-02-27", 70},
			{52, "K8s Deployment", "Planning", "High", "Bob", "2024-05-25", 15},
			{53, "CI/CD Pipeline", "In Progress", "High", "Carol", "2024-02-19", 85},
			{54, "Code Coverage", "On Hold", "Low", "David", "2024-06-10", 40},
			{55, "Linting Setup", "Not Started", "Medium", "Eve", "2024-04-28", 0},
			{56, "Documentation", "In Progress", "Low", "Frank", "2024-03-01", 60},
			{57, "Testing Framework", "Completed", "Medium", "Grace", "2024-02-01", 100},
			{58, "Performance Monitoring", "In Progress", "High", "Henry", "2024-03-18", 50},
			{59, "Security Scanning", "Planning", "High", "Ivy", "2024-05-18", 20},
			{60, "Compliance Check", "Not Started", "Medium", "Jack", "2024-04-15", 0},
			{61, "Disaster Recovery", "In Progress", "High", "Alice", "2024-02-25", 45},
			{62, "Encryption Setup", "Planning", "High", "Bob", "2024-06-05", 25},
			{63, "Access Control", "In Progress", "Medium", "Carol", "2024-02-21", 75},
			{64, "Session Management", "On Hold", "Medium", "David", "2024-05-30", 50},
			{65, "Password Policy", "Not Started", "Low", "Eve", "2024-05-08", 0},
			{66, "LDAP Integration", "In Progress", "Medium", "Frank", "2024-03-03", 35},
			{67, "SSO Implementation", "Completed", "High", "Grace", "2024-02-05", 100},
			{68, "Token Management", "In Progress", "Medium", "Henry", "2024-03-20", 65},
			{69, "CORS Setup", "Planning", "Low", "Ivy", "2024-05-22", 30},
			{70, "API Versioning", "Not Started", "Medium", "Jack", "2024-04-25", 0},
			{71, "Schema Validation", "In Progress", "Medium", "Alice", "2024-02-28", 55},
			{72, "Data Migration", "Planning", "High", "Bob", "2024-06-12", 18},
			{73, "Index Optimization", "In Progress", "High", "Carol", "2024-02-23", 80},
			{74, "Query Performance", "On Hold", "Medium", "David", "2024-06-15", 45},
			{75, "Connection Pooling", "Not Started", "High", "Eve", "2024-05-15", 0},
			{76, "Replication Setup", "In Progress", "High", "Frank", "2024-03-05", 40},
			{77, "Sharding Strategy", "Completed", "Medium", "Grace", "2024-02-08", 100},
			{78, "Transaction Handling", "In Progress", "Medium", "Henry", "2024-03-22", 70},
			{79, "Data Archiving", "Planning", "Low", "Ivy", "2024-06-20", 35},
			{80, "Cleanup Scripts", "Not Started", "Low", "Jack", "2024-05-28", 0},
		},
	}
}

func (p *ProjectProvider) GetTotal() int {
	return len(p.projects)
}

func (p *ProjectProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[vtable.TableRow], error) {
	start := request.Start
	count := request.Count

	if start >= len(p.projects) {
		return []vtable.Data[vtable.TableRow]{}, nil
	}

	if start+count > len(p.projects) {
		count = len(p.projects) - start
	}

	result := make([]vtable.Data[vtable.TableRow], count)
	for i := 0; i < count; i++ {
		project := p.projects[start+i]

		row := vtable.TableRow{
			Cells: []string{
				fmt.Sprintf("%d", project.ID),
				project.Name,
				project.Status,
				project.Priority,
				project.Assignee,
				project.Deadline,
				fmt.Sprintf("%d%%", project.Completion),
			},
		}

		result[i] = vtable.Data[vtable.TableRow]{
			ID:       fmt.Sprintf("project-%d", project.ID),
			Item:     row,
			Selected: false,
			Metadata: vtable.NewTypedMetadata(),
		}
	}
	return result, nil
}

// Selection methods for DataProvider interface
func (p *ProjectProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionMultiple
}

func (p *ProjectProvider) SetSelected(index int, selected bool) bool {
	return true
}

func (p *ProjectProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	return true
}

func (p *ProjectProvider) SelectRange(startID, endID string) bool {
	return true
}

func (p *ProjectProvider) SelectAll() bool {
	return true
}

func (p *ProjectProvider) ClearSelection() {
}

func (p *ProjectProvider) GetSelectedIndices() []int {
	return []int{}
}

func (p *ProjectProvider) GetSelectedIDs() []string {
	return []string{}
}

func (p *ProjectProvider) GetItemID(item *vtable.TableRow) string {
	return ""
}

// Table model with vim navigation
type TableKeybindingModel struct {
	table    *vtable.TeaTable
	provider *ProjectProvider
	status   string
}

func newTableKeybindingModel() *TableKeybindingModel {
	provider := NewProjectProvider()

	columns := []vtable.TableColumn{
		vtable.NewRightColumn("ID", 4),
		vtable.NewColumn("Project", 18),
		vtable.NewColumn("Status", 12),
		vtable.NewColumn("Priority", 8),
		vtable.NewColumn("Assignee", 10),
		vtable.NewColumn("Deadline", 12),
		vtable.NewRightColumn("Progress", 8),
	}

	table, err := vtable.NewTeaTableWithHeight(columns, provider, 12)
	if err != nil {
		panic(err)
	}

	return &TableKeybindingModel{
		table:    table,
		provider: provider,
		status:   "Vim navigation: h/l jump 5 rows, j/k move 1 row, space to select, q to quit",
	}
}

func (m *TableKeybindingModel) Init() tea.Cmd {
	return m.table.Init()
}

func (m *TableKeybindingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return BackToMenuMsg{} }

		// Vim navigation
		case "h":
			// Jump up by 5 rows
			for i := 0; i < 5; i++ {
				m.table.MoveUp()
			}
			m.status = "h - jumped up 5 rows"
			return m, nil

		case "j":
			m.table.MoveDown()
			m.status = "j - moved down 1 row"
			return m, nil

		case "k":
			m.table.MoveUp()
			m.status = "k - moved up 1 row"
			return m, nil

		case "l":
			// Jump down by 5 rows
			for i := 0; i < 5; i++ {
				m.table.MoveDown()
			}
			m.status = "l - jumped down 5 rows"
			return m, nil

		case " ", "space":
			if m.table.ToggleCurrentSelection() {
				state := m.table.GetState()
				count := m.table.GetSelectionCount()
				m.status = fmt.Sprintf("Toggled selection for row %d (total selected: %d)", state.CursorIndex, count)
			}
			return m, nil
		}
	}

	// Let the table handle other navigation
	newTable, cmd := m.table.Update(msg)
	m.table = newTable.(*vtable.TeaTable)
	return m, cmd
}

func (m *TableKeybindingModel) View() string {
	var sb strings.Builder

	sb.WriteString("VTable Example 05: Vim Navigation - Table Demo\n\n")

	// Table
	sb.WriteString(m.table.View())
	sb.WriteString("\n\n")

	// Status
	sb.WriteString(m.status)
	sb.WriteString("\n\n")

	// Help
	sb.WriteString("Vim Keys: h=up5 j=down k=up l=down5 • space=select • q=quit")

	return sb.String()
}

// ===== MAIN =====

func main() {
	app := newAppModel()

	p := tea.NewProgram(app)

	if _, err := p.Run(); err != nil {
		panic(err)
	}

	// Clean exit
	fmt.Print("\033[H\033[2J")
	fmt.Print("\033[?25h")
	fmt.Print("\n\n")
}
