package vtable

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// FilterMsg is a message type to apply or update filters.
type FilterMsg struct {
	Field  string
	Value  any
	Clear  bool
	Remove bool
}

// SortMsg is a message type to apply or update sorting.
type SortMsg struct {
	Field     string
	Direction string
	Clear     bool
	Remove    bool
	Add       bool
}

// TeaList is a Bubble Tea model wrapping a List.
type TeaList[T any] struct {
	// The underlying list model
	list *List[T]

	// Key mappings
	keyMap NavigationKeyMap

	// Whether the component is focused
	focused bool

	// Event callbacks
	onSelectItem     func(item T, index int)
	onHighlight      func(item T, index int)
	onScroll         func(state ViewportState)
	onFiltersChanged func(filters map[string]any)
	onSortChanged    func(field, direction string)

	// Animation support
	animationEngine   *AnimationEngine
	animatedFormatter ItemFormatterAnimated[T]
	animationConfig   AnimationConfig
	lastAnimationTime time.Time

	// Real-time data updates
	realTimeUpdates    bool
	realTimeInterval   time.Duration
	lastRealTimeUpdate time.Time

	// Animation content cache
	cachedAnimationContent map[string]string

	// Track cursor position for cache invalidation
	lastCursorIndex int
}

// NewTeaList creates a new Bubble Tea model for a virtualized list.
func NewTeaList[T any](
	config ViewportConfig,
	provider DataProvider[T],
	styleConfig StyleConfig,
	formatter ItemFormatter[T],
) (*TeaList[T], error) {
	// Create the underlying list
	list, err := NewList(config, provider, styleConfig, formatter)
	if err != nil {
		return nil, err
	}

	// Initialize animation system
	animConfig := DefaultAnimationConfig()
	animEngine := NewAnimationEngine(animConfig)

	return &TeaList[T]{
		list:                   list,
		keyMap:                 PlatformKeyMap(), // Use platform-specific key bindings
		focused:                true,
		animationEngine:        animEngine,
		animationConfig:        animConfig,
		cachedAnimationContent: make(map[string]string),
		lastCursorIndex:        0, // Initialize cursor tracking
	}, nil
}

// Init initializes the Tea model.
func (m *TeaList[T]) Init() tea.Cmd {
	// Start the global animation loop
	return StartGlobalAnimationLoop()
}

