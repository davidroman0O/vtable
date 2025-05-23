// Package vtable provides a virtualized table and list component for Bubble Tea.
// It efficiently handles large datasets by only loading and rendering the visible portion.
package vtable

import (
	"time"
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

// Data wraps an item with its state and metadata for rendering
type Data[T any] struct {
	// Item is the actual data item
	Item T

	// Selected indicates if this item is selected
	Selected bool

	// Metadata contains custom rendering metadata (colors, icons, badges, etc.)
	Metadata map[string]any

	// Disabled indicates if this item should be rendered as disabled
	Disabled bool

	// Hidden indicates if this item should be hidden from view
	Hidden bool
}

// RenderContext provides dimensional constraints and utilities for formatting
type RenderContext struct {
	// MaxWidth is the available width for this cell/row
	MaxWidth int

	// MaxHeight is the available height (1 for single-line, more for multi-line)
	MaxHeight int

	// ColumnIndex indicates which column we're rendering (for tables)
	ColumnIndex int

	// ColumnConfig contains the column configuration (for tables)
	ColumnConfig *TableColumn
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

	// SelectAll selects all items
	SelectAll() bool

	// ClearSelection clears all selections
	ClearSelection()

	// GetSelectedIndices returns the indices of all selected items
	GetSelectedIndices() []int

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
	isCursor bool,
	isTopThreshold bool,
	isBottomThreshold bool,
	animationState map[string]any,
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

// RenderResult contains both the visual output and refresh instructions
type RenderResult struct {
	// Content is the rendered string
	Content string

	// RefreshAfter specifies when to re-render (0 = no refresh)
	RefreshAfter time.Duration

	// TriggerType specifies what type of refresh trigger this is
	TriggerType TriggerType

	// AnimationState stores state between renders
	AnimationState map[string]any
}
