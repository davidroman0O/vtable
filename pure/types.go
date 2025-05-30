package vtable

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ================================
// CORE DATA TYPES
// ================================

// TableRow represents a row of data in the table.
type TableRow struct {
	// ID is a stable unique identifier for this row
	ID string

	// Cells contains the string values for each column in the row.
	Cells []string
}

// TableColumn represents a column configuration in a table.
type TableColumn struct {
	// Title is the column header text
	Title string

	// Width is the column width in characters
	Width int

	// Alignment defines how text is aligned in the column cells (left, right, center)
	Alignment int

	// Field is the identifier used for sorting/filtering operations
	Field string

	// Header configuration (independent from cell alignment)
	HeaderAlignment  int            // Alignment for the header text (can be different from Alignment)
	HeaderConstraint CellConstraint // Constraints for header cell formatting
}

// Data wraps an item with its state and metadata for rendering
type Data[T any] struct {
	// ID is a stable unique identifier for this item
	ID string

	// Item is the actual data item
	Item T

	// Selected indicates if this item is selected
	Selected bool

	// Metadata contains custom rendering metadata with type safety
	Metadata TypedMetadata

	// Disabled indicates if this item should be rendered as disabled
	Disabled bool

	// Hidden indicates if this item should be hidden from view
	Hidden bool

	// Error contains any error state for this item
	Error error

	// Loading indicates if this item is currently loading
	Loading bool
}

// ================================
// VIEWPORT & NAVIGATION TYPES
// ================================

// ViewportState contains component viewport information
type ViewportState struct {
	// ViewportStartIndex is the absolute index of the first item in the viewport.
	ViewportStartIndex int

	// CursorIndex is the absolute index of the selected item in the dataset.
	CursorIndex int

	// CursorViewportIndex is the relative index of the cursor within the viewport.
	CursorViewportIndex int

	// IsAtTopThreshold indicates if the cursor is at the top threshold.
	IsAtTopThreshold bool

	// IsAtBottomThreshold indicates if the cursor is at the bottom threshold.
	IsAtBottomThreshold bool

	// AtDatasetStart indicates if the viewport is at the start of the dataset.
	AtDatasetStart bool

	// AtDatasetEnd indicates if the viewport is at the end of the dataset.
	AtDatasetEnd bool
}

// ViewportConfig defines the configuration for the viewport.
type ViewportConfig struct {
	// Height is the number of items visible in the viewport
	Height int

	// TopThreshold is the offset from viewport start where scrolling up triggers
	TopThreshold int

	// BottomThreshold is the offset from viewport end where scrolling down triggers
	BottomThreshold int

	// ChunkSize is the number of items to load in each chunk
	ChunkSize int

	// InitialIndex is the starting cursor position
	InitialIndex int

	// BoundingAreaBefore is the number of items to keep loaded before the viewport top
	BoundingAreaBefore int

	// BoundingAreaAfter is the number of items to keep loaded after the viewport bottom
	BoundingAreaAfter int
}

// ================================
// DATA REQUEST & FILTERING
// ================================

// DataRequest represents a request for data with optional filtering and sorting.
type DataRequest struct {
	// Start is the index of the first item to return
	Start int

	// Count is the number of items to return
	Count int

	// SortFields specifies the fields to sort by in order of priority
	SortFields []string

	// SortDirections specifies the sort directions ("asc" or "desc") corresponding to SortFields
	SortDirections []string

	// Filters is a map of field names to filter values
	Filters map[string]any
}

// ================================
// CHUNK MANAGEMENT
// ================================

// Chunk represents a chunk of data loaded from the data source.
type Chunk[T any] struct {
	// StartIndex is the absolute index of the first item in the chunk.
	StartIndex int

	// EndIndex is the absolute index of the last item in the chunk.
	EndIndex int

	// Items is the slice of Data items in the chunk.
	Items []Data[T]

	// LoadedAt is when this chunk was loaded
	LoadedAt time.Time

	// Request is the DataRequest that created this chunk (for validation)
	Request DataRequest
}

// ChunkInfo provides information about a loaded chunk.
type ChunkInfo struct {
	// StartIndex is the absolute index of the first item in the chunk.
	StartIndex int

	// EndIndex is the absolute index of the last item in the chunk.
	EndIndex int

	// ItemCount is the number of items in the chunk.
	ItemCount int
}

// ================================
// SELECTION TYPES
// ================================

// SelectionMode defines how selection behaves
type SelectionMode int

const (
	// SelectionSingle allows only one item to be selected at a time
	SelectionSingle SelectionMode = iota
	// SelectionMultiple allows multiple items to be selected
	SelectionMultiple
	// SelectionNone disables selection
	SelectionNone
)

// ================================
// METADATA SYSTEM
// ================================

