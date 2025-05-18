package vtable

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Chunk represents a chunk of data loaded from the data provider.
type chunk[T any] struct {
	// StartIndex is the absolute index of the first item in the chunk.
	StartIndex int

	// EndIndex is the absolute index of the last item in the chunk.
	EndIndex int

	// Items is the slice of items in the chunk.
	Items []T
}

// List is a virtualized list component that efficiently handles large datasets.
type List[T any] struct {
	// Config is the configuration for the viewport.
	Config ViewportConfig

	// DataProvider is the data provider for the list.
	DataProvider DataProvider[T]

	// StyleConfig defines the styles for the list.
	StyleConfig StyleConfig

	// State is the current state of the viewport.
	State ViewportState

	// Formatter is the function used to format items.
	Formatter ItemFormatter[T]

	// chunks is a map of chunks loaded from the data provider.
	// The key is the start index of the chunk.
	chunks map[int]*chunk[T]

	// totalItems is the total number of items in the dataset.
	totalItems int

	// visibleItems is the slice of items currently visible in the viewport.
	visibleItems []T
}

// NewList creates a new virtualized list component.
func NewList[T any](
	config ViewportConfig,
	provider DataProvider[T],
	styleConfig StyleConfig,
	formatter ItemFormatter[T],
) (*List[T], error) {
	if config.Height <= 0 {
		return nil, fmt.Errorf("viewport height must be greater than 0")
	}

	if config.TopThresholdIndex < 0 || config.TopThresholdIndex >= config.Height {
		return nil, fmt.Errorf("top threshold must be within viewport bounds")
	}

	if config.BottomThresholdIndex < 0 || config.BottomThresholdIndex >= config.Height {
		return nil, fmt.Errorf("bottom threshold must be within viewport bounds")
	}

	if config.BottomThresholdIndex <= config.TopThresholdIndex {
		return nil, fmt.Errorf("bottom threshold must be below top threshold")
	}

	if config.ChunkSize <= 0 {
		return nil, fmt.Errorf("chunk size must be greater than 0")
	}

	totalItems := provider.GetTotal()
	if totalItems <= 0 {
		return nil, fmt.Errorf("dataset must have at least one item")
	}

	initialIndex := config.InitialIndex
	if initialIndex < 0 || initialIndex >= totalItems {
		initialIndex = 0
	}

	viewportStartIndex := 0
	cursorViewportIndex := initialIndex

	// Adjust if initialIndex would place cursor beyond viewport
	if initialIndex >= config.Height {
		viewportStartIndex = initialIndex - config.TopThresholdIndex
		cursorViewportIndex = config.TopThresholdIndex
	}

	list := &List[T]{
		Config:       config,
		DataProvider: provider,
		StyleConfig:  styleConfig,
		Formatter:    formatter,
		State: ViewportState{
			ViewportStartIndex:  viewportStartIndex,
			CursorIndex:         initialIndex,
			CursorViewportIndex: cursorViewportIndex,
			IsAtTopThreshold:    cursorViewportIndex == config.TopThresholdIndex,
			IsAtBottomThreshold: cursorViewportIndex == config.BottomThresholdIndex,
			AtDatasetStart:      viewportStartIndex == 0,
			AtDatasetEnd:        viewportStartIndex+config.Height >= totalItems,
		},
		chunks:       make(map[int]*chunk[T]),
		totalItems:   totalItems,
		visibleItems: make([]T, 0, config.Height),
	}

	// Load initial chunk
	chunkStartIndex := (viewportStartIndex / config.ChunkSize) * config.ChunkSize
	err := list.loadChunk(chunkStartIndex)
	if err != nil {
		return nil, err
	}

	// Initialize visible items
	list.updateVisibleItems()

	return list, nil
}

// loadChunk loads a chunk of data from the data provider.
func (l *List[T]) loadChunk(startIndex int) error {
	// Check if the chunk is already loaded
	if _, ok := l.chunks[startIndex]; ok {
		return nil
	}

	// Calculate the number of items to load
	count := l.Config.ChunkSize
	if startIndex+count > l.totalItems {
		count = l.totalItems - startIndex
	}

	// Load the items from the data provider
	items, err := l.DataProvider.GetItems(startIndex, count)
	if err != nil {
		return err
	}

	// Create and store the chunk
	l.chunks[startIndex] = &chunk[T]{
		StartIndex: startIndex,
		EndIndex:   startIndex + len(items) - 1,
		Items:      items,
	}

	return nil
}

// unloadChunks unloads chunks that are no longer needed.
func (l *List[T]) unloadChunks() {
	// Calculate the bounds of chunks that should be kept
	viewportChunkIndex := (l.State.ViewportStartIndex / l.Config.ChunkSize) * l.Config.ChunkSize
	keepLowerBound := viewportChunkIndex - l.Config.ChunkSize
	if keepLowerBound < 0 {
		keepLowerBound = 0
	}
	keepUpperBound := viewportChunkIndex + (2 * l.Config.ChunkSize)

	// Unload chunks outside the bounds
	for startIndex := range l.chunks {
		if startIndex < keepLowerBound || startIndex > keepUpperBound {
			delete(l.chunks, startIndex)
		}
	}
}

