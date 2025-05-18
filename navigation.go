package vtable

// MoveUp moves the cursor up one position.
func (l *List[T]) MoveUp() {
	// Can't move up if already at the beginning
	if l.State.CursorIndex <= 0 {
		return
	}

	previousState := l.State

	// Update cursor index
	l.State.CursorIndex--

	// Handle top threshold logic
	if previousState.IsAtTopThreshold && !l.State.AtDatasetStart {
		// Cursor is at the top threshold, so we need to move the viewport up
		l.State.ViewportStartIndex--
		l.State.IsAtTopThreshold = true
		// Cursor stays at the top threshold row in the viewport
		l.State.CursorViewportIndex = l.Config.TopThresholdIndex
	} else if previousState.CursorViewportIndex > 0 {
		// Cursor is not at the top threshold, so just move it up within the viewport
		l.State.CursorViewportIndex--
		// Check if we're now at a threshold
		l.State.IsAtTopThreshold = l.State.CursorViewportIndex == l.Config.TopThresholdIndex
		l.State.IsAtBottomThreshold = l.State.CursorViewportIndex == l.Config.BottomThresholdIndex
	} else {
		// This handles the case when we're at the top of the viewport but not at the top threshold
		// This should rarely happen but is included for completeness
		l.State.ViewportStartIndex--
		// Cursor stays at the top row in the viewport
		l.State.CursorViewportIndex = 0
	}

	// Update dataset boundary flags
	l.State.AtDatasetStart = l.State.ViewportStartIndex == 0
	l.State.AtDatasetEnd = l.State.ViewportStartIndex+l.Config.Height >= l.totalItems

	// Update visible items if viewport changed
	if l.State.ViewportStartIndex != previousState.ViewportStartIndex {
		// Make sure chunks are loaded
		chunkStartIndex := (l.State.ViewportStartIndex / l.Config.ChunkSize) * l.Config.ChunkSize
		l.loadChunk(chunkStartIndex)
		l.updateVisibleItems()
		l.unloadChunks()
	}
}

// MoveDown moves the cursor down one position.
func (l *List[T]) MoveDown() {
	// Can't move down if already at the end
	if l.State.CursorIndex >= l.totalItems-1 {
		return
	}

	previousState := l.State

	// Update cursor index
	l.State.CursorIndex++

	// Handle bottom threshold logic
	if previousState.IsAtBottomThreshold && !l.State.AtDatasetEnd {
		// Cursor is at the bottom threshold, so we need to move the viewport down
		l.State.ViewportStartIndex++
		l.State.IsAtBottomThreshold = true
		// Cursor stays at the bottom threshold row in the viewport
		l.State.CursorViewportIndex = l.Config.BottomThresholdIndex
	} else if previousState.CursorViewportIndex < l.Config.Height-1 &&
		l.State.ViewportStartIndex+previousState.CursorViewportIndex+1 < l.totalItems {
		// Cursor is not at the bottom threshold, so just move it down within the viewport
		l.State.CursorViewportIndex++
		// Check if we're now at a threshold
		l.State.IsAtTopThreshold = l.State.CursorViewportIndex == l.Config.TopThresholdIndex
		l.State.IsAtBottomThreshold = l.State.CursorViewportIndex == l.Config.BottomThresholdIndex
	} else {
		// This handles the case when we're at the bottom of the viewport but not at the bottom threshold
		// This should rarely happen but is included for completeness
		if l.State.ViewportStartIndex+l.Config.Height < l.totalItems {
			l.State.ViewportStartIndex++
			// Cursor stays at the bottom row in the viewport
			l.State.CursorViewportIndex = l.Config.Height - 1
		}
	}

	// Update dataset boundary flags
	l.State.AtDatasetStart = l.State.ViewportStartIndex == 0
	l.State.AtDatasetEnd = l.State.ViewportStartIndex+l.Config.Height >= l.totalItems

	// Update visible items if viewport changed
	if l.State.ViewportStartIndex != previousState.ViewportStartIndex {
		// Make sure chunks are loaded
		chunkStartIndex := (l.State.ViewportStartIndex / l.Config.ChunkSize) * l.Config.ChunkSize
		l.loadChunk(chunkStartIndex)
		l.updateVisibleItems()
		l.unloadChunks()
	}
}

