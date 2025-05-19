package vtable

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// TeaList is a Bubble Tea model wrapping a List.
type TeaList[T any] struct {
	// The underlying list model
	list *List[T]

	// Key mappings
	keyMap NavigationKeyMap

	// Whether the component is focused
	focused bool

	// Event callbacks
	onSelectItem func(item T, index int)
	onHighlight  func(item T, index int)
	onScroll     func(state ViewportState)
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

	return &TeaList[T]{
		list:    list,
		keyMap:  PlatformKeyMap(), // Use platform-specific key bindings
		focused: true,
	}, nil
}

// Init initializes the Tea model.
func (m *TeaList[T]) Init() tea.Cmd {
	return nil
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
	return m.list.Render()
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
	return m.list.GetVisibleItems()
}

// GetCurrentItem returns the currently selected item.
func (m *TeaList[T]) GetCurrentItem() (T, bool) {
	return m.list.GetCurrentItem()
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

// HandleKeypress programmatically simulates pressing a key.
func (m *TeaList[T]) HandleKeypress(keyStr string) {
	// Create a key message and update
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyStr)}
	m.Update(keyMsg)
}