// Update updates the Tea model based on messages.
func (m *TeaList[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If not focused, don't handle messages
	if !m.focused {
		return m, nil
	}

	var cmds []tea.Cmd
	previousState := m.list.GetState()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Up):
			m.list.MoveUp()
		case key.Matches(msg, m.keyMap.Down):
			m.list.MoveDown()
		case key.Matches(msg, m.keyMap.PageUp):
			m.list.PageUp()
		case key.Matches(msg, m.keyMap.PageDown):
			m.list.PageDown()
		case key.Matches(msg, m.keyMap.Home):
			m.list.JumpToStart()
		case key.Matches(msg, m.keyMap.End):
			m.list.JumpToEnd()
		case key.Matches(msg, m.keyMap.Select):
			if m.onSelectItem != nil {
				if item, ok := m.GetCurrentItem(); ok {
					m.onSelectItem(item, m.GetState().CursorIndex)
				}
			}
		}
	case GlobalAnimationTickMsg:
		// Handle global animation tick - this runs continuously while animations are active
		if cmd := m.animationEngine.ProcessGlobalTick(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}

		// Handle real-time data updates if enabled
		if m.realTimeUpdates {
			now := msg.Timestamp
			if now.Sub(m.lastRealTimeUpdate) >= m.realTimeInterval {
				// Time for a real-time data refresh
				m.lastRealTimeUpdate = now
				m.ForceDataRefresh()
			}
		}
	case AnimationUpdateMsg:
		// Animations have been updated - trigger re-render by doing nothing
		// The View() method will automatically pick up the changes

		// CRITICAL FIX: Only update animation content when we receive animation update messages
		// This decouples animation updates from cursor movements
		if m.animatedFormatter != nil {
			m.updateAnimationContent()
		}
	case FilterMsg:
		previousFilters := make(map[string]any)
		for k, v := range m.list.dataRequest.Filters {
			previousFilters[k] = v
		}

		// Handle the filter message
		if msg.Clear {
			m.list.ClearFilters()
		} else if msg.Remove {
			m.list.RemoveFilter(msg.Field)
		} else {
			m.list.SetFilter(msg.Field, msg.Value)
		}

		// After filter changes, ensure the visual state is properly updated
		// If the number of items changes dramatically, we may need to adjust cursor position
		if m.list.totalItems == 0 {
			// No matching items after filter
			cmds = append(cmds, func() tea.Msg {
				return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("home")}
			})
		} else if m.list.totalItems <= m.list.Config.Height {
			// Small enough dataset to show everything, jump to start
			cmds = append(cmds, func() tea.Msg {
				return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("home")}
			})
		}

		// Call the filter changed callback if filters changed
		if m.onFiltersChanged != nil {
			hasChanged := len(previousFilters) != len(m.list.dataRequest.Filters)
			if !hasChanged {
				// Check if any values changed
				for k, v := range previousFilters {
					if newV, ok := m.list.dataRequest.Filters[k]; !ok || newV != v {
						hasChanged = true
						break
					}
				}
			}

			if hasChanged {
				m.onFiltersChanged(m.list.dataRequest.Filters)
			}
		}
	case SortMsg:
		// Store previous sorts for callback comparison
		previousSortFields := make([]string, len(m.list.dataRequest.SortFields))
		previousSortDirections := make([]string, len(m.list.dataRequest.SortDirections))
		copy(previousSortFields, m.list.dataRequest.SortFields)
		copy(previousSortDirections, m.list.dataRequest.SortDirections)

		// Handle the sort message
		if msg.Clear {
			m.list.ClearSort()
		} else if msg.Remove {
			m.list.RemoveSort(msg.Field)
		} else if msg.Add {
			m.list.AddSort(msg.Field, msg.Direction)
		} else {
			m.list.SetSort(msg.Field, msg.Direction)
		}

		// Call the sort changed callback if sort changed
		if m.onSortChanged != nil {
			// Check if sorting has changed
			changed := len(previousSortFields) != len(m.list.dataRequest.SortFields)
			if !changed {
				for i, field := range previousSortFields {
					if i >= len(m.list.dataRequest.SortFields) ||
						field != m.list.dataRequest.SortFields[i] ||
						previousSortDirections[i] != m.list.dataRequest.SortDirections[i] {
						changed = true
						break
					}
				}
			}

			if changed && len(m.list.dataRequest.SortFields) > 0 {
				m.onSortChanged(
					strings.Join(m.list.dataRequest.SortFields, ","),
					strings.Join(m.list.dataRequest.SortDirections, ","),
				)
			} else if changed {
				m.onSortChanged("", "")
			}
		}
	}

	// Check if we need to trigger callbacks based on state changes
	currentState := m.list.GetState()

	// Call onScroll if viewport changed
	if m.onScroll != nil && (previousState.ViewportStartIndex != currentState.ViewportStartIndex) {
		m.onScroll(currentState)
	}

	// Call onHighlight if highlighted item changed
	if m.onHighlight != nil && previousState.CursorIndex != currentState.CursorIndex {
		if item, ok := m.GetCurrentItem(); ok {
			m.onHighlight(item, currentState.CursorIndex)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the Tea model.
func (m *TeaList[T]) View() string {
	// Handle animations if we have an animated formatter
	if m.animatedFormatter != nil {
		// Ensure we have fresh data before rendering
		m.list.updateVisibleItems()

		// Check if cursor position has changed
		currentCursorIndex := m.list.GetState().CursorIndex
		if currentCursorIndex != m.lastCursorIndex {
			// Cursor moved - update cache immediately for smooth movement
			m.updateCursorInCache(m.lastCursorIndex, currentCursorIndex)
			m.lastCursorIndex = currentCursorIndex
		}

		// Ensure cache is populated for all visible items
		m.ensureAnimationCachePopulated()

		// Use cached animation content - only updated when animation update messages are received
		result := m.list.RenderWithAnimatedContent(m.cachedAnimationContent)
		return result
	}

	return m.list.Render()
}

// ensureAnimationCachePopulated ensures all visible items have cached animation content
func (m *TeaList[T]) ensureAnimationCachePopulated() {
	visibleItems := m.list.GetVisibleItems()
	state := m.list.GetState()

	for i := range visibleItems {
		absoluteIndex := state.ViewportStartIndex + i
		animationKey := fmt.Sprintf("item-%d", absoluteIndex)

		// If we don't have cached content for this item, it means it's newly visible
		// We need to populate the cache without triggering animation acceleration
		if _, exists := m.cachedAnimationContent[animationKey]; !exists {
			// This is a new visible item - we need to populate its cache
			// But we should only call the formatter once for initial setup
			m.populateInitialAnimationContent(i, absoluteIndex, animationKey)
		}
	}
}

// populateInitialAnimationContent populates cache for a newly visible item
func (m *TeaList[T]) populateInitialAnimationContent(viewportIndex, absoluteIndex int, animationKey string) {
	visibleItems := m.list.GetVisibleItems()
	if viewportIndex >= len(visibleItems) {
		return
	}

	dataItem := visibleItems[viewportIndex]
	state := m.list.GetState()

	// Create render context
	ctx := DefaultRenderContext()
	ctx.CurrentTime = time.Now()
	ctx.DeltaTime = 0 // No delta time for initial render
	ctx.MaxWidth = 80

	// Get animation state (will be empty for new animations)
	animState := m.animationEngine.GetAnimationState(animationKey)

	// Determine cursor state
	isCursor := viewportIndex == state.CursorViewportIndex
	isTopThreshold := viewportIndex == m.list.Config.TopThresholdIndex
	isBottomThreshold := viewportIndex == m.list.Config.BottomThresholdIndex

	// Call animated formatter ONCE for initial setup
	result := m.animatedFormatter(dataItem, absoluteIndex, ctx, animState, isCursor, isTopThreshold, isBottomThreshold)

	// Cache the content
	m.cachedAnimationContent[animationKey] = result.Content

	// Register animation if needed (this should only happen once)
	if len(result.RefreshTriggers) > 0 && !m.animationEngine.IsVisible(animationKey) {
		if cmd := m.animationEngine.RegisterAnimation(animationKey, result.RefreshTriggers, result.AnimationState); cmd != nil {
			_ = cmd
		}
	}

	// Set initial animation state
	if len(result.AnimationState) > 0 {
		m.animationEngine.UpdateAnimationState(animationKey, result.AnimationState)
	}

	// Make animation visible
	m.animationEngine.SetVisible(animationKey, true)
}

// processAnimations handles the animation lifecycle for visible items and returns rendered content
func (m *TeaList[T]) processAnimations() map[string]string {
	visibleItems := m.list.GetVisibleItems()
	state := m.list.GetState()

	// Track which animations should be active
	activeAnimationKeys := make(map[string]bool)

	// Calculate delta time
	now := time.Now()
	deltaTime := time.Duration(0)
	if !m.lastAnimationTime.IsZero() {
		deltaTime = now.Sub(m.lastAnimationTime)
	}
	m.lastAnimationTime = now

	// Process each visible item
	for i, dataItem := range visibleItems {
		absoluteIndex := state.ViewportStartIndex + i
		animationKey := fmt.Sprintf("item-%d", absoluteIndex)
		activeAnimationKeys[animationKey] = true

		// Create render context with delta time
		ctx := DefaultRenderContext()
		ctx.CurrentTime = now
		ctx.DeltaTime = deltaTime
		ctx.MaxWidth = 80 // Default width - could be made configurable

		// Get animation state
		animState := m.animationEngine.GetAnimationState(animationKey)

		// Determine cursor state
		isCursor := i == state.CursorViewportIndex
		isTopThreshold := i == m.list.Config.TopThresholdIndex
		isBottomThreshold := i == m.list.Config.BottomThresholdIndex

		// Call animated formatter
		result := m.animatedFormatter(dataItem, absoluteIndex, ctx, animState, isCursor, isTopThreshold, isBottomThreshold)

		// Cache the rendered content
		m.cachedAnimationContent[animationKey] = result.Content

		// CRITICAL FIX: Only register animations ONCE when they first become visible
		// Do NOT re-register on every view render (this causes acceleration)
		if len(result.RefreshTriggers) > 0 && !m.animationEngine.IsVisible(animationKey) {
			// Register animation only if it doesn't exist yet
			if cmd := m.animationEngine.RegisterAnimation(animationKey, result.RefreshTriggers, result.AnimationState); cmd != nil {
				// Animation loop started - this only happens once
				_ = cmd
			}
		}

		// Update animation state ONLY if it actually changed
		// Don't update state on every render to prevent animation reset
		if len(result.AnimationState) > 0 {
			currentState := m.animationEngine.GetAnimationState(animationKey)
			hasChanges := false

			// Check if state actually changed
			for k, newValue := range result.AnimationState {
				if currentValue, exists := currentState[k]; !exists || currentValue != newValue {
					hasChanges = true
					break
				}
			}

			// Only update if there are actual changes
			if hasChanges {
				m.animationEngine.UpdateAnimationState(animationKey, result.AnimationState)
			}
		}

		// Make sure the animation is visible (this is safe to call repeatedly)
		m.animationEngine.SetVisible(animationKey, true)
	}

	// Clean up animations for items that are no longer visible
	activeAnimations := m.animationEngine.GetActiveAnimations()
	for _, animKey := range activeAnimations {
		if !activeAnimationKeys[animKey] {
			m.animationEngine.SetVisible(animKey, false)
			// Remove cached content for invisible items
			delete(m.cachedAnimationContent, animKey)
		}
	}

	// Clear dirty flags after processing
	m.animationEngine.ClearDirtyFlags()

	return m.cachedAnimationContent
}

// Focus sets the focus state of the component.
func (m *TeaList[T]) Focus() {
	m.focused = true
}

// Blur removes focus from the component.
func (m *TeaList[T]) Blur() {
	m.focused = false
}

// IsFocused returns whether the component is focused.
func (m *TeaList[T]) IsFocused() bool {
	return m.focused
}

// GetState returns the current viewport state.
func (m *TeaList[T]) GetState() ViewportState {
	return m.list.GetState()
}

// GetVisibleItems returns the slice of items currently visible in the viewport.
func (m *TeaList[T]) GetVisibleItems() []T {
	dataItems := m.list.GetVisibleItems()
	items := make([]T, len(dataItems))
	for i, data := range dataItems {
		items[i] = data.Item
	}
	return items
}

// GetCurrentItem returns the currently selected item.
func (m *TeaList[T]) GetCurrentItem() (T, bool) {
	data, ok := m.list.GetCurrentItem()
	if !ok {
		var zero T
		return zero, false
	}
	return data.Item, true
}

// SetKeyMap sets the key mappings for the component.
func (m *TeaList[T]) SetKeyMap(keyMap NavigationKeyMap) {
	m.keyMap = keyMap
}

// GetKeyMap returns the current key mappings for the component.
func (m *TeaList[T]) GetKeyMap() NavigationKeyMap {
	return m.keyMap
}

// JumpToItem jumps to an item with the specified key-value pair.
// Returns true if the item was found and jumped to, false otherwise.
func (m *TeaList[T]) JumpToItem(key string, value any) bool {
	return m.list.JumpToItem(key, value)
}

// JumpToIndex jumps to the specified index in the dataset.
func (m *TeaList[T]) JumpToIndex(index int) {
	m.list.JumpToIndex(index)
}

// MoveUp moves the cursor up one position.
func (m *TeaList[T]) MoveUp() {
	m.list.MoveUp()
}

// MoveDown moves the cursor down one position.
func (m *TeaList[T]) MoveDown() {
	m.list.MoveDown()
}

// PageUp moves the cursor up by a page.
func (m *TeaList[T]) PageUp() {
	m.list.PageUp()
}

// PageDown moves the cursor down by a page.
func (m *TeaList[T]) PageDown() {
	m.list.PageDown()
}

// JumpToStart jumps to the start of the dataset.
func (m *TeaList[T]) JumpToStart() {
	m.list.JumpToStart()
}

// JumpToEnd jumps to the end of the dataset.
func (m *TeaList[T]) JumpToEnd() {
	m.list.JumpToEnd()
}

// GetHelpView returns a string describing the key bindings.
func (m *TeaList[T]) GetHelpView() string {
	return GetKeyMapDescription(m.keyMap)
}

// SetStyle updates the style configuration without recreating the list.
// This is much better than creating a new list when only the theme changes.
func (m *TeaList[T]) SetStyle(styleConfig StyleConfig) {
	m.list.StyleConfig = styleConfig
}

// SetFormatter updates the item formatter function.
func (m *TeaList[T]) SetFormatter(formatter ItemFormatter[T]) {
	m.list.Formatter = formatter
}

// SetAnimatedFormatter sets an animated formatter that supports dynamic content.
func (m *TeaList[T]) SetAnimatedFormatter(formatter ItemFormatterAnimated[T]) {
	m.animatedFormatter = formatter

	// Clear existing cache and trigger initial animation setup
	m.cachedAnimationContent = make(map[string]string)

	// Trigger initial animation setup if formatter is set
	if formatter != nil {
		m.processAnimations()
	}
}

// ClearAnimatedFormatter removes the animated formatter and stops all animations.
func (m *TeaList[T]) ClearAnimatedFormatter() {
	m.animatedFormatter = nil
	m.animationEngine.Cleanup()
	// Explicitly stop the loop since we're no longer using animations
	m.animationEngine.StopLoop()
}

// SetAnimationConfig updates the animation configuration.
func (m *TeaList[T]) SetAnimationConfig(config AnimationConfig) tea.Cmd {
	m.animationConfig = config
	return m.animationEngine.UpdateConfig(config)
}

// EnableAnimations enables the animation system and starts the loop if needed.
func (m *TeaList[T]) EnableAnimations() tea.Cmd {
	m.animationConfig.Enabled = true
	return m.animationEngine.UpdateConfig(m.animationConfig)
}

// DisableAnimations disables the animation system and stops the loop.
func (m *TeaList[T]) DisableAnimations() {
	m.animationConfig.Enabled = false
	m.animationEngine.UpdateConfig(m.animationConfig)
}

// IsAnimationEnabled returns whether animations are currently enabled.
func (m *TeaList[T]) IsAnimationEnabled() bool {
	return m.animationConfig.Enabled
}

// IsAnimationLoopRunning returns whether the animation loop is currently running.
func (m *TeaList[T]) IsAnimationLoopRunning() bool {
	return m.animationEngine.IsRunning()
}

// GetAnimationConfig returns the current animation configuration.
func (m *TeaList[T]) GetAnimationConfig() AnimationConfig {
	return m.animationConfig
}

// SetTickInterval sets the animation tick interval for smoother or more efficient animations.
func (m *TeaList[T]) SetTickInterval(interval time.Duration) tea.Cmd {
	m.animationConfig.TickInterval = interval
	return m.animationEngine.UpdateConfig(m.animationConfig)
}

// GetTickInterval returns the current animation tick interval.
func (m *TeaList[T]) GetTickInterval() time.Duration {
	return m.animationConfig.TickInterval
}

// SetDataProvider updates the data provider.
// Note: This will reset the viewport position to the beginning.
func (m *TeaList[T]) SetDataProvider(provider DataProvider[T]) {
	// Store current position
	currentPos := m.list.State.CursorIndex

	// Update provider
	m.list.DataProvider = provider
	m.list.totalItems = provider.GetTotal()

	// Clear chunks and reload data
	m.list.chunks = make(map[int]*chunk[T])

	// Try to restore position or adjust if needed
	if currentPos >= m.list.totalItems {
		currentPos = m.list.totalItems - 1
	}
	if currentPos < 0 {
		currentPos = 0
	}

	m.JumpToIndex(currentPos)
}

// RefreshData forces a reload of data from the provider.
// Useful when the underlying data has changed.
func (m *TeaList[T]) RefreshData() {
	// Store current position
	currentPos := m.list.State.CursorIndex

	// Update total items count
	m.list.totalItems = m.list.DataProvider.GetTotal()

	// Clear chunks and reload data
	m.list.chunks = make(map[int]*chunk[T])

	// Restore position or adjust if needed
	if currentPos >= m.list.totalItems {
		currentPos = m.list.totalItems - 1
	}
	if currentPos < 0 {
		currentPos = 0
	}

	m.JumpToIndex(currentPos)
}

// OnSelect sets a callback function that will be called when an item is selected.
func (m *TeaList[T]) OnSelect(callback func(item T, index int)) {
	m.onSelectItem = callback
}

// OnHighlight sets a callback function that will be called when the highlighted item changes.
func (m *TeaList[T]) OnHighlight(callback func(item T, index int)) {
	m.onHighlight = callback
}

// OnScroll sets a callback function that will be called when the viewport scrolls.
func (m *TeaList[T]) OnScroll(callback func(state ViewportState)) {
	m.onScroll = callback
}

// OnFiltersChanged sets a callback function that will be called when filters change.
func (m *TeaList[T]) OnFiltersChanged(callback func(filters map[string]any)) {
	m.onFiltersChanged = callback
}

// OnSortChanged sets a callback function that will be called when sorting changes.
func (m *TeaList[T]) OnSortChanged(callback func(field, direction string)) {
	m.onSortChanged = callback
}

// HandleKeypress programmatically simulates pressing a key.
func (m *TeaList[T]) HandleKeypress(keyStr string) {
	// Create a key message and update
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyStr)}
	m.Update(keyMsg)
}

