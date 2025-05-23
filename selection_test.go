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
			Item:     p.items[i],
			Selected: p.selection[i],
			Metadata: nil,
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

	table, err := NewTeaTable(config, provider, DefaultTheme())
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

// ------------------------
// Helper functions
// ------------------------

func containsSelectedIndicator(view string) bool {
	return strings.Contains(view, "✓") || strings.Contains(view, ">")
}

// Test table provider
type TestTableRow struct {
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
			Item:     TableRow{Cells: item.Cells},
			Selected: p.selection[i],
			Metadata: nil,
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
