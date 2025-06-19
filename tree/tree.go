// Package tree provides a hierarchical list component for Bubble Tea applications,
// extending the core functionalities of the list component with tree-specific
// features like node expansion, indentation, and parent-child relationships.
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

// TreeData represents a single node in a hierarchical dataset. It contains the
// item's data, its unique ID, and a slice of its children, forming the fundamental
// structure for the tree.
type TreeData[T any] struct {
	ID       string
	Item     T
	Children []TreeData[T]
	Expanded bool
}

// TreeDataSource defines the contract for providing hierarchical data to the
// TreeList component. Implementations are responsible for fetching the tree
// structure and handling operations like selection.
type TreeDataSource[T any] interface {
	// GetRootNodes returns the top-level nodes of the tree.
	GetRootNodes() []TreeData[T]

	// GetItemByID retrieves a specific tree node by its unique ID.
	GetItemByID(id string) (TreeData[T], bool)
	// SetSelected sends a command to update the selection state of a node by its ID.
	SetSelected(id string, selected bool) tea.Cmd
	// SetSelectedByID is an alias for SetSelected.
	SetSelectedByID(id string, selected bool) tea.Cmd
	// SelectAll sends a command to select all nodes.
	SelectAll() tea.Cmd
	// ClearSelection sends a command to clear all selections.
	ClearSelection() tea.Cmd
	// SelectRange sends a command to select a range of nodes between two IDs.
	SelectRange(startID, endID string) tea.Cmd
}

// FlatTreeItem represents a tree node within the flattened, linear view used for
// rendering. It contains the original item data, along with metadata about its
// position in the tree, such as depth and expansion state.
type FlatTreeItem[T any] struct {
	ID            string
	Item          T
	Depth         int
	HasChildNodes bool // Renamed to avoid conflict with method
	Expanded      bool
	ParentID      string
}

// GetDepth returns the indentation level of this tree item.
func (f FlatTreeItem[T]) GetDepth() int {
	return f.Depth
}

// HasChildren returns true if this item has child nodes.
func (f FlatTreeItem[T]) HasChildren() bool {
	return f.HasChildNodes
}

// IsExpanded returns true if this item is currently expanded.
func (f FlatTreeItem[T]) IsExpanded() bool {
	return f.Expanded
}

// TreeList is a stateful Bubble Tea component that displays a scrollable,
// hierarchical list. It manages tree-specific state like node expansion and
// selection, flattens the tree structure for efficient rendering, and reuses
// core list functionalities for viewport management, data virtualization, and
// user interactions.
type TreeList[T any] struct {
	// Core state for hierarchical data
	treeDataSource TreeDataSource[T]
	chunks         map[int]core.Chunk[any] // Reuses the same chunk system as the List
	totalItems     int                     // Total number of *visible* items in the flattened view

	// Viewport state - managed identically to the List
	viewport core.ViewportState

	// Configuration - reuses the core ListConfig
	config core.ListConfig

	// Tree-specific state
	rootNodes     []TreeData[T]     // The original hierarchical data
	expandedNodes map[string]bool   // A set of IDs for currently expanded nodes
	selectedNodes map[string]bool   // A set of IDs for currently selected nodes
	flattenedView []FlatTreeItem[T] // The cached linear representation of the visible tree

	// Rendering - uses a tree-specific component system
	formatter         core.ItemFormatter[any]
	animatedFormatter core.ItemFormatterAnimated[any]
	renderContext     core.RenderContext

	// Focus state
	focused bool

	// Tree-specific configuration
	treeConfig TreeConfig

	// Chunk and visibility management - identical to the List
	visibleItems     []core.Data[any]
	chunkAccessTime  map[int]time.Time
	loadingChunks    map[int]bool
	hasLoadingChunks bool
	canScroll        bool

	// Error handling
	lastError error
}

// TreeConfig contains configuration options specific to the TreeList component,
// primarily related to rendering and interaction with the tree structure.
type TreeConfig struct {
	// RenderConfig defines the component-based rendering pipeline for the tree.
	RenderConfig TreeRenderConfig

	// CascadingSelection, when true, causes selecting a parent node to
	// automatically select all of its descendant nodes.
	CascadingSelection bool
	// AutoExpand, when true, automatically expands a collapsed node when the
	// cursor moves to it.
	AutoExpand bool
	// ShowRoot, when true, applies special styling to root-level nodes.
	ShowRoot bool

	// ExpandOnSelect, when true, expands or collapses a node when it is selected.
	ExpandOnSelect bool

	// The fields below are legacy and kept for backward compatibility. The
	// component-based rendering system in `TreeRenderConfig` is now the
	// preferred way to control appearance.
	Enumerator            tree.Enumerator
	Indenter              tree.Indenter
	RootStyle             lipgloss.Style
	ItemStyle             lipgloss.Style
	EnumeratorStyle       lipgloss.Style
	CursorIndicator       string
	CursorSpacing         string
	NormalSpacing         string
	ShowCursor            bool
	EnableCursorStyling   bool
	CursorBackgroundStyle lipgloss.Style
}

