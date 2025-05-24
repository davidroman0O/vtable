package vtable

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Chunk represents a chunk of data loaded from the data provider.
type chunk[T any] struct {
	// StartIndex is the absolute index of the first item in the chunk.
	StartIndex int

	// EndIndex is the absolute index of the last item in the chunk.
	EndIndex int

	// Items is the slice of Data items in the chunk.
	Items []Data[T]
}

// List is a virtualized list component that efficiently handles large datasets.
type List[T any] struct {
	// Config is the configuration for the viewport.
	Config ViewportConfig

	// DataProvider is the data provider for the list.
	DataProvider DataProvider[T]

	// StyleConfig defines the styles for the list.
	StyleConfig StyleConfig

	// State is the current state of the viewport.
	State ViewportState

	// Formatter is the function used to format items.
	Formatter ItemFormatter[T]

	// chunks is a map of chunks loaded from the data provider.
	// The key is the start index of the chunk.
	chunks map[int]*chunk[T]

	// totalItems is the total number of items in the dataset.
	totalItems int

	// totalItemsValid tracks whether the cached totalItems is still valid
	// This prevents unnecessary GetTotal() calls when the dataset hasn't changed
	totalItemsValid bool

	// visibleItems is the slice of Data items currently visible in the viewport.
	visibleItems []Data[T]

	// dataRequest holds the current data request configuration
	// including filtering and sorting options
	dataRequest DataRequest

	// lastDataRequest tracks the last data request to detect when filters/sorts change
	lastDataRequest DataRequest
}

// NewList creates a new virtualized list component.
func NewList[T any](
	config ViewportConfig,
	provider DataProvider[T],
	styleConfig StyleConfig,
	formatter ItemFormatter[T],
) (*List[T], error) {
	if config.Height <= 0 {
		return nil, fmt.Errorf("viewport height must be greater than 0")
	}

	if config.TopThresholdIndex < 0 || config.TopThresholdIndex >= config.Height {
		return nil, fmt.Errorf("top threshold must be within viewport bounds")
	}

	if config.BottomThresholdIndex < 0 || config.BottomThresholdIndex >= config.Height {
		return nil, fmt.Errorf("bottom threshold must be within viewport bounds")
	}

	if config.BottomThresholdIndex <= config.TopThresholdIndex {
		return nil, fmt.Errorf("bottom threshold must be below top threshold")
	}

	if config.ChunkSize <= 0 {
		return nil, fmt.Errorf("chunk size must be greater than 0")
	}

	totalItems := provider.GetTotal()
	if totalItems <= 0 {
		return nil, fmt.Errorf("dataset must have at least one item")
	}

	initialIndex := config.InitialIndex
	if initialIndex < 0 || initialIndex >= totalItems {
		initialIndex = 0
	}

	viewportStartIndex := 0
	cursorViewportIndex := initialIndex

	// Adjust if initialIndex would place cursor beyond viewport
	if initialIndex >= config.Height {
		viewportStartIndex = initialIndex - config.TopThresholdIndex
		cursorViewportIndex = config.TopThresholdIndex
	}

	// Initialize with default data request
	dataRequest := DataRequest{
		Start:          0,
		Count:          config.ChunkSize,
		Filters:        make(map[string]any),
		SortFields:     []string{},
		SortDirections: []string{},
	}

	list := &List[T]{
		Config:          config,
		DataProvider:    provider,
		StyleConfig:     styleConfig,
		Formatter:       formatter,
		chunks:          make(map[int]*chunk[T]),
		totalItems:      totalItems,
		totalItemsValid: true, // Valid since we just fetched it
		visibleItems:    make([]Data[T], 0, config.Height),
		State: ViewportState{
			ViewportStartIndex:  viewportStartIndex,
			CursorIndex:         initialIndex,
			CursorViewportIndex: cursorViewportIndex,
			IsAtTopThreshold:    cursorViewportIndex == config.TopThresholdIndex,
			IsAtBottomThreshold: cursorViewportIndex == config.BottomThresholdIndex,
			AtDatasetStart:      viewportStartIndex == 0,
			AtDatasetEnd:        viewportStartIndex+config.Height >= totalItems,
		},
		dataRequest: dataRequest,
		// IMPORTANT: Deep copy dataRequest to avoid reference sharing
		lastDataRequest: DataRequest{
			Start:          dataRequest.Start,
			Count:          dataRequest.Count,
			SortFields:     append([]string(nil), dataRequest.SortFields...),
			SortDirections: append([]string(nil), dataRequest.SortDirections...),
			Filters:        make(map[string]any),
		},
	}

	// Copy filters map for lastDataRequest
	for k, v := range dataRequest.Filters {
		list.lastDataRequest.Filters[k] = v
	}

	// Load initial chunk
	chunkStartIndex := (viewportStartIndex / config.ChunkSize) * config.ChunkSize
	err := list.loadChunk(chunkStartIndex)
	if err != nil {
		return nil, err
	}

	// Initialize visible items
	list.updateVisibleItems()

	return list, nil
}

