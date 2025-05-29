package vtable

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ================================
// TABLE MODEL IMPLEMENTATION
// ================================

// Table represents a pure Tea table component that reuses the List infrastructure
type Table struct {
	// Core state - reuse List infrastructure
	dataSource DataSource[any]
	chunks     map[int]Chunk[any] // Map of start index to chunk
	totalItems int

	// Viewport state - same as List
	viewport ViewportState

	// Configuration
	config TableConfig

	// Table-specific configuration
	columns []TableColumn

	// Rendering
	cellFormatters      map[int]CellFormatter // Column index -> formatter
	rowFormatter        RowFormatter
	headerFormatter     HeaderFormatter
	headerCellFormatter HeaderCellFormatter
	loadingFormatter    LoadingRowFormatter
	renderContext       RenderContext

	// Selection state
	selectedItems map[string]bool
	selectedOrder []string

	// Focus state
	focused bool

	lastError error

	// Filtering and sorting
	filters     map[string]any
	sortFields  []string
	sortDirs    []string
	searchQuery string
	searchField string

	// Search results
	searchResults []int

	// visibleItems is the slice of Data items currently visible in the viewport
	visibleItems []Data[any]

	// Chunk access tracking for LRU management
	chunkAccessTime map[int]time.Time

	// Loading state tracking
	loadingChunks    map[int]bool
	hasLoadingChunks bool
	canScroll        bool
}

// ================================
// CONSTRUCTOR
// ================================

// NewTable creates a new Table with the given configuration and data source
func NewTable(config TableConfig, dataSource DataSource[any]) *Table {
	// Validate and fix config
	errors := ValidateTableConfig(&config)
	if len(errors) > 0 {
		FixTableConfig(&config)
	}

	table := &Table{
		dataSource:       dataSource,
		chunks:           make(map[int]Chunk[any]),
		config:           config,
		columns:          config.Columns,
		cellFormatters:   make(map[int]CellFormatter),
		selectedItems:    make(map[string]bool),
		selectedOrder:    make([]string, 0),
		filters:          make(map[string]any),
		chunkAccessTime:  make(map[int]time.Time),
		visibleItems:     make([]Data[any], 0),
		loadingChunks:    make(map[int]bool),
		hasLoadingChunks: false,
		canScroll:        true,
		viewport: ViewportState{
			ViewportStartIndex:  0,
			CursorIndex:         config.ViewportConfig.InitialIndex,
			CursorViewportIndex: 0,
			IsAtTopThreshold:    false,
			IsAtBottomThreshold: false,
			AtDatasetStart:      true,
			AtDatasetEnd:        false,
		},
	}

	// Set up render context
	table.setupRenderContext()

	return table
}

// ================================
// TEA MODEL INTERFACE
// ================================

// Init initializes the table model
func (t *Table) Init() tea.Cmd {
	return t.loadInitialData()
}

