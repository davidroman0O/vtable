package tree

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/data"
	"github.com/davidroman0O/vtable/viewport"
)

// ================================
// TREE DATA STRUCTURES
// ================================

// TreeData represents a hierarchical data item
type TreeData[T any] struct {
	ID       string
	Item     T
	Children []TreeData[T]
	Expanded bool
}

// TreeDataSource provides hierarchical data
type TreeDataSource[T any] interface {
	// Get the root nodes of the tree
	GetRootNodes() []TreeData[T]

	// Standard data source operations for individual items
	GetItemByID(id string) (TreeData[T], bool)
	SetSelected(id string, selected bool) tea.Cmd
	SetSelectedByID(id string, selected bool) tea.Cmd
	SelectAll() tea.Cmd
	ClearSelection() tea.Cmd
	SelectRange(startID, endID string) tea.Cmd
}

// ================================
// INTERNAL FLAT DATA SOURCE
// ================================

// ================================
// FLATTENED TREE ITEM
// ================================

// FlatTreeItem represents a tree item in flattened form
type FlatTreeItem[T any] struct {
	ID            string
	Item          T
	Depth         int
	HasChildNodes bool // Renamed to avoid conflict
	Expanded      bool
	ParentID      string
}

// GetDepth returns the depth of this tree item
func (f FlatTreeItem[T]) GetDepth() int {
	return f.Depth
}

// HasChildren returns whether this item has children
func (f FlatTreeItem[T]) HasChildren() bool {
	return f.HasChildNodes
}

// IsExpanded returns whether this item is expanded
func (f FlatTreeItem[T]) IsExpanded() bool {
	return f.Expanded
}

// ================================
// TREE LIST COMPONENT
// ================================

// TreeList is an independent component that handles hierarchical data
type TreeList[T any] struct {
	// Core state - same as List but for tree data
	treeDataSource TreeDataSource[T]
	chunks         map[int]core.Chunk[any] // Reuse same chunk system
	totalItems     int

	// Viewport state - same as List
	viewport core.ViewportState

	// Configuration - reuse List config
	config core.ListConfig

	// Tree-specific state
	rootNodes     []TreeData[T]
	expandedNodes map[string]bool
	selectedNodes map[string]bool
	flattenedView []FlatTreeItem[T] // Cached flattened view

	// Rendering - ENHANCED with component system
	formatter         core.ItemFormatter[any]
	animatedFormatter core.ItemFormatterAnimated[any]
	renderContext     core.RenderContext

	// Focus state
	focused bool

	// Tree configuration - SIMPLIFIED (component system handles most rendering)
	treeConfig TreeConfig

	// Chunk management - same as List
	visibleItems     []core.Data[any]
	chunkAccessTime  map[int]time.Time
	loadingChunks    map[int]bool
	hasLoadingChunks bool
	canScroll        bool

	// Error handling
	lastError error
}

// TreeConfig contains tree-specific configuration
type TreeConfig struct {
	// Component-based rendering configuration
	RenderConfig TreeRenderConfig // Use tree-specific component system

	// Tree-specific behavior
	CascadingSelection bool // When true, selecting a parent also selects all children
	AutoExpand         bool // When true, automatically expand nodes when navigating to them
	ShowRoot           bool // When true, show root nodes with special styling

	// Tree navigation behavior
	ExpandOnSelect bool // When true, selecting a node also expands it

	// Tree symbols (legacy - component system handles these now)
	Enumerator tree.Enumerator
	Indenter   tree.Indenter

	// Styling (legacy - component system handles these now)
	RootStyle       lipgloss.Style
	ItemStyle       lipgloss.Style
	EnumeratorStyle lipgloss.Style

	// Cursor and styling configuration (legacy - component system handles these now)
	CursorIndicator       string
	CursorSpacing         string
	NormalSpacing         string
	ShowCursor            bool
	EnableCursorStyling   bool
	CursorBackgroundStyle lipgloss.Style
}

// DefaultTreeConfig returns sensible defaults for tree configuration
func DefaultTreeConfig() TreeConfig {
	config := DefaultTreeRenderConfig()

	return TreeConfig{
		RenderConfig:          config,
		CascadingSelection:    true,
		AutoExpand:            true,
		ShowRoot:              true,
		ExpandOnSelect:        true,
		Enumerator:            tree.DefaultEnumerator,
		Indenter:              tree.DefaultIndenter,
		RootStyle:             lipgloss.NewStyle(),
		ItemStyle:             lipgloss.NewStyle(),
		EnumeratorStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		CursorIndicator:       "► ",
		CursorSpacing:         "  ",
		NormalSpacing:         "  ",
		ShowCursor:            true,
		EnableCursorStyling:   true,
		CursorBackgroundStyle: lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("15")),
	}
}

