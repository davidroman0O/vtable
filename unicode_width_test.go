package vtable

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// TestEmojiWidthCalculations tests width calculations for the specific emoji cases
// that are causing table deformation in animations
func TestEmojiWidthCalculations(t *testing.T) {
	level := "WARN"

	// The two cases from the animation that have different widths
	case1 := "⚠️ " + level // "⚠️ WARN"
	case2 := "🟡 " + level  // "🟡 WARN"

	// Test lipgloss.Width vs go-runewidth
	lipglossWidth1 := lipgloss.Width(case1)
	lipglossWidth2 := lipgloss.Width(case2)

	runeWidth1 := runewidth.StringWidth(case1)
	runeWidth2 := runewidth.StringWidth(case2)

	t.Logf("Case 1: '%s'", case1)
	t.Logf("  lipgloss.Width: %d", lipglossWidth1)
	t.Logf("  runewidth.StringWidth: %d", runeWidth1)

	t.Logf("Case 2: '%s'", case2)
	t.Logf("  lipgloss.Width: %d", lipglossWidth2)
	t.Logf("  runewidth.StringWidth: %d", runeWidth2)

	// Check if lipgloss gives different results
	t.Logf("Lipgloss width difference: %d", lipglossWidth1-lipglossWidth2)
	t.Logf("Runewidth width difference: %d", runeWidth1-runeWidth2)

	// Test other emoji cases
	testCases := []string{
		"✅ SUCCESS",
		"❌ ERROR",
		"⚠️ WARNING",
		"🟡 PENDING",
		"🔄 LOADING",
		"📊 DATA",
		"🚀 ROCKET",
		"💰 MONEY",
		"🎯 TARGET",
	}

	t.Log("\nComparing width calculations for various emoji strings:")
	for _, testCase := range testCases {
		lipWidth := lipgloss.Width(testCase)
		runeWidth := runewidth.StringWidth(testCase)

		t.Logf("'%s': lipgloss=%d, runewidth=%d, diff=%d",
			testCase, lipWidth, runeWidth, lipWidth-runeWidth)
	}
}

// TestCJKCharacterWidths tests width calculations for CJK characters
func TestCJKCharacterWidths(t *testing.T) {
	testCases := []string{
		"你好世界",   // Chinese
		"こんにちは",  // Japanese Hiragana
		"カタカナ",   // Japanese Katakana
		"안녕하세요",  // Korean
		"中文测试",   // Mixed Chinese
		"日本語テスト", // Mixed Japanese
	}

	t.Log("Comparing width calculations for CJK characters:")
	for _, testCase := range testCases {
		lipWidth := lipgloss.Width(testCase)
		runeWidth := runewidth.StringWidth(testCase)

		t.Logf("'%s': lipgloss=%d, runewidth=%d, diff=%d",
			testCase, lipWidth, runeWidth, lipWidth-runeWidth)
	}
}

// TestProperWidthVsLipgloss compares our proper width function with lipgloss
func TestProperWidthVsLipgloss(t *testing.T) {
	testCases := []string{
		"simple text",
		"⚠️ WARN",
		"🟡 WARN",
		"✅ SUCCESS",
		"❌ FAILED",
		"🔄 Loading...",
		"你好",
		"こんにちは",
		"mixed text with 🎯 emoji and 中文",
		"ANSI \033[31mred\033[0m text",
	}

	t.Log("Comparing properDisplayWidth (go-runewidth) vs lipgloss.Width:")
	for _, testCase := range testCases {
		properWidth := properDisplayWidth(testCase)
		lipWidth := lipgloss.Width(testCase)

		equal := ""
		if properWidth != lipWidth {
			equal = " ❌ DIFFERENT"
		} else {
			equal = " ✅ SAME"
		}

		t.Logf("'%s': proper=%d, lipgloss=%d%s",
			testCase, properWidth, lipWidth, equal)
	}
}

// TestEmojiConstraintConsistency tests that the specific emoji cases from animations
// are properly constrained to the same width
func TestEmojiConstraintConsistency(t *testing.T) {
	level := "WARN"
	targetWidth := 8

	// The two cases from the animation that should both fit in the same width
	case1 := "⚠️ " + level // "⚠️ WARN" - runewidth 6
	case2 := "🟡 " + level  // "🟡 WARN" - runewidth 7

	constraint := CellConstraint{
		Width:     targetWidth,
		Height:    1,
		Alignment: AlignLeft,
	}

	result1 := enforceCellConstraints(case1, constraint)
	result2 := enforceCellConstraints(case2, constraint)

	actualWidth1 := properDisplayWidth(result1)
	actualWidth2 := properDisplayWidth(result2)

	t.Logf("Input case 1: '%s' (width: %d)", case1, properDisplayWidth(case1))
	t.Logf("Output case 1: '%s' (width: %d)", result1, actualWidth1)
	t.Logf("Input case 2: '%s' (width: %d)", case2, properDisplayWidth(case2))
	t.Logf("Output case 2: '%s' (width: %d)", result2, actualWidth2)

	// Both results should have the same width (the target width)
	if actualWidth1 != targetWidth {
		t.Errorf("Case 1 width mismatch: expected %d, got %d", targetWidth, actualWidth1)
	}
	if actualWidth2 != targetWidth {
		t.Errorf("Case 2 width mismatch: expected %d, got %d", targetWidth, actualWidth2)
	}
	if actualWidth1 != actualWidth2 {
		t.Errorf("Width inconsistency: case1=%d, case2=%d", actualWidth1, actualWidth2)
	}

	// Test various widths to ensure consistency
	testWidths := []int{6, 7, 8, 10, 12}
	for _, width := range testWidths {
		constraint.Width = width

		result1 := enforceCellConstraints(case1, constraint)
		result2 := enforceCellConstraints(case2, constraint)

		width1 := properDisplayWidth(result1)
		width2 := properDisplayWidth(result2)

		if width1 != width2 {
			t.Errorf("Width %d: inconsistent results - case1=%d, case2=%d", width, width1, width2)
		}
		if width1 != width {
			t.Errorf("Width %d: case1 wrong width - expected %d, got %d", width, width, width1)
		}
		if width2 != width {
			t.Errorf("Width %d: case2 wrong width - expected %d, got %d", width, width, width2)
		}

		t.Logf("Width %d: Both cases properly constrained to %d characters", width, width1)
	}
}