// Update handles all messages and updates the table state
func (t *Table) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// ===== Lifecycle Messages =====
	case InitMsg:
		return t, t.Init()

	case DestroyMsg:
		return t, nil

	case ResetMsg:
		t.reset()
		return t, t.Init()

	// ===== Navigation Messages - Reuse List logic =====
	case CursorUpMsg:
		cmd := t.handleCursorUp()
		return t, cmd

	case CursorDownMsg:
		cmd := t.handleCursorDown()
		return t, cmd

	case PageUpMsg:
		cmd := t.handlePageUp()
		return t, cmd

	case PageDownMsg:
		cmd := t.handlePageDown()
		return t, cmd

	case JumpToStartMsg:
		cmd := t.handleJumpToStart()
		return t, cmd

	case JumpToEndMsg:
		cmd := t.handleJumpToEnd()
		return t, cmd

	case JumpToMsg:
		cmd := t.handleJumpTo(msg.Index)
		return t, cmd

	// ===== Data Messages - Reuse List logic =====
	case DataRefreshMsg:
		cmd := t.handleDataRefresh()
		return t, cmd

	case DataChunksRefreshMsg:
		t.chunks = make(map[int]Chunk[any])
		t.loadingChunks = make(map[int]bool)
		t.hasLoadingChunks = false
		t.canScroll = true
		return t, t.smartChunkManagement()

	case DataChunkLoadedMsg:
		cmd := t.handleDataChunkLoaded(msg)
		return t, cmd

	case DataChunkErrorMsg:
		t.lastError = msg.Error
		return t, ErrorCmd(msg.Error, "chunk_load")

	case DataTotalMsg:
		t.totalItems = msg.Total
		t.updateViewportBounds()
		t.viewport.ViewportStartIndex = 0
		t.viewport.CursorIndex = t.config.ViewportConfig.InitialIndex
		t.viewport.CursorViewportIndex = t.config.ViewportConfig.InitialIndex
		return t, t.smartChunkManagement()

	case DataTotalUpdateMsg:
		oldTotal := t.totalItems
		t.totalItems = msg.Total
		t.updateViewportBounds()

		if t.viewport.CursorIndex >= t.totalItems && t.totalItems > 0 {
			t.viewport.CursorIndex = t.totalItems - 1
			t.viewport.CursorViewportIndex = t.viewport.CursorIndex - t.viewport.ViewportStartIndex
			if t.viewport.CursorViewportIndex < 0 {
				t.viewport.ViewportStartIndex = t.viewport.CursorIndex
				t.viewport.CursorViewportIndex = 0
			}
		}

		if oldTotal != t.totalItems {
			return t, t.smartChunkManagement()
		}
		return t, nil

	case DataLoadErrorMsg:
		t.lastError = msg.Error
		return t, ErrorCmd(msg.Error, "data_load")

	case DataTotalRequestMsg:
		if t.dataSource != nil {
			return t, t.dataSource.GetTotal()
		}
		return t, nil

	case DataSourceSetMsg:
		t.dataSource = msg.DataSource
		return t, t.dataSource.GetTotal()

	case ChunkUnloadedMsg:
		// Handle chunk unloaded notification (for UI feedback)
		return t, nil

	// ===== Selection Messages - Reuse List logic =====
	case SelectCurrentMsg:
		cmd := t.handleSelectCurrent()
		return t, cmd

	case SelectToggleMsg:
		cmd := t.handleSelectToggle(msg.Index)
		return t, cmd

	case SelectAllMsg:
		cmd := t.handleSelectAll()
		return t, cmd

	case SelectClearMsg:
		if t.dataSource == nil {
			return t, nil
		}
		return t, t.dataSource.ClearSelection()

	case SelectRangeMsg:
		cmd := t.handleSelectRange(msg.StartID, msg.EndID)
		return t, cmd

	case SelectionModeSetMsg:
		t.config.SelectionMode = msg.Mode
		if msg.Mode == SelectionNone {
			t.clearSelection()
		}
		return t, nil

	case SelectionResponseMsg:
		cmd := t.refreshChunks()
		return t, cmd

	// ===== Table-specific Messages =====
	case ColumnSetMsg:
		t.columns = msg.Columns
		t.config.Columns = msg.Columns
		return t, nil

	case ColumnUpdateMsg:
		if msg.Index >= 0 && msg.Index < len(t.columns) {
			t.columns[msg.Index] = msg.Column
			t.config.Columns[msg.Index] = msg.Column
		}
		return t, nil

	case HeaderVisibilityMsg:
		t.config.ShowHeader = msg.Visible
		return t, nil

	case BorderVisibilityMsg:
		t.config.ShowBorders = msg.Visible
		return t, nil

	case CellFormatterSetMsg:
		if msg.ColumnIndex >= 0 {
			t.cellFormatters[msg.ColumnIndex] = msg.Formatter
		}
		return t, nil

	case RowFormatterSetMsg:
		t.rowFormatter = msg.Formatter
		return t, nil

	case HeaderFormatterSetMsg:
		t.headerFormatter = msg.Formatter
		return t, nil

	case LoadingFormatterSetMsg:
		t.loadingFormatter = msg.Formatter
		return t, nil

	case HeaderCellFormatterSetMsg:
		t.headerCellFormatter = msg.Formatter
		return t, nil

	case TableThemeSetMsg:
		t.config.Theme = msg.Theme
		return t, nil

	// ===== Configuration Messages =====
	case ViewportConfigMsg:
		t.config.ViewportConfig = msg.Config
		t.updateViewportBounds()
		return t, nil

	case KeyMapSetMsg:
		t.config.KeyMap = msg.KeyMap
		return t, nil

	// ===== Filter Messages - Reuse List logic =====
	case FilterSetMsg:
		t.filters[msg.Field] = msg.Value
		cmd := t.handleFilterChange()
		return t, cmd

	case FilterClearMsg:
		delete(t.filters, msg.Field)
		cmd := t.handleFilterChange()
		return t, cmd

	case FiltersClearAllMsg:
		t.filters = make(map[string]any)
		cmd := t.handleFilterChange()
		return t, cmd

	// ===== Sort Messages - Reuse List logic =====
	case SortToggleMsg:
		cmd := t.handleSortToggle(msg.Field)
		return t, cmd

	case SortSetMsg:
		cmd := t.handleSortSet(msg.Field, msg.Direction)
		return t, cmd

	case SortAddMsg:
		cmd := t.handleSortAdd(msg.Field, msg.Direction)
		return t, cmd

	case SortRemoveMsg:
		cmd := t.handleSortRemove(msg.Field)
		return t, cmd

	case SortsClearAllMsg:
		t.sortFields = nil
		t.sortDirs = nil
		cmd := t.handleFilterChange()
		return t, cmd

	// ===== Focus Messages =====
	case FocusMsg:
		t.focused = true
		return t, nil

	case BlurMsg:
		t.focused = false
		return t, nil

	// ===== Search Messages =====
	case SearchSetMsg:
		t.searchQuery = msg.Query
		t.searchField = msg.Field
		cmd := t.handleSearch()
		return t, cmd

	case SearchClearMsg:
		t.searchQuery = ""
		t.searchField = ""
		t.searchResults = nil
		return t, nil

	case SearchResultMsg:
		t.searchResults = msg.Results
		return t, nil

	// ===== Error Messages =====
	case ErrorMsg:
		t.lastError = msg.Error
		return t, nil

	// ===== Viewport Messages =====
	case ViewportResizeMsg:
		t.config.ViewportConfig.Height = msg.Height
		t.updateViewportBounds()
		return t, nil

	// ===== Batch Messages =====
	case BatchMsg:
		for _, subMsg := range msg.Messages {
			var cmd tea.Cmd
			_, cmd = t.Update(subMsg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return t, tea.Batch(cmds...)

	// ===== Keyboard Input =====
	case tea.KeyMsg:
		cmd := t.handleKeyPress(msg)
		return t, cmd
	}

	return t, nil
}

