package tree

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/core"
)

// This file implements a flexible, component-based rendering system specialized
// for the Tree component. Each part of a tree item's visual representation
// (e.g., cursor, indentation, expand/collapse symbol, content) is a distinct,
// optional component. This allows for extensive customization and reordering of
// how tree items are displayed. This system is distinct from the rendering
// pipelines for the List and Table components.

// TreeRenderComponent defines the contract for a single, modular part of a
// rendered tree item. Implementations of this interface are responsible for
// generating a specific piece of the final string output, such as the cursor
// indicator or the indentation.
type TreeRenderComponent interface {
	// Render generates the string content for this component based on the
	// provided context.
	Render(ctx TreeComponentContext) string
	// GetType returns the component's unique type identifier, used for ordering
	// and configuration.
	GetType() TreeComponentType
	// IsEnabled returns whether this component should be rendered.
	IsEnabled() bool
	// SetEnabled allows enabling or disabling this component at runtime.
	SetEnabled(enabled bool)
}

// TreeComponentType is a string identifier for a specific type of tree
// rendering component. It is used to define the rendering order and to configure
// individual components.
type TreeComponentType string

// Constants defining the available component types for the tree.
const (
	// TreeComponentCursor handles the rendering of the cursor indicator.
	TreeComponentCursor TreeComponentType = "cursor"
	// TreeComponentPreSpacing adds spacing before the main content.
	TreeComponentPreSpacing TreeComponentType = "pre_spacing"
	// TreeComponentIndentation renders the indentation based on the node's depth.
	TreeComponentIndentation TreeComponentType = "indentation"
	// TreeComponentTreeSymbol renders the expand/collapse/leaf symbol.
	TreeComponentTreeSymbol TreeComponentType = "tree_symbol"
	// TreeComponentEnumerator renders a prefix like a bullet or number.
	TreeComponentEnumerator TreeComponentType = "enumerator"
	// TreeComponentContent renders the main item content.
	TreeComponentContent TreeComponentType = "content"
	// TreeComponentPostSpacing adds spacing after the main content.
	TreeComponentPostSpacing TreeComponentType = "post_spacing"
	// TreeComponentBackground applies background styling as a final step.
	TreeComponentBackground TreeComponentType = "background"
)

// TreeComponentContext provides all the necessary data for a TreeRenderComponent
// to render its output. It encapsulates item-specific data, tree structure
// information, and global rendering settings.
type TreeComponentContext struct {
	// Item is the core data for the node being rendered.
	Item core.Data[any]
	// Index is the absolute linear index of the item in the flattened view.
	Index int
	// IsCursor is true if this item is currently under the cursor.
	IsCursor bool
	// IsSelected is true if this item is currently selected.
	IsSelected bool
	// IsThreshold is true if this item is at a scroll threshold.
	IsThreshold bool

	// Depth is the level of the node in the tree hierarchy (root is 0).
	Depth int
	// HasChildren is true if the node has child nodes.
	HasChildren bool
	// IsExpanded is true if the node is currently expanded to show its children.
	IsExpanded bool
	// ParentID is the ID of the parent node.
	ParentID string

	// RenderContext provides global rendering information like theming and
	// utility functions.
	RenderContext core.RenderContext

	// ComponentData is a map containing the rendered output of preceding
	// components in the pipeline, allowing subsequent components to make layout
	// decisions based on the width of earlier ones (e.g., for text wrapping).
	ComponentData map[TreeComponentType]string

	// TreeConfig holds the current rendering configuration for the tree.
	TreeConfig TreeRenderConfig
}

// TreeRenderConfig holds the complete configuration for the component-based
// rendering pipeline of the tree. It defines which components are active, their
// order, and their individual settings.
type TreeRenderConfig struct {
	// ComponentOrder defines the sequence in which the components are rendered.
	ComponentOrder []TreeComponentType

	// Component configurations for each part of the tree item.
	CursorConfig      TreeCursorConfig
	PreSpacingConfig  TreeSpacingConfig
	IndentationConfig TreeIndentationConfig
	TreeSymbolConfig  TreeSymbolConfig
	EnumeratorConfig  TreeEnumeratorConfig
	ContentConfig     TreeContentConfig
	PostSpacingConfig TreeSpacingConfig
	BackgroundConfig  TreeBackgroundConfig
}