// PageUp moves the cursor up by a page (viewport height).
func (l *List[T]) PageUp() {
	// Can't move up if already at the beginning
	if l.State.CursorIndex <= 0 {
		return
	}

	previousState := l.State

	// Calculate how many items to move up
	moveCount := l.Config.Height

	// Don't move past the beginning
	if moveCount > l.State.CursorIndex {
		moveCount = l.State.CursorIndex
	}

	// Update cursor index
	l.State.CursorIndex -= moveCount

	// Calculate new viewport start
	if l.State.CursorIndex < l.State.ViewportStartIndex+l.Config.TopThresholdIndex {
		// Position the cursor at the top threshold
		l.State.ViewportStartIndex = l.State.CursorIndex - l.Config.TopThresholdIndex
		if l.State.ViewportStartIndex < 0 {
			l.State.ViewportStartIndex = 0
		}
		l.State.CursorViewportIndex = l.State.CursorIndex - l.State.ViewportStartIndex
		l.State.IsAtTopThreshold = l.State.CursorViewportIndex == l.Config.TopThresholdIndex
		l.State.IsAtBottomThreshold = l.State.CursorViewportIndex == l.Config.BottomThresholdIndex
	} else {
		// Keep cursor at current position in viewport
		l.State.ViewportStartIndex -= moveCount
		if l.State.ViewportStartIndex < 0 {
			l.State.ViewportStartIndex = 0
		}
		l.State.CursorViewportIndex = l.State.CursorIndex - l.State.ViewportStartIndex
		l.State.IsAtTopThreshold = l.State.CursorViewportIndex == l.Config.TopThresholdIndex
		l.State.IsAtBottomThreshold = l.State.CursorViewportIndex == l.Config.BottomThresholdIndex
	}

	// Update dataset boundary flags
	l.State.AtDatasetStart = l.State.ViewportStartIndex == 0
	l.State.AtDatasetEnd = l.State.ViewportStartIndex+l.Config.Height >= l.totalItems

	// Update visible items if viewport changed
	if l.State.ViewportStartIndex != previousState.ViewportStartIndex {
		// Make sure chunks are loaded
		chunkStartIndex := (l.State.ViewportStartIndex / l.Config.ChunkSize) * l.Config.ChunkSize
		l.loadChunk(chunkStartIndex)
		l.updateVisibleItems()
		l.unloadChunks()
	}
}

// PageDown moves the cursor down by a page (viewport height).
func (l *List[T]) PageDown() {
	// Can't move down if already at the end
	if l.State.CursorIndex >= l.totalItems-1 {
		return
	}

	previousState := l.State

	// Calculate how many items to move down
	moveCount := l.Config.Height

	// Don't move past the end
	if l.State.CursorIndex+moveCount >= l.totalItems {
		moveCount = l.totalItems - l.State.CursorIndex - 1
	}

	// Update cursor index
	l.State.CursorIndex += moveCount

	// Calculate new viewport start
	if l.State.CursorIndex > l.State.ViewportStartIndex+l.Config.BottomThresholdIndex {
		// Position the cursor at the bottom threshold
		newViewportStart := l.State.CursorIndex - l.Config.BottomThresholdIndex

		// Don't let viewport show beyond end of data
		maxViewportStart := l.totalItems - l.Config.Height
		if maxViewportStart < 0 {
			maxViewportStart = 0
		}

		if newViewportStart > maxViewportStart {
			newViewportStart = maxViewportStart
		}

		l.State.ViewportStartIndex = newViewportStart
		l.State.CursorViewportIndex = l.State.CursorIndex - l.State.ViewportStartIndex
		l.State.IsAtTopThreshold = l.State.CursorViewportIndex == l.Config.TopThresholdIndex
		l.State.IsAtBottomThreshold = l.State.CursorViewportIndex == l.Config.BottomThresholdIndex
	} else {
		// Keep cursor at current position in viewport
		l.State.ViewportStartIndex += moveCount

		// Don't let viewport show beyond end of data
		maxViewportStart := l.totalItems - l.Config.Height
		if maxViewportStart < 0 {
			maxViewportStart = 0
		}

		if l.State.ViewportStartIndex > maxViewportStart {
			l.State.ViewportStartIndex = maxViewportStart
		}

		l.State.CursorViewportIndex = l.State.CursorIndex - l.State.ViewportStartIndex
		l.State.IsAtTopThreshold = l.State.CursorViewportIndex == l.Config.TopThresholdIndex
		l.State.IsAtBottomThreshold = l.State.CursorViewportIndex == l.Config.BottomThresholdIndex
	}

	// Update dataset boundary flags
	l.State.AtDatasetStart = l.State.ViewportStartIndex == 0
	l.State.AtDatasetEnd = l.State.ViewportStartIndex+l.Config.Height >= l.totalItems

	// Update visible items if viewport changed
	if l.State.ViewportStartIndex != previousState.ViewportStartIndex {
		// Make sure chunks are loaded
		chunkStartIndex := (l.State.ViewportStartIndex / l.Config.ChunkSize) * l.Config.ChunkSize
		l.loadChunk(chunkStartIndex)
		l.updateVisibleItems()
		l.unloadChunks()
	}
}