// SetFilter sets a filter for a specific field.
func (m *TeaList[T]) SetFilter(field string, value any) {
	m.list.SetFilter(field, value)
}

// RemoveFilter removes a filter for a specific field.
func (m *TeaList[T]) RemoveFilter(field string) {
	m.list.RemoveFilter(field)
}

// ClearFilters removes all filters.
func (m *TeaList[T]) ClearFilters() {
	m.list.ClearFilters()
}

// SetSort sets the sort field and direction, clearing any existing sorts.
func (m *TeaList[T]) SetSort(field, direction string) {
	m.list.SetSort(field, direction)
}

// AddSort adds a sort field and direction without clearing existing sorts.
// This allows for multi-column sorting.
func (m *TeaList[T]) AddSort(field, direction string) {
	m.list.AddSort(field, direction)
}

// RemoveSort removes a specific sort field.
func (m *TeaList[T]) RemoveSort(field string) {
	m.list.RemoveSort(field)
}

// ClearSort removes sorting.
func (m *TeaList[T]) ClearSort() {
	m.list.ClearSort()
}

// GetDataRequest returns the current data request configuration.
func (m *TeaList[T]) GetDataRequest() DataRequest {
	return m.list.GetDataRequest()
}

// FilterCommand returns a command that will trigger filter application.
func FilterCommand(field string, value any) tea.Cmd {
	return func() tea.Msg {
		return FilterMsg{
			Field: field,
			Value: value,
		}
	}
}

