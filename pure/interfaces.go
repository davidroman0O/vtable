package vtable

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ================================
// DATA SOURCE INTERFACE
// ================================

// DataSource provides data loading capabilities for pure Tea components.
// Unlike the old DataProvider, this returns commands instead of blocking calls.
type DataSource[T any] interface {
	// LoadChunk loads a chunk of data asynchronously via a command
	LoadChunk(request DataRequest) tea.Cmd

	// GetTotal returns the total number of items via a command
	GetTotal() tea.Cmd

	// RefreshTotal refreshes the total count via a command
	RefreshTotal() tea.Cmd

	// Selection operations (async via commands)
	SetSelected(index int, selected bool) tea.Cmd
	SetSelectedByID(id string, selected bool) tea.Cmd
	SelectAll() tea.Cmd
	ClearSelection() tea.Cmd
	SelectRange(startIndex, endIndex int) tea.Cmd

	// Pure functions (no state mutation, can be called directly)
	GetItemID(item T) string
}

// SearchableDataSource extends DataSource with search capabilities
type SearchableDataSource[T any] interface {
	DataSource[T]

	// FindItemIndex searches for an item and returns a command
	FindItemIndex(key string, value any) tea.Cmd
}

// ================================
// FORMATTER INTERFACES
// ================================

// ItemFormatter formats a single list item
type ItemFormatter[T any] func(
	data Data[T],
	index int,
	ctx RenderContext,
	isCursor bool,
	isTopThreshold bool,
	isBottomThreshold bool,
) string

// ItemFormatterAnimated formats a list item with animation support
type ItemFormatterAnimated[T any] func(
	data Data[T],
	index int,
	ctx RenderContext,
	animationState map[string]any,
	isCursor bool,
	isTopThreshold bool,
	isBottomThreshold bool,
) RenderResult

// CellFormatter formats a single table cell
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

// CellFormatterAnimated formats a table cell with animation support
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

// RowFormatter formats an entire table row
type RowFormatter func(
	row TableRow,
	columns []TableColumn,
	cellResults []CellRenderResult,
	ctx RenderContext,
	isCursor bool,
	isSelected bool,
) string

// HeaderFormatter formats the table header
type HeaderFormatter func(
	columns []TableColumn,
	ctx RenderContext,
) string

// LoadingRowFormatter formats loading placeholder rows in tables
type LoadingRowFormatter func(
	index int,
	columns []TableColumn,
	ctx RenderContext,
	isCursor bool,
) string

// HeaderCellFormatter formats individual header cells in tables
type HeaderCellFormatter func(
	column TableColumn,
	columnIndex int,
	ctx RenderContext,
) string

// ================================
// MODEL INTERFACES
// ================================

// TeaModel represents the core interface that both List and Table implement
type TeaModel interface {
	tea.Model

	// Focus management
	Focus() tea.Cmd
	Blur() tea.Cmd
	IsFocused() bool

	// State access
	GetState() ViewportState
	GetTotalItems() int

	// Selection
	GetSelectedIndices() []int
	GetSelectedIDs() []string
	GetSelectionCount() int
}

// ListModel extends TeaModel with list-specific operations
type ListModel interface {
	TeaModel

	// List-specific operations
	SetFormatter(formatter ItemFormatter[any]) tea.Cmd
	SetAnimatedFormatter(formatter ItemFormatterAnimated[any]) tea.Cmd
	SetMaxWidth(width int) tea.Cmd
	GetCurrentItem() (Data[any], bool)
}

// TableModel extends TeaModel with table-specific operations
type TableModel interface {
	TeaModel

	// Table-specific operations
	SetColumns(columns []TableColumn) tea.Cmd
	SetHeaderVisibility(visible bool) tea.Cmd
	SetBorderVisibility(visible bool) tea.Cmd
	SetCellFormatter(columnIndex int, formatter CellFormatter) tea.Cmd
	SetCellAnimatedFormatter(columnIndex int, formatter CellFormatterAnimated) tea.Cmd
	SetRowFormatter(formatter RowFormatter) tea.Cmd
	SetHeaderFormatter(formatter HeaderFormatter) tea.Cmd
	SetColumnConstraints(columnIndex int, constraints CellConstraint) tea.Cmd
	GetCurrentRow() (TableRow, bool)
}