// loadChunk loads a chunk of data from the data provider.
func (l *List[T]) loadChunk(startIndex int) error {
	// Check if the chunk is already loaded
	if _, ok := l.chunks[startIndex]; ok {
		return nil
	}

	// Calculate the number of items to load
	count := l.Config.ChunkSize
	if startIndex+count > l.totalItems {
		count = l.totalItems - startIndex
	}

	// Make sure we don't request invalid ranges
	if startIndex < 0 || count <= 0 || startIndex >= l.totalItems {
		return fmt.Errorf("invalid chunk range: %d-%d (total: %d)",
			startIndex, startIndex+count, l.totalItems)
	}

	// Create a data request with the current filters and sorting
	request := l.dataRequest
	request.Start = startIndex
	request.Count = count

	// Load the items from the data provider
	items, err := l.DataProvider.GetItems(request)
	if err != nil {
		return err
	}

	// Validate return data - ensure we got the expected number of items
	// or at most what we asked for
	if len(items) > count {
		// Truncate if provider returned too many items
		items = items[:count]
	}

	// Create and store the chunk
	l.chunks[startIndex] = &chunk[T]{
		StartIndex: startIndex,
		EndIndex:   startIndex + len(items) - 1,
		Items:      items,
	}

	return nil
}

// unloadChunks unloads chunks that are no longer needed.
func (l *List[T]) unloadChunks() {
	// Calculate the bounds of chunks that should be kept
	viewportChunkIndex := (l.State.ViewportStartIndex / l.Config.ChunkSize) * l.Config.ChunkSize
	keepLowerBound := viewportChunkIndex - l.Config.ChunkSize
	if keepLowerBound < 0 {
		keepLowerBound = 0
	}
	keepUpperBound := viewportChunkIndex + (2 * l.Config.ChunkSize)

	// Unload chunks outside the bounds
	for startIndex := range l.chunks {
		if startIndex < keepLowerBound || startIndex > keepUpperBound {
			delete(l.chunks, startIndex)
		}
	}
}

