package main

import (
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/tree"
	vtable "github.com/davidroman0O/vtable/pure"
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
	rootNodes     []vtable.TreeData[Task]
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
	var projects []vtable.TreeData[Task]

	// Generate 10 major projects
	for projectIdx := 0; projectIdx < 10; projectIdx++ {
		projectID := fmt.Sprintf("project-%d", projectIdx)
		projectTitle := fmt.Sprintf("📁 Project %d - %s", projectIdx+1, getProjectName(projectIdx))

		var modules []vtable.TreeData[Task]

		// Each project has 5-8 modules
		moduleCount := 5 + (projectIdx % 4)
		for moduleIdx := 0; moduleIdx < moduleCount; moduleIdx++ {
			moduleID := fmt.Sprintf("%s-module-%d", projectID, moduleIdx)
			moduleTitle := fmt.Sprintf("📦 Module %d - %s", moduleIdx+1, getModuleName(moduleIdx))

			var tasks []vtable.TreeData[Task]

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

				tasks = append(tasks, vtable.TreeData[Task]{
					ID:       taskID,
					Item:     task,
					Expanded: false,
					Children: nil,
				})
			}

			modules = append(modules, vtable.TreeData[Task]{
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

		projects = append(projects, vtable.TreeData[Task]{
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
func (ds *TreeTaskDataSource) GetRootNodes() []vtable.TreeData[Task] {
	return ds.rootNodes
}

func (ds *TreeTaskDataSource) GetItemByID(id string) (vtable.TreeData[Task], bool) {
	return ds.findNodeByID(ds.rootNodes, id)
}

func (ds *TreeTaskDataSource) findNodeByID(nodes []vtable.TreeData[Task], id string) (vtable.TreeData[Task], bool) {
	for _, node := range nodes {
		if node.ID == id {
			return node, true
		}
		if found, ok := ds.findNodeByID(node.Children, id); ok {
			return found, true
		}
	}
	return vtable.TreeData[Task]{}, false
}

func (ds *TreeTaskDataSource) SetSelected(id string, selected bool) tea.Cmd {
	if selected {
		ds.selectedNodes[id] = true
	} else {
		delete(ds.selectedNodes, id)
	}
	return vtable.SelectionResponseCmd(true, -1, id, selected, "toggle", nil, nil)
}

func (ds *TreeTaskDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	return ds.SetSelected(id, selected)
}

func (ds *TreeTaskDataSource) SelectAll() tea.Cmd {
	// Select all nodes in the tree
	ds.selectAllNodes(ds.rootNodes)
	return vtable.SelectionResponseCmd(true, -1, "", true, "selectAll", nil, nil)
}

func (ds *TreeTaskDataSource) selectAllNodes(nodes []vtable.TreeData[Task]) {
	for _, node := range nodes {
		ds.selectedNodes[node.ID] = true
		ds.selectAllNodes(node.Children)
	}
}

func (ds *TreeTaskDataSource) ClearSelection() tea.Cmd {
	ds.selectedNodes = make(map[string]bool)
	return vtable.SelectionResponseCmd(true, -1, "", false, "clear", nil, nil)
}

func (ds *TreeTaskDataSource) SelectRange(startID, endID string) tea.Cmd {
	// For trees, range selection is more complex - for now, just select both nodes
	ds.selectedNodes[startID] = true
	ds.selectedNodes[endID] = true
	return vtable.SelectionResponseCmd(true, -1, "", true, "range", nil, []string{startID, endID})
}

// TreeAppModel wraps our tree list
type TreeAppModel struct {
	treeList      *vtable.TreeList[Task]
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
	config := vtable.ListConfig{
		ViewportConfig: vtable.ViewportConfig{
			Height:             15, // Larger viewport for better tree viewing
			TopThreshold:       3,  // More threshold space for trees
			BottomThreshold:    3,
			ChunkSize:          20, // Larger chunks for better performance
			InitialIndex:       0,
			BoundingAreaBefore: 10, // Larger bounding area for smoother scrolling
			BoundingAreaAfter:  10,
		},
		SelectionMode: vtable.SelectionMultiple,
		KeyMap:        vtable.DefaultNavigationKeyMap(),
		MaxWidth:      120,
	}

	// Create tree configuration
	treeConfig := vtable.DefaultTreeConfig()

	// Create tree list
	treeList := vtable.NewTreeList(config, treeConfig, dataSource)

	// Define cursor style options
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
		styleNames:    []string{"Default", "Rounded"},
		cursorStyle:   0,
		cursorStyles:  cursorStyles,
	}

	// Run the program
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
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
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "enter":
			// Toggle current node
			return m, m.treeList.ToggleCurrentNode()

		case "r":
			// Force refresh to see chunk loading again
			m.statusMessage = "🔄 Refreshing tree data..."
			return m, vtable.DataRefreshCmd()

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
			currentStyling := m.treeList.GetEnableCursorStyling()
			m.treeList.SetEnableCursorStyling(!currentStyling)
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
			return m, vtable.JumpToStartCmd()
		case "G":
			return m, vtable.JumpToEndCmd()
		case "h":
			return m, vtable.PageUpCmd()
		case "l":
			return m, vtable.PageDownCmd()

		// Navigation keys - let TreeList handle these
		case "j", "down":
			return m, vtable.CursorDownCmd()
		case "k", "up":
			return m, vtable.CursorUpCmd()

		// Selection keys
		case " ":
			return m, vtable.SelectCurrentCmd()
		case "a":
			return m, vtable.SelectAllCmd()
		case "x":
			return m, vtable.SelectClearCmd()
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
	case vtable.ChunkLoadingStartedMsg:
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

	case vtable.ChunkLoadingCompletedMsg:
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

	case vtable.ChunkUnloadedMsg:
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
	case vtable.SelectionResponseMsg:
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
	case vtable.PageUpMsg, vtable.PageDownMsg, vtable.JumpToMsg, vtable.JumpToStartMsg, vtable.JumpToEndMsg:
		var cmd tea.Cmd
		_, cmd = m.treeList.Update(msg)
		state := m.treeList.GetState()
		m.statusMessage = fmt.Sprintf("🌳 Position: %d (Viewport: %d-%d)",
			state.CursorIndex+1,
			state.ViewportStartIndex,
			state.ViewportStartIndex+14)
		return m, cmd

	case vtable.CursorUpMsg, vtable.CursorDownMsg:
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
		m.treeList.SetEnumerator(tree.DefaultEnumerator)
	case 1: // Rounded
		m.treeList.SetEnumerator(tree.RoundedEnumerator)
	}
}

func (m *TreeAppModel) applyCursorStyle() {
	style := m.cursorStyles[m.cursorStyle]
	m.treeList.SetCursorIndicator(style.CursorIndicator)
	m.treeList.SetCursorSpacing(style.CursorSpacing)
	m.treeList.SetNormalSpacing(style.NormalSpacing)
	m.treeList.SetShowCursor(style.ShowCursor)
	m.treeList.SetEnableCursorStyling(style.EnableStyling)

	// Set background styling if enabled
	if style.EnableStyling && style.BackgroundColor != "" && style.ForegroundColor != "" {
		m.treeList.SetCursorStyle(style.ShowCursor, style.BackgroundColor, style.ForegroundColor)
	}
}

func (m TreeAppModel) View() string {
	var view strings.Builder

	// Show help if enabled
	if m.showHelp {
		view.WriteString(m.renderHelp())
		view.WriteString("\n")
	}

	// Show status message
	view.WriteString(m.statusMessage)
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
	help.WriteString("🌳 === LARGE TREE DEMO WITH CHUNKING ===\n")
	help.WriteString("Dataset: 10 projects × 5-8 modules × 8-15 tasks = ~500-1200 items total\n")
	help.WriteString("Visual: ► = cursor • ✅ = selected • Watch chunk loading in debug!\n")
	help.WriteString("Tree: Enter=expand/collapse • e=cycle tree styles • c=cycle cursor styles (9 options) • C=toggle cascading selection\n")
	help.WriteString("Cursor Styles: Arrow, Pointer, Star, Bullet, Bracket, Background Only, Subtle, Bright, None\n")
	help.WriteString("Styling: b=toggle background cursor styling on/off\n")
	help.WriteString("Navigation: j/k or ↑/↓ move • h/l page • g=start • G=end\n")
	help.WriteString("Selection: Space=toggle • a=select all • x=clear • s=show selection info\n")
	help.WriteString("Cascading: When ON, selecting a parent automatically selects all children\n")
	help.WriteString("Other: r=refresh • d=debug • ?=help • q=quit")
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
	debug.WriteString(fmt.Sprintf("   ShowCursor: %v, EnableStyling: %v\n", currentCursorStyle.ShowCursor, m.treeList.GetEnableCursorStyling()))
	if m.treeList.GetEnableCursorStyling() && currentCursorStyle.BackgroundColor != "" {
		debug.WriteString(fmt.Sprintf("   Background: %s, Foreground: %s\n", currentCursorStyle.BackgroundColor, currentCursorStyle.ForegroundColor))
	}
	cascadingState := "OFF"
	if m.treeList.GetCascadingSelection() {
		cascadingState = "ON"
	}
	debug.WriteString(fmt.Sprintf("🔗 Cascading Selection: %s\n", cascadingState))

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