// ================================
// ANIMATION INTERFACES
// ================================

// AnimationEngine provides animation management capabilities
type AnimationEngine interface {
	// Lifecycle
	StartLoop() tea.Cmd
	StopLoop()
	IsRunning() bool

	// Animation management
	RegisterAnimation(id string, triggers []RefreshTrigger, initialState map[string]any) tea.Cmd
	UnregisterAnimation(id string) tea.Cmd
	SetVisible(id string, visible bool)
	IsVisible(id string) bool

	// State management
	GetAnimationState(id string) map[string]any
	UpdateAnimationState(id string, newState map[string]any)
	HasUpdates() bool
	ClearDirtyFlags()

	// Configuration
	GetConfig() AnimationConfig
	UpdateConfig(config AnimationConfig) tea.Cmd

	// Processing
	ProcessGlobalTick(msg GlobalAnimationTickMsg) tea.Cmd
	ProcessEvent(event string) []string
	CheckConditionalTriggers() []string

	// Cleanup
	Cleanup()
}

// ================================
// VALIDATION INTERFACES
// ================================

// Validator provides validation capabilities for configuration
type Validator interface {
	// Validate checks if the configuration is valid
	Validate() error

	// Fix attempts to fix common configuration issues
	Fix() error
}

// ViewportConfigValidator validates viewport configuration
type ViewportConfigValidator interface {
	Validator
	ValidateViewportConfig(config *ViewportConfig) error
	FixViewportConfig(config *ViewportConfig) error
}

// TableConfigValidator validates table configuration
type TableConfigValidator interface {
	Validator
	ValidateTableConfig(config *TableConfig) error
	FixTableConfig(config *TableConfig) error
}

// ================================
// METADATA INTERFACES
// ================================

// MetadataManager provides type-safe metadata operations
type MetadataManager interface {
	// Type-safe operations
	Set(key string, value any) error
	Get(key string) (any, bool)
	Delete(key string)
	Clear()

	// Utility operations
	GetRaw() map[string]any
	SetRaw(key string, value any)
	Copy() MetadataManager
}

// ================================
// UTILITY INTERFACES
// ================================

// Truncator provides text truncation capabilities
type Truncator interface {
	Truncate(text string, maxWidth int) string
}

// Wrapper provides text wrapping capabilities
type Wrapper interface {
	Wrap(text string, maxWidth int) []string
}

// Measurer provides text measurement capabilities
type Measurer interface {
	Measure(text string) (width int, height int)
}

// TextProcessor combines truncation, wrapping, and measurement
type TextProcessor interface {
	Truncator
	Wrapper
	Measurer
}

// ================================
// SELECTION INTERFACES
// ================================

// SelectionManager handles selection state and operations
type SelectionManager interface {
	// Selection state
	IsSelected(id string) bool
	GetSelectedIDs() []string
	GetSelectedIndices() []int
	GetSelectionCount() int

	// Selection operations
	Select(id string) bool
	Deselect(id string) bool
	Toggle(id string) bool
	SelectAll(ids []string) bool
	ClearSelection()
	SelectRange(startID, endID string) bool

	// Configuration
	GetSelectionMode() SelectionMode
	SetSelectionMode(mode SelectionMode)
}

// ================================
// THEME INTERFACES
// ================================

// ThemeProvider provides theming capabilities
type ThemeProvider interface {
	// Theme access
	GetTheme() Theme
	SetTheme(theme Theme) tea.Cmd

	// Style access
	GetStyleConfig() StyleConfig
	SetStyleConfig(config StyleConfig) tea.Cmd

	// Utility methods
	ApplyTheme(content string, style string) string
	GetBorderChars() BorderChars
	SetBorderChars(chars BorderChars) tea.Cmd
}

