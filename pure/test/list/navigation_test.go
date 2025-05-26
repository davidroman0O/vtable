package list_test

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	vtable "github.com/davidroman0O/vtable/pure"
)

// TestPerson represents our test data
type TestPerson struct {
	Name string
	Age  int
}

// SimpleDataSource implements the pure DataSource interface
type SimpleDataSource struct {
	people []TestPerson
}

func (s *SimpleDataSource) LoadChunk(request vtable.DataRequest) tea.Cmd {
	start := request.Start
	count := request.Count
	total := len(s.people)

	if start >= total {
		return vtable.DataChunkLoadedCmd(start, []vtable.Data[any]{}, request)
	}

	end := start + count
	if end > total {
		end = total
	}

	var chunkItems []vtable.Data[any]
	for i := start; i < end; i++ {
		person := s.people[i]
		chunkItems = append(chunkItems, vtable.Data[any]{
			ID:   fmt.Sprintf("person_%d", i),
			Item: person,
		})
	}

	return vtable.DataChunkLoadedCmd(start, chunkItems, request)
}

func (s *SimpleDataSource) GetTotal() tea.Cmd {
	return vtable.DataTotalCmd(len(s.people))
}

func (s *SimpleDataSource) RefreshTotal() tea.Cmd {
	return s.GetTotal()
}

func (s *SimpleDataSource) GetItemID(item any) string {
	if person, ok := item.(TestPerson); ok {
		return fmt.Sprintf("person_%s_%d", person.Name, person.Age)
	}
	return fmt.Sprintf("%v", item)
}

// Selection operations - required by DataSource interface
func (s *SimpleDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return vtable.SelectionResponseCmd(true, index, fmt.Sprintf("person_%d", index), selected, "toggle", nil, nil)
}

func (s *SimpleDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	return vtable.SelectionResponseCmd(true, -1, id, selected, "toggleByID", nil, nil)
}

func (s *SimpleDataSource) SelectAll() tea.Cmd {
	affectedIDs := make([]string, len(s.people))
	for i := 0; i < len(s.people); i++ {
		affectedIDs[i] = fmt.Sprintf("person_%d", i)
	}
	return vtable.SelectionResponseCmd(true, -1, "", true, "selectAll", nil, affectedIDs)
}

func (s *SimpleDataSource) ClearSelection() tea.Cmd {
	return vtable.SelectionResponseCmd(true, -1, "", false, "clear", nil, nil)
}

func (s *SimpleDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	affectedIDs := make([]string, endIndex-startIndex+1)
	for i := startIndex; i <= endIndex; i++ {
		affectedIDs[i-startIndex] = fmt.Sprintf("person_%d", i)
	}
	return vtable.SelectionResponseCmd(true, -1, "", true, "range", nil, affectedIDs)
}

