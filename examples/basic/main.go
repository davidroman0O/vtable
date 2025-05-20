package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable"
)

// Example data providers
type StringListProvider struct {
	items []string
}

func NewStringListProvider(count int) *StringListProvider {
	items := make([]string, count)
	for i := 0; i < count; i++ {
		items[i] = fmt.Sprintf("Item %d", i)
	}
	return &StringListProvider{items: items}
}

func (p *StringListProvider) GetTotal() int {
	// In a real implementation, this would adjust based on filters
	return len(p.items)
}

func (p *StringListProvider) GetItems(request vtable.DataRequest) ([]string, error) {
	start := request.Start
	count := request.Count

	if start >= len(p.items) {
		return []string{}, nil
	}

	end := start + count
	if end > len(p.items) {
		end = len(p.items)
	}

	return p.items[start:end], nil
}

func (p *StringListProvider) FindItemIndex(key string, value any) (int, bool) {
	if key != "id" {
		return -1, false
	}

	// Try to parse as integer
	var id int
	switch v := value.(type) {
	case int:
		id = v
	case string:
		var err error
		id, err = strconv.Atoi(v)
		if err != nil {
			return -1, false
		}
	default:
		return -1, false
	}

	if id >= 0 && id < len(p.items) {
		return id, true
	}

	return -1, false
}

// Table data provider
type TableDataProvider struct {
	rows []vtable.TableRow
}

func NewTableDataProvider(count int) *TableDataProvider {
	rows := make([]vtable.TableRow, count)
	for i := 0; i < count; i++ {
		rows[i] = vtable.TableRow{
			Cells: []string{
				fmt.Sprintf("Row %d", i),
				fmt.Sprintf("Value %d", i*10),
				fmt.Sprintf("Description for item %d", i),
			},
		}
	}
	return &TableDataProvider{rows: rows}
}

func (p *TableDataProvider) GetTotal() int {
	// In a real implementation, this would adjust based on filters
	return len(p.rows)
}

func (p *TableDataProvider) GetItems(request vtable.DataRequest) ([]vtable.TableRow, error) {
	start := request.Start
	count := request.Count

	if start >= len(p.rows) {
		return []vtable.TableRow{}, nil
	}

	end := start + count
	if end > len(p.rows) {
		end = len(p.rows)
	}

	return p.rows[start:end], nil
}

func (p *TableDataProvider) FindItemIndex(key string, value any) (int, bool) {
	if key != "id" {
		return -1, false
	}

	// Try to parse as integer
	var id int
	switch v := value.(type) {
	case int:
		id = v
	case string:
		var err error
		id, err = strconv.Atoi(v)
		if err != nil {
			return -1, false
		}
	default:
		return -1, false
	}

	if id >= 0 && id < len(p.rows) {
		return id, true
	}

	return -1, false
}

// Main application
type appView int

const (
	viewList appView = iota
	viewTable
)

type Model struct {
	// Current view
	activeView appView

	// List component
	listModel *vtable.TeaList[string]

	// Table component
	tableModel *vtable.TeaTable

	// Search
	searchInput  textinput.Model
	searching    bool
	searchResult string

	// Display options
	debug bool

	// Theme
	currentTheme string
	themes       map[string]vtable.Theme

	// Currently pressed key (for help text highlighting)
	activeKey string

	// Terminal dimensions
	termWidth int
}

// Search result message
type searchResultMsg struct {
	found bool
	index int
}

// KeyReleasedMsg is sent when a key should no longer be highlighted
type KeyReleasedMsg struct{}

// SendKeyReleasedMsg creates a command that will reset the active key
func SendKeyReleasedMsg() tea.Cmd {
	return func() tea.Msg {
		return KeyReleasedMsg{}
	}
}

// Search command for list
func searchListCommand(list *vtable.TeaList[string], value string) tea.Cmd {
	return func() tea.Msg {
		found := list.JumpToItem("id", value)
		if found {
			index := list.GetState().CursorIndex
			return searchResultMsg{found: true, index: index}
		}
		return searchResultMsg{found: false, index: -1}
	}
}

