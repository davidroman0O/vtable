// This file contains tests for the table's border rendering functionality.
// It specifically verifies the granular controls for showing and hiding
// individual borders (top, bottom, header separator) and the feature for
// removing the empty lines that can appear above or below the table's borders.
package table

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman0O/vtable/config"
	"github.com/davidroman0O/vtable/core"
)

// TestGranularBorderControls tests the new granular border control functionality
func TestGranularBorderControls(t *testing.T) {
	// Create a test table
	columns := []core.TableColumn{
		{Title: "Name", Field: "name", Width: 10},
		{Title: "Value", Field: "value", Width: 8},
	}

	rows := []core.TableRow{
		{ID: "row-1", Cells: []string{"Item 1", "Value 1"}},
		{ID: "row-2", Cells: []string{"Item 2", "Value 2"}},
	}

	dataSource := &BorderTestDataSource{items: rows}

	tableConfig := core.TableConfig{
		Columns:             columns,
		ShowHeader:          true,
		ShowBorders:         true,
		ShowTopBorder:       true,
		ShowBottomBorder:    true,
		ShowHeaderSeparator: true,
		ViewportConfig:      config.DefaultViewportConfig(),
		Theme:               config.DefaultTheme(),
		SelectionMode:       core.SelectionNone,
		KeyMap:              core.DefaultNavigationKeyMap(),
	}

	table := NewTable(tableConfig, dataSource)

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
	table.Update(core.TopBorderVisibilityMsg{Visible: false})
	output = table.View()
	fmt.Printf("=== TEST 2: No top border ===\n%s\n\n", output)

	if strings.Contains(output, "┌") {
		t.Error("Should not have top border when disabled")
	}
	if !strings.Contains(output, "└") {
		t.Error("Should still have bottom border when only top is disabled")
	}

	// Test 3: Disable bottom border only (enable top back first)
	table.Update(core.TopBorderVisibilityMsg{Visible: true})
	table.Update(core.BottomBorderVisibilityMsg{Visible: false})
	output = table.View()
	fmt.Printf("=== TEST 3: No bottom border ===\n%s\n\n", output)

	if !strings.Contains(output, "┌") {
		t.Error("Should have top border when only bottom is disabled")
	}
	if strings.Contains(output, "└") {
		t.Error("Should not have bottom border when disabled")
	}

	// Test 4: Disable header separator only (enable bottom back first)
	table.Update(core.BottomBorderVisibilityMsg{Visible: true})
	table.Update(core.HeaderSeparatorVisibilityMsg{Visible: false})
	output = table.View()
	fmt.Printf("=== TEST 4: No header separator ===\n%s\n\n", output)

	if !strings.Contains(output, "┌") || !strings.Contains(output, "└") {
		t.Error("Should have top and bottom borders when only header separator is disabled")
	}
	if strings.Contains(output, "├") || strings.Contains(output, "┤") {
		t.Error("Should not have header separator when disabled")
	}

	// Test 5: Disable all borders
	table.Update(core.TopBorderVisibilityMsg{Visible: false})
	table.Update(core.BottomBorderVisibilityMsg{Visible: false})
	table.Update(core.HeaderSeparatorVisibilityMsg{Visible: false})
	output = table.View()
	fmt.Printf("=== TEST 5: No borders at all ===\n%s\n\n", output)

	if strings.Contains(output, "┌") || strings.Contains(output, "└") ||
		strings.Contains(output, "├") || strings.Contains(output, "┤") {
		t.Error("Should have no borders when all are disabled")
	}

	// Test 6: Test the commands work
	cmd := core.TopBorderVisibilityCmd(true)
	msg := cmd()
	if _, ok := msg.(core.TopBorderVisibilityMsg); !ok {
		t.Error("TopBorderVisibilityCmd should return TopBorderVisibilityMsg")
	}

	cmd = core.BottomBorderVisibilityCmd(true)
	msg = cmd()
	if _, ok := msg.(core.BottomBorderVisibilityMsg); !ok {
		t.Error("BottomBorderVisibilityCmd should return BottomBorderVisibilityMsg")
	}

	cmd = core.HeaderSeparatorVisibilityCmd(true)
	msg = cmd()
	if _, ok := msg.(core.HeaderSeparatorVisibilityMsg); !ok {
		t.Error("HeaderSeparatorVisibilityCmd should return HeaderSeparatorVisibilityMsg")
	}

	fmt.Println("✅ All granular border control tests passed!")
}

