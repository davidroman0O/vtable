package vtable

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// HorizontalScrollTestDataSource implements DataSource for testing horizontal scrolling
type HorizontalScrollTestDataSource struct {
	data []TableRow
}

func (ds *HorizontalScrollTestDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return DataTotalMsg{Total: len(ds.data)}
	}
}

func (ds *HorizontalScrollTestDataSource) LoadChunk(request DataRequest) tea.Cmd {
	return func() tea.Msg {
		start := request.Start
		end := start + request.Count
		if end > len(ds.data) {
			end = len(ds.data)
		}

		var items []Data[any]
		for i := start; i < end; i++ {
			items = append(items, Data[any]{
				ID:       ds.data[i].ID,
				Item:     ds.data[i],
				Selected: false,
				Metadata: NewTypedMetadata(),
			})
		}

		return DataChunkLoadedMsg{
			StartIndex: start,
			Items:      items,
			Request:    request,
		}
	}
}

func (ds *HorizontalScrollTestDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		return SelectionResponseMsg{Success: true, Index: index, Selected: selected}
	}
}

func (ds *HorizontalScrollTestDataSource) SetSelectedByID(id string, selected bool) tea.Cmd {
	return ds.SetSelected(0, selected)
}

func (ds *HorizontalScrollTestDataSource) ClearSelection() tea.Cmd {
	return func() tea.Msg {
		return SelectionResponseMsg{Success: true, Operation: "clear"}
	}
}

func (ds *HorizontalScrollTestDataSource) SelectAll() tea.Cmd {
	return func() tea.Msg {
		return SelectionResponseMsg{Success: true, Operation: "selectAll"}
	}
}

func (ds *HorizontalScrollTestDataSource) SelectRange(startIndex, endIndex int) tea.Cmd {
	return ds.SelectAll()
}

func (ds *HorizontalScrollTestDataSource) GetItemID(item any) string {
	if row, ok := item.(TableRow); ok {
		return row.ID
	}
	return ""
}

func (ds *HorizontalScrollTestDataSource) RefreshTotal() tea.Cmd {
	return ds.GetTotal()
}

func createHorizontalScrollTestTable() *Table {
	longText := "This is a very long piece of text that should definitely be longer than any reasonable column width to test horizontal scrolling functionality."

	dataSource := &HorizontalScrollTestDataSource{
		data: []TableRow{
			{
				ID:    "test-1",
				Cells: []string{longText},
			},
		},
	}

	columns := []TableColumn{
		{
			Title:     "Long Text",
			Field:     "text",
			Width:     25, // Narrow width to force scrolling
			Alignment: AlignLeft,
		},
	}

	config := TableConfig{
		Columns:     columns,
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: ViewportConfig{
			Height:    3,
			ChunkSize: 10,
		},
		Theme:         DefaultTheme(),
		SelectionMode: SelectionNone,
	}

	table := NewTable(config, dataSource)
	table.Focus() // CRITICAL: Focus the table
	return table
}

func initializeTestTable(table *Table) {
	// Initialize table
	initCmd := table.Init()
	if initCmd != nil {
		msg := initCmd()
		if msg != nil {
			table.Update(msg)
		}
	}

	// Load total
	totalCmd := table.dataSource.GetTotal()
	if totalCmd != nil {
		msg := totalCmd()
		if msg != nil {
			table.Update(msg)
		}
	}

	// Load data
	chunkCmd := table.dataSource.LoadChunk(DataRequest{Start: 0, Count: 1})
	if chunkCmd != nil {
		msg := chunkCmd()
		if msg != nil {
			table.Update(msg)
		}
	}
}

func TestHorizontalScrollCharacterMode(t *testing.T) {
	table := createHorizontalScrollTestTable()
	initializeTestTable(table)

	// Set character mode
	table.horizontalScrollMode = "character"

	// Get initial view
	initialView := table.View()
	initialContent := extractCellContentFromView(initialView)

	if !strings.Contains(initialContent, "This is a very") {
		t.Errorf("Initial view should contain start of text, got: %s", initialContent)
	}

	// Test scrolling right character by character
	for i := 0; i < 5; i++ {
		table.horizontalScrollOffsets[0]++
		view := table.View()
		content := extractCellContentFromView(view)

		// Content should change with each scroll
		if content == initialContent {
			t.Errorf("Content should change after scrolling %d characters, but got same content: %s", i+1, content)
		}

		// Should no longer start with "This"
		if i >= 2 && strings.HasPrefix(content, "This") {
			t.Errorf("After scrolling %d characters, content should not start with 'This', got: %s", i+1, content)
		}
	}
}

