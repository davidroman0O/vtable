package vtable

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// ================================
// VIEWPORT COMPONENT
// ================================

// Viewport manages viewport state, navigation, and boundary detection
// It's a pure Bubble Tea component that can be embedded in List/Table
type Viewport struct {
	config ViewportConfig
	state  ViewportState

	// Dataset information
	totalItems int

	// Event callbacks (optional - for notifications)
	onViewportChanged func(ViewportState)
	onCursorChanged   func(int)

	// Debug mode
	debugMode bool
}

// ================================
// NAVIGATION TYPES
// ================================

// NavigationType represents different navigation operations
type NavigationType int

const (
	NavigationUp NavigationType = iota
	NavigationDown
	NavigationPageUp
	NavigationPageDown
	NavigationStart
	NavigationEnd
	NavigationJump
)

// String returns string representation of navigation type
func (n NavigationType) String() string {
	switch n {
	case NavigationUp:
		return "up"
	case NavigationDown:
		return "down"
	case NavigationPageUp:
		return "page_up"
	case NavigationPageDown:
		return "page_down"
	case NavigationStart:
		return "start"
	case NavigationEnd:
		return "end"
	case NavigationJump:
		return "jump"
	default:
		return "unknown"
	}
}

// ================================
// VIEWPORT MESSAGES
// ================================

// ViewportNavigationMsg represents navigation commands
type ViewportNavigationMsg struct {
	Type   NavigationType
	Amount int // For jump operations
}

// Specific navigation message constructors for type safety
type ViewportUpMsg struct{}
type ViewportDownMsg struct{}
type ViewportPageUpMsg struct{}
type ViewportPageDownMsg struct{}
type ViewportStartMsg struct{}
type ViewportEndMsg struct{}
type ViewportJumpMsg struct{ Index int }

// ViewportResizedMsg notifies viewport of size changes
type ViewportResizedMsg struct {
	Height int
}

// ViewportDataChangedMsg notifies viewport when dataset changes
type ViewportDataChangedMsg struct {
	TotalItems int
}

// ViewportStateChangedMsg is emitted when viewport state changes
type ViewportStateChangedMsg struct {
	State ViewportState
	// What changed flags
	ViewportMoved    bool
	CursorMoved      bool
	ThresholdChanged bool
}

// ================================
// CONSTRUCTOR
// ================================

// NewViewport creates a new viewport component
func NewViewport(config ViewportConfig) *Viewport {
	// Validate and fix config
	ValidateAndFixViewportConfig(&config)

	return &Viewport{
		config: config,
		state: ViewportState{
			ViewportStartIndex:  0,
			CursorIndex:         config.InitialIndex,
			CursorViewportIndex: 0,
			IsAtTopThreshold:    false,
			IsAtBottomThreshold: false,
			AtDatasetStart:      true,
			AtDatasetEnd:        false,
		},
		totalItems: 0,
		debugMode:  false,
	}
}

// SetDebugMode enables/disables debug mode
func (v *Viewport) SetDebugMode(enabled bool) {
	v.debugMode = enabled
}

// ================================
// BUBBLE TEA INTERFACE
// ================================

// Init initializes the viewport
func (v *Viewport) Init() tea.Cmd {
	return nil
}

// Update handles viewport messages
func (v *Viewport) Update(msg tea.Msg) (*Viewport, tea.Cmd) {
	switch msg := msg.(type) {

	case ViewportNavigationMsg:
		return v.handleNavigation(msg)

	// Type-safe specific navigation messages
	case ViewportUpMsg:
		return v.handleNavigation(ViewportNavigationMsg{Type: NavigationUp})
	case ViewportDownMsg:
		return v.handleNavigation(ViewportNavigationMsg{Type: NavigationDown})
	case ViewportPageUpMsg:
		return v.handleNavigation(ViewportNavigationMsg{Type: NavigationPageUp})
	case ViewportPageDownMsg:
		return v.handleNavigation(ViewportNavigationMsg{Type: NavigationPageDown})
	case ViewportStartMsg:
		return v.handleNavigation(ViewportNavigationMsg{Type: NavigationStart})
	case ViewportEndMsg:
		return v.handleNavigation(ViewportNavigationMsg{Type: NavigationEnd})
	case ViewportJumpMsg:
		return v.handleNavigation(ViewportNavigationMsg{Type: NavigationJump, Amount: msg.Index})

	case ViewportResizedMsg:
		return v.handleResize(msg)

	case ViewportDataChangedMsg:
		return v.handleDataChanged(msg)
	}

	return v, nil
}

