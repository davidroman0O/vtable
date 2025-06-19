// Package core provides the fundamental types, interfaces, and messages for the
// vtable library. It defines the shared data structures and contracts used by
// different components like List and Table, ensuring a consistent and
// interoperable architecture. This package is the foundation upon which all other
// vtable modules are built.
package core

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// DataSource defines the contract for providing data to vtable components like
// List and Table. It uses a command-based, asynchronous pattern to load data,
// preventing the UI from blocking. Implementations of this interface are
// responsible for fetching data from any source, such as a database, an API, or
// an in-memory slice.
type DataSource[T any] interface {
	// LoadChunk requests a chunk of data from the data source. It should return a
	// tea.Cmd that, when executed, produces either a DataChunkLoadedMsg on
	// success or a DataChunkErrorMsg on failure.
	LoadChunk(request DataRequest) tea.Cmd

	// GetTotal requests the total number of items available in the data source.
	// It should return a tea.Cmd that resolves to a DataTotalMsg.
	GetTotal() tea.Cmd

	// RefreshTotal requests an updated total count from the data source. This is
	// useful when the dataset size changes dynamically. It should return a
	// tea.Cmd that resolves to a DataTotalMsg.
	RefreshTotal() tea.Cmd

	// SetSelected sends a command to update the selection state of an item at a
	// specific index.
	SetSelected(index int, selected bool) tea.Cmd
	// SetSelectedByID sends a command to update the selection state of an item
	// with a specific ID.
	SetSelectedByID(id string, selected bool) tea.Cmd
	// SelectAll sends a command to select all items in the data source.
	SelectAll() tea.Cmd
	// ClearSelection sends a command to clear all selections.
	ClearSelection() tea.Cmd
	// SelectRange sends a command to select a range of items between two indices.
	SelectRange(startIndex, endIndex int) tea.Cmd

	// GetItemID returns the stable, unique ID for a given data item. This is a
	// pure function and can be called directly.
	GetItemID(item T) string
}

// SearchableDataSource extends the DataSource interface with search capabilities.
type SearchableDataSource[T any] interface {
	DataSource[T]

	// FindItemIndex searches for an item based on a key-value pair and returns a
	// tea.Cmd that resolves to a message containing the index of the found item.
	FindItemIndex(key string, value any) tea.Cmd
}

// ItemFormatter is a function that defines how a single list item is rendered
// into a string. It receives the item's data, its state (cursor, selection),
// and the render context.
type ItemFormatter[T any] func(
	data Data[T],
	index int,
	ctx RenderContext,
	isCursor bool,
	isTopThreshold bool,
	isBottomThreshold bool,
) string

// ItemFormatterAnimated is a function that defines how a list item is rendered
// with animation support. It returns a RenderResult containing the string
// content and animation metadata.
type ItemFormatterAnimated[T any] func(
	data Data[T],
	index int,
	ctx RenderContext,
	animationState map[string]any,
	isCursor bool,
	isTopThreshold bool,
	isBottomThreshold bool,
) RenderResult

// SimpleCellFormatter is a function that defines how a table cell's content is
// rendered. It is "simple" because it receives the final cell value and does
// not need to handle complex data structures; the table component manages data
// extraction. It automatically handles text truncation based on column width.
type SimpleCellFormatter func(
	cellValue string,
	rowIndex int,
	column TableColumn,
	ctx RenderContext,
	isCursor bool,
	isSelected bool,
	isActiveCell bool,
) string

// SimpleHeaderFormatter is a function that defines how a table header cell is
// rendered. It automatically handles text truncation based on column width.
type SimpleHeaderFormatter func(
	column TableColumn,
	ctx RenderContext,
) string

// CellFormatter defines how a single table cell is rendered.
//
// Deprecated: Use SimpleCellFormatter instead, which simplifies the function
// signature and automatically handles truncation.
type CellFormatter func(
	cellValue string,
	rowIndex int,
	columnIndex int,
	column TableColumn,
	ctx RenderContext,
	isCursor bool,
	isSelected bool,
	isTopThreshold bool,
	isBottomThreshold bool,
) string