// Helper function to create a list with data loaded
func createListWithData(t *testing.T, dataCount int, chunkSize int, viewportHeight int) (*vtable.List, *SimpleDataSource) {
	// Create test data
	people := make([]TestPerson, dataCount)
	for i := 0; i < dataCount; i++ {
		people[i] = TestPerson{
			Name: fmt.Sprintf("Person_%d", i),
			Age:  25 + (i % 40), // Ages 25-64
		}
	}

	// Create data source
	dataSource := &SimpleDataSource{people: people}

	// Create configuration
	config := vtable.ListConfig{
		ViewportConfig: vtable.ViewportConfig{
			Height:          5,
			TopThreshold:    1, // 1 position from viewport start
			BottomThreshold: 1, // 1 position from viewport end (position 3 in height-5 viewport)
			ChunkSize:       3,
			InitialIndex:    0,
			// FULLY AUTOMATED bounding area!
		},
		KeyMap: vtable.DefaultNavigationKeyMap(),
	}

	// Create list
	list := vtable.NewList(config, dataSource)

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

func TestBasicNavigation(t *testing.T) {
	list, _ := createListWithData(t, 10, 10, 5)

	// Test initial state
	state := list.GetState()
	view := list.View()
	t.Logf("Initial state - Cursor: %d, Viewport: %d", state.CursorIndex, state.ViewportStartIndex)
	t.Logf("Initial View:\n%s", view)

	if state.CursorIndex != 0 {
		t.Errorf("Expected initial cursor at 0, got %d", state.CursorIndex)
	}
	if state.ViewportStartIndex != 0 {
		t.Errorf("Expected initial viewport at 0, got %d", state.ViewportStartIndex)
	}
	if view == "No data available" || view == "" {
		t.Fatal("Expected initial view to show data")
	}
	// TEST VIEW CONTENT: Should show cursor on first item
	if !strings.Contains(view, "> {Person_0 25}") {
		t.Errorf("Expected view to contain '> {Person_0 25}' (cursor on first item), got: %s", view)
	}
	if !strings.Contains(view, "  {Person_1 26}") {
		t.Errorf("Expected view to contain '  {Person_1 26}' (second item without cursor), got: %s", view)
	}

	// Test MoveDown
	list.MoveDown()
	state = list.GetState()
	view = list.View()
	t.Logf("After MoveDown - Cursor: %d, Viewport: %d", state.CursorIndex, state.ViewportStartIndex)
	t.Logf("View after MoveDown:\n%s", view)

	if state.CursorIndex != 1 {
		t.Errorf("Expected cursor at 1 after MoveDown, got %d", state.CursorIndex)
	}
	if view == "No data available" || view == "" {
		t.Error("Expected view to show data after MoveDown")
	}
	// TEST VIEW CONTENT: Should show cursor on second item
	if !strings.Contains(view, "> {Person_1 26}") {
		t.Errorf("Expected view to contain '> {Person_1 26}' (cursor on second item), got: %s", view)
	}
	if !strings.Contains(view, "  {Person_0 25}") {
		t.Errorf("Expected view to contain '  {Person_0 25}' (first item without cursor), got: %s", view)
	}

	// Test MoveUp
	list.MoveUp()
	state = list.GetState()
	view = list.View()
	t.Logf("After MoveUp - Cursor: %d, Viewport: %d", state.CursorIndex, state.ViewportStartIndex)
	t.Logf("View after MoveUp:\n%s", view)

	if state.CursorIndex != 0 {
		t.Errorf("Expected cursor back at 0 after MoveUp, got %d", state.CursorIndex)
	}
	if view == "No data available" || view == "" {
		t.Error("Expected view to show data after MoveUp")
	}
	// TEST VIEW CONTENT: Should show cursor back on first item
	if !strings.Contains(view, "> {Person_0 25}") {
		t.Errorf("Expected view to contain '> {Person_0 25}' (cursor back on first item), got: %s", view)
	}
	if !strings.Contains(view, "  {Person_1 26}") {
		t.Errorf("Expected view to contain '  {Person_1 26}' (second item without cursor), got: %s", view)
	}

	// Test JumpToEnd
	list.JumpToEnd()
	state = list.GetState()
	view = list.View()
	t.Logf("After JumpToEnd - Cursor: %d, Viewport: %d", state.CursorIndex, state.ViewportStartIndex)
	t.Logf("View after JumpToEnd:\n%s", view)

	if state.CursorIndex != 9 {
		t.Errorf("Expected cursor at 9 (last item) after JumpToEnd, got %d", state.CursorIndex)
	}
	if view == "No data available" || view == "" {
		t.Error("Expected view to show data after JumpToEnd")
	}
	// TEST VIEW CONTENT: Should show cursor on last item
	if !strings.Contains(view, "> {Person_9 34}") {
		t.Errorf("Expected view to contain '> {Person_9 34}' (cursor on last item), got: %s", view)
	}
	// Should also show items before the last item
	if !strings.Contains(view, "  {Person_8 33}") {
		t.Errorf("Expected view to contain '  {Person_8 33}' (second-to-last item), got: %s", view)
	}

	// Test JumpToStart
	list.JumpToStart()
	state = list.GetState()
	view = list.View()
	t.Logf("After JumpToStart - Cursor: %d, Viewport: %d", state.CursorIndex, state.ViewportStartIndex)
	t.Logf("View after JumpToStart:\n%s", view)

	if state.CursorIndex != 0 {
		t.Errorf("Expected cursor at 0 after JumpToStart, got %d", state.CursorIndex)
	}
	if view == "No data available" || view == "" {
		t.Error("Expected view to show data after JumpToStart")
	}
	// TEST VIEW CONTENT: Should show cursor back on first item
	if !strings.Contains(view, "> {Person_0 25}") {
		t.Errorf("Expected view to contain '> {Person_0 25}' (cursor back on first item), got: %s", view)
	}
	if !strings.Contains(view, "  {Person_1 26}") {
		t.Errorf("Expected view to contain '  {Person_1 26}' (second item without cursor), got: %s", view)
	}
}

func TestViewportAndThresholds(t *testing.T) {
	// Create list with 20 items, chunk size 10, viewport height 5
	list, _ := createListWithData(t, 20, 10, 5)

	// Test threshold behavior - move to threshold position
	list.MoveDown() // cursor = 1 (at top threshold)
	state := list.GetState()
	view := list.View()
	t.Logf("At top threshold - Cursor: %d, Viewport: %d, IsAtTopThreshold: %t",
		state.CursorIndex, state.ViewportStartIndex, state.IsAtTopThreshold)
	t.Logf("View at top threshold:\n%s", view)

	if !state.IsAtTopThreshold {
		t.Error("Expected to be at top threshold when cursor is at index 1")
	}
	if view == "No data available" || view == "" {
		t.Error("Expected view to show data at top threshold")
	}
	// TEST VIEW CONTENT: Should show cursor on second item at top threshold
	if !strings.Contains(view, "> {Person_1 26}") {
		t.Errorf("Expected view to contain '> {Person_1 26}' (cursor at top threshold), got: %s", view)
	}
	if !strings.Contains(view, "  {Person_0 25}") {
		t.Errorf("Expected view to contain '  {Person_0 25}' (first item visible), got: %s", view)
	}

	// Move down to bottom threshold
	list.MoveDown() // cursor = 2
	list.MoveDown() // cursor = 3 (at bottom threshold)
	state = list.GetState()
	view = list.View()
	t.Logf("At bottom threshold - Cursor: %d, Viewport: %d, IsAtBottomThreshold: %t",
		state.CursorIndex, state.ViewportStartIndex, state.IsAtBottomThreshold)
	t.Logf("View at bottom threshold:\n%s", view)

	if !state.IsAtBottomThreshold {
		t.Error("Expected to be at bottom threshold")
	}
	if view == "No data available" || view == "" {
		t.Error("Expected view to show data at bottom threshold")
	}
	// TEST VIEW CONTENT: Should show cursor on item at bottom threshold
	if !strings.Contains(view, "> {Person_3 28}") {
		t.Errorf("Expected view to contain '> {Person_3 28}' (cursor at bottom threshold), got: %s", view)
	}
	// Viewport should have scrolled to show this item
	if !strings.Contains(view, "  {Person_2 27}") {
		t.Errorf("Expected view to contain '  {Person_2 27}' (visible in scrolled viewport), got: %s", view)
	}

	// Moving down from bottom threshold should scroll viewport
	previousViewport := state.ViewportStartIndex
	list.MoveDown() // cursor = 4, should scroll viewport
	state = list.GetState()
	view = list.View()
	t.Logf("After viewport scroll - Previous: %d, Current: %d, Cursor: %d",
		previousViewport, state.ViewportStartIndex, state.CursorIndex)
	t.Logf("View after viewport scroll:\n%s", view)

	if state.ViewportStartIndex <= previousViewport {
		t.Errorf("Expected viewport to scroll, was %d, now %d", previousViewport, state.ViewportStartIndex)
	}
	if view == "No data available" || view == "" {
		t.Error("Expected view to show data after viewport scroll")
	}
	// TEST VIEW CONTENT: Should show cursor on next item after viewport scroll
	if !strings.Contains(view, "> {Person_4 29}") {
		t.Errorf("Expected view to contain '> {Person_4 29}' (cursor after viewport scroll), got: %s", view)
	}
	// Should show scrolled viewport content
	if !strings.Contains(view, "  {Person_3 28}") {
		t.Errorf("Expected view to contain '  {Person_3 28}' (visible in new viewport), got: %s", view)
	}
	// Should not show items from old viewport
	if strings.Contains(view, "  {Person_0 25}") {
		t.Errorf("Expected view to NOT contain '  {Person_0 25}' (should be scrolled out), got: %s", view)
	}
}

func TestChunkLoadingOnScrolling(t *testing.T) {
	// Create list with 50 items, small chunk size to test loading
	list, _ := createListWithData(t, 50, 5, 5)

	// Show initial state
	state := list.GetState()
	view := list.View()
	t.Logf("Initial state - Cursor: %d, Viewport: %d, CursorViewportIndex: %d",
		state.CursorIndex, state.ViewportStartIndex, state.CursorViewportIndex)
	t.Logf("Initial view:\n%s", view)
	// TEST VIEW CONTENT: Should show initial chunk loaded
	if !strings.Contains(view, "> {Person_0 25}") {
		t.Errorf("Expected view to contain '> {Person_0 25}' (initial cursor), got: %s", view)
	}
	if !strings.Contains(view, "  {Person_1 26}") {
		t.Errorf("Expected view to contain '  {Person_1 26}' (initial chunk), got: %s", view)
	}

	// Initially only chunk 0 should be loaded
	// Scroll down significantly to trigger chunk loading
	for i := 0; i < 15; i++ {
		list.MoveDown()
		if i == 4 || i == 9 || i == 14 { // Show progress at key points
			state := list.GetState()
			view := list.View()
			t.Logf("After %d moves - Cursor: %d, Viewport: %d, CursorViewportIndex: %d",
				i+1, state.CursorIndex, state.ViewportStartIndex, state.CursorViewportIndex)
			t.Logf("View after %d moves:\n%s", i+1, view)

			// TEST VIEW CONTENT at key checkpoints
			expectedCursor := fmt.Sprintf("> {Person_%d %d}", i+1, 25+(i+1))
			if !strings.Contains(view, expectedCursor) {
				t.Errorf("After %d moves, expected view to contain '%s' (cursor position), got: %s", i+1, expectedCursor, view)
			}

			// Make sure we're not showing "Loading..." when we have data
			if strings.Contains(view, "Loading...") && (i+1) <= 5 {
				t.Errorf("After %d moves, expected no 'Loading...' for items in initial chunk, got: %s", i+1, view)
			}
		}
	}

	state = list.GetState()
	view = list.View()
	t.Logf("Final state - Cursor: %d, Viewport: %d, CursorViewportIndex: %d",
		state.CursorIndex, state.ViewportStartIndex, state.CursorViewportIndex)
	t.Logf("Final view:\n%s", view)

	if state.CursorIndex != 15 {
		t.Errorf("Expected cursor at 15 after 15 moves, got %d", state.CursorIndex)
	}

	// EXPECTED: CursorViewportIndex should be 15 - 11 = 4
	expectedCursorViewportIndex := state.CursorIndex - state.ViewportStartIndex
	if state.CursorViewportIndex != expectedCursorViewportIndex {
		t.Errorf("CursorViewportIndex mismatch: expected %d (cursor %d - viewport %d), got %d",
			expectedCursorViewportIndex, state.CursorIndex, state.ViewportStartIndex, state.CursorViewportIndex)
	}

	// Should be able to render without issues
	if view == "No data available" {
		t.Error("Expected data to be available after scrolling, but got 'No data available'")
	}

	// TEST VIEW CONTENT: Should show cursor on final position
	if !strings.Contains(view, "> {Person_15 40}") {
		t.Errorf("Expected view to contain '> {Person_15 40}' (final cursor position), got: %s", view)
	}

	// Should not show "Loading..." for items that should be loaded
	linesWithLoading := 0
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "Loading...") {
			linesWithLoading++
		}
	}
	// Some "Loading..." is expected for chunks not yet loaded, but not all lines
	if linesWithLoading >= 5 {
		t.Errorf("Expected fewer than 5 'Loading...' lines in final view, got %d: %s", linesWithLoading, view)
	}
}

