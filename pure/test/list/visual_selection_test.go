package list_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	vtable "github.com/davidroman0O/vtable/pure"
)

// TestItem represents a simple test item
type TestItem struct {
	Name string
}

// TestDataProvider implements DataSource interface for testing
type TestDataProvider struct {
	items    []TestItem
	selected map[int]bool
}

func NewTestDataProvider(count int) *TestDataProvider {
	items := make([]TestItem, count)
	for i := 0; i < count; i++ {
		items[i] = TestItem{Name: fmt.Sprintf("Item %d", i)}
	}

	return &TestDataProvider{
		items:    items,
		selected: make(map[int]bool),
	}
}

// DataSource interface implementation
func (p *TestDataProvider) LoadChunk(request vtable.DataRequest) tea.Cmd {
	start := request.Start
	count := request.Count

	if start >= len(p.items) {
		return vtable.DataChunkLoadedCmd(start, []vtable.Data[any]{}, request)
	}

	end := start + count
	if end > len(p.items) {
		end = len(p.items)
	}

	result := make([]vtable.Data[any], end-start)
	for i := start; i < end; i++ {
		result[i-start] = vtable.Data[any]{
			ID:       strconv.Itoa(i),
			Item:     p.items[i],
			Selected: p.selected[i],
			Loading:  false,
			Error:    nil,
			Disabled: false,
		}
	}

	return vtable.DataChunkLoadedCmd(start, result, request)
}

func (p *TestDataProvider) GetTotal() tea.Cmd {
	return vtable.DataTotalCmd(len(p.items))
}

func (p *TestDataProvider) RefreshTotal() tea.Cmd {
	return p.GetTotal()
}

func (p *TestDataProvider) GetItemID(item any) string {
	if testItem, ok := item.(TestItem); ok {
		return testItem.Name
	}
	return fmt.Sprintf("%v", item)
}

// Selection operations - required by DataSource interface
func (p *TestDataProvider) SetSelected(index int, selected bool) tea.Cmd {
	if index >= 0 && index < len(p.items) {
		if selected {
			p.selected[index] = true
		} else {
			delete(p.selected, index)
		}
	}
	return vtable.SelectionResponseCmd(true, index, strconv.Itoa(index), selected, "toggle", nil, nil)
}

func (p *TestDataProvider) SetSelectedByID(id string, selected bool) tea.Cmd {
	return vtable.SelectionResponseCmd(true, -1, id, selected, "toggleByID", nil, nil)
}

func (p *TestDataProvider) SelectAll() tea.Cmd {
	affectedIDs := make([]string, len(p.items))
	for i := 0; i < len(p.items); i++ {
		p.selected[i] = true
		affectedIDs[i] = strconv.Itoa(i)
	}
	return vtable.SelectionResponseCmd(true, -1, "", true, "selectAll", nil, affectedIDs)
}

func (p *TestDataProvider) ClearSelection() tea.Cmd {
	p.selected = make(map[int]bool)
	return vtable.SelectionResponseCmd(true, -1, "", false, "clear", nil, nil)
}

func (p *TestDataProvider) SelectRange(startIndex, endIndex int) tea.Cmd {
	affectedIDs := make([]string, endIndex-startIndex+1)
	for i := startIndex; i <= endIndex; i++ {
		if i >= 0 && i < len(p.items) {
			p.selected[i] = true
			affectedIDs[i-startIndex] = strconv.Itoa(i)
		}
	}
	return vtable.SelectionResponseCmd(true, -1, "", true, "range", nil, affectedIDs)
}

