package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/tree"
)

// Task represents our hierarchical task data
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

// String implements fmt.Stringer for clean task display
func (t Task) String() string {
	return t.Title
}

// TreeTaskDataSource implements TreeDataSource for hierarchical tasks
type TreeTaskDataSource struct {
	rootNodes     []tree.TreeData[Task]
	selectedNodes map[string]bool
}

func NewTreeTaskDataSource() *TreeTaskDataSource {
	ds := &TreeTaskDataSource{
		selectedNodes: make(map[string]bool),
	}
	ds.generateLargeDataset()
	return ds
}

func (ds *TreeTaskDataSource) generateLargeDataset() {
	// Create a much larger hierarchical dataset to demonstrate chunking
	var projects []tree.TreeData[Task]

	// Generate 10 major projects
	for projectIdx := 0; projectIdx < 10; projectIdx++ {
		projectID := fmt.Sprintf("project-%d", projectIdx)
		projectTitle := fmt.Sprintf("📁 Project %d - %s", projectIdx+1, getProjectName(projectIdx))

		var modules []tree.TreeData[Task]

		// Each project has 5-8 modules
		moduleCount := 5 + (projectIdx % 4)
		for moduleIdx := 0; moduleIdx < moduleCount; moduleIdx++ {
			moduleID := fmt.Sprintf("%s-module-%d", projectID, moduleIdx)
			moduleTitle := fmt.Sprintf("📦 Module %d - %s", moduleIdx+1, getModuleName(moduleIdx))

			var tasks []tree.TreeData[Task]

			// Each module has 8-15 tasks
			taskCount := 8 + (moduleIdx % 8)
			for taskIdx := 0; taskIdx < taskCount; taskIdx++ {
				taskID := fmt.Sprintf("%s-task-%d", moduleID, taskIdx)
				task := Task{
					ID:          taskID,
					Title:       fmt.Sprintf("🔧 %s", getTaskName(taskIdx)),
					Description: fmt.Sprintf("Task %d in module %d of project %d", taskIdx+1, moduleIdx+1, projectIdx+1),
					Priority:    getPriority(taskIdx),
					Status:      getStatus(taskIdx),
					Category:    getCategory(moduleIdx),
					Assignee:    getAssignee(taskIdx),
					Progress:    getProgress(taskIdx),
				}

				tasks = append(tasks, tree.TreeData[Task]{
					ID:       taskID,
					Item:     task,
					Expanded: false,
					Children: nil,
				})
			}

			modules = append(modules, tree.TreeData[Task]{
				ID: moduleID,
				Item: Task{
					ID:       moduleID,
					Title:    moduleTitle,
					Category: "Module",
					Status:   "In Progress",
				},
				Expanded: moduleIdx < 2, // Expand first 2 modules by default
				Children: tasks,
			})
		}

		projects = append(projects, tree.TreeData[Task]{
			ID: projectID,
			Item: Task{
				ID:       projectID,
				Title:    projectTitle,
				Category: "Project",
				Status:   "Active",
			},
			Expanded: projectIdx < 3, // Expand first 3 projects by default
			Children: modules,
		})
	}

	ds.rootNodes = projects
}

// Helper functions for generating varied data
func getProjectName(idx int) string {
	names := []string{
		"E-Commerce Platform", "Mobile Banking App", "AI Analytics Dashboard",
		"Cloud Infrastructure", "IoT Device Manager", "Social Media Platform",
		"Healthcare Portal", "Education System", "Gaming Engine", "Blockchain Network",
	}
	return names[idx%len(names)]
}

func getModuleName(idx int) string {
	names := []string{
		"Authentication", "User Management", "Payment Processing", "Data Analytics",
		"Notification System", "File Storage", "API Gateway", "Security Layer",
		"Reporting Engine", "Integration Hub", "Monitoring Dashboard", "Cache Layer",
	}
	return names[idx%len(names)]
}

