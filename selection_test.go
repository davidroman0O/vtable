package vtable

import (
	"fmt"
	"strings"
	"testing"
)

// ------------------------
// Test Data Provider
// ------------------------

type TestItem struct {
	ID   int
	Name string
}

type TestDataProvider struct {
	items     []TestItem
	selection map[int]bool
}

func NewTestDataProvider(count int) *TestDataProvider {
	items := make([]TestItem, count)
	for i := 0; i < count; i++ {
		items[i] = TestItem{ID: i, Name: fmt.Sprintf("Item %d", i)}
	}
	return &TestDataProvider{
		items:     items,
		selection: make(map[int]bool),
	}
}

func (p *TestDataProvider) GetTotal() int {
	return len(p.items)
}

func (p *TestDataProvider) GetItems(request DataRequest) ([]Data[TestItem], error) {
	start := request.Start
	count := request.Count

	if start >= len(p.items) {
		return []Data[TestItem]{}, nil
	}

	end := start + count
	if end > len(p.items) {
		end = len(p.items)
	}

	result := make([]Data[TestItem], end-start)
	for i := start; i < end; i++ {
		result[i-start] = Data[TestItem]{
			ID:       fmt.Sprintf("%d", p.items[i].ID),
			Item:     p.items[i],
			Selected: p.selection[i],
			Metadata: NewTypedMetadata(),
			Disabled: false,
			Hidden:   false,
		}
	}

	return result, nil
}

func (p *TestDataProvider) GetSelectionMode() SelectionMode {
	return SelectionMultiple
}

func (p *TestDataProvider) SetSelected(index int, selected bool) bool {
	if index < 0 || index >= len(p.items) {
		return false
	}
	if selected {
		p.selection[index] = true
	} else {
		delete(p.selection, index)
	}
	return true
}

func (p *TestDataProvider) SelectAll() bool {
	for i := 0; i < len(p.items); i++ {
		p.selection[i] = true
	}
	return true
}

func (p *TestDataProvider) ClearSelection() {
	p.selection = make(map[int]bool)
}

func (p *TestDataProvider) GetSelectedIndices() []int {
	indices := make([]int, 0, len(p.selection))
	for idx := range p.selection {
		indices = append(indices, idx)
	}
	return indices
}

func (p *TestDataProvider) GetItemID(item *TestItem) string {
	return fmt.Sprintf("%d", item.ID)
}

// Add missing methods for DataProvider interface
func (p *TestDataProvider) GetSelectedIDs() []string {
	ids := make([]string, 0, len(p.selection))
	for idx := range p.selection {
		if idx < len(p.items) {
			ids = append(ids, fmt.Sprintf("%d", p.items[idx].ID))
		}
	}
	return ids
}

func (p *TestDataProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	// Map IDs to indices and set their selection state
	for _, id := range ids {
		for i, item := range p.items {
			if fmt.Sprintf("%d", item.ID) == id {
				if selected {
					p.selection[i] = true
				} else {
					delete(p.selection, i)
				}
				break
			}
		}
	}
	return true
}

func (p *TestDataProvider) SelectRange(startID, endID string) bool {
	// Find start and end indices
	startIndex := -1
	endIndex := -1

	for i, item := range p.items {
		itemID := fmt.Sprintf("%d", item.ID)
		if itemID == startID {
			startIndex = i
		}
		if itemID == endID {
			endIndex = i
		}
	}

	if startIndex == -1 || endIndex == -1 {
		return false
	}

	// Ensure startIndex <= endIndex
	if startIndex > endIndex {
		startIndex, endIndex = endIndex, startIndex
	}

	// Select the range
	for i := startIndex; i <= endIndex; i++ {
		p.selection[i] = true
	}
	return true
}

// ------------------------
// Unit Tests
// ------------------------

func TestDataProviderSelection(t *testing.T) {
	provider := NewTestDataProvider(5)

	// Test initial state - no selections
	items, err := provider.GetItems(DataRequest{Start: 0, Count: 5})
	if err != nil {
		t.Fatalf("GetItems failed: %v", err)
	}

	for i, item := range items {
		if item.Selected {
			t.Errorf("Item %d should not be selected initially", i)
		}
	}

	// Test selecting an item
	if !provider.SetSelected(1, true) {
		t.Error("SetSelected should return true for valid index")
	}

	items, err = provider.GetItems(DataRequest{Start: 0, Count: 5})
	if err != nil {
		t.Fatalf("GetItems failed: %v", err)
	}

	if !items[1].Selected {
		t.Error("Item 1 should be selected after SetSelected(1, true)")
	}

	// Check other items are not selected
	for i, item := range items {
		if i != 1 && item.Selected {
			t.Errorf("Item %d should not be selected", i)
		}
	}
}

