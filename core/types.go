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

// TableRow represents a row of data in the table.
// It contains a stable unique identifier and the cell data for that row.
type TableRow struct {
	// ID is a stable unique identifier for this row. It is crucial for
	// maintaining state across data refreshes and for operations like selection
	// and animation.
	ID string

	// Cells contains the string values for each column in the row. The order of
	// cells should correspond to the order of columns defined in the table
	// configuration.
	Cells []string
}

// TableColumn represents the configuration for a single column in a table.
// It defines properties like the title, width, alignment, and the data field it
// corresponds to.
type TableColumn struct {
	// Title is the column header text.
	Title string

	// Width is the column width in characters.
	Width int

	// Alignment defines how text is aligned in the column cells (left, right,
	// center). Use the AlignLeft, AlignCenter, or AlignRight constants.
	Alignment int

	// Field is the identifier used for sorting/filtering operations. This should
	// correspond to a key in the underlying data source.
	Field string

	// HeaderAlignment defines alignment for the header text, which can be
	// different from the cell alignment.
	HeaderAlignment int
	// HeaderConstraint defines formatting constraints for the header cell.
	HeaderConstraint CellConstraint
}

// Data is a generic wrapper for any data item managed by a vtable component.
// It augments the original item with state information essential for rendering
// and interaction, such as selection status, loading state, and associated
// errors.
type Data[T any] struct {
	// ID is a stable unique identifier for this item, crucial for state
	// management.
	ID string

	// Item is the actual data item of type T.
	Item T

	// Selected indicates if this item is currently selected.
	Selected bool

	// Metadata contains custom, type-safe rendering metadata for advanced styling
	// or behavior.
	Metadata TypedMetadata

	// Disabled indicates if this item should be rendered as disabled and be
	// non-interactive.
	Disabled bool

	// Hidden indicates if this item should be hidden from view.
	Hidden bool

	// Error contains any error state associated with this item, which can be
	// used for special styling.
	Error error

	// Loading indicates if this item is currently in a loading state, for
	// example, if its data is being fetched.
	Loading bool
}

// BoundingArea represents the area around the viewport where data chunks should be
// pre-emptively loaded to ensure smooth scrolling. It is defined by absolute
// item indices and chunk boundaries.
type BoundingArea struct {
	// StartIndex is the absolute start index of the bounding area.
	StartIndex int
	// EndIndex is the absolute end index (inclusive) of the bounding area.
	EndIndex int
	// ChunkStart is the start index of the first chunk within the bounding area.
	ChunkStart int
	// ChunkEnd is the boundary for the last chunk in the area (exclusive).
	ChunkEnd int
}

// ViewportState contains the current positional information of a component's
// viewport, such as cursor position and scroll offsets.
type ViewportState struct {
	// ViewportStartIndex is the absolute index of the first item visible in the
	// viewport.
	ViewportStartIndex int

	// CursorIndex is the absolute index of the selected item in the entire
	// dataset.
	CursorIndex int

	// CursorViewportIndex is the relative index of the cursor within the visible
	// viewport (0 to Height-1).
	CursorViewportIndex int

	// IsAtTopThreshold indicates if the cursor is at the top scroll-triggering
	// threshold.
	IsAtTopThreshold bool

	// IsAtBottomThreshold indicates if the cursor is at the bottom
	// scroll-triggering threshold.
	IsAtBottomThreshold bool

	// AtDatasetStart indicates if the viewport is at the very beginning of the
	// dataset.
	AtDatasetStart bool

	// AtDatasetEnd indicates if the viewport is at the very end of the dataset.
	AtDatasetEnd bool
}

// ViewportConfig defines the configuration for the viewport's behavior,
// including its size, scrolling thresholds, and data chunking strategy.
type ViewportConfig struct {
	// Height is the number of items visible in the viewport.
	Height int

	// TopThreshold is the offset from the viewport's start where scrolling up is
	// triggered. A value of -1 disables it.
	TopThreshold int

	// BottomThreshold is the offset from the viewport's end where scrolling down
	// is triggered. A value of -1 disables it.
	BottomThreshold int

	// ChunkSize is the number of items to load in each data chunk.
	ChunkSize int

	// InitialIndex is the starting cursor position when the component is
	// initialized.
	InitialIndex int

	// BoundingAreaBefore is the number of items to keep loaded before the
	// viewport top.
	BoundingAreaBefore int

	// BoundingAreaAfter is the number of items to keep loaded after the viewport
	// bottom.
	BoundingAreaAfter int
}