func getTaskName(idx int) string {
	names := []string{
		"Setup Database Schema", "Implement REST API", "Create User Interface",
		"Add Authentication", "Write Unit Tests", "Setup CI/CD Pipeline",
		"Configure Monitoring", "Optimize Performance", "Add Documentation",
		"Security Audit", "Load Testing", "Bug Fixes", "Code Review",
		"Deploy to Staging", "User Acceptance Testing", "Production Deployment",
	}
	return names[idx%len(names)]
}

func getPriority(idx int) string {
	priorities := []string{"Low", "Medium", "High", "Critical"}
	return priorities[idx%len(priorities)]
}

func getStatus(idx int) string {
	statuses := []string{"Todo", "In Progress", "Review", "Done", "Blocked"}
	return statuses[idx%len(statuses)]
}

func getCategory(idx int) string {
	categories := []string{"Frontend", "Backend", "DevOps", "Testing", "Documentation"}
	return categories[idx%len(categories)]
}

func getAssignee(idx int) string {
	assignees := []string{
		"Alice", "Bob", "Carol", "David", "Eve", "Frank", "Grace", "Henry",
		"Ivy", "Jack", "Kate", "Liam", "Mia", "Noah", "Olivia", "Paul",
	}
	return assignees[idx%len(assignees)]
}

func getProgress(idx int) int {
	return (idx * 13) % 101 // 0-100%
}

// Implement TreeDataSource interface
func (ds *TreeTaskDataSource) GetRootNodes() []tree.TreeData[Task] {
	return ds.rootNodes
}

func (ds *TreeTaskDataSource) GetItemByID(id string) (tree.TreeData[Task], bool) {
	return ds.findNodeByID(ds.rootNodes, id)
}

func (ds *TreeTaskDataSource) findNodeByID(nodes []tree.TreeData[Task], id string) (tree.TreeData[Task], bool) {
	for _, node := range nodes {
		if node.ID == id {
			return node, true
		}
		if found, ok := ds.findNodeByID(node.Children, id); ok {
			return found, true
		}
	}
	return tree.TreeData[Task]{}, false
}

func (ds *TreeTaskDataSource) SetSelected(id string, selected bool) tea.Cmd {
	if selected {
		ds.selectedNodes[id] = true
	} else {
		delete(ds.selectedNodes, id)
	}
	return core.SelectionResponseCmd(true, -1, id, selected, "toggle", nil, nil)
}

func (ds *TreeTaskDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	return ds.SetSelected(id, selected)
}

func (ds *TreeTaskDataSource) SelectAll() tea.Cmd {
	// Select all nodes in the tree
	ds.selectAllNodes(ds.rootNodes)
	return core.SelectionResponseCmd(true, -1, "", true, "selectAll", nil, nil)
}

func (ds *TreeTaskDataSource) selectAllNodes(nodes []tree.TreeData[Task]) {
	for _, node := range nodes {
		ds.selectedNodes[node.ID] = true
		ds.selectAllNodes(node.Children)
	}
}

func (ds *TreeTaskDataSource) ClearSelection() tea.Cmd {
	ds.selectedNodes = make(map[string]bool)
	return core.SelectionResponseCmd(true, -1, "", false, "clear", nil, nil)
}

func (ds *TreeTaskDataSource) SelectRange(startID, endID string) tea.Cmd {
	// For trees, range selection is more complex - for now, just select both nodes
	ds.selectedNodes[startID] = true
	ds.selectedNodes[endID] = true
	return core.SelectionResponseCmd(true, -1, "", true, "range", nil, []string{startID, endID})
}

// TreeAppModel wraps our tree list
type TreeAppModel struct {
	treeList      *tree.TreeList[Task]
	dataSource    *TreeTaskDataSource
	loadingChunks map[int]bool
	chunkHistory  []string
	showDebug     bool
	showHelp      bool
	statusMessage string
	currentStyle  int
	styleNames    []string
	cursorStyle   int
	cursorStyles  []CursorStyleConfig
	// Add jump to index functionality
	inputMode bool   // true when entering a number for JumpToIndex
	jumpInput string // the input string being built
}

