// Package core provides the fundamental types, interfaces, and messages for the
// vtable library. It defines the shared data structures and contracts used by
// different components like List and Table, ensuring a consistent and
// interoperable architecture. This package is the foundation upon which all other
// vtable modules are built.
package core

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// CursorUpMsg is a message sent to move the cursor up by one position.
type CursorUpMsg struct{}

// CursorDownMsg is a message sent to move the cursor down by one position.
type CursorDownMsg struct{}

// CursorLeftMsg is a message sent to move the cursor left by one position.
type CursorLeftMsg struct{}

// CursorRightMsg is a message sent to move the cursor right by one position.
type CursorRightMsg struct{}

// PageUpMsg is a message sent to move the cursor up by one page/viewport height.
type PageUpMsg struct{}

// PageDownMsg is a message sent to move the cursor down by one page/viewport height.
type PageDownMsg struct{}

// PageLeftMsg is a message sent to move the cursor left by one page/viewport width.
type PageLeftMsg struct{}

// PageRightMsg is a message sent to move the cursor right by one page/viewport width.
type PageRightMsg struct{}

// JumpToStartMsg is a message sent to move the cursor to the first item in the dataset.
type JumpToStartMsg struct{}

// JumpToEndMsg is a message sent to move the cursor to the last item in the dataset.
type JumpToEndMsg struct{}

// JumpToMsg is a message sent to move the cursor to a specific absolute index.
type JumpToMsg struct {
	Index int
}

// TreeJumpToIndexMsg is a message sent to move the cursor to a specific index
// in a tree component, with an option to expand parent nodes to make the target visible.
type TreeJumpToIndexMsg struct {
	Index         int
	ExpandParents bool // If true, expand all parent nodes to make the target item visible
}

// DataRefreshMsg is a message sent to trigger a full refresh of the component's
// data, including reloading the total count and all visible chunks.
type DataRefreshMsg struct{}

// DataChunksRefreshMsg is a message sent to refresh only the currently loaded
// data chunks, preserving the cursor position. This is useful for reflecting
// state changes (like selection) without a full reload.
type DataChunksRefreshMsg struct{}

// DataChunkLoadedMsg is a message sent by a DataSource when a requested chunk
// of data has been successfully loaded.
type DataChunkLoadedMsg struct {
	StartIndex int
	Items      []Data[any]
	Request    DataRequest // The original request, for validation
}

// DataChunkErrorMsg is a message sent by a DataSource when a requested chunk
// of data failed to load.
type DataChunkErrorMsg struct {
	StartIndex int
	Error      error
	Request    DataRequest
}

// DataTotalMsg is a message sent by a DataSource containing the total number of
// items in the dataset.
type DataTotalMsg struct {
	Total int
}

// DataTotalUpdateMsg is a message sent to update the total number of items
// while preserving the current cursor and viewport position.
type DataTotalUpdateMsg struct {
	Total int
}

// DataLoadErrorMsg is a message indicating a general error occurred during
// data loading, not specific to a single chunk.
type DataLoadErrorMsg struct {
	Error error
}

// DataSourceSetMsg is a message sent to replace the component's current
// DataSource with a new one.
type DataSourceSetMsg struct {
	DataSource DataSource[any]
}

// ChunkUnloadedMsg is a message indicating that a data chunk has been unloaded
// from memory, usually as part of a cache-clearing strategy.
type ChunkUnloadedMsg struct {
	ChunkStart int
}

// ChunkLoadingStartedMsg is a message indicating that a request to load a data
// chunk has been initiated. Useful for showing loading indicators.
type ChunkLoadingStartedMsg struct {
	ChunkStart int
	Request    DataRequest
}

// ChunkLoadingCompletedMsg is a message indicating that a data chunk has
// finished loading, successfully or not.
type ChunkLoadingCompletedMsg struct {
	ChunkStart int
	ItemCount  int
	Request    DataRequest
}

// DataTotalRequestMsg is a message sent to explicitly request the total item
// count from the DataSource.
type DataTotalRequestMsg struct{}

// SelectCurrentMsg is a message to select or toggle the item at the current
// cursor position.
type SelectCurrentMsg struct{}

// SelectToggleMsg is a message to toggle the selection state of an item at a
// specific index.
type SelectToggleMsg struct {
	Index int
}

// SelectAllMsg is a message to select all items in the dataset.
type SelectAllMsg struct{}

// SelectClearMsg is a message to clear all current selections.
type SelectClearMsg struct{}

// SelectRangeMsg is a message to select a range of items between two item IDs.
type SelectRangeMsg struct {
	StartID string
	EndID   string
}

