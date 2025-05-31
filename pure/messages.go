package vtable

import (
	"time"
)

// ================================
// SHARED MESSAGES (Both List & Table)
// ================================

// ===== Navigation Messages =====

// CursorUpMsg moves the cursor up one position
type CursorUpMsg struct{}

// CursorDownMsg moves the cursor down one position
type CursorDownMsg struct{}

// PageUpMsg moves the cursor up one page
type PageUpMsg struct{}

// PageDownMsg moves the cursor down one page
type PageDownMsg struct{}

// JumpToStartMsg moves the cursor to the first item
type JumpToStartMsg struct{}

// JumpToEndMsg moves the cursor to the last item
type JumpToEndMsg struct{}

// JumpToMsg moves the cursor to a specific index
type JumpToMsg struct {
	Index int
}

// TreeJumpToIndexMsg moves the cursor to a specific index in a tree, expanding parent nodes as needed
type TreeJumpToIndexMsg struct {
	Index         int
	ExpandParents bool // If true, expand all parent nodes to make the target item visible
}

// ===== Data Messages =====

// DataRefreshMsg triggers a complete data refresh
type DataRefreshMsg struct{}

// DataChunksRefreshMsg triggers chunk refresh while preserving cursor position
type DataChunksRefreshMsg struct{}

// DataChunkLoadedMsg indicates that a chunk of data has been loaded
type DataChunkLoadedMsg struct {
	StartIndex int
	Items      []Data[any]
	Request    DataRequest // For validation
}

// DataChunkErrorMsg indicates that a chunk failed to load
type DataChunkErrorMsg struct {
	StartIndex int
	Error      error
	Request    DataRequest
}

// DataTotalMsg provides the total number of items
type DataTotalMsg struct {
	Total int
}

// DataTotalUpdateMsg updates the total number of items while preserving cursor position
type DataTotalUpdateMsg struct {
	Total int
}

// DataLoadErrorMsg indicates a general data loading error
type DataLoadErrorMsg struct {
	Error error
}

// DataSourceSetMsg sets a new data source
type DataSourceSetMsg struct {
	DataSource DataSource[any]
}

// ChunkUnloadedMsg indicates that a chunk was unloaded from memory
type ChunkUnloadedMsg struct {
	ChunkStart int
}

// ChunkLoadingStartedMsg indicates that a chunk has started loading
type ChunkLoadingStartedMsg struct {
	ChunkStart int
	Request    DataRequest
}

// ChunkLoadingCompletedMsg indicates that a chunk has finished loading
type ChunkLoadingCompletedMsg struct {
	ChunkStart int
	ItemCount  int
	Request    DataRequest
}

// DataTotalRequestMsg requests the total count of items from the data source
type DataTotalRequestMsg struct{}

// ===== Selection Messages =====

// SelectCurrentMsg selects the item at the current cursor position
type SelectCurrentMsg struct{}

// SelectToggleMsg toggles selection for a specific item
type SelectToggleMsg struct {
	Index int
}

// SelectAllMsg selects all items
type SelectAllMsg struct{}

// SelectClearMsg clears all selections
type SelectClearMsg struct{}

// SelectRangeMsg selects a range of items
type SelectRangeMsg struct {
	StartID string
	EndID   string
}

// SelectionModeSetMsg sets the selection mode
type SelectionModeSetMsg struct {
	Mode SelectionMode
}

// SelectionResponseMsg indicates the result of a selection operation
type SelectionResponseMsg struct {
	Success     bool
	Index       int
	ID          string
	Selected    bool
	Operation   string // "toggle", "select", "deselect", "selectAll", "clear", "range"
	Error       error
	AffectedIDs []string // For operations that affect multiple items
}

// SelectionChangedMsg indicates that selection state has changed in the data source
type SelectionChangedMsg struct {
	SelectedIndices []int
	SelectedIDs     []string
	TotalSelected   int
}

// ===== Filter Messages =====

// FilterSetMsg sets a filter on a specific field
type FilterSetMsg struct {
	Field string
	Value any
}

// FilterClearMsg clears a filter on a specific field
type FilterClearMsg struct {
	Field string
}

// FiltersClearAllMsg clears all filters
type FiltersClearAllMsg struct{}

// ===== Sort Messages =====

// SortToggleMsg toggles sorting on a field
type SortToggleMsg struct {
	Field string
}

// SortSetMsg sets sorting on a field with a specific direction
type SortSetMsg struct {
	Field     string
	Direction string // "asc" or "desc"
}

// SortAddMsg adds a sort field to multi-column sorting
type SortAddMsg struct {
	Field     string
	Direction string
}

// SortRemoveMsg removes a field from sorting
type SortRemoveMsg struct {
	Field string
}

// SortsClearAllMsg clears all sorting
type SortsClearAllMsg struct{}