// NewTreeList creates a new TreeList component
func NewTreeList[T any](listConfig core.ListConfig, treeConfig TreeConfig, dataSource TreeDataSource[T]) *TreeList[T] {
	// Validate and fix config - reuse List validation
	errors := config.ValidateListConfig(&listConfig)
	if len(errors) > 0 {
		config.FixListConfig(&listConfig)
	}

	treeList := &TreeList[T]{
		treeDataSource:   dataSource,
		chunks:           make(map[int]core.Chunk[any]),
		config:           listConfig,
		rootNodes:        dataSource.GetRootNodes(),
		expandedNodes:    make(map[string]bool),
		selectedNodes:    make(map[string]bool),
		treeConfig:       treeConfig,
		chunkAccessTime:  make(map[int]time.Time),
		visibleItems:     make([]core.Data[any], 0),
		loadingChunks:    make(map[int]bool),
		hasLoadingChunks: false,
		canScroll:        true,
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

	// Set up render context - reuse List setup
	treeList.setupRenderContext()

	// Initialize flattened view
	treeList.updateFlattenedView()

	return treeList
}

// ================================
// TEA MODEL INTERFACE - Same as List
// ================================

// Init initializes the tree list model
func (tl *TreeList[T]) Init() tea.Cmd {
	return tl.loadInitialData()
}

// Update handles all messages - reuse List message handling patterns
func (tl *TreeList[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// ===== Navigation Messages - Same as List =====
	case core.CursorUpMsg:
		cmd := tl.handleCursorUp()
		return tl, cmd

	case core.CursorDownMsg:
		cmd := tl.handleCursorDown()
		return tl, cmd

	case core.PageUpMsg:
		cmd := tl.handlePageUp()
		return tl, cmd

	case core.PageDownMsg:
		cmd := tl.handlePageDown()
		return tl, cmd

	case core.JumpToStartMsg:
		cmd := tl.handleJumpToStart()
		return tl, cmd

	case core.JumpToEndMsg:
		cmd := tl.handleJumpToEnd()
		return tl, cmd

	case core.JumpToMsg:
		cmd := tl.handleJumpTo(msg.Index)
		return tl, cmd

	case core.TreeJumpToIndexMsg:
		cmd := tl.handleTreeJumpToIndex(msg.Index, msg.ExpandParents)
		return tl, cmd

	// ===== Data Messages - Same as List =====
	case core.DataRefreshMsg:
		cmd := tl.handleDataRefresh()
		return tl, cmd

	case core.DataChunksRefreshMsg:
		// Refresh chunks while preserving cursor position
		tl.chunks = make(map[int]core.Chunk[any])
		tl.loadingChunks = make(map[int]bool)
		tl.hasLoadingChunks = false
		tl.canScroll = true
		return tl, tl.smartChunkManagement()

	case core.DataChunkLoadedMsg:
		cmd := tl.handleDataChunkLoaded(msg)
		return tl, cmd

	case core.DataTotalMsg:
		tl.totalItems = msg.Total
		tl.updateViewportBounds()
		// Reset viewport for initial load
		tl.viewport.ViewportStartIndex = 0
		tl.viewport.CursorIndex = tl.config.ViewportConfig.InitialIndex
		tl.viewport.CursorViewportIndex = tl.config.ViewportConfig.InitialIndex
		return tl, tl.smartChunkManagement()

	case core.DataTotalUpdateMsg:
		// Update total while preserving cursor position
		oldTotal := tl.totalItems
		tl.totalItems = msg.Total
		tl.updateViewportBounds()

		// Ensure cursor stays within bounds
		if tl.viewport.CursorIndex >= tl.totalItems && tl.totalItems > 0 {
			tl.viewport.CursorIndex = tl.totalItems - 1
			tl.viewport.CursorViewportIndex = tl.viewport.CursorIndex - tl.viewport.ViewportStartIndex
			if tl.viewport.CursorViewportIndex < 0 {
				tl.viewport.ViewportStartIndex = tl.viewport.CursorIndex
				tl.viewport.CursorViewportIndex = 0
			}
		}

		if oldTotal != tl.totalItems {
			return tl, tl.smartChunkManagement()
		}
		return tl, nil

	// ===== Selection Messages - Same as List =====
	case core.SelectCurrentMsg:
		cmd := tl.handleSelectCurrent()
		return tl, cmd

	case core.SelectAllMsg:
		cmd := tl.handleSelectAll()
		return tl, cmd

	case core.SelectClearMsg:
		cmd := tl.handleSelectClear()
		return tl, cmd

	case core.SelectionResponseMsg:
		// Handle selection response - refresh chunks to get updated selection state
		cmd := tl.refreshChunks()
		return tl, cmd

	// ===== Focus Messages - Same as List =====
	case core.FocusMsg:
		tl.focused = true
		return tl, nil

	case core.BlurMsg:
		tl.focused = false
		return tl, nil

	// ===== Keyboard Input - Same as List =====
	case tea.KeyMsg:
		cmd := tl.handleKeyPress(msg)
		return tl, cmd
	}

	return tl, nil
}

// View renders the tree list - similar to List but with tree formatting
func (tl *TreeList[T]) View() string {
	var builder strings.Builder

	// Special case for empty dataset
	if tl.totalItems == 0 {
		return "No data available"
	}

	// Ensure visible items are up to date
	tl.updateVisibleItems()

	// If we have no visible items, render empty or continue
	if len(tl.visibleItems) == 0 {
		// Don't show "Loading..." - let chunk loading happen silently
		// The data will appear automatically when chunks load
	}

	// Render each visible item using component-based system
	for i, item := range tl.visibleItems {
		absoluteIndex := tl.viewport.ViewportStartIndex + i

		if absoluteIndex >= tl.totalItems {
			break
		}

		isCursor := i == tl.viewport.CursorViewportIndex

		var renderedItem string

		if tl.formatter != nil {
			// Use custom formatter
			renderedItem = tl.formatter(
				item,
				absoluteIndex,
				tl.renderContext,
				isCursor,
				tl.viewport.IsAtTopThreshold,
				tl.viewport.IsAtBottomThreshold,
			)
		} else {
			// Use component-based tree rendering system for full customization
			enhancedFormatter := EnhancedTreeFormatter(tl.treeConfig.RenderConfig)
			ctx := tl.renderContext
			ctx.MaxWidth = tl.treeConfig.RenderConfig.ContentConfig.MaxWidth

			// Extract tree-specific data from the flattened item
			flatItem, ok := item.Item.(FlatTreeItem[T])
			if !ok {
				// Fallback to simple tree formatter if type assertion fails
				renderedItem = tl.formatTreeItem(
					item,
					absoluteIndex,
					tl.renderContext,
					isCursor,
					tl.viewport.IsAtTopThreshold,
					tl.viewport.IsAtBottomThreshold,
				)
			} else {
				// Use component-based rendering with tree-specific data
				renderedItem = enhancedFormatter(
					item,
					absoluteIndex,
					flatItem.Depth,
					flatItem.HasChildren(),
					flatItem.IsExpanded(),
					ctx,
					isCursor,
					tl.viewport.IsAtTopThreshold,
					tl.viewport.IsAtBottomThreshold,
				)
			}
		}

		builder.WriteString(renderedItem)

		if i < len(tl.visibleItems)-1 && absoluteIndex < tl.totalItems-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// ================================
// TREE FORMATTING
// ================================

// formatTreeItem formats a tree item with proper tree structure
func (tl *TreeList[T]) formatTreeItem(
	item core.Data[any],
	index int,
	ctx core.RenderContext,
	isCursor, isTopThreshold, isBottomThreshold bool,
) string {
	// Type assert to FlatTreeItem
	flatItem, ok := item.Item.(FlatTreeItem[T])
	if !ok {
		return fmt.Sprintf("%s Invalid tree item: %v", ctx.ErrorIndicator, item.Item)
	}

	var prefix strings.Builder

	// Add cursor indicator or spacing
	if tl.treeConfig.ShowCursor && isCursor {
		prefix.WriteString(tl.treeConfig.CursorIndicator)
	} else {
		prefix.WriteString(tl.treeConfig.NormalSpacing)
	}

	// Add indentation based on depth
	for i := 0; i < flatItem.Depth; i++ {
		prefix.WriteString("  ")
	}

	// Add tree connector
	if flatItem.HasChildNodes {
		if flatItem.Expanded {
			prefix.WriteString("▼ ")
		} else {
			prefix.WriteString("▶ ")
		}
	} else {
		prefix.WriteString("• ")
	}

	// Format the item content
	content := tl.formatItemContent(flatItem.Item)

	// Build the complete line content
	fullContent := prefix.String() + content

	// Apply cursor styling if enabled and this is the cursor line
	if isCursor && tl.treeConfig.EnableCursorStyling {
		if !tl.treeConfig.ShowCursor {
			// No cursor indicator - apply background style to entire line
			fullContent = tl.treeConfig.CursorBackgroundStyle.Render(fullContent)
		} else {
			// Has cursor indicator - apply background style to content part only
			styledContent := tl.treeConfig.CursorBackgroundStyle.Render(content)
			// Rebuild with styled content, preserving the prefix structure
			prefixWithoutIndicator := prefix.String()[len(tl.treeConfig.CursorIndicator):]
			fullContent = tl.treeConfig.CursorIndicator + prefixWithoutIndicator + styledContent
		}
	} else if item.Selected && !isCursor {
		// Apply selection styling only if not cursor (cursor styling takes precedence)
		styledContent := lipgloss.NewStyle().
			Background(lipgloss.Color("240")).
			Foreground(lipgloss.Color("15")).
			Render(content)
		fullContent = prefix.String() + styledContent
	}

	// Add selection indicator if selected
	if item.Selected {
		fullContent += " " + ctx.SelectedIndicator
	}

	return fullContent
}

// formatItemContent formats the content of a tree item
func (tl *TreeList[T]) formatItemContent(item T) string {
	if stringer, ok := any(item).(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%v", item)
}

// ================================
// TREE OPERATIONS
// ================================

// ExpandNode expands a tree node
func (tl *TreeList[T]) ExpandNode(id string) tea.Cmd {
	tl.expandedNodes[id] = true
	tl.updateFlattenedView()
	// Update total and refresh chunks
	return tea.Batch(
		core.DataTotalUpdateCmd(len(tl.flattenedView)),
		core.DataChunksRefreshCmd(),
	)
}

// CollapseNode collapses a tree node
func (tl *TreeList[T]) CollapseNode(id string) tea.Cmd {
	delete(tl.expandedNodes, id)
	tl.updateFlattenedView()
	// Update total and refresh chunks
	return tea.Batch(
		core.DataTotalUpdateCmd(len(tl.flattenedView)),
		core.DataChunksRefreshCmd(),
	)
}

// ToggleNode toggles a tree node's expanded state
func (tl *TreeList[T]) ToggleNode(id string) tea.Cmd {
	if tl.expandedNodes[id] {
		return tl.CollapseNode(id)
	}
	return tl.ExpandNode(id)
}

// ToggleCurrentNode toggles the currently selected node
func (tl *TreeList[T]) ToggleCurrentNode() tea.Cmd {
	if tl.viewport.CursorIndex >= 0 && tl.viewport.CursorIndex < len(tl.flattenedView) {
		currentItem := tl.flattenedView[tl.viewport.CursorIndex]
		if currentItem.HasChildren() {
			return tl.ToggleNode(currentItem.ID)
		}
	}
	return nil
}

// createFullyExpandedView creates a flattened view as if all nodes were expanded
func (tl *TreeList[T]) createFullyExpandedView() []FlatTreeItem[T] {
	var fullyExpandedView []FlatTreeItem[T]
	tl.flattenNodesFullyExpanded(tl.rootNodes, "", 0, &fullyExpandedView)
	return fullyExpandedView
}

// flattenNodesFullyExpanded recursively flattens tree nodes with all nodes expanded
func (tl *TreeList[T]) flattenNodesFullyExpanded(nodes []TreeData[T], parentID string, depth int, result *[]FlatTreeItem[T]) {
	for _, node := range nodes {
		// Add the node itself
		*result = append(*result, FlatTreeItem[T]{
			ID:            node.ID,
			Item:          node.Item,
			Depth:         depth,
			HasChildNodes: len(node.Children) > 0,
			Expanded:      true, // Always expanded in this view
			ParentID:      parentID,
		})

		// Always add children (fully expanded)
		if len(node.Children) > 0 {
			tl.flattenNodesFullyExpanded(node.Children, node.ID, depth+1, result)
		}
	}
}

// findPathToItem finds the path of parent IDs to reach a specific item
func (tl *TreeList[T]) findPathToItem(targetID string, nodes []TreeData[T], currentPath []string) []string {
	for _, node := range nodes {
		if node.ID == targetID {
			// Found the target item, return the current path (which doesn't include the target itself)
			return currentPath
		}

		if len(node.Children) > 0 {
			// Search in children with this node added to the path
			newPath := append(currentPath, node.ID)
			if result := tl.findPathToItem(targetID, node.Children, newPath); result != nil {
				return result
			}
		}
	}

	return nil // Not found
}

// findItemIndexInFlattenedView finds the index of an item in the current flattened view
func (tl *TreeList[T]) findItemIndexInFlattenedView(itemID string) int {
	for i, item := range tl.flattenedView {
		if item.ID == itemID {
			return i
		}
	}
	return -1 // Not found
}

// ================================
// TREE FLATTENING - TreeList specific
// ================================

// updateFlattenedView updates the cached flattened view
func (tl *TreeList[T]) updateFlattenedView() {
	tl.flattenedView = nil
	tl.flattenNodes(tl.rootNodes, "", 0)
	tl.totalItems = len(tl.flattenedView)
}

// flattenNodes recursively flattens tree nodes
func (tl *TreeList[T]) flattenNodes(nodes []TreeData[T], parentID string, depth int) {
	for _, node := range nodes {
		// Add the node itself
		tl.flattenedView = append(tl.flattenedView, FlatTreeItem[T]{
			ID:            node.ID,
			Item:          node.Item,
			Depth:         depth,
			HasChildNodes: len(node.Children) > 0,
			Expanded:      tl.expandedNodes[node.ID],
			ParentID:      parentID,
		})

		// Add children if expanded
		if tl.expandedNodes[node.ID] && len(node.Children) > 0 {
			tl.flattenNodes(node.Children, node.ID, depth+1)
		}
	}
}

// ================================
// CHUNK MANAGEMENT - Reuse List functions
// ================================

// loadInitialData loads the total count and initial chunk
func (tl *TreeList[T]) loadInitialData() tea.Cmd {
	// Set initial total
	return core.DataTotalCmd(len(tl.flattenedView))
}

// smartChunkManagement - reuse List logic
func (tl *TreeList[T]) smartChunkManagement() tea.Cmd {
	// Calculate bounding area - reuse List function
	boundingArea := viewport.CalculateBoundingArea(tl.viewport, tl.config.ViewportConfig, tl.totalItems)
	chunkSize := tl.config.ViewportConfig.ChunkSize
	var cmds []tea.Cmd

	// Get chunks that need to be loaded
	chunksToLoad := data.CalculateChunksInBoundingArea(boundingArea, chunkSize, tl.totalItems)

	// Load chunks that aren't already loaded
	for _, chunkStart := range chunksToLoad {
		if !data.IsChunkLoaded(chunkStart, tl.chunks) && !tl.loadingChunks[chunkStart] {
			tl.loadingChunks[chunkStart] = true

			// Create chunk data from flattened view
			cmd := tl.loadChunkFromFlattenedView(chunkStart, chunkSize)
			cmds = append(cmds, cmd)
		}
	}

	// Update loading state
	if len(chunksToLoad) > 0 {
		tl.hasLoadingChunks = true
		tl.canScroll = !data.IsLoadingCriticalChunks(tl.viewport, tl.config.ViewportConfig, tl.loadingChunks)
	}

	// Unload chunks outside bounding area
	chunksToUnload := data.FindChunksToUnload(tl.chunks, boundingArea, chunkSize)
	for _, chunkStart := range chunksToUnload {
		delete(tl.chunks, chunkStart)
		delete(tl.chunkAccessTime, chunkStart)
		cmds = append(cmds, core.ChunkUnloadedCmd(chunkStart))
	}

	return tea.Batch(cmds...)
}

// loadChunkFromFlattenedView creates a chunk from the flattened view
func (tl *TreeList[T]) loadChunkFromFlattenedView(chunkStart, chunkSize int) tea.Cmd {
	return func() tea.Msg {
		start := chunkStart
		count := chunkSize
		total := len(tl.flattenedView)

		if start >= total {
			return core.DataChunkLoadedMsg{
				StartIndex: start,
				Items:      []core.Data[any]{},
				Request:    core.DataRequest{Start: start, Count: count},
			}
		}

		end := start + count
		if end > total {
			end = total
		}

		var chunkItems []core.Data[any]
		for i := start; i < end; i++ {
			flatItem := tl.flattenedView[i]
			chunkItems = append(chunkItems, core.Data[any]{
				ID:       flatItem.ID,
				Item:     flatItem,
				Selected: tl.selectedNodes[flatItem.ID],
				Error:    nil,
				Loading:  false,
				Disabled: false,
			})
		}

		return core.DataChunkLoadedMsg{
			StartIndex: start,
			Items:      chunkItems,
			Request:    core.DataRequest{Start: start, Count: count},
		}
	}
}

// ================================
// NAVIGATION HANDLERS - Reuse List logic
// ================================

// handleCursorUp - reuse List logic
func (tl *TreeList[T]) handleCursorUp() tea.Cmd {
	if tl.totalItems == 0 || !tl.canScroll || tl.viewport.CursorIndex <= 0 {
		return nil
	}

	previousState := tl.viewport
	tl.viewport = viewport.CalculateCursorUp(tl.viewport, tl.config.ViewportConfig, tl.totalItems)

	if tl.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		tl.updateVisibleItems()
		return tl.smartChunkManagement()
	}
	return nil
}

// handleCursorDown - reuse List logic
func (tl *TreeList[T]) handleCursorDown() tea.Cmd {
	if tl.totalItems == 0 || !tl.canScroll || tl.viewport.CursorIndex >= tl.totalItems-1 {
		return nil
	}

	previousState := tl.viewport
	tl.viewport = viewport.CalculateCursorDown(tl.viewport, tl.config.ViewportConfig, tl.totalItems)

	if tl.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		tl.updateVisibleItems()
		return tl.smartChunkManagement()
	}
	return nil
}

// handlePageUp - reuse List logic
func (tl *TreeList[T]) handlePageUp() tea.Cmd {
	if tl.totalItems == 0 || !tl.canScroll {
		return nil
	}

	previousState := tl.viewport
	tl.viewport = viewport.CalculatePageUp(tl.viewport, tl.config.ViewportConfig, tl.totalItems)

	if tl.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		tl.updateVisibleItems()
	}
	return tl.smartChunkManagement()
}

// handlePageDown - reuse List logic
func (tl *TreeList[T]) handlePageDown() tea.Cmd {
	if tl.viewport.CursorIndex >= tl.totalItems-1 {
		return nil
	}

	previousState := tl.viewport
	tl.viewport = viewport.CalculatePageDown(tl.viewport, tl.config.ViewportConfig, tl.totalItems)

	if tl.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		tl.updateVisibleItems()
	}
	return tl.smartChunkManagement()
}

