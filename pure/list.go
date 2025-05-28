package vtable

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ================================
// LIST MODEL IMPLEMENTATION
// ================================

// List represents a pure Tea list component
type List struct {
	// Core state
	dataSource DataSource[any]
	chunks     map[int]Chunk[any] // Map of start index to chunk
	totalItems int

	// Viewport state
	viewport ViewportState

	// Configuration
	config ListConfig

	// Rendering
	formatter         ItemFormatter[any]
	animatedFormatter ItemFormatterAnimated[any]
	renderContext     RenderContext

	// Selection state
	selectedItems map[string]bool
	selectedOrder []string // Maintain selection order

	// Focus state
	focused bool

	// Animation state
	animationEngine AnimationEngine
	animations      map[string]ListAnimation
	lastError       error

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

	// Loading state tracking - CRITICAL for UX!
	loadingChunks    map[int]bool // Track chunks currently being loaded
	hasLoadingChunks bool         // Quick check if any chunks are loading
	canScroll        bool         // Whether scrolling is allowed (blocked during critical loads)
}

// ================================
// CONSTRUCTOR
// ================================

// NewList creates a new List with the given configuration and data source
func NewList(config ListConfig, dataSource DataSource[any], formatter ...ItemFormatter[any]) *List {
	// Validate and fix config
	errors := ValidateListConfig(&config)
	if len(errors) > 0 {
		FixListConfig(&config)
	}

	list := &List{
		dataSource:       dataSource,
		chunks:           make(map[int]Chunk[any]),
		config:           config,
		selectedItems:    make(map[string]bool),
		selectedOrder:    make([]string, 0),
		animations:       make(map[string]ListAnimation),
		filters:          make(map[string]any),
		chunkAccessTime:  make(map[int]time.Time),
		visibleItems:     make([]Data[any], 0), // Initialize visible items
		loadingChunks:    make(map[int]bool),   // Initialize loading state tracking
		hasLoadingChunks: false,
		canScroll:        true, // Allow scrolling initially
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

	// Set formatter if provided
	if len(formatter) > 0 {
		list.formatter = formatter[0]
	}

	// Set up render context
	list.setupRenderContext()

	return list
}

// ================================
// TEA MODEL INTERFACE
// ================================

// Init initializes the list model
func (l *List) Init() tea.Cmd {
	return l.loadInitialData()
}

// Update handles all messages and updates the list state
func (l *List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// ===== Lifecycle Messages =====
	case InitMsg:
		return l, l.Init()

	case DestroyMsg:
		if l.animationEngine != nil {
			l.animationEngine.Cleanup()
		}
		return l, nil

	case ResetMsg:
		l.reset()
		return l, l.Init()

	// ===== Navigation Messages =====
	case CursorUpMsg:
		cmd := l.handleCursorUp()
		return l, cmd

	case CursorDownMsg:
		cmd := l.handleCursorDown()
		return l, cmd

	case PageUpMsg:
		cmd := l.handlePageUp()
		return l, cmd

	case PageDownMsg:
		cmd := l.handlePageDown()
		return l, cmd

	case JumpToStartMsg:
		cmd := l.handleJumpToStart()
		return l, cmd

	case JumpToEndMsg:
		cmd := l.handleJumpToEnd()
		return l, cmd

	case JumpToMsg:
		cmd := l.handleJumpTo(msg.Index)
		return l, cmd

	// ===== Data Messages =====
	case DataRefreshMsg:
		cmd := l.handleDataRefresh()
		return l, cmd

	case DataChunksRefreshMsg:
		// Refresh chunks while preserving cursor position
		l.chunks = make(map[int]Chunk[any])
		l.loadingChunks = make(map[int]bool)
		l.hasLoadingChunks = false
		l.canScroll = true
		// Don't reset cursor position - just reload chunks
		return l, l.smartChunkManagement()

	case DataChunkLoadedMsg:
		cmd := l.handleDataChunkLoaded(msg)
		return l, cmd

	case DataChunkErrorMsg:
		l.lastError = msg.Error
		return l, ErrorCmd(msg.Error, "chunk_load")

	case DataTotalMsg:
		l.totalItems = msg.Total
		l.updateViewportBounds()
		// Ensure viewport starts at the configured initial position
		l.viewport.ViewportStartIndex = 0
		l.viewport.CursorIndex = l.config.ViewportConfig.InitialIndex
		l.viewport.CursorViewportIndex = l.config.ViewportConfig.InitialIndex
		// After getting total, load the initial chunks using smart chunk management
		return l, l.smartChunkManagement()

	case DataTotalUpdateMsg:
		// Update total while preserving cursor position
		oldTotal := l.totalItems
		l.totalItems = msg.Total
		l.updateViewportBounds()

		// Ensure cursor stays within bounds if total decreased
		if l.viewport.CursorIndex >= l.totalItems && l.totalItems > 0 {
			l.viewport.CursorIndex = l.totalItems - 1
			// Recalculate viewport position based on new cursor
			l.viewport.CursorViewportIndex = l.viewport.CursorIndex - l.viewport.ViewportStartIndex
			if l.viewport.CursorViewportIndex < 0 {
				l.viewport.ViewportStartIndex = l.viewport.CursorIndex
				l.viewport.CursorViewportIndex = 0
			}
		}

		// Only reload chunks if we need to refresh data (not just for cursor preservation)
		if oldTotal != l.totalItems {
			return l, l.smartChunkManagement()
		}
		return l, nil

	case DataLoadErrorMsg:
		l.lastError = msg.Error
		return l, ErrorCmd(msg.Error, "data_load")

	case DataTotalRequestMsg:
		// Handle DataTotalRequestMsg by calling the actual dataSource
		if l.dataSource != nil {
			return l, l.dataSource.GetTotal()
		}
		return l, nil

	case DataSourceSetMsg:
		l.dataSource = msg.DataSource
		return l, l.dataSource.GetTotal()

	case ChunkUnloadedMsg:
		// Handle chunk unloaded notification (for UI feedback)
		return l, nil

	// ===== Selection Messages =====
	case SelectCurrentMsg:
		cmd := l.handleSelectCurrent()
		return l, cmd

	case SelectToggleMsg:
		cmd := l.handleSelectToggle(msg.Index)
		return l, cmd

	case SelectAllMsg:
		cmd := l.handleSelectAll()
		return l, cmd

	case SelectClearMsg:
		if l.dataSource == nil {
			return l, nil
		}
		// Return the command to be processed by Tea model loop
		return l, l.dataSource.ClearSelection()

	case SelectRangeMsg:
		cmd := l.handleSelectRange(msg.StartID, msg.EndID)
		return l, cmd

	case SelectionModeSetMsg:
		l.config.SelectionMode = msg.Mode
		if msg.Mode == SelectionNone {
			l.clearSelection()
		}
		return l, nil

	case SelectionResponseMsg:
		// Handle selection response from DataSource
		// The DataSource has updated its internal state, now we need to refresh chunks
		// to get the updated selection state in the Data[T].Selected fields
		cmd := l.refreshChunks()
		return l, cmd

	// ===== Filter Messages =====
	case FilterSetMsg:
		l.filters[msg.Field] = msg.Value
		cmd := l.handleFilterChange()
		return l, cmd

	case FilterClearMsg:
		delete(l.filters, msg.Field)
		cmd := l.handleFilterChange()
		return l, cmd

	case FiltersClearAllMsg:
		l.filters = make(map[string]any)
		cmd := l.handleFilterChange()
		return l, cmd

	// ===== Sort Messages =====
	case SortToggleMsg:
		cmd := l.handleSortToggle(msg.Field)
		return l, cmd

	case SortSetMsg:
		cmd := l.handleSortSet(msg.Field, msg.Direction)
		return l, cmd

	case SortAddMsg:
		cmd := l.handleSortAdd(msg.Field, msg.Direction)
		return l, cmd

	case SortRemoveMsg:
		cmd := l.handleSortRemove(msg.Field)
		return l, cmd

	case SortsClearAllMsg:
		l.sortFields = nil
		l.sortDirs = nil
		cmd := l.handleFilterChange() // Refresh data
		return l, cmd

	// ===== Focus Messages =====
	case FocusMsg:
		l.focused = true
		return l, nil

	case BlurMsg:
		l.focused = false
		return l, nil

	// ===== Animation Messages =====
	case GlobalAnimationTickMsg:
		if l.animationEngine != nil {
			cmd := l.animationEngine.ProcessGlobalTick(msg)
			return l, cmd
		}
		return l, nil

	case AnimationUpdateMsg:
		// Handle animation updates
		return l, nil

	case AnimationConfigMsg:
		l.config.AnimationConfig = msg.Config
		if l.animationEngine != nil {
			cmd := l.animationEngine.UpdateConfig(msg.Config)
			return l, cmd
		}
		return l, nil

	case ItemAnimationStartMsg:
		l.animations[msg.ItemID] = msg.Animation
		return l, nil

	case ItemAnimationStopMsg:
		delete(l.animations, msg.ItemID)
		return l, nil

	// ===== Configuration Messages =====
	case FormatterSetMsg:
		l.formatter = msg.Formatter
		return l, nil

	case AnimatedFormatterSetMsg:
		l.animatedFormatter = msg.Formatter
		return l, nil

	case MaxWidthSetMsg:
		l.config.MaxWidth = msg.Width
		l.setupRenderContext()
		return l, nil

	case StyleConfigSetMsg:
		l.config.StyleConfig = msg.Config
		return l, nil

	case ViewportConfigMsg:
		l.config.ViewportConfig = msg.Config
		l.updateViewportBounds()
		return l, nil

	case KeyMapSetMsg:
		l.config.KeyMap = msg.KeyMap
		return l, nil

	// ===== Search Messages =====
	case SearchSetMsg:
		l.searchQuery = msg.Query
		l.searchField = msg.Field
		cmd := l.handleSearch()
		return l, cmd

	case SearchClearMsg:
		l.searchQuery = ""
		l.searchField = ""
		l.searchResults = nil
		return l, nil

	case SearchResultMsg:
		l.searchResults = msg.Results
		return l, nil

	// ===== Error Messages =====
	case ErrorMsg:
		l.lastError = msg.Error
		return l, nil

	// ===== Viewport Messages =====
	case ViewportResizeMsg:
		l.config.ViewportConfig.Height = msg.Height
		l.updateViewportBounds()
		return l, nil

	// ===== Batch Messages =====
	case BatchMsg:
		for _, subMsg := range msg.Messages {
			var cmd tea.Cmd
			_, cmd = l.Update(subMsg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return l, tea.Batch(cmds...)

	// ===== Keyboard Input =====
	case tea.KeyMsg:
		cmd := l.handleKeyPress(msg)
		return l, cmd
	}

	return l, nil
}

// View renders the list
func (l *List) View() string {
	var builder strings.Builder

	// Special case for empty dataset
	if l.totalItems == 0 {
		return "No data available"
	}

	// Ensure visible items are up to date
	l.updateVisibleItems()

	// If we have no visible items, it means chunks are not loaded yet
	if len(l.visibleItems) == 0 {
		return "Loading initial data..."
	}

	// Render each visible item
	for i, item := range l.visibleItems {
		absoluteIndex := l.viewport.ViewportStartIndex + i

		// Skip if we've rendered all real data
		if absoluteIndex >= l.totalItems {
			break
		}

		isCursor := i == l.viewport.CursorViewportIndex
		isSelected := item.Selected

		var renderedItem string

		// Always try to render actual data if available, even if chunk is "loading"
		if l.formatter != nil {
			// Use custom formatter if available
			ctx := RenderContext{
				MaxWidth:          l.config.MaxWidth,
				MaxHeight:         1,
				ErrorIndicator:    "❌",
				LoadingIndicator:  "⏳",
				DisabledIndicator: "🚫",
				SelectedIndicator: "✅",
			}
			renderedItem = l.formatter(
				item,
				absoluteIndex,
				ctx,
				isCursor,
				l.viewport.IsAtTopThreshold,
				l.viewport.IsAtBottomThreshold,
			)
		} else {
			// Use enhanced rendering system with enumerators
			enhancedFormatter := EnhancedListFormatter(l.config.RenderConfig)
			ctx := l.renderContext
			ctx.MaxWidth = l.config.RenderConfig.ContentConfig.MaxWidth

			renderedItem = enhancedFormatter(
				item,
				absoluteIndex,
				ctx,
				isCursor,
				l.viewport.IsAtTopThreshold,
				l.viewport.IsAtBottomThreshold,
			)
		}

		// Apply item styling
		renderedItem = l.applyItemStyle(renderedItem, isCursor, isSelected, item)

		builder.WriteString(renderedItem)

		// Add a newline unless it's the last actual item
		if i < len(l.visibleItems)-1 && absoluteIndex < l.totalItems-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// ================================
// LIST MODEL INTERFACE
// ================================

// Focus sets the list as focused
func (l *List) Focus() tea.Cmd {
	l.focused = true
	return nil
}

// Blur removes focus from the list
func (l *List) Blur() {
	l.focused = false
}

// IsFocused returns whether the list has focus
func (l *List) IsFocused() bool {
	return l.focused
}

// GetState returns the current viewport state
func (l *List) GetState() ViewportState {
	return l.viewport
}

// GetSelectionCount returns the number of selected items
func (l *List) GetSelectionCount() int {
	return GetSelectionCount(l.chunks)
}

// ================================
// PRIVATE HELPER METHODS
// ================================

// setupRenderContext initializes the render context
func (l *List) setupRenderContext() {
	l.renderContext = RenderContext{
		MaxWidth:       l.config.MaxWidth,
		MaxHeight:      1,   // Single line for list items
		Theme:          nil, // Lists use StyleConfig instead
		BaseStyle:      l.config.StyleConfig.DefaultStyle,
		ColorSupport:   true,
		UnicodeSupport: true,
		CurrentTime:    time.Now(),
		FocusState:     FocusState{HasFocus: l.focused},

		// Default state indicators
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
			// Simple word wrapping
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
			l.lastError = err
		},
	}
}

// reset resets the list to its initial state
func (l *List) reset() {
	l.chunks = make(map[int]Chunk[any])
	l.totalItems = 0
	// Selection state is managed by DataSource, not the List
	l.loadingChunks = make(map[int]bool)
	l.hasLoadingChunks = false
	l.canScroll = true
	l.viewport = ViewportState{
		ViewportStartIndex:  0,
		CursorIndex:         l.config.ViewportConfig.InitialIndex,
		CursorViewportIndex: 0,
		IsAtTopThreshold:    false,
		IsAtBottomThreshold: false,
		AtDatasetStart:      true,
		AtDatasetEnd:        false,
	}
	l.lastError = nil
	l.filters = make(map[string]any)
	l.sortFields = nil
	l.sortDirs = nil
	l.searchQuery = ""
	l.searchField = ""
	l.searchResults = nil
}

// ================================
// NAVIGATION HELPERS
// ================================

// loadInitialData loads the total count and initial chunk
func (l *List) loadInitialData() tea.Cmd {
	if l.dataSource == nil {
		return nil
	}

	// First get the total count
	return l.dataSource.GetTotal()
}

// loadInitialChunk loads the first chunk of data
func (l *List) loadInitialChunk() tea.Cmd {
	if l.dataSource == nil {
		return nil
	}

	request := DataRequest{
		Start:          0,
		Count:          l.config.ViewportConfig.ChunkSize,
		SortFields:     l.sortFields,
		SortDirections: l.sortDirs,
		Filters:        l.filters,
	}

	return l.dataSource.LoadChunk(request)
}

// handleCursorUp moves cursor up one position with proper threshold handling
func (l *List) handleCursorUp() tea.Cmd {
	if l.totalItems == 0 || !l.canScroll {
		return nil
	}

	// Can't move up if already at the beginning
	if l.viewport.CursorIndex <= 0 {
		return nil
	}

	previousState := l.viewport
	l.viewport = CalculateCursorUp(l.viewport, l.config.ViewportConfig, l.totalItems)

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
		return l.smartChunkManagement()
	}

	return nil
}

// handleCursorDown moves cursor down one position with proper threshold handling
func (l *List) handleCursorDown() tea.Cmd {
	if l.totalItems == 0 || !l.canScroll {
		return nil
	}

	// Can't move down if already at the end
	if l.viewport.CursorIndex >= l.totalItems-1 {
		return nil
	}

	previousState := l.viewport
	l.viewport = CalculateCursorDown(l.viewport, l.config.ViewportConfig, l.totalItems)

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
		return l.smartChunkManagement()
	}

	return nil
}

// handlePageUp moves cursor up one page
func (l *List) handlePageUp() tea.Cmd {
	if l.totalItems == 0 || !l.canScroll {
		return nil
	}

	previousState := l.viewport
	l.viewport = CalculatePageUp(l.viewport, l.config.ViewportConfig, l.totalItems)

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
	}

	return l.smartChunkManagement()
}

