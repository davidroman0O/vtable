package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/list"
)

// Person represents our data model
type Person struct {
	Name string
	Age  int
	City string
	Job  string
}

// Sample data for generation
var firstNames = []string{
	"Alice", "Bob", "Carol", "David", "Eve", "Frank", "Grace", "Henry", "Ivy", "Jack",
	"Kate", "Liam", "Mia", "Noah", "Olivia", "Paul", "Quinn", "Ruby", "Sam", "Tina",
	"Uma", "Victor", "Wendy", "Xavier", "Yara", "Zack", "Anna", "Ben", "Clara", "Dan",
}

var lastNames = []string{
	"Johnson", "Smith", "Williams", "Brown", "Davis", "Miller", "Wilson", "Taylor",
	"Anderson", "Thomas", "Jackson", "White", "Harris", "Martin", "Thompson", "Garcia",
	"Martinez", "Robinson", "Clark", "Rodriguez", "Lewis", "Lee", "Walker", "Hall",
	"Allen", "Young", "Hernandez", "King", "Wright", "Lopez", "Hill", "Scott",
}

var cities = []string{
	"New York", "San Francisco", "Chicago", "Boston", "Seattle", "Denver", "Austin",
	"Miami", "Portland", "Phoenix", "Los Angeles", "Dallas", "Atlanta", "Detroit",
	"Philadelphia", "Houston", "San Diego", "Las Vegas", "Orlando", "Nashville",
}

var jobs = []string{
	"Engineer", "Designer", "Manager", "Developer", "Analyst", "Architect", "Writer",
	"Sales", "Artist", "Consultant", "Teacher", "Doctor", "Lawyer", "Chef", "Nurse",
	"Scientist", "Photographer", "Musician", "Accountant", "Marketing", "HR", "Finance",
}

// GeneratedDataSource implements the pure DataSource interface with generated data
type GeneratedDataSource struct {
	totalPeople int          // Only data, no state!
	selected    map[int]bool // Selection state owned by DataSource
}

func NewGeneratedDataSource(total int) *GeneratedDataSource {
	return &GeneratedDataSource{
		totalPeople: total,
		selected:    make(map[int]bool),
	}
}

func (s *GeneratedDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		// Simulate loading delay asynchronously
		time.Sleep(10 * time.Millisecond)

		start := request.Start
		count := request.Count
		total := s.totalPeople

		if start >= total {
			// Return empty chunk if start is beyond data
			return core.DataChunkLoadedMsg{
				StartIndex: start,
				Items:      []core.Data[any]{},
				Request:    request,
			}
		}

		end := start + count
		if end > total {
			end = total
		}

		var chunkItems []core.Data[any]
		for i := start; i < end; i++ {
			person := s.generatePerson(i)
			chunkItems = append(chunkItems, core.Data[any]{
				ID:       fmt.Sprintf("person-%d", i),
				Item:     person,
				Selected: s.selected[i], // Include current selection state
			})
		}

		return core.DataChunkLoadedMsg{
			StartIndex: start,
			Items:      chunkItems,
			Request:    request,
		}
	}
}

// LoadChunkImmediate loads a chunk synchronously - FULLY AUTOMATED!
func (s *GeneratedDataSource) LoadChunkImmediate(request core.DataRequest) core.DataChunkLoadedMsg {
	start := request.Start
	count := request.Count
	total := s.totalPeople

	if start >= total {
		// Return empty chunk if start is beyond data
		return core.DataChunkLoadedMsg{
			StartIndex: start,
			Items:      []core.Data[any]{},
			Request:    request,
		}
	}

	end := start + count
	if end > total {
		end = total
	}

	var chunkItems []core.Data[any]
	for i := start; i < end; i++ {
		person := s.generatePerson(i)
		chunkItems = append(chunkItems, core.Data[any]{
			ID:       fmt.Sprintf("person-%d", i),
			Item:     person,
			Selected: s.selected[i], // Include current selection state
		})
	}

	return core.DataChunkLoadedMsg{
		StartIndex: start,
		Items:      chunkItems,
		Request:    request,
	}
}

