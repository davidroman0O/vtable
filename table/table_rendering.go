// This file implements a flexible, component-based rendering system specialized
// for the Table component. Each part of a table row's visual representation
// (e.g., cursor, selection marker, cells) is a distinct, optional component
// that can be customized and reordered. This allows for extensive customization
// of how table rows are displayed, separate from the rendering pipelines for the
// List and Tree components.
package table

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/core"
)

// TableRenderComponent defines the contract for a single, modular part of a
// rendered table row. Implementations of this interface are responsible for
// generating a specific piece of the final string output, such as the cursor
// indicator or the main cell content.
type TableRenderComponent interface {
	// Render generates the string content for this component based on the
	// provided context.
	Render(ctx TableComponentContext) string
	// GetType returns the component's unique type identifier, used for ordering
	// and configuration.
	GetType() TableComponentType
	// IsEnabled returns whether this component should be rendered.
	IsEnabled() bool
	// SetEnabled allows enabling or disabling this component at runtime.
	SetEnabled(enabled bool)
}

// TableComponentType is a string identifier for a specific type of table
// rendering component. It is used to define the rendering order and to configure
// individual components.
type TableComponentType string

// Constants defining the available component types for the table.
const (
	// TableComponentCursor handles the rendering of the cursor indicator.
	TableComponentCursor TableComponentType = "cursor"
	// TableComponentPreSpacing adds spacing before the main content.
	TableComponentPreSpacing TableComponentType = "pre_spacing"
	// TableComponentRowNumber renders the row number.
	TableComponentRowNumber TableComponentType = "row_number"
	// TableComponentSelectionMarker renders a marker for selected rows.
	TableComponentSelectionMarker TableComponentType = "selection_marker"
	// TableComponentCells renders the main content of the table cells.
	TableComponentCells TableComponentType = "cells"
	// TableComponentPostSpacing adds spacing after the main content.
	TableComponentPostSpacing TableComponentType = "post_spacing"
	// TableComponentBackground applies background styling as a final step.
	TableComponentBackground TableComponentType = "background"
	// TableComponentBorder renders borders around the row.
	TableComponentBorder TableComponentType = "border"
)

// TableComponentContext provides all the necessary data for a TableRenderComponent
// to render its output. It encapsulates item-specific data, table structure
// information, and global rendering settings.
type TableComponentContext struct {
	// Item is the core data for the row being rendered.
	Item core.Data[any]
	// Index is the absolute linear index of the item in the dataset.
	Index int
	// IsCursor is true if this item is currently under the cursor.
	IsCursor bool
	// IsSelected is true if this item is currently selected.
	IsSelected bool
	// IsThreshold is true if this item is at a scroll threshold.
	IsThreshold bool

	// RowData contains the raw string data for each cell in the current row.
	RowData []string
	// ColumnCount is the total number of columns in the table.
	ColumnCount int
	// ColumnData provides configuration details for each column.
	ColumnData []TableColumnData

	// RenderContext provides global rendering information like theming and
	// utility functions.
	RenderContext core.RenderContext

	// ComponentData is a map containing the rendered output of preceding
	// components in the pipeline. This is not typically used in table rendering
	// but is included for consistency with other component systems.
	ComponentData map[TableComponentType]string

	// TableConfig holds the current rendering configuration for the table.
	TableConfig ComponentTableRenderConfig
}

// TableColumnData contains configuration information for a specific column,
// passed to rendering components.
type TableColumnData struct {
	// Header is the title of the column.
	Header string
	// Width is the configured width of the column.
	Width int
	// Alignment specifies the horizontal alignment for the column's content.
	Alignment TableColumnAlignment
	// Column is a reference to the original column definition.
	Column core.TableColumn
}

// ComponentTableRenderConfig holds the complete configuration for the
// component-based rendering pipeline of the table. It defines which components
// are active, their order, and their individual settings.
type ComponentTableRenderConfig struct {
	// ComponentOrder defines the sequence in which the components are rendered.
	ComponentOrder []TableComponentType

	// Component configurations for each part of the table row.
	CursorConfig          TableCursorConfig
	PreSpacingConfig      TableSpacingConfig
	RowNumberConfig       TableRowNumberConfig
	SelectionMarkerConfig TableSelectionMarkerConfig
	CellsConfig           TableCellsConfig
	PostSpacingConfig     TableSpacingConfig
	BackgroundConfig      TableBackgroundConfig
	BorderConfig          TableBorderConfig
}

