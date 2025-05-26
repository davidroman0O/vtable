package vtable

import (
	"strings"
	"time"
	"unicode"
)

// SimpleScrollConfig contains configuration for simple horizontal scrolling
type SimpleScrollConfig struct {
	Speed          float64       // Characters per second (default: 5.0)
	WordAware      bool          // Stop at word boundaries (default: true)
	PauseDuration  time.Duration // Pause at edges before bouncing (default: 500ms)
	MinScrollWidth int           // Don't scroll if text is only slightly longer (default: 3)
}

// DefaultScrollConfig returns a sensible default configuration
func DefaultScrollConfig() SimpleScrollConfig {
	return SimpleScrollConfig{
		Speed:          2.0, // Smooth and pleasant - 2 characters per second
		WordAware:      true,
		PauseDuration:  800 * time.Millisecond, // Reasonable pause
		MinScrollWidth: 3,
	}
}

// ScrollDirection represents the direction of scrolling
type ScrollDirection int

const (
	ScrollRight ScrollDirection = 1
	ScrollLeft  ScrollDirection = -1
)

// SimpleScrollState holds the state for simple horizontal scrolling
type SimpleScrollState struct {
	Position      float64         // Current scroll position (in characters)
	Direction     ScrollDirection // Current scroll direction
	PauseUntil    time.Time       // Pause until this time
	LastUpdate    time.Time       // Last update timestamp
	TextLength    int             // Length of the text being scrolled
	MaxPosition   float64         // Maximum scroll position
	WordPositions []int           // Positions of word boundaries (if word-aware)
}

// CreateSimpleHorizontalScrolling creates a simple bouncing horizontal scroll
func CreateSimpleHorizontalScrolling(
	text string,
	maxWidth int,
	config SimpleScrollConfig,
	state map[string]any,
	deltaTime time.Duration,
) (string, map[string]any) {
	// Handle empty or short text
	if text == "" || maxWidth <= 0 {
		return padToWidth("", maxWidth), state
	}

	// Calculate actual text width
	textWidth := properDisplayWidth(text)

	// If text fits within maxWidth, just return it padded
	if textWidth <= maxWidth || textWidth-maxWidth <= config.MinScrollWidth {
		return padToWidth(text, maxWidth), state
	}

	// Get or initialize scroll state
	scrollState := getOrInitScrollState(state, text, maxWidth, config)

	// Update scroll state
	now := time.Now()
	if !scrollState.LastUpdate.IsZero() && deltaTime > 0 {
		updateScrollState(scrollState, config, deltaTime, now)
	}
	scrollState.LastUpdate = now

	// Extract the visible window of text
	visibleText := extractVisibleWindow(text, scrollState, maxWidth, config)

	// Update state map
	newState := make(map[string]any)
	for k, v := range state {
		newState[k] = v
	}
	newState["scroll_state"] = scrollState

	return visibleText, newState
}

// getOrInitScrollState gets existing scroll state or initializes a new one
func getOrInitScrollState(state map[string]any, text string, maxWidth int, config SimpleScrollConfig) *SimpleScrollState {
	if stateData, exists := state["scroll_state"]; exists {
		if scrollState, ok := stateData.(*SimpleScrollState); ok {
			// Check if text changed - if so, reinitialize
			textWidth := properDisplayWidth(text)
			if scrollState.TextLength != textWidth {
				return initScrollState(text, maxWidth, config)
			}
			return scrollState
		}
	}

	return initScrollState(text, maxWidth, config)
}

// initScrollState initializes a new scroll state
func initScrollState(text string, maxWidth int, config SimpleScrollConfig) *SimpleScrollState {
	textWidth := properDisplayWidth(text)
	maxPosition := float64(textWidth - maxWidth)
	if maxPosition < 0 {
		maxPosition = 0
	}

	scrollState := &SimpleScrollState{
		Position:    0.0,
		Direction:   ScrollRight,
		TextLength:  textWidth,
		MaxPosition: maxPosition,
		LastUpdate:  time.Now(),
	}

	// Calculate word positions if word-aware
	if config.WordAware {
		scrollState.WordPositions = findWordBoundaries(text)
	}

	return scrollState
}

// updateScrollState updates the scroll position based on time and configuration
func updateScrollState(scrollState *SimpleScrollState, config SimpleScrollConfig, deltaTime time.Duration, now time.Time) {
	// Check if we're in a pause
	if now.Before(scrollState.PauseUntil) {
		return // Still pausing
	}

	// Calculate movement distance based purely on speed and deltaTime
	secondsElapsed := deltaTime.Seconds()
	movement := config.Speed * secondsElapsed

	// Update position based on direction
	oldPosition := scrollState.Position
	if scrollState.Direction == ScrollRight {
		scrollState.Position += movement
	} else {
		scrollState.Position -= movement
	}

	// Check for bouncing at edges
	bounced := false

	if scrollState.Position <= 0 {
		scrollState.Position = 0
		if scrollState.Direction == ScrollLeft {
			scrollState.Direction = ScrollRight
			scrollState.PauseUntil = now.Add(config.PauseDuration)
			bounced = true
		}
	}

	if scrollState.Position >= scrollState.MaxPosition {
		scrollState.Position = scrollState.MaxPosition
		if scrollState.Direction == ScrollRight {
			scrollState.Direction = ScrollLeft
			scrollState.PauseUntil = now.Add(config.PauseDuration)
			bounced = true
		}
	}

	// Apply word-aware positioning if enabled and we're not bouncing
	if config.WordAware && !bounced && len(scrollState.WordPositions) > 0 {
		scrollState.Position = snapToNearestWord(scrollState.Position, oldPosition, scrollState.WordPositions, scrollState.Direction)
	}
}

