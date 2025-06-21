package main

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/table"
)

// Debug mode constants
const (
	DebugOff      = 0 // No debug info
	DebugBasic    = 1 // Basic operations
	DebugDetailed = 2 // Detailed timing
	DebugVerbose  = 3 // Full message logs
)

// Debug and observability structures
type ActivityLog struct {
	Timestamp time.Time
	Type      string        // "user", "system", "performance"
	Action    string        // "navigation", "sort", "filter", "chunk_load"
	Details   string        // Specific information
	Duration  time.Duration // For performance tracking
}

type PerformanceMetrics struct {
	ChunkLoadTime   time.Duration
	FilterTime      time.Duration
	SortTime        time.Duration
	RenderTime      time.Duration
	MemoryUsage     int64
	ActiveChunks    int
	TotalOperations int
}

type ChunkState struct {
	StartIndex    int
	Size          int
	LoadStartTime time.Time
	LoadEndTime   time.Time
	Status        string // "loading", "loaded", "unloaded"
	AccessCount   int
}

type MessageLog struct {
	Timestamp time.Time
	Type      string
	Message   string
	Data      interface{}
}

// SimpleDataSource with debug capabilities - BASED ON WORKING EXAMPLE
type SimpleDataSource struct {
	totalItems     int
	data           []core.TableRow
	selectedItems  map[string]bool
	recentActivity []string
	// Sorting and filtering state
	sortFields    []string
	sortDirs      []string
	filters       map[string]any
	filteredData  []core.TableRow
	filteredTotal int
	// Debug tracking
	operationStartTimes map[string]time.Time
}

func NewSimpleDataSource() *SimpleDataSource {
	totalItems := 50
	data := make([]core.TableRow, totalItems)

	// Create simple items like the working example
	for i := 0; i < totalItems; i++ {
		data[i] = core.TableRow{
			ID: fmt.Sprintf("row-%d", i),
			Cells: []string{
				fmt.Sprintf("Item %d", i+1), // Just the item name
			},
		}
	}

	ds := &SimpleDataSource{
		totalItems:          totalItems,
		data:                data,
		selectedItems:       make(map[string]bool),
		recentActivity:      make([]string, 0),
		sortFields:          []string{},
		sortDirs:            []string{},
		filters:             make(map[string]any),
		filteredData:        data, // Start with all data
		filteredTotal:       totalItems,
		operationStartTimes: make(map[string]time.Time),
	}

	return ds
}

// GetTotal returns the total number of items
func (ds *SimpleDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return core.DataTotalMsg{Total: ds.filteredTotal}
	}
}

// RefreshTotal refreshes the total count
func (ds *SimpleDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

// LoadChunk loads a chunk of data - EXACTLY LIKE WORKING EXAMPLE
func (ds *SimpleDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		// Update sorting and filtering if changed in request
		if len(request.SortFields) > 0 || len(request.Filters) > 0 {
			// Apply new sort/filter settings
			if len(request.SortFields) > 0 {
				ds.sortFields = request.SortFields
				ds.sortDirs = request.SortDirections
			}
			if len(request.Filters) > 0 {
				ds.filters = request.Filters
			}
			ds.rebuildFilteredData()
		}

		// Simulate loading delay
		time.Sleep(10 * time.Millisecond)

		start := request.Start
		end := start + request.Count
		if end > ds.filteredTotal {
			end = ds.filteredTotal
		}

		var items []core.Data[any]
		for i := start; i < end; i++ {
			if i < len(ds.filteredData) {
				items = append(items, core.Data[any]{
					ID:       ds.filteredData[i].ID,
					Item:     ds.filteredData[i],
					Selected: ds.selectedItems[ds.filteredData[i].ID],
					Metadata: core.NewTypedMetadata(),
				})
			}
		}

		return core.DataChunkLoadedMsg{
			StartIndex: start,
			Items:      items,
			Request:    request,
		}
	}
}