// TableCursorConfig configures the appearance and behavior of the cursor component.
type TableCursorConfig struct {
	// Enabled toggles the rendering of this component.
	Enabled bool
	// CursorIndicator is the string shown when the item is under the cursor.
	CursorIndicator string
	// NormalSpacing is the string used for alignment when the item is not under
	// the cursor.
	NormalSpacing string
	// Style is the lipgloss style applied to the component's output.
	Style lipgloss.Style
}

// TableSpacingConfig configures a spacing component, used for adding horizontal
// space (padding) within the rendered item line.
type TableSpacingConfig struct {
	// Enabled toggles the rendering of this component.
	Enabled bool
	// Spacing is the string to be rendered, typically composed of spaces.
	Spacing string
	// Style is the lipgloss style applied to the spacing.
	Style lipgloss.Style
}

// TableRowNumberConfig configures the component responsible for rendering row numbers.
type TableRowNumberConfig struct {
	// Enabled toggles the rendering of this component.
	Enabled bool
	// ShowHeader, if true, displays a header for the row number column.
	ShowHeader bool
	// Width is the fixed width for the row number column, ensuring alignment.
	Width int
	// Style is the lipgloss style applied to the row number.
	Style lipgloss.Style
	// Alignment specifies the horizontal alignment of the row number.
	Alignment TableColumnAlignment
	// Formatter is a custom function to format the row number string.
	Formatter RowNumberFormatter
}

// TableSelectionMarkerConfig configures the component that displays a marker for
// selected rows.
type TableSelectionMarkerConfig struct {
	// Enabled toggles the rendering of this component.
	Enabled bool
	// SelectedMarker is the string shown for a selected row.
	SelectedMarker string
	// UnselectedMarker is the string shown for an unselected row.
	UnselectedMarker string
	// Style is the lipgloss style applied to the marker.
	Style lipgloss.Style
	// Width is the fixed width for the marker column.
	Width int
}

// TableCellsConfig configures the component that renders the main content of all
// cells in a row.
type TableCellsConfig struct {
	// Enabled toggles the rendering of this component.
	Enabled bool
	// CellSeparator is the string used to separate columns (e.g., " │ ").
	CellSeparator string
	// CellPadding is the number of spaces to add inside each cell.
	CellPadding int
	// Style is the base lipgloss style for cells.
	Style lipgloss.Style
	// HeaderStyle is the style applied to header cells.
	HeaderStyle lipgloss.Style
	// AlternateStyle is a style applied to alternating rows if UseAlternating is true.
	AlternateStyle lipgloss.Style
	// UseAlternating toggles the application of AlternateStyle.
	UseAlternating bool
}

// TableBackgroundConfig configures the component that applies a background style
// to the rendered row, acting as a post-processing step.
type TableBackgroundConfig struct {
	// Enabled toggles the rendering of this component.
	Enabled bool
	// Style is the background style to apply.
	Style lipgloss.Style
	// ApplyToComponents specifies which other components the background should
	// be applied to when using `TableBackgroundSelectiveComponents` mode.
	ApplyToComponents []TableComponentType
	// Mode determines how the background style is applied.
	Mode TableBackgroundMode
}

// TableBorderConfig configures the component that renders borders around the row.
type TableBorderConfig struct {
	// Enabled toggles the rendering of this component.
	Enabled bool
	// LeftBorder is the character(s) for the left border.
	LeftBorder string
	// RightBorder is the character(s) for the right border.
	RightBorder string
	// CellBorder is the character(s) used between cells when borders are enabled.
	CellBorder string
	// Style is the lipgloss style for the border characters.
	Style lipgloss.Style
}

// TableColumnAlignment defines the horizontal alignment for cell content.
type TableColumnAlignment int

// Constants for column alignment.
const (
	TableAlignLeft TableColumnAlignment = iota
	TableAlignCenter
	TableAlignRight
)

// TableBackgroundMode defines how a background style is applied to a row.
type TableBackgroundMode int

