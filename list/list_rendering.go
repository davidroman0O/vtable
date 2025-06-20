// Package list provides a feature-rich, data-virtualized list component for
// Bubble Tea applications. It is designed for performance and flexibility,
// capable of handling very large datasets by loading data in chunks as needed.
// The list supports various item styles, selection modes, configurable keymaps,
// and a component-based rendering pipeline for easy customization.
package list

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
	"github.com/davidroman0O/vtable/render"
)

// ListCursorComponent is a render component responsible for displaying the cursor
// indicator. It shows a specific string (`CursorIndicator`) when the item is
// under the cursor and a different string (`NormalSpacing`) otherwise, ensuring
// proper alignment.
type ListCursorComponent struct {
	config core.ListCursorConfig
}

// NewListCursorComponent creates a new cursor component with the given configuration.
func NewListCursorComponent(config core.ListCursorConfig) *ListCursorComponent {
	return &ListCursorComponent{config: config}
}

// Render returns the cursor indicator string based on the context.
func (c *ListCursorComponent) Render(ctx core.ListComponentContext) string {
	var content string
	if ctx.IsCursor {
		content = c.config.CursorIndicator
	} else {
		content = c.config.NormalSpacing
	}

	// Apply component styling first
	styledContent := c.config.Style.Render(content)

	// Apply background styling based on state while preserving text styling
	if ctx.IsCursor && c.config.ApplyCursorBg {
		// Use width-aware background styling that works with complex content
		bgStyle := c.config.CursorBackground.Copy().Width(lipgloss.Width(styledContent))
		return bgStyle.Render(styledContent)
	} else if ctx.IsSelected && c.config.ApplySelectedBg {
		// Use width-aware background styling that works with complex content
		bgStyle := c.config.SelectedBackground.Copy().Width(lipgloss.Width(styledContent))
		return bgStyle.Render(styledContent)
	} else if c.config.ApplyNormalBg {
		// Use width-aware background styling that works with complex content
		bgStyle := c.config.NormalBackground.Copy().Width(lipgloss.Width(styledContent))
		return bgStyle.Render(styledContent)
	}

	return styledContent
}

// GetType returns the unique type identifier for this component.
func (c *ListCursorComponent) GetType() core.ListComponentType {
	return core.ListComponentCursor
}