// DataRequest represents a request for a segment of data from a DataSource.
// It supports pagination, sorting, and filtering.
type DataRequest struct {
	// Start is the index of the first item to return.
	Start int

	// Count is the number of items to return.
	Count int

	// SortFields specifies the fields to sort by, in order of priority.
	SortFields []string

	// SortDirections specifies the sort directions ("asc" or "desc")
	// corresponding to SortFields.
	SortDirections []string

	// Filters is a map of field names to their corresponding filter values.
	Filters map[string]any
}

// Chunk represents a block of data loaded from a DataSource. Components use
// chunks to manage large datasets efficiently without keeping everything in memory.
type Chunk[T any] struct {
	// StartIndex is the absolute index of the first item in the chunk.
	StartIndex int

	// EndIndex is the absolute index of the last item in the chunk.
	EndIndex int

	// Items is the slice of Data items contained in the chunk.
	Items []Data[T]

	// LoadedAt is the timestamp when this chunk was loaded into memory.
	LoadedAt time.Time

	// Request is the DataRequest that was used to load this chunk, useful for
	// validation and debugging.
	Request DataRequest
}

// ChunkInfo provides metadata about a loaded chunk.
type ChunkInfo struct {
	// StartIndex is the absolute index of the first item in the chunk.
	StartIndex int

	// EndIndex is the absolute index of the last item in the chunk.
	EndIndex int

	// ItemCount is the number of items in the chunk.
	ItemCount int
}

// SelectionMode defines the selection behavior of a component.
type SelectionMode int

const (
	// SelectionSingle allows only one item to be selected at a time.
	SelectionSingle SelectionMode = iota
	// SelectionMultiple allows multiple items to be selected simultaneously.
	SelectionMultiple
	// SelectionNone disables item selection entirely.
	SelectionNone
)

// MetadataKey represents a type-safe key for storing and retrieving values from
// TypedMetadata. It includes a default value and an optional validator function.
type MetadataKey[T any] struct {
	// Key is the string identifier for the metadata.
	Key string
	// DefaultValue is the value returned if the key is not found.
	DefaultValue T
	// Validator is a function to validate values before they are set.
	Validator func(T) error
}

// TypedMetadata provides a type-safe container for custom metadata. It uses
// MetadataKey to ensure type correctness at compile time.
type TypedMetadata struct {
	data map[string]any
}

// NewTypedMetadata creates a new, empty TypedMetadata instance.
func NewTypedMetadata() TypedMetadata {
	return TypedMetadata{
		data: make(map[string]any),
	}
}

// SetTypedMetadata stores a value with its associated type-safe key.
// It runs the key's validator function if one is provided.
func SetTypedMetadata[T any](tm *TypedMetadata, key MetadataKey[T], value T) error {
	if key.Validator != nil {
		if err := key.Validator(value); err != nil {
			return err
		}
	}
	tm.data[key.Key] = value
	return nil
}

// GetTypedMetadata retrieves a value using its type-safe key.
// If the key is not found or the type does not match, it returns the key's
// default value.
func GetTypedMetadata[T any](tm *TypedMetadata, key MetadataKey[T]) T {
	if val, ok := tm.data[key.Key]; ok {
		if typedVal, ok := val.(T); ok {
			return typedVal
		}
	}
	return key.DefaultValue
}

// GetRaw returns the raw map[string]any for situations where type safety is not
// required, such as during serialization.
func (tm *TypedMetadata) GetRaw() map[string]any {
	return tm.data
}

// SetRaw sets a value without type-safety. Use with caution.
func (tm *TypedMetadata) SetRaw(key string, value any) {
	tm.data[key] = value
}

// Common metadata keys are predefined for frequent use cases.
var (
	// StatusColorKey is for storing a color string (e.g., "#FF0000") for status indicators.
	StatusColorKey = MetadataKey[string]{"status_color", "#ffffff", nil}
	// PriorityKey is for storing an integer priority level.
	PriorityKey = MetadataKey[int]{"priority", 0, nil}
	// TooltipKey is for storing a string tooltip.
	TooltipKey = MetadataKey[string]{"tooltip", "", nil}
	// IconKey is for storing an icon character or string.
	IconKey = MetadataKey[string]{"icon", "", nil}
	// BadgeKey is for storing a badge string.
	BadgeKey = MetadataKey[string]{"badge", "", nil}
)

