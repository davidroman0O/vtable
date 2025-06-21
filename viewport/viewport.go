// Package viewport provides the logic for managing the visible area of a component,
// handling scrolling, cursor movement, and calculating which data chunks are
// needed to display the current view. It is a core dependency for components
// like List and Table that virtualize their data.
package viewport

import "github.com/davidroman0O/vtable/core"

// BoundingArea represents the area around the viewport where data chunks should
// be pre-emptively loaded to ensure smooth scrolling. It is defined by absolute
// item indices and chunk boundaries.
//
// Deprecated: This struct is defined in the core package. This local definition
// is redundant and will be removed in a future version. Use core.BoundingArea instead.
type BoundingArea struct {
	StartIndex int // Absolute start index of bounding area
	EndIndex   int // Absolute end index of bounding area (inclusive)
	ChunkStart int // First chunk index in bounding area
	ChunkEnd   int // Last chunk index in bounding area
}

// UpdateViewportBounds calculates and updates the boundary flags of the viewport state.
// These flags (e.g., `IsAtTopThreshold`, `AtDatasetEnd`) are essential for
// controlling scroll behavior and providing feedback to the user. The function
// operates on a copy of the state, ensuring immutability.
func UpdateViewportBounds(viewport core.ViewportState, viewportConfig core.ViewportConfig, totalItems int) core.ViewportState {
	height := viewportConfig.Height
	topThreshold := viewportConfig.TopThreshold
	bottomThreshold := viewportConfig.BottomThreshold

	// Create a copy to avoid modifying the input
	result := viewport

	// Update threshold flags using offset semantics
	// TopThreshold: offset from viewport start (e.g., TopThreshold=2 means position 2)
	// BottomThreshold: offset from viewport end (e.g., BottomThreshold=2 means position height-2-1)
	result.IsAtTopThreshold = false
	result.IsAtBottomThreshold = false

	if topThreshold >= 0 && topThreshold < height {
		result.IsAtTopThreshold = result.CursorViewportIndex == topThreshold
	}

	if bottomThreshold >= 0 && bottomThreshold < height {
		// BottomThreshold is offset from end: if height=8 and bottomThreshold=2, then position is 8-2-1=5
		bottomPosition := height - bottomThreshold - 1
		if bottomPosition >= 0 && bottomPosition < height {
			result.IsAtBottomThreshold = result.CursorViewportIndex == bottomPosition
		}
	}

	// Update dataset boundary flags
	result.AtDatasetStart = result.ViewportStartIndex == 0
	result.AtDatasetEnd = result.ViewportStartIndex+height >= totalItems

	return result
}

// CalculateBoundingArea determines the range of data that should be loaded into
// memory, based on the current viewport position and the configured buffer sizes
// (`BoundingAreaBefore` and `BoundingAreaAfter`). This "bounding area" is larger
// than the visible viewport, allowing for seamless scrolling as data is pre-fetched.
func CalculateBoundingArea(viewport core.ViewportState, viewportConfig core.ViewportConfig, totalItems int) core.BoundingArea {
	if totalItems == 0 {
		return core.BoundingArea{}
	}

	chunkSize := viewportConfig.ChunkSize
	viewportHeight := viewportConfig.Height
	boundingBefore := viewportConfig.BoundingAreaBefore
	boundingAfter := viewportConfig.BoundingAreaAfter

	// Calculate viewport bounds (item indices)
	viewportStart := viewport.ViewportStartIndex
	viewportEnd := viewportStart + viewportHeight - 1

	// FULLY AUTOMATED BOUNDING AREA CALCULATION
	// Automatically calculate bounding area based on current viewport position
	// using the configured distances (boundingBefore/boundingAfter items)
	boundingStartIndex := viewportStart - boundingBefore
	boundingEndIndex := viewportEnd + boundingAfter

	// Clamp to dataset bounds
	if boundingStartIndex < 0 {
		boundingStartIndex = 0
	}
	if boundingEndIndex >= totalItems {
		boundingEndIndex = totalItems - 1
	}

	// Find which chunks intersect with this bounding area
	firstChunkStart := (boundingStartIndex / chunkSize) * chunkSize
	lastChunkStart := (boundingEndIndex / chunkSize) * chunkSize

	// ChunkEnd is the boundary for the loop (exclusive)
	chunkEnd := lastChunkStart + chunkSize

	return core.BoundingArea{
		StartIndex: boundingStartIndex,
		EndIndex:   boundingEndIndex,
		ChunkStart: firstChunkStart,
		ChunkEnd:   chunkEnd,
	}
}

// UpdateViewportPosition recalculates the viewport's start index based on the
// absolute cursor position. If the cursor has moved outside the visible area,
// this function adjusts the viewport to bring the cursor back into view. It's
// a key function for ensuring the cursor remains visible during navigation.
func UpdateViewportPosition(viewport core.ViewportState, viewportConfig core.ViewportConfig, totalItems int) core.ViewportState {
	if totalItems == 0 {
		return viewport
	}

	height := viewportConfig.Height
	result := viewport

	// Calculate relative position within viewport
	result.CursorViewportIndex = result.CursorIndex - result.ViewportStartIndex

	// Adjust viewport if cursor is outside
	if result.CursorViewportIndex < 0 {
		result.ViewportStartIndex = result.CursorIndex
		result.CursorViewportIndex = 0
	} else if result.CursorViewportIndex >= height {
		result.ViewportStartIndex = result.CursorIndex - height + 1
		result.CursorViewportIndex = height - 1
	}

	// Update bounds using the extracted function
	result = UpdateViewportBounds(result, viewportConfig, totalItems)

	return result
}