func TestSelectAll(t *testing.T) {
	provider := NewTestDataProvider(3)

	if !provider.SelectAll() {
		t.Error("SelectAll should return true")
	}

	items, err := provider.GetItems(DataRequest{Start: 0, Count: 3})
	if err != nil {
		t.Fatalf("GetItems failed: %v", err)
	}

	for i, item := range items {
		if !item.Selected {
			t.Errorf("Item %d should be selected after SelectAll", i)
		}
	}

	selectedIndices := provider.GetSelectedIndices()
	if len(selectedIndices) != 3 {
		t.Errorf("Expected 3 selected indices, got %d", len(selectedIndices))
	}
}

func TestClearSelection(t *testing.T) {
	provider := NewTestDataProvider(3)

	// Select all first
	provider.SelectAll()

	// Clear selection
	provider.ClearSelection()

	items, err := provider.GetItems(DataRequest{Start: 0, Count: 3})
	if err != nil {
		t.Fatalf("GetItems failed: %v", err)
	}

	for i, item := range items {
		if item.Selected {
			t.Errorf("Item %d should not be selected after ClearSelection", i)
		}
	}

	selectedIndices := provider.GetSelectedIndices()
	if len(selectedIndices) != 0 {
		t.Errorf("Expected 0 selected indices after clear, got %d", len(selectedIndices))
	}
}

func TestTeaListSelection(t *testing.T) {
	provider := NewTestDataProvider(10)

	config := ViewportConfig{
		Height:               5,
		TopThresholdIndex:    1,
		BottomThresholdIndex: 3,
		ChunkSize:            10,
		InitialIndex:         0,
		Debug:                false,
	}

	formatter := func(data Data[TestItem], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
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
		return fmt.Sprintf("%s%s", prefix, data.Item.Name)
	}

	styleConfig := StyleConfig{} // Default style

	list, err := NewTeaList(config, provider, styleConfig, formatter)
	if err != nil {
		t.Fatalf("NewTeaList failed: %v", err)
	}

	// Test toggle selection
	if !list.ToggleCurrentSelection() {
		t.Error("ToggleCurrentSelection should return true")
	}

	// Check that selection is reflected in rendered output
	view := list.View()
	if !containsSelectedIndicator(view) {
		t.Error("View should contain selection indicator after toggle")
	}

	// Test selection count
	count := list.GetSelectionCount()
	if count != 1 {
		t.Errorf("Expected selection count 1, got %d", count)
	}

	// Test clear selection
	list.ClearSelection()
	count = list.GetSelectionCount()
	if count != 0 {
		t.Errorf("Expected selection count 0 after clear, got %d", count)
	}
}

func TestTeaTableSelection(t *testing.T) {
	// Test table selection similar to list
	provider := &TableTestProvider{
		items:     []TestTableRow{{Cells: []string{"1", "Item 1"}}, {Cells: []string{"2", "Item 2"}}},
		selection: make(map[int]bool),
	}

	columns := []TableColumn{
		{Title: "ID", Width: 5, Alignment: AlignLeft},
		{Title: "Name", Width: 10, Alignment: AlignLeft},
	}

	config := TableConfig{
		Columns:     columns,
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: ViewportConfig{
			Height:               3,
			TopThresholdIndex:    0,
			BottomThresholdIndex: 1,
			ChunkSize:            10,
		},
	}

	theme := *DefaultTheme()
	table, err := NewTeaTable(config, provider, theme)
	if err != nil {
		t.Fatalf("NewTeaTable failed: %v", err)
	}

	// Test toggle selection
	if !table.ToggleCurrentSelection() {
		t.Error("ToggleCurrentSelection should return true")
	}

	count := table.GetSelectionCount()
	if count != 1 {
		t.Errorf("Expected selection count 1, got %d", count)
	}
}