// SetSelected sets the selection state of an item
func (ds *SimpleDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		if index >= 0 && index < len(ds.data) {
			id := ds.data[index].ID

			if selected {
				ds.selectedItems[id] = true
				ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected: %s", ds.data[index].Cells[0]))
			} else {
				delete(ds.selectedItems, id)
				ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Deselected: %s", ds.data[index].Cells[0]))
			}

			// Keep only last 10 activities
			if len(ds.recentActivity) > 10 {
				ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
			}

			return core.SelectionResponseMsg{
				Success:   true,
				Index:     index,
				ID:        id,
				Selected:  selected,
				Operation: "toggle",
			}
		}

		return core.SelectionResponseMsg{
			Success:   false,
			Index:     index,
			ID:        "",
			Selected:  false,
			Operation: "toggle",
			Error:     fmt.Errorf("invalid index: %d", index),
		}
	}
}

// SetSelectedByID sets the selection state of an item by ID
func (ds *SimpleDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	return func() tea.Msg {
		for i, row := range ds.data {
			if row.ID == id {
				if selected {
					ds.selectedItems[id] = true
					ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected: %s", row.Cells[0]))
				} else {
					delete(ds.selectedItems, id)
					ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Deselected: %s", row.Cells[0]))
				}

				if len(ds.recentActivity) > 10 {
					ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
				}

				return core.SelectionResponseMsg{
					Success:   true,
					Index:     i,
					ID:        id,
					Selected:  selected,
					Operation: "toggle",
				}
			}
		}

		return core.SelectionResponseMsg{
			Success:   false,
			Index:     -1,
			ID:        id,
			Selected:  false,
			Operation: "toggle",
			Error:     fmt.Errorf("item not found: %s", id),
		}
	}
}

// ClearSelection clears all selections
func (ds *SimpleDataSource) ClearSelection() tea.Cmd {
	return func() tea.Msg {
		count := len(ds.selectedItems)
		ds.selectedItems = make(map[string]bool)
		ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Cleared %d selections", count))

		if len(ds.recentActivity) > 10 {
			ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
		}

		return core.SelectionResponseMsg{
			Success:   true,
			Index:     -1,
			ID:        "",
			Selected:  false,
			Operation: "clear",
		}
	}
}

// SelectAll selects all items
func (ds *SimpleDataSource) SelectAll() tea.Cmd {
	return func() tea.Msg {
		for _, row := range ds.data {
			ds.selectedItems[row.ID] = true
		}

		ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected all %d items", len(ds.data)))

		if len(ds.recentActivity) > 10 {
			ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
		}

		return core.SelectionResponseMsg{
			Success:   true,
			Index:     -1,
			ID:        "",
			Selected:  true,
			Operation: "selectAll",
		}
	}
}

// SelectRange selects a range of items
func (ds *SimpleDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	return func() tea.Msg {
		var affectedIDs []string
		count := 0

		for i := startIndex; i <= endIndex && i < len(ds.data); i++ {
			ds.selectedItems[ds.data[i].ID] = true
			affectedIDs = append(affectedIDs, ds.data[i].ID)
			count++
		}

		ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected range: %d items", count))

		if len(ds.recentActivity) > 10 {
			ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
		}

		return core.SelectionResponseMsg{
			Success:     true,
			Index:       startIndex,
			ID:          "",
			Selected:    true,
			Operation:   "range",
			AffectedIDs: affectedIDs,
		}
	}
}

// GetItemID returns the ID for a given item
func (ds *SimpleDataSource) GetItemID(item any) string {
	if row, ok := item.(core.TableRow); ok {
		return row.ID
	}
	return ""
}

// Helper methods for debug info
func (ds *SimpleDataSource) GetRecentActivity() []string {
	return ds.recentActivity
}

func (ds *SimpleDataSource) GetSelectionCount() int {
	return len(ds.selectedItems)
}

// Filtering and sorting methods
func (ds *SimpleDataSource) SetFilter(field string, value any) {
	ds.operationStartTimes["filter"] = time.Now()
	ds.filters[field] = value
	ds.rebuildFilteredData()
}

func (ds *SimpleDataSource) ClearFilter(field string) {
	ds.operationStartTimes["filter"] = time.Now()
	delete(ds.filters, field)
	ds.rebuildFilteredData()
}

func (ds *SimpleDataSource) ClearAllFilters() {
	ds.operationStartTimes["filter"] = time.Now()
	ds.filters = make(map[string]any)
	ds.rebuildFilteredData()
}

