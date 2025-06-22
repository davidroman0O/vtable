package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/tree"
)

// FileItem represents our data structure for files and folders
type FileItem struct {
	Name     string
	IsFolder bool
}

func (f FileItem) String() string {
	if f.IsFolder {
		return "📁 " + f.Name
	}
	return "📄 " + f.Name
}

// FileTreeDataSource implements TreeDataSource for hierarchical file data
type FileTreeDataSource struct {
	rootNodes     []tree.TreeData[FileItem]
	selectedNodes map[string]bool
}

func NewFileTreeDataSource() *FileTreeDataSource {
	return &FileTreeDataSource{
		rootNodes: []tree.TreeData[FileItem]{
			// Project 1: Web Application
			{
				ID:   "webapp",
				Item: FileItem{Name: "Web Application", IsFolder: true},
				Children: []tree.TreeData[FileItem]{
					{
						ID:   "webapp_src",
						Item: FileItem{Name: "src", IsFolder: true},
						Children: []tree.TreeData[FileItem]{
							{
								ID:   "webapp_main",
								Item: FileItem{Name: "main.go", IsFolder: false},
							},
							{
								ID:   "webapp_app",
								Item: FileItem{Name: "app.go", IsFolder: false},
							},
							{
								ID:   "webapp_handlers",
								Item: FileItem{Name: "handlers", IsFolder: true},
								Children: []tree.TreeData[FileItem]{
									{
										ID:   "webapp_user_handler",
										Item: FileItem{Name: "user_handler.go", IsFolder: false},
									},
									{
										ID:   "webapp_auth_handler",
										Item: FileItem{Name: "auth_handler.go", IsFolder: false},
									},
									{
										ID:   "webapp_middleware",
										Item: FileItem{Name: "middleware.go", IsFolder: false},
									},
								},
							},
							{
								ID:   "webapp_models",
								Item: FileItem{Name: "models", IsFolder: true},
								Children: []tree.TreeData[FileItem]{
									{
										ID:   "webapp_user_model",
										Item: FileItem{Name: "user.go", IsFolder: false},
									},
									{
										ID:   "webapp_product_model",
										Item: FileItem{Name: "product.go", IsFolder: false},
									},
								},
							},
						},
					},
					{
						ID:   "webapp_tests",
						Item: FileItem{Name: "tests", IsFolder: true},
						Children: []tree.TreeData[FileItem]{
							{
								ID:   "webapp_unit_tests",
								Item: FileItem{Name: "unit_test.go", IsFolder: false},
							},
							{
								ID:   "webapp_integration_tests",
								Item: FileItem{Name: "integration_test.go", IsFolder: false},
							},
						},
					},
					{
						ID:   "webapp_config",
						Item: FileItem{Name: "config", IsFolder: true},
						Children: []tree.TreeData[FileItem]{
							{
								ID:   "webapp_env",
								Item: FileItem{Name: ".env", IsFolder: false},
							},
							{
								ID:   "webapp_yaml",
								Item: FileItem{Name: "config.yaml", IsFolder: false},
							},
						},
					},
				},
			},
			// Project 2: CLI Tool
			{
				ID:   "cli_tool",
				Item: FileItem{Name: "CLI Tool", IsFolder: true},
				Children: []tree.TreeData[FileItem]{
					{
						ID:   "cli_cmd",
						Item: FileItem{Name: "cmd", IsFolder: true},
						Children: []tree.TreeData[FileItem]{
							{
								ID:   "cli_root",
								Item: FileItem{Name: "root.go", IsFolder: false},
							},
							{
								ID:   "cli_version",
								Item: FileItem{Name: "version.go", IsFolder: false},
							},
							{
								ID:   "cli_config_cmd",
								Item: FileItem{Name: "config.go", IsFolder: false},
							},
						},
					},
					{
						ID:   "cli_internal",
						Item: FileItem{Name: "internal", IsFolder: true},
						Children: []tree.TreeData[FileItem]{
							{
								ID:   "cli_config",
								Item: FileItem{Name: "config.go", IsFolder: false},
							},
							{
								ID:   "cli_utils",
								Item: FileItem{Name: "utils", IsFolder: true},
								Children: []tree.TreeData[FileItem]{
									{
										ID:   "cli_helpers",
										Item: FileItem{Name: "helpers.go", IsFolder: false},
									},
									{
										ID:   "cli_logger",
										Item: FileItem{Name: "logger.go", IsFolder: false},
									},
								},
							},
						},
					},
				},
			},
			// Project 3: Documentation
			{
				ID:   "docs",
				Item: FileItem{Name: "Documentation", IsFolder: true},
				Children: []tree.TreeData[FileItem]{
					{
						ID:   "docs_guides",
						Item: FileItem{Name: "guides", IsFolder: true},
						Children: []tree.TreeData[FileItem]{
							{
								ID:   "docs_getting_started",
								Item: FileItem{Name: "getting-started.md", IsFolder: false},
							},
							{
								ID:   "docs_advanced",
								Item: FileItem{Name: "advanced.md", IsFolder: false},
							},
							{
								ID:   "docs_examples",
								Item: FileItem{Name: "examples", IsFolder: true},
								Children: []tree.TreeData[FileItem]{
									{
										ID:   "docs_basic_example",
										Item: FileItem{Name: "basic.md", IsFolder: false},
									},
									{
										ID:   "docs_advanced_example",
										Item: FileItem{Name: "advanced.md", IsFolder: false},
									},
								},
							},
						},
					},
					{
						ID:   "docs_api",
						Item: FileItem{Name: "api", IsFolder: true},
						Children: []tree.TreeData[FileItem]{
							{
								ID:   "docs_reference",
								Item: FileItem{Name: "reference.md", IsFolder: false},
							},
							{
								ID:   "docs_changelog",
								Item: FileItem{Name: "CHANGELOG.md", IsFolder: false},
							},
						},
					},
				},
			},
		},
		selectedNodes: make(map[string]bool),
	}
}