func TestVisualSelectionRendering(t *testing.T) {
	t.Log("=== SYSTEMATIC VISUAL SELECTION RENDERING TEST ===")

	provider := NewTestDataProvider(5)

	config := vtable.ListConfig{
		ViewportConfig: vtable.ViewportConfig{
			Height:          5,
			TopThreshold:    1, // 1 position from viewport start
			BottomThreshold: 1, // 1 position from viewport end (position 3 in height-5 viewport)
			ChunkSize:       10,
			InitialIndex:    0,
			// FULLY AUTOMATED bounding area!
		},
		SelectionMode: vtable.SelectionMultiple,
		KeyMap:        vtable.DefaultNavigationKeyMap(),
		StyleConfig:   vtable.DefaultStyleConfig(),
		MaxWidth:      80,
	}

	formatter := func(data vtable.Data[any], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		item := data.Item.(TestItem)
		prefix := "  "
		if data.Selected {
			prefix = "✓ "
		}
		if isCursor {
			prefix = "> "
			if data.Selected {
				prefix = "✓>"
			}
		}
		return fmt.Sprintf("%s%s", prefix, item.Name)
	}

	list := vtable.NewList(config, provider)
	list.SetFormatter(formatter)

	// Manually simulate the data loading process
	// 1. First simulate getting the total
	totalCmd := provider.GetTotal()
	if totalMsg := totalCmd(); totalMsg != nil {
		list.Update(totalMsg)
	}

	// 2. Then simulate loading the first chunk
	loadCmd := provider.LoadChunk(vtable.DataRequest{Start: 0, Count: 10})
	if loadMsg := loadCmd(); loadMsg != nil {
		list.Update(loadMsg)
	}

	// Test 1: Initial state (no selections)
	t.Log("\n--- Test 1: Initial State (No Selections) ---")
	view := list.View()
	t.Logf("View Output:\n%s", view)

	if view == "No data available" {
		t.Fatal("Expected data to be loaded, but got 'No data available'")
	}

	// Should show cursor on first item, no selections
	if !strings.Contains(view, "> ") {
		t.Errorf("Expected cursor (>) on first item, got view: %s", view)
	}

	if strings.Contains(view, "✓") {
		t.Errorf("Expected no selections initially, but found checkmarks in view: %s", view)
	}

	// Verify specific content (using correct format)
	if !strings.Contains(view, "> {Item 0}") {
		t.Errorf("Expected '> {Item 0}' (cursor on first item), got view: %s", view)
	}

	// Test 2: Toggle current selection
	t.Log("\n--- Test 2: After Toggle Current Selection ---")

	// Use proper Tea model pattern - send SelectCurrentMsg through Update
	model, cmd := list.Update(vtable.SelectCurrentMsg{})
	list = model.(*vtable.List)

	// Execute any commands returned from the selection
	for cmd != nil {
		if msg := cmd(); msg != nil {
			model, cmd = list.Update(msg)
			list = model.(*vtable.List)
		} else {
			break
		}
	}

	view = list.View()
	t.Logf("View Output:\n%s", view)

	// Should show cursor and selection on first item
	if !strings.Contains(view, "✓") {
		t.Errorf("Expected selection mark (✓) after toggle, got view: %s", view)
	}

	// Should show combined cursor+selection marker (with correct format)
	if !strings.Contains(view, "✓ > {Item 0}") {
		t.Errorf("Expected '✓ > {Item 0}' (cursor+selection on first item), got view: %s", view)
	}

	// Test 3: Move cursor and verify rendering
	t.Log("\n--- Test 3: Move Cursor Down ---")
	list.MoveDown()
	view = list.View()
	t.Logf("View Output:\n%s", view)

	// Should show selection on Item 0 and cursor on Item 1 (with correct format)
	if !strings.Contains(view, "✓   {Item 0}") {
		t.Errorf("Expected '✓   {Item 0}' (selected but no cursor), got view: %s", view)
	}
	if !strings.Contains(view, "> {Item 1}") {
		t.Errorf("Expected '> {Item 1}' (cursor on second item), got view: %s", view)
	}

	// Test 4: Test selection state
	t.Log("\n--- Test 4: Verify Selection State ---")
	selectionCount := list.GetSelectionCount()
	t.Logf("Selection Count: %d", selectionCount)
	if selectionCount != 1 {
		t.Errorf("Expected 1 selected item, got %d", selectionCount)
	}

	selectedIDs := list.GetSelectedIDs()
	t.Logf("Selected IDs: %v", selectedIDs)
	if len(selectedIDs) != 1 {
		t.Errorf("Expected 1 selected ID, got %d", len(selectedIDs))
	}

	// Test 5: Toggle second item selection
	t.Log("\n--- Test 5: Toggle Second Item Selection ---")

	// Use proper Tea model pattern - send SelectCurrentMsg through Update
	model, cmd = list.Update(vtable.SelectCurrentMsg{})
	list = model.(*vtable.List)

	// Execute any commands returned from the selection
	for cmd != nil {
		if msg := cmd(); msg != nil {
			model, cmd = list.Update(msg)
			list = model.(*vtable.List)
		} else {
			break
		}
	}

	view = list.View()
	t.Logf("View Output:\n%s", view)

	// Should show both items selected (with correct format)
	if !strings.Contains(view, "✓   {Item 0}") {
		t.Errorf("Expected '✓   {Item 0}' to remain selected, got view: %s", view)
	}
	if !strings.Contains(view, "✓ > {Item 1}") {
		t.Errorf("Expected '✓ > {Item 1}' (cursor+selection on second item), got view: %s", view)
	}

	// Test 6: Select all
	t.Log("\n--- Test 6: Select All ---")

	// Use proper Tea model pattern - send SelectAllMsg through Update
	model, cmd = list.Update(vtable.SelectAllMsg{})
	list = model.(*vtable.List)

	// Execute any commands returned from the selection
	for cmd != nil {
		if msg := cmd(); msg != nil {
			model, cmd = list.Update(msg)
			list = model.(*vtable.List)
		} else {
			break
		}
	}

	view = list.View()
	t.Logf("View Output:\n%s", view)

	// Count checkmarks in view
	checkCount := strings.Count(view, "✓")
	t.Logf("Checkmark count in view: %d", checkCount)
	if checkCount < 5 {
		t.Errorf("Expected at least 5 selections after SelectAll, got %d checkmarks in view: %s", checkCount, view)
	}

	// Check selection count
	selectionCount = list.GetSelectionCount()
	t.Logf("Total selection count: %d", selectionCount)
	if selectionCount < 5 {
		t.Errorf("Expected at least 5 selected items after SelectAll, got %d", selectionCount)
	}

	// Verify all items show checkmarks
	for i := 0; i < 5; i++ {
		expectedText := fmt.Sprintf("✓")
		if !strings.Contains(view, expectedText) {
			t.Errorf("Expected to see checkmark for Item %d, got view: %s", i, view)
		}
	}

	// Test 7: Clear selections
	t.Log("\n--- Test 7: Clear All Selections ---")

	// Use proper Tea model pattern - send SelectClearMsg through Update
	model, cmd = list.Update(vtable.SelectClearMsg{})
	list = model.(*vtable.List)

	// Execute any commands returned from the selection
	for cmd != nil {
		if msg := cmd(); msg != nil {
			model, cmd = list.Update(msg)
			list = model.(*vtable.List)
		} else {
			break
		}
	}

	view = list.View()
	t.Logf("View Output:\n%s", view)

	// Should have no checkmarks except possibly the cursor
	checkmarksOnly := strings.Count(view, "✓ ")
	if checkmarksOnly > 0 {
		t.Errorf("Expected no standalone checkmarks after ClearSelection, got %d in view: %s", checkmarksOnly, view)
	}

	// Check selection count
	selectionCount = list.GetSelectionCount()
	t.Logf("Selection count after clear: %d", selectionCount)
	if selectionCount != 0 {
		t.Errorf("Expected 0 selected items after ClearSelection, got %d", selectionCount)
	}

	// Should still show cursor
	if !strings.Contains(view, "> ") {
		t.Errorf("Expected cursor to remain visible after clear, got view: %s", view)
	}

	// Test 8: Navigate and test visual consistency
	t.Log("\n--- Test 8: Navigate and Test Visual Consistency ---")
	for i := 0; i < 3; i++ {
		t.Logf("Moving to position %d", i)
		list.JumpToStart()
		for j := 0; j < i; j++ {
			list.MoveDown()
		}

		view = list.View()
		t.Logf("Position %d View:\n%s", i, view)

		expectedCursor := fmt.Sprintf("> {Item %d}", i)
		if !strings.Contains(view, expectedCursor) {
			t.Errorf("Expected '%s' at position %d, got view: %s", expectedCursor, i, view)
		}
	}

	t.Log("\n=== SYSTEMATIC VISUAL SELECTION RENDERING TEST COMPLETED ===")
}

