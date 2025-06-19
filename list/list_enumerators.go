package list

import (
	"fmt"
	"strings"

	"github.com/davidroman0O/vtable/core"
)

// This file provides a collection of `ListEnumerator` functions and helpers for
// the List component. Enumerators are responsible for generating the prefix for
// each list item, such as bullets, numbers, or checkboxes. This allows for easy
// customization of list styles. The file includes standard enumerators, as well
// as utilities for creating conditional, padded, and styled enumerators.

// Package list provides a feature-rich, data-virtualized list component for
// Bubble Tea applications. It is designed for performance and flexibility,
// capable of handling very large datasets by loading data in chunks as needed.
// The list supports various item styles, selection modes, configurable keymaps,
// and a component-based rendering pipeline for easy customization.

// BulletEnumerator is a `ListEnumerator` that creates a classic bullet point ("• ")
// for each list item. It is a simple and common style for unordered lists.
func BulletEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	return "• "
}

// DashEnumerator is a `ListEnumerator` that creates a dash ("- ") for each list item.
// It serves as an alternative style for unordered lists.
func DashEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	return "- "
}

// AsteriskEnumerator is a `ListEnumerator` that creates an asterisk ("* ") for
// each list item, another common style for unordered lists.
func AsteriskEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	return "* "
}

// ArabicEnumerator is a `ListEnumerator` that creates a numbered list using
// Arabic numerals (e.g., "1. ", "2. ", "3. ").
func ArabicEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	return fmt.Sprintf("%d. ", index+1)
}

// AlphabetEnumerator is a `ListEnumerator` that creates an alphabetical list
// (e.g., "a. ", "b. ", "z. ", "aa. "). It supports single, double, and triple
// character representations for large lists.
func AlphabetEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	const abcLen = 26

	if index >= abcLen*abcLen+abcLen {
		return fmt.Sprintf("%c%c%c. ", 'a'+index/abcLen/abcLen-1, 'a'+(index/abcLen)%abcLen-1, 'a'+index%abcLen)
	}
	if index >= abcLen {
		return fmt.Sprintf("%c%c. ", 'a'+index/abcLen-1, 'a'+(index)%abcLen)
	}
	return fmt.Sprintf("%c. ", 'a'+index%abcLen)
}

// RomanEnumerator is a `ListEnumerator` that creates a Roman numeral list
// (e.g., "i. ", "ii. ", "x. "). It converts the item's index to its lowercase
// Roman numeral representation.
func RomanEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	var (
		roman  = []string{"m", "cm", "d", "cd", "c", "xc", "l", "xl", "x", "ix", "v", "iv", "i"}
		arabic = []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
		result strings.Builder
	)

	num := index + 1
	for v, value := range arabic {
		for num >= value {
			num -= value
			result.WriteString(roman[v])
		}
	}
	result.WriteString(". ")
	return result.String()
}

// CheckboxEnumerator is a `ListEnumerator` that creates a checkbox-style list.
// It displays a checked box ("☑ ") for selected items and an unchecked box
// ("☐ ") for unselected items, based on the `item.Selected` field.
func CheckboxEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	if item.Selected {
		return "☑ "
	}
	return "☐ "
}

// ArrowEnumerator is a `ListEnumerator` that creates an arrow ("→ ") for each
// list item, which can be used to indicate focus or direction.
func ArrowEnumerator(item core.Data[any], index int, ctx core.RenderContext) string {
	return "→ "
}

// CustomEnumerator creates a `ListEnumerator` from a pattern string. This allows
// for flexible, user-defined enumerators. The following placeholders are replaced:
// - {index}: The 0-based index of the item.
// - {index1}: The 1-based index of the item.
// - {id}: The unique ID of the item.
func CustomEnumerator(pattern string) core.ListEnumerator {
	return func(item core.Data[any], index int, ctx core.RenderContext) string {
		// Replace placeholders in pattern
		result := pattern
		result = strings.ReplaceAll(result, "{index}", fmt.Sprintf("%d", index))
		result = strings.ReplaceAll(result, "{index1}", fmt.Sprintf("%d", index+1))
		result = strings.ReplaceAll(result, "{id}", item.ID)
		return result
	}
}

// ConditionalEnumerator allows for using different enumerators based on a set
// of dynamic conditions. It evaluates conditions in the order they are added
// and uses the enumerator from the first matching condition. A fallback
// enumerator is used if no conditions match, making it highly versatile.
type ConditionalEnumerator struct {
	conditions []EnumeratorCondition
	fallback   core.ListEnumerator
}

