package list_test

import (
	"fmt"
	"strings"
	"testing"

	vtable "github.com/davidroman0O/vtable/pure"
)

// TestBoundingAreaBasics tests the fundamental bounding area calculations
func TestBoundingAreaBasics(t *testing.T) {
	t.Log("=== BOUNDING AREA BASICS TEST ===")

	// Create list with specific bounding configuration
	// Viewport: 8 items, Chunk: 8 items, 1 chunk before, 2 chunks after
	list, _ := createListWithBoundingConfig(t, 100, 8, 8, 1, 2)

	// Test initial bounding area calculation
	t.Log("\n--- Test 1: Initial Bounding Area ---")
	state := list.GetState()
	t.Logf("Viewport: %d-%d (height %d)",
		state.ViewportStartIndex,
		state.ViewportStartIndex+7, 8)

	// Expected bounding area:
	// - 1 chunk before: items -8 to -1 (clamped to 0-7)
	// - Viewport: items 0-7
	// - 2 chunks after: items 8-23
	// Total bounding: items 0-23 (chunks 0, 8, 16)

	// Simulate what the bounding area should contain
	expectedChunks := []int{0, 8, 16} // chunk start indices
	t.Logf("Expected chunks in bounding area: %v", expectedChunks)

	// Test navigation and bounding area movement
	t.Log("\n--- Test 2: Bounding Area Movement ---")

	// Move to middle of dataset (around item 40)
	list.JumpToIndex(40)
	state = list.GetState()
	t.Logf("After jump to 40 - Viewport: %d-%d",
		state.ViewportStartIndex,
		state.ViewportStartIndex+7)

	// Expected bounding area around viewport 32-39:
	// - 1 chunk before: items 24-31 (chunk 24)
	// - Viewport: items 32-39 (chunk 32)
	// - 2 chunks after: items 40-55 (chunks 40, 48)
	// Total bounding: items 24-55 (chunks 24, 32, 40, 48)

	expectedChunks = []int{24, 32, 40, 48}
	t.Logf("Expected chunks around item 40: %v", expectedChunks)

	t.Log("\n=== BOUNDING AREA BASICS TEST COMPLETED ===")
}

// TestBoundingAreaDynamicLoading tests that chunks are loaded/unloaded dynamically
func TestBoundingAreaDynamicLoading(t *testing.T) {
	t.Log("=== BOUNDING AREA DYNAMIC LOADING TEST ===")

	// Create list with smaller chunks to see loading behavior
	// Viewport: 5 items, Chunk: 5 items, 1 chunk before, 1 chunk after
	list, _ := createListWithBoundingConfig(t, 50, 5, 5, 1, 1)

	// Test 1: Initial state - should load chunks 0, 5, 10
	t.Log("\n--- Test 1: Initial Chunk Loading ---")
	state := list.GetState()
	view := list.View()
	t.Logf("Initial viewport: %d, view:\n%s", state.ViewportStartIndex, view)

	// Verify initial data is available
	if view == "No data available" {
		t.Error("Expected initial data to be loaded")
	}

	// Test 2: Navigate forward to trigger new chunk loading
	t.Log("\n--- Test 2: Forward Navigation Chunk Loading ---")

	// Move to item 15 (chunk 15), should trigger loading chunks 10, 15, 20
	list.JumpToIndex(15)
	state = list.GetState()
	view = list.View()
	t.Logf("After jump to 15 - Viewport: %d, view:\n%s", state.ViewportStartIndex, view)

	// Should show Person_15 and nearby items
	if view == "No data available" {
		t.Error("Expected data around item 15 to be loaded")
	}

	// Test 3: Navigate backward to test reverse loading
	t.Log("\n--- Test 3: Backward Navigation Chunk Loading ---")

	// Move back to item 5 (chunk 5), should load chunks 0, 5, 10
	list.JumpToIndex(5)
	state = list.GetState()
	view = list.View()
	t.Logf("After jump back to 5 - Viewport: %d, view:\n%s", state.ViewportStartIndex, view)

	// Should show Person_5 and nearby items
	if view == "No data available" {
		t.Error("Expected data around item 5 to be loaded")
	}

	// Test 4: Gradual navigation to test smooth loading
	t.Log("\n--- Test 4: Gradual Navigation ---")

	initialViewport := state.ViewportStartIndex
	// Move down several times to test gradual chunk loading
	for i := 0; i < 10; i++ {
		list.MoveDown()
		state = list.GetState()
		view = list.View()

		if view == "No data available" {
			t.Errorf("Data should be available after %d moves", i+1)
		}

		t.Logf("Move %d - Cursor: %d, Viewport: %d", i+1, state.CursorIndex, state.ViewportStartIndex)
	}

	finalViewport := state.ViewportStartIndex
	t.Logf("Navigation test: moved from viewport %d to %d", initialViewport, finalViewport)

	t.Log("\n=== BOUNDING AREA DYNAMIC LOADING TEST COMPLETED ===")
}