// View returns viewport debug information (optional)
func (v *Viewport) View() string {
	if !v.debugMode {
		return ""
	}

	return fmt.Sprintf("Viewport[%d:%d] Cursor[%d:%d] T[%t,%t] B[%t,%t]",
		v.state.ViewportStartIndex,
		v.state.ViewportStartIndex+v.config.Height-1,
		v.state.CursorIndex,
		v.state.CursorViewportIndex,
		v.state.IsAtTopThreshold,
		v.state.IsAtBottomThreshold,
		v.state.AtDatasetStart,
		v.state.AtDatasetEnd,
	)
}

// ================================
// PUBLIC INTERFACE
// ================================

// GetState returns current viewport state
func (v *Viewport) GetState() ViewportState {
	return v.state
}

// GetConfig returns viewport configuration
func (v *Viewport) GetConfig() ViewportConfig {
	return v.config
}

// SetTotalItems updates the total dataset size
func (v *Viewport) SetTotalItems(total int) tea.Cmd {
	return ViewportDataChangedCmd(total)
}

// SetCallbacks sets optional event callbacks
func (v *Viewport) SetCallbacks(onViewportChanged func(ViewportState), onCursorChanged func(int)) {
	v.onViewportChanged = onViewportChanged
	v.onCursorChanged = onCursorChanged
}

// GetVisibleRange returns the range of items that should be visible
func (v *Viewport) GetVisibleRange() (start, end int) {
	start = v.state.ViewportStartIndex
	end = start + v.config.Height
	if end > v.totalItems {
		end = v.totalItems
	}
	return start, end
}

// GetThresholdIndices returns the absolute indices of threshold positions
func (v *Viewport) GetThresholdIndices() (topThreshold, bottomThreshold int) {
	topThreshold = -1
	bottomThreshold = -1

	if v.config.TopThreshold >= 0 && v.config.TopThreshold < v.config.Height {
		topThreshold = v.state.ViewportStartIndex + v.config.TopThreshold
	}

	if v.config.BottomThreshold >= 0 && v.config.BottomThreshold < v.config.Height {
		// BottomThreshold is offset from end, so calculate actual position
		bottomPosition := v.config.Height - v.config.BottomThreshold - 1
		bottomThreshold = v.state.ViewportStartIndex + bottomPosition
	}

	return topThreshold, bottomThreshold
}

// ================================
// NAVIGATION HANDLERS
// ================================

// handleNavigation processes navigation messages
func (v *Viewport) handleNavigation(msg ViewportNavigationMsg) (*Viewport, tea.Cmd) {
	if v.totalItems == 0 {
		return v, nil
	}

	previousState := v.state

	switch msg.Type {
	case NavigationUp:
		v.moveUp()
	case NavigationDown:
		v.moveDown()
	case NavigationPageUp:
		v.pageUp()
	case NavigationPageDown:
		v.pageDown()
	case NavigationStart:
		v.jumpToStart()
	case NavigationEnd:
		v.jumpToEnd()
	case NavigationJump:
		v.jumpToIndex(msg.Amount)
	}

	// Check what changed and emit appropriate messages
	var cmds []tea.Cmd

	viewportMoved := v.state.ViewportStartIndex != previousState.ViewportStartIndex
	cursorMoved := v.state.CursorIndex != previousState.CursorIndex
	thresholdChanged := v.state.IsAtTopThreshold != previousState.IsAtTopThreshold ||
		v.state.IsAtBottomThreshold != previousState.IsAtBottomThreshold

	// Emit state changed message
	if viewportMoved || cursorMoved || thresholdChanged {
		cmds = append(cmds, ViewportStateChangedCmd(ViewportStateChangedMsg{
			State:            v.state,
			ViewportMoved:    viewportMoved,
			CursorMoved:      cursorMoved,
			ThresholdChanged: thresholdChanged,
		}))
	}

	// Call callbacks if set
	if v.onViewportChanged != nil && viewportMoved {
		v.onViewportChanged(v.state)
	}
	if v.onCursorChanged != nil && cursorMoved {
		v.onCursorChanged(v.state.CursorIndex)
	}

	return v, tea.Batch(cmds...)
}

