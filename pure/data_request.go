package vtable

// ================================
// DATA REQUEST FUNCTIONS
// ================================

// CreateDataRequest creates a standardized data request with the given parameters
func CreateDataRequest(start, count int, sortFields, sortDirections []string, filters map[string]any) DataRequest {
	return DataRequest{
		Start:          start,
		Count:          count,
		SortFields:     sortFields,
		SortDirections: sortDirections,
		Filters:        filters,
	}
}

// CalculateActualChunkSize calculates the actual chunk size, accounting for end-of-dataset
func CalculateActualChunkSize(chunkStart, chunkSize, totalItems int) int {
	actualChunkSize := chunkSize
	if chunkStart+chunkSize > totalItems {
		actualChunkSize = totalItems - chunkStart
	}
	return actualChunkSize
}

// CreateChunkRequest creates a data request for a specific chunk
func CreateChunkRequest(chunkStart, chunkSize, totalItems int, sortFields, sortDirections []string, filters map[string]any) DataRequest {
	actualChunkSize := CalculateActualChunkSize(chunkStart, chunkSize, totalItems)
	return CreateDataRequest(chunkStart, actualChunkSize, sortFields, sortDirections, filters)
}

// CalculateChunksInBoundingArea calculates which chunks need to be loaded for a bounding area
func CalculateChunksInBoundingArea(boundingArea BoundingArea, chunkSize, totalItems int) []int {
	var chunks []int

	for chunkStart := boundingArea.ChunkStart; chunkStart < boundingArea.ChunkEnd; chunkStart += chunkSize {
		if chunkStart >= totalItems {
			break // Don't include chunks beyond dataset
		}
		chunks = append(chunks, chunkStart)
	}

	return chunks
}

// CheckChunkIntersection checks if a chunk intersects with a bounding area
func CheckChunkIntersection(chunkStart, chunkSize int, boundingArea BoundingArea) bool {
	chunkEnd := chunkStart + chunkSize - 1
	// A chunk intersects if: chunkStart <= boundingArea.EndIndex AND chunkEnd >= boundingArea.StartIndex
	return chunkStart <= boundingArea.EndIndex && chunkEnd >= boundingArea.StartIndex
}

// FindChunksToUnload finds chunks that should be unloaded based on bounding area
func FindChunksToUnload(loadedChunks map[int]Chunk[any], boundingArea BoundingArea, chunkSize int) []int {
	var chunksToUnload []int

	for chunkStart := range loadedChunks {
		if !CheckChunkIntersection(chunkStart, chunkSize, boundingArea) {
			chunksToUnload = append(chunksToUnload, chunkStart)
		}
	}

	return chunksToUnload
}

// CreatePlaceholderItem creates a placeholder item for missing data
func CreatePlaceholderItem(index int, itemType string) Data[any] {
	return Data[any]{
		ID:   itemType + "-" + string(rune(index)),
		Item: itemType + " item " + string(rune(index)),
	}
}

// LoadingState represents the current loading state
type LoadingState struct {
	LoadingChunks    map[int]bool
	HasLoadingChunks bool
	CanScroll        bool
}

// UpdateLoadingState updates the loading state based on new loading chunks
func UpdateLoadingState(currentState LoadingState, newLoadingChunks []int, viewport ViewportState, config ViewportConfig) LoadingState {
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

// ClearLoadingChunk removes a chunk from the loading state
func ClearLoadingChunk(currentState LoadingState, chunkStart int) LoadingState {
	newState := currentState
	delete(newState.LoadingChunks, chunkStart)
	newState.HasLoadingChunks = len(newState.LoadingChunks) > 0

	if !newState.HasLoadingChunks {
		newState.CanScroll = true
	}

	return newState
}
