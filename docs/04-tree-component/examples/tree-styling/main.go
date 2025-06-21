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

// Theme formatter functions - using proper component-based approach
func createDefaultFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
	return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
		if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
			content := flatItem.Item.String()

			// Selection styling (highest priority)
			if item.Selected {
				return lipgloss.NewStyle().
					Background(lipgloss.Color("12")). // Blue background
					Foreground(lipgloss.Color("15")). // White text
					Bold(true).
					Render(content)
			}

			// Content-based styling (no cursor handling - let TreeBackgroundComponent handle it)
			if flatItem.Item.IsFolder {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("12")). // Blue for folders
					Bold(true).
					Render(content)
			} else {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("10")). // Green for files
					Render(content)
			}
		}

		return fmt.Sprintf("%v", item.Item)
	}
}

func createContentOnlyFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
	return createDefaultFormatter() // Same as default - content-only cursor is handled by background config
}

func createDynamicCursorFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
	return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
		if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
			content := flatItem.Item.String()

			// Selection always gets full treatment
			if item.Selected {
				return lipgloss.NewStyle().
					Background(lipgloss.Color("12")).
					Foreground(lipgloss.Color("15")).
					Bold(true).
					Render(content)
			}

			// Different cursor styles for folders vs files
			if isCursor {
				if flatItem.Item.IsFolder {
					// Full-row cursor for folders - but we'll handle this via TreeBackgroundComponent
					// Just return styled content, TreeBackgroundComponent will apply full-row background
					return lipgloss.NewStyle().
						Foreground(lipgloss.Color("15")).
						Bold(true).
						Render(content)
				} else {
					// Content-only cursor for files - apply background directly
					return lipgloss.NewStyle().
						Background(lipgloss.Color("240")).
						Foreground(lipgloss.Color("15")).
						Bold(true).
						Render(content)
				}
			}

			// Regular styling...
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

func createDarkThemeFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
	return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
		if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
			content := flatItem.Item.String()

			// Selection styling (bright highlight)
			if item.Selected {
				return lipgloss.NewStyle().
					Background(lipgloss.Color("33")). // Bright blue
					Foreground(lipgloss.Color("0")).  // Black text
					Bold(true).
					Render(content)
			}

			// Dark theme colors
			if flatItem.Item.IsFolder {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("39")). // Bright cyan for folders
					Bold(true).
					Render(content)
			} else {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("46")). // Bright green for files
					Render(content)
			}
		}

		return fmt.Sprintf("%v", item.Item)
	}
}

func createProfessionalThemeFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
	return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
		if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
			content := flatItem.Item.String()

			// Professional selection styling
			if item.Selected {
				return lipgloss.NewStyle().
					Background(lipgloss.Color("153")). // Soft blue
					Foreground(lipgloss.Color("0")).   // Black text
					Render(content)
			}

			// Professional colors
			if flatItem.Item.IsFolder {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("24")). // Dark blue for folders
					Bold(true).
					Render(content)
			} else {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")). // Dark gray for files
					Render(content)
			}
		}

		return fmt.Sprintf("%v", item.Item)
	}
}

func createHighContrastFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
	return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
		if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
			content := flatItem.Item.String()

			// High contrast selection
			if item.Selected {
				return lipgloss.NewStyle().
					Background(lipgloss.Color("0")).  // Black background
					Foreground(lipgloss.Color("15")). // White text
					Bold(true).
					Underline(true).
					Render(content)
			}

			// High contrast colors
			if flatItem.Item.IsFolder {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("0")). // Black for folders
					Bold(true).
					Render(content)
			} else {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("8")). // Dark gray for files
					Render(content)
			}
		}

		return fmt.Sprintf("%v", item.Item)
	}
}

func createColorfulFormatter() func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string {
	return func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
		if flatItem, ok := item.Item.(tree.FlatTreeItem[FileItem]); ok {
			content := flatItem.Item.String()

			// Colorful selection
			if item.Selected {
				return lipgloss.NewStyle().
					Background(lipgloss.Color("201")). // Bright magenta
					Foreground(lipgloss.Color("15")).  // White text
					Bold(true).
					Render(content)
			}

			// Colorful theme
			if flatItem.Item.IsFolder {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("208")). // Orange for folders
					Bold(true).
					Italic(true).
					Render(content)
			} else {
				return lipgloss.NewStyle().
					Foreground(lipgloss.Color("82")). // Bright green for files
					Render(content)
			}
		}

		return fmt.Sprintf("%v", item.Item)
	}
}

// Theme configurations using proper component-based approach
type ThemeStyle struct {
	Name           string
	Formatter      func(core.Data[any], int, int, bool, bool, core.RenderContext, bool, bool, bool) string
	SymbolStyle    lipgloss.Style
	CursorStyle    lipgloss.Style
	Description    string
	CursorType     string                  // "content-only", "full-row", "dynamic"
	BackgroundMode tree.TreeBackgroundMode // How to apply cursor background
}