// CellFormatterAnimated is a function that defines how a table cell is rendered
// with animation support. It returns a CellRenderResult containing the string
// content and animation metadata.
type CellFormatterAnimated func(
	cellValue string,
	rowIndex int,
	columnIndex int,
	column TableColumn,
	ctx RenderContext,
	animationState map[string]any,
	isCursor bool,
	isSelected bool,
	isTopThreshold bool,
	isBottomThreshold bool,
) CellRenderResult

// RowFormatter is a function that defines how an entire table row is rendered
// from its constituent cell results. This allows for custom row-level styling
// or layout.
type RowFormatter func(
	row TableRow,
	columns []TableColumn,
	cellResults []CellRenderResult,
	ctx RenderContext,
	isCursor bool,
	isSelected bool,
) string

// HeaderFormatter is a function that defines how the entire table header is
// rendered from its constituent column titles.
type HeaderFormatter func(
	columns []TableColumn,
	ctx RenderContext,
) string

// LoadingRowFormatter is a function that defines how a placeholder row is
// rendered while data is being loaded.
type LoadingRowFormatter func(
	index int,
	columns []TableColumn,
	ctx RenderContext,
	isCursor bool,
) string

// HeaderCellFormatter defines how an individual header cell is rendered.
//
// Deprecated: Use SimpleHeaderFormatter instead for a simpler API and automatic
// truncation.
type HeaderCellFormatter func(
	column TableColumn,
	columnIndex int,
	ctx RenderContext,
) string

// TeaModel is the core interface that all vtable components (List, Table, etc.)
// implement. It extends the standard bubbletea.Model with common functionality
// for focus management, state access, and selection.
type TeaModel interface {
	tea.Model

	// Focus sets the component to a focused state, allowing it to receive
	// keyboard input.
	Focus() tea.Cmd
	// Blur removes focus from the component.
	Blur()
	// IsFocused returns true if the component currently has focus.
	IsFocused() bool

	// State access
	// GetState returns the current viewport state of the component.
	GetState() ViewportState
	// GetTotalItems returns the total number of items in the dataset.
	GetTotalItems() int

	// Selection
	// GetSelectedIndices returns the indices of all selected items.
	GetSelectedIndices() []int
	// GetSelectedIDs returns the stable IDs of all selected items.
	GetSelectedIDs() []string
	// GetSelectionCount returns the total number of selected items.
	GetSelectionCount() int
}

// ListModel extends TeaModel with methods specific to the List component.
type ListModel interface {
	TeaModel

	// SetFormatter sets a custom item formatter and returns a command.
	SetFormatter(formatter ItemFormatter[any]) tea.Cmd
	// SetAnimatedFormatter sets a custom animated item formatter and returns a command.
	SetAnimatedFormatter(formatter ItemFormatterAnimated[any]) tea.Cmd
	// SetMaxWidth sets the maximum width of the list and returns a command.
	SetMaxWidth(width int) tea.Cmd
	// GetCurrentItem returns the data item currently under the cursor.
	GetCurrentItem() (Data[any], bool)
}

// TableModel extends TeaModel with methods specific to the Table component.
type TableModel interface {
	TeaModel

	// SetColumns sets the table's column configuration and returns a command.
	SetColumns(columns []TableColumn) tea.Cmd
	// SetHeaderVisibility sets the visibility of the table header and returns a command.
	SetHeaderVisibility(visible bool) tea.Cmd
	// SetBorderVisibility sets the visibility of table borders and returns a command.
	SetBorderVisibility(visible bool) tea.Cmd
	// SetCellFormatter sets a custom formatter for a specific column and returns a command.
	SetCellFormatter(columnIndex int, formatter CellFormatter) tea.Cmd
	// SetCellAnimatedFormatter sets a custom animated formatter for a specific column and returns a command.
	SetCellAnimatedFormatter(columnIndex int, formatter CellFormatterAnimated) tea.Cmd
	// SetRowFormatter sets a custom row formatter and returns a command.
	SetRowFormatter(formatter RowFormatter) tea.Cmd
	// SetHeaderFormatter sets a custom header formatter and returns a command.
	SetHeaderFormatter(formatter HeaderFormatter) tea.Cmd
	// SetColumnConstraints sets layout constraints for a specific column and returns a command.
	SetColumnConstraints(columnIndex int, constraints CellConstraint) tea.Cmd
	// GetCurrentRow returns the data for the row currently under the cursor.
	GetCurrentRow() (TableRow, bool)
}