// View renders the table
func (t *Table) View() string {
	var builder strings.Builder

	// Special case for empty dataset
	if t.totalItems == 0 {
		return "No data available"
	}

	// Render header if enabled
	if t.config.ShowHeader {
		header := t.renderHeader()
		if header != "" {
			builder.WriteString(header)
			builder.WriteString("\n")
		}
	}

	// Ensure visible items are up to date
	t.updateVisibleItems()

	// If we have no visible items, render empty rows or placeholders
	if len(t.visibleItems) == 0 {
		// Don't show "Loading..." - just render empty table structure
		// The chunk loading will happen automatically through smartChunkManagement
	}

	// Render each visible row
	for i, item := range t.visibleItems {
		absoluteIndex := t.viewport.ViewportStartIndex + i

		if absoluteIndex >= t.totalItems {
			break
		}

		isCursor := i == t.viewport.CursorViewportIndex

		renderedRow := t.renderRow(item, absoluteIndex, isCursor)

		builder.WriteString(renderedRow)

		if i < len(t.visibleItems)-1 && absoluteIndex < t.totalItems-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// ================================
// TABLE MODEL INTERFACE
// ================================

// Focus sets the table as focused
func (t *Table) Focus() tea.Cmd {
	t.focused = true
	return nil
}

// Blur removes focus from the table
func (t *Table) Blur() {
	t.focused = false
}

// IsFocused returns whether the table has focus
func (t *Table) IsFocused() bool {
	return t.focused
}

// GetState returns the current viewport state
func (t *Table) GetState() ViewportState {
	return t.viewport
}

// GetTotalItems returns the total number of items
func (t *Table) GetTotalItems() int {
	return t.totalItems
}

// GetSelectionCount returns the number of selected items
func (t *Table) GetSelectionCount() int {
	return GetSelectionCount(t.chunks)
}

// GetSelectedIndices returns the indices of selected items
func (t *Table) GetSelectedIndices() []int {
	var indices []int
	for _, chunk := range t.chunks {
		for i, item := range chunk.Items {
			if item.Selected {
				indices = append(indices, chunk.StartIndex+i)
			}
		}
	}
	return indices
}

// GetSelectedIDs returns the IDs of selected items
func (t *Table) GetSelectedIDs() []string {
	var ids []string
	for _, chunk := range t.chunks {
		for _, item := range chunk.Items {
			if item.Selected {
				ids = append(ids, item.ID)
			}
		}
	}
	return ids
}

// GetCurrentRow returns the currently selected row
func (t *Table) GetCurrentRow() (TableRow, bool) {
	item, exists := t.getItemAtIndex(t.viewport.CursorIndex)
	if !exists {
		return TableRow{}, false
	}

	// Convert Data[any] to TableRow
	if row, ok := item.Item.(TableRow); ok {
		return row, true
	}

	// If item is not a TableRow, try to convert it
	return TableRow{}, false
}

// ================================
// PRIVATE HELPER METHODS - Reuse List patterns
// ================================

// setupRenderContext initializes the render context
func (t *Table) setupRenderContext() {
	t.renderContext = RenderContext{
		MaxWidth:       120, // Default table width
		MaxHeight:      1,   // Single line for table rows
		Theme:          &t.config.Theme,
		BaseStyle:      t.config.Theme.CellStyle,
		ColorSupport:   true,
		UnicodeSupport: true,
		CurrentTime:    time.Now(),
		FocusState:     FocusState{HasFocus: t.focused},

		ErrorIndicator:    "❌",
		LoadingIndicator:  "⏳",
		DisabledIndicator: "🚫",
		SelectedIndicator: "✅",

		Truncate: func(text string, maxWidth int) string {
			if len(text) <= maxWidth {
				return text
			}
			if maxWidth < 3 {
				return text[:maxWidth]
			}
			return text[:maxWidth-3] + "..."
		},
		Wrap: func(text string, maxWidth int) []string {
			words := strings.Fields(text)
			if len(words) == 0 {
				return []string{""}
			}

			var lines []string
			currentLine := ""

			for _, word := range words {
				if len(currentLine) == 0 {
					currentLine = word
				} else if len(currentLine)+1+len(word) <= maxWidth {
					currentLine += " " + word
				} else {
					lines = append(lines, currentLine)
					currentLine = word
				}
			}

			if currentLine != "" {
				lines = append(lines, currentLine)
			}

			return lines
		},
		Measure: func(text string, maxWidth int) (int, int) {
			lines := strings.Split(text, "\n")
			width := 0
			for _, line := range lines {
				if len(line) > width {
					width = len(line)
				}
			}
			return width, len(lines)
		},
		OnError: func(err error) {
			t.lastError = err
		},
	}
}

// reset resets the table to its initial state
func (t *Table) reset() {
	t.chunks = make(map[int]Chunk[any])
	t.totalItems = 0
	t.loadingChunks = make(map[int]bool)
	t.hasLoadingChunks = false
	t.canScroll = true
	t.viewport = ViewportState{
		ViewportStartIndex:  0,
		CursorIndex:         t.config.ViewportConfig.InitialIndex,
		CursorViewportIndex: 0,
		IsAtTopThreshold:    false,
		IsAtBottomThreshold: false,
		AtDatasetStart:      true,
		AtDatasetEnd:        false,
	}
	t.lastError = nil
	t.filters = make(map[string]any)
	t.sortFields = nil
	t.sortDirs = nil
	t.searchQuery = ""
	t.searchField = ""
	t.searchResults = nil
}

// ================================
// NAVIGATION HELPERS - Reuse List logic
// ================================

// loadInitialData loads the total count and initial chunk
func (t *Table) loadInitialData() tea.Cmd {
	if t.dataSource == nil {
		return nil
	}
	return t.dataSource.GetTotal()
}

// handleCursorUp moves cursor up one position
func (t *Table) handleCursorUp() tea.Cmd {
	if t.totalItems == 0 || !t.canScroll {
		return nil
	}

	if t.viewport.CursorIndex <= 0 {
		return nil
	}

	previousState := t.viewport
	t.viewport = CalculateCursorUp(t.viewport, t.config.ViewportConfig, t.totalItems)

	if t.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		t.updateVisibleItems()
		return t.smartChunkManagement()
	}

	return nil
}

// handleCursorDown moves cursor down one position
func (t *Table) handleCursorDown() tea.Cmd {
	if t.totalItems == 0 || !t.canScroll {
		return nil
	}

	if t.viewport.CursorIndex >= t.totalItems-1 {
		return nil
	}

	previousState := t.viewport
	t.viewport = CalculateCursorDown(t.viewport, t.config.ViewportConfig, t.totalItems)

	if t.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		t.updateVisibleItems()
		return t.smartChunkManagement()
	}

	return nil
}