// TreeCursorConfig configures the appearance and behavior of the cursor component.
type TreeCursorConfig struct {
	// Enabled toggles the rendering of this component.
	Enabled bool
	// CursorIndicator is the string shown when the item is under the cursor.
	CursorIndicator string
	// NormalSpacing is the string used for alignment when the item is not under
	// the cursor.
	NormalSpacing string
	// Style is the lipgloss style applied to the component's output.
	Style lipgloss.Style
	// ShowOnlyAtRoot, if true, restricts the cursor indicator to only be shown
	// for root-level nodes.
	ShowOnlyAtRoot bool
}

// TreeSpacingConfig configures a spacing component, used for adding horizontal
// space (padding) within the rendered item line.
type TreeSpacingConfig struct {
	// Enabled toggles the rendering of this component.
	Enabled bool
	// Spacing is the string to be rendered, typically composed of spaces.
	Spacing string
	// Style is the lipgloss style applied to the spacing.
	Style lipgloss.Style
}

// TreeIndentationConfig configures the component responsible for rendering the
// indentation that visually represents the tree's hierarchy.
type TreeIndentationConfig struct {
	// Enabled toggles the rendering of this component.
	Enabled bool
	// IndentString is the string repeated for each level of indentation when
	// UseConnectors is false.
	IndentString string
	// IndentSize is the number of spaces used for each indentation level if
	// IndentString is empty.
	IndentSize int
	// Style is the lipgloss style for the indentation.
	Style lipgloss.Style
	// ConnectorStyle is the style applied to box-drawing characters if
	// UseConnectors is true.
	ConnectorStyle lipgloss.Style
	// UseConnectors, if true, renders indentation using box-drawing characters
	// to create a classic tree look.
	UseConnectors bool
}

// TreeSymbolConfig configures the component that displays symbols indicating a
// node's state (expanded, collapsed, or leaf).
type TreeSymbolConfig struct {
	// Enabled toggles the rendering of this component.
	Enabled bool
	// ExpandedSymbol is the string shown for an expanded node (e.g., "▼").
	ExpandedSymbol string
	// CollapsedSymbol is the string shown for a collapsed node (e.g., "▶").
	CollapsedSymbol string
	// LeafSymbol is the string shown for a node with no children.
	LeafSymbol string
	// Style is the lipgloss style applied to the symbol.
	Style lipgloss.Style
	// ShowForLeaves, if true, renders the LeafSymbol for nodes without children.
	ShowForLeaves bool
	// SymbolSpacing is a string (usually a space) added after the symbol for
	// padding.
	SymbolSpacing string
}

// TreeEnumeratorConfig configures the component that renders an enumerator
// (e.g., bullet, number) for each tree item.
type TreeEnumeratorConfig struct {
	// Enabled toggles the rendering of this component.
	Enabled bool
	// Enumerator is the function that generates the enumerator string.
	Enumerator TreeEnumeratorFunc
	// Style is the lipgloss style applied to the enumerator.
	Style lipgloss.Style
	// Alignment controls the horizontal alignment of the enumerator string.
	Alignment TreeEnumeratorAlignment
	// MaxWidth is the maximum width for the enumerator, used for alignment.
	MaxWidth int
}

// TreeContentConfig configures the component that renders the main content of
// the tree item.
type TreeContentConfig struct {
	// Enabled toggles the rendering of this component.
	Enabled bool
	// Formatter is the function that generates the content string from the item data.
	Formatter TreeItemFormatter
	// Style is the lipgloss style applied to the content.
	Style lipgloss.Style
	// WrapText enables or disables text wrapping for the content.
	WrapText bool
	// MaxWidth is the width at which text wrapping will occur.
	MaxWidth int
}

// TreeBackgroundConfig configures the component that applies a background style
// to the rendered item, acting as a post-processing step.
type TreeBackgroundConfig struct {
	// Enabled toggles the rendering of this component.
	Enabled bool
	// Style is the background style to apply.
	Style lipgloss.Style
	// ApplyToComponents specifies which other components the background should
	// be applied to when using `TreeBackgroundSelectiveComponents` mode.
	ApplyToComponents []TreeComponentType
	// Mode determines how the background style is applied.
	Mode TreeBackgroundMode
}

// TreeEnumeratorAlignment defines the horizontal alignment for the enumerator.
type TreeEnumeratorAlignment int