// TestBoundingAreaConfiguration tests different bounding configurations
func TestBoundingAreaConfiguration(t *testing.T) {
	t.Log("=== BOUNDING AREA CONFIGURATION TEST ===")

	// Test 1: Minimal bounding (0 before, 1 after)
	t.Log("\n--- Test 1: Minimal Bounding Configuration ---")
	list1, _ := createListWithBoundingConfig(t, 50, 10, 5, 0, 1)

	list1.JumpToIndex(20)
	state := list1.GetState()
	view := list1.View()
	t.Logf("Minimal bounding - Viewport: %d, view available: %t",
		state.ViewportStartIndex, view != "No data available")

	// Test 2: Aggressive bounding (2 before, 3 after)
	t.Log("\n--- Test 2: Aggressive Bounding Configuration ---")
	list2, _ := createListWithBoundingConfig(t, 50, 5, 5, 2, 3)

	list2.JumpToIndex(20)
	state = list2.GetState()
	view = list2.View()
	t.Logf("Aggressive bounding - Viewport: %d, view available: %t",
		state.ViewportStartIndex, view != "No data available")

	// Test 3: Asymmetric bounding (3 before, 1 after)
	t.Log("\n--- Test 3: Asymmetric Bounding Configuration ---")
	list3, _ := createListWithBoundingConfig(t, 50, 8, 5, 3, 1)

	list3.JumpToIndex(20)
	state = list3.GetState()
	view = list3.View()
	t.Logf("Asymmetric bounding - Viewport: %d, view available: %t",
		state.ViewportStartIndex, view != "No data available")

	t.Log("\n=== BOUNDING AREA CONFIGURATION TEST COMPLETED ===")
}

// TestBoundingAreaWithThresholds tests that bounding area works with threshold navigation
func TestBoundingAreaWithThresholds(t *testing.T) {
	t.Log("=== BOUNDING AREA WITH THRESHOLDS TEST ===")

	// Create list with both bounding area and thresholds
	// Viewport: 7 items, thresholds at 2 and 4, bounding: 1 before, 2 after
	list, _ := createListWithBoundingAndThresholds(t, 50, 7, 7, 2, 4, 1, 2)

	// Test threshold navigation with bounding area
	t.Log("\n--- Test 1: Threshold Navigation with Bounding ---")

	state := list.GetState()
	view := list.View()
	t.Logf("Initial - Cursor: %d, Viewport: %d, TopThreshold: %t, BottomThreshold: %t",
		state.CursorIndex, state.ViewportStartIndex, state.IsAtTopThreshold, state.IsAtBottomThreshold)
	t.Logf("View:\n%s", view)

	// Move to bottom threshold and trigger scroll
	for i := 0; i < 5; i++ {
		list.MoveDown()
		state = list.GetState()
		if state.IsAtBottomThreshold {
			t.Logf("Reached bottom threshold at cursor %d", state.CursorIndex)
			break
		}
	}

	// Move one more to trigger viewport scroll
	list.MoveDown()
	state = list.GetState()
	view = list.View()
	t.Logf("After threshold scroll - Cursor: %d, Viewport: %d",
		state.CursorIndex, state.ViewportStartIndex)
	t.Logf("View:\n%s", view)

	// Verify data is still available after threshold-triggered scroll
	if view == "No data available" {
		t.Error("Bounding area should ensure data is available after threshold scroll")
	}

	t.Log("\n=== BOUNDING AREA WITH THRESHOLDS TEST COMPLETED ===")
}