// handlePageUp moves cursor up one page
func (t *Table) handlePageUp() tea.Cmd {
	if t.totalItems == 0 || !t.canScroll {
		return nil
	}

	previousState := t.viewport
	t.viewport = CalculatePageUp(t.viewport, t.config.ViewportConfig, t.totalItems)

	if t.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		t.updateVisibleItems()
	}

	return t.smartChunkManagement()
}

// handlePageDown moves cursor down one page
func (t *Table) handlePageDown() tea.Cmd {
	if t.viewport.CursorIndex >= t.totalItems-1 {
		return nil
	}

	previousState := t.viewport
	t.viewport = CalculatePageDown(t.viewport, t.config.ViewportConfig, t.totalItems)

	if t.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		t.updateVisibleItems()
	}

	return t.smartChunkManagement()
}

// handleJumpToStart moves cursor to the start
func (t *Table) handleJumpToStart() tea.Cmd {
	if t.totalItems == 0 || !t.canScroll {
		return nil
	}

	t.viewport = CalculateJumpToStart(t.config.ViewportConfig, t.totalItems)
	return t.smartChunkManagement()
}

// handleJumpToEnd moves cursor to the end
func (t *Table) handleJumpToEnd() tea.Cmd {
	if t.totalItems <= 0 || !t.canScroll {
		return nil
	}

	previousState := t.viewport
	t.viewport = CalculateJumpToEnd(t.config.ViewportConfig, t.totalItems)

	if t.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		t.updateVisibleItems()
		return t.smartChunkManagement()
	}
	return nil
}

// handleJumpTo moves cursor to a specific index
func (t *Table) handleJumpTo(index int) tea.Cmd {
	if t.totalItems == 0 || index < 0 || index >= t.totalItems || !t.canScroll {
		return nil
	}

	t.viewport = CalculateJumpTo(index, t.config.ViewportConfig, t.totalItems)
	return t.smartChunkManagement()
}

// ================================
// DATA MANAGEMENT HELPERS - Reuse List logic
// ================================

// handleDataRefresh refreshes all data
func (t *Table) handleDataRefresh() tea.Cmd {
	t.chunks = make(map[int]Chunk[any])

	if t.dataSource == nil {
		return nil
	}

	var cmds []tea.Cmd
	cmds = append(cmds, t.dataSource.GetTotal())

	return tea.Batch(cmds...)
}

// handleDataChunkLoaded processes a loaded data chunk
func (t *Table) handleDataChunkLoaded(msg DataChunkLoadedMsg) tea.Cmd {
	chunk := Chunk[any]{
		StartIndex: msg.StartIndex,
		EndIndex:   msg.StartIndex + len(msg.Items) - 1,
		Items:      msg.Items,
		LoadedAt:   time.Now(),
		Request:    msg.Request,
	}

	t.chunks[msg.StartIndex] = chunk

	delete(t.loadingChunks, msg.StartIndex)

	t.hasLoadingChunks = len(t.loadingChunks) > 0
	if !t.hasLoadingChunks {
		t.canScroll = true
	} else {
		t.canScroll = !t.isLoadingCriticalChunks()
	}

	t.updateVisibleItems()
	t.updateViewportBounds()

	var cmds []tea.Cmd

	cmds = append(cmds, ChunkLoadingCompletedCmd(msg.StartIndex, len(msg.Items), msg.Request))

	if unloadCmd := t.unloadOldChunks(); unloadCmd != nil {
		cmds = append(cmds, unloadCmd)
	}

	return tea.Batch(cmds...)
}

// ================================
// SELECTION HELPERS - Reuse List logic
// ================================

// handleSelectCurrent selects the current item
func (t *Table) handleSelectCurrent() tea.Cmd {
	if t.config.SelectionMode == SelectionNone || t.totalItems == 0 {
		return nil
	}

	item, exists := t.getItemAtIndex(t.viewport.CursorIndex)
	if !exists {
		return nil
	}

	return t.toggleItemSelection(item.ID)
}

// handleSelectToggle toggles selection for a specific item
func (t *Table) handleSelectToggle(index int) tea.Cmd {
	if t.config.SelectionMode == SelectionNone || index < 0 || index >= t.totalItems {
		return nil
	}

	item, exists := t.getItemAtIndex(index)
	if !exists {
		return nil
	}

	return t.toggleItemSelection(item.ID)
}

