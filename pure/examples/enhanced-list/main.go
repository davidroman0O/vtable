package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	vtable "github.com/davidroman0O/vtable/pure"
)

// Task represents our enhanced data model with more variety
type Task struct {
	ID          string
	Title       string
	Description string
	Priority    string
	Status      string
	Category    string
	DueDate     string
	Assignee    string
	Progress    int
}

// Sample data for generation with more variety
var taskTitles = []string{
	"Fix authentication bug", "Implement user dashboard", "Design new logo", "Write API documentation",
	"Optimize database queries", "Create unit tests", "Update dependencies", "Review pull requests",
	"Deploy to production", "Setup monitoring", "Refactor legacy code", "Add error handling",
	"Implement search feature", "Design mobile layout", "Configure CI/CD", "Update user guide",
	"Fix memory leak", "Add logging", "Implement caching", "Security audit",
	"Performance testing", "Code review", "Bug triage", "Feature planning",
	"Database migration", "API versioning", "Load testing", "Documentation update",
	"UI improvements", "Backend optimization", "Frontend refactoring", "Integration testing",
}

var priorities = []string{"Low", "Medium", "High", "Critical"}
var statuses = []string{"Todo", "In Progress", "Review", "Done", "Blocked"}
var categories = []string{"Frontend", "Backend", "DevOps", "Design", "Testing", "Documentation"}
var assignees = []string{
	"Alice Johnson", "Bob Smith", "Carol Davis", "David Wilson", "Eve Brown",
	"Frank Miller", "Grace Taylor", "Henry Clark", "Ivy Rodriguez", "Jack Lee",
}

// EnhancedDataSource implements the pure DataSource interface with rich task data
type EnhancedDataSource struct {
	totalTasks int          // Only data, no state!
	selected   map[int]bool // Selection state owned by DataSource
}

func NewEnhancedDataSource(total int) *EnhancedDataSource {
	return &EnhancedDataSource{
		totalTasks: total,
		selected:   make(map[int]bool),
	}
}

func (s *EnhancedDataSource) LoadChunk(request vtable.DataRequest) tea.Cmd {
	return func() tea.Msg {
		// Simulate loading delay asynchronously
		time.Sleep(15 * time.Millisecond)

		start := request.Start
		count := request.Count
		total := s.totalTasks

		if start >= total {
			return vtable.DataChunkLoadedMsg{
				StartIndex: start,
				Items:      []vtable.Data[any]{},
				Request:    request,
			}
		}

		end := start + count
		if end > total {
			end = total
		}

		var chunkItems []vtable.Data[any]
		for i := start; i < end; i++ {
			task := s.generateTask(i)

			// Add some variety to item states for demonstration
			var itemError error
			var loading bool
			var disabled bool

			// Every 15th item has an error
			if i%15 == 0 && i > 0 {
				itemError = fmt.Errorf("sync failed")
			}

			// Every 20th item is loading
			if i%20 == 0 && i > 0 {
				loading = true
			}

			// Every 25th item is disabled
			if i%25 == 0 && i > 0 {
				disabled = true
			}

			chunkItems = append(chunkItems, vtable.Data[any]{
				ID:       task.ID,
				Item:     task,
				Selected: s.selected[i],
				Error:    itemError,
				Loading:  loading,
				Disabled: disabled,
			})
		}

		return vtable.DataChunkLoadedMsg{
			StartIndex: start,
			Items:      chunkItems,
			Request:    request,
		}
	}
}

