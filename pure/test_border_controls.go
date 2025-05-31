package vtable

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestGranularBorderControls tests the new granular border control functionality
func TestGranularBorderControls(t *testing.T) {
	// Create a test table
	columns := []TableColumn{
		{Title: "Name", Field: "name", Width: 10},
		{Title: "Value", Field: "value", Width: 8},
	}

	rows := []TableRow{
		{ID: "row-1", Cells: []string{"Item 1", "Value 1"}},
		{ID: "row-2", Cells: []string{"Item 2", "Value 2"}},
	}

	dataSource := &BorderTestDataSource{items: rows}

	config := TableConfig{
		Columns:             columns,
		ShowHeader:          true,
		ShowBorders:         true,
		ShowTopBorder:       true,
		ShowBottomBorder:    true,
		ShowHeaderSeparator: true,
		ViewportConfig:      DefaultViewportConfig(),
		Theme:               DefaultTheme(),
		SelectionMode:       SelectionNone,
		KeyMap:              DefaultNavigationKeyMap(),
	}

	table := NewTable(config, dataSource)

	// Test 1: Default state should have all borders
	output := table.View()
	fmt.Printf("=== TEST 1: Default state (all borders) ===\n%s\n\n", output)

	if !strings.Contains(output, "┌") || !strings.Contains(output, "└") {
		t.Error("Expected top and bottom borders in default state")
	}
	if !strings.Contains(output, "├") || !strings.Contains(output, "┤") {
		t.Error("Expected header separator in default state")
	}

	// Test 2: Disable top border only
	table.Update(TopBorderVisibilityMsg{Visible: false})
	output = table.View()
	fmt.Printf("=== TEST 2: No top border ===\n%s\n\n", output)

	if strings.Contains(output, "┌") {
		t.Error("Should not have top border when disabled")
	}
	if !strings.Contains(output, "└") {
		t.Error("Should still have bottom border when only top is disabled")
	}

	// Test 3: Disable bottom border only (enable top back first)
	table.Update(TopBorderVisibilityMsg{Visible: true})
	table.Update(BottomBorderVisibilityMsg{Visible: false})
	output = table.View()
	fmt.Printf("=== TEST 3: No bottom border ===\n%s\n\n", output)

	if !strings.Contains(output, "┌") {
		t.Error("Should have top border when only bottom is disabled")
	}
	if strings.Contains(output, "└") {
		t.Error("Should not have bottom border when disabled")
	}

	// Test 4: Disable header separator only (enable bottom back first)
	table.Update(BottomBorderVisibilityMsg{Visible: true})
	table.Update(HeaderSeparatorVisibilityMsg{Visible: false})
	output = table.View()
	fmt.Printf("=== TEST 4: No header separator ===\n%s\n\n", output)

	if !strings.Contains(output, "┌") || !strings.Contains(output, "└") {
		t.Error("Should have top and bottom borders when only header separator is disabled")
	}
	if strings.Contains(output, "├") || strings.Contains(output, "┤") {
		t.Error("Should not have header separator when disabled")
	}

	// Test 5: Disable all borders
	table.Update(TopBorderVisibilityMsg{Visible: false})
	table.Update(BottomBorderVisibilityMsg{Visible: false})
	table.Update(HeaderSeparatorVisibilityMsg{Visible: false})
	output = table.View()
	fmt.Printf("=== TEST 5: No borders at all ===\n%s\n\n", output)

	if strings.Contains(output, "┌") || strings.Contains(output, "└") ||
		strings.Contains(output, "├") || strings.Contains(output, "┤") {
		t.Error("Should have no borders when all are disabled")
	}

	// Test 6: Test the commands work
	cmd := TopBorderVisibilityCmd(true)
	msg := cmd()
	if _, ok := msg.(TopBorderVisibilityMsg); !ok {
		t.Error("TopBorderVisibilityCmd should return TopBorderVisibilityMsg")
	}

	cmd = BottomBorderVisibilityCmd(true)
	msg = cmd()
	if _, ok := msg.(BottomBorderVisibilityMsg); !ok {
		t.Error("BottomBorderVisibilityCmd should return BottomBorderVisibilityMsg")
	}

	cmd = HeaderSeparatorVisibilityCmd(true)
	msg = cmd()
	if _, ok := msg.(HeaderSeparatorVisibilityMsg); !ok {
		t.Error("HeaderSeparatorVisibilityCmd should return HeaderSeparatorVisibilityMsg")
	}

	fmt.Println("✅ All granular border control tests passed!")
}