// AnimationEngine defines the contract for an animation management system. It
// handles the registration, state updates, and rendering lifecycle of animations.
type AnimationEngine interface {
	// StartLoop starts the animation engine's update loop, returning a command
	// that produces tick messages.
	StartLoop() tea.Cmd
	// StopLoop stops the animation engine's update loop.
	StopLoop()
	// IsRunning returns true if the animation loop is active.
	IsRunning() bool

	// RegisterAnimation registers a new animation with the engine.
	RegisterAnimation(id string, triggers []RefreshTrigger, initialState map[string]any) tea.Cmd
	// UnregisterAnimation removes an animation from the engine.
	UnregisterAnimation(id string) tea.Cmd
	// SetVisible notifies the engine whether an animated item is currently visible.
	SetVisible(id string, visible bool)
	// IsVisible checks if an animated item is marked as visible.
	IsVisible(id string) bool

	// GetAnimationState retrieves the current state of a specific animation.
	GetAnimationState(id string) map[string]any
	// UpdateAnimationState updates the state of a specific animation.
	UpdateAnimationState(id string, newState map[string]any)
	// HasUpdates returns true if any animations have changed since the last check.
	HasUpdates() bool
	// ClearDirtyFlags resets the update flags for all animations.
	ClearDirtyFlags()

	// GetConfig returns the current animation configuration.
	GetConfig() AnimationConfig
	// UpdateConfig applies a new configuration to the engine.
	UpdateConfig(config AnimationConfig) tea.Cmd

	// ProcessGlobalTick processes a global tick message to update timer-based animations.
	ProcessGlobalTick(msg GlobalAnimationTickMsg) tea.Cmd
	// ProcessEvent processes an event, triggering any event-based animations.
	ProcessEvent(event string) []string
	// CheckConditionalTriggers evaluates all conditional triggers.
	CheckConditionalTriggers() []string

	// Cleanup releases any resources used by the animation engine.
	Cleanup()
}

// Validator provides methods for validating and fixing configuration structs.
type Validator interface {
	// Validate checks if the configuration is valid and returns an error if not.
	Validate() error

	// Fix attempts to correct common configuration issues.
	Fix() error
}

// ViewportConfigValidator validates ViewportConfig structs.
type ViewportConfigValidator interface {
	Validator
	ValidateViewportConfig(config *ViewportConfig) error
	FixViewportConfig(config *ViewportConfig) error
}

// TableConfigValidator validates TableConfig structs.
type TableConfigValidator interface {
	Validator
	ValidateTableConfig(config *TableConfig) error
	FixTableConfig(config *TableConfig) error
}

// MetadataManager provides a type-safe way to manage metadata.
type MetadataManager interface {
	// Set stores a value with its associated key.
	Set(key string, value any) error
	// Get retrieves a value by its key.
	Get(key string) (any, bool)
	// Delete removes a key-value pair.
	Delete(key string)
	// Clear removes all key-value pairs.
	Clear()

	// GetRaw returns the raw, untyped metadata map.
	GetRaw() map[string]any
	// SetRaw sets a raw key-value pair.
	SetRaw(key string, value any)
	// Copy creates a deep copy of the metadata manager.
	Copy() MetadataManager
}

// Truncator provides a method for truncating text.
type Truncator interface {
	// Truncate shortens the given text to the specified maximum width.
	Truncate(text string, maxWidth int) string
}

// Wrapper provides a method for wrapping text.
type Wrapper interface {
	// Wrap breaks the given text into multiple lines, each no wider than the
	// specified maximum width.
	Wrap(text string, maxWidth int) []string
}

