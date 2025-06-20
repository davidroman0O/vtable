// Package config provides default configurations and builders for vtable
// components. It centralizes the initial setup for Lists (list and tree list) and Tables, making it
// easier to create new components with sensible defaults. This package also
// includes validation and fixing logic to ensure that configurations are always
// in a valid state.
package config

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/core"
)

// DefaultListConfig returns a sensible default configuration for a List component.
// It initializes the list with a standard viewport, styles, selection mode, and keymap.
func DefaultListConfig() core.ListConfig {
	return core.ListConfig{
		ViewportConfig: DefaultViewportConfig(),
		StyleConfig:    DefaultStyleConfig(),
		RenderConfig:   DefaultListRenderConfig(), // Will be set by list package
		SelectionMode:  core.SelectionSingle,
		KeyMap:         core.DefaultNavigationKeyMap(),
		MaxWidth:       80,
	}
}

// DefaultListRenderConfig returns the default configuration for the
// component-based list rendering pipeline. It specifies the default order of
// components (cursor, enumerator, content) and their individual default settings.
func DefaultListRenderConfig() core.ListRenderConfig {
	return core.ListRenderConfig{
		ComponentOrder: []core.ListComponentType{
			core.ListComponentCursor,
			core.ListComponentEnumerator,
			core.ListComponentContent,
		},
		CursorConfig: core.ListCursorConfig{
			Enabled:         true,
			CursorIndicator: "► ",
			NormalSpacing:   "  ",
			Style:           lipgloss.NewStyle(),
			Alignment:       core.ListAlignmentNone,
			MaxWidth:        0,
		},
		PreSpacingConfig: core.ListSpacingConfig{
			Enabled: false,
			Spacing: "",
			Style:   lipgloss.NewStyle(),
		},
		EnumeratorConfig: core.ListEnumeratorConfig{
			Enabled:    true,
			Enumerator: func(item core.Data[any], index int, ctx core.RenderContext) string { return "" },
			Style:      lipgloss.NewStyle(),
			Alignment:  core.ListAlignmentNone,
			MaxWidth:   0,
		},
		ContentConfig: core.ListContentConfig{
			Enabled:   true,
			Formatter: nil,
			Style:     lipgloss.NewStyle(),
			WrapText:  false,
			MaxWidth:  80,
		},
		PostSpacingConfig: core.ListSpacingConfig{
			Enabled: false,
			Spacing: "",
			Style:   lipgloss.NewStyle(),
		},
		BackgroundConfig: core.ListBackgroundConfig{
			Enabled:           false,
			Style:             lipgloss.NewStyle(),
			ApplyToComponents: []core.ListComponentType{core.ListComponentCursor, core.ListComponentEnumerator, core.ListComponentContent},
			Mode:              core.ListBackgroundEntireLine,
		},
	}
}

// DefaultTableConfig returns a sensible default configuration for a Table component.
// It sets up a table with visible headers and borders, a default theme, and
// standard viewport settings.
func DefaultTableConfig() core.TableConfig {
	return core.TableConfig{
		Columns:                 []core.TableColumn{},
		ShowHeader:              true,
		ShowBorders:             true,
		FullRowHighlighting:     true,
		ShowTopBorder:           true,  // Default to enabled when borders are on
		ShowBottomBorder:        true,  // Default to enabled when borders are on
		ShowHeaderSeparator:     true,  // Default to enabled when borders are on
		RemoveTopBorderSpace:    false, // Default to preserving space
		RemoveBottomBorderSpace: false, // Default to preserving space
		ViewportConfig:          DefaultViewportConfig(),
		Theme:                   DefaultTheme(),
		// TODO: animation system is not implemented yet
		// AnimationConfig:         core.DefaultAnimationConfig(),
		SelectionMode: core.SelectionSingle,
		KeyMap:        core.DefaultNavigationKeyMap(),
	}
}

// DefaultViewportConfig returns a sensible default viewport configuration,
// suitable for both lists and tables.
func DefaultViewportConfig() core.ViewportConfig {
	return core.ViewportConfig{
		Height:          10,
		TopThreshold:    2, // 2 positions from viewport start
		BottomThreshold: 2, // 2 positions from viewport end
		ChunkSize:       100,
		InitialIndex:    0,
	}
}

// DefaultStyleConfig returns the default style configuration for a List component,
// defining colors and attributes for different item states.
func DefaultStyleConfig() core.StyleConfig {
	return core.StyleConfig{
		CursorStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true),
		SelectedStyle:  lipgloss.NewStyle().Background(lipgloss.Color("57")).Foreground(lipgloss.Color("230")),
		DefaultStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		ThresholdStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true),
		DisabledStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
		LoadingStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Italic(true),
		ErrorStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
	}
}