// ===== Focus Messages =====

// FocusMsg gives focus to the component
type FocusMsg struct{}

// BlurMsg removes focus from the component
type BlurMsg struct{}

// ===== Animation Messages =====

// GlobalAnimationTickMsg is sent periodically to update all animations
type GlobalAnimationTickMsg struct {
	Timestamp time.Time
}

// AnimationUpdateMsg indicates that animations have been updated
type AnimationUpdateMsg struct {
	UpdatedAnimations []string
}

// AnimationConfigMsg sets the animation configuration
type AnimationConfigMsg struct {
	Config AnimationConfig
}

// AnimationStartMsg starts a specific animation
type AnimationStartMsg struct {
	AnimationID string
}

// AnimationStopMsg stops a specific animation
type AnimationStopMsg struct {
	AnimationID string
}

// ===== Theme Messages =====

// ThemeSetMsg sets the theme
type ThemeSetMsg struct {
	Theme interface{} // Theme or StyleConfig
}

// ===== Real-time Update Messages =====

// RealTimeUpdateMsg triggers a real-time update
type RealTimeUpdateMsg struct{}

// RealTimeConfigMsg configures real-time updates
type RealTimeConfigMsg struct {
	Enabled  bool
	Interval time.Duration
}

// ===== Viewport Messages =====

// ViewportResizeMsg indicates that the viewport has been resized
type ViewportResizeMsg struct {
	Width  int
	Height int
}

// ViewportConfigMsg sets viewport configuration
type ViewportConfigMsg struct {
	Config ViewportConfig
}

// ================================
// TABLE-SPECIFIC MESSAGES
// ================================

// ===== Column Messages =====

// ColumnSetMsg sets the table columns
type ColumnSetMsg struct {
	Columns []TableColumn
}

// ColumnUpdateMsg updates a specific column
type ColumnUpdateMsg struct {
	Index  int
	Column TableColumn
}

// ===== Header & Border Messages =====

// HeaderVisibilityMsg sets header visibility
type HeaderVisibilityMsg struct {
	Visible bool
}

// BorderVisibilityMsg sets border visibility
type BorderVisibilityMsg struct {
	Visible bool
}

// TopBorderVisibilityMsg sets top border visibility
type TopBorderVisibilityMsg struct {
	Visible bool
}

// BottomBorderVisibilityMsg sets bottom border visibility
type BottomBorderVisibilityMsg struct {
	Visible bool
}

// HeaderSeparatorVisibilityMsg sets header separator visibility
type HeaderSeparatorVisibilityMsg struct {
	Visible bool
}

// TopBorderSpaceRemovalMsg controls whether top border space is completely removed
type TopBorderSpaceRemovalMsg struct {
	Remove bool
}

// BottomBorderSpaceRemovalMsg controls whether bottom border space is completely removed
type BottomBorderSpaceRemovalMsg struct {
	Remove bool
}

// ===== Formatter Messages =====

// CellFormatterSetMsg sets a cell formatter for a specific column
type CellFormatterSetMsg struct {
	ColumnIndex int // -1 for all columns
	Formatter   SimpleCellFormatter
}

// CellAnimatedFormatterSetMsg sets an animated cell formatter
type CellAnimatedFormatterSetMsg struct {
	ColumnIndex int
	Formatter   CellFormatterAnimated
}

// RowFormatterSetMsg sets the loading row formatter
type RowFormatterSetMsg struct {
	Formatter LoadingRowFormatter
}

// HeaderFormatterSetMsg sets the header formatter for a specific column
type HeaderFormatterSetMsg struct {
	ColumnIndex int
	Formatter   SimpleHeaderFormatter
}

// LoadingFormatterSetMsg sets the loading row formatter (DEPRECATED - use RowFormatterSetMsg)
type LoadingFormatterSetMsg struct {
	Formatter LoadingRowFormatter
}

// HeaderCellFormatterSetMsg sets the header cell formatter (DEPRECATED - use HeaderFormatterSetMsg)
type HeaderCellFormatterSetMsg struct {
	Formatter HeaderCellFormatter
}

// ===== Constraint Messages =====

// ColumnConstraintsSetMsg sets constraints for a column
type ColumnConstraintsSetMsg struct {
	ColumnIndex int
	Constraints CellConstraint
}

// ===== Table Theme Messages =====

// TableThemeSetMsg sets the table theme
type TableThemeSetMsg struct {
	Theme Theme
}

// ================================
// LIST-SPECIFIC MESSAGES
// ================================

// ===== List Formatter Messages =====

// FormatterSetMsg sets the list item formatter
type FormatterSetMsg struct {
	Formatter ItemFormatter[any]
}

// AnimatedFormatterSetMsg sets the animated list item formatter
type AnimatedFormatterSetMsg struct {
	Formatter ItemFormatterAnimated[any]
}

