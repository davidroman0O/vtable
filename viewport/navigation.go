package viewport

import (
	"github.com/davidroman0O/vtable/core"
)

// ================================
// NAVIGATION CALCULATION FUNCTIONS
// ================================

// CalculateCursorUp calculates the new viewport state after moving cursor up one position
func CalculateCursorUp(viewport core.ViewportState, viewportConfig core.ViewportConfig, totalItems int) core.ViewportState {
	if totalItems <= 0 || viewport.CursorIndex <= 0 {
		return viewport
	}

	result := viewport
	topThreshold := viewportConfig.TopThreshold

	// Handle top threshold logic (only if thresholds are enabled)
	if viewport.IsAtTopThreshold && !viewport.AtDatasetStart && topThreshold >= 0 {
		// Cursor was at the top threshold, scroll viewport up while keeping cursor at threshold
		if result.ViewportStartIndex > 0 {
			result.ViewportStartIndex--
			result.CursorViewportIndex = topThreshold // LOCK cursor at threshold
			// Update absolute cursor position based on new viewport
			result.CursorIndex = result.ViewportStartIndex + result.CursorViewportIndex
		} else {
			// Can't scroll viewport up anymore, move cursor normally
			result.CursorIndex--
			result.CursorViewportIndex--
		}
	} else if topThreshold < 0 {
		// Thresholds disabled - use pure edge-based scrolling
		result.CursorIndex-- // Move cursor normally
		// Move cursor within viewport if possible, otherwise scroll
		if viewport.CursorViewportIndex > 0 {
			// Cursor can move within viewport
			result.CursorViewportIndex--
		} else {
			// Cursor is at top edge of viewport - scroll if possible
			if result.ViewportStartIndex > 0 {
				result.ViewportStartIndex--
				result.CursorViewportIndex = 0
			}
		}
	} else {
		// Thresholds enabled - move cursor normally, let viewport follow
		result.CursorIndex-- // Move cursor first

		if viewport.CursorViewportIndex > 0 {
			// Cursor not at threshold, move within viewport
			result.CursorViewportIndex--
		} else {
			// At viewport top edge, scroll if possible
			if result.ViewportStartIndex > 0 {
				result.ViewportStartIndex--
				result.CursorViewportIndex = 0
			} else {
				// Can't scroll, cursor stays at top
				result.CursorViewportIndex = 0
			}
		}
	}

	// Final safety check - ensure cursor doesn't go negative
	if result.CursorIndex < 0 {
		result.CursorIndex = 0
		result.CursorViewportIndex = 0
	}

	// Update bounds using existing function
	result = UpdateViewportBounds(result, viewportConfig, totalItems)

	return result
}

// CalculateCursorDown calculates the new viewport state after moving cursor down one position
func CalculateCursorDown(viewport core.ViewportState, viewportConfig core.ViewportConfig, totalItems int) core.ViewportState {
	if totalItems <= 0 || viewport.CursorIndex >= totalItems-1 {
		return viewport
	}

	result := viewport
	bottomThreshold := viewportConfig.BottomThreshold

	// Handle bottom threshold logic (only if thresholds are enabled)
	if viewport.IsAtBottomThreshold && !viewport.AtDatasetEnd && bottomThreshold >= 0 {
		// Cursor was at the bottom threshold, scroll viewport down while keeping cursor at threshold
		result.ViewportStartIndex++
		bottomPosition := viewportConfig.Height - bottomThreshold - 1
		result.CursorViewportIndex = bottomPosition // LOCK cursor at threshold
		// Update absolute cursor position based on new viewport
		result.CursorIndex = result.ViewportStartIndex + result.CursorViewportIndex
	} else if bottomThreshold < 0 {
		// Thresholds disabled - use pure edge-based scrolling
		result.CursorIndex++ // Move cursor normally
		// Move cursor within viewport if possible, otherwise scroll
		if viewport.CursorViewportIndex < viewportConfig.Height-1 {
			// Cursor can move within viewport
			result.CursorViewportIndex++
		} else {
			// Cursor is at bottom edge of viewport - scroll if possible
			if result.ViewportStartIndex+viewportConfig.Height < totalItems {
				result.ViewportStartIndex++
				result.CursorViewportIndex = viewportConfig.Height - 1
			}
		}
	} else {
		// Thresholds enabled - move cursor normally, let viewport follow
		result.CursorIndex++ // Move cursor first

		// Ensure we don't exceed actual data count
		if result.CursorIndex >= totalItems {
			result.CursorIndex = totalItems - 1
			// If we're already at the last item, no change needed
			if result.CursorIndex == viewport.CursorIndex {
				return viewport
			}
		}

		if viewport.CursorViewportIndex < viewportConfig.Height-1 &&
			result.ViewportStartIndex+viewport.CursorViewportIndex+1 < totalItems {
			// Cursor not at threshold, move within viewport
			result.CursorViewportIndex++
		} else {
			// At viewport bottom edge, scroll if possible
			if result.ViewportStartIndex+viewportConfig.Height < totalItems {
				result.ViewportStartIndex++
			}
			result.CursorViewportIndex = result.CursorIndex - result.ViewportStartIndex
		}
	}

	// Final boundary check - ensure we're not beyond data
	if result.CursorIndex >= totalItems {
		result.CursorIndex = totalItems - 1
		result.CursorViewportIndex = result.CursorIndex - result.ViewportStartIndex
	}

	// Ensure cursor viewport index is within bounds
	if result.CursorViewportIndex < 0 {
		result.CursorViewportIndex = 0
		result.CursorIndex = result.ViewportStartIndex
	}

	// Update bounds using existing function
	result = UpdateViewportBounds(result, viewportConfig, totalItems)

	return result
}