func (s *GeneratedDataSource) generatePerson(index int) Person {
	// Use index as seed for consistent generation
	firstName := firstNames[index%len(firstNames)]
	lastName := lastNames[(index*7)%len(lastNames)] // Use different multiplier for variety
	fullName := firstName + " " + lastName

	age := 22 + (index*3)%43 // Ages from 22 to 64
	city := cities[(index*11)%len(cities)]
	job := jobs[(index*13)%len(jobs)]

	return Person{
		Name: fullName,
		Age:  age,
		City: city,
		Job:  job,
	}
}

func (s *GeneratedDataSource) GetTotal() tea.Cmd {
	return core.DataTotalCmd(s.totalPeople)
}

func (s *GeneratedDataSource) RefreshTotal() tea.Cmd {
	return s.GetTotal()
}

func (s *GeneratedDataSource) GetItemID(item any) string {
	if person, ok := item.(Person); ok {
		return fmt.Sprintf("person-%s-%d", person.Name, person.Age)
	}
	return fmt.Sprintf("%v", item)
}

// Selection operations - DataSource handles selection state
func (s *GeneratedDataSource) SetSelected(index int, selected bool) tea.Cmd {
	// Update internal selection state
	if index >= 0 && index < s.totalPeople {
		if selected {
			s.selected[index] = true
		} else {
			delete(s.selected, index)
		}
	}
	return core.SelectionResponseCmd(true, index, fmt.Sprintf("person-%d", index), selected, "toggle", nil, nil)
}

func (s *GeneratedDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	// Extract index from ID (person-123 format)
	if strings.HasPrefix(id, "person-") {
		if index, err := strconv.Atoi(strings.TrimPrefix(id, "person-")); err == nil {
			return s.SetSelected(index, selected)
		}
	}
	return core.SelectionResponseCmd(false, -1, id, selected, "toggleByID", fmt.Errorf("invalid ID format"), nil)
}

func (s *GeneratedDataSource) SelectAll() tea.Cmd {
	// Select all items in the data source
	for i := 0; i < s.totalPeople; i++ {
		s.selected[i] = true
	}
	affectedIDs := make([]string, s.totalPeople)
	for i := 0; i < s.totalPeople; i++ {
		affectedIDs[i] = fmt.Sprintf("person-%d", i)
	}
	return core.SelectionResponseCmd(true, -1, "", true, "selectAll", nil, affectedIDs)
}

func (s *GeneratedDataSource) ClearSelection() tea.Cmd {
	// Clear all selections in the data source
	s.selected = make(map[int]bool)
	return core.SelectionResponseCmd(true, -1, "", false, "clear", nil, nil)
}

func (s *GeneratedDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	// Select a range of items
	affectedIDs := make([]string, 0, endIndex-startIndex+1)
	for i := startIndex; i <= endIndex && i < s.totalPeople; i++ {
		if i >= 0 {
			s.selected[i] = true
			affectedIDs = append(affectedIDs, fmt.Sprintf("person-%d", i))
		}
	}
	return core.SelectionResponseCmd(true, -1, "", true, "range", nil, affectedIDs)
}

// AppModel wraps our list for the Tea application and manages ALL STATE
type AppModel struct {
	list          *list.List
	dataSource    *GeneratedDataSource
	loadingChunks map[int]bool
	chunkHistory  []string
	showDebug     bool
	showHelp      bool
	statusMessage string
	indexInput    string
	inputMode     bool // true when entering a number for JumpToIndex
}