// SelectionModeSetMsg is a message to change the component's selection mode
// (e.g., single, multiple, none).
type SelectionModeSetMsg struct {
	Mode SelectionMode
}

// SelectionResponseMsg is a message from a DataSource indicating the result of
// a selection operation.
type SelectionResponseMsg struct {
	Success     bool
	Index       int
	ID          string
	Selected    bool
	Operation   string // e.g., "toggle", "selectAll", "clear"
	Error       error
	AffectedIDs []string // For operations that affect multiple items
}

// SelectionChangedMsg is a message indicating that the selection state has
// changed within the data source.
type SelectionChangedMsg struct {
	SelectedIndices []int
	SelectedIDs     []string
	TotalSelected   int
}

// FilterSetMsg is a message to apply or update a filter on a specific data field.
type FilterSetMsg struct {
	Field string
	Value any
}

// FilterClearMsg is a message to remove a filter from a specific data field.
type FilterClearMsg struct {
	Field string
}

// FiltersClearAllMsg is a message to remove all active filters.
type FiltersClearAllMsg struct{}

// SortToggleMsg is a message to toggle the sort order of a field (e.g., asc ->
// desc -> none).
type SortToggleMsg struct {
	Field string
}

// SortSetMsg is a message to set the sorting for a specific field and direction.
type SortSetMsg struct {
	Field     string
	Direction string // "asc" or "desc"
}

// SortAddMsg is a message to add a field to a multi-column sort configuration.
type SortAddMsg struct {
	Field     string
	Direction string
}

// SortRemoveMsg is a message to remove a field from the sorting configuration.
type SortRemoveMsg struct {
	Field string
}

// SortsClearAllMsg is a message to clear all sorting configurations.
type SortsClearAllMsg struct{}

// FocusMsg is a message to give focus to the component, making it active.
type FocusMsg struct{}

// BlurMsg is a message to remove focus from the component, making it inactive.
type BlurMsg struct{}

// GlobalAnimationTickMsg is a message sent periodically by the animation engine
// to drive time-based animations.
type GlobalAnimationTickMsg struct {
	Timestamp time.Time
}

// AnimationUpdateMsg is a message indicating that one or more animations have
// been updated and may require a re-render.
type AnimationUpdateMsg struct {
	UpdatedAnimations []string
}

// AnimationConfigMsg is a message to set a new configuration for the animation
// engine.
type AnimationConfigMsg struct {
	Config AnimationConfig
}

// AnimationStartMsg is a message to start a specific registered animation.
type AnimationStartMsg struct {
	AnimationID string
}

// AnimationStopMsg is a message to stop a specific running animation.
type AnimationStopMsg struct {
	AnimationID string
}

// ThemeSetMsg is a message to apply a new theme or style configuration to a
// component.
type ThemeSetMsg struct {
	Theme interface{} // Can be Theme or StyleConfig
}

// RealTimeUpdateMsg is a message to trigger a real-time update of the component.
type RealTimeUpdateMsg struct{}

// RealTimeConfigMsg is a message to configure real-time update behavior.
type RealTimeConfigMsg struct {
	Enabled  bool
	Interval time.Duration
}

// ViewportResizeMsg is a message indicating that the component's available
// size has changed.
type ViewportResizeMsg struct {
	Width  int
	Height int
}

// ViewportConfigMsg is a message to apply a new ViewportConfig.
type ViewportConfigMsg struct {
	Config ViewportConfig
}

// ColumnSetMsg is a message to set the columns for a table component.
type ColumnSetMsg struct {
	Columns []TableColumn
}

// ColumnUpdateMsg is a message to update the configuration of a single table column.
type ColumnUpdateMsg struct {
	Index  int
	Column TableColumn
}

// HeaderVisibilityMsg is a message to set the visibility of the table header.
type HeaderVisibilityMsg struct {
	Visible bool
}

// BorderVisibilityMsg is a message to set the visibility of table borders.
type BorderVisibilityMsg struct {
	Visible bool
}

// TopBorderVisibilityMsg is a message to set the visibility of the table's top border.
type TopBorderVisibilityMsg struct {
	Visible bool
}

// BottomBorderVisibilityMsg is a message to set the visibility of the table's bottom border.
type BottomBorderVisibilityMsg struct {
	Visible bool
}

// HeaderSeparatorVisibilityMsg is a message to set the visibility of the
// separator line between the header and the table body.
type HeaderSeparatorVisibilityMsg struct {
	Visible bool
}

// TopBorderSpaceRemovalMsg is a message to control whether the space for the
// top border is completely removed when it's not visible.
type TopBorderSpaceRemovalMsg struct {
	Remove bool
}

