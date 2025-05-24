package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable"
)

// Task represents a task in our application
type Task struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
	Progress int    `json:"progress"`
}

// TaskDataProvider provides task data with selection support and dynamic updates
type TaskDataProvider struct {
	tasks         []Task
	selection     map[int]bool
	lastUpdate    time.Time
	taskIDCounter int
}

func NewTaskDataProvider() *TaskDataProvider {
	return &TaskDataProvider{
		tasks:         generateTasks(),
		selection:     make(map[int]bool),
		lastUpdate:    time.Now(),
		taskIDCounter: 10, // Start after initial tasks
	}
}

func generateTasks() []Task {
	return []Task{
		{ID: 1, Title: "Design homepage layout", Status: "active", Priority: "high", Progress: 75},
		{ID: 2, Title: "Implement user authentication", Status: "completed", Priority: "high", Progress: 100},
		{ID: 3, Title: "Write API documentation", Status: "pending", Priority: "medium", Progress: 0},
		{ID: 4, Title: "Set up CI/CD pipeline", Status: "active", Priority: "high", Progress: 50},
		{ID: 5, Title: "Create unit tests", Status: "active", Priority: "medium", Progress: 30},
		{ID: 6, Title: "Update database schema", Status: "pending", Priority: "low", Progress: 0},
		{ID: 7, Title: "Optimize query performance", Status: "active", Priority: "urgent", Progress: 25},
		{ID: 8, Title: "Review code changes", Status: "pending", Priority: "medium", Progress: 0},
		{ID: 9, Title: "Deploy to staging", Status: "completed", Priority: "high", Progress: 100},
		{ID: 10, Title: "Monitor system metrics", Status: "active", Priority: "medium", Progress: 60},
	}
}

func (p *TaskDataProvider) GetTotal() int {
	return len(p.tasks)
}

// GetFilteredTotal returns the count after applying current filters
func (p *TaskDataProvider) GetFilteredTotal(filters map[string]any) int {
	if len(filters) == 0 {
		return len(p.tasks)
	}

	count := 0
	for _, task := range p.tasks {
		// Check status filter
		if statusFilter, exists := filters["status"]; exists {
			if statusStr, ok := statusFilter.(string); ok && statusStr != task.Status {
				continue
			}
		}

		// Check priority filter
		if priorityFilter, exists := filters["priority"]; exists {
			if priorityStr, ok := priorityFilter.(string); ok && priorityStr != task.Priority {
				continue
			}
		}

		count++
	}
	return count
}

func (p *TaskDataProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[Task], error) {
	start := request.Start
	count := request.Count

	// Apply filters first
	filteredTasks := make([]Task, 0, len(p.tasks))
	for _, task := range p.tasks {
		// Check status filter
		if statusFilter, exists := request.Filters["status"]; exists {
			if statusStr, ok := statusFilter.(string); ok && statusStr != task.Status {
				continue // Skip this task
			}
		}

		// Check priority filter
		if priorityFilter, exists := request.Filters["priority"]; exists {
			if priorityStr, ok := priorityFilter.(string); ok && priorityStr != task.Priority {
				continue // Skip this task
			}
		}

		filteredTasks = append(filteredTasks, task)
	}

	if start >= len(filteredTasks) {
		return []vtable.Data[Task]{}, nil
	}

	end := start + count
	if end > len(filteredTasks) {
		end = len(filteredTasks)
	}

	result := make([]vtable.Data[Task], end-start)
	for i := start; i < end; i++ {
		// Find original index for selection state
		originalIndex := -1
		for j, original := range p.tasks {
			if original.ID == filteredTasks[i].ID {
				originalIndex = j
				break
			}
		}

		result[i-start] = vtable.Data[Task]{
			ID:       fmt.Sprintf("%d", filteredTasks[i].ID),
			Item:     filteredTasks[i],
			Selected: originalIndex >= 0 && p.selection[originalIndex],
			Metadata: vtable.NewTypedMetadata(),
			Disabled: false,
			Hidden:   false,
		}
	}

	return result, nil
}

