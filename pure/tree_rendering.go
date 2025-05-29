package vtable

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ================================
// COMPONENT-BASED TREE RENDERING
// ================================
// A flexible rendering system where each part of a tree item is a distinct,
// optional component that can be customized and reordered as needed.
// This is specialized for trees - lists and tables have their own systems.

// TreeRenderComponent represents a single rendering component for tree items
type TreeRenderComponent interface {
	// Render generates the string content for this component
	Render(ctx TreeComponentContext) string
	// GetType returns the component type for identification
	GetType() TreeComponentType
	// IsEnabled returns whether this component should be rendered
	IsEnabled() bool
	// SetEnabled enables or disables this component
	SetEnabled(enabled bool)
}

// TreeComponentType identifies different types of tree rendering components
type TreeComponentType string

const (
	TreeComponentCursor      TreeComponentType = "cursor"
	TreeComponentPreSpacing  TreeComponentType = "pre_spacing"
	TreeComponentIndentation TreeComponentType = "indentation"
	TreeComponentTreeSymbol  TreeComponentType = "tree_symbol"
	TreeComponentEnumerator  TreeComponentType = "enumerator"
	TreeComponentContent     TreeComponentType = "content"
	TreeComponentPostSpacing TreeComponentType = "post_spacing"
	TreeComponentBackground  TreeComponentType = "background"
)

// TreeComponentContext provides all the context needed for tree component rendering
type TreeComponentContext struct {
	// Item data
	Item        Data[any]
	Index       int
	IsCursor    bool
	IsSelected  bool
	IsThreshold bool

	// Tree-specific data
	Depth       int
	HasChildren bool
	IsExpanded  bool
	ParentID    string

	// Rendering context
	RenderContext RenderContext

	// Component-specific data (populated by other components during rendering)
	ComponentData map[TreeComponentType]string

	// Tree-specific configuration
	TreeConfig TreeRenderConfig
}

// TreeRenderConfig contains configuration for component-based tree rendering
type TreeRenderConfig struct {
	// Component order - defines which components render and in what order
	ComponentOrder []TreeComponentType

	// Component configurations
	CursorConfig      TreeCursorConfig
	PreSpacingConfig  TreeSpacingConfig
	IndentationConfig TreeIndentationConfig
	TreeSymbolConfig  TreeSymbolConfig
	EnumeratorConfig  TreeEnumeratorConfig
	ContentConfig     TreeContentConfig
	PostSpacingConfig TreeSpacingConfig
	BackgroundConfig  TreeBackgroundConfig
}

// Individual component configurations for trees
type TreeCursorConfig struct {
	Enabled         bool
	CursorIndicator string
	NormalSpacing   string
	Style           lipgloss.Style
	ShowOnlyAtRoot  bool // Only show cursor indicator at root level
}

type TreeSpacingConfig struct {
	Enabled bool
	Spacing string
	Style   lipgloss.Style
}

type TreeIndentationConfig struct {
	Enabled        bool
	IndentString   string // What to use for each level of indentation
	IndentSize     int    // How many spaces per level
	Style          lipgloss.Style
	ConnectorStyle lipgloss.Style
	UseConnectors  bool // Whether to use box-drawing characters
}

type TreeSymbolConfig struct {
	Enabled         bool
	ExpandedSymbol  string
	CollapsedSymbol string
	LeafSymbol      string
	Style           lipgloss.Style
	ShowForLeaves   bool   // Whether to show symbols for leaf nodes
	SymbolSpacing   string // Spacing after symbol
}

type TreeEnumeratorConfig struct {
	Enabled    bool
	Enumerator TreeEnumeratorFunc
	Style      lipgloss.Style
	Alignment  TreeEnumeratorAlignment
	MaxWidth   int
}

type TreeContentConfig struct {
	Enabled   bool
	Formatter TreeItemFormatter
	Style     lipgloss.Style
	WrapText  bool
	MaxWidth  int
}

type TreeBackgroundConfig struct {
	Enabled           bool
	Style             lipgloss.Style
	ApplyToComponents []TreeComponentType
	Mode              TreeBackgroundMode
}

type TreeEnumeratorAlignment int