// MetadataKey represents a type-safe key for metadata values
type MetadataKey[T any] struct {
	Key          string
	DefaultValue T
	Validator    func(T) error
}

// TypedMetadata provides type-safe metadata operations
type TypedMetadata struct {
	data map[string]any
}

// NewTypedMetadata creates a new TypedMetadata instance
func NewTypedMetadata() TypedMetadata {
	return TypedMetadata{
		data: make(map[string]any),
	}
}

// Common metadata keys
var (
	StatusColorKey = MetadataKey[string]{"status_color", "#ffffff", nil}
	PriorityKey    = MetadataKey[int]{"priority", 0, nil}
	TooltipKey     = MetadataKey[string]{"tooltip", "", nil}
	IconKey        = MetadataKey[string]{"icon", "", nil}
	BadgeKey       = MetadataKey[string]{"badge", "", nil}
)

// ================================
// FOCUS MANAGEMENT
// ================================

// FocusState contains component focus information
type FocusState struct {
	HasFocus    bool
	FocusedCell string // ID of focused cell (for tables)
}

// ================================
// RENDERING TYPES
// ================================

// RenderContext provides dimensional constraints and utilities for formatting
type RenderContext struct {
	// Dimensional constraints
	MaxWidth  int
	MaxHeight int

	// Component context
	ColumnIndex int

	// Styling & theming
	Theme     *Theme
	BaseStyle lipgloss.Style

	// Terminal capabilities
	ColorSupport   bool
	UnicodeSupport bool

	// Accessibility
	HighContrast  bool
	ReducedMotion bool
	ScreenReader  bool

	// Global state
	CurrentTime time.Time
	FocusState  FocusState
	DeltaTime   time.Duration // Time since last render for smooth animations

	// State indicators (configurable)
	ErrorIndicator    string
	LoadingIndicator  string
	DisabledIndicator string
	SelectedIndicator string

	// Utility functions
	Truncate func(string, int) string
	Wrap     func(string, int) []string
	Measure  func(string, int) (int, int)

	// Error handling
	OnError func(error)
}

// RenderResult contains the result of an animated rendering operation
type RenderResult struct {
	// Content is the rendered string
	Content string

	// RefreshTriggers specify when this content should be re-rendered
	RefreshTriggers []RefreshTrigger

	// AnimationState stores state between renders
	AnimationState map[string]any

	// Error contains any rendering error
	Error error

	// Fallback content to use if there's an error
	Fallback string
}

// CellRenderResult contains the result of cell rendering with constraints
type CellRenderResult struct {
	// Content is the rendered cell content
	Content string

	// ActualWidth is the actual width used
	ActualWidth int

	// ActualHeight is the actual height used
	ActualHeight int

	// Overflow indicates whether content was truncated
	Overflow bool

	// RefreshTriggers specify when this cell should be re-rendered
	RefreshTriggers []RefreshTrigger

	// AnimationState stores state between renders for this specific cell
	AnimationState map[string]any

	// Error contains any rendering error
	Error error

	// Fallback content to use if there's an error
	Fallback string
}

// ================================
// CONSTRAINT MANAGEMENT
// ================================

// CellConstraint represents the constraints for a cell
type CellConstraint struct {
	Width     int           // Exact width (enforced)
	Height    int           // Usually 1 for tables (multi-line support future)
	Alignment int           // Use alignment constants
	Padding   PaddingConfig // Padding configuration
	MaxLines  int           // For future multi-line support
}

// PaddingConfig defines padding for cells
type PaddingConfig struct {
	Left   int
	Right  int
	Top    int
	Bottom int
}

// Alignment constants
const (
	AlignLeft = iota
	AlignCenter
	AlignRight
)

// ================================
// ANIMATION TYPES
// ================================

// Animation represents a single animation instance
type Animation struct {
	State      map[string]any
	Triggers   []RefreshTrigger
	LastRender time.Time
	IsVisible  bool
}

// AnimationState holds the current state of an animation
type AnimationState struct {
	ID         string
	State      map[string]any
	Triggers   []RefreshTrigger
	LastUpdate time.Time
	NextUpdate time.Time
	IsActive   bool
	IsVisible  bool
	IsDirty    bool // Whether the animation has changed since last render
}

// AnimationConfig controls animation behavior
type AnimationConfig struct {
	Enabled       bool
	ReducedMotion bool
	MaxAnimations int
	BatchUpdates  bool
	TickInterval  time.Duration // Time between animation ticks
}

// ListAnimation represents an animation for list items
type ListAnimation struct {
	ItemID        string
	AnimationType string         // "fade", "slide", "highlight"
	State         map[string]any // Animation-specific state
	Triggers      []RefreshTrigger
}

// CellAnimation represents an animation for table cells
type CellAnimation struct {
	RowID         string
	ColumnIndex   int
	AnimationType string
	State         map[string]any
	Triggers      []RefreshTrigger
}