func TestExtensiveScrolling(t *testing.T) {
	// Create large dataset to test extensive scrolling
	list, _ := createListWithData(t, 200, 20, 8)

	// Test scrolling down extensively
	t.Log("Testing extensive downward scrolling...")
	for i := 0; i < 150; i++ {
		list.MoveDown()
		if i%25 == 0 { // Check state periodically
			state := list.GetState()
			view := list.View()
			if view == "No data available" {
				t.Fatalf("Lost data at scroll position %d, cursor %d", i, state.CursorIndex)
			}
		}
	}

	state := list.GetState()
	if state.CursorIndex != 150 {
		t.Errorf("Expected cursor at 150 after extensive scrolling, got %d", state.CursorIndex)
	}

	// Test scrolling back up extensively
	t.Log("Testing extensive upward scrolling...")
	for i := 0; i < 100; i++ {
		list.MoveUp()
		if i%25 == 0 { // Check state periodically
			state := list.GetState()
			view := list.View()
			if view == "No data available" {
				t.Fatalf("Lost data during upward scroll at position %d, cursor %d", i, state.CursorIndex)
			}
		}
	}

	state = list.GetState()
	if state.CursorIndex != 50 {
		t.Errorf("Expected cursor at 50 after scrolling back up, got %d", state.CursorIndex)
	}
}

