package vtable

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ================================
// COMPONENT-BASED LIST RENDERING
// ================================
// A flexible rendering system where each part of a list item is a distinct,
// optional component that can be customized and reordered as needed.
// This is specialized for lists only - tables and trees will have their own systems.

// ListRenderComponent represents a single rendering component for list items
type ListRenderComponent interface {
	// Render generates the string content for this component
	Render(ctx ListComponentContext) string
	// GetType returns the component type for identification
	GetType() ListComponentType
	// IsEnabled returns whether this component should be rendered
	IsEnabled() bool
	// SetEnabled enables or disables this component
	SetEnabled(enabled bool)
}

// ListComponentType identifies different types of list rendering components
type ListComponentType string

const (
	ListComponentCursor      ListComponentType = "cursor"
	ListComponentPreSpacing  ListComponentType = "pre_spacing"
	ListComponentEnumerator  ListComponentType = "enumerator"
	ListComponentContent     ListComponentType = "content"
	ListComponentPostSpacing ListComponentType = "post_spacing"
	ListComponentBackground  ListComponentType = "background"
)

// ListComponentContext provides all the context needed for list component rendering
type ListComponentContext struct {
	// Item data
	Item        Data[any]
	Index       int
	IsCursor    bool
	IsSelected  bool
	IsThreshold bool

	// Rendering context
	RenderContext RenderContext

	// Component-specific data (populated by other components during rendering)
	ComponentData map[ListComponentType]string

	// List-specific configuration
	ListConfig ListRenderConfig
}

// ListRenderConfig contains configuration for component-based list rendering
type ListRenderConfig struct {
	// Component order - defines which components render and in what order
	ComponentOrder []ListComponentType

	// Component configurations
	CursorConfig      ListCursorConfig
	PreSpacingConfig  ListSpacingConfig
	EnumeratorConfig  ListEnumeratorConfig
	ContentConfig     ListContentConfig
	PostSpacingConfig ListSpacingConfig
	BackgroundConfig  ListBackgroundConfig
}

// Individual component configurations
type ListCursorConfig struct {
	Enabled         bool
	CursorIndicator string
	NormalSpacing   string
	Style           lipgloss.Style
}

type ListSpacingConfig struct {
	Enabled bool
	Spacing string
	Style   lipgloss.Style
}

type ListEnumeratorConfig struct {
	Enabled    bool
	Enumerator ListEnumerator
	Style      lipgloss.Style
	Alignment  ListEnumeratorAlignment
	MaxWidth   int
}

type ListContentConfig struct {
	Enabled   bool
	Formatter ItemFormatter[any]
	Style     lipgloss.Style
	WrapText  bool
	MaxWidth  int
}

type ListBackgroundConfig struct {
	Enabled           bool
	Style             lipgloss.Style
	ApplyToComponents []ListComponentType
	Mode              ListBackgroundMode
}

type ListEnumeratorAlignment int

const (
	ListAlignmentNone ListEnumeratorAlignment = iota
	ListAlignmentLeft
	ListAlignmentRight
)

type ListBackgroundMode int

const (
	ListBackgroundEntireLine ListBackgroundMode = iota
	ListBackgroundSelectiveComponents
	ListBackgroundContentOnly
	ListBackgroundIndicatorOnly
)

// ================================
// COMPONENT IMPLEMENTATIONS
// ================================

// ListCursorComponent handles cursor indicator rendering
type ListCursorComponent struct {
	config ListCursorConfig
}

func NewListCursorComponent(config ListCursorConfig) *ListCursorComponent {
	return &ListCursorComponent{config: config}
}

func (c *ListCursorComponent) Render(ctx ListComponentContext) string {
	if ctx.IsCursor {
		return c.config.Style.Render(c.config.CursorIndicator)
	}
	return c.config.Style.Render(c.config.NormalSpacing)
}

func (c *ListCursorComponent) GetType() ListComponentType {
	return ListComponentCursor
}