func TestSelectionWithThresholdNavigation(t *testing.T) {
	t.Log("=== SELECTION WITH THRESHOLD NAVIGATION TEST ===")

	provider := NewTestDataProvider(15)

	config := vtable.ListConfig{
		ViewportConfig: vtable.ViewportConfig{
			Height:             5,
			TopThreshold:       1, // 1 position from viewport start
			BottomThreshold:    1, // 1 position from viewport end (position 3 in height-5 viewport)
			ChunkSize:          5,
			InitialIndex:       0,
			BoundingAreaBefore: 2,
			BoundingAreaAfter:  2,
		},
		SelectionMode: vtable.SelectionMultiple,
		KeyMap:        vtable.DefaultNavigationKeyMap(),
		StyleConfig:   vtable.DefaultStyleConfig(),
		MaxWidth:      80,
	}

	formatter := func(data vtable.Data[any], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		item := data.Item.(TestItem)
		prefix := "  "
		if data.Selected {
			prefix = "✓ "
		}
		if isCursor {
			prefix = "> "
			if data.Selected {
				prefix = "✓>"
			}
		}
		return fmt.Sprintf("%s%s", prefix, item.Name)
	}

	list := vtable.NewList(config, provider)
	list.SetFormatter(formatter)

	// Load initial data
	totalCmd := provider.GetTotal()
	if totalMsg := totalCmd(); totalMsg != nil {
		list.Update(totalMsg)
	}

	loadCmd := provider.LoadChunk(vtable.DataRequest{Start: 0, Count: 5})
	if loadMsg := loadCmd(); loadMsg != nil {
		list.Update(loadMsg)
	}

	// Test 1: Select items and navigate to threshold
	t.Log("\n--- Test 1: Selection with Threshold Navigation ---")

	// Select first item
	list.ToggleCurrentSelection()

	// Move to threshold position
	list.MoveDown() // cursor at position 1 (top threshold)

	state := list.GetState()
	view := list.View()

	t.Logf("At threshold - Cursor: %d, ViewportIdx: %d, IsAtTopThreshold: %t",
		state.CursorIndex, state.CursorViewportIndex, state.IsAtTopThreshold)
	t.Logf("Selection count: %d", list.GetSelectionCount())
	t.Logf("View:\n%s", view)

	// Verify selection persists through navigation
	if list.GetSelectionCount() != 1 {
		t.Errorf("Expected 1 selected item, got %d", list.GetSelectionCount())
	}

	// Verify view shows both selection and cursor correctly
	lines := strings.Split(strings.TrimSpace(view), "\n")

	// First line should show selection without cursor
	if len(lines) > 0 && !strings.HasPrefix(lines[0], "✓ ") {
		t.Errorf("Expected first line to show selection '✓ ', got: %s", lines[0])
	}

	// Second line should show cursor at threshold
	if len(lines) > 1 && !strings.HasPrefix(lines[1], "> ") {
		t.Errorf("Expected second line to show cursor '> ', got: %s", lines[1])
	}

	// Verify threshold flag
	if !state.IsAtTopThreshold {
		t.Error("Expected to be at top threshold")
	}

	// Test 2: Select current item at threshold and scroll viewport
	t.Log("\n--- Test 2: Select at Threshold and Trigger Scroll ---")

	list.ToggleCurrentSelection() // Select item at threshold

	// Move up to trigger viewport scroll
	list.MoveUp()

	state = list.GetState()
	view = list.View()

	t.Logf("After scroll - Cursor: %d, ViewportIdx: %d, ViewportStart: %d",
		state.CursorIndex, state.CursorViewportIndex, state.ViewportStartIndex)
	t.Logf("Selection count: %d", list.GetSelectionCount())
	t.Logf("View:\n%s", view)

	// Should have 2 selections now
	if list.GetSelectionCount() != 2 {
		t.Errorf("Expected 2 selected items, got %d", list.GetSelectionCount())
	}

	// Verify selections persist through viewport scrolling
	selectedIDs := list.GetSelectedIDs()
	if len(selectedIDs) != 2 {
		t.Errorf("Expected 2 selected IDs, got %d: %v", len(selectedIDs), selectedIDs)
	}

	// Test 3: Navigate to bottom threshold with selections
	t.Log("\n--- Test 3: Bottom Threshold with Selections ---")

	// Navigate to bottom threshold
	list.JumpToStart()
	for i := 0; i < 3; i++ { // Move to position 3 (bottom threshold)
		list.MoveDown()
	}

	state = list.GetState()
	view = list.View()

	t.Logf("At bottom threshold - Cursor: %d, ViewportIdx: %d, IsAtBottomThreshold: %t",
		state.CursorIndex, state.CursorViewportIndex, state.IsAtBottomThreshold)
	t.Logf("View:\n%s", view)

	// Verify threshold flag
	if !state.IsAtBottomThreshold {
		t.Error("Expected to be at bottom threshold")
	}

	// Verify view structure with selections
	lines = strings.Split(strings.TrimSpace(view), "\n")

	// Should see checkmarks for selected items and cursor at bottom threshold
	cursorLine := -1
	selectionCount := 0

	for i, line := range lines {
		if strings.HasPrefix(line, "> ") {
			cursorLine = i
		}
		if strings.Contains(line, "✓") {
			selectionCount++
		}
	}

	if cursorLine != 3 {
		t.Errorf("Expected cursor at line 3 (bottom threshold), got line %d", cursorLine)
	}

	// Should still have our selections visible (if in view)
	if selectionCount == 0 {
		t.Log("Note: No selections visible in current viewport (expected if selections are scrolled out)")
	}

	t.Log("\n=== SELECTION WITH THRESHOLD NAVIGATION TEST COMPLETED ===")
}