func (ds *SimpleDataSource) SetSort(fields []string, directions []string) {
	ds.operationStartTimes["sort"] = time.Now()
	ds.sortFields = fields
	ds.sortDirs = directions
	ds.rebuildFilteredData()
}

func (ds *SimpleDataSource) ClearSort() {
	ds.operationStartTimes["sort"] = time.Now()
	ds.sortFields = []string{}
	ds.sortDirs = []string{}
	ds.rebuildFilteredData()
}

func (ds *SimpleDataSource) rebuildFilteredData() {
	start := time.Now()

	// Start with all data
	result := make([]core.TableRow, 0, len(ds.data))

	// Apply filters
	for _, row := range ds.data {
		include := true

		for field, filterValue := range ds.filters {
			switch field {
			case "search":
				if filterStr, ok := filterValue.(string); ok {
					searchTerm := strings.ToLower(filterStr)
					if !strings.Contains(strings.ToLower(row.Cells[0]), searchTerm) {
						include = false
						break
					}
				}
			}
		}

		if include {
			result = append(result, row)
		}
	}

	// Apply sorting
	if len(ds.sortFields) > 0 {
		sort.Slice(result, func(i, j int) bool {
			for idx, field := range ds.sortFields {
				dir := "asc"
				if idx < len(ds.sortDirs) {
					dir = ds.sortDirs[idx]
				}

				var cellI, cellJ string
				switch field {
				case "name":
					if len(result[i].Cells) > 0 {
						cellI = result[i].Cells[0]
					}
					if len(result[j].Cells) > 0 {
						cellJ = result[j].Cells[0]
					}
				}

				var cmp int
				if cellI < cellJ {
					cmp = -1
				} else if cellI > cellJ {
					cmp = 1
				}

				if cmp != 0 {
					if dir == "desc" {
						return cmp > 0
					}
					return cmp < 0
				}
			}
			return false
		})
	}

	ds.filteredData = result
	ds.filteredTotal = len(result)

	// Track rebuild time
	duration := time.Since(start)
	if duration > time.Millisecond {
		fmt.Printf("Data rebuild took: %v\n", duration)
	}
}

type AppModel struct {
	table      *table.Table
	dataSource *SimpleDataSource

	// Core state
	statusMessage  string
	activeFilters  map[string]bool
	currentSort    string
	currentSortDir string
	searchMode     bool
	searchTerm     string
	searchActive   bool

	// Debug and observability
	debugMode          int
	showDebugOverlay   bool
	activityLog        []ActivityLog
	performanceMetrics PerformanceMetrics
	chunkStates        map[int]ChunkState
	messageLog         []MessageLog

	// Timing tracking
	lastOperationStart time.Time
	operationTimings   map[string][]time.Duration
}

func main() {
	// Create data source
	dataSource := NewSimpleDataSource()

	// Create table configuration - EXACTLY LIKE WORKING EXAMPLE
	columns := []core.TableColumn{
		{
			Title:           "Item",
			Field:           "name",
			Width:           25,
			Alignment:       core.AlignLeft,
			HeaderAlignment: core.AlignLeft,
		},
	}

	theme := core.Theme{
		HeaderStyle:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("57")),
		CellStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		CursorStyle:        lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("21")),
		SelectedStyle:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("57")),
		FullRowCursorStyle: lipgloss.NewStyle().Background(lipgloss.Color("21")).Foreground(lipgloss.Color("15")),
		BorderChars: core.BorderChars{
			Horizontal: "─", Vertical: "│", TopLeft: "┌", TopRight: "┐",
			BottomLeft: "└", BottomRight: "┘", TopT: "┬", BottomT: "┴",
			LeftT: "├", RightT: "┤", Cross: "┼",
		},
		BorderColor: "8",
	}

	config := core.TableConfig{
		Columns:     columns,
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: core.ViewportConfig{
			Height:             8, // Smaller to leave room for debug info
			TopThreshold:       2,
			BottomThreshold:    2,
			ChunkSize:          5, // Smaller chunks to see more loading activity
			InitialIndex:       0,
			BoundingAreaBefore: 25,
			BoundingAreaAfter:  25,
		},
		Theme:                       theme,
		SelectionMode:               core.SelectionNone,
		ActiveCellIndicationEnabled: true,
	}

	// Create table with data source - EXACTLY LIKE WORKING EXAMPLE
	tbl := table.NewTable(config, dataSource)

	// CRITICAL: Focus the table so it can receive key events
	tbl.Focus()

	model := AppModel{
		table:            tbl,
		dataSource:       dataSource,
		statusMessage:    "Debug & Observability Demo - 50 items loaded - Press d for debug modes, D for overlay",
		activeFilters:    make(map[string]bool),
		debugMode:        DebugBasic, // Start with basic debug enabled
		chunkStates:      make(map[int]ChunkState),
		operationTimings: make(map[string][]time.Duration),
	}

	// Initialize performance metrics
	model.updateMemoryUsage()

	// Run the interactive program without alt screen
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
	}
}

