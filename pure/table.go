package vtable

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ================================
// TABLE MODEL IMPLEMENTATION
// ================================

// Table represents a pure Tea table component
type Table struct {
	// Core state
	dataSource DataSource[any]
	chunks     map[int]Chunk[any] // Map of start index to chunk
	totalItems int

	// Viewport state
	viewport ViewportState

	// Configuration
	config TableConfig

	// Rendering formatters
	cellFormatters     map[int]CellFormatter         // Column index to formatter
	animatedFormatters map[int]CellFormatterAnimated // Column index to animated formatter
	rowFormatter       RowFormatter
	headerFormatter    HeaderFormatter

	// Column constraints
	columnConstraints map[int]CellConstraint

	// Selection state
	selectedItems map[string]bool
	selectedOrder []string // Maintain selection order

	// Focus state
	focused    bool
	focusedCol int // For cell-level focus (future feature)

	// Animation state
	animationEngine AnimationEngine
	cellAnimations  map[string]CellAnimation // Key: "rowID:colIndex"
	rowAnimations   map[string]RowAnimation

	// Error state
	lastError error

	// Filtering and sorting
	filters    map[string]any
	sortFields []string
	sortDirs   []string

	// Search state
	searchQuery   string
	searchField   string
	searchResults []int

	// Performance monitoring
	renderContext RenderContext
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
		dataSource:         dataSource,
		chunks:             make(map[int]Chunk[any]),
		config:             config,
		cellFormatters:     make(map[int]CellFormatter),
		animatedFormatters: make(map[int]CellFormatterAnimated),
		columnConstraints:  make(map[int]CellConstraint),
		selectedItems:      make(map[string]bool),
		selectedOrder:      make([]string, 0),
		cellAnimations:     make(map[string]CellAnimation),
		rowAnimations:      make(map[string]RowAnimation),
		filters:            make(map[string]any),
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

	// Initialize column constraints with defaults
	table.initializeColumnConstraints()

	// Set up render context
	table.setupRenderContext()

	return table
}

// ================================
// TEA MODEL INTERFACE
// ================================

// Init initializes the table model
func (t *Table) Init() tea.Cmd {
	cmds := []tea.Cmd{
		InitCmd(),
	}

	// Load initial data
	if t.dataSource != nil {
		cmds = append(cmds, t.dataSource.GetTotal())
		cmds = append(cmds, t.loadInitialChunk())
	}

	// Start animation engine if enabled
	if t.config.AnimationConfig.Enabled && t.animationEngine != nil {
		cmds = append(cmds, t.animationEngine.StartLoop())
	}

	return tea.Batch(cmds...)
}