// updateVisibleItems updates the slice of items currently visible in the viewport.
func (l *List[T]) updateVisibleItems() {
	l.visibleItems = make([]T, 0, l.Config.Height)

	// Calculate the end index of the viewport
	viewportEndIndex := l.State.ViewportStartIndex + l.Config.Height
	if viewportEndIndex > l.totalItems {
		viewportEndIndex = l.totalItems
	}

	// Collect visible items from chunks
	for i := l.State.ViewportStartIndex; i < viewportEndIndex; i++ {
		chunkStartIndex := (i / l.Config.ChunkSize) * l.Config.ChunkSize
		chunk, ok := l.chunks[chunkStartIndex]
		if !ok {
			// Load the chunk if it's not loaded yet
			err := l.loadChunk(chunkStartIndex)
			if err != nil {
				// If we can't load the chunk, add a zero value as a placeholder
				var zero T
				l.visibleItems = append(l.visibleItems, zero)
				continue
			}
			chunk = l.chunks[chunkStartIndex]
		}

		// Add the item to the visible items
		itemIndex := i - chunk.StartIndex
		if itemIndex >= 0 && itemIndex < len(chunk.Items) {
			l.visibleItems = append(l.visibleItems, chunk.Items[itemIndex])
		} else {
			// This should not happen, but add a zero value as a placeholder just in case
			var zero T
			l.visibleItems = append(l.visibleItems, zero)
		}
	}
}

// GetVisibleItems returns the slice of items currently visible in the viewport.
func (l *List[T]) GetVisibleItems() []T {
	return l.visibleItems
}

// GetTotalItems returns the total number of items in the dataset.
func (l *List[T]) GetTotalItems() int {
	return l.totalItems
}

// GetCurrentItem returns the currently selected item.
func (l *List[T]) GetCurrentItem() (T, bool) {
	if l.State.CursorIndex < 0 || l.State.CursorIndex >= l.totalItems {
		var zero T
		return zero, false
	}

	chunkStartIndex := (l.State.CursorIndex / l.Config.ChunkSize) * l.Config.ChunkSize
	chunk, ok := l.chunks[chunkStartIndex]
	if !ok {
		var zero T
		return zero, false
	}

	itemIndex := l.State.CursorIndex - chunk.StartIndex
	if itemIndex < 0 || itemIndex >= len(chunk.Items) {
		var zero T
		return zero, false
	}

	return chunk.Items[itemIndex], true
}

// GetState returns the current viewport state.
func (l *List[T]) GetState() ViewportState {
	return l.State
}

// GetLoadedChunks returns information about the currently loaded chunks.
func (l *List[T]) GetLoadedChunks() []ChunkInfo {
	chunks := make([]ChunkInfo, 0, len(l.chunks))
	for _, chunk := range l.chunks {
		chunks = append(chunks, ChunkInfo{
			StartIndex: chunk.StartIndex,
			EndIndex:   chunk.EndIndex,
			ItemCount:  len(chunk.Items),
		})
	}
	return chunks
}

// Render renders the list to a string using the provided formatter.
func (l *List[T]) Render() string {
	var builder strings.Builder

	// Set up styles
	rowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(l.StyleConfig.RowStyle))
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(l.StyleConfig.SelectedRowStyle))

	// Render items in the viewport
	for i, item := range l.visibleItems {
		absoluteIndex := l.State.ViewportStartIndex + i
		isCursor := i == l.State.CursorViewportIndex
		isTopThreshold := i == l.Config.TopThresholdIndex
		isBottomThreshold := i == l.Config.BottomThresholdIndex

		// Format the item using the formatter
		var renderedItem string
		if l.Formatter != nil {
			renderedItem = l.Formatter(item, absoluteIndex, isCursor, isTopThreshold, isBottomThreshold)
		} else {
			// Default formatting if no formatter is provided
			itemStr := fmt.Sprintf("%v", item)
			prefix := fmt.Sprintf("%d - ", absoluteIndex)

			if isCursor {
				renderedItem = selectedStyle.Render(prefix + "[ " + itemStr + " ]")
			} else {
				renderedItem = rowStyle.Render(prefix + itemStr)
			}
		}

		builder.WriteString(renderedItem)

		// Add a newline unless it's the last item
		if i < len(l.visibleItems)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// RenderDebugInfo renders debug information about the list.
func (l *List[T]) RenderDebugInfo() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Total Items: %d\n", l.totalItems))
	builder.WriteString(fmt.Sprintf("Viewport Start: %d\n", l.State.ViewportStartIndex))
	builder.WriteString(fmt.Sprintf("Viewport End: %d\n", l.State.ViewportStartIndex+l.Config.Height))
	builder.WriteString(fmt.Sprintf("Cursor Index: %d\n", l.State.CursorIndex))
	builder.WriteString(fmt.Sprintf("Cursor Viewport Index: %d\n", l.State.CursorViewportIndex))
	builder.WriteString(fmt.Sprintf("At Top Threshold: %t\n", l.State.IsAtTopThreshold))
	builder.WriteString(fmt.Sprintf("At Bottom Threshold: %t\n", l.State.IsAtBottomThreshold))
	builder.WriteString(fmt.Sprintf("At Dataset Start: %t\n", l.State.AtDatasetStart))
	builder.WriteString(fmt.Sprintf("At Dataset End: %t\n", l.State.AtDatasetEnd))

	// Show information about loaded chunks
	builder.WriteString("Loaded Chunks:\n")
	for _, chunk := range l.chunks {
		builder.WriteString(fmt.Sprintf("  Chunk %d-%d (%d items)\n", chunk.StartIndex, chunk.EndIndex, len(chunk.Items)))
	}

	return builder.String()
}
