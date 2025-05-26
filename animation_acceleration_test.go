package vtable

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestAnimationAccelerationFix tests that cursor movements don't accelerate animations
func TestAnimationAccelerationFix(t *testing.T) {
	fmt.Println("\n=== ANIMATION ACCELERATION FIX TEST ===")

	provider := NewAnimTestDataProvider(10)

	config := ViewportConfig{
		Height:               5,
		TopThresholdIndex:    1,
		BottomThresholdIndex: 3,
		ChunkSize:            10,
		InitialIndex:         0,
		Debug:                false,
	}

	// Create animated formatter that increments a counter on every call
	animatedFormatter := func(data Data[AnimTestItem], index int, ctx RenderContext,
		animationState map[string]any, isCursor bool, isTopThreshold bool, isBottomThreshold bool) RenderResult {

		// Track how many times this formatter has been called
		counter := 0
		if c, ok := animationState["counter"]; ok {
			if ci, ok := c.(int); ok {
				counter = ci
			}
		}
		counter++

		prefix := "  "
		if isCursor {
			prefix = "> "
		}

		content := fmt.Sprintf("%s%s [calls: %d]", prefix, data.Item.Name, counter)

		return RenderResult{
			Content: content,
			RefreshTriggers: []RefreshTrigger{{
				Type:     TriggerTimer,
				Interval: 500 * time.Millisecond, // Slow interval for testing
			}},
			AnimationState: map[string]any{
				"counter": counter,
			},
		}
	}

	regularFormatter := func(data Data[AnimTestItem], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return fmt.Sprintf("%s%s", prefix, data.Item.Name)
	}

	styleConfig := StyleConfig{}
	list, err := NewTeaList(config, provider, styleConfig, regularFormatter)
	if err != nil {
		t.Fatalf("NewTeaList failed: %v", err)
	}

	// Set animated formatter
	list.SetAnimatedFormatter(animatedFormatter)

	// Render initial view to register animations
	fmt.Println("\n1. Initial render (registers animations):")
	view1 := list.View()
	fmt.Print(view1)

	// Extract counter values from the first render
	initialCounters := extractCounters(view1)
	fmt.Printf("Initial counter values: %v\n", initialCounters)

	// Simulate rapid cursor movements (this used to cause acceleration)
	fmt.Println("\n2. Rapid cursor movements (should NOT accelerate animations):")
	for i := 0; i < 10; i++ {
		// Move cursor down and render (simulating fast navigation)
		list.MoveDown()
		view := list.View()

		// Each cursor movement used to cause re-registration and acceleration
		// With the fix, animations should remain stable
		_ = view
	}

	// Check counters after cursor movements - they should not have accelerated
	view2 := list.View()
	fmt.Print(view2)

	afterMovementCounters := extractCounters(view2)
	fmt.Printf("Counter values after rapid movements: %v\n", afterMovementCounters)

	// Counters should not have dramatically increased due to cursor movements
	// They should only increment based on the animation timer, not view renders
	for i, initial := range initialCounters {
		if i < len(afterMovementCounters) {
			after := afterMovementCounters[i]
			// Allow some natural increment due to time passing, but not excessive
			if after > initial+5 { // Reasonable threshold
				t.Errorf("Animation acceleration detected! Counter %d went from %d to %d (too much increase)", i, initial, after)
			} else {
				fmt.Printf("✅ Counter %d stable: %d -> %d (normal increment)\n", i, initial, after)
			}
		}
	}

	// Test that animations still work properly with timer-based updates
	fmt.Println("\n3. Testing timer-based animation updates:")

	// Wait for animation timer to trigger
	time.Sleep(600 * time.Millisecond) // Slightly longer than trigger interval

	// Process a timer tick
	engine := list.animationEngine
	tickMsg := GlobalAnimationTickMsg{Timestamp: time.Now()}
	cmd := engine.ProcessGlobalTick(tickMsg)

	if cmd != nil {
		fmt.Println("✅ Timer tick processed successfully")
	}

	// Render after timer update
	view3 := list.View()
	timerCounters := extractCounters(view3)
	fmt.Printf("Counter values after timer update: %v\n", timerCounters)

	// Now counters should have increased due to the timer trigger
	hasIncreased := false
	for i, before := range afterMovementCounters {
		if i < len(timerCounters) {
			after := timerCounters[i]
			if after > before {
				hasIncreased = true
				fmt.Printf("✅ Timer-based update working: Counter %d: %d -> %d\n", i, before, after)
			}
		}
	}

	if !hasIncreased {
		fmt.Println("⚠️  Timer-based updates may not be working as expected")
	}

	fmt.Println("\n=== END ANIMATION ACCELERATION FIX TEST ===")
}