// Search command for table
func searchTableCommand(table *vtable.TeaTable, value string) tea.Cmd {
	return func() tea.Msg {
		found := table.JumpToItem("id", value)
		if found {
			index := table.GetState().CursorIndex
			return searchResultMsg{found: true, index: index}
		}
		return searchResultMsg{found: false, index: -1}
	}
}

func initialModel() (Model, error) {
	// Configure common viewport settings
	viewportConfig := vtable.ViewportConfig{
		Height:               12,
		TopThresholdIndex:    2,
		BottomThresholdIndex: 9,
		ChunkSize:            20,
		InitialIndex:         0,
		Debug:                false,
	}

	// Setup different themes
	themes := make(map[string]vtable.Theme)
	themes["default"] = vtable.DefaultTheme()
	themes["dark"] = vtable.DarkTheme()
	themes["light"] = vtable.LightTheme()
	themes["colorful"] = vtable.ColorfulTheme()

	// Create theme with custom border characters
	roundedTheme := vtable.DefaultTheme()
	roundedTheme.BorderChars = vtable.RoundedBorderCharacters()
	themes["rounded"] = roundedTheme

	doubleTheme := vtable.ColorfulTheme()
	doubleTheme.BorderChars = vtable.DoubleBorderCharacters()
	themes["double"] = doubleTheme

	thickTheme := vtable.DarkTheme()
	thickTheme.BorderChars = vtable.ThickBorderCharacters()
	themes["thick"] = thickTheme

	currentTheme := "default"
	theme := themes[currentTheme]

	// Create list provider
	listProvider := NewStringListProvider(1000)

	// Create list formatter
	listFormatter := func(item string, index int, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		var style lipgloss.Style
		if isCursor {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Background(lipgloss.Color("63"))
		} else {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		}

		result := fmt.Sprintf("%d: %s", index, item)
		if isTopThreshold && isBottomThreshold {
			result = fmt.Sprintf("%s (T+B)", result)
		} else if isTopThreshold {
			result = fmt.Sprintf("%s (T)", result)
		} else if isBottomThreshold {
			result = fmt.Sprintf("%s (B)", result)
		}

		if isCursor {
			result = fmt.Sprintf("> %s", result)
		} else {
			result = fmt.Sprintf("  %s", result)
		}

		return style.Render(result)
	}

	// Create style config for backward compatibility
	styleConfig := vtable.ThemeToStyleConfig(theme)

	// Create list
	listModel, err := vtable.NewTeaList(viewportConfig, listProvider, styleConfig, listFormatter)
	if err != nil {
		return Model{}, err
	}

	// Create table provider
	tableProvider := NewTableDataProvider(1000)

	// Create table config
	tableConfig := vtable.TableConfig{
		Columns: []vtable.TableColumn{
			{Title: "ID", Width: 10, Alignment: vtable.AlignLeft},
			{Title: "Value", Width: 15, Alignment: vtable.AlignRight},
			{Title: "Description", Width: 30, Alignment: vtable.AlignLeft},
		},
		ShowHeader:     true,
		ShowBorders:    true,
		ViewportConfig: viewportConfig,
	}

	// Create table
	tableModel, err := vtable.NewTeaTable(tableConfig, tableProvider, theme)
	if err != nil {
		return Model{}, err
	}

	// Create search input
	ti := textinput.New()
	ti.Placeholder = "Enter ID to jump to"
	ti.Width = 30
	ti.Prompt = "Jump to ID: "

	// Create model with list view initially active
	return Model{
		activeView:   viewList,
		listModel:    listModel,
		tableModel:   tableModel,
		searchInput:  ti,
		searching:    false,
		searchResult: "",
		debug:        false,
		currentTheme: currentTheme,
		themes:       themes,
		activeKey:    "", // No key is initially active
		termWidth:    0,  // Initialize termWidth
	}, nil
}