// Auto-expand configuration - now handled by TreeList directly
type AutoExpandConfig struct {
	ExpandRoot    bool     // Always expand root nodes
	ExpandFolders []string // Specific folder names to auto-expand
	MaxDepth      int      // Maximum depth to auto-expand
	ExpandEmpty   bool     // Auto-expand folders with no children
}

// Implement TreeDataSource interface
func (ds *FileTreeDataSource) GetRootNodes() []tree.TreeData[FileItem] {
	return ds.rootNodes
}

func (ds *FileTreeDataSource) GetItemByID(id string) (tree.TreeData[FileItem], bool) {
	return ds.findNodeByID(ds.rootNodes, id)
}

func (ds *FileTreeDataSource) findNodeByID(nodes []tree.TreeData[FileItem], id string) (tree.TreeData[FileItem], bool) {
	for _, node := range nodes {
		if node.ID == id {
			return node, true
		}
		if found, ok := ds.findNodeByID(node.Children, id); ok {
			return found, true
		}
	}
	return tree.TreeData[FileItem]{}, false
}

func (ds *FileTreeDataSource) SetSelected(id string, selected bool) tea.Cmd {
	if selected {
		ds.selectedNodes[id] = true
	} else {
		delete(ds.selectedNodes, id)
	}
	return core.SelectionResponseCmd(true, -1, id, selected, "toggle", nil, nil)
}

func (ds *FileTreeDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	return ds.SetSelected(id, selected)
}

func (ds *FileTreeDataSource) SelectAll() tea.Cmd {
	ds.selectAllNodes(ds.rootNodes)
	return core.SelectionResponseCmd(true, -1, "", true, "selectAll", nil, nil)
}

func (ds *FileTreeDataSource) selectAllNodes(nodes []tree.TreeData[FileItem]) {
	for _, node := range nodes {
		ds.selectedNodes[node.ID] = true
		ds.selectAllNodes(node.Children)
	}
}

func (ds *FileTreeDataSource) ClearSelection() tea.Cmd {
	ds.selectedNodes = make(map[string]bool)
	return core.SelectionResponseCmd(true, -1, "", false, "clear", nil, nil)
}

func (ds *FileTreeDataSource) SelectRange(startID, endID string) tea.Cmd {
	ds.selectedNodes[startID] = true
	ds.selectedNodes[endID] = true
	return core.SelectionResponseCmd(true, -1, "", true, "range", nil, []string{startID, endID})
}

// Advanced content formatter with item counts
func createAdvancedFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
	return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
		if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
			content := flatItem.Item.String()

			// Add visual indicator for folders with children
			if flatItem.Item.IsFolder && hasChildren {
				if isExpanded {
					content = content + " (expanded)"
				} else {
					content = content + " (...)"
				}
			}

			// Apply selection styling (highest priority)
			if item.Selected {
				return lipgloss.NewStyle().
					Background(lipgloss.Color("12")).
					Foreground(lipgloss.Color("15")).
					Bold(true).
					Render(content)
			}

			// Content styling
			if flatItem.Item.IsFolder {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("12")).
					Bold(true).
					Render(content)
			} else {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("10")).
					Render(content)
			}
		}

		return fmt.Sprintf("%v", item.Item)
	}
}

// App wraps our tree component with advanced features
type App struct {
	tree             *tree.TreeList[FileItem]
	status           string
	dataSource       *FileTreeDataSource
	autoExpandConfig AutoExpandConfig
}

func (app *App) Init() tea.Cmd {
	return app.tree.Init()
}

// Advanced navigation functions - now using TreeList methods directly
func (app *App) JumpToFirstChild() tea.Cmd {
	// Try to expand current node to show its children
	return app.tree.ToggleCurrentNode()
}

func (app *App) JumpToParent() tea.Cmd {
	// Navigate up in the tree
	return core.CursorUpCmd()
}

func (app *App) JumpToNextSibling() tea.Cmd {
	// Navigate down in the tree
	return core.CursorDownCmd()
}

func (app *App) JumpToPrevSibling() tea.Cmd {
	// Navigate up in the tree
	return core.CursorUpCmd()
}

