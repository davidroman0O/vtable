package vtable

import "fmt"

// ================================
// VISIBLE ITEMS FUNCTIONS
// ================================

// VisibleItemsResult represents the result of calculating visible items
type VisibleItemsResult struct {
	Items               []Data[any]
	AdjustedViewport    ViewportState
	PlaceholdersCreated int
}

// CreateLoadingPlaceholder creates a loading placeholder item
func CreateLoadingPlaceholder(index int) Data[any] {
	return Data[any]{
		ID:   fmt.Sprintf("loading-%d", index),
		Item: fmt.Sprintf("Loading item %d...", index),
	}
}

// CreateMissingPlaceholder creates a missing item placeholder
func CreateMissingPlaceholder(index int) Data[any] {
	return Data[any]{
		ID:   fmt.Sprintf("missing-%d", index),
		Item: fmt.Sprintf("Missing item %d", index),
	}
}

// CalculateVisibleItemsFromChunks calculates visible items from loaded chunks
func CalculateVisibleItemsFromChunks(
	viewport ViewportState,
	config ViewportConfig,
	totalItems int,
	chunks map[int]Chunk[any],
	ensureChunkLoaded func(int),
) VisibleItemsResult {
	if totalItems == 0 {
		return VisibleItemsResult{
			Items:            []Data[any]{},
			AdjustedViewport: viewport,
		}
	}

	// Calculate bounds using existing function
	_, viewportEnd, maxStart := CalculateVisibleItemsBounds(viewport, config, totalItems)

	// Ensure viewport doesn't extend beyond dataset
	adjustedViewport := viewport
	if adjustedViewport.ViewportStartIndex > maxStart {
		adjustedViewport.ViewportStartIndex = maxStart
	}
	if adjustedViewport.ViewportStartIndex < 0 {
		adjustedViewport.ViewportStartIndex = 0
	}

	// Create a new slice to hold visible items
	visibleItems := make([]Data[any], 0, viewportEnd-adjustedViewport.ViewportStartIndex)
	placeholdersCreated := 0

	// Fill the visible items slice with actual data
	for i := adjustedViewport.ViewportStartIndex; i < viewportEnd; i++ {
		// Get the chunk that contains this item
		chunkStartIndex := CalculateChunkStartIndex(i, config.ChunkSize)
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

// ValidateVisibleItemsBounds ensures cursor stays within visible items bounds
func ValidateVisibleItemsBounds(viewport ViewportState, visibleItemsCount int) ViewportState {
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