// JumpToIndex jumps to the specified index in the dataset.
func (l *List[T]) JumpToIndex(index int) {
	// Ensure the index is within bounds
	if index < 0 {
		index = 0
	} else if index >= l.totalItems {
		index = l.totalItems - 1
	}

	previousState := l.State

	// Update cursor index
	l.State.CursorIndex = index

	// Calculate new viewport start
	// Try to position the cursor at the middle of the viewport
	middleViewportIndex := l.Config.Height / 2
	newViewportStart := index - middleViewportIndex

	// Don't let viewport start before beginning of data
	if newViewportStart < 0 {
		newViewportStart = 0
	}

	// Don't let viewport show beyond end of data
	maxViewportStart := l.totalItems - l.Config.Height
	if maxViewportStart < 0 {
		maxViewportStart = 0
	}

	if newViewportStart > maxViewportStart {
		newViewportStart = maxViewportStart
	}

	l.State.ViewportStartIndex = newViewportStart
	l.State.CursorViewportIndex = l.State.CursorIndex - l.State.ViewportStartIndex

	// Check if we're at a threshold
	l.State.IsAtTopThreshold = l.State.CursorViewportIndex == l.Config.TopThresholdIndex
	l.State.IsAtBottomThreshold = l.State.CursorViewportIndex == l.Config.BottomThresholdIndex

	// Update dataset boundary flags
	l.State.AtDatasetStart = l.State.ViewportStartIndex == 0
	l.State.AtDatasetEnd = l.State.ViewportStartIndex+l.Config.Height >= l.totalItems

	// Update visible items if viewport changed
	if l.State.ViewportStartIndex != previousState.ViewportStartIndex {
		// Make sure chunks are loaded
		chunkStartIndex := (l.State.ViewportStartIndex / l.Config.ChunkSize) * l.Config.ChunkSize
		l.loadChunk(chunkStartIndex)
		l.updateVisibleItems()
		l.unloadChunks()
	}
}

// JumpToStart jumps to the start of the dataset.
func (l *List[T]) JumpToStart() {
	l.JumpToIndex(0)
}

// JumpToEnd jumps to the end of the dataset.
func (l *List[T]) JumpToEnd() {
	l.JumpToIndex(l.totalItems - 1)
}

// JumpToItem jumps to an item with the specified key-value pair.
// This method requires a SearchableDataProvider.
// Returns true if the item was found and jumped to, false otherwise.
func (l *List[T]) JumpToItem(key string, value any) bool {
	// Check if the data provider supports searching
	searchable, ok := l.DataProvider.(SearchableDataProvider[T])
	if !ok {
		return false
	}

	// Find the item index
	index, found := searchable.FindItemIndex(key, value)
	if !found {
		return false
	}

	// Jump to the found index
	l.JumpToIndex(index)
	return true
}