// handleJumpToStart - reuse List logic
func (tl *TreeList[T]) handleJumpToStart() tea.Cmd {
	if tl.totalItems == 0 || !tl.canScroll {
		return nil
	}

	tl.viewport = viewport.CalculateJumpToStart(tl.config.ViewportConfig, tl.totalItems)
	return tl.smartChunkManagement()
}

// handleJumpToEnd - reuse List logic
func (tl *TreeList[T]) handleJumpToEnd() tea.Cmd {
	if tl.totalItems <= 0 || !tl.canScroll {
		return nil
	}

	previousState := tl.viewport
	tl.viewport = viewport.CalculateJumpToEnd(tl.config.ViewportConfig, tl.totalItems)

	if tl.viewport.ViewportStartIndex != previousState.ViewportStartIndex {
		tl.updateVisibleItems()
		return tl.smartChunkManagement()
	}
	return nil
}

// handleJumpTo - reuse List logic
func (tl *TreeList[T]) handleJumpTo(index int) tea.Cmd {
	if tl.totalItems == 0 || index < 0 || index >= tl.totalItems || !tl.canScroll {
		return nil
	}

	tl.viewport = viewport.CalculateJumpTo(index, tl.config.ViewportConfig, tl.totalItems)
	return tl.smartChunkManagement()
}