// handlePageDown moves cursor down one page
func (l *List) handlePageDown() tea.Cmd {
	if l.viewport.CursorIndex >= l.totalItems-1 {
		return nil
	}

	previousState := l.viewport
	l.viewport = CalculatePageDown(l.viewport, l.config.ViewportConfig, l.totalItems)

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
	}

	return l.smartChunkManagement()
}

// handleJumpToStart moves cursor to the start
func (l *List) handleJumpToStart() tea.Cmd {
	if l.totalItems == 0 || !l.canScroll {
		return nil
	}

	l.viewport = CalculateJumpToStart(l.config.ViewportConfig, l.totalItems)
	return l.smartChunkManagement()
}

// handleJumpToEnd moves cursor to the end
func (l *List) handleJumpToEnd() tea.Cmd {
	if l.totalItems <= 0 || !l.canScroll {
		return nil
	}

	previousState := l.viewport
	l.viewport = CalculateJumpToEnd(l.config.ViewportConfig, l.totalItems)

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
		// Use smart chunk management for proper loading feedback
		return l.smartChunkManagement()
	}
	return nil
}

// handleJumpTo moves cursor to a specific index
func (l *List) handleJumpTo(index int) tea.Cmd {
	if l.totalItems == 0 || index < 0 || index >= l.totalItems || !l.canScroll {
		return nil
	}

	l.viewport = CalculateJumpTo(index, l.config.ViewportConfig, l.totalItems)
	return l.smartChunkManagement()
}