func (c *ListCursorComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *ListCursorComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// ListPreSpacingComponent handles spacing before the main content
type ListPreSpacingComponent struct {
	config ListSpacingConfig
}

func NewListPreSpacingComponent(config ListSpacingConfig) *ListPreSpacingComponent {
	return &ListPreSpacingComponent{config: config}
}

func (c *ListPreSpacingComponent) Render(ctx ListComponentContext) string {
	return c.config.Style.Render(c.config.Spacing)
}

func (c *ListPreSpacingComponent) GetType() ListComponentType {
	return ListComponentPreSpacing
}

func (c *ListPreSpacingComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *ListPreSpacingComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// ListEnumeratorComponent handles enumeration (bullets, numbers, etc.)
type ListEnumeratorComponent struct {
	config ListEnumeratorConfig
}

func NewListEnumeratorComponent(config ListEnumeratorConfig) *ListEnumeratorComponent {
	return &ListEnumeratorComponent{config: config}
}

func (c *ListEnumeratorComponent) Render(ctx ListComponentContext) string {
	if c.config.Enumerator == nil {
		return ""
	}

	enumText := c.config.Enumerator(ctx.Item, ctx.Index, ctx.RenderContext)

	// Apply alignment if configured
	if c.config.Alignment != ListAlignmentNone && c.config.MaxWidth > 0 {
		switch c.config.Alignment {
		case ListAlignmentRight:
			padding := c.config.MaxWidth - len(enumText)
			if padding > 0 {
				enumText = strings.Repeat(" ", padding) + enumText
			}
		case ListAlignmentLeft:
			padding := c.config.MaxWidth - len(enumText)
			if padding > 0 {
				enumText = enumText + strings.Repeat(" ", padding)
			}
		}
	}

	return c.config.Style.Render(enumText)
}

func (c *ListEnumeratorComponent) GetType() ListComponentType {
	return ListComponentEnumerator
}

func (c *ListEnumeratorComponent) IsEnabled() bool {
	return c.config.Enabled && c.config.Enumerator != nil
}

func (c *ListEnumeratorComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// ListContentComponent handles the main item content
type ListContentComponent struct {
	config ListContentConfig
}

func NewListContentComponent(config ListContentConfig) *ListContentComponent {
	return &ListContentComponent{config: config}
}

func (c *ListContentComponent) Render(ctx ListComponentContext) string {
	var content string

	if c.config.Formatter != nil {
		content = c.config.Formatter(
			ctx.Item,
			ctx.Index,
			ctx.RenderContext,
			ctx.IsCursor,
			ctx.IsThreshold,
			ctx.IsThreshold, // Using same for both top/bottom for simplicity
		)
	} else {
		// Default content formatting
		content = FormatItemContent(
			ctx.Item,
			ctx.Index,
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
			for _, compType := range []ListComponentType{ListComponentCursor, ListComponentPreSpacing, ListComponentEnumerator} {
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

func (c *ListContentComponent) GetType() ListComponentType {
	return ListComponentContent
}

func (c *ListContentComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *ListContentComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// ListPostSpacingComponent handles spacing after the main content
type ListPostSpacingComponent struct {
	config ListSpacingConfig
}

func NewListPostSpacingComponent(config ListSpacingConfig) *ListPostSpacingComponent {
	return &ListPostSpacingComponent{config: config}
}

func (c *ListPostSpacingComponent) Render(ctx ListComponentContext) string {
	return c.config.Style.Render(c.config.Spacing)
}

func (c *ListPostSpacingComponent) GetType() ListComponentType {
	return ListComponentPostSpacing
}

func (c *ListPostSpacingComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *ListPostSpacingComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// ListBackgroundComponent handles background styling as a post-process
type ListBackgroundComponent struct {
	config ListBackgroundConfig
}

func NewListBackgroundComponent(config ListBackgroundConfig) *ListBackgroundComponent {
	return &ListBackgroundComponent{config: config}
}

func (c *ListBackgroundComponent) Render(ctx ListComponentContext) string {
	switch c.config.Mode {
	case ListBackgroundEntireLine:
		// Apply background to the entire combined content
		var parts []string
		for _, compType := range ctx.ListConfig.ComponentOrder {
			if compType == ListComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				parts = append(parts, content)
			}
		}
		fullContent := strings.Join(parts, "")
		return c.config.Style.Render(fullContent)

	case ListBackgroundSelectiveComponents:
		// Apply background only to specified components
		var result strings.Builder

		// Ensure we have valid component data before proceeding
		hasValidData := false
		for _, compType := range ctx.ListConfig.ComponentOrder {
			if compType == ListComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				hasValidData = true
				break
			}
		}

		// If no valid component data, fall back to normal rendering
		if !hasValidData {
			for _, compType := range ctx.ListConfig.ComponentOrder {
				if compType == ListComponentBackground {
					continue
				}
				if content, exists := ctx.ComponentData[compType]; exists {
					result.WriteString(content)
				}
			}
			return result.String()
		}

		// Apply selective background styling
		for _, compType := range ctx.ListConfig.ComponentOrder {
			if compType == ListComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				shouldStyle := false
				for _, targetType := range c.config.ApplyToComponents {
					if compType == targetType {
						shouldStyle = true
						break
					}
				}

				if shouldStyle {
					// Ensure the style is properly applied without interfering with other components
					styledContent := c.config.Style.Render(content)
					result.WriteString(styledContent)
				} else {
					result.WriteString(content)
				}
			}
		}
		return result.String()

	case ListBackgroundContentOnly:
		// Apply background only to content
		var result strings.Builder
		for _, compType := range ctx.ListConfig.ComponentOrder {
			if compType == ListComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				if compType == ListComponentContent {
					result.WriteString(c.config.Style.Render(content))
				} else {
					result.WriteString(content)
				}
			}
		}
		return result.String()

	case ListBackgroundIndicatorOnly:
		// Apply background only to cursor indicator
		var result strings.Builder
		for _, compType := range ctx.ListConfig.ComponentOrder {
			if compType == ListComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				if compType == ListComponentCursor {
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
		for _, compType := range ctx.ListConfig.ComponentOrder {
			if compType == ListComponentBackground {
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

func (c *ListBackgroundComponent) GetType() ListComponentType {
	return ListComponentBackground
}

func (c *ListBackgroundComponent) IsEnabled() bool {
	return c.config.Enabled
}

func (c *ListBackgroundComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// ================================
// LIST COMPONENT RENDERER
// ================================

// ListComponentRenderer orchestrates the rendering of all list components
type ListComponentRenderer struct {
	components map[ListComponentType]ListRenderComponent
	config     ListRenderConfig
}

func NewListComponentRenderer(config ListRenderConfig) *ListComponentRenderer {
	renderer := &ListComponentRenderer{
		components: make(map[ListComponentType]ListRenderComponent),
		config:     config,
	}

	// Create components based on config
	renderer.components[ListComponentCursor] = NewListCursorComponent(config.CursorConfig)
	renderer.components[ListComponentPreSpacing] = NewListPreSpacingComponent(config.PreSpacingConfig)
	renderer.components[ListComponentEnumerator] = NewListEnumeratorComponent(config.EnumeratorConfig)
	renderer.components[ListComponentContent] = NewListContentComponent(config.ContentConfig)
	renderer.components[ListComponentPostSpacing] = NewListPostSpacingComponent(config.PostSpacingConfig)
	renderer.components[ListComponentBackground] = NewListBackgroundComponent(config.BackgroundConfig)

	return renderer
}

// SetComponentOrder sets the rendering order of components
func (r *ListComponentRenderer) SetComponentOrder(order []ListComponentType) {
	r.config.ComponentOrder = order
}

// EnableComponent enables a specific component
func (r *ListComponentRenderer) EnableComponent(componentType ListComponentType) {
	if comp, exists := r.components[componentType]; exists {
		comp.SetEnabled(true)
	}
}

// DisableComponent disables a specific component
func (r *ListComponentRenderer) DisableComponent(componentType ListComponentType) {
	if comp, exists := r.components[componentType]; exists {
		comp.SetEnabled(false)
	}
}

// GetComponent returns a component by type
func (r *ListComponentRenderer) GetComponent(componentType ListComponentType) ListRenderComponent {
	return r.components[componentType]
}

// UpdateConfig updates the renderer configuration
func (r *ListComponentRenderer) UpdateConfig(config ListRenderConfig) {
	r.config = config
	// Recreate components with new config
	r.components[ListComponentCursor] = NewListCursorComponent(config.CursorConfig)
	r.components[ListComponentPreSpacing] = NewListPreSpacingComponent(config.PreSpacingConfig)
	r.components[ListComponentEnumerator] = NewListEnumeratorComponent(config.EnumeratorConfig)
	r.components[ListComponentContent] = NewListContentComponent(config.ContentConfig)
	r.components[ListComponentPostSpacing] = NewListPostSpacingComponent(config.PostSpacingConfig)
	r.components[ListComponentBackground] = NewListBackgroundComponent(config.BackgroundConfig)
}

// Render renders all components in the specified order
func (r *ListComponentRenderer) Render(
	item Data[any],
	index int,
	renderContext RenderContext,
	isCursor, isTopThreshold, isBottomThreshold bool,
) string {
	ctx := ListComponentContext{
		Item:          item,
		Index:         index,
		IsCursor:      isCursor,
		IsSelected:    item.Selected,
		IsThreshold:   isTopThreshold || isBottomThreshold,
		RenderContext: renderContext,
		ComponentData: make(map[ListComponentType]string),
		ListConfig:    r.config,
	}

	// First pass: render all non-background components
	for _, compType := range r.config.ComponentOrder {
		if compType == ListComponentBackground {
			continue // Handle background separately
		}

		if comp, exists := r.components[compType]; exists && comp.IsEnabled() {
			content := comp.Render(ctx)
			ctx.ComponentData[compType] = content
		}
	}

	// Check if we have a background component and apply it if enabled
	backgroundComp := r.components[ListComponentBackground]
	if backgroundComp != nil && backgroundComp.IsEnabled() {
		// Apply background styling - but only for cursor items to maintain expected behavior
		if isCursor {
			return backgroundComp.Render(ctx)
		}
	}

	// No background styling - combine components normally
	var result strings.Builder
	for _, compType := range r.config.ComponentOrder {
		if compType == ListComponentBackground {
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

// DefaultListRenderConfig returns sensible defaults for component-based list rendering
func DefaultListRenderConfig() ListRenderConfig {
	return ListRenderConfig{
		ComponentOrder: []ListComponentType{
			ListComponentCursor,
			ListComponentEnumerator,
			ListComponentContent,
		},
		CursorConfig: ListCursorConfig{
			Enabled:         true,
			CursorIndicator: "► ",
			NormalSpacing:   "  ",
			Style:           lipgloss.NewStyle(),
		},
		PreSpacingConfig: ListSpacingConfig{
			Enabled: false,
			Spacing: "",
			Style:   lipgloss.NewStyle(),
		},
		EnumeratorConfig: ListEnumeratorConfig{
			Enabled:    true,
			Enumerator: BulletEnumerator,
			Style:      lipgloss.NewStyle(),
			Alignment:  ListAlignmentNone,
			MaxWidth:   0,
		},
		ContentConfig: ListContentConfig{
			Enabled:   true,
			Formatter: nil,
			Style:     lipgloss.NewStyle(),
			WrapText:  false,
			MaxWidth:  80,
		},
		PostSpacingConfig: ListSpacingConfig{
			Enabled: false,
			Spacing: "",
			Style:   lipgloss.NewStyle(),
		},
		BackgroundConfig: ListBackgroundConfig{
			Enabled:           false,
			Style:             lipgloss.NewStyle(),
			ApplyToComponents: []ListComponentType{ListComponentCursor, ListComponentEnumerator, ListComponentContent},
			Mode:              ListBackgroundEntireLine,
		},
	}
}

// BulletListConfig creates a config for bullet lists
func BulletListConfig() ListRenderConfig {
	config := DefaultListRenderConfig()
	config.EnumeratorConfig.Enumerator = BulletEnumerator
	return config
}

// NumberedListConfig creates a config for numbered lists
func NumberedListConfig() ListRenderConfig {
	config := DefaultListRenderConfig()
	config.EnumeratorConfig.Enumerator = ArabicEnumerator
	config.EnumeratorConfig.Alignment = ListAlignmentRight
	config.EnumeratorConfig.MaxWidth = 8
	return config
}

// ChecklistConfig creates a config for checklist-style lists
func ChecklistConfig() ListRenderConfig {
	config := DefaultListRenderConfig()
	config.EnumeratorConfig.Enumerator = CheckboxEnumerator
	return config
}

// MinimalListConfig creates a config with just content (no indicators)
func MinimalListConfig() ListRenderConfig {
	config := DefaultListRenderConfig()
	config.ComponentOrder = []ListComponentType{ListComponentContent}
	config.CursorConfig.Enabled = false
	config.EnumeratorConfig.Enabled = false
	return config
}

// CustomOrderListConfig creates a config with custom component order
func CustomOrderListConfig(order []ListComponentType) ListRenderConfig {
	config := DefaultListRenderConfig()
	config.ComponentOrder = order
	return config
}

// BackgroundStyledListConfig creates a config with background styling
func BackgroundStyledListConfig(style lipgloss.Style, mode ListBackgroundMode) ListRenderConfig {
	config := DefaultListRenderConfig()
	config.BackgroundConfig.Enabled = true
	config.BackgroundConfig.Style = style
	config.BackgroundConfig.Mode = mode
	return config
}

// ================================
// INTEGRATION WITH EXISTING SYSTEM
// ================================

// ComponentBasedListFormatter creates an ItemFormatter that uses the component system
func ComponentBasedListFormatter(config ListRenderConfig) ItemFormatter[any] {
	renderer := NewListComponentRenderer(config)
	return func(
		item Data[any],
		index int,
		ctx RenderContext,
		isCursor, isTopThreshold, isBottomThreshold bool,
	) string {
		return renderer.Render(item, index, ctx, isCursor, isTopThreshold, isBottomThreshold)
	}
}

// EnhancedListFormatter creates an ItemFormatter using the component system
// This is the main entry point for list rendering
func EnhancedListFormatter(config ListRenderConfig) ItemFormatter[any] {
	return ComponentBasedListFormatter(config)
}

// Legacy type aliases for backward compatibility
type BackgroundStylingMode = ListBackgroundMode
type BackgroundStyler func(cursorIndicator, enumerator, content string, isCursor bool, style lipgloss.Style) string

const (
	BackgroundStyleEntireLine     = ListBackgroundEntireLine
	BackgroundStyleContentOnly    = ListBackgroundContentOnly
	BackgroundStyleIndicatorOnly  = ListBackgroundIndicatorOnly
	BackgroundStyleWithEnumerator = ListBackgroundSelectiveComponents
	BackgroundStyleCustom         = ListBackgroundSelectiveComponents
)