func (s *EnhancedDataSource) LoadChunkImmediate(request vtable.DataRequest) vtable.DataChunkLoadedMsg {
	start := request.Start
	count := request.Count
	total := s.totalTasks

	if start >= total {
		return vtable.DataChunkLoadedMsg{
			StartIndex: start,
			Items:      []vtable.Data[any]{},
			Request:    request,
		}
	}

	end := start + count
	if end > total {
		end = total
	}

	var chunkItems []vtable.Data[any]
	for i := start; i < end; i++ {
		task := s.generateTask(i)

		// Add some variety to item states
		var itemError error
		var loading bool
		var disabled bool

		if i%15 == 0 && i > 0 {
			itemError = fmt.Errorf("sync failed")
		}
		if i%20 == 0 && i > 0 {
			loading = true
		}
		if i%25 == 0 && i > 0 {
			disabled = true
		}

		chunkItems = append(chunkItems, vtable.Data[any]{
			ID:       task.ID,
			Item:     task,
			Selected: s.selected[i],
			Error:    itemError,
			Loading:  loading,
			Disabled: disabled,
		})
	}

	return vtable.DataChunkLoadedMsg{
		StartIndex: start,
		Items:      chunkItems,
		Request:    request,
	}
}

func (s *EnhancedDataSource) generateTask(index int) Task {
	title := taskTitles[index%len(taskTitles)]
	priority := priorities[index%len(priorities)]
	status := statuses[index%len(statuses)]
	category := categories[index%len(categories)]
	assignee := assignees[index%len(assignees)]

	// Generate varied descriptions
	descriptions := []string{
		"Quick fix needed for production issue",
		"Complex feature requiring multiple components and careful testing",
		"Simple task that should be completed quickly",
		"Important milestone with dependencies on other tasks",
		"Routine maintenance work",
	}
	description := descriptions[index%len(descriptions)]

	// Generate due dates
	dueDates := []string{
		"Today", "Tomorrow", "This Week", "Next Week", "Next Month", "No Due Date",
	}
	dueDate := dueDates[index%len(dueDates)]

	progress := (index * 7) % 101 // Progress from 0-100

	return Task{
		ID:          fmt.Sprintf("task-%d", index),
		Title:       title,
		Description: description,
		Priority:    priority,
		Status:      status,
		Category:    category,
		DueDate:     dueDate,
		Assignee:    assignee,
		Progress:    progress,
	}
}

func (s *EnhancedDataSource) GetTotal() tea.Cmd {
	return vtable.DataTotalCmd(s.totalTasks)
}

func (s *EnhancedDataSource) RefreshTotal() tea.Cmd {
	return s.GetTotal()
}

func (s *EnhancedDataSource) GetItemID(item any) string {
	if task, ok := item.(Task); ok {
		return task.ID
	}
	return fmt.Sprintf("%v", item)
}

// Selection operations
func (s *EnhancedDataSource) SetSelected(index int, selected bool) tea.Cmd {
	if index >= 0 && index < s.totalTasks {
		if selected {
			s.selected[index] = true
		} else {
			delete(s.selected, index)
		}
	}
	return vtable.SelectionResponseCmd(true, index, fmt.Sprintf("task-%d", index), selected, "toggle", nil, nil)
}

func (s *EnhancedDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	if strings.HasPrefix(id, "task-") {
		if index, err := strconv.Atoi(strings.TrimPrefix(id, "task-")); err == nil {
			return s.SetSelected(index, selected)
		}
	}
	return vtable.SelectionResponseCmd(false, -1, id, selected, "toggleByID", fmt.Errorf("invalid ID format"), nil)
}

func (s *EnhancedDataSource) SelectAll() tea.Cmd {
	for i := 0; i < s.totalTasks; i++ {
		s.selected[i] = true
	}
	affectedIDs := make([]string, s.totalTasks)
	for i := 0; i < s.totalTasks; i++ {
		affectedIDs[i] = fmt.Sprintf("task-%d", i)
	}
	return vtable.SelectionResponseCmd(true, -1, "", true, "selectAll", nil, affectedIDs)
}

func (s *EnhancedDataSource) ClearSelection() tea.Cmd {
	s.selected = make(map[int]bool)
	return vtable.SelectionResponseCmd(true, -1, "", false, "clear", nil, nil)
}