func (m AppModel) Init() tea.Cmd {
	m.logActivity("system", "init", "Application started")

	var cmds []tea.Cmd
	cmds = append(cmds, m.table.Init())
	cmds = append(cmds, m.table.Focus())
	cmds = append(cmds, core.DataRefreshCmd()) // Trigger initial data load

	return tea.Batch(cmds...)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Log all messages if in verbose debug mode
	if m.debugMode >= DebugVerbose {
		m.logMessage(msg)
	}

	// Intercept specific messages for debug tracking but don't interfere
	switch msg := msg.(type) {
	case core.ChunkLoadingStartedMsg:
		m.chunkStates[msg.ChunkStart] = ChunkState{
			StartIndex:    msg.ChunkStart,
			Size:          msg.Request.Count,
			LoadStartTime: time.Now(),
			Status:        "loading",
		}
		m.logActivity("system", "chunk_load_start",
			fmt.Sprintf("Chunk %d (size: %d)", msg.ChunkStart, msg.Request.Count))

	case core.ChunkLoadingCompletedMsg:
		if state, exists := m.chunkStates[msg.ChunkStart]; exists {
			state.LoadEndTime = time.Now()
			state.Status = "loaded"
			state.AccessCount++
			m.chunkStates[msg.ChunkStart] = state

			duration := state.LoadEndTime.Sub(state.LoadStartTime)
			m.performanceMetrics.ChunkLoadTime = duration
			m.logActivity("system", "chunk_load_complete",
				fmt.Sprintf("Chunk %d loaded in %v", msg.ChunkStart, duration))
		}

	case core.ChunkUnloadedMsg:
		if state, exists := m.chunkStates[msg.ChunkStart]; exists {
			state.Status = "unloaded"
			m.chunkStates[msg.ChunkStart] = state
			m.logActivity("system", "chunk_unload",
				fmt.Sprintf("Chunk %d unloaded", msg.ChunkStart))
		}

	case core.DataRefreshMsg:
		m.lastOperationStart = time.Now()
		m.logActivity("system", "data_refresh", "Data refresh triggered")

	case core.DataTotalMsg:
		m.logActivity("system", "data_total", fmt.Sprintf("Total: %d items", msg.Total))

	case core.DataChunkLoadedMsg:
		m.logActivity("system", "chunk_data_loaded",
			fmt.Sprintf("Chunk %d: %d items loaded", msg.StartIndex, len(msg.Items)))

	case core.SelectionResponseMsg:
		m.logActivity("system", "selection",
			fmt.Sprintf("Selection %s: index %d", msg.Operation, msg.Index))

	case core.CursorUpMsg, core.CursorDownMsg:
		m.logActivity("user", "navigation", "Cursor movement")

	case core.PageUpMsg, core.PageDownMsg:
		m.logActivity("user", "navigation", "Page movement")

	case tea.KeyMsg:
		// Handle search mode first
		if m.searchMode {
			switch msg.String() {
			case "enter":
				m.searchMode = false
				if m.searchTerm != "" {
					m.dataSource.SetFilter("search", m.searchTerm)
					m.searchActive = true
					m.statusMessage = fmt.Sprintf("Searching for: %s (%d results)", m.searchTerm, m.dataSource.filteredTotal)
					m.logActivity("user", "search", fmt.Sprintf("Search: %s", m.searchTerm))
					return m, core.DataRefreshCmd()
				} else {
					m.statusMessage = "Search cancelled"
					return m, nil
				}
			case "escape":
				m.searchMode = false
				m.searchTerm = ""
				m.statusMessage = "Search cancelled"
				m.logActivity("user", "search", "Search cancelled")
				return m, nil
			case "backspace":
				if len(m.searchTerm) > 0 {
					m.searchTerm = m.searchTerm[:len(m.searchTerm)-1]
				}
				return m, nil
			default:
				if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
					m.searchTerm += msg.String()
				}
				return m, nil
			}
		}

		// Track user input
		m.trackUserInput(msg.String())

		// Normal key handling
		switch msg.String() {
		case "q", "ctrl+c":
			m.logActivity("user", "quit", "Application exit")
			return m, tea.Quit

		// ===================
		// DEBUG CONTROLS
		// ===================
		case "d":
			m.debugMode = (m.debugMode + 1) % 4
			m.statusMessage = fmt.Sprintf("Debug mode: %s", m.getDebugModeName())
			m.logActivity("user", "debug_mode", fmt.Sprintf("Changed to: %s", m.getDebugModeName()))

		case "D":
			m.showDebugOverlay = !m.showDebugOverlay
			if m.showDebugOverlay {
				m.statusMessage = "Debug overlay enabled"
			} else {
				m.statusMessage = "Debug overlay disabled"
			}
			m.logActivity("user", "debug_overlay", fmt.Sprintf("Overlay: %v", m.showDebugOverlay))

		case "ctrl+r":
			// Reset all debug data
			m.activityLog = []ActivityLog{}
			m.messageLog = []MessageLog{}
			m.chunkStates = make(map[int]ChunkState)
			m.operationTimings = make(map[string][]time.Duration)
			m.performanceMetrics = PerformanceMetrics{}
			m.statusMessage = "Debug data reset"
			m.logActivity("user", "debug_reset", "All debug data cleared")

		// ===================
		// TABLE CONTROLS
		// ===================
		case "s":
			return m.sortByActiveColumn()

		case "S":
			return m.clearSorting()

		case "1":
			return m.handleNumberFilter("1")

		case "0":
			return m.clearAllFilters()

		case "/":
			return m.enterSearchMode()

		case "j", "down":
			return m, core.CursorDownCmd()

		case "k", "up":
			return m, core.CursorUpCmd()
		}
	}

	// ALWAYS pass ALL messages to the table - let it handle everything normally
	var cmd tea.Cmd
	_, cmd = m.table.Update(msg)

	// Update performance metrics after table processes the message
	m.updateMemoryUsage()
	m.performanceMetrics.ActiveChunks = len(m.chunkStates)

	return m, cmd
}