// handleSelectAll selects all items via DataSource
func (t *Table) handleSelectAll() tea.Cmd {
	if t.config.SelectionMode != SelectionMultiple || t.dataSource == nil {
		return nil
	}

	return t.dataSource.SelectAll()
}

// handleSelectRange selects a range of items
func (t *Table) handleSelectRange(startID, endID string) tea.Cmd {
	if t.config.SelectionMode != SelectionMultiple {
		return nil
	}

	startIndex := t.findItemIndex(startID)
	endIndex := t.findItemIndex(endID)

	if startIndex < 0 || endIndex < 0 {
		return nil
	}

	if startIndex > endIndex {
		startIndex, endIndex = endIndex, startIndex
	}

	for i := startIndex; i <= endIndex; i++ {
		item, exists := t.getItemAtIndex(i)
		if exists && !t.selectedItems[item.ID] {
			t.selectedItems[item.ID] = true
			t.selectedOrder = append(t.selectedOrder, item.ID)
		}
	}

	return nil
}

// ================================
// FILTER AND SORT HELPERS - Reuse List logic
// ================================

// handleFilterChange triggers data refresh when filters change
func (t *Table) handleFilterChange() tea.Cmd {
	return t.handleDataRefresh()
}

// handleSortToggle toggles sorting on a field
func (t *Table) handleSortToggle(field string) tea.Cmd {
	currentSort := SortState{
		Fields:     t.sortFields,
		Directions: t.sortDirs,
	}

	newSort := ToggleSortField(currentSort, field)
	t.sortFields = newSort.Fields
	t.sortDirs = newSort.Directions

	return t.handleDataRefresh()
}

// handleSortSet sets sorting on a field
func (t *Table) handleSortSet(field, direction string) tea.Cmd {
	newSort := SetSortField(field, direction)
	t.sortFields = newSort.Fields
	t.sortDirs = newSort.Directions
	return t.handleDataRefresh()
}

// handleSortAdd adds a sort field
func (t *Table) handleSortAdd(field, direction string) tea.Cmd {
	currentSort := SortState{
		Fields:     t.sortFields,
		Directions: t.sortDirs,
	}

	newSort := AddSortField(currentSort, field, direction)
	t.sortFields = newSort.Fields
	t.sortDirs = newSort.Directions

	return t.handleDataRefresh()
}

// handleSortRemove removes a sort field
func (t *Table) handleSortRemove(field string) tea.Cmd {
	currentSort := SortState{
		Fields:     t.sortFields,
		Directions: t.sortDirs,
	}

	newSort := RemoveSortField(currentSort, field)
	t.sortFields = newSort.Fields
	t.sortDirs = newSort.Directions

	return t.handleDataRefresh()
}

// ================================
// SEARCH HELPERS
// ================================

// handleSearch performs a search
func (t *Table) handleSearch() tea.Cmd {
	if t.dataSource == nil {
		return nil
	}

	return SearchResultCmd([]int{}, t.searchQuery, 0)
}

// ================================
// KEYBOARD HANDLING
// ================================

// handleKeyPress handles keyboard input
func (t *Table) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	if !t.focused {
		return nil
	}

	key := msg.String()

	// Check navigation keys
	for _, upKey := range t.config.KeyMap.Up {
		if key == upKey {
			return t.handleCursorUp()
		}
	}

	for _, downKey := range t.config.KeyMap.Down {
		if key == downKey {
			return t.handleCursorDown()
		}
	}

	for _, pageUpKey := range t.config.KeyMap.PageUp {
		if key == pageUpKey {
			return t.handlePageUp()
		}
	}

	for _, pageDownKey := range t.config.KeyMap.PageDown {
		if key == pageDownKey {
			return t.handlePageDown()
		}
	}

	for _, homeKey := range t.config.KeyMap.Home {
		if key == homeKey {
			return t.handleJumpToStart()
		}
	}

	for _, endKey := range t.config.KeyMap.End {
		if key == endKey {
			return t.handleJumpToEnd()
		}
	}

	for _, selectKey := range t.config.KeyMap.Select {
		if key == selectKey {
			return SelectCurrentCmd()
		}
	}

	for _, selectAllKey := range t.config.KeyMap.SelectAll {
		if key == selectAllKey {
			return SelectAllCmd()
		}
	}

	return nil
}

// ================================
// RENDERING HELPERS
// ================================

