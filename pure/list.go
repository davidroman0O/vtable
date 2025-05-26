package vtable

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
			// Use formatter if available
			ctx := RenderContext{
				MaxWidth:  80,
				MaxHeight: 1,
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
			// Default formatting
			prefix := "  "
			if isCursor {
				prefix = "> "
			}

			if isSelected {
				prefix = "✓ " + prefix
			}

			renderedItem = prefix + fmt.Sprintf("%v", item.Item)
		}

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

// GetTotalItems returns the total number of items
func (l *List) GetTotalItems() int {
	return l.totalItems
}

// GetSelectedIndices returns the indices of selected items
func (l *List) GetSelectedIndices() []int {
	var indices []int
	// Read selection state from chunks (DataSource owns the state)
	for _, chunk := range l.chunks {
		for i, item := range chunk.Items {
			if item.Selected {
				indices = append(indices, chunk.StartIndex+i)
			}
		}
	}
	return indices
}

// GetSelectedIDs returns the IDs of selected items
func (l *List) GetSelectedIDs() []string {
	var ids []string
	// Read selection state from chunks (DataSource owns the state)
	for _, chunk := range l.chunks {
		for _, item := range chunk.Items {
			if item.Selected {
				ids = append(ids, item.ID)
			}
		}
	}
	return ids
}

// GetSelectionCount returns the number of selected items
func (l *List) GetSelectionCount() int {
	count := 0
	// Read selection state from chunks (DataSource owns the state)
	for _, chunk := range l.chunks {
		for _, item := range chunk.Items {
			if item.Selected {
				count++
			}
		}
	}
	return count
}

// SetFormatter sets the item formatter
func (l *List) SetFormatter(formatter ItemFormatter[any]) tea.Cmd {
	return FormatterSetCmd(formatter)
}

// SetAnimatedFormatter sets the animated item formatter
func (l *List) SetAnimatedFormatter(formatter ItemFormatterAnimated[any]) tea.Cmd {
	return AnimatedFormatterSetCmd(formatter)
}

// SetMaxWidth sets the maximum width
func (l *List) SetMaxWidth(width int) tea.Cmd {
	return MaxWidthSetCmd(width)
}

// GetCurrentItem returns the item at the cursor position
func (l *List) GetCurrentItem() (Data[any], bool) {
	item, exists := l.getItemAtIndex(l.viewport.CursorIndex)
	return item, exists
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
	topThreshold := l.config.ViewportConfig.TopThreshold

	// Handle top threshold logic (only if thresholds are enabled)
	if previousState.IsAtTopThreshold && !l.viewport.AtDatasetStart && topThreshold >= 0 {
		// Cursor was at the top threshold, scroll viewport up while keeping cursor at threshold
		if l.viewport.ViewportStartIndex > 0 {
			l.viewport.ViewportStartIndex--
			l.viewport.CursorViewportIndex = topThreshold // LOCK cursor at threshold
			// Update absolute cursor position based on new viewport
			l.viewport.CursorIndex = l.viewport.ViewportStartIndex + l.viewport.CursorViewportIndex
		} else {
			// Can't scroll viewport up anymore, move cursor normally
			l.viewport.CursorIndex--
			l.viewport.CursorViewportIndex--
		}
	} else if topThreshold < 0 {
		// Thresholds disabled - use pure edge-based scrolling
		l.viewport.CursorIndex-- // Move cursor normally
		// Move cursor within viewport if possible, otherwise scroll
		if previousState.CursorViewportIndex > 0 {
			// Cursor can move within viewport
			l.viewport.CursorViewportIndex--
		} else {
			// Cursor is at top edge of viewport - scroll if possible
			if l.viewport.ViewportStartIndex > 0 {
				l.viewport.ViewportStartIndex--
				l.viewport.CursorViewportIndex = 0
			}
		}
	} else {
		// Thresholds enabled - move cursor normally, let viewport follow
		l.viewport.CursorIndex-- // Move cursor first

		if previousState.CursorViewportIndex > 0 {
			// Cursor not at threshold, move within viewport
			l.viewport.CursorViewportIndex--
		} else {
			// At viewport top edge, scroll if possible
			if l.viewport.ViewportStartIndex > 0 {
				l.viewport.ViewportStartIndex--
				l.viewport.CursorViewportIndex = 0
			} else {
				// Can't scroll, cursor stays at top
				l.viewport.CursorViewportIndex = 0
			}
		}
	}

	// Final safety check - ensure cursor doesn't go negative
	if l.viewport.CursorIndex < 0 {
		l.viewport.CursorIndex = 0
		l.viewport.CursorViewportIndex = 0
	}

	l.updateViewportBounds()

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
	bottomThreshold := l.config.ViewportConfig.BottomThreshold

	// Handle bottom threshold logic (only if thresholds are enabled)
	if previousState.IsAtBottomThreshold && !l.viewport.AtDatasetEnd && bottomThreshold >= 0 {
		// Cursor was at the bottom threshold, scroll viewport down while keeping cursor at threshold
		l.viewport.ViewportStartIndex++
		bottomPosition := l.config.ViewportConfig.Height - bottomThreshold - 1
		l.viewport.CursorViewportIndex = bottomPosition // LOCK cursor at threshold
		// Update absolute cursor position based on new viewport
		l.viewport.CursorIndex = l.viewport.ViewportStartIndex + l.viewport.CursorViewportIndex
	} else if bottomThreshold < 0 {
		// Thresholds disabled - use pure edge-based scrolling
		l.viewport.CursorIndex++ // Move cursor normally
		// Move cursor within viewport if possible, otherwise scroll
		if previousState.CursorViewportIndex < l.config.ViewportConfig.Height-1 {
			// Cursor can move within viewport
			l.viewport.CursorViewportIndex++
		} else {
			// Cursor is at bottom edge of viewport - scroll if possible
			if l.viewport.ViewportStartIndex+l.config.ViewportConfig.Height < l.totalItems {
				l.viewport.ViewportStartIndex++
				l.viewport.CursorViewportIndex = l.config.ViewportConfig.Height - 1
			}
		}
	} else {
		// Thresholds enabled - move cursor normally, let viewport follow
		l.viewport.CursorIndex++ // Move cursor first

		// Ensure we don't exceed actual data count
		if l.viewport.CursorIndex >= l.totalItems {
			l.viewport.CursorIndex = l.totalItems - 1
			// If we're already at the last item, no need to continue
			if l.viewport.CursorIndex == previousState.CursorIndex {
				return nil
			}
		}

		if previousState.CursorViewportIndex < l.config.ViewportConfig.Height-1 &&
			l.viewport.ViewportStartIndex+previousState.CursorViewportIndex+1 < l.totalItems {
			// Cursor not at threshold, move within viewport
			l.viewport.CursorViewportIndex++
		} else {
			// At viewport bottom edge, scroll if possible
			if l.viewport.ViewportStartIndex+l.config.ViewportConfig.Height < l.totalItems {
				l.viewport.ViewportStartIndex++
			}
			l.viewport.CursorViewportIndex = l.viewport.CursorIndex - l.viewport.ViewportStartIndex
		}
	}

	// Final boundary check - ensure we're not beyond data
	if l.viewport.CursorIndex >= l.totalItems {
		l.viewport.CursorIndex = l.totalItems - 1
		l.viewport.CursorViewportIndex = l.viewport.CursorIndex - l.viewport.ViewportStartIndex
	}

	// Ensure cursor viewport index is within bounds
	if l.viewport.CursorViewportIndex < 0 {
		l.viewport.CursorViewportIndex = 0
		l.viewport.CursorIndex = l.viewport.ViewportStartIndex
	}

	l.updateViewportBounds()

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

	pageSize := l.config.ViewportConfig.Height
	newIndex := l.viewport.CursorIndex - pageSize
	if newIndex < 0 {
		newIndex = 0
	}

	l.viewport.CursorIndex = newIndex
	l.updateViewportPosition()
	return l.smartChunkManagement()
}

