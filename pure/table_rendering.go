package vtable

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ================================
// COMPONENT-BASED TABLE RENDERING
// ================================
// A flexible rendering system where each part of a table row is a distinct,
// optional component that can be customized and reordered as needed.
// This is specialized for tables - lists and trees have their own systems.

// TableRenderComponent represents a single rendering component for table rows
type TableRenderComponent interface {
	// Render generates the string content for this component
	Render(ctx TableComponentContext) string
	// GetType returns the component type for identification
	GetType() TableComponentType
	// IsEnabled returns whether this component should be rendered
	IsEnabled() bool
	// SetEnabled enables or disables this component
	SetEnabled(enabled bool)
}

// TableComponentType identifies different types of table rendering components
type TableComponentType string

const (
	TableComponentCursor          TableComponentType = "cursor"
	TableComponentPreSpacing      TableComponentType = "pre_spacing"
	TableComponentRowNumber       TableComponentType = "row_number"
	TableComponentSelectionMarker TableComponentType = "selection_marker"
	TableComponentCells           TableComponentType = "cells"
	TableComponentPostSpacing     TableComponentType = "post_spacing"
	TableComponentBackground      TableComponentType = "background"
	TableComponentBorder          TableComponentType = "border"
)

// TableComponentContext provides all the context needed for table component rendering
type TableComponentContext struct {
	// Item data
	Item        Data[any]
	Index       int
	IsCursor    bool
	IsSelected  bool
	IsThreshold bool

	// Table-specific data
	RowData     []string          // Raw cell data
	ColumnCount int               // Number of columns
	ColumnData  []TableColumnData // Column configuration

	// Rendering context
	RenderContext RenderContext

	// Component-specific data (populated by other components during rendering)
	ComponentData map[TableComponentType]string

	// Table-specific configuration
	TableConfig TableRenderConfig
}

// TableColumnData contains information about a specific column
type TableColumnData struct {
	Header    string
	Width     int
	Alignment TableColumnAlignment
	Column    TableColumn // Reference to the actual column
}

// TableRenderConfig contains configuration for component-based table rendering
type TableRenderConfig struct {
	// Component order - defines which components render and in what order
	ComponentOrder []TableComponentType

	// Component configurations
	CursorConfig          TableCursorConfig
	PreSpacingConfig      TableSpacingConfig
	RowNumberConfig       TableRowNumberConfig
	SelectionMarkerConfig TableSelectionMarkerConfig
	CellsConfig           TableCellsConfig
	PostSpacingConfig     TableSpacingConfig
	BackgroundConfig      TableBackgroundConfig
	BorderConfig          TableBorderConfig
}

// Individual component configurations for tables
type TableCursorConfig struct {
	Enabled         bool
	CursorIndicator string
	NormalSpacing   string
	Style           lipgloss.Style
}

type TableSpacingConfig struct {
	Enabled bool
	Spacing string
	Style   lipgloss.Style
}

type TableRowNumberConfig struct {
	Enabled    bool
	ShowHeader bool // Show "Row" in header
	Width      int  // Fixed width for row numbers
	Style      lipgloss.Style
	Alignment  TableColumnAlignment
	Formatter  RowNumberFormatter
}

type TableSelectionMarkerConfig struct {
	Enabled          bool
	SelectedMarker   string
	UnselectedMarker string
	Style            lipgloss.Style
	Width            int
}

type TableCellsConfig struct {
	Enabled        bool
	CellSeparator  string
	CellPadding    int
	Style          lipgloss.Style
	HeaderStyle    lipgloss.Style
	AlternateStyle lipgloss.Style // For alternating row colors
	UseAlternating bool
}

type TableBackgroundConfig struct {
	Enabled           bool
	Style             lipgloss.Style
	ApplyToComponents []TableComponentType
	Mode              TableBackgroundMode
}

type TableBorderConfig struct {
	Enabled     bool
	LeftBorder  string
	RightBorder string
	CellBorder  string
	Style       lipgloss.Style
}

type TableColumnAlignment int

const (
	TableAlignLeft TableColumnAlignment = iota
	TableAlignCenter
	TableAlignRight
)

type TableBackgroundMode int

