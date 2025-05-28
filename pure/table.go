package vtable

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ================================
// TABLE DATA STRUCTURES
// ================================

// TableDataSource provides tabular data
type TableDataSource[T any] interface {
	// Standard data source operations
	LoadChunk(request DataRequest) tea.Cmd
	LoadChunkImmediate(request DataRequest) DataChunkLoadedMsg
	GetTotal() tea.Cmd
	RefreshTotal() tea.Cmd
	GetItemID(item any) string

	// Selection operations
	SetSelected(index int, selected bool) tea.Cmd
	SetSelectedByID(id string, selected bool) tea.Cmd
	SelectAll() tea.Cmd
	ClearSelection() tea.Cmd
	SelectRange(startIndex, endIndex int) tea.Cmd

	// Table-specific operations
	GetColumns() []TableColumn
	GetCellValue(item T, columnField string) any
	SortBy(fields []string, directions []string) tea.Cmd
	FilterBy(filters map[string]any) tea.Cmd
}

// ================================
// TABLE COMPONENT
// ================================

// Table is an independent component that handles tabular data
type Table[T any] struct {
	// Core state - same as List but for table data
	tableDataSource TableDataSource[T]
	chunks          map[int]Chunk[any] // Reuse same chunk system
	totalItems      int

	// Viewport state - same as List
	viewport ViewportState

	// Configuration - reuse List config + table-specific config
	config      ListConfig
	tableConfig TableConfig

	// Table-specific state
	columns       []TableColumn
	selectedItems map[string]bool
	sortFields    []string
	sortDirs      []string
	filters       map[string]any

	// Rendering
	formatter     ItemFormatter[any]
	renderContext RenderContext

	// Focus state
	focused bool

	// Chunk management - same as List
	visibleItems     []Data[any]
	chunkAccessTime  map[int]time.Time
	loadingChunks    map[int]bool
	hasLoadingChunks bool
	canScroll        bool

	// Error handling
	lastError error
}

// TableBorderChars defines the characters used for table borders
type TableBorderChars struct {
	Horizontal  string
	Vertical    string
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Cross       string
	TopCross    string
	BottomCross string
	LeftCross   string
	RightCross  string
}

// DefaultTableBorderChars returns default border characters
func DefaultTableBorderChars() TableBorderChars {
	return TableBorderChars{
		Horizontal:  "─",
		Vertical:    "│",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
		Cross:       "┼",
		TopCross:    "┬",
		BottomCross: "┴",
		LeftCross:   "├",
		RightCross:  "┤",
	}
}

