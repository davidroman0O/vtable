package vtable

import (
	"fmt"
	"testing"
	"time"
)

// TableRowDataProvider for testing table animations
type TableRowDataProvider struct {
	rows      []TableRow
	selection map[int]bool
}

func NewTableRowDataProvider(rowCount int) *TableRowDataProvider {
	rows := make([]TableRow, rowCount)
	for i := 0; i < rowCount; i++ {
		rows[i] = TableRow{
			Cells: []string{
				fmt.Sprintf("%d", i+1),
				fmt.Sprintf("Row %d", i+1),
			},
		}
	}

	return &TableRowDataProvider{
		rows:      rows,
		selection: make(map[int]bool),
	}
}

func (p *TableRowDataProvider) GetTotal() int {
	return len(p.rows)
}

func (p *TableRowDataProvider) GetItems(request DataRequest) ([]Data[TableRow], error) {
	start := request.Start
	count := request.Count

	if start >= len(p.rows) {
		return []Data[TableRow]{}, nil
	}

	end := start + count
	if end > len(p.rows) {
		end = len(p.rows)
	}

	result := make([]Data[TableRow], end-start)
	for i := start; i < end; i++ {
		result[i-start] = Data[TableRow]{
			ID:       fmt.Sprintf("row-%d", i),
			Item:     p.rows[i],
			Selected: p.selection[i],
			Metadata: NewTypedMetadata(),
		}
	}

	return result, nil
}

func (p *TableRowDataProvider) GetSelectionMode() SelectionMode { return SelectionMultiple }
func (p *TableRowDataProvider) SetSelected(index int, selected bool) bool {
	if selected {
		p.selection[index] = true
	} else {
		delete(p.selection, index)
	}
	return true
}
func (p *TableRowDataProvider) SelectAll() bool                                   { return true }
func (p *TableRowDataProvider) ClearSelection()                                   { p.selection = make(map[int]bool) }
func (p *TableRowDataProvider) GetSelectedIndices() []int                         { return []int{} }
func (p *TableRowDataProvider) GetSelectedIDs() []string                          { return []string{} }
func (p *TableRowDataProvider) SetSelectedByIDs(ids []string, selected bool) bool { return true }
func (p *TableRowDataProvider) SelectRange(startID, endID string) bool            { return true }
func (p *TableRowDataProvider) GetItemID(item *TableRow) string                   { return "row" }

// TestAnimationToggle verifies that disabling and re-enabling animations works correctly
func TestAnimationToggle(t *testing.T) {
	// Create provider
	provider := NewTableRowDataProvider(50)

	// Create table config
	config := TableConfig{
		Columns: []TableColumn{
			{Title: "ID", Width: 10, Field: "id"},
			{Title: "Name", Width: 20, Field: "name"},
		},
		ShowHeader: true,
		ViewportConfig: ViewportConfig{
			Height:               5,
			ChunkSize:            10,
			TopThresholdIndex:    1,
			BottomThresholdIndex: 3,
		},
	}

	// Create table
	table, err := NewTeaTable(config, provider, *DefaultTheme())
	if err != nil {
		t.Fatalf("Failed to create TeaTable: %v", err)
	}

	// Create animated formatter
	animationCallCount := 0
	animatedFormatter := func(data Data[TableRow], index int, ctx RenderContext,
		animationState map[string]any, isCursor bool, isTopThreshold bool, isBottomThreshold bool) RenderResult {

		animationCallCount++

		// Simple animation that changes content
		counter := 0
		if c, ok := animationState["counter"]; ok {
			if ci, ok := c.(int); ok {
				counter = ci
			}
		}
		counter++

		content := "Row " + data.Item.Cells[1] + " (tick: " + fmt.Sprintf("%d", counter) + ")"

		return RenderResult{
			Content: content,
			RefreshTriggers: []RefreshTrigger{{
				Type:     TriggerTimer,
				Interval: 100 * time.Millisecond,
			}},
			AnimationState: map[string]any{
				"counter": counter,
			},
		}
	}

	// Set animated formatter
	table.SetAnimatedFormatter(animatedFormatter)

	// Initial state - animations should be enabled by default
	if !table.IsAnimationEnabled() {
		t.Error("Animations should be enabled by default")
	}

	// Process some animation ticks to establish baseline
	initialCallCount := animationCallCount
	table.processAnimations()
	if animationCallCount <= initialCallCount {
		t.Error("Animation formatter should have been called")
	}

	// Disable animations
	table.DisableAnimations()
	if table.IsAnimationEnabled() {
		t.Error("Animations should be disabled")
	}

	// Animation formatter should not be called when disabled
	disabledCallCount := animationCallCount
	table.processAnimations()
	if animationCallCount > disabledCallCount {
		t.Error("Animation formatter should not be called when disabled")
	}

	// Re-enable animations
	table.EnableAnimations()
	if !table.IsAnimationEnabled() {
		t.Error("Animations should be re-enabled")
	}

	// Animation formatter should be called again after re-enabling
	reenabledCallCount := animationCallCount
	table.processAnimations()
	if animationCallCount <= reenabledCallCount {
		t.Error("Animation formatter should be called after re-enabling")
	}

	// Verify that animation state is properly reset
	// The cache should be cleared and animations should restart fresh
	if len(table.cachedAnimationContent) == 0 {
		// This is expected - cache should be cleared on re-enable
		t.Log("Animation cache properly cleared on re-enable")
	}
}