// ================================
// DATA MANAGEMENT HELPERS
// ================================

// handleDataRefresh refreshes all data
func (l *List) handleDataRefresh() tea.Cmd {
	l.chunks = make(map[int]Chunk[any])

	if l.dataSource == nil {
		return nil
	}

	var cmds []tea.Cmd
	cmds = append(cmds, l.dataSource.GetTotal())
	cmds = append(cmds, l.loadInitialChunk())

	return tea.Batch(cmds...)
}

// handleDataChunkLoaded processes a loaded data chunk
func (l *List) handleDataChunkLoaded(msg DataChunkLoadedMsg) tea.Cmd {
	chunk := Chunk[any]{
		StartIndex: msg.StartIndex,
		EndIndex:   msg.StartIndex + len(msg.Items) - 1,
		Items:      msg.Items,
		LoadedAt:   time.Now(),
		Request:    msg.Request,
	}

	l.chunks[msg.StartIndex] = chunk

	// Clear loading state for this chunk
	delete(l.loadingChunks, msg.StartIndex)

	// Update loading flags
	l.hasLoadingChunks = len(l.loadingChunks) > 0
	if !l.hasLoadingChunks {
		l.canScroll = true // Re-enable scrolling when all chunks finish loading
	} else {
		// Check if critical chunks are still loading
		l.canScroll = !l.isLoadingCriticalChunks()
	}

	l.updateVisibleItems()
	l.updateViewportBounds()

	var cmds []tea.Cmd

	// Emit chunk loading completed message for observability
	cmds = append(cmds, ChunkLoadingCompletedCmd(msg.StartIndex, len(msg.Items), msg.Request))

	// Unload old chunks
	if unloadCmd := l.unloadOldChunks(); unloadCmd != nil {
		cmds = append(cmds, unloadCmd)
	}

	return tea.Batch(cmds...)
}