// BottomBorderSpaceRemovalMsg is a message to control whether the space for the
// bottom border is completely removed when it's not visible.
type BottomBorderSpaceRemovalMsg struct {
	Remove bool
}

// CellFormatterSetMsg is a message to set a custom cell formatter for a table column.
type CellFormatterSetMsg struct {
	ColumnIndex int // -1 applies to all columns
	Formatter   SimpleCellFormatter
}

// CellAnimatedFormatterSetMsg is a message to set a custom animated cell
// formatter for a table column.
type CellAnimatedFormatterSetMsg struct {
	ColumnIndex int
	Formatter   CellFormatterAnimated
}

// RowFormatterSetMsg is a message to set a custom formatter for loading placeholder rows.
type RowFormatterSetMsg struct {
	Formatter LoadingRowFormatter
}

// HeaderFormatterSetMsg is a message to set a custom header formatter for a table column.
type HeaderFormatterSetMsg struct {
	ColumnIndex int
	Formatter   SimpleHeaderFormatter
}

// LoadingFormatterSetMsg is a message to set a custom loading row formatter.
//
// Deprecated: Use RowFormatterSetMsg instead.
type LoadingFormatterSetMsg struct {
	Formatter LoadingRowFormatter
}

// HeaderCellFormatterSetMsg is a message to set a custom header cell formatter.
//
// Deprecated: Use HeaderFormatterSetMsg instead.
type HeaderCellFormatterSetMsg struct {
	Formatter HeaderCellFormatter
}

// ColumnConstraintsSetMsg is a message to set layout constraints for a table column.
type ColumnConstraintsSetMsg struct {
	ColumnIndex int
	Constraints CellConstraint
}

// TableThemeSetMsg is a message to apply a new theme to a table component.
type TableThemeSetMsg struct {
	Theme Theme
}

// FormatterSetMsg is a message to set a custom item formatter for a list component.
type FormatterSetMsg struct {
	Formatter ItemFormatter[any]
}

// AnimatedFormatterSetMsg is a message to set a custom animated item formatter
// for a list component.
type AnimatedFormatterSetMsg struct {
	Formatter ItemFormatterAnimated[any]
}

// ChunkSizeSetMsg is a message to set the chunk size for data loading.
type ChunkSizeSetMsg struct {
	Size int
}

// MaxWidthSetMsg is a message to set the maximum width for a list component.
type MaxWidthSetMsg struct {
	Width int
}

// StyleConfigSetMsg is a message to apply a new style configuration to a list component.
type StyleConfigSetMsg struct {
	Config StyleConfig
}

// CellAnimationStartMsg is a message to start an animation for a specific table cell.
type CellAnimationStartMsg struct {
	RowID       string
	ColumnIndex int
	Animation   CellAnimation
}

// CellAnimationStopMsg is a message to stop an animation for a specific table cell.
type CellAnimationStopMsg struct {
	RowID       string
	ColumnIndex int
}

// RowAnimationStartMsg is a message to start an animation for a specific table row.
type RowAnimationStartMsg struct {
	RowID     string
	Animation RowAnimation
}

// RowAnimationStopMsg is a message to stop an animation for a specific table row.
type RowAnimationStopMsg struct {
	RowID string
}

// ItemAnimationStartMsg is a message to start an animation for a specific list item.
type ItemAnimationStartMsg struct {
	ItemID    string
	Animation ListAnimation
}

// ItemAnimationStopMsg is a message to stop an animation for a specific list item.
type ItemAnimationStopMsg struct {
	ItemID string
}

// KeyMapSetMsg is a message to apply a new key map for navigation and actions.
type KeyMapSetMsg struct {
	KeyMap NavigationKeyMap
}

// PerformanceConfigMsg is a message to configure performance monitoring.
type PerformanceConfigMsg struct {
	Enabled           bool
	MonitorMemory     bool
	MonitorRenderTime bool
	ReportInterval    time.Duration
}

// DebugEnableMsg is a message to enable or disable debugging features.
type DebugEnableMsg struct {
	Enabled bool
}

// DebugLevelSetMsg is a message to set the verbosity level for debugging output.
type DebugLevelSetMsg struct {
	Level DebugLevel
}

// ErrorMsg is a message representing a generic error that has occurred.
type ErrorMsg struct {
	Error   error
	Context string
}

// ValidationErrorMsg is a message representing a validation error for a
// specific field.
type ValidationErrorMsg struct {
	Field   string
	Value   any
	Error   error
	Context string
}

// StatusMsg is a message used to display status information to the user.
type StatusMsg struct {
	Message string
	Type    StatusType
}

// StatusType defines the category of a status message.
type StatusType int

// Constants for different status message types.
const (
	StatusInfo StatusType = iota
	StatusWarning
	StatusError
	StatusSuccess
)

