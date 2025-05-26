package vtable

import (
	"fmt"
	"testing"
)

// TestGetCachedTotal verifies that GetCachedTotal() returns cached values without triggering data provider calls
func TestGetCachedTotal(t *testing.T) {
	// Create a provider that tracks GetTotal() calls
	provider := &TrackingDataProvider{
		items:         generateTestItems(100),
		getTotalCalls: 0,
		filters:       make(map[string]any),
	}

	// Create list
	config := ViewportConfig{
		Height:               5,
		ChunkSize:            10,
		TopThresholdIndex:    1,
		BottomThresholdIndex: 3,
	}

	list, err := NewList(config, provider, DefaultStyleConfig(), simpleFormatter)
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Initial creation should call GetTotal() once
	if provider.getTotalCalls != 1 {
		t.Errorf("Expected 1 GetTotal() call after creation, got %d", provider.getTotalCalls)
	}

	// Reset counter for testing
	provider.getTotalCalls = 0

	// Test 1: GetCachedTotal() should NOT call GetTotal()
	t.Log("=== Testing GetCachedTotal() ===")

	cachedTotal := list.GetCachedTotal()
	if cachedTotal != 100 {
		t.Errorf("Expected cached total of 100, got %d", cachedTotal)
	}

	if provider.getTotalCalls != 0 {
		t.Errorf("GetCachedTotal() should not call GetTotal(), but got %d calls", provider.getTotalCalls)
	}
	t.Logf("✅ GetCachedTotal(): returned %d with 0 GetTotal() calls", cachedTotal)

	// Test 2: Multiple GetCachedTotal() calls should still not trigger provider calls
	for i := 0; i < 10; i++ {
		total := list.GetCachedTotal()
		if total != 100 {
			t.Errorf("Expected cached total of 100 on call %d, got %d", i, total)
		}
	}

	if provider.getTotalCalls != 0 {
		t.Errorf("Multiple GetCachedTotal() calls should not call GetTotal(), but got %d calls", provider.getTotalCalls)
	}
	t.Log("✅ Multiple GetCachedTotal() calls: 0 GetTotal() calls")

	// Test 3: Apply filter - this should invalidate cache and call GetTotal() once
	list.SetFilter("name", "Item 5")

	// Now GetCachedTotal() should return the updated cached value without additional calls
	provider.getTotalCalls = 0 // Reset after filter operation

	cachedTotal = list.GetCachedTotal()
	if provider.getTotalCalls != 0 {
		t.Errorf("GetCachedTotal() after filter should not call GetTotal(), but got %d calls", provider.getTotalCalls)
	}
	t.Logf("✅ GetCachedTotal() after filter: returned %d with 0 GetTotal() calls", cachedTotal)

	// Test 4: Cache invalidation - GetCachedTotal() should still return cached value
	list.InvalidateTotalItemsCache()

	cachedTotal = list.GetCachedTotal()
	if provider.getTotalCalls != 0 {
		t.Errorf("GetCachedTotal() after cache invalidation should not call GetTotal(), but got %d calls", provider.getTotalCalls)
	}
	t.Logf("✅ GetCachedTotal() after invalidation: returned %d with 0 GetTotal() calls", cachedTotal)

	t.Log("=== GetCachedTotal() Test Complete ===")
	t.Log("🚀 GetCachedTotal() provides efficient access to total count without data provider overhead")
}

// TestTeaListGetCachedTotal verifies GetCachedTotal() works with TeaList
func TestTeaListGetCachedTotal(t *testing.T) {
	provider := &TrackingDataProvider{
		items:         generateTestItems(50),
		getTotalCalls: 0,
		filters:       make(map[string]any),
	}

	config := ViewportConfig{
		Height:               5,
		ChunkSize:            10,
		TopThresholdIndex:    1,
		BottomThresholdIndex: 3,
	}

	teaList, err := NewTeaList(config, provider, DefaultStyleConfig(), simpleFormatter)
	if err != nil {
		t.Fatalf("Failed to create TeaList: %v", err)
	}

	// Reset counter after creation
	provider.getTotalCalls = 0

	// Test GetCachedTotal() on TeaList
	cachedTotal := teaList.GetCachedTotal()
	if cachedTotal != 50 {
		t.Errorf("Expected cached total of 50, got %d", cachedTotal)
	}

	if provider.getTotalCalls != 0 {
		t.Errorf("TeaList.GetCachedTotal() should not call GetTotal(), but got %d calls", provider.getTotalCalls)
	}

	t.Logf("✅ TeaList.GetCachedTotal(): returned %d with 0 GetTotal() calls", cachedTotal)
}