// RemoveFilterCommand returns a command to remove a filter.
func RemoveFilterCommand(field string) tea.Cmd {
	return func() tea.Msg {
		return FilterMsg{
			Field:  field,
			Remove: true,
		}
	}
}

// ClearFiltersCommand returns a command to clear all filters.
func ClearFiltersCommand() tea.Cmd {
	return func() tea.Msg {
		return FilterMsg{
			Clear: true,
		}
	}
}

// SortCommand returns a command that will trigger sorting and replace any existing sorts.
func SortCommand(field, direction string) tea.Cmd {
	return func() tea.Msg {
		return SortMsg{
			Field:     field,
			Direction: direction,
		}
	}
}

// AddSortCommand returns a command that will add a sort field without clearing existing sorts.
func AddSortCommand(field, direction string) tea.Cmd {
	return func() tea.Msg {
		return SortMsg{
			Field:     field,
			Direction: direction,
			Add:       true,
		}
	}
}

// RemoveSortCommand returns a command to remove a specific sort field.
func RemoveSortCommand(field string) tea.Cmd {
	return func() tea.Msg {
		return SortMsg{
			Field:  field,
			Remove: true,
		}
	}
}

// ClearSortCommand returns a command to clear sorting.
func ClearSortCommand() tea.Cmd {
	return func() tea.Msg {
		return SortMsg{
			Clear: true,
		}
	}
}

