package vtable

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ================================
// DEFAULT CONFIGURATIONS
// ================================

// DefaultListConfig returns a sensible default configuration for lists
func DefaultListConfig() ListConfig {
	return ListConfig{
		ViewportConfig:  DefaultViewportConfig(),
		StyleConfig:     DefaultStyleConfig(),
		RenderConfig:    DefaultListRenderConfig(),
		AnimationConfig: DefaultAnimationConfig(),
		SelectionMode:   SelectionSingle,
		KeyMap:          DefaultNavigationKeyMap(),
		MaxWidth:        80,
	}
}

// DefaultTableConfig returns a sensible default configuration for tables
func DefaultTableConfig() TableConfig {
	return TableConfig{
		Columns:         []TableColumn{},
		ShowHeader:      true,
		ShowBorders:     true,
		ViewportConfig:  DefaultViewportConfig(),
		Theme:           DefaultTheme(),
		AnimationConfig: DefaultAnimationConfig(),
		SelectionMode:   SelectionSingle,
		KeyMap:          DefaultNavigationKeyMap(),
	}
}

// DefaultViewportConfig returns a sensible default viewport configuration
func DefaultViewportConfig() ViewportConfig {
	return ViewportConfig{
		Height:          10,
		TopThreshold:    2, // 2 positions from viewport start
		BottomThreshold: 2, // 2 positions from viewport end
		ChunkSize:       100,
		InitialIndex:    0,
	}
}

// DefaultStyleConfig returns a sensible default style configuration for lists
func DefaultStyleConfig() StyleConfig {
	return StyleConfig{
		CursorStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true),
		SelectedStyle:  lipgloss.NewStyle().Background(lipgloss.Color("57")).Foreground(lipgloss.Color("230")),
		DefaultStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		ThresholdStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true),
		DisabledStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
		LoadingStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Italic(true),
		ErrorStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
	}
}

// DefaultTheme returns a sensible default theme for tables
func DefaultTheme() Theme {
	return Theme{
		HeaderStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true),
		CellStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		CursorStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true),
		SelectedStyle:     lipgloss.NewStyle().Background(lipgloss.Color("57")).Foreground(lipgloss.Color("230")),
		BorderChars:       DefaultBorderChars(),
		BorderColor:       "241",
		HeaderColor:       "99",
		AlternateRowStyle: lipgloss.NewStyle().Background(lipgloss.Color("235")),
		DisabledStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
		LoadingStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Italic(true),
		ErrorStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
	}
}

// ================================
// CONFIGURATION VALIDATION
// ================================

// ValidateViewportConfig validates viewport configuration and returns any errors
func ValidateViewportConfig(config *ViewportConfig) []error {
	var errors []error

	if config.Height <= 0 {
		errors = append(errors, fmt.Errorf("viewport height must be positive, got %d", config.Height))
	}

	// Allow -1 as a valid threshold value (means disabled)
	if config.TopThreshold < -1 {
		errors = append(errors, fmt.Errorf("top threshold must be -1 (disabled) or non-negative, got %d", config.TopThreshold))
	}

	if config.BottomThreshold < -1 {
		errors = append(errors, fmt.Errorf("bottom threshold must be -1 (disabled) or non-negative, got %d", config.BottomThreshold))
	}

	if config.TopThreshold >= config.Height {
		errors = append(errors, fmt.Errorf("top threshold (%d) must be less than height (%d)", config.TopThreshold, config.Height))
	}

	if config.BottomThreshold >= config.Height {
		errors = append(errors, fmt.Errorf("bottom threshold (%d) must be less than height (%d)", config.BottomThreshold, config.Height))
	}

	if config.ChunkSize <= 0 {
		errors = append(errors, fmt.Errorf("chunk size must be positive, got %d", config.ChunkSize))
	}

	if config.InitialIndex < 0 {
		errors = append(errors, fmt.Errorf("initial index must be non-negative, got %d", config.InitialIndex))
	}

	return errors
}

