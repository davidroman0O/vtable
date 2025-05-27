package vtable

// ================================
// NAVIGATION CALCULATION FUNCTIONS
// ================================

// CalculateCursorUp calculates the new viewport state after moving cursor up one position
func CalculateCursorUp(viewport ViewportState, config ViewportConfig, totalItems int) ViewportState {
	if totalItems <= 0 || viewport.CursorIndex <= 0 {
		return viewport
	}

	result := viewport
	topThreshold := config.TopThreshold

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
	result = UpdateViewportBounds(result, config, totalItems)

	return result
}

// CalculateCursorDown calculates the new viewport state after moving cursor down one position
func CalculateCursorDown(viewport ViewportState, config ViewportConfig, totalItems int) ViewportState {
	if totalItems <= 0 || viewport.CursorIndex >= totalItems-1 {
		return viewport
	}

	result := viewport
	bottomThreshold := config.BottomThreshold

	// Handle bottom threshold logic (only if thresholds are enabled)
	if viewport.IsAtBottomThreshold && !viewport.AtDatasetEnd && bottomThreshold >= 0 {
		// Cursor was at the bottom threshold, scroll viewport down while keeping cursor at threshold
		result.ViewportStartIndex++
		bottomPosition := config.Height - bottomThreshold - 1
		result.CursorViewportIndex = bottomPosition // LOCK cursor at threshold
		// Update absolute cursor position based on new viewport
		result.CursorIndex = result.ViewportStartIndex + result.CursorViewportIndex
	} else if bottomThreshold < 0 {
		// Thresholds disabled - use pure edge-based scrolling
		result.CursorIndex++ // Move cursor normally
		// Move cursor within viewport if possible, otherwise scroll
		if viewport.CursorViewportIndex < config.Height-1 {
			// Cursor can move within viewport
			result.CursorViewportIndex++
		} else {
			// Cursor is at bottom edge of viewport - scroll if possible
			if result.ViewportStartIndex+config.Height < totalItems {
				result.ViewportStartIndex++
				result.CursorViewportIndex = config.Height - 1
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

		if viewport.CursorViewportIndex < config.Height-1 &&
			result.ViewportStartIndex+viewport.CursorViewportIndex+1 < totalItems {
			// Cursor not at threshold, move within viewport
			result.CursorViewportIndex++
		} else {
			// At viewport bottom edge, scroll if possible
			if result.ViewportStartIndex+config.Height < totalItems {
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
	result = UpdateViewportBounds(result, config, totalItems)

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
func CalculatePageUp(viewport ViewportState, config ViewportConfig, totalItems int) ViewportState {
	if totalItems <= 0 || viewport.CursorIndex <= 0 {
		return viewport
	}

	result := viewport

	// Move cursor up by a full page
	newCursorIndex := viewport.CursorIndex - config.Height
	if newCursorIndex < 0 {
		newCursorIndex = 0
	}

	result.CursorIndex = newCursorIndex

	// Position cursor at top threshold if thresholds are enabled
	if config.TopThreshold >= 0 {
		// Try to position cursor at top threshold
		result.ViewportStartIndex = newCursorIndex - config.TopThreshold
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
	result = UpdateViewportBounds(result, config, totalItems)

	return result
}

// CalculatePageDown calculates viewport state for page down with threshold awareness
func CalculatePageDown(viewport ViewportState, config ViewportConfig, totalItems int) ViewportState {
	if totalItems <= 0 || viewport.CursorIndex >= totalItems-1 {
		return viewport
	}

	result := viewport

	// Move cursor down by a full page
	newCursorIndex := viewport.CursorIndex + config.Height
	if newCursorIndex >= totalItems {
		newCursorIndex = totalItems - 1
	}

	result.CursorIndex = newCursorIndex

	// Position cursor at bottom threshold if thresholds are enabled
	if config.BottomThreshold >= 0 {
		// Try to position cursor at bottom threshold
		bottomPosition := config.Height - config.BottomThreshold - 1
		result.ViewportStartIndex = newCursorIndex - bottomPosition
		if result.ViewportStartIndex < 0 {
			result.ViewportStartIndex = 0
		}
		// Ensure viewport doesn't go beyond data
		if result.ViewportStartIndex+config.Height > totalItems {
			result.ViewportStartIndex = totalItems - config.Height
			if result.ViewportStartIndex < 0 {
				result.ViewportStartIndex = 0
			}
		}
		result.CursorViewportIndex = newCursorIndex - result.ViewportStartIndex
	} else {
		// No thresholds - position cursor at bottom of viewport
		result.ViewportStartIndex = newCursorIndex - config.Height + 1
		if result.ViewportStartIndex < 0 {
			result.ViewportStartIndex = 0
		}
		result.CursorViewportIndex = newCursorIndex - result.ViewportStartIndex
	}

	// Update bounds using existing function
	result = UpdateViewportBounds(result, config, totalItems)

	return result
}

// CalculateJumpToEnd calculates viewport state for jumping to the end of the dataset
func CalculateJumpToEnd(config ViewportConfig, totalItems int) ViewportState {
	if totalItems <= 0 {
		return ViewportState{}
	}

	result := ViewportState{
		CursorIndex: totalItems - 1,
	}

	// Calculate viewport start to show the cursor at the bottom threshold (or bottom if small dataset)
	if totalItems <= config.Height {
		result.ViewportStartIndex = 0
		result.CursorViewportIndex = totalItems - 1
	} else {
		result.ViewportStartIndex = totalItems - config.Height
		result.CursorViewportIndex = config.Height - 1
	}

	// Update bounds using existing function
	result = UpdateViewportBounds(result, config, totalItems)

	return result
}

// CalculateJumpToStart calculates viewport state for jumping to the start of the dataset
func CalculateJumpToStart(config ViewportConfig, totalItems int) ViewportState {
	if totalItems <= 0 {
		return ViewportState{}
	}

	result := ViewportState{
		CursorIndex:         0,
		ViewportStartIndex:  0,
		CursorViewportIndex: 0,
	}

	// Update bounds using existing function
	result = UpdateViewportBounds(result, config, totalItems)

	return result
}