// DefaultTheme returns the default theme for a Table component, defining styles
// for headers, cells, borders, and various states.
func DefaultTheme() core.Theme {
	return core.Theme{
		HeaderStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("99")).Bold(true),
		CellStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		CursorStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true),
		SelectedStyle:      lipgloss.NewStyle().Background(lipgloss.Color("57")).Foreground(lipgloss.Color("230")),
		FullRowCursorStyle: lipgloss.NewStyle().Background(lipgloss.Color("12")).Foreground(lipgloss.Color("15")).Bold(true),
		BorderChars:        core.DefaultBorderChars(),
		BorderColor:        "241",
		HeaderColor:        "99",
		AlternateRowStyle:  lipgloss.NewStyle().Background(lipgloss.Color("235")),
		DisabledStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("243")),
		LoadingStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Italic(true),
		ErrorStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
	}
}

// ValidateViewportConfig checks a ViewportConfig for valid values and returns a
// slice of errors if any issues are found.
func ValidateViewportConfig(config *core.ViewportConfig) []error {
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

// ValidateTableConfig checks a TableConfig for valid values and returns a slice
// of errors. It validates the viewport, columns, and animation settings.
func ValidateTableConfig(config *core.TableConfig) []error {
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
	// TODO: animation system is not implemented yet
	// errors = append(errors, ValidateAnimationConfig(&config.AnimationConfig)...)

	return errors
}

// ValidateListConfig checks a ListConfig for valid values and returns a slice
// of errors. It validates the viewport and display settings.
func ValidateListConfig(config *core.ListConfig) []error {
	var errors []error

	// Validate viewport config
	errors = append(errors, ValidateViewportConfig(&config.ViewportConfig)...)

	// Validate max width
	if config.MaxWidth <= 0 {
		errors = append(errors, fmt.Errorf("max width must be positive, got %d", config.MaxWidth))
	}

	return errors
}

// ValidateAnimationConfig checks an AnimationConfig for valid values and returns
// a slice of errors.
func ValidateAnimationConfig(config *core.AnimationConfig) []error {
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

// FixViewportConfig corrects common issues in a ViewportConfig, such as
// negative heights or out-of-bounds thresholds, resetting them to valid defaults.
func FixViewportConfig(config *core.ViewportConfig) {
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

// FixTableConfig corrects common issues in a TableConfig by fixing its
// child configurations (viewport, columns, animation).
func FixTableConfig(config *core.TableConfig) {
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
	// TODO: animation system is not implemented yet
	// FixAnimationConfig(&config.AnimationConfig)
}

// FixListConfig corrects common issues in a ListConfig by fixing its
// child configurations (viewport) and validating its own properties.
func FixListConfig(config *core.ListConfig) {
	// Fix viewport config
	FixViewportConfig(&config.ViewportConfig)

	// Fix max width
	if config.MaxWidth <= 0 {
		config.MaxWidth = 80
	}
}

// FixAnimationConfig corrects common issues in an AnimationConfig, ensuring
// sensible values for performance and stability.
func FixAnimationConfig(config *core.AnimationConfig) {
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

// ListConfigBuilder provides a fluent API for constructing a core.ListConfig.
type ListConfigBuilder struct {
	config core.ListConfig
}

// NewListConfigBuilder creates a new ListConfigBuilder, initialized with the
// default list configuration.
func NewListConfigBuilder() *ListConfigBuilder {
	return &ListConfigBuilder{
		config: DefaultListConfig(),
	}
}

// WithViewportHeight sets the viewport height in the configuration.
func (b *ListConfigBuilder) WithViewportHeight(height int) *ListConfigBuilder {
	b.config.ViewportConfig.Height = height
	return b
}

// WithChunkSize sets the data chunk size in the configuration.
func (b *ListConfigBuilder) WithChunkSize(size int) *ListConfigBuilder {
	b.config.ViewportConfig.ChunkSize = size
	return b
}

// WithSelectionMode sets the selection mode in the configuration.
func (b *ListConfigBuilder) WithSelectionMode(mode core.SelectionMode) *ListConfigBuilder {
	b.config.SelectionMode = mode
	return b
}

// WithMaxWidth sets the maximum width of the list in the configuration.
func (b *ListConfigBuilder) WithMaxWidth(width int) *ListConfigBuilder {
	b.config.MaxWidth = width
	return b
}

// Build returns the final, constructed core.ListConfig.
func (b *ListConfigBuilder) Build() core.ListConfig {
	return b.config
}

// TableConfigBuilder provides a fluent API for constructing a core.TableConfig.
type TableConfigBuilder struct {
	config core.TableConfig
}

// NewTableConfigBuilder creates a new TableConfigBuilder, initialized with the
// default table configuration.
func NewTableConfigBuilder() *TableConfigBuilder {
	return &TableConfigBuilder{
		config: DefaultTableConfig(),
	}
}

// WithColumns sets the table columns in the configuration.
func (b *TableConfigBuilder) WithColumns(columns []core.TableColumn) *TableConfigBuilder {
	b.config.Columns = columns
	return b
}

// WithColumn adds a single column to the configuration.
func (b *TableConfigBuilder) WithColumn(title, field string, width int) *TableConfigBuilder {
	b.config.Columns = append(b.config.Columns, core.TableColumn{
		Title:     title,
		Field:     field,
		Width:     width,
		Alignment: core.AlignLeft,
	})
	return b
}

// WithViewportHeight sets the viewport height in the configuration.
func (b *TableConfigBuilder) WithViewportHeight(height int) *TableConfigBuilder {
	b.config.ViewportConfig.Height = height
	return b
}

// WithChunkSize sets the data chunk size in the configuration.
func (b *TableConfigBuilder) WithChunkSize(size int) *TableConfigBuilder {
	b.config.ViewportConfig.ChunkSize = size
	return b
}

// WithSelectionMode sets the selection mode in the configuration.
func (b *TableConfigBuilder) WithSelectionMode(mode core.SelectionMode) *TableConfigBuilder {
	b.config.SelectionMode = mode
	return b
}

// WithHeaderVisible sets the header visibility in the configuration.
func (b *TableConfigBuilder) WithHeaderVisible(visible bool) *TableConfigBuilder {
	b.config.ShowHeader = visible
	return b
}

// WithBordersVisible sets the border visibility in the configuration.
func (b *TableConfigBuilder) WithBordersVisible(visible bool) *TableConfigBuilder {
	b.config.ShowBorders = visible
	return b
}

// WithTopBorderVisible sets the top border visibility in the configuration.
func (b *TableConfigBuilder) WithTopBorderVisible(visible bool) *TableConfigBuilder {
	b.config.ShowTopBorder = visible
	return b
}

// WithBottomBorderVisible sets the bottom border visibility in the configuration.
func (b *TableConfigBuilder) WithBottomBorderVisible(visible bool) *TableConfigBuilder {
	b.config.ShowBottomBorder = visible
	return b
}

// WithHeaderSeparatorVisible sets the header separator visibility in the configuration.
func (b *TableConfigBuilder) WithHeaderSeparatorVisible(visible bool) *TableConfigBuilder {
	b.config.ShowHeaderSeparator = visible
	return b
}

// WithTopBorderSpaceRemoved controls whether top border space is completely removed.
func (b *TableConfigBuilder) WithTopBorderSpaceRemoved(removed bool) *TableConfigBuilder {
	b.config.RemoveTopBorderSpace = removed
	return b
}

// WithBottomBorderSpaceRemoved controls whether bottom border space is completely removed.
func (b *TableConfigBuilder) WithBottomBorderSpaceRemoved(removed bool) *TableConfigBuilder {
	b.config.RemoveBottomBorderSpace = removed
	return b
}

// Build returns the final, constructed core.TableConfig.
func (b *TableConfigBuilder) Build() core.TableConfig {
	return b.config
}

// MergeListConfigs merges two list configurations. Values from the `override`
// config take precedence over the `base` config.
func MergeListConfigs(base, override core.ListConfig) core.ListConfig {
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
	result.StyleConfig = override.StyleConfig
	result.KeyMap = override.KeyMap

	return result
}

// MergeTableConfigs merges two table configurations. Values from the `override`
// config take precedence over the `base` config.
func MergeTableConfigs(base, override core.TableConfig) core.TableConfig {
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
	// TODO: animation system is not implemented yet
	// result.AnimationConfig = override.AnimationConfig
	result.Theme = override.Theme
	result.KeyMap = override.KeyMap

	return result
}

// CloneListConfig creates a deep copy of a list configuration.
func CloneListConfig(config core.ListConfig) core.ListConfig {
	// Create a copy with the same values
	return core.ListConfig{
		ViewportConfig: config.ViewportConfig,
		StyleConfig:    config.StyleConfig,
		RenderConfig:   config.RenderConfig,
		SelectionMode:  config.SelectionMode,
		KeyMap:         config.KeyMap,
		MaxWidth:       config.MaxWidth,
	}
}

// CloneTableConfig creates a deep copy of a table configuration.
func CloneTableConfig(config core.TableConfig) core.TableConfig {
	// Copy columns slice
	columns := make([]core.TableColumn, len(config.Columns))
	copy(columns, config.Columns)

	return core.TableConfig{
		Columns:        columns,
		ShowHeader:     config.ShowHeader,
		ShowBorders:    config.ShowBorders,
		ViewportConfig: config.ViewportConfig,
		Theme:          config.Theme,
		// TODO: animation system is not implemented yet
		// AnimationConfig: config.AnimationConfig,
		SelectionMode: config.SelectionMode,
		KeyMap:        config.KeyMap,
	}
}