// RowAnimation represents an animation for table rows
type RowAnimation struct {
	RowID         string
	AnimationType string
	State         map[string]any
	Triggers      []RefreshTrigger
}

// TriggerType defines different types of animation triggers
type TriggerType int

const (
	TriggerTimer TriggerType = iota
	TriggerEvent
	TriggerConditional
)

// RefreshTrigger defines when content should be refreshed
type RefreshTrigger struct {
	Type      TriggerType
	Interval  time.Duration // For timer-based triggers
	Event     string        // For event-based triggers
	Condition func() bool   // For conditional triggers
}

// ================================
// KEYBINDING TYPES
// ================================

// NavigationKeyMap defines key mappings for navigation
type NavigationKeyMap struct {
	Up        []string
	Down      []string
	PageUp    []string
	PageDown  []string
	Home      []string
	End       []string
	Select    []string
	SelectAll []string
	Filter    []string
	Sort      []string
	Quit      []string
}

// ================================
// THEME & STYLING TYPES
// ================================

// StyleConfig defines the styles for list components
type StyleConfig struct {
	CursorStyle    lipgloss.Style
	SelectedStyle  lipgloss.Style
	DefaultStyle   lipgloss.Style
	ThresholdStyle lipgloss.Style
	DisabledStyle  lipgloss.Style
	LoadingStyle   lipgloss.Style
	ErrorStyle     lipgloss.Style
}

// Theme defines the visual appearance for table components
type Theme struct {
	HeaderStyle        lipgloss.Style
	CellStyle          lipgloss.Style
	CursorStyle        lipgloss.Style
	SelectedStyle      lipgloss.Style
	FullRowCursorStyle lipgloss.Style // Style applied to entire row when full row highlighting is enabled
	BorderChars        BorderChars
	BorderColor        string
	HeaderColor        string
	AlternateRowStyle  lipgloss.Style
	DisabledStyle      lipgloss.Style
	LoadingStyle       lipgloss.Style
	ErrorStyle         lipgloss.Style
}

// BorderChars defines the characters used for table borders
type BorderChars struct {
	Horizontal  string
	Vertical    string
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	TopT        string
	BottomT     string
	LeftT       string
	RightT      string
	Cross       string
}

// ================================
// CONFIGURATION TYPES
// ================================

// TableConfig contains all configuration for a table
type TableConfig struct {
	// Core table configuration
	Columns     []TableColumn
	ShowHeader  bool
	ShowBorders bool

	// Highlighting configuration
	FullRowHighlighting bool // Enable full row highlighting mode

	// Viewport configuration
	ViewportConfig ViewportConfig

	// Theme and styling
	Theme Theme

	// Animation configuration
	AnimationConfig AnimationConfig

	// Selection configuration
	SelectionMode SelectionMode

	// Key bindings
	KeyMap NavigationKeyMap
}

// ListConfig contains all configuration for a list
type ListConfig struct {
	// Viewport configuration
	ViewportConfig ViewportConfig

	// Styling
	StyleConfig StyleConfig

	// Enhanced rendering configuration
	RenderConfig ListRenderConfig

	// Animation configuration
	AnimationConfig AnimationConfig

	// Selection configuration
	SelectionMode SelectionMode

	// Key bindings
	KeyMap NavigationKeyMap

	// Display configuration
	MaxWidth int
}

// ================================
// UTILITY FUNCTIONS
// ================================

// DefaultAnimationConfig returns sensible defaults for animation configuration
func DefaultAnimationConfig() AnimationConfig {
	return AnimationConfig{
		Enabled:       true,
		ReducedMotion: false,
		MaxAnimations: 100,
		BatchUpdates:  true,
		TickInterval:  100 * time.Millisecond,
	}
}

// DefaultNavigationKeyMap returns default key mappings
func DefaultNavigationKeyMap() NavigationKeyMap {
	return NavigationKeyMap{
		Up:        []string{"up", "k"},
		Down:      []string{"down", "j"},
		PageUp:    []string{"pgup", "ctrl+u"},
		PageDown:  []string{"pgdown", "ctrl+d"},
		Home:      []string{"home", "g"},
		End:       []string{"end", "G"},
		Select:    []string{" ", "enter"},
		SelectAll: []string{"ctrl+a"},
		Filter:    []string{"/"},
		Sort:      []string{"s"},
		Quit:      []string{"q", "ctrl+c"},
	}
}

// DefaultBorderChars returns default border characters
func DefaultBorderChars() BorderChars {
	return BorderChars{
		Horizontal:  "─",
		Vertical:    "│",
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
		TopT:        "┬",
		BottomT:     "┴",
		LeftT:       "├",
		RightT:      "┤",
		Cross:       "┼",
	}
}

// ================================
// KEYBINDING TYPES
// ================================
