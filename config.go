package vtable

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

// DefaultViewportConfig returns a default viewport configuration with sensible values.
func DefaultViewportConfig() ViewportConfig {
	return ViewportConfig{
		Height:               10,
		TopThresholdIndex:    2,
		BottomThresholdIndex: 7,
		ChunkSize:            20,
		InitialIndex:         0,
		Debug:                false,
	}
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

// DefaultTableConfig returns a default table configuration with sensible values.
func DefaultTableConfig() TableConfig {
	return TableConfig{
		ShowHeader:     true,
		ShowBorders:    true,
		ViewportConfig: DefaultViewportConfig(),
	}
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