func TestVisualSelectionRendering(t *testing.T) {
	fmt.Println("\n=== VISUAL SELECTION TEST ===")

	provider := NewTestDataProvider(5)

	config := ViewportConfig{
		Height:               5,
		TopThresholdIndex:    1,
		BottomThresholdIndex: 3,
		ChunkSize:            10,
		InitialIndex:         0,
		Debug:                false,
	}

	formatter := func(data Data[TestItem], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
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
		return fmt.Sprintf("%s%s", prefix, data.Item.Name)
	}

	styleConfig := StyleConfig{} // Default style

	list, err := NewTeaList(config, provider, styleConfig, formatter)
	if err != nil {
		t.Fatalf("NewTeaList failed: %v", err)
	}

	// Test 1: Initial state
	fmt.Println("\n1. Initial state (no selections):")
	view := list.View()
	fmt.Print(view)

	// Test 2: Toggle current selection
	fmt.Println("\n\n2. After toggling current selection:")
	success := list.ToggleCurrentSelection()
	fmt.Printf("ToggleCurrentSelection returned: %t\n", success)
	view = list.View()
	fmt.Print(view)

	// Test 3: Select multiple items
	fmt.Println("\n\n3. After selecting item 2 manually:")
	provider.SetSelected(2, true)
	list.RefreshData() // Force refresh
	view = list.View()
	fmt.Print(view)

	// Test 4: Select all
	fmt.Println("\n\n4. After select all:")
	list.SelectAll()
	view = list.View()
	fmt.Print(view)

	// Test 5: Clear selections
	fmt.Println("\n\n5. After clear selection:")
	list.ClearSelection()
	view = list.View()
	fmt.Print(view)

	fmt.Println("\n\n=== END VISUAL TEST ===")
}

func TestSelectionTrackingDuringScroll(t *testing.T) {
	fmt.Println("\n=== SELECTION TRACKING DURING SCROLL TEST ===")

	// Create a provider with 50 items
	provider := NewTestDataProvider(50)

	config := ViewportConfig{
		Height:               10, // Show 10 items at a time
		TopThresholdIndex:    2,
		BottomThresholdIndex: 7,
		ChunkSize:            20,
		InitialIndex:         0,
		Debug:                false,
	}

	formatter := func(data Data[TestItem], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
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
		return fmt.Sprintf("%s%d: %s", prefix, index, data.Item.Name)
	}

	styleConfig := StyleConfig{}
	list, err := NewTeaList(config, provider, styleConfig, formatter)
	if err != nil {
		t.Fatalf("NewTeaList failed: %v", err)
	}

	// Step 1: Select item at index 5 (should be "Item 5")
	fmt.Println("\n1. Selecting item at index 5:")
	provider.SetSelected(5, true)
	list.RefreshData()

	// Verify the selection
	selectedIndices := provider.GetSelectedIndices()
	if len(selectedIndices) != 1 || selectedIndices[0] != 5 {
		t.Errorf("Expected item 5 to be selected, got indices: %v", selectedIndices)
	}

	// Jump to show the selected item
	list.JumpToIndex(5)
	view := list.View()
	fmt.Print(view)

	// Verify that index 5 shows "Item 5" with selection marker
	if !strings.Contains(view, "✓>5: Item 5") && !strings.Contains(view, "✓ 5: Item 5") {
		t.Errorf("Expected to see '✓ 5: Item 5' or '✓>5: Item 5' in view, got:\n%s", view)
	}

	// Step 2: Scroll down significantly (simulate Page Down several times)
	fmt.Println("\n\n2. Scrolling down to index 25:")
	list.JumpToIndex(25)
	view = list.View()
	fmt.Print(view)

	// Verify item 5 is still selected (not item 25)
	selectedIndices = provider.GetSelectedIndices()
	if len(selectedIndices) != 1 || selectedIndices[0] != 5 {
		t.Errorf("After scrolling, expected item 5 to still be selected, got indices: %v", selectedIndices)
	}

	// Step 3: Scroll back to show the selected item
	fmt.Println("\n\n3. Scrolling back to show selected item (index 5):")
	list.JumpToIndex(5)
	view = list.View()
	fmt.Print(view)

	// Verify that index 5 still shows "Item 5" with selection
	if !strings.Contains(view, "✓>5: Item 5") && !strings.Contains(view, "✓ 5: Item 5") {
		t.Errorf("After scrolling back, expected to see '✓ 5: Item 5' in view, got:\n%s", view)
	}

	// Step 4: Test multiple selections
	fmt.Println("\n\n4. Adding more selections and scrolling:")
	provider.SetSelected(15, true)
	provider.SetSelected(35, true)
	list.RefreshData()

	// Check all selections are tracked correctly
	selectedIndices = provider.GetSelectedIndices()
	expectedSelections := map[int]bool{5: true, 15: true, 35: true}

	if len(selectedIndices) != 3 {
		t.Errorf("Expected 3 selections, got %d: %v", len(selectedIndices), selectedIndices)
	}

	for _, idx := range selectedIndices {
		if !expectedSelections[idx] {
			t.Errorf("Unexpected selection at index %d", idx)
		}
	}

	// Test viewing each selected item
	for _, idx := range []int{5, 15, 35} {
		fmt.Printf("\n\nViewing selected item at index %d:\n", idx)
		list.JumpToIndex(idx)
		view = list.View()
		fmt.Print(view)

		expectedText := fmt.Sprintf("%d: Item %d", idx, idx)
		if !strings.Contains(view, expectedText) {
			t.Errorf("Expected to see '%s' when viewing index %d, got:\n%s", expectedText, idx, view)
		}

		// Should have selection marker
		if !strings.Contains(view, "✓") {
			t.Errorf("Expected to see selection marker when viewing selected index %d, got:\n%s", idx, view)
		}
	}

	fmt.Println("\n\n=== END SELECTION TRACKING TEST ===")
}