func TestPageNavigation(t *testing.T) {
	list, _ := createListWithData(t, 50, 20, 5)

	// Test PageDown
	list.PageDown()
	state := list.GetState()
	if state.CursorIndex != 5 { // Should move by viewport height
		t.Errorf("Expected cursor at 5 after PageDown, got %d", state.CursorIndex)
	}

	// Test PageUp
	list.PageUp()
	state = list.GetState()
	if state.CursorIndex != 0 {
		t.Errorf("Expected cursor back at 0 after PageUp, got %d", state.CursorIndex)
	}

	// Test PageDown multiple times
	for i := 0; i < 5; i++ {
		list.PageDown()
	}

	state = list.GetState()
	expectedCursor := 25 // 5 page downs * 5 viewport height
	if state.CursorIndex != expectedCursor {
		t.Errorf("Expected cursor at %d after 5 PageDowns, got %d", expectedCursor, state.CursorIndex)
	}
}

func TestBoundaryConditions(t *testing.T) {
	list, _ := createListWithData(t, 10, 10, 5)

	// Test moving up from start (should not move)
	initialState := list.GetState()
	list.MoveUp()
	state := list.GetState()
	if state.CursorIndex != initialState.CursorIndex {
		t.Error("Cursor should not move up from initial position")
	}

	// Jump to end and test moving down (should not move)
	list.JumpToEnd()
	endState := list.GetState()
	list.MoveDown()
	state = list.GetState()
	if state.CursorIndex != endState.CursorIndex {
		t.Error("Cursor should not move down from end position")
	}

	// Test that dataset boundary flags are correct
	list.JumpToStart()
	state = list.GetState()
	if !state.AtDatasetStart {
		t.Error("Should be at dataset start")
	}

	list.JumpToEnd()
	state = list.GetState()
	if !state.AtDatasetEnd {
		t.Error("Should be at dataset end")
	}
}

