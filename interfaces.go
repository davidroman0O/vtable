// Package vtable provides a virtualized table and list component for Bubble Tea.
// It efficiently handles large datasets by only loading and rendering the visible portion.
package vtable

// DataProvider is an interface for providing data to virtualized components.
// It abstracts the data source and allows for different implementations.
type DataProvider[T any] interface {
	// GetTotal returns the total number of items in the dataset.
	GetTotal() int

	// GetItems returns a slice of items in the specified range.
	// start is the index of the first item to return.
	// count is the number of items to return.
	GetItems(start, count int) ([]T, error)
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
type ItemFormatter[T any] func(item T, index int, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string

// TableColumn defines a table column configuration.
type TableColumn struct {
	// Title is the column header text
	Title string

	// Width is the column width in characters
	Width int

	// Alignment defines how text is aligned in the column (left, right, center)
	Alignment int
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