// CalculatePageMovement calculates cursor movement for page up/down operations
func CalculatePageMovement(currentIndex int, pageSize int, totalItems int, direction int) int {
	newIndex := currentIndex + (direction * pageSize)

	if newIndex < 0 {
		return 0
	}
	if newIndex >= totalItems {
		return totalItems - 1
	}

	return newIndex
}

// CalculatePageUp calculates viewport state for page up with threshold awareness
func CalculatePageUp(viewport core.ViewportState, viewportConfig core.ViewportConfig, totalItems int) core.ViewportState {
	if totalItems <= 0 || viewport.CursorIndex <= 0 {
		return viewport
	}

	result := viewport

	// Move cursor up by a full page
	newCursorIndex := viewport.CursorIndex - viewportConfig.Height
	if newCursorIndex < 0 {
		newCursorIndex = 0
	}

	result.CursorIndex = newCursorIndex

	// Position cursor at top threshold if thresholds are enabled
	if viewportConfig.TopThreshold >= 0 {
		// Try to position cursor at top threshold
		result.ViewportStartIndex = newCursorIndex - viewportConfig.TopThreshold
		if result.ViewportStartIndex < 0 {
			result.ViewportStartIndex = 0
		}
		result.CursorViewportIndex = newCursorIndex - result.ViewportStartIndex
	} else {
		// No thresholds - position cursor at top of viewport
		result.ViewportStartIndex = newCursorIndex
		result.CursorViewportIndex = 0
	}

	// Update bounds using existing function
	result = UpdateViewportBounds(result, viewportConfig, totalItems)

	return result
}

// CalculatePageDown calculates viewport state for page down with threshold awareness
func CalculatePageDown(viewport core.ViewportState, viewportConfig core.ViewportConfig, totalItems int) core.ViewportState {
	if totalItems <= 0 || viewport.CursorIndex >= totalItems-1 {
		return viewport
	}

	result := viewport

	// Move cursor down by a full page
	newCursorIndex := viewport.CursorIndex + viewportConfig.Height
	if newCursorIndex >= totalItems {
		newCursorIndex = totalItems - 1
	}

	result.CursorIndex = newCursorIndex

	// Position cursor at bottom threshold if thresholds are enabled
	if viewportConfig.BottomThreshold >= 0 {
		// Try to position cursor at bottom threshold
		bottomPosition := viewportConfig.Height - viewportConfig.BottomThreshold - 1
		result.ViewportStartIndex = newCursorIndex - bottomPosition
		if result.ViewportStartIndex < 0 {
			result.ViewportStartIndex = 0
		}
		// Ensure viewport doesn't go beyond data
		if result.ViewportStartIndex+viewportConfig.Height > totalItems {
			result.ViewportStartIndex = totalItems - viewportConfig.Height
			if result.ViewportStartIndex < 0 {
				result.ViewportStartIndex = 0
			}
		}
		result.CursorViewportIndex = newCursorIndex - result.ViewportStartIndex
	} else {
		// No thresholds - position cursor at bottom of viewport
		result.ViewportStartIndex = newCursorIndex - viewportConfig.Height + 1
		if result.ViewportStartIndex < 0 {
			result.ViewportStartIndex = 0
		}
		result.CursorViewportIndex = newCursorIndex - result.ViewportStartIndex
	}

	// Update bounds using existing function
	result = UpdateViewportBounds(result, viewportConfig, totalItems)

	return result
}

// CalculateJumpToEnd calculates viewport state for jumping to the end of the dataset
func CalculateJumpToEnd(viewportConfig core.ViewportConfig, totalItems int) core.ViewportState {
	if totalItems <= 0 {
		return core.ViewportState{}
	}

	result := core.ViewportState{
		CursorIndex: totalItems - 1,
	}

	// Calculate viewport start to show the cursor at the bottom threshold (or bottom if small dataset)
	if totalItems <= viewportConfig.Height {
		result.ViewportStartIndex = 0
		result.CursorViewportIndex = totalItems - 1
	} else {
		result.ViewportStartIndex = totalItems - viewportConfig.Height
		result.CursorViewportIndex = viewportConfig.Height - 1
	}

	// Update bounds using existing function
	result = UpdateViewportBounds(result, viewportConfig, totalItems)

	return result
}