// CursorStyleConfig represents different cursor styling options
type CursorStyleConfig struct {
	Name            string
	CursorIndicator string
	CursorSpacing   string
	NormalSpacing   string
	ShowCursor      bool
	BackgroundColor string
	ForegroundColor string
	EnableStyling   bool
	Description     string
}

func main() {
	// Create tree data source with large dataset
	dataSource := NewTreeTaskDataSource()

	// Create list configuration optimized for large datasets
	config := core.ListConfig{
		ViewportConfig: core.ViewportConfig{
			Height:             15, // Larger viewport for better tree viewing
			TopThreshold:       3,  // More threshold space for trees
			BottomThreshold:    3,
			ChunkSize:          20, // Larger chunks for better performance
			InitialIndex:       0,
			BoundingAreaBefore: 10, // Larger bounding area for smoother scrolling
			BoundingAreaAfter:  10,
		},
		SelectionMode: core.SelectionMultiple,
		KeyMap:        core.DefaultNavigationKeyMap(),
		MaxWidth:      120,
	}

	// Create tree configuration with component-based rendering
	treeConfig := tree.DefaultTreeConfig()

	// Set up a custom tree formatter that enhances task display
	treeConfig.RenderConfig.ContentConfig.Formatter = createTaskTreeFormatter()

	// Create tree list
	treeList := tree.NewTreeList(config, treeConfig, dataSource)

	// Define cursor style options (updated for component system)
	cursorStyles := []CursorStyleConfig{
		{
			Name: "Arrow", CursorIndicator: "► ", CursorSpacing: "  ", NormalSpacing: "  ",
			ShowCursor: true, BackgroundColor: "240", ForegroundColor: "15", EnableStyling: true,
			Description: "Arrow indicator with background highlight",
		},
		{
			Name: "Pointer", CursorIndicator: "→ ", CursorSpacing: "  ", NormalSpacing: "  ",
			ShowCursor: true, BackgroundColor: "33", ForegroundColor: "15", EnableStyling: true,
			Description: "Pointer indicator with blue background",
		},
		{
			Name: "Star", CursorIndicator: "★ ", CursorSpacing: "  ", NormalSpacing: "  ",
			ShowCursor: true, BackgroundColor: "214", ForegroundColor: "0", EnableStyling: true,
			Description: "Star indicator with orange background",
		},
		{
			Name: "Bullet", CursorIndicator: "• ", CursorSpacing: "  ", NormalSpacing: "  ",
			ShowCursor: true, BackgroundColor: "28", ForegroundColor: "15", EnableStyling: true,
			Description: "Bullet indicator with green background",
		},
		{
			Name: "Bracket", CursorIndicator: "[>] ", CursorSpacing: "    ", NormalSpacing: "    ",
			ShowCursor: true, BackgroundColor: "196", ForegroundColor: "15", EnableStyling: true,
			Description: "Bracket indicator with red background",
		},
		{
			Name: "Background Only", CursorIndicator: "", CursorSpacing: "", NormalSpacing: "",
			ShowCursor: false, BackgroundColor: "240", ForegroundColor: "15", EnableStyling: true,
			Description: "No indicator, only background highlight",
		},
		{
			Name: "Subtle Background", CursorIndicator: "", CursorSpacing: "", NormalSpacing: "",
			ShowCursor: false, BackgroundColor: "235", ForegroundColor: "250", EnableStyling: true,
			Description: "No indicator, subtle gray background",
		},
		{
			Name: "Bright Background", CursorIndicator: "", CursorSpacing: "", NormalSpacing: "",
			ShowCursor: false, BackgroundColor: "51", ForegroundColor: "0", EnableStyling: true,
			Description: "No indicator, bright cyan background",
		},
		{
			Name: "No Cursor", CursorIndicator: "", CursorSpacing: "", NormalSpacing: "",
			ShowCursor: false, BackgroundColor: "", ForegroundColor: "", EnableStyling: false,
			Description: "No visual cursor indication",
		},
	}

	// Create app model
	app := TreeAppModel{
		treeList:      treeList,
		dataSource:    dataSource,
		loadingChunks: make(map[int]bool),
		chunkHistory:  make([]string, 0),
		showDebug:     true,
		showHelp:      true,
		statusMessage: "🌳 Large Tree Demo! Navigate with j/k, expand/collapse with Enter. Watch chunk loading!",
		currentStyle:  0,
		styleNames:    []string{"Default", "Standard", "Minimal", "Enumerated"},
		cursorStyle:   0,
		cursorStyles:  cursorStyles,
		inputMode:     false,
		jumpInput:     "",
	}

	// Run the program
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

// createTaskTreeFormatter creates a custom tree formatter for tasks
func createTaskTreeFormatter() tree.TreeItemFormatter {
	return func(
		item core.Data[any],
		index int,
		depth int,
		hasChildren, isExpanded bool,
		ctx core.RenderContext,
		isCursor, isTopThreshold, isBottomThreshold bool,
	) string {
		// Type assert to our Task type
		flatItem, ok := item.Item.(tree.FlatTreeItem[Task])
		if !ok {
			return fmt.Sprintf("Invalid item: %v", item.Item)
		}

		task := flatItem.Item

		// Just the task title - no selection indicators here
		var content strings.Builder
		content.WriteString(task.Title)

		// Add status indicator for tasks (not projects/modules)
		if !strings.Contains(task.Title, "📁") && !strings.Contains(task.Title, "📦") {
			switch task.Status {
			case "Done":
				content.WriteString(" ✅")
			case "In Progress":
				content.WriteString(" 🔄")
			case "Blocked":
				content.WriteString(" ❌")
			case "Review":
				content.WriteString(" 👀")
			case "Todo":
				content.WriteString(" 📝")
			}

			// Add priority indicator for high/critical tasks
			if task.Priority == "High" {
				content.WriteString(" 🔥")
			} else if task.Priority == "Critical" {
				content.WriteString(" 🚨")
			}

			// Add progress for in-progress tasks
			if task.Status == "In Progress" && task.Progress > 0 {
				content.WriteString(fmt.Sprintf(" (%d%%)", task.Progress))
			}
		}

		// Add assignee for leaf tasks
		if !hasChildren && task.Assignee != "" && task.Assignee != task.Title {
			content.WriteString(fmt.Sprintf(" [@%s]", task.Assignee))
		}

		// Add selection indicator AT THE END - simple and clean
		if item.Selected {
			content.WriteString(" ✅")
		}

		result := content.String()

		// Apply background styling for selected items
		if item.Selected {
			style := lipgloss.NewStyle().
				Background(lipgloss.Color("240")).
				Foreground(lipgloss.Color("15"))
			result = style.Render(result)
		}

		return result
	}
}

func (m TreeAppModel) Init() tea.Cmd {
	return tea.Batch(
		m.treeList.Init(),
		m.treeList.Focus(),
	)
}

func (m TreeAppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle input mode for JumpToIndex
		if m.inputMode {
			switch msg.String() {
			case "enter":
				// Parse the input and jump to index
				if index, err := strconv.Atoi(m.jumpInput); err == nil && index >= 0 {
					m.inputMode = false
					m.jumpInput = ""
					// Get the total possible indices in fully expanded tree
					fullyExpandedCount := m.treeList.GetFullyExpandedItemCount()
					if index < fullyExpandedCount {
						m.statusMessage = fmt.Sprintf("🎯 Jumping to index %d with parent expansion", index)
						return m, m.treeList.JumpToIndexExpandingParents(index)
					} else {
						m.statusMessage = fmt.Sprintf("❌ Index %d is beyond tree range (max: %d)", index, fullyExpandedCount-1)
						return m, nil
					}
				} else {
					m.statusMessage = "❌ Invalid index! Please enter a valid number"
					m.inputMode = false
					m.jumpInput = ""
					return m, nil
				}
			case "escape":
				m.inputMode = false
				m.jumpInput = ""
				m.statusMessage = "🚫 Jump cancelled"
				return m, nil
			case "backspace":
				if len(m.jumpInput) > 0 {
					m.jumpInput = m.jumpInput[:len(m.jumpInput)-1]
				}
				return m, nil
			default:
				// Only allow digits
				if len(msg.String()) == 1 && msg.String() >= "0" && msg.String() <= "9" {
					if len(m.jumpInput) < 5 { // Limit to 5 digits for very large trees
						m.jumpInput += msg.String()
					}
				}
				return m, nil
			}
		}

		// Normal key handling
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			// Toggle current node
			return m, m.treeList.ToggleCurrentNode()

		case "r":
			// Force refresh to see chunk loading again
			m.statusMessage = "🔄 Refreshing tree data..."
			return m, core.DataRefreshCmd()

		case "e":
			// Cycle tree styles
			m.currentStyle = (m.currentStyle + 1) % len(m.styleNames)
			m.applyCurrentStyle()
			m.statusMessage = fmt.Sprintf("🎨 Tree Style: %s", m.styleNames[m.currentStyle])
			return m, nil

		case "c":
			// Cycle cursor styles
			m.cursorStyle = (m.cursorStyle + 1) % len(m.cursorStyles)
			m.applyCursorStyle()
			currentStyle := m.cursorStyles[m.cursorStyle]
			m.statusMessage = fmt.Sprintf("🎯 Cursor Style: %s - %s", currentStyle.Name, currentStyle.Description)
			return m, nil

		case "C":
			// Toggle cascading selection (uppercase C)
			currentCascading := m.treeList.GetCascadingSelection()
			m.treeList.SetCascadingSelection(!currentCascading)
			if !currentCascading {
				m.statusMessage = "🔗 Cascading Selection: ON - selecting parents will select all children"
			} else {
				m.statusMessage = "🔗 Cascading Selection: OFF - only individual items are selected"
			}
			return m, nil

		case "b":
			// Toggle background cursor styling
			config := m.treeList.GetRenderConfig()
			currentStyling := config.BackgroundConfig.Enabled
			config.BackgroundConfig.Enabled = !currentStyling
			m.treeList.SetRenderConfig(config)
			if !currentStyling {
				m.statusMessage = "🎨 Background Cursor Styling: ON - cursor line has background color"
			} else {
				m.statusMessage = "🎨 Background Cursor Styling: OFF - no background highlighting"
			}
			return m, nil

		case "d":
			m.showDebug = !m.showDebug
			if m.showDebug {
				m.statusMessage = "🐛 Debug mode ON - watch chunk loading activity!"
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

		// Quick navigation shortcuts
		case "g":
			return m, core.JumpToStartCmd()
		case "G":
			return m, core.JumpToEndCmd()
		case "h":
			return m, core.PageUpCmd()
		case "l":
			return m, core.PageDownCmd()

		case "J":
			// Enter jump-to-index mode (uppercase J)
			m.inputMode = true
			m.jumpInput = ""
			fullyExpandedCount := m.treeList.GetFullyExpandedItemCount()
			m.statusMessage = fmt.Sprintf("🎯 Jump to index (0-%d): ", fullyExpandedCount-1)
			return m, nil

		// Navigation keys - let TreeList handle these
		case "j", "down":
			return m, core.CursorDownCmd()
		case "k", "up":
			return m, core.CursorUpCmd()

		// Selection keys
		case " ":
			return m, core.SelectCurrentCmd()
		case "a":
			return m, core.SelectAllCmd()
		case "x":
			return m, core.SelectClearCmd()
		case "s":
			selectionCount := m.treeList.GetSelectionCount()
			if selectionCount > 0 {
				m.statusMessage = fmt.Sprintf("✅ SELECTED: %d items total", selectionCount)
			} else {
				m.statusMessage = "📝 No items selected - use Space to select"
			}
			return m, nil

		default:
			// Let TreeList handle other keys
			var cmd tea.Cmd
			_, cmd = m.treeList.Update(msg)

			// Update status with current position
			state := m.treeList.GetState()
			m.statusMessage = fmt.Sprintf("🌳 Position: %d (Viewport: %d-%d)",
				state.CursorIndex+1,
				state.ViewportStartIndex,
				state.ViewportStartIndex+14) // viewport height - 1

			return m, cmd
		}

	// Handle chunk loading observability messages
	case core.ChunkLoadingStartedMsg:
		m.loadingChunks[msg.ChunkStart] = true
		historyEntry := fmt.Sprintf("🔄 Started loading chunk %d (size: %d)", msg.ChunkStart, msg.Request.Count)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		// Also pass to tree list
		var cmd tea.Cmd
		_, cmd = m.treeList.Update(msg)
		return m, cmd

	case core.ChunkLoadingCompletedMsg:
		delete(m.loadingChunks, msg.ChunkStart)
		historyEntry := fmt.Sprintf("✅ Completed chunk %d (%d items)", msg.ChunkStart, msg.ItemCount)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		// Also pass to tree list
		var cmd tea.Cmd
		_, cmd = m.treeList.Update(msg)
		return m, cmd

	case core.ChunkUnloadedMsg:
		historyEntry := fmt.Sprintf("🗑️ Unloaded chunk %d", msg.ChunkStart)
		m.chunkHistory = append(m.chunkHistory, historyEntry)
		// Keep only last 10 entries
		if len(m.chunkHistory) > 10 {
			m.chunkHistory = m.chunkHistory[1:]
		}
		// Also pass to tree list
		var cmd tea.Cmd
		_, cmd = m.treeList.Update(msg)
		return m, cmd

	// Handle selection response messages
	case core.SelectionResponseMsg:
		// Update status based on selection operation
		switch msg.Operation {
		case "toggle":
			selectionCount := m.treeList.GetSelectionCount()
			state := m.treeList.GetState()
			if msg.Selected {
				m.statusMessage = fmt.Sprintf("✅ Selected item at position %d - %d items selected total", state.CursorIndex+1, selectionCount)
			} else {
				m.statusMessage = fmt.Sprintf("❌ Deselected item at position %d - %d items selected total", state.CursorIndex+1, selectionCount)
			}
		case "selectAll":
			selectionCount := m.treeList.GetSelectionCount()
			m.statusMessage = fmt.Sprintf("✅ Selected ALL %d items in tree!", selectionCount)
		case "clear":
			m.statusMessage = "🧹 All selections cleared"
		}
		// Also pass to tree list
		var cmd tea.Cmd
		_, cmd = m.treeList.Update(msg)
		return m, cmd

	// Handle navigation messages to update status
	case core.PageUpMsg, core.PageDownMsg, core.JumpToMsg, core.JumpToStartMsg, core.JumpToEndMsg:
		var cmd tea.Cmd
		_, cmd = m.treeList.Update(msg)
		state := m.treeList.GetState()
		m.statusMessage = fmt.Sprintf("🌳 Position: %d (Viewport: %d-%d)",
			state.CursorIndex+1,
			state.ViewportStartIndex,
			state.ViewportStartIndex+14)
		return m, cmd

	case core.TreeJumpToIndexMsg:
		var cmd tea.Cmd
		_, cmd = m.treeList.Update(msg)
		state := m.treeList.GetState()
		m.statusMessage = fmt.Sprintf("🎯 Jumped to index %d (Position: %d, Viewport: %d-%d)",
			msg.Index,
			state.CursorIndex+1,
			state.ViewportStartIndex,
			state.ViewportStartIndex+14)
		return m, cmd

	case core.CursorUpMsg, core.CursorDownMsg:
		var cmd tea.Cmd
		_, cmd = m.treeList.Update(msg)
		state := m.treeList.GetState()
		m.statusMessage = fmt.Sprintf("🌳 Position: %d (Viewport: %d-%d)",
			state.CursorIndex+1,
			state.ViewportStartIndex,
			state.ViewportStartIndex+14)
		return m, cmd

	default:
		// Let TreeList handle all other messages
		var cmd tea.Cmd
		_, cmd = m.treeList.Update(msg)
		return m, cmd
	}
}