// ===== List Configuration Messages =====

// ChunkSizeSetMsg sets the chunk size for data loading
type ChunkSizeSetMsg struct {
	Size int
}

// MaxWidthSetMsg sets the maximum width for list items
type MaxWidthSetMsg struct {
	Width int
}

// ===== List Style Messages =====

// StyleConfigSetMsg sets the style configuration for lists
type StyleConfigSetMsg struct {
	Config StyleConfig
}

// ================================
// ANIMATION CONTROL MESSAGES
// ================================

// ===== Cell Animation Messages =====

// CellAnimationStartMsg starts an animation for a specific cell
type CellAnimationStartMsg struct {
	RowID       string
	ColumnIndex int
	Animation   CellAnimation
}

// CellAnimationStopMsg stops an animation for a specific cell
type CellAnimationStopMsg struct {
	RowID       string
	ColumnIndex int
}

// ===== Row Animation Messages =====

// RowAnimationStartMsg starts an animation for a specific row
type RowAnimationStartMsg struct {
	RowID     string
	Animation RowAnimation
}

// RowAnimationStopMsg stops an animation for a specific row
type RowAnimationStopMsg struct {
	RowID string
}

// ===== Item Animation Messages =====

// ItemAnimationStartMsg starts an animation for a specific list item
type ItemAnimationStartMsg struct {
	ItemID    string
	Animation ListAnimation
}

// ItemAnimationStopMsg stops an animation for a specific list item
type ItemAnimationStopMsg struct {
	ItemID string
}

// ================================
// CONFIGURATION MESSAGES
// ================================

// ===== Key Binding Messages =====

// KeyMapSetMsg sets the key map configuration
type KeyMapSetMsg struct {
	KeyMap NavigationKeyMap
}

// ===== Performance Messages =====

// PerformanceConfigMsg configures performance monitoring
type PerformanceConfigMsg struct {
	Enabled           bool
	MonitorMemory     bool
	MonitorRenderTime bool
	ReportInterval    time.Duration
}

// ===== Debug Messages =====

// DebugEnableMsg enables or disables debugging
type DebugEnableMsg struct {
	Enabled bool
}

// DebugLevelSetMsg sets the debug level
type DebugLevelSetMsg struct {
	Level DebugLevel
}

// ================================
// ERROR MESSAGES
// ================================

// ErrorMsg represents a general error
type ErrorMsg struct {
	Error   error
	Context string
}

// ValidationErrorMsg represents a validation error
type ValidationErrorMsg struct {
	Field   string
	Value   any
	Error   error
	Context string
}

// ================================
// STATUS MESSAGES
// ================================

// StatusMsg provides status information
type StatusMsg struct {
	Message string
	Type    StatusType
}

// StatusType defines different types of status messages
type StatusType int

const (
	StatusInfo StatusType = iota
	StatusWarning
	StatusError
	StatusSuccess
)

// ================================
// SEARCH MESSAGES
// ================================

// SearchSetMsg sets a search query
type SearchSetMsg struct {
	Query string
	Field string // Optional: search in specific field
}

// SearchClearMsg clears the search
type SearchClearMsg struct{}

// SearchResultMsg provides search results
type SearchResultMsg struct {
	Results []int // Indices of matching items
	Query   string
	Total   int
}

// ================================
// ACCESSIBILITY MESSAGES
// ================================

// AccessibilityConfigMsg configures accessibility features
type AccessibilityConfigMsg struct {
	ScreenReader  bool
	HighContrast  bool
	ReducedMotion bool
}

// AriaLabelSetMsg sets the ARIA label
type AriaLabelSetMsg struct {
	Label string
}

// DescriptionSetMsg sets the description
type DescriptionSetMsg struct {
	Description string
}

// ================================
// BATCH MESSAGES
// ================================

// BatchMsg allows multiple messages to be sent at once
type BatchMsg struct {
	Messages []interface{}
}

// ================================
// LIFECYCLE MESSAGES
// ================================

// InitMsg initializes the component
type InitMsg struct{}

// DestroyMsg destroys the component and cleans up resources
type DestroyMsg struct{}

// ResetMsg resets the component to its initial state
type ResetMsg struct{}

// ================================
// UTILITY FUNCTIONS
// ================================

// Batch creates a BatchMsg from multiple messages
func Batch(messages ...interface{}) BatchMsg {
	return BatchMsg{Messages: messages}
}

// ActiveCellIndicationModeSetMsg sets the active cell indication mode
type ActiveCellIndicationModeSetMsg struct {
	Enabled bool // Simple boolean: enabled or disabled
}

// ActiveCellBackgroundColorSetMsg sets the active cell background color
type ActiveCellBackgroundColorSetMsg struct {
	Color string // lipgloss color value
}
