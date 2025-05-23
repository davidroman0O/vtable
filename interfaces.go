// Package vtable provides a virtualized table and list component for Bubble Tea.
// It efficiently handles large datasets by only loading and rendering the visible portion.
package vtable

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

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

// AddSortField adds a sort field with direction to a DataRequest
func (r *DataRequest) AddSortField(field string, direction string) {
	// Normalize direction
	if direction != "asc" && direction != "desc" {
		direction = "asc" // Default to ascending
	}

	// Add the sort field and direction
	r.SortFields = append(r.SortFields, field)
	r.SortDirections = append(r.SortDirections, direction)
}

// ClearSort clears all sort fields and directions
func (r *DataRequest) ClearSort() {
	r.SortFields = nil
	r.SortDirections = nil
}

// HasSort returns true if the request has any sort fields
func (r *DataRequest) HasSort() bool {
	return len(r.SortFields) > 0
}

// IsFieldSortedAscending checks if a field is sorted ascending
// Returns: sorted ascending (true), sorted descending (false), not sorted (false, false)
func (r *DataRequest) IsFieldSortedAscending(field string) (bool, bool) {
	for i, f := range r.SortFields {
		if f == field {
			return r.SortDirections[i] == "asc", true
		}
	}
	return false, false
}

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

// Set stores a value with type safety and validation
func SetTypedMetadata[T any](tm *TypedMetadata, key MetadataKey[T], value T) error {
	if key.Validator != nil {
		if err := key.Validator(value); err != nil {
			return err
		}
	}
	tm.data[key.Key] = value
	return nil
}

// Get retrieves a value with type safety
func GetTypedMetadata[T any](tm *TypedMetadata, key MetadataKey[T]) T {
	if val, ok := tm.data[key.Key]; ok {
		if typedVal, ok := val.(T); ok {
			return typedVal
		}
	}
	return key.DefaultValue
}

// GetRaw returns the raw map for backward compatibility
func (tm *TypedMetadata) GetRaw() map[string]any {
	return tm.data
}

// SetRaw sets a raw value (for backward compatibility)
func (tm *TypedMetadata) SetRaw(key string, value any) {
	tm.data[key] = value
}

// Common metadata keys
var (
	StatusColorKey = MetadataKey[string]{"status_color", "#ffffff", nil}
	PriorityKey    = MetadataKey[int]{"priority", 0, nil}
	TooltipKey     = MetadataKey[string]{"tooltip", "", nil}
	IconKey        = MetadataKey[string]{"icon", "", nil}
	BadgeKey       = MetadataKey[string]{"badge", "", nil}
)

// FocusState contains component focus information
type FocusState struct {
	HasFocus    bool
	FocusedCell string // ID of focused cell (for tables)
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

// RenderContext provides dimensional constraints and utilities for formatting
type RenderContext struct {
	// Dimensional constraints
	MaxWidth  int
	MaxHeight int

	// Component context
	ColumnIndex  int
	ColumnConfig *TableColumn

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

	// Utility functions
	Truncate func(string, int) string
	Wrap     func(string, int) []string
	Measure  func(string) (int, int)

	// Error handling
	OnError func(error)
}

// DefaultRenderContext creates a RenderContext with sensible defaults
func DefaultRenderContext() RenderContext {
	return RenderContext{
		MaxWidth:       80,
		MaxHeight:      1,
		ColorSupport:   true,
		UnicodeSupport: true,
		HighContrast:   false,
		ReducedMotion:  false,
		ScreenReader:   false,
		CurrentTime:    time.Now(),
		Truncate:       defaultTruncate,
		Wrap:           defaultWrap,
		Measure:        defaultMeasure,
		OnError:        defaultOnError,
	}
}

// Default utility functions
func defaultTruncate(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return strings.Repeat(".", maxWidth)
	}
	return text[:maxWidth-3] + "..."
}

func defaultWrap(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
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

	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}

	return lines
}

func defaultMeasure(text string) (int, int) {
	lines := strings.Split(text, "\n")
	height := len(lines)
	width := 0
	for _, line := range lines {
		if w := utf8.RuneCountInString(line); w > width {
			width = w
		}
	}
	return width, height
}

func defaultOnError(err error) {
	// Default: silent ignore (can be overridden)
	_ = err
}

// TableColumn defines a table column configuration.
type TableColumn struct {
	// Title is the column header text
	Title string

	// Width is the column width in characters
	Width int

	// Alignment defines how text is aligned in the column (left, right, center)
	Alignment int

	// Field is the identifier used for sorting/filtering operations
	Field string
}

// ViewportState represents the current state of the viewport.
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