const (
	TableBackgroundEntireRow TableBackgroundMode = iota
	TableBackgroundSelectiveComponents
	TableBackgroundCellsOnly
	TableBackgroundIndicatorOnly
)

// RowNumberFormatter formats row numbers
type RowNumberFormatter func(index int, isHeader bool) string

// ================================
// COMPONENT IMPLEMENTATIONS
// ================================

// TableCursorComponent handles cursor indicator rendering for tables
type TableCursorComponent struct {
	config TableCursorConfig
}

func NewTableCursorComponent(config TableCursorConfig) *TableCursorComponent {
	return &TableCursorComponent{config: config}
}

func (c *TableCursorComponent) Render(ctx TableComponentContext) string {
	if ctx.IsCursor {
		return c.config.Style.Render(c.config.CursorIndicator)
	}
	return c.config.Style.Render(c.config.NormalSpacing)
}

func (c *TableCursorComponent) GetType() TableComponentType {
	return TableComponentCursor
}

func (c *TableCursorComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *TableCursorComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TablePreSpacingComponent handles spacing before table content
type TablePreSpacingComponent struct {
	config TableSpacingConfig
}

func NewTablePreSpacingComponent(config TableSpacingConfig) *TablePreSpacingComponent {
	return &TablePreSpacingComponent{config: config}
}

func (c *TablePreSpacingComponent) Render(ctx TableComponentContext) string {
	return c.config.Style.Render(c.config.Spacing)
}

func (c *TablePreSpacingComponent) GetType() TableComponentType {
	return TableComponentPreSpacing
}

func (c *TablePreSpacingComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *TablePreSpacingComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TableRowNumberComponent handles row numbering
type TableRowNumberComponent struct {
	config TableRowNumberConfig
}

func NewTableRowNumberComponent(config TableRowNumberConfig) *TableRowNumberComponent {
	return &TableRowNumberComponent{config: config}
}

func (c *TableRowNumberComponent) Render(ctx TableComponentContext) string {
	var content string

	if c.config.Formatter != nil {
		content = c.config.Formatter(ctx.Index, false)
	} else {
		content = fmt.Sprintf("%d", ctx.Index+1)
	}

	// Apply width and alignment
	if c.config.Width > 0 {
		switch c.config.Alignment {
		case TableAlignRight:
			if len(content) < c.config.Width {
				content = strings.Repeat(" ", c.config.Width-len(content)) + content
			}
		case TableAlignCenter:
			if len(content) < c.config.Width {
				padding := c.config.Width - len(content)
				leftPad := padding / 2
				rightPad := padding - leftPad
				content = strings.Repeat(" ", leftPad) + content + strings.Repeat(" ", rightPad)
			}
		case TableAlignLeft:
			if len(content) < c.config.Width {
				content = content + strings.Repeat(" ", c.config.Width-len(content))
			}
		}
	}

	return c.config.Style.Render(content)
}

func (c *TableRowNumberComponent) GetType() TableComponentType {
	return TableComponentRowNumber
}

func (c *TableRowNumberComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *TableRowNumberComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TableSelectionMarkerComponent handles selection indicators
type TableSelectionMarkerComponent struct {
	config TableSelectionMarkerConfig
}

func NewTableSelectionMarkerComponent(config TableSelectionMarkerConfig) *TableSelectionMarkerComponent {
	return &TableSelectionMarkerComponent{config: config}
}

func (c *TableSelectionMarkerComponent) Render(ctx TableComponentContext) string {
	var marker string
	if ctx.IsSelected {
		marker = c.config.SelectedMarker
	} else {
		marker = c.config.UnselectedMarker
	}

	// Apply width
	if c.config.Width > 0 && len(marker) < c.config.Width {
		marker = marker + strings.Repeat(" ", c.config.Width-len(marker))
	}

	return c.config.Style.Render(marker)
}

func (c *TableSelectionMarkerComponent) GetType() TableComponentType {
	return TableComponentSelectionMarker
}

func (c *TableSelectionMarkerComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *TableSelectionMarkerComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TableCellsComponent handles the main table cell content
type TableCellsComponent struct {
	config TableCellsConfig
}

func NewTableCellsComponent(config TableCellsConfig) *TableCellsComponent {
	return &TableCellsComponent{config: config}
}

func (c *TableCellsComponent) Render(ctx TableComponentContext) string {
	if len(ctx.RowData) == 0 {
		return ""
	}

	var cells []string

	for i, cellData := range ctx.RowData {
		var content string

		// For component-based rendering, we don't use formatters here
		// Formatters are handled at a higher level in the actual table implementation
		content = cellData

		// Apply column width and alignment if configured
		if i < len(ctx.ColumnData) {
			colData := ctx.ColumnData[i]
			if colData.Width > 0 {
				switch colData.Alignment {
				case TableAlignRight:
					if len(content) < colData.Width {
						content = strings.Repeat(" ", colData.Width-len(content)) + content
					}
				case TableAlignCenter:
					if len(content) < colData.Width {
						padding := colData.Width - len(content)
						leftPad := padding / 2
						rightPad := padding - leftPad
						content = strings.Repeat(" ", leftPad) + content + strings.Repeat(" ", rightPad)
					}
				case TableAlignLeft:
					if len(content) < colData.Width {
						content = content + strings.Repeat(" ", colData.Width-len(content))
					}
				}
			}
		}

		// Apply cell padding
		if c.config.CellPadding > 0 {
			padding := strings.Repeat(" ", c.config.CellPadding)
			content = padding + content + padding
		}

		cells = append(cells, content)
	}

	// Join cells with separator
	result := strings.Join(cells, c.config.CellSeparator)

	// Apply alternating row style if enabled
	if c.config.UseAlternating && ctx.Index%2 == 1 {
		return c.config.AlternateStyle.Render(result)
	}

	return c.config.Style.Render(result)
}

func (c *TableCellsComponent) GetType() TableComponentType {
	return TableComponentCells
}

func (c *TableCellsComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *TableCellsComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TablePostSpacingComponent handles spacing after table content
type TablePostSpacingComponent struct {
	config TableSpacingConfig
}

func NewTablePostSpacingComponent(config TableSpacingConfig) *TablePostSpacingComponent {
	return &TablePostSpacingComponent{config: config}
}

func (c *TablePostSpacingComponent) Render(ctx TableComponentContext) string {
	return c.config.Style.Render(c.config.Spacing)
}

func (c *TablePostSpacingComponent) GetType() TableComponentType {
	return TableComponentPostSpacing
}

func (c *TablePostSpacingComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *TablePostSpacingComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TableBorderComponent handles table borders
type TableBorderComponent struct {
	config TableBorderConfig
}

func NewTableBorderComponent(config TableBorderConfig) *TableBorderComponent {
	return &TableBorderComponent{config: config}
}

func (c *TableBorderComponent) Render(ctx TableComponentContext) string {
	// Get the content from other components
	var parts []string
	for _, compType := range ctx.TableConfig.ComponentOrder {
		if compType == TableComponentBorder || compType == TableComponentBackground {
			continue
		}
		if content, exists := ctx.ComponentData[compType]; exists && content != "" {
			parts = append(parts, content)
		}
	}

	content := strings.Join(parts, "")

	// Add borders
	result := c.config.LeftBorder + content + c.config.RightBorder

	return c.config.Style.Render(result)
}

func (c *TableBorderComponent) GetType() TableComponentType {
	return TableComponentBorder
}

func (c *TableBorderComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *TableBorderComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TableBackgroundComponent handles background styling for table rows
type TableBackgroundComponent struct {
	config TableBackgroundConfig
}

func NewTableBackgroundComponent(config TableBackgroundConfig) *TableBackgroundComponent {
	return &TableBackgroundComponent{config: config}
}

func (c *TableBackgroundComponent) Render(ctx TableComponentContext) string {
	switch c.config.Mode {
	case TableBackgroundEntireRow:
		// Apply background to the entire combined content
		var parts []string
		for _, compType := range ctx.TableConfig.ComponentOrder {
			if compType == TableComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				parts = append(parts, content)
			}
		}
		fullContent := strings.Join(parts, "")
		return c.config.Style.Render(fullContent)

	case TableBackgroundSelectiveComponents:
		// Apply background only to specified components
		var result strings.Builder
		for _, compType := range ctx.TableConfig.ComponentOrder {
			if compType == TableComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				shouldApplyBackground := false
				for _, applyToComp := range c.config.ApplyToComponents {
					if compType == applyToComp {
						shouldApplyBackground = true
						break
					}
				}

				if shouldApplyBackground {
					result.WriteString(c.config.Style.Render(content))
				} else {
					result.WriteString(content)
				}
			}
		}
		return result.String()

	case TableBackgroundCellsOnly:
		// Apply background only to cells
		var result strings.Builder
		for _, compType := range ctx.TableConfig.ComponentOrder {
			if compType == TableComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				if compType == TableComponentCells {
					result.WriteString(c.config.Style.Render(content))
				} else {
					result.WriteString(content)
				}
			}
		}
		return result.String()

	case TableBackgroundIndicatorOnly:
		// Apply background only to cursor indicator
		var result strings.Builder
		for _, compType := range ctx.TableConfig.ComponentOrder {
			if compType == TableComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				if compType == TableComponentCursor {
					result.WriteString(c.config.Style.Render(content))
				} else {
					result.WriteString(content)
				}
			}
		}
		return result.String()

	default:
		// Fallback to entire row
		var parts []string
		for _, compType := range ctx.TableConfig.ComponentOrder {
			if compType == TableComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				parts = append(parts, content)
			}
		}
		fullContent := strings.Join(parts, "")
		return c.config.Style.Render(fullContent)
	}
}

func (c *TableBackgroundComponent) GetType() TableComponentType {
	return TableComponentBackground
}

func (c *TableBackgroundComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *TableBackgroundComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// ================================
// TABLE COMPONENT RENDERER
// ================================

// TableComponentRenderer orchestrates the rendering of all table components
type TableComponentRenderer struct {
	components map[TableComponentType]TableRenderComponent
	config     TableRenderConfig
}

func NewTableComponentRenderer(config TableRenderConfig) *TableComponentRenderer {
	renderer := &TableComponentRenderer{
		components: make(map[TableComponentType]TableRenderComponent),
		config:     config,
	}

	// Create components based on config
	renderer.components[TableComponentCursor] = NewTableCursorComponent(config.CursorConfig)
	renderer.components[TableComponentPreSpacing] = NewTablePreSpacingComponent(config.PreSpacingConfig)
	renderer.components[TableComponentRowNumber] = NewTableRowNumberComponent(config.RowNumberConfig)
	renderer.components[TableComponentSelectionMarker] = NewTableSelectionMarkerComponent(config.SelectionMarkerConfig)
	renderer.components[TableComponentCells] = NewTableCellsComponent(config.CellsConfig)
	renderer.components[TableComponentPostSpacing] = NewTablePostSpacingComponent(config.PostSpacingConfig)
	renderer.components[TableComponentBorder] = NewTableBorderComponent(config.BorderConfig)
	renderer.components[TableComponentBackground] = NewTableBackgroundComponent(config.BackgroundConfig)

	return renderer
}

// SetComponentOrder sets the rendering order of components
func (r *TableComponentRenderer) SetComponentOrder(order []TableComponentType) {
	r.config.ComponentOrder = order
}

// EnableComponent enables a specific component
func (r *TableComponentRenderer) EnableComponent(componentType TableComponentType) {
	if comp, exists := r.components[componentType]; exists {
		comp.SetEnabled(true)
	}
}

// DisableComponent disables a specific component
func (r *TableComponentRenderer) DisableComponent(componentType TableComponentType) {
	if comp, exists := r.components[componentType]; exists {
		comp.SetEnabled(false)
	}
}

// GetComponent returns a component by type
func (r *TableComponentRenderer) GetComponent(componentType TableComponentType) TableRenderComponent {
	return r.components[componentType]
}

// UpdateConfig updates the renderer configuration
func (r *TableComponentRenderer) UpdateConfig(config TableRenderConfig) {
	r.config = config
	// Recreate components with new config
	r.components[TableComponentCursor] = NewTableCursorComponent(config.CursorConfig)
	r.components[TableComponentPreSpacing] = NewTablePreSpacingComponent(config.PreSpacingConfig)
	r.components[TableComponentRowNumber] = NewTableRowNumberComponent(config.RowNumberConfig)
	r.components[TableComponentSelectionMarker] = NewTableSelectionMarkerComponent(config.SelectionMarkerConfig)
	r.components[TableComponentCells] = NewTableCellsComponent(config.CellsConfig)
	r.components[TableComponentPostSpacing] = NewTablePostSpacingComponent(config.PostSpacingConfig)
	r.components[TableComponentBorder] = NewTableBorderComponent(config.BorderConfig)
	r.components[TableComponentBackground] = NewTableBackgroundComponent(config.BackgroundConfig)
}

// Render renders all components in the specified order
func (r *TableComponentRenderer) Render(
	item Data[any],
	index int,
	rowData []string,
	columnData []TableColumnData,
	renderContext RenderContext,
	isCursor, isTopThreshold, isBottomThreshold bool,
) string {
	ctx := TableComponentContext{
		Item:          item,
		Index:         index,
		IsCursor:      isCursor,
		IsSelected:    item.Selected,
		IsThreshold:   isTopThreshold || isBottomThreshold,
		RowData:       rowData,
		ColumnCount:   len(columnData),
		ColumnData:    columnData,
		RenderContext: renderContext,
		ComponentData: make(map[TableComponentType]string),
		TableConfig:   r.config,
	}

	// First pass: render all non-background/border components
	for _, compType := range r.config.ComponentOrder {
		if compType == TableComponentBackground || compType == TableComponentBorder {
			continue // Handle these separately
		}

		if comp, exists := r.components[compType]; exists && comp.IsEnabled() {
			content := comp.Render(ctx)
			ctx.ComponentData[compType] = content
		}
	}

	// Check if we have a border component and apply it if enabled
	borderComp := r.components[TableComponentBorder]
	if borderComp != nil && borderComp.IsEnabled() {
		return borderComp.Render(ctx)
	}

	// Check if we have a background component and apply it if enabled
	backgroundComp := r.components[TableComponentBackground]
	if backgroundComp != nil && backgroundComp.IsEnabled() {
		// Apply background styling - but only for cursor items to maintain expected behavior
		if isCursor {
			return backgroundComp.Render(ctx)
		}
	}

	// No border or background styling - combine components normally
	var result strings.Builder
	for _, compType := range r.config.ComponentOrder {
		if compType == TableComponentBackground || compType == TableComponentBorder {
			continue
		}
		if content, exists := ctx.ComponentData[compType]; exists {
			result.WriteString(content)
		}
	}

	return result.String()
}

// ================================
// PRESET CONFIGURATIONS
// ================================

// DefaultTableRenderConfig returns sensible defaults for component-based table rendering
func DefaultTableRenderConfig() TableRenderConfig {
	return TableRenderConfig{
		ComponentOrder: []TableComponentType{
			TableComponentCursor,
			TableComponentSelectionMarker,
			TableComponentCells,
		},
		CursorConfig: TableCursorConfig{
			Enabled:         true,
			CursorIndicator: "► ",
			NormalSpacing:   "  ",
			Style:           lipgloss.NewStyle(),
		},
		PreSpacingConfig: TableSpacingConfig{
			Enabled: false,
			Spacing: "",
			Style:   lipgloss.NewStyle(),
		},
		RowNumberConfig: TableRowNumberConfig{
			Enabled:    false,
			ShowHeader: true,
			Width:      4,
			Style:      lipgloss.NewStyle(),
			Alignment:  TableAlignRight,
			Formatter:  nil,
		},
		SelectionMarkerConfig: TableSelectionMarkerConfig{
			Enabled:          true,
			SelectedMarker:   "[✓]",
			UnselectedMarker: "[ ]",
			Style:            lipgloss.NewStyle(),
			Width:            4,
		},
		CellsConfig: TableCellsConfig{
			Enabled:        true,
			CellSeparator:  " │ ",
			CellPadding:    1,
			Style:          lipgloss.NewStyle(),
			HeaderStyle:    lipgloss.NewStyle().Bold(true),
			AlternateStyle: lipgloss.NewStyle().Background(lipgloss.Color("240")),
			UseAlternating: false,
		},
		PostSpacingConfig: TableSpacingConfig{
			Enabled: false,
			Spacing: "",
			Style:   lipgloss.NewStyle(),
		},
		BackgroundConfig: TableBackgroundConfig{
			Enabled:           false,
			Style:             lipgloss.NewStyle(),
			ApplyToComponents: []TableComponentType{TableComponentCursor, TableComponentCells},
			Mode:              TableBackgroundEntireRow,
		},
		BorderConfig: TableBorderConfig{
			Enabled:     false,
			LeftBorder:  "│ ",
			RightBorder: " │",
			CellBorder:  " │ ",
			Style:       lipgloss.NewStyle(),
		},
	}
}

// NumberedTableConfig creates a config for tables with row numbers
func NumberedTableConfig() TableRenderConfig {
	config := DefaultTableRenderConfig()
	config.ComponentOrder = []TableComponentType{
		TableComponentCursor,
		TableComponentRowNumber,
		TableComponentSelectionMarker,
		TableComponentCells,
	}
	config.RowNumberConfig.Enabled = true
	return config
}

// MinimalTableConfig creates a config for minimal tables
func MinimalTableConfig() TableRenderConfig {
	config := DefaultTableRenderConfig()
	config.ComponentOrder = []TableComponentType{
		TableComponentCells,
	}
	config.CursorConfig.Enabled = false
	config.SelectionMarkerConfig.Enabled = false
	return config
}

// BorderedTableConfig creates a config for tables with borders
func BorderedTableConfig() TableRenderConfig {
	config := DefaultTableRenderConfig()
	config.ComponentOrder = []TableComponentType{
		TableComponentBorder,
	}
	config.BorderConfig.Enabled = true
	return config
}

// AlternatingTableConfig creates a config for tables with alternating row colors
func AlternatingTableConfig() TableRenderConfig {
	config := DefaultTableRenderConfig()
	config.CellsConfig.UseAlternating = true
	return config
}

// BackgroundStyledTableConfig creates a config with background styling
func BackgroundStyledTableConfig(style lipgloss.Style, mode TableBackgroundMode) TableRenderConfig {
	config := DefaultTableRenderConfig()
	config.BackgroundConfig.Enabled = true
	config.BackgroundConfig.Style = style
	config.BackgroundConfig.Mode = mode
	return config
}

// ================================
// INTEGRATION WITH EXISTING SYSTEM
// ================================

// ComponentBasedTableFormatter creates a CellFormatter that uses the component system
func ComponentBasedTableFormatter(config TableRenderConfig) func(Data[any], int, []string, []TableColumnData, RenderContext, bool, bool, bool) string {
	renderer := NewTableComponentRenderer(config)
	return func(
		item Data[any],
		index int,
		rowData []string,
		columnData []TableColumnData,
		ctx RenderContext,
		isCursor, isTopThreshold, isBottomThreshold bool,
	) string {
		return renderer.Render(item, index, rowData, columnData, ctx, isCursor, isTopThreshold, isBottomThreshold)
	}
}

// EnhancedTableFormatter creates a table formatter using the component system
// This is the main entry point for table rendering
func EnhancedTableFormatter(config TableRenderConfig) func(Data[any], int, []string, []TableColumnData, RenderContext, bool, bool, bool) string {
	return ComponentBasedTableFormatter(config)
}

// ================================
// UTILITY FUNCTIONS
// ================================

// DefaultRowNumberFormatter provides default row number formatting
func DefaultRowNumberFormatter(index int, isHeader bool) string {
	if isHeader {
		return "Row"
	}
	return fmt.Sprintf("%d", index+1)
}

// AlphabeticalRowNumberFormatter provides alphabetical row numbering
func AlphabeticalRowNumberFormatter(index int, isHeader bool) string {
	if isHeader {
		return "Row"
	}
	if index < 26 {
		return string(rune('A' + index))
	}
	return fmt.Sprintf("%d", index+1)
}

// RomanRowNumberFormatter provides roman numeral row numbering (up to 20)
func RomanRowNumberFormatter(index int, isHeader bool) string {
	if isHeader {
		return "Row"
	}
	romans := []string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X",
		"XI", "XII", "XIII", "XIV", "XV", "XVI", "XVII", "XVIII", "XIX", "XX"}
	if index < len(romans) {
		return romans[index]
	}
	return fmt.Sprintf("%d", index+1)
}