// TestTeaTableGetCachedTotal verifies GetCachedTotal() works with TeaTable
func TestTeaTableGetCachedTotal(t *testing.T) {
	// Create provider for TableRow
	provider := &TableRowTrackingProvider{
		items:         generateTestTableRows(30),
		getTotalCalls: 0,
		filters:       make(map[string]any),
	}

	config := TableConfig{
		Columns: []TableColumn{
			{Title: "Name", Width: 20, Field: "name"},
			{Title: "ID", Width: 5, Field: "id"},
		},
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: ViewportConfig{
			Height:               5,
			TopThresholdIndex:    1,
			BottomThresholdIndex: 3,
			ChunkSize:            10,
		},
	}

	teaTable, err := NewTeaTable(config, provider, *DefaultTheme())
	if err != nil {
		t.Fatalf("Failed to create TeaTable: %v", err)
	}

	// Reset counter after creation
	provider.getTotalCalls = 0

	// Test GetCachedTotal() on TeaTable
	cachedTotal := teaTable.GetCachedTotal()
	if cachedTotal != 30 {
		t.Errorf("Expected cached total of 30, got %d", cachedTotal)
	}

	if provider.getTotalCalls != 0 {
		t.Errorf("TeaTable.GetCachedTotal() should not call GetTotal(), but got %d calls", provider.getTotalCalls)
	}

	t.Logf("✅ TeaTable.GetCachedTotal(): returned %d with 0 GetTotal() calls", cachedTotal)
}

// TableRowTrackingProvider for testing TeaTable
type TableRowTrackingProvider struct {
	items         []TableRow
	getTotalCalls int
	filters       map[string]any
}

func (p *TableRowTrackingProvider) GetTotal() int {
	p.getTotalCalls++
	return len(p.items)
}

func (p *TableRowTrackingProvider) GetItems(request DataRequest) ([]Data[TableRow], error) {
	p.filters = request.Filters

	start := request.Start
	end := start + request.Count
	if start >= len(p.items) {
		return []Data[TableRow]{}, nil
	}
	if end > len(p.items) {
		end = len(p.items)
	}

	result := make([]Data[TableRow], end-start)
	for i := start; i < end; i++ {
		result[i-start] = Data[TableRow]{
			ID:       fmt.Sprintf("row-%d", i),
			Item:     p.items[i],
			Selected: false,
			Metadata: NewTypedMetadata(),
		}
	}

	return result, nil
}

func (p *TableRowTrackingProvider) GetSelectionMode() SelectionMode                   { return SelectionMultiple }
func (p *TableRowTrackingProvider) SetSelected(index int, selected bool) bool         { return true }
func (p *TableRowTrackingProvider) SetSelectedByIDs(ids []string, selected bool) bool { return true }
func (p *TableRowTrackingProvider) SelectRange(startID, endID string) bool            { return true }
func (p *TableRowTrackingProvider) SelectAll() bool                                   { return true }
func (p *TableRowTrackingProvider) ClearSelection()                                   {}
func (p *TableRowTrackingProvider) GetSelectedIndices() []int                         { return []int{} }
func (p *TableRowTrackingProvider) GetSelectedIDs() []string                          { return []string{} }
func (p *TableRowTrackingProvider) GetItemID(item *TableRow) string {
	return fmt.Sprintf("row-%s", item.Cells[0])
}

// Helper function to generate test table rows
func generateTestTableRows(count int) []TableRow {
	rows := make([]TableRow, count)
	for i := 0; i < count; i++ {
		rows[i] = TableRow{
			Cells: []string{fmt.Sprintf("Item %d", i), fmt.Sprintf("%d", i)},
		}
	}
	return rows
}