func main() {
	// Create generated data source with 100 people
	dataSource := NewGeneratedDataSource(100)

	// Create list configuration using proper defaults
	config := config.DefaultListConfig()
	config.ViewportConfig.Height = 8
	config.ViewportConfig.TopThreshold = 2
	config.ViewportConfig.BottomThreshold = 2
	config.ViewportConfig.ChunkSize = 5
	config.ViewportConfig.InitialIndex = 0
	config.ViewportConfig.BoundingAreaBefore = 4
	config.ViewportConfig.BoundingAreaAfter = 4
	config.SelectionMode = core.SelectionMultiple
	config.MaxWidth = 85

	// Set the formatter in the config
	config.RenderConfig.ContentConfig.Formatter = personFormatter

	// Set up basic cursor indicators so we can see the list!
	// The default enumerator returns empty string, so nothing shows up without this
	config.RenderConfig.CursorConfig.Enabled = true
	config.RenderConfig.CursorConfig.CursorIndicator = "► "
	config.RenderConfig.CursorConfig.NormalSpacing = "  "

	// Create app model
	app := AppModel{
		list:          list.NewList(config, dataSource),
		dataSource:    dataSource,
		loadingChunks: make(map[int]bool),
		chunkHistory:  make([]string, 0),
		showDebug:     true, // Show debug by default all the time
		showHelp:      true, // Start with help visible
		statusMessage: "Welcome! Use arrow keys to navigate, space to select, ? to toggle help",
		indexInput:    "",
		inputMode:     false,
	}

	// Run the program
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (m AppModel) Init() tea.Cmd {
	// just init and focus!
	return tea.Batch(
		m.list.Init(), // Automatically calls GetTotal() and loads initial chunk
		m.list.Focus(),
	)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle input mode for JumpToIndex
		if m.inputMode {
			switch msg.String() {
			case "enter":
				// Parse the input and jump to index
				if index, err := strconv.Atoi(m.indexInput); err == nil && index >= 0 && index < 100 {
					m.inputMode = false
					m.indexInput = ""
					m.statusMessage = fmt.Sprintf("Jumping to index %d", index)
					return m, core.JumpToCmd(index)
				} else {
					m.statusMessage = "Invalid index! Please enter a number between 0-99"
					m.inputMode = false
					m.indexInput = ""
					return m, nil
				}
			case "escape":
				m.inputMode = false
				m.indexInput = ""
				m.statusMessage = "Jump cancelled"
				return m, nil
			case "backspace":
				if len(m.indexInput) > 0 {
					m.indexInput = m.indexInput[:len(m.indexInput)-1]
				}
				return m, nil
			default:
				// Only allow digits
				if len(msg.String()) == 1 && msg.String() >= "0" && msg.String() <= "9" {
					if len(m.indexInput) < 3 { // Limit to 3 digits (0-999)
						m.indexInput += msg.String()
					}
				}
				return m, nil
			}
		}

		// Normal key handling
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "r":
			// Force refresh to see loading again
			m.statusMessage = "Refreshing data..."
			return m, core.DataRefreshCmd()

		case "d":
			// Toggle debug display
			m.showDebug = !m.showDebug
			if m.showDebug {
				m.statusMessage = "Debug mode ON"
			} else {
				m.statusMessage = "Debug mode OFF"
			}
			return m, nil

		case "?":
			// Toggle help display
			m.showHelp = !m.showHelp
			if m.showHelp {
				m.statusMessage = "Help visible - press ? to hide"
			} else {
				m.statusMessage = "Help hidden - press ? to show"
			}
			return m, nil

		// === NAVIGATION KEYS ===
		case "g":
			// Jump to start (like vim)
			return m, core.JumpToStartCmd()

		case "G":
			// Jump to end (like vim)
			return m, core.JumpToEndCmd()

		case "J":
			// Enter jump-to-index mode (uppercase J)
			m.inputMode = true
			m.indexInput = ""
			m.statusMessage = "Enter index to jump to (0-99): "
			return m, nil

		case "h":
			// Page up using proper command
			return m, core.PageUpCmd()

		case "l":
			// Page down using proper command
			return m, core.PageDownCmd()

		case "j", "up":
			// Move up using proper command
			return m, core.CursorUpCmd()

		case "k", "down":
			// Move down using proper command
			return m, core.CursorDownCmd()

		// === SELECTION KEYS ===
		case " ":
			// Toggle current selection using proper Tea message
			return m, core.SelectCurrentCmd()

		case "a":
			// Select all using proper Tea message
			return m, core.SelectAllCmd()

		case "c":
			// Clear selection using proper Tea message
			return m, core.SelectClearCmd()

		case "s":
			// Show selection info
			selectionCount := m.list.GetSelectionCount()
			if selectionCount > 0 {
				m.statusMessage = fmt.Sprintf("SELECTED: %d items total (visible items show [✓] and ◄ SELECTED)", selectionCount)
			} else {
				m.statusMessage = "No items selected - use Space to select items"
			}
			return m, nil

		// === QUICK JUMP SHORTCUTS ===
		case "1":
			return m, core.JumpToCmd(10)

		case "2":
			return m, core.JumpToCmd(25)

		case "3":
			return m, core.JumpToCmd(50)

		case "4":
			return m, core.JumpToCmd(75)

		case "5":
			return m, core.JumpToCmd(90)

		default:
			// Let the list handle all other key presses (arrow keys, etc.)
			var cmd tea.Cmd
			_, cmd = m.list.Update(msg)

			// Update status with current position
			state := m.list.GetState()
			m.statusMessage = fmt.Sprintf("Position: %d/%d (Viewport: %d-%d)",
				state.CursorIndex+1, 100,
				state.ViewportStartIndex,
				state.ViewportStartIndex+7)

			return m, cmd
		}

	// Handle chunk loading observability messages
	case core.ChunkLoadingStartedMsg:
		m.loadingChunks[msg.ChunkStart] = true
		historyEntry := fmt.Sprintf("Started loading chunk %d (size: %d)", msg.ChunkStart, msg.Request.Count)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		// Also pass to list
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		return m, cmd

	case core.ChunkLoadingCompletedMsg:
		delete(m.loadingChunks, msg.ChunkStart)
		historyEntry := fmt.Sprintf("Completed chunk %d (%d items)", msg.ChunkStart, msg.ItemCount)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		// Also pass to list
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		return m, cmd

	case core.ChunkUnloadedMsg:
		historyEntry := fmt.Sprintf("Unloaded chunk %d", msg.ChunkStart)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		// Also pass to list
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		return m, cmd

	// Handle selection response messages
	case core.SelectionResponseMsg:
		// Update status based on selection operation
		switch msg.Operation {
		case "toggle":
			selectionCount := m.list.GetSelectionCount()
			state := m.list.GetState()
			if msg.Selected {
				m.statusMessage = fmt.Sprintf("Selected item at index %d - %d items selected total (look for [✓] and ◄ SELECTED)", state.CursorIndex, selectionCount)
			} else {
				m.statusMessage = fmt.Sprintf("Deselected item at index %d - %d items selected total", state.CursorIndex, selectionCount)
			}
		case "selectAll":
			selectionCount := m.list.GetSelectionCount()
			m.statusMessage = fmt.Sprintf("Selected ALL %d items in datasource (look for [✓] indicators!)", selectionCount)
		case "clear":
			m.statusMessage = "All selections cleared - [✓] indicators removed"
		}
		// Also pass to list
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		return m, cmd

	// Handle navigation messages to update status
	case core.PageUpMsg:
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		state := m.list.GetState()
		m.statusMessage = fmt.Sprintf("Page up - now at index %d", state.CursorIndex)
		return m, cmd

	case core.PageDownMsg:
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		state := m.list.GetState()
		m.statusMessage = fmt.Sprintf("Page down - now at index %d", state.CursorIndex)
		return m, cmd

	case core.JumpToMsg:
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		state := m.list.GetState()
		m.statusMessage = fmt.Sprintf("Jumped to index %d", state.CursorIndex)
		return m, cmd

	case core.JumpToStartMsg:
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		m.statusMessage = "Jumped to start"
		return m, cmd

	case core.JumpToEndMsg:
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		m.statusMessage = "Jumped to end"
		return m, cmd

	case core.CursorUpMsg, core.CursorDownMsg:
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		state := m.list.GetState()
		m.statusMessage = fmt.Sprintf("Position: %d/%d (Viewport: %d-%d)",
			state.CursorIndex+1, 100,
			state.ViewportStartIndex,
			state.ViewportStartIndex+7)
		return m, cmd

	default:
		// Pass all other messages to the list
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		return m, cmd
	}
}