// DefaultTreeConfig returns a set of sensible default configurations for a TreeList.
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

// NewTreeList creates a new TreeList component with the given configurations and
// data source. It initializes the tree's state, flattens the initial view, and
// sets up the rendering context.
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

// Init initializes the TreeList component, loading the initial data. It is part
// of the bubbletea.Model interface.
func (tl *TreeList[T]) Init() tea.Cmd {
	return tl.loadInitialData()
}

// Update is the central message handler for the TreeList component. It processes
// messages for navigation, data manipulation, tree expansion, and other state
// changes. It implements the bubbletea.Model interface.
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

// View renders the TreeList component into a string. It calculates the visible
// items based on the current viewport, formats each item using the configured
// tree rendering pipeline, and assembles the final output.
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

// formatTreeItem formats a single tree item using the legacy, non-component-based
// rendering system. It constructs the item string with indentation, symbols,
// and content, applying cursor and selection styles.
//
// Deprecated: The component-based rendering system configured via
// `TreeRenderConfig` is now the preferred method for rendering.
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

// formatItemContent provides default formatting for the main content of an item.
// It uses fmt.Stringer if available, otherwise it uses a default format.
func (tl *TreeList[T]) formatItemContent(item T) string {
	if stringer, ok := any(item).(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%v", item)
}

// ExpandNode expands a tree node specified by its ID, revealing its children.
// It then updates the flattened view and refreshes the data.
func (tl *TreeList[T]) ExpandNode(id string) tea.Cmd {
	tl.expandedNodes[id] = true
	tl.updateFlattenedView()
	// Update total and refresh chunks
	return tea.Batch(
		core.DataTotalUpdateCmd(len(tl.flattenedView)),
		core.DataChunksRefreshCmd(),
	)
}

// CollapseNode collapses a tree node specified by its ID, hiding its children.
// It then updates the flattened view and refreshes the data.
func (tl *TreeList[T]) CollapseNode(id string) tea.Cmd {
	delete(tl.expandedNodes, id)
	tl.updateFlattenedView()
	// Update total and refresh chunks
	return tea.Batch(
		core.DataTotalUpdateCmd(len(tl.flattenedView)),
		core.DataChunksRefreshCmd(),
	)
}

// ToggleNode toggles the expansion state of a tree node specified by its ID.
func (tl *TreeList[T]) ToggleNode(id string) tea.Cmd {
	if tl.expandedNodes[id] {
		return tl.CollapseNode(id)
	}
	return tl.ExpandNode(id)
}

// ToggleCurrentNode toggles the expansion state of the node currently under the
// cursor.
func (tl *TreeList[T]) ToggleCurrentNode() tea.Cmd {
	if tl.viewport.CursorIndex >= 0 && tl.viewport.CursorIndex < len(tl.flattenedView) {
		currentItem := tl.flattenedView[tl.viewport.CursorIndex]
		if currentItem.HasChildren() {
			return tl.ToggleNode(currentItem.ID)
		}
	}
	return nil
}

// createFullyExpandedView generates a flattened view of the tree as if all
// nodes were expanded. This is used for operations like `JumpToIndexExpandingParents`.
func (tl *TreeList[T]) createFullyExpandedView() []FlatTreeItem[T] {
	var fullyExpandedView []FlatTreeItem[T]
	tl.flattenNodesFullyExpanded(tl.rootNodes, "", 0, &fullyExpandedView)
	return fullyExpandedView
}

// flattenNodesFullyExpanded is a recursive helper to flatten the tree with all
// nodes treated as expanded.
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

// findPathToItem finds the sequence of parent IDs leading to a target node.
// This is used to expand all ancestors of a node when jumping to it.
func (tl *TreeList[T]) findPathToItem(targetID string, nodes []TreeData[T], currentPath []string) []string {
	for _, node := range nodes {
		if node.ID == targetID {
			// Found the target item, return the current path
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

// findItemIndexInFlattenedView gets the linear index of a node by its ID in the
// current flattened view.
func (tl *TreeList[T]) findItemIndexInFlattenedView(itemID string) int {
	for i, item := range tl.flattenedView {
		if item.ID == itemID {
			return i
		}
	}
	return -1 // Not found
}

// updateFlattenedView rebuilds the cached `flattenedView` based on the current
// expansion state of the nodes. This is a critical operation performed whenever
// the tree structure changes.
func (tl *TreeList[T]) updateFlattenedView() {
	tl.flattenedView = nil
	tl.flattenNodes(tl.rootNodes, "", 0)
	tl.totalItems = len(tl.flattenedView)
}

// flattenNodes is a recursive helper function that traverses the tree data and
// builds the `flattenedView`, respecting the current expansion state of each node.
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

// loadInitialData prepares the initial state of the tree by setting the total
// number of items based on the initial flattened view.
func (tl *TreeList[T]) loadInitialData() tea.Cmd {
	// Set initial total
	return core.DataTotalCmd(len(tl.flattenedView))
}

// smartChunkManagement is the core logic for data virtualization. It determines
// which chunks of the flattened view to load based on the viewport position. It
// reuses the core list's chunking logic.
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

// loadChunkFromFlattenedView creates a `DataChunkLoadedMsg` command for a
// specific segment of the `flattenedView`. This simulates chunk loading for the
// in-memory flattened tree structure.
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

// handleCursorUp processes a "cursor up" event, recalculating the viewport
// and cursor positions. It reuses the core viewport logic.
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

// handleCursorDown processes a "cursor down" event, adjusting the viewport for
// downward movement. It reuses the core viewport logic.
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

// handlePageUp processes a "page up" event, moving the cursor and viewport up
// by one page. It reuses the core viewport logic.
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

// handlePageDown processes a "page down" event, moving the cursor and viewport
// down by one page. It reuses the core viewport logic.
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

// handleJumpToStart moves the cursor and viewport to the very beginning of the
// list. It reuses the core viewport logic.
func (tl *TreeList[T]) handleJumpToStart() tea.Cmd {
	if tl.totalItems == 0 || !tl.canScroll {
		return nil
	}

	tl.viewport = viewport.CalculateJumpToStart(tl.config.ViewportConfig, tl.totalItems)
	return tl.smartChunkManagement()
}

// handleJumpToEnd moves the cursor and viewport to the very end of the list.
// It reuses the core viewport logic.
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

// handleJumpTo moves the cursor and viewport to a specific item index in the
// flattened view. It reuses the core viewport logic.
func (tl *TreeList[T]) handleJumpTo(index int) tea.Cmd {
	if tl.totalItems == 0 || index < 0 || index >= tl.totalItems || !tl.canScroll {
		return nil
	}

	tl.viewport = viewport.CalculateJumpTo(index, tl.config.ViewportConfig, tl.totalItems)
	return tl.smartChunkManagement()
}

// handleTreeJumpToIndex handles a jump to a specific index in the conceptual
// "fully expanded" tree, automatically expanding parent nodes if required.
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

// updateViewportPosition ensures the viewport is correctly positioned to keep
// the cursor visible. It reuses the core viewport logic.
func (tl *TreeList[T]) updateViewportPosition() {
	tl.viewport = viewport.UpdateViewportPosition(tl.viewport, tl.config.ViewportConfig, tl.totalItems)
}

// updateViewportBounds recalculates the boundary flags of the viewport. It reuses
// the core viewport logic.
func (tl *TreeList[T]) updateViewportBounds() {
	tl.viewport = viewport.UpdateViewportBounds(tl.viewport, tl.config.ViewportConfig, tl.totalItems)
}

// updateVisibleItems recalculates the `visibleItems` slice based on the
// current viewport and loaded chunks. It reuses the core list logic.
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

// ensureChunkLoadedImmediate is a helper to request a chunk if it's not loaded,
// used to fill in missing data for the current view. It reuses the core list logic.
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

// handleDataRefresh performs a hard refresh of the tree's data. It clears all
// local caches and re-initiates the data loading process.
func (tl *TreeList[T]) handleDataRefresh() tea.Cmd {
	tl.chunks = make(map[int]core.Chunk[any])
	tl.updateFlattenedView()
	return core.DataTotalCmd(tl.totalItems)
}

// handleDataChunkLoaded processes a newly loaded data chunk. It adds the chunk
// to the local cache and updates the loading state.
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

// handleSelectCurrent toggles the selection state of the item currently under
// the cursor. If cascading selection is enabled, it also toggles all descendants.
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

// cascadeSelection recursively applies the selection state to a node's descendants.
func (tl *TreeList[T]) cascadeSelection(parentID string, selected bool) {
	// Find the parent node in the tree structure
	parentNode, found := tl.findNodeInTree(tl.rootNodes, parentID)
	if !found {
		return
	}

	// Recursively select/deselect all children
	tl.cascadeSelectionRecursive(parentNode.Children, selected)
}

// cascadeSelectionRecursive is a helper that recursively applies selection
// state to a slice of nodes and their children.
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

// findNodeInTree recursively searches for a node by its ID in the original
// hierarchical tree data.
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

// handleSelectAll selects all currently visible items in the flattened view.
func (tl *TreeList[T]) handleSelectAll() tea.Cmd {
	if tl.config.SelectionMode != core.SelectionMultiple {
		return nil
	}

	for _, item := range tl.flattenedView {
		tl.selectedNodes[item.ID] = true
	}
	return tl.refreshChunks()
}

// handleSelectClear clears all node selections.
func (tl *TreeList[T]) handleSelectClear() tea.Cmd {
	tl.selectedNodes = make(map[string]bool)
	return tl.refreshChunks()
}

// refreshChunks forces a reload of all currently loaded data chunks. This is
// useful for reflecting state changes (like selection) in the view.
func (tl *TreeList[T]) refreshChunks() tea.Cmd {
	var cmds []tea.Cmd

	// Reload all currently loaded chunks
	for chunkStart := range tl.chunks {
		cmd := tl.loadChunkFromFlattenedView(chunkStart, tl.config.ViewportConfig.ChunkSize)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// handleKeyPress processes raw key presses, mapping them to tree actions based
// on the current keymap. It reuses the core list's key handling logic.
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

// TreeEnumerator is a legacy enumerator function that creates a tree-style
// prefix with indentation and expand/collapse symbols.
//
// Deprecated: Use the component-based rendering system with `TreeIndentationComponent`
// and `TreeSymbolComponent` for more flexibility.
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

// TreeExpandedEnumerator is a legacy enumerator that uses box-drawing characters
// to create a connected tree look.
//
// Deprecated: Use the component-based rendering system and set
// `TreeIndentationConfig.UseConnectors` to true.
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

// TreeMinimalEnumerator provides a minimal tree visualization with simple
// indentation and +/- symbols.
//
// Deprecated: This style can be achieved and customized via the component system.
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

// SetTreeEnumerator configures the tree to use a basic tree-style enumerator.
func (tl *TreeList[T]) SetTreeEnumerator() {
	// Create a wrapper function that matches TreeEnumeratorFunc signature
	tl.treeConfig.RenderConfig.EnumeratorConfig.Enumerator = func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext) string {
		return TreeEnumerator(item, index, ctx)
	}
}

// SetTreeExpandedEnumerator configures the tree to use an enumerator with
// box-drawing characters for a connected look.
func (tl *TreeList[T]) SetTreeExpandedEnumerator() {
	// Create a wrapper function that matches TreeEnumeratorFunc signature
	tl.treeConfig.RenderConfig.EnumeratorConfig.Enumerator = func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext) string {
		return TreeExpandedEnumerator(item, index, ctx)
	}
}

// SetTreeMinimalEnumerator configures the tree to use a minimal enumerator style.
func (tl *TreeList[T]) SetTreeMinimalEnumerator() {
	// Create a wrapper function that matches TreeEnumeratorFunc signature
	tl.treeConfig.RenderConfig.EnumeratorConfig.Enumerator = func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext) string {
		return TreeMinimalEnumerator(item, index, ctx)
	}
}

// SetCascadingSelection enables or disables the automatic selection of child
// nodes when a parent is selected.
func (tl *TreeList[T]) SetCascadingSelection(enabled bool) {
	tl.treeConfig.CascadingSelection = enabled
}

// GetCascadingSelection returns whether cascading selection is currently enabled.
func (tl *TreeList[T]) GetCascadingSelection() bool {
	return tl.treeConfig.CascadingSelection
}

// SetRenderConfig applies a completely new rendering configuration for the tree.
func (tl *TreeList[T]) SetRenderConfig(config TreeRenderConfig) {
	tl.treeConfig.RenderConfig = config
}

// GetRenderConfig returns the current tree rendering configuration.
func (tl *TreeList[T]) GetRenderConfig() TreeRenderConfig {
	return tl.treeConfig.RenderConfig
}

// SetAutoExpand enables or disables the automatic expansion of nodes when the
// cursor moves to them.
func (tl *TreeList[T]) SetAutoExpand(enabled bool) {
	tl.treeConfig.AutoExpand = enabled
}

// GetAutoExpand returns whether auto-expansion is currently enabled.
func (tl *TreeList[T]) GetAutoExpand() bool {
	return tl.treeConfig.AutoExpand
}

// SetExpandOnSelect enables or disables expanding/collapsing a node when it is selected.
func (tl *TreeList[T]) SetExpandOnSelect(enabled bool) {
	tl.treeConfig.ExpandOnSelect = enabled
}

// GetExpandOnSelect returns whether expand-on-select is currently enabled.
func (tl *TreeList[T]) GetExpandOnSelect() bool {
	return tl.treeConfig.ExpandOnSelect
}

// setupRenderContext initializes the render context with values from the list's
// configuration. It reuses the core list's setup logic.
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

// Focus sets the tree to a focused state, allowing it to receive and handle
// keyboard inputs.
func (tl *TreeList[T]) Focus() tea.Cmd {
	tl.focused = true
	return nil
}

// Blur removes focus from the tree, preventing it from handling keyboard inputs.
func (tl *TreeList[T]) Blur() {
	tl.focused = false
}

// IsFocused returns true if the tree is currently focused.
func (tl *TreeList[T]) IsFocused() bool {
	return tl.focused
}

// GetState returns the current state of the viewport, including cursor position
// and scroll offset.
func (tl *TreeList[T]) GetState() core.ViewportState {
	return tl.viewport
}

// GetSelectionCount returns the number of currently selected nodes.
func (tl *TreeList[T]) GetSelectionCount() int {
	return len(tl.selectedNodes)
}

// // SetEnumerator sets the tree enumerator function.
// //
// // Deprecated: Use the component-based rendering system via SetRenderConfig.
// func (tl *TreeList[T]) SetEnumerator(enum tree.Enumerator) {
// 	tl.treeConfig.Enumerator = enum
// }

// // SetCursorIndicator sets the cursor indicator string.
// //
// // Deprecated: Use the component-based rendering system via SetRenderConfig.
// func (tl *TreeList[T]) SetCursorIndicator(indicator string) {
// 	tl.treeConfig.CursorIndicator = indicator
// }

// // SetCursorSpacing sets the spacing for cursor lines.
// //
// // Deprecated: Use the component-based rendering system via SetRenderConfig.
// func (tl *TreeList[T]) SetCursorSpacing(spacing string) {
// 	tl.treeConfig.CursorSpacing = spacing
// }

// // SetNormalSpacing sets the spacing for non-cursor lines.
// //
// // Deprecated: Use the component-based rendering system via SetRenderConfig.
// func (tl *TreeList[T]) SetNormalSpacing(spacing string) {
// 	tl.treeConfig.NormalSpacing = spacing
// }

// // SetShowCursor enables or disables cursor indicators.
// //
// // Deprecated: Use the component-based rendering system via SetRenderConfig.
// func (tl *TreeList[T]) SetShowCursor(show bool) {
// 	tl.treeConfig.ShowCursor = show
// }

// // SetEnableCursorStyling enables or disables cursor background styling.
// //
// // Deprecated: Use the component-based rendering system via SetRenderConfig.
// func (tl *TreeList[T]) SetEnableCursorStyling(enabled bool) {
// 	tl.treeConfig.EnableCursorStyling = enabled
// }

// // GetEnableCursorStyling returns whether cursor background styling is enabled.
// //
// // Deprecated: Use the component-based rendering system via SetRenderConfig.
// func (tl *TreeList[T]) GetEnableCursorStyling() bool {
// 	return tl.treeConfig.EnableCursorStyling
// }

// // SetCursorStyle is a convenience method to set common cursor styling options.
// //
// // Deprecated: Use the component-based rendering system via SetRenderConfig.
// func (tl *TreeList[T]) SetCursorStyle(showIndicator bool, backgroundColor, foregroundColor string) {
// 	tl.treeConfig.ShowCursor = showIndicator
// 	tl.treeConfig.EnableCursorStyling = true
// 	tl.treeConfig.CursorBackgroundStyle = lipgloss.NewStyle().
// 		Background(lipgloss.Color(backgroundColor)).
// 		Foreground(lipgloss.Color(foregroundColor))
// }

// JumpToIndexExpandingParents creates a command to jump to a specific index in the
// conceptual "fully expanded" tree. It automatically expands all parent nodes
// necessary to make the target item visible.
func (tl *TreeList[T]) JumpToIndexExpandingParents(index int) tea.Cmd {
	return core.TreeJumpToIndexCmd(index, true)
}

// GetFullyExpandedItemCount returns the total number of items the tree would have
// if all nodes were expanded.
func (tl *TreeList[T]) GetFullyExpandedItemCount() int {
	fullyExpandedView := tl.createFullyExpandedView()
	return len(fullyExpandedView)
}
