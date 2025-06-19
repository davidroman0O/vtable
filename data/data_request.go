// Package data provides the core data handling capabilities for the vtable component.
// It includes functionalities for managing data requests, chunking, sorting, and caching,
// forming the backbone of the data virtualization layer. This package is designed to
// efficiently handle large datasets by loading data in manageable chunks, only when needed.
package data

import (
	"github.com/davidroman0O/vtable/core"
)

// CreateDataRequest creates a standardized data request with the given parameters.
// This function is a simple factory for core.DataRequest, ensuring consistency
// across the application when requesting data.
func CreateDataRequest(start, count int, sortFields, sortDirections []string, filters map[string]any) core.DataRequest {
	return core.DataRequest{
		Start:          start,
		Count:          count,
		SortFields:     sortFields,
		SortDirections: sortDirections,
		Filters:        filters,
	}
}

// CalculateActualChunkSize calculates the actual chunk size, accounting for the end of the dataset.
// When the requested chunk size would exceed the total number of items, this function
// adjusts the size to prevent requesting non-existent data.
func CalculateActualChunkSize(chunkStart, chunkSize, totalItems int) int {
	actualChunkSize := chunkSize
	if chunkStart+chunkSize > totalItems {
		actualChunkSize = totalItems - chunkStart
	}
	return actualChunkSize
}

// CreateChunkRequest creates a data request for a specific chunk.
// It uses CalculateActualChunkSize to ensure the request is valid and
// properly bounded by the total number of items.
func CreateChunkRequest(chunkStart, chunkSize, totalItems int, sortFields, sortDirections []string, filters map[string]any) core.DataRequest {
	actualChunkSize := CalculateActualChunkSize(chunkStart, chunkSize, totalItems)
	return CreateDataRequest(chunkStart, actualChunkSize, sortFields, sortDirections, filters)
}

// CalculateChunksInBoundingArea calculates which chunks need to be loaded for a given bounding area.
// A bounding area defines the visible data range, and this function identifies all chunks
// that fall within this area, which are candidates for loading.
func CalculateChunksInBoundingArea(boundingArea core.BoundingArea, chunkSize, totalItems int) []int {
	var chunks []int

	for chunkStart := boundingArea.ChunkStart; chunkStart < boundingArea.ChunkEnd; chunkStart += chunkSize {
		if chunkStart >= totalItems {
			break // Don't include chunks beyond dataset
		}
		chunks = append(chunks, chunkStart)
	}

	return chunks
}

// CheckChunkIntersection checks if a chunk intersects with a bounding area.
// This is crucial for determining which loaded chunks are still relevant to the
// current viewport and which can potentially be unloaded.
func CheckChunkIntersection(chunkStart, chunkSize int, boundingArea core.BoundingArea) bool {
	chunkEnd := chunkStart + chunkSize - 1
	// A chunk intersects if: chunkStart <= boundingArea.EndIndex AND chunkEnd >= boundingArea.StartIndex
	return chunkStart <= boundingArea.EndIndex && chunkEnd >= boundingArea.StartIndex
}

// FindChunksToUnload finds chunks that should be unloaded based on the current bounding area.
// It iterates through all currently loaded chunks and identifies those that no longer
// intersect with the visible area, marking them for removal to free up resources.
func FindChunksToUnload(loadedChunks map[int]core.Chunk[any], boundingArea core.BoundingArea, chunkSize int) []int {
	var chunksToUnload []int

	for chunkStart := range loadedChunks {
		if !CheckChunkIntersection(chunkStart, chunkSize, boundingArea) {
			chunksToUnload = append(chunksToUnload, chunkStart)
		}
	}

	return chunksToUnload
}

// CreatePlaceholderItem creates a placeholder item for missing data.
// When data is being loaded, these placeholders can be used to populate the
// UI, indicating that content is on its way.
func CreatePlaceholderItem(index int, itemType string) core.Data[any] {
	return core.Data[any]{
		ID:   itemType + "-" + string(rune(index)),
		Item: itemType + " item " + string(rune(index)),
	}
}

// LoadingState represents the current loading state of data chunks.
// It tracks which chunks are currently being fetched, and whether the UI
// should be allowed to scroll based on the loading status of critical chunks.
type LoadingState struct {
	LoadingChunks    map[int]bool // A map to track chunks that are in the process of being loaded.
	HasLoadingChunks bool         // A flag indicating if there are any chunks currently loading.
	CanScroll        bool         // A flag that determines if the user is allowed to scroll.
}

// UpdateLoadingState updates the loading state based on new loading chunks.
// It adds new chunks to the loading set and recalculates the `HasLoadingChunks`
// and `CanScroll` flags accordingly.
func UpdateLoadingState(currentState LoadingState, newLoadingChunks []int, viewport core.ViewportState, config core.ViewportConfig) LoadingState {
	newState := currentState

	// Add new loading chunks
	for _, chunkStart := range newLoadingChunks {
		newState.LoadingChunks[chunkStart] = true
	}

	// Update flags
	newState.HasLoadingChunks = len(newState.LoadingChunks) > 0
	if newState.HasLoadingChunks {
		// Block scrolling if we're loading chunks that affect current viewport
		newState.CanScroll = !IsLoadingCriticalChunks(viewport, config, newState.LoadingChunks)
	} else {
		newState.CanScroll = true
	}

	return newState
}

// ClearLoadingChunk removes a chunk from the loading state.
// This is called when a chunk has finished loading, and it updates the
// state to reflect that this chunk is no longer in-flight.
func ClearLoadingChunk(currentState LoadingState, chunkStart int) LoadingState {
	newState := currentState
	delete(newState.LoadingChunks, chunkStart)
	newState.HasLoadingChunks = len(newState.LoadingChunks) > 0

	if !newState.HasLoadingChunks {
		newState.CanScroll = true
	}

	return newState
}