// EnumeratorCondition holds a predicate function and the `ListEnumerator` to use
// if the condition returns true. It is the building block for the
// `ConditionalEnumerator`.
type EnumeratorCondition struct {
	Condition func(item core.Data[any], index int, ctx core.RenderContext) bool
	Enum      core.ListEnumerator
}

// NewConditionalEnumerator creates a new `ConditionalEnumerator` with a required
// fallback enumerator, which is used when no other conditions are met.
func NewConditionalEnumerator(fallback core.ListEnumerator) *ConditionalEnumerator {
	return &ConditionalEnumerator{
		fallback: fallback,
	}
}

// When adds a new condition and its corresponding enumerator to the chain.
func (ce *ConditionalEnumerator) When(condition func(item core.Data[any], index int, ctx core.RenderContext) bool, enum core.ListEnumerator) *ConditionalEnumerator {
	ce.conditions = append(ce.conditions, EnumeratorCondition{
		Condition: condition,
		Enum:      enum,
	})
	return ce
}

// Enumerate evaluates the conditions in order and returns the string from the
// first matching enumerator. If no conditions match, it uses the fallback.
func (ce *ConditionalEnumerator) Enumerate(item core.Data[any], index int, ctx core.RenderContext) string {
	for _, cond := range ce.conditions {
		if cond.Condition(item, index, ctx) {
			return cond.Enum(item, index, ctx)
		}
	}
	return ce.fallback(item, index, ctx)
}

// PaddedEnumerator wraps another `ListEnumerator` to ensure its output is padded
// with spaces to a minimum width. This is useful for creating visually aligned
// lists, especially with numbered enumerators of varying digit lengths.
func PaddedEnumerator(enum core.ListEnumerator, width int) core.ListEnumerator {
	return func(item core.Data[any], index int, ctx core.RenderContext) string {
		prefix := enum(item, index, ctx)
		if len(prefix) < width {
			prefix = prefix + strings.Repeat(" ", width-len(prefix))
		}
		return prefix
	}
}

// StyledEnumerator wraps another `ListEnumerator`, applying a dynamically
// generated style string to its output. The style is determined by the
// `styleFunc`, allowing for context-aware styling (e.g., based on selection or
// cursor state). Note: The `styleFunc` should return terminal style sequences.
func StyledEnumerator(enum core.ListEnumerator, styleFunc func(item core.Data[any], index int, ctx core.RenderContext) string) core.ListEnumerator {
	return func(item core.Data[any], index int, ctx core.RenderContext) string {
		prefix := enum(item, index, ctx)
		style := styleFunc(item, index, ctx)
		// Apply style to prefix (this would use lipgloss styling)
		return style + prefix
	}
}

// IsSelected is a condition function for `ConditionalEnumerator` that returns
// true if the item is marked as selected.
func IsSelected(item core.Data[any], index int, ctx core.RenderContext) bool {
	return item.Selected
}

// IsError is a condition function for `ConditionalEnumerator` that returns true
// if the item has a non-nil error.
func IsError(item core.Data[any], index int, ctx core.RenderContext) bool {
	return item.Error != nil
}

// IsLoading is a condition function for `ConditionalEnumerator` that returns
// true if the item is in a loading state.
func IsLoading(item core.Data[any], index int, ctx core.RenderContext) bool {
	return item.Loading
}

// IsDisabled is a condition function for `ConditionalEnumerator` that returns
// true if the item is disabled.
func IsDisabled(item core.Data[any], index int, ctx core.RenderContext) bool {
	return item.Disabled
}

// IsEven is a condition function for `ConditionalEnumerator` that returns true
// if the item's index is even. Useful for creating striped lists.
func IsEven(item core.Data[any], index int, ctx core.RenderContext) bool {
	return index%2 == 0
}

// IsOdd is a condition function for `ConditionalEnumerator` that returns true if
// the item's index is odd. Useful for creating striped lists.
func IsOdd(item core.Data[any], index int, ctx core.RenderContext) bool {
	return index%2 == 1
}

// IsCursor is a placeholder condition function that demonstrates how a condition
// could check for the cursor state. In a real implementation, this requires the
// viewport state to be passed via the `RenderContext` to determine if the item
// at `index` is currently under the cursor.
func IsCursor(item core.Data[any], index int, ctx core.RenderContext) bool {
	// This would need to be passed in the context
	return false // Placeholder - would need viewport state
}