// ================================
// CONFIGURATION INTERFACES
// ================================

// Configurable represents something that can be configured
type Configurable interface {
	// Configuration access
	GetConfig() any
	SetConfig(config any) tea.Cmd
	ValidateConfig(config any) error
	ResetToDefaults() tea.Cmd
}

// ListConfigurable extends Configurable for lists
type ListConfigurable interface {
	Configurable
	GetListConfig() ListConfig
	SetListConfig(config ListConfig) tea.Cmd
}

// TableConfigurable extends Configurable for tables
type TableConfigurable interface {
	Configurable
	GetTableConfig() TableConfig
	SetTableConfig(config TableConfig) tea.Cmd
}

// ================================
// ERROR HANDLING INTERFACES
// ================================

// ErrorHandler provides error handling capabilities
type ErrorHandler interface {
	// Error handling
	HandleError(err error)
	GetLastError() error
	ClearErrors()
	HasErrors() bool

	// Error reporting
	GetErrorCount() int
	GetErrors() []error
}

// ================================
// PERFORMANCE INTERFACES
// ================================

// PerformanceMonitor tracks performance metrics
type PerformanceMonitor interface {
	// Metrics
	GetRenderTime() time.Duration
	GetUpdateTime() time.Duration
	GetMemoryUsage() uint64

	// Monitoring
	StartMonitoring()
	StopMonitoring()
	IsMonitoring() bool

	// Reporting
	GetReport() PerformanceReport
	ResetMetrics()
}

// PerformanceReport contains performance metrics
type PerformanceReport struct {
	AverageRenderTime time.Duration
	AverageUpdateTime time.Duration
	PeakMemoryUsage   uint64
	TotalOperations   int64
	Uptime            time.Duration
}

// ================================
// CALLBACK INTERFACES
// ================================

// EventHandler handles various events
type EventHandler interface {
	// Selection events
	OnSelect(callback func(item any, index int))
	OnDeselect(callback func(item any, index int))
	OnSelectionChange(callback func(selectedItems []any))

	// Navigation events
	OnCursorMove(callback func(from, to int))
	OnScroll(callback func(state ViewportState))

	// Data events
	OnDataLoad(callback func(items []any))
	OnDataError(callback func(err error))

	// Focus events
	OnFocus(callback func())
	OnBlur(callback func())
}

// ================================
// DEBUGGING INTERFACES
// ================================

// Debugger provides debugging capabilities
type Debugger interface {
	// Debug information
	GetDebugInfo() string
	GetStateInfo() string
	GetChunkInfo() string
	GetAnimationInfo() string

	// Debug controls
	EnableDebug(enabled bool)
	IsDebugEnabled() bool
	SetDebugLevel(level DebugLevel)
	GetDebugLevel() DebugLevel
}

// DebugLevel defines different levels of debugging information
type DebugLevel int

const (
	DebugLevelNone DebugLevel = iota
	DebugLevelBasic
	DebugLevelDetailed
	DebugLevelVerbose
)

// ================================
// ACCESSIBILITY INTERFACES
// ================================

// AccessibilityProvider provides accessibility features
type AccessibilityProvider interface {
	// Screen reader support
	GetAriaLabel() string
	SetAriaLabel(label string)
	GetDescription() string
	SetDescription(desc string)

	// High contrast support
	IsHighContrast() bool
	SetHighContrast(enabled bool) tea.Cmd

	// Reduced motion support
	IsReducedMotion() bool
	SetReducedMotion(enabled bool) tea.Cmd

	// Keyboard navigation
	GetKeyboardShortcuts() []KeyboardShortcut
	SetKeyboardShortcuts(shortcuts []KeyboardShortcut) tea.Cmd
}

// KeyboardShortcut represents a keyboard shortcut
type KeyboardShortcut struct {
	Key         string
	Description string
	Action      string
}