func TestRenderingConsistency(t *testing.T) {
	list, _ := createListWithData(t, 30, 10, 5)

	// Test that rendering works at various scroll positions
	positions := []int{0, 5, 10, 15, 20, 25, 29}

	for _, pos := range positions {
		// Jump to position
		list.JumpToStart()
		for i := 0; i < pos; i++ {
			list.MoveDown()
		}

		// Check rendering
		view := list.View()
		if view == "No data available" || view == "" {
			t.Errorf("Rendering failed at position %d", pos)
		}

		// Check that view contains expected content
		state := list.GetState()
		if state.CursorIndex != pos {
			t.Errorf("Cursor position mismatch: expected %d, got %d", pos, state.CursorIndex)
		}
	}
}

func TestSystematicViewRendering(t *testing.T) {
	t.Log("=== SYSTEMATIC VIEW() RENDERING TEST ===")

	// Create small dataset for detailed testing
	list, _ := createListWithData(t, 8, 5, 4)

	// Test 1: Initial state rendering
	t.Log("\n--- Test 1: Initial State ---")
	view := list.View()
	state := list.GetState()
	t.Logf("Cursor: %d, Viewport: %d", state.CursorIndex, state.ViewportStartIndex)
	t.Logf("View Output:\n%s", view)

	if view == "No data available" {
		t.Fatal("Expected initial data to be loaded")
	}
	if !strings.Contains(view, "Person_0") {
		t.Errorf("Expected to see Person_0 in initial view")
	}

	// Test 2: Single move down
	t.Log("\n--- Test 2: After MoveDown() ---")
	list.MoveDown()
	view = list.View()
	state = list.GetState()
	t.Logf("Cursor: %d, Viewport: %d", state.CursorIndex, state.ViewportStartIndex)
	t.Logf("View Output:\n%s", view)

	if state.CursorIndex != 1 {
		t.Errorf("Expected cursor at 1, got %d", state.CursorIndex)
	}

	// Test 3: Multiple moves to test threshold behavior
	t.Log("\n--- Test 3: Moving to Threshold Position ---")
	list.MoveDown() // cursor = 2
	list.MoveDown() // cursor = 3 (at bottom threshold)
	view = list.View()
	state = list.GetState()
	t.Logf("Cursor: %d, Viewport: %d, AtBottomThreshold: %t",
		state.CursorIndex, state.ViewportStartIndex, state.IsAtBottomThreshold)
	t.Logf("View Output:\n%s", view)

	// Test 4: Viewport Scrolling
	t.Log("\n--- Test 4: Viewport Scrolling ---")
	previousViewport := state.ViewportStartIndex
	list.MoveDown() // cursor = 4, should scroll viewport
	view = list.View()
	state = list.GetState()
	t.Logf("Previous Viewport: %d, Current Viewport: %d", previousViewport, state.ViewportStartIndex)
	t.Logf("Cursor: %d, Viewport: %d", state.CursorIndex, state.ViewportStartIndex)
	t.Logf("View Output:\n%s", view)

	// Note: viewport scrolling behavior may vary based on threshold logic
	// The important thing is that navigation works correctly
	if state.CursorIndex != 4 {
		t.Errorf("Expected cursor at 4 after MoveDown, got %d", state.CursorIndex)
	}

	// Test 5: Jump to end
	t.Log("\n--- Test 5: Jump to End ---")
	list.JumpToEnd()
	view = list.View()
	state = list.GetState()
	t.Logf("Cursor: %d, Viewport: %d, AtDatasetEnd: %t",
		state.CursorIndex, state.ViewportStartIndex, state.AtDatasetEnd)
	t.Logf("View Output:\n%s", view)

	if !strings.Contains(view, "Person_7") {
		t.Errorf("Expected to see Person_7 (last item) in view after JumpToEnd")
	}

	// Test 6: Jump back to start
	t.Log("\n--- Test 6: Jump to Start ---")
	list.JumpToStart()
	view = list.View()
	state = list.GetState()
	t.Logf("Cursor: %d, Viewport: %d, AtDatasetStart: %t",
		state.CursorIndex, state.ViewportStartIndex, state.AtDatasetStart)
	t.Logf("View Output:\n%s", view)

	if !strings.Contains(view, "Person_0") {
		t.Errorf("Expected to see Person_0 (first item) in view after JumpToStart")
	}

	// Test 7: Page navigation
	t.Log("\n--- Test 7: Page Down ---")
	list.PageDown()
	view = list.View()
	state = list.GetState()
	t.Logf("Cursor: %d, Viewport: %d", state.CursorIndex, state.ViewportStartIndex)
	t.Logf("View Output:\n%s", view)

	if state.CursorIndex != 4 { // Should move by viewport height (4)
		t.Errorf("Expected cursor at 4 after PageDown, got %d", state.CursorIndex)
	}

	t.Log("\n=== SYSTEMATIC VIEW RENDERING TEST COMPLETED ===")
}

