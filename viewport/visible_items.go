// Package viewport provides the logic for managing the visible area of a component,
// handling scrolling, cursor movement, and calculating which data chunks are
// needed to display the current view. It is a core dependency for components
// like List and Table that virtualize their data.
package viewport

import (
	"fmt"

	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/data"
)

// VisibleItemsResult holds the outcome of calculating the set of items that
// should be currently visible in the viewport. It includes the slice of items
// (which may contain placeholders), an adjusted viewport state, and a count of
// any placeholders that were created.
type VisibleItemsResult struct {
	Items               []core.Data[any]   // The slice of items to be displayed.
	AdjustedViewport    core.ViewportState // The viewport state, possibly corrected for boundary conditions.
	PlaceholdersCreated int                // The number of placeholder items created due to missing data.
}

// CreateLoadingPlaceholder generates a placeholder `core.Data` item that indicates
// its content is currently being loaded. This is used to populate the viewport
// while waiting for a data chunk to arrive.
func CreateLoadingPlaceholder(index int) core.Data[any] {
	return core.Data[any]{
		ID:   fmt.Sprintf("loading-%d", index),
		Item: fmt.Sprintf("Loading item %d...", index),
	}
}

// CreateMissingPlaceholder generates a placeholder `core.Data` item for data that
// could not be found, even after a load attempt. This indicates a potential
// inconsistency between the expected total items and the actual data available.
func CreateMissingPlaceholder(index int) core.Data[any] {
	return core.Data[any]{
		ID:   fmt.Sprintf("missing-%d", index),
		Item: fmt.Sprintf("Missing item %d", index),
	}
}

// CalculateVisibleItemsFromChunks constructs the slice of items that should be
// visible in the viewport based on the current set of loaded data chunks. If an
// item is not found in the loaded chunks, it can trigger an immediate load via
// the `ensureChunkLoaded` callback and, if still unavailable, will generate a
// placeholder item. The function returns the final slice of items and an
// adjusted viewport state.
func CalculateVisibleItemsFromChunks(
	viewport core.ViewportState,
	config core.ViewportConfig,
	totalItems int,
	chunks map[int]core.Chunk[any],
	ensureChunkLoaded func(int),
) VisibleItemsResult {
	if totalItems == 0 {
		return VisibleItemsResult{
			Items:            []core.Data[any]{},
			AdjustedViewport: viewport,
		}
	}

	// Calculate bounds using local function
	_, viewportEnd, maxStart := data.CalculateVisibleItemsBounds(viewport, config, totalItems)

	// Ensure viewport doesn't extend beyond dataset
	adjustedViewport := viewport
	if adjustedViewport.ViewportStartIndex > maxStart {
		adjustedViewport.ViewportStartIndex = maxStart
	}
	if adjustedViewport.ViewportStartIndex < 0 {
		adjustedViewport.ViewportStartIndex = 0
	}

	// Create a new slice to hold visible items
	visibleItems := make([]core.Data[any], 0, viewportEnd-adjustedViewport.ViewportStartIndex)
	placeholdersCreated := 0

	// Fill the visible items slice with actual data
	for i := adjustedViewport.ViewportStartIndex; i < viewportEnd; i++ {
		// Get the chunk that contains this item
		chunkStartIndex := data.CalculateChunkStartIndex(i, config.ChunkSize)
		chunk, ok := chunks[chunkStartIndex]

		// If chunk isn't loaded, try to load it immediately
		if !ok && ensureChunkLoaded != nil {
			ensureChunkLoaded(chunkStartIndex)
			// Try to get the chunk again after loading
			chunk, ok = chunks[chunkStartIndex]
		}

		// If we still don't have the chunk, create a placeholder
		if !ok {
			visibleItems = append(visibleItems, CreateLoadingPlaceholder(i))
			placeholdersCreated++
			continue
		}

		// Calculate item index within the chunk
		itemIndex := i - chunk.StartIndex

		// Only add the item if it's within the chunk's bounds
		if itemIndex >= 0 && itemIndex < len(chunk.Items) {
			visibleItems = append(visibleItems, chunk.Items[itemIndex])
		} else {
			// Item not in chunk bounds - create placeholder
			visibleItems = append(visibleItems, CreateMissingPlaceholder(i))
			placeholdersCreated++
		}
	}

	// Adjust cursor if needed
	if adjustedViewport.CursorViewportIndex >= len(visibleItems) {
		if len(visibleItems) > 0 {
			adjustedViewport.CursorViewportIndex = len(visibleItems) - 1
		} else {
			adjustedViewport.CursorViewportIndex = 0
		}
		// Adjust absolute cursor position
		adjustedViewport.CursorIndex = adjustedViewport.ViewportStartIndex + adjustedViewport.CursorViewportIndex
	}

	return VisibleItemsResult{
		Items:               visibleItems,
		AdjustedViewport:    adjustedViewport,
		PlaceholdersCreated: placeholdersCreated,
	}
}

// ValidateVisibleItemsBounds corrects the cursor position if it falls outside
// the bounds of the currently visible items. This is a crucial safety check to
// prevent out-of-bounds panics when the number of visible items changes, for
// example, after a data refresh or filter operation.
func ValidateVisibleItemsBounds(viewport core.ViewportState, visibleItemsCount int) core.ViewportState {
	result := viewport

	if result.CursorViewportIndex >= visibleItemsCount {
		if visibleItemsCount > 0 {
			result.CursorViewportIndex = visibleItemsCount - 1
		} else {
			result.CursorViewportIndex = 0
		}
		// Adjust absolute cursor position
		result.CursorIndex = result.ViewportStartIndex + result.CursorViewportIndex
	}

	return result
}
