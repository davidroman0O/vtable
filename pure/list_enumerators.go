package vtable

import (
	"fmt"
	"strings"
)

// ================================
// LIST ENUMERATORS
// ================================
// Inspired by lipgloss/list but adapted for our vtable List component

// ListEnumerator enumerates list items. Given the item data, index, and context,
// it returns the prefix that should be displayed for the current item.
type ListEnumerator func(item Data[any], index int, ctx RenderContext) string

// ================================
// PREDEFINED ENUMERATORS
// ================================

// BulletEnumerator creates bullet points for list items
func BulletEnumerator(item Data[any], index int, ctx RenderContext) string {
	return "• "
}

// DashEnumerator creates dash points for list items
func DashEnumerator(item Data[any], index int, ctx RenderContext) string {
	return "- "
}

// AsteriskEnumerator creates asterisk points for list items
func AsteriskEnumerator(item Data[any], index int, ctx RenderContext) string {
	return "* "
}

// ArabicEnumerator creates numbered list items (1. 2. 3.)
func ArabicEnumerator(item Data[any], index int, ctx RenderContext) string {
	return fmt.Sprintf("%d. ", index+1)
}

// AlphabetEnumerator creates alphabetical list items (a. b. c.)
func AlphabetEnumerator(item Data[any], index int, ctx RenderContext) string {
	const abcLen = 26

	if index >= abcLen*abcLen+abcLen {
		return fmt.Sprintf("%c%c%c. ", 'a'+index/abcLen/abcLen-1, 'a'+(index/abcLen)%abcLen-1, 'a'+index%abcLen)
	}
	if index >= abcLen {
		return fmt.Sprintf("%c%c. ", 'a'+index/abcLen-1, 'a'+(index)%abcLen)
	}
	return fmt.Sprintf("%c. ", 'a'+index%abcLen)
}