func (m *TreeAppModel) applyCurrentStyle() {
	switch m.currentStyle {
	case 0: // Default
		// Use default tree configuration
		config := tree.DefaultTreeRenderConfig()
		config.ContentConfig.Formatter = createTaskTreeFormatter()
		m.treeList.SetRenderConfig(config)
	case 1: // Standard (with box-drawing connectors)
		config := tree.StandardTreeConfig()
		config.ContentConfig.Formatter = createTaskTreeFormatter()
		m.treeList.SetRenderConfig(config)
	case 2: // Minimal (no tree symbols, just indentation)
		config := tree.MinimalTreeConfig()
		config.ContentConfig.Formatter = createTaskTreeFormatter()
		m.treeList.SetRenderConfig(config)
	case 3: // Enumerated (with bullet enumeration)
		bulletEnum := func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext) string {
			return "• "
		}
		config := tree.EnumeratedTreeConfig(bulletEnum)
		config.ContentConfig.Formatter = createTaskTreeFormatter()
		m.treeList.SetRenderConfig(config)
	}
}

func (m *TreeAppModel) applyCursorStyle() {
	style := m.cursorStyles[m.cursorStyle]

	// Get current render config and update cursor settings
	config := m.treeList.GetRenderConfig()

	// Configure cursor indicators that show selection state
	if style.ShowCursor {
		config.CursorConfig.CursorIndicator = style.CursorIndicator
		config.CursorConfig.NormalSpacing = style.NormalSpacing
		config.CursorConfig.Enabled = true
	} else {
		config.CursorConfig.Enabled = false
		config.CursorConfig.CursorIndicator = ""
		config.CursorConfig.NormalSpacing = ""
	}

	// Apply background styling that differentiates selected vs unselected items
	if style.EnableStyling && style.BackgroundColor != "" && style.ForegroundColor != "" {
		config.BackgroundConfig.Enabled = true
		config.BackgroundConfig.Style = lipgloss.NewStyle().
			Background(lipgloss.Color(style.BackgroundColor)).
			Foreground(lipgloss.Color(style.ForegroundColor))
		config.BackgroundConfig.Mode = tree.TreeBackgroundEntireLine
	} else {
		config.BackgroundConfig.Enabled = false
	}

	m.treeList.SetRenderConfig(config)
}