// ValidateTableConfig validates table configuration and returns any errors
func ValidateTableConfig(config *TableConfig) []error {
	var errors []error

	// Validate viewport config
	errors = append(errors, ValidateViewportConfig(&config.ViewportConfig)...)

	// Validate columns
	if len(config.Columns) == 0 {
		errors = append(errors, fmt.Errorf("table must have at least one column"))
	}

	for i, col := range config.Columns {
		if col.Width <= 0 {
			errors = append(errors, fmt.Errorf("column %d width must be positive, got %d", i, col.Width))
		}
		if col.Title == "" {
			errors = append(errors, fmt.Errorf("column %d must have a title", i))
		}
		if col.Field == "" {
			errors = append(errors, fmt.Errorf("column %d must have a field", i))
		}
	}

	// Validate animation config
	errors = append(errors, ValidateAnimationConfig(&config.AnimationConfig)...)

	return errors
}

// ValidateListConfig validates list configuration and returns any errors
func ValidateListConfig(config *ListConfig) []error {
	var errors []error

	// Validate viewport config
	errors = append(errors, ValidateViewportConfig(&config.ViewportConfig)...)

	// Validate animation config
	errors = append(errors, ValidateAnimationConfig(&config.AnimationConfig)...)

	// Validate max width
	if config.MaxWidth <= 0 {
		errors = append(errors, fmt.Errorf("max width must be positive, got %d", config.MaxWidth))
	}

	return errors
}

// ValidateAnimationConfig validates animation configuration and returns any errors
func ValidateAnimationConfig(config *AnimationConfig) []error {
	var errors []error

	if config.MaxAnimations < 0 {
		errors = append(errors, fmt.Errorf("max animations must be non-negative, got %d", config.MaxAnimations))
	}

	if config.TickInterval <= 0 {
		errors = append(errors, fmt.Errorf("tick interval must be positive, got %v", config.TickInterval))
	}

	if config.TickInterval < 10*time.Millisecond {
		errors = append(errors, fmt.Errorf("tick interval too small (may cause performance issues), got %v", config.TickInterval))
	}

	return errors
}

// ================================
// CONFIGURATION FIXING
// ================================

// FixViewportConfig attempts to fix common issues in viewport configuration
func FixViewportConfig(config *ViewportConfig) {
	if config.Height <= 0 {
		config.Height = 10
	}

	// Don't "fix" thresholds that are -1 since -1 means disabled
	// Only fix thresholds that are < -1 (invalid) or >= height (out of bounds)
	if config.TopThreshold < -1 {
		config.TopThreshold = -1 // Set to disabled
	}

	if config.BottomThreshold < -1 {
		config.BottomThreshold = -1 // Set to disabled
	}

	if config.TopThreshold >= config.Height {
		config.TopThreshold = config.Height - 1
		if config.TopThreshold < 0 {
			config.TopThreshold = -1 // Set to disabled if height is too small
		}
	}

	if config.BottomThreshold >= config.Height {
		config.BottomThreshold = config.Height - 1
		if config.BottomThreshold < 0 {
			config.BottomThreshold = -1 // Set to disabled if height is too small
		}
	}

	if config.ChunkSize <= 0 {
		config.ChunkSize = 100
	}

	if config.InitialIndex < 0 {
		config.InitialIndex = 0
	}
}

// FixTableConfig attempts to fix common issues in table configuration
func FixTableConfig(config *TableConfig) {
	// Fix viewport config
	FixViewportConfig(&config.ViewportConfig)

	// Fix columns
	for i := range config.Columns {
		if config.Columns[i].Width <= 0 {
			config.Columns[i].Width = 10
		}
		if config.Columns[i].Title == "" {
			config.Columns[i].Title = fmt.Sprintf("Column %d", i+1)
		}
		if config.Columns[i].Field == "" {
			config.Columns[i].Field = fmt.Sprintf("field_%d", i)
		}
	}

	// Fix animation config
	FixAnimationConfig(&config.AnimationConfig)
}