// ================================
// SELECTION HELPERS
// ================================

// handleSelectCurrent selects the current item
func (l *List) handleSelectCurrent() tea.Cmd {
	if l.config.SelectionMode == SelectionNone || l.totalItems == 0 {
		return nil
	}

	item, exists := l.getItemAtIndex(l.viewport.CursorIndex)
	if !exists {
		return nil
	}

	return l.toggleItemSelection(item.ID)
}

// handleSelectToggle toggles selection for a specific item
func (l *List) handleSelectToggle(index int) tea.Cmd {
	if l.config.SelectionMode == SelectionNone || index < 0 || index >= l.totalItems {
		return nil
	}

	item, exists := l.getItemAtIndex(index)
	if !exists {
		return nil
	}

	return l.toggleItemSelection(item.ID)
}

// handleSelectAll selects all items via DataSource
func (l *List) handleSelectAll() tea.Cmd {
	if l.config.SelectionMode != SelectionMultiple || l.dataSource == nil {
		return nil
	}

	// Return the command to be processed by Tea model loop
	return l.dataSource.SelectAll()
}

// handleSelectRange selects a range of items
func (l *List) handleSelectRange(startID, endID string) tea.Cmd {
	if l.config.SelectionMode != SelectionMultiple {
		return nil
	}

	startIndex := l.findItemIndex(startID)
	endIndex := l.findItemIndex(endID)

	if startIndex < 0 || endIndex < 0 {
		return nil
	}

	if startIndex > endIndex {
		startIndex, endIndex = endIndex, startIndex
	}

	// Select all items in range (only loaded ones)
	for i := startIndex; i <= endIndex; i++ {
		item, exists := l.getItemAtIndex(i)
		if exists && !l.selectedItems[item.ID] {
			l.selectedItems[item.ID] = true
			l.selectedOrder = append(l.selectedOrder, item.ID)
		}
	}

	return nil
}

