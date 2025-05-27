package vtable

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ================================
// ENHANCED LIST RENDERING
// ================================
// Advanced rendering options inspired by lipgloss but integrated with our system

// ListRenderConfig contains configuration for enhanced list rendering
type ListRenderConfig struct {
	Enumerator      ListEnumerator
	ShowEnumerator  bool
	IndentSize      int
	ItemSpacing     int
	MaxWidth        int
	WrapText        bool
	AlignEnumerator bool // Whether to right-align enumerators for consistent spacing
	CursorIndicator string
	NormalSpacing   string
}

// DefaultListRenderConfig returns sensible defaults for list rendering
func DefaultListRenderConfig() ListRenderConfig {
	return ListRenderConfig{
		Enumerator:      BulletEnumerator,
		ShowEnumerator:  true,
		IndentSize:      2,
		ItemSpacing:     0,
		MaxWidth:        80,
		WrapText:        true,
		AlignEnumerator: true,
		CursorIndicator: "► ",
		NormalSpacing:   "  ",
	}
}

// EnhancedListFormatter creates an ItemFormatter that uses enumerators and advanced styling
func EnhancedListFormatter(config ListRenderConfig) ItemFormatter[any] {
	return func(
		data Data[any],
		index int,
		ctx RenderContext,
		isCursor bool,
		isTopThreshold bool,
		isBottomThreshold bool,
	) string {
		var parts []string

		// Add cursor indicator if this is the cursor line
		if isCursor {
			parts = append(parts, config.CursorIndicator)
		} else {
			parts = append(parts, config.NormalSpacing) // Add spacing for alignment
		}

		// Generate enumerator if enabled
		if config.ShowEnumerator && config.Enumerator != nil {
			enum := config.Enumerator(data, index, ctx)
			if enum != "" {
				parts = append(parts, enum)
			}
		}

		// Format the main content
		content := FormatItemContent(data, index, ctx, isCursor, isTopThreshold, isBottomThreshold, nil)

		// Handle text wrapping if enabled
		if config.WrapText && config.MaxWidth > 0 {
			if ctx.Wrap != nil {
				lines := ctx.Wrap(content, config.MaxWidth)
				if len(lines) > 1 {
					// Multi-line content - handle indentation
					content = strings.Join(lines, "\n"+strings.Repeat(" ", config.IndentSize))
				} else if len(lines) == 1 {
					content = lines[0]
				}
			}
		}

		parts = append(parts, content)

		result := strings.Join(parts, "")

		// Apply spacing if configured
		if config.ItemSpacing > 0 {
			result += strings.Repeat("\n", config.ItemSpacing)
		}

		return result
	}
}

// CalculateEnumeratorWidth calculates the maximum width needed for enumerators
func CalculateEnumeratorWidth(items []Data[any], enum ListEnumerator, ctx RenderContext) int {
	maxWidth := 0
	for i, item := range items {
		if enum != nil {
			enumText := enum(item, i, ctx)
			if len(enumText) > maxWidth {
				maxWidth = len(enumText)
			}
		}
	}
	return maxWidth
}

// AlignedEnumeratorFormatter creates a formatter with aligned enumerators
func AlignedEnumeratorFormatter(config ListRenderConfig, items []Data[any]) ItemFormatter[any] {
	// Calculate the maximum enumerator width for alignment
	maxEnumWidth := 0
	if config.AlignEnumerator && config.ShowEnumerator && config.Enumerator != nil {
		ctx := RenderContext{
			ErrorIndicator:    "❌",
			LoadingIndicator:  "⏳",
			DisabledIndicator: "🚫",
			SelectedIndicator: "✅",
		} // Basic context for width calculation
		maxEnumWidth = CalculateEnumeratorWidth(items, config.Enumerator, ctx)
	}

	return func(
		data Data[any],
		index int,
		ctx RenderContext,
		isCursor bool,
		isTopThreshold bool,
		isBottomThreshold bool,
	) string {
		var parts []string

		// Generate aligned enumerator if enabled
		if config.ShowEnumerator && config.Enumerator != nil {
			enum := config.Enumerator(data, index, ctx)
			if config.AlignEnumerator && maxEnumWidth > 0 {
				// Right-align the enumerator
				enum = strings.Repeat(" ", maxEnumWidth-len(enum)) + enum
			}
			if enum != "" {
				parts = append(parts, enum)
			}
		}

		// Format the main content
		content := FormatItemContent(data, index, ctx, isCursor, isTopThreshold, isBottomThreshold, nil)
		parts = append(parts, content)

		return strings.Join(parts, "")
	}
}

// ================================
// SPECIALIZED FORMATTERS
// ================================

// ChecklistFormatter creates a checklist-style formatter
func ChecklistFormatter() ItemFormatter[any] {
	config := DefaultListRenderConfig()
	config.Enumerator = CheckboxEnumerator
	return EnhancedListFormatter(config)
}