// renderHeader renders the table header
func (t *Table) renderHeader() string {
	if !t.config.ShowHeader || len(t.columns) == 0 {
		return ""
	}

	if t.headerFormatter != nil {
		return t.headerFormatter(t.columns, t.renderContext)
	}

	// Default header rendering with cell-by-cell formatting
	var parts []string
	for i, col := range t.columns {
		var headerText string

		// Use HeaderCellFormatter if available
		if t.headerCellFormatter != nil {
			headerText = t.headerCellFormatter(col, i, t.renderContext)
		} else {
			// Default header cell rendering
			headerText = col.Title

			// Add sort indicator if this column is sorted
			for j, field := range t.sortFields {
				if field == col.Field {
					if t.sortDirs[j] == "asc" {
						headerText += " ↑"
					} else {
						headerText += " ↓"
					}
					break
				}
			}

			// Determine which alignment and constraint to use
			headerAlignment := col.HeaderAlignment
			if headerAlignment == 0 {
				headerAlignment = col.Alignment // Fall back to column alignment if not specified
			}

			// Use header constraint if specified, otherwise create default constraint
			var constraint CellConstraint
			if col.HeaderConstraint.Width > 0 || col.HeaderConstraint.Alignment > 0 {
				constraint = col.HeaderConstraint
				// Override alignment if HeaderAlignment is specified
				if headerAlignment != 0 {
					constraint.Alignment = headerAlignment
				}
			} else {
				// Create default constraint with header alignment
				constraint = CellConstraint{
					Width:     col.Width,
					Height:    1,
					Alignment: headerAlignment,
				}
			}

			// Apply constraints to header text
			styledHeader := t.applyCellConstraints(headerText, constraint)

			// Apply header styling
			headerText = t.config.Theme.HeaderStyle.Render(styledHeader)
		}

		parts = append(parts, headerText)
	}

	result := strings.Join(parts, t.getBorderChar())

	if t.config.ShowBorders {
		result = t.getBorderChar() + result + t.getBorderChar()
	}

	return result
}

// renderRow renders a single table row
func (t *Table) renderRow(item Data[any], absoluteIndex int, isCursor bool) string {
	// Handle loading placeholders with custom formatter
	if strings.HasPrefix(item.ID, "loading-") || strings.HasPrefix(item.ID, "missing-") {
		if t.loadingFormatter != nil {
			return t.loadingFormatter(absoluteIndex, t.columns, t.renderContext, isCursor)
		}
		// Default loading behavior - show empty cells with proper column widths
		return t.renderDefaultLoadingRow(absoluteIndex, isCursor)
	}

	if t.rowFormatter != nil {
		// Convert item to TableRow for row formatter
		if row, ok := item.Item.(TableRow); ok {
			cellResults := t.renderCellsForRow(row, absoluteIndex, isCursor, item.Selected)
			return t.rowFormatter(row, t.columns, cellResults, t.renderContext, isCursor, item.Selected)
		}
	}

	// Default row rendering
	var parts []string

	// Convert item to TableRow
	var row TableRow
	if r, ok := item.Item.(TableRow); ok {
		row = r
	} else {
		// Create a single-cell row if item is not a TableRow
		row = TableRow{
			ID:    item.ID,
			Cells: []string{fmt.Sprintf("%v", item.Item)},
		}
	}

	// Render each cell
	for i, col := range t.columns {
		var cellValue string
		if i < len(row.Cells) {
			cellValue = row.Cells[i]
		}

		var styledCell string

		// Use regular formatter or default
		if formatter, exists := t.cellFormatters[i]; exists {
			// Calculate threshold flags for the current cursor position
			isTopThreshold := isCursor && t.viewport.IsAtTopThreshold
			isBottomThreshold := isCursor && t.viewport.IsAtBottomThreshold

			styledCell = formatter(cellValue, absoluteIndex, i, col, t.renderContext, isCursor, item.Selected, isTopThreshold, isBottomThreshold)
		} else {
			// No formatter - apply default styling
			constraint := CellConstraint{
				Width:     col.Width,
				Height:    1,
				Alignment: col.Alignment,
			}

			styledCell = t.applyCellConstraints(cellValue, constraint)

			// Apply default row styling only when no formatter is present
			if isCursor {
				styledCell = t.config.Theme.CursorStyle.Render(styledCell)
			} else if item.Selected {
				styledCell = t.config.Theme.SelectedStyle.Render(styledCell)
			} else {
				styledCell = t.config.Theme.CellStyle.Render(styledCell)
			}
		}

		parts = append(parts, styledCell)
	}

	result := strings.Join(parts, t.getBorderChar())

	if t.config.ShowBorders {
		result = t.getBorderChar() + result + t.getBorderChar()
	}

	return result
}

// renderCellsForRow renders all cells for a row and returns CellRenderResults
func (t *Table) renderCellsForRow(row TableRow, absoluteIndex int, isCursor, isSelected bool) []CellRenderResult {
	var results []CellRenderResult

	for i, col := range t.columns {
		var cellValue string
		if i < len(row.Cells) {
			cellValue = row.Cells[i]
		}

		// Use regular formatter or default
		if formatter, exists := t.cellFormatters[i]; exists {
			// Calculate threshold flags for the current cursor position
			isTopThreshold := isCursor && t.viewport.IsAtTopThreshold
			isBottomThreshold := isCursor && t.viewport.IsAtBottomThreshold

			cellValue = formatter(cellValue, absoluteIndex, i, col, t.renderContext, isCursor, isSelected, isTopThreshold, isBottomThreshold)
		}

		result := CellRenderResult{
			Content:         cellValue,
			RefreshTriggers: nil,
			AnimationState:  nil,
			Error:           nil,
			Fallback:        cellValue,
		}

		results = append(results, result)
	}

	return results
}

// renderDefaultLoadingRow renders a default loading row with empty cells
func (t *Table) renderDefaultLoadingRow(absoluteIndex int, isCursor bool) string {
	var parts []string

	// Create empty cells for each column
	for _, col := range t.columns {
		constraint := CellConstraint{
			Width:     col.Width,
			Height:    1,
			Alignment: col.Alignment,
		}

		// Use loading indicator or empty space
		loadingText := ""
		if col.Width >= 10 {
			loadingText = "Loading..."
		}

		styledCell := t.applyCellConstraints(loadingText, constraint)

		// Apply cursor styling if this is the cursor row
		if isCursor {
			styledCell = t.config.Theme.CursorStyle.Render(styledCell)
		} else {
			styledCell = t.config.Theme.CellStyle.Render(styledCell)
		}

		parts = append(parts, styledCell)
	}

	result := strings.Join(parts, t.getBorderChar())

	if t.config.ShowBorders {
		result = t.getBorderChar() + result + t.getBorderChar()
	}

	return result
}