func (m AppModel) View() string {
	var view strings.Builder

	// Show help if enabled
	if m.showHelp {
		view.WriteString(m.renderHelp())
		view.WriteString("\n")
	}

	// Show status message or input prompt
	if m.inputMode {
		view.WriteString(fmt.Sprintf("%s%s_", m.statusMessage, m.indexInput))
	} else {
		view.WriteString(m.statusMessage)
	}
	view.WriteString("\n\n")

	// Show main list content
	content := m.list.View()
	view.WriteString(content)

	// Show selection info
	selectionCount := m.list.GetSelectionCount()
	if selectionCount > 0 {
		view.WriteString(fmt.Sprintf("\n\nSelected: %d items", selectionCount))
	}

	// Show debug info if enabled
	if m.showDebug {
		view.WriteString("\n\n")
		view.WriteString(m.renderDebugInfo())
	}

	return view.String()
}

// renderHelp renders the help text
func (m AppModel) renderHelp() string {
	var help strings.Builder
	help.WriteString("=== NAVIGATION & SELECTION DEMO ===\n")
	help.WriteString("Visual Indicators: ► = cursor • [✓] = selected • ◄ SELECTED = selected item\n")
	help.WriteString("Navigation: j/k or ↑/↓ move • h/l page up/down • g=start • G=end • J=jump to index • 1-5=quick jumps\n")
	help.WriteString("Selection: Space=toggle • a=select ALL in datasource • c=clear • s=show selection\n")
	help.WriteString("Other: r=refresh • d=debug • ?=help • q=quit")
	return help.String()
}