// ===================
// DEBUG HELPER METHODS
// ===================

func (m *AppModel) logActivity(activityType, action, details string) {
	activity := ActivityLog{
		Timestamp: time.Now(),
		Type:      activityType,
		Action:    action,
		Details:   details,
	}

	m.activityLog = append(m.activityLog, activity)

	// Keep only last 50 activities
	if len(m.activityLog) > 50 {
		m.activityLog = m.activityLog[len(m.activityLog)-50:]
	}

	m.performanceMetrics.TotalOperations++
}

func (m *AppModel) logMessage(msg tea.Msg) {
	msgLog := MessageLog{
		Timestamp: time.Now(),
		Type:      fmt.Sprintf("%T", msg),
		Message:   fmt.Sprintf("%+v", msg),
		Data:      msg,
	}

	m.messageLog = append(m.messageLog, msgLog)

	// Keep only last 20 messages
	if len(m.messageLog) > 20 {
		m.messageLog = m.messageLog[len(m.messageLog)-20:]
	}
}

func (m *AppModel) trackUserInput(key string) {
	var action string
	switch key {
	case "s":
		action = "sort_column"
	case "1":
		action = "toggle_filter"
	case "0":
		action = "clear_filters"
	case "/":
		action = "search_mode"
	case "up", "down", "j", "k":
		action = "row_navigation"
	case "d", "D":
		action = "debug_control"
	default:
		action = "other_input"
	}

	m.logActivity("user", action, fmt.Sprintf("Key: %s", key))
}

func (m *AppModel) updateMemoryUsage() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.performanceMetrics.MemoryUsage = int64(memStats.Alloc)
}