// ================================
// FILTER AND SORT HELPERS
// ================================

// handleFilterChange triggers data refresh when filters change
func (l *List) handleFilterChange() tea.Cmd {
	return l.handleDataRefresh()
}

// handleSortToggle toggles sorting on a field
func (l *List) handleSortToggle(field string) tea.Cmd {
	currentSort := SortState{
		Fields:     l.sortFields,
		Directions: l.sortDirs,
	}

	newSort := ToggleSortField(currentSort, field)
	l.sortFields = newSort.Fields
	l.sortDirs = newSort.Directions

	return l.handleDataRefresh()
}

// handleSortSet sets sorting on a field
func (l *List) handleSortSet(field, direction string) tea.Cmd {
	newSort := SetSortField(field, direction)
	l.sortFields = newSort.Fields
	l.sortDirs = newSort.Directions
	return l.handleDataRefresh()
}

// handleSortAdd adds a sort field
func (l *List) handleSortAdd(field, direction string) tea.Cmd {
	currentSort := SortState{
		Fields:     l.sortFields,
		Directions: l.sortDirs,
	}

	newSort := AddSortField(currentSort, field, direction)
	l.sortFields = newSort.Fields
	l.sortDirs = newSort.Directions

	return l.handleDataRefresh()
}

// handleSortRemove removes a sort field
func (l *List) handleSortRemove(field string) tea.Cmd {
	currentSort := SortState{
		Fields:     l.sortFields,
		Directions: l.sortDirs,
	}

	newSort := RemoveSortField(currentSort, field)
	l.sortFields = newSort.Fields
	l.sortDirs = newSort.Directions

	return l.handleDataRefresh()
}

// ================================
// SEARCH HELPERS
// ================================

// handleSearch performs a search
func (l *List) handleSearch() tea.Cmd {
	if l.dataSource == nil {
		return nil
	}

	// For now, return a simple search command
	// Real implementation would depend on DataSource capabilities
	return SearchResultCmd([]int{}, l.searchQuery, 0)
}

// ================================
// KEYBOARD HANDLING
// ================================

// handleKeyPress handles keyboard input
func (l *List) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	if !l.focused {
		return nil
	}

	key := msg.String()

	// Check if key matches any of our navigation keys
	for _, upKey := range l.config.KeyMap.Up {
		if key == upKey {
			return l.handleCursorUp() // Use message handler to emit commands
		}
	}

	for _, downKey := range l.config.KeyMap.Down {
		if key == downKey {
			return l.handleCursorDown() // Use message handler to emit commands
		}
	}

	for _, pageUpKey := range l.config.KeyMap.PageUp {
		if key == pageUpKey {
			return l.handlePageUp() // Use message handler to emit commands
		}
	}

	for _, pageDownKey := range l.config.KeyMap.PageDown {
		if key == pageDownKey {
			return l.handlePageDown() // Use message handler to emit commands
		}
	}

	for _, homeKey := range l.config.KeyMap.Home {
		if key == homeKey {
			return l.handleJumpToStart() // Use message handler to emit commands
		}
	}

	for _, endKey := range l.config.KeyMap.End {
		if key == endKey {
			return l.handleJumpToEnd() // Use message handler to emit commands
		}
	}

	for _, selectKey := range l.config.KeyMap.Select {
		if key == selectKey {
			return SelectCurrentCmd()
		}
	}

	for _, selectAllKey := range l.config.KeyMap.SelectAll {
		if key == selectAllKey {
			return SelectAllCmd()
		}
	}

	for _, filterKey := range l.config.KeyMap.Filter {
		if key == filterKey {
			// Return command to start filtering
			return StatusCmd("Filter mode", StatusInfo)
		}
	}

	for _, sortKey := range l.config.KeyMap.Sort {
		if key == sortKey {
			// Return command to start sorting
			return StatusCmd("Sort mode", StatusInfo)
		}
	}

	return nil
}

// ================================
// RENDERING HELPERS
// ================================