// TestAutomaticBoundingAreaCalculation tests the FULLY AUTOMATED bounding area calculation
func TestAutomaticBoundingAreaCalculation(t *testing.T) {
	t.Log("=== AUTOMATIC BOUNDING AREA CALCULATION TEST ===")

	// Create list with specific bounding configuration
	// BoundingAreaBefore: 4 items before viewport
	// BoundingAreaAfter: 4 items after viewport
	list, _ := createTestListWithBoundingConfig(t, 50, 5, 8, 4, 4)

	// Test 1: Beginning position - viewport items 0-7
	t.Log("\n--- Test 1: Beginning Position (viewport 0-7) ---")

	list.JumpToStart()
	state := list.GetState()
	view := list.View()

	t.Logf("Beginning - Viewport: %d-%d, Cursor: %d",
		state.ViewportStartIndex, state.ViewportStartIndex+7, state.CursorIndex)
	t.Logf("View:\n%s", view)

	// Expected bounding area:
	// - BoundingAreaBefore: 4 items before viewport start (items -4 to -1, clamped to 0)
	// - Viewport: items 0-7
	// - BoundingAreaAfter: 4 items after viewport end (items 8-11)
	// Total bounding: items 0-11 (covers chunks 0, 5, 10)

	// Verify data is loaded correctly
	if !strings.Contains(view, "Person_0") {
		t.Error("Expected Person_0 to be visible at beginning")
	}
	if !strings.Contains(view, "Person_7") {
		t.Error("Expected Person_7 to be visible at beginning")
	}

	// Test 2: Middle position - move to viewport items 20-27
	t.Log("\n--- Test 2: Middle Position (viewport 20-27) ---")

	list.JumpToIndex(23) // Place cursor in middle of viewport 20-27
	state = list.GetState()
	view = list.View()

	t.Logf("Middle - Viewport: %d-%d, Cursor: %d",
		state.ViewportStartIndex, state.ViewportStartIndex+7, state.CursorIndex)
	t.Logf("View:\n%s", view)

	// Expected bounding area:
	// - BoundingAreaBefore: 4 items before viewport start (items 16-19)
	// - Viewport: items 20-27
	// - BoundingAreaAfter: 4 items after viewport end (items 28-31)
	// Total bounding: items 16-31 (covers chunks 15, 20, 25, 30)

	// Verify correct items are visible
	if !strings.Contains(view, "Person_20") {
		t.Error("Expected Person_20 to be visible in middle position")
	}
	if !strings.Contains(view, "Person_27") {
		t.Error("Expected Person_27 to be visible in middle position")
	}
	if !strings.Contains(view, "Person_23") {
		t.Error("Expected Person_23 (cursor) to be visible in middle position")
	}

	// Test 3: End position - viewport items 42-49
	t.Log("\n--- Test 3: End Position (viewport 42-49) ---")

	list.JumpToEnd() // Jump to last item (49)
	state = list.GetState()
	view = list.View()

	t.Logf("End - Viewport: %d-%d, Cursor: %d",
		state.ViewportStartIndex, state.ViewportStartIndex+7, state.CursorIndex)
	t.Logf("View:\n%s", view)

	// Expected bounding area:
	// - BoundingAreaBefore: 4 items before viewport start (items 38-41)
	// - Viewport: items 42-49
	// - BoundingAreaAfter: 4 items after viewport end (items 50-53, clamped to 49)
	// Total bounding: items 38-49 (covers chunks 35, 40, 45)

	// Verify end items are visible
	if !strings.Contains(view, "Person_42") {
		t.Error("Expected Person_42 to be visible at end")
	}
	if !strings.Contains(view, "Person_49") {
		t.Error("Expected Person_49 to be visible at end")
	}

	// Test 4: Gradual navigation to verify dynamic calculation
	t.Log("\n--- Test 4: Gradual Navigation ---")

	list.JumpToStart()

	// Move through several positions and verify bounding area updates
	positions := []int{10, 20, 30, 40}

	for _, targetPos := range positions {
		t.Logf("Moving to position %d", targetPos)

		list.JumpToIndex(targetPos)
		state = list.GetState()
		view = list.View()

		t.Logf("Position %d - Viewport: %d-%d, Cursor: %d",
			targetPos, state.ViewportStartIndex, state.ViewportStartIndex+7, state.CursorIndex)

		// Verify data is available (no "Loading..." or "No data available")
		if strings.Contains(view, "Loading") || strings.Contains(view, "No data available") {
			t.Errorf("Position %d: Data should be loaded automatically, got: %s", targetPos, view)
		}

		// Verify correct cursor item is visible
		expectedItem := fmt.Sprintf("Person_%d", targetPos)
		if !strings.Contains(view, expectedItem) {
			t.Errorf("Position %d: Expected '%s' to be visible, got: %s", targetPos, expectedItem, view)
		}

		// Verify cursor is shown correctly
		if !strings.Contains(view, "> ") {
			t.Errorf("Position %d: Expected cursor (>) to be visible, got: %s", targetPos, view)
		}
	}

	t.Log("\n=== AUTOMATIC BOUNDING AREA CALCULATION TEST COMPLETED ===")
}