// handlePageDown moves cursor down one page
func (l *List) handlePageDown() tea.Cmd {
	if l.totalItems == 0 || !l.canScroll {
		return nil
	}

	pageSize := l.config.ViewportConfig.Height
	newIndex := l.viewport.CursorIndex + pageSize
	if newIndex >= l.totalItems {
		newIndex = l.totalItems - 1
	}

	l.viewport.CursorIndex = newIndex
	l.updateViewportPosition()
	return l.smartChunkManagement()
}

// handleJumpToStart moves cursor to the start
func (l *List) handleJumpToStart() tea.Cmd {
	if l.totalItems == 0 || !l.canScroll {
		return nil
	}

	l.viewport.CursorIndex = 0
	l.updateViewportPosition()
	return l.smartChunkManagement()
}

// handleJumpToEnd moves cursor to the end
func (l *List) handleJumpToEnd() tea.Cmd {
	if l.totalItems <= 0 || !l.canScroll {
		return nil
	}

	previousState := l.viewport

	l.viewport.CursorIndex = l.totalItems - 1

	// Calculate viewport start to show the cursor at the bottom threshold (or bottom if small dataset)
	if l.totalItems <= l.config.ViewportConfig.Height {
		l.viewport.ViewportStartIndex = 0
		l.viewport.CursorViewportIndex = l.totalItems - 1
	} else {
		l.viewport.ViewportStartIndex = l.totalItems - l.config.ViewportConfig.Height
		l.viewport.CursorViewportIndex = l.config.ViewportConfig.Height - 1
	}

	l.viewport.IsAtTopThreshold = false
	l.viewport.IsAtBottomThreshold = false

	// Only set threshold flags if thresholds are enabled
	if l.config.ViewportConfig.TopThreshold >= 0 && l.config.ViewportConfig.TopThreshold < l.config.ViewportConfig.Height {
		l.viewport.IsAtTopThreshold = l.viewport.CursorViewportIndex == l.config.ViewportConfig.TopThreshold
	}
	if l.config.ViewportConfig.BottomThreshold >= 0 && l.config.ViewportConfig.BottomThreshold < l.config.ViewportConfig.Height {
		l.viewport.IsAtBottomThreshold = l.viewport.CursorViewportIndex == l.config.ViewportConfig.BottomThreshold
	}

	// Update dataset boundary flags
	l.viewport.AtDatasetStart = l.viewport.ViewportStartIndex == 0
	l.viewport.AtDatasetEnd = true

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

	l.viewport.CursorIndex = index
	l.updateViewportPosition()
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
	// Find field in current sort
	for i, sortField := range l.sortFields {
		if sortField == field {
			// Toggle direction
			if l.sortDirs[i] == "asc" {
				l.sortDirs[i] = "desc"
			} else {
				l.sortDirs[i] = "asc"
			}
			return l.handleDataRefresh()
		}
	}

	// Field not found, add it
	l.sortFields = append(l.sortFields, field)
	l.sortDirs = append(l.sortDirs, "asc")
	return l.handleDataRefresh()
}