func TestDataConsistencyDuringScroll(t *testing.T) {
	fmt.Println("\n=== DATA CONSISTENCY DURING SCROLL TEST ===")

	// Create a provider with 100 items
	provider := NewTestDataProvider(100)

	config := ViewportConfig{
		Height:               5,
		TopThresholdIndex:    1,
		BottomThresholdIndex: 3,
		ChunkSize:            10,
		InitialIndex:         0,
		Debug:                false,
	}

	formatter := func(data Data[TestItem], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		return fmt.Sprintf("%d: %s", index, data.Item.Name)
	}

	styleConfig := StyleConfig{}
	list, err := NewTeaList(config, provider, styleConfig, formatter)
	if err != nil {
		t.Fatalf("NewTeaList failed: %v", err)
	}

	// Test multiple scroll positions to ensure data consistency
	testPositions := []int{0, 10, 25, 50, 75, 95}

	for _, pos := range testPositions {
		fmt.Printf("\n\nTesting position %d:\n", pos)
		list.JumpToIndex(pos)
		view := list.View()
		fmt.Print(view)

		// Verify that the index matches the item content
		lines := strings.Split(strings.TrimSpace(view), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Extract index from the line (format: "index: Item X")
			if strings.Contains(line, ":") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					// Extract the index number (remove cursor indicators)
					indexStr := strings.TrimSpace(parts[0])
					indexStr = strings.ReplaceAll(indexStr, ">", "")
					indexStr = strings.ReplaceAll(indexStr, "✓", "")
					indexStr = strings.TrimSpace(indexStr)

					// Extract the expected item name
					contentPart := strings.TrimSpace(parts[1])

					// The content should be "Item X" where X matches the index
					expectedContent := fmt.Sprintf("Item %s", indexStr)

					if !strings.Contains(contentPart, expectedContent) {
						t.Errorf("Data mismatch at position %d: index %s should show '%s' but shows '%s'",
							pos, indexStr, expectedContent, contentPart)
						t.Errorf("Full line: '%s'", line)
					}
				}
			}
		}
	}

	fmt.Println("\n\n=== END DATA CONSISTENCY TEST ===")
}

// ------------------------
// Helper functions
// ------------------------

func containsSelectedIndicator(view string) bool {
	return strings.Contains(view, "✓") || strings.Contains(view, ">")
}

// Test table provider
type TestTableRow struct {
	ID    int // Add ID field
	Cells []string
}

type TableTestProvider struct {
	items     []TestTableRow
	selection map[int]bool
}

func (p *TableTestProvider) GetTotal() int {
	return len(p.items)
}

func (p *TableTestProvider) GetItems(request DataRequest) ([]Data[TableRow], error) {
	result := make([]Data[TableRow], len(p.items))
	for i, item := range p.items {
		result[i] = Data[TableRow]{
			ID:       p.GetItemID(&TableRow{Cells: item.Cells}),
			Item:     TableRow{Cells: item.Cells},
			Selected: p.selection[i],
			Metadata: NewTypedMetadata(),
			Disabled: false,
			Hidden:   false,
		}
	}
	return result, nil
}