func TestHorizontalScrollWordMode(t *testing.T) {
	table := createHorizontalScrollTestTable()
	initializeTestTable(table)

	// Set word mode
	table.horizontalScrollMode = "word"

	// Get initial view
	initialView := table.View()
	initialContent := extractCellContentFromView(initialView)

	if !strings.Contains(initialContent, "This is a very") {
		t.Errorf("Initial view should contain start of text, got: %s", initialContent)
	}

	// Test scrolling right word by word
	table.horizontalScrollOffsets[0] = 1
	view := table.View()
	content := extractCellContentFromView(view)

	// Strip ANSI codes for comparison since styled text contains escape sequences
	plainContent := stripANSIForTest(content)

	// After scrolling one word, should start with "is" and not start with "This"
	if strings.HasPrefix(plainContent, "This") {
		t.Errorf("After scrolling 1 word, content should not start with 'This', got: %s", plainContent)
	}
	if !strings.HasPrefix(plainContent, "is") {
		t.Errorf("After scrolling 1 word, content should start with 'is', got: %s", plainContent)
	}

	// Scroll another word
	table.horizontalScrollOffsets[0] = 2
	view = table.View()
	content = extractCellContentFromView(view)

	// Strip ANSI codes for comparison
	plainContent = stripANSIForTest(content)

	// After scrolling two words, should start with "a" and not start with "This" or "is"
	if strings.HasPrefix(plainContent, "This") || strings.HasPrefix(plainContent, "is") {
		t.Errorf("After scrolling 2 words, content should not start with 'This' or 'is', got: %s", plainContent)
	}
	if !strings.HasPrefix(plainContent, "a") {
		t.Errorf("After scrolling 2 words, content should start with 'a', got: %s", plainContent)
	}
}

// stripANSIForTest removes ANSI escape codes from a string for testing
func stripANSIForTest(s string) string {
	result := ""
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result += string(r)
	}
	return result
}

func TestHorizontalScrollSmartMode(t *testing.T) {
	table := createHorizontalScrollTestTable()
	initializeTestTable(table)

	// Set smart mode
	table.horizontalScrollMode = "smart"

	// Get initial view
	initialView := table.View()
	initialContent := extractCellContentFromView(initialView)

	if !strings.Contains(initialContent, "This is a very") {
		t.Errorf("Initial view should contain start of text, got: %s", initialContent)
	}

	// Test scrolling right by smart boundaries
	table.horizontalScrollOffsets[0] = 1
	view := table.View()
	content := extractCellContentFromView(view)

	// Content should change
	if content == initialContent {
		t.Errorf("Content should change after smart scrolling, but got same content: %s", content)
	}
}

func TestHorizontalScrollScopeCurrentOnly(t *testing.T) {
	table := createHorizontalScrollTestTable()
	initializeTestTable(table)

	// Set scope to current row only
	table.scrollAllRows = false
	table.horizontalScrollMode = "character"

	// Apply scrolling - should only affect current row
	table.horizontalScrollOffsets[0] = 5

	view := table.View()
	content := extractCellContentFromView(view)

	// Content should be scrolled
	if strings.HasPrefix(content, "This is") {
		t.Errorf("Content should be scrolled when scope is 'current', got: %s", content)
	}
}

func TestHorizontalScrollReset(t *testing.T) {
	table := createHorizontalScrollTestTable()
	initializeTestTable(table)

	// Apply some scrolling
	table.horizontalScrollOffsets[0] = 5

	// Verify scrolling is applied
	view := table.View()
	scrolledContent := extractCellContentFromView(view)

	// Reset scrolling
	table.horizontalScrollOffsets = make(map[int]int)

	// Verify reset
	view = table.View()
	resetContent := extractCellContentFromView(view)

	if !strings.Contains(resetContent, "This is a very") {
		t.Errorf("After reset, content should contain start of text, got: %s", resetContent)
	}

	if resetContent == scrolledContent {
		t.Errorf("Content should be different after reset")
	}
}