// handleResize processes viewport resize
func (v *Viewport) handleResize(msg ViewportResizedMsg) (*Viewport, tea.Cmd) {
	if msg.Height <= 0 {
		return v, nil
	}

	previousHeight := v.config.Height
	v.config.Height = msg.Height

	// Recalculate thresholds for new height
	if v.config.TopThreshold >= msg.Height || v.config.BottomThreshold >= msg.Height {
		v.config.TopThreshold, v.config.BottomThreshold = CalculateThresholds(msg.Height)
	}

	// Adjust viewport if needed
	v.updateViewportBounds()

	// If height changed significantly, emit notification
	if previousHeight != msg.Height {
		return v, ViewportStateChangedCmd(ViewportStateChangedMsg{
			State:            v.state,
			ViewportMoved:    true,
			CursorMoved:      false,
			ThresholdChanged: true,
		})
	}

	return v, nil
}

// handleDataChanged processes dataset changes
func (v *Viewport) handleDataChanged(msg ViewportDataChangedMsg) (*Viewport, tea.Cmd) {
	v.totalItems = msg.TotalItems

	// Ensure cursor doesn't exceed new bounds
	if v.state.CursorIndex >= v.totalItems && v.totalItems > 0 {
		v.state.CursorIndex = v.totalItems - 1
	}

	// Recalculate viewport position
	v.updateViewportPosition()

	return v, ViewportStateChangedCmd(ViewportStateChangedMsg{
		State:            v.state,
		ViewportMoved:    true,
		CursorMoved:      true,
		ThresholdChanged: true,
	})
}

// ================================
// NAVIGATION METHODS (Core Logic)
// ================================

// moveUp moves cursor up one position
func (v *Viewport) moveUp() {
	if v.state.CursorIndex <= 0 {
		return
	}

	v.state.CursorIndex--

	// Handle top threshold logic
	topThreshold := v.config.TopThreshold
	if v.state.IsAtTopThreshold && !v.state.AtDatasetStart && topThreshold >= 0 {
		// At top threshold - scroll viewport up
		v.state.ViewportStartIndex--
		v.state.CursorViewportIndex = topThreshold
	} else if v.state.CursorViewportIndex > 0 {
		// Not at threshold - move cursor within viewport
		v.state.CursorViewportIndex--
	} else {
		// At top of viewport - scroll up
		v.state.ViewportStartIndex--
		v.state.CursorViewportIndex = 0
	}

	v.updateViewportBounds()
}

// moveDown moves cursor down one position
func (v *Viewport) moveDown() {
	if v.state.CursorIndex >= v.totalItems-1 {
		return
	}

	v.state.CursorIndex++

	// Handle bottom threshold logic
	bottomThreshold := v.config.BottomThreshold
	if v.state.IsAtBottomThreshold && !v.state.AtDatasetEnd && bottomThreshold >= 0 {
		// At bottom threshold - scroll viewport down
		v.state.ViewportStartIndex++
		v.state.CursorViewportIndex = bottomThreshold
	} else if v.state.CursorViewportIndex < v.config.Height-1 &&
		v.state.ViewportStartIndex+v.state.CursorViewportIndex+1 < v.totalItems {
		// Not at threshold - move cursor within viewport
		v.state.CursorViewportIndex++
	} else {
		// At bottom of viewport - scroll down
		if v.state.ViewportStartIndex+v.config.Height < v.totalItems {
			v.state.ViewportStartIndex++
		}
		v.state.CursorViewportIndex = v.state.CursorIndex - v.state.ViewportStartIndex
	}

	v.updateViewportBounds()
}