func (m Model) Init() tea.Cmd {
	// Request the window size on initialization
	cmds := []tea.Cmd{
		textinput.Blink,
	}

	// Add a command to get window size
	cmds = append(cmds, func() tea.Msg {
		return tea.WindowSizeMsg{
			Width:  0, // We don't know the size yet, the terminal will fill it in
			Height: 0,
		}
	})

	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Store the terminal width for our help text
		m.termWidth = msg.Width
		return m, nil
	case tea.KeyMsg:
		// If we're searching, handle search input
		if m.searching {
			// Set the active key for highlighting purposes (except for text entry)
			keyStr := msg.String()
			if keyStr == "enter" || keyStr == "esc" {
				m.activeKey = keyStr
				cmds = append(cmds, tea.Tick(200*time.Millisecond, func(_ time.Time) tea.Msg {
					return KeyReleasedMsg{}
				}))
			} else {
				// Don't highlight regular typing keys during search
				m.activeKey = ""
			}

			switch msg.String() {
			case "enter":
				// Perform the search when Enter is pressed
				value := m.searchInput.Value()
				m.searching = false
				m.searchResult = "" // Clear previous result

				// Search in the appropriate view
				if m.activeView == viewList {
					return m, searchListCommand(m.listModel, value)
				} else {
					return m, searchTableCommand(m.tableModel, value)
				}
			case "esc":
				// Cancel search on Escape
				m.searching = false
				m.searchResult = ""
				return m, nil
			default:
				// Update the search input
				var inputCmd tea.Cmd
				m.searchInput, inputCmd = m.searchInput.Update(msg)
				return m, inputCmd
			}
		}

		// Store the pressed key for highlighting
		m.activeKey = msg.String()
		cmds = append(cmds, tea.Tick(200*time.Millisecond, func(_ time.Time) tea.Msg {
			return KeyReleasedMsg{}
		}))

		// Regular key handlers when not searching
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "D":
			m.debug = !m.debug
		case "m":
			// Example of customizing keymaps
			if m.activeView == viewList {
				// Get current keymap
				currentKeyMap := m.listModel.GetKeyMap()

				// Create custom keymap based on current one
				customKeyMap := currentKeyMap

				// Modify some bindings - for example, swap up and down keys
				upKeys := customKeyMap.Up.Keys()
				downKeys := customKeyMap.Down.Keys()

				customKeyMap.Up = key.NewBinding(
					key.WithKeys(downKeys...),
					key.WithHelp("↑", "up (customized)"),
				)

				customKeyMap.Down = key.NewBinding(
					key.WithKeys(upKeys...),
					key.WithHelp("↓", "down (customized)"),
				)

				// Apply the custom keymap
				m.listModel.SetKeyMap(customKeyMap)
				m.searchResult = "Custom keymap applied to list - up/down keys are swapped!"
			} else {
				// Do the same for table
				currentKeyMap := m.tableModel.GetKeyMap()

				// Create custom keymap based on current one
				customKeyMap := currentKeyMap

				// Modify some bindings
				customKeyMap.PageUp = key.NewBinding(
					key.WithKeys("u", "b", "space"), // Add space as PageUp
					key.WithHelp("space/u/b", "page up (customized)"),
				)

				customKeyMap.PageDown = key.NewBinding(
					key.WithKeys("d", "enter"), // Change to use enter for PageDown
					key.WithHelp("enter/d", "page down (customized)"),
				)

				// Apply the custom keymap
				m.tableModel.SetKeyMap(customKeyMap)
				m.searchResult = "Custom keymap applied to table - PageUp/PageDown modified!"
			}
		case "tab":
			// Simply toggle between views - each maintains its own state
			if m.activeView == viewList {
				m.activeView = viewTable
			} else {
				m.activeView = viewList
			}
		case "t":
			// Cycle through themes
			themeKeys := []string{"default", "dark", "light", "colorful", "rounded", "double", "thick"}
			currentIndex := 0
			for i, key := range themeKeys {
				if key == m.currentTheme {
					currentIndex = i
					break
				}
			}
			nextIndex := (currentIndex + 1) % len(themeKeys)
			m.currentTheme = themeKeys[nextIndex]
			newTheme := m.themes[m.currentTheme]

			// Update theme for Table and List WITHOUT recreating them
			// This preserves cursor position exactly
			if m.activeView == viewTable {
				// For table, update the theme directly
				m.tableModel.SetTheme(newTheme)
				m.searchResult = fmt.Sprintf("Theme changed to %s (Table view)", m.currentTheme)
			} else {
				// For list, convert theme to style and update
				styleConfig := vtable.ThemeToStyleConfig(newTheme)
				m.listModel.SetStyle(styleConfig)
				m.searchResult = fmt.Sprintf("Theme changed to %s (List view)", m.currentTheme)
			}
		default:
			// Check if this is a search key based on current model's key bindings
			var isSearchKey bool
			if m.activeView == viewList {
				isSearchKey = key.Matches(msg, m.listModel.GetKeyMap().Search)
			} else {
				isSearchKey = key.Matches(msg, m.tableModel.GetKeyMap().Search)
			}

			if isSearchKey {
				// For 'f' key specifically, ensure it gets highlighted
				// before entering search mode
				if msg.String() == "f" || msg.String() == "/" || msg.String() == "slash" {
					// Start the search mode after a brief delay to allow the key highlight to be visible
					cmds = append(cmds, tea.Tick(100*time.Millisecond, func(_ time.Time) tea.Msg {
						// This custom message will trigger search mode
						return StartSearchMsg{}
					}))
					return m, tea.Batch(cmds...)
				}
			}
		}
	case StartSearchMsg:
		// Start search mode
		m.searching = true
		m.searchInput.SetValue("")
		m.searchInput.Focus()
		// Clear the highlight after entering search mode
		m.activeKey = ""
		return m, textinput.Blink
	case KeyReleasedMsg:
		// Reset the active key
		m.activeKey = ""
		return m, nil
	case searchResultMsg:
		if msg.found {
			m.searchResult = fmt.Sprintf("Found at index %d", msg.index)
		} else {
			m.searchResult = "Not found"
		}
		return m, nil
	}

	// If we're searching, we don't want to update the components
	if m.searching {
		return m, nil
	}

	// Update only the active component
	if m.activeView == viewList {
		// We're in list view - update only the list model
		var cmd tea.Cmd
		listModel, cmd := m.listModel.Update(msg)
		if listM, ok := listModel.(*vtable.TeaList[string]); ok {
			m.listModel = listM
		}
		cmds = append(cmds, cmd)
	} else {
		// We're in table view - update only the table model
		var cmd tea.Cmd
		tableModel, cmd := m.tableModel.Update(msg)
		if tableM, ok := tableModel.(*vtable.TeaTable); ok {
			m.tableModel = tableM
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var sb strings.Builder

	// Add a title based on current view
	var title string
	if m.activeView == viewList {
		title = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Render(fmt.Sprintf("Virtualized List Example - Theme: %s (TAB to switch)", m.currentTheme))
	} else {
		title = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Render(fmt.Sprintf("Virtualized Table Example - Theme: %s (TAB to switch)", m.currentTheme))
	}

	sb.WriteString(title)
	sb.WriteString("\n\n")

	// If searching, show the search input
	if m.searching {
		searchBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Render(m.searchInput.View())
		sb.WriteString(searchBox)
		sb.WriteString("\n\n")
	}

	// If we have a search result, show it
	if m.searchResult != "" {
		result := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Render(m.searchResult)
		sb.WriteString(result)
		sb.WriteString("\n\n")
	}

	// Render only the active component
	if m.activeView == viewList {
		sb.WriteString(m.listModel.View())
	} else {
		sb.WriteString(m.tableModel.View())
	}

	// Add debug info if enabled
	if m.debug {
		sb.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true).
			Render("\n\n=== Debug Information ===\n"))

		if m.activeView == viewList {
			state := m.listModel.GetState()
			sb.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#5F87FF")).
				Bold(true).
				Render("\nList State:\n"))

			sb.WriteString(fmt.Sprintf("Selected Item: %d\n", state.CursorIndex))
			sb.WriteString(fmt.Sprintf("Viewport Start: %d\n", state.ViewportStartIndex))
			sb.WriteString(fmt.Sprintf("Cursor Viewport Index: %d\n", state.CursorViewportIndex))
			sb.WriteString(fmt.Sprintf("At Top Threshold: %t\n", state.IsAtTopThreshold))
			sb.WriteString(fmt.Sprintf("At Bottom Threshold: %t\n", state.IsAtBottomThreshold))
			sb.WriteString(fmt.Sprintf("At Dataset Start: %t\n", state.AtDatasetStart))
			sb.WriteString(fmt.Sprintf("At Dataset End: %t\n", state.AtDatasetEnd))
		} else {
			state := m.tableModel.GetState()
			sb.WriteString(lipgloss.NewStyle().
				Foreground(lipgloss.Color("#5F87FF")).
				Bold(true).
				Render("\nTable State:\n"))

			sb.WriteString(fmt.Sprintf("Selected Item: %d\n", state.CursorIndex))
			sb.WriteString(fmt.Sprintf("Viewport Start: %d\n", state.ViewportStartIndex))
			sb.WriteString(fmt.Sprintf("Cursor Viewport Index: %d\n", state.CursorViewportIndex))
			sb.WriteString(fmt.Sprintf("At Top Threshold: %t\n", state.IsAtTopThreshold))
			sb.WriteString(fmt.Sprintf("At Bottom Threshold: %t\n", state.IsAtBottomThreshold))
			sb.WriteString(fmt.Sprintf("At Dataset Start: %t\n", state.AtDatasetStart))
			sb.WriteString(fmt.Sprintf("At Dataset End: %t\n", state.AtDatasetEnd))
		}
	}

	// Add help text with active key highlighted
	sb.WriteString("\n\n")
	sb.WriteString(m.renderHelpText())

	return sb.String()
}