// TestBorderSpaceRemoval tests the new border space removal functionality
func TestBorderSpaceRemoval(t *testing.T) {
	// Create a test table
	columns := []core.TableColumn{
		{Title: "Name", Field: "name", Width: 10},
		{Title: "Value", Field: "value", Width: 8},
	}

	rows := []core.TableRow{
		{ID: "row-1", Cells: []string{"Item 1", "Value 1"}},
		{ID: "row-2", Cells: []string{"Item 2", "Value 2"}},
	}

	dataSource := &BorderTestDataSource{items: rows}

	tableConfig := core.TableConfig{
		Columns:                 columns,
		ShowHeader:              true,
		ShowBorders:             true,
		ShowTopBorder:           true,
		ShowBottomBorder:        true,
		ShowHeaderSeparator:     true,
		RemoveTopBorderSpace:    false,
		RemoveBottomBorderSpace: false,
		ViewportConfig:          config.DefaultViewportConfig(),
		Theme:                   config.DefaultTheme(),
		SelectionMode:           core.SelectionNone,
		KeyMap:                  core.DefaultNavigationKeyMap(),
	}

	table := NewTable(tableConfig, dataSource)

	// Test 1: Default state - borders visible, space preserved
	output := table.View()
	fmt.Printf("=== TEST 1: Default state (borders visible, space preserved) ===\n%s\n\n", output)

	lines := strings.Split(output, "\n")
	if len(lines) < 6 {
		t.Error("Expected at least 6 lines with full borders")
	}

	// Test 2: Remove top border space completely
	table.Update(core.TopBorderSpaceRemovalMsg{Remove: true})
	output = table.View()
	fmt.Printf("=== TEST 2: Top border space removed ===\n%s\n\n", output)

	// Should have fewer lines since top border space is removed
	newLines := strings.Split(output, "\n")
	if len(newLines) >= len(lines) {
		t.Error("Expected fewer lines when top border space is removed")
	}

	// Test 3: Remove bottom border space too
	table.Update(core.BottomBorderSpaceRemovalMsg{Remove: true})
	output = table.View()
	fmt.Printf("=== TEST 3: Both top and bottom border space removed ===\n%s\n\n", output)

	// Should have even fewer lines
	finalLines := strings.Split(output, "\n")
	if len(finalLines) >= len(newLines) {
		t.Error("Expected even fewer lines when both border spaces are removed")
	}

	// Test 4: Restore space preservation
	table.Update(core.TopBorderSpaceRemovalMsg{Remove: false})
	table.Update(core.BottomBorderSpaceRemovalMsg{Remove: false})
	output = table.View()
	fmt.Printf("=== TEST 4: Border space restored ===\n%s\n\n", output)

	restoredLines := strings.Split(output, "\n")
	if len(restoredLines) != len(lines) {
		t.Error("Expected same number of lines when border space is restored")
	}

	// Test 5: Test commands work
	cmd := core.TopBorderSpaceRemovalCmd(true)
	msg := cmd()
	if _, ok := msg.(core.TopBorderSpaceRemovalMsg); !ok {
		t.Error("TopBorderSpaceRemovalCmd should return TopBorderSpaceRemovalMsg")
	}

	cmd = core.BottomBorderSpaceRemovalCmd(true)
	msg = cmd()
	if _, ok := msg.(core.BottomBorderSpaceRemovalMsg); !ok {
		t.Error("BottomBorderSpaceRemovalCmd should return BottomBorderSpaceRemovalMsg")
	}

	fmt.Println("✅ All border space removal tests passed!")
}

// TestDataSource for testing
type BorderTestDataSource struct {
	items []core.TableRow
}

func (ds *BorderTestDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return core.DataTotalMsg{Total: len(ds.items)}
	}
}

func (ds *BorderTestDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

func (ds *BorderTestDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		var items []core.Data[any]
		for i, row := range ds.items {
			if i >= request.Start && i < request.Start+request.Count {
				items = append(items, core.Data[any]{
					ID:       row.ID,
					Item:     row,
					Selected: false,
					Metadata: core.NewTypedMetadata(),
				})
			}
		}
		return core.DataChunkLoadedMsg{
			StartIndex: request.Start,
			Items:      items,
			Request:    request,
		}
	}
}

func (ds *BorderTestDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		return core.SelectionResponseMsg{Success: true, Index: index, Selected: selected}
	}
}

func (ds *BorderTestDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	return func() tea.Msg {
		return core.SelectionResponseMsg{Success: true, ID: id, Selected: selected}
	}
}

func (ds *BorderTestDataSource) ClearSelection() tea.Cmd {
	return func() tea.Msg {
		return core.SelectionResponseMsg{Success: true, Operation: "clear"}
	}
}

func (ds *BorderTestDataSource) SelectAll() tea.Cmd {
	return func() tea.Msg {
		return core.SelectionResponseMsg{Success: true, Operation: "selectAll"}
	}
}

func (ds *BorderTestDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	return func() tea.Msg {
		return core.SelectionResponseMsg{Success: true, Operation: "range"}
	}
}

func (ds *BorderTestDataSource) GetItemID(item any) string {
	if row, ok := item.(core.TableRow); ok {
		return row.ID
	}
	return ""
}