var themes = []ThemeStyle{
	{
		Name:           "Default",
		Formatter:      createDefaultFormatter(),
		SymbolStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		CursorStyle:    lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("15")),
		Description:    "Clean blue and green theme",
		CursorType:     "content-only",
		BackgroundMode: tree.TreeBackgroundContentOnly,
	},
	{
		Name:           "Full-Row",
		Formatter:      createContentOnlyFormatter(),
		SymbolStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		CursorStyle:    lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("15")),
		Description:    "Full-row cursor highlighting",
		CursorType:     "full-row",
		BackgroundMode: tree.TreeBackgroundEntireLine,
	},
	{
		Name:           "Dynamic",
		Formatter:      createDynamicCursorFormatter(),
		SymbolStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		CursorStyle:    lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("15")),
		Description:    "Different cursor styles by content type",
		CursorType:     "dynamic",
		BackgroundMode: tree.TreeBackgroundSelectiveComponents, // We'll specify which components get background
	},
	{
		Name:           "Dark",
		Formatter:      createDarkThemeFormatter(),
		SymbolStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Bold(true),
		CursorStyle:    lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color("15")).Bold(true),
		Description:    "High contrast dark theme",
		CursorType:     "content-only",
		BackgroundMode: tree.TreeBackgroundContentOnly,
	},
	{
		Name:           "Professional",
		Formatter:      createProfessionalThemeFormatter(),
		SymbolStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		CursorStyle:    lipgloss.NewStyle().Background(lipgloss.Color("250")).Foreground(lipgloss.Color("0")),
		Description:    "Subdued business theme",
		CursorType:     "content-only",
		BackgroundMode: tree.TreeBackgroundContentOnly,
	},
	{
		Name:           "High Contrast",
		Formatter:      createHighContrastFormatter(),
		SymbolStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Bold(true),
		CursorStyle:    lipgloss.NewStyle().Background(lipgloss.Color("7")).Foreground(lipgloss.Color("0")).Bold(true),
		Description:    "Maximum accessibility contrast",
		CursorType:     "content-only",
		BackgroundMode: tree.TreeBackgroundContentOnly,
	},
	{
		Name:           "Colorful",
		Formatter:      createColorfulFormatter(),
		SymbolStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("129")).Bold(true),
		CursorStyle:    lipgloss.NewStyle().Background(lipgloss.Color("93")).Foreground(lipgloss.Color("0")).Bold(true),
		Description:    "Vibrant colors and styling",
		CursorType:     "content-only",
		BackgroundMode: tree.TreeBackgroundContentOnly,
	},
}

// App wraps our tree component
type App struct {
	tree         *tree.TreeList[FileItem]
	status       string
	currentTheme int
	dataSource   *FileTreeDataSource
}

func (app *App) Init() tea.Cmd {
	return app.tree.Init()
}

func (app *App) applyTheme() {
	theme := themes[app.currentTheme]

	// Get current config
	treeConfig := app.tree.GetRenderConfig()

	// Apply theme formatting
	treeConfig.ContentConfig.Formatter = theme.Formatter
	treeConfig.TreeSymbolConfig.Style = theme.SymbolStyle

	// Apply cursor styling using proper component-based approach
	treeConfig.BackgroundConfig.Enabled = true
	treeConfig.BackgroundConfig.Style = theme.CursorStyle
	treeConfig.BackgroundConfig.Mode = theme.BackgroundMode

	// For dynamic cursor, configure which components get background
	if theme.CursorType == "dynamic" {
		// For dynamic cursor, we want to apply background to different components based on content
		// This would need custom logic in a specialized formatter or component
		treeConfig.BackgroundConfig.ApplyToComponents = []tree.TreeComponentType{
			tree.TreeComponentCursor,
			tree.TreeComponentIndentation,
			tree.TreeComponentTreeSymbol,
			tree.TreeComponentContent,
		}
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
		case "t":
			// Cycle through themes
			app.currentTheme = (app.currentTheme + 1) % len(themes)
			app.applyTheme()
			theme := themes[app.currentTheme]
			app.status = fmt.Sprintf("Theme: %s (%s) - %s", theme.Name, theme.CursorType, theme.Description)
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
	title := "🎨 Tree Styling & Cursor Demo"

	// Show current theme info
	currentTheme := themes[app.currentTheme]
	themeInfo := fmt.Sprintf("Theme: %s (%s) - %s", currentTheme.Name, currentTheme.CursorType, currentTheme.Description)

	help := "Navigate: ↑/↓/j/k, Enter: expand/collapse, Space: select, t: cycle themes, c: clear, q: quit"
	status := fmt.Sprintf("Status: %s", app.status)

	return fmt.Sprintf("%s\n%s\n\n%s\n\n%s\n%s",
		title,
		themeInfo,
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

	// Enable background styling for cursor items
	treeConfig.RenderConfig.BackgroundConfig.Enabled = true

	// Create the tree
	treeComponent := tree.NewTreeList(listConfig, treeConfig, dataSource)

	// Create the app
	app := &App{
		tree:         treeComponent,
		status:       "Ready! Press 't' to cycle through different visual themes",
		currentTheme: 0, // Start with default theme
		dataSource:   dataSource,
	}

	// Apply initial theme
	app.applyTheme()

	// Run the application
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