// findWordBoundaries finds positions of word boundaries in the text
func findWordBoundaries(text string) []int {
	var positions []int
	runes := []rune(text)

	positions = append(positions, 0) // Start position

	for i, r := range runes {
		if unicode.IsSpace(r) {
			// Add position after space as word boundary
			if i+1 < len(runes) {
				positions = append(positions, i+1)
			}
		}
	}

	if len(positions) == 0 || positions[len(positions)-1] != len(runes) {
		positions = append(positions, len(runes)) // End position
	}

	return positions
}

// snapToNearestWord snaps the position to the nearest word boundary in the direction of movement
// Only snaps if the movement is significant enough (prevents ultra-slow speeds from jumping)
func snapToNearestWord(newPos, oldPos float64, wordPositions []int, direction ScrollDirection) float64 {
	if len(wordPositions) <= 1 {
		return newPos
	}

	// Only snap to word boundaries if we've moved at least 0.5 characters
	// This prevents ultra-slow speeds from causing unwanted jumps
	if abs(newPos-oldPos) < 0.5 {
		return newPos
	}

	// Find the closest word boundary in the direction of movement
	intPos := int(newPos)

	if direction == ScrollRight {
		// Moving right - find next word boundary
		for _, wordPos := range wordPositions {
			if wordPos > intPos {
				return float64(wordPos)
			}
		}
	} else {
		// Moving left - find previous word boundary
		for i := len(wordPositions) - 1; i >= 0; i-- {
			wordPos := wordPositions[i]
			if wordPos < intPos {
				return float64(wordPos)
			}
		}
	}

	return newPos // No suitable word boundary found
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// extractVisibleWindow extracts the visible portion of text based on scroll state
func extractVisibleWindow(text string, scrollState *SimpleScrollState, maxWidth int, config SimpleScrollConfig) string {
	if maxWidth <= 0 {
		return ""
	}

	runes := []rune(text)
	textLen := len(runes)
	startPos := int(scrollState.Position)

	// Ensure we don't go out of bounds
	if startPos < 0 {
		startPos = 0
	}
	if startPos >= textLen {
		startPos = textLen - 1
	}

	// Extract characters to fill the maxWidth
	var result strings.Builder
	pos := startPos
	targetWidth := maxWidth

	for targetWidth > 0 && pos < textLen {
		char := runes[pos]
		charWidth := properDisplayWidth(string(char))

		if charWidth > targetWidth {
			// Character is wider than remaining space
			break
		}

		result.WriteRune(char)
		targetWidth -= charWidth
		pos++
	}

	// If we still have space and we're at the end, we might need to wrap around
	// or just pad with spaces to ensure consistent width
	resultStr := result.String()
	currentWidth := properDisplayWidth(resultStr)

	if currentWidth < maxWidth {
		// Pad with spaces to maintain consistent width
		padding := maxWidth - currentWidth
		resultStr += strings.Repeat(" ", padding)
	}

	return resultStr
}

// padToWidth pads text to the specified width
func padToWidth(text string, width int) string {
	if width <= 0 {
		return ""
	}

	currentWidth := properDisplayWidth(text)
	if currentWidth >= width {
		// Truncate if too long
		runes := []rune(text)
		result := ""
		for _, r := range runes {
			candidate := result + string(r)
			if properDisplayWidth(candidate) <= width {
				result = candidate
			} else {
				break
			}
		}
		// Pad to exact width
		resultWidth := properDisplayWidth(result)
		if resultWidth < width {
			result += strings.Repeat(" ", width-resultWidth)
		}
		return result
	}

	// Pad with spaces
	padding := width - currentWidth
	return text + strings.Repeat(" ", padding)
}

// Convenience functions for easy usage

// CreateBounceScrollCell creates a simple bouncing scroll cell with default config
func CreateBounceScrollCell(text string, maxWidth int, state map[string]any, deltaTime time.Duration) (string, map[string]any) {
	config := DefaultScrollConfig()
	return CreateSimpleHorizontalScrolling(text, maxWidth, config, state, deltaTime)
}

// CreateFastScrollCell creates a faster scrolling cell
func CreateFastScrollCell(text string, maxWidth int, state map[string]any, deltaTime time.Duration) (string, map[string]any) {
	config := DefaultScrollConfig()
	config.Speed = 4.0 // Fast but readable
	return CreateSimpleHorizontalScrolling(text, maxWidth, config, state, deltaTime)
}

// CreateSlowScrollCell creates a slower scrolling cell
func CreateSlowScrollCell(text string, maxWidth int, state map[string]any, deltaTime time.Duration) (string, map[string]any) {
	config := DefaultScrollConfig()
	config.Speed = 1.0 // Slow but not painful
	return CreateSimpleHorizontalScrolling(text, maxWidth, config, state, deltaTime)
}