// pageUp moves cursor up by page size
func (v *Viewport) pageUp() {
	if v.state.CursorIndex <= 0 {
		return
	}

	moveCount := v.config.Height
	if moveCount > v.state.CursorIndex {
		moveCount = v.state.CursorIndex
	}

	v.state.CursorIndex -= moveCount
	v.updateViewportPosition()
}

// pageDown moves cursor down by page size
func (v *Viewport) pageDown() {
	if v.state.CursorIndex >= v.totalItems-1 {
		return
	}

	moveCount := v.config.Height
	if v.state.CursorIndex+moveCount >= v.totalItems {
		moveCount = v.totalItems - 1 - v.state.CursorIndex
	}

	v.state.CursorIndex += moveCount
	v.updateViewportPosition()
}

// jumpToStart moves to dataset start
func (v *Viewport) jumpToStart() {
	v.state.CursorIndex = 0
	v.state.ViewportStartIndex = 0
	v.state.CursorViewportIndex = 0
	v.updateViewportBounds()
}

// jumpToEnd moves to dataset end
func (v *Viewport) jumpToEnd() {
	if v.totalItems <= 0 {
		return
	}

	v.state.CursorIndex = v.totalItems - 1

	if v.totalItems <= v.config.Height {
		v.state.ViewportStartIndex = 0
		v.state.CursorViewportIndex = v.totalItems - 1
	} else {
		v.state.ViewportStartIndex = v.totalItems - v.config.Height
		v.state.CursorViewportIndex = v.config.Height - 1
	}

	v.updateViewportBounds()
}

// jumpToIndex moves to specific index
func (v *Viewport) jumpToIndex(index int) {
	if index < 0 || index >= v.totalItems {
		return
	}

	v.state.CursorIndex = index
	v.updateViewportPosition()
}

// ================================
// HELPER METHODS
// ================================

// updateViewportPosition calculates viewport position based on cursor
func (v *Viewport) updateViewportPosition() {
	// Calculate cursor position within viewport
	v.state.CursorViewportIndex = v.state.CursorIndex - v.state.ViewportStartIndex

	// Adjust viewport if cursor is outside
	if v.state.CursorViewportIndex < 0 {
		v.state.ViewportStartIndex = v.state.CursorIndex
		v.state.CursorViewportIndex = 0
	} else if v.state.CursorViewportIndex >= v.config.Height {
		v.state.ViewportStartIndex = v.state.CursorIndex - v.config.Height + 1
		v.state.CursorViewportIndex = v.config.Height - 1
	}

	v.updateViewportBounds()
}

// updateViewportBounds updates threshold and boundary flags
func (v *Viewport) updateViewportBounds() {
	height := v.config.Height
	topThreshold := v.config.TopThreshold
	bottomThreshold := v.config.BottomThreshold

	// Update threshold flags (optional thresholds)
	v.state.IsAtTopThreshold = false
	v.state.IsAtBottomThreshold = false

	if topThreshold >= 0 && topThreshold < height {
		v.state.IsAtTopThreshold = v.state.CursorViewportIndex == topThreshold
	}

	if bottomThreshold >= 0 && bottomThreshold < height {
		v.state.IsAtBottomThreshold = v.state.CursorViewportIndex == bottomThreshold
	}

	// Update dataset boundary flags
	v.state.AtDatasetStart = v.state.ViewportStartIndex == 0
	v.state.AtDatasetEnd = v.state.ViewportStartIndex+height >= v.totalItems

	// Ensure viewport doesn't extend beyond dataset
	maxStart := v.totalItems - height
	if maxStart < 0 {
		maxStart = 0
	}
	if v.state.ViewportStartIndex > maxStart {
		v.state.ViewportStartIndex = maxStart
		v.state.CursorViewportIndex = v.state.CursorIndex - v.state.ViewportStartIndex
	}
}