// updateVisibleItems updates the slice of items currently visible in the viewport.
func (l *List[T]) updateVisibleItems() {
	// If there's no data, return an empty slice
	if l.totalItems == 0 {
		l.visibleItems = []Data[T]{}
		return
	}

	// Calculate how many actual items we can show
	maxVisibleItems := l.Config.Height
	if l.totalItems < maxVisibleItems {
		maxVisibleItems = l.totalItems
	}

	// Ensure viewport doesn't extend beyond dataset
	maxStart := l.totalItems - maxVisibleItems
	if l.State.ViewportStartIndex > maxStart {
		l.State.ViewportStartIndex = maxStart
	}
	if l.State.ViewportStartIndex < 0 {
		l.State.ViewportStartIndex = 0
	}

	// Calculate endpoint of visible area (exclusive)
	viewportEnd := l.State.ViewportStartIndex + maxVisibleItems
	if viewportEnd > l.totalItems {
		viewportEnd = l.totalItems
	}

	// Create a new slice to hold visible items
	l.visibleItems = make([]Data[T], 0, viewportEnd-l.State.ViewportStartIndex)

	// Fill the visible items slice with actual data
	for i := l.State.ViewportStartIndex; i < viewportEnd; i++ {
		// Get the chunk that contains this item
		chunkStartIndex := (i / l.Config.ChunkSize) * l.Config.ChunkSize
		chunk, ok := l.chunks[chunkStartIndex]

		// If chunk isn't loaded, load it
		if !ok {
			// Try to load the chunk
			err := l.loadChunk(chunkStartIndex)
			if err != nil {
				// Skip this item if we can't load its chunk
				continue
			}
			chunk = l.chunks[chunkStartIndex]
		}

		// Skip if chunk is nil (should never happen, but safety first)
		if chunk == nil {
			continue
		}

		// Calculate item index within the chunk
		itemIndex := i - chunk.StartIndex

		// Only add the item if it's within the chunk's bounds
		if itemIndex >= 0 && itemIndex < len(chunk.Items) {
			l.visibleItems = append(l.visibleItems, chunk.Items[itemIndex])
		} else {
			// Try reloading this chunk - it may have changed due to filtering
			delete(l.chunks, chunkStartIndex)
			err := l.loadChunk(chunkStartIndex)
			if err == nil && l.chunks[chunkStartIndex] != nil {
				// Try again with fresh chunk
				chunk = l.chunks[chunkStartIndex]
				if itemIndex >= 0 && itemIndex < len(chunk.Items) {
					l.visibleItems = append(l.visibleItems, chunk.Items[itemIndex])
				}
			}
		}
	}

	// Ensure cursor stays within bounds of visible data
	if l.State.CursorViewportIndex >= len(l.visibleItems) {
		if len(l.visibleItems) > 0 {
			l.State.CursorViewportIndex = len(l.visibleItems) - 1
		} else {
			l.State.CursorViewportIndex = 0
		}
		// Adjust absolute cursor position
		l.State.CursorIndex = l.State.ViewportStartIndex + l.State.CursorViewportIndex
	}
}

// GetVisibleItems returns the slice of items currently visible in the viewport.
func (l *List[T]) GetVisibleItems() []Data[T] {
	return l.visibleItems
}

// GetTotalItems returns the total number of items in the dataset.
func (l *List[T]) GetTotalItems() int {
	return l.totalItems
}

// getTotalItemsFromProvider intelligently fetches total items count
// Only calls DataProvider.GetTotal() when the dataset structure has actually changed
func (l *List[T]) getTotalItemsFromProvider() int {
	// Check if our cached total is still valid by comparing data requests
	if l.totalItemsValid && l.dataRequestsEqual(l.dataRequest, l.lastDataRequest) {
		// Dataset structure hasn't changed, use cached value
		return l.totalItems
	}

	// Dataset structure has changed, need to fetch fresh count
	newTotal := l.DataProvider.GetTotal()
	l.totalItems = newTotal
	l.totalItemsValid = true

	// IMPORTANT: Deep copy the dataRequest to avoid reference sharing
	// The Filters map needs to be copied, not just referenced
	l.lastDataRequest = DataRequest{
		Start:          l.dataRequest.Start,
		Count:          l.dataRequest.Count,
		SortFields:     append([]string(nil), l.dataRequest.SortFields...),
		SortDirections: append([]string(nil), l.dataRequest.SortDirections...),
		Filters:        make(map[string]any),
	}

	// Copy filters map
	for k, v := range l.dataRequest.Filters {
		l.lastDataRequest.Filters[k] = v
	}

	return newTotal
}