// renderDebugInfo renders chunk loading debug information
func (m AppModel) renderDebugInfo() string {
	var debug strings.Builder
	debug.WriteString("=== CHUNK LOADING DEBUG ===\n")

	// Show viewport and bounding area details
	state := m.list.GetState()
	debug.WriteString(fmt.Sprintf("Viewport: start=%d, cursor=%d (viewport_idx=%d)\n",
		state.ViewportStartIndex, state.CursorIndex, state.CursorViewportIndex))

	// Show threshold flags
	debug.WriteString(fmt.Sprintf("Thresholds: top=%v, bottom=%v\n",
		state.IsAtTopThreshold, state.IsAtBottomThreshold))

	// Show currently loading chunks
	if len(m.loadingChunks) > 0 {
		debug.WriteString("Loading chunks: ")
		var chunks []string
		for chunk := range m.loadingChunks {
			chunks = append(chunks, fmt.Sprintf("%d", chunk))
		}
		debug.WriteString(strings.Join(chunks, ", ") + "\n")
	}

	// Show recent chunk history
	if len(m.chunkHistory) > 0 {
		debug.WriteString("Recent activity:\n")
		for _, entry := range m.chunkHistory {
			debug.WriteString("  " + entry + "\n")
		}
	}

	if len(m.loadingChunks) == 0 && len(m.chunkHistory) == 0 {
		debug.WriteString("No chunk activity yet\n")
	}

	return debug.String()
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// personFormatter formats a person item for display - CONTENT ONLY
// The component system handles cursor, selection indicators, and enumerators
func personFormatter(item core.Data[any], index int, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
	person, ok := item.Item.(Person)
	if !ok {
		return fmt.Sprintf("Invalid data: %v", item.Item)
	}

	// ONLY format the main content - no prefixes or suffixes
	// The component system will add cursor (►), enumerators (1. 2. •), and selection indicators
	content := fmt.Sprintf("%-20s | Age: %-3d | %-15s | %s",
		person.Name, person.Age, person.City, person.Job)

	// Add threshold indicators for demo purposes (this is content, not cursor indicators)
	if isCursor {
		if isTopThreshold {
			content += " [TOP]"
		} else if isBottomThreshold {
			content += " [BOT]"
		}
	}

	// Add selection indicator as part of content (not prefix)
	if item.Selected {
		content += " ◄ SELECTED"
	}

	return content
}