// handleTreeJumpToIndex - handles tree-specific jumping with parent expansion
func (tl *TreeList[T]) handleTreeJumpToIndex(index int, expandParents bool) tea.Cmd {
	if tl.totalItems == 0 || index < 0 || !tl.canScroll {
		return nil
	}

	// If expandParents is false, just use regular jump
	if !expandParents {
		return tl.handleJumpTo(index)
	}

	// We need to jump to a specific index in the "fully expanded" tree
	// This means we need to:
	// 1. Create a temporary fully expanded view to find what item is at that index
	// 2. Find the path to that item in the tree
	// 3. Expand all parents in that path
	// 4. Re-flatten the view with the new expansions
	// 5. Find the new index of the target item in the re-flattened view
	// 6. Jump to that new index

	// Step 1: Create a fully expanded view to map the target index to an actual item
	fullyExpandedView := tl.createFullyExpandedView()

	if index >= len(fullyExpandedView) {
		// Index is beyond the fully expanded tree, jump to the end
		return tl.handleJumpToEnd()
	}

	targetItem := fullyExpandedView[index]

	// Step 2: Find the path to this item (all parent IDs)
	parentPath := tl.findPathToItem(targetItem.ID, tl.rootNodes, []string{})

	// Step 3: Expand all parents in the path
	var cmds []tea.Cmd
	expansionNeeded := false
	for _, parentID := range parentPath {
		if !tl.expandedNodes[parentID] {
			tl.expandedNodes[parentID] = true
			expansionNeeded = true
		}
	}

	// Step 4: Re-flatten the view if we made any expansions
	if expansionNeeded {
		tl.updateFlattenedView()
		cmds = append(cmds, core.DataTotalUpdateCmd(len(tl.flattenedView)))
		cmds = append(cmds, core.DataChunksRefreshCmd())
	}

	// Step 5: Find the new index of the target item in the current flattened view
	newIndex := tl.findItemIndexInFlattenedView(targetItem.ID)
	if newIndex == -1 {
		// Item not found in flattened view (shouldn't happen), fallback to regular jump
		return tea.Batch(cmds...)
	}

	// Step 6: Jump to the new index
	tl.viewport = viewport.CalculateJumpTo(newIndex, tl.config.ViewportConfig, tl.totalItems)
	cmds = append(cmds, tl.smartChunkManagement())

	return tea.Batch(cmds...)
}

