package table

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/data"
	"github.com/davidroman0O/vtable/render"
	"github.com/davidroman0O/vtable/viewport"
)

// TextSegment represents a segment of text (either visible text or ANSI escape codes)
type TextSegment struct {
	Text    string
	IsANSI  bool
	IsSpace bool
	Index   int
	Rune    rune
	Next    *TextSegment
	Prev    *TextSegment
	IsLower bool
}

// Table represents a pure Tea table component that reuses the List infrastructure
type Table struct {
	// Core state - reuse List infrastructure
	dataSource core.DataSource[any]
	chunks     map[int]core.Chunk[any] // Map of start index to chunk
	totalItems int

	// Viewport state - same as List
	viewport core.ViewportState

	// Configuration
	config core.TableConfig

	// Table-specific configuration
	columns              []core.TableColumn
	cellFormatters       map[int]core.SimpleCellFormatter // Column index -> simplified formatter
	rowFormatter         core.RowFormatter
	headerFormatter      core.HeaderFormatter
	headerCellFormatters map[int]core.SimpleHeaderFormatter // Column index -> header formatter
	loadingFormatter     core.LoadingRowFormatter
	renderContext        core.RenderContext

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
	visibleItems []core.Data[any]

	// Chunk access tracking for LRU management
	chunkAccessTime map[int]time.Time

	// Loading state tracking
	loadingChunks    map[int]bool
	hasLoadingChunks bool
	canScroll        bool

	// Component-based rendering system
	componentRenderer *TableComponentRenderer // Optional component-based renderer

	// Horizontal scrolling state
	horizontalScrollOffsets map[int]int // Column index -> scroll offset
	horizontalScrollMode    string      // "character", "word", "smart"
	scrollAllRows           bool        // true = scroll all rows together, false = only current row
	currentColumn           int         // Currently focused column for scrolling
	previousCursorIndex     int         // Track previous cursor position for scroll reset
}

// TableLayout handles proper column width calculation and cell alignment
type TableLayout struct {
	columns      []core.TableColumn
	totalWidth   int
	columnWidths []int
	borderWidth  int
}

// NewTableLayout creates a new table layout calculator
func NewTableLayout(columns []core.TableColumn, showBorders bool) *TableLayout {
	borderWidth := 0
	if showBorders {
		// Account for left border + column separators + right border
		borderWidth = 1 + (len(columns) - 1) + 1
	}

	layout := &TableLayout{
		columns:      columns,
		borderWidth:  borderWidth,
		columnWidths: make([]int, len(columns)),
	}

	// Calculate initial column widths
	layout.calculateColumnWidths()

	return layout
}

// calculateColumnWidths calculates the actual column widths
func (tl *TableLayout) calculateColumnWidths() {
	// Use the defined column widths directly
	for i, col := range tl.columns {
		tl.columnWidths[i] = col.Width
	}

	// Calculate total table width
	tl.totalWidth = tl.borderWidth
	for _, width := range tl.columnWidths {
		tl.totalWidth += width
	}
}

// RenderCell renders a cell with proper width and alignment constraints
func (tl *TableLayout) RenderCell(content string, columnIndex int, style lipgloss.Style) string {
	if columnIndex >= len(tl.columns) {
		return content
	}

	col := tl.columns[columnIndex]
	width := tl.columnWidths[columnIndex]

	// Apply width and alignment constraints
	constrainedContent := tl.applyColumnConstraints(content, width, col.Alignment)

	// Apply the style to the constrained content
	return style.Render(constrainedContent)
}

// applyColumnConstraints applies width and alignment constraints to content
func (tl *TableLayout) applyColumnConstraints(content string, width int, alignment int) string {
	// Clean up content (remove newlines, collapse spaces)
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", " ")
	content = strings.ReplaceAll(content, "\t", " ")
	for strings.Contains(content, "  ") {
		content = strings.ReplaceAll(content, "  ", " ")
	}
	content = strings.TrimSpace(content)

	// Measure actual display width
	actualWidth := render.MeasureText(content)

	// Truncate if too long
	if actualWidth > width {
		if width <= 0 {
			return ""
		}
		if width <= 3 {
			return strings.Repeat(".", width)
		}

		// Use ellipsis for truncation
		content = render.TruncateText(content, width)
		actualWidth = render.MeasureText(content)
	}

	// Apply alignment and padding
	return render.PadText(content, width, alignment)
}

// GetColumnWidths returns the calculated column widths
func (tl *TableLayout) GetColumnWidths() []int {
	return tl.columnWidths
}

// GetTotalWidth returns the total table width
func (tl *TableLayout) GetTotalWidth() int {
	return tl.totalWidth
}