// Constants for background styling modes.
const (
	// TableBackgroundEntireRow applies the style to the entire rendered row.
	TableBackgroundEntireRow TableBackgroundMode = iota
	// TableBackgroundSelectiveComponents applies the style only to a specified
	// subset of components.
	TableBackgroundSelectiveComponents
	// TableBackgroundCellsOnly applies the style only to the main cell content.
	TableBackgroundCellsOnly
	// TableBackgroundIndicatorOnly applies the style only to the cursor indicator.
	TableBackgroundIndicatorOnly
)

// RowNumberFormatter is a function type for generating a row number string.
type RowNumberFormatter func(index int, isHeader bool) string

// TableCursorComponent is a render component responsible for displaying the cursor
// indicator for a table row. It shows a specific string when the item is under
// the cursor and a different string otherwise to ensure alignment.
type TableCursorComponent struct {
	config TableCursorConfig
}

// NewTableCursorComponent creates a new cursor component with the given configuration.
func NewTableCursorComponent(config TableCursorConfig) *TableCursorComponent {
	return &TableCursorComponent{config: config}
}

// Render returns the cursor indicator string based on the context.
func (c *TableCursorComponent) Render(ctx TableComponentContext) string {
	if ctx.IsCursor {
		return c.config.Style.Render(c.config.CursorIndicator)
	}
	return c.config.Style.Render(c.config.NormalSpacing)
}