// NewTable creates a new Table component
func NewTable[T any](config ListConfig, tableConfig TableConfig, dataSource TableDataSource[T]) *Table[T] {
	// Validate and fix config - reuse List validation
	errors := ValidateListConfig(&config)
	if len(errors) > 0 {
		FixListConfig(&config)
	}

	table := &Table[T]{
		tableDataSource:  dataSource,
		chunks:           make(map[int]Chunk[any]),
		config:           config,
		tableConfig:      tableConfig,
		columns:          dataSource.GetColumns(),
		selectedItems:    make(map[string]bool),
		sortFields:       make([]string, 0),
		sortDirs:         make([]string, 0),
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

	// Set up render context - reuse List setup
	table.setupRenderContext()

	return table
}

// ================================
// TEA MODEL INTERFACE - Same as List
// ================================

// Init initializes the table model
func (t *Table[T]) Init() tea.Cmd {
	return t.loadInitialData()
}

// Update handles all messages - reuse List message handling patterns
func (t *Table[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// ===== Navigation Messages - Same as List =====
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

	// ===== Data Messages - Same as List =====
	case DataRefreshMsg:
		cmd := t.handleDataRefresh()
		return t, cmd

	case DataChunksRefreshMsg:
		// Refresh chunks while preserving cursor position
		t.chunks = make(map[int]Chunk[any])
		t.loadingChunks = make(map[int]bool)
		t.hasLoadingChunks = false
		t.canScroll = true
		return t, t.smartChunkManagement()

	case DataChunkLoadedMsg:
		cmd := t.handleDataChunkLoaded(msg)
		return t, cmd

	case DataTotalMsg:
		t.totalItems = msg.Total
		t.updateViewportBounds()
		// Reset viewport for initial load
		t.viewport.ViewportStartIndex = 0
		t.viewport.CursorIndex = t.config.ViewportConfig.InitialIndex
		t.viewport.CursorViewportIndex = t.config.ViewportConfig.InitialIndex
		return t, t.smartChunkManagement()

	case DataTotalUpdateMsg:
		// Update total while preserving cursor position
		oldTotal := t.totalItems
		t.totalItems = msg.Total
		t.updateViewportBounds()

		// Ensure cursor stays within bounds
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

	// ===== Selection Messages - Same as List =====
	case SelectCurrentMsg:
		cmd := t.handleSelectCurrent()
		return t, cmd

	case SelectAllMsg:
		cmd := t.handleSelectAll()
		return t, cmd

	case SelectClearMsg:
		cmd := t.handleSelectClear()
		return t, cmd

	case SelectionResponseMsg:
		// Handle selection response - refresh chunks to get updated selection state
		cmd := t.refreshChunks()
		return t, cmd

	// ===== Focus Messages - Same as List =====
	case FocusMsg:
		t.focused = true
		return t, nil

	case BlurMsg:
		t.focused = false
		return t, nil

	// ===== Keyboard Input - Same as List =====
	case tea.KeyMsg:
		cmd := t.handleKeyPress(msg)
		return t, cmd
	}

	return t, nil
}

// View renders the table - similar to List but with table formatting
func (t *Table[T]) View() string {
	var builder strings.Builder

	// Special case for empty dataset
	if t.totalItems == 0 {
		return "No data available"
	}

	// Ensure visible items are up to date
	t.updateVisibleItems()

	// If we have no visible items, chunks are not loaded yet
	if len(t.visibleItems) == 0 {
		return "Loading initial data..."
	}

	// Calculate column widths
	columnWidths := t.calculateColumnWidths()

	// Render header if enabled
	if t.tableConfig.ShowHeader {
		header := t.renderHeader(columnWidths)
		builder.WriteString(header)
		builder.WriteString("\n")

		if t.tableConfig.ShowBorders {
			separator := t.renderHeaderSeparator(columnWidths)
			builder.WriteString(separator)
			builder.WriteString("\n")
		}
	}

	// Render each visible row
	for i, item := range t.visibleItems {
		absoluteIndex := t.viewport.ViewportStartIndex + i

		if absoluteIndex >= t.totalItems {
			break
		}

		isCursor := i == t.viewport.CursorViewportIndex

		var renderedRow string

		if t.formatter != nil {
			// Use custom formatter
			renderedRow = t.formatter(
				item,
				absoluteIndex,
				t.renderContext,
				isCursor,
				t.viewport.IsAtTopThreshold,
				t.viewport.IsAtBottomThreshold,
			)
		} else {
			// Use default table formatter
			renderedRow = t.formatTableRow(
				item,
				absoluteIndex,
				t.renderContext,
				isCursor,
				t.viewport.IsAtTopThreshold,
				t.viewport.IsAtBottomThreshold,
				columnWidths,
			)
		}

		builder.WriteString(renderedRow)

		if i < len(t.visibleItems)-1 && absoluteIndex < t.totalItems-1 {
			builder.WriteString("\n")
		}
	}

	// Render footer border if enabled
	if t.tableConfig.ShowBorders {
		builder.WriteString("\n")
		footer := t.renderFooterBorder(columnWidths)
		builder.WriteString(footer)
	}

	return builder.String()
}

// ================================
// TABLE OPERATIONS
// ================================

// SortByColumn sorts the table by a specific column
func (t *Table[T]) SortByColumn(columnField string, direction string) tea.Cmd {
	t.sortFields = []string{columnField}
	t.sortDirs = []string{direction}
	return tea.Batch(
		t.tableDataSource.SortBy(t.sortFields, t.sortDirs),
		DataTotalUpdateCmd(t.totalItems),
		DataChunksRefreshCmd(),
	)
}

// AddSort adds a sort field to the existing sort criteria
func (t *Table[T]) AddSort(columnField string, direction string) tea.Cmd {
	t.sortFields = append(t.sortFields, columnField)
	t.sortDirs = append(t.sortDirs, direction)
	return tea.Batch(
		t.tableDataSource.SortBy(t.sortFields, t.sortDirs),
		DataTotalUpdateCmd(t.totalItems),
		DataChunksRefreshCmd(),
	)
}

// ClearSort clears all sorting
func (t *Table[T]) ClearSort() tea.Cmd {
	t.sortFields = make([]string, 0)
	t.sortDirs = make([]string, 0)
	return tea.Batch(
		t.tableDataSource.SortBy(t.sortFields, t.sortDirs),
		DataTotalUpdateCmd(t.totalItems),
		DataChunksRefreshCmd(),
	)
}

// FilterByColumn filters the table by a specific column
func (t *Table[T]) FilterByColumn(columnField string, value any) tea.Cmd {
	t.filters[columnField] = value
	return tea.Batch(
		t.tableDataSource.FilterBy(t.filters),
		DataTotalUpdateCmd(t.totalItems),
		DataChunksRefreshCmd(),
	)
}

// ClearFilter clears a specific filter
func (t *Table[T]) ClearFilter(columnField string) tea.Cmd {
	delete(t.filters, columnField)
	return tea.Batch(
		t.tableDataSource.FilterBy(t.filters),
		DataTotalUpdateCmd(t.totalItems),
		DataChunksRefreshCmd(),
	)
}

// ClearAllFilters clears all filters
func (t *Table[T]) ClearAllFilters() tea.Cmd {
	t.filters = make(map[string]any)
	return tea.Batch(
		t.tableDataSource.FilterBy(t.filters),
		DataTotalUpdateCmd(t.totalItems),
		DataChunksRefreshCmd(),
	)
}

// ================================
// CHUNK MANAGEMENT - Reuse List functions
// ================================

// loadInitialData loads the total count and initial chunk
func (t *Table[T]) loadInitialData() tea.Cmd {
	return t.tableDataSource.GetTotal()
}

// smartChunkManagement - reuse List logic
func (t *Table[T]) smartChunkManagement() tea.Cmd {
	// Calculate bounding area - reuse List function
	boundingArea := CalculateBoundingArea(t.viewport, t.config.ViewportConfig, t.totalItems)
	chunkSize := t.config.ViewportConfig.ChunkSize
	var cmds []tea.Cmd

	// Get chunks that need to be loaded
	chunksToLoad := CalculateChunksInBoundingArea(boundingArea, chunkSize, t.totalItems)

	// Load chunks that aren't already loaded
	for _, chunkStart := range chunksToLoad {
		if !IsChunkLoaded(chunkStart, t.chunks) && !t.loadingChunks[chunkStart] {
			t.loadingChunks[chunkStart] = true

			// Create data request
			request := DataRequest{
				Start: chunkStart,
				Count: chunkSize,
			}

			// Send chunk loading started message
			cmds = append(cmds, ChunkLoadingStartedCmd(chunkStart, request))

			// Load chunk from data source
			cmd := t.tableDataSource.LoadChunk(request)
			cmds = append(cmds, cmd)
		}
	}

	// Update loading state
	if len(chunksToLoad) > 0 {
		t.hasLoadingChunks = true
		t.canScroll = !IsLoadingCriticalChunks(t.viewport, t.config.ViewportConfig, t.loadingChunks)
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

// ================================
// NAVIGATION HANDLERS - Reuse List logic
// ================================

// handleCursorUp - reuse List logic
func (t *Table[T]) handleCursorUp() tea.Cmd {
	if t.totalItems == 0 || !t.canScroll || t.viewport.CursorIndex <= 0 {
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

// handleCursorDown - reuse List logic
func (t *Table[T]) handleCursorDown() tea.Cmd {
	if t.totalItems == 0 || !t.canScroll || t.viewport.CursorIndex >= t.totalItems-1 {
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

// handlePageUp - reuse List logic
func (t *Table[T]) handlePageUp() tea.Cmd {
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

// handlePageDown - reuse List logic
func (t *Table[T]) handlePageDown() tea.Cmd {
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

// handleJumpToStart - reuse List logic
func (t *Table[T]) handleJumpToStart() tea.Cmd {
	if t.totalItems == 0 || !t.canScroll {
		return nil
	}

	t.viewport = CalculateJumpToStart(t.config.ViewportConfig, t.totalItems)
	return t.smartChunkManagement()
}

// handleJumpToEnd - reuse List logic
func (t *Table[T]) handleJumpToEnd() tea.Cmd {
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

// handleJumpTo - reuse List logic
func (t *Table[T]) handleJumpTo(index int) tea.Cmd {
	if t.totalItems == 0 || index < 0 || index >= t.totalItems || !t.canScroll {
		return nil
	}

	t.viewport = CalculateJumpTo(index, t.config.ViewportConfig, t.totalItems)
	return t.smartChunkManagement()
}

// ================================
// HELPER METHODS - Reuse List logic
// ================================

// updateViewportPosition - reuse List logic
func (t *Table[T]) updateViewportPosition() {
	t.viewport = UpdateViewportPosition(t.viewport, t.config.ViewportConfig, t.totalItems)
}

// updateViewportBounds - reuse List logic
func (t *Table[T]) updateViewportBounds() {
	t.viewport = UpdateViewportBounds(t.viewport, t.config.ViewportConfig, t.totalItems)
}

// updateVisibleItems - reuse List logic
func (t *Table[T]) updateVisibleItems() {
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

// ensureChunkLoadedImmediate - reuse List logic
func (t *Table[T]) ensureChunkLoadedImmediate(index int) {
	chunkStartIndex := CalculateChunkStartIndex(index, t.config.ViewportConfig.ChunkSize)
	if _, exists := t.chunks[chunkStartIndex]; !exists {
		// Load chunk immediately from data source
		request := DataRequest{
			Start: chunkStartIndex,
			Count: t.config.ViewportConfig.ChunkSize,
		}
		msg := t.tableDataSource.LoadChunkImmediate(request)
		t.handleDataChunkLoaded(msg)
	}
}

// ================================
// REMAINING HANDLERS - Similar to List
// ================================

// handleDataRefresh refreshes all data
func (t *Table[T]) handleDataRefresh() tea.Cmd {
	t.chunks = make(map[int]Chunk[any])
	return t.tableDataSource.GetTotal()
}

// handleDataChunkLoaded processes a loaded data chunk
func (t *Table[T]) handleDataChunkLoaded(msg DataChunkLoadedMsg) tea.Cmd {
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
		t.canScroll = !IsLoadingCriticalChunks(t.viewport, t.config.ViewportConfig, t.loadingChunks)
	}

	t.updateVisibleItems()
	t.updateViewportBounds()

	return ChunkLoadingCompletedCmd(msg.StartIndex, len(msg.Items), msg.Request)
}

// handleSelectCurrent selects the current item
func (t *Table[T]) handleSelectCurrent() tea.Cmd {
	if t.config.SelectionMode == SelectionNone || t.totalItems == 0 {
		return nil
	}

	if t.viewport.CursorIndex >= 0 && t.viewport.CursorIndex < t.totalItems {
		return t.tableDataSource.SetSelected(t.viewport.CursorIndex, !t.selectedItems[fmt.Sprintf("%d", t.viewport.CursorIndex)])
	}
	return nil
}

// handleSelectAll selects all items
func (t *Table[T]) handleSelectAll() tea.Cmd {
	if t.config.SelectionMode != SelectionMultiple {
		return nil
	}

	return t.tableDataSource.SelectAll()
}

// handleSelectClear clears all selections
func (t *Table[T]) handleSelectClear() tea.Cmd {
	t.selectedItems = make(map[string]bool)
	return t.tableDataSource.ClearSelection()
}

// refreshChunks reloads existing chunks to get updated selection state
func (t *Table[T]) refreshChunks() tea.Cmd {
	var cmds []tea.Cmd

	// Reload all currently loaded chunks
	for chunkStart := range t.chunks {
		request := DataRequest{
			Start: chunkStart,
			Count: t.config.ViewportConfig.ChunkSize,
		}
		cmd := t.tableDataSource.LoadChunk(request)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// handleKeyPress handles keyboard input - reuse List logic
func (t *Table[T]) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	if !t.focused {
		return nil
	}

	key := msg.String()

	// Check navigation keys - reuse List key mapping logic
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

	for _, filterKey := range t.config.KeyMap.Filter {
		if key == filterKey {
			// Return command to start filtering
			return StatusCmd("Filter mode", StatusInfo)
		}
	}

	for _, sortKey := range t.config.KeyMap.Sort {
		if key == sortKey {
			// Return command to start sorting
			return StatusCmd("Sort mode", StatusInfo)
		}
	}

	return nil
}

// ================================
// TABLE FORMATTING
// ================================

// calculateColumnWidths calculates the width for each column
func (t *Table[T]) calculateColumnWidths() []int {
	widths := make([]int, len(t.columns))

	// Use configured widths or auto-size
	for i, col := range t.columns {
		if col.Width > 0 {
			widths[i] = col.Width
		} else {
			// Auto-size based on header and content
			headerWidth := len(col.Title)

			// Check content width from visible items
			maxContentWidth := headerWidth
			for _, item := range t.visibleItems {
				if cellValue := t.tableDataSource.GetCellValue(item.Item.(T), col.Field); cellValue != nil {
					contentWidth := len(fmt.Sprintf("%v", cellValue))
					if contentWidth > maxContentWidth {
						maxContentWidth = contentWidth
					}
				}
			}

			// Use the larger of header or content width, with minimum of 8
			width := maxContentWidth
			if width < 8 {
				width = 8
			}
			widths[i] = width
		}
	}

	return widths
}

// renderHeader renders the table header
func (t *Table[T]) renderHeader(columnWidths []int) string {
	var cells []string

	for i, col := range t.columns {
		width := columnWidths[i]
		title := col.Title

		// Apply alignment based on column alignment
		switch col.Alignment {
		case AlignCenter:
			title = t.centerAlign(title, width)
		case AlignRight:
			title = t.rightAlign(title, width)
		default: // AlignLeft
			title = t.leftAlign(title, width)
		}

		// Apply header styling
		styledTitle := t.tableConfig.Theme.HeaderStyle.Render(title)
		cells = append(cells, styledTitle)
	}

	// Join with borders if enabled
	if t.tableConfig.ShowBorders {
		borderChar := t.tableConfig.Theme.BorderChars.Vertical
		return borderChar + strings.Join(cells, borderChar) + borderChar
	}

	return strings.Join(cells, " ")
}

// renderHeaderSeparator renders the separator line below the header
func (t *Table[T]) renderHeaderSeparator(columnWidths []int) string {
	var segments []string

	for i := range t.columns {
		width := columnWidths[i]
		segment := strings.Repeat(t.tableConfig.Theme.BorderChars.Horizontal, width)
		segments = append(segments, segment)
	}

	if t.tableConfig.ShowBorders {
		left := t.tableConfig.Theme.BorderChars.LeftT
		right := t.tableConfig.Theme.BorderChars.RightT
		separator := t.tableConfig.Theme.BorderChars.TopT

		return left + strings.Join(segments, separator) + right
	}

	return strings.Join(segments, " ")
}

// renderFooterBorder renders the bottom border of the table
func (t *Table[T]) renderFooterBorder(columnWidths []int) string {
	var segments []string

	for i := range t.columns {
		width := columnWidths[i]
		segment := strings.Repeat(t.tableConfig.Theme.BorderChars.Horizontal, width)
		segments = append(segments, segment)
	}

	if t.tableConfig.ShowBorders {
		left := t.tableConfig.Theme.BorderChars.BottomLeft
		right := t.tableConfig.Theme.BorderChars.BottomRight
		separator := t.tableConfig.Theme.BorderChars.BottomT

		return left + strings.Join(segments, separator) + right
	}

	return strings.Join(segments, " ")
}

// formatTableRow formats a table row with proper column alignment
func (t *Table[T]) formatTableRow(
	item Data[any],
	index int,
	ctx RenderContext,
	isCursor, isTopThreshold, isBottomThreshold bool,
	columnWidths []int,
) string {
	var cells []string

	for i, col := range t.columns {
		width := columnWidths[i]

		// Get cell value
		var cellValue any
		if typedItem, ok := item.Item.(T); ok {
			cellValue = t.tableDataSource.GetCellValue(typedItem, col.Field)
		}

		// Format cell content
		cellContent := fmt.Sprintf("%v", cellValue)

		// Apply alignment
		switch col.Alignment {
		case AlignCenter:
			cellContent = t.centerAlign(cellContent, width)
		case AlignRight:
			cellContent = t.rightAlign(cellContent, width)
		default: // AlignLeft
			cellContent = t.leftAlign(cellContent, width)
		}

		// Apply styling
		if isCursor {
			cellContent = t.tableConfig.Theme.CursorStyle.Render(cellContent)
		} else if item.Selected {
			cellContent = t.tableConfig.Theme.SelectedStyle.Render(cellContent)
		} else {
			cellContent = t.tableConfig.Theme.CellStyle.Render(cellContent)
		}

		cells = append(cells, cellContent)
	}

	// Join with borders if enabled
	if t.tableConfig.ShowBorders {
		borderChar := t.tableConfig.Theme.BorderChars.Vertical
		return borderChar + strings.Join(cells, borderChar) + borderChar
	}

	return strings.Join(cells, " ")
}

// Helper alignment functions
func (t *Table[T]) leftAlign(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	return text + strings.Repeat(" ", width-len(text))
}

func (t *Table[T]) rightAlign(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	return strings.Repeat(" ", width-len(text)) + text
}

func (t *Table[T]) centerAlign(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}
	padding := width - len(text)
	leftPad := padding / 2
	rightPad := padding - leftPad
	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}

// setupRenderContext - reuse List logic
func (t *Table[T]) setupRenderContext() {
	t.renderContext = RenderContext{
		MaxWidth:          t.config.MaxWidth,
		MaxHeight:         1,
		Theme:             &t.tableConfig.Theme,
		BaseStyle:         t.config.StyleConfig.DefaultStyle,
		ColorSupport:      true,
		UnicodeSupport:    true,
		CurrentTime:       time.Now(),
		FocusState:        FocusState{HasFocus: t.focused},
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
		OnError: func(err error) {
			t.lastError = err
		},
	}
}

// ================================
// PUBLIC INTERFACE - Same as List
// ================================

// Focus sets the table as focused
func (t *Table[T]) Focus() tea.Cmd {
	t.focused = true
	return nil
}

// Blur removes focus from the table
func (t *Table[T]) Blur() {
	t.focused = false
}

// IsFocused returns whether the table has focus
func (t *Table[T]) IsFocused() bool {
	return t.focused
}

// GetState returns the current viewport state
func (t *Table[T]) GetState() ViewportState {
	return t.viewport
}

// GetSelectionCount returns the number of selected items
func (t *Table[T]) GetSelectionCount() int {
	return len(t.selectedItems)
}

// ================================
// TABLE CONFIGURATION
// ================================

// SetShowHeader enables or disables the header
func (t *Table[T]) SetShowHeader(show bool) {
	t.tableConfig.ShowHeader = show
}

// SetShowBorders enables or disables borders
func (t *Table[T]) SetShowBorders(show bool) {
	t.tableConfig.ShowBorders = show
}

// GetColumns returns the current columns
func (t *Table[T]) GetColumns() []TableColumn {
	return t.columns
}

// SetColumns updates the table columns
func (t *Table[T]) SetColumns(columns []TableColumn) {
	t.columns = columns
}