// Selection methods - delegate to the underlying DataProvider

// ToggleSelection toggles the selection state of the item at the given index.
func (m *TeaList[T]) ToggleSelection(index int) bool {
	// Get current selection state
	if data, ok := m.list.GetCurrentItem(); ok && m.list.State.CursorIndex == index {
		newSelected := !data.Selected
		if m.list.DataProvider.SetSelected(index, newSelected) {
			// Use efficient cache refresh instead of full data reload
			m.refreshCachedData()
			return true
		}
	} else {
		// If not the current item, we need to determine current state differently
		// For now, just set as selected
		if m.list.DataProvider.SetSelected(index, true) {
			// Use efficient cache refresh instead of full data reload
			m.refreshCachedData()
			return true
		}
	}
	return false
}

// ToggleCurrentSelection toggles the selection state of the currently highlighted item.
func (m *TeaList[T]) ToggleCurrentSelection() bool {
	currentIndex := m.list.State.CursorIndex
	if data, ok := m.list.GetCurrentItem(); ok {
		newSelected := !data.Selected
		if m.list.DataProvider.SetSelected(currentIndex, newSelected) {
			// Force refresh of cached data and update visible items
			m.refreshCachedData()
			return true
		}
	}
	return false
}

// SelectAll selects all items.
func (m *TeaList[T]) SelectAll() bool {
	if m.list.DataProvider.SelectAll() {
		// Just invalidate cache - let normal render cycle fetch fresh data
		m.refreshCachedData()
		return true
	}
	return false
}