func (s *EnhancedDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	affectedIDs := make([]string, 0, endIndex-startIndex+1)
	for i := startIndex; i <= endIndex && i < s.totalTasks; i++ {
		if i >= 0 {
			s.selected[i] = true
			affectedIDs = append(affectedIDs, fmt.Sprintf("task-%d", i))
		}
	}
	return vtable.SelectionResponseCmd(true, -1, "", true, "range", nil, affectedIDs)
}

// EnhancedAppModel wraps our list with enhanced rendering features
type EnhancedAppModel struct {
	list            *vtable.List
	dataSource      *EnhancedDataSource
	loadingChunks   map[int]bool
	chunkHistory    []string
	showDebug       bool
	showHelp        bool
	statusMessage   string
	indexInput      string
	inputMode       bool
	currentStyle    int
	styleNames      []string
	showStyleInfo   bool
	customFormatter vtable.ItemFormatter[any]
}

func main() {
	// Create enhanced data source with 150 tasks
	dataSource := NewEnhancedDataSource(150)

	// Create list configuration with enhanced rendering
	config := vtable.ListConfig{
		ViewportConfig: vtable.ViewportConfig{
			Height:             10,
			TopThreshold:       2,
			BottomThreshold:    2,
			ChunkSize:          8,
			InitialIndex:       0,
			BoundingAreaBefore: 6,
			BoundingAreaAfter:  6,
		},
		SelectionMode: vtable.SelectionMultiple,
		KeyMap:        vtable.DefaultNavigationKeyMap(),
		MaxWidth:      100,
		// Enhanced rendering configuration
		RenderConfig: vtable.ListRenderConfig{
			Enumerator:      vtable.BulletEnumerator,
			ShowEnumerator:  true,
			IndentSize:      2,
			ItemSpacing:     0,
			MaxWidth:        90,
			WrapText:        true,
			AlignEnumerator: true,
		},
	}

	// Create list WITHOUT a formatter - let enhanced rendering handle everything
	list := vtable.NewList(config, dataSource)

	// Create app model
	app := EnhancedAppModel{
		list:          list,
		dataSource:    dataSource,
		loadingChunks: make(map[int]bool),
		chunkHistory:  make([]string, 0),
		showDebug:     true,
		showHelp:      true,
		statusMessage: "🎨 Enhanced List Demo! Press 'e' to cycle enumerator styles, '?' for help",
		indexInput:    "",
		inputMode:     false,
		currentStyle:  0,
		styleNames: []string{
			"Bullets", "Numbers", "Checkboxes", "Alphabetical", "Dashes",
			"Arrows", "Conditional", "Custom Pattern", "Roman Numerals", "Custom Formatter",
		},
		showStyleInfo: true,
	}

	// Run the program
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (m EnhancedAppModel) Init() tea.Cmd {
	return tea.Batch(
		m.list.Init(),
		m.list.Focus(),
	)
}

func (m EnhancedAppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle input mode for JumpToIndex
		if m.inputMode {
			switch msg.String() {
			case "enter":
				if index, err := strconv.Atoi(m.indexInput); err == nil && index >= 0 && index < 150 {
					m.inputMode = false
					m.indexInput = ""
					m.statusMessage = fmt.Sprintf("🎯 Jumping to task %d", index)
					return m, vtable.JumpToCmd(index)
				} else {
					m.statusMessage = "❌ Invalid index! Please enter a number between 0-149"
					m.inputMode = false
					m.indexInput = ""
					return m, nil
				}
			case "escape":
				m.inputMode = false
				m.indexInput = ""
				m.statusMessage = "🚫 Jump cancelled"
				return m, nil
			case "backspace":
				if len(m.indexInput) > 0 {
					m.indexInput = m.indexInput[:len(m.indexInput)-1]
				}
				return m, nil
			default:
				if len(msg.String()) == 1 && msg.String() >= "0" && msg.String() <= "9" {
					if len(m.indexInput) < 3 {
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

		case "e":
			// Cycle through enumerator styles
			m.currentStyle = (m.currentStyle + 1) % len(m.styleNames)
			m.applyCurrentStyle()
			m.statusMessage = fmt.Sprintf("🎨 Style: %s", m.styleNames[m.currentStyle])
			return m, nil

		case "t":
			// Toggle style info display
			m.showStyleInfo = !m.showStyleInfo
			if m.showStyleInfo {
				m.statusMessage = "📊 Style info visible"
			} else {
				m.statusMessage = "📊 Style info hidden"
			}
			return m, nil

		case "w":
			// Toggle text wrapping
			config := m.list.GetRenderConfig()
			config.WrapText = !config.WrapText
			m.list.SetRenderConfig(config)
			if config.WrapText {
				m.statusMessage = "📝 Text wrapping enabled"
			} else {
				m.statusMessage = "📝 Text wrapping disabled"
			}
			return m, nil

		case "i":
			// Toggle indent size
			config := m.list.GetRenderConfig()
			if config.IndentSize == 2 {
				config.IndentSize = 4
			} else {
				config.IndentSize = 2
			}
			m.list.SetRenderConfig(config)
			m.statusMessage = fmt.Sprintf("📐 Indent size: %d", config.IndentSize)
			return m, nil

		case "r":
			m.statusMessage = "🔄 Refreshing data..."
			return m, vtable.DataRefreshCmd()

		case "d":
			m.showDebug = !m.showDebug
			if m.showDebug {
				m.statusMessage = "🐛 Debug mode ON"
			} else {
				m.statusMessage = "🐛 Debug mode OFF"
			}
			return m, nil

		case "?":
			m.showHelp = !m.showHelp
			if m.showHelp {
				m.statusMessage = "❓ Help visible - press ? to hide"
			} else {
				m.statusMessage = "❓ Help hidden - press ? to show"
			}
			return m, nil

		// Navigation keys
		case "g":
			return m, vtable.JumpToStartCmd()
		case "G":
			return m, vtable.JumpToEndCmd()
		case "J":
			m.inputMode = true
			m.indexInput = ""
			m.statusMessage = "🎯 Enter task index to jump to (0-149): "
			return m, nil
		case "h":
			return m, vtable.PageUpCmd()
		case "l":
			return m, vtable.PageDownCmd()
		case "j", "up":
			return m, vtable.CursorUpCmd()
		case "k", "down":
			return m, vtable.CursorDownCmd()

		// Selection keys
		case " ":
			return m, vtable.SelectCurrentCmd()
		case "a":
			return m, vtable.SelectAllCmd()
		case "c":
			return m, vtable.SelectClearCmd()
		case "s":
			selectionCount := m.list.GetSelectionCount()
			if selectionCount > 0 {
				m.statusMessage = fmt.Sprintf("✅ SELECTED: %d tasks total", selectionCount)
			} else {
				m.statusMessage = "📝 No tasks selected - use Space to select"
			}
			return m, nil

		// Quick jump shortcuts
		case "1":
			return m, vtable.JumpToCmd(20)
		case "2":
			return m, vtable.JumpToCmd(50)
		case "3":
			return m, vtable.JumpToCmd(80)
		case "4":
			return m, vtable.JumpToCmd(120)
		case "5":
			return m, vtable.JumpToCmd(140)

		default:
			var cmd tea.Cmd
			_, cmd = m.list.Update(msg)
			state := m.list.GetState()
			m.statusMessage = fmt.Sprintf("📍 Position: %d/%d (Viewport: %d-%d)",
				state.CursorIndex+1, 150,
				state.ViewportStartIndex,
				state.ViewportStartIndex+9)
			return m, cmd
		}

	// Handle chunk loading messages
	case vtable.ChunkLoadingStartedMsg:
		m.loadingChunks[msg.ChunkStart] = true
		historyEntry := fmt.Sprintf("⏳ Loading chunk %d (size: %d)", msg.ChunkStart, msg.Request.Count)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		return m, cmd

	case vtable.ChunkLoadingCompletedMsg:
		delete(m.loadingChunks, msg.ChunkStart)
		historyEntry := fmt.Sprintf("✅ Completed chunk %d (%d items)", msg.ChunkStart, msg.ItemCount)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		return m, cmd

	case vtable.ChunkUnloadedMsg:
		historyEntry := fmt.Sprintf("🗑️  Unloaded chunk %d", msg.ChunkStart)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		return m, cmd

	// Handle selection responses
	case vtable.SelectionResponseMsg:
		switch msg.Operation {
		case "toggle":
			selectionCount := m.list.GetSelectionCount()
			state := m.list.GetState()
			if msg.Selected {
				m.statusMessage = fmt.Sprintf("✅ Selected task %d - %d total selected", state.CursorIndex, selectionCount)
			} else {
				m.statusMessage = fmt.Sprintf("❌ Deselected task %d - %d total selected", state.CursorIndex, selectionCount)
			}
		case "selectAll":
			selectionCount := m.list.GetSelectionCount()
			m.statusMessage = fmt.Sprintf("🎯 Selected ALL %d tasks!", selectionCount)
		case "clear":
			m.statusMessage = "🧹 All selections cleared"
		}
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		return m, cmd

	// Handle navigation messages
	case vtable.JumpToMsg:
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		state := m.list.GetState()
		m.statusMessage = fmt.Sprintf("🎯 Jumped to task %d", state.CursorIndex)
		return m, cmd

	case vtable.JumpToStartMsg:
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		m.statusMessage = "⬆️  Jumped to start"
		return m, cmd

	case vtable.JumpToEndMsg:
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		m.statusMessage = "⬇️  Jumped to end"
		return m, cmd

	default:
		var cmd tea.Cmd
		_, cmd = m.list.Update(msg)
		return m, cmd
	}
}

func (m *EnhancedAppModel) applyCurrentStyle() {
	switch m.currentStyle {
	case 0: // Bullets
		m.list.SetFormatter(nil) // Use enhanced rendering system
		m.list.SetBulletStyle()
		m.customFormatter = nil
	case 1: // Numbers
		m.list.SetFormatter(nil) // Use enhanced rendering system
		m.list.SetNumberedStyle()
		m.customFormatter = nil
	case 2: // Checkboxes
		m.list.SetFormatter(nil) // Use enhanced rendering system
		m.list.SetChecklistStyle()
		m.customFormatter = nil
	case 3: // Alphabetical
		m.list.SetFormatter(nil) // Use enhanced rendering system
		m.list.SetAlphabeticalStyle()
		m.customFormatter = nil
	case 4: // Dashes
		m.list.SetFormatter(nil) // Use enhanced rendering system
		m.list.SetDashStyle()
		m.customFormatter = nil
	case 5: // Arrows
		m.list.SetFormatter(nil) // Use enhanced rendering system
		m.list.SetCustomEnumerator("→ ")
		m.customFormatter = nil
	case 6: // Conditional
		m.list.SetFormatter(nil) // Use enhanced rendering system
		m.list.SetConditionalStyle()
		m.customFormatter = nil
	case 7: // Custom Pattern
		m.list.SetFormatter(nil) // Use enhanced rendering system
		m.list.SetCustomEnumerator("[{index1}] ")
		m.customFormatter = nil
	case 8: // Roman Numerals
		m.list.SetFormatter(nil) // Use enhanced rendering system
		m.list.SetEnumerator(vtable.RomanEnumerator)
		m.customFormatter = nil
	case 9: // Custom Formatter
		m.customFormatter = createCustomTaskFormatter()
		m.list.SetFormatter(m.customFormatter) // Use full custom formatter that handles everything
		// Reset to no enumerator since custom formatter handles everything
		config := m.list.GetRenderConfig()
		config.ShowEnumerator = false
		m.list.SetRenderConfig(config)
	}
}

func (m EnhancedAppModel) View() string {
	var view strings.Builder

	// Show help if enabled
	if m.showHelp {
		view.WriteString(m.renderHelp())
		view.WriteString("\n")
	}

	// Show style info if enabled
	if m.showStyleInfo {
		view.WriteString(m.renderStyleInfo())
		view.WriteString("\n")
	}

	// Show status message or input prompt
	if m.inputMode {
		view.WriteString(fmt.Sprintf("%s%s_", m.statusMessage, m.indexInput))
	} else {
		view.WriteString(m.statusMessage)
	}
	view.WriteString("\n\n")

	// Show main list content - now always uses a formatter
	content := m.list.View()
	view.WriteString(content)

	// Show selection info
	selectionCount := m.list.GetSelectionCount()
	if selectionCount > 0 {
		view.WriteString(fmt.Sprintf("\n\n✅ Selected: %d tasks", selectionCount))
	}

	// Show debug info if enabled
	if m.showDebug {
		view.WriteString("\n\n")
		view.WriteString(m.renderDebugInfo())
	}

	return view.String()
}

func (m EnhancedAppModel) renderHelp() string {
	var help strings.Builder
	help.WriteString("🎨 === ENHANCED LIST RENDERING DEMO ===\n")
	help.WriteString("Visual: ► = cursor • Various enumerators based on style • Error/Loading states shown\n")
	help.WriteString("Styles: e=cycle enumerator styles • t=toggle style info • w=toggle wrapping • i=toggle indent\n")
	help.WriteString("Navigation: j/k or ↑/↓ move • h/l page • g=start • G=end • J=jump • 1-5=quick jumps\n")
	help.WriteString("Selection: Space=toggle • a=select all • c=clear • s=show selection info\n")
	help.WriteString("Other: r=refresh • d=debug • ?=help • q=quit")
	return help.String()
}

func (m EnhancedAppModel) renderStyleInfo() string {
	var info strings.Builder
	info.WriteString(fmt.Sprintf("🎨 Current Style: %s (%d/%d)\n",
		m.styleNames[m.currentStyle], m.currentStyle+1, len(m.styleNames)))

	config := m.list.GetRenderConfig()
	info.WriteString(fmt.Sprintf("📐 Config: Wrapping=%v • Indent=%d • Alignment=%v • MaxWidth=%d",
		config.WrapText, config.IndentSize, config.AlignEnumerator, config.MaxWidth))

	return info.String()
}

func (m EnhancedAppModel) renderDebugInfo() string {
	var debug strings.Builder
	debug.WriteString("🐛 === CHUNK LOADING DEBUG ===\n")

	state := m.list.GetState()
	debug.WriteString(fmt.Sprintf("📍 Viewport: start=%d, cursor=%d (viewport_idx=%d)\n",
		state.ViewportStartIndex, state.CursorIndex, state.CursorViewportIndex))

	debug.WriteString(fmt.Sprintf("🎯 Thresholds: top=%v, bottom=%v\n",
		state.IsAtTopThreshold, state.IsAtBottomThreshold))

	if len(m.loadingChunks) > 0 {
		debug.WriteString("⏳ Loading chunks: ")
		var chunks []string
		for chunk := range m.loadingChunks {
			chunks = append(chunks, fmt.Sprintf("%d", chunk))
		}
		debug.WriteString(strings.Join(chunks, ", ") + "\n")
	}

	if len(m.chunkHistory) > 0 {
		debug.WriteString("📜 Recent activity:\n")
		for _, entry := range m.chunkHistory {
			debug.WriteString("  " + entry + "\n")
		}
	}

	if len(m.loadingChunks) == 0 && len(m.chunkHistory) == 0 {
		debug.WriteString("💤 No chunk activity yet\n")
	}

	return debug.String()
}

// createCustomTaskFormatter creates a completely custom formatter that doesn't use enumerators
func createCustomTaskFormatter() vtable.ItemFormatter[any] {
	return func(
		item vtable.Data[any],
		index int,
		ctx vtable.RenderContext,
		isCursor, isTopThreshold, isBottomThreshold bool,
	) string {
		task, ok := item.Item.(Task)
		if !ok {
			return fmt.Sprintf("❌ Invalid data: %v", item.Item)
		}

		// Custom prefix based on item state
		var prefix string
		var suffix string

		// Handle different states with custom icons
		switch {
		case item.Error != nil:
			if isCursor {
				prefix = "► 🚨 "
			} else {
				prefix = "  🚨 "
			}
			suffix = " (ERROR)"
		case item.Loading:
			if isCursor {
				prefix = "► ⏳ "
			} else {
				prefix = "  ⏳ "
			}
			suffix = " (LOADING)"
		case item.Disabled:
			if isCursor {
				prefix = "► 🚫 "
			} else {
				prefix = "  🚫 "
			}
			suffix = " (DISABLED)"
		case item.Selected && isCursor:
			prefix = "► ✅ "
			suffix = " ◄ SELECTED"
		case item.Selected:
			prefix = "  ✅ "
			suffix = " ◄ SELECTED"
		case isCursor:
			prefix = "► "
			suffix = ""
		default:
			// Use priority-based icons
			switch task.Priority {
			case "Critical":
				prefix = "  🔴 "
			case "High":
				prefix = "  🟠 "
			case "Medium":
				prefix = "  🟡 "
			case "Low":
				prefix = "  🟢 "
			default:
				prefix = "  ⚪ "
			}
		}

		// Add status icon
		var statusIcon string
		switch task.Status {
		case "Done":
			statusIcon = "✅"
		case "In Progress":
			statusIcon = "🔄"
		case "Review":
			statusIcon = "👀"
		case "Blocked":
			statusIcon = "🚫"
		default:
			statusIcon = "📝"
		}

		// Format progress bar
		progressBar := createProgressBar(task.Progress, 10)

		// Add threshold indicators
		thresholdIndicator := ""
		if isCursor {
			if isTopThreshold {
				thresholdIndicator = " [TOP]"
			} else if isBottomThreshold {
				thresholdIndicator = " [BOT]"
			}
		}

		// Format the main content with rich information
		content := fmt.Sprintf("%-30s | %s %s | %s | %s",
			task.Title, statusIcon, task.Status, progressBar, task.Category)

		return fmt.Sprintf("%s%s%s%s", prefix, content, thresholdIndicator, suffix)
	}
}

// createProgressBar creates a visual progress bar
func createProgressBar(progress, width int) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}

	filled := (progress * width) / 100
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("[%s] %d%%", bar, progress)
}

// createSimpleTaskFormatter creates a simple formatter that just formats the Task content nicely
func createSimpleTaskFormatter() vtable.ItemFormatter[any] {
	return func(
		item vtable.Data[any],
		index int,
		ctx vtable.RenderContext,
		isCursor, isTopThreshold, isBottomThreshold bool,
	) string {
		task, ok := item.Item.(Task)
		if !ok {
			return fmt.Sprintf("❌ Invalid data: %v", item.Item)
		}

		// Format the basic task content without cursor handling
		// The enhanced rendering system will handle enumerators and cursor styling
		content := fmt.Sprintf("%s | %s | %s | %s",
			task.Title, task.Priority, task.Status, task.Category)

		// Add state indicators
		var stateIndicator string
		switch {
		case item.Error != nil:
			stateIndicator = " ❌"
		case item.Loading:
			stateIndicator = " ⏳"
		case item.Disabled:
			stateIndicator = " 🚫"
		case item.Selected:
			stateIndicator = " ✅"
		}

		return content + stateIndicator
	}
}