func (m TreeAppModel) View() string {
	var view strings.Builder

	// Show help if enabled
	if m.showHelp {
		view.WriteString(m.renderHelp())
		view.WriteString("\n")
	}

	// Show status message
	if m.inputMode {
		// Show input prompt with current input
		view.WriteString(fmt.Sprintf("%s%s_", m.statusMessage, m.jumpInput))
	} else {
		view.WriteString(m.statusMessage)
	}
	view.WriteString("\n\n")

	// Show main tree list content
	content := m.treeList.View()
	view.WriteString(content)

	// Show selection info
	selectionCount := m.treeList.GetSelectionCount()
	if selectionCount > 0 {
		view.WriteString(fmt.Sprintf("\n\n✅ Selected: %d items", selectionCount))
	}

	// Show debug info if enabled
	if m.showDebug {
		view.WriteString("\n\n")
		view.WriteString(m.renderDebugInfo())
	}

	return view.String()
}

func (m TreeAppModel) renderHelp() string {
	var help strings.Builder
	help.WriteString("🌳 === LARGE TREE DEMO WITH COMPONENT-BASED RENDERING ===\n")
	help.WriteString("Dataset: 10 projects × 5-8 modules × 8-15 tasks = ~500-1200 items total\n")
	help.WriteString("Visual: ► = cursor • Background = selected • 🔄 = in progress • 🔥 = high priority\n")
	help.WriteString("Selection: Selected items have background styling - no text indicators!\n")
	help.WriteString("Tree: Enter=expand/collapse • e=cycle tree styles (4 options) • c=cycle cursor styles (9 options)\n")
	help.WriteString("Tree Styles: Default, Standard (box connectors), Minimal, Enumerated\n")
	help.WriteString("Cursor Styles: Arrow, Pointer, Star, Bullet, Bracket, Background Only, Subtle, Bright, None\n")
	help.WriteString("Component: C=toggle cascading selection • b=toggle background cursor styling\n")
	help.WriteString("Navigation: j/k or ↑/↓ move • h/l page • g=start • G=end • J=jump to index\n")
	help.WriteString("Selection: Space=toggle • a=select all • x=clear • s=show selection info\n")
	help.WriteString("Cascading: When ON, selecting a parent automatically selects all children\n")
	help.WriteString("Other: r=refresh • d=debug (shows component config) • ?=help • q=quit")
	return help.String()
}