// ChunkInfo represents information about a loaded data chunk.
type ChunkInfo struct {
	// StartIndex is the absolute index of the first item in the chunk.
	StartIndex int

	// EndIndex is the absolute index of the last item in the chunk.
	EndIndex int

	// ItemCount is the number of items in the chunk.
	ItemCount int
}

const (
	// AlignLeft aligns text to the left
	AlignLeft = iota
	// AlignCenter aligns text to the center
	AlignCenter
	// AlignRight aligns text to the right
	AlignRight
)

// DataProvider is an interface for providing data to virtualized components.
// It returns Data[T] objects that contain the item plus all rendering state.
type DataProvider[T any] interface {
	// GetTotal returns the total number of items in the dataset.
	GetTotal() int

	// GetItems returns a slice of Data objects based on the provided request.
	// Each Data object contains the item plus selection state and metadata.
	GetItems(request DataRequest) ([]Data[T], error)

	// GetSelectionMode returns the current selection mode
	GetSelectionMode() SelectionMode

	// SetSelected sets the selection state for an item at the given index.
	// Returns true if the operation was successful.
	SetSelected(index int, selected bool) bool

	// SetSelectedByIDs sets the selection state for items with the given IDs.
	// Returns true if the operation was successful.
	SetSelectedByIDs(ids []string, selected bool) bool

	// SelectRange selects items between startID and endID (inclusive).
	// Returns true if the operation was successful.
	SelectRange(startID, endID string) bool

	// SelectAll selects all items
	SelectAll() bool

	// ClearSelection clears all selections
	ClearSelection()

	// GetSelectedIndices returns the indices of all selected items
	GetSelectedIndices() []int

	// GetSelectedIDs returns the IDs of all selected items
	GetSelectedIDs() []string

	// GetItemID extracts the unique identifier from an item.
	// We don't know the ID field by default, so each provider must implement this.
	GetItemID(item *T) string
}

// SearchableDataProvider extends DataProvider with search capabilities.
type SearchableDataProvider[T any] interface {
	DataProvider[T]

	// FindItemIndex searches for an item based on the given criteria and returns its index.
	// The criteria is provided as a key-value pair.
	// If the item is found, its index and true are returned.
	// If the item is not found, -1 and false are returned.
	FindItemIndex(key string, value any) (int, bool)
}

// ItemFormatter is a function type for formatting items in the viewport.
// It receives the Data object which contains the item, selection state, and metadata.
type ItemFormatter[T any] func(data Data[T], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string

// ItemFormatterAnimated is an enhanced formatter that supports animations.
// It returns a RenderResult that can trigger refreshes for dynamic content.
type ItemFormatterAnimated[T any] func(
	data Data[T],
	index int,
	ctx RenderContext,
	animationState map[string]any,
	isCursor bool,
	isTopThreshold bool,
	isBottomThreshold bool,
) RenderResult

// TriggerType defines when a render refresh should occur
type TriggerType int

const (
	// TriggerTimer refreshes at regular intervals
	TriggerTimer TriggerType = iota
	// TriggerEvent refreshes when specific events occur
	TriggerEvent
	// TriggerConditional refreshes when conditions are met
	TriggerConditional
	// TriggerOnce refreshes only once after initial render
	TriggerOnce
)

// RefreshTrigger defines when a cell should be re-rendered
type RefreshTrigger struct {
	Type      TriggerType
	Interval  time.Duration // For timer-based triggers
	Event     string        // For event-based triggers
	Condition func() bool   // For conditional triggers
}

// RenderResult contains both the visual output and refresh instructions
type RenderResult struct {
	// Content is the rendered string
	Content string

	// RefreshTriggers specify when this cell should be re-rendered
	RefreshTriggers []RefreshTrigger

	// AnimationState stores state between renders
	AnimationState map[string]any

	// Error contains any rendering error
	Error error

	// Fallback content to use if there's an error
	Fallback string
}

// Animation represents an active animation
type Animation struct {
	State      map[string]any
	Triggers   []RefreshTrigger
	LastRender time.Time
	IsVisible  bool
}

// AnimationConfig configures animation behavior
type AnimationConfig struct {
	Enabled       bool
	ReducedMotion bool
	MaxAnimations int
	BatchUpdates  bool
	TickInterval  time.Duration // Time between animation ticks
}

// DefaultAnimationConfig returns sensible animation defaults
func DefaultAnimationConfig() AnimationConfig {
	return AnimationConfig{
		Enabled:       true,
		ReducedMotion: false,
		MaxAnimations: 50,
		BatchUpdates:  true,
		TickInterval:  100 * time.Millisecond, // Default tick interval
	}
}