// Measurer provides a method for measuring text dimensions.
type Measurer interface {
	// Measure returns the width and height of the given text.
	Measure(text string) (width int, height int)
}

// TextProcessor combines the Truncator, Wrapper, and Measurer interfaces.
type TextProcessor interface {
	Truncator
	Wrapper
	Measurer
}

// SelectionManager defines the contract for managing item selections.
type SelectionManager interface {
	// IsSelected checks if an item with the given ID is selected.
	IsSelected(id string) bool
	// GetSelectedIDs returns a slice of IDs for all selected items.
	GetSelectedIDs() []string
	// GetSelectedIndices returns a slice of indices for all selected items.
	GetSelectedIndices() []int
	// GetSelectionCount returns the total number of selected items.
	GetSelectionCount() int

	// Select marks an item as selected.
	Select(id string) bool
	// Deselect removes an item from the selection.
	Deselect(id string) bool
	// Toggle flips the selection state of an item.
	Toggle(id string) bool
	// SelectAll marks all given IDs as selected.
	SelectAll(ids []string) bool
	// ClearSelection removes all selections.
	ClearSelection()
	// SelectRange selects all items between a start and end ID.
	SelectRange(startID, endID string) bool

	// GetSelectionMode returns the current selection mode.
	GetSelectionMode() SelectionMode
	// SetSelectionMode sets the selection mode.
	SetSelectionMode(mode SelectionMode)
}

// ThemeProvider defines the contract for managing themes and styles.
type ThemeProvider interface {
	// GetTheme returns the currently active theme.
	GetTheme() Theme
	// SetTheme sets a new theme and returns a command.
	SetTheme(theme Theme) tea.Cmd

	// GetStyleConfig returns the currently active style configuration.
	GetStyleConfig() StyleConfig
	// SetStyleConfig sets a new style configuration and returns a command.
	SetStyleConfig(config StyleConfig) tea.Cmd

	// ApplyTheme applies a named style from the theme to the given content.
	ApplyTheme(content string, style string) string
	// GetBorderChars returns the character set used for borders.
	GetBorderChars() BorderChars
	// SetBorderChars sets a new character set for borders and returns a command.
	SetBorderChars(chars BorderChars) tea.Cmd
}

// Configurable represents a component whose configuration can be managed.
type Configurable interface {
	// GetConfig returns the component's current configuration.
	GetConfig() any
	// SetConfig applies a new configuration and returns a command.
	SetConfig(config any) tea.Cmd
	// ValidateConfig checks if a given configuration is valid.
	ValidateConfig(config any) error
	// ResetToDefaults resets the component's configuration to its default
	// values and returns a command.
	ResetToDefaults() tea.Cmd
}

// ListConfigurable extends Configurable for List components.
type ListConfigurable interface {
	Configurable
	GetListConfig() ListConfig
	SetListConfig(config ListConfig) tea.Cmd
}

// TableConfigurable extends Configurable for Table components.
type TableConfigurable interface {
	Configurable
	GetTableConfig() TableConfig
	SetTableConfig(config TableConfig) tea.Cmd
}

// ErrorHandler defines the contract for components that manage error states.
type ErrorHandler interface {
	// HandleError processes and stores an error.
	HandleError(err error)
	// GetLastError returns the most recent error.
	GetLastError() error
	// ClearErrors removes all stored errors.
	ClearErrors()
	// HasErrors returns true if there are any stored errors.
	HasErrors() bool

	// GetErrorCount returns the total number of stored errors.
	GetErrorCount() int
	// GetErrors returns all stored errors.
	GetErrors() []error
}

// PerformanceMonitor defines the contract for tracking performance metrics.
type PerformanceMonitor interface {
	// GetRenderTime returns the time taken for the last render cycle.
	GetRenderTime() time.Duration
	// GetUpdateTime returns the time taken for the last update cycle.
	GetUpdateTime() time.Duration
	// GetMemoryUsage returns the current memory usage in bytes.
	GetMemoryUsage() uint64

	// StartMonitoring begins tracking performance metrics.
	StartMonitoring()
	// StopMonitoring stops tracking performance metrics.
	StopMonitoring()
	// IsMonitoring returns true if the monitor is currently active.
	IsMonitoring() bool

	// GetReport returns a summary of performance metrics over a period.
	GetReport() PerformanceReport
	// ResetMetrics resets all tracked metrics to zero.
	ResetMetrics()
}