// renderEmpty renders the empty state
func (l *List) renderEmpty() string {
	return RenderEmptyState(l.config.StyleConfig, l.lastError)
}

// renderItem renders a single list item
func (l *List) renderItem(absoluteIndex, viewportIndex int) string {
	item, exists := l.getItemAtIndex(absoluteIndex)
	if !exists {
		return RenderLoadingPlaceholder(l.config.StyleConfig)
	}

	isCursor := absoluteIndex == l.viewport.CursorIndex
	isSelected := item.Selected

	// Use the already-calculated threshold flags from updateViewportBounds()
	// These are the authoritative threshold values calculated based on cursor position
	isTopThreshold := isCursor && l.viewport.IsAtTopThreshold
	isBottomThreshold := isCursor && l.viewport.IsAtBottomThreshold

	// Use animated formatter if available and animations are enabled
	if l.animatedFormatter != nil && l.config.AnimationConfig.Enabled {
		animationState := make(map[string]any)
		if animation, exists := l.animations[item.ID]; exists {
			animationState = animation.State
		}

		result := l.animatedFormatter(
			item,
			absoluteIndex,
			l.renderContext,
			animationState,
			isCursor,
			isTopThreshold,
			isBottomThreshold,
		)

		return l.applyItemStyle(result.Content, isCursor, isSelected, item)
	}

	// Use regular formatter
	var content string
	if l.formatter != nil {
		content = l.formatter(
			item,
			absoluteIndex,
			l.renderContext,
			isCursor,
			isTopThreshold,
			isBottomThreshold,
		)
	} else {
		// Default formatting
		content = fmt.Sprintf("%v", item.Item)
	}

	return l.applyItemStyle(content, isCursor, isSelected, item)
}

// applyItemStyle applies the appropriate style to an item
func (l *List) applyItemStyle(content string, isCursor, isSelected bool, item Data[any]) string {
	return ApplyItemStyle(content, isCursor, isSelected, item, l.config.StyleConfig, l.config.MaxWidth, l.renderContext.Truncate)
}

// ================================
// UTILITY HELPERS
// ================================

// updateViewportPosition updates the viewport based on cursor position
func (l *List) updateViewportPosition() {
	l.viewport = UpdateViewportPosition(l.viewport, l.config.ViewportConfig, l.totalItems)
}

// updateViewportBounds updates viewport boundary flags
func (l *List) updateViewportBounds() {
	l.viewport = UpdateViewportBounds(l.viewport, l.config.ViewportConfig, l.totalItems)
}

// ================================
// BOUNDING CHUNK MANAGEMENT SYSTEM
// ================================

// BoundingArea represents the area around the viewport where chunks should be loaded
type BoundingArea struct {
	StartIndex int // Absolute start index of bounding area
	EndIndex   int // Absolute end index of bounding area (inclusive)
	ChunkStart int // First chunk index in bounding area
	ChunkEnd   int // Last chunk index in bounding area
}

// calculateBoundingArea calculates the bounding area around the current viewport automatically
func (l *List) calculateBoundingArea() BoundingArea {
	return CalculateBoundingArea(l.viewport, l.config.ViewportConfig, l.totalItems)
}

// unloadChunksOutsideBoundingArea unloads chunks that are outside the bounding area
func (l *List) unloadChunksOutsideBoundingArea() tea.Cmd {
	boundingArea := l.calculateBoundingArea()
	chunkSize := l.config.ViewportConfig.ChunkSize
	var cmds []tea.Cmd

	// Find and unload chunks outside the bounding area
	chunksToUnload := FindChunksToUnload(l.chunks, boundingArea, chunkSize)
	for _, chunkStart := range chunksToUnload {
		delete(l.chunks, chunkStart)
		delete(l.chunkAccessTime, chunkStart)
		cmds = append(cmds, ChunkUnloadedCmd(chunkStart))
	}

	return tea.Batch(cmds...)
}

// isChunkLoaded checks if a chunk containing the given index is loaded
func (l *List) isChunkLoaded(index int) bool {
	return IsChunkLoaded(index, l.chunks)
}

// getItemAtIndex retrieves an item at a specific index
func (l *List) getItemAtIndex(index int) (Data[any], bool) {
	return GetItemAtIndex(index, l.chunks, l.totalItems, l.chunkAccessTime)
}

// findItemIndex finds the index of an item by ID
func (l *List) findItemIndex(id string) int {
	return FindItemIndex(id, l.chunks)
}

// toggleItemSelection toggles selection for an item via DataSource
func (l *List) toggleItemSelection(id string) tea.Cmd {
	if l.config.SelectionMode == SelectionNone || l.dataSource == nil {
		return nil
	}

	// Find the item to determine current selection state
	var currentlySelected bool
	var itemIndex int = -1

	for _, chunk := range l.chunks {
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
		return l.dataSource.SetSelected(itemIndex, !currentlySelected)
	}

	return nil
}

// clearSelection clears all selections via DataSource
func (l *List) clearSelection() {
	if l.dataSource == nil {
		return
	}
	// Delegate to DataSource - this will trigger SelectionResponseMsg when completed
	if cmd := l.dataSource.ClearSelection(); cmd != nil {
		// Execute the command immediately since this is a public method
		if msg := cmd(); msg != nil {
			l.Update(msg)
		}
	}
}