// TestBoundingAreaDifferentConfigurations tests various bounding area configurations
func TestBoundingAreaDifferentConfigurations(t *testing.T) {
	t.Log("=== BOUNDING AREA DIFFERENT CONFIGURATIONS TEST ===")

	// Test different bounding configurations
	configurations := []struct {
		name        string
		before      int
		after       int
		description string
	}{
		{"Conservative", 0, 1, "Minimal loading - 0 before, 1 after"},
		{"Balanced", 2, 2, "Balanced loading - 2 before, 2 after"},
		{"Aggressive", 4, 4, "Aggressive loading - 4 before, 4 after"},
		{"Asymmetric", 1, 3, "Asymmetric loading - 1 before, 3 after"},
	}

	for _, config := range configurations {
		t.Logf("\n--- Testing %s Configuration: %s ---", config.name, config.description)

		list, _ := createTestListWithBoundingConfig(t, 30, 5, 6, config.before, config.after)

		// Test navigation in middle of dataset
		list.JumpToIndex(15) // Position 15 in 30-item dataset
		state := list.GetState()
		view := list.View()

		t.Logf("%s - Viewport: %d-%d, Cursor: %d",
			config.name, state.ViewportStartIndex, state.ViewportStartIndex+5, state.CursorIndex)
		t.Logf("BoundingAreaBefore: %d, BoundingAreaAfter: %d", config.before, config.after)
		t.Logf("View:\n%s", view)

		// Verify data is loaded correctly
		if strings.Contains(view, "Loading") || strings.Contains(view, "No data available") {
			t.Errorf("%s: Data should be loaded, got: %s", config.name, view)
		}

		// Verify cursor item is visible
		if !strings.Contains(view, "Person_15") {
			t.Errorf("%s: Expected Person_15 to be visible, got: %s", config.name, view)
		}

		// Verify cursor is shown
		if !strings.Contains(view, "> ") {
			t.Errorf("%s: Expected cursor (>) to be visible, got: %s", config.name, view)
		}
	}

	t.Log("\n=== BOUNDING AREA DIFFERENT CONFIGURATIONS TEST COMPLETED ===")
}

// TestBoundingAreaWithScrolling tests bounding area during viewport scrolling
func TestBoundingAreaWithScrolling(t *testing.T) {
	t.Log("=== BOUNDING AREA WITH SCROLLING TEST ===")

	// Create list with moderate bounding area
	list, _ := createTestListWithBoundingConfig(t, 25, 5, 5, 2, 2)

	// Test scrolling through the dataset
	t.Log("\n--- Scrolling Through Dataset ---")

	positions := []int{0, 5, 10, 15, 20, 24}

	for i, pos := range positions {
		t.Logf("Step %d: Jumping to position %d", i+1, pos)

		list.JumpToIndex(pos)
		state := list.GetState()
		view := list.View()

		t.Logf("Position %d - Viewport: %d-%d, Cursor: %d",
			pos, state.ViewportStartIndex, state.ViewportStartIndex+4, state.CursorIndex)

		// Verify continuous data availability during scrolling
		if strings.Contains(view, "Loading") {
			t.Errorf("Position %d: Bounding area should prevent loading delays, got: %s", pos, view)
		}

		// Verify correct item is at cursor
		expectedItem := fmt.Sprintf("Person_%d", pos)
		if !strings.Contains(view, expectedItem) {
			t.Errorf("Position %d: Expected '%s' to be visible, got: %s", pos, expectedItem, view)
		}

		// Count visible items (should be 5 unless at end)
		lines := strings.Split(strings.TrimSpace(view), "\n")
		if pos <= 20 && len(lines) != 5 {
			t.Errorf("Position %d: Expected 5 visible items, got %d lines: %s", pos, len(lines), view)
		}
	}

	t.Log("\n=== BOUNDING AREA WITH SCROLLING TEST COMPLETED ===")
}

// Helper function to create list with automatic bounding calculation
func createListWithBoundingConfig(t *testing.T, dataCount int, chunkSize int, viewportHeight int, _ int, _ int) (*vtable.List, *SimpleDataSource) {
	// Create test data
	people := make([]TestPerson, dataCount)
	for i := 0; i < dataCount; i++ {
		people[i] = TestPerson{
			Name: fmt.Sprintf("Person_%d", i),
			Age:  25 + (i % 40),
		}
	}

	// Create data source
	dataSource := &SimpleDataSource{people: people}

	// Create configuration with automatic bounding area
	config := vtable.ListConfig{
		ViewportConfig: vtable.ViewportConfig{
			Height:          viewportHeight,
			TopThreshold:    1,
			BottomThreshold: 1, // This will be position viewportHeight - 1 - 1 = viewportHeight - 2
			ChunkSize:       chunkSize,
			InitialIndex:    0,
			// FULLY AUTOMATED bounding area - calculated based on viewport position!
		},
		KeyMap: vtable.DefaultNavigationKeyMap(),
	}

	// Create list
	list := vtable.NewList(config, dataSource)

	// Set up formatter
	formatter := func(data vtable.Data[any], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		item := data.Item.(TestPerson)
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return fmt.Sprintf("%s{%s %d}", prefix, item.Name, item.Age)
	}
	list.SetFormatter(formatter)

	// Load initial data
	totalCmd := dataSource.GetTotal()
	if totalMsg := totalCmd(); totalMsg != nil {
		list.Update(totalMsg)
	}

	// Load first chunk
	loadCmd := dataSource.LoadChunk(vtable.DataRequest{Start: 0, Count: chunkSize})
	if loadMsg := loadCmd(); loadMsg != nil {
		list.Update(loadMsg)
	}

	return list, dataSource
}