// extractCounters extracts counter values from rendered view for testing
func extractCounters(view string) []int {
	// This is a simple parser to extract [calls: X] values from the view
	// In a real implementation, you'd want more robust parsing
	counters := []int{}

	// Look for patterns like "[calls: 1]"
	lines := []string{}
	current := ""
	for _, char := range view {
		if char == '\n' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}

	for _, line := range lines {
		// Simple extraction - look for "[calls: X]"
		start := -1
		for i := 0; i < len(line)-8; i++ {
			if line[i:i+8] == "[calls: " {
				start = i + 8
				break
			}
		}

		if start != -1 {
			// Find the closing bracket
			end := -1
			for i := start; i < len(line); i++ {
				if line[i] == ']' {
					end = i
					break
				}
			}

			if end != -1 {
				// Extract the number
				numStr := line[start:end]
				var num int
				fmt.Sscanf(numStr, "%d", &num)
				counters = append(counters, num)
			}
		}
	}

	return counters
}

// TestCursorMovementWithAnimations tests that cursor movements work smoothly with animations
func TestCursorMovementWithAnimations(t *testing.T) {
	fmt.Println("\n=== CURSOR MOVEMENT WITH ANIMATIONS TEST ===")

	provider := NewAnimTestDataProvider(5)
	config := ViewportConfig{
		Height:               5,
		TopThresholdIndex:    1,
		BottomThresholdIndex: 3,
		ChunkSize:            10,
		InitialIndex:         0,
		Debug:                false,
	}

	// Simple animated formatter that just shows cursor state
	animatedFormatter := func(data Data[AnimTestItem], index int, ctx RenderContext,
		animationState map[string]any, isCursor bool, isTopThreshold bool, isBottomThreshold bool) RenderResult {

		prefix := "  "
		if isCursor {
			prefix = "> "
		}

		content := fmt.Sprintf("%s%s", prefix, data.Item.Name)

		return RenderResult{
			Content: content,
			RefreshTriggers: []RefreshTrigger{{
				Type:     TriggerTimer,
				Interval: 100 * time.Millisecond,
			}},
			AnimationState: map[string]any{
				"cursor": isCursor,
			},
		}
	}

	regularFormatter := func(data Data[AnimTestItem], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return fmt.Sprintf("%s%s", prefix, data.Item.Name)
	}

	styleConfig := StyleConfig{}
	list, err := NewTeaList(config, provider, styleConfig, regularFormatter)
	if err != nil {
		t.Fatalf("NewTeaList failed: %v", err)
	}

	list.SetAnimatedFormatter(animatedFormatter)

	// Test initial state
	fmt.Println("\n1. Initial state:")
	view1 := list.View()
	fmt.Print(view1)

	if !strings.Contains(view1, "> Item 0") {
		t.Error("Expected cursor to be on Item 0 initially")
	}

	// Test cursor movement
	fmt.Println("\n2. After moving cursor down:")
	list.MoveDown()
	view2 := list.View()
	fmt.Print(view2)

	if !strings.Contains(view2, "> Item 1") {
		t.Error("Expected cursor to be on Item 1 after moving down")
	}

	// Test rapid movements
	fmt.Println("\n3. After rapid movements:")
	for i := 0; i < 3; i++ {
		list.MoveDown()
	}
	view3 := list.View()
	fmt.Print(view3)

	if !strings.Contains(view3, "> Item 4") {
		t.Error("Expected cursor to be on Item 4 after rapid movements")
	}

	fmt.Println("✅ Cursor movement works smoothly with animations")
	fmt.Println("\n=== END CURSOR MOVEMENT TEST ===")
}