// FixListConfig attempts to fix common issues in list configuration
func FixListConfig(config *ListConfig) {
	// Fix viewport config
	FixViewportConfig(&config.ViewportConfig)

	// Fix animation config
	FixAnimationConfig(&config.AnimationConfig)

	// Fix max width
	if config.MaxWidth <= 0 {
		config.MaxWidth = 80
	}
}

// FixAnimationConfig attempts to fix common issues in animation configuration
func FixAnimationConfig(config *AnimationConfig) {
	if config.MaxAnimations < 0 {
		config.MaxAnimations = 100
	}

	if config.TickInterval <= 0 {
		config.TickInterval = 100 * time.Millisecond
	}

	if config.TickInterval < 10*time.Millisecond {
		config.TickInterval = 10 * time.Millisecond
	}
}

// ================================
// CONFIGURATION BUILDERS
// ================================

// ListConfigBuilder provides a fluent API for building list configurations
type ListConfigBuilder struct {
	config ListConfig
}

// NewListConfigBuilder creates a new list configuration builder with defaults
func NewListConfigBuilder() *ListConfigBuilder {
	return &ListConfigBuilder{
		config: DefaultListConfig(),
	}
}

// WithViewportHeight sets the viewport height
func (b *ListConfigBuilder) WithViewportHeight(height int) *ListConfigBuilder {
	b.config.ViewportConfig.Height = height
	return b
}

// WithChunkSize sets the chunk size
func (b *ListConfigBuilder) WithChunkSize(size int) *ListConfigBuilder {
	b.config.ViewportConfig.ChunkSize = size
	return b
}

// WithSelectionMode sets the selection mode
func (b *ListConfigBuilder) WithSelectionMode(mode SelectionMode) *ListConfigBuilder {
	b.config.SelectionMode = mode
	return b
}

// WithMaxWidth sets the maximum width
func (b *ListConfigBuilder) WithMaxWidth(width int) *ListConfigBuilder {
	b.config.MaxWidth = width
	return b
}

// WithAnimationEnabled enables or disables animations
func (b *ListConfigBuilder) WithAnimationEnabled(enabled bool) *ListConfigBuilder {
	b.config.AnimationConfig.Enabled = enabled
	return b
}

// Build returns the configured ListConfig
func (b *ListConfigBuilder) Build() ListConfig {
	return b.config
}

// TableConfigBuilder provides a fluent API for building table configurations
type TableConfigBuilder struct {
	config TableConfig
}

// NewTableConfigBuilder creates a new table configuration builder with defaults
func NewTableConfigBuilder() *TableConfigBuilder {
	return &TableConfigBuilder{
		config: DefaultTableConfig(),
	}
}

// WithColumns sets the table columns
func (b *TableConfigBuilder) WithColumns(columns []TableColumn) *TableConfigBuilder {
	b.config.Columns = columns
	return b
}

// WithColumn adds a single column
func (b *TableConfigBuilder) WithColumn(title, field string, width int) *TableConfigBuilder {
	b.config.Columns = append(b.config.Columns, TableColumn{
		Title:     title,
		Field:     field,
		Width:     width,
		Alignment: AlignLeft,
	})
	return b
}

// WithViewportHeight sets the viewport height
func (b *TableConfigBuilder) WithViewportHeight(height int) *TableConfigBuilder {
	b.config.ViewportConfig.Height = height
	return b
}

// WithChunkSize sets the chunk size
func (b *TableConfigBuilder) WithChunkSize(size int) *TableConfigBuilder {
	b.config.ViewportConfig.ChunkSize = size
	return b
}

// WithSelectionMode sets the selection mode
func (b *TableConfigBuilder) WithSelectionMode(mode SelectionMode) *TableConfigBuilder {
	b.config.SelectionMode = mode
	return b
}

// WithHeaderVisible sets header visibility
func (b *TableConfigBuilder) WithHeaderVisible(visible bool) *TableConfigBuilder {
	b.config.ShowHeader = visible
	return b
}

// WithBordersVisible sets border visibility
func (b *TableConfigBuilder) WithBordersVisible(visible bool) *TableConfigBuilder {
	b.config.ShowBorders = visible
	return b
}