// PerformanceReport contains a summary of performance metrics.
type PerformanceReport struct {
	AverageRenderTime time.Duration
	AverageUpdateTime time.Duration
	PeakMemoryUsage   uint64
	TotalOperations   int64
	Uptime            time.Duration
}

// EventHandler defines callbacks for various component events.
type EventHandler interface {
	// OnSelect registers a callback to be invoked when an item is selected.
	OnSelect(callback func(item any, index int))
	// OnDeselect registers a callback to be invoked when an item is deselected.
	OnDeselect(callback func(item any, index int))
	// OnSelectionChange registers a callback to be invoked when the overall
	// selection changes.
	OnSelectionChange(callback func(selectedItems []any))

	// OnCursorMove registers a callback for when the cursor moves.
	OnCursorMove(callback func(from, to int))
	// OnScroll registers a callback for when the viewport scrolls.
	OnScroll(callback func(state ViewportState))

	// OnDataLoad registers a callback for when data is successfully loaded.
	OnDataLoad(callback func(items []any))
	// OnDataError registers a callback for when a data loading error occurs.
	OnDataError(callback func(err error))

	// OnFocus registers a callback for when the component gains focus.
	OnFocus(callback func())
	// OnBlur registers a callback for when the component loses focus.
	OnBlur(callback func())
}

// Debugger provides methods for accessing internal debugging information.
type Debugger interface {
	// GetDebugInfo returns a string with general debugging information.
	GetDebugInfo() string
	// GetStateInfo returns a string with detailed state information.
	GetStateInfo() string
	// GetChunkInfo returns a string with information about loaded data chunks.
	GetChunkInfo() string
	// GetAnimationInfo returns a string with information about active animations.
	GetAnimationInfo() string

	// EnableDebug toggles the debugging mode.
	EnableDebug(enabled bool)
	// IsDebugEnabled returns true if debugging is enabled.
	IsDebugEnabled() bool
	// SetDebugLevel sets the verbosity of debugging information.
	SetDebugLevel(level DebugLevel)
	// GetDebugLevel returns the current debug level.
	GetDebugLevel() DebugLevel
}

// DebugLevel defines the verbosity of debugging information.
type DebugLevel int

// Constants for different debug levels.
const (
	DebugLevelNone DebugLevel = iota
	DebugLevelBasic
	DebugLevelDetailed
	DebugLevelVerbose
)

// AccessibilityProvider defines the contract for managing accessibility features.
type AccessibilityProvider interface {
	// GetAriaLabel returns the ARIA label for the component.
	GetAriaLabel() string
	// SetAriaLabel sets the ARIA label for the component.
	SetAriaLabel(label string)
	// GetDescription returns the accessible description for the component.
	GetDescription() string
	// SetDescription sets the accessible description for the component.
	SetDescription(desc string)

	// IsHighContrast returns true if high contrast mode is enabled.
	IsHighContrast() bool
	// SetHighContrast enables or disables high contrast mode.
	SetHighContrast(enabled bool) tea.Cmd

	// IsReducedMotion returns true if reduced motion mode is enabled.
	IsReducedMotion() bool
	// SetReducedMotion enables or disables reduced motion mode.
	SetReducedMotion(enabled bool) tea.Cmd

	// GetKeyboardShortcuts returns the list of available keyboard shortcuts.
	GetKeyboardShortcuts() []KeyboardShortcut
	// SetKeyboardShortcuts sets a new list of keyboard shortcuts.
	SetKeyboardShortcuts(shortcuts []KeyboardShortcut) tea.Cmd
}

// KeyboardShortcut represents a single keyboard shortcut and its description.
type KeyboardShortcut struct {
	Key         string
	Description string
	Action      string
}