// ClearSelection clears all selections.
func (m *TeaList[T]) ClearSelection() {
	m.list.DataProvider.ClearSelection()
	// Just invalidate cache - let normal render cycle fetch fresh data
	m.refreshCachedData()
}

// refreshCachedData invalidates cached chunks to force refresh of visible data
// This should ONLY affect visual representation, NOT trigger data provider calls
func (m *TeaList[T]) refreshCachedData() {
	// DON'T clear chunks - that would trigger unnecessary data fetching
	// Instead, we should only update the selection state in existing chunks
	m.updateSelectionInVisibleChunks()
}

// updateSelectionInVisibleChunks updates selection state in currently loaded chunks
// without triggering any data provider calls
func (m *TeaList[T]) updateSelectionInVisibleChunks() {
	selectedIndices := m.list.DataProvider.GetSelectedIndices()
	selectedSet := make(map[int]bool)
	for _, idx := range selectedIndices {
		selectedSet[idx] = true
	}

	// Update selection state in all loaded chunks
	for _, chunk := range m.list.chunks {
		for i := range chunk.Items {
			absoluteIndex := chunk.StartIndex + i
			chunk.Items[i].Selected = selectedSet[absoluteIndex]
		}
	}

	// Update visible items to reflect selection changes
	m.list.updateVisibleItems()
}