// NewTable creates a new Table with the given configuration and data source
func NewTable(tableConfig core.TableConfig, dataSource core.DataSource[any]) *Table {
	// Validate and fix config
	errors := config.ValidateTableConfig(&tableConfig)
	if len(errors) > 0 {
		config.FixTableConfig(&tableConfig)
	}

	// Set default values for active cell indication if not specified
	if tableConfig.ActiveCellBackgroundColor == "" {
		tableConfig.ActiveCellBackgroundColor = "226" // Default bright yellow background
	}

	table := &Table{
		dataSource:           dataSource,
		chunks:               make(map[int]core.Chunk[any]),
		config:               tableConfig,
		columns:              tableConfig.Columns,
		cellFormatters:       make(map[int]core.SimpleCellFormatter),
		headerCellFormatters: make(map[int]core.SimpleHeaderFormatter),
		selectedItems:        make(map[string]bool),
		selectedOrder:        make([]string, 0),
		filters:              make(map[string]any),
		chunkAccessTime:      make(map[int]time.Time),
		visibleItems:         make([]core.Data[any], 0),
		loadingChunks:        make(map[int]bool),
		hasLoadingChunks:     false,
		canScroll:            true,
		componentRenderer:    NewTableComponentRenderer(DefaultComponentTableRenderConfig()), // Always enabled
		// Initialize horizontal scrolling state
		horizontalScrollOffsets: make(map[int]int),
		horizontalScrollMode:    "character",                             // Default to character-by-character
		scrollAllRows:           false,                                   // Default to scroll all rows together
		currentColumn:           0,                                       // Start with first column
		previousCursorIndex:     tableConfig.ViewportConfig.InitialIndex, // Track for scroll reset
		viewport: core.ViewportState{
			ViewportStartIndex:  0,
			CursorIndex:         tableConfig.ViewportConfig.InitialIndex,
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

// Init initializes the table model
func (t *Table) Init() tea.Cmd {
	return t.loadInitialData()
}

// Update handles all messages and updates the table state
func (t *Table) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// ===== Lifecycle Messages =====
	case core.InitMsg:
		return t, t.Init()

	case core.DestroyMsg:
		return t, nil

	case core.ResetMsg:
		t.reset()
		return t, t.Init()

	// ===== Navigation Messages - Reuse List logic =====
	case core.CursorUpMsg:
		cmd := t.handleCursorUp()
		return t, cmd

	case core.CursorDownMsg:
		cmd := t.handleCursorDown()
		return t, cmd

	case core.CursorLeftMsg:
		cmd := t.handlePrevColumn()
		return t, cmd

	case core.CursorRightMsg:
		cmd := t.handleNextColumn()
		return t, cmd

	case core.PageUpMsg:
		cmd := t.handlePageUp()
		return t, cmd

	case core.PageDownMsg:
		cmd := t.handlePageDown()
		return t, cmd

	case core.PageLeftMsg:
		cmd := t.handleHorizontalScrollPageLeft()
		return t, cmd

	case core.PageRightMsg:
		cmd := t.handleHorizontalScrollPageRight()
		return t, cmd

	case core.JumpToStartMsg:
		cmd := t.handleJumpToStart()
		return t, cmd

	case core.JumpToEndMsg:
		cmd := t.handleJumpToEnd()
		return t, cmd

	case core.JumpToMsg:
		cmd := t.handleJumpTo(msg.Index)
		return t, cmd

	// === HORIZONTAL SCROLLING MESSAGES ===
	case core.HorizontalScrollLeftMsg:
		cmd := t.handleHorizontalScrollLeft()
		return t, cmd

	case core.HorizontalScrollRightMsg:
		cmd := t.handleHorizontalScrollRight()
		return t, cmd

	case core.HorizontalScrollWordLeftMsg:
		cmd := t.handleHorizontalScrollWordLeft()
		return t, cmd

	case core.HorizontalScrollWordRightMsg:
		cmd := t.handleHorizontalScrollWordRight()
		return t, cmd

	case core.HorizontalScrollSmartLeftMsg:
		cmd := t.handleHorizontalScrollSmartLeft()
		return t, cmd

	case core.HorizontalScrollSmartRightMsg:
		cmd := t.handleHorizontalScrollSmartRight()
		return t, cmd

	case core.HorizontalScrollPageLeftMsg:
		cmd := t.handleHorizontalScrollPageLeft()
		return t, cmd

	case core.HorizontalScrollPageRightMsg:
		cmd := t.handleHorizontalScrollPageRight()
		return t, cmd

	case core.HorizontalScrollModeToggleMsg:
		cmd := t.handleToggleScrollMode()
		return t, cmd

	case core.HorizontalScrollScopeToggleMsg:
		cmd := t.handleToggleScrollScope()
		return t, cmd

	case core.HorizontalScrollResetMsg:
		cmd := t.handleResetScrolling()
		return t, cmd

	// === COLUMN NAVIGATION MESSAGES ===
	case core.NextColumnMsg:
		cmd := t.handleNextColumn()
		return t, cmd

	case core.PrevColumnMsg:
		cmd := t.handlePrevColumn()
		return t, cmd

	// ===== Data Messages - Reuse List logic =====
	case core.DataRefreshMsg:
		cmd := t.handleDataRefresh()
		return t, cmd

	case core.DataChunksRefreshMsg:
		t.chunks = make(map[int]core.Chunk[any])
		t.loadingChunks = make(map[int]bool)
		t.hasLoadingChunks = false
		t.canScroll = true
		return t, t.smartChunkManagement()

	case core.DataChunkLoadedMsg:
		cmd := t.handleDataChunkLoaded(msg)
		return t, cmd

	case core.DataChunkErrorMsg:
		t.lastError = msg.Error
		return t, core.ErrorCmd(msg.Error, "chunk_load")

	case core.DataTotalMsg:
		t.totalItems = msg.Total
		t.updateViewportBounds()
		t.viewport.ViewportStartIndex = 0
		t.viewport.CursorIndex = t.config.ViewportConfig.InitialIndex
		t.viewport.CursorViewportIndex = t.config.ViewportConfig.InitialIndex
		return t, t.smartChunkManagement()

	case core.DataTotalUpdateMsg:
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

	case core.DataLoadErrorMsg:
		t.lastError = msg.Error
		return t, core.ErrorCmd(msg.Error, "data_load")

	case core.DataTotalRequestMsg:
		if t.dataSource != nil {
			return t, t.dataSource.GetTotal()
		}
		return t, nil

	case core.DataSourceSetMsg:
		t.dataSource = msg.DataSource
		return t, t.dataSource.GetTotal()

	case core.ChunkUnloadedMsg:
		// Handle chunk unloaded notification (for UI feedback)
		return t, nil

	// ===== Selection Messages - Reuse List logic =====
	case core.SelectCurrentMsg:
		cmd := t.handleSelectCurrent()
		return t, cmd

	case core.SelectToggleMsg:
		cmd := t.handleSelectToggle(msg.Index)
		return t, cmd

	case core.SelectAllMsg:
		cmd := t.handleSelectAll()
		return t, cmd

	case core.SelectClearMsg:
		if t.dataSource == nil {
			return t, nil
		}
		return t, t.dataSource.ClearSelection()

	case core.SelectRangeMsg:
		cmd := t.handleSelectRange(msg.StartID, msg.EndID)
		return t, cmd

	case core.SelectionModeSetMsg:
		t.config.SelectionMode = msg.Mode
		if msg.Mode == core.SelectionNone {
			t.clearSelection()
		}
		return t, nil

	case core.SelectionResponseMsg:
		cmd := t.refreshChunks()
		return t, cmd

	// ===== Table-specific Messages =====
	case core.ColumnSetMsg:
		t.columns = msg.Columns
		t.config.Columns = msg.Columns
		return t, nil

	case core.ColumnUpdateMsg:
		if msg.Index >= 0 && msg.Index < len(t.columns) {
			t.columns[msg.Index] = msg.Column
			t.config.Columns[msg.Index] = msg.Column
		}
		return t, nil

	case core.HeaderVisibilityMsg:
		t.config.ShowHeader = msg.Visible
		return t, nil

	case core.BorderVisibilityMsg:
		t.config.ShowBorders = msg.Visible
		return t, nil

	case core.TopBorderVisibilityMsg:
		t.config.ShowTopBorder = msg.Visible
		return t, nil

	case core.BottomBorderVisibilityMsg:
		t.config.ShowBottomBorder = msg.Visible
		return t, nil

	case core.HeaderSeparatorVisibilityMsg:
		t.config.ShowHeaderSeparator = msg.Visible
		return t, nil

	case core.TopBorderSpaceRemovalMsg:
		t.config.RemoveTopBorderSpace = msg.Remove
		return t, nil

	case core.BottomBorderSpaceRemovalMsg:
		t.config.RemoveBottomBorderSpace = msg.Remove
		return t, nil

	case core.CellFormatterSetMsg:
		if msg.ColumnIndex >= 0 {
			t.cellFormatters[msg.ColumnIndex] = msg.Formatter
		}
		return t, nil

	case core.RowFormatterSetMsg:
		t.loadingFormatter = msg.Formatter
		return t, nil

	case core.HeaderFormatterSetMsg:
		t.headerCellFormatters[msg.ColumnIndex] = msg.Formatter
		return t, nil

	case core.LoadingFormatterSetMsg:
		t.loadingFormatter = msg.Formatter
		return t, nil

	case core.HeaderCellFormatterSetMsg:
		// Convert old HeaderCellFormatter to SimpleHeaderFormatter
		oldFormatter := msg.Formatter
		t.headerCellFormatters[0] = func(column core.TableColumn, ctx core.RenderContext) string {
			return oldFormatter(column, 0, ctx) // Pass 0 as columnIndex
		}
		return t, nil

	case core.TableThemeSetMsg:
		t.config.Theme = msg.Theme
		return t, nil

	case core.FullRowHighlightToggleMsg:
		t.config.FullRowHighlighting = !t.config.FullRowHighlighting
		return t, nil

	case core.FullRowHighlightEnableMsg:
		t.config.FullRowHighlighting = msg.Enabled
		return t, nil

	case core.ActiveCellIndicationModeSetMsg:
		t.config.ActiveCellIndicationEnabled = msg.Enabled
		return t, nil

	case core.ActiveCellBackgroundColorSetMsg:
		t.config.ActiveCellBackgroundColor = msg.Color
		return t, nil

	// ===== Configuration Messages =====
	case core.ViewportConfigMsg:
		t.config.ViewportConfig = msg.Config
		t.updateViewportBounds()
		return t, nil

	case core.KeyMapSetMsg:
		t.config.KeyMap = msg.KeyMap
		return t, nil

	// ===== Filter Messages - Reuse List logic =====
	case core.FilterSetMsg:
		t.filters[msg.Field] = msg.Value
		cmd := t.handleFilterChange()
		return t, cmd

	case core.FilterClearMsg:
		delete(t.filters, msg.Field)
		cmd := t.handleFilterChange()
		return t, cmd

	case core.FiltersClearAllMsg:
		t.filters = make(map[string]any)
		cmd := t.handleFilterChange()
		return t, cmd

	// ===== Sort Messages - Reuse List logic =====
	case core.SortToggleMsg:
		cmd := t.handleSortToggle(msg.Field)
		return t, cmd

	case core.SortSetMsg:
		cmd := t.handleSortSet(msg.Field, msg.Direction)
		return t, cmd

	case core.SortAddMsg:
		cmd := t.handleSortAdd(msg.Field, msg.Direction)
		return t, cmd

	case core.SortRemoveMsg:
		cmd := t.handleSortRemove(msg.Field)
		return t, cmd

	case core.SortsClearAllMsg:
		t.sortFields = nil
		t.sortDirs = nil
		cmd := t.handleFilterChange()
		return t, cmd

	// ===== Focus Messages =====
	case core.FocusMsg:
		t.focused = true
		return t, nil

	case core.BlurMsg:
		t.focused = false
		return t, nil

	// ===== Search Messages =====
	case core.SearchSetMsg:
		t.searchQuery = msg.Query
		t.searchField = msg.Field
		cmd := t.handleSearch()
		return t, cmd

	case core.SearchClearMsg:
		t.searchQuery = ""
		t.searchField = ""
		t.searchResults = nil
		return t, nil

	case core.SearchResultMsg:
		t.searchResults = msg.Results
		return t, nil

	// ===== Error Messages =====
	case core.ErrorMsg:
		t.lastError = msg.Error
		return t, nil

	// ===== Viewport Messages =====
	case core.ViewportResizeMsg:
		t.config.ViewportConfig.Height = msg.Height
		t.updateViewportBounds()
		return t, nil

	// ===== Batch Messages =====
	case core.BatchMsg:
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

	// Add top border if enabled
	if t.config.ShowTopBorder && !t.config.RemoveTopBorderSpace {
		builder.WriteString(t.constructTopBorder())
		builder.WriteString("\n")
	}

	// Render header if enabled
	if t.config.ShowHeader {
		header := t.renderHeader()
		if header != "" {
			builder.WriteString(header)
			builder.WriteString("\n")

			// Add header separator border if enabled
			if t.config.ShowHeaderSeparator {
				builder.WriteString(t.constructHeaderSeparator())
				builder.WriteString("\n")
			}
		}
	}

	// Ensure visible items are up to date
	t.updateVisibleItems()

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

	// Add bottom border if enabled
	if t.config.ShowBottomBorder && !t.config.RemoveBottomBorderSpace {
		builder.WriteString("\n")
		builder.WriteString(t.constructBottomBorder())
	}

	return builder.String()
}

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
func (t *Table) GetState() core.ViewportState {
	return t.viewport
}

// GetTotalItems returns the total number of items
func (t *Table) GetTotalItems() int {
	return t.totalItems
}

// GetSelectionCount returns the number of selected items
func (t *Table) GetSelectionCount() int {
	var count int
	for _, chunk := range t.chunks {
		for _, item := range chunk.Items {
			if item.Selected {
				count++
			}
		}
	}
	return count
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
func (t *Table) GetCurrentRow() (core.TableRow, bool) {
	item, exists := t.getItemAtIndex(t.viewport.CursorIndex)
	if !exists {
		return core.TableRow{}, false
	}

	// Convert Data[any] to TableRow
	if row, ok := item.Item.(core.TableRow); ok {
		return row, true
	}

	// If item is not a TableRow, try to convert it
	return core.TableRow{}, false
}

// setupRenderContext initializes the render context
func (t *Table) setupRenderContext() {
	t.renderContext = core.RenderContext{
		MaxWidth:       120, // Default table width
		MaxHeight:      1,   // Single line for table rows
		Theme:          &t.config.Theme,
		BaseStyle:      t.config.Theme.CellStyle,
		ColorSupport:   true,
		UnicodeSupport: true,
		CurrentTime:    time.Now(),
		FocusState:     core.FocusState{HasFocus: t.focused},

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
	t.chunks = make(map[int]core.Chunk[any])
	t.totalItems = 0
	t.loadingChunks = make(map[int]bool)
	t.hasLoadingChunks = false
	t.canScroll = true
	t.viewport = core.ViewportState{
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

// handleScrollResetOnNavigation resets scroll offsets when navigating between rows if enabled
func (t *Table) handleScrollResetOnNavigation() {
	// Only reset if the feature is enabled and we actually moved to a different row
	if !t.config.ResetScrollOnNavigation || t.viewport.CursorIndex == t.previousCursorIndex {
		return
	}

	// Reset scroll offsets when ResetScrollOnNavigation is enabled
	// The scrollAllRows scope only affects which cells get scrolled during rendering,
	// not whether to reset on navigation
	t.horizontalScrollOffsets = make(map[int]int)

	// Update the previous cursor position
	t.previousCursorIndex = t.viewport.CursorIndex
}

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
	t.viewport = viewport.CalculateCursorUp(t.viewport, t.config.ViewportConfig, t.totalItems)

	// Handle scroll reset if enabled and cursor position changed
	t.handleScrollResetOnNavigation()

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
	t.viewport = viewport.CalculateCursorDown(t.viewport, t.config.ViewportConfig, t.totalItems)

	// Handle scroll reset if enabled and cursor position changed
	t.handleScrollResetOnNavigation()

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
	t.viewport = viewport.CalculatePageUp(t.viewport, t.config.ViewportConfig, t.totalItems)

	// Handle scroll reset if enabled and cursor position changed
	t.handleScrollResetOnNavigation()

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
	t.viewport = viewport.CalculatePageDown(t.viewport, t.config.ViewportConfig, t.totalItems)

	// Handle scroll reset if enabled and cursor position changed
	t.handleScrollResetOnNavigation()

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

	t.viewport = viewport.CalculateJumpToStart(t.config.ViewportConfig, t.totalItems)

	// Handle scroll reset if enabled and cursor position changed
	t.handleScrollResetOnNavigation()

	return t.smartChunkManagement()
}

// handleJumpToEnd moves cursor to the end
func (t *Table) handleJumpToEnd() tea.Cmd {
	if t.totalItems <= 0 || !t.canScroll {
		return nil
	}

	previousState := t.viewport
	t.viewport = viewport.CalculateJumpToEnd(t.config.ViewportConfig, t.totalItems)

	// Handle scroll reset if enabled and cursor position changed
	t.handleScrollResetOnNavigation()

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

	t.viewport = viewport.CalculateJumpTo(index, t.config.ViewportConfig, t.totalItems)

	// Handle scroll reset if enabled and cursor position changed
	t.handleScrollResetOnNavigation()

	return t.smartChunkManagement()
}

// handleDataRefresh refreshes all data
func (t *Table) handleDataRefresh() tea.Cmd {
	t.chunks = make(map[int]core.Chunk[any])

	if t.dataSource == nil {
		return nil
	}

	var cmds []tea.Cmd
	cmds = append(cmds, t.dataSource.GetTotal())

	return tea.Batch(cmds...)
}

// handleDataChunkLoaded processes a loaded data chunk
func (t *Table) handleDataChunkLoaded(msg core.DataChunkLoadedMsg) tea.Cmd {
	chunk := core.Chunk[any]{
		StartIndex: msg.StartIndex,
		EndIndex:   msg.StartIndex + len(msg.Items) - 1,
		Items:      msg.Items,
		LoadedAt:   time.Now(),

		Request: msg.Request,
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

	cmds = append(cmds, core.ChunkLoadingCompletedCmd(msg.StartIndex, len(msg.Items), msg.Request))

	if unloadCmd := t.unloadOldChunks(); unloadCmd != nil {
		cmds = append(cmds, unloadCmd)
	}

	return tea.Batch(cmds...)
}

// handleSelectCurrent selects the current item
func (t *Table) handleSelectCurrent() tea.Cmd {
	if t.config.SelectionMode == core.SelectionNone || t.totalItems == 0 {
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
	if t.config.SelectionMode == core.SelectionNone || index < 0 || index >= t.totalItems {
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
	if t.config.SelectionMode != core.SelectionMultiple || t.dataSource == nil {
		return nil
	}

	return t.dataSource.SelectAll()
}

// handleSelectRange selects a range of items
func (t *Table) handleSelectRange(startID, endID string) tea.Cmd {
	if t.config.SelectionMode != core.SelectionMultiple {
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

// handleFilterChange triggers data refresh when filters change
func (t *Table) handleFilterChange() tea.Cmd {
	return t.handleDataRefresh()
}

// handleSortToggle toggles sorting on a field
func (t *Table) handleSortToggle(field string) tea.Cmd {
	// Simplified implementation - just toggle between asc/desc for now
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
	// Field not found, add it
	t.sortFields = append(t.sortFields, field)
	t.sortDirs = append(t.sortDirs, "asc")
	return t.handleDataRefresh()
}

// handleSortSet sets sorting on a field
func (t *Table) handleSortSet(field, direction string) tea.Cmd {
	t.sortFields = []string{field}
	t.sortDirs = []string{direction}
	return t.handleDataRefresh()
}

// handleSortAdd adds a sort field
func (t *Table) handleSortAdd(field, direction string) tea.Cmd {
	t.sortFields = append(t.sortFields, field)
	t.sortDirs = append(t.sortDirs, direction)
	return t.handleDataRefresh()
}

// handleSortRemove removes a sort field
func (t *Table) handleSortRemove(field string) tea.Cmd {
	for i, sortField := range t.sortFields {
		if sortField == field {
			t.sortFields = append(t.sortFields[:i], t.sortFields[i+1:]...)
			t.sortDirs = append(t.sortDirs[:i], t.sortDirs[i+1:]...)
			break
		}
	}
	return t.handleDataRefresh()
}

// handleSearch performs a search
func (t *Table) handleSearch() tea.Cmd {
	if t.dataSource == nil {
		return nil
	}

	return core.SearchResultCmd([]int{}, t.searchQuery, 0)
}

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
			return core.SelectCurrentCmd()
		}
	}

	for _, selectAllKey := range t.config.KeyMap.SelectAll {
		if key == selectAllKey {
			return core.SelectAllCmd()
		}
	}

	// // === HORIZONTAL SCROLLING KEYS ===
	// switch key {
	// case "left":
	// 	return t.handleHorizontalScrollLeft()
	// case "right":
	// 	return t.handleHorizontalScrollRight()
	// case "[":
	// 	return t.handleHorizontalScrollWordLeft()
	// case "]":
	// 	return t.handleHorizontalScrollWordRight()
	// case "{":
	// 	return t.handleHorizontalScrollSmartLeft()
	// case "}":
	// 	return t.handleHorizontalScrollSmartRight()
	// case ".":
	// 	return t.handleNextColumn()
	// case ",":
	// 	return t.handlePrevColumn()
	// case "m", "M":
	// 	return t.handleToggleScrollMode()
	// case "v", "V":
	// 	return t.handleToggleScrollScope()
	// case "backspace", "delete":
	// 	return t.handleResetScrolling()
	// }

	return nil
}

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

	// Add indicator column header since component renderer is always enabled
	indicatorWidth := 4
	indicatorHeader := "●" // Use a dot/bullet as indicator

	// Create constraint for indicator header
	indicatorConstraint := core.CellConstraint{
		Width:     indicatorWidth,
		Height:    1,
		Alignment: core.AlignCenter,
	}

	styledIndicatorHeader := t.applyCellConstraints(indicatorHeader, indicatorConstraint, -1) // Use -1 for indicator column
	styledIndicatorHeader = t.config.Theme.HeaderStyle.Render(styledIndicatorHeader)
	parts = append(parts, styledIndicatorHeader)

	for i, col := range t.columns {
		var headerText string

		// Use HeaderCellFormatter if available for this specific column
		// IMPORTANT: Use the original column index i, NOT shifted by indicator column
		if formatter, exists := t.headerCellFormatters[i]; exists {
			// Get the formatted header content
			formattedHeader := formatter(col, t.renderContext)

			// Determine which alignment and constraint to use
			headerAlignment := col.HeaderAlignment
			if headerAlignment == 0 {
				headerAlignment = col.Alignment // Fall back to column alignment if not specified
			}

			// Use header constraint if specified, otherwise create default constraint
			var constraint core.CellConstraint
			if col.HeaderConstraint.Width > 0 || col.HeaderConstraint.Alignment > 0 {
				constraint = col.HeaderConstraint
				// Override alignment if HeaderAlignment is specified
				if headerAlignment != 0 {
					constraint.Alignment = headerAlignment
				}
			} else {
				// Create default constraint with header alignment
				constraint = core.CellConstraint{
					Width:     col.Width,
					Height:    1,
					Alignment: headerAlignment,
				}
			}

			// CRITICAL: Apply constraints to the formatted header to ensure exact column width
			constrainedHeader := t.applyCellConstraints(formattedHeader, constraint, -1) // Use -1 to skip horizontal scrolling for headers

			// Apply header styling to the constrained content
			headerText = t.config.Theme.HeaderStyle.Render(constrainedHeader)
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
			var constraint core.CellConstraint
			if col.HeaderConstraint.Width > 0 || col.HeaderConstraint.Alignment > 0 {
				constraint = col.HeaderConstraint
				// Override alignment if HeaderAlignment is specified
				if headerAlignment != 0 {
					constraint.Alignment = headerAlignment
				}
			} else {
				// Create default constraint with header alignment
				constraint = core.CellConstraint{
					Width:     col.Width,
					Height:    1,
					Alignment: headerAlignment,
				}
			}

			// Apply constraints to header text
			styledHeader := t.applyCellConstraints(headerText, constraint, -1) // Use -1 to skip horizontal scrolling for headers

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

// renderRow renders a single table row using proper table layout
func (t *Table) renderRow(item core.Data[any], absoluteIndex int, isCursor bool) string {
	// Handle loading placeholders with custom formatter
	if strings.HasPrefix(item.ID, "loading-") || strings.HasPrefix(item.ID, "missing-") {
		if t.loadingFormatter != nil {
			return t.loadingFormatter(absoluteIndex, t.columns, t.renderContext, isCursor)
		}
		// Default loading behavior - show empty cells with proper column widths
		return t.renderDefaultLoadingRow(absoluteIndex, isCursor)
	}

	// Convert item to TableRow
	var row core.TableRow
	if r, ok := item.Item.(core.TableRow); ok {
		row = r
	} else {
		// Create a single-cell row if item is not a TableRow
		row = core.TableRow{
			ID:    item.ID,
			Cells: []string{fmt.Sprintf("%v", item.Item)},
		}
	}

	// Use component rendering for all rows - it's now the only system
	var parts []string

	// FIRST: Add a separate indicator column for cursor/selection
	indicatorWidth := 4 // Width for "► ✓ "
	var indicatorContent string

	// Build indicators separately from content
	if isCursor && item.Selected {
		indicatorContent = "►✓"
	} else if isCursor {
		indicatorContent = "► "
	} else if item.Selected {
		indicatorContent = " ✓"
	} else {
		indicatorContent = "  "
	}

	// Apply constraint to indicator column
	indicatorConstraint := core.CellConstraint{
		Width:     indicatorWidth,
		Height:    1,
		Alignment: core.AlignCenter,
	}
	constrainedIndicator := t.applyCellConstraints(indicatorContent, indicatorConstraint, -1) // Use -1 for indicator column

	// Style the indicator column
	var styledIndicator string
	if isCursor {
		styledIndicator = t.config.Theme.CursorStyle.Render(constrainedIndicator)
	} else if item.Selected {
		styledIndicator = t.config.Theme.SelectedStyle.Render(constrainedIndicator)
	} else {
		styledIndicator = t.config.Theme.CellStyle.Render(constrainedIndicator)
	}

	parts = append(parts, styledIndicator)

	// THEN: Render each actual data cell WITHOUT contamination
	for i, col := range t.columns {
		var cellValue string
		if i < len(row.Cells) {
			cellValue = row.Cells[i]
		}

		// Apply cell formatter to original content (NO prefix contamination!)
		var formattedContent string
		if formatter, exists := t.cellFormatters[i]; exists {
			isActiveCell := t.isActiveCell(i, isCursor)
			formattedContent = formatter(cellValue, absoluteIndex, col, t.renderContext, isCursor, item.Selected, isActiveCell)
		} else {
			formattedContent = cellValue
		}

		// Apply cell constraints to maintain column width
		constraint := core.CellConstraint{
			Width:     col.Width,
			Height:    1,
			Alignment: col.Alignment,
		}

		constrainedContent := t.applyCellConstraintsWithRowInfo(formattedContent, constraint, i, isCursor)

		// When full-row highlighting is on, the active cell indication must be layered on top.
		// This block ensures the active cell's background overrides the full-row highlight.
		// Apply full row highlighting if enabled (overrides all other styling)
		var styledCell string
		if t.config.FullRowHighlighting && isCursor {
			// Full row highlighting takes over - strip existing styling and apply uniform background
			plainContent := stripANSI(constrainedContent)
			fullRowStyle := t.config.Theme.FullRowCursorStyle

			// Check for active cell and override background if needed
			isActiveCell := t.isActiveCell(i, isCursor)
			if isActiveCell && t.config.ActiveCellIndicationEnabled {
				// Active cell background overrides full row cursor background
				activeCellStyle := fullRowStyle.Copy().
					Background(lipgloss.Color(t.config.ActiveCellBackgroundColor))
				styledCell = activeCellStyle.Render(plainContent)
			} else {
				styledCell = fullRowStyle.Render(plainContent)
			}
		} else if item.Selected {
			// Apply full-row selection styling - strip existing styling and apply uniform selection background
			plainContent := stripANSI(constrainedContent)
			selectionStyle := t.config.Theme.SelectedStyle
			styledCell = selectionStyle.Render(plainContent)
		} else if isCursor {
			// Check if this is an active cell that should override cursor styling
			isActiveCell := t.isActiveCell(i, isCursor)
			if isActiveCell && t.config.ActiveCellIndicationEnabled {
				// Active cell background overrides cursor background
				activeCellStyle := lipgloss.NewStyle().
					Background(lipgloss.Color(t.config.ActiveCellBackgroundColor)).
					Foreground(t.config.Theme.CursorStyle.GetForeground())
				styledCell = activeCellStyle.Render(stripANSI(constrainedContent))
			} else {
				// Apply normal cursor styling to formatted content
				styledCell = t.config.Theme.CursorStyle.Render(constrainedContent)
			}
		} else {
			// Use the formatted and constrained content as-is
			styledCell = constrainedContent
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
func (t *Table) renderCellsForRow(row core.TableRow, absoluteIndex int, isCursor, isSelected bool) []core.CellRenderResult {
	var results []core.CellRenderResult

	for i, col := range t.columns {
		var cellValue string
		if i < len(row.Cells) {
			cellValue = row.Cells[i]
		}

		// Use regular formatter or default
		var finalCellValue string
		if formatter, exists := t.cellFormatters[i]; exists {
			isActiveCell := t.isActiveCell(i, isCursor)
			formattedValue := formatter(cellValue, absoluteIndex, col, t.renderContext, isCursor, isSelected, isActiveCell)

			// Apply full row highlighting if enabled (overrides formatter styling)
			if t.config.FullRowHighlighting && isCursor {
				// Full row highlighting takes over - strip existing styling and apply uniform background
				plainContent := stripANSI(formattedValue)
				fullRowStyle := t.config.Theme.FullRowCursorStyle

				isActiveCell := t.isActiveCell(i, isCursor)
				if isActiveCell && t.config.ActiveCellIndicationEnabled {
					activeCellStyle := fullRowStyle.Copy().
						Background(lipgloss.Color(t.config.ActiveCellBackgroundColor))
					finalCellValue = activeCellStyle.Render(plainContent)
				} else {
					finalCellValue = fullRowStyle.Render(plainContent)
				}
			} else {
				finalCellValue = formattedValue
			}
		} else {
			// Apply full row highlighting if enabled, otherwise use plain value
			if t.config.FullRowHighlighting && isCursor {
				plainContent := stripANSI(cellValue)
				fullRowStyle := t.config.Theme.FullRowCursorStyle

				isActiveCell := t.isActiveCell(i, isCursor)
				if isActiveCell && t.config.ActiveCellIndicationEnabled {
					activeCellStyle := fullRowStyle.Copy().
						Background(lipgloss.Color(t.config.ActiveCellBackgroundColor))
					finalCellValue = activeCellStyle.Render(plainContent)
				} else {
					finalCellValue = fullRowStyle.Render(plainContent)
				}
			} else {
				finalCellValue = cellValue
			}
		}

		result := core.CellRenderResult{
			Content:         finalCellValue,
			RefreshTriggers: nil,
			AnimationState:  nil,
			Error:           nil,
			Fallback:        finalCellValue,
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
		constraint := core.CellConstraint{
			Width:     col.Width,
			Height:    1,
			Alignment: col.Alignment,
		}

		// Use loading indicator or empty space
		loadingText := ""
		if col.Width >= 10 {
			loadingText = "Loading..."
		}

		constrainedContent := t.applyCellConstraints(loadingText, constraint, -1) // Use -1 for loading cells

		// Apply styling - full row highlighting takes precedence
		var styledCell string
		if t.config.FullRowHighlighting && isCursor {
			// Apply full row highlighting to loading rows too
			fullRowStyle := t.config.Theme.FullRowCursorStyle
			styledCell = fullRowStyle.Render(constrainedContent)
		} else if isCursor {
			styledCell = t.config.Theme.CursorStyle.Render(constrainedContent)
		} else {
			styledCell = t.config.Theme.CellStyle.Render(constrainedContent)
		}

		parts = append(parts, styledCell)
	}

	result := strings.Join(parts, t.getBorderChar())

	if t.config.ShowBorders {
		result = t.getBorderChar() + result + t.getBorderChar()
	}

	return result
}

// applyCellConstraints applies width and alignment constraints to cell content
func (t *Table) applyCellConstraints(text string, constraint core.CellConstraint, columnIndex int) string {
	return t.applyCellConstraintsWithRowInfo(text, constraint, columnIndex, false)
}

// applyCellConstraintsWithRowInfo applies width and alignment constraints with row context for scope-aware scrolling
func (t *Table) applyCellConstraintsWithRowInfo(text string, constraint core.CellConstraint, columnIndex int, isCurrentRow bool) string {
	width := constraint.Width

	// For cursor row mode, only apply scrolling to the cell that's BOTH:
	// 1. On the cursor row (isCurrentRow = true)
	// 2. AND in the currently focused column (columnIndex == t.currentColumn)
	var shouldApplyHorizontalScrolling bool
	if t.scrollAllRows {
		// All rows mode: scroll all cells in the focused column
		shouldApplyHorizontalScrolling = columnIndex == t.currentColumn
	} else {
		// Current row mode: only scroll the cell at cursor row + focused column
		shouldApplyHorizontalScrolling = isCurrentRow && columnIndex == t.currentColumn
	}

	// Apply horizontal scrolling first (before any constraints)
	var scrolledText string
	if shouldApplyHorizontalScrolling {
		scrolledText = t.applyHorizontalScrollWithScope(text, columnIndex, isCurrentRow)
	} else {
		scrolledText = text // No horizontal scrolling for this cell
	}

	// Track original text for ellipsis decision
	originalText := text

	// Check if horizontal scrolling was applied (i.e., we have a scroll offset)
	hasHorizontalScrolling := false
	if shouldApplyHorizontalScrolling && columnIndex >= 0 && t.horizontalScrollOffsets[columnIndex] > 0 {
		hasHorizontalScrolling = true
	}

	// Clean up scrolled content (same as before but on scrolled text)
	scrolledText = strings.ReplaceAll(scrolledText, "\n", " ")
	scrolledText = strings.ReplaceAll(scrolledText, "\r", " ")
	scrolledText = strings.ReplaceAll(scrolledText, "\t", " ")
	for strings.Contains(scrolledText, "  ") {
		scrolledText = strings.ReplaceAll(scrolledText, "  ", " ")
	}

	// CRITICAL FIX: Don't trim leading spaces when horizontal scrolling is active
	// Those leading spaces are intentional results of scrolling to show them
	if !hasHorizontalScrolling {
		scrolledText = strings.TrimSpace(scrolledText)
	} else {
		// Only trim trailing spaces when horizontal scrolling is active
		scrolledText = strings.TrimRight(scrolledText, " \t")
	}

	// Detect if text contains ANSI escape codes (styling)
	hasANSI := strings.Contains(scrolledText, "\x1b")

	// Choose appropriate width measurement function
	var measureWidth func(string) int
	var truncateFunc func(string, int, string) string

	if hasANSI {
		// For styled text: use lipgloss for ANSI-aware width measurement
		measureWidth = lipgloss.Width
		truncateFunc = ansiTruncate
	} else {
		// For plain text: use runewidth for proper Unicode handling
		measureWidth = runewidth.StringWidth
		truncateFunc = ansiTruncateWithRunewidth
	}

	// Check if we need to truncate
	if measureWidth(scrolledText) > width {
		// Determine if we should show ellipsis
		showEllipsis := t.shouldShowEllipsis(originalText, columnIndex, scrolledText, isCurrentRow)

		if showEllipsis {
			scrolledText = truncateFunc(scrolledText, width, "...")
		} else {
			// No ellipsis - we're at the end of the content

			scrolledText = truncateFunc(scrolledText, width, "")
		}
	}

	// Then apply alignment and padding to exact width
	actualWidth := measureWidth(scrolledText)
	if actualWidth < width {
		padding := width - actualWidth

		if hasANSI {
			// For styled text, we need to extend the styling to cover padding
			paddingSpaces := strings.Repeat(" ", padding)

			// Extract background color from the styled text to apply to padding
			backgroundStyle := extractBackgroundStyle(scrolledText)
			var styledPadding string
			if backgroundStyle != "" {
				// Apply same background to padding spaces
				styledPadding = "\x1b[" + backgroundStyle + "m" + paddingSpaces + "\x1b[0m"
			} else {
				styledPadding = paddingSpaces
			}

			switch constraint.Alignment {
			case core.AlignRight:
				scrolledText = styledPadding + scrolledText
			case core.AlignCenter:
				leftPad := padding / 2
				rightPad := padding - leftPad
				leftPaddingSpaces := strings.Repeat(" ", leftPad)
				rightPaddingSpaces := strings.Repeat(" ", rightPad)

				var leftStyledPadding, rightStyledPadding string
				if backgroundStyle != "" {
					leftStyledPadding = "\x1b[" + backgroundStyle + "m" + leftPaddingSpaces + "\x1b[0m"
					rightStyledPadding = "\x1b[" + backgroundStyle + "m" + rightPaddingSpaces + "\x1b[0m"
				} else {
					leftStyledPadding = leftPaddingSpaces
					rightStyledPadding = rightPaddingSpaces
				}

				scrolledText = leftStyledPadding + scrolledText + rightStyledPadding
			default: // core.AlignLeft
				scrolledText = scrolledText + styledPadding
			}
		} else {
			// For plain text, use regular padding
			switch constraint.Alignment {
			case core.AlignRight:
				scrolledText = strings.Repeat(" ", padding) + scrolledText
			case core.AlignCenter:
				leftPad := padding / 2
				rightPad := padding - leftPad
				scrolledText = strings.Repeat(" ", leftPad) + scrolledText + strings.Repeat(" ", rightPad)
			default: // core.AlignLeft
				scrolledText = scrolledText + strings.Repeat(" ", padding)
			}
		}
	}

	// Apply active cell indication for built-in modes (brackets, background)
	// Note: Custom mode is handled in the formatter itself
	isActiveCellForIndication := t.isActiveCell(columnIndex, isCurrentRow)
	scrolledText = t.applyActiveCellIndication(scrolledText, isActiveCellForIndication)

	return scrolledText
}

// shouldShowEllipsis determines if ellipsis should be shown based on scroll position and scope
func (t *Table) shouldShowEllipsis(originalText string, columnIndex int, scrolledText string, isCurrentRow bool) bool {
	// Check if this specific row/column combination is actually being scrolled
	var isBeingScrolled bool
	if t.scrollAllRows {
		// All rows mode: all rows in the focused column are scrolled
		isBeingScrolled = columnIndex == t.currentColumn
	} else {
		// Current row mode: only the cursor row in the focused column is scrolled
		isBeingScrolled = isCurrentRow && columnIndex == t.currentColumn
	}

	// If this row/column is not being scrolled, use normal ellipsis logic
	if !isBeingScrolled {
		return true // Show ellipsis for non-scrolled cells that are truncated
	}

	scrollOffset := t.horizontalScrollOffsets[columnIndex]

	// If no scrolling, use normal ellipsis logic
	if scrollOffset <= 0 {
		return true
	}

	// Get the column width to determine how much content is visible
	columnWidth := 25 // Default, should get from actual column
	if columnIndex < len(t.columns) {
		columnWidth = t.columns[columnIndex].Width
	}

	// Check if we're at the end of the content based on scroll mode
	switch t.horizontalScrollMode {
	case "word":
		words := strings.Fields(originalText)
		// If we've scrolled past or to the last few words, don't show ellipsis
		return scrollOffset < len(words)-2
	case "smart":
		boundaries := t.findSmartBoundariesInSegments(t.parseTextWithANSI(originalText))
		// If we're at the last boundary, don't show ellipsis
		return scrollOffset < len(boundaries)-1
	default: // "character"
		runes := []rune(originalText)
		remainingContentLength := len(runes) - scrollOffset

		// If the remaining content (after scrolling) can't fill the column width,
		// then we're near the end and shouldn't show ellipsis
		// Reserve 3 characters for the ellipsis itself
		return remainingContentLength > columnWidth
	}
}

// extractBackgroundStyle extracts background color codes from ANSI styled text
func extractBackgroundStyle(text string) string {
	// Look for background color patterns: 48;5;{color}, 48;2;{r};{g};{b}, or standard codes like 40-47

	// Find ANSI escape sequences
	start := strings.Index(text, "\x1b[")
	if start == -1 {
		return ""
	}

	end := strings.Index(text[start:], "m")
	if end == -1 {
		return ""
	}

	// Extract the style codes between \x1b[ and m
	codes := text[start+2 : start+end]

	// Split by semicolon to find background codes
	parts := strings.Split(codes, ";")

	for i, part := range parts {
		if part == "48" {
			// Found background color start
			if i+1 < len(parts) && parts[i+1] == "5" && i+2 < len(parts) {
				// 256-color background: 48;5;{color}
				return "48;5;" + parts[i+2]
			} else if i+1 < len(parts) && parts[i+1] == "2" && i+4 < len(parts) {
				// RGB background: 48;2;{r};{g};{b}
				return "48;2;" + parts[i+2] + ";" + parts[i+3] + ";" + parts[i+4]
			}
		} else if strings.HasPrefix(part, "4") && len(part) == 2 {
			// Standard background colors: 40-47
			return part
		} else if strings.HasPrefix(part, "10") && len(part) == 3 {
			// Bright background colors: 100-107
			return part
		}
	}

	return ""
}

// ansiTruncate truncates text accounting for ANSI escape codes (like lipgloss ansi.Truncate)
func ansiTruncate(text string, maxWidth int, suffix string) string {
	if maxWidth <= 0 {
		return ""
	}

	textWidth := lipgloss.Width(text)
	if textWidth <= maxWidth {
		return text
	}

	suffixWidth := lipgloss.Width(suffix)
	if maxWidth <= suffixWidth {
		// If there's no room for content, just return dots
		return strings.Repeat(".", maxWidth)
	}

	// We need to truncate - account for ANSI codes
	targetWidth := maxWidth - suffixWidth
	var result strings.Builder
	var currentWidth int
	inAnsi := false

	for _, r := range text {
		if r == '\x1b' {
			inAnsi = true
			result.WriteRune(r)
			continue
		}

		if inAnsi {
			result.WriteRune(r)
			if r == 'm' {
				inAnsi = false
			}
			continue
		}

		charWidth := lipgloss.Width(string(r))
		if currentWidth+charWidth > targetWidth {
			break
		}

		result.WriteRune(r)
		currentWidth += charWidth
	}

	result.WriteString(suffix)
	return result.String()
}

// ansiTruncateWithRunewidth truncates text accounting for ANSI escape codes (like lipgloss ansi.Truncate)
func ansiTruncateWithRunewidth(text string, maxWidth int, suffix string) string {
	if maxWidth <= 0 {
		return ""
	}

	textWidth := runewidth.StringWidth(text)
	if textWidth <= maxWidth {
		return text
	}

	suffixWidth := runewidth.StringWidth(suffix)
	if maxWidth <= suffixWidth {
		// If there's no room for content, just return dots
		return strings.Repeat(".", maxWidth)
	}

	// We need to truncate - account for ANSI codes
	targetWidth := maxWidth - suffixWidth
	var result strings.Builder
	var currentWidth int
	inAnsi := false

	for _, r := range text {
		if r == '\x1b' {
			inAnsi = true
			result.WriteRune(r)
			continue
		}

		if inAnsi {
			result.WriteRune(r)
			if r == 'm' {
				inAnsi = false
			}
			continue
		}

		charWidth := runewidth.StringWidth(string(r))
		if currentWidth+charWidth > targetWidth {
			break
		}

		result.WriteRune(r)
		currentWidth += charWidth
	}

	result.WriteString(suffix)
	return result.String()
}

// getBorderChar returns the appropriate border character
func (t *Table) getBorderChar() string {
	if t.config.ShowBorders {
		return t.config.Theme.BorderChars.Vertical
	}
	return " "
}

// applyAutomaticTruncation applies automatic truncation with ellipsis using proper Unicode width calculation
func (t *Table) applyAutomaticTruncation(text string, maxWidth int) string {
	// Use the shared rendering utility for proper Unicode width handling
	return render.TruncateText(text, maxWidth)
}

// SetColumns sets the table columns
func (t *Table) SetColumns(columns []core.TableColumn) tea.Cmd {
	return core.ColumnSetCmd(columns)
}

// SetHeaderVisibility sets header visibility
func (t *Table) SetHeaderVisibility(visible bool) tea.Cmd {
	return core.HeaderVisibilityCmd(visible)
}

// SetBorderVisibility sets border visibility
func (t *Table) SetBorderVisibility(visible bool) tea.Cmd {
	return core.BorderVisibilityCmd(visible)
}

// SetTopBorderVisibility sets top border visibility
func (t *Table) SetTopBorderVisibility(visible bool) tea.Cmd {
	return core.TopBorderVisibilityCmd(visible)
}

// SetBottomBorderVisibility sets bottom border visibility
func (t *Table) SetBottomBorderVisibility(visible bool) tea.Cmd {
	return core.BottomBorderVisibilityCmd(visible)
}

// SetHeaderSeparatorVisibility sets header separator visibility
func (t *Table) SetHeaderSeparatorVisibility(visible bool) tea.Cmd {
	return core.HeaderSeparatorVisibilityCmd(visible)
}

// SetCellFormatter sets a cell formatter for a specific column
func (t *Table) SetCellFormatter(columnIndex int, formatter core.SimpleCellFormatter) tea.Cmd {
	return core.CellFormatterSetCmd(columnIndex, formatter)
}

// SetRowFormatter sets the row formatter
func (t *Table) SetRowFormatter(formatter core.LoadingRowFormatter) tea.Cmd {
	return core.RowFormatterSetCmd(formatter)
}

// SetHeaderFormatter sets the header formatter
func (t *Table) SetHeaderFormatter(columnIndex int, formatter core.SimpleHeaderFormatter) tea.Cmd {
	return core.HeaderFormatterSetCmd(columnIndex, formatter)
}

// SetTheme sets the table theme
func (t *Table) SetTheme(theme core.Theme) tea.Cmd {
	return core.TableThemeSetCmd(theme)
}

// SetColumnFormatter sets a formatter for a specific column with automatic truncation
func (t *Table) SetColumnFormatter(columnIndex int, formatter core.SimpleCellFormatter) tea.Cmd {
	return t.SetCellFormatter(columnIndex, formatter)
}

// SetHeaderFormatterForColumn sets a header formatter for a specific column with automatic truncation
func (t *Table) SetHeaderFormatterForColumn(columnIndex int, formatter core.SimpleHeaderFormatter) tea.Cmd {
	return t.SetHeaderFormatter(columnIndex, formatter)
}

// CreateSimpleCellFormatter creates a SimpleCellFormatter from a basic formatting function
// This is a helper to make it easier to create formatters that just need to transform the cell value
func CreateSimpleCellFormatter(formatFunc func(cellValue string) string) core.SimpleCellFormatter {
	return func(cellValue string, rowIndex int, column core.TableColumn, ctx core.RenderContext, isCursor bool, isSelected bool, isActiveCell bool) string {
		// Apply the basic formatting
		formatted := formatFunc(cellValue)

		// Apply default styling based on row state
		var style lipgloss.Style
		if isCursor && isSelected {
			style = ctx.Theme.CursorStyle.Copy().Background(ctx.Theme.SelectedStyle.GetBackground())
		} else if isCursor {
			style = ctx.Theme.CursorStyle
		} else if isSelected {
			style = ctx.Theme.SelectedStyle
		} else {
			style = ctx.Theme.CellStyle
		}

		return style.Render(formatted)
	}
}

// CreateSimpleHeaderFormatter creates a SimpleHeaderFormatter from a basic formatting function
func CreateSimpleHeaderFormatter(formatFunc func(columnTitle string) string) core.SimpleHeaderFormatter {
	return func(column core.TableColumn, ctx core.RenderContext) string {
		// Apply the basic formatting
		formatted := formatFunc(column.Title)

		// Apply header styling
		return formatted
	}
}

// calculateBoundingArea calculates the bounding area around the current viewport automatically
func (t *Table) calculateBoundingArea() core.BoundingArea {
	return viewport.CalculateBoundingArea(t.viewport, t.config.ViewportConfig, t.totalItems)
}

// unloadChunksOutsideBoundingArea unloads chunks that are outside the bounding area
func (t *Table) unloadChunksOutsideBoundingArea() tea.Cmd {
	boundingArea := t.calculateBoundingArea()
	chunkSize := t.config.ViewportConfig.ChunkSize
	var cmds []tea.Cmd

	// Find and unload chunks outside the bounding area
	chunksToUnload := data.FindChunksToUnload(t.chunks, boundingArea, chunkSize)
	for _, chunkStart := range chunksToUnload {
		delete(t.chunks, chunkStart)
		delete(t.chunkAccessTime, chunkStart)
		cmds = append(cmds, core.ChunkUnloadedCmd(chunkStart))
	}

	return tea.Batch(cmds...)
}

// isChunkLoaded checks if a chunk containing the given index is loaded
func (t *Table) isChunkLoaded(index int) bool {
	return data.IsChunkLoaded(index, t.chunks)
}

// updateViewportPosition updates the viewport based on cursor position
func (t *Table) updateViewportPosition() {
	t.viewport = viewport.UpdateViewportPosition(t.viewport, t.config.ViewportConfig, t.totalItems)
}

// updateViewportBounds updates viewport boundary flags
func (t *Table) updateViewportBounds() {
	t.viewport = viewport.UpdateViewportBounds(t.viewport, t.config.ViewportConfig, t.totalItems)
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
	chunksToLoad := data.CalculateChunksInBoundingArea(boundingArea, chunkSize, t.totalItems)

	// Load chunks that aren't already loaded or loading
	for _, chunkStart := range chunksToLoad {
		if !t.isChunkLoaded(chunkStart) && !t.loadingChunks[chunkStart] {
			// Mark chunk as loading
			t.loadingChunks[chunkStart] = true
			newLoadingChunks = append(newLoadingChunks, chunkStart)

			request := data.CreateChunkRequest(
				chunkStart,
				chunkSize,
				t.totalItems,
				t.sortFields,
				t.sortDirs,
				t.filters,
			)

			// Emit chunk loading started message for observability
			cmds = append(cmds, core.ChunkLoadingStartedCmd(chunkStart, request))
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
	chunksToUnload := data.FindChunksToUnload(t.chunks, boundingArea, chunkSize)
	for _, chunkStart := range chunksToUnload {
		delete(t.chunks, chunkStart)
		delete(t.chunkAccessTime, chunkStart)
		cmds = append(cmds, core.ChunkUnloadedCmd(chunkStart))
	}

	return tea.Batch(cmds...)
}

// isLoadingCriticalChunks checks if we're loading chunks that affect the current viewport
func (t *Table) isLoadingCriticalChunks() bool {
	return data.IsLoadingCriticalChunks(t.viewport, t.config.ViewportConfig, t.loadingChunks)
}

// updateVisibleItems updates the slice of items currently visible in the viewport
func (t *Table) updateVisibleItems() {
	result := viewport.CalculateVisibleItemsFromChunks(
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
	chunkStartIndex := data.CalculateChunkStartIndex(index, t.config.ViewportConfig.ChunkSize)
	if _, exists := t.chunks[chunkStartIndex]; !exists {
		// DON'T load chunks immediately here!
		// This function is called during rendering/visible item calculation
		// Let smartChunkManagement handle proper loading with observability messages
		// The chunk will be loaded on the next smartChunkManagement call
	}
}

// getItemAtIndex retrieves an item at a specific index
func (t *Table) getItemAtIndex(index int) (core.Data[any], bool) {
	return data.GetItemAtIndex(index, t.chunks, t.totalItems, t.chunkAccessTime)
}

// findItemIndex finds the index of an item by ID
func (t *Table) findItemIndex(id string) int {
	return data.FindItemIndex(id, t.chunks)
}

// toggleItemSelection toggles selection for an item via DataSource
func (t *Table) toggleItemSelection(id string) tea.Cmd {
	if t.config.SelectionMode == core.SelectionNone || t.dataSource == nil {
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
	keepLowerBound, keepUpperBound := data.CalculateUnloadBounds(t.viewport, t.config.ViewportConfig)

	var unloadedChunks []int

	// Unload chunks outside the bounds
	for startIndex := range t.chunks {
		if data.ShouldUnloadChunk(startIndex, keepLowerBound, keepUpperBound) {
			delete(t.chunks, startIndex)
			delete(t.chunkAccessTime, startIndex)
			unloadedChunks = append(unloadedChunks, startIndex)
		}
	}

	// Return commands for unloaded chunks (for UI feedback)
	var cmds []tea.Cmd
	for _, chunkStart := range unloadedChunks {
		cmds = append(cmds, core.ChunkUnloadedCmd(chunkStart))
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
		request := data.CreateChunkRequest(
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

// constructBottomBorder constructs the bottom border like lipgloss table
func (t *Table) constructBottomBorder() string {
	var parts []string
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.config.Theme.BorderColor))

	// Left corner
	parts = append(parts, borderStyle.Render(t.config.Theme.BorderChars.BottomLeft))

	// Add indicator column border since component renderer is always enabled
	indicatorWidth := 4
	// Horizontal line for indicator column width
	parts = append(parts, borderStyle.Render(strings.Repeat(t.config.Theme.BorderChars.Horizontal, indicatorWidth)))
	// Column separator
	parts = append(parts, borderStyle.Render(t.config.Theme.BorderChars.BottomT))

	// Column borders
	for i, col := range t.columns {
		// Horizontal line for column width
		parts = append(parts, borderStyle.Render(strings.Repeat(t.config.Theme.BorderChars.Horizontal, col.Width)))

		// Column separator or right corner
		if i < len(t.columns)-1 {
			parts = append(parts, borderStyle.Render(t.config.Theme.BorderChars.BottomT))
		} else {
			parts = append(parts, borderStyle.Render(t.config.Theme.BorderChars.BottomRight))
		}
	}

	return strings.Join(parts, "")
}

// constructTopBorder constructs the top border like the bottom border but with top characters
func (t *Table) constructTopBorder() string {
	var parts []string
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.config.Theme.BorderColor))

	// Left corner
	parts = append(parts, borderStyle.Render(t.config.Theme.BorderChars.TopLeft))

	// Add indicator column border since component renderer is always enabled
	indicatorWidth := 4
	// Horizontal line for indicator column width
	parts = append(parts, borderStyle.Render(strings.Repeat(t.config.Theme.BorderChars.Horizontal, indicatorWidth)))
	// Column separator
	parts = append(parts, borderStyle.Render(t.config.Theme.BorderChars.TopT))

	// Column borders
	for i, col := range t.columns {
		// Horizontal line for column width
		parts = append(parts, borderStyle.Render(strings.Repeat(t.config.Theme.BorderChars.Horizontal, col.Width)))

		// Column separator or right corner
		if i < len(t.columns)-1 {
			parts = append(parts, borderStyle.Render(t.config.Theme.BorderChars.TopT))
		} else {
			parts = append(parts, borderStyle.Render(t.config.Theme.BorderChars.TopRight))
		}
	}

	return strings.Join(parts, "")
}

// constructHeaderSeparator constructs the separator border between header and data
func (t *Table) constructHeaderSeparator() string {
	var parts []string
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.config.Theme.BorderColor))

	// Left T-junction
	parts = append(parts, borderStyle.Render(t.config.Theme.BorderChars.LeftT))

	// Add indicator column border since component renderer is always enabled
	indicatorWidth := 4
	// Horizontal line for indicator column width
	parts = append(parts, borderStyle.Render(strings.Repeat(t.config.Theme.BorderChars.Horizontal, indicatorWidth)))
	// Column separator (cross)
	parts = append(parts, borderStyle.Render(t.config.Theme.BorderChars.Cross))

	// Column borders
	for i, col := range t.columns {
		// Horizontal line for column width
		parts = append(parts, borderStyle.Render(strings.Repeat(t.config.Theme.BorderChars.Horizontal, col.Width)))

		// Column separator or right T-junction
		if i < len(t.columns)-1 {
			parts = append(parts, borderStyle.Render(t.config.Theme.BorderChars.Cross))
		} else {
			parts = append(parts, borderStyle.Render(t.config.Theme.BorderChars.RightT))
		}
	}

	return strings.Join(parts, "")
}

// EnableComponentRenderer is deprecated - component rendering is now always enabled
// func (t *Table) EnableComponentRenderer() tea.Cmd {
//	config := DefaultComponentTableRenderConfig()
//	t.componentRenderer = NewTableComponentRenderer(config)
//	return nil
// }

// EnableComponentRendererWithConfig is deprecated - use UpdateComponentConfig instead
// func (t *Table) EnableComponentRendererWithConfig(config ComponentTableRenderConfig) tea.Cmd {
//	t.componentRenderer = NewTableComponentRenderer(config)
//	return nil
// }

// DisableComponentRenderer is deprecated - component rendering is now always enabled
// func (t *Table) DisableComponentRenderer() tea.Cmd {
//	t.componentRenderer = nil
//	return nil
// }

// UpdateComponentConfig updates the component renderer configuration
func (t *Table) UpdateComponentConfig(config ComponentTableRenderConfig) tea.Cmd {
	t.componentRenderer.UpdateConfig(config)
	return nil
}

// GetComponentRenderer returns the component renderer
func (t *Table) GetComponentRenderer() *TableComponentRenderer {
	return t.componentRenderer
}

// IsComponentRenderingEnabled returns whether component rendering is enabled (always true now)
func (t *Table) IsComponentRenderingEnabled() bool {
	return true
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(s string) string {
	result := ""
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result += string(r)
	}
	return result
}

// SetTopBorderSpaceRemoval controls whether top border space is completely removed
func (t *Table) SetTopBorderSpaceRemoval(remove bool) tea.Cmd {
	return core.TopBorderSpaceRemovalCmd(remove)
}

// SetBottomBorderSpaceRemoval controls whether bottom border space is completely removed
func (t *Table) SetBottomBorderSpaceRemoval(remove bool) tea.Cmd {
	return core.BottomBorderSpaceRemovalCmd(remove)
}

// applyHorizontalScroll applies horizontal scrolling offset to text content with ANSI awareness
func (t *Table) applyHorizontalScroll(text string, columnIndex int) string {
	// Skip scrolling for special column indices (headers, indicators, loading cells)
	if columnIndex < 0 {
		return text
	}

	// If scope is "current", only apply scrolling to the current row
	if t.scrollAllRows {
		// We need to know if this is the current row - this will be handled in renderRow
		// For now, let the calling code handle this logic
	}

	// Get scroll offset for this column
	scrollOffset := t.horizontalScrollOffsets[columnIndex]
	if scrollOffset <= 0 {
		return text // No scrolling needed
	}

	// Clean text for processing (same as in applyCellConstraints)
	cleanText := strings.ReplaceAll(text, "\n", " ")
	cleanText = strings.ReplaceAll(cleanText, "\r", " ")
	cleanText = strings.ReplaceAll(cleanText, "\t", " ")
	for strings.Contains(cleanText, "  ") {
		cleanText = strings.ReplaceAll(cleanText, "  ", " ")
	}
	cleanText = strings.TrimSpace(cleanText)

	// Use simpler, more robust ANSI-aware scrolling
	return t.applySimpleANSIAwareScroll(cleanText, scrollOffset)
}

// applySimpleANSIAwareScroll applies horizontal scrolling with simple approach
func (t *Table) applySimpleANSIAwareScroll(text string, scrollOffset int) string {
	if scrollOffset <= 0 {
		return text
	}

	// For styled text, strip ANSI codes to prevent scrolling from getting stuck
	plainText := text
	if strings.Contains(text, "\x1b") {
		plainText = stripANSI(text)
	}

	// Apply scrolling to the plain text based on mode
	switch t.horizontalScrollMode {
	case "word":
		words := strings.Fields(plainText)
		if scrollOffset >= len(words) {
			return ""
		}
		return strings.Join(words[scrollOffset:], " ")
	case "smart":
		runes := []rune(plainText)
		boundaries := []int{0}

		for i, r := range runes {
			if i > 0 {
				prev := runes[i-1]
				if strings.ContainsRune(".,;:!?-", prev) && r == ' ' {
					boundaries = append(boundaries, i+1)
				}
				if r >= 'A' && r <= 'Z' && prev >= 'a' && prev <= 'z' {
					boundaries = append(boundaries, i)
				}
				if prev == ' ' && r != ' ' {
					boundaries = append(boundaries, i)
				}
			}
		}

		if scrollOffset >= len(boundaries) {
			return ""
		}
		boundaryPos := boundaries[scrollOffset]
		if boundaryPos >= len(runes) {
			return ""
		}
		return string(runes[boundaryPos:])
	default: // "character"
		runes := []rune(plainText)
		if scrollOffset >= len(runes) {
			return ""
		}
		return string(runes[scrollOffset:])
	}
}

// applyCharacterScrollWithANSI scrolls text character by character while preserving ANSI styling
func (t *Table) applyCharacterScrollWithANSI(text string, offset int) string {
	if offset <= 0 {
		return text
	}

	// Parse text into segments (visible chars and ANSI codes)
	segments := t.parseTextWithANSI(text)
	if len(segments) == 0 {
		return ""
	}

	// Find the scroll position in terms of visible characters
	visibleCount := 0
	startSegmentIndex := 0
	startCharOffset := 0

	for i, segment := range segments {
		if segment.IsANSI {
			continue // Skip ANSI codes when counting
		}

		segmentLength := runewidth.StringWidth(segment.Text)
		if visibleCount+segmentLength > offset {
			// We found the segment where scrolling starts
			startSegmentIndex = i
			startCharOffset = offset - visibleCount
			break
		}
		visibleCount += segmentLength

		// If we've reached the end without finding enough characters
		if i == len(segments)-1 {
			return "" // Scrolled past end
		}
	}

	// Build the scrolled text starting from the calculated position
	return t.buildScrolledText(segments, startSegmentIndex, startCharOffset)
}

// applyWordScrollWithANSI scrolls text word by word while preserving ANSI styling
func (t *Table) applyWordScrollWithANSI(text string, offset int) string {
	if offset <= 0 {
		return text
	}

	// Parse text into segments
	segments := t.parseTextWithANSI(text)

	// Find word boundaries in visible text only
	wordBoundaries := t.findWordBoundariesInSegments(segments)

	if offset >= len(wordBoundaries) {
		return "" // Scrolled past end
	}

	// Find the segment and position for the target word boundary
	targetPosition := wordBoundaries[offset]
	return t.scrollToVisiblePosition(segments, targetPosition)
}

// applySmartScrollWithANSI scrolls to meaningful boundaries while preserving ANSI styling
func (t *Table) applySmartScrollWithANSI(text string, offset int) string {
	if offset <= 0 {
		return text
	}

	// Parse text into segments
	segments := t.parseTextWithANSI(text)

	// Find smart boundaries in visible text only
	smartBoundaries := t.findSmartBoundariesInSegments(segments)

	if offset >= len(smartBoundaries) {
		return "" // Scrolled past end
	}

	// Find the segment and position for the target boundary
	targetPosition := smartBoundaries[offset]
	return t.scrollToVisiblePosition(segments, targetPosition)
}

// findSmartBoundariesInSegments finds meaningful scroll positions in text
func (t *Table) findSmartBoundariesInSegments(segments []TextSegment) []int {
	var boundaries []int
	boundaries = append(boundaries, 0) // Always start at beginning

	for _, segment := range segments {
		// Add boundary after punctuation
		if strings.ContainsRune(".,;:!?-", segment.Rune) && segment.Next != nil && segment.Next.IsSpace {
			boundaries = append(boundaries, segment.Next.Index)
		}
		// Add boundary at word starts after spaces
		if segment.IsSpace && segment.Next != nil && !segment.Next.IsSpace {
			boundaries = append(boundaries, segment.Next.Index)
		}
		// Add boundary at uppercase letters (camelCase support)
		if segment.Rune >= 'A' && segment.Rune <= 'Z' && segment.Prev != nil && segment.Prev.IsLower {
			boundaries = append(boundaries, segment.Index)
		}
	}

	return boundaries
}

// scrollToVisiblePosition scrolls to a specific position in the text
func (t *Table) scrollToVisiblePosition(segments []TextSegment, targetPosition int) string {
	var result strings.Builder
	currentPosition := 0

	for _, segment := range segments {
		if segment.IsANSI {
			result.WriteString(segment.Text)
			continue
		}

		segmentLength := runewidth.StringWidth(segment.Text)
		if currentPosition+segmentLength > targetPosition {
			// We found the segment where scrolling starts
			result.WriteString(segment.Text[targetPosition-currentPosition:])
			break
		}
		currentPosition += segmentLength
	}

	return result.String()
}

// parseTextWithANSI parses text into segments (visible chars and ANSI codes)
func (t *Table) parseTextWithANSI(text string) []TextSegment {
	var segments []TextSegment
	runes := []rune(text)

	i := 0
	for i < len(runes) {
		r := runes[i]

		if r == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
			// Found ANSI escape sequence
			start := i
			i += 2 // Skip \x1b[

			// Find the end of the ANSI sequence
			for i < len(runes) && runes[i] != 'm' {
				i++
			}
			if i < len(runes) {
				i++ // Include the 'm'
			}

			// Add ANSI segment
			segments = append(segments, TextSegment{
				Text:   string(runes[start:i]),
				IsANSI: true,
				Index:  start,
			})
		} else {
			// Regular character
			segments = append(segments, TextSegment{
				Text:    string(r),
				IsANSI:  false,
				IsSpace: r == ' ' || r == '\t',
				Index:   i,
				Rune:    r,
				IsLower: r >= 'a' && r <= 'z',
			})
			i++
		}
	}

	// Link segments with Next/Prev pointers
	for i := range segments {
		if i > 0 {
			segments[i].Prev = &segments[i-1]
		}
		if i < len(segments)-1 {
			segments[i].Next = &segments[i+1]
		}
	}

	return segments
}

// buildScrolledText builds the scrolled text from segments
func (t *Table) buildScrolledText(segments []TextSegment, startSegmentIndex, startCharOffset int) string {
	var result strings.Builder
	activeStyles := make(map[string]string) // Track active ANSI styles

	// First pass: collect all ANSI styles that should be active at the start position
	visibleChars := 0
	for i, segment := range segments {
		if segment.IsANSI {
			// Parse and store the style
			if styleType, styleValue := t.parseANSIStyle(segment.Text); styleType != "" {
				activeStyles[styleType] = styleValue
			}
			continue
		}

		segmentLength := runewidth.StringWidth(segment.Text)
		if i >= startSegmentIndex {
			break
		}
		visibleChars += segmentLength
	}

	// Apply accumulated styles at the start
	for _, styleValue := range activeStyles {
		result.WriteString(styleValue)
	}

	// Second pass: build the visible text starting from the scroll position
	for i, segment := range segments {
		if i < startSegmentIndex {
			continue
		}

		if segment.IsANSI {
			result.WriteString(segment.Text)
			continue
		}

		if i == startSegmentIndex && startCharOffset > 0 {
			// Handle partial segment at start position
			runes := []rune(segment.Text)
			if startCharOffset < len(runes) {
				result.WriteString(string(runes[startCharOffset:]))
			}
		} else {
			// Include full segment
			result.WriteString(segment.Text)
		}
	}

	return result.String()
}

// parseANSIStyle parses an ANSI escape sequence and returns style type and value
func (t *Table) parseANSIStyle(ansiCode string) (string, string) {
	// Simple ANSI parsing - can be extended for more complex cases
	if strings.Contains(ansiCode, "38;5;") {
		return "foreground", ansiCode
	}
	if strings.Contains(ansiCode, "48;5;") {
		return "background", ansiCode
	}
	if strings.Contains(ansiCode, "1m") {
		return "bold", ansiCode
	}
	if strings.Contains(ansiCode, "3m") {
		return "italic", ansiCode
	}
	return "", ""
}

// findWordBoundariesInSegments finds word boundaries in visible text only
func (t *Table) findWordBoundariesInSegments(segments []TextSegment) []int {
	var boundaries []int
	boundaries = append(boundaries, 0) // Always start at beginning

	visiblePos := 0
	inWord := false

	for _, segment := range segments {
		if segment.IsANSI {
			continue // Skip ANSI codes
		}

		for _, r := range segment.Text {
			isSpace := r == ' ' || r == '\t'

			if inWord && isSpace {
				// End of word - add boundary at next non-space
				inWord = false
			} else if !inWord && !isSpace {
				// Start of word
				boundaries = append(boundaries, visiblePos)
				inWord = true
			}

			if !isSpace {
				visiblePos++
			}
		}
	}

	return boundaries
}

// handleHorizontalScrollLeft scrolls left in current scroll mode
func (t *Table) handleHorizontalScrollLeft() tea.Cmd {
	if t.horizontalScrollOffsets[t.currentColumn] > 0 {
		t.horizontalScrollOffsets[t.currentColumn]--
	}
	return nil
}

// handleHorizontalScrollRight scrolls right in current scroll mode
func (t *Table) handleHorizontalScrollRight() tea.Cmd {
	// Get max scroll for current column
	maxScroll := t.getMaxScrollForColumn(t.currentColumn)
	if t.horizontalScrollOffsets[t.currentColumn] < maxScroll {
		t.horizontalScrollOffsets[t.currentColumn]++
	}
	return nil
}

// handleHorizontalScrollPageLeft scrolls left by a larger amount (page-based)
func (t *Table) handleHorizontalScrollPageLeft() tea.Cmd {
	pageSize := 5 // Scroll by 5 units at a time for page-based navigation
	currentOffset := t.horizontalScrollOffsets[t.currentColumn]
	if currentOffset >= pageSize {
		t.horizontalScrollOffsets[t.currentColumn] = currentOffset - pageSize
	} else {
		t.horizontalScrollOffsets[t.currentColumn] = 0
	}
	return nil
}

// handleHorizontalScrollPageRight scrolls right by a larger amount (page-based)
func (t *Table) handleHorizontalScrollPageRight() tea.Cmd {
	pageSize := 5 // Scroll by 5 units at a time for page-based navigation
	maxScroll := t.getMaxScrollForColumn(t.currentColumn)
	currentOffset := t.horizontalScrollOffsets[t.currentColumn]
	if currentOffset+pageSize <= maxScroll {
		t.horizontalScrollOffsets[t.currentColumn] = currentOffset + pageSize
	} else {
		t.horizontalScrollOffsets[t.currentColumn] = maxScroll
	}
	return nil
}

// handleHorizontalScrollWordLeft scrolls left by word
func (t *Table) handleHorizontalScrollWordLeft() tea.Cmd {
	// Get current content for the focused column to find word boundaries
	if t.currentColumn < 0 || t.currentColumn >= len(t.columns) {
		return nil
	}

	currentOffset := t.horizontalScrollOffsets[t.currentColumn]
	if currentOffset <= 0 {
		return nil // Already at start
	}

	// Find content from current row or any visible row if no content in current row
	var cellText string
	for _, item := range t.visibleItems {
		if row, ok := item.Item.(core.TableRow); ok && t.currentColumn < len(row.Cells) {
			cellText = row.Cells[t.currentColumn]
			if cellText != "" {
				break // Use first non-empty cell content
			}
		}
	}

	if cellText == "" {
		return nil // No content to scroll
	}

	// Clean text for processing
	cleanText := strings.ReplaceAll(cellText, "\n", " ")
	cleanText = strings.ReplaceAll(cleanText, "\r", " ")
	cleanText = strings.ReplaceAll(cleanText, "\t", " ")
	for strings.Contains(cleanText, "  ") {
		cleanText = strings.ReplaceAll(cleanText, "  ", " ")
	}
	cleanText = strings.TrimSpace(cleanText)

	// Find the previous word boundary before current position
	runes := []rune(cleanText)
	if currentOffset > len(runes) {
		currentOffset = len(runes)
	}

	// Find previous word boundary
	prevWordStart := 0 // Default to beginning
	inWord := false

	for i := 0; i < currentOffset; i++ {
		isSpace := runes[i] == ' ' || runes[i] == '\t'

		if !inWord && !isSpace {
			// Found start of word
			prevWordStart = i
			inWord = true
		} else if inWord && isSpace {
			// End of word
			inWord = false
		}
	}

	t.horizontalScrollOffsets[t.currentColumn] = prevWordStart
	return nil
}

// handleHorizontalScrollWordRight scrolls right by word
func (t *Table) handleHorizontalScrollWordRight() tea.Cmd {
	// Get current content for the focused column to find word boundaries
	if t.currentColumn < 0 || t.currentColumn >= len(t.columns) {
		return nil
	}

	// Find content from current row or any visible row if no content in current row
	var cellText string
	for _, item := range t.visibleItems {
		if row, ok := item.Item.(core.TableRow); ok && t.currentColumn < len(row.Cells) {
			cellText = row.Cells[t.currentColumn]
			if cellText != "" {
				break // Use first non-empty cell content
			}
		}
	}

	if cellText == "" {
		return nil // No content to scroll
	}

	// Clean text for processing
	cleanText := strings.ReplaceAll(cellText, "\n", " ")
	cleanText = strings.ReplaceAll(cleanText, "\r", " ")
	cleanText = strings.ReplaceAll(cleanText, "\t", " ")
	for strings.Contains(cleanText, "  ") {
		cleanText = strings.ReplaceAll(cleanText, "  ", " ")
	}
	cleanText = strings.TrimSpace(cleanText)

	// Get current character offset
	currentOffset := t.horizontalScrollOffsets[t.currentColumn]

	// Find the next word boundary after current position
	runes := []rune(cleanText)
	if currentOffset >= len(runes) {
		return nil // Already at or past end
	}

	// Find next word boundary
	nextWordStart := -1
	inWord := false

	for i := currentOffset; i < len(runes); i++ {
		isSpace := runes[i] == ' ' || runes[i] == '\t'

		if inWord && isSpace {
			// End of current word, look for next word start
			inWord = false
		} else if !inWord && !isSpace {
			// Found start of next word
			if i > currentOffset { // Must be after current position
				nextWordStart = i
				break
			}
			inWord = true
		}
	}

	if nextWordStart > currentOffset {
		t.horizontalScrollOffsets[t.currentColumn] = nextWordStart
	}

	return nil
}

// handleHorizontalScrollSmartLeft scrolls left by smart boundaries
func (t *Table) handleHorizontalScrollSmartLeft() tea.Cmd {
	// Get current content for the focused column to find smart boundaries
	if t.currentColumn < 0 || t.currentColumn >= len(t.columns) {
		return nil
	}

	currentOffset := t.horizontalScrollOffsets[t.currentColumn]
	if currentOffset <= 0 {
		return nil // Already at start
	}

	// Find content from current row or any visible row if no content in current row
	var cellText string
	for _, item := range t.visibleItems {
		if row, ok := item.Item.(core.TableRow); ok && t.currentColumn < len(row.Cells) {
			cellText = row.Cells[t.currentColumn]
			if cellText != "" {
				break // Use first non-empty cell content
			}
		}
	}

	if cellText == "" {
		return nil // No content to scroll
	}

	// Clean text for processing
	cleanText := strings.ReplaceAll(cellText, "\n", " ")
	cleanText = strings.ReplaceAll(cleanText, "\r", " ")
	cleanText = strings.ReplaceAll(cleanText, "\t", " ")
	for strings.Contains(cleanText, "  ") {
		cleanText = strings.ReplaceAll(cleanText, "  ", " ")
	}
	cleanText = strings.TrimSpace(cleanText)

	// Find smart boundaries
	segments := t.parseTextWithANSI(cleanText)
	boundaries := t.findSmartBoundariesInSegments(segments)

	// Find the previous boundary before current position
	prevBoundary := 0 // Default to beginning
	for _, boundary := range boundaries {
		if boundary < currentOffset {
			prevBoundary = boundary
		} else {
			break // We've passed the current position
		}
	}

	t.horizontalScrollOffsets[t.currentColumn] = prevBoundary
	return nil
}

// handleHorizontalScrollSmartRight scrolls right by smart boundaries
func (t *Table) handleHorizontalScrollSmartRight() tea.Cmd {
	// Get current content for the focused column to find smart boundaries
	if t.currentColumn < 0 || t.currentColumn >= len(t.columns) {
		return nil
	}

	// Find content from current row or any visible row if no content in current row
	var cellText string
	for _, item := range t.visibleItems {
		if row, ok := item.Item.(core.TableRow); ok && t.currentColumn < len(row.Cells) {
			cellText = row.Cells[t.currentColumn]
			if cellText != "" {
				break // Use first non-empty cell content
			}
		}
	}

	if cellText == "" {
		return nil // No content to scroll
	}

	// Clean text for processing
	cleanText := strings.ReplaceAll(cellText, "\n", " ")
	cleanText = strings.ReplaceAll(cleanText, "\r", " ")
	cleanText = strings.ReplaceAll(cleanText, "\t", " ")
	for strings.Contains(cleanText, "  ") {
		cleanText = strings.ReplaceAll(cleanText, "  ", " ")
	}
	cleanText = strings.TrimSpace(cleanText)

	// Get current character offset
	currentOffset := t.horizontalScrollOffsets[t.currentColumn]

	// Find smart boundaries
	segments := t.parseTextWithANSI(cleanText)
	boundaries := t.findSmartBoundariesInSegments(segments)

	// Find the next boundary after current position
	nextBoundary := -1
	for _, boundary := range boundaries {
		if boundary > currentOffset {
			nextBoundary = boundary
			break
		}
	}

	if nextBoundary > currentOffset {
		t.horizontalScrollOffsets[t.currentColumn] = nextBoundary
	}

	return nil
}

// handleNextColumn switches to next column for scrolling
func (t *Table) handleNextColumn() tea.Cmd {
	t.currentColumn = (t.currentColumn + 1) % len(t.columns)
	return nil
}

// handlePrevColumn switches to previous column for scrolling
func (t *Table) handlePrevColumn() tea.Cmd {
	t.currentColumn = (t.currentColumn - 1 + len(t.columns)) % len(t.columns)
	return nil
}

// handleToggleScrollMode cycles through scroll modes
func (t *Table) handleToggleScrollMode() tea.Cmd {
	switch t.horizontalScrollMode {
	case "character":
		t.horizontalScrollMode = "word"
	case "word":
		t.horizontalScrollMode = "smart"
	case "smart":
		t.horizontalScrollMode = "character"
	}
	return nil
}

// handleResetScrolling resets all scroll offsets
func (t *Table) handleResetScrolling() tea.Cmd {
	t.horizontalScrollOffsets = make(map[int]int)
	return nil
}

// getMaxScrollForColumn calculates the maximum scroll offset for a column
func (t *Table) getMaxScrollForColumn(columnIndex int) int {
	if columnIndex < 0 || columnIndex >= len(t.columns) {
		return 0
	}

	column := t.columns[columnIndex]
	columnWidth := column.Width

	// Check if any content in this column actually needs scrolling
	hasScrollableContent := false
	maxScroll := 0

	for _, item := range t.visibleItems {
		if row, ok := item.Item.(core.TableRow); ok && columnIndex < len(row.Cells) {
			cellText := row.Cells[columnIndex]

			// Apply the same text cleaning as in applyCellConstraintsWithRowInfo
			cleanText := strings.ReplaceAll(cellText, "\n", " ")
			cleanText = strings.ReplaceAll(cleanText, "\r", " ")
			cleanText = strings.ReplaceAll(cleanText, "\t", " ")
			for strings.Contains(cleanText, "  ") {
				cleanText = strings.ReplaceAll(cleanText, "  ", " ")
			}
			cleanText = strings.TrimSpace(cleanText)

			// Check if content actually exceeds column width (is truncated)
			var measureWidth func(string) int
			if strings.Contains(cleanText, "\x1b") {
				measureWidth = lipgloss.Width // ANSI-aware measurement
			} else {
				measureWidth = runewidth.StringWidth // Unicode-aware measurement
			}

			contentWidth := measureWidth(cleanText)
			if contentWidth <= columnWidth {
				// Content fits completely - no scrolling needed for this cell
				continue
			}

			// Content is truncated - calculate scroll potential
			hasScrollableContent = true

			switch t.horizontalScrollMode {
			case "word":
				words := strings.Fields(cleanText)
				if len(words) > 1 { // Only allow scrolling if there are multiple words
					wordScroll := len(words) - 1
					if wordScroll > maxScroll {
						maxScroll = wordScroll
					}
				}
			case "smart":
				boundaries := t.findSmartBoundariesInSegments(t.parseTextWithANSI(cleanText))
				if len(boundaries) > 1 { // Only allow scrolling if there are multiple boundaries
					smartScroll := len(boundaries) - 1
					if smartScroll > maxScroll {
						maxScroll = smartScroll
					}
				}
			default: // "character"
				runes := []rune(cleanText)
				// Calculate how many characters we can scroll while still showing useful content
				// We want to ensure the user can scroll to see all the content that was truncated
				charactersHidden := len(runes) - columnWidth
				if charactersHidden > 0 {
					// Allow scrolling to reveal all hidden content, plus a small buffer
					charScroll := charactersHidden + 3 // Small buffer to see the end clearly
					if charScroll > maxScroll {
						maxScroll = charScroll
					}
				}
			}
		}
	}

	// Only allow scrolling if we found content that actually needs it
	if !hasScrollableContent {
		return 0
	}

	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// handleToggleScrollScope toggles between all rows and current row scroll scope
func (t *Table) handleToggleScrollScope() tea.Cmd {
	t.scrollAllRows = !t.scrollAllRows
	return nil
}

// applyHorizontalScrollWithScope applies horizontal scrolling with scope awareness
func (t *Table) applyHorizontalScrollWithScope(text string, columnIndex int, isCurrentRow bool) string {
	// Skip scrolling for special column indices (headers, indicators, loading cells)
	if columnIndex < 0 {
		return text
	}

	// If scope is "current", only apply scrolling to the current row
	if !t.scrollAllRows && !isCurrentRow {
		return text // Don't scroll non-current rows
	}

	// Get scroll offset for this column
	scrollOffset := t.horizontalScrollOffsets[columnIndex]
	if scrollOffset <= 0 {
		return text // No scrolling needed
	}

	// Clean text for processing (same as in applyCellConstraints)
	cleanText := strings.ReplaceAll(text, "\n", " ")
	cleanText = strings.ReplaceAll(cleanText, "\r", " ")
	cleanText = strings.ReplaceAll(cleanText, "\t", " ")
	for strings.Contains(cleanText, "  ") {
		cleanText = strings.ReplaceAll(cleanText, "  ", " ")
	}
	cleanText = strings.TrimSpace(cleanText)

	// Use simpler, more robust ANSI-aware scrolling
	return t.applySimpleANSIAwareScroll(cleanText, scrollOffset)
}

// TestHorizontalScrollRight is a temporary public method for testing horizontal scrolling
func (t *Table) TestHorizontalScrollRight() tea.Cmd {
	return t.handleHorizontalScrollRight()
}

// TestResetHorizontalScroll is a temporary public method for testing
func (t *Table) TestResetHorizontalScroll() {
	t.horizontalScrollOffsets = make(map[int]int)
}

// TestGetScrollState is a temporary public method for debugging
func (t *Table) TestGetScrollState() (map[int]int, string, bool, int) {
	return t.horizontalScrollOffsets, t.horizontalScrollMode, t.scrollAllRows, t.currentColumn
}

// TestSetScrollMode is a temporary public method for testing
func (t *Table) TestSetScrollMode(mode string) {
	t.horizontalScrollMode = mode
}

// GetHorizontalScrollState returns the current horizontal scrolling state
func (t *Table) GetHorizontalScrollState() (mode string, scrollAllRows bool, currentColumn int, offsets map[int]int) {
	// Return copies to prevent external modification
	offsetsCopy := make(map[int]int)
	for k, v := range t.horizontalScrollOffsets {
		offsetsCopy[k] = v
	}

	return t.horizontalScrollMode, t.scrollAllRows, t.currentColumn, offsetsCopy
}

// SetResetScrollOnNavigation controls whether horizontal scroll offsets reset when navigating between rows
func (t *Table) SetResetScrollOnNavigation(enabled bool) {
	t.config.ResetScrollOnNavigation = enabled
}

// isActiveCell determines if a cell at the given position is the active cell for horizontal scrolling
func (t *Table) isActiveCell(columnIndex int, isCurrentRow bool) bool {
	// Skip for special columns (headers, indicators, loading cells)
	if columnIndex < 0 || columnIndex >= len(t.columns) {
		return false
	}

	// Check if this is the focused column for horizontal scrolling
	if columnIndex != t.currentColumn {
		return false
	}

	// Apply scope-aware logic
	switch t.scrollAllRows {
	case true:
		// All rows mode: all cells in the focused column can be considered active
		// But we might want to limit this to just show the focused column without row restriction
		return true
	case false:
		// Current row mode: only the cell at cursor row + focused column is active
		return isCurrentRow
	default:
		return false
	}
}

// applyActiveCellIndication applies the configured active cell indication to content
func (t *Table) applyActiveCellIndication(content string, isActiveCell bool) string {
	if !isActiveCell || !t.config.ActiveCellIndicationEnabled {
		return content
	}

	// When enabled, apply background color that should override cursor background
	// This is now handled within the main render loops to properly layer over
	// full-row highlighting. This function remains for other indication modes if added later.
	// For now, we return content as-is because background is handled elsewhere.
	return content
}
