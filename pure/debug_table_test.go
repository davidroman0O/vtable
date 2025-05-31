package vtable

import (
	"fmt"
	"strings"
	"testing"
)

func TestFullRowHighlighting_Simple_Working(t *testing.T) {
	fmt.Println("\n=== SIMPLE WORKING FULL ROW HIGHLIGHTING TEST ===")

	rows := []TableRow{
		{ID: "row-1", Cells: []string{"Item 1", "25", "Status0"}},
		{ID: "row-2", Cells: []string{"Item 2", "75", "Status1"}},
		{ID: "row-3", Cells: []string{"Item 3", "50", "Status2"}},
	}
	table := createTestTable(rows)

	// Test 1: Verify feature can be enabled
	cmd := FullRowHighlightEnableCmd(true)
	msg := cmd()
	table.Update(msg)

	if !table.config.FullRowHighlighting {
		t.Errorf("❌ Failed to enable full row highlighting")
		return
	}
	fmt.Printf("✅ Full row highlighting enabled: %v\n", table.config.FullRowHighlighting)

	// Test 2: Verify feature can be disabled
	cmd2 := FullRowHighlightEnableCmd(false)
	msg2 := cmd2()
	table.Update(msg2)

	if table.config.FullRowHighlighting {
		t.Errorf("❌ Failed to disable full row highlighting")
		return
	}
	fmt.Printf("✅ Full row highlighting disabled: %v\n", table.config.FullRowHighlighting)

	// Test 3: Enable and render table
	table.Update(FullRowHighlightEnableCmd(true)())
	table.viewport.CursorIndex = 1
	table.viewport.CursorViewportIndex = 1

	output := table.View()
	fmt.Printf("Table output with full row highlighting:\n%s\n", output)

	// Test 4: Basic sanity check - output should contain data
	if !strings.Contains(output, "Item 1") || !strings.Contains(output, "Item 2") {
		t.Errorf("❌ Table output doesn't contain expected data")
		return
	}

	fmt.Printf("✅ Table renders with data\n")
	fmt.Printf("✅ Full row highlighting feature test PASSED\n")
}