// Implement remaining DataProvider methods
func (p *TaskDataProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionMultiple
}

func (p *TaskDataProvider) SetSelected(index int, selected bool) bool {
	if index < 0 || index >= len(p.tasks) {
		return false
	}
	if selected {
		p.selection[index] = true
	} else {
		delete(p.selection, index)
	}
	return true
}

func (p *TaskDataProvider) SelectAll() bool {
	for i := 0; i < len(p.tasks); i++ {
		p.selection[i] = true
	}
	return true
}

func (p *TaskDataProvider) ClearSelection() {
	p.selection = make(map[int]bool)
}

func (p *TaskDataProvider) GetSelectedIndices() []int {
	indices := make([]int, 0, len(p.selection))
	for idx := range p.selection {
		indices = append(indices, idx)
	}
	return indices
}

func (p *TaskDataProvider) GetItemID(item *Task) string {
	return fmt.Sprintf("%d", item.ID)
}

func (p *TaskDataProvider) GetSelectedIDs() []string {
	ids := make([]string, 0, len(p.selection))
	for idx := range p.selection {
		if idx < len(p.tasks) {
			ids = append(ids, fmt.Sprintf("%d", p.tasks[idx].ID))
		}
	}
	return ids
}

func (p *TaskDataProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	for _, id := range ids {
		for i, task := range p.tasks {
			if fmt.Sprintf("%d", task.ID) == id {
				if selected {
					p.selection[i] = true
				} else {
					delete(p.selection, i)
				}
				break
			}
		}
	}
	return true
}

func (p *TaskDataProvider) SelectRange(startID, endID string) bool {
	return true
}

// UpdateTasks simulates real task progression
func (p *TaskDataProvider) UpdateTasks() {
	now := time.Now()
	elapsed := now.Sub(p.lastUpdate)

	// Only update every 500ms to avoid too frequent changes
	if elapsed < 500*time.Millisecond {
		return
	}

	for i := range p.tasks {
		task := &p.tasks[i]

		// Progress active tasks based on priority
		if task.Status == "active" {
			progressSpeed := p.getProgressSpeed(task.Priority)
			// Add some randomness to make it feel real
			if rand.Float32() < 0.7 { // 70% chance to progress each update
				task.Progress += progressSpeed
				if task.Progress > 100 {
					task.Progress = 100
					task.Status = "completed"
				}
			}
		}

		// Activate some pending tasks occasionally
		if task.Status == "pending" && rand.Float32() < 0.05 { // 5% chance
			task.Status = "active"
		}
	}

	// Add new tasks occasionally (every ~10 seconds on average)
	if rand.Float32() < 0.02 && len(p.tasks) < 25 { // 2% chance, max 25 tasks
		p.addNewTask()
	}

	p.lastUpdate = now
}

func (p *TaskDataProvider) getProgressSpeed(priority string) int {
	switch priority {
	case "urgent":
		return rand.Intn(8) + 3 // 3-10 progress per update
	case "high":
		return rand.Intn(5) + 2 // 2-6 progress per update
	case "medium":
		return rand.Intn(3) + 1 // 1-3 progress per update
	case "low":
		return rand.Intn(2) + 1 // 1-2 progress per update
	default:
		return 1
	}
}