// Update handles all messages and updates the table state
func (t *Table) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// ===== Lifecycle Messages =====
	case InitMsg:
		return t, t.Init()

	case DestroyMsg:
		if t.animationEngine != nil {
			t.animationEngine.Cleanup()
		}
		return t, nil

	case ResetMsg:
		t.reset()
		return t, t.Init()

	// ===== Navigation Messages =====
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

	// ===== Data Messages =====
	case DataRefreshMsg:
		cmd := t.handleDataRefresh()
		return t, cmd

	case DataChunkLoadedMsg:
		cmd := t.handleDataChunkLoaded(msg)
		return t, cmd

	case DataChunkErrorMsg:
		t.lastError = msg.Error
		return t, ErrorCmd(msg.Error, "chunk_load")

	case DataTotalMsg:
		t.totalItems = msg.Total
		t.updateViewportBounds()
		return t, nil

	case DataLoadErrorMsg:
		t.lastError = msg.Error
		return t, ErrorCmd(msg.Error, "data_load")

	case DataSourceSetMsg:
		t.dataSource = msg.DataSource
		return t, t.dataSource.GetTotal()

	// ===== Selection Messages =====
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
		t.clearSelection()
		return t, nil

	case SelectRangeMsg:
		cmd := t.handleSelectRange(msg.StartID, msg.EndID)
		return t, cmd

	case SelectionModeSetMsg:
		t.config.SelectionMode = msg.Mode
		if msg.Mode == SelectionNone {
			t.clearSelection()
		}
		return t, nil

	// ===== Filter Messages =====
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

	// ===== Sort Messages =====
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
		cmd := t.handleFilterChange() // Refresh data
		return t, cmd

	// ===== Focus Messages =====
	case FocusMsg:
		t.focused = true
		return t, nil

	case BlurMsg:
		t.focused = false
		return t, nil

	// ===== Animation Messages =====
	case GlobalAnimationTickMsg:
		if t.animationEngine != nil {
			cmd := t.animationEngine.ProcessGlobalTick(msg)
			return t, cmd
		}
		return t, nil

	case AnimationConfigMsg:
		t.config.AnimationConfig = msg.Config
		if t.animationEngine != nil {
			cmd := t.animationEngine.UpdateConfig(msg.Config)
			return t, cmd
		}
		return t, nil

	case CellAnimationStartMsg:
		key := fmt.Sprintf("%s:%d", msg.RowID, msg.ColumnIndex)
		t.cellAnimations[key] = msg.Animation
		return t, nil

	case CellAnimationStopMsg:
		key := fmt.Sprintf("%s:%d", msg.RowID, msg.ColumnIndex)
		delete(t.cellAnimations, key)
		return t, nil

	case RowAnimationStartMsg:
		t.rowAnimations[msg.RowID] = msg.Animation
		return t, nil

	case RowAnimationStopMsg:
		delete(t.rowAnimations, msg.RowID)
		return t, nil

	// ===== Table-specific Messages =====
	case ColumnSetMsg:
		t.config.Columns = msg.Columns
		t.initializeColumnConstraints()
		return t, nil

	case ColumnUpdateMsg:
		if msg.Index >= 0 && msg.Index < len(t.config.Columns) {
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
		} else {
			// Set for all columns
			for i := range t.config.Columns {
				t.cellFormatters[i] = msg.Formatter
			}
		}
		return t, nil

	case CellAnimatedFormatterSetMsg:
		t.animatedFormatters[msg.ColumnIndex] = msg.Formatter
		return t, nil

	case RowFormatterSetMsg:
		t.rowFormatter = msg.Formatter
		return t, nil

	case HeaderFormatterSetMsg:
		t.headerFormatter = msg.Formatter
		return t, nil

	case ColumnConstraintsSetMsg:
		t.columnConstraints[msg.ColumnIndex] = msg.Constraints
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
	if t.totalItems == 0 {
		return t.renderEmpty()
	}

	var lines []string

	// Render header if enabled
	if t.config.ShowHeader {
		header := t.renderHeader()
		lines = append(lines, header)

		// Add separator line if borders are enabled
		if t.config.ShowBorders {
			separator := t.renderHeaderSeparator()
			lines = append(lines, separator)
		}
	}

	// Render data rows
	visibleHeight := t.config.ViewportConfig.Height
	for i := 0; i < visibleHeight; i++ {
		absoluteIndex := t.viewport.ViewportStartIndex + i
		if absoluteIndex >= t.totalItems {
			break
		}

		rowLine := t.renderRow(absoluteIndex, i)
		lines = append(lines, rowLine)
	}

	// Add bottom border if enabled
	if t.config.ShowBorders {
		border := t.renderBottomBorder()
		lines = append(lines, border)
	}

	return strings.Join(lines, "\n")
}

// ================================
// TABLE MODEL INTERFACE
// ================================

// Focus gives focus to the table
func (t *Table) Focus() tea.Cmd {
	return FocusCmd()
}