// GetSelectedIndices returns the indices of all selected items.
func (m *TeaList[T]) GetSelectedIndices() []int {
	return m.list.DataProvider.GetSelectedIndices()
}

// GetSelectionCount returns the number of selected items.
func (m *TeaList[T]) GetSelectionCount() int {
	return len(m.list.DataProvider.GetSelectedIndices())
}

// EnableRealTimeUpdates enables periodic data refreshing for dynamic data sources
func (m *TeaList[T]) EnableRealTimeUpdates(interval time.Duration) {
	m.realTimeUpdates = true
	m.realTimeInterval = interval
	m.lastRealTimeUpdate = time.Now()
}

// DisableRealTimeUpdates disables periodic data refreshing
func (m *TeaList[T]) DisableRealTimeUpdates() {
	m.realTimeUpdates = false
}

// IsRealTimeUpdatesEnabled returns whether real-time updates are enabled
func (m *TeaList[T]) IsRealTimeUpdatesEnabled() bool {
	return m.realTimeUpdates
}

// ForceDataRefresh forces a complete data reload - use sparingly!
// This should only be called when you know the data structure has changed
func (m *TeaList[T]) ForceDataRefresh() {
	m.list.refreshData()
}

// updateAnimationContent updates animation content when animation update messages are received
func (m *TeaList[T]) updateAnimationContent() {
	// Handle animations when animation update messages are received
	// This decouples animation updates from cursor movements
	if m.animatedFormatter != nil {
		m.processAnimations()
	}
}