// IsEnabled checks if the component is configured to be rendered.
func (c *ListCursorComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *ListCursorComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// ListPreSpacingComponent is a render component that adds a fixed-width space
// before the main content, useful for creating indentation or alignment.
type ListPreSpacingComponent struct {
	config core.ListSpacingConfig
}

// NewListPreSpacingComponent creates a new spacing component.
func NewListPreSpacingComponent(config core.ListSpacingConfig) *ListPreSpacingComponent {
	return &ListPreSpacingComponent{config: config}
}

// Render returns the configured spacing string.
func (c *ListPreSpacingComponent) Render(ctx core.ListComponentContext) string {
	return c.config.Style.Render(c.config.Spacing)
}

// GetType returns the unique type identifier for this component.
func (c *ListPreSpacingComponent) GetType() core.ListComponentType {
	return core.ListComponentPreSpacing
}

// IsEnabled checks if the component is configured to be rendered.
func (c *ListPreSpacingComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *ListPreSpacingComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// ListEnumeratorComponent is a render component that displays an enumerator, such
// as a bullet point, number, or checkbox, next to the list item. The appearance
// is determined by a provided `ListEnumerator` function.
type ListEnumeratorComponent struct {
	config core.ListEnumeratorConfig
}

// NewListEnumeratorComponent creates a new enumerator component.
func NewListEnumeratorComponent(config core.ListEnumeratorConfig) *ListEnumeratorComponent {
	return &ListEnumeratorComponent{config: config}
}

// Render generates the enumerator string using the configured function and
// applies any specified alignment.
func (c *ListEnumeratorComponent) Render(ctx core.ListComponentContext) string {
	if c.config.Enumerator == nil {
		return ""
	}

	enumText := c.config.Enumerator(ctx.Item, ctx.Index, ctx.RenderContext)

	// Apply alignment if configured
	if c.config.Alignment != core.ListAlignmentNone && c.config.MaxWidth > 0 {
		switch c.config.Alignment {
		case core.ListAlignmentRight:
			padding := c.config.MaxWidth - len(enumText)
			if padding > 0 {
				enumText = strings.Repeat(" ", padding) + enumText
			}
		case core.ListAlignmentLeft:
			padding := c.config.MaxWidth - len(enumText)
			if padding > 0 {
				enumText = enumText + strings.Repeat(" ", padding)
			}
		}
	}

	// Apply component styling first
	styledContent := c.config.Style.Render(enumText)

	// Apply background styling based on state
	if ctx.IsCursor && c.config.ApplyCursorBg {
		return c.config.CursorBackground.Render(styledContent)
	} else if ctx.IsSelected && c.config.ApplySelectedBg {
		return c.config.SelectedBackground.Render(styledContent)
	} else if c.config.ApplyNormalBg {
		return c.config.NormalBackground.Render(styledContent)
	}

	return styledContent
}

// GetType returns the unique type identifier for this component.
func (c *ListEnumeratorComponent) GetType() core.ListComponentType {
	return core.ListComponentEnumerator
}

// IsEnabled checks if the component is configured to be rendered.
func (c *ListEnumeratorComponent) IsEnabled() bool {
	return c.config.Enabled && c.config.Enumerator != nil
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *ListEnumeratorComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// ListContentComponent is the render component responsible for displaying the
// main content of the list item. It uses a provided formatter or a default
// formatting logic and can handle text wrapping.
type ListContentComponent struct {
	config core.ListContentConfig
}

// NewListContentComponent creates a new content component.
func NewListContentComponent(config core.ListContentConfig) *ListContentComponent {
	return &ListContentComponent{config: config}
}

// Render generates the item's main content string. If text wrapping is enabled,
// it will wrap the text and indent subsequent lines to align with the start of
// the content.
func (c *ListContentComponent) Render(ctx core.ListComponentContext) string {
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
		content = render.FormatItemContent(
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
			for _, compType := range []core.ListComponentType{core.ListComponentCursor, core.ListComponentPreSpacing, core.ListComponentEnumerator} {
				if compContent, exists := ctx.ComponentData[compType]; exists {
					indentSize += len(compContent)
				}
			}
			content = strings.Join(lines, "\n"+strings.Repeat(" ", indentSize))
		} else if len(lines) == 1 {
			content = lines[0]
		}
	}

	// Apply component styling first
	styledContent := c.config.Style.Render(content)

	// Apply background styling based on state with more aggressive approach
	if ctx.IsCursor && c.config.ApplyCursorBg {
		// Strip all existing styling and apply background
		plainContent := stripAnsiCodes(styledContent)
		return c.config.CursorBackground.Render(plainContent)
	} else if ctx.IsSelected && c.config.ApplySelectedBg {
		// Strip all existing styling and apply background
		plainContent := stripAnsiCodes(styledContent)
		return c.config.SelectedBackground.Render(plainContent)
	} else if c.config.ApplyNormalBg {
		// Strip all existing styling and apply background
		plainContent := stripAnsiCodes(styledContent)
		return c.config.NormalBackground.Render(plainContent)
	}

	return styledContent
}

// stripAnsiCodes removes ANSI escape sequences from a string to get plain text
func stripAnsiCodes(s string) string {
	// Simple regex to remove ANSI escape sequences
	// This is a basic implementation - might need improvement for complex cases
	result := ""
	inEscape := false

	for _, r := range s {
		if r == '\033' { // ESC character
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' { // End of escape sequence
				inEscape = false
			}
			continue
		}
		result += string(r)
	}
	return result
}

// GetType returns the unique type identifier for this component.
func (c *ListContentComponent) GetType() core.ListComponentType {
	return core.ListComponentContent
}

// IsEnabled checks if the component is configured to be rendered.
func (c *ListContentComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *ListContentComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// ListPostSpacingComponent is a render component that adds a fixed-width space
// after the main content.
type ListPostSpacingComponent struct {
	config core.ListSpacingConfig
}

// NewListPostSpacingComponent creates a new spacing component.
func NewListPostSpacingComponent(config core.ListSpacingConfig) *ListPostSpacingComponent {
	return &ListPostSpacingComponent{config: config}
}

// Render returns the configured spacing string.
func (c *ListPostSpacingComponent) Render(ctx core.ListComponentContext) string {
	return c.config.Style.Render(c.config.Spacing)
}

// GetType returns the unique type identifier for this component.
func (c *ListPostSpacingComponent) GetType() core.ListComponentType {
	return core.ListComponentPostSpacing
}

// IsEnabled checks if the component is configured to be rendered.
func (c *ListPostSpacingComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *ListPostSpacingComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// ListBackgroundComponent is a special render component that applies a background
// style as a post-processing step. It can apply the style to the entire line or
// to a select subset of the other rendered components.
type ListBackgroundComponent struct {
	config core.ListBackgroundConfig
}

// NewListBackgroundComponent creates a new background component.
func NewListBackgroundComponent(config core.ListBackgroundConfig) *ListBackgroundComponent {
	return &ListBackgroundComponent{config: config}
}

// Render applies the background style according to the configured mode. It does
// not return its own content but rather modifies the combined output of other
// components.
func (c *ListBackgroundComponent) Render(ctx core.ListComponentContext) string {
	switch c.config.Mode {
	case core.ListBackgroundEntireLine:
		// Apply background to the entire combined content
		var parts []string
		for _, compType := range ctx.ListConfig.ComponentOrder {
			if compType == core.ListComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				parts = append(parts, content)
			}
		}
		fullContent := strings.Join(parts, "")
		return c.config.Style.Render(fullContent)

	case core.ListBackgroundSelectiveComponents:
		// Apply background only to specified components
		var result strings.Builder

		// Ensure we have valid component data before proceeding
		hasValidData := false
		for _, compType := range ctx.ListConfig.ComponentOrder {
			if compType == core.ListComponentBackground {
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
				if compType == core.ListComponentBackground {
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
			if compType == core.ListComponentBackground {
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

	case core.ListBackgroundContentOnly:
		// Apply background only to content
		var result strings.Builder
		for _, compType := range ctx.ListConfig.ComponentOrder {
			if compType == core.ListComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				if compType == core.ListComponentContent {
					result.WriteString(c.config.Style.Render(content))
				} else {
					result.WriteString(content)
				}
			}
		}
		return result.String()

	case core.ListBackgroundIndicatorOnly:
		// Apply background only to cursor indicator
		var result strings.Builder
		for _, compType := range ctx.ListConfig.ComponentOrder {
			if compType == core.ListComponentBackground {
				continue
			}
			if content, exists := ctx.ComponentData[compType]; exists && content != "" {
				if compType == core.ListComponentCursor {
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
			if compType == core.ListComponentBackground {
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
func (c *ListBackgroundComponent) GetType() core.ListComponentType {
	return core.ListComponentBackground
}

// IsEnabled checks if the component is configured to be rendered.
func (c *ListBackgroundComponent) IsEnabled() bool {
	return c.config.Enabled
}

// SetEnabled allows enabling or disabling this component at runtime.
func (c *ListBackgroundComponent) SetEnabled(enabled bool) {
	c.config.Enabled = enabled
}

// ListComponentRenderer manages the rendering pipeline for a list item. It holds
// all the individual render components and processes them in a defined order to
// construct the final string for a single list item.
type ListComponentRenderer struct {
	components map[core.ListComponentType]core.ListRenderComponent
	config     core.ListRenderConfig
}

// NewListComponentRenderer creates a new renderer with a given configuration,
// initializing all the necessary components.
func NewListComponentRenderer(config core.ListRenderConfig) *ListComponentRenderer {
	renderer := &ListComponentRenderer{
		components: make(map[core.ListComponentType]core.ListRenderComponent),
		config:     config,
	}

	// Create components based on config
	renderer.components[core.ListComponentCursor] = NewListCursorComponent(config.CursorConfig)
	renderer.components[core.ListComponentPreSpacing] = NewListPreSpacingComponent(config.PreSpacingConfig)
	renderer.components[core.ListComponentEnumerator] = NewListEnumeratorComponent(config.EnumeratorConfig)
	renderer.components[core.ListComponentContent] = NewListContentComponent(config.ContentConfig)
	renderer.components[core.ListComponentPostSpacing] = NewListPostSpacingComponent(config.PostSpacingConfig)
	renderer.components[core.ListComponentBackground] = NewListBackgroundComponent(config.BackgroundConfig)

	return renderer
}

// SetComponentOrder defines the sequence in which the components are rendered.
func (r *ListComponentRenderer) SetComponentOrder(order []core.ListComponentType) {
	r.config.ComponentOrder = order
}

// EnableComponent activates a specific component in the rendering pipeline.
func (r *ListComponentRenderer) EnableComponent(componentType core.ListComponentType) {
	if comp, exists := r.components[componentType]; exists {
		comp.SetEnabled(true)
	}
}

// DisableComponent deactivates a specific component.
func (r *ListComponentRenderer) DisableComponent(componentType core.ListComponentType) {
	if comp, exists := r.components[componentType]; exists {
		comp.SetEnabled(false)
	}
}

// GetComponent retrieves a specific component from the renderer, allowing for
// direct configuration changes.
func (r *ListComponentRenderer) GetComponent(componentType core.ListComponentType) core.ListRenderComponent {
	return r.components[componentType]
}

// UpdateConfig applies a new configuration to the renderer, recreating all
// its internal components to reflect the changes.
func (r *ListComponentRenderer) UpdateConfig(config core.ListRenderConfig) {
	r.config = config
	// Re-initialize components with the new config
	r.components[core.ListComponentCursor] = NewListCursorComponent(config.CursorConfig)
	r.components[core.ListComponentPreSpacing] = NewListPreSpacingComponent(config.PreSpacingConfig)
	r.components[core.ListComponentEnumerator] = NewListEnumeratorComponent(config.EnumeratorConfig)
	r.components[core.ListComponentContent] = NewListContentComponent(config.ContentConfig)
	r.components[core.ListComponentPostSpacing] = NewListPostSpacingComponent(config.PostSpacingConfig)
	r.components[core.ListComponentBackground] = NewListBackgroundComponent(config.BackgroundConfig)
}

// Render executes the full rendering pipeline for a single list item. It iterates
// through the components in the configured order, calls their `Render` methods,
// and assembles the final string. It handles special cases like the background
// component, which modifies the output of other components.
func (r *ListComponentRenderer) Render(
	item core.Data[any],
	index int,
	renderContext core.RenderContext,
	isCursor, isTopThreshold, isBottomThreshold bool,
) string {
	ctx := core.ListComponentContext{
		Item:          item,
		Index:         index,
		IsCursor:      isCursor,
		IsSelected:    item.Selected,
		IsThreshold:   isTopThreshold || isBottomThreshold,
		RenderContext: renderContext,
		ComponentData: make(map[core.ListComponentType]string),
		ListConfig:    r.config,
	}

	// First pass: render all non-background components
	for _, compType := range r.config.ComponentOrder {
		if compType == core.ListComponentBackground {
			continue // Handle background separately
		}

		if comp, exists := r.components[compType]; exists && comp.IsEnabled() {
			content := comp.Render(ctx)
			ctx.ComponentData[compType] = content
		}
	}

	// Check if we have a background component and apply it if enabled
	backgroundComp := r.components[core.ListComponentBackground]
	if backgroundComp != nil && backgroundComp.IsEnabled() {
		// Apply background styling for cursor items OR selected items
		if isCursor || ctx.IsSelected {
			return backgroundComp.Render(ctx)
		}
	}

	// No background styling - combine components normally
	var result strings.Builder
	for _, compType := range r.config.ComponentOrder {
		if compType == core.ListComponentBackground {
			continue
		}
		if content, exists := ctx.ComponentData[compType]; exists {
			result.WriteString(content)
		}
	}
	return result.String()
}

// BulletListConfig provides a pre-configured `ListRenderConfig` for a classic
// bullet-point list style.
func BulletListConfig() core.ListRenderConfig {
	cfg := config.DefaultListRenderConfig()
	cfg.EnumeratorConfig.Enumerator = BulletEnumerator
	return cfg
}

// NumberedListConfig provides a pre-configured `ListRenderConfig` for a numbered
// list style (e.g., "1.", "2.", ...).
func NumberedListConfig() core.ListRenderConfig {
	cfg := config.DefaultListRenderConfig()
	cfg.EnumeratorConfig.Enumerator = ArabicEnumerator
	cfg.EnumeratorConfig.Alignment = core.ListAlignmentRight
	cfg.EnumeratorConfig.MaxWidth = 8
	return cfg
}

// ChecklistConfig provides a pre-configured `ListRenderConfig` for a checklist
// style, showing "[x]" for selected items and "[ ]" for others.
func ChecklistConfig() core.ListRenderConfig {
	cfg := config.DefaultListRenderConfig()
	cfg.EnumeratorConfig.Enumerator = CheckboxEnumerator
	return cfg
}

// MinimalListConfig provides a pre-configured `ListRenderConfig` with only the
// cursor and content components enabled, for a clean, simple look.
func MinimalListConfig() core.ListRenderConfig {
	cfg := config.DefaultListRenderConfig()
	cfg.ComponentOrder = []core.ListComponentType{core.ListComponentCursor, core.ListComponentContent}
	cfg.EnumeratorConfig.Enabled = false
	return cfg
}

// CustomOrderListConfig creates a `ListRenderConfig` with a user-defined
// component rendering order.
func CustomOrderListConfig(order []core.ListComponentType) core.ListRenderConfig {
	cfg := config.DefaultListRenderConfig()
	cfg.ComponentOrder = order
	return cfg
}

// BackgroundStyledListConfig creates a `ListRenderConfig` that applies a
// background style to list items. The mode determines whether the style applies
// to the entire line or just specific components.
func BackgroundStyledListConfig(style lipgloss.Style, mode core.ListBackgroundMode) core.ListRenderConfig {
	cfg := config.DefaultListRenderConfig()
	cfg.ComponentOrder = append(cfg.ComponentOrder, core.ListComponentBackground)
	cfg.BackgroundConfig.Enabled = true
	cfg.BackgroundConfig.Style = style
	cfg.BackgroundConfig.Mode = mode
	return cfg
}

// ComponentBasedListFormatter creates an `ItemFormatter` function from a
// `ListRenderConfig`. This allows the component-based rendering pipeline to be
// used as a standard formatter.
func ComponentBasedListFormatter(config core.ListRenderConfig) core.ItemFormatter[any] {
	renderer := NewListComponentRenderer(config)
	return func(item core.Data[any], index int, renderContext core.RenderContext, isCursor, isTopThreshold, isBottomThreshold bool) string {
		return renderer.Render(
			item,
			index,
			renderContext,
			isCursor,
			isTopThreshold,
			isBottomThreshold,
		)
	}
}

// EnhancedListFormatter is an alias for `ComponentBasedListFormatter` for clarity.
func EnhancedListFormatter(config core.ListRenderConfig) core.ItemFormatter[any] {
	return ComponentBasedListFormatter(config)
}

// BackgroundStylingMode is an alias for `core.ListBackgroundMode` for convenience.
type BackgroundStylingMode = core.ListBackgroundMode

// BackgroundStyler defines a function type for custom background styling logic.
type BackgroundStyler func(cursorIndicator, enumerator, content string, isCursor bool, style lipgloss.Style) string

const (
	// BackgroundStyleEntireLine applies the background to the entire rendered line.
	BackgroundStyleEntireLine = core.ListBackgroundEntireLine
	// BackgroundStyleContentOnly applies the background only to the main content component.
	BackgroundStyleContentOnly = core.ListBackgroundContentOnly
	// BackgroundStyleIndicatorOnly applies the background only to the cursor indicator component.
	BackgroundStyleIndicatorOnly = core.ListBackgroundIndicatorOnly
	// BackgroundStyleWithEnumerator applies the background to both the content and enumerator.
	BackgroundStyleWithEnumerator = core.ListBackgroundSelectiveComponents
	// BackgroundStyleCustom allows applying the background to a custom set of components.
	BackgroundStyleCustom = core.ListBackgroundSelectiveComponents
)