func (p *TableTestProvider) GetSelectionMode() SelectionMode {
	return SelectionMultiple
}

func (p *TableTestProvider) SetSelected(index int, selected bool) bool {
	if index < 0 || index >= len(p.items) {
		return false
	}
	if selected {
		p.selection[index] = true
	} else {
		delete(p.selection, index)
	}
	return true
}

func (p *TableTestProvider) SelectAll() bool {
	for i := 0; i < len(p.items); i++ {
		p.selection[i] = true
	}
	return true
}

func (p *TableTestProvider) ClearSelection() {
	p.selection = make(map[int]bool)
}

func (p *TableTestProvider) GetSelectedIndices() []int {
	indices := make([]int, 0, len(p.selection))
	for idx := range p.selection {
		indices = append(indices, idx)
	}
	return indices
}

func (p *TableTestProvider) GetItemID(item *TableRow) string {
	if len(item.Cells) > 0 {
		return item.Cells[0]
	}
	return ""
}

// Add missing methods for DataProvider interface
func (p *TableTestProvider) GetSelectedIDs() []string {
	ids := make([]string, 0, len(p.selection))
	for idx := range p.selection {
		if idx < len(p.items) {
			ids = append(ids, p.GetItemID(&TableRow{Cells: p.items[idx].Cells}))
		}
	}
	return ids
}

func (p *TableTestProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	// Map IDs to indices and set their selection state
	for _, id := range ids {
		for i, item := range p.items {
			if p.GetItemID(&TableRow{Cells: item.Cells}) == id {
				if selected {
					p.selection[i] = true
				} else {
					delete(p.selection, i)
				}
				break
			}
		}
	}
	return true
}

func (p *TableTestProvider) SelectRange(startID, endID string) bool {
	// Find start and end indices
	startIndex := -1
	endIndex := -1

	for i, item := range p.items {
		itemID := p.GetItemID(&TableRow{Cells: item.Cells})
		if itemID == startID {
			startIndex = i
		}
		if itemID == endID {
			endIndex = i
		}
	}

	if startIndex == -1 || endIndex == -1 {
		return false
	}

	// Ensure startIndex <= endIndex
	if startIndex > endIndex {
		startIndex, endIndex = endIndex, startIndex
	}

	// Select the range
	for i := startIndex; i <= endIndex; i++ {
		p.selection[i] = true
	}
	return true
}

func TestDataIDFieldHandling(t *testing.T) {
	fmt.Println("\n=== DATA ID FIELD HANDLING TEST ===")

	// Create provider that DOES NOT populate ID field in GetItems
	provider := &SimpleTestProvider{
		items:     make([]TestItem, 20),
		selection: make(map[int]bool),
	}

	// Initialize items
	for i := 0; i < 20; i++ {
		provider.items[i] = TestItem{ID: i, Name: fmt.Sprintf("Item %d", i)}
	}

	config := ViewportConfig{
		Height:               5,
		TopThresholdIndex:    1,
		BottomThresholdIndex: 3,
		ChunkSize:            10,
		InitialIndex:         0,
		Debug:                false,
	}

	formatter := func(data Data[TestItem], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
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
		return fmt.Sprintf("%s%d: %s (ID:%s)", prefix, index, data.Item.Name, data.ID)
	}

	styleConfig := StyleConfig{}
	list, err := NewTeaList(config, provider, styleConfig, formatter)
	if err != nil {
		t.Fatalf("NewTeaList failed: %v", err)
	}

	// Test 1: Select an item and verify everything aligns
	fmt.Println("\n1. Selecting item at index 8:")
	provider.SetSelected(8, true)
	list.JumpToIndex(8)
	list.RefreshData()
	view := list.View()
	fmt.Print(view)

	// Verify proper alignment: index 8 should show "Item 8"
	if !strings.Contains(view, "8: Item 8") {
		t.Errorf("Expected to see '8: Item 8' in view when positioned at index 8, got:\n%s", view)
	}

	// Test 2: Scroll and verify selection persists
	fmt.Println("\n\n2. Scrolling to index 15:")
	list.JumpToIndex(15)
	view = list.View()
	fmt.Print(view)

	// Verify index 15 shows "Item 15"
	if !strings.Contains(view, "15: Item 15") {
		t.Errorf("Expected to see '15: Item 15' in view when positioned at index 15, got:\n%s", view)
	}

	// Verify selection is still on index 8
	selectedIndices := provider.GetSelectedIndices()
	if len(selectedIndices) != 1 || selectedIndices[0] != 8 {
		t.Errorf("Expected item 8 to still be selected, got indices: %v", selectedIndices)
	}

	// Test 3: Go back to selected item
	fmt.Println("\n\n3. Going back to selected item (index 8):")
	list.JumpToIndex(8)
	view = list.View()
	fmt.Print(view)

	// Should show selection marker at index 8 with "Item 8"
	if !strings.Contains(view, "✓>8: Item 8") && !strings.Contains(view, "✓ 8: Item 8") {
		t.Errorf("Expected to see selection marker with '8: Item 8', got:\n%s", view)
	}

	fmt.Println("\n\n=== END DATA ID FIELD HANDLING TEST ===")
}