func (m *AppModel) getDebugModeName() string {
	switch m.debugMode {
	case DebugOff:
		return "Off"
	case DebugBasic:
		return "Basic"
	case DebugDetailed:
		return "Detailed"
	case DebugVerbose:
		return "Verbose"
	default:
		return "Unknown"
	}
}

// ===================
// TABLE OPERATION METHODS
// ===================

func (m AppModel) sortByActiveColumn() (tea.Model, tea.Cmd) {
	start := time.Now()

	// Only one column - always sort by name
	field := "name"
	columnName := "Item"

	if m.currentSort == field {
		if m.currentSortDir == "asc" {
			m.currentSortDir = "desc"
			m.statusMessage = fmt.Sprintf("Sort: %s (Z→A)", columnName)
		} else {
			m.currentSort = ""
			m.currentSortDir = ""
			m.dataSource.ClearSort()
			m.statusMessage = "Sorting cleared"
			duration := time.Since(start)
			m.logActivity("user", "sort_clear", fmt.Sprintf("Duration: %v", duration))
			return m, core.DataRefreshCmd()
		}
	} else {
		m.currentSort = field
		m.currentSortDir = "asc"
		m.statusMessage = fmt.Sprintf("Sort: %s (A→Z)", columnName)
	}

	m.dataSource.SetSort([]string{field}, []string{m.currentSortDir})
	duration := time.Since(start)
	m.logActivity("user", "sort_column", fmt.Sprintf("Field: %s, Dir: %s, Duration: %v", field, m.currentSortDir, duration))
	return m, core.DataRefreshCmd()
}

func (m AppModel) handleNumberFilter(key string) (tea.Model, tea.Cmd) {
	// Simplified - just enable search for demo
	if key == "1" {
		return m.enterSearchMode()
	}
	return m, nil
}

func (m AppModel) clearAllFilters() (tea.Model, tea.Cmd) {
	start := time.Now()
	m.activeFilters = make(map[string]bool)
	m.searchActive = false
	m.dataSource.ClearAllFilters()
	m.statusMessage = fmt.Sprintf("All filters cleared (%d results)", m.dataSource.filteredTotal)
	duration := time.Since(start)
	m.logActivity("user", "clear_all_filters", fmt.Sprintf("Duration: %v", duration))
	return m, core.DataRefreshCmd()
}

func (m AppModel) clearSorting() (tea.Model, tea.Cmd) {
	start := time.Now()
	m.currentSort = ""
	m.currentSortDir = ""
	m.dataSource.ClearSort()
	m.statusMessage = "Sorting cleared - original order"
	duration := time.Since(start)
	m.logActivity("user", "clear_sorting", fmt.Sprintf("Duration: %v", duration))
	return m, core.DataRefreshCmd()
}

func (m AppModel) enterSearchMode() (tea.Model, tea.Cmd) {
	m.searchMode = true
	m.searchTerm = ""
	m.statusMessage = "Search mode: Type to filter data, Enter to apply, Esc to cancel"
	m.logActivity("user", "search_mode", "Entered search mode")
	return m, nil
}

// ===================
// VIEW RENDERING
// ===================

func (m AppModel) View() string {
	var view strings.Builder

	// Main interface
	view.WriteString(m.renderMainInterface())

	// Debug information based on mode
	if m.debugMode > DebugOff {
		view.WriteString("\n" + m.renderDebugInfo())
	}

	// Debug overlay
	if m.showDebugOverlay {
		view.WriteString("\n" + m.renderDebugOverlay())
	}

	return view.String()
}

func (m *AppModel) renderMainInterface() string {
	var view strings.Builder

	view.WriteString("=== DEBUG & OBSERVABILITY DEMO ===\n")

	// Current state with debug info - simplified for one column
	view.WriteString(fmt.Sprintf("Data: %d/%d | Sort: %s | Debug: %s\n",
		m.dataSource.filteredTotal,
		m.dataSource.totalItems,
		m.getSortDescription(),
		m.getDebugModeName(),
	))

	// Controls - simplified
	view.WriteString("Controls: s=sort S=clear-sort | 1=search /=search | d=debug D=overlay | Ctrl+R=reset | ↑↓=navigate | q=quit\n")

	// Status or search prompt
	if m.searchMode {
		view.WriteString(fmt.Sprintf("🔍 Search: %s_\n", m.searchTerm))
	} else {
		view.WriteString(fmt.Sprintf("Status: %s\n", m.statusMessage))
	}

	view.WriteString("\n")

	// Table
	view.WriteString(m.table.View())

	return view.String()
}