// SearchSetMsg is a message to initiate a search with a given query.
type SearchSetMsg struct {
	Query string
	Field string // Optional: search within a specific field
}

// SearchClearMsg is a message to clear the current search query and results.
type SearchClearMsg struct{}

// SearchResultMsg is a message containing the results of a search operation.
type SearchResultMsg struct {
	Results []int // Indices of matching items
	Query   string
	Total   int
}

// AccessibilityConfigMsg is a message to configure accessibility features.
type AccessibilityConfigMsg struct {
	ScreenReader  bool
	HighContrast  bool
	ReducedMotion bool
}

// AriaLabelSetMsg is a message to set the ARIA label for a component, improving
// screen reader support.
type AriaLabelSetMsg struct {
	Label string
}

// DescriptionSetMsg is a message to set the accessible description for a component.
type DescriptionSetMsg struct {
	Description string
}

// BatchMsg is a message that wraps multiple other messages, allowing them to be
// processed in a single update cycle.
type BatchMsg struct {
	Messages []interface{}
}

// InitMsg is a message to trigger the initial state setup of a component.
type InitMsg struct{}

// DestroyMsg is a message to trigger the cleanup and resource release of a component.
type DestroyMsg struct{}

// ResetMsg is a message to reset a component to its initial state.
type ResetMsg struct{}

// Batch creates a new BatchMsg from a variadic list of messages.
func Batch(messages ...interface{}) BatchMsg {
	return BatchMsg{Messages: messages}
}

// ActiveCellIndicationModeSetMsg is a message to set the active cell
// indication mode for tables.
type ActiveCellIndicationModeSetMsg struct {
	Enabled bool // Simple boolean: enabled or disabled
}

// ActiveCellBackgroundColorSetMsg is a message to set the background color of
// the active cell in a table.
type ActiveCellBackgroundColorSetMsg struct {
	Color string // lipgloss color value
}

// SetFullRowSelectionMsg is a message to enable/disable full row selection background styling
type SetFullRowSelectionMsg struct {
	Enabled    bool
	Background lipgloss.Style
}

// SetCursorRowStylingMsg is a message to enable/disable full row cursor background styling
type SetCursorRowStylingMsg struct {
	Enabled    bool
	Background lipgloss.Style
}

// SetComponentBackgroundMsg is a message to configure background styling for a specific component
type SetComponentBackgroundMsg struct {
	ComponentType ListComponentType
	CursorBg      lipgloss.Style
	SelectedBg    lipgloss.Style
	NormalBg      lipgloss.Style
	ApplyCursor   bool
	ApplySelected bool
	ApplyNormal   bool
}

// === HORIZONTAL SCROLLING MESSAGES ===

// HorizontalScrollLeftMsg is a message sent to scroll horizontally left within the current column.
type HorizontalScrollLeftMsg struct{}

// HorizontalScrollRightMsg is a message sent to scroll horizontally right within the current column.
type HorizontalScrollRightMsg struct{}

// HorizontalScrollWordLeftMsg is a message sent to scroll horizontally left by word boundaries.
type HorizontalScrollWordLeftMsg struct{}

// HorizontalScrollWordRightMsg is a message sent to scroll horizontally right by word boundaries.
type HorizontalScrollWordRightMsg struct{}

// HorizontalScrollSmartLeftMsg is a message sent to scroll horizontally left using smart boundaries.
type HorizontalScrollSmartLeftMsg struct{}

// HorizontalScrollSmartRightMsg is a message sent to scroll horizontally right using smart boundaries.
type HorizontalScrollSmartRightMsg struct{}

// HorizontalScrollPageLeftMsg is a message sent to scroll horizontally left by a page amount.
type HorizontalScrollPageLeftMsg struct{}

// HorizontalScrollPageRightMsg is a message sent to scroll horizontally right by a page amount.
type HorizontalScrollPageRightMsg struct{}

// HorizontalScrollModeToggleMsg is a message sent to cycle through horizontal scroll modes (character/word/smart).
type HorizontalScrollModeToggleMsg struct{}

// HorizontalScrollScopeToggleMsg is a message sent to toggle horizontal scroll scope (current row/all rows).
type HorizontalScrollScopeToggleMsg struct{}

// HorizontalScrollResetMsg is a message sent to reset all horizontal scroll offsets.
type HorizontalScrollResetMsg struct{}

// === COLUMN NAVIGATION MESSAGES ===

// NextColumnMsg is a message sent to move to the next column for horizontal navigation/scrolling focus.
type NextColumnMsg struct{}

// PrevColumnMsg is a message sent to move to the previous column for horizontal navigation/scrolling focus.
type PrevColumnMsg struct{}