// SimpleTestProvider that does NOT populate the ID field in Data structs
type SimpleTestProvider struct {
	items     []TestItem
	selection map[int]bool
}

func (p *SimpleTestProvider) GetTotal() int {
	return len(p.items)
}

func (p *SimpleTestProvider) GetItems(request DataRequest) ([]Data[TestItem], error) {
	start := request.Start
	count := request.Count

	if start >= len(p.items) {
		return []Data[TestItem]{}, nil
	}

	end := start + count
	if end > len(p.items) {
		end = len(p.items)
	}

	result := make([]Data[TestItem], end-start)
	for i := start; i < end; i++ {
		// INTENTIONALLY leaving ID empty - this is what examples should do
		result[i-start] = Data[TestItem]{
			ID:       "",
			Item:     p.items[i],
			Selected: p.selection[i],
			Metadata: NewTypedMetadata(),
			Disabled: false,
			Hidden:   false,
		}
	}

	return result, nil
}

func (p *SimpleTestProvider) GetSelectionMode() SelectionMode {
	return SelectionMultiple
}

func (p *SimpleTestProvider) SetSelected(index int, selected bool) bool {
	if index < 0 || index >= len(p.items) {
		return false
	}
	if selected {
		p.selection[index] = true
	} else {
		delete(p.selection, index)
	}
	return true
}

func (p *SimpleTestProvider) SelectAll() bool {
	for i := 0; i < len(p.items); i++ {
		p.selection[i] = true
	}
	return true
}

func (p *SimpleTestProvider) ClearSelection() {
	p.selection = make(map[int]bool)
}

func (p *SimpleTestProvider) GetSelectedIndices() []int {
	indices := make([]int, 0, len(p.selection))
	for idx := range p.selection {
		indices = append(indices, idx)
	}
	return indices
}

func (p *SimpleTestProvider) GetItemID(item *TestItem) string {
	return fmt.Sprintf("%d", item.ID)
}

func (p *SimpleTestProvider) GetSelectedIDs() []string {
	ids := make([]string, 0, len(p.selection))
	for idx := range p.selection {
		if idx < len(p.items) {
			ids = append(ids, fmt.Sprintf("%d", p.items[idx].ID))
		}
	}
	return ids
}

func (p *SimpleTestProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	for _, id := range ids {
		for i, item := range p.items {
			if fmt.Sprintf("%d", item.ID) == id {
				if selected {
					p.selection[i] = true
				} else {
					delete(p.selection, i)
				}
				break
			}
		}
	}
	return true
}

func (p *SimpleTestProvider) SelectRange(startID, endID string) bool {
	startIndex := -1
	endIndex := -1

	for i, item := range p.items {
		itemID := fmt.Sprintf("%d", item.ID)
		if itemID == startID {
			startIndex = i
		}
		if itemID == endID {
			endIndex = i
		}
	}

	if startIndex == -1 || endIndex == -1 {
		return false
	}

	if startIndex > endIndex {
		startIndex, endIndex = endIndex, startIndex
	}

	for i := startIndex; i <= endIndex; i++ {
		p.selection[i] = true
	}
	return true
}