// NumberedListFormatter creates a numbered list formatter
func NumberedListFormatter() ItemFormatter[any] {
	config := DefaultListRenderConfig()
	config.Enumerator = ArabicEnumerator
	config.AlignEnumerator = true
	return EnhancedListFormatter(config)
}

// BulletListFormatter creates a bullet list formatter
func BulletListFormatter() ItemFormatter[any] {
	config := DefaultListRenderConfig()
	config.Enumerator = BulletEnumerator
	return EnhancedListFormatter(config)
}

// AlphabeticalListFormatter creates an alphabetical list formatter
func AlphabeticalListFormatter() ItemFormatter[any] {
	config := DefaultListRenderConfig()
	config.Enumerator = AlphabetEnumerator
	config.AlignEnumerator = true
	return EnhancedListFormatter(config)
}

// ================================
// CONDITIONAL FORMATTING
// ================================

// ConditionalListFormatter creates a formatter with conditional styling
func ConditionalListFormatter() ItemFormatter[any] {
	// Create a conditional enumerator
	conditionalEnum := NewConditionalEnumerator(BulletEnumerator).
		When(IsSelected, CheckboxEnumerator).
		When(IsError, func(item Data[any], index int, ctx RenderContext) string {
			return ctx.ErrorIndicator + " "
		}).
		When(IsLoading, func(item Data[any], index int, ctx RenderContext) string {
			return ctx.LoadingIndicator + " "
		})

	config := DefaultListRenderConfig()
	config.Enumerator = conditionalEnum.Enumerate

	return EnhancedListFormatter(config)
}

// ================================
// MULTI-LINE SUPPORT
// ================================

// MultiLineListFormatter creates a formatter that handles multi-line content properly
func MultiLineListFormatter(indentSize int) ItemFormatter[any] {
	return func(
		data Data[any],
		index int,
		ctx RenderContext,
		isCursor bool,
		isTopThreshold bool,
		isBottomThreshold bool,
	) string {
		// Generate enumerator
		enum := BulletEnumerator(data, index, ctx)

		// Format the main content
		content := FormatItemContent(data, index, ctx, isCursor, isTopThreshold, isBottomThreshold, nil)

		// Split content into lines
		lines := strings.Split(content, "\n")
		if len(lines) <= 1 {
			// Single line - simple case
			return enum + content
		}

		// Multi-line - indent continuation lines
		result := enum + lines[0]
		indent := strings.Repeat(" ", len(enum)+indentSize)

		for _, line := range lines[1:] {
			result += "\n" + indent + line
		}

		return result
	}
}

// ================================
// STYLE INTEGRATION
// ================================

// StyledListFormatter creates a formatter with lipgloss styling integration
func StyledListFormatter(styleConfig StyleConfig) ItemFormatter[any] {
	return func(
		data Data[any],
		index int,
		ctx RenderContext,
		isCursor bool,
		isTopThreshold bool,
		isBottomThreshold bool,
	) string {
		// Generate enumerator
		enum := BulletEnumerator(data, index, ctx)

		// Format the main content
		content := FormatItemContent(data, index, ctx, isCursor, isTopThreshold, isBottomThreshold, nil)

		// Apply styling based on item state
		var style lipgloss.Style
		switch {
		case data.Error != nil:
			style = styleConfig.ErrorStyle
		case data.Loading:
			style = styleConfig.LoadingStyle
		case data.Disabled:
			style = styleConfig.DisabledStyle
		case isCursor && data.Selected:
			style = styleConfig.CursorStyle.Copy().
				Background(styleConfig.SelectedStyle.GetBackground())
		case isCursor:
			style = styleConfig.CursorStyle
		case data.Selected:
			style = styleConfig.SelectedStyle
		default:
			style = styleConfig.DefaultStyle
		}

		// Combine enumerator and content
		fullContent := enum + content

		return style.Render(fullContent)
	}
}

// ================================
// UTILITY FUNCTIONS
// ================================

// WrapListContent wraps list content to specified width while preserving enumerators
func WrapListContent(content string, enum string, maxWidth int, wrapFunc func(string, int) []string) string {
	if wrapFunc == nil || maxWidth <= 0 {
		return enum + content
	}

	// Calculate available width for content (subtract enumerator width)
	availableWidth := maxWidth - len(enum)
	if availableWidth <= 0 {
		return enum + content
	}

	// Wrap the content
	lines := wrapFunc(content, availableWidth)
	if len(lines) <= 1 {
		if len(lines) == 1 {
			return enum + lines[0]
		}
		return enum + content
	}

	// Multi-line - indent continuation lines
	result := enum + lines[0]
	indent := strings.Repeat(" ", len(enum))

	for _, line := range lines[1:] {
		result += "\n" + indent + line
	}

	return result
}