// WithAnimationEnabled enables or disables animations
func (b *TableConfigBuilder) WithAnimationEnabled(enabled bool) *TableConfigBuilder {
	b.config.AnimationConfig.Enabled = enabled
	return b
}

// Build returns the configured TableConfig
func (b *TableConfigBuilder) Build() TableConfig {
	return b.config
}

// ================================
// CONFIGURATION UTILITIES
// ================================

// MergeListConfigs merges two list configurations, with override taking precedence
func MergeListConfigs(base, override ListConfig) ListConfig {
	result := base

	// Merge viewport config
	if override.ViewportConfig.Height > 0 {
		result.ViewportConfig.Height = override.ViewportConfig.Height
	}
	if override.ViewportConfig.ChunkSize > 0 {
		result.ViewportConfig.ChunkSize = override.ViewportConfig.ChunkSize
	}
	if override.ViewportConfig.TopThreshold >= 0 {
		result.ViewportConfig.TopThreshold = override.ViewportConfig.TopThreshold
	}
	if override.ViewportConfig.BottomThreshold >= 0 {
		result.ViewportConfig.BottomThreshold = override.ViewportConfig.BottomThreshold
	}
	if override.ViewportConfig.InitialIndex >= 0 {
		result.ViewportConfig.InitialIndex = override.ViewportConfig.InitialIndex
	}

	// Merge other configs
	if override.MaxWidth > 0 {
		result.MaxWidth = override.MaxWidth
	}

	result.SelectionMode = override.SelectionMode
	result.AnimationConfig = override.AnimationConfig
	result.StyleConfig = override.StyleConfig
	result.KeyMap = override.KeyMap

	return result
}

// MergeTableConfigs merges two table configurations, with override taking precedence
func MergeTableConfigs(base, override TableConfig) TableConfig {
	result := base

	// Merge columns if provided
	if len(override.Columns) > 0 {
		result.Columns = override.Columns
	}

	// Merge viewport config
	if override.ViewportConfig.Height > 0 {
		result.ViewportConfig.Height = override.ViewportConfig.Height
	}
	if override.ViewportConfig.ChunkSize > 0 {
		result.ViewportConfig.ChunkSize = override.ViewportConfig.ChunkSize
	}
	if override.ViewportConfig.TopThreshold >= 0 {
		result.ViewportConfig.TopThreshold = override.ViewportConfig.TopThreshold
	}
	if override.ViewportConfig.BottomThreshold >= 0 {
		result.ViewportConfig.BottomThreshold = override.ViewportConfig.BottomThreshold
	}
	if override.ViewportConfig.InitialIndex >= 0 {
		result.ViewportConfig.InitialIndex = override.ViewportConfig.InitialIndex
	}

	// Merge other configs
	result.ShowHeader = override.ShowHeader
	result.ShowBorders = override.ShowBorders
	result.SelectionMode = override.SelectionMode
	result.AnimationConfig = override.AnimationConfig
	result.Theme = override.Theme
	result.KeyMap = override.KeyMap

	return result
}

// CloneListConfig creates a deep copy of a list configuration
func CloneListConfig(config ListConfig) ListConfig {
	// Create a copy with the same values
	return ListConfig{
		ViewportConfig:  config.ViewportConfig,
		StyleConfig:     config.StyleConfig,
		AnimationConfig: config.AnimationConfig,
		SelectionMode:   config.SelectionMode,
		KeyMap:          config.KeyMap,
		MaxWidth:        config.MaxWidth,
	}
}

// CloneTableConfig creates a deep copy of a table configuration
func CloneTableConfig(config TableConfig) TableConfig {
	// Copy columns slice
	columns := make([]TableColumn, len(config.Columns))
	copy(columns, config.Columns)

	return TableConfig{
		Columns:         columns,
		ShowHeader:      config.ShowHeader,
		ShowBorders:     config.ShowBorders,
		ViewportConfig:  config.ViewportConfig,
		Theme:           config.Theme,
		AnimationConfig: config.AnimationConfig,
		SelectionMode:   config.SelectionMode,
		KeyMap:          config.KeyMap,
	}
}