// ================================
// HELPER METHODS - Reuse List logic
// ================================

// updateViewportPosition - reuse List logic
func (tl *TreeList[T]) updateViewportPosition() {
	tl.viewport = viewport.UpdateViewportPosition(tl.viewport, tl.config.ViewportConfig, tl.totalItems)
}

// updateViewportBounds - reuse List logic
func (tl *TreeList[T]) updateViewportBounds() {
	tl.viewport = viewport.UpdateViewportBounds(tl.viewport, tl.config.ViewportConfig, tl.totalItems)
}

// updateVisibleItems - reuse List logic
func (tl *TreeList[T]) updateVisibleItems() {
	result := viewport.CalculateVisibleItemsFromChunks(
		tl.viewport,
		tl.config.ViewportConfig,
		tl.totalItems,
		tl.chunks,
		tl.ensureChunkLoadedImmediate,
	)

	tl.visibleItems = result.Items
	tl.viewport = result.AdjustedViewport
}

// ensureChunkLoadedImmediate - reuse List logic
func (tl *TreeList[T]) ensureChunkLoadedImmediate(index int) {
	chunkStartIndex := data.CalculateChunkStartIndex(index, tl.config.ViewportConfig.ChunkSize)
	if _, exists := tl.chunks[chunkStartIndex]; !exists {
		// Load chunk immediately from flattened view
		cmd := tl.loadChunkFromFlattenedView(chunkStartIndex, tl.config.ViewportConfig.ChunkSize)
		if msg := cmd(); msg != nil {
			if chunkMsg, ok := msg.(core.DataChunkLoadedMsg); ok {
				tl.handleDataChunkLoaded(chunkMsg)
			}
		}
	}
}

