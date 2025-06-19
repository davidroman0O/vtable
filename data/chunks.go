// Package data provides the core data handling capabilities for the vtable component.
// It includes functionalities for managing data requests, chunking, sorting, and caching,
// forming the backbone of the data virtualization layer. This package is designed to
// efficiently handle large datasets by loading data in manageable chunks, only when needed.
package data

import (
	"time"

	"github.com/davidroman0O/vtable/core"
)

// IsLoadingCriticalChunks checks if any of the currently loading chunks are
// "critical" to the current viewport. A chunk is considered critical if it
// is currently visible or in the immediate overscan area. This function helps
// decide whether to block user interactions, like scrolling, while essential
// data is being fetched, preventing a disjointed user experience.
func IsLoadingCriticalChunks(viewport core.ViewportState, viewportConfig core.ViewportConfig, loadingChunks map[int]bool) bool {
	chunkSize := viewportConfig.ChunkSize
	viewportStart := viewport.ViewportStartIndex
	viewportEnd := viewportStart + viewportConfig.Height

	for chunkStart := range loadingChunks {
		chunkEnd := chunkStart + chunkSize
		// Check if this loading chunk overlaps with viewport
		if !(chunkEnd <= viewportStart || chunkStart >= viewportEnd) {
			return true
		}
	}
	return false
}

// CalculateUnloadBounds determines the range of chunks that should be kept in
// memory based on the current viewport position. Chunks outside this range are
// candidates for unloading to conserve memory. The bounds are calculated to
// include the current viewport's chunk and a buffer of one chunk before and
// after.
func CalculateUnloadBounds(viewport core.ViewportState, viewportConfig core.ViewportConfig) (keepLowerBound, keepUpperBound int) {
	viewportChunkIndex := (viewport.ViewportStartIndex / viewportConfig.ChunkSize) * viewportConfig.ChunkSize
	keepLowerBound = viewportChunkIndex - viewportConfig.ChunkSize
	if keepLowerBound < 0 {
		keepLowerBound = 0
	}
	keepUpperBound = viewportChunkIndex + (2 * viewportConfig.ChunkSize)
	return keepLowerBound, keepUpperBound
}

// CalculateVisibleItemsBounds calculates the exact boundaries of the items that
// should be visible in the viewport. It considers the viewport's configured height
// and the total number of items in the dataset to prevent out-of-bounds errors.
// It returns the number of items that can be displayed, the end index of the
// viewport, and the maximum possible start index.
func CalculateVisibleItemsBounds(viewport core.ViewportState, viewportConfig core.ViewportConfig, totalItems int) (maxVisibleItems, viewportEnd, maxStart int) {
	// Calculate how many actual items we can show
	maxVisibleItems = viewportConfig.Height
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

// CalculateChunkStartIndex computes the starting index of the data chunk that
// contains the given item index. This is a pure calculation based on the item's
// position and the uniform chunk size.
func CalculateChunkStartIndex(itemIndex, chunkSize int) int {
	return (itemIndex / chunkSize) * chunkSize
}

// ShouldUnloadChunk determines whether a specific chunk should be unloaded based
// on pre-calculated keep-alive bounds. This is used in memory management routines
// to decide which chunks to discard.
func ShouldUnloadChunk(chunkStart, keepLowerBound, keepUpperBound int) bool {
	return chunkStart < keepLowerBound || chunkStart > keepUpperBound
}

// IsChunkLoaded checks if a data chunk containing a specific item index is
// currently loaded in memory. This is a convenience function to quickly determine
// if a data fetch is needed for a given item.
func IsChunkLoaded[T any](index int, chunks map[int]core.Chunk[T]) bool {
	for _, chunk := range chunks {
		if index >= chunk.StartIndex && index <= chunk.EndIndex {
			return true
		}
	}
	return false
}

// GetItemAtIndex retrieves a single item from the loaded chunks by its absolute index.
// It searches through the in-memory chunks to find the item. If the chunk is part
// of a larger caching strategy, this function also updates the chunk's last access
// time for LRU (Least Recently Used) cache eviction policies.
func GetItemAtIndex[T any](index int, chunks map[int]core.Chunk[T], totalItems int, chunkAccessTime map[int]time.Time) (core.Data[T], bool) {
	if index < 0 || index >= totalItems {
		return core.Data[T]{}, false
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

	return core.Data[T]{}, false
}

// FindItemIndex searches for an item by its unique ID across all loaded chunks.
// It returns the absolute index of the item if found, or -1 if the item is not
// present in any of the currently loaded chunks.
func FindItemIndex[T any](id string, chunks map[int]core.Chunk[T]) int {
	for _, chunk := range chunks {
		for i, item := range chunk.Items {
			if item.ID == id {
				return chunk.StartIndex + i
			}
		}
	}
	return -1
}