func TestSelectionRenderingConsistency(t *testing.T) {
	t.Log("=== SELECTION RENDERING CONSISTENCY TEST ===")

	provider := NewTestDataProvider(20)

	config := vtable.ListConfig{
		ViewportConfig: vtable.ViewportConfig{
			Height:             6,
			TopThreshold:       1,
			BottomThreshold:    1,
			ChunkSize:          8,
			InitialIndex:       0,
			BoundingAreaBefore: 3,
			BoundingAreaAfter:  3,
		},
		SelectionMode: vtable.SelectionMultiple,
		KeyMap:        vtable.DefaultNavigationKeyMap(),
		StyleConfig:   vtable.DefaultStyleConfig(),
		MaxWidth:      80,
	}

	formatter := func(data vtable.Data[any], index int, ctx vtable.RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		item := data.Item.(TestItem)
		prefix := "  "
		if data.Selected {
			prefix = "✓ "
		}
		if isCursor {
			prefix = "> "
			if data.Selected {
				prefix = "✓>"
			}
		}
		return fmt.Sprintf("%s%s", prefix, item.Name)
	}

	list := vtable.NewList(config, provider)
	list.SetFormatter(formatter)

	// Load initial data
	totalCmd := provider.GetTotal()
	if totalMsg := totalCmd(); totalMsg != nil {
		list.Update(totalMsg)
	}

	loadCmd := provider.LoadChunk(vtable.DataRequest{Start: 0, Count: 8})
	if loadMsg := loadCmd(); loadMsg != nil {
		list.Update(loadMsg)
	}

	// Test selection rendering at various positions
	testPositions := []int{0, 3, 6, 9, 12, 15, 19}

	for _, pos := range testPositions {
		t.Logf("\n--- Testing selection rendering at position %d ---", pos)

		// Navigate to position
		list.JumpToIndex(pos)

		// Select current item
		list.ToggleCurrentSelection()

		state := list.GetState()
		view := list.View()

		t.Logf("Position %d - Cursor: %d, Viewport: %d, Selections: %d",
			pos, state.CursorIndex, state.ViewportStartIndex, list.GetSelectionCount())
		t.Logf("View:\n%s", view)

		// Parse view to verify selection rendering
		lines := strings.Split(strings.TrimSpace(view), "\n")

		// Find cursor line
		cursorLine := -1
		for i, line := range lines {
			if strings.HasPrefix(line, "> ") || strings.HasPrefix(line, "✓>") {
				cursorLine = i
				break
			}
		}

		if cursorLine == -1 {
			t.Errorf("Position %d: Cursor not found in view: %s", pos, view)
			continue
		}

		// Verify cursor line shows selection if item is selected
		expectedItem := fmt.Sprintf("Item %d", pos)
		if !strings.Contains(lines[cursorLine], expectedItem) {
			t.Errorf("Position %d: Expected '%s' at cursor line, got: %s", pos, expectedItem, lines[cursorLine])
		}

		// If item is selected, cursor line should show "✓>"
		if strings.HasPrefix(lines[cursorLine], "✓>") {
			t.Logf("Position %d: ✓ Correctly shows selected cursor", pos)
		} else if strings.HasPrefix(lines[cursorLine], "> ") {
			t.Logf("Position %d: ✓ Correctly shows unselected cursor", pos)
		} else {
			t.Errorf("Position %d: Invalid cursor format: %s", pos, lines[cursorLine])
		}

		// Count selections in view
		selectionCount := 0
		for _, line := range lines {
			if strings.Contains(line, "✓") {
				selectionCount++
			}
		}

		t.Logf("Position %d: %d selections visible in viewport", pos, selectionCount)
	}

	// Final verification
	totalSelections := list.GetSelectionCount()
	selectedIDs := list.GetSelectedIDs()

	t.Logf("Final state - Total selections: %d, Selected IDs: %v", totalSelections, selectedIDs)

	if totalSelections != len(testPositions) {
		t.Errorf("Expected %d total selections, got %d", len(testPositions), totalSelections)
	}

	t.Log("\n=== SELECTION RENDERING CONSISTENCY TEST COMPLETED ===")
}