// handleSortSet sets sorting on a field
func (l *List) handleSortSet(field, direction string) tea.Cmd {
	l.sortFields = []string{field}
	l.sortDirs = []string{direction}
	return l.handleDataRefresh()
}

// handleSortAdd adds a sort field
func (l *List) handleSortAdd(field, direction string) tea.Cmd {
	// Remove field if it already exists
	for i, sortField := range l.sortFields {
		if sortField == field {
			l.sortFields = append(l.sortFields[:i], l.sortFields[i+1:]...)
			l.sortDirs = append(l.sortDirs[:i], l.sortDirs[i+1:]...)
			break
		}
	}

	// Add to end
	l.sortFields = append(l.sortFields, field)
	l.sortDirs = append(l.sortDirs, direction)
	return l.handleDataRefresh()
}

// handleSortRemove removes a sort field
func (l *List) handleSortRemove(field string) tea.Cmd {
	for i, sortField := range l.sortFields {
		if sortField == field {
			l.sortFields = append(l.sortFields[:i], l.sortFields[i+1:]...)
			l.sortDirs = append(l.sortDirs[:i], l.sortDirs[i+1:]...)
			return l.handleDataRefresh()
		}
	}
	return nil
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
	style := l.config.StyleConfig.DefaultStyle
	if l.lastError != nil {
		style = l.config.StyleConfig.ErrorStyle
		return style.Render("Error: " + l.lastError.Error())
	}
	return style.Render("No items")
}