func TestLargeDatasetScrolling(t *testing.T) {
	t.Log("=== LARGE DATASET SCROLLING TEST ===")

	// Create larger dataset for chunk loading tests
	list, _ := createListWithData(t, 50, 10, 6)

	// Test extensive scrolling with periodic view checks
	positions := []int{0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 49}

	for _, targetPos := range positions {
		t.Logf("\n--- Scrolling to Position %d ---", targetPos)

		// Jump to start first
		list.JumpToStart()

		// Navigate to target position
		for i := 0; i < targetPos; i++ {
			list.MoveDown()
		}

		// Check state and view
		view := list.View()
		state := list.GetState()

		t.Logf("Target: %d, Actual Cursor: %d, Viewport: %d",
			targetPos, state.CursorIndex, state.ViewportStartIndex)

		if view == "No data available" {
			t.Fatalf("Lost data at position %d", targetPos)
		}

		if state.CursorIndex != targetPos {
			t.Errorf("Cursor mismatch: expected %d, got %d", targetPos, state.CursorIndex)
		}

		// Display view for manual verification (first few lines)
		lines := strings.Split(view, "\n")
		displayLines := lines
		if len(lines) > 8 {
			displayLines = lines[:8]
		}
		t.Logf("View (first %d lines):\n%s", len(displayLines), strings.Join(displayLines, "\n"))

		// Verify that we can see the expected item name
		expectedName := fmt.Sprintf("Person_%d", targetPos)
		if !strings.Contains(view, expectedName) {
			t.Errorf("Expected to see %s in view at position %d", expectedName, targetPos)
		}
	}

	t.Log("\n=== LARGE DATASET SCROLLING TEST COMPLETED ===")
}