// TestBorderSpaceRemoval tests the new border space removal functionality
func TestBorderSpaceRemoval(t *testing.T) {
	// Create a test table
	columns := []TableColumn{
		{Title: "Name", Field: "name", Width: 10},
		{Title: "Value", Field: "value", Width: 8},
	}

	rows := []TableRow{
		{ID: "row-1", Cells: []string{"Item 1", "Value 1"}},
		{ID: "row-2", Cells: []string{"Item 2", "Value 2"}},
	}

	dataSource := &BorderTestDataSource{items: rows}

	config := TableConfig{
		Columns:                 columns,
		ShowHeader:              true,
		ShowBorders:             true,
		ShowTopBorder:           true,
		ShowBottomBorder:        true,
		ShowHeaderSeparator:     true,
		RemoveTopBorderSpace:    false,
		RemoveBottomBorderSpace: false,
		ViewportConfig:          DefaultViewportConfig(),
		Theme:                   DefaultTheme(),
		SelectionMode:           SelectionNone,
		KeyMap:                  DefaultNavigationKeyMap(),
	}

	table := NewTable(config, dataSource)

	// Test 1: Default state - borders visible, space preserved
	output := table.View()
	fmt.Printf("=== TEST 1: Default state (borders visible, space preserved) ===\n%s\n\n", output)

	lines := strings.Split(output, "\n")
	if len(lines) < 6 {
		t.Error("Expected at least 6 lines with full borders")
	}

	// Test 2: Remove top border space completely
	table.Update(TopBorderSpaceRemovalMsg{Remove: true})
	output = table.View()
	fmt.Printf("=== TEST 2: Top border space removed ===\n%s\n\n", output)

	// Should have fewer lines since top border space is removed
	newLines := strings.Split(output, "\n")
	if len(newLines) >= len(lines) {
		t.Error("Expected fewer lines when top border space is removed")
	}

	// Test 3: Remove bottom border space too
	table.Update(BottomBorderSpaceRemovalMsg{Remove: true})
	output = table.View()
	fmt.Printf("=== TEST 3: Both top and bottom border space removed ===\n%s\n\n", output)

	// Should have even fewer lines
	finalLines := strings.Split(output, "\n")
	if len(finalLines) >= len(newLines) {
		t.Error("Expected even fewer lines when both border spaces are removed")
	}

	// Test 4: Restore space preservation
	table.Update(TopBorderSpaceRemovalMsg{Remove: false})
	table.Update(BottomBorderSpaceRemovalMsg{Remove: false})
	output = table.View()
	fmt.Printf("=== TEST 4: Border space restored ===\n%s\n\n", output)

	restoredLines := strings.Split(output, "\n")
	if len(restoredLines) != len(lines) {
		t.Error("Expected same number of lines when border space is restored")
	}

	// Test 5: Test commands work
	cmd := TopBorderSpaceRemovalCmd(true)
	msg := cmd()
	if _, ok := msg.(TopBorderSpaceRemovalMsg); !ok {
		t.Error("TopBorderSpaceRemovalCmd should return TopBorderSpaceRemovalMsg")
	}

	cmd = BottomBorderSpaceRemovalCmd(true)
	msg = cmd()
	if _, ok := msg.(BottomBorderSpaceRemovalMsg); !ok {
		t.Error("BottomBorderSpaceRemovalCmd should return BottomBorderSpaceRemovalMsg")
	}

	fmt.Println("✅ All border space removal tests passed!")
}

// TestDataSource for testing
type BorderTestDataSource struct {
	items []TableRow
}

func (ds *BorderTestDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return DataTotalMsg{Total: len(ds.items)}
	}
}

func (ds *BorderTestDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

func (ds *BorderTestDataSource) LoadChunk(request DataRequest) tea.Cmd {
	return func() tea.Msg {
		var items []Data[any]
		for i, row := range ds.items {
			if i >= request.Start && i < request.Start+request.Count {
				items = append(items, Data[any]{
					ID:       row.ID,
					Item:     row,
					Selected: false,
					Metadata: NewTypedMetadata(),
				})
			}
		}
		return DataChunkLoadedMsg{
			StartIndex: request.Start,
			Items:      items,
			Request:    request,
		}
	}
}

func (ds *BorderTestDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		return SelectionResponseMsg{Success: true, Index: index, Selected: selected}
	}
}

func (ds *BorderTestDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	return func() tea.Msg {
		return SelectionResponseMsg{Success: true, ID: id, Selected: selected}
	}
}

func (ds *BorderTestDataSource) ClearSelection() tea.Cmd {
	return func() tea.Msg {
		return SelectionResponseMsg{Success: true, Operation: "clear"}
	}
}

func (ds *BorderTestDataSource) SelectAll() tea.Cmd {
	return func() tea.Msg {
		return SelectionResponseMsg{Success: true, Operation: "selectAll"}
	}
}

func (ds *BorderTestDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	return func() tea.Msg {
		return SelectionResponseMsg{Success: true, Operation: "range"}
	}
}

func (ds *BorderTestDataSource) GetItemID(item any) string {
	if row, ok := item.(TableRow); ok {
		return row.ID
	}
	return ""
}