func (p *TaskDataProvider) addNewTask() {
	p.taskIDCounter++

	// Realistic new task templates
	templates := []struct {
		title    string
		priority string
		status   string
		progress int
	}{
		{"Fix production bug", "urgent", "pending", 0},
		{"Code review PR #" + fmt.Sprintf("%d", rand.Intn(1000)+100), "medium", "pending", 0},
		{"Update dependencies", "low", "pending", 0},
		{"Security audit", "high", "active", rand.Intn(20)},
		{"Performance optimization", "medium", "pending", 0},
		{"Write documentation", "low", "pending", 0},
		{"Database migration", "high", "pending", 0},
		{"User testing session", "medium", "active", rand.Intn(30)},
		{"Backup verification", "low", "pending", 0},
		{"Mobile app update", "high", "pending", 0},
	}

	template := templates[rand.Intn(len(templates))]

	newTask := Task{
		ID:       p.taskIDCounter,
		Title:    template.title,
		Status:   template.status,
		Priority: template.priority,
		Progress: template.progress,
	}

	p.tasks = append(p.tasks, newTask)
}

// RestartTasks resets the status and progress of tasks at given indices
func (p *TaskDataProvider) RestartTasks(indices []int) int {
	restarted := 0
	for _, index := range indices {
		if index >= 0 && index < len(p.tasks) {
			task := &p.tasks[index]
			// Only restart non-completed tasks or reset completed ones
			task.Status = "pending"
			task.Progress = 0
			restarted++
		}
	}
	return restarted
}

// Model for the animated example
type animatedModel struct {
	taskList          *vtable.TeaList[Task]
	provider          *TaskDataProvider // Add reference to provider
	currentView       int
	quitting          bool
	lastUpdate        time.Time
	tickCount         int
	activeFilter      string // Track current filter for display
	statusMessage     string // Show feedback messages
	animationsEnabled bool   // Track animation state
}

func newAnimatedModel() *animatedModel {
	// Create data provider
	provider := NewTaskDataProvider()

	// Create viewport config
	config := vtable.ViewportConfig{
		Height:               8,
		TopThresholdIndex:    1,
		BottomThresholdIndex: 6,
		ChunkSize:            50,
		InitialIndex:         0,
		Debug:                false,
	}

	// Create style config
	styleConfig := vtable.StyleConfig{
		BorderStyle:      "245",             // Gray
		HeaderStyle:      "bold 252 on 238", // Bold white on dark gray
		RowStyle:         "252",             // Light white
		SelectedRowStyle: "bold 252 on 63",  // Bold white on blue
	}

	// Regular formatter
	regularFormatter := func(data vtable.Data[Task], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		task := data.Item
		prefix := "  "
		if data.Selected {
			prefix = "✓ "
		}
		if isCursor {
			prefix = "> "
			if data.Selected {
				prefix = "✓>"
			}
		}

		status := getStatusEmoji(task.Status)
		priority := getPriorityColor(task.Priority)

		return fmt.Sprintf("%s%s %s [%s] %d%%",
			prefix,
			priority.Render(task.Title),
			status,
			task.Status,
			task.Progress,
		)
	}

	// Create the list
	list, err := vtable.NewTeaList(config, provider, styleConfig, regularFormatter)
	if err != nil {
		log.Fatal(err)
	}

	// Set a slower tick rate for the animation (250ms instead of default 100ms)
	list.SetTickInterval(250 * time.Millisecond)

	// Create animated formatter that shows live updates
	animatedFormatter := func(data vtable.Data[Task], index int, ctx vtable.RenderContext,
		animationState map[string]any, isCursor bool, isTopThreshold bool, isBottomThreshold bool) vtable.RenderResult {

		task := data.Item
		prefix := "  "
		if data.Selected {
			prefix = "✓ "
		}
		if isCursor {
			prefix = "> "
			if data.Selected {
				prefix = "✓>"
			}
		}

		// Get animation counter and use delta time for smooth progression
		counter := 0
		if c, ok := animationState["counter"]; ok {
			if ci, ok := c.(int); ok {
				counter = ci
			}
		}
		counter++

		// Use delta time for smooth animations (available since last render)
		deltaMs := ctx.DeltaTime.Milliseconds()

		// Add animated progress bar for active tasks
		progressBar := ""
		if task.Progress > 0 {
			barWidth := 10
			filled := (task.Progress * barWidth) / 100
			for i := 0; i < barWidth; i++ {
				if i < filled {
					progressBar += "█"
				} else {
					progressBar += "░"
				}
			}
		}

		// Add spinner for active tasks
		spinner := ""
		if task.Status == "active" || task.Status == "urgent" {
			spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
			spinner = spinnerFrames[counter%len(spinnerFrames)] + " "
		}

		// Blinking alert for urgent tasks (use delta time aware blinking)
		urgent := ""
		if task.Priority == "urgent" {
			// Blink based on time, not just counter (smoother blinking)
			if (ctx.CurrentTime.UnixMilli()/500)%2 == 0 {
				urgent = " 🚨"
			}
		}

		// Live timestamp with delta time info
		timestamp := ctx.CurrentTime.Format("15:04:05")
		deltaInfo := ""
		if deltaMs > 0 {
			// deltaInfo = fmt.Sprintf(" (Δ%dms)", deltaMs)
		}

		status := getStatusEmoji(task.Status)
		priority := getPriorityColor(task.Priority)

		var content string
		if progressBar != "" {
			content = fmt.Sprintf("%s%s%s %s [%s] %s (%s%s)%s",
				prefix,
				spinner,
				priority.Render(task.Title),
				status,
				task.Status,
				progressBar,
				timestamp,
				deltaInfo,
				urgent,
			)
		} else {
			content = fmt.Sprintf("%s%s%s %s [%s] %d%% (%s%s)%s",
				prefix,
				spinner,
				priority.Render(task.Title),
				status,
				task.Status,
				task.Progress,
				timestamp,
				deltaInfo,
				urgent,
			)
		}

		return vtable.RenderResult{
			Content: content,
			RefreshTriggers: []vtable.RefreshTrigger{{
				Type:     vtable.TriggerTimer,
				Interval: 250 * time.Millisecond, // Match the tick interval
			}},
			AnimationState: map[string]any{
				"counter":   counter,
				"timestamp": timestamp,
				"deltaMs":   deltaMs,
			},
		}
	}

	list.SetAnimatedFormatter(animatedFormatter)

	model := &animatedModel{
		taskList:          list,
		provider:          provider,
		currentView:       1, // Start in animated mode since we set animated formatter
		lastUpdate:        time.Now(),
		activeFilter:      "all", // Initialize with no filter
		animationsEnabled: true,  // Animations start enabled
	}

	return model
}