// dataRequestsEqual compares two DataRequest objects to see if they would affect total count
func (l *List[T]) dataRequestsEqual(a, b DataRequest) bool {
	// Check if filters are the same
	if len(a.Filters) != len(b.Filters) {
		return false
	}
	for k, v := range a.Filters {
		if bVal, exists := b.Filters[k]; !exists || v != bVal {
			return false
		}
	}

	// Sort fields don't affect total count, so we don't compare them
	// Only filters can change the total number of items

	return true
}

// GetCurrentItem returns the currently selected item.
func (l *List[T]) GetCurrentItem() (Data[T], bool) {
	if l.State.CursorIndex < 0 || l.State.CursorIndex >= l.totalItems {
		var zero Data[T]
		return zero, false
	}

	chunkStartIndex := (l.State.CursorIndex / l.Config.ChunkSize) * l.Config.ChunkSize
	chunk, ok := l.chunks[chunkStartIndex]
	if !ok {
		var zero Data[T]
		return zero, false
	}

	itemIndex := l.State.CursorIndex - chunk.StartIndex
	if itemIndex < 0 || itemIndex >= len(chunk.Items) {
		var zero Data[T]
		return zero, false
	}

	return chunk.Items[itemIndex], true
}

// GetState returns the current viewport state.
func (l *List[T]) GetState() ViewportState {
	return l.State
}

// GetLoadedChunks returns information about the currently loaded chunks.
func (l *List[T]) GetLoadedChunks() []ChunkInfo {
	chunks := make([]ChunkInfo, 0, len(l.chunks))
	for _, chunk := range l.chunks {
		chunks = append(chunks, ChunkInfo{
			StartIndex: chunk.StartIndex,
			EndIndex:   chunk.EndIndex,
			ItemCount:  len(chunk.Items),
		})
	}
	return chunks
}

// Render renders the list to a string using the provided formatter.
func (l *List[T]) Render() string {
	return l.RenderWithAnimatedContent(nil)
}