func (m TreeAppModel) renderDebugInfo() string {
	var debug strings.Builder
	debug.WriteString("🐛 === TREE CHUNK LOADING DEBUG ===\n")

	// Show viewport and bounding area details
	state := m.treeList.GetState()
	debug.WriteString(fmt.Sprintf("📍 Viewport: start=%d, cursor=%d (viewport_idx=%d)\n",
		state.ViewportStartIndex, state.CursorIndex, state.CursorViewportIndex))

	// Show threshold flags
	debug.WriteString(fmt.Sprintf("🎯 Thresholds: top=%v, bottom=%v\n",
		state.IsAtTopThreshold, state.IsAtBottomThreshold))

	// Show tree configuration
	debug.WriteString(fmt.Sprintf("🌳 Tree Style: %s\n", m.styleNames[m.currentStyle]))
	currentCursorStyle := m.cursorStyles[m.cursorStyle]
	debug.WriteString(fmt.Sprintf("🎯 Cursor Style: %s (%s)\n", currentCursorStyle.Name, currentCursorStyle.Description))

	// Show component-based configuration
	config := m.treeList.GetRenderConfig()
	debug.WriteString(fmt.Sprintf("   ShowCursor: %v, EnableBackground: %v\n",
		config.CursorConfig.Enabled, config.BackgroundConfig.Enabled))
	debug.WriteString(fmt.Sprintf("   CursorIndicator: %q, NormalSpacing: %q\n",
		config.CursorConfig.CursorIndicator, config.CursorConfig.NormalSpacing))

	if config.BackgroundConfig.Enabled {
		debug.WriteString(fmt.Sprintf("   Background Mode: %v\n", config.BackgroundConfig.Mode))
	}

	cascadingState := "OFF"
	if m.treeList.GetCascadingSelection() {
		cascadingState = "ON"
	}
	debug.WriteString(fmt.Sprintf("🔗 Cascading Selection: %s\n", cascadingState))

	// Show component order
	debug.WriteString(fmt.Sprintf("🧩 Component Order: %v\n", config.ComponentOrder))

	// Show currently loading chunks
	if len(m.loadingChunks) > 0 {
		debug.WriteString("⏳ Loading chunks: ")
		var chunks []string
		for chunk := range m.loadingChunks {
			chunks = append(chunks, fmt.Sprintf("%d", chunk))
		}
		debug.WriteString(strings.Join(chunks, ", ") + "\n")
	}

	// Show recent chunk history
	if len(m.chunkHistory) > 0 {
		debug.WriteString("📋 Recent chunk activity:\n")
		for _, entry := range m.chunkHistory {
			debug.WriteString("  " + entry + "\n")
		}
	}

	if len(m.loadingChunks) == 0 && len(m.chunkHistory) == 0 {
		debug.WriteString("💤 No chunk activity yet - navigate around to see chunking in action!\n")
	}

	return debug.String()
}