// ================================
// UTILITY HELPERS - Reuse List patterns
// ================================

// calculateBoundingArea calculates the bounding area around the current viewport automatically
func (t *Table) calculateBoundingArea() BoundingArea {
	return CalculateBoundingArea(t.viewport, t.config.ViewportConfig, t.totalItems)
}

// unloadChunksOutsideBoundingArea unloads chunks that are outside the bounding area
func (t *Table) unloadChunksOutsideBoundingArea() tea.Cmd {
	boundingArea := t.calculateBoundingArea()
	chunkSize := t.config.ViewportConfig.ChunkSize
	var cmds []tea.Cmd

	// Find and unload chunks outside the bounding area
	chunksToUnload := FindChunksToUnload(t.chunks, boundingArea, chunkSize)
	for _, chunkStart := range chunksToUnload {
		delete(t.chunks, chunkStart)
		delete(t.chunkAccessTime, chunkStart)
		cmds = append(cmds, ChunkUnloadedCmd(chunkStart))
	}

	return tea.Batch(cmds...)
}

// isChunkLoaded checks if a chunk containing the given index is loaded
func (t *Table) isChunkLoaded(index int) bool {
	return IsChunkLoaded(index, t.chunks)
}

// updateViewportPosition updates the viewport based on cursor position
func (t *Table) updateViewportPosition() {
	t.viewport = UpdateViewportPosition(t.viewport, t.config.ViewportConfig, t.totalItems)
}

// updateViewportBounds updates viewport boundary flags
func (t *Table) updateViewportBounds() {
	t.viewport = UpdateViewportBounds(t.viewport, t.config.ViewportConfig, t.totalItems)
}

// smartChunkManagement provides intelligent chunk loading with user feedback
func (t *Table) smartChunkManagement() tea.Cmd {
	if t.dataSource == nil {
		return nil
	}

	// Calculate what chunks we need for bounding area
	boundingArea := t.calculateBoundingArea()
	chunkSize := t.config.ViewportConfig.ChunkSize
	var cmds []tea.Cmd
	var newLoadingChunks []int

	// Get chunks that need to be loaded
	chunksToLoad := CalculateChunksInBoundingArea(boundingArea, chunkSize, t.totalItems)

	// Load chunks that aren't already loaded or loading
	for _, chunkStart := range chunksToLoad {
		if !t.isChunkLoaded(chunkStart) && !t.loadingChunks[chunkStart] {
			// Mark chunk as loading
			t.loadingChunks[chunkStart] = true
			newLoadingChunks = append(newLoadingChunks, chunkStart)

			request := CreateChunkRequest(
				chunkStart,
				chunkSize,
				t.totalItems,
				t.sortFields,
				t.sortDirs,
				t.filters,
			)

			// Emit chunk loading started message for observability
			cmds = append(cmds, ChunkLoadingStartedCmd(chunkStart, request))
			cmds = append(cmds, t.dataSource.LoadChunk(request))
		}
	}

	// Update loading state
	if len(newLoadingChunks) > 0 {
		t.hasLoadingChunks = true
		// Block scrolling if we're loading chunks that affect current viewport
		t.canScroll = !t.isLoadingCriticalChunks()
	}

	// Unload chunks outside bounding area
	chunksToUnload := FindChunksToUnload(t.chunks, boundingArea, chunkSize)
	for _, chunkStart := range chunksToUnload {
		delete(t.chunks, chunkStart)
		delete(t.chunkAccessTime, chunkStart)
		cmds = append(cmds, ChunkUnloadedCmd(chunkStart))
	}

	return tea.Batch(cmds...)
}

// isLoadingCriticalChunks checks if we're loading chunks that affect the current viewport
func (t *Table) isLoadingCriticalChunks() bool {
	return IsLoadingCriticalChunks(t.viewport, t.config.ViewportConfig, t.loadingChunks)
}

// updateVisibleItems updates the slice of items currently visible in the viewport
func (t *Table) updateVisibleItems() {
	result := CalculateVisibleItemsFromChunks(
		t.viewport,
		t.config.ViewportConfig,
		t.totalItems,
		t.chunks,
		t.ensureChunkLoadedImmediate,
	)

	t.visibleItems = result.Items
	t.viewport = result.AdjustedViewport
}

// ensureChunkLoadedImmediate loads the chunk containing the given index immediately
func (t *Table) ensureChunkLoadedImmediate(index int) {
	chunkStartIndex := CalculateChunkStartIndex(index, t.config.ViewportConfig.ChunkSize)
	if _, exists := t.chunks[chunkStartIndex]; !exists {
		// DON'T load chunks immediately here!
		// This function is called during rendering/visible item calculation
		// Let smartChunkManagement handle proper loading with observability messages
		// The chunk will be loaded on the next smartChunkManagement call
	}
}

// getItemAtIndex retrieves an item at a specific index
func (t *Table) getItemAtIndex(index int) (Data[any], bool) {
	return GetItemAtIndex(index, t.chunks, t.totalItems, t.chunkAccessTime)
}