func TestHorizontalScrollMaxBounds(t *testing.T) {
	table := createHorizontalScrollTestTable()
	initializeTestTable(table)

	// Set character mode
	table.horizontalScrollMode = "character"

	// Try to scroll way past the end
	table.horizontalScrollOffsets[0] = 1000

	view := table.View()
	content := extractCellContentFromView(view)

	// Debug output to see what we're actually getting
	t.Logf("View when scrolled to offset 1000:\n%s", view)
	t.Logf("Extracted content: %q", content)

	// Should not crash - empty content is valid when scrolled past end
	// The important thing is that we don't get an error/crash
	if view == "" {
		t.Errorf("View should not be completely empty")
	}

	// Content can be empty when scrolled past the end - that's expected behavior
	t.Logf("Successfully scrolled past end without errors. Content: %q", content)
}

func TestToggleScrollMode(t *testing.T) {
	table := createHorizontalScrollTestTable()
	initializeTestTable(table)

	// Start with character mode
	if table.horizontalScrollMode != "character" {
		t.Errorf("Expected initial mode to be 'character', got: %s", table.horizontalScrollMode)
	}

	// Toggle to word mode
	table.handleToggleScrollMode()
	if table.horizontalScrollMode != "word" {
		t.Errorf("Expected mode to be 'word' after first toggle, got: %s", table.horizontalScrollMode)
	}

	// Toggle to smart mode
	table.handleToggleScrollMode()
	if table.horizontalScrollMode != "smart" {
		t.Errorf("Expected mode to be 'smart' after second toggle, got: %s", table.horizontalScrollMode)
	}

	// Toggle back to character mode
	table.handleToggleScrollMode()
	if table.horizontalScrollMode != "character" {
		t.Errorf("Expected mode to be 'character' after third toggle, got: %s", table.horizontalScrollMode)
	}
}

func TestToggleScrollScope(t *testing.T) {
	table := createHorizontalScrollTestTable()

	// Start with current row scope (default is false = current row only)
	if table.scrollAllRows {
		t.Errorf("Expected initial scroll all rows to be false, got %v", table.scrollAllRows)
	}

	// Toggle to all rows scope
	table.handleToggleScrollScope()
	if !table.scrollAllRows {
		t.Errorf("Expected scroll all rows to be true after first toggle, got %v", table.scrollAllRows)
	}

	// Toggle back to current row scope
	table.handleToggleScrollScope()
	if table.scrollAllRows {
		t.Errorf("Expected scroll all rows to be false after second toggle, got %v", table.scrollAllRows)
	}
}

func TestKeyboardHorizontalScrolling(t *testing.T) {
	table := createHorizontalScrollTestTable()
	initializeTestTable(table)

	// Test left/right arrow keys
	initialOffset := table.horizontalScrollOffsets[0]

	// Scroll right
	rightMsg := tea.KeyMsg{Type: tea.KeyRight}
	table.Update(rightMsg)

	rightOffset := table.horizontalScrollOffsets[0]
	if rightOffset != initialOffset+1 {
		t.Errorf("Expected offset to increase by 1 after right arrow, got: %d", rightOffset)
	}

	// Scroll left
	leftMsg := tea.KeyMsg{Type: tea.KeyLeft}
	table.Update(leftMsg)

	leftOffset := table.horizontalScrollOffsets[0]
	if leftOffset != initialOffset {
		t.Errorf("Expected offset to return to initial value after left arrow, got: %d", leftOffset)
	}

	// Test backspace reset
	table.horizontalScrollOffsets[0] = 5
	backspaceMsg := tea.KeyMsg{Type: tea.KeyBackspace}
	table.Update(backspaceMsg)

	resetOffset := table.horizontalScrollOffsets[0]
	if resetOffset != 0 {
		t.Errorf("Expected offset to be 0 after backspace reset, got: %d", resetOffset)
	}
}

func extractCellContentFromView(tableView string) string {
	lines := strings.Split(tableView, "\n")

	// Look for the data row (contains our test text)
	for _, line := range lines {
		if strings.Contains(line, "This") || strings.Contains(line, "very") || strings.Contains(line, "text") {
			// Remove table border characters and trim
			content := strings.Trim(line, "│ ")
			content = strings.TrimSpace(content)
			return content
		}
	}

	// If we can't find the main content, return the line that looks like data
	for _, line := range lines {
		if strings.Contains(line, "│") && !strings.Contains(line, "Long Text") && !strings.Contains(line, "─") {
			content := strings.Trim(line, "│ ")
			content = strings.TrimSpace(content)
			// Return empty content instead of ERROR - this is valid when scrolled past end
			return content
		}
	}

	return ""
}

func BenchmarkHorizontalScrolling(b *testing.B) {
	table := createHorizontalScrollTestTable()
	initializeTestTable(table)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		table.horizontalScrollOffsets[0] = i % 50
		_ = table.View()
	}
}