func (app *App) ExpandCurrentSubtree() tea.Cmd {
	// Use TreeList's new method to expand entire subtree
	return app.tree.ExpandCurrentSubtree()
}

func (app *App) CollapseCurrentSubtree() tea.Cmd {
	// Use TreeList's new method to collapse entire subtree
	return app.tree.CollapseCurrentSubtree()
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return app, tea.Quit

		// Basic navigation
		case "enter":
			app.status = "Toggled expand/collapse"
			return app, app.tree.ToggleCurrentNode()
		case " ":
			app.status = "Toggled selection"
			return app, core.SelectCurrentCmd()

		// Advanced expand/collapse operations using TreeList methods
		case "E":
			app.status = "Expanded entire tree"
			return app, app.tree.ExpandAll()
		case "C":
			app.status = "Collapsed entire tree"
			return app, app.tree.CollapseAll()
		case "e":
			app.status = "Expanded current subtree"
			return app, app.ExpandCurrentSubtree()
		case "c":
			app.status = "Collapsed current subtree"
			return app, app.CollapseCurrentSubtree()

		// Smart navigation (simplified)
		case "right", "l":
			app.status = "Expanded/toggled current node"
			return app, app.JumpToFirstChild()
		case "left", "h":
			app.status = "Moved up (simulated parent jump)"
			return app, app.JumpToParent()
		case "n":
			app.status = "Moved down (simulated sibling jump)"
			return app, app.JumpToNextSibling()
		case "p":
			app.status = "Moved up (simulated sibling jump)"
			return app, app.JumpToPrevSibling()

		// Selection operations
		case "a":
			app.status = "Selected all items"
			return app, core.SelectAllCmd()
		case "x":
			app.status = "Cleared all selections"
			return app, core.SelectClearCmd()

		// Basic navigation
		case "up", "k":
			app.status = "Moved up"
			return app, core.CursorUpCmd()
		case "down", "j":
			app.status = "Moved down"
			return app, core.CursorDownCmd()
		case "pgup":
			app.status = "Page up"
			return app, core.PageUpCmd()
		case "pgdn":
			app.status = "Page down"
			return app, core.PageDownCmd()
		case "home", "g":
			app.status = "Jump to start"
			return app, core.JumpToStartCmd()
		case "end", "G":
			app.status = "Jump to end"
			return app, core.JumpToEndCmd()
		}
	}

	// Pass all other messages to the tree
	var cmd tea.Cmd
	_, cmd = app.tree.Update(msg)
	return app, cmd
}

func (app *App) View() string {
	title := "🌳 Advanced Tree Features Demo"

	// Show keyboard shortcuts
	shortcuts := "Advanced: E/C: expand/collapse all | e/c: subtree | h/l/n/p: smart nav | a/x: select"
	help := "Basic: ↑/↓/j/k: navigate | Enter: toggle | Space: select | q: quit"
	status := fmt.Sprintf("Status: %s", app.status)

	return fmt.Sprintf("%s\n%s\n\n%s\n\n%s\n%s",
		title,
		shortcuts,
		app.tree.View(),
		status,
		help)
}

// `04-tree-component/examples/advanced-features/main.go`
func main() {
	// Create the data source
	dataSource := NewFileTreeDataSource()

	// Configure the list component
	listConfig := core.ListConfig{
		ViewportConfig: core.ViewportConfig{
			Height:    15,
			ChunkSize: 20,
		},
		SelectionMode: core.SelectionMultiple,
		KeyMap:        core.DefaultNavigationKeyMap(),
	}

	// Configure tree with connected lines and styling
	treeConfig := tree.DefaultTreeConfig()
	treeConfig.RenderConfig.IndentationConfig.Enabled = true
	treeConfig.RenderConfig.IndentationConfig.UseConnectors = true
	treeConfig.RenderConfig.IndentationConfig.ConnectorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	// Enhanced content formatting with item counts
	treeConfig.RenderConfig.ContentConfig.Formatter = createAdvancedFormatter()

	// Background styling for cursor items
	treeConfig.RenderConfig.BackgroundConfig.Enabled = true
	treeConfig.RenderConfig.BackgroundConfig.Style = lipgloss.NewStyle().
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("15"))

	treeConfig.CascadingSelection = true

	// Create the tree
	treeComponent := tree.NewTreeList(listConfig, treeConfig, dataSource)

	// Auto-expand some nodes on startup using TreeList methods
	var autoExpandCommands []tea.Cmd
	// Auto-expand the first root node
	if len(dataSource.rootNodes) > 0 {
		autoExpandCommands = append(autoExpandCommands, treeComponent.ExpandNode(dataSource.rootNodes[0].ID))
	}

	// Create the app
	app := &App{
		tree:   treeComponent,
		status: "Advanced tree ready! Try E/C to expand/collapse all, e/c for subtrees",
	}

	// Run the application with auto-expand
	p := tea.NewProgram(app, tea.WithoutSignalHandler())

	// Apply auto-expand after starting
	go func() {
		for _, cmd := range autoExpandCommands {
			if cmd != nil {
				p.Send(cmd())
			}
		}
	}()

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