// ================================
// REMAINING HANDLERS - Similar to List
// ================================

// handleDataRefresh refreshes all data
func (tl *TreeList[T]) handleDataRefresh() tea.Cmd {
	tl.chunks = make(map[int]core.Chunk[any])
	tl.updateFlattenedView()
	return core.DataTotalCmd(tl.totalItems)
}

// handleDataChunkLoaded processes a loaded data chunk
func (tl *TreeList[T]) handleDataChunkLoaded(msg core.DataChunkLoadedMsg) tea.Cmd {
	chunk := core.Chunk[any]{
		StartIndex: msg.StartIndex,
		EndIndex:   msg.StartIndex + len(msg.Items) - 1,
		Items:      msg.Items,
		LoadedAt:   time.Now(),
		Request:    msg.Request,
	}

	tl.chunks[msg.StartIndex] = chunk
	delete(tl.loadingChunks, msg.StartIndex)
	tl.hasLoadingChunks = len(tl.loadingChunks) > 0
	if !tl.hasLoadingChunks {
		tl.canScroll = true
	} else {
		tl.canScroll = !data.IsLoadingCriticalChunks(tl.viewport, tl.config.ViewportConfig, tl.loadingChunks)
	}

	tl.updateVisibleItems()
	tl.updateViewportBounds()

	return core.ChunkLoadingCompletedCmd(msg.StartIndex, len(msg.Items), msg.Request)
}

// handleSelectCurrent selects the current item with optional cascading
func (tl *TreeList[T]) handleSelectCurrent() tea.Cmd {
	if tl.config.SelectionMode == core.SelectionNone || tl.totalItems == 0 {
		return nil
	}

	if tl.viewport.CursorIndex >= 0 && tl.viewport.CursorIndex < len(tl.flattenedView) {
		currentItem := tl.flattenedView[tl.viewport.CursorIndex]

		// Toggle the current item's selection
		newSelectionState := !tl.selectedNodes[currentItem.ID]
		tl.selectedNodes[currentItem.ID] = newSelectionState

		// If cascading selection is enabled and this item has children, cascade the selection
		if tl.treeConfig.CascadingSelection && currentItem.HasChildren() {
			tl.cascadeSelection(currentItem.ID, newSelectionState)
		}

		return tl.refreshChunks()
	}
	return nil
}

// cascadeSelection recursively selects/deselects all children of a node
func (tl *TreeList[T]) cascadeSelection(parentID string, selected bool) {
	// Find the parent node in the tree structure
	parentNode, found := tl.findNodeInTree(tl.rootNodes, parentID)
	if !found {
		return
	}

	// Recursively select/deselect all children
	tl.cascadeSelectionRecursive(parentNode.Children, selected)
}

// cascadeSelectionRecursive recursively applies selection to all descendant nodes
func (tl *TreeList[T]) cascadeSelectionRecursive(nodes []TreeData[T], selected bool) {
	for _, node := range nodes {
		// Set selection state for this node
		if selected {
			tl.selectedNodes[node.ID] = true
		} else {
			delete(tl.selectedNodes, node.ID)
		}

		// Recursively apply to children
		if len(node.Children) > 0 {
			tl.cascadeSelectionRecursive(node.Children, selected)
		}
	}
}

