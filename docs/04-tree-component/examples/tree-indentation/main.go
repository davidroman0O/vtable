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
			// Project 3: API Service
			{
				ID:   "api_service",
				Item: FileItem{Name: "API Service", IsFolder: true},
				Children: []tree.TreeData[FileItem]{
					{
						ID:   "api_endpoints",
						Item: FileItem{Name: "endpoints", IsFolder: true},
						Children: []tree.TreeData[FileItem]{
							{
								ID:   "api_users",
								Item: FileItem{Name: "users.go", IsFolder: false},
							},
							{
								ID:   "api_products",
								Item: FileItem{Name: "products.go", IsFolder: false},
							},
						},
					},
					{
						ID:   "api_middleware",
						Item: FileItem{Name: "middleware", IsFolder: true},
						Children: []tree.TreeData[FileItem]{
							{
								ID:   "api_auth_middleware",
								Item: FileItem{Name: "auth.go", IsFolder: false},
							},
							{
								ID:   "api_cors_middleware",
								Item: FileItem{Name: "cors.go", IsFolder: false},
							},
						},
					},
				},
			},
		},
		selectedNodes: make(map[string]bool),
	}
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

// Content formatter that supports depth indicators for deep trees
func createDepthAwareFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
	return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
		if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
			content := flatItem.Item.String()

			// Add depth indicators for very deep items (level 3+)
			var depthIndicator string
			if depth > 2 {
				depthIndicator = fmt.Sprintf("[L%d] ", depth)
			}

			// Apply selection styling (highest priority)
			if item.Selected {
				return lipgloss.NewStyle().
					Background(lipgloss.Color("12")). // Blue background
					Foreground(lipgloss.Color("15")). // White text
					Bold(true).
					Render(depthIndicator + content)
			}

			// Content-based styling
			if flatItem.Item.IsFolder {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("12")). // Blue for folders
					Bold(true).
					Render(depthIndicator + content)
			} else {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("10")). // Green for files
					Render(depthIndicator + content)
			}
		}

		return fmt.Sprintf("%v", item.Item)
	}
}

// Indentation themes configuration
type IndentationTheme struct {
	Name         string
	IndentString string
	IndentSize   int
	Style        lipgloss.Style
	Description  string
}

var indentationThemes = []IndentationTheme{
	{
		Name:         "Minimal",
		IndentString: "",
		IndentSize:   2,
		Style:        lipgloss.NewStyle(),
		Description:  "Clean 2-space indentation",
	},
	{
		Name:         "Spacious",
		IndentString: "",
		IndentSize:   4,
		Style:        lipgloss.NewStyle(),
		Description:  "Wide 4-space indentation for clarity",
	},
	{
		Name:         "Compact",
		IndentString: "",
		IndentSize:   1,
		Style:        lipgloss.NewStyle(),
		Description:  "Minimal 1-space indentation for dense trees",
	},
	{
		Name:         "Dotted",
		IndentString: "··",
		IndentSize:   0,
		Style:        lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Description:  "Gray dots show hierarchy clearly",
	},
	{
		Name:         "Dashed",
		IndentString: "- ",
		IndentSize:   0,
		Style:        lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		Description:  "Dashes for distinctive hierarchy",
	},
	{
		Name:         "Arrows",
		IndentString: "→ ",
		IndentSize:   0,
		Style:        lipgloss.NewStyle().Foreground(lipgloss.Color("242")),
		Description:  "Arrow indicators pointing to content",
	},
	{
		Name:         "Bullets",
		IndentString: "• ",
		IndentSize:   0,
		Style:        lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		Description:  "Bullet points for each level",
	},
	{
		Name:         "Boxed",
		IndentString: "│ ",
		IndentSize:   0,
		Style:        lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Bold(true),
		Description:  "Box-drawing characters for structure",
	},
}

// App wraps our tree component
type App struct {
	tree               *tree.TreeList[FileItem]
	status             string
	currentIndentation int
	dataSource         *FileTreeDataSource
}

func (app *App) Init() tea.Cmd {
	return app.tree.Init()
}

func (app *App) applyIndentationTheme() {
	theme := indentationThemes[app.currentIndentation]

	// Get current config
	treeConfig := app.tree.GetRenderConfig()

	// Apply indentation theme
	treeConfig.IndentationConfig.IndentString = theme.IndentString
	treeConfig.IndentationConfig.IndentSize = theme.IndentSize
	treeConfig.IndentationConfig.Style = theme.Style

	// Apply the updated config
	app.tree.SetRenderConfig(treeConfig)
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return app, tea.Quit
		case "enter":
			app.status = "Toggled expand/collapse"
			return app, app.tree.ToggleCurrentNode()
		case " ":
			app.status = "Toggled selection"
			return app, core.SelectCurrentCmd()
		case "i":
			// Cycle through indentation themes
			app.currentIndentation = (app.currentIndentation + 1) % len(indentationThemes)
			app.applyIndentationTheme()
			theme := indentationThemes[app.currentIndentation]
			app.status = fmt.Sprintf("Indentation: %s - %s", theme.Name, theme.Description)
			return app, nil
		case "c":
			app.status = "Cleared all selections"
			return app, core.SelectClearCmd()

		// Navigation commands
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
	title := "🌳 Tree Indentation Demo"

	// Show current indentation info
	currentTheme := indentationThemes[app.currentIndentation]
	indentInfo := fmt.Sprintf("Indentation: %s - %s", currentTheme.Name, currentTheme.Description)

	help := "Navigate: ↑/↓/j/k, Enter: expand/collapse, Space: select, i: cycle indentation, c: clear, q: quit"
	status := fmt.Sprintf("Status: %s", app.status)

	return fmt.Sprintf("%s\n%s\n\n%s\n\n%s\n%s",
		title,
		indentInfo,
		app.tree.View(),
		status,
		help)
}

// `04-tree-component/examples/tree-indentation/main.go`
func main() {
	// Create the data source
	dataSource := NewFileTreeDataSource()

	// Configure the list component
	listConfig := core.ListConfig{
		ViewportConfig: core.ViewportConfig{
			Height:    12,
			ChunkSize: 20,
		},
		SelectionMode: core.SelectionMultiple,
		KeyMap:        core.DefaultNavigationKeyMap(),
	}

	// Start with default tree configuration
	treeConfig := tree.DefaultTreeConfig()

	// Enable indentation and configure basic styling
	treeConfig.RenderConfig.IndentationConfig.Enabled = true
	treeConfig.RenderConfig.ContentConfig.Formatter = createDepthAwareFormatter()

	// Enable background styling for cursor items
	treeConfig.RenderConfig.BackgroundConfig.Enabled = true
	treeConfig.RenderConfig.BackgroundConfig.Style = lipgloss.NewStyle().
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("15"))

	// Create the tree
	treeComponent := tree.NewTreeList(listConfig, treeConfig, dataSource)

	// Create the app
	app := &App{
		tree:               treeComponent,
		status:             "Ready! Press 'i' to cycle through indentation styles",
		currentIndentation: 0, // Start with minimal theme
		dataSource:         dataSource,
	}

	// Apply initial indentation theme
	app.applyIndentationTheme()

	// Run the application
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