// findItemIndex finds the index of an item by ID
func (t *Table) findItemIndex(id string) int {
	return FindItemIndex(id, t.chunks)
}

// toggleItemSelection toggles selection for an item via DataSource
func (t *Table) toggleItemSelection(id string) tea.Cmd {
	if t.config.SelectionMode == SelectionNone || t.dataSource == nil {
		return nil
	}

	// Find the item to determine current selection state
	var currentlySelected bool
	var itemIndex int = -1

	for _, chunk := range t.chunks {
		for i, item := range chunk.Items {
			if item.ID == id {
				currentlySelected = item.Selected
				itemIndex = chunk.StartIndex + i
				break
			}
		}
		if itemIndex >= 0 {
			break
		}
	}

	if itemIndex >= 0 {
		// Delegate to DataSource
		return t.dataSource.SetSelected(itemIndex, !currentlySelected)
	}

	return nil
}

// clearSelection clears all selections via DataSource
func (t *Table) clearSelection() {
	if t.dataSource == nil {
		return
	}
	// Delegate to DataSource - this will trigger SelectionResponseMsg when completed
	if cmd := t.dataSource.ClearSelection(); cmd != nil {
		// Execute the command immediately since this is a public method
		if msg := cmd(); msg != nil {
			t.Update(msg)
		}
	}
}

// unloadOldChunks unloads chunks that are no longer needed based on smart strategy
func (t *Table) unloadOldChunks() tea.Cmd {
	// Calculate the bounds of chunks that should be kept
	keepLowerBound, keepUpperBound := CalculateUnloadBounds(t.viewport, t.config.ViewportConfig)

	var unloadedChunks []int

	// Unload chunks outside the bounds
	for startIndex := range t.chunks {
		if ShouldUnloadChunk(startIndex, keepLowerBound, keepUpperBound) {
			delete(t.chunks, startIndex)
			delete(t.chunkAccessTime, startIndex)
			unloadedChunks = append(unloadedChunks, startIndex)
		}
	}

	// Return commands for unloaded chunks (for UI feedback)
	var cmds []tea.Cmd
	for _, chunkStart := range unloadedChunks {
		cmds = append(cmds, ChunkUnloadedCmd(chunkStart))
	}

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

// refreshChunks reloads existing chunks to get updated selection state
func (t *Table) refreshChunks() tea.Cmd {
	if t.dataSource == nil {
		return nil
	}

	var cmds []tea.Cmd

	// Reload all currently loaded chunks to get updated selection state
	for chunkStart := range t.chunks {
		request := CreateChunkRequest(
			chunkStart,
			t.config.ViewportConfig.ChunkSize,
			t.totalItems,
			t.sortFields,
			t.sortDirs,
			t.filters,
		)

		// Reload this chunk to get updated selection state
		cmds = append(cmds, t.dataSource.LoadChunk(request))
	}

	return tea.Batch(cmds...)
}

// ================================
// TABLE-SPECIFIC UTILITIES
// ================================

// applyCellConstraints applies width and alignment constraints to cell content
func (t *Table) applyCellConstraints(text string, constraint CellConstraint) string {
	// Use the shared rendering utility
	options := CellRenderOptions{
		Width:     constraint.Width,
		Height:    constraint.Height,
		Alignment: constraint.Alignment,
		Padding:   PaddingConfig{Left: 0, Right: 0, Top: 0, Bottom: 0},
		Style:     t.config.Theme.CellStyle,
		Truncate:  true,
		Wrap:      false,
	}

	result := RenderInCell(text, options)
	return result.Content
}

// getBorderChar returns the appropriate border character
func (t *Table) getBorderChar() string {
	if t.config.ShowBorders {
		return t.config.Theme.BorderChars.Vertical
	}
	return " "
}

// ================================
// PUBLIC CONFIGURATION METHODS
// ================================

// SetColumns sets the table columns
func (t *Table) SetColumns(columns []TableColumn) tea.Cmd {
	return ColumnSetCmd(columns)
}

// SetHeaderVisibility sets header visibility
func (t *Table) SetHeaderVisibility(visible bool) tea.Cmd {
	return HeaderVisibilityCmd(visible)
}

// SetBorderVisibility sets border visibility
func (t *Table) SetBorderVisibility(visible bool) tea.Cmd {
	return BorderVisibilityCmd(visible)
}

// SetCellFormatter sets a cell formatter for a specific column
func (t *Table) SetCellFormatter(columnIndex int, formatter CellFormatter) tea.Cmd {
	return CellFormatterSetCmd(columnIndex, formatter)
}

// SetRowFormatter sets the row formatter
func (t *Table) SetRowFormatter(formatter RowFormatter) tea.Cmd {
	return RowFormatterSetCmd(formatter)
}

// SetHeaderFormatter sets the header formatter
func (t *Table) SetHeaderFormatter(formatter HeaderFormatter) tea.Cmd {
	return HeaderFormatterSetCmd(formatter)
}

// SetLoadingFormatter sets the loading formatter
func (t *Table) SetLoadingFormatter(formatter LoadingRowFormatter) tea.Cmd {
	return LoadingFormatterSetCmd(formatter)
}

// SetHeaderCellFormatter sets the header cell formatter
func (t *Table) SetHeaderCellFormatter(formatter HeaderCellFormatter) tea.Cmd {
	return HeaderCellFormatterSetCmd(formatter)
}

// SetTheme sets the table theme
func (t *Table) SetTheme(theme Theme) tea.Cmd {
	return TableThemeSetCmd(theme)
}
