// Package list provides a feature-rich, data-virtualized list component for
// Bubble Tea applications. It is designed for performance and flexibility,
// capable of handling very large datasets by loading data in chunks as needed.
// The list supports various item styles, selection modes, configurable keymaps,
// and a component-based rendering pipeline for easy customization.
package list

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/data"
	"github.com/davidroman0O/vtable/render"
	"github.com/davidroman0O/vtable/viewport"
)

// List is a stateful Bubble Tea component that displays a scrollable list of
// items. It manages data fetching, viewport state, user interactions, and
// rendering. The list is highly configurable and uses data virtualization to
// efficiently handle large datasets.
type List struct {
	// Core state
	dataSource core.DataSource[any]    // The source from which the list fetches its data.
	chunks     map[int]core.Chunk[any] // In-memory cache of data chunks, keyed by start index.
	totalItems int                     // The total number of items in the dataset.

	// Viewport state
	viewport core.ViewportState // Manages the visible portion of the list.

	// Configuration
	config core.ListConfig // Holds all configuration for the list's behavior and appearance.

	// Rendering
	formatter         core.ItemFormatter[any]         // A function to custom-render a list item.
	animatedFormatter core.ItemFormatterAnimated[any] // A function to render an item with animation.
	renderContext     core.RenderContext              // Global context for rendering operations.

	// Selection state
	selectedItems map[string]bool // A set of selected item IDs for quick lookups.
	selectedOrder []string        // Maintains the order in which items were selected.

	// Focus state
	focused bool // True if the list is currently handling user input.

	// Animation state
	animationEngine core.AnimationEngine          // The engine that manages animations.
	animations      map[string]core.ListAnimation // A map of active animations for list items.
	lastError       error                         // The last error that occurred, for display purposes.

	// Filtering and sorting
	filters     map[string]any // A map of active filters applied to the data.
	sortFields  []string       // The fields to sort by.
	sortDirs    []string       // The corresponding sort directions.
	searchQuery string         // The current search query.
	searchField string         // The field to search within.

	// Search results
	searchResults []int // A slice of indices that match the current search query.

	// visibleItems is the slice of Data items currently visible in the viewport
	visibleItems []core.Data[any]

	// Chunk access tracking for LRU management
	chunkAccessTime map[int]time.Time // Tracks the last access time for each chunk.

	// Loading state tracking - CRITICAL for UX!
	loadingChunks    map[int]bool // Tracks chunks that are currently being loaded.
	hasLoadingChunks bool         // A quick flag to check if any chunks are loading.
	canScroll        bool         // Whether scrolling is allowed (blocked during critical data loads).
}

