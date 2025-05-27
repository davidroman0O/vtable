package vtable

import "time"

// ================================
// CHUNK MANAGEMENT FUNCTIONS
// ================================

// IsLoadingCriticalChunks checks if we're loading chunks that affect the current viewport
func IsLoadingCriticalChunks(viewport ViewportState, config ViewportConfig, loadingChunks map[int]bool) bool {
	chunkSize := config.ChunkSize
	viewportStart := viewport.ViewportStartIndex
	viewportEnd := viewportStart + config.Height

	for chunkStart := range loadingChunks {
		chunkEnd := chunkStart + chunkSize
		// Check if this loading chunk overlaps with viewport
		if !(chunkEnd <= viewportStart || chunkStart >= viewportEnd) {
			return true
		}
	}
	return false
}

// CalculateUnloadBounds calculates which chunks should be kept based on viewport position
func CalculateUnloadBounds(viewport ViewportState, config ViewportConfig) (keepLowerBound, keepUpperBound int) {
	viewportChunkIndex := (viewport.ViewportStartIndex / config.ChunkSize) * config.ChunkSize
	keepLowerBound = viewportChunkIndex - config.ChunkSize
	if keepLowerBound < 0 {
		keepLowerBound = 0
	}
	keepUpperBound = viewportChunkIndex + (2 * config.ChunkSize)
	return keepLowerBound, keepUpperBound
}

// CalculateVisibleItemsBounds calculates the bounds for visible items
func CalculateVisibleItemsBounds(viewport ViewportState, config ViewportConfig, totalItems int) (maxVisibleItems, viewportEnd, maxStart int) {
	// Calculate how many actual items we can show
	maxVisibleItems = config.Height
	if totalItems < maxVisibleItems {
		maxVisibleItems = totalItems
	}

	// Ensure viewport doesn't extend beyond dataset
	maxStart = totalItems - maxVisibleItems
	if maxStart < 0 {
		maxStart = 0
	}

	// Calculate endpoint of visible area (exclusive)
	viewportEnd = viewport.ViewportStartIndex + maxVisibleItems
	if viewportEnd > totalItems {
		viewportEnd = totalItems
	}

	return maxVisibleItems, viewportEnd, maxStart
}

// CalculateChunkStartIndex calculates the start index of the chunk containing the given item index
func CalculateChunkStartIndex(itemIndex, chunkSize int) int {
	return (itemIndex / chunkSize) * chunkSize
}

// ShouldUnloadChunk determines if a chunk should be unloaded based on bounds
func ShouldUnloadChunk(chunkStart, keepLowerBound, keepUpperBound int) bool {
	return chunkStart < keepLowerBound || chunkStart > keepUpperBound
}

// ================================
// CHUNK ACCESS FUNCTIONS
// ================================

// IsChunkLoaded checks if a chunk containing the given index is loaded
func IsChunkLoaded[T any](index int, chunks map[int]Chunk[T]) bool {
	for _, chunk := range chunks {
		if index >= chunk.StartIndex && index <= chunk.EndIndex {
			return true
		}
	}
	return false
}

// GetItemAtIndex retrieves an item at a specific index
func GetItemAtIndex[T any](index int, chunks map[int]Chunk[T], totalItems int, chunkAccessTime map[int]time.Time) (Data[T], bool) {
	if index < 0 || index >= totalItems {
		return Data[T]{}, false
	}

	// Find chunk containing this index
	for chunkStart, chunk := range chunks {
		if index >= chunk.StartIndex && index <= chunk.EndIndex {
			// Update access time for LRU management
			if chunkAccessTime != nil {
				chunkAccessTime[chunkStart] = time.Now()
			}

			chunkIndex := index - chunk.StartIndex
			if chunkIndex < len(chunk.Items) {
				return chunk.Items[chunkIndex], true
			}
		}
	}

	return Data[T]{}, false
}

// FindItemIndex finds the index of an item by ID
func FindItemIndex[T any](id string, chunks map[int]Chunk[T]) int {
	for _, chunk := range chunks {
		for i, item := range chunk.Items {
			if item.ID == id {
				return chunk.StartIndex + i
			}
		}
	}
	return -1
}