// RomanEnumerator creates roman numeral list items (i. ii. iii.)
func RomanEnumerator(item Data[any], index int, ctx RenderContext) string {
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

// CheckboxEnumerator creates checkbox-style list items
func CheckboxEnumerator(item Data[any], index int, ctx RenderContext) string {
	if item.Selected {
		return "☑ "
	}
	return "☐ "
}

// ArrowEnumerator creates arrow-style list items
func ArrowEnumerator(item Data[any], index int, ctx RenderContext) string {
	return "→ "
}

// CustomEnumerator allows for custom enumeration patterns
func CustomEnumerator(pattern string) ListEnumerator {
	return func(item Data[any], index int, ctx RenderContext) string {
		// Replace placeholders in pattern
		result := pattern
		result = strings.ReplaceAll(result, "{index}", fmt.Sprintf("%d", index))
		result = strings.ReplaceAll(result, "{index1}", fmt.Sprintf("%d", index+1))
		result = strings.ReplaceAll(result, "{id}", item.ID)
		return result
	}
}

// ================================
// CONDITIONAL ENUMERATORS
// ================================

// ConditionalEnumerator allows different enumerators based on conditions
type ConditionalEnumerator struct {
	conditions []EnumeratorCondition
	fallback   ListEnumerator
}

// EnumeratorCondition represents a condition and its associated enumerator
type EnumeratorCondition struct {
	Condition func(item Data[any], index int, ctx RenderContext) bool
	Enum      ListEnumerator
}

// NewConditionalEnumerator creates a new conditional enumerator
func NewConditionalEnumerator(fallback ListEnumerator) *ConditionalEnumerator {
	return &ConditionalEnumerator{
		fallback: fallback,
	}
}

// When adds a condition and enumerator
func (ce *ConditionalEnumerator) When(condition func(item Data[any], index int, ctx RenderContext) bool, enum ListEnumerator) *ConditionalEnumerator {
	ce.conditions = append(ce.conditions, EnumeratorCondition{
		Condition: condition,
		Enum:      enum,
	})
	return ce
}

// Enumerate executes the conditional enumeration
func (ce *ConditionalEnumerator) Enumerate(item Data[any], index int, ctx RenderContext) string {
	for _, cond := range ce.conditions {
		if cond.Condition(item, index, ctx) {
			return cond.Enum(item, index, ctx)
		}
	}
	return ce.fallback(item, index, ctx)
}

// ================================
// ENUMERATOR UTILITIES
// ================================

// PaddedEnumerator wraps an enumerator to ensure consistent width
func PaddedEnumerator(enum ListEnumerator, width int) ListEnumerator {
	return func(item Data[any], index int, ctx RenderContext) string {
		prefix := enum(item, index, ctx)
		if len(prefix) < width {
			prefix = prefix + strings.Repeat(" ", width-len(prefix))
		}
		return prefix
	}
}

// StyledEnumerator wraps an enumerator with styling
func StyledEnumerator(enum ListEnumerator, styleFunc func(item Data[any], index int, ctx RenderContext) string) ListEnumerator {
	return func(item Data[any], index int, ctx RenderContext) string {
		prefix := enum(item, index, ctx)
		style := styleFunc(item, index, ctx)
		// Apply style to prefix (this would use lipgloss styling)
		return style + prefix
	}
}

// ================================
// COMMON CONDITION HELPERS
// ================================

// IsSelected returns true if the item is selected
func IsSelected(item Data[any], index int, ctx RenderContext) bool {
	return item.Selected
}

// IsError returns true if the item has an error
func IsError(item Data[any], index int, ctx RenderContext) bool {
	return item.Error != nil
}

// IsLoading returns true if the item is loading
func IsLoading(item Data[any], index int, ctx RenderContext) bool {
	return item.Loading
}

// IsDisabled returns true if the item is disabled
func IsDisabled(item Data[any], index int, ctx RenderContext) bool {
	return item.Disabled
}

// IsEven returns true if the index is even
func IsEven(item Data[any], index int, ctx RenderContext) bool {
	return index%2 == 0
}

// IsOdd returns true if the index is odd
func IsOdd(item Data[any], index int, ctx RenderContext) bool {
	return index%2 == 1
}

// IsCursor returns true if this is the cursor position
func IsCursor(item Data[any], index int, ctx RenderContext) bool {
	// This would need to be passed in the context
	return false // Placeholder - would need viewport state
}

// ================================
// TREE-SPECIFIC ENUMERATORS
// ================================

// TreeEnumerator creates tree-style enumeration with proper indentation and tree symbols
func TreeEnumerator(item Data[any], index int, ctx RenderContext) string {
	// Type assert to check if this is a tree item
	if flatItem, ok := item.Item.(interface {
		GetDepth() int
		HasChildren() bool
		IsExpanded() bool
	}); ok {
		var prefix strings.Builder

		// Add indentation based on depth
		depth := flatItem.GetDepth()
		for i := 0; i < depth; i++ {
			prefix.WriteString("  ")
		}

		// Add tree connector
		if flatItem.HasChildren() {
			if flatItem.IsExpanded() {
				prefix.WriteString("▼ ")
			} else {
				prefix.WriteString("▶ ")
			}
		} else {
			prefix.WriteString("• ")
		}

		return prefix.String()
	}

	// Fallback to bullet for non-tree items
	return "• "
}

// TreeExpandedEnumerator shows different symbols for expanded/collapsed nodes
func TreeExpandedEnumerator(item Data[any], index int, ctx RenderContext) string {
	if flatItem, ok := item.Item.(interface {
		GetDepth() int
		HasChildren() bool
		IsExpanded() bool
	}); ok {
		var prefix strings.Builder

		// Add indentation
		depth := flatItem.GetDepth()
		for i := 0; i < depth; i++ {
			prefix.WriteString("│ ")
		}

		// Add tree connector with box drawing characters
		if flatItem.HasChildren() {
			if flatItem.IsExpanded() {
				prefix.WriteString("├─")
			} else {
				prefix.WriteString("├+")
			}
		} else {
			prefix.WriteString("└─")
		}

		return prefix.String()
	}

	return "• "
}

// TreeMinimalEnumerator provides minimal tree visualization
func TreeMinimalEnumerator(item Data[any], index int, ctx RenderContext) string {
	if flatItem, ok := item.Item.(interface {
		GetDepth() int
		HasChildren() bool
		IsExpanded() bool
	}); ok {
		var prefix strings.Builder

		// Add simple indentation
		depth := flatItem.GetDepth()
		for i := 0; i < depth; i++ {
			prefix.WriteString("  ")
		}

		// Simple symbols
		if flatItem.HasChildren() {
			if flatItem.IsExpanded() {
				prefix.WriteString("- ")
			} else {
				prefix.WriteString("+ ")
			}
		} else {
			prefix.WriteString("  ")
		}

		return prefix.String()
	}

	return ""
}