func (m *AppModel) renderDebugInfo() string {
	var debug strings.Builder

	debug.WriteString("=== DEBUG INFO ===\n")
	debug.WriteString(fmt.Sprintf("Mode: %s | Operations: %d | Memory: %s | Active Chunks: %d\n",
		m.getDebugModeName(),
		m.performanceMetrics.TotalOperations,
		m.formatBytes(m.performanceMetrics.MemoryUsage),
		len(m.chunkStates)))

	if m.debugMode >= DebugBasic {
		debug.WriteString("\nRecent Activity:\n")
		debug.WriteString(m.renderActivityLog())
	}

	if m.debugMode >= DebugDetailed {
		debug.WriteString("\n\nChunk States:\n")
		debug.WriteString(m.renderChunkStates())
	}

	if m.debugMode >= DebugVerbose {
		debug.WriteString("\n\nMessage Log:\n")
		debug.WriteString(m.renderMessageLog())
	}

	return debug.String()
}

func (m *AppModel) renderDebugOverlay() string {
	return fmt.Sprintf("=== OVERLAY === Ops: %d | Mem: %s | Chunks: %d | Mode: %s",
		m.performanceMetrics.TotalOperations,
		m.formatBytes(m.performanceMetrics.MemoryUsage),
		len(m.chunkStates),
		m.getDebugModeName())
}

func (m *AppModel) renderActivityLog() string {
	if len(m.activityLog) == 0 {
		return "No recent activity"
	}

	var activities []string
	count := len(m.activityLog)
	start := 0
	if count > 8 {
		start = count - 8
	}

	for i := start; i < count; i++ {
		activity := m.activityLog[i]
		timeStr := activity.Timestamp.Format("15:04:05")
		activities = append(activities,
			fmt.Sprintf("%s [%s] %s: %s",
				timeStr, activity.Type, activity.Action, activity.Details))
	}

	return strings.Join(activities, "\n")
}

func (m *AppModel) renderChunkStates() string {
	if len(m.chunkStates) == 0 {
		return "No chunks loaded"
	}

	var chunks []string
	for start, state := range m.chunkStates {
		status := state.Status
		if state.Status == "loaded" && !state.LoadEndTime.IsZero() {
			duration := state.LoadEndTime.Sub(state.LoadStartTime)
			status = fmt.Sprintf("loaded (%v)", duration)
		}

		chunks = append(chunks,
			fmt.Sprintf("Chunk %d: %s (size: %d, accessed: %d)",
				start, status, state.Size, state.AccessCount))
	}

	return strings.Join(chunks, "\n")
}

func (m *AppModel) renderMessageLog() string {
	if len(m.messageLog) == 0 {
		return "No messages logged"
	}

	var messages []string
	count := len(m.messageLog)
	start := 0
	if count > 5 {
		start = count - 5
	}

	for i := start; i < count; i++ {
		msg := m.messageLog[i]
		timeStr := msg.Timestamp.Format("15:04:05.000")
		messages = append(messages,
			fmt.Sprintf("%s %s", timeStr, msg.Type))
	}

	return strings.Join(messages, "\n")
}

func (m *AppModel) getSortDescription() string {
	if m.currentSort == "" {
		return "None"
	}

	direction := "↑"
	if m.currentSortDir == "desc" {
		direction = "↓"
	}

	fieldNames := map[string]string{
		"name": "Item",
	}

	if name, exists := fieldNames[m.currentSort]; exists {
		return fmt.Sprintf("%s%s", name, direction)
	}

	return "Custom"
}

func (m *AppModel) getActiveFiltersDescription() string {
	if !m.searchActive {
		return "None"
	}
	return "Search"
}

func (m *AppModel) formatBytes(bytes int64) string {
	const kb = 1024
	const mb = kb * 1024

	if bytes < kb {
		return fmt.Sprintf("%dB", bytes)
	} else if bytes < mb {
		return fmt.Sprintf("%.1fKB", float64(bytes)/kb)
	} else {
		return fmt.Sprintf("%.1fMB", float64(bytes)/mb)
	}
}