// FocusState contains information about the component's focus state.
type FocusState struct {
	// HasFocus is true if the component is currently focused.
	HasFocus bool
	// FocusedCell identifies the currently focused cell in a table, typically in
	// "rowID:columnField" format.
	FocusedCell string
}

// RenderContext provides dimensional constraints, styling, and utility functions
// to formatters, ensuring consistent rendering across the component.
type RenderContext struct {
	// MaxWidth is the maximum width available for rendering.
	MaxWidth int
	// MaxHeight is the maximum height available for rendering.
	MaxHeight int

	// Component context
	ColumnIndex int

	// Styling & theming
	// Theme provides the active theme for table components.
	Theme *Theme
	// BaseStyle is the default style for list components.
	BaseStyle lipgloss.Style

	// Terminal capabilities
	// ColorSupport is true if the terminal supports colors.
	ColorSupport bool
	// UnicodeSupport is true if the terminal supports unicode characters.
	UnicodeSupport bool

	// Accessibility
	// HighContrast is true if high contrast mode is enabled.
	HighContrast bool
	// ReducedMotion is true if reduced motion mode is enabled.
	ReducedMotion bool
	// ScreenReader is true if screen reader support is enabled.
	ScreenReader bool

	// Global state
	// CurrentTime is the time of the current render cycle.
	CurrentTime time.Time
	// FocusState contains information about the component's focus.
	FocusState FocusState
	// DeltaTime is the duration since the last render, useful for animations.
	DeltaTime time.Duration

	// State indicators (configurable)
	// ErrorIndicator is the string used to indicate an error state.
	ErrorIndicator string
	// LoadingIndicator is the string used to indicate a loading state.
	LoadingIndicator string
	// DisabledIndicator is the string used to indicate a disabled state.
	DisabledIndicator string
	// SelectedIndicator is the string used to indicate a selected state.
	SelectedIndicator string

	// Utility functions
	// Truncate shortens a string to a given width.
	Truncate func(string, int) string
	// Wrap wraps a string to a given width.
	Wrap func(string, int) []string
	// Measure calculates the width and height of a string.
	Measure func(string, int) (int, int)

	// Error handling
	// OnError is a callback for reporting rendering errors.
	OnError func(error)
}

// RenderResult contains the output of an animated rendering operation. It includes
// the content and metadata needed for re-rendering and state management.
type RenderResult struct {
	// Content is the rendered string output.
	Content string

	// RefreshTriggers specify conditions under which this content should be
	// re-rendered.
	RefreshTriggers []RefreshTrigger

	// AnimationState stores state between renders for this specific item.
	AnimationState map[string]any

	// Error contains any error that occurred during rendering.
	Error error

	// Fallback content to use if a rendering error occurs.
	Fallback string
}

// CellRenderResult contains the result of a table cell rendering operation,
// including the content and its dimensional properties.
type CellRenderResult struct {
	// Content is the rendered cell content.
	Content string

	// ActualWidth is the actual width of the rendered content.
	ActualWidth int

	// ActualHeight is the actual height of the rendered content.
	ActualHeight int

	// Overflow indicates whether the content was truncated or wrapped.
	Overflow bool

	// RefreshTriggers specify when this cell should be re-rendered.
	RefreshTriggers []RefreshTrigger

	// AnimationState stores state between renders for this specific cell.
	AnimationState map[string]any

	// Error contains any rendering error.
	Error error

	// Fallback content to use if a rendering error occurs.
	Fallback string
}

// CellConstraint defines the dimensional and alignment constraints for a table cell.
type CellConstraint struct {
	// Width is the exact width the cell must occupy.
	Width int
	// Height is the exact height the cell must occupy (usually 1 for tables).
	Height int
	// Alignment specifies text alignment (AlignLeft, AlignCenter, AlignRight).
	Alignment int
	// Padding defines the padding within the cell.
	Padding PaddingConfig
	// MaxLines is the maximum number of lines for multi-line content.
	MaxLines int
}

// PaddingConfig defines the padding for each side of a cell.
type PaddingConfig struct {
	Left   int
	Right  int
	Top    int
	Bottom int
}

// Alignment constants for text.
const (
	AlignLeft   = 0
	AlignCenter = 1
	AlignRight  = 2
)