// unloadOldChunks unloads chunks that are no longer needed based on smart strategy
func (l *List) unloadOldChunks() tea.Cmd {
	// Calculate the bounds of chunks that should be kept
	keepLowerBound, keepUpperBound := CalculateUnloadBounds(l.viewport, l.config.ViewportConfig)

	var unloadedChunks []int

	// Unload chunks outside the bounds
	for startIndex := range l.chunks {
		if ShouldUnloadChunk(startIndex, keepLowerBound, keepUpperBound) {
			delete(l.chunks, startIndex)
			delete(l.chunkAccessTime, startIndex)
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

// updateVisibleItems updates the slice of items currently visible in the viewport
func (l *List) updateVisibleItems() {
	result := CalculateVisibleItemsFromChunks(
		l.viewport,
		l.config.ViewportConfig,
		l.totalItems,
		l.chunks,
		l.ensureChunkLoadedImmediate,
	)

	l.visibleItems = result.Items
	l.viewport = result.AdjustedViewport
}

// ================================
// HELPER METHODS
// ================================

// ensureChunkLoadedImmediate loads the chunk containing the given index immediately
func (l *List) ensureChunkLoadedImmediate(index int) {
	chunkStartIndex := CalculateChunkStartIndex(index, l.config.ViewportConfig.ChunkSize)
	if _, exists := l.chunks[chunkStartIndex]; !exists {
		// Load this chunk immediately - NO WAITING!
		if l.dataSource != nil {
			request := CreateChunkRequest(
				chunkStartIndex,
				l.config.ViewportConfig.ChunkSize,
				l.totalItems,
				l.sortFields,
				l.sortDirs,
				l.filters,
			)

			// Check if the data source supports immediate loading
			if immediateLoader, ok := l.dataSource.(interface {
				LoadChunkImmediate(DataRequest) DataChunkLoadedMsg
			}); ok {
				// Use immediate loading - FULLY AUTOMATED!
				chunkMsg := immediateLoader.LoadChunkImmediate(request)
				l.handleDataChunkLoaded(chunkMsg)
			} else {
				// Fallback to async loading (not ideal but better than nothing)
				loadCmd := l.dataSource.LoadChunk(request)
				if loadCmd != nil {
					if msg := loadCmd(); msg != nil {
						if chunkMsg, ok := msg.(DataChunkLoadedMsg); ok {
							l.handleDataChunkLoaded(chunkMsg)
						}
					}
				}
			}
		}
	}
}

// smartChunkManagement provides intelligent chunk loading with user feedback
func (l *List) smartChunkManagement() tea.Cmd {
	if l.dataSource == nil {
		return nil
	}

	// Calculate what chunks we need for bounding area
	boundingArea := l.calculateBoundingArea()
	chunkSize := l.config.ViewportConfig.ChunkSize
	var cmds []tea.Cmd
	var newLoadingChunks []int

	// Get chunks that need to be loaded
	chunksToLoad := CalculateChunksInBoundingArea(boundingArea, chunkSize, l.totalItems)

	// Load chunks that aren't already loaded or loading
	for _, chunkStart := range chunksToLoad {
		if !l.isChunkLoaded(chunkStart) && !l.loadingChunks[chunkStart] {
			// Mark chunk as loading
			l.loadingChunks[chunkStart] = true
			newLoadingChunks = append(newLoadingChunks, chunkStart)

			request := CreateChunkRequest(
				chunkStart,
				chunkSize,
				l.totalItems,
				l.sortFields,
				l.sortDirs,
				l.filters,
			)

			// Emit chunk loading started message for observability
			cmds = append(cmds, ChunkLoadingStartedCmd(chunkStart, request))
			cmds = append(cmds, l.dataSource.LoadChunk(request))
		}
	}

	// Update loading state
	if len(newLoadingChunks) > 0 {
		l.hasLoadingChunks = true
		// Block scrolling if we're loading chunks that affect current viewport
		l.canScroll = !l.isLoadingCriticalChunks()
	}

	// Unload chunks outside bounding area
	chunksToUnload := FindChunksToUnload(l.chunks, boundingArea, chunkSize)
	for _, chunkStart := range chunksToUnload {
		delete(l.chunks, chunkStart)
		delete(l.chunkAccessTime, chunkStart)
		cmds = append(cmds, ChunkUnloadedCmd(chunkStart))
	}

	return tea.Batch(cmds...)
}

// isLoadingCriticalChunks checks if we're loading chunks that affect the current viewport
func (l *List) isLoadingCriticalChunks() bool {
	return IsLoadingCriticalChunks(l.viewport, l.config.ViewportConfig, l.loadingChunks)
}

// refreshChunks reloads existing chunks to get updated selection state
func (l *List) refreshChunks() tea.Cmd {
	if l.dataSource == nil {
		return nil
	}

	var cmds []tea.Cmd

	// Reload all currently loaded chunks to get updated selection state
	for chunkStart := range l.chunks {
		request := CreateChunkRequest(
			chunkStart,
			l.config.ViewportConfig.ChunkSize,
			l.totalItems,
			l.sortFields,
			l.sortDirs,
			l.filters,
		)

		// Reload this chunk to get updated selection state
		cmds = append(cmds, l.dataSource.LoadChunk(request))
	}

	return tea.Batch(cmds...)
}

// ================================
// ENHANCED RENDERING METHODS
// ================================

// SetEnumerator sets the list enumerator
func (l *List) SetEnumerator(enum ListEnumerator) {
	l.config.RenderConfig.EnumeratorConfig.Enumerator = enum
}

// SetBulletStyle sets the list to use bullet points
func (l *List) SetBulletStyle() {
	l.config.RenderConfig.EnumeratorConfig.Enumerator = BulletEnumerator
}

// SetNumberedStyle sets the list to use numbered items
func (l *List) SetNumberedStyle() {
	l.config.RenderConfig.EnumeratorConfig.Enumerator = ArabicEnumerator
	l.config.RenderConfig.EnumeratorConfig.Alignment = ListAlignmentRight
}

// SetChecklistStyle sets the list to use checkbox-style items
func (l *List) SetChecklistStyle() {
	l.config.RenderConfig.EnumeratorConfig.Enumerator = CheckboxEnumerator
}

// SetAlphabeticalStyle sets the list to use alphabetical enumeration
func (l *List) SetAlphabeticalStyle() {
	l.config.RenderConfig.EnumeratorConfig.Enumerator = AlphabetEnumerator
	l.config.RenderConfig.EnumeratorConfig.Alignment = ListAlignmentRight
}

// SetDashStyle sets the list to use dash points
func (l *List) SetDashStyle() {
	l.config.RenderConfig.EnumeratorConfig.Enumerator = DashEnumerator
}

// SetConditionalStyle sets the list to use conditional formatting
func (l *List) SetConditionalStyle() {
	conditionalEnum := NewConditionalEnumerator(BulletEnumerator).
		When(IsSelected, CheckboxEnumerator).
		When(IsError, func(item Data[any], index int, ctx RenderContext) string {
			return "✗ "
		}).
		When(IsLoading, func(item Data[any], index int, ctx RenderContext) string {
			return "⟳ "
		})

	l.config.RenderConfig.EnumeratorConfig.Enumerator = conditionalEnum.Enumerate
}

// SetCustomEnumerator sets a custom enumerator pattern
func (l *List) SetCustomEnumerator(pattern string) {
	l.config.RenderConfig.EnumeratorConfig.Enumerator = CustomEnumerator(pattern)
}

// SetRenderConfig sets the complete render configuration
func (l *List) SetRenderConfig(config ListRenderConfig) {
	l.config.RenderConfig = config
}

// GetRenderConfig returns the current render configuration
func (l *List) GetRenderConfig() ListRenderConfig {
	return l.config.RenderConfig
}

// SetEnumeratorAlignment sets whether enumerators should be aligned
func (l *List) SetEnumeratorAlignment(align bool) {
	if align {
		l.config.RenderConfig.EnumeratorConfig.Alignment = ListAlignmentRight
		l.config.RenderConfig.EnumeratorConfig.MaxWidth = 4
	} else {
		l.config.RenderConfig.EnumeratorConfig.Alignment = ListAlignmentNone
		l.config.RenderConfig.EnumeratorConfig.MaxWidth = 0
	}
}

// SetTextWrapping sets whether text should be wrapped
func (l *List) SetTextWrapping(wrap bool) {
	l.config.RenderConfig.ContentConfig.WrapText = wrap
}

// SetIndentSize sets the indentation size for multi-line content
func (l *List) SetIndentSize(size int) {
	// In the new system, indent size is handled automatically by the content component
	// based on the width of preceding components, but we can set max width
	if size > 0 {
		l.config.RenderConfig.ContentConfig.MaxWidth = 80 - size
	}
}

// SetFormatter sets a custom formatter and returns the previous one
func (l *List) SetFormatter(formatter ItemFormatter[any]) ItemFormatter[any] {
	previous := l.formatter
	l.formatter = formatter
	return previous
}

// GetFormatter returns the current formatter
func (l *List) GetFormatter() ItemFormatter[any] {
	return l.formatter
}

// ================================
// STATE INDICATOR CONFIGURATION
// ================================

// SetErrorIndicator sets the error state indicator
func (l *List) SetErrorIndicator(indicator string) {
	l.renderContext.ErrorIndicator = indicator
}

// SetLoadingIndicator sets the loading state indicator
func (l *List) SetLoadingIndicator(indicator string) {
	l.renderContext.LoadingIndicator = indicator
}

// SetDisabledIndicator sets the disabled state indicator
func (l *List) SetDisabledIndicator(indicator string) {
	l.renderContext.DisabledIndicator = indicator
}

// SetSelectedIndicator sets the selected state indicator
func (l *List) SetSelectedIndicator(indicator string) {
	l.renderContext.SelectedIndicator = indicator
}

// GetErrorIndicator returns the current error indicator
func (l *List) GetErrorIndicator() string {
	return l.renderContext.ErrorIndicator
}

// GetLoadingIndicator returns the current loading indicator
func (l *List) GetLoadingIndicator() string {
	return l.renderContext.LoadingIndicator
}

// GetDisabledIndicator returns the current disabled indicator
func (l *List) GetDisabledIndicator() string {
	return l.renderContext.DisabledIndicator
}

// GetSelectedIndicator returns the current selected indicator
func (l *List) GetSelectedIndicator() string {
	return l.renderContext.SelectedIndicator
}