// findNodeInTree recursively searches for a node by ID in the tree structure
func (tl *TreeList[T]) findNodeInTree(nodes []TreeData[T], id string) (TreeData[T], bool) {
	for _, node := range nodes {
		if node.ID == id {
			return node, true
		}
		if found, ok := tl.findNodeInTree(node.Children, id); ok {
			return found, true
		}
	}
	return TreeData[T]{}, false
}

// handleSelectAll selects all items
func (tl *TreeList[T]) handleSelectAll() tea.Cmd {
	if tl.config.SelectionMode != core.SelectionMultiple {
		return nil
	}

	for _, item := range tl.flattenedView {
		tl.selectedNodes[item.ID] = true
	}
	return tl.refreshChunks()
}

// handleSelectClear clears all selections
func (tl *TreeList[T]) handleSelectClear() tea.Cmd {
	tl.selectedNodes = make(map[string]bool)
	return tl.refreshChunks()
}

// refreshChunks reloads existing chunks to get updated selection state
func (tl *TreeList[T]) refreshChunks() tea.Cmd {
	var cmds []tea.Cmd

	// Reload all currently loaded chunks
	for chunkStart := range tl.chunks {
		cmd := tl.loadChunkFromFlattenedView(chunkStart, tl.config.ViewportConfig.ChunkSize)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// handleKeyPress handles keyboard input - reuse List logic
func (tl *TreeList[T]) handleKeyPress(msg tea.KeyMsg) tea.Cmd {
	if !tl.focused {
		return nil
	}

	key := msg.String()

	// Check navigation keys - reuse List key mapping logic
	for _, upKey := range tl.config.KeyMap.Up {
		if key == upKey {
			return tl.handleCursorUp()
		}
	}

	for _, downKey := range tl.config.KeyMap.Down {
		if key == downKey {
			return tl.handleCursorDown()
		}
	}

	// Add other key handlers...
	return nil
}

// ================================
// TREE-SPECIFIC ENUMERATORS
// ================================

// TreeEnumerator creates tree-style enumeration with proper indentation and tree symbols
func TreeEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	// Type assert to check if this is a tree item
	if flatItem, ok := item.Item.(interface {
		GetDepth() int
		HasChildren() bool
		IsExpanded() bool
	}); ok {
		var prefix strings.Builder

		// Add indentation based on depth
		depth := flatItem.GetDepth()
		for i := 0; i < depth; i++ {
			prefix.WriteString("  ")
		}

		// Add tree connector
		if flatItem.HasChildren() {
			if flatItem.IsExpanded() {
				prefix.WriteString("▼ ")
			} else {
				prefix.WriteString("▶ ")
			}
		} else {
			prefix.WriteString("• ")
		}

		return prefix.String()
	}

	// Fallback to bullet for non-tree items
	return "• "
}

// TreeExpandedEnumerator shows different symbols for expanded/collapsed nodes
func TreeExpandedEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	if flatItem, ok := item.Item.(interface {
		GetDepth() int
		HasChildren() bool
		IsExpanded() bool
	}); ok {
		var prefix strings.Builder

		// Add indentation
		depth := flatItem.GetDepth()
		for i := 0; i < depth; i++ {
			prefix.WriteString("│ ")
		}

		// Add tree connector with box drawing characters
		if flatItem.HasChildren() {
			if flatItem.IsExpanded() {
				prefix.WriteString("├─")
			} else {
				prefix.WriteString("├+")
			}
		} else {
			prefix.WriteString("└─")
		}

		return prefix.String()
	}

	return "• "
}

// TreeMinimalEnumerator provides minimal tree visualization
func TreeMinimalEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	if flatItem, ok := item.Item.(interface {
		GetDepth() int
		HasChildren() bool
		IsExpanded() bool
	}); ok {
		var prefix strings.Builder

		// Add simple indentation
		depth := flatItem.GetDepth()
		for i := 0; i < depth; i++ {
			prefix.WriteString("  ")
		}

		// Simple symbols
		if flatItem.HasChildren() {
			if flatItem.IsExpanded() {
				prefix.WriteString("- ")
			} else {
				prefix.WriteString("+ ")
			}
		} else {
			prefix.WriteString("  ")
		}

		return prefix.String()
	}

	return ""
}

// ================================
// TREE CONFIGURATION - Component-based
// ================================

// SetTreeEnumerator sets the tree to use tree-style enumeration
func (tl *TreeList[T]) SetTreeEnumerator() {
	// Create a wrapper function that matches TreeEnumeratorFunc signature
	tl.treeConfig.RenderConfig.EnumeratorConfig.Enumerator = func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext) string {
		return TreeEnumerator(item, index, ctx)
	}
}

// SetTreeExpandedEnumerator sets the tree to use expanded tree-style enumeration
func (tl *TreeList[T]) SetTreeExpandedEnumerator() {
	// Create a wrapper function that matches TreeEnumeratorFunc signature
	tl.treeConfig.RenderConfig.EnumeratorConfig.Enumerator = func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext) string {
		return TreeExpandedEnumerator(item, index, ctx)
	}
}

