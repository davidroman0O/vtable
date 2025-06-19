// Package render provides a collection of utility functions for rendering vtable
// components. It encapsulates common rendering logic, such as applying styles,
// formatting content, and handling different item states (e.g., loading, error,
// selected). This package promotes consistency and simplifies the rendering
// process within individual components like List and Table.
package render

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable/core"
)

// RenderEmptyState generates the string to be displayed when a component has no
// items. It uses the appropriate style from the style configuration, displaying a
// custom error message if an error is present, or a default "No items" message
// otherwise.
func RenderEmptyState(styleConfig core.StyleConfig, lastError error) string {
	style := styleConfig.DefaultStyle
	if lastError != nil {
		style = styleConfig.ErrorStyle
		return style.Render("Error: " + lastError.Error())
	}
	return style.Render("No items")
}

// ApplyItemStyle selects and applies the correct `lipgloss.Style` to a content
// string based on the item's state (e.g., cursor, selected, disabled, error).
// It handles combined states, such as when an item is both under the cursor and
// selected. The function also truncates the content to a maximum width if specified.
func ApplyItemStyle(content string, isCursor, isSelected bool, item core.Data[any], styleConfig core.StyleConfig, maxWidth int, truncateFunc func(string, int) string) string {
	var style lipgloss.Style

	switch {
	case item.Error != nil:
		style = styleConfig.ErrorStyle
	case item.Loading:
		style = styleConfig.LoadingStyle
	case item.Disabled:
		style = styleConfig.DisabledStyle
	case isCursor && isSelected:
		// Combine cursor and selected styles
		style = styleConfig.CursorStyle.Copy().
			Background(styleConfig.SelectedStyle.GetBackground())
	case isCursor:
		style = styleConfig.CursorStyle
	case isSelected:
		style = styleConfig.SelectedStyle
	default:
		style = styleConfig.DefaultStyle
	}

	// Truncate content to max width
	if maxWidth > 0 && len(content) > maxWidth {
		content = truncateFunc(content, maxWidth)
	}

	return style.Render(content)
}

// RenderLoadingPlaceholder generates a styled "Loading..." string. This is a
// convenience function for displaying a consistent loading indicator.
func RenderLoadingPlaceholder(styleConfig core.StyleConfig) string {
	return styleConfig.LoadingStyle.Render("Loading...")
}

// FormatItemContent prepares an item's data for display. If a custom
// `ItemFormatter` is provided, it is used; otherwise, the function applies
// default formatting rules, including intelligent string conversion for common
// types and the addition of configurable state indicators (e.g., for loading,
// errors, or selection) based on the render context.
func FormatItemContent(
	item core.Data[any],
	absoluteIndex int,
	renderContext core.RenderContext,
	isCursor, isTopThreshold, isBottomThreshold bool,
	formatter core.ItemFormatter[any],
) string {
	if formatter != nil {
		return formatter(
			item,
			absoluteIndex,
			renderContext,
			isCursor,
			isTopThreshold,
			isBottomThreshold,
		)
	}

	// Enhanced default formatting for common types
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

// FormatAnimatedItemContent processes an item using a provided animated formatter.
// It passes all necessary state and context to the formatter and returns the
// resulting content string from the `RenderResult`.
func FormatAnimatedItemContent(
	item core.Data[any],
	absoluteIndex int,
	renderContext core.RenderContext,
	animationState map[string]any,
	isCursor, isTopThreshold, isBottomThreshold bool,
	animatedFormatter core.ItemFormatterAnimated[any],
) string {
	result := animatedFormatter(
		item,
		absoluteIndex,
		renderContext,
		animationState,
		isCursor,
		isTopThreshold,
		isBottomThreshold,
	)

	return result.Content
}

// CalculateThresholdFlags determines if an item at a given absolute index is
// currently at the top or bottom scroll threshold. It requires the current
// viewport state to make this determination.
func CalculateThresholdFlags(absoluteIndex int, viewport core.ViewportState) (isTopThreshold, isBottomThreshold bool) {
	isCursor := absoluteIndex == viewport.CursorIndex
	isTopThreshold = isCursor && viewport.IsAtTopThreshold
	isBottomThreshold = isCursor && viewport.IsAtBottomThreshold
	return isTopThreshold, isBottomThreshold
}