// Animation represents a single animation instance.
type Animation struct {
	// State holds the current values for the animation (e.g., opacity, position).
	State map[string]any
	// Triggers define when the animation should be updated.
	Triggers []RefreshTrigger
	// LastRender is the timestamp of the last render.
	LastRender time.Time
	// IsVisible indicates if the animated item is currently in the viewport.
	IsVisible bool
}

// AnimationState holds the complete state of a managed animation.
type AnimationState struct {
	// ID is the unique identifier for the animation.
	ID string
	// State holds the animation's current values.
	State map[string]any
	// Triggers define when the animation updates.
	Triggers []RefreshTrigger
	// LastUpdate is the timestamp of the last state change.
	LastUpdate time.Time
	// NextUpdate is the scheduled time for the next update (for timer triggers).
	NextUpdate time.Time
	// IsActive indicates if the animation is currently running.
	IsActive bool
	// IsVisible indicates if the animated item is in the viewport.
	IsVisible bool
	// IsDirty is true if the animation state has changed since the last render.
	IsDirty bool
}

// AnimationConfig controls the global behavior of the animation engine.
type AnimationConfig struct {
	// Enabled globally enables or disables animations.
	Enabled bool
	// ReducedMotion, when true, disables non-essential animations for accessibility.
	ReducedMotion bool
	// MaxAnimations is the maximum number of concurrent animations.
	MaxAnimations int
	// BatchUpdates, when true, batches multiple animation updates into a single message.
	BatchUpdates bool
	// TickInterval is the time between global animation ticks.
	TickInterval time.Duration
}

// ListAnimation represents an animation for a list item.
type ListAnimation struct {
	// ItemID is the ID of the list item to animate.
	ItemID string
	// AnimationType is the type of animation (e.g., "fade", "slide").
	AnimationType string
	// State holds the animation's current values.
	State map[string]any
	// Triggers define when the animation should update.
	Triggers []RefreshTrigger
}

// CellAnimation represents an animation for a table cell.
type CellAnimation struct {
	// RowID is the ID of the row containing the cell.
	RowID string
	// ColumnIndex is the index of the column containing the cell.
	ColumnIndex int
	// AnimationType is the type of animation.
	AnimationType string
	// State holds the animation's current values.
	State map[string]any
	// Triggers define when the animation should update.
	Triggers []RefreshTrigger
}

// RowAnimation represents an animation for an entire table row.
type RowAnimation struct {
	// RowID is the ID of the row to animate.
	RowID string
	// AnimationType is the type of animation.
	AnimationType string
	// State holds the animation's current values.
	State map[string]any
	// Triggers define when the animation should update.
	Triggers []RefreshTrigger
}

// TriggerType defines the different kinds of animation triggers.
type TriggerType int

const (
	// TriggerTimer updates the animation periodically.
	TriggerTimer TriggerType = iota
	// TriggerEvent updates the animation in response to a specific event.
	TriggerEvent
	// TriggerConditional updates the animation when a specific condition is met.
	TriggerConditional
)

// RefreshTrigger defines a condition that, when met, causes an item to be
// re-rendered. This is essential for animations and real-time updates.
type RefreshTrigger struct {
	// Type is the kind of trigger.
	Type TriggerType
	// Interval is the duration for timer-based triggers.
	Interval time.Duration
	// Event is the name for event-based triggers.
	Event string
	// Condition is the function for conditional triggers.
	Condition func() bool
}

// ================================
// KEYBINDING TYPES
// ================================

// NavigationKeyMap defines the key mappings for component navigation and actions.
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

// StyleConfig defines the styles for various states of list items.
type StyleConfig struct {
	// CursorStyle is the style for the item under the cursor.
	CursorStyle lipgloss.Style
	// SelectedStyle is the style for a selected item.
	SelectedStyle lipgloss.Style
	// DefaultStyle is the style for a normal, unselected item.
	DefaultStyle lipgloss.Style
	// ThresholdStyle is the style for items at the scroll threshold.
	ThresholdStyle lipgloss.Style
	// DisabledStyle is the style for a disabled item.
	DisabledStyle lipgloss.Style
	// LoadingStyle is the style for a loading item.
	LoadingStyle lipgloss.Style
	// ErrorStyle is the style for an item with an error.
	ErrorStyle lipgloss.Style
}