func getStatusEmoji(status string) string {
	switch status {
	case "completed":
		return "✅"
	case "active":
		return "🔄"
	case "urgent":
		return "🔥"
	case "pending":
		return "⏳"
	default:
		return "❓"
	}
}

func getPriorityColor(priority string) lipgloss.Style {
	switch priority {
	case "urgent":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) // Red
	case "high":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Orange
	case "medium":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
	case "low":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // Default
	}
}

func (m *animatedModel) Init() tea.Cmd {
	return m.taskList.Init()
}

func (m *animatedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "a":
			// Toggle animations dynamically
			if m.animationsEnabled {
				m.taskList.DisableAnimations()
				m.animationsEnabled = false
				m.statusMessage = "Animations disabled"
			} else {
				if cmd := m.taskList.EnableAnimations(); cmd != nil {
					cmds = append(cmds, cmd)
				}
				m.animationsEnabled = true
				m.statusMessage = "Animations enabled"
			}
			return m, tea.Batch(cmds...)
		case "tab":
			// Toggle between regular and animated view
			m.currentView = (m.currentView + 1) % 2
			if m.currentView == 0 {
				m.taskList.ClearAnimatedFormatter()
			} else {
				// Re-set animated formatter
				animatedFormatter := func(data vtable.Data[Task], index int, ctx vtable.RenderContext,
					animationState map[string]any, isCursor bool, isTopThreshold bool, isBottomThreshold bool) vtable.RenderResult {

					task := data.Item
					prefix := "  "
					if data.Selected {
						prefix = "✓ "
					}
					if isCursor {
						prefix = "> "
						if data.Selected {
							prefix = "✓>"
						}
					}

					// Get animation counter and use delta time for smooth progression
					counter := 0
					if c, ok := animationState["counter"]; ok {
						if ci, ok := c.(int); ok {
							counter = ci
						}
					}
					counter++

					// Use delta time for smooth animations (available since last render)
					deltaMs := ctx.DeltaTime.Milliseconds()

					// Add animated progress bar for active tasks
					progressBar := ""
					if task.Progress > 0 {
						barWidth := 10
						filled := (task.Progress * barWidth) / 100
						for i := 0; i < barWidth; i++ {
							if i < filled {
								progressBar += "█"
							} else {
								progressBar += "░"
							}
						}
					}

					// Add spinner for active tasks
					spinner := ""
					if task.Status == "active" || task.Status == "urgent" {
						spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
						spinner = spinnerFrames[counter%len(spinnerFrames)] + " "
					}

					// Blinking alert for urgent tasks (use delta time aware blinking)
					urgent := ""
					if task.Priority == "urgent" {
						// Blink based on time, not just counter (smoother blinking)
						if (ctx.CurrentTime.UnixMilli()/500)%2 == 0 {
							urgent = " 🚨"
						}
					}

					// Live timestamp with delta time info
					timestamp := ctx.CurrentTime.Format("15:04:05")
					deltaInfo := ""
					if deltaMs > 0 {
						deltaInfo = fmt.Sprintf(" (Δ%dms)", deltaMs)
					}

					status := getStatusEmoji(task.Status)
					priority := getPriorityColor(task.Priority)

					var content string
					if progressBar != "" {
						content = fmt.Sprintf("%s%s%s %s [%s] %s (%s%s)%s",
							prefix,
							spinner,
							priority.Render(task.Title),
							status,
							task.Status,
							progressBar,
							timestamp,
							deltaInfo,
							urgent,
						)
					} else {
						content = fmt.Sprintf("%s%s%s %s [%s] %d%% (%s%s)%s",
							prefix,
							spinner,
							priority.Render(task.Title),
							status,
							task.Status,
							task.Progress,
							timestamp,
							deltaInfo,
							urgent,
						)
					}

					return vtable.RenderResult{
						Content: content,
						RefreshTriggers: []vtable.RefreshTrigger{{
							Type:     vtable.TriggerTimer,
							Interval: 250 * time.Millisecond, // Match the tick interval
						}},
						AnimationState: map[string]any{
							"counter":   counter,
							"timestamp": timestamp,
							"deltaMs":   deltaMs,
						},
					}
				}
				m.taskList.SetAnimatedFormatter(animatedFormatter)
			}
			return m, nil
		case " ", "space":
			m.taskList.ToggleCurrentSelection()
			// Return early to prevent component processing the same key
			return m, tea.Batch(cmds...)
		case "enter":
			// Restart selected tasks
			selectedIndices := m.taskList.GetSelectedIndices()
			if len(selectedIndices) > 0 {
				restarted := m.provider.RestartTasks(selectedIndices)
				m.taskList.ClearSelection()
				// Refresh to show changes
				m.taskList.RefreshData()
				// You could add a status message here if desired
				m.statusMessage = fmt.Sprintf("Restarted %d tasks", restarted)
			}
			return m, tea.Batch(cmds...)
		case "1":
			// Show all tasks
			m.taskList.ClearFilters()
			m.activeFilter = "all"
			return m, tea.Batch(cmds...)
		case "2":
			// Show only active tasks
			m.taskList.ClearFilters()
			m.taskList.SetFilter("status", "active")
			m.activeFilter = "active"
			return m, tea.Batch(cmds...)
		case "3":
			// Show only pending tasks
			m.taskList.ClearFilters()
			m.taskList.SetFilter("status", "pending")
			m.activeFilter = "pending"
			return m, tea.Batch(cmds...)
		case "4":
			// Show only completed tasks
			m.taskList.ClearFilters()
			m.taskList.SetFilter("status", "completed")
			m.activeFilter = "completed"
			return m, tea.Batch(cmds...)
		case "5":
			// Show only urgent priority
			m.taskList.ClearFilters()
			m.taskList.SetFilter("priority", "urgent")
			m.activeFilter = "urgent priority"
			return m, tea.Batch(cmds...)
		}
	case vtable.GlobalAnimationTickMsg:
		// Track animation ticks
		m.tickCount++
		m.lastUpdate = time.Now()

		// Update task data - this makes the example truly dynamic!
		m.provider.UpdateTasks()

		// Refresh the list data to show real changes
		m.taskList.RefreshData()
	case vtable.AnimationUpdateMsg:
		// Animation updates received - no action needed, View() will handle
	}

	// Update the list
	updatedList, cmd := m.taskList.Update(msg)
	m.taskList = updatedList.(*vtable.TeaList[Task])
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *animatedModel) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var sb strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Render("Animated VTable Example")

	sb.WriteString(title + "\n\n")

	// Mode indicator
	mode := "Regular"
	modeDescription := "Static view with progress percentages"
	if m.currentView == 1 {
		mode = "Animated"
		modeDescription = "Live animations with delta time, configurable tick rate (250ms)"
	}

	modeStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("15")).
		Padding(0, 1)

	// Animation status indicator
	animStatus := "OFF"
	animColor := lipgloss.Color("9") // Red
	if m.animationsEnabled {
		animStatus = "ON"
		animColor = lipgloss.Color("10") // Green
	}

	animLoopRunning := m.taskList.IsAnimationLoopRunning()
	loopStatus := "stopped"
	if animLoopRunning {
		loopStatus = "running"
	}

	animStyle := lipgloss.NewStyle().
		Background(animColor).
		Foreground(lipgloss.Color("0")).
		Padding(0, 1)

	sb.WriteString(fmt.Sprintf("Mode: %s - %s | Animations: %s (loop %s)\n",
		modeStyle.Render(mode),
		modeDescription,
		animStyle.Render(animStatus),
		loopStatus,
	))

	// Filter indicator
	filterStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("208")).
		Foreground(lipgloss.Color("0")).
		Padding(0, 1)

	// Get current filter count
	currentFilters := m.taskList.GetDataRequest().Filters
	filteredCount := m.provider.GetFilteredTotal(currentFilters)
	totalCount := m.provider.GetTotal()

	filterInfo := fmt.Sprintf("Filter: %s (%d/%d tasks)",
		filterStyle.Render(m.activeFilter),
		filteredCount,
		totalCount,
	)

	sb.WriteString(fmt.Sprintf("%s | Ticks: %d | Last Update: %s\n\n",
		filterInfo,
		m.tickCount,
		m.lastUpdate.Format("15:04:05.000"),
	))

	// List
	sb.WriteString(m.taskList.View())

	// Help
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		MarginTop(1)

	help := helpStyle.Render(
		"Controls: ↑/↓ navigate • SPACE select • ENTER restart selected • TAB toggle animation • A toggle anim engine • q quit\n" +
			"Filters: 1 all • 2 active • 3 pending • 4 completed • 5 urgent priority\n" +
			"Features: Real-time filtering • Delta time animations • Dynamic data • Live progress • Dynamic anim control")
	sb.WriteString("\n" + help)

	// Status message
	if m.statusMessage != "" {
		sb.WriteString("\n\n")
		sb.WriteString(m.statusMessage)
		m.statusMessage = ""
	}

	return sb.String()
}

func main() {
	// Initialize random seed for realistic task progression
	rand.Seed(time.Now().UnixNano())

	model := newAnimatedModel()

	// Configure the program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