const (
	TreeAlignmentNone TreeEnumeratorAlignment = iota
	TreeAlignmentLeft
	TreeAlignmentRight
)

type TreeBackgroundMode int

const (
	TreeBackgroundEntireLine TreeBackgroundMode = iota
	TreeBackgroundSelectiveComponents
	TreeBackgroundContentOnly
	TreeBackgroundIndicatorOnly
)

// TreeEnumeratorFunc is like ListEnumerator but with tree-specific context
type TreeEnumeratorFunc func(item Data[any], index int, depth int, hasChildren, isExpanded bool, ctx RenderContext) string

// TreeItemFormatter is like ItemFormatter but with tree-specific context
type TreeItemFormatter func(item Data[any], index int, depth int, hasChildren, isExpanded bool, ctx RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string

// ================================
// COMPONENT IMPLEMENTATIONS
// ================================

// TreeCursorComponent handles cursor indicator rendering for trees
type TreeCursorComponent struct {
	config TreeCursorConfig
}

func NewTreeCursorComponent(config TreeCursorConfig) *TreeCursorComponent {
	return &TreeCursorComponent{config: config}
}

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

func (c *TreeCursorComponent) GetType() TreeComponentType {
	return TreeComponentCursor
}

func (c *TreeCursorComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *TreeCursorComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TreePreSpacingComponent handles spacing before tree content
type TreePreSpacingComponent struct {
	config TreeSpacingConfig
}

func NewTreePreSpacingComponent(config TreeSpacingConfig) *TreePreSpacingComponent {
	return &TreePreSpacingComponent{config: config}
}

func (c *TreePreSpacingComponent) Render(ctx TreeComponentContext) string {
	return c.config.Style.Render(c.config.Spacing)
}

func (c *TreePreSpacingComponent) GetType() TreeComponentType {
	return TreeComponentPreSpacing
}

func (c *TreePreSpacingComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *TreePreSpacingComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TreeIndentationComponent handles tree indentation
type TreeIndentationComponent struct {
	config TreeIndentationConfig
}

func NewTreeIndentationComponent(config TreeIndentationConfig) *TreeIndentationComponent {
	return &TreeIndentationComponent{config: config}
}

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

	return c.config.Style.Render(indent.String())
}

func (c *TreeIndentationComponent) GetType() TreeComponentType {
	return TreeComponentIndentation
}

func (c *TreeIndentationComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *TreeIndentationComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TreeSymbolComponent handles tree expansion/collapse symbols
type TreeSymbolComponent struct {
	config TreeSymbolConfig
}

func NewTreeSymbolComponent(config TreeSymbolConfig) *TreeSymbolComponent {
	return &TreeSymbolComponent{config: config}
}

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

func (c *TreeSymbolComponent) GetType() TreeComponentType {
	return TreeComponentTreeSymbol
}

func (c *TreeSymbolComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *TreeSymbolComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TreeEnumeratorComponent handles enumeration for tree items
type TreeEnumeratorComponent struct {
	config TreeEnumeratorConfig
}

func NewTreeEnumeratorComponent(config TreeEnumeratorConfig) *TreeEnumeratorComponent {
	return &TreeEnumeratorComponent{config: config}
}

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

func (c *TreeEnumeratorComponent) GetType() TreeComponentType {
	return TreeComponentEnumerator
}

func (c *TreeEnumeratorComponent) IsEnabled() bool {
	return c.config.Enabled && c.config.Enumerator != nil
}

func (c *TreeEnumeratorComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TreeContentComponent handles the main tree item content
type TreeContentComponent struct {
	config TreeContentConfig
}

func NewTreeContentComponent(config TreeContentConfig) *TreeContentComponent {
	return &TreeContentComponent{config: config}
}

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

func (c *TreeContentComponent) GetType() TreeComponentType {
	return TreeComponentContent
}

func (c *TreeContentComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *TreeContentComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TreePostSpacingComponent handles spacing after tree content
type TreePostSpacingComponent struct {
	config TreeSpacingConfig
}

func NewTreePostSpacingComponent(config TreeSpacingConfig) *TreePostSpacingComponent {
	return &TreePostSpacingComponent{config: config}
}

func (c *TreePostSpacingComponent) Render(ctx TreeComponentContext) string {
	return c.config.Style.Render(c.config.Spacing)
}

func (c *TreePostSpacingComponent) GetType() TreeComponentType {
	return TreeComponentPostSpacing
}

func (c *TreePostSpacingComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *TreePostSpacingComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// TreeBackgroundComponent handles background styling for tree items
type TreeBackgroundComponent struct {
	config TreeBackgroundConfig
}

func NewTreeBackgroundComponent(config TreeBackgroundConfig) *TreeBackgroundComponent {
	return &TreeBackgroundComponent{config: config}
}

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

func (c *TreeBackgroundComponent) GetType() TreeComponentType {
	return TreeComponentBackground
}

func (c *TreeBackgroundComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *TreeBackgroundComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// ================================
// TREE COMPONENT RENDERER
// ================================

// TreeComponentRenderer orchestrates the rendering of all tree components
type TreeComponentRenderer struct {
	components map[TreeComponentType]TreeRenderComponent
	config     TreeRenderConfig
}

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

// SetComponentOrder sets the rendering order of components
func (r *TreeComponentRenderer) SetComponentOrder(order []TreeComponentType) {
	r.config.ComponentOrder = order
}

// EnableComponent enables a specific component
func (r *TreeComponentRenderer) EnableComponent(componentType TreeComponentType) {
	if comp, exists := r.components[componentType]; exists {
		comp.SetEnabled(true)
	}
}

// DisableComponent disables a specific component
func (r *TreeComponentRenderer) DisableComponent(componentType TreeComponentType) {
	if comp, exists := r.components[componentType]; exists {
		comp.SetEnabled(false)
	}
}

// GetComponent returns a component by type
func (r *TreeComponentRenderer) GetComponent(componentType TreeComponentType) TreeRenderComponent {
	return r.components[componentType]
}

// UpdateConfig updates the renderer configuration
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

// Render renders all components in the specified order
func (r *TreeComponentRenderer) Render(
	item Data[any],
	index int,
	depth int,
	hasChildren, isExpanded bool,
	renderContext RenderContext,
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

// ================================
// PRESET CONFIGURATIONS
// ================================

// DefaultTreeRenderConfig returns sensible defaults for component-based tree rendering
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

// StandardTreeConfig creates a config for standard trees with box-drawing characters
func StandardTreeConfig() TreeRenderConfig {
	config := DefaultTreeRenderConfig()
	config.IndentationConfig.UseConnectors = true
	return config
}

// MinimalTreeConfig creates a config for minimal trees
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

// EnumeratedTreeConfig creates a config for trees with enumeration
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

// BackgroundStyledTreeConfig creates a config with background styling
func BackgroundStyledTreeConfig(style lipgloss.Style, mode TreeBackgroundMode) TreeRenderConfig {
	config := DefaultTreeRenderConfig()
	config.BackgroundConfig.Enabled = true
	config.BackgroundConfig.Style = style
	config.BackgroundConfig.Mode = mode
	return config
}

// ================================
// INTEGRATION WITH EXISTING SYSTEM
// ================================

// ComponentBasedTreeFormatter creates a TreeItemFormatter that uses the component system
func ComponentBasedTreeFormatter(config TreeRenderConfig) TreeItemFormatter {
	renderer := NewTreeComponentRenderer(config)
	return func(
		item Data[any],
		index int,
		depth int,
		hasChildren, isExpanded bool,
		ctx RenderContext,
		isCursor, isTopThreshold, isBottomThreshold bool,
	) string {
		return renderer.Render(item, index, depth, hasChildren, isExpanded, ctx, isCursor, isTopThreshold, isBottomThreshold)
	}
}

// EnhancedTreeFormatter creates a TreeItemFormatter using the component system
// This is the main entry point for tree rendering
func EnhancedTreeFormatter(config TreeRenderConfig) TreeItemFormatter {
	return ComponentBasedTreeFormatter(config)
}

// ================================
// UTILITY FUNCTIONS
// ================================

// FormatTreeItemContent formats tree item content with default behavior
func FormatTreeItemContent(
	item Data[any],
	index int,
	depth int,
	hasChildren, isExpanded bool,
	renderContext RenderContext,
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