// Theme defines the visual appearance and character set for table components.
type Theme struct {
	// HeaderStyle is the style for header cells.
	HeaderStyle lipgloss.Style
	// CellStyle is the style for non-header cells.
	CellStyle lipgloss.Style
	// CursorStyle is the style for the cell under the cursor.
	CursorStyle lipgloss.Style
	// SelectedStyle is the style for selected rows.
	SelectedStyle lipgloss.Style
	// FullRowCursorStyle is the style applied to the entire row when full-row
	// highlighting is enabled.
	FullRowCursorStyle lipgloss.Style
	// BorderChars defines the characters used for drawing table borders.
	BorderChars BorderChars
	// BorderColor is the color for table borders.
	BorderColor string
	// HeaderColor is the color for header text.
	HeaderColor string
	// AlternateRowStyle is a style applied to alternating rows for readability.
	AlternateRowStyle lipgloss.Style
	// DisabledStyle is the style for disabled rows.
	DisabledStyle lipgloss.Style
	// LoadingStyle is the style for loading placeholder rows.
	LoadingStyle lipgloss.Style
	// ErrorStyle is the style for rows with errors.
	ErrorStyle lipgloss.Style
}

// BorderChars defines the characters used for drawing table borders.
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

// TableConfig contains all configuration options for a table component.
type TableConfig struct {
	// Columns defines the structure of the table columns.
	Columns []TableColumn
	// ShowHeader controls the visibility of the table header.
	ShowHeader bool
	// ShowBorders is a global toggle for all table borders.
	ShowBorders bool

	// Granular border configuration
	// ShowTopBorder controls the visibility of the top border.
	ShowTopBorder bool
	// ShowBottomBorder controls the visibility of the bottom border.
	ShowBottomBorder bool
	// ShowHeaderSeparator controls the visibility of the line between the header
	// and the body.
	ShowHeaderSeparator bool

	// Space removal for borders (when true, completely removes the line space)
	// RemoveTopBorderSpace, if true, removes the line where the top border would be.
	RemoveTopBorderSpace bool
	// RemoveBottomBorderSpace, if true, removes the line where the bottom border would be.
	RemoveBottomBorderSpace bool

	// FullRowHighlighting enables a mode where the entire row is highlighted by the cursor.
	FullRowHighlighting bool

	// ResetScrollOnNavigation, if true, resets horizontal scroll offsets when
	// navigating between rows.
	ResetScrollOnNavigation bool

	// ActiveCellIndicationEnabled toggles the background highlighting of the active cell.
	ActiveCellIndicationEnabled bool
	// ActiveCellBackgroundColor sets the background color for the active cell.
	ActiveCellBackgroundColor string

	// ViewportConfig defines the viewport behavior.
	ViewportConfig ViewportConfig

	// Theme defines the visual style of the table.
	Theme Theme

	// AnimationConfig controls animation behavior.
	// TODO: i need more ideas before doing that
	// AnimationConfig AnimationConfig

	// SelectionMode defines the selection behavior.
	SelectionMode SelectionMode

	// KeyMap defines the keybindings for navigation and actions.
	KeyMap NavigationKeyMap
}

// ListConfig contains all configuration options for a list component.
type ListConfig struct {
	// ViewportConfig defines the viewport behavior.
	ViewportConfig ViewportConfig

	// StyleConfig defines the styles for list items.
	StyleConfig StyleConfig

	// RenderConfig defines the component-based rendering pipeline.
	RenderConfig ListRenderConfig

	// AnimationConfig controls animation behavior.
	AnimationConfig AnimationConfig

	// SelectionMode defines the selection behavior.
	SelectionMode SelectionMode

	// KeyMap defines the keybindings for navigation and actions.
	KeyMap NavigationKeyMap

	// MaxWidth is the maximum width of the list.
	MaxWidth int
}

// ListRenderConfig contains the configuration for the component-based list
// rendering pipeline. It defines which visual components are rendered and in
// what order.
type ListRenderConfig struct {
	// ComponentOrder defines the sequence of components to render for each list
	// item (e.g., cursor, enumerator, content).
	ComponentOrder []ListComponentType

	// Component configurations
	CursorConfig      ListCursorConfig
	PreSpacingConfig  ListSpacingConfig
	EnumeratorConfig  ListEnumeratorConfig
	ContentConfig     ListContentConfig
	PostSpacingConfig ListSpacingConfig
	BackgroundConfig  ListBackgroundConfig
}

