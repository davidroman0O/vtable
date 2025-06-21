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

// Symbol style configurations
type SymbolStyle struct {
	Name            string
	ExpandedSymbol  string
	CollapsedSymbol string
	LeafSymbol      string
	SymbolSpacing   string
	ShowForLeaves   bool
	Style           lipgloss.Style
	Description     string
}

var symbolStyles = []SymbolStyle{
	{
		Name:            "Default",
		ExpandedSymbol:  "▼",
		CollapsedSymbol: "▶",
		LeafSymbol:      "•",
		SymbolSpacing:   " ",
		ShowForLeaves:   true,
		Style:           lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Description:     "VTable default arrow style",
	},
	{
		Name:            "Plus/Minus",
		ExpandedSymbol:  "-",
		CollapsedSymbol: "+",
		LeafSymbol:      " ",
		SymbolSpacing:   " ",
		ShowForLeaves:   true,
		Style:           lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true),
		Description:     "Classic file explorer style",
	},
	{
		Name:            "Boxed",
		ExpandedSymbol:  "[-]",
		CollapsedSymbol: "[+]",
		LeafSymbol:      "   ",
		SymbolSpacing:   " ",
		ShowForLeaves:   false,
		Style:           lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true),
		Description:     "Boxed symbols for clarity",
	},
	{
		Name:            "Unicode",
		ExpandedSymbol:  "◢",
		CollapsedSymbol: "◤",
		LeafSymbol:      "◦",
		SymbolSpacing:   " ",
		ShowForLeaves:   true,
		Style:           lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true),
		Description:     "Modern Unicode symbols",
	},
	{
		Name:            "Emoji",
		ExpandedSymbol:  "📂",
		CollapsedSymbol: "📁",
		LeafSymbol:      "📄",
		SymbolSpacing:   " ",
		ShowForLeaves:   true,
		Style:           lipgloss.NewStyle(),
		Description:     "Emoji folder/file theme",
	},
	{
		Name:            "Minimal",
		ExpandedSymbol:  "",
		CollapsedSymbol: "",
		LeafSymbol:      "",
		SymbolSpacing:   "",
		ShowForLeaves:   false,
		Style:           lipgloss.NewStyle(),
		Description:     "No symbols, clean look",
	},
}

// Custom formatter to display clean file names with selection styling
func fileTreeFormatter(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
	// Extract the FileItem from the data
	if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
		content := flatItem.Item.String() // Use our FileItem's String() method

		// Apply selection styling if item is selected
		if item.Selected {
			return lipgloss.NewStyle().
				Background(lipgloss.Color("12")). // Blue background for selected
				Foreground(lipgloss.Color("15")). // White text
				Render(content)
		}

		return content
	}

	// Fallback to default formatting
	return fmt.Sprintf("%v", item.Item)
}

// App wraps our tree component
type App struct {
	tree         *tree.TreeList[FileItem]
	status       string
	currentStyle int
	dataSource   *FileTreeDataSource
	listConfig   core.ListConfig
}

func (app *App) Init() tea.Cmd {
	return app.tree.Init()
}

func (app *App) applySymbolStyle() {
	style := symbolStyles[app.currentStyle]

	// Get current config
	treeConfig := app.tree.GetRenderConfig()

	// Apply symbol configuration
	if style.Name == "Minimal" {
		treeConfig.TreeSymbolConfig.Enabled = false
	} else {
		treeConfig.TreeSymbolConfig.Enabled = true
		treeConfig.TreeSymbolConfig.ExpandedSymbol = style.ExpandedSymbol
		treeConfig.TreeSymbolConfig.CollapsedSymbol = style.CollapsedSymbol
		treeConfig.TreeSymbolConfig.LeafSymbol = style.LeafSymbol
		treeConfig.TreeSymbolConfig.SymbolSpacing = style.SymbolSpacing
		treeConfig.TreeSymbolConfig.ShowForLeaves = style.ShowForLeaves
		treeConfig.TreeSymbolConfig.Style = style.Style
	}

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
		case "s":
			// Cycle through symbol styles
			app.currentStyle = (app.currentStyle + 1) % len(symbolStyles)
			app.applySymbolStyle()
			style := symbolStyles[app.currentStyle]
			app.status = fmt.Sprintf("Switched to %s style - %s", style.Name, style.Description)
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
	title := "🌳 Tree Symbol Styles Demo"

	// Show current style info
	currentStyle := symbolStyles[app.currentStyle]
	styleInfo := fmt.Sprintf("Current Style: %s - %s", currentStyle.Name, currentStyle.Description)

	help := "Navigate: ↑/↓/j/k, Enter: expand/collapse, Space: select, s: cycle symbols, c: clear, q: quit"
	status := fmt.Sprintf("Status: %s", app.status)

	return fmt.Sprintf("%s\n%s\n\n%s\n\n%s\n%s",
		title,
		styleInfo,
		app.tree.View(),
		status,
		help)
}

func main() {
	// Create the data source
	dataSource := NewFileTreeDataSource()

	// Configure the list component
	listConfig := core.ListConfig{
		ViewportConfig: core.ViewportConfig{
			Height:    10,
			ChunkSize: 20,
		},
		SelectionMode: core.SelectionMultiple,
		KeyMap:        core.DefaultNavigationKeyMap(),
	}

	// Start with default tree configuration
	treeConfig := tree.DefaultTreeConfig()
	treeConfig.RenderConfig.ContentConfig.Formatter = fileTreeFormatter

	// Enable background styling for cursor items
	treeConfig.RenderConfig.BackgroundConfig.Enabled = true
	treeConfig.RenderConfig.BackgroundConfig.Style = lipgloss.NewStyle().
		Background(lipgloss.Color("240")). // Gray background for cursor
		Foreground(lipgloss.Color("15"))   // White text

	// Create the tree
	treeComponent := tree.NewTreeList(listConfig, treeConfig, dataSource)

	// Create the app
	app := &App{
		tree:         treeComponent,
		status:       "Ready! Press 's' to cycle through different symbol styles",
		currentStyle: 0, // Start with default style
		dataSource:   dataSource,
		listConfig:   listConfig,
	}

	// Apply initial symbol style
	app.applySymbolStyle()

	// Run the application
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