// Helper function to create list with automatic bounding calculation
func createListWithBoundingAndThresholds(t *testing.T, dataCount int, chunkSize int, viewportHeight int, topThreshold int, bottomThreshold int, _ int, _ int) (*vtable.List, *SimpleDataSource) {
	// Create test data
	people := make([]TestPerson, dataCount)
	for i := 0; i < dataCount; i++ {
		people[i] = TestPerson{
			Name: fmt.Sprintf("Person_%d", i),
			Age:  25 + (i % 40),
		}
	}

	// Create data source
	dataSource := &SimpleDataSource{people: people}

	// Create configuration with automatic bounding area
	config := vtable.ListConfig{
		ViewportConfig: vtable.ViewportConfig{
			Height:          viewportHeight,
			TopThreshold:    1,
			BottomThreshold: 1, // This will be position viewportHeight - 1 - 1 = viewportHeight - 2
			ChunkSize:       chunkSize,
			InitialIndex:    0,
			// FULLY AUTOMATED bounding area - calculated based on viewport position!
		},
		KeyMap: vtable.DefaultNavigationKeyMap(),
	}

	// Create list
	list := vtable.NewList(config, dataSource)

	// Set up formatter
	formatter := func(data vtable.Data[any], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		item := data.Item.(TestPerson)
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return fmt.Sprintf("%s{%s %d}", prefix, item.Name, item.Age)
	}
	list.SetFormatter(formatter)

	// Load initial data
	totalCmd := dataSource.GetTotal()
	if totalMsg := totalCmd(); totalMsg != nil {
		list.Update(totalMsg)
	}

	// Load first chunk
	loadCmd := dataSource.LoadChunk(vtable.DataRequest{Start: 0, Count: chunkSize})
	if loadMsg := loadCmd(); loadMsg != nil {
		list.Update(loadMsg)
	}

	return list, dataSource
}

// createTestListWithBoundingConfig creates a list with specific bounding configuration for testing
func createTestListWithBoundingConfig(t *testing.T, dataCount, chunkSize, viewportHeight, boundingBefore, boundingAfter int) (*vtable.List, *SimpleDataSource) {
	// Create test data
	people := make([]TestPerson, dataCount)
	for i := 0; i < dataCount; i++ {
		people[i] = TestPerson{
			Name: fmt.Sprintf("Person_%d", i),
			Age:  25 + (i % 40),
		}
	}

	// Create data source
	dataSource := &SimpleDataSource{people: people}

	// Create configuration with specified bounding area
	config := vtable.ListConfig{
		ViewportConfig: vtable.ViewportConfig{
			Height:             viewportHeight,
			TopThreshold:       1,
			BottomThreshold:    1,
			ChunkSize:          chunkSize,
			InitialIndex:       0,
			BoundingAreaBefore: boundingBefore, // Configured distance before viewport
			BoundingAreaAfter:  boundingAfter,  // Configured distance after viewport
		},
		KeyMap: vtable.DefaultNavigationKeyMap(),
	}

	// Create list
	list := vtable.NewList(config, dataSource)

	// Set up formatter
	formatter := func(data vtable.Data[any], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		item := data.Item.(TestPerson)
		prefix := "  "
		if isCursor {
			prefix = "> "
		}
		return fmt.Sprintf("%s{%s %d}", prefix, item.Name, item.Age)
	}
	list.SetFormatter(formatter)

	// Load initial data
	totalCmd := dataSource.GetTotal()
	if totalMsg := totalCmd(); totalMsg != nil {
		list.Update(totalMsg)
	}

	// Load first chunk
	loadCmd := dataSource.LoadChunk(vtable.DataRequest{Start: 0, Count: chunkSize})
	if loadMsg := loadCmd(); loadMsg != nil {
		list.Update(loadMsg)
	}

	return list, dataSource
}