// SetTreeMinimalEnumerator sets the tree to use minimal tree-style enumeration
func (tl *TreeList[T]) SetTreeMinimalEnumerator() {
	// Create a wrapper function that matches TreeEnumeratorFunc signature
	tl.treeConfig.RenderConfig.EnumeratorConfig.Enumerator = func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext) string {
		return TreeMinimalEnumerator(item, index, ctx)
	}
}

// SetCascadingSelection enables or disables cascading selection
func (tl *TreeList[T]) SetCascadingSelection(enabled bool) {
	tl.treeConfig.CascadingSelection = enabled
}

// GetCascadingSelection returns whether cascading selection is enabled
func (tl *TreeList[T]) GetCascadingSelection() bool {
	return tl.treeConfig.CascadingSelection
}

// SetRenderConfig sets the complete render configuration
func (tl *TreeList[T]) SetRenderConfig(config TreeRenderConfig) {
	tl.treeConfig.RenderConfig = config
}

// GetRenderConfig returns the current render configuration
func (tl *TreeList[T]) GetRenderConfig() TreeRenderConfig {
	return tl.treeConfig.RenderConfig
}

// SetAutoExpand enables or disables auto-expansion of nodes
func (tl *TreeList[T]) SetAutoExpand(enabled bool) {
	tl.treeConfig.AutoExpand = enabled
}

// GetAutoExpand returns whether auto-expansion is enabled
func (tl *TreeList[T]) GetAutoExpand() bool {
	return tl.treeConfig.AutoExpand
}

// SetExpandOnSelect enables or disables expanding nodes when selected
func (tl *TreeList[T]) SetExpandOnSelect(enabled bool) {
	tl.treeConfig.ExpandOnSelect = enabled
}

// GetExpandOnSelect returns whether expand-on-select is enabled
func (tl *TreeList[T]) GetExpandOnSelect() bool {
	return tl.treeConfig.ExpandOnSelect
}

// setupRenderContext - reuse List logic
func (tl *TreeList[T]) setupRenderContext() {
	tl.renderContext = core.RenderContext{
		MaxWidth:          tl.config.MaxWidth,
		MaxHeight:         1,
		Theme:             nil,
		BaseStyle:         tl.config.StyleConfig.DefaultStyle,
		ColorSupport:      true,
		UnicodeSupport:    true,
		CurrentTime:       time.Now(),
		FocusState:        core.FocusState{HasFocus: tl.focused},
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
			tl.lastError = err
		},
	}
}

// ================================
// PUBLIC INTERFACE - Same as List
// ================================

// Focus sets the tree list as focused
func (tl *TreeList[T]) Focus() tea.Cmd {
	tl.focused = true
	return nil
}

// Blur removes focus from the tree list
func (tl *TreeList[T]) Blur() {
	tl.focused = false
}

// IsFocused returns whether the tree list has focus
func (tl *TreeList[T]) IsFocused() bool {
	return tl.focused
}

// GetState returns the current viewport state
func (tl *TreeList[T]) GetState() core.ViewportState {
	return tl.viewport
}

// GetSelectionCount returns the number of selected items
func (tl *TreeList[T]) GetSelectionCount() int {
	return len(tl.selectedNodes)
}

// ================================
// CURSOR AND STYLING METHODS
// ================================

// SetEnumerator sets the tree enumerator function
func (tl *TreeList[T]) SetEnumerator(enum tree.Enumerator) {
	tl.treeConfig.Enumerator = enum
}

// SetCursorIndicator sets the cursor indicator string
func (tl *TreeList[T]) SetCursorIndicator(indicator string) {
	tl.treeConfig.CursorIndicator = indicator
}

// SetCursorSpacing sets the spacing for cursor lines
func (tl *TreeList[T]) SetCursorSpacing(spacing string) {
	tl.treeConfig.CursorSpacing = spacing
}

// SetNormalSpacing sets the spacing for non-cursor lines
func (tl *TreeList[T]) SetNormalSpacing(spacing string) {
	tl.treeConfig.NormalSpacing = spacing
}

// SetShowCursor enables or disables cursor indicators
func (tl *TreeList[T]) SetShowCursor(show bool) {
	tl.treeConfig.ShowCursor = show
}

// SetEnableCursorStyling enables or disables cursor background styling
func (tl *TreeList[T]) SetEnableCursorStyling(enabled bool) {
	tl.treeConfig.EnableCursorStyling = enabled
}

// GetEnableCursorStyling returns whether cursor background styling is enabled
func (tl *TreeList[T]) GetEnableCursorStyling() bool {
	return tl.treeConfig.EnableCursorStyling
}

// SetCursorStyle is a convenience method to set common cursor styling options
func (tl *TreeList[T]) SetCursorStyle(showIndicator bool, backgroundColor, foregroundColor string) {
	tl.treeConfig.ShowCursor = showIndicator
	tl.treeConfig.EnableCursorStyling = true
	tl.treeConfig.CursorBackgroundStyle = lipgloss.NewStyle().
		Background(lipgloss.Color(backgroundColor)).
		Foreground(lipgloss.Color(foregroundColor))
}

// JumpToIndexExpandingParents jumps to a specific index in the fully expanded tree,
// automatically expanding all parent nodes necessary to make the target item visible
func (tl *TreeList[T]) JumpToIndexExpandingParents(index int) tea.Cmd {
	return core.TreeJumpToIndexCmd(index, true)
}

// GetFullyExpandedItemCount returns the total number of items if the tree were fully expanded
func (tl *TreeList[T]) GetFullyExpandedItemCount() int {
	fullyExpandedView := tl.createFullyExpandedView()
	return len(fullyExpandedView)
}