// renderHelpText creates the help text with the currently pressed key highlighted
func (m Model) renderHelpText() string {
	// Regular style for inactive keys
	regularKeyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	// Highlighted style for keys shown in help but not pressed
	highlightedKeyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF9900")).
		Bold(true)

	// Active style for the currently pressed key
	activeKeyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Background(lipgloss.Color("#FFFF00")).
		Bold(true)

	// Helper function to style a key
	styleKey := func(key string) string {
		if key == m.activeKey {
			return activeKeyStyle.Render(key)
		}
		return highlightedKeyStyle.Render(key)
	}

	// All help text items
	helpItems := []string{
		fmt.Sprintf("%s/%s: navigate", styleKey("j"), styleKey("k")),
		fmt.Sprintf("%s: search", styleKey("f")),
		fmt.Sprintf("%s: quit", styleKey("q")),
		fmt.Sprintf("%s/%s: page up/down", styleKey("u"), styleKey("d")),
		fmt.Sprintf("%s/%s: top/bottom", styleKey("g"), styleKey("G")),
		fmt.Sprintf("%s: switch view", styleKey("tab")),
		fmt.Sprintf("%s: cycle themes", styleKey("t")),
		fmt.Sprintf("%s: keymap customize", styleKey("m")),
		fmt.Sprintf("%s: toggle debug", styleKey("D")),
	}

	// Get current terminal width (or use default)
	width := m.termWidth
	if width <= 0 {
		width = 80 // reasonable default
	}

	// Format help text with explicit line breaks for cleaner display
	var lines []string
	currentLine := ""
	separator := " • "

	for _, item := range helpItems {
		// Check if adding this item would exceed the line width
		testLine := currentLine
		if len(currentLine) > 0 {
			testLine += separator
		}
		testLine += item

		// If we would exceed width, start a new line
		if lipgloss.Width(testLine) > width && len(currentLine) > 0 {
			lines = append(lines, currentLine)
			currentLine = item
		} else {
			// Otherwise add to current line with separator if needed
			if len(currentLine) > 0 {
				currentLine += separator
			}
			currentLine += item
		}
	}

	// Add the last line if not empty
	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}

	// Join lines with carriage returns
	helpText := strings.Join(lines, "\n")

	return regularKeyStyle.Render(helpText)
}

// StartSearchMsg is a message to start search mode after key highlight is shown
type StartSearchMsg struct{}

func main() {
	model, err := initialModel()
	if err != nil {
		fmt.Printf("Error creating model: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