// NewList creates a new List component with the given configuration and data
// source. It initializes the list's state, validates the provided configuration,
// and sets up the rendering context. An optional `ItemFormatter` can be provided
// to customize how list items are displayed.
func NewList(listConfig core.ListConfig, dataSource core.DataSource[any], formatter ...core.ItemFormatter[any]) *List {
	// Validate and fix config
	errors := config.ValidateListConfig(&listConfig)
	if len(errors) > 0 {
		config.FixListConfig(&listConfig)
	}

	list := &List{
		dataSource:       dataSource,
		chunks:           make(map[int]core.Chunk[any]),
		config:           listConfig,
		selectedItems:    make(map[string]bool),
		selectedOrder:    make([]string, 0),
		animations:       make(map[string]core.ListAnimation),
		filters:          make(map[string]any),
		chunkAccessTime:  make(map[int]time.Time),
		visibleItems:     make([]core.Data[any], 0), // Initialize visible items
		loadingChunks:    make(map[int]bool),        // Initialize loading state tracking
		hasLoadingChunks: false,
		canScroll:        true, // Allow scrolling initially
		viewport: core.ViewportState{
			ViewportStartIndex:  0,
			CursorIndex:         listConfig.ViewportConfig.InitialIndex,
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

// Init triggers the initial data loading for the list. It is part of the
// bubbletea.Model interface and should be called when the component is first created.
func (l *List) Init() tea.Cmd {
	return l.loadInitialData()
}

// Update is the central message handler for the List component. It processes
// messages for navigation, data loading, selection, and other state changes,
// returning an updated model and any necessary commands. It is the core of the
// component's logic and implements the bubbletea.Model interface.
func (l *List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// ===== Lifecycle Messages =====
	case core.InitMsg:
		return l, l.Init()

	case core.DestroyMsg:
		if l.animationEngine != nil {
			l.animationEngine.Cleanup()
		}
		return l, nil

	case core.ResetMsg:
		l.reset()
		return l, l.Init()

	// ===== Navigation Messages =====
	case core.CursorUpMsg:
		cmd := l.handleCursorUp()
		return l, cmd

	case core.CursorDownMsg:
		cmd := l.handleCursorDown()
		return l, cmd

	case core.PageUpMsg:
		cmd := l.handlePageUp()
		return l, cmd

	case core.PageDownMsg:
		cmd := l.handlePageDown()
		return l, cmd

	case core.JumpToStartMsg:
		cmd := l.handleJumpToStart()
		return l, cmd

	case core.JumpToEndMsg:
		cmd := l.handleJumpToEnd()
		return l, cmd

	case core.JumpToMsg:
		cmd := l.handleJumpTo(msg.Index)
		return l, cmd

	// ===== Data Messages =====
	case core.DataRefreshMsg:
		cmd := l.handleDataRefresh()
		return l, cmd

	case core.DataChunksRefreshMsg:
		// Refresh chunks while preserving cursor position
		l.chunks = make(map[int]core.Chunk[any])
		l.loadingChunks = make(map[int]bool)
		l.hasLoadingChunks = false
		l.canScroll = true
		// Don't reset cursor position - just reload chunks
		return l, l.smartChunkManagement()

	case core.DataChunkLoadedMsg:
		cmd := l.handleDataChunkLoaded(msg)
		return l, cmd

	case core.DataChunkErrorMsg:
		l.lastError = msg.Error
		return l, core.ErrorCmd(msg.Error, "chunk_load")

	case core.DataTotalMsg:
		l.totalItems = msg.Total
		l.updateViewportBounds()
		// Ensure viewport starts at the configured initial position
		l.viewport.ViewportStartIndex = 0
		l.viewport.CursorIndex = l.config.ViewportConfig.InitialIndex
		l.viewport.CursorViewportIndex = l.config.ViewportConfig.InitialIndex
		// After getting total, load the initial chunks using smart chunk management
		return l, l.smartChunkManagement()

	case core.DataTotalUpdateMsg:
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

	case core.DataLoadErrorMsg:
		l.lastError = msg.Error
		return l, core.ErrorCmd(msg.Error, "data_load")

	case core.DataTotalRequestMsg:
		// Handle DataTotalRequestMsg by calling the actual dataSource
		if l.dataSource != nil {
			return l, l.dataSource.GetTotal()
		}
		return l, nil

	case core.DataSourceSetMsg:
		l.dataSource = msg.DataSource
		return l, l.dataSource.GetTotal()

	case core.ChunkUnloadedMsg:
		// Handle chunk unloaded notification (for UI feedback)
		return l, nil

	// ===== Selection Messages =====
	case core.SelectCurrentMsg:
		cmd := l.handleSelectCurrent()
		return l, cmd

	case core.SelectToggleMsg:
		cmd := l.handleSelectToggle(msg.Index)
		return l, cmd

	case core.SelectAllMsg:
		cmd := l.handleSelectAll()
		return l, cmd

	case core.SelectClearMsg:
		if l.dataSource == nil {
			return l, nil
		}
		// Return the command to be processed by Tea model loop
		return l, l.dataSource.ClearSelection()

	case core.SelectRangeMsg:
		cmd := l.handleSelectRange(msg.StartID, msg.EndID)
		return l, cmd

	case core.SelectionModeSetMsg:
		l.config.SelectionMode = msg.Mode
		if msg.Mode == core.SelectionNone {
			l.clearSelection()
		}
		return l, nil

	case core.SelectionResponseMsg:
		// Handle selection response from DataSource
		// The DataSource has updated its internal state, now we need to refresh chunks
		// to get the updated selection state in the Data[T].Selected fields
		cmd := l.refreshChunks()
		return l, cmd

	// ===== Filter Messages =====
	case core.FilterSetMsg:
		l.filters[msg.Field] = msg.Value
		cmd := l.handleFilterChange()
		return l, cmd

	case core.FilterClearMsg:
		delete(l.filters, msg.Field)
		cmd := l.handleFilterChange()
		return l, cmd

	case core.FiltersClearAllMsg:
		l.filters = make(map[string]any)
		cmd := l.handleFilterChange()
		return l, cmd

	// ===== Sort Messages =====
	case core.SortToggleMsg:
		cmd := l.handleSortToggle(msg.Field)
		return l, cmd

	case core.SortSetMsg:
		cmd := l.handleSortSet(msg.Field, msg.Direction)
		return l, cmd

	case core.SortAddMsg:
		cmd := l.handleSortAdd(msg.Field, msg.Direction)
		return l, cmd

	case core.SortRemoveMsg:
		cmd := l.handleSortRemove(msg.Field)
		return l, cmd

	case core.SortsClearAllMsg:
		l.sortFields = nil
		l.sortDirs = nil
		cmd := l.handleFilterChange() // Refresh data
		return l, cmd

	// ===== Focus Messages =====
	case core.FocusMsg:
		l.focused = true
		return l, nil

	case core.BlurMsg:
		l.focused = false
		return l, nil

	// ===== Animation Messages =====
	case core.GlobalAnimationTickMsg:
		if l.animationEngine != nil {
			cmd := l.animationEngine.ProcessGlobalTick(msg)
			return l, cmd
		}
		return l, nil

	case core.AnimationUpdateMsg:
		// Handle animation updates
		return l, nil

	case core.AnimationConfigMsg:
		l.config.AnimationConfig = msg.Config
		if l.animationEngine != nil {
			cmd := l.animationEngine.UpdateConfig(msg.Config)
			return l, cmd
		}
		return l, nil

	case core.ItemAnimationStartMsg:
		l.animations[msg.ItemID] = msg.Animation
		return l, nil

	case core.ItemAnimationStopMsg:
		delete(l.animations, msg.ItemID)
		return l, nil

	// ===== Configuration Messages =====
	case core.FormatterSetMsg:
		l.formatter = msg.Formatter
		return l, nil

	case core.AnimatedFormatterSetMsg:
		l.animatedFormatter = msg.Formatter
		return l, nil

	case core.MaxWidthSetMsg:
		l.config.MaxWidth = msg.Width
		l.setupRenderContext()
		return l, nil

	case core.StyleConfigSetMsg:
		l.config.StyleConfig = msg.Config
		return l, nil

	case core.ViewportConfigMsg:
		l.config.ViewportConfig = msg.Config
		l.updateViewportBounds()
		return l, nil

	case core.KeyMapSetMsg:
		l.config.KeyMap = msg.KeyMap
		return l, nil

	// ===== Search Messages =====
	case core.SearchSetMsg:
		l.searchQuery = msg.Query
		l.searchField = msg.Field
		cmd := l.handleSearch()
		return l, cmd

	case core.SearchClearMsg:
		l.searchQuery = ""
		l.searchField = ""
		l.searchResults = nil
		return l, nil

	case core.SearchResultMsg:
		l.searchResults = msg.Results
		return l, nil

	// ===== Error Messages =====
	case core.ErrorMsg:
		l.lastError = msg.Error
		return l, nil

	// ===== Viewport Messages =====
	case core.ViewportResizeMsg:
		l.config.ViewportConfig.Height = msg.Height
		l.updateViewportBounds()
		return l, nil

	// ===== Batch Messages =====
	case core.BatchMsg:
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

// View renders the list component into a string. It calculates the visible
// items based on the current viewport, formats each item using the configured
// rendering pipeline, and assembles the final output.
func (l *List) View() string {
	var builder strings.Builder

	// Special case for empty dataset
	if l.totalItems == 0 {
		return "No data available"
	}

	// Ensure visible items are up to date
	l.updateVisibleItems()

	// If we have no visible items, render empty or continue
	if len(l.visibleItems) == 0 {
		// Don't show "Loading..." - let chunk loading happen silently
		// The data will appear automatically when chunks load
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
			//  // Use custom formatter if available
			//  ctx := RenderContext{
			//     MaxWidth:          l.config.MaxWidth,
			//     MaxHeight:         1,
			//     ErrorIndicator:    "❌",
			//     LoadingIndicator:  "⏳",
			//     DisabledIndicator: "🚫",
			//     SelectedIndicator: "✅",
			// }
			// Use custom formatter
			renderedItem = l.formatter(
				item,
				absoluteIndex,
				// ctx,
				l.renderContext,
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

// Focus sets the list to a focused state, allowing it to receive and handle
// keyboard inputs.
func (l *List) Focus() tea.Cmd {
	l.focused = true
	return nil
}

// Blur removes focus from the list, preventing it from handling keyboard inputs.
func (l *List) Blur() {
	l.focused = false
}

// IsFocused returns true if the list is currently focused and ready to handle
// user input.
func (l *List) IsFocused() bool {
	return l.focused
}

// GetState returns the current state of the viewport, including cursor position,
// scroll offset, and boundary flags.
func (l *List) GetState() core.ViewportState {
	return l.viewport
}

// GetSelectionCount returns the number of currently selected items.
func (l *List) GetSelectionCount() int {
	return data.GetSelectionCount(l.chunks)
}

// setupRenderContext initializes the render context with values from the list's
// configuration. This context is then passed to rendering functions to ensure
// consistent styling and behavior.
func (l *List) setupRenderContext() {
	l.renderContext = core.RenderContext{
		MaxWidth:       l.config.MaxWidth,
		MaxHeight:      1,   // Single line for list items
		Theme:          nil, // Lists use StyleConfig instead
		BaseStyle:      l.config.StyleConfig.DefaultStyle,
		ColorSupport:   true,
		UnicodeSupport: true,
		CurrentTime:    time.Now(),
		FocusState:     core.FocusState{HasFocus: l.focused},

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

// reset returns the list to its initial state. It clears all cached data,
// selections, and errors, and resets the viewport to its starting position.
func (l *List) reset() {
	l.chunks = make(map[int]core.Chunk[any])
	l.totalItems = 0
	// Selection state is managed by DataSource, not the List
	l.loadingChunks = make(map[int]bool)
	l.hasLoadingChunks = false
	l.canScroll = true
	l.viewport = core.ViewportState{
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

// loadInitialData is the command that starts the data loading process. It
// requests the total number of items from the data source, which in turn
// will trigger the initial chunk loading.
func (l *List) loadInitialData() tea.Cmd {
	if l.dataSource == nil {
		return nil
	}

	// First get the total count
	return l.dataSource.GetTotal()
}

// loadInitialChunk loads the first chunk of data. This is typically called
// after the total number of items is known.
func (l *List) loadInitialChunk() tea.Cmd {
	if l.dataSource == nil {
		return nil
	}

	request := core.DataRequest{
		Start:          0,
		Count:          l.config.ViewportConfig.ChunkSize,
		SortFields:     l.sortFields,
		SortDirections: l.sortDirs,
		Filters:        l.filters,
	}

	return l.dataSource.LoadChunk(request)
}

// handleCursorUp processes a "cursor up" event. It recalculates the viewport
// and cursor positions and then triggers smart chunk management to ensure the
// necessary data is loaded for the new view.
func (l *List) handleCursorUp() tea.Cmd {
	if l.totalItems == 0 || !l.canScroll {
		return nil
	}

	// Can't move up if already at the beginning
	if l.viewport.CursorIndex <= 0 {
		return nil
	}

	previousState := l.viewport
	l.viewport = viewport.CalculateCursorUp(l.viewport, l.config.ViewportConfig, l.totalItems)

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
		return l.smartChunkManagement()
	}

	return nil
}

// handleCursorDown processes a "cursor down" event, adjusting the viewport for
// downward movement and loading new data as needed.
func (l *List) handleCursorDown() tea.Cmd {
	if l.totalItems == 0 || !l.canScroll {
		return nil
	}

	// Can't move down if already at the end
	if l.viewport.CursorIndex >= l.totalItems-1 {
		return nil
	}

	previousState := l.viewport
	l.viewport = viewport.CalculateCursorDown(l.viewport, l.config.ViewportConfig, l.totalItems)

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
		return l.smartChunkManagement()
	}

	return nil
}

// handlePageUp processes a "page up" event, moving the cursor and viewport up
// by one page (equivalent to the viewport height).
func (l *List) handlePageUp() tea.Cmd {
	if l.totalItems == 0 || !l.canScroll {
		return nil
	}

	previousState := l.viewport
	l.viewport = viewport.CalculatePageUp(l.viewport, l.config.ViewportConfig, l.totalItems)

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
	}

	return l.smartChunkManagement()
}

// handlePageDown processes a "page down" event, moving the cursor and viewport
// down by one page.
func (l *List) handlePageDown() tea.Cmd {
	if l.viewport.CursorIndex >= l.totalItems-1 {
		return nil
	}

	previousState := l.viewport
	l.viewport = viewport.CalculatePageDown(l.viewport, l.config.ViewportConfig, l.totalItems)

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
	}

	return l.smartChunkManagement()
}

// handleJumpToStart moves the cursor and viewport to the very beginning of the list.
func (l *List) handleJumpToStart() tea.Cmd {
	if l.totalItems == 0 || !l.canScroll {
		return nil
	}

	l.viewport = viewport.CalculateJumpToStart(l.config.ViewportConfig, l.totalItems)
	return l.smartChunkManagement()
}

// handleJumpToEnd moves the cursor and viewport to the very end of the list.
func (l *List) handleJumpToEnd() tea.Cmd {
	if l.totalItems <= 0 || !l.canScroll {
		return nil
	}

	previousState := l.viewport
	l.viewport = viewport.CalculateJumpToEnd(l.config.ViewportConfig, l.totalItems)

	// Update visible items if viewport changed
	if l.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		l.updateVisibleItems()
		// Use smart chunk management for proper loading feedback
		return l.smartChunkManagement()
	}
	return nil
}

// handleJumpTo moves the cursor and viewport to a specific item index.
func (l *List) handleJumpTo(index int) tea.Cmd {
	if l.totalItems == 0 || index < 0 || index >= l.totalItems || !l.canScroll {
		return nil
	}

	l.viewport = viewport.CalculateJumpTo(index, l.config.ViewportConfig, l.totalItems)
	return l.smartChunkManagement()
}

// handleDataRefresh performs a hard refresh of the list's data. It clears all
// local caches and re-initiates the data loading process.
func (l *List) handleDataRefresh() tea.Cmd {
	l.chunks = make(map[int]core.Chunk[any])

	if l.dataSource == nil {
		return nil
	}

	var cmds []tea.Cmd
	cmds = append(cmds, l.dataSource.GetTotal())
	cmds = append(cmds, l.loadInitialChunk())

	return tea.Batch(cmds...)
}

// handleDataChunkLoaded processes a newly loaded data chunk. It adds the chunk
// to the local cache and updates the loading state.
func (l *List) handleDataChunkLoaded(msg core.DataChunkLoadedMsg) tea.Cmd {
	chunk := core.Chunk[any]{
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
	cmds = append(cmds, core.ChunkLoadingCompletedCmd(msg.StartIndex, len(msg.Items), msg.Request))

	// Unload old chunks
	if unloadCmd := l.unloadOldChunks(); unloadCmd != nil {
		cmds = append(cmds, unloadCmd)
	}

	return tea.Batch(cmds...)
}

// handleSelectCurrent toggles the selection state of the item currently under
// the cursor.
func (l *List) handleSelectCurrent() tea.Cmd {
	if l.config.SelectionMode == core.SelectionNone || l.totalItems == 0 {
		return nil
	}

	item, exists := l.getItemAtIndex(l.viewport.CursorIndex)
	if !exists {
		return nil
	}

	return l.toggleItemSelection(item.ID)
}

// handleSelectToggle toggles the selection state of an item at a specific index.
func (l *List) handleSelectToggle(index int) tea.Cmd {
	if l.config.SelectionMode == core.SelectionNone || index < 0 || index >= l.totalItems {
		return nil
	}

	item, exists := l.getItemAtIndex(index)
	if !exists {
		return nil
	}

	return l.toggleItemSelection(item.ID)
}

// handleSelectAll selects all items in the list via the data source.
func (l *List) handleSelectAll() tea.Cmd {
	if l.config.SelectionMode != core.SelectionMultiple || l.dataSource == nil {
		return nil
	}

	// Return the command to be processed by Tea model loop
	return l.dataSource.SelectAll()
}

// handleSelectRange selects a range of items between a start and end ID.
func (l *List) handleSelectRange(startID, endID string) tea.Cmd {
	if l.config.SelectionMode != core.SelectionMultiple {
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

// handleFilterChange triggers a data refresh when filters change.
func (l *List) handleFilterChange() tea.Cmd {
	return l.handleDataRefresh()
}

// handleSortToggle toggles the sort order for a given field.
func (l *List) handleSortToggle(field string) tea.Cmd {
	currentSort := data.SortState{
		Fields:     l.sortFields,
		Directions: l.sortDirs,
	}

	newSort := data.ToggleSortField(currentSort, field)
	l.sortFields = newSort.Fields
	l.sortDirs = newSort.Directions

	return l.handleDataRefresh()
}

// handleSortSet sets the sort to a single field and direction.
func (l *List) handleSortSet(field, direction string) tea.Cmd {
	newSort := data.SetSortField(field, direction)
	l.sortFields = newSort.Fields
	l.sortDirs = newSort.Directions
	return l.handleDataRefresh()
}

// handleSortAdd adds a new field to the multi-level sort configuration.
func (l *List) handleSortAdd(field, direction string) tea.Cmd {
	currentSort := data.SortState{
		Fields:     l.sortFields,
		Directions: l.sortDirs,
	}

	newSort := data.AddSortField(currentSort, field, direction)
	l.sortFields = newSort.Fields
	l.sortDirs = newSort.Directions

	return l.handleDataRefresh()
}

// handleSortRemove removes a field from the sort configuration.
func (l *List) handleSortRemove(field string) tea.Cmd {
	currentSort := data.SortState{
		Fields:     l.sortFields,
		Directions: l.sortDirs,
	}

	newSort := data.RemoveSortField(currentSort, field)
	l.sortFields = newSort.Fields
	l.sortDirs = newSort.Directions

	return l.handleDataRefresh()
}

// handleSearch performs a search, which typically involves refreshing the data
// with search parameters.
func (l *List) handleSearch() tea.Cmd {
	if l.dataSource == nil {
		return nil
	}

	// For now, return a simple search command
	// Real implementation would depend on DataSource capabilities
	return core.SearchResultCmd([]int{}, l.searchQuery, 0)
}

// handleKeyPress processes raw key presses, mapping them to list actions based
// on the current keymap.
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
			return core.SelectCurrentCmd()
		}
	}

	for _, selectAllKey := range l.config.KeyMap.SelectAll {
		if key == selectAllKey {
			return core.SelectAllCmd()
		}
	}

	for _, filterKey := range l.config.KeyMap.Filter {
		if key == filterKey {
			// Return command to start filtering
			return core.StatusCmd("Filter mode", core.StatusInfo)
		}
	}

	for _, sortKey := range l.config.KeyMap.Sort {
		if key == sortKey {
			// Return command to start sorting
			return core.StatusCmd("Sort mode", core.StatusInfo)
		}
	}

	return nil
}

// renderEmpty renders the view shown when the list has no items.
func (l *List) renderEmpty() string {
	return render.RenderEmptyState(l.config.StyleConfig, l.lastError)
}

// renderItem renders a single item at a given index.
func (l *List) renderItem(absoluteIndex, viewportIndex int) string {
	item, exists := l.getItemAtIndex(absoluteIndex)
	if !exists {
		return render.RenderLoadingPlaceholder(l.config.StyleConfig)
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

// applyItemStyle is a wrapper around the render package's ApplyItemStyle.
func (l *List) applyItemStyle(content string, isCursor, isSelected bool, item core.Data[any]) string {
	return render.ApplyItemStyle(content, isCursor, isSelected, item, l.config.StyleConfig, l.config.MaxWidth, l.renderContext.Truncate)
}

// updateViewportPosition ensures the viewport is correctly positioned to keep
// the cursor visible.
func (l *List) updateViewportPosition() {
	l.viewport = viewport.UpdateViewportPosition(l.viewport, l.config.ViewportConfig, l.totalItems)
}

// updateViewportBounds recalculates the boundary flags of the viewport.
func (l *List) updateViewportBounds() {
	l.viewport = viewport.UpdateViewportBounds(l.viewport, l.config.ViewportConfig, l.totalItems)
}

// calculateBoundingArea determines the range of data that should be pre-fetched
// around the current viewport.
func (l *List) calculateBoundingArea() core.BoundingArea {
	return viewport.CalculateBoundingArea(l.viewport, l.config.ViewportConfig, l.totalItems)
}

// unloadChunksOutsideBoundingArea unloads data chunks that are no longer within
// the active bounding area.
func (l *List) unloadChunksOutsideBoundingArea() tea.Cmd {
	boundingArea := l.calculateBoundingArea()
	chunkSize := l.config.ViewportConfig.ChunkSize
	var cmds []tea.Cmd

	// Find and unload chunks outside the bounding area
	chunksToUnload := data.FindChunksToUnload(l.chunks, boundingArea, chunkSize)
	for _, chunkStart := range chunksToUnload {
		delete(l.chunks, chunkStart)
		delete(l.chunkAccessTime, chunkStart)
		cmds = append(cmds, core.ChunkUnloadedCmd(chunkStart))
	}

	return tea.Batch(cmds...)
}

// isChunkLoaded checks if the data chunk for a given item index is in memory.
func (l *List) isChunkLoaded(index int) bool {
	return data.IsChunkLoaded(index, l.chunks)
}

// getItemAtIndex retrieves an item by its absolute index from cached chunks.
func (l *List) getItemAtIndex(index int) (core.Data[any], bool) {
	return data.GetItemAtIndex(index, l.chunks, l.totalItems, l.chunkAccessTime)
}

// findItemIndex searches for an item by its ID across all loaded chunks.
func (l *List) findItemIndex(id string) int {
	return data.FindItemIndex(id, l.chunks)
}

// toggleItemSelection changes the selection state of an item via the data source.
func (l *List) toggleItemSelection(id string) tea.Cmd {
	if l.config.SelectionMode == core.SelectionNone || l.dataSource == nil {
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

// clearSelection deselects all currently selected items via the data source.
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

// unloadOldChunks implements a cache eviction policy to remove chunks that are
// no longer needed.
func (l *List) unloadOldChunks() tea.Cmd {
	// Calculate the bounds of chunks that should be kept
	keepLowerBound, keepUpperBound := data.CalculateUnloadBounds(l.viewport, l.config.ViewportConfig)

	var unloadedChunks []int

	// Unload chunks outside the bounds
	for startIndex := range l.chunks {
		if data.ShouldUnloadChunk(startIndex, keepLowerBound, keepUpperBound) {
			delete(l.chunks, startIndex)
			delete(l.chunkAccessTime, startIndex)
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

// updateVisibleItems recalculates the `visibleItems` slice based on the
// current viewport and loaded chunks.
func (l *List) updateVisibleItems() {
	result := viewport.CalculateVisibleItemsFromChunks(
		l.viewport,
		l.config.ViewportConfig,
		l.totalItems,
		l.chunks,
		l.ensureChunkLoadedImmediate,
	)

	l.visibleItems = result.Items
	l.viewport = result.AdjustedViewport
}

// ensureChunkLoadedImmediate is a helper to request a chunk if it's not loaded,
// used to fill in missing data for the current view.
func (l *List) ensureChunkLoadedImmediate(index int) {
	chunkStartIndex := data.CalculateChunkStartIndex(index, l.config.ViewportConfig.ChunkSize)
	if _, exists := l.chunks[chunkStartIndex]; !exists {
		// Load this chunk immediately - NO WAITING!
		if l.dataSource != nil {
			request := data.CreateChunkRequest(
				chunkStartIndex,
				l.config.ViewportConfig.ChunkSize,
				l.totalItems,
				l.sortFields,
				l.sortDirs,
				l.filters,
			)

			// Check if the data source supports immediate loading
			if immediateLoader, ok := l.dataSource.(interface {
				LoadChunkImmediate(core.DataRequest) core.DataChunkLoadedMsg
			}); ok {
				// Use immediate loading - FULLY AUTOMATED!
				chunkMsg := immediateLoader.LoadChunkImmediate(request)
				l.handleDataChunkLoaded(chunkMsg)
			} else {
				// Fallback to async loading (not ideal but better than nothing)
				loadCmd := l.dataSource.LoadChunk(request)
				if loadCmd != nil {
					if msg := loadCmd(); msg != nil {
						if chunkMsg, ok := msg.(core.DataChunkLoadedMsg); ok {
							l.handleDataChunkLoaded(chunkMsg)
						}
					}
				}
			}
		}
	}
}

// smartChunkManagement is the core logic for data virtualization. It determines
// which chunks to load and unload based on viewport position.
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
	chunksToLoad := data.CalculateChunksInBoundingArea(boundingArea, chunkSize, l.totalItems)

	// Load chunks that aren't already loaded or loading
	for _, chunkStart := range chunksToLoad {
		if !l.isChunkLoaded(chunkStart) && !l.loadingChunks[chunkStart] {
			// Mark chunk as loading
			l.loadingChunks[chunkStart] = true
			newLoadingChunks = append(newLoadingChunks, chunkStart)

			request := data.CreateChunkRequest(
				chunkStart,
				chunkSize,
				l.totalItems,
				l.sortFields,
				l.sortDirs,
				l.filters,
			)

			// Emit chunk loading started message for observability
			cmds = append(cmds, core.ChunkLoadingStartedCmd(chunkStart, request))
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
	chunksToUnload := data.FindChunksToUnload(l.chunks, boundingArea, chunkSize)
	for _, chunkStart := range chunksToUnload {
		delete(l.chunks, chunkStart)
		delete(l.chunkAccessTime, chunkStart)
		cmds = append(cmds, core.ChunkUnloadedCmd(chunkStart))
	}

	return tea.Batch(cmds...)
}

// isLoadingCriticalChunks checks if any chunks currently being loaded are
// within the visible viewport.
func (l *List) isLoadingCriticalChunks() bool {
	return data.IsLoadingCriticalChunks(l.viewport, l.config.ViewportConfig, l.loadingChunks)
}

// refreshChunks forces a reload of all currently loaded data chunks.
func (l *List) refreshChunks() tea.Cmd {
	if l.dataSource == nil {
		return nil
	}

	var cmds []tea.Cmd

	// Reload all currently loaded chunks to get updated selection state
	for chunkStart := range l.chunks {
		request := data.CreateChunkRequest(
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

// SetEnumerator sets a custom function to generate the enumerator for each item.
func (l *List) SetEnumerator(enum core.ListEnumerator) {
	l.config.RenderConfig.EnumeratorConfig.Enumerator = enum
}

// SetBulletStyle configures the list to use a standard bullet point enumerator.
func (l *List) SetBulletStyle() {
	l.config.RenderConfig.EnumeratorConfig.Enumerator = BulletEnumerator
}

// SetNumberedStyle configures the list to use a numbered enumerator.
func (l *List) SetNumberedStyle() {
	l.config.RenderConfig.EnumeratorConfig.Enumerator = ArabicEnumerator
	l.config.RenderConfig.EnumeratorConfig.Alignment = core.ListAlignmentRight
}

// SetChecklistStyle configures the list to use a checkbox enumerator.
func (l *List) SetChecklistStyle() {
	l.config.RenderConfig.EnumeratorConfig.Enumerator = CheckboxEnumerator
}

// SetAlphabeticalStyle configures the list to use an alphabetical enumerator.
func (l *List) SetAlphabeticalStyle() {
	l.config.RenderConfig.EnumeratorConfig.Enumerator = AlphabetEnumerator
	l.config.RenderConfig.EnumeratorConfig.Alignment = core.ListAlignmentRight
}

// SetDashStyle configures the list to use a dash enumerator.
func (l *List) SetDashStyle() {
	l.config.RenderConfig.EnumeratorConfig.Enumerator = DashEnumerator
}

// SetConditionalStyle sets up a conditional enumerator, allowing different
// enumerators based on item state.
func (l *List) SetConditionalStyle() {
	conditionalEnum := NewConditionalEnumerator(BulletEnumerator).
		When(IsSelected, CheckboxEnumerator).
		When(IsError, func(item core.Data[any], index int, ctx core.RenderContext) string {
			return "✗ "
		}).
		When(IsLoading, func(item core.Data[any], index int, ctx core.RenderContext) string {
			return "⟳ "
		})

	l.config.RenderConfig.EnumeratorConfig.Enumerator = conditionalEnum.Enumerate
}

// SetCustomEnumerator sets an enumerator based on a string pattern.
func (l *List) SetCustomEnumerator(pattern string) {
	l.config.RenderConfig.EnumeratorConfig.Enumerator = CustomEnumerator(pattern)
}

// SetRenderConfig applies a completely new rendering configuration.
func (l *List) SetRenderConfig(config core.ListRenderConfig) {
	l.config.RenderConfig = config
}

// GetRenderConfig returns the current rendering configuration.
func (l *List) GetRenderConfig() core.ListRenderConfig {
	return l.config.RenderConfig
}

// SetEnumeratorAlignment enables or disables right-alignment for the enumerator.
func (l *List) SetEnumeratorAlignment(align bool) {
	if align {
		l.config.RenderConfig.EnumeratorConfig.Alignment = core.ListAlignmentRight
		l.config.RenderConfig.EnumeratorConfig.MaxWidth = 4
	} else {
		l.config.RenderConfig.EnumeratorConfig.Alignment = core.ListAlignmentNone
		l.config.RenderConfig.EnumeratorConfig.MaxWidth = 0
	}
}

// SetTextWrapping enables or disables text wrapping for item content.
func (l *List) SetTextWrapping(wrap bool) {
	l.config.RenderConfig.ContentConfig.WrapText = wrap
}

// SetIndentSize sets the indentation size for multi-line content.
func (l *List) SetIndentSize(size int) {
	// In the new system, indent size is handled automatically by the content component
	// based on the width of preceding components, but we can set max width
	if size > 0 {
		l.config.RenderConfig.ContentConfig.MaxWidth = 80 - size
	}
}

// SetFormatter sets a custom function for rendering an item's content.
func (l *List) SetFormatter(formatter core.ItemFormatter[any]) core.ItemFormatter[any] {
	previous := l.formatter
	l.formatter = formatter
	return previous
}

// GetFormatter returns the currently configured item formatter.
func (l *List) GetFormatter() core.ItemFormatter[any] {
	return l.formatter
}

// SetErrorIndicator sets the string used to indicate an error state.
func (l *List) SetErrorIndicator(indicator string) {
	l.renderContext.ErrorIndicator = indicator
}

// SetLoadingIndicator sets the string used to indicate a loading state.
func (l *List) SetLoadingIndicator(indicator string) {
	l.renderContext.LoadingIndicator = indicator
}

// SetDisabledIndicator sets the string used to indicate a disabled item.
func (l *List) SetDisabledIndicator(indicator string) {
	l.renderContext.DisabledIndicator = indicator
}

// SetSelectedIndicator sets the string used to indicate a selected item.
func (l *List) SetSelectedIndicator(indicator string) {
	l.renderContext.SelectedIndicator = indicator
}

// GetErrorIndicator returns the current error indicator string.
func (l *List) GetErrorIndicator() string {
	return l.renderContext.ErrorIndicator
}

// GetLoadingIndicator returns the current loading indicator string.
func (l *List) GetLoadingIndicator() string {
	return l.renderContext.LoadingIndicator
}

// GetDisabledIndicator returns the current disabled indicator string.
func (l *List) GetDisabledIndicator() string {
	return l.renderContext.DisabledIndicator
}

// GetSelectedIndicator returns the current selected indicator string.
func (l *List) GetSelectedIndicator() string {
	return l.renderContext.SelectedIndicator
}