// updateCursorInCache updates the cache for the cursor position
func (m *TeaList[T]) updateCursorInCache(oldIndex, newIndex int) {
	visibleItems := m.list.GetVisibleItems()
	state := m.list.GetState()

	// Find which viewport positions the old and new cursor positions map to
	oldViewportIndex := -1
	newViewportIndex := -1

	for i := range visibleItems {
		absoluteIndex := state.ViewportStartIndex + i
		if absoluteIndex == oldIndex {
			oldViewportIndex = i
		}
		if absoluteIndex == newIndex {
			newViewportIndex = i
		}
	}

	// Update cache for old cursor position (remove cursor)
	if oldViewportIndex >= 0 && oldViewportIndex < len(visibleItems) {
		m.updateSingleItemCache(oldViewportIndex, false) // Not cursor anymore
	}

	// Update cache for new cursor position (add cursor)
	if newViewportIndex >= 0 && newViewportIndex < len(visibleItems) {
		m.updateSingleItemCache(newViewportIndex, true) // Now cursor
	}
}

// updateSingleItemCache updates the cache for a single item with cursor state
func (m *TeaList[T]) updateSingleItemCache(viewportIndex int, isCursor bool) {
	visibleItems := m.list.GetVisibleItems()
	state := m.list.GetState()

	if viewportIndex >= len(visibleItems) {
		return
	}

	absoluteIndex := state.ViewportStartIndex + viewportIndex
	animationKey := fmt.Sprintf("item-%d", absoluteIndex)
	dataItem := visibleItems[viewportIndex]

	// Create render context
	ctx := DefaultRenderContext()
	ctx.CurrentTime = time.Now()
	ctx.DeltaTime = 0 // No delta time for cursor updates
	ctx.MaxWidth = 80

	// Get existing animation state (preserve it!)
	animState := m.animationEngine.GetAnimationState(animationKey)

	// Determine threshold states
	isTopThreshold := viewportIndex == m.list.Config.TopThresholdIndex
	isBottomThreshold := viewportIndex == m.list.Config.BottomThresholdIndex

	// Call animated formatter with updated cursor state
	result := m.animatedFormatter(dataItem, absoluteIndex, ctx, animState, isCursor, isTopThreshold, isBottomThreshold)

	// Update only the content in cache (DON'T re-register animations or update state)
	m.cachedAnimationContent[animationKey] = result.Content

	// DON'T call RegisterAnimation or UpdateAnimationState here - that would cause acceleration
	// The animation state should persist, only the visual content changes
}