// renderItem renders a single list item
func (l *List) renderItem(absoluteIndex, viewportIndex int) string {
	item, exists := l.getItemAtIndex(absoluteIndex)
	if !exists {
		return l.config.StyleConfig.LoadingStyle.Render("Loading...")
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
	var style lipgloss.Style

	switch {
	case item.Error != nil:
		style = l.config.StyleConfig.ErrorStyle
	case item.Loading:
		style = l.config.StyleConfig.LoadingStyle
	case item.Disabled:
		style = l.config.StyleConfig.DisabledStyle
	case isCursor && isSelected:
		// Combine cursor and selected styles
		style = l.config.StyleConfig.CursorStyle.Copy().
			Background(l.config.StyleConfig.SelectedStyle.GetBackground())
	case isCursor:
		style = l.config.StyleConfig.CursorStyle
	case isSelected:
		style = l.config.StyleConfig.SelectedStyle
	default:
		style = l.config.StyleConfig.DefaultStyle
	}

	// Truncate content to max width
	if l.config.MaxWidth > 0 && len(content) > l.config.MaxWidth {
		content = l.renderContext.Truncate(content, l.config.MaxWidth)
	}

	return style.Render(content)
}

// ================================
// UTILITY HELPERS
// ================================

// updateViewportPosition updates the viewport based on cursor position
func (l *List) updateViewportPosition() {
	if l.totalItems == 0 {
		return
	}

	height := l.config.ViewportConfig.Height

	// Calculate relative position within viewport
	l.viewport.CursorViewportIndex = l.viewport.CursorIndex - l.viewport.ViewportStartIndex

	// Adjust viewport if cursor is outside
	if l.viewport.CursorViewportIndex < 0 {
		l.viewport.ViewportStartIndex = l.viewport.CursorIndex
		l.viewport.CursorViewportIndex = 0
	} else if l.viewport.CursorViewportIndex >= height {
		l.viewport.ViewportStartIndex = l.viewport.CursorIndex - height + 1
		l.viewport.CursorViewportIndex = height - 1
	}

	l.updateViewportBounds()
}

// updateViewportBounds updates viewport boundary flags
func (l *List) updateViewportBounds() {
	height := l.config.ViewportConfig.Height
	topThreshold := l.config.ViewportConfig.TopThreshold
	bottomThreshold := l.config.ViewportConfig.BottomThreshold

	// Update threshold flags using offset semantics
	// TopThreshold: offset from viewport start (e.g., TopThreshold=2 means position 2)
	// BottomThreshold: offset from viewport end (e.g., BottomThreshold=2 means position height-2-1)
	l.viewport.IsAtTopThreshold = false
	l.viewport.IsAtBottomThreshold = false

	if topThreshold >= 0 && topThreshold < height {
		l.viewport.IsAtTopThreshold = l.viewport.CursorViewportIndex == topThreshold
	}

	if bottomThreshold >= 0 && bottomThreshold < height {
		// BottomThreshold is offset from end: if height=8 and bottomThreshold=2, then position is 8-2-1=5
		bottomPosition := height - bottomThreshold - 1
		if bottomPosition >= 0 && bottomPosition < height {
			l.viewport.IsAtBottomThreshold = l.viewport.CursorViewportIndex == bottomPosition
		}
	}

	// Update dataset boundary flags
	l.viewport.AtDatasetStart = l.viewport.ViewportStartIndex == 0
	l.viewport.AtDatasetEnd = l.viewport.ViewportStartIndex+height >= l.totalItems
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
	if l.totalItems == 0 {
		return BoundingArea{}
	}

	chunkSize := l.config.ViewportConfig.ChunkSize
	viewportHeight := l.config.ViewportConfig.Height
	boundingBefore := l.config.ViewportConfig.BoundingAreaBefore
	boundingAfter := l.config.ViewportConfig.BoundingAreaAfter

	// Calculate viewport bounds (item indices)
	viewportStart := l.viewport.ViewportStartIndex
	viewportEnd := viewportStart + viewportHeight - 1

	// FULLY AUTOMATED BOUNDING AREA CALCULATION
	// Automatically calculate bounding area based on current viewport position
	// using the configured distances (boundingBefore/boundingAfter items)
	boundingStartIndex := viewportStart - boundingBefore
	boundingEndIndex := viewportEnd + boundingAfter

	// Clamp to dataset bounds
	if boundingStartIndex < 0 {
		boundingStartIndex = 0
	}
	if boundingEndIndex >= l.totalItems {
		boundingEndIndex = l.totalItems - 1
	}

	// Find which chunks intersect with this bounding area
	firstChunkStart := (boundingStartIndex / chunkSize) * chunkSize
	lastChunkStart := (boundingEndIndex / chunkSize) * chunkSize

	// ChunkEnd is the boundary for the loop (exclusive)
	chunkEnd := lastChunkStart + chunkSize

	return BoundingArea{
		StartIndex: boundingStartIndex,
		EndIndex:   boundingEndIndex,
		ChunkStart: firstChunkStart,
		ChunkEnd:   chunkEnd,
	}
}

// ensureBoundingAreaLoaded ensures all chunks in the bounding area are loaded
func (l *List) ensureBoundingAreaLoaded() tea.Cmd {
	if l.dataSource == nil {
		return nil
	}

	boundingArea := l.calculateBoundingArea()
	chunkSize := l.config.ViewportConfig.ChunkSize
	var cmds []tea.Cmd

	// Load all chunks that should be in the bounding area
	for chunkStart := boundingArea.ChunkStart; chunkStart < boundingArea.ChunkEnd; chunkStart += chunkSize {
		if chunkStart >= l.totalItems {
			break // Don't load chunks beyond dataset
		}

		if !l.isChunkLoaded(chunkStart) {
			// Calculate actual chunk size (might be smaller at the end)
			actualChunkSize := chunkSize
			if chunkStart+chunkSize > l.totalItems {
				actualChunkSize = l.totalItems - chunkStart
			}

			request := DataRequest{
				Start:          chunkStart,
				Count:          actualChunkSize,
				SortFields:     l.sortFields,
				SortDirections: l.sortDirs,
				Filters:        l.filters,
			}
			cmds = append(cmds, l.dataSource.LoadChunk(request))
		}
	}

	return tea.Batch(cmds...)
}

// unloadChunksOutsideBoundingArea unloads chunks that are outside the bounding area
func (l *List) unloadChunksOutsideBoundingArea() tea.Cmd {
	boundingArea := l.calculateBoundingArea()
	chunkSize := l.config.ViewportConfig.ChunkSize
	var cmds []tea.Cmd

	// Unload chunks that do NOT intersect with the bounding area
	for chunkStart := range l.chunks {
		chunkEnd := chunkStart + chunkSize - 1

		// Check if this chunk intersects with the bounding area
		// A chunk intersects if: chunkStart <= boundingArea.EndIndex AND chunkEnd >= boundingArea.StartIndex
		doesIntersect := chunkStart <= boundingArea.EndIndex && chunkEnd >= boundingArea.StartIndex

		if !doesIntersect {
			// This chunk is completely outside the bounding area, unload it
			delete(l.chunks, chunkStart)
			delete(l.chunkAccessTime, chunkStart)
			cmds = append(cmds, ChunkUnloadedCmd(chunkStart))
		}
	}

	return tea.Batch(cmds...)
}

// manageBoundingArea is the main method that manages the bounding area
// It ensures chunks are loaded proactively and unloads distant chunks
func (l *List) manageBoundingArea() tea.Cmd {
	var cmds []tea.Cmd

	// First, ensure all chunks in the bounding area are loaded
	if loadCmd := l.ensureBoundingAreaLoaded(); loadCmd != nil {
		cmds = append(cmds, loadCmd)
	}

	// Then, unload chunks outside the bounding area
	if unloadCmd := l.unloadChunksOutsideBoundingArea(); unloadCmd != nil {
		cmds = append(cmds, unloadCmd)
	}

	return tea.Batch(cmds...)
}

// checkAndLoadChunks is now replaced by the bounding area system
func (l *List) checkAndLoadChunks() tea.Cmd {
	if l.dataSource == nil {
		return nil
	}

	// Calculate bounding area and load necessary chunks
	boundingArea := l.calculateBoundingArea()
	chunkSize := l.config.ViewportConfig.ChunkSize
	var cmds []tea.Cmd

	// Load all chunks that should be in the bounding area
	for chunkStart := boundingArea.ChunkStart; chunkStart < boundingArea.ChunkEnd; chunkStart += chunkSize {
		if chunkStart >= l.totalItems {
			break // Don't load chunks beyond dataset
		}

		if !l.isChunkLoaded(chunkStart) {
			// Calculate actual chunk size (might be smaller at the end)
			actualChunkSize := chunkSize
			if chunkStart+chunkSize > l.totalItems {
				actualChunkSize = l.totalItems - chunkStart
			}

			request := DataRequest{
				Start:          chunkStart,
				Count:          actualChunkSize,
				SortFields:     l.sortFields,
				SortDirections: l.sortDirs,
				Filters:        l.filters,
			}
			cmds = append(cmds, l.dataSource.LoadChunk(request))
		}
	}

	// Also unload chunks outside bounding area
	for chunkStart := range l.chunks {
		chunkEnd := chunkStart + chunkSize - 1

		// Check if this chunk intersects with the bounding area
		// A chunk intersects if: chunkStart <= boundingArea.EndIndex AND chunkEnd >= boundingArea.StartIndex
		doesIntersect := chunkStart <= boundingArea.EndIndex && chunkEnd >= boundingArea.StartIndex

		if !doesIntersect {
			// This chunk is completely outside the bounding area, unload it
			delete(l.chunks, chunkStart)
			delete(l.chunkAccessTime, chunkStart)
			cmds = append(cmds, ChunkUnloadedCmd(chunkStart))
		}
	}

	return tea.Batch(cmds...)
}

// isChunkLoaded checks if a chunk containing the given index is loaded
func (l *List) isChunkLoaded(index int) bool {
	for _, chunk := range l.chunks {
		if index >= chunk.StartIndex && index <= chunk.EndIndex {
			return true
		}
	}
	return false
}

// getItemAtIndex retrieves an item at a specific index
func (l *List) getItemAtIndex(index int) (Data[any], bool) {
	if index < 0 || index >= l.totalItems {
		return Data[any]{}, false
	}

	// Find chunk containing this index
	for chunkStart, chunk := range l.chunks {
		if index >= chunk.StartIndex && index <= chunk.EndIndex {
			// Update access time for LRU management
			l.chunkAccessTime[chunkStart] = time.Now()

			chunkIndex := index - chunk.StartIndex
			if chunkIndex < len(chunk.Items) {
				return chunk.Items[chunkIndex], true
			}
		}
	}

	return Data[any]{}, false
}

// findItemIndex finds the index of an item by ID
func (l *List) findItemIndex(id string) int {
	for _, chunk := range l.chunks {
		for i, item := range chunk.Items {
			if item.ID == id {
				return chunk.StartIndex + i
			}
		}
	}
	return -1
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

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// unloadOldChunks unloads chunks that are no longer needed based on smart strategy
func (l *List) unloadOldChunks() tea.Cmd {
	// Calculate the bounds of chunks that should be kept
	viewportChunkIndex := (l.viewport.ViewportStartIndex / l.config.ViewportConfig.ChunkSize) * l.config.ViewportConfig.ChunkSize
	keepLowerBound := viewportChunkIndex - l.config.ViewportConfig.ChunkSize
	if keepLowerBound < 0 {
		keepLowerBound = 0
	}
	keepUpperBound := viewportChunkIndex + (2 * l.config.ViewportConfig.ChunkSize)

	var unloadedChunks []int

	// Unload chunks outside the bounds
	for startIndex := range l.chunks {
		if startIndex < keepLowerBound || startIndex > keepUpperBound {
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
	// If there's no data, return an empty slice
	if l.totalItems == 0 {
		l.visibleItems = []Data[any]{}
		return
	}

	// Calculate how many actual items we can show
	maxVisibleItems := l.config.ViewportConfig.Height
	if l.totalItems < maxVisibleItems {
		maxVisibleItems = l.totalItems
	}

	// Ensure viewport doesn't extend beyond dataset
	maxStart := l.totalItems - maxVisibleItems
	if l.viewport.ViewportStartIndex > maxStart {
		l.viewport.ViewportStartIndex = maxStart
	}
	if l.viewport.ViewportStartIndex < 0 {
		l.viewport.ViewportStartIndex = 0
	}

	// Calculate endpoint of visible area (exclusive)
	viewportEnd := l.viewport.ViewportStartIndex + maxVisibleItems
	if viewportEnd > l.totalItems {
		viewportEnd = l.totalItems
	}

	// Create a new slice to hold visible items
	l.visibleItems = make([]Data[any], 0, viewportEnd-l.viewport.ViewportStartIndex)

	// Fill the visible items slice with actual data - ENSURE CHUNKS ARE LOADED!
	for i := l.viewport.ViewportStartIndex; i < viewportEnd; i++ {
		// Get the chunk that contains this item
		chunkStartIndex := (i / l.config.ViewportConfig.ChunkSize) * l.config.ViewportConfig.ChunkSize
		chunk, ok := l.chunks[chunkStartIndex]

		// If chunk isn't loaded, load it immediately - NO WAITING!
		if !ok {
			l.ensureChunkLoadedImmediate(chunkStartIndex)
			// Try to get the chunk again after loading
			chunk, ok = l.chunks[chunkStartIndex]
		}

		// If we still don't have the chunk, something is wrong - create a placeholder
		if !ok {
			l.visibleItems = append(l.visibleItems, Data[any]{
				ID:   fmt.Sprintf("loading-%d", i),
				Item: fmt.Sprintf("Loading item %d...", i),
			})
			continue
		}

		// Calculate item index within the chunk
		itemIndex := i - chunk.StartIndex

		// Only add the item if it's within the chunk's bounds
		if itemIndex >= 0 && itemIndex < len(chunk.Items) {
			l.visibleItems = append(l.visibleItems, chunk.Items[itemIndex])
		} else {
			// Item not in chunk bounds - create placeholder
			l.visibleItems = append(l.visibleItems, Data[any]{
				ID:   fmt.Sprintf("missing-%d", i),
				Item: fmt.Sprintf("Missing item %d", i),
			})
		}
	}

	// Ensure cursor stays within bounds of visible data
	if l.viewport.CursorViewportIndex >= len(l.visibleItems) {
		if len(l.visibleItems) > 0 {
			l.viewport.CursorViewportIndex = len(l.visibleItems) - 1
		} else {
			l.viewport.CursorViewportIndex = 0
		}
		// Adjust absolute cursor position
		l.viewport.CursorIndex = l.viewport.ViewportStartIndex + l.viewport.CursorViewportIndex
	}
}

// ================================
// HELPER METHODS
// ================================

// ensureChunkLoaded loads the chunk containing the given index if not already loaded
func (l *List) ensureChunkLoaded(index int) {
	chunkStartIndex := (index / l.config.ViewportConfig.ChunkSize) * l.config.ViewportConfig.ChunkSize
	if _, exists := l.chunks[chunkStartIndex]; !exists {
		// Load this chunk synchronously - no "Loading..." placeholders!
		if l.dataSource != nil {
			loadCmd := l.dataSource.LoadChunk(DataRequest{
				Start: chunkStartIndex,
				Count: l.config.ViewportConfig.ChunkSize,
			})
			// Execute the command immediately to get the data
			if loadCmd != nil {
				if msg := loadCmd(); msg != nil {
					// Process the chunk loaded message immediately
					if chunkMsg, ok := msg.(DataChunkLoadedMsg); ok {
						l.handleDataChunkLoaded(chunkMsg)
					}
				}
			}
		}
	}
}

// ensureChunkLoadedImmediate loads the chunk containing the given index immediately
func (l *List) ensureChunkLoadedImmediate(index int) {
	chunkStartIndex := (index / l.config.ViewportConfig.ChunkSize) * l.config.ViewportConfig.ChunkSize
	if _, exists := l.chunks[chunkStartIndex]; !exists {
		// Load this chunk immediately - NO WAITING!
		if l.dataSource != nil {
			// Calculate actual chunk size (might be smaller at the end)
			actualChunkSize := l.config.ViewportConfig.ChunkSize
			if chunkStartIndex+actualChunkSize > l.totalItems {
				actualChunkSize = l.totalItems - chunkStartIndex
			}

			request := DataRequest{
				Start:          chunkStartIndex,
				Count:          actualChunkSize,
				SortFields:     l.sortFields,
				SortDirections: l.sortDirs,
				Filters:        l.filters,
			}

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

// ================================
// NAVIGATION METHODS (Direct state manipulation)
// ================================

// MoveUp moves the cursor up one position
func (l *List) MoveUp() {
	if l.totalItems <= 0 || l.viewport.CursorIndex <= 0 {
		return
	}

	previousState := l.viewport
	l.viewport.CursorIndex--

	topThreshold := l.config.ViewportConfig.TopThreshold

	// Handle top threshold logic (only if thresholds are enabled)
	if previousState.IsAtTopThreshold && !l.viewport.AtDatasetStart && topThreshold >= 0 {
		// Cursor was at the top threshold, scroll viewport up while keeping cursor at threshold
		l.viewport.ViewportStartIndex--
		l.viewport.CursorViewportIndex = topThreshold // LOCK cursor at threshold
		// Update absolute cursor position based on new viewport
		l.viewport.CursorIndex = l.viewport.ViewportStartIndex + l.viewport.CursorViewportIndex
	} else if topThreshold < 0 {
		// Thresholds disabled - use pure edge-based scrolling
		l.viewport.CursorIndex-- // Move cursor normally
		// Move cursor within viewport if possible, otherwise scroll
		if previousState.CursorViewportIndex > 0 {
			// Cursor can move within viewport
			l.viewport.CursorViewportIndex--
		} else {
			// Cursor is at top edge of viewport - scroll if possible
			if l.viewport.ViewportStartIndex > 0 {
				l.viewport.ViewportStartIndex--
				l.viewport.CursorViewportIndex = 0
			}
		}
	} else {
		// Thresholds enabled - move cursor normally, let viewport follow
		l.viewport.CursorIndex-- // Move cursor first

		if previousState.CursorViewportIndex > 0 {
			// Cursor not at threshold, move within viewport
			l.viewport.CursorViewportIndex--
		} else {
			// At viewport top edge, scroll if possible
			if l.viewport.ViewportStartIndex > 0 {
				l.viewport.ViewportStartIndex--
				l.viewport.CursorViewportIndex = 0
			} else {
				// Can't scroll, cursor stays at top
				l.viewport.CursorViewportIndex = 0
			}
		}
	}

	// Final safety check - ensure cursor doesn't go negative
	if l.viewport.CursorIndex < 0 {
		l.viewport.CursorIndex = 0
		l.viewport.CursorViewportIndex = 0
	}

	l.updateViewportBounds()

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
		// NOTE: manageBoundingArea() removed - public methods can't handle async commands
		// Bounding area management happens through the command handlers
	}
}

// MoveDown moves the cursor down one position
func (l *List) MoveDown() {
	if l.totalItems <= 0 || l.viewport.CursorIndex >= l.totalItems-1 {
		return
	}

	previousState := l.viewport
	l.viewport.CursorIndex++

	bottomThreshold := l.config.ViewportConfig.BottomThreshold

	// Handle bottom threshold logic (only if thresholds are enabled)
	if previousState.IsAtBottomThreshold && !l.viewport.AtDatasetEnd && bottomThreshold >= 0 {
		// Cursor was at the bottom threshold, scroll viewport down
		l.viewport.ViewportStartIndex++
		l.viewport.CursorViewportIndex = bottomThreshold
	} else if bottomThreshold < 0 {
		// Thresholds disabled - use pure edge-based scrolling
		// Move cursor within viewport if possible, otherwise scroll
		if previousState.CursorViewportIndex < l.config.ViewportConfig.Height-1 {
			// Cursor can move within viewport
			l.viewport.CursorViewportIndex++
		} else {
			// Cursor is at bottom edge of viewport - scroll if possible
			if l.viewport.ViewportStartIndex+l.config.ViewportConfig.Height < l.totalItems {
				l.viewport.ViewportStartIndex++
				l.viewport.CursorViewportIndex = l.config.ViewportConfig.Height - 1
			}
		}
	} else {
		// Thresholds enabled - traditional threshold-based logic
		if previousState.CursorViewportIndex < l.config.ViewportConfig.Height-1 &&
			l.viewport.ViewportStartIndex+previousState.CursorViewportIndex+1 < l.totalItems {
			// Cursor not at threshold, move within viewport
			l.viewport.CursorViewportIndex++
		} else {
			// At viewport bottom edge, scroll if possible
			if l.viewport.ViewportStartIndex+l.config.ViewportConfig.Height < l.totalItems {
				l.viewport.ViewportStartIndex++
			}
			l.viewport.CursorViewportIndex = l.viewport.CursorIndex - l.viewport.ViewportStartIndex
		}
	}

	// Final boundary check - ensure we're not beyond data
	if l.viewport.CursorIndex >= l.totalItems {
		l.viewport.CursorIndex = l.totalItems - 1
		l.viewport.CursorViewportIndex = l.viewport.CursorIndex - l.viewport.ViewportStartIndex
	}

	// Ensure cursor viewport index is within bounds
	if l.viewport.CursorViewportIndex < 0 {
		l.viewport.CursorViewportIndex = 0
		l.viewport.CursorIndex = l.viewport.ViewportStartIndex
	}

	l.updateViewportBounds()

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
		// NOTE: manageBoundingArea() removed - public methods can't handle async commands
		// Bounding area management happens through the command handlers
	}
}

// PageUp moves the cursor up by a page (viewport height)
func (l *List) PageUp() {
	if l.viewport.CursorIndex <= 0 {
		return
	}

	previousState := l.viewport
	moveCount := l.config.ViewportConfig.Height

	if moveCount > l.viewport.CursorIndex {
		moveCount = l.viewport.CursorIndex
	}

	l.viewport.CursorIndex -= moveCount
	l.updateViewportPosition()

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
		// NOTE: manageBoundingArea() removed - public methods can't handle async commands
	}
}

// PageDown moves the cursor down by a page (viewport height)
func (l *List) PageDown() {
	if l.viewport.CursorIndex >= l.totalItems-1 {
		return
	}

	previousState := l.viewport
	moveCount := l.config.ViewportConfig.Height

	if l.viewport.CursorIndex+moveCount >= l.totalItems {
		moveCount = l.totalItems - 1 - l.viewport.CursorIndex
	}

	l.viewport.CursorIndex += moveCount
	l.updateViewportPosition()

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
		// NOTE: manageBoundingArea() removed - public methods can't handle async commands
	}
}

// JumpToStart jumps to the start of the dataset
func (l *List) JumpToStart() {
	if l.totalItems <= 0 {
		return
	}

	previousState := l.viewport
	l.viewport.CursorIndex = 0
	l.viewport.ViewportStartIndex = 0
	l.viewport.CursorViewportIndex = 0
	l.viewport.AtDatasetStart = true
	l.viewport.AtDatasetEnd = l.totalItems <= l.config.ViewportConfig.Height

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
		// NOTE: manageBoundingArea() removed - public methods can't handle async commands
	}
}

// JumpToEnd jumps to the end of the dataset
func (l *List) JumpToEnd() {
	if l.totalItems <= 0 {
		return
	}

	previousState := l.viewport
	l.viewport.CursorIndex = l.totalItems - 1

	if l.totalItems <= l.config.ViewportConfig.Height {
		l.viewport.ViewportStartIndex = 0
		l.viewport.CursorViewportIndex = l.totalItems - 1
	} else {
		l.viewport.ViewportStartIndex = l.totalItems - l.config.ViewportConfig.Height
		l.viewport.CursorViewportIndex = l.config.ViewportConfig.Height - 1
	}

	l.viewport.AtDatasetStart = l.viewport.ViewportStartIndex == 0
	l.viewport.AtDatasetEnd = true

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
		// NOTE: manageBoundingArea() removed - public methods can't handle async commands
	}
}

// JumpToIndex jumps to a specific index in the dataset
func (l *List) JumpToIndex(index int) {
	if l.totalItems == 0 || index < 0 || index >= l.totalItems {
		return
	}

	l.viewport.CursorIndex = index
	l.updateViewportPosition()

	// CRITICAL FIX: Use smart chunk management instead of immediate loading
	// This respects the bounding area configuration
	if cmd := l.smartChunkManagement(); cmd != nil {
		// Execute the chunk management command immediately
		if msg := cmd(); msg != nil {
			// Process any resulting messages
			l.Update(msg)
		}
	}

	// Only update visible items after chunk management
	l.updateVisibleItems()
}

// ToggleCurrentSelection toggles the selection of the current item
func (l *List) ToggleCurrentSelection() bool {
	if l.config.SelectionMode == SelectionNone || l.totalItems == 0 || l.dataSource == nil {
		return false
	}

	item, exists := l.getItemAtIndex(l.viewport.CursorIndex)
	if !exists {
		return false
	}

	// Delegate to DataSource and execute command immediately
	cmd := l.dataSource.SetSelected(l.viewport.CursorIndex, !item.Selected)
	if cmd != nil {
		if msg := cmd(); msg != nil {
			l.Update(msg)
		}
	}
	return true
}

// SelectAll selects all items via DataSource
func (l *List) SelectAll() {
	if l.config.SelectionMode != SelectionMultiple || l.dataSource == nil {
		return
	}

	// Delegate to DataSource and execute command immediately
	cmd := l.dataSource.SelectAll()
	if cmd != nil {
		if msg := cmd(); msg != nil {
			l.Update(msg)
		}
	}
}

// ClearSelection clears all selections
func (l *List) ClearSelection() {
	l.clearSelection()
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

	// Load all chunks that should be in the bounding area
	for chunkStart := boundingArea.ChunkStart; chunkStart < boundingArea.ChunkEnd; chunkStart += chunkSize {
		if chunkStart >= l.totalItems {
			break // Don't load chunks beyond dataset
		}

		if !l.isChunkLoaded(chunkStart) && !l.loadingChunks[chunkStart] {
			// Mark chunk as loading
			l.loadingChunks[chunkStart] = true
			newLoadingChunks = append(newLoadingChunks, chunkStart)

			// Calculate actual chunk size (might be smaller at the end)
			actualChunkSize := chunkSize
			if chunkStart+chunkSize > l.totalItems {
				actualChunkSize = l.totalItems - chunkStart
			}

			request := DataRequest{
				Start:          chunkStart,
				Count:          actualChunkSize,
				SortFields:     l.sortFields,
				SortDirections: l.sortDirs,
				Filters:        l.filters,
			}

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

	// Also unload chunks outside bounding area
	for chunkStart := range l.chunks {
		chunkEnd := chunkStart + chunkSize - 1

		// Check if this chunk intersects with the bounding area
		// A chunk intersects if: chunkStart <= boundingArea.EndIndex AND chunkEnd >= boundingArea.StartIndex
		doesIntersect := chunkStart <= boundingArea.EndIndex && chunkEnd >= boundingArea.StartIndex

		if !doesIntersect {
			// This chunk is completely outside the bounding area, unload it
			delete(l.chunks, chunkStart)
			delete(l.chunkAccessTime, chunkStart)
			cmds = append(cmds, ChunkUnloadedCmd(chunkStart))
		}
	}

	return tea.Batch(cmds...)
}

// isLoadingCriticalChunks checks if we're loading chunks that affect the current viewport
func (l *List) isLoadingCriticalChunks() bool {
	chunkSize := l.config.ViewportConfig.ChunkSize
	viewportStart := l.viewport.ViewportStartIndex
	viewportEnd := viewportStart + l.config.ViewportConfig.Height

	for chunkStart := range l.loadingChunks {
		chunkEnd := chunkStart + chunkSize
		// Check if this loading chunk overlaps with viewport
		if !(chunkEnd <= viewportStart || chunkStart >= viewportEnd) {
			return true
		}
	}
	return false
}

// refreshChunks reloads existing chunks to get updated selection state
func (l *List) refreshChunks() tea.Cmd {
	if l.dataSource == nil {
		return nil
	}

	var cmds []tea.Cmd

	// Reload all currently loaded chunks to get updated selection state
	for chunkStart := range l.chunks {
		// Calculate actual chunk size (might be smaller at the end)
		chunkSize := l.config.ViewportConfig.ChunkSize
		actualChunkSize := chunkSize
		if chunkStart+chunkSize > l.totalItems {
			actualChunkSize = l.totalItems - chunkStart
		}

		request := DataRequest{
			Start:          chunkStart,
			Count:          actualChunkSize,
			SortFields:     l.sortFields,
			SortDirections: l.sortDirs,
			Filters:        l.filters,
		}

		// Reload this chunk to get updated selection state
		cmds = append(cmds, l.dataSource.LoadChunk(request))
	}

	return tea.Batch(cmds...)
}