func TestExactBugScenario(t *testing.T) {
	fmt.Println("\n=== EXACT BUG SCENARIO TEST ===")

	// Create a provider with enough items to test scrolling
	provider := NewTestDataProvider(50)

	config := ViewportConfig{
		Height:               10, // Show 10 items at a time
		TopThresholdIndex:    2,
		BottomThresholdIndex: 7,
		ChunkSize:            20,
		InitialIndex:         0,
		Debug:                false,
	}

	formatter := func(data Data[TestItem], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
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
		return fmt.Sprintf("%sIndex %d: %s", prefix, index, data.Item.Name)
	}

	styleConfig := StyleConfig{}
	list, err := NewTeaList(config, provider, styleConfig, formatter)
	if err != nil {
		t.Fatalf("NewTeaList failed: %v", err)
	}

	// Step 1: Scroll down to item 3
	fmt.Println("\n1. Scrolling down to item 3:")
	list.JumpToIndex(3)
	view := list.View()
	fmt.Print(view)

	// Verify we can see item 3 and it shows "Item 3"
	if !strings.Contains(view, "Index 3: Item 3") {
		t.Errorf("Expected to see 'Index 3: Item 3' when positioned at index 3, got:\n%s", view)
	}

	// Step 2: Read the View to check no selection yet
	fmt.Println("\n2. Checking that item 3 is not selected yet:")
	if strings.Contains(view, "✓") {
		t.Errorf("Expected no selection markers, but found them in:\n%s", view)
	}

	// Step 3: Select item 3
	fmt.Println("\n3. Selecting item 3:")
	provider.SetSelected(3, true)
	list.RefreshData()
	view = list.View()
	fmt.Print(view)

	// Step 4: Read the View to check we indeed have selection on item 3
	fmt.Println("\n4. Verifying item 3 is now selected:")
	if !strings.Contains(view, "✓>Index 3: Item 3") && !strings.Contains(view, "✓ Index 3: Item 3") {
		t.Errorf("Expected to see selection marker on 'Index 3: Item 3', got:\n%s", view)
	}

	selectedIndices := provider.GetSelectedIndices()
	if len(selectedIndices) != 1 || selectedIndices[0] != 3 {
		t.Errorf("Expected item 3 to be selected in provider, got indices: %v", selectedIndices)
	}

	// Step 5: Scroll down A LOT
	fmt.Println("\n5. Scrolling down A LOT to index 25:")
	list.JumpToIndex(25)
	view = list.View()
	fmt.Print(view)

	// Step 6: Read the View to check if we have the bug
	fmt.Println("\n6. Checking for the bug - verifying indices align with item names:")

	// Parse each line and verify index matches item name
	lines := strings.Split(strings.TrimSpace(view), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for pattern "Index X: Item Y"
		if strings.Contains(line, "Index") && strings.Contains(line, "Item") {
			// Extract the index number
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				// Get the index part (remove selection markers)
				indexPart := strings.TrimSpace(parts[0])
				indexPart = strings.ReplaceAll(indexPart, ">", "")
				indexPart = strings.ReplaceAll(indexPart, "✓", "")
				indexPart = strings.TrimSpace(indexPart)
				indexPart = strings.ReplaceAll(indexPart, "Index ", "")

				// Get the item part
				itemPart := strings.TrimSpace(parts[1])

				// The item should be "Item X" where X matches the index
				expectedItem := fmt.Sprintf("Item %s", indexPart)

				if itemPart != expectedItem {
					t.Errorf("BUG DETECTED! Index %s shows '%s' but should show '%s'",
						indexPart, itemPart, expectedItem)
					t.Errorf("Full line: '%s'", line)
					fmt.Printf("🐛 BUG FOUND: Index %s shows '%s' but should show '%s'\n",
						indexPart, itemPart, expectedItem)
				} else {
					fmt.Printf("✅ Correct: Index %s shows '%s'\n", indexPart, itemPart)
				}
			}
		}
	}

	// Also verify our original selection is still intact
	selectedIndices = provider.GetSelectedIndices()
	if len(selectedIndices) != 1 || selectedIndices[0] != 3 {
		t.Errorf("Selection tracking broken! Expected item 3 to still be selected, got indices: %v", selectedIndices)
	} else {
		fmt.Printf("✅ Selection tracking intact: item 3 still selected\n")
	}

	// Go back to item 3 to double-check
	fmt.Println("\n7. Going back to item 3 to verify it's still selected:")
	list.JumpToIndex(3)
	view = list.View()
	fmt.Print(view)

	if !strings.Contains(view, "✓>Index 3: Item 3") && !strings.Contains(view, "✓ Index 3: Item 3") {
		t.Errorf("When returning to item 3, expected selection marker on 'Index 3: Item 3', got:\n%s", view)
	} else {
		fmt.Printf("✅ Item 3 still shows as selected correctly\n")
	}

	fmt.Println("\n=== END EXACT BUG SCENARIO TEST ===")
}
