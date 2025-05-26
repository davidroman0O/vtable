package vtable

import "fmt"

// ViewportConfig defines the configuration for the viewport.
type ViewportConfig struct {
	// Height is the number of rows in the viewport.
	Height int

	// TopThresholdIndex is the index of the top threshold row within the viewport.
	TopThresholdIndex int

	// BottomThresholdIndex is the index of the bottom threshold row within the viewport.
	BottomThresholdIndex int

	// ChunkSize is the number of items to load in a chunk.
	ChunkSize int

	// InitialIndex is the initial selected item index.
	InitialIndex int

	// Debug enables special visual markers for debugging.
	Debug bool
}

// CalculateThresholds calculates reasonable threshold values for a given viewport height.
// This ensures proper spacing and navigation behavior.
func CalculateThresholds(height int) (topThreshold, bottomThreshold int) {
	if height <= 1 {
		return 0, 0 // Special case: single row, both thresholds are the same
	}
	if height == 2 {
		return 0, 1
	}
	if height == 3 {
		return 0, 2
	}
	if height <= 5 {
		return 1, height - 2
	}

	// For larger heights, use proportional spacing
	// Top threshold is about 20% down, bottom threshold is about 20% up from bottom
	topThreshold = height / 5
	if topThreshold < 1 {
		topThreshold = 1
	}

	bottomThreshold = height - 1 - (height / 5)
	if bottomThreshold <= topThreshold {
		bottomThreshold = topThreshold + 1
	}
	if bottomThreshold >= height {
		bottomThreshold = height - 1
	}

	return topThreshold, bottomThreshold
}

// ValidateAndFixViewportConfig validates a viewport config and automatically fixes any issues.
// This prevents configuration errors by auto-correcting invalid values.
func ValidateAndFixViewportConfig(config *ViewportConfig) {
	// Ensure minimum height
	if config.Height <= 0 {
		config.Height = 10 // Reasonable default
	}

	// Auto-calculate chunk size if not set or invalid
	if config.ChunkSize <= 0 {
		config.ChunkSize = config.Height * 2 // Load 2 viewports worth
		if config.ChunkSize < 20 {
			config.ChunkSize = 20 // Minimum reasonable chunk size
		}
	}

	// Ensure initial index is valid
	if config.InitialIndex < 0 {
		config.InitialIndex = 0
	}

	// Auto-calculate thresholds if they're invalid
	if config.TopThresholdIndex < 0 ||
		config.TopThresholdIndex >= config.Height ||
		config.BottomThresholdIndex < 0 ||
		config.BottomThresholdIndex >= config.Height ||
		config.BottomThresholdIndex <= config.TopThresholdIndex {

		config.TopThresholdIndex, config.BottomThresholdIndex = CalculateThresholds(config.Height)
	}
}

// NewViewportConfig creates a viewport config with the specified height and auto-calculated thresholds.
// This is the easiest way to create a viewport config - just specify the height.
func NewViewportConfig(height int) ViewportConfig {
	topThreshold, bottomThreshold := CalculateThresholds(height)

	chunkSize := height * 2
	if chunkSize < 20 {
		chunkSize = 20
	}

	return ViewportConfig{
		Height:               height,
		TopThresholdIndex:    topThreshold,
		BottomThresholdIndex: bottomThreshold,
		ChunkSize:            chunkSize,
		InitialIndex:         0,
		Debug:                false,
	}
}

// DefaultViewportConfig returns a default viewport configuration with sensible values.
func DefaultViewportConfig() ViewportConfig {
	return NewViewportConfig(10)
}

// TableConfig defines the configuration for the table component.
type TableConfig struct {
	// Columns defines the columns in the table
	Columns []TableColumn

	// ShowHeader determines whether the header is displayed
	ShowHeader bool

	// ShowBorders determines whether borders are displayed
	ShowBorders bool

	// ViewportConfig defines the configuration for the viewport
	ViewportConfig ViewportConfig
}

// NewSimpleTableConfig creates a table config with just columns and reasonable defaults.
// This is the easiest way to create a table - just provide the columns.
func NewSimpleTableConfig(columns []TableColumn) TableConfig {
	return TableConfig{
		Columns:        columns,
		ShowHeader:     true,
		ShowBorders:    true,
		ViewportConfig: DefaultViewportConfig(),
	}
}

// NewTableConfig creates a table config with columns and specified viewport height.
// Thresholds are automatically calculated for the given height.
func NewTableConfig(columns []TableColumn, viewportHeight int) TableConfig {
	return TableConfig{
		Columns:        columns,
		ShowHeader:     true,
		ShowBorders:    true,
		ViewportConfig: NewViewportConfig(viewportHeight),
	}
}

// NewTableConfigWithOptions creates a table config with full control over options.
// This provides maximum flexibility while still auto-calculating thresholds.
func NewTableConfigWithOptions(columns []TableColumn, viewportHeight int, showHeader, showBorders bool) TableConfig {
	return TableConfig{
		Columns:        columns,
		ShowHeader:     showHeader,
		ShowBorders:    showBorders,
		ViewportConfig: NewViewportConfig(viewportHeight),
	}
}

// DefaultTableConfig returns a default table configuration with sensible values.
func DefaultTableConfig() TableConfig {
	return TableConfig{
		ShowHeader:     true,
		ShowBorders:    true,
		ViewportConfig: DefaultViewportConfig(),
	}
}

// ValidateAndFixTableConfig validates a table config and automatically fixes any issues.
func ValidateAndFixTableConfig(config *TableConfig) error {
	// Validate columns
	if len(config.Columns) == 0 {
		return fmt.Errorf("table must have at least one column")
	}

	// Auto-fix viewport config
	ValidateAndFixViewportConfig(&config.ViewportConfig)

	return nil
}

// StyleConfig defines styling options for the virtualized components.
type StyleConfig struct {
	// BorderStyle defines the style for the borders (color, etc.)
	BorderStyle string

	// HeaderStyle defines the style for the headers (color, bold, etc.)
	HeaderStyle string

	// RowStyle defines the style for normal rows
	RowStyle string

	// SelectedRowStyle defines the style for the selected row
	SelectedRowStyle string
}

// DefaultStyleConfig returns a default style configuration.
func DefaultStyleConfig() StyleConfig {
	return StyleConfig{
		BorderStyle:      "245",             // Gray
		HeaderStyle:      "bold 252 on 238", // Bold white on dark gray
		RowStyle:         "252",             // Light white
		SelectedRowStyle: "bold 252 on 63",  // Bold white on blue
	}
}