// ListRenderComponent represents a single, pluggable piece of the list item
// rendering pipeline, such as the cursor indicator or the item content.
type ListRenderComponent interface {
	// Render generates the string content for this component.
	Render(ctx ListComponentContext) string
	// GetType returns the component's unique type identifier.
	GetType() ListComponentType
	// IsEnabled returns whether this component should be rendered.
	IsEnabled() bool
	// SetEnabled enables or disables this component.
	SetEnabled(enabled bool)
}

// ListComponentType is a unique identifier for each type of list rendering
// component.
type ListComponentType string

// Constants for all available list component types.
const (
	ListComponentCursor      ListComponentType = "cursor"
	ListComponentPreSpacing  ListComponentType = "pre_spacing"
	ListComponentEnumerator  ListComponentType = "enumerator"
	ListComponentContent     ListComponentType = "content"
	ListComponentPostSpacing ListComponentType = "post_spacing"
	ListComponentBackground  ListComponentType = "background"
)

// ListComponentContext provides all necessary data for a ListRenderComponent to
// render itself. It is passed to the Render method of each component.
type ListComponentContext struct {
	// Item is the data for the item being rendered.
	Item Data[any]
	// Index is the absolute index of the item in the dataset.
	Index int
	// IsCursor is true if the item is under the cursor.
	IsCursor bool
	// IsSelected is true if the item is selected.
	IsSelected bool
	// IsThreshold is true if the item is at a scroll threshold.
	IsThreshold bool

	// RenderContext provides global rendering information.
	RenderContext RenderContext

	// ComponentData is a map containing the rendered output of preceding
	// components in the pipeline.
	ComponentData map[ListComponentType]string

	// ListConfig holds the current rendering configuration for the list.
	ListConfig ListRenderConfig
}

// ListEnumerator is a function type that generates a prefix for a list item,
// such as a bullet point, number, or checkbox.
type ListEnumerator func(item Data[any], index int, ctx RenderContext) string

// ListCursorConfig configures the cursor component.
type ListCursorConfig struct {
	Enabled         bool
	CursorIndicator string
	NormalSpacing   string
	Style           lipgloss.Style
}

// ListSpacingConfig configures spacing components.
type ListSpacingConfig struct {
	Enabled bool
	Spacing string
	Style   lipgloss.Style
}

// ListEnumeratorConfig configures the enumerator component.
type ListEnumeratorConfig struct {
	Enabled    bool
	Enumerator ListEnumerator
	Style      lipgloss.Style
	Alignment  ListEnumeratorAlignment
	MaxWidth   int
}

// ListContentConfig configures the main content component.
type ListContentConfig struct {
	Enabled   bool
	Formatter ItemFormatter[any]
	Style     lipgloss.Style
	WrapText  bool
	MaxWidth  int
}

// ListBackgroundConfig configures the background styling component.
type ListBackgroundConfig struct {
	Enabled           bool
	Style             lipgloss.Style
	ApplyToComponents []ListComponentType
	Mode              ListBackgroundMode
}

// ListEnumeratorAlignment defines the text alignment for enumerators.
type ListEnumeratorAlignment int

// Constants for list enumerator alignment.
const (
	ListAlignmentNone ListEnumeratorAlignment = iota
	ListAlignmentLeft
	ListAlignmentRight
)

// ListBackgroundMode defines how background styling is applied.
type ListBackgroundMode int

// Constants for list background rendering modes.
const (
	// ListBackgroundEntireLine applies the background to the entire line.
	ListBackgroundEntireLine ListBackgroundMode = iota
	// ListBackgroundSelectiveComponents applies the background to a specified
	// subset of components.
	ListBackgroundSelectiveComponents
	// ListBackgroundContentOnly applies the background only to the content component.
	ListBackgroundContentOnly
	// ListBackgroundIndicatorOnly applies the background only to the cursor
	// component.
	ListBackgroundIndicatorOnly
)

// DefaultAnimationConfig returns a sensible default configuration for animations.
func DefaultAnimationConfig() AnimationConfig {
	return AnimationConfig{
		Enabled:       true,
		ReducedMotion: false,
		MaxAnimations: 100,
		BatchUpdates:  true,
		TickInterval:  100 * time.Millisecond,
	}
}

// DefaultNavigationKeyMap returns the default key mappings for navigation.
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

// DefaultBorderChars returns the default characters used for table borders.
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