// CalculateJumpToStart calculates viewport state for jumping to the start of the dataset
func CalculateJumpToStart(viewportConfig core.ViewportConfig, totalItems int) core.ViewportState {
	if totalItems <= 0 {
		return core.ViewportState{}
	}

	result := core.ViewportState{
		CursorIndex:         0,
		ViewportStartIndex:  0,
		CursorViewportIndex: 0,
	}

	// Update bounds using existing function
	result = UpdateViewportBounds(result, viewportConfig, totalItems)

	return result
}

// CalculateJumpTo calculates optimal viewport state for jumping to an arbitrary index
// This function ensures proper threshold positioning and centering when possible
func CalculateJumpTo(targetIndex int, viewportConfig core.ViewportConfig, totalItems int) core.ViewportState {
	if totalItems <= 0 {
		return core.ViewportState{}
	}

	// Ensure target index is within bounds
	if targetIndex < 0 {
		targetIndex = 0
	}
	if targetIndex >= totalItems {
		targetIndex = totalItems - 1
	}

	result := core.ViewportState{
		CursorIndex: targetIndex,
	}

	// Strategy: Try to position cursor optimally based on dataset position and thresholds

	// Case 1: Small dataset - show everything from start
	if totalItems <= viewportConfig.Height {
		result.ViewportStartIndex = 0
		result.CursorViewportIndex = targetIndex
		result = UpdateViewportBounds(result, viewportConfig, totalItems)
		return result
	}

	// Case 2: Near the beginning - position at start with cursor at top threshold if possible
	if targetIndex < viewportConfig.TopThreshold {
		result.ViewportStartIndex = 0
		result.CursorViewportIndex = targetIndex
	} else if targetIndex < viewportConfig.Height {
		// Still near beginning but can use threshold positioning
		result.ViewportStartIndex = 0
		result.CursorViewportIndex = targetIndex
	} else {
		// Case 3: Near the end - position at end with cursor at bottom threshold if possible
		if targetIndex >= totalItems-viewportConfig.BottomThreshold {
			// Very close to end - position viewport at end
			result.ViewportStartIndex = totalItems - viewportConfig.Height
			if result.ViewportStartIndex < 0 {
				result.ViewportStartIndex = 0
			}
			result.CursorViewportIndex = targetIndex - result.ViewportStartIndex
		} else if targetIndex >= totalItems-viewportConfig.Height {
			// Close to end but can use threshold positioning
			result.ViewportStartIndex = totalItems - viewportConfig.Height
			if result.ViewportStartIndex < 0 {
				result.ViewportStartIndex = 0
			}
			result.CursorViewportIndex = targetIndex - result.ViewportStartIndex
		} else {
			// Case 4: Middle of dataset - center cursor optimally
			// Try to position cursor at top threshold for best navigation experience
			if viewportConfig.TopThreshold >= 0 && viewportConfig.TopThreshold < viewportConfig.Height {
				result.ViewportStartIndex = targetIndex - viewportConfig.TopThreshold
				result.CursorViewportIndex = viewportConfig.TopThreshold
			} else {
				// No thresholds or invalid viewportConfig - center in viewport
				result.ViewportStartIndex = targetIndex - viewportConfig.Height/2
				result.CursorViewportIndex = viewportConfig.Height / 2
			}

			// Ensure viewport doesn't go negative
			if result.ViewportStartIndex < 0 {
				result.ViewportStartIndex = 0
				result.CursorViewportIndex = targetIndex
			}

			// Ensure viewport doesn't exceed dataset
			if result.ViewportStartIndex+viewportConfig.Height > totalItems {
				result.ViewportStartIndex = totalItems - viewportConfig.Height
				if result.ViewportStartIndex < 0 {
					result.ViewportStartIndex = 0
				}
				result.CursorViewportIndex = targetIndex - result.ViewportStartIndex
			}
		}
	}

	// Final safety checks
	if result.CursorViewportIndex < 0 {
		result.CursorViewportIndex = 0
		result.CursorIndex = result.ViewportStartIndex
	}
	if result.CursorViewportIndex >= viewportConfig.Height {
		result.CursorViewportIndex = viewportConfig.Height - 1
		result.CursorIndex = result.ViewportStartIndex + result.CursorViewportIndex
	}
	if result.CursorIndex >= totalItems {
		result.CursorIndex = totalItems - 1
		result.CursorViewportIndex = result.CursorIndex - result.ViewportStartIndex
	}

	// Update bounds using existing function
	result = UpdateViewportBounds(result, viewportConfig, totalItems)

	return result
}