// GetType returns the unique type identifier for this component.
func (c *TableCursorComponent) GetType() TableComponentType {
	return TableComponentCursor
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TableCursorComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TableCursorComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TablePreSpacingComponent is a render component that adds a fixed-width space
// before the main content, useful for creating indentation or alignment.
type TablePreSpacingComponent struct {
	config TableSpacingConfig
}

// NewTablePreSpacingComponent creates a new pre-spacing component.
func NewTablePreSpacingComponent(config TableSpacingConfig) *TablePreSpacingComponent {
	return &TablePreSpacingComponent{config: config}
}

// Render returns the configured spacing string.
func (c *TablePreSpacingComponent) Render(ctx TableComponentContext) string {
	return c.config.Style.Render(c.config.Spacing)
}

// GetType returns the unique type identifier for this component.
func (c *TablePreSpacingComponent) GetType() TableComponentType {
	return TableComponentPreSpacing
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TablePreSpacingComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TablePreSpacingComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TableRowNumberComponent is a render component that generates the row number
// string, formatted according to its configuration.
type TableRowNumberComponent struct {
	config TableRowNumberConfig
}

// NewTableRowNumberComponent creates a new row number component.
func NewTableRowNumberComponent(config TableRowNumberConfig) *TableRowNumberComponent {
	return &TableRowNumberComponent{config: config}
}

// Render generates the row number string using a formatter or a default, and
// applies width and alignment constraints.
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

// GetType returns the unique type identifier for this component.
func (c *TableRowNumberComponent) GetType() TableComponentType {
	return TableComponentRowNumber
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TableRowNumberComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TableRowNumberComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TableSelectionMarkerComponent is a render component that displays a marker
// to indicate whether a row is selected.
type TableSelectionMarkerComponent struct {
	config TableSelectionMarkerConfig
}

// NewTableSelectionMarkerComponent creates a new selection marker component.
func NewTableSelectionMarkerComponent(config TableSelectionMarkerConfig) *TableSelectionMarkerComponent {
	return &TableSelectionMarkerComponent{config: config}
}

// Render returns the appropriate marker string based on the row's selection state.
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

// GetType returns the unique type identifier for this component.
func (c *TableSelectionMarkerComponent) GetType() TableComponentType {
	return TableComponentSelectionMarker
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TableSelectionMarkerComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TableSelectionMarkerComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TableCellsComponent is the render component responsible for displaying the main
// content of the table cells, joined by a separator.
type TableCellsComponent struct {
	config TableCellsConfig
}

// NewTableCellsComponent creates a new cells component.
func NewTableCellsComponent(config TableCellsConfig) *TableCellsComponent {
	return &TableCellsComponent{config: config}
}

// Render processes the raw cell data, applies column constraints and padding,
// and joins the cells into a single string. It can also apply alternating
// row styling.
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

// GetType returns the unique type identifier for this component.
func (c *TableCellsComponent) GetType() TableComponentType {
	return TableComponentCells
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TableCellsComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TableCellsComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TablePostSpacingComponent is a render component that adds a fixed-width space
// after the main content.
type TablePostSpacingComponent struct {
	config TableSpacingConfig
}

// NewTablePostSpacingComponent creates a new post-spacing component.
func NewTablePostSpacingComponent(config TableSpacingConfig) *TablePostSpacingComponent {
	return &TablePostSpacingComponent{config: config}
}

// Render returns the configured spacing string.
func (c *TablePostSpacingComponent) Render(ctx TableComponentContext) string {
	return c.config.Style.Render(c.config.Spacing)
}

// GetType returns the unique type identifier for this component.
func (c *TablePostSpacingComponent) GetType() TableComponentType {
	return TableComponentPostSpacing
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TablePostSpacingComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TablePostSpacingComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TableBorderComponent is a special render component that wraps the output of
// other components with border characters.
type TableBorderComponent struct {
	config TableBorderConfig
}

// NewTableBorderComponent creates a new border component.
func NewTableBorderComponent(config TableBorderConfig) *TableBorderComponent {
	return &TableBorderComponent{config: config}
}

// Render collects the output of other components and wraps it with the configured
// border characters.
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

// GetType returns the unique type identifier for this component.
func (c *TableBorderComponent) GetType() TableComponentType {
	return TableComponentBorder
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TableBorderComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TableBorderComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TableBackgroundComponent is a special render component that applies a background
// style as a post-processing step. It can apply the style to the entire row or
// to a select subset of the other rendered components.
type TableBackgroundComponent struct {
	config TableBackgroundConfig
}

// NewTableBackgroundComponent creates a new background component.
func NewTableBackgroundComponent(config TableBackgroundConfig) *TableBackgroundComponent {
	return &TableBackgroundComponent{config: config}
}

// Render applies the background style according to the configured mode. It does
// not return its own content but rather modifies the combined output of other
// components.
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

// GetType returns the unique type identifier for this component.
func (c *TableBackgroundComponent) GetType() TableComponentType {
	return TableComponentBackground
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TableBackgroundComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TableBackgroundComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TableComponentRenderer orchestrates the rendering of all table components.
// It holds all the individual render components and processes them in a defined
// order to construct the final string for a single table row.
type TableComponentRenderer struct {
	components map[TableComponentType]TableRenderComponent
	config     ComponentTableRenderConfig
}

// NewTableComponentRenderer creates a new renderer with a given configuration,
// initializing all the necessary components for table rendering.
func NewTableComponentRenderer(config ComponentTableRenderConfig) *TableComponentRenderer {
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

// SetComponentOrder defines the sequence in which the components are rendered.
func (r *TableComponentRenderer) SetComponentOrder(order []TableComponentType) {
	r.config.ComponentOrder = order
}

// EnableComponent activates a specific component in the rendering pipeline.
func (r *TableComponentRenderer) EnableComponent(componentType TableComponentType) {
	if comp, exists := r.components[componentType]; exists {
		comp.SetEnabled(true)
	}
}

// DisableComponent deactivates a specific component.
func (r *TableComponentRenderer) DisableComponent(componentType TableComponentType) {
	if comp, exists := r.components[componentType]; exists {
		comp.SetEnabled(false)
	}
}

// GetComponent retrieves a specific component from the renderer, allowing for
// direct configuration changes.
func (r *TableComponentRenderer) GetComponent(componentType TableComponentType) TableRenderComponent {
	return r.components[componentType]
}

// UpdateConfig applies a new configuration to the renderer, recreating all
// its internal components to reflect the changes.
func (r *TableComponentRenderer) UpdateConfig(config ComponentTableRenderConfig) {
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

// Render executes the full rendering pipeline for a single table row. It iterates
// through the components in the configured order, calls their `Render` methods,
// and assembles the final string. It handles special components like border and
// background, which modify the output of other components.
func (r *TableComponentRenderer) Render(
	item core.Data[any],
	index int,
	rowData []string,
	columnData []TableColumnData,
	renderContext core.RenderContext,
	isCursor, isTopThreshold, isBottomThreshold bool,
) string {
	// Build context for all components
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

// DefaultComponentTableRenderConfig returns a sensible default configuration for
// component-based table rendering, featuring a cursor, selection marker, and cells.
func DefaultComponentTableRenderConfig() ComponentTableRenderConfig {
	return ComponentTableRenderConfig{
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

// NumberedTableConfig provides a pre-configured `ComponentTableRenderConfig` that
// includes a row number column.
func NumberedTableConfig() ComponentTableRenderConfig {
	config := DefaultComponentTableRenderConfig()
	config.ComponentOrder = []TableComponentType{
		TableComponentCursor,
		TableComponentRowNumber,
		TableComponentSelectionMarker,
		TableComponentCells,
	}
	config.RowNumberConfig.Enabled = true
	return config
}

// MinimalTableConfig provides a pre-configured `ComponentTableRenderConfig` with
// only the essential cell content, for a clean, simple look.
func MinimalTableConfig() ComponentTableRenderConfig {
	config := DefaultComponentTableRenderConfig()
	config.ComponentOrder = []TableComponentType{
		TableComponentCells,
	}
	config.CursorConfig.Enabled = false
	config.SelectionMarkerConfig.Enabled = false
	return config
}

// BorderedTableConfig provides a pre-configured `ComponentTableRenderConfig` that
// wraps rows in borders.
func BorderedTableConfig() ComponentTableRenderConfig {
	config := DefaultComponentTableRenderConfig()
	config.ComponentOrder = []TableComponentType{
		TableComponentBorder,
	}
	config.BorderConfig.Enabled = true
	return config
}

// AlternatingTableConfig provides a pre-configured `ComponentTableRenderConfig`
// that applies a background color to alternating rows.
func AlternatingTableConfig() ComponentTableRenderConfig {
	config := DefaultComponentTableRenderConfig()
	config.CellsConfig.UseAlternating = true
	return config
}

// BackgroundStyledTableConfig creates a `ComponentTableRenderConfig` that applies
// a background style to table rows. The mode determines whether the style applies
// to the entire row or just specific components.
func BackgroundStyledTableConfig(style lipgloss.Style, mode TableBackgroundMode) ComponentTableRenderConfig {
	config := DefaultComponentTableRenderConfig()
	config.BackgroundConfig.Enabled = true
	config.BackgroundConfig.Style = style
	config.BackgroundConfig.Mode = mode
	return config
}

// ComponentBasedTableFormatter creates a table formatter function from a
// `ComponentTableRenderConfig`. This allows the component-based rendering pipeline
// to be used as a standard formatter, providing an integration point with systems
// that expect a single formatting function.
func ComponentBasedTableFormatter(config ComponentTableRenderConfig) func(core.Data[any], int, []string, []TableColumnData, core.RenderContext, bool, bool, bool) string {
	renderer := NewTableComponentRenderer(config)
	return func(
		item core.Data[any],
		index int,
		rowData []string,
		columnData []TableColumnData,
		ctx core.RenderContext,
		isCursor, isTopThreshold, isBottomThreshold bool,
	) string {
		return renderer.Render(item, index, rowData, columnData, ctx, isCursor, isTopThreshold, isBottomThreshold)
	}
}

// EnhancedTableFormatter creates a table formatter using the component system.
// This is the main entry point for using the component-based rendering system.
func EnhancedTableFormatter(config ComponentTableRenderConfig) func(core.Data[any], int, []string, []TableColumnData, core.RenderContext, bool, bool, bool) string {
	return ComponentBasedTableFormatter(config)
}

// DefaultRowNumberFormatter provides default numeric row number formatting.
func DefaultRowNumberFormatter(index int, isHeader bool) string {
	if isHeader {
		return "Row"
	}
	return fmt.Sprintf("%d", index+1)
}

// AlphabeticalRowNumberFormatter provides alphabetical row numbering (A, B, C, ...).
func AlphabeticalRowNumberFormatter(index int, isHeader bool) string {
	if isHeader {
		return "Row"
	}
	if index < 26 {
		return string(rune('A' + index))
	}
	return fmt.Sprintf("%d", index+1)
}

// RomanRowNumberFormatter provides Roman numeral row numbering (I, II, III, ...)
// for the first 20 rows, falling back to numeric after that.
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
