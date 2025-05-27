package vtable

import (
	"fmt"
	"strings"

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

	// Enhanced default formatting for common types
	var content string
	switch v := item.Item.(type) {
	case string:
		content = v
	case fmt.Stringer:
		content = v.String()
	default:
		// Check if it's a struct with common fields we can format nicely
		if taskContent := tryFormatAsTask(v); taskContent != "" {
			content = taskContent
		} else {
			// Fallback to default formatting
			content = fmt.Sprintf("%v", item.Item)
		}
	}

	// Add state indicators - show both selection and error/loading/disabled states
	var stateIndicator string

	// Add error/loading/disabled indicators
	switch {
	case item.Error != nil:
		stateIndicator += " ❌"
	case item.Loading:
		stateIndicator += " ⏳"
	case item.Disabled:
		stateIndicator += " 🚫"
	}

	// Add selection indicator if selected
	if item.Selected {
		stateIndicator += " ✅"
	}

	return content + stateIndicator
}

// tryFormatAsTask attempts to format an item as a task-like structure
func tryFormatAsTask(item any) string {
	// Use reflection to check for common task fields
	// This is a simple approach that works for our Task struct
	if taskMap, ok := item.(map[string]interface{}); ok {
		// Handle map-based tasks
		if title, hasTitle := taskMap["Title"].(string); hasTitle {
			if priority, hasPriority := taskMap["Priority"].(string); hasPriority {
				if status, hasStatus := taskMap["Status"].(string); hasStatus {
					if category, hasCategory := taskMap["Category"].(string); hasCategory {
						return fmt.Sprintf("%s | %s | %s | %s", title, priority, status, category)
					}
				}
			}
		}
	}

	// For our specific Task struct, we can use a type assertion
	// This is a bit hacky but works for the example
	taskStr := fmt.Sprintf("%+v", item)
	if strings.Contains(taskStr, "Title:") && strings.Contains(taskStr, "Priority:") {
		// Parse the struct string to extract fields
		// This is not ideal but works for the demo
		fields := strings.Fields(taskStr)
		var title, priority, status, category string

		for _, field := range fields {
			if strings.HasPrefix(field, "Title:") {
				title = strings.TrimPrefix(field, "Title:")
			} else if strings.HasPrefix(field, "Priority:") {
				priority = strings.TrimPrefix(field, "Priority:")
			} else if strings.HasPrefix(field, "Status:") {
				status = strings.TrimPrefix(field, "Status:")
			} else if strings.HasPrefix(field, "Category:") {
				category = strings.TrimPrefix(field, "Category:")
			}
		}

		if title != "" && priority != "" && status != "" && category != "" {
			return fmt.Sprintf("%s | %s | %s | %s", title, priority, status, category)
		}
	}

	return "" // Couldn't format as task
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
