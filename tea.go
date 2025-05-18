package vtable

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// TeaList is a Bubble Tea model wrapping a List.
type TeaList[T any] struct {
	// The underlying list model
	List *List[T]

	// Key mappings
	KeyMap NavigationKeyMap

	// Whether the component is focused
	Focused bool
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
		List:    list,
		KeyMap:  PlatformKeyMap(), // Use platform-specific key bindings
		Focused: true,
	}, nil
}

// Init initializes the Tea model.
func (m *TeaList[T]) Init() tea.Cmd {
	return nil
}

// Update updates the Tea model based on messages.
func (m *TeaList[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// If not focused, don't handle messages
	if !m.Focused {
		return m, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Up):
			m.List.MoveUp()
		case key.Matches(msg, m.KeyMap.Down):
			m.List.MoveDown()
		case key.Matches(msg, m.KeyMap.PageUp):
			m.List.PageUp()
		case key.Matches(msg, m.KeyMap.PageDown):
			m.List.PageDown()
		case key.Matches(msg, m.KeyMap.Home):
			m.List.JumpToStart()
		case key.Matches(msg, m.KeyMap.End):
			m.List.JumpToEnd()
			// Note: Search and Select are handled by the parent application
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the Tea model.
func (m *TeaList[T]) View() string {
	return m.List.Render()
}

// Focus sets the focus state of the component.
func (m *TeaList[T]) Focus() {
	m.Focused = true
}

// Blur removes focus from the component.
func (m *TeaList[T]) Blur() {
	m.Focused = false
}

// IsFocused returns whether the component is focused.
func (m *TeaList[T]) IsFocused() bool {
	return m.Focused
}

// GetState returns the current viewport state.
func (m *TeaList[T]) GetState() ViewportState {
	return m.List.GetState()
}

// GetVisibleItems returns the slice of items currently visible in the viewport.
func (m *TeaList[T]) GetVisibleItems() []T {
	return m.List.GetVisibleItems()
}

// GetCurrentItem returns the currently selected item.
func (m *TeaList[T]) GetCurrentItem() (T, bool) {
	return m.List.GetCurrentItem()
}

// SetKeyMap sets the key mappings for the component.
func (m *TeaList[T]) SetKeyMap(keyMap NavigationKeyMap) {
	m.KeyMap = keyMap
}

// JumpToItem jumps to an item with the specified key-value pair.
// Returns true if the item was found and jumped to, false otherwise.
func (m *TeaList[T]) JumpToItem(key string, value any) bool {
	return m.List.JumpToItem(key, value)
}

// JumpToIndex jumps to the specified index in the dataset.
func (m *TeaList[T]) JumpToIndex(index int) {
	m.List.JumpToIndex(index)
}

// GetHelpView returns a string describing the key bindings.
func (m *TeaList[T]) GetHelpView() string {
	return GetKeyMapDescription(m.KeyMap)
}