// RenderWithAnimatedContent renders the list with optional animated content
func (l *List[T]) RenderWithAnimatedContent(animatedContent map[string]string) string {
	var builder strings.Builder

	// Special case for empty dataset
	if l.totalItems == 0 {
		// Return an empty string for empty dataset
		return ""
	}

	// Set up styles
	rowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(l.StyleConfig.RowStyle))
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(l.StyleConfig.SelectedRowStyle))

	// Only render rows that actually have data
	for i, item := range l.visibleItems {
		absoluteIndex := l.State.ViewportStartIndex + i

		// Skip if we've rendered all real data
		if absoluteIndex >= l.totalItems {
			break
		}

		isCursor := i == l.State.CursorViewportIndex
		isTopThreshold := i == l.Config.TopThresholdIndex
		isBottomThreshold := i == l.Config.BottomThresholdIndex

		var renderedItem string

		// Check if we have animated content for this item
		animKey := fmt.Sprintf("item-%d", absoluteIndex)
		if animatedContent != nil && animatedContent[animKey] != "" {
			renderedItem = animatedContent[animKey]
		} else if l.Formatter != nil {
			// Use regular formatter if no animated content
			ctx := RenderContext{
				MaxWidth:     80,  // Reasonable default for lists
				MaxHeight:    1,   // Single line for list items
				ColumnIndex:  0,   // Lists don't have columns
				ColumnConfig: nil, // Lists don't have column config
			}
			renderedItem = l.Formatter(item, absoluteIndex, ctx, isCursor, isTopThreshold, isBottomThreshold)
		} else {
			// Default formatting if no formatter is provided
			itemStr := fmt.Sprintf("%v", item)
			prefix := fmt.Sprintf("%d - ", absoluteIndex)

			if isCursor {
				renderedItem = selectedStyle.Render(prefix + "[ " + itemStr + " ]")
			} else {
				renderedItem = rowStyle.Render(prefix + itemStr)
			}
		}

		builder.WriteString(renderedItem)

		// Add a newline unless it's the last actual item
		if i < len(l.visibleItems)-1 && absoluteIndex < l.totalItems-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// RenderDebugInfo renders debug information about the list.
func (l *List[T]) RenderDebugInfo() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Total Items: %d\n", l.totalItems))
	builder.WriteString(fmt.Sprintf("Viewport Start: %d\n", l.State.ViewportStartIndex))
	builder.WriteString(fmt.Sprintf("Viewport End: %d\n", l.State.ViewportStartIndex+l.Config.Height))
	builder.WriteString(fmt.Sprintf("Cursor Index: %d\n", l.State.CursorIndex))
	builder.WriteString(fmt.Sprintf("Cursor Viewport Index: %d\n", l.State.CursorViewportIndex))
	builder.WriteString(fmt.Sprintf("At Top Threshold: %t\n", l.State.IsAtTopThreshold))
	builder.WriteString(fmt.Sprintf("At Bottom Threshold: %t\n", l.State.IsAtBottomThreshold))
	builder.WriteString(fmt.Sprintf("At Dataset Start: %t\n", l.State.AtDatasetStart))
	builder.WriteString(fmt.Sprintf("At Dataset End: %t\n", l.State.AtDatasetEnd))

	// Show information about loaded chunks
	builder.WriteString("Loaded Chunks:\n")
	for _, chunk := range l.chunks {
		builder.WriteString(fmt.Sprintf("  Chunk %d-%d (%d items)\n", chunk.StartIndex, chunk.EndIndex, len(chunk.Items)))
	}

	// Show sorting and filtering info
	if len(l.dataRequest.SortFields) > 0 {
		builder.WriteString("Sort:\n")
		for i, field := range l.dataRequest.SortFields {
			builder.WriteString(fmt.Sprintf("  %s (%s)\n", field, l.dataRequest.SortDirections[i]))
		}
	}

	if len(l.dataRequest.Filters) > 0 {
		builder.WriteString("Filters:\n")
		for key, value := range l.dataRequest.Filters {
			builder.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	return builder.String()
}

// SetFilter sets a filter for a specific field.
// Applying a filter will reload all data.
func (l *List[T]) SetFilter(field string, value any) {
	// Initialize filters map if it doesn't exist
	if l.dataRequest.Filters == nil {
		l.dataRequest.Filters = make(map[string]any)
	}

	// Set the filter
	l.dataRequest.Filters[field] = value

	// Reload all data
	l.refreshData()
}

// ClearFilters removes all filters.
// This will reload all data.
func (l *List[T]) ClearFilters() {
	// Only reload if we had filters before
	if len(l.dataRequest.Filters) > 0 {
		l.dataRequest.Filters = make(map[string]any)
		l.refreshData()
	}
}

// RemoveFilter removes a specific filter.
// This will reload all data.
func (l *List[T]) RemoveFilter(field string) {
	// Only reload if the filter existed
	if _, exists := l.dataRequest.Filters[field]; exists {
		delete(l.dataRequest.Filters, field)
		l.refreshData()
	}
}

// SetSort sets a sort field and direction.
// Direction should be "asc" or "desc".
// Applying a sort will reload all data.
func (l *List[T]) SetSort(field string, direction string) {
	// Validate direction
	if direction != "asc" && direction != "desc" {
		direction = "asc" // Default to ascending if invalid
	}

	// Check if we're already sorting by this field
	for i, existingField := range l.dataRequest.SortFields {
		if existingField == field {
			// Update the direction
			l.dataRequest.SortDirections[i] = direction

			// Move to front if not already at front
			if i > 0 {
				// Copy out this field/direction
				tempField := l.dataRequest.SortFields[i]
				tempDir := l.dataRequest.SortDirections[i]

				// Remove from current position
				l.dataRequest.SortFields = append(l.dataRequest.SortFields[:i], l.dataRequest.SortFields[i+1:]...)
				l.dataRequest.SortDirections = append(l.dataRequest.SortDirections[:i], l.dataRequest.SortDirections[i+1:]...)

				// Add to front as primary sort
				l.dataRequest.SortFields = append([]string{tempField}, l.dataRequest.SortFields...)
				l.dataRequest.SortDirections = append([]string{tempDir}, l.dataRequest.SortDirections...)
			}

			// Reload data with updated sorts
			l.refreshData()
			return
		}
	}

	// Clear existing sorts and add the new one - THIS IS A COMPLETELY NEW SORT
	l.dataRequest.SortFields = []string{field}
	l.dataRequest.SortDirections = []string{direction}

	// Reload all data
	l.refreshData()
}

// AddSort adds a sort field and direction without clearing existing sorts.
// This allows for multi-column sorting.
// Direction should be "asc" or "desc".
// Applying a sort will reload all data.
func (l *List[T]) AddSort(field string, direction string) {
	// Check if we're already sorting by this field
	for i, existingField := range l.dataRequest.SortFields {
		if existingField == field {
			// Update the direction and move it to the front
			l.dataRequest.SortDirections[i] = direction

			// Move this field to the front (most important)
			if i > 0 {
				// Make a copy of the field and direction
				tempField := l.dataRequest.SortFields[i]
				tempDir := l.dataRequest.SortDirections[i]

				// Shift all fields before it
				for j := i; j > 0; j-- {
					l.dataRequest.SortFields[j] = l.dataRequest.SortFields[j-1]
					l.dataRequest.SortDirections[j] = l.dataRequest.SortDirections[j-1]
				}

				// Place the field at the front
				l.dataRequest.SortFields[0] = tempField
				l.dataRequest.SortDirections[0] = tempDir
			}

			// Reload all data
			l.refreshData()
			return
		}
	}

	// Validate direction
	if direction != "asc" && direction != "desc" {
		direction = "asc" // Default to ascending if invalid
	}

	// Add new sort field at the beginning (most important)
	l.dataRequest.SortFields = append([]string{field}, l.dataRequest.SortFields...)
	l.dataRequest.SortDirections = append([]string{direction}, l.dataRequest.SortDirections...)

	// Reload all data
	l.refreshData()
}

// RemoveSort removes a sort field.
// This will reload all data if the field was being sorted on.
func (l *List[T]) RemoveSort(field string) {
	// Check if we're sorting by this field
	for i, existingField := range l.dataRequest.SortFields {
		if existingField == field {
			// Remove this field from sorting
			newSortFields := make([]string, 0, len(l.dataRequest.SortFields)-1)
			newSortDirections := make([]string, 0, len(l.dataRequest.SortDirections)-1)

			for j, f := range l.dataRequest.SortFields {
				if j != i {
					newSortFields = append(newSortFields, f)
					newSortDirections = append(newSortDirections, l.dataRequest.SortDirections[j])
				}
			}

			l.dataRequest.SortFields = newSortFields
			l.dataRequest.SortDirections = newSortDirections

			// Reload all data
			l.refreshData()
			return
		}
	}
}

// ClearSort removes any sorting criteria.
// This will reload all data.
func (l *List[T]) ClearSort() {
	// Only reload if we had a sort before
	if len(l.dataRequest.SortFields) > 0 {
		l.dataRequest.SortFields = []string{}
		l.dataRequest.SortDirections = []string{}
		l.refreshData()
	}
}

// GetDataRequest returns the current data request configuration.
func (l *List[T]) GetDataRequest() DataRequest {
	return l.dataRequest
}

// refreshData reloads the data with current filter and sort settings.
// This should ONLY be called for structural changes (filters, sorts)
// NOT for selection changes or navigation within loaded chunks
func (l *List[T]) refreshData() {
	// Store current position - we'll try to keep this position in view
	currentPos := l.State.CursorIndex

	// Mark all chunks as dirty but DON'T reload them immediately
	// They will be loaded lazily when actually needed for display
	for chunkStart := range l.chunks {
		delete(l.chunks, chunkStart)
	}

	// SMART PERFORMANCE FIX: Only call GetTotal() when dataset structure actually changed
	// This prevents unnecessary data provider calls on every animation update
	newTotalItems := l.getTotalItemsFromProvider()

	// Handle the case where the dataset is now empty after filtering
	if newTotalItems == 0 {
		// Reset everything to default state for empty dataset
		l.State.ViewportStartIndex = 0
		l.State.CursorIndex = 0
		l.State.CursorViewportIndex = 0
		l.State.IsAtTopThreshold = false
		l.State.IsAtBottomThreshold = false
		l.State.AtDatasetStart = true
		l.State.AtDatasetEnd = true
		l.visibleItems = []Data[T]{}
		return
	}

	// If we have filtered data, reset cursor to beginning if current position is invalid
	if currentPos >= newTotalItems {
		currentPos = newTotalItems - 1
	}
	if currentPos < 0 {
		currentPos = 0
	}

	// Calculate optimal viewport start - try to keep cursor in the middle
	var viewportStartIndex int
	if l.totalItems <= l.Config.Height {
		// For small datasets, always start at first record
		viewportStartIndex = 0
	} else {
		// For large datasets, try to position cursor in the middle
		viewportStartIndex = currentPos - (l.Config.Height / 2)

		// Don't go below 0
		if viewportStartIndex < 0 {
			viewportStartIndex = 0
		}

		// Don't show beyond dataset end
		maxStart := l.totalItems - l.Config.Height
		if viewportStartIndex > maxStart {
			viewportStartIndex = maxStart
		}
	}

	// Update cursor positions
	l.State.ViewportStartIndex = viewportStartIndex
	l.State.CursorIndex = currentPos
	l.State.CursorViewportIndex = currentPos - viewportStartIndex

	// Fix cursor viewport index if it's out of bounds
	if l.State.CursorViewportIndex < 0 {
		l.State.CursorViewportIndex = 0
		l.State.CursorIndex = viewportStartIndex // Adjust absolute cursor position
	}
	if l.State.CursorViewportIndex >= l.Config.Height {
		l.State.CursorViewportIndex = l.Config.Height - 1
		l.State.CursorIndex = viewportStartIndex + l.State.CursorViewportIndex
	}

	// Make sure cursor doesn't go beyond dataset
	if l.State.CursorIndex >= l.totalItems {
		l.State.CursorIndex = l.totalItems - 1
		l.State.CursorViewportIndex = l.State.CursorIndex - viewportStartIndex
	}

	// Update threshold and boundary flags
	l.State.IsAtTopThreshold = l.State.CursorViewportIndex == l.Config.TopThresholdIndex
	l.State.IsAtBottomThreshold = l.State.CursorViewportIndex == l.Config.BottomThresholdIndex
	l.State.AtDatasetStart = viewportStartIndex == 0
	l.State.AtDatasetEnd = viewportStartIndex+l.Config.Height >= l.totalItems

	// DON'T force load chunks here - let updateVisibleItems() load them lazily
	// This is the key performance improvement: only load when actually needed for display

	// Update the visible items (this will trigger lazy chunk loading)
	l.updateVisibleItems()
}

// moveToIndex is a helper method for positioning the viewport and cursor at a specific index
func (l *List[T]) moveToIndex(index int) {
	// Handle empty dataset case
	if l.totalItems <= 0 {
		// Reset everything to zero position
		l.State.ViewportStartIndex = 0
		l.State.CursorIndex = 0
		l.State.CursorViewportIndex = 0
		l.State.IsAtTopThreshold = false
		l.State.IsAtBottomThreshold = false
		l.State.AtDatasetStart = true
		l.State.AtDatasetEnd = true

		// No data to display
		l.visibleItems = nil
		return
	}

	// Ensure index is within bounds
	if index < 0 {
		index = 0
	}
	if index >= l.totalItems {
		index = l.totalItems - 1
	}

	// Store current position
	newCursorIndex := index

	// Calculate new viewport start position
	var newViewportStartIndex int
	if l.totalItems <= l.Config.Height {
		// For small datasets, always show everything from the beginning
		newViewportStartIndex = 0
	} else if newCursorIndex < l.Config.TopThresholdIndex {
		// Near the beginning of the dataset
		newViewportStartIndex = 0
	} else if newCursorIndex >= l.totalItems-l.Config.BottomThresholdIndex {
		// Near the end of the dataset
		newViewportStartIndex = l.totalItems - l.Config.Height
		if newViewportStartIndex < 0 {
			newViewportStartIndex = 0
		}
	} else {
		// Middle of the dataset - position the cursor at the top threshold index
		newViewportStartIndex = newCursorIndex - l.Config.TopThresholdIndex
	}

	// Double-check that viewport start index is in bounds
	if newViewportStartIndex < 0 {
		newViewportStartIndex = 0
	}

	// Make sure viewportStartIndex doesn't go beyond the possible range
	maxStart := l.totalItems - 1
	if l.totalItems > l.Config.Height {
		maxStart = l.totalItems - l.Config.Height
	}
	if newViewportStartIndex > maxStart {
		newViewportStartIndex = maxStart
	}

	// Calculate cursor viewport index (relative position in visible area)
	newCursorViewportIndex := newCursorIndex - newViewportStartIndex

	// Make sure cursorViewportIndex doesn't exceed the viewport height
	if newCursorViewportIndex >= l.Config.Height {
		newCursorViewportIndex = l.Config.Height - 1
		// Also adjust absolute cursor position to match
		newCursorIndex = newViewportStartIndex + newCursorViewportIndex
	}

	// Additional check: make sure we don't exceed total number of items
	if newCursorIndex >= l.totalItems {
		newCursorIndex = l.totalItems - 1
		newCursorViewportIndex = newCursorIndex - newViewportStartIndex
	}

	// Make sure cursorViewportIndex doesn't go negative
	if newCursorViewportIndex < 0 {
		newCursorViewportIndex = 0
		// Also adjust absolute cursor position to match
		newCursorIndex = newViewportStartIndex
	}

	// Update state
	l.State.ViewportStartIndex = newViewportStartIndex
	l.State.CursorIndex = newCursorIndex
	l.State.CursorViewportIndex = newCursorViewportIndex

	// Update threshold flags
	l.State.IsAtTopThreshold = newCursorViewportIndex == l.Config.TopThresholdIndex
	l.State.IsAtBottomThreshold = newCursorViewportIndex == l.Config.BottomThresholdIndex

	// Update dataset boundary flags - critical for proper navigation
	l.State.AtDatasetStart = newViewportStartIndex == 0
	l.State.AtDatasetEnd = newViewportStartIndex+l.Config.Height >= l.totalItems

	// Load chunks if they're not loaded
	if l.totalItems > 0 {
		chunkStartIndex := (newViewportStartIndex / l.Config.ChunkSize) * l.Config.ChunkSize
		_ = l.loadChunk(chunkStartIndex)

		// Load an additional chunk if needed
		nextChunkIndex := chunkStartIndex + l.Config.ChunkSize
		if nextChunkIndex < l.totalItems && newViewportStartIndex+l.Config.Height > nextChunkIndex {
			_ = l.loadChunk(nextChunkIndex)
		}
	}

	// Update visible items
	l.updateVisibleItems()

	// Clean up unused chunks
	l.unloadChunks()
}

// JumpToIndex jumps to the specified index in the dataset.
func (l *List[T]) JumpToIndex(index int) {
	l.moveToIndex(index)
}

// InvalidateTotalItemsCache marks the cached total items count as invalid
// This should be called when the dataset is known to have changed externally
func (l *List[T]) InvalidateTotalItemsCache() {
	l.totalItemsValid = false
}
