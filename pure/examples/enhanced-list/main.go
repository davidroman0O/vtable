package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	vtable "github.com/davidroman0O/vtable/pure"
)

// ================================
// PURE TEA MODEL MESSAGES
// ================================

// Configuration change messages - PURE TEA MODEL
type RenderStyleChangeMsg struct {
	StyleIndex int
}

type BackgroundModeChangeMsg struct {
	ModeIndex int
}

type ComponentOrderChangeMsg struct {
	OrderIndex int
}

type CursorIndicatorChangeMsg struct {
	IndicatorIndex int
}

type EnumeratorAlignmentToggleMsg struct{}

type TextWrappingToggleMsg struct{}

type ComponentInfoToggleMsg struct{}

type DebugToggleMsg struct{}

type HelpToggleMsg struct{}

type InputModeToggleMsg struct {
	Enabled bool
	Prompt  string
}

type IndexInputMsg struct {
	Input string
}

type StatusUpdateMsg struct {
	Message string
}

// ================================
// DATA MODEL
// ================================

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

// ================================
// PURE TEA MODEL
// ================================

// ComponentDemoModel - PURE TEA MODEL IMPLEMENTATION
type ComponentDemoModel struct {
	list          *vtable.List
	dataSource    *EnhancedDataSource
	loadingChunks map[int]bool
	chunkHistory  []string
	showDebug     bool
	showHelp      bool
	statusMessage string
	indexInput    string
	inputMode     bool

	// Component rendering demo state - IMMUTABLE
	currentRenderStyle int
	renderStyleNames   []string
	renderConfigs      []vtable.ListRenderConfig

	// Background demo state - IMMUTABLE
	currentBackgroundMode int
	backgroundModeNames   []string

	// Component order demo state - IMMUTABLE
	currentOrderStyle int
	orderStyleNames   []string
	componentOrders   [][]vtable.ListComponentType

	// Cursor indicators - IMMUTABLE
	cursorIndicators    []string
	currentIndicatorIdx int

	showComponentInfo bool
}

func main() {
	// Create enhanced data source with 150 tasks
	dataSource := NewEnhancedDataSource(150)

	// Create different rendering configurations to showcase component system
	renderConfigs := []vtable.ListRenderConfig{
		// 1. Standard bullet list
		vtable.BulletListConfig(),

		// 2. Numbered list with alignment
		vtable.NumberedListConfig(),

		// 3. Checklist style
		vtable.ChecklistConfig(),

		// 4. Minimal (content only)
		vtable.MinimalListConfig(),

		// 5. Custom enumerator with arrows
		func() vtable.ListRenderConfig {
			config := vtable.DefaultListRenderConfig()
			config.EnumeratorConfig.Enumerator = func(data vtable.Data[any], index int, ctx vtable.RenderContext) string {
				return "→ "
			}
			return config
		}(),

		// 6. Conditional enumerator (changes based on state)
		func() vtable.ListRenderConfig {
			config := vtable.DefaultListRenderConfig()
			conditionalEnum := vtable.NewConditionalEnumerator(vtable.BulletEnumerator).
				When(vtable.IsSelected, vtable.CheckboxEnumerator).
				When(vtable.IsError, func(item vtable.Data[any], index int, ctx vtable.RenderContext) string {
					return "❌ "
				}).
				When(vtable.IsLoading, func(item vtable.Data[any], index int, ctx vtable.RenderContext) string {
					return "⏳ "
				})
			config.EnumeratorConfig.Enumerator = conditionalEnum.Enumerate
			return config
		}(),
	}

	renderStyleNames := []string{
		"Bullet List", "Numbered List", "Checklist", "Minimal (Content Only)",
		"Arrow Enumerator", "Conditional Enumerator",
	}

	// Different component orders to demonstrate flexibility
	componentOrders := [][]vtable.ListComponentType{
		// Standard order
		{vtable.ListComponentCursor, vtable.ListComponentEnumerator, vtable.ListComponentContent},
		// Content first
		{vtable.ListComponentContent, vtable.ListComponentEnumerator, vtable.ListComponentCursor},
		// Enumerator first
		{vtable.ListComponentEnumerator, vtable.ListComponentCursor, vtable.ListComponentContent},
		// Content only
		{vtable.ListComponentContent},
		// Cursor and content only
		{vtable.ListComponentCursor, vtable.ListComponentContent},
		// With spacing
		{vtable.ListComponentCursor, vtable.ListComponentPreSpacing, vtable.ListComponentEnumerator, vtable.ListComponentContent, vtable.ListComponentPostSpacing},
	}

	orderStyleNames := []string{
		"Standard (Cursor→Enum→Content)", "Content First (Content→Enum→Cursor)",
		"Enum First (Enum→Cursor→Content)", "Content Only", "Cursor+Content Only",
		"With Spacing (Cursor→Pre→Enum→Content→Post)",
	}

	backgroundModeNames := []string{
		"No Background", "Entire Line", "Content Only", "Indicator Only",
	}

	cursorIndicators := []string{"► ", "→ ", "* ", "• ", "▶ ", ""}

	// Create list configuration
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
		RenderConfig:  renderConfigs[0], // Start with bullet list
	}

	// Set content formatter for tasks
	config.RenderConfig.ContentConfig.Formatter = createTaskFormatter()

	// Create list - don't set a custom formatter so it uses the component system
	list := vtable.NewList(config, dataSource)

	// Create app model - PURE TEA MODEL
	app := ComponentDemoModel{
		list:          list,
		dataSource:    dataSource,
		loadingChunks: make(map[int]bool),
		chunkHistory:  make([]string, 0),
		showDebug:     true,
		showHelp:      true,
		statusMessage: "🎨 Component-Based Rendering Demo! Press keys to explore different rendering modes",
		indexInput:    "",
		inputMode:     false,

		// Component demo state - IMMUTABLE
		currentRenderStyle:    0,
		renderStyleNames:      renderStyleNames,
		renderConfigs:         renderConfigs,
		currentBackgroundMode: 0,
		backgroundModeNames:   backgroundModeNames,
		currentOrderStyle:     0,
		orderStyleNames:       orderStyleNames,
		componentOrders:       componentOrders,
		cursorIndicators:      cursorIndicators,
		currentIndicatorIdx:   0,
		showComponentInfo:     true,
	}

	// Run the program
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (m ComponentDemoModel) Init() tea.Cmd {
	return tea.Batch(
		m.list.Init(),
		m.list.Focus(),
	)
}

