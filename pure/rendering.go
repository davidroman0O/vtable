package vtable

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// ================================
// RENDERING FUNCTIONS
// ================================

// RenderEmptyState renders the empty state with appropriate styling
func RenderEmptyState(styleConfig StyleConfig, lastError error) string {
	style := styleConfig.DefaultStyle
	if lastError != nil {
		style = styleConfig.ErrorStyle
		return style.Render("Error: " + lastError.Error())
	}
	return style.Render("No items")
}

// ApplyItemStyle applies the appropriate style to an item based on its state
func ApplyItemStyle(content string, isCursor, isSelected bool, item Data[any], styleConfig StyleConfig, maxWidth int, truncateFunc func(string, int) string) string {
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

// RenderLoadingPlaceholder renders a loading placeholder with appropriate styling
func RenderLoadingPlaceholder(styleConfig StyleConfig) string {
	return styleConfig.LoadingStyle.Render("Loading...")
}

// FormatItemContent formats item content using the provided formatter or default formatting
func FormatItemContent(
	item Data[any],
	absoluteIndex int,
	renderContext RenderContext,
	isCursor, isTopThreshold, isBottomThreshold bool,
	formatter ItemFormatter[any],
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

	// Default formatting
	return fmt.Sprintf("%v", item.Item)
}

// FormatAnimatedItemContent formats item content using animated formatter
func FormatAnimatedItemContent(
	item Data[any],
	absoluteIndex int,
	renderContext RenderContext,
	animationState map[string]any,
	isCursor, isTopThreshold, isBottomThreshold bool,
	animatedFormatter ItemFormatterAnimated[any],
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

// CalculateThresholdFlags calculates threshold flags for an item
func CalculateThresholdFlags(absoluteIndex int, viewport ViewportState) (isTopThreshold, isBottomThreshold bool) {
	isCursor := absoluteIndex == viewport.CursorIndex
	isTopThreshold = isCursor && viewport.IsAtTopThreshold
	isBottomThreshold = isCursor && viewport.IsAtBottomThreshold
	return isTopThreshold, isBottomThreshold
}
