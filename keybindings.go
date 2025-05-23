package vtable

import (
	"fmt"
	"runtime"

	"github.com/charmbracelet/bubbles/key"
)

// PlatformType represents a specific operating system platform
type PlatformType string

const (
	PlatformMacOS   PlatformType = "darwin"
	PlatformLinux   PlatformType = "linux"
	PlatformWindows PlatformType = "windows"
	PlatformUnknown PlatformType = "unknown"
)

// DetectPlatform returns the current platform based on runtime.GOOS
func DetectPlatform() PlatformType {
	switch runtime.GOOS {
	case "darwin":
		return PlatformMacOS
	case "linux":
		return PlatformLinux
	case "windows":
		return PlatformWindows
	default:
		return PlatformUnknown
	}
}

// NavigationKeyMap defines the key mappings for navigation in components.
type NavigationKeyMap struct {
	// Navigation keys
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Home     key.Binding
	End      key.Binding

	// Search
	Search key.Binding

	// Additional actions
	Select         key.Binding
	SelectAll      key.Binding
	ClearSelection key.Binding
	Back           key.Binding
}

// NavigationKeyMapWithHelp sets help text for all keys
func NavigationKeyMapWithHelp(km NavigationKeyMap) NavigationKeyMap {
	// Update keys only if they don't already have help text
	if km.Up.Help().Key == "" {
		km.Up = key.NewBinding(
			key.WithKeys(km.Up.Keys()...),
			key.WithHelp("↑", "up"),
		)
	}

	if km.Down.Help().Key == "" {
		km.Down = key.NewBinding(
			key.WithKeys(km.Down.Keys()...),
			key.WithHelp("↓", "down"),
		)
	}

	if km.PageUp.Help().Key == "" {
		km.PageUp = key.NewBinding(
			key.WithKeys(km.PageUp.Keys()...),
			key.WithHelp("pgup", "page up"),
		)
	}

	if km.PageDown.Help().Key == "" {
		km.PageDown = key.NewBinding(
			key.WithKeys(km.PageDown.Keys()...),
			key.WithHelp("pgdn", "page down"),
		)
	}

	if km.Home.Help().Key == "" {
		km.Home = key.NewBinding(
			key.WithKeys(km.Home.Keys()...),
			key.WithHelp("home", "go to top"),
		)
	}

	if km.End.Help().Key == "" {
		km.End = key.NewBinding(
			key.WithKeys(km.End.Keys()...),
			key.WithHelp("end", "go to bottom"),
		)
	}

	if km.Search.Help().Key == "" {
		km.Search = key.NewBinding(
			key.WithKeys(km.Search.Keys()...),
			key.WithHelp("f", "search"),
		)
	}

	if km.Select.Help().Key == "" {
		km.Select = key.NewBinding(
			key.WithKeys(km.Select.Keys()...),
			key.WithHelp("enter", "select"),
		)
	}

	if km.SelectAll.Help().Key == "" {
		km.SelectAll = key.NewBinding(
			key.WithKeys(km.SelectAll.Keys()...),
			key.WithHelp("ctrl+a", "select all"),
		)
	}

	if km.ClearSelection.Help().Key == "" {
		km.ClearSelection = key.NewBinding(
			key.WithKeys(km.ClearSelection.Keys()...),
			key.WithHelp("ctrl+x", "clear selection"),
		)
	}

	if km.Back.Help().Key == "" {
		km.Back = key.NewBinding(
			key.WithKeys(km.Back.Keys()...),
			key.WithHelp("esc", "back"),
		)
	}

	return km
}

// MacOSKeyMap returns key bindings optimized for macOS
func MacOSKeyMap() NavigationKeyMap {
	km := NavigationKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("u", "b"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("d", "space"),
		),
		Home: key.NewBinding(
			key.WithKeys("g"),
		),
		End: key.NewBinding(
			key.WithKeys("G"),
		),
		Search: key.NewBinding(
			key.WithKeys("f", "slash", "/"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("ctrl+a"),
		),
		ClearSelection: key.NewBinding(
			key.WithKeys("ctrl+x"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "q"),
		),
	}

	// Set Mac-specific help text
	km.PageUp = key.NewBinding(
		key.WithKeys(km.PageUp.Keys()...),
		key.WithHelp("u/b", "page up"),
	)

	km.PageDown = key.NewBinding(
		key.WithKeys(km.PageDown.Keys()...),
		key.WithHelp("d/space", "page down"),
	)

	km.Home = key.NewBinding(
		key.WithKeys(km.Home.Keys()...),
		key.WithHelp("g", "go to top"),
	)

	km.End = key.NewBinding(
		key.WithKeys(km.End.Keys()...),
		key.WithHelp("G", "go to bottom"),
	)

	// Apply standard help text to other keys
	km = NavigationKeyMapWithHelp(km)

	return km
}

// LinuxKeyMap returns key bindings optimized for Linux
func LinuxKeyMap() NavigationKeyMap {
	km := NavigationKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("u", "b", "pgup"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("d", "space", "pgdown"),
		),
		Home: key.NewBinding(
			key.WithKeys("g", "home"),
		),
		End: key.NewBinding(
			key.WithKeys("G", "end"),
		),
		Search: key.NewBinding(
			key.WithKeys("f", "slash", "/"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("ctrl+a"),
		),
		ClearSelection: key.NewBinding(
			key.WithKeys("ctrl+x"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "q"),
		),
	}

	return NavigationKeyMapWithHelp(km)
}

// WindowsKeyMap returns key bindings optimized for Windows
func WindowsKeyMap() NavigationKeyMap {
	km := NavigationKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("u", "b", "pgup"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("d", "space", "pgdown"),
		),
		Home: key.NewBinding(
			key.WithKeys("g", "home"),
		),
		End: key.NewBinding(
			key.WithKeys("G", "end"),
		),
		Search: key.NewBinding(
			key.WithKeys("f", "slash", "/"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("ctrl+a"),
		),
		ClearSelection: key.NewBinding(
			key.WithKeys("ctrl+x"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "q"),
		),
	}

	return NavigationKeyMapWithHelp(km)
}

// PlatformKeyMap returns the appropriate key map for the current platform
func PlatformKeyMap() NavigationKeyMap {
	platform := DetectPlatform()

	switch platform {
	case PlatformMacOS:
		return MacOSKeyMap()
	case PlatformLinux:
		return LinuxKeyMap()
	case PlatformWindows:
		return WindowsKeyMap()
	default:
		// Use Linux bindings as default for unknown platforms
		return LinuxKeyMap()
	}
}

// GetKeyMapDescription returns a user-friendly description of the key bindings
func GetKeyMapDescription(km NavigationKeyMap) string {
	return fmt.Sprintf(
		"↑/↓: navigate • %s/%s: page up/down • %s/%s: top/bottom • %s: search • %s: select • %s: select all • %s: clear selection • %s: back",
		km.PageUp.Help().Key,
		km.PageDown.Help().Key,
		km.Home.Help().Key,
		km.End.Help().Key,
		km.Search.Help().Key,
		km.Select.Help().Key,
		km.SelectAll.Help().Key,
		km.ClearSelection.Help().Key,
		km.Back.Help().Key,
	)
}