// Blur removes focus from the table
func (t *Table) Blur() tea.Cmd {
	return BlurCmd()
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

// GetSelectedIndices returns the indices of selected items
func (t *Table) GetSelectedIndices() []int {
	var indices []int
	for id := range t.selectedItems {
		if index := t.findItemIndex(id); index >= 0 {
			indices = append(indices, index)
		}
	}
	return indices
}

// GetSelectedIDs returns the IDs of selected items
func (t *Table) GetSelectedIDs() []string {
	return t.selectedOrder
}

// GetSelectionCount returns the number of selected items
func (t *Table) GetSelectionCount() int {
	return len(t.selectedItems)
}

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

// SetCellFormatter sets a cell formatter for a column
func (t *Table) SetCellFormatter(columnIndex int, formatter CellFormatter) tea.Cmd {
	return CellFormatterSetCmd(columnIndex, formatter)
}

// SetCellAnimatedFormatter sets an animated cell formatter for a column
func (t *Table) SetCellAnimatedFormatter(columnIndex int, formatter CellFormatterAnimated) tea.Cmd {
	return CellAnimatedFormatterSetCmd(columnIndex, formatter)
}

// SetRowFormatter sets the row formatter
func (t *Table) SetRowFormatter(formatter RowFormatter) tea.Cmd {
	return RowFormatterSetCmd(formatter)
}

// SetHeaderFormatter sets the header formatter
func (t *Table) SetHeaderFormatter(formatter HeaderFormatter) tea.Cmd {
	return HeaderFormatterSetCmd(formatter)
}

// SetColumnConstraints sets constraints for a column
func (t *Table) SetColumnConstraints(columnIndex int, constraints CellConstraint) tea.Cmd {
	return ColumnConstraintsSetCmd(columnIndex, constraints)
}

// GetCurrentRow returns the row at the cursor position
func (t *Table) GetCurrentRow() (TableRow, bool) {
	row, exists := t.getRowAtIndex(t.viewport.CursorIndex)
	return row, exists
}

// ================================
// NAVIGATION HELPERS (reusing List logic with Table-specific adaptations)
// ================================

// loadInitialChunk loads the initial chunk of data
func (t *Table) loadInitialChunk() tea.Cmd {
	if t.dataSource == nil {
		return nil
	}

	request := DataRequest{
		Start:          0,
		Count:          t.config.ViewportConfig.ChunkSize,
		SortFields:     t.sortFields,
		SortDirections: t.sortDirs,
		Filters:        t.filters,
	}

	return t.dataSource.LoadChunk(request)
}

// Navigation helpers (similar to List but for Table context)
func (t *Table) handleCursorUp() tea.Cmd {
	if t.totalItems == 0 {
		return nil
	}

	if t.viewport.CursorIndex > 0 {
		t.viewport.CursorIndex--
		t.updateViewportPosition()
		return t.checkAndLoadChunks()
	}

	return nil
}

func (t *Table) handleCursorDown() tea.Cmd {
	if t.totalItems == 0 {
		return nil
	}

	if t.viewport.CursorIndex < t.totalItems-1 {
		t.viewport.CursorIndex++
		t.updateViewportPosition()
		return t.checkAndLoadChunks()
	}

	return nil
}

func (t *Table) handlePageUp() tea.Cmd {
	if t.totalItems == 0 {
		return nil
	}

	pageSize := t.config.ViewportConfig.Height
	newIndex := t.viewport.CursorIndex - pageSize
	if newIndex < 0 {
		newIndex = 0
	}

	t.viewport.CursorIndex = newIndex
	t.updateViewportPosition()
	return t.checkAndLoadChunks()
}

func (t *Table) handlePageDown() tea.Cmd {
	if t.totalItems == 0 {
		return nil
	}

	pageSize := t.config.ViewportConfig.Height
	newIndex := t.viewport.CursorIndex + pageSize
	if newIndex >= t.totalItems {
		newIndex = t.totalItems - 1
	}

	t.viewport.CursorIndex = newIndex
	t.updateViewportPosition()
	return t.checkAndLoadChunks()
}

func (t *Table) handleJumpToStart() tea.Cmd {
	if t.totalItems == 0 {
		return nil
	}

	t.viewport.CursorIndex = 0
	t.updateViewportPosition()
	return t.checkAndLoadChunks()
}

func (t *Table) handleJumpToEnd() tea.Cmd {
	if t.totalItems == 0 {
		return nil
	}

	t.viewport.CursorIndex = t.totalItems - 1
	t.updateViewportPosition()
	return t.checkAndLoadChunks()
}

func (t *Table) handleJumpTo(index int) tea.Cmd {
	if t.totalItems == 0 || index < 0 || index >= t.totalItems {
		return nil
	}

	t.viewport.CursorIndex = index
	t.updateViewportPosition()
	return t.checkAndLoadChunks()
}

// ================================
// DATA MANAGEMENT HELPERS
// ================================

func (t *Table) handleDataRefresh() tea.Cmd {
	t.chunks = make(map[int]Chunk[any])

	if t.dataSource == nil {
		return nil
	}

	var cmds []tea.Cmd
	cmds = append(cmds, t.dataSource.GetTotal())
	cmds = append(cmds, t.loadInitialChunk())

	return tea.Batch(cmds...)
}

func (t *Table) handleDataChunkLoaded(msg DataChunkLoadedMsg) tea.Cmd {
	chunk := Chunk[any]{
		StartIndex: msg.StartIndex,
		EndIndex:   msg.StartIndex + len(msg.Items) - 1,
		Items:      msg.Items,
		LoadedAt:   time.Now(),
		Request:    msg.Request,
	}

	t.chunks[msg.StartIndex] = chunk
	t.updateViewportBounds()

	return nil
}

// ================================
// SELECTION HELPERS
// ================================

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

func (t *Table) handleSelectAll() tea.Cmd {
	if t.config.SelectionMode != SelectionMultiple {
		return nil
	}

	for _, chunk := range t.chunks {
		for _, item := range chunk.Items {
			if !t.selectedItems[item.ID] {
				t.selectedItems[item.ID] = true
				t.selectedOrder = append(t.selectedOrder, item.ID)
			}
		}
	}

	return nil
}

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
// FILTER AND SORT HELPERS
// ================================

func (t *Table) handleFilterChange() tea.Cmd {
	return t.handleDataRefresh()
}

func (t *Table) handleSortToggle(field string) tea.Cmd {
	for i, sortField := range t.sortFields {
		if sortField == field {
			if t.sortDirs[i] == "asc" {
				t.sortDirs[i] = "desc"
			} else {
				t.sortDirs[i] = "asc"
			}
			return t.handleDataRefresh()
		}
	}

	t.sortFields = append(t.sortFields, field)
	t.sortDirs = append(t.sortDirs, "asc")
	return t.handleDataRefresh()
}

func (t *Table) handleSortSet(field, direction string) tea.Cmd {
	t.sortFields = []string{field}
	t.sortDirs = []string{direction}
	return t.handleDataRefresh()
}

func (t *Table) handleSortAdd(field, direction string) tea.Cmd {
	for i, sortField := range t.sortFields {
		if sortField == field {
			t.sortFields = append(t.sortFields[:i], t.sortFields[i+1:]...)
			t.sortDirs = append(t.sortDirs[:i], t.sortDirs[i+1:]...)
			break
		}
	}

	t.sortFields = append(t.sortFields, field)
	t.sortDirs = append(t.sortDirs, direction)
	return t.handleDataRefresh()
}

func (t *Table) handleSortRemove(field string) tea.Cmd {
	for i, sortField := range t.sortFields {
		if sortField == field {
			t.sortFields = append(t.sortFields[:i], t.sortFields[i+1:]...)
			t.sortDirs = append(t.sortDirs[:i], t.sortDirs[i+1:]...)
			return t.handleDataRefresh()
		}
	}
	return nil
}

// ================================
// SEARCH HELPERS
// ================================

func (t *Table) handleSearch() tea.Cmd {
	if t.dataSource == nil {
		return nil
	}
	return SearchResultCmd([]int{}, t.searchQuery, 0)
}

// ================================
// KEYBOARD HANDLING
// ================================

func (t *Table) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	if !t.focused {
		return nil
	}

	key := msg.String()

	for _, upKey := range t.config.KeyMap.Up {
		if key == upKey {
			return CursorUpCmd()
		}
	}

	for _, downKey := range t.config.KeyMap.Down {
		if key == downKey {
			return CursorDownCmd()
		}
	}

	for _, pageUpKey := range t.config.KeyMap.PageUp {
		if key == pageUpKey {
			return PageUpCmd()
		}
	}

	for _, pageDownKey := range t.config.KeyMap.PageDown {
		if key == pageDownKey {
			return PageDownCmd()
		}
	}

	for _, homeKey := range t.config.KeyMap.Home {
		if key == homeKey {
			return JumpToStartCmd()
		}
	}

	for _, endKey := range t.config.KeyMap.End {
		if key == endKey {
			return JumpToEndCmd()
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

func (t *Table) renderEmpty() string {
	style := t.config.Theme.CellStyle
	if t.lastError != nil {
		style = t.config.Theme.ErrorStyle
		return style.Render("Error: " + t.lastError.Error())
	}
	return style.Render("No data")
}

func (t *Table) renderHeader() string {
	if t.headerFormatter != nil {
		return t.headerFormatter(t.config.Columns, t.renderContext)
	}

	var cells []string
	for i, col := range t.config.Columns {
		cellContent := col.Title

		// Apply column width constraint
		if constraint, exists := t.columnConstraints[i]; exists {
			cellContent = t.renderContext.Truncate(cellContent, constraint.Width)
		}

		// Apply header style
		styled := t.config.Theme.HeaderStyle.Render(cellContent)
		cells = append(cells, styled)
	}

	if t.config.ShowBorders {
		return "│ " + strings.Join(cells, " │ ") + " │"
	}
	return strings.Join(cells, " ")
}

func (t *Table) renderHeaderSeparator() string {
	var parts []string

	if t.config.ShowBorders {
		parts = append(parts, "├")
	}

	for i, col := range t.config.Columns {
		width := col.Width
		if constraint, exists := t.columnConstraints[i]; exists {
			width = constraint.Width
		}

		separator := strings.Repeat("─", width)
		parts = append(parts, separator)

		if i < len(t.config.Columns)-1 && t.config.ShowBorders {
			parts = append(parts, "┼")
		}
	}

	if t.config.ShowBorders {
		parts = append(parts, "┤")
	}

	return strings.Join(parts, "")
}

func (t *Table) renderRow(absoluteIndex, viewportIndex int) string {
	item, exists := t.getItemAtIndex(absoluteIndex)
	if !exists {
		return t.renderLoadingRow()
	}

	row := t.convertItemToRow(item)

	isCursor := absoluteIndex == t.viewport.CursorIndex
	isSelected := t.selectedItems[item.ID]

	// Use row formatter if available
	if t.rowFormatter != nil {
		// First render all cells to get CellRenderResults
		var cellResults []CellRenderResult
		for i, col := range t.config.Columns {
			cellValue := t.getCellValue(row, col.Field)
			cellResult := t.renderCellWithResult(cellValue, i, item.ID, isCursor, isSelected, absoluteIndex, col)
			cellResults = append(cellResults, cellResult)
		}
		return t.rowFormatter(row, t.config.Columns, cellResults, t.renderContext, isCursor, isSelected)
	}

	// Default row rendering
	var cells []string
	for i, col := range t.config.Columns {
		cellValue := t.getCellValue(row, col.Field)
		cellContent := t.renderCell(cellValue, i, item.ID, isCursor, isSelected, absoluteIndex, col)
		cells = append(cells, cellContent)
	}

	if t.config.ShowBorders {
		return "│ " + strings.Join(cells, " │ ") + " │"
	}
	return strings.Join(cells, " ")
}

func (t *Table) renderCell(value any, columnIndex int, rowID string, isCursor, isSelected bool, absoluteIndex int, col TableColumn) string {
	var content string

	// Convert value to string for formatters
	stringValue := fmt.Sprintf("%v", value)

	// Use animated formatter if available and animations are enabled
	if formatter, exists := t.animatedFormatters[columnIndex]; exists && t.config.AnimationConfig.Enabled {
		key := fmt.Sprintf("%s:%d", rowID, columnIndex)
		animationState := make(map[string]any)
		if animation, exists := t.cellAnimations[key]; exists {
			animationState = animation.State
		}

		result := formatter(stringValue, absoluteIndex, columnIndex, col, t.renderContext, animationState, isCursor, isSelected, false, false)
		content = result.Content
	} else if formatter, exists := t.cellFormatters[columnIndex]; exists {
		// Use regular cell formatter
		content = formatter(stringValue, absoluteIndex, columnIndex, col, t.renderContext, isCursor, isSelected)
	} else {
		// Default formatting
		content = stringValue
	}

	// Apply constraints
	if constraint, exists := t.columnConstraints[columnIndex]; exists {
		content = t.applyCellConstraints(content, constraint)
	}

	// Apply styling
	return t.applyCellStyle(content, isCursor, isSelected)
}

func (t *Table) renderCellWithResult(value any, columnIndex int, rowID string, isCursor, isSelected bool, absoluteIndex int, col TableColumn) CellRenderResult {
	// Convert value to string for formatters
	stringValue := fmt.Sprintf("%v", value)

	// Use animated formatter if available and animations are enabled
	if formatter, exists := t.animatedFormatters[columnIndex]; exists && t.config.AnimationConfig.Enabled {
		key := fmt.Sprintf("%s:%d", rowID, columnIndex)
		animationState := make(map[string]any)
		if animation, exists := t.cellAnimations[key]; exists {
			animationState = animation.State
		}

		result := formatter(stringValue, absoluteIndex, columnIndex, col, t.renderContext, animationState, isCursor, isSelected, false, false)
		return CellRenderResult{
			Content:         result.Content,
			ActualWidth:     len(result.Content),
			ActualHeight:    1,
			Overflow:        false,
			RefreshTriggers: result.RefreshTriggers,
			AnimationState:  result.AnimationState,
			Error:           result.Error,
			Fallback:        result.Fallback,
		}
	} else if formatter, exists := t.cellFormatters[columnIndex]; exists {
		// Use regular cell formatter
		content := formatter(stringValue, absoluteIndex, columnIndex, col, t.renderContext, isCursor, isSelected)
		return CellRenderResult{
			Content:      content,
			ActualWidth:  len(content),
			ActualHeight: 1,
			Overflow:     false,
		}
	} else {
		// Default formatting
		return CellRenderResult{
			Content:      stringValue,
			ActualWidth:  len(stringValue),
			ActualHeight: 1,
			Overflow:     false,
		}
	}
}

func (t *Table) renderLoadingRow() string {
	var cells []string
	for i := range t.config.Columns {
		loading := "Loading..."
		if constraint, exists := t.columnConstraints[i]; exists {
			loading = t.renderContext.Truncate(loading, constraint.Width)
		}
		styled := t.config.Theme.LoadingStyle.Render(loading)
		cells = append(cells, styled)
	}

	if t.config.ShowBorders {
		return "│ " + strings.Join(cells, " │ ") + " │"
	}
	return strings.Join(cells, " ")
}

func (t *Table) renderBottomBorder() string {
	var parts []string

	parts = append(parts, "└")

	for i, col := range t.config.Columns {
		width := col.Width
		if constraint, exists := t.columnConstraints[i]; exists {
			width = constraint.Width
		}

		border := strings.Repeat("─", width)
		parts = append(parts, border)

		if i < len(t.config.Columns)-1 {
			parts = append(parts, "┴")
		}
	}

	parts = append(parts, "┘")

	return strings.Join(parts, "")
}

// ================================
// UTILITY HELPERS
// ================================

func (t *Table) setupRenderContext() {
	totalWidth := 0
	for _, col := range t.config.Columns {
		totalWidth += col.Width
	}
	if t.config.ShowBorders {
		totalWidth += len(t.config.Columns) + 1
	}

	t.renderContext = RenderContext{
		MaxWidth:       totalWidth,
		MaxHeight:      1,
		Theme:          &t.config.Theme,
		BaseStyle:      t.config.Theme.CellStyle,
		ColorSupport:   true,
		UnicodeSupport: true,
		CurrentTime:    time.Now(),
		FocusState:     FocusState{HasFocus: t.focused},
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
		Measure: func(text string) (int, int) {
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

func (t *Table) initializeColumnConstraints() {
	for i, col := range t.config.Columns {
		if _, exists := t.columnConstraints[i]; !exists {
			t.columnConstraints[i] = CellConstraint{
				Width:     col.Width,
				Height:    1,
				Alignment: col.Alignment,
				Padding: PaddingConfig{
					Left:   1,
					Right:  1,
					Top:    0,
					Bottom: 0,
				},
				MaxLines: 1,
			}
		}
	}
}

func (t *Table) reset() {
	t.chunks = make(map[int]Chunk[any])
	t.totalItems = 0
	t.clearSelection()
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
	t.cellAnimations = make(map[string]CellAnimation)
	t.rowAnimations = make(map[string]RowAnimation)
}

func (t *Table) updateViewportPosition() {
	if t.totalItems == 0 {
		return
	}

	height := t.config.ViewportConfig.Height

	t.viewport.CursorViewportIndex = t.viewport.CursorIndex - t.viewport.ViewportStartIndex

	if t.viewport.CursorViewportIndex < 0 {
		t.viewport.ViewportStartIndex = t.viewport.CursorIndex
		t.viewport.CursorViewportIndex = 0
	} else if t.viewport.CursorViewportIndex >= height {
		t.viewport.ViewportStartIndex = t.viewport.CursorIndex - height + 1
		t.viewport.CursorViewportIndex = height - 1
	}

	t.updateViewportBounds()
}

func (t *Table) updateViewportBounds() {
	height := t.config.ViewportConfig.Height
	topThreshold := t.config.ViewportConfig.TopThreshold
	bottomThreshold := t.config.ViewportConfig.BottomThreshold

	t.viewport.IsAtTopThreshold = t.viewport.CursorViewportIndex <= topThreshold
	// BottomThreshold is offset from end, so calculate actual position
	if bottomThreshold >= 0 {
		bottomPosition := height - bottomThreshold - 1
		t.viewport.IsAtBottomThreshold = t.viewport.CursorViewportIndex >= bottomPosition
	} else {
		t.viewport.IsAtBottomThreshold = false
	}

	t.viewport.AtDatasetStart = t.viewport.ViewportStartIndex == 0
	t.viewport.AtDatasetEnd = t.viewport.ViewportStartIndex+height >= t.totalItems
}

func (t *Table) checkAndLoadChunks() tea.Cmd {
	if t.dataSource == nil {
		return nil
	}

	var cmds []tea.Cmd
	height := t.config.ViewportConfig.Height
	chunkSize := t.config.ViewportConfig.ChunkSize

	if t.viewport.IsAtTopThreshold && t.viewport.ViewportStartIndex > 0 {
		startIndex := t.viewport.ViewportStartIndex - chunkSize
		if startIndex < 0 {
			startIndex = 0
		}

		if !t.isChunkLoaded(startIndex) {
			request := DataRequest{
				Start:          startIndex,
				Count:          chunkSize,
				SortFields:     t.sortFields,
				SortDirections: t.sortDirs,
				Filters:        t.filters,
			}
			cmds = append(cmds, t.dataSource.LoadChunk(request))
		}
	}

	if t.viewport.IsAtBottomThreshold && t.viewport.ViewportStartIndex+height < t.totalItems {
		startIndex := t.viewport.ViewportStartIndex + height

		if !t.isChunkLoaded(startIndex) {
			request := DataRequest{
				Start:          startIndex,
				Count:          chunkSize,
				SortFields:     t.sortFields,
				SortDirections: t.sortDirs,
				Filters:        t.filters,
			}
			cmds = append(cmds, t.dataSource.LoadChunk(request))
		}
	}

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}

	return nil
}

func (t *Table) isChunkLoaded(index int) bool {
	for _, chunk := range t.chunks {
		if index >= chunk.StartIndex && index <= chunk.EndIndex {
			return true
		}
	}
	return false
}

func (t *Table) getItemAtIndex(index int) (Data[any], bool) {
	if index < 0 || index >= t.totalItems {
		return Data[any]{}, false
	}

	for _, chunk := range t.chunks {
		if index >= chunk.StartIndex && index <= chunk.EndIndex {
			chunkIndex := index - chunk.StartIndex
			if chunkIndex < len(chunk.Items) {
				return chunk.Items[chunkIndex], true
			}
		}
	}

	return Data[any]{}, false
}

func (t *Table) getRowAtIndex(index int) (TableRow, bool) {
	item, exists := t.getItemAtIndex(index)
	if !exists {
		return TableRow{}, false
	}
	return t.convertItemToRow(item), true
}

func (t *Table) findItemIndex(id string) int {
	for _, chunk := range t.chunks {
		for i, item := range chunk.Items {
			if item.ID == id {
				return chunk.StartIndex + i
			}
		}
	}
	return -1
}

func (t *Table) toggleItemSelection(id string) tea.Cmd {
	if t.config.SelectionMode == SelectionNone {
		return nil
	}

	if t.selectedItems[id] {
		delete(t.selectedItems, id)
		for i, selectedID := range t.selectedOrder {
			if selectedID == id {
				t.selectedOrder = append(t.selectedOrder[:i], t.selectedOrder[i+1:]...)
				break
			}
		}
	} else {
		if t.config.SelectionMode == SelectionSingle {
			t.clearSelection()
		}
		t.selectedItems[id] = true
		t.selectedOrder = append(t.selectedOrder, id)
	}

	return nil
}

func (t *Table) clearSelection() {
	t.selectedItems = make(map[string]bool)
	t.selectedOrder = make([]string, 0)
}

func (t *Table) convertItemToRow(item Data[any]) TableRow {
	// Convert the generic item to a TableRow
	// TableRow has Cells as []string, so we need to convert based on columns
	cells := make([]string, len(t.config.Columns))

	// If item.Item is a map, extract values for each column
	if itemMap, ok := item.Item.(map[string]any); ok {
		for i, col := range t.config.Columns {
			if value, exists := itemMap[col.Field]; exists {
				cells[i] = fmt.Sprintf("%v", value)
			} else {
				cells[i] = ""
			}
		}
	} else {
		// Otherwise, put the item value in the first cell
		if len(cells) > 0 {
			cells[0] = fmt.Sprintf("%v", item.Item)
		}
	}

	return TableRow{
		ID:    item.ID,
		Cells: cells,
	}
}

func (t *Table) getCellValue(row TableRow, fieldName string) any {
	// Find the column index for the field name
	for i, col := range t.config.Columns {
		if col.Field == fieldName {
			if i < len(row.Cells) {
				return row.Cells[i]
			}
			return ""
		}
	}
	return ""
}

func (t *Table) applyCellConstraints(content string, constraint CellConstraint) string {
	// Apply width constraint
	if len(content) > constraint.Width {
		content = t.renderContext.Truncate(content, constraint.Width)
	}

	// Apply padding
	leftPadding := strings.Repeat(" ", constraint.Padding.Left)
	rightPadding := strings.Repeat(" ", constraint.Padding.Right)
	content = leftPadding + content + rightPadding

	// Ensure minimum width
	if len(content) < constraint.Width {
		switch constraint.Alignment {
		case AlignCenter:
			totalPadding := constraint.Width - len(content)
			leftPad := totalPadding / 2
			rightPad := totalPadding - leftPad
			content = strings.Repeat(" ", leftPad) + content + strings.Repeat(" ", rightPad)
		case AlignRight:
			padding := constraint.Width - len(content)
			content = strings.Repeat(" ", padding) + content
		default: // AlignLeft
			padding := constraint.Width - len(content)
			content = content + strings.Repeat(" ", padding)
		}
	}

	return content
}

func (t *Table) applyCellStyle(content string, isCursor, isSelected bool) string {
	var style lipgloss.Style

	switch {
	case isCursor && isSelected:
		style = t.config.Theme.SelectedStyle.Copy().
			Background(t.config.Theme.CursorStyle.GetBackground())
	case isCursor:
		style = t.config.Theme.CursorStyle
	case isSelected:
		style = t.config.Theme.SelectedStyle
	default:
		style = t.config.Theme.CellStyle
	}

	return style.Render(content)
}