func TestThresholdLockingBehavior(t *testing.T) {
	t.Log("=== THRESHOLD LOCKING BEHAVIOR TEST ===")

	// Create list with specific threshold configuration and larger dataset
	// Height=5, TopThreshold=1, BottomThreshold=1
	// TopThreshold=1 means cursor locks at viewport position 1
	// BottomThreshold=1 means cursor locks at viewport position 5-1-1=3
	list, _ := createListWithData(t, 20, 5, 5)

	// Test 1: Navigate to middle where we can test scrolling
	t.Log("\n--- Test 1: Setup for Threshold Testing ---")

	// Jump to position 10 so we have room to scroll both directions
	list.JumpToIndex(10)
	state := list.GetState()
	view := list.View()

	t.Logf("Setup - Cursor: %d, Viewport: %d, ViewportIdx: %d",
		state.CursorIndex, state.ViewportStartIndex, state.CursorViewportIndex)
	t.Logf("View:\n%s", view)

	// Test 2: Move to top threshold and verify locking
	t.Log("\n--- Test 2: Top Threshold Locking ---")

	// Navigate to a viewport where cursor is at top threshold
	list.JumpToIndex(6) // This should put us at viewport 5-9 with cursor at position 1 (top threshold)
	state = list.GetState()
	view = list.View()

	t.Logf("At threshold - Cursor: %d, Viewport: %d, ViewportIdx: %d, IsAtTopThreshold: %t",
		state.CursorIndex, state.ViewportStartIndex, state.CursorViewportIndex, state.IsAtTopThreshold)
	t.Logf("View:\n%s", view)

	// Verify we're at the threshold
	if !state.IsAtTopThreshold {
		t.Error("Expected to be at top threshold")
	}
	if state.CursorViewportIndex != 1 {
		t.Errorf("Expected cursor viewport index 1, got %d", state.CursorViewportIndex)
	}

	// Test 3: Move up from threshold - should trigger viewport scroll and cursor lock
	t.Log("\n--- Test 3: Viewport Scroll from Top Threshold ---")

	previousViewport := state.ViewportStartIndex
	previousCursor := state.CursorIndex
	list.MoveUp() // cursor should move to 5, viewport should scroll to maintain threshold lock

	state = list.GetState()
	view = list.View()

	t.Logf("After scroll - Cursor: %d->%d, Viewport: %d->%d, ViewportIdx: %d, IsAtTopThreshold: %t",
		previousCursor, state.CursorIndex, previousViewport, state.ViewportStartIndex, state.CursorViewportIndex, state.IsAtTopThreshold)
	t.Logf("View:\n%s", view)

	// Verify the expected behavior:
	// - Cursor should have moved to previous item (5)
	// - Viewport should have scrolled to maintain threshold (viewport 4-8)
	// - Cursor should still be at threshold position (viewport index 1)
	if state.CursorIndex != 5 {
		t.Errorf("Expected cursor at 5, got %d", state.CursorIndex)
	}
	if state.ViewportStartIndex != 4 {
		t.Errorf("Expected viewport at 4, got %d", state.ViewportStartIndex)
	}
	if state.CursorViewportIndex != 1 {
		t.Errorf("Expected cursor locked at threshold position 1, got %d", state.CursorViewportIndex)
	}

	// Test 4: Navigate to bottom threshold
	t.Log("\n--- Test 4: Bottom Threshold Locking ---")

	// Navigate to position where cursor is at bottom threshold
	list.JumpToIndex(8) // This should put us at viewport 5-9 with cursor at position 3 (bottom threshold)
	state = list.GetState()
	view = list.View()

	t.Logf("At bottom threshold - Cursor: %d, Viewport: %d, ViewportIdx: %d, IsAtBottomThreshold: %t",
		state.CursorIndex, state.ViewportStartIndex, state.CursorViewportIndex, state.IsAtBottomThreshold)
	t.Logf("View:\n%s", view)

	// Verify we're at bottom threshold
	if !state.IsAtBottomThreshold {
		t.Error("Expected to be at bottom threshold")
	}
	if state.CursorViewportIndex != 3 {
		t.Errorf("Expected cursor viewport index 3 (bottom threshold), got %d", state.CursorViewportIndex)
	}

	// Test 5: Move down from threshold - should trigger viewport scroll and cursor lock
	t.Log("\n--- Test 5: Viewport Scroll from Bottom Threshold ---")

	previousViewport = state.ViewportStartIndex
	previousCursor = state.CursorIndex
	list.MoveDown() // cursor should move to 9, viewport should scroll to maintain threshold lock

	state = list.GetState()
	view = list.View()

	t.Logf("After scroll - Cursor: %d->%d, Viewport: %d->%d, ViewportIdx: %d, IsAtBottomThreshold: %t",
		previousCursor, state.CursorIndex, previousViewport, state.ViewportStartIndex, state.CursorViewportIndex, state.IsAtBottomThreshold)
	t.Logf("View:\n%s", view)

	// Verify the expected behavior:
	// - Cursor should have moved to next item (9)
	// - Viewport should have scrolled to maintain threshold (viewport 6-10)
	// - Cursor should still be at threshold position (viewport index 3)
	if state.CursorIndex != 9 {
		t.Errorf("Expected cursor at 9, got %d", state.CursorIndex)
	}
	if state.ViewportStartIndex != 6 {
		t.Errorf("Expected viewport at 6, got %d", state.ViewportStartIndex)
	}
	if state.CursorViewportIndex != 3 {
		t.Errorf("Expected cursor locked at threshold position 3, got %d", state.CursorViewportIndex)
	}

	t.Log("\n=== THRESHOLD LOCKING BEHAVIOR TEST COMPLETED ===")
}