// ================================
// COMMAND HELPERS
// ================================

// ViewportNavigationCmd creates navigation command
func ViewportNavigationCmd(navType NavigationType, amount int) tea.Cmd {
	return func() tea.Msg {
		return ViewportNavigationMsg{
			Type:   navType,
			Amount: amount,
		}
	}
}

// Type-safe navigation command constructors
func ViewportUpCmd() tea.Cmd {
	return func() tea.Msg { return ViewportUpMsg{} }
}

func ViewportDownCmd() tea.Cmd {
	return func() tea.Msg { return ViewportDownMsg{} }
}

func ViewportPageUpCmd() tea.Cmd {
	return func() tea.Msg { return ViewportPageUpMsg{} }
}

func ViewportPageDownCmd() tea.Cmd {
	return func() tea.Msg { return ViewportPageDownMsg{} }
}

func ViewportStartCmd() tea.Cmd {
	return func() tea.Msg { return ViewportStartMsg{} }
}

func ViewportEndCmd() tea.Cmd {
	return func() tea.Msg { return ViewportEndMsg{} }
}

func ViewportJumpCmd(index int) tea.Cmd {
	return func() tea.Msg { return ViewportJumpMsg{Index: index} }
}

// ViewportHeightResizeCmd creates resize command
func ViewportHeightResizeCmd(height int) tea.Cmd {
	return func() tea.Msg {
		return ViewportResizedMsg{Height: height}
	}
}

// ViewportDataChangedCmd creates data changed command
func ViewportDataChangedCmd(totalItems int) tea.Cmd {
	return func() tea.Msg {
		return ViewportDataChangedMsg{TotalItems: totalItems}
	}
}

// ViewportStateChangedCmd creates state changed command
func ViewportStateChangedCmd(msg ViewportStateChangedMsg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

// ================================
// UTILITY FUNCTIONS
// ================================

// ValidateAndFixViewportConfig validates and fixes viewport configuration
func ValidateAndFixViewportConfig(config *ViewportConfig) {
	// Ensure minimum height
	if config.Height <= 0 {
		config.Height = 10
	}

	// Auto-calculate chunk size if invalid
	if config.ChunkSize <= 0 {
		config.ChunkSize = config.Height * 2
		if config.ChunkSize < 20 {
			config.ChunkSize = 20
		}
	}

	// Ensure initial index is valid
	if config.InitialIndex < 0 {
		config.InitialIndex = 0
	}

	// Auto-calculate thresholds if invalid
	if config.TopThreshold < -1 ||
		config.TopThreshold >= config.Height ||
		config.BottomThreshold < -1 ||
		config.BottomThreshold >= config.Height ||
		(config.BottomThreshold >= 0 && config.TopThreshold >= 0 &&
			config.BottomThreshold <= config.TopThreshold) {

		config.TopThreshold, config.BottomThreshold = CalculateThresholds(config.Height)
	}

	// Set bounding defaults if not configured
	if config.BoundingAreaBefore == 0 {
		config.BoundingAreaBefore = 1
	}
	if config.BoundingAreaAfter == 0 {
		config.BoundingAreaAfter = 2
	}
}

// CalculateThresholds calculates reasonable threshold values
func CalculateThresholds(height int) (topThreshold, bottomThreshold int) {
	if height <= 1 {
		return 0, 0
	}
	if height == 2 {
		return 0, 1
	}
	if height == 3 {
		return 0, 2
	}
	if height <= 5 {
		return 1, height - 2
	}

	// For larger heights, use proportional spacing
	topThreshold = height / 5
	if topThreshold < 1 {
		topThreshold = 1
	}

	bottomThreshold = height - 1 - (height / 5)
	if bottomThreshold <= topThreshold {
		bottomThreshold = topThreshold + 1
	}
	if bottomThreshold >= height {
		bottomThreshold = height - 1
	}

	return topThreshold, bottomThreshold
}