// Constants for enumerator alignment.
const (
	TreeAlignmentNone TreeEnumeratorAlignment = iota
	TreeAlignmentLeft
	TreeAlignmentRight
)

// TreeBackgroundMode defines how the background style is applied to an item.
type TreeBackgroundMode int

// Constants for background styling modes.
const (
	// TreeBackgroundEntireLine applies the style to the entire rendered line.
	TreeBackgroundEntireLine TreeBackgroundMode = iota
	// TreeBackgroundSelectiveComponents applies the style only to a specified
	// subset of components.
	TreeBackgroundSelectiveComponents
	// TreeBackgroundContentOnly applies the style only to the main content.
	TreeBackgroundContentOnly
	// TreeBackgroundIndicatorOnly applies the style only to the cursor indicator.
	TreeBackgroundIndicatorOnly
)

// TreeEnumeratorFunc is a function type for generating an enumerator string,
// receiving tree-specific context like depth and expansion state.
type TreeEnumeratorFunc func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext) string

// TreeItemFormatter is a function type for rendering the main content of a tree
// item, receiving full tree-specific context.
type TreeItemFormatter func(item core.Data[any], index int, depth int, hasChildren, isExpanded bool, ctx core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string

// TreeCursorComponent is a render component responsible for displaying the cursor
// indicator for a tree item. It shows a specific string when the item is under
// the cursor and a different string otherwise to ensure alignment.
type TreeCursorComponent struct {
	config TreeCursorConfig
}

// NewTreeCursorComponent creates a new cursor component with the given configuration.
func NewTreeCursorComponent(config TreeCursorConfig) *TreeCursorComponent {
	return &TreeCursorComponent{config: config}
}

// Render returns the cursor indicator string based on the context.
func (c *TreeCursorComponent) Render(ctx TreeComponentContext) string {
	// Check if we should only show cursor at root level
	if c.config.ShowOnlyAtRoot && ctx.Depth > 0 && ctx.IsCursor {
		return c.config.Style.Render(c.config.NormalSpacing)
	}

	if ctx.IsCursor {
		return c.config.Style.Render(c.config.CursorIndicator)
	}
	return c.config.Style.Render(c.config.NormalSpacing)
}

// GetType returns the unique type identifier for this component.
func (c *TreeCursorComponent) GetType() TreeComponentType {
	return TreeComponentCursor
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TreeCursorComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TreeCursorComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TreePreSpacingComponent is a render component that adds a fixed-width space
// before the main content, useful for creating indentation or alignment.
type TreePreSpacingComponent struct {
	config TreeSpacingConfig
}

// NewTreePreSpacingComponent creates a new pre-spacing component.
func NewTreePreSpacingComponent(config TreeSpacingConfig) *TreePreSpacingComponent {
	return &TreePreSpacingComponent{config: config}
}

// Render returns the configured spacing string.
func (c *TreePreSpacingComponent) Render(ctx TreeComponentContext) string {
	return c.config.Style.Render(c.config.Spacing)
}

// GetType returns the unique type identifier for this component.
func (c *TreePreSpacingComponent) GetType() TreeComponentType {
	return TreeComponentPreSpacing
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TreePreSpacingComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TreePreSpacingComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TreeIndentationComponent is a render component that generates the indentation
// string for a tree node, visually representing its depth in the hierarchy. It
// can use either simple spacing or box-drawing characters for a connected look.
type TreeIndentationComponent struct {
	config TreeIndentationConfig
}

// NewTreeIndentationComponent creates a new indentation component.
func NewTreeIndentationComponent(config TreeIndentationConfig) *TreeIndentationComponent {
	return &TreeIndentationComponent{config: config}
}

// Render generates the indentation string based on the node's depth and the
// component's configuration.
func (c *TreeIndentationComponent) Render(ctx TreeComponentContext) string {
	if ctx.Depth == 0 {
		return ""
	}

	var indent strings.Builder

	if c.config.UseConnectors {
		// Use box-drawing characters for indentation
		for i := 0; i < ctx.Depth; i++ {
			if i == ctx.Depth-1 {
				// This is the last level, use appropriate connector
				if ctx.HasChildren {
					if ctx.IsExpanded {
						indent.WriteString("├─ ")
					} else {
						indent.WriteString("├+ ")
					}
				} else {
					indent.WriteString("├─ ")
				}
			} else {
				// Intermediate levels
				indent.WriteString("│  ")
			}
		}
	} else {
		// Use simple string-based indentation
		if c.config.IndentString != "" {
			for i := 0; i < ctx.Depth; i++ {
				indent.WriteString(c.config.IndentString)
			}
		} else {
			// Use spaces based on IndentSize
			indentSize := c.config.IndentSize
			if indentSize <= 0 {
				indentSize = 2 // Default
			}
			spaces := strings.Repeat(" ", indentSize*ctx.Depth)
			indent.WriteString(spaces)
		}
	}

	return c.config.ConnectorStyle.Render(indent.String())
}

// GetType returns the unique type identifier for this component.
func (c *TreeIndentationComponent) GetType() TreeComponentType {
	return TreeComponentIndentation
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TreeIndentationComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TreeIndentationComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TreeSymbolComponent is a render component that displays a symbol indicating
// the state of a tree node (e.g., expanded, collapsed, or a leaf).
type TreeSymbolComponent struct {
	config TreeSymbolConfig
}

// NewTreeSymbolComponent creates a new tree symbol component.
func NewTreeSymbolComponent(config TreeSymbolConfig) *TreeSymbolComponent {
	return &TreeSymbolComponent{config: config}
}

// Render generates the symbol string based on the node's state (e.g., has
// children, is expanded).
func (c *TreeSymbolComponent) Render(ctx TreeComponentContext) string {
	var symbol string

	if ctx.HasChildren {
		if ctx.IsExpanded {
			symbol = c.config.ExpandedSymbol
		} else {
			symbol = c.config.CollapsedSymbol
		}
	} else if c.config.ShowForLeaves {
		symbol = c.config.LeafSymbol
	}

	if symbol != "" && c.config.SymbolSpacing != "" {
		symbol += c.config.SymbolSpacing
	}

	return c.config.Style.Render(symbol)
}

// GetType returns the unique type identifier for this component.
func (c *TreeSymbolComponent) GetType() TreeComponentType {
	return TreeComponentTreeSymbol
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TreeSymbolComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TreeSymbolComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TreeEnumeratorComponent is a render component that displays an enumerator, such
// as a bullet point or number, next to the tree item. The appearance is
// determined by a provided `TreeEnumeratorFunc`.
type TreeEnumeratorComponent struct {
	config TreeEnumeratorConfig
}

// NewTreeEnumeratorComponent creates a new enumerator component for trees.
func NewTreeEnumeratorComponent(config TreeEnumeratorConfig) *TreeEnumeratorComponent {
	return &TreeEnumeratorComponent{config: config}
}

// Render generates the enumerator string using the configured function and
// applies any specified alignment.
func (c *TreeEnumeratorComponent) Render(ctx TreeComponentContext) string {
	if c.config.Enumerator == nil {
		return ""
	}

	enumText := c.config.Enumerator(
		ctx.Item,
		ctx.Index,
		ctx.Depth,
		ctx.HasChildren,
		ctx.IsExpanded,
		ctx.RenderContext,
	)

	// Apply alignment if configured
	if c.config.Alignment != TreeAlignmentNone && c.config.MaxWidth > 0 {
		switch c.config.Alignment {
		case TreeAlignmentRight:
			padding := c.config.MaxWidth - len(enumText)
			if padding > 0 {
				enumText = strings.Repeat(" ", padding) + enumText
			}
		case TreeAlignmentLeft:
			padding := c.config.MaxWidth - len(enumText)
			if padding > 0 {
				enumText = enumText + strings.Repeat(" ", padding)
			}
		}
	}

	return c.config.Style.Render(enumText)
}

// GetType returns the unique type identifier for this component.
func (c *TreeEnumeratorComponent) GetType() TreeComponentType {
	return TreeComponentEnumerator
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TreeEnumeratorComponent) IsEnabled() bool {
	return c.config.Enabled && c.config.Enumerator != nil
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TreeEnumeratorComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TreeContentComponent is the render component responsible for displaying the
// main content of the tree item. It uses a provided formatter or a default
// formatting logic and can handle text wrapping.
type TreeContentComponent struct {
	config TreeContentConfig
}

// NewTreeContentComponent creates a new content component for trees.
func NewTreeContentComponent(config TreeContentConfig) *TreeContentComponent {
	return &TreeContentComponent{config: config}
}

// Render generates the item's main content string. If text wrapping is enabled,
// it will wrap the text and indent subsequent lines to align with the start of
// the content.
func (c *TreeContentComponent) Render(ctx TreeComponentContext) string {
	var content string

	if c.config.Formatter != nil {
		content = c.config.Formatter(
			ctx.Item,
			ctx.Index,
			ctx.Depth,
			ctx.HasChildren,
			ctx.IsExpanded,
			ctx.RenderContext,
			ctx.IsCursor,
			ctx.IsThreshold,
			ctx.IsThreshold, // Using same for both top/bottom for simplicity
		)
	} else {
		// Default tree content formatting
		content = FormatTreeItemContent(
			ctx.Item,
			ctx.Index,
			ctx.Depth,
			ctx.HasChildren,
			ctx.IsExpanded,
			ctx.RenderContext,
			ctx.IsCursor,
			ctx.IsThreshold,
			ctx.IsThreshold,
			nil,
		)
	}

	// Handle text wrapping if enabled
	if c.config.WrapText && c.config.MaxWidth > 0 && ctx.RenderContext.Wrap != nil {
		lines := ctx.RenderContext.Wrap(content, c.config.MaxWidth)
		if len(lines) > 1 {
			// Multi-line content - handle indentation
			// Calculate indent based on previous components
			var indentSize int
			for _, compType := range []TreeComponentType{
				TreeComponentCursor, TreeComponentPreSpacing, TreeComponentIndentation,
				TreeComponentTreeSymbol, TreeComponentEnumerator,
			} {
				if compContent, exists := ctx.ComponentData[compType]; exists {
					indentSize += len(compContent)
				}
			}
			content = strings.Join(lines, "\n"+strings.Repeat(" ", indentSize))
		} else if len(lines) == 1 {
			content = lines[0]
		}
	}

	return c.config.Style.Render(content)
}

// GetType returns the unique type identifier for this component.
func (c *TreeContentComponent) GetType() TreeComponentType {
	return TreeComponentContent
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TreeContentComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TreeContentComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TreePostSpacingComponent is a render component that adds a fixed-width space
// after the main content.
type TreePostSpacingComponent struct {
	config TreeSpacingConfig
}

// NewTreePostSpacingComponent creates a new post-spacing component.
func NewTreePostSpacingComponent(config TreeSpacingConfig) *TreePostSpacingComponent {
	return &TreePostSpacingComponent{config: config}
}

// Render returns the configured spacing string.
func (c *TreePostSpacingComponent) Render(ctx TreeComponentContext) string {
	return c.config.Style.Render(c.config.Spacing)
}

// GetType returns the unique type identifier for this component.
func (c *TreePostSpacingComponent) GetType() TreeComponentType {
	return TreeComponentPostSpacing
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TreePostSpacingComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TreePostSpacingComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TreeBackgroundComponent is a special render component that applies a background
// style as a post-processing step. It can apply the style to the entire line or
// to a select subset of the other rendered components.
type TreeBackgroundComponent struct {
	config TreeBackgroundConfig
}

// NewTreeBackgroundComponent creates a new background component for trees.
func NewTreeBackgroundComponent(config TreeBackgroundConfig) *TreeBackgroundComponent {
	return &TreeBackgroundComponent{config: config}
}

// Render applies the background style according to the configured mode. It does
// not return its own content but rather modifies the combined output of other
// components.
func (c *TreeBackgroundComponent) Render(ctx TreeComponentContext) string {
	switch c.config.Mode {
	case TreeBackgroundEntireLine:
		// Apply background to the entire combined content
		var parts []string
		for _, compType := range ctx.TreeConfig.ComponentOrder {
			if compType == TreeComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				parts = append(parts, content)
			}
		}
		fullContent := strings.Join(parts, "")
		return c.config.Style.Render(fullContent)

	case TreeBackgroundSelectiveComponents:
		// Apply background only to specified components
		var result strings.Builder
		for _, compType := range ctx.TreeConfig.ComponentOrder {
			if compType == TreeComponentBackground {
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

	case TreeBackgroundContentOnly:
		// Apply background only to content
		var result strings.Builder
		for _, compType := range ctx.TreeConfig.ComponentOrder {
			if compType == TreeComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				if compType == TreeComponentContent {
					result.WriteString(c.config.Style.Render(content))
				} else {
					result.WriteString(content)
				}
			}
		}
		return result.String()

	case TreeBackgroundIndicatorOnly:
		// Apply background only to cursor indicator
		var result strings.Builder
		for _, compType := range ctx.TreeConfig.ComponentOrder {
			if compType == TreeComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				if compType == TreeComponentCursor {
					result.WriteString(c.config.Style.Render(content))
				} else {
					result.WriteString(content)
				}
			}
		}
		return result.String()

	default:
		// Fallback to entire line
		var parts []string
		for _, compType := range ctx.TreeConfig.ComponentOrder {
			if compType == TreeComponentBackground {
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
func (c *TreeBackgroundComponent) GetType() TreeComponentType {
	return TreeComponentBackground
}

// IsEnabled checks if the component is configured to be rendered.
func (c *TreeBackgroundComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *TreeBackgroundComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TreeComponentRenderer manages the rendering pipeline for a tree item. It holds
// all the individual render components and processes them in a defined order to
// construct the final string for a single tree item.
type TreeComponentRenderer struct {
	components map[TreeComponentType]TreeRenderComponent
	config     TreeRenderConfig
}

// NewTreeComponentRenderer creates a new renderer with a given configuration,
// initializing all the necessary components for tree rendering.
func NewTreeComponentRenderer(config TreeRenderConfig) *TreeComponentRenderer {
	renderer := &TreeComponentRenderer{
		components: make(map[TreeComponentType]TreeRenderComponent),
		config:     config,
	}

	// Create components based on config
	renderer.components[TreeComponentCursor] = NewTreeCursorComponent(config.CursorConfig)
	renderer.components[TreeComponentPreSpacing] = NewTreePreSpacingComponent(config.PreSpacingConfig)
	renderer.components[TreeComponentIndentation] = NewTreeIndentationComponent(config.IndentationConfig)
	renderer.components[TreeComponentTreeSymbol] = NewTreeSymbolComponent(config.TreeSymbolConfig)
	renderer.components[TreeComponentEnumerator] = NewTreeEnumeratorComponent(config.EnumeratorConfig)
	renderer.components[TreeComponentContent] = NewTreeContentComponent(config.ContentConfig)
	renderer.components[TreeComponentPostSpacing] = NewTreePostSpacingComponent(config.PostSpacingConfig)
	renderer.components[TreeComponentBackground] = NewTreeBackgroundComponent(config.BackgroundConfig)

	return renderer
}

// SetComponentOrder defines the sequence in which the components are rendered.
func (r *TreeComponentRenderer) SetComponentOrder(order []TreeComponentType) {
	r.config.ComponentOrder = order
}

// EnableComponent activates a specific component in the rendering pipeline.
func (r *TreeComponentRenderer) EnableComponent(componentType TreeComponentType) {
	if comp, exists := r.components[componentType]; exists {
		comp.SetEnabled(true)
	}
}

// DisableComponent deactivates a specific component.
func (r *TreeComponentRenderer) DisableComponent(componentType TreeComponentType) {
	if comp, exists := r.components[componentType]; exists {
		comp.SetEnabled(false)
	}
}

// GetComponent retrieves a specific component from the renderer, allowing for
// direct configuration changes.
func (r *TreeComponentRenderer) GetComponent(componentType TreeComponentType) TreeRenderComponent {
	return r.components[componentType]
}

// UpdateConfig applies a new configuration to the renderer, recreating all
// its internal components to reflect the changes.
func (r *TreeComponentRenderer) UpdateConfig(config TreeRenderConfig) {
	r.config = config
	// Recreate components with new config
	r.components[TreeComponentCursor] = NewTreeCursorComponent(config.CursorConfig)
	r.components[TreeComponentPreSpacing] = NewTreePreSpacingComponent(config.PreSpacingConfig)
	r.components[TreeComponentIndentation] = NewTreeIndentationComponent(config.IndentationConfig)
	r.components[TreeComponentTreeSymbol] = NewTreeSymbolComponent(config.TreeSymbolConfig)
	r.components[TreeComponentEnumerator] = NewTreeEnumeratorComponent(config.EnumeratorConfig)
	r.components[TreeComponentContent] = NewTreeContentComponent(config.ContentConfig)
	r.components[TreeComponentPostSpacing] = NewTreePostSpacingComponent(config.PostSpacingConfig)
	r.components[TreeComponentBackground] = NewTreeBackgroundComponent(config.BackgroundConfig)
}

// Render executes the full rendering pipeline for a single tree item. It iterates
// through the components in the configured order, calls their `Render` methods,
// and assembles the final string. It handles special cases like the background
// component, which modifies the output of other components.
func (r *TreeComponentRenderer) Render(
	item core.Data[any],
	index int,
	depth int,
	hasChildren, isExpanded bool,
	renderContext core.RenderContext,
	isCursor, isTopThreshold, isBottomThreshold bool,
) string {
	ctx := TreeComponentContext{
		Item:          item,
		Index:         index,
		IsCursor:      isCursor,
		IsSelected:    item.Selected,
		IsThreshold:   isTopThreshold || isBottomThreshold,
		Depth:         depth,
		HasChildren:   hasChildren,
		IsExpanded:    isExpanded,
		RenderContext: renderContext,
		ComponentData: make(map[TreeComponentType]string),
		TreeConfig:    r.config,
	}

	// First pass: render all non-background components
	for _, compType := range r.config.ComponentOrder {
		if compType == TreeComponentBackground {
			continue // Handle background separately
		}

		if comp, exists := r.components[compType]; exists && comp.IsEnabled() {
			content := comp.Render(ctx)
			ctx.ComponentData[compType] = content
		}
	}

	// Check if we have a background component and apply it if enabled
	backgroundComp := r.components[TreeComponentBackground]
	if backgroundComp != nil && backgroundComp.IsEnabled() {
		// Apply background styling - but only for cursor items to maintain expected behavior
		if isCursor {
			return backgroundComp.Render(ctx)
		}
	}

	// No background styling - combine components normally
	var result strings.Builder
	for _, compType := range r.config.ComponentOrder {
		if compType == TreeComponentBackground {
			continue
		}
		if content, exists := ctx.ComponentData[compType]; exists {
			result.WriteString(content)
		}
	}

	return result.String()
}

// DefaultTreeRenderConfig returns a sensible default configuration for a
// component-based tree, featuring a cursor, indentation, tree symbols, and content.
func DefaultTreeRenderConfig() TreeRenderConfig {
	return TreeRenderConfig{
		ComponentOrder: []TreeComponentType{
			TreeComponentCursor,
			TreeComponentIndentation,
			TreeComponentTreeSymbol,
			TreeComponentContent,
		},
		CursorConfig: TreeCursorConfig{
			Enabled:         true,
			CursorIndicator: "► ",
			NormalSpacing:   "  ",
			Style:           lipgloss.NewStyle(),
			ShowOnlyAtRoot:  false,
		},
		PreSpacingConfig: TreeSpacingConfig{
			Enabled: false,
			Spacing: "",
			Style:   lipgloss.NewStyle(),
		},
		IndentationConfig: TreeIndentationConfig{
			Enabled:        true,
			IndentString:   "",
			IndentSize:     2,
			Style:          lipgloss.NewStyle(),
			ConnectorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
			UseConnectors:  false,
		},
		TreeSymbolConfig: TreeSymbolConfig{
			Enabled:         true,
			ExpandedSymbol:  "▼",
			CollapsedSymbol: "▶",
			LeafSymbol:      "•",
			Style:           lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
			ShowForLeaves:   true,
			SymbolSpacing:   " ",
		},
		EnumeratorConfig: TreeEnumeratorConfig{
			Enabled:    false,
			Enumerator: nil,
			Style:      lipgloss.NewStyle(),
			Alignment:  TreeAlignmentNone,
			MaxWidth:   0,
		},
		ContentConfig: TreeContentConfig{
			Enabled:   true,
			Formatter: nil,
			Style:     lipgloss.NewStyle(),
			WrapText:  false,
			MaxWidth:  80,
		},
		PostSpacingConfig: TreeSpacingConfig{
			Enabled: false,
			Spacing: "",
			Style:   lipgloss.NewStyle(),
		},
		BackgroundConfig: TreeBackgroundConfig{
			Enabled:           false,
			Style:             lipgloss.NewStyle(),
			ApplyToComponents: []TreeComponentType{TreeComponentCursor, TreeComponentContent},
			Mode:              TreeBackgroundEntireLine,
		},
	}
}

// StandardTreeConfig provides a pre-configured `TreeRenderConfig` that uses
// box-drawing characters for a classic, connected tree appearance.
func StandardTreeConfig() TreeRenderConfig {
	config := DefaultTreeRenderConfig()
	config.IndentationConfig.UseConnectors = true
	return config
}

// MinimalTreeConfig provides a pre-configured `TreeRenderConfig` with only
// indentation and content enabled, for a clean, simple look.
func MinimalTreeConfig() TreeRenderConfig {
	config := DefaultTreeRenderConfig()
	config.ComponentOrder = []TreeComponentType{
		TreeComponentIndentation,
		TreeComponentContent,
	}
	config.CursorConfig.Enabled = false
	config.TreeSymbolConfig.Enabled = false
	return config
}

// EnumeratedTreeConfig creates a `TreeRenderConfig` that includes an enumerator
// component, allowing for numbered or bulleted tree items.
func EnumeratedTreeConfig(enumerator TreeEnumeratorFunc) TreeRenderConfig {
	config := DefaultTreeRenderConfig()
	config.ComponentOrder = []TreeComponentType{
		TreeComponentCursor,
		TreeComponentIndentation,
		TreeComponentEnumerator,
		TreeComponentContent,
	}
	config.EnumeratorConfig.Enabled = true
	config.EnumeratorConfig.Enumerator = enumerator
	config.TreeSymbolConfig.Enabled = false
	return config
}

// BackgroundStyledTreeConfig creates a `TreeRenderConfig` that applies a
// background style to tree items. The mode determines whether the style applies
// to the entire line or just specific components.
func BackgroundStyledTreeConfig(style lipgloss.Style, mode TreeBackgroundMode) TreeRenderConfig {
	config := DefaultTreeRenderConfig()
	config.BackgroundConfig.Enabled = true
	config.BackgroundConfig.Style = style
	config.BackgroundConfig.Mode = mode
	return config
}

// ComponentBasedTreeFormatter creates a `TreeItemFormatter` function from a
// `TreeRenderConfig`. This allows the component-based rendering pipeline to be
// used as a standard formatter, providing an integration point with systems
// that expect a single formatting function.
func ComponentBasedTreeFormatter(config TreeRenderConfig) TreeItemFormatter {
	renderer := NewTreeComponentRenderer(config)
	return func(
		item core.Data[any],
		index int,
		depth int,
		hasChildren, isExpanded bool,
		ctx core.RenderContext,
		isCursor, isTopThreshold, isBottomThreshold bool,
	) string {
		return renderer.Render(item, index, depth, hasChildren, isExpanded, ctx, isCursor, isTopThreshold, isBottomThreshold)
	}
}

// EnhancedTreeFormatter is an alias for `ComponentBasedTreeFormatter`. It serves
// as the primary entry point for using the component-based rendering system.
func EnhancedTreeFormatter(config TreeRenderConfig) TreeItemFormatter {
	return ComponentBasedTreeFormatter(config)
}

// FormatTreeItemContent provides a default way to format the main content of a
// tree item. It converts the item's data to a string and appends indicators
// for states like error, loading, or selection based on the render context.
func FormatTreeItemContent(
	item core.Data[any],
	index int,
	depth int,
	hasChildren, isExpanded bool,
	renderContext core.RenderContext,
	isCursor, isTopThreshold, isBottomThreshold bool,
	formatter TreeItemFormatter,
) string {
	if formatter != nil {
		return formatter(
			item,
			index,
			depth,
			hasChildren,
			isExpanded,
			renderContext,
			isCursor,
			isTopThreshold,
			isBottomThreshold,
		)
	}

	// Enhanced default formatting for tree items
	var content string
	switch v := item.Item.(type) {
	case string:
		content = v
	case fmt.Stringer:
		content = v.String()
	default:
		// Use standard Go formatting for any type
		content = fmt.Sprintf("%v", item.Item)
	}

	// Add configurable state indicators using render context
	var stateIndicator string

	// Add error/loading/disabled indicators
	switch {
	case item.Error != nil:
		if renderContext.ErrorIndicator != "" {
			stateIndicator += " " + renderContext.ErrorIndicator
		}
	case item.Loading:
		if renderContext.LoadingIndicator != "" {
			stateIndicator += " " + renderContext.LoadingIndicator
		}
	case item.Disabled:
		if renderContext.DisabledIndicator != "" {
			stateIndicator += " " + renderContext.DisabledIndicator
		}
	}

	// Add selection indicator if selected
	if item.Selected && renderContext.SelectedIndicator != "" {
		stateIndicator += " " + renderContext.SelectedIndicator
	}

	return content + stateIndicator
}