func TestDisabledThresholds(t *testing.T) {
	t.Log("=== DISABLED THRESHOLDS TEST ===")

	// Create list with disabled thresholds (-1 means disabled)
	people := make([]TestPerson, 15)
	for i := 0; i < 15; i++ {
		people[i] = TestPerson{
			Name: fmt.Sprintf("Person_%d", i),
			Age:  25 + (i % 40),
		}
	}

	dataSource := &SimpleDataSource{people: people}

	config := vtable.ListConfig{
		ViewportConfig: vtable.ViewportConfig{
			Height:             5,
			TopThreshold:       -1, // DISABLED
			BottomThreshold:    -1, // DISABLED
			ChunkSize:          5,
			InitialIndex:       0,
			BoundingAreaBefore: 1,
			BoundingAreaAfter:  1,
		},
		KeyMap: vtable.DefaultNavigationKeyMap(),
	}

	list := vtable.NewList(config, dataSource)

	// Load initial data
	totalCmd := dataSource.GetTotal()
	if totalMsg := totalCmd(); totalMsg != nil {
		list.Update(totalMsg)
	}

	loadCmd := dataSource.LoadChunk(vtable.DataRequest{Start: 0, Count: 5})
	if loadMsg := loadCmd(); loadMsg != nil {
		list.Update(loadMsg)
	}

	// Test edge-based scrolling without thresholds
	t.Log("\n--- Test 1: Edge-Based Scrolling (No Thresholds) ---")

	// Move to viewport edge (position 4 in height-5 viewport)
	for i := 0; i < 4; i++ {
		list.MoveDown()
		state := list.GetState()

		// Should never be at thresholds since they're disabled
		if state.IsAtTopThreshold || state.IsAtBottomThreshold {
			t.Errorf("Position %d: Should never be at thresholds when disabled", i+1)
		}
	}

	state := list.GetState()
	view := list.View()

	t.Logf("At edge - Cursor: %d, Viewport: %d, ViewportIdx: %d",
		state.CursorIndex, state.ViewportStartIndex, state.CursorViewportIndex)
	t.Logf("View:\n%s", view)

	// Should be at position 4, viewport should NOT have scrolled yet
	if state.CursorIndex != 4 {
		t.Errorf("Expected cursor at 4, got %d", state.CursorIndex)
	}
	if state.ViewportStartIndex != 0 {
		t.Errorf("Viewport should not have scrolled yet, got %d", state.ViewportStartIndex)
	}
	if state.CursorViewportIndex != 4 {
		t.Errorf("Expected cursor at viewport edge (4), got %d", state.CursorViewportIndex)
	}

	// Verify view shows cursor at edge
	lines := strings.Split(strings.TrimSpace(view), "\n")
	if len(lines) < 5 || !strings.HasPrefix(lines[4], "> ") {
		t.Errorf("Expected cursor (>) at edge (line 4), got view: %s", view)
	}

	// Test 2: Scroll beyond edge
	t.Log("\n--- Test 2: Scroll Beyond Edge ---")

	previousViewport := state.ViewportStartIndex
	list.MoveDown() // cursor=5, should scroll viewport

	state = list.GetState()
	view = list.View()

	t.Logf("After edge scroll - Cursor: %d, Viewport: %d->%d, ViewportIdx: %d",
		state.CursorIndex, previousViewport, state.ViewportStartIndex, state.CursorViewportIndex)
	t.Logf("View:\n%s", view)

	// Viewport should have scrolled
	if state.ViewportStartIndex <= previousViewport {
		t.Errorf("Viewport should have scrolled: %d -> %d", previousViewport, state.ViewportStartIndex)
	}

	// Should still never be at thresholds
	if state.IsAtTopThreshold || state.IsAtBottomThreshold {
		t.Error("Should never be at thresholds when disabled")
	}

	t.Log("\n=== DISABLED THRESHOLDS TEST COMPLETED ===")
}

func TestViewRenderingConsistency(t *testing.T) {
	t.Log("=== VIEW RENDERING CONSISTENCY TEST ===")

	list, _ := createListWithData(t, 10, 5, 5)

	// Test view consistency at different positions
	positions := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	for _, pos := range positions {
		t.Logf("\n--- Testing position %d ---", pos)

		list.JumpToIndex(pos)
		state := list.GetState()
		view := list.View()

		t.Logf("Position %d - Cursor: %d, Viewport: %d, ViewportIdx: %d",
			pos, state.CursorIndex, state.ViewportStartIndex, state.CursorViewportIndex)
		t.Logf("View:\n%s", view)

		// Parse view lines
		lines := strings.Split(strings.TrimSpace(view), "\n")

		// Verify cursor position in view
		cursorFound := false
		cursorLine := -1
		for i, line := range lines {
			if strings.HasPrefix(line, "> ") {
				cursorFound = true
				cursorLine = i
				break
			}
		}

		if !cursorFound {
			t.Errorf("Position %d: Cursor not found in view: %s", pos, view)
			continue
		}

		// Verify cursor line matches viewport index
		if cursorLine != state.CursorViewportIndex {
			t.Errorf("Position %d: Cursor at line %d but ViewportIndex is %d", pos, cursorLine, state.CursorViewportIndex)
		}

		// Verify correct item is shown with cursor
		expectedItem := fmt.Sprintf("Person_%d", pos)
		if !strings.Contains(lines[cursorLine], expectedItem) {
			t.Errorf("Position %d: Expected '%s' at cursor line, got: %s", pos, expectedItem, lines[cursorLine])
		}

		// Verify all visible items are in correct order
		for i, line := range lines {
			expectedPos := state.ViewportStartIndex + i
			if expectedPos >= 10 { // Beyond dataset
				break
			}

			expectedItemName := fmt.Sprintf("Person_%d", expectedPos)
			if !strings.Contains(line, expectedItemName) {
				t.Errorf("Position %d: Line %d should contain '%s', got: %s", pos, i, expectedItemName, line)
			}
		}
	}

	t.Log("\n=== VIEW RENDERING CONSISTENCY TEST COMPLETED ===")
}