// PURE TEA MODEL UPDATE - NO DIRECT MUTATIONS
func (m ComponentDemoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle input mode for JumpToIndex
		if m.inputMode {
			switch msg.String() {
			case "enter":
				if index, err := strconv.Atoi(m.indexInput); err == nil && index >= 0 && index < 150 {
					m.inputMode = false
					m.indexInput = ""
					return m, tea.Batch(
						vtable.JumpToCmd(index),
						func() tea.Msg {
							return StatusUpdateMsg{Message: fmt.Sprintf("🎯 Jumping to task %d", index)}
						},
					)
				} else {
					m.inputMode = false
					m.indexInput = ""
					return m, func() tea.Msg {
						return StatusUpdateMsg{Message: "❌ Invalid index! Please enter a number between 0-149"}
					}
				}
			case "escape":
				m.inputMode = false
				m.indexInput = ""
				return m, func() tea.Msg {
					return StatusUpdateMsg{Message: "🚫 Jump cancelled"}
				}
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

		// Normal key handling - PURE TEA MODEL
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		// Component rendering style cycling - PURE TEA MODEL
		case "e":
			newStyle := (m.currentRenderStyle + 1) % len(m.renderConfigs)
			return m, func() tea.Msg {
				return RenderStyleChangeMsg{StyleIndex: newStyle}
			}

		// Background mode cycling - PURE TEA MODEL
		case "b":
			newMode := (m.currentBackgroundMode + 1) % len(m.backgroundModeNames)
			return m, func() tea.Msg {
				return BackgroundModeChangeMsg{ModeIndex: newMode}
			}

		// Component order cycling - PURE TEA MODEL
		case "o":
			newOrder := (m.currentOrderStyle + 1) % len(m.componentOrders)
			return m, func() tea.Msg {
				return ComponentOrderChangeMsg{OrderIndex: newOrder}
			}

		// Toggle component info display - PURE TEA MODEL
		case "t":
			return m, func() tea.Msg {
				return ComponentInfoToggleMsg{}
			}

		// Cursor indicator cycling - PURE TEA MODEL
		case "w":
			newIndicator := (m.currentIndicatorIdx + 1) % len(m.cursorIndicators)
			return m, func() tea.Msg {
				return CursorIndicatorChangeMsg{IndicatorIndex: newIndicator}
			}

		// Toggle enumerator alignment - PURE TEA MODEL
		case "a":
			return m, func() tea.Msg {
				return EnumeratorAlignmentToggleMsg{}
			}

		// Toggle text wrapping - PURE TEA MODEL
		case "W":
			return m, func() tea.Msg {
				return TextWrappingToggleMsg{}
			}

		case "r":
			return m, tea.Batch(
				vtable.DataRefreshCmd(),
				func() tea.Msg {
					return StatusUpdateMsg{Message: "🔄 Refreshing data..."}
				},
			)

		case "d":
			return m, func() tea.Msg {
				return DebugToggleMsg{}
			}

		case "?":
			return m, func() tea.Msg {
				return HelpToggleMsg{}
			}

		// Navigation keys
		case "g":
			return m, vtable.JumpToStartCmd()
		case "G":
			return m, vtable.JumpToEndCmd()
		case "J":
			return m, func() tea.Msg {
				return InputModeToggleMsg{
					Enabled: true,
					Prompt:  "🎯 Enter task index to jump to (0-149): ",
				}
			}
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
		case "A":
			return m, vtable.SelectAllCmd()
		case "c":
			return m, vtable.SelectClearCmd()
		case "s":
			selectionCount := m.list.GetSelectionCount()
			if selectionCount > 0 {
				return m, func() tea.Msg {
					return StatusUpdateMsg{Message: fmt.Sprintf("✅ SELECTED: %d tasks total", selectionCount)}
				}
			} else {
				return m, func() tea.Msg {
					return StatusUpdateMsg{Message: "📝 No tasks selected - use Space to select"}
				}
			}

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
			return m, tea.Batch(
				cmd,
				func() tea.Msg {
					return StatusUpdateMsg{
						Message: fmt.Sprintf("📍 Position: %d/%d (Viewport: %d-%d)",
							state.CursorIndex+1, 150,
							state.ViewportStartIndex,
							state.ViewportStartIndex+9),
					}
				},
			)
		}

	// PURE TEA MODEL MESSAGE HANDLING
	case RenderStyleChangeMsg:
		m.currentRenderStyle = msg.StyleIndex
		config := m.renderConfigs[m.currentRenderStyle]
		// Preserve current background and order settings
		currentConfig := m.list.GetRenderConfig()
		config.BackgroundConfig = currentConfig.BackgroundConfig
		config.ComponentOrder = currentConfig.ComponentOrder
		config.ContentConfig.Formatter = createTaskFormatter()

		// Clear any custom list formatter so the component system is used
		m.list.SetFormatter(nil)
		m.list.SetRenderConfig(config)

		return m, func() tea.Msg {
			return StatusUpdateMsg{Message: fmt.Sprintf("🎨 Render Style: %s", m.renderStyleNames[m.currentRenderStyle])}
		}

	case BackgroundModeChangeMsg:
		m.currentBackgroundMode = msg.ModeIndex
		config := m.list.GetRenderConfig()

		switch m.currentBackgroundMode {
		case 0: // No Background
			config.BackgroundConfig.Enabled = false
		case 1: // Entire Line
			config.BackgroundConfig.Enabled = true
			config.BackgroundConfig.Mode = vtable.ListBackgroundEntireLine
			config.BackgroundConfig.Style = lipgloss.NewStyle().
				Background(lipgloss.Color("240")).
				Foreground(lipgloss.Color("15"))
		case 2: // Content Only
			config.BackgroundConfig.Enabled = true
			config.BackgroundConfig.Mode = vtable.ListBackgroundContentOnly
			config.BackgroundConfig.Style = lipgloss.NewStyle().
				Background(lipgloss.Color("33")).
				Foreground(lipgloss.Color("15"))
		case 3: // Indicator Only
			config.BackgroundConfig.Enabled = true
			config.BackgroundConfig.Mode = vtable.ListBackgroundIndicatorOnly
			config.BackgroundConfig.Style = lipgloss.NewStyle().
				Background(lipgloss.Color("196")).
				Foreground(lipgloss.Color("15"))
		}

		m.list.SetRenderConfig(config)
		return m, func() tea.Msg {
			return StatusUpdateMsg{Message: fmt.Sprintf("🎨 Background Mode: %s", m.backgroundModeNames[m.currentBackgroundMode])}
		}

	case ComponentOrderChangeMsg:
		m.currentOrderStyle = msg.OrderIndex
		config := m.list.GetRenderConfig()
		config.ComponentOrder = m.componentOrders[m.currentOrderStyle]

		// Enable/disable components based on order
		config.CursorConfig.Enabled = false
		config.PreSpacingConfig.Enabled = false
		config.EnumeratorConfig.Enabled = false
		config.ContentConfig.Enabled = false
		config.PostSpacingConfig.Enabled = false

		for _, compType := range config.ComponentOrder {
			switch compType {
			case vtable.ListComponentCursor:
				config.CursorConfig.Enabled = true
			case vtable.ListComponentPreSpacing:
				config.PreSpacingConfig.Enabled = true
				config.PreSpacingConfig.Spacing = "  "
			case vtable.ListComponentEnumerator:
				config.EnumeratorConfig.Enabled = true
			case vtable.ListComponentContent:
				config.ContentConfig.Enabled = true
			case vtable.ListComponentPostSpacing:
				config.PostSpacingConfig.Enabled = true
				config.PostSpacingConfig.Spacing = " "
			}
		}

		m.list.SetRenderConfig(config)
		return m, func() tea.Msg {
			return StatusUpdateMsg{Message: fmt.Sprintf("🔄 Component Order: %s", m.orderStyleNames[m.currentOrderStyle])}
		}

	case CursorIndicatorChangeMsg:
		m.currentIndicatorIdx = msg.IndicatorIndex
		config := m.list.GetRenderConfig()
		indicator := m.cursorIndicators[m.currentIndicatorIdx]
		config.CursorConfig.CursorIndicator = indicator
		if indicator == "" {
			config.CursorConfig.NormalSpacing = ""
		} else {
			config.CursorConfig.NormalSpacing = "  "
		}
		m.list.SetRenderConfig(config)
		return m, func() tea.Msg {
			return StatusUpdateMsg{Message: fmt.Sprintf("🎨 Cursor Indicator: %q", indicator)}
		}

	case EnumeratorAlignmentToggleMsg:
		config := m.list.GetRenderConfig()
		if config.EnumeratorConfig.Alignment == vtable.ListAlignmentNone {
			config.EnumeratorConfig.Alignment = vtable.ListAlignmentRight
			config.EnumeratorConfig.MaxWidth = 8
			m.list.SetRenderConfig(config)
			return m, func() tea.Msg {
				return StatusUpdateMsg{Message: "📐 Enumerator alignment: RIGHT (width: 8)"}
			}
		} else {
			config.EnumeratorConfig.Alignment = vtable.ListAlignmentNone
			config.EnumeratorConfig.MaxWidth = 0
			m.list.SetRenderConfig(config)
			return m, func() tea.Msg {
				return StatusUpdateMsg{Message: "📐 Enumerator alignment: NONE"}
			}
		}

	case TextWrappingToggleMsg:
		config := m.list.GetRenderConfig()
		config.ContentConfig.WrapText = !config.ContentConfig.WrapText
		m.list.SetRenderConfig(config)
		return m, func() tea.Msg {
			return StatusUpdateMsg{Message: fmt.Sprintf("📝 Text wrapping: %v", config.ContentConfig.WrapText)}
		}

	case ComponentInfoToggleMsg:
		m.showComponentInfo = !m.showComponentInfo
		if m.showComponentInfo {
			return m, func() tea.Msg {
				return StatusUpdateMsg{Message: "📊 Component info visible"}
			}
		} else {
			return m, func() tea.Msg {
				return StatusUpdateMsg{Message: "📊 Component info hidden"}
			}
		}

	case DebugToggleMsg:
		m.showDebug = !m.showDebug
		if m.showDebug {
			return m, func() tea.Msg {
				return StatusUpdateMsg{Message: "🐛 Debug mode ON"}
			}
		} else {
			return m, func() tea.Msg {
				return StatusUpdateMsg{Message: "🐛 Debug mode OFF"}
			}
		}

	case HelpToggleMsg:
		m.showHelp = !m.showHelp
		if m.showHelp {
			return m, func() tea.Msg {
				return StatusUpdateMsg{Message: "❓ Help visible - press ? to hide"}
			}
		} else {
			return m, func() tea.Msg {
				return StatusUpdateMsg{Message: "❓ Help hidden - press ? to show"}
			}
		}

	case InputModeToggleMsg:
		m.inputMode = msg.Enabled
		m.statusMessage = msg.Prompt
		m.indexInput = ""
		return m, nil

	case StatusUpdateMsg:
		m.statusMessage = msg.Message
		return m, nil

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

func (m ComponentDemoModel) View() string {
	var view strings.Builder

	// Show help if enabled
	if m.showHelp {
		view.WriteString(m.renderHelp())
		view.WriteString("\n")
	}

	// Show component info if enabled
	if m.showComponentInfo {
		view.WriteString(m.renderComponentInfo())
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
		view.WriteString(fmt.Sprintf("\n\n✅ Selected: %d tasks", selectionCount))
	}

	// Show debug info if enabled
	if m.showDebug {
		view.WriteString("\n\n")
		view.WriteString(m.renderDebugInfo())
	}

	return view.String()
}

func (m ComponentDemoModel) renderHelp() string {
	var help strings.Builder
	help.WriteString("🎨 === COMPONENT-BASED RENDERING DEMO ===\n")
	help.WriteString("Rendering: e=cycle render styles • b=cycle background modes • o=cycle component orders\n")
	help.WriteString("Components: w=cycle cursor indicators • a=toggle enumerator alignment • W=toggle text wrapping\n")
	help.WriteString("Display: t=toggle component info • d=toggle debug • ?=toggle help\n")
	help.WriteString("Navigation: j/k or ↑/↓ move • h/l page • g=start • G=end • J=jump • 1-5=quick jumps\n")
	help.WriteString("Selection: Space=toggle • A=select all • c=clear • s=show selection info\n")
	help.WriteString("Other: r=refresh • q=quit")
	return help.String()
}

func (m ComponentDemoModel) renderComponentInfo() string {
	var info strings.Builder
	info.WriteString("🎨 === COMPONENT CONFIGURATION ===\n")

	config := m.list.GetRenderConfig()

	// Current styles
	info.WriteString(fmt.Sprintf("🎨 Render Style: %s (%d/%d)\n",
		m.renderStyleNames[m.currentRenderStyle], m.currentRenderStyle+1, len(m.renderStyleNames)))
	info.WriteString(fmt.Sprintf("🎨 Background Mode: %s (%d/%d)\n",
		m.backgroundModeNames[m.currentBackgroundMode], m.currentBackgroundMode+1, len(m.backgroundModeNames)))
	info.WriteString(fmt.Sprintf("🔄 Component Order: %s (%d/%d)\n",
		m.orderStyleNames[m.currentOrderStyle], m.currentOrderStyle+1, len(m.orderStyleNames)))

	// Component details
	info.WriteString(fmt.Sprintf("📐 Components: Order=%v\n", config.ComponentOrder))
	info.WriteString(fmt.Sprintf("🎯 Cursor: %v (indicator: %q, spacing: %q)\n",
		config.CursorConfig.Enabled, config.CursorConfig.CursorIndicator, config.CursorConfig.NormalSpacing))
	info.WriteString(fmt.Sprintf("📝 Enumerator: %v (alignment: %v, maxwidth: %d)\n",
		config.EnumeratorConfig.Enabled, config.EnumeratorConfig.Alignment, config.EnumeratorConfig.MaxWidth))
	info.WriteString(fmt.Sprintf("📄 Content: %v (wrap: %v, maxwidth: %d)\n",
		config.ContentConfig.Enabled, config.ContentConfig.WrapText, config.ContentConfig.MaxWidth))
	info.WriteString(fmt.Sprintf("🎨 Background: %v (mode: %v)\n",
		config.BackgroundConfig.Enabled, config.BackgroundConfig.Mode))

	return info.String()
}

func (m ComponentDemoModel) renderDebugInfo() string {
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

// createTaskFormatter creates a formatter for task content
func createTaskFormatter() vtable.ItemFormatter[any] {
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

		// Add state indicators
		var stateIndicator string

		// Add error/loading/disabled indicators
		if item.Error != nil {
			stateIndicator += " ❌"
		} else if item.Loading {
			stateIndicator += " ⏳"
		} else if item.Disabled {
			stateIndicator += " 🚫"
		}

		// Add selection indicator (independent of state)
		if item.Selected {
			stateIndicator += " ✅"
		}

		// Format the basic task content
		content := fmt.Sprintf("%s | %s | %s | %s",
			task.Title, task.Priority, task.Status, task.Category)

		return content + stateIndicator
	}
}
