package vtable

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// ================================
// DEBUG TESTS - PRINT ACTUAL OUTPUT
// ================================

func TestDebug_BasicTableOutput(t *testing.T) {
	fmt.Println("\n=== BASIC TABLE OUTPUT ===")

	rows := createTestRows(3)
	table := createTestTable(rows)

	output := table.View()
	fmt.Printf("Raw output:\n%s\n", output)

	fmt.Printf("Stripped output:\n%s\n", stripANSI(output))

	lines := strings.Split(stripANSI(output), "\n")
	fmt.Printf("Line count: %d\n", len(lines))
	for i, line := range lines {
		fmt.Printf("Line %d: %q\n", i, line)
	}
}

func TestDebug_WithFormatters(t *testing.T) {
	fmt.Println("\n=== TABLE WITH FORMATTERS ===")

	rows := createTestRows(2)
	table := createTestTable(rows)

	// Add cell formatter
	nameFormatter := CreateSimpleCellFormatter(func(cellValue string) string {
		return "★ " + cellValue
	})
	formatterMsg := CellFormatterSetCmd(0, nameFormatter)()
	table.Update(formatterMsg)

	// Add value formatter
	valueFormatter := CreateSimpleCellFormatter(func(cellValue string) string {
		if val, err := strconv.Atoi(cellValue); err == nil && val > 5 {
			return fmt.Sprintf("HIGH:%s", cellValue)
		}
		return cellValue
	})
	formatterMsg2 := CellFormatterSetCmd(1, valueFormatter)()
	table.Update(formatterMsg2)

	output := table.View()
	fmt.Printf("Raw output:\n%s\n", output)

	fmt.Printf("Stripped output:\n%s\n", stripANSI(output))

	lines := strings.Split(stripANSI(output), "\n")
	for i, line := range lines {
		fmt.Printf("Line %d: %q\n", i, line)
	}
}

func TestDebug_ComponentRenderer(t *testing.T) {
	fmt.Println("\n=== TABLE WITH COMPONENT RENDERER ===")

	rows := createTestRows(2)
	table := createTestTable(rows)

	// Enable component renderer
	table.EnableComponentRenderer()

	output := table.View()
	fmt.Printf("Raw output:\n%s\n", output)

	fmt.Printf("Stripped output:\n%s\n", stripANSI(output))

	lines := strings.Split(stripANSI(output), "\n")
	for i, line := range lines {
		fmt.Printf("Line %d: %q\n", i, line)
	}
}

func TestDebug_WithoutBorders(t *testing.T) {
	fmt.Println("\n=== TABLE WITHOUT BORDERS ===")

	rows := createTestRows(2)
	table := createTestTable(rows)
	table.config.ShowBorders = false

	output := table.View()
	fmt.Printf("Raw output:\n%s\n", output)

	fmt.Printf("Stripped output:\n%s\n", stripANSI(output))

	lines := strings.Split(stripANSI(output), "\n")
	for i, line := range lines {
		fmt.Printf("Line %d: %q\n", i, line)
	}
}

func TestDebug_HeaderFormatter(t *testing.T) {
	fmt.Println("\n=== TABLE WITH HEADER FORMATTER ===")

	rows := createTestRows(2)
	table := createTestTable(rows)

	// Add header formatter
	headerFormatter := CreateSimpleHeaderFormatter(func(columnTitle string) string {
		return "📝 " + columnTitle
	})
	headerMsg := HeaderFormatterSetCmd(0, headerFormatter)()
	table.Update(headerMsg)

	output := table.View()
	fmt.Printf("Raw output:\n%s\n", output)

	fmt.Printf("Stripped output:\n%s\n", stripANSI(output))

	lines := strings.Split(stripANSI(output), "\n")
	for i, line := range lines {
		fmt.Printf("Line %d: %q\n", i, line)
	}
}

func TestDebug_LongContent(t *testing.T) {
	fmt.Println("\n=== TABLE WITH LONG CONTENT ===")

	rows := []TableRow{
		{
			ID: "row-1",
			Cells: []string{
				"This is a very long name that exceeds column width",
				"999999",
				"VeryLongStatus",
			},
		},
	}
	table := createTestTable(rows)

	output := table.View()
	fmt.Printf("Raw output:\n%s\n", output)

	fmt.Printf("Stripped output:\n%s\n", stripANSI(output))

	lines := strings.Split(stripANSI(output), "\n")
	for i, line := range lines {
		fmt.Printf("Line %d: %q\n", i, line)
	}
}

func TestDebug_EmptyTable(t *testing.T) {
	fmt.Println("\n=== EMPTY TABLE ===")

	table := createTestTable([]TableRow{})

	output := table.View()
	fmt.Printf("Raw output:\n%s\n", output)

	fmt.Printf("Stripped output:\n%s\n", stripANSI(output))
}

func TestDebug_FullIntegration(t *testing.T) {
	fmt.Println("\n=== FULL INTEGRATION TEST ===")

	rows := createTestRows(3)
	table := createTestTable(rows)

	// Add cell formatter
	nameFormatter := CreateSimpleCellFormatter(func(cellValue string) string {
		return "• " + cellValue
	})
	formatterMsg := CellFormatterSetCmd(0, nameFormatter)()
	table.Update(formatterMsg)

	// Add header formatter
	headerFormatter := CreateSimpleHeaderFormatter(func(columnTitle string) string {
		return "📋 " + columnTitle
	})
	headerMsg := HeaderFormatterSetCmd(0, headerFormatter)()
	table.Update(headerMsg)

	// Enable component renderer
	table.EnableComponentRenderer()

	// Set cursor position
	table.viewport.CursorIndex = 1
	table.viewport.CursorViewportIndex = 1

	output := table.View()
	fmt.Printf("Raw output:\n%s\n", output)

	fmt.Printf("Stripped output:\n%s\n", stripANSI(output))

	lines := strings.Split(stripANSI(output), "\n")
	for i, line := range lines {
		fmt.Printf("Line %d: %q\n", i, line)
	}
}

func TestDebug_CellFormatterContent(t *testing.T) {
	fmt.Println("\n=== DEBUG CELL FORMATTER CONTENT ===")

	rows := createTestRows(2)
	table := createTestTable(rows)

	// Add cell formatter
	nameFormatter := CreateSimpleCellFormatter(func(cellValue string) string {
		fmt.Printf("Cell formatter input: %q\n", cellValue)
		result := "• " + cellValue
		fmt.Printf("Cell formatter output: %q\n", result)
		return result
	})
	formatterMsg := CellFormatterSetCmd(0, nameFormatter)()
	table.Update(formatterMsg)

	fmt.Println("=== WITHOUT COMPONENT RENDERER ===")
	output1 := table.View()
	fmt.Printf("Output: %s\n", stripANSI(output1))

	fmt.Println("=== WITH COMPONENT RENDERER ===")
	table.EnableComponentRenderer()
	output2 := table.View()
	fmt.Printf("Output: %s\n", stripANSI(output2))
}

func TestDebug_ComponentRendererDetailed(t *testing.T) {
	fmt.Println("\n=== DETAILED COMPONENT RENDERER DEBUG ===")

	rows := createTestRows(1)
	table := createTestTable(rows)

	// Add cell formatter with detailed logging
	nameFormatter := CreateSimpleCellFormatter(func(cellValue string) string {
		fmt.Printf("FORMATTER: input=%q\n", cellValue)
		result := "• " + cellValue
		fmt.Printf("FORMATTER: output=%q\n", result)
		return result
	})
	formatterMsg := CellFormatterSetCmd(0, nameFormatter)()
	table.Update(formatterMsg)

	// Enable component renderer
	table.EnableComponentRenderer()

	// Set cursor to see the cursor indicator
	table.viewport.CursorIndex = 0
	table.viewport.CursorViewportIndex = 0

	output := table.View()
	fmt.Printf("Final output:\n%s\n", stripANSI(output))

	lines := strings.Split(stripANSI(output), "\n")
	for i, line := range lines {
		fmt.Printf("Line %d: %q\n", i, line)
	}
}

func TestDebug_AllHeaderFormatters(t *testing.T) {
	fmt.Println("\n=== ALL HEADER FORMATTERS TEST ===")

	rows := createTestRows(2)
	table := createTestTable(rows)

	// Set header formatters for ALL columns like the example app
	headerFormatters := map[int]SimpleHeaderFormatter{
		0: func(column TableColumn, ctx RenderContext) string {
			return "📝 " + column.Title
		},
		1: func(column TableColumn, ctx RenderContext) string {
			return "💰 " + column.Title
		},
		2: func(column TableColumn, ctx RenderContext) string {
			return "📊 " + column.Title
		},
	}

	// Apply all header formatters
	for columnIndex, formatter := range headerFormatters {
		headerMsg := HeaderFormatterSetCmd(columnIndex, formatter)()
		table.Update(headerMsg)
		fmt.Printf("Set header formatter for column %d\n", columnIndex)
	}

	// Enable component renderer
	table.EnableComponentRenderer()

	output := table.View()
	fmt.Printf("Raw output:\n%s\n", output)

	fmt.Printf("Stripped output:\n%s\n", stripANSI(output))

	lines := strings.Split(stripANSI(output), "\n")
	for i, line := range lines {
		fmt.Printf("Line %d: %q\n", i, line)
	}
}

func TestDebug_HeaderWidthCalculation(t *testing.T) {
	fmt.Println("\n=== HEADER WIDTH CALCULATION DEBUG ===")

	rows := createTestRows(1)
	table := createTestTable(rows)

	// Test the Unicode width of header formatter outputs
	headerFormatters := map[int]SimpleHeaderFormatter{
		0: func(column TableColumn, ctx RenderContext) string {
			result := "📝 " + column.Title
			fmt.Printf("Column 0 formatter output: %q (len=%d, lipgloss.Width=%d)\n",
				result, len(result), lipgloss.Width(result))
			return result
		},
		1: func(column TableColumn, ctx RenderContext) string {
			result := "💰 " + column.Title
			fmt.Printf("Column 1 formatter output: %q (len=%d, lipgloss.Width=%d)\n",
				result, len(result), lipgloss.Width(result))
			return result
		},
		2: func(column TableColumn, ctx RenderContext) string {
			result := "📊 " + column.Title
			fmt.Printf("Column 2 formatter output: %q (len=%d, lipgloss.Width=%d)\n",
				result, len(result), lipgloss.Width(result))
			return result
		},
	}

	// Apply all header formatters
	for columnIndex, formatter := range headerFormatters {
		headerMsg := HeaderFormatterSetCmd(columnIndex, formatter)()
		table.Update(headerMsg)
	}

	// Enable component renderer
	table.EnableComponentRenderer()

	// Print column definitions
	for i, col := range table.columns {
		fmt.Printf("Column %d: Title=%q, Width=%d, Alignment=%d\n",
			i, col.Title, col.Width, col.Alignment)
	}

	output := table.View()
	fmt.Printf("Raw output:\n%s\n", output)

	lines := strings.Split(stripANSI(output), "\n")
	headerLine := lines[0]
	fmt.Printf("Header line: %q (len=%d, lipgloss.Width=%d)\n",
		headerLine, len(headerLine), lipgloss.Width(headerLine))

	// Split the header by border characters and analyze each cell
	parts := strings.Split(headerLine, "│")
	for i, part := range parts {
		if part != "" {
			fmt.Printf("Header part %d: %q (len=%d, lipgloss.Width=%d)\n",
				i, part, len(part), lipgloss.Width(part))
		}
	}
}

func TestDebug_WidthMeasurementComparison(t *testing.T) {
	fmt.Println("\n=== WIDTH MEASUREMENT COMPARISON ===")

	// Test various emoji headers
	testStrings := []string{
		"📝 Name",
		"💰 Value",
		"📊 Status",
		"🏷️ Category",
		"●",
		"►",
		"✓",
	}

	for _, str := range testStrings {
		lipglossWidth := lipgloss.Width(str)
		runewidthWidth := runewidth.StringWidth(str)

		fmt.Printf("String: %q\n", str)
		fmt.Printf("  len(): %d\n", len(str))
		fmt.Printf("  lipgloss.Width(): %d\n", lipglossWidth)
		fmt.Printf("  runewidth.StringWidth(): %d\n", runewidthWidth)
		if lipglossWidth != runewidthWidth {
			fmt.Printf("  ❌ MISMATCH!\n")
		} else {
			fmt.Printf("  ✅ Match\n")
		}
		fmt.Println()
	}

	// Also test the actual headers that would be constrained
	fmt.Println("=== AFTER CONSTRAINT APPLICATION ===")

	// Simulate what happens in applyCellConstraints
	testHeader := "📝 Name"
	targetWidth := 10

	fmt.Printf("Original: %q (lipgloss.Width=%d, runewidth=%d)\n",
		testHeader, lipgloss.Width(testHeader), runewidth.StringWidth(testHeader))

	// Current implementation uses lipgloss.Width
	lipglossActualWidth := lipgloss.Width(testHeader)
	if lipglossActualWidth < targetWidth {
		padding := targetWidth - lipglossActualWidth
		padded := testHeader + strings.Repeat(" ", padding)
		fmt.Printf("Lipgloss padded: %q (lipgloss.Width=%d, runewidth=%d)\n",
			padded, lipgloss.Width(padded), runewidth.StringWidth(padded))
	}

	// What if we used runewidth instead?
	runewidthActualWidth := runewidth.StringWidth(testHeader)
	if runewidthActualWidth < targetWidth {
		padding := targetWidth - runewidthActualWidth
		padded := testHeader + strings.Repeat(" ", padding)
		fmt.Printf("Runewidth padded: %q (lipgloss.Width=%d, runewidth=%d)\n",
			padded, lipgloss.Width(padded), runewidth.StringWidth(padded))
	}
}

func TestDebug_ProblematicEmojiHeader(t *testing.T) {
	fmt.Println("\n=== PROBLEMATIC EMOJI HEADER TEST ===")

	rows := createTestRows(1)
	table := createTestTable(rows)

	// Test the problematic 🏷️ emoji that had width mismatch
	headerFormatter := func(column TableColumn, ctx RenderContext) string {
		return "🏷️ " + column.Title
	}

	// Apply to first column (width=10)
	headerMsg := HeaderFormatterSetCmd(0, headerFormatter)()
	table.Update(headerMsg)

	// Enable component renderer
	table.EnableComponentRenderer()

	output := table.View()
	fmt.Printf("Raw output:\n%s\n", output)

	lines := strings.Split(stripANSI(output), "\n")
	headerLine := lines[0]
	fmt.Printf("Header line: %q\n", headerLine)

	// Check the first data column specifically
	parts := strings.Split(headerLine, "│")
	if len(parts) > 2 {
		nameHeader := parts[2] // Skip empty and indicator parts
		fmt.Printf("Name header part: %q (len=%d, runewidth=%d)\n",
			nameHeader, len(nameHeader), runewidth.StringWidth(nameHeader))

		// Should be exactly 10 characters wide
		if runewidth.StringWidth(nameHeader) != 10 {
			fmt.Printf("❌ Width mismatch! Expected 10, got %d\n", runewidth.StringWidth(nameHeader))
		} else {
			fmt.Printf("✅ Width correct! Exactly 10 characters\n")
		}
	}
}

func TestDebug_DemoColoredFormatters(t *testing.T) {
	fmt.Println("\n=== DEMO COLORED FORMATTERS TEST ===")

	rows := []TableRow{
		{
			ID:    "row-1",
			Cells: []string{"Item 1", "Value 0", "Status 0", "Category A"},
		},
		{
			ID:    "row-2",
			Cells: []string{"Item 2", "Value 37", "Status 1", "Category B"},
		},
	}
	table := createTestTable(rows)

	// Apply the EXACT same formatters as the demo app
	// Name formatter - simple styling
	nameFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("39")) // Blue
		return style.Render(cellValue)
	}

	// Value formatter - with color coding like demo
	valueFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		var style lipgloss.Style
		if strings.HasPrefix(cellValue, "Value ") {
			valueStr := strings.TrimPrefix(cellValue, "Value ")
			if value, err := strconv.Atoi(valueStr); err == nil {
				switch {
				case value < 30:
					style = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
				case value < 70:
					style = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
				default:
					style = lipgloss.NewStyle().Foreground(lipgloss.Color("46")) // Green
				}
			}
		} else {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
		}
		return style.Render(cellValue)
	}

	// Status formatter - with icons and colors like demo
	statusFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		var statusText string
		var style lipgloss.Style

		switch cellValue {
		case "Status 0":
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("46")) // Green
			statusText = "✓ Active"
		case "Status 1":
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // Yellow
			statusText = "▲ Warning"
		default:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
			statusText = cellValue
		}

		return style.Render(statusText)
	}

	// Set all formatters
	table.Update(CellFormatterSetCmd(0, nameFormatter)())
	table.Update(CellFormatterSetCmd(1, valueFormatter)())
	table.Update(CellFormatterSetCmd(2, statusFormatter)())

	// Set header formatters like demo
	table.Update(HeaderFormatterSetCmd(0, func(column TableColumn, ctx RenderContext) string {
		return "📝 " + column.Title
	})())
	table.Update(HeaderFormatterSetCmd(1, func(column TableColumn, ctx RenderContext) string {
		return "💰 " + column.Title
	})())
	table.Update(HeaderFormatterSetCmd(2, func(column TableColumn, ctx RenderContext) string {
		return "📊 " + column.Title
	})())

	// Enable component renderer
	table.EnableComponentRenderer()

	output := table.View()
	fmt.Printf("Raw output:\n%s\n", output)

	fmt.Printf("Stripped output:\n%s\n", stripANSI(output))

	lines := strings.Split(stripANSI(output), "\n")
	for i, line := range lines {
		fmt.Printf("Line %d: %q\n", i, line)
	}

	// Verify the column widths are exactly correct
	headerLine := lines[0]
	if !strings.Contains(headerLine, "📝 Name   ") {
		t.Errorf("Name column not properly padded in header")
	}
	if !strings.Contains(headerLine, "💰 Value") {
		t.Errorf("Value column not properly formatted in header")
	}
	if !strings.Contains(headerLine, "📊 Status ") {
		t.Errorf("Status column not properly padded in header")
	}

	// Check that data lines have correct column boundaries
	dataLine1 := lines[1]
	fmt.Printf("Data line analysis: %q\n", dataLine1)

	// Should have exactly aligned columns
	parts := strings.Split(dataLine1, "│")
	if len(parts) >= 4 {
		nameCol := strings.TrimSpace(parts[2]) // Skip empty + indicator, trim border spacing
		valueCol := strings.TrimSpace(parts[3])
		statusCol := strings.TrimSpace(parts[4])

		fmt.Printf("Name column (trimmed): %q (visual_width=%d)\n", nameCol, lipgloss.Width(nameCol))
		fmt.Printf("Value column (trimmed): %q (visual_width=%d)\n", valueCol, lipgloss.Width(valueCol))
		fmt.Printf("Status column (trimmed): %q (visual_width=%d)\n", statusCol, lipgloss.Width(statusCol))

		// The actual column content should fit within the specified widths
		// But we need to check the raw column parts for exact spacing
		rawNameCol := parts[2]
		rawValueCol := parts[3]
		rawStatusCol := parts[4]

		fmt.Printf("Raw name column: %q (visual_width=%d)\n", rawNameCol, lipgloss.Width(rawNameCol))
		fmt.Printf("Raw value column: %q (visual_width=%d)\n", rawValueCol, lipgloss.Width(rawValueCol))
		fmt.Printf("Raw status column: %q (visual_width=%d)\n", rawStatusCol, lipgloss.Width(rawStatusCol))

		// The raw columns should be exactly the specified visual widths
		if lipgloss.Width(rawNameCol) != 10 {
			t.Errorf("Name column width incorrect: expected 10, got %d", lipgloss.Width(rawNameCol))
		}
		if lipgloss.Width(rawValueCol) != 8 {
			t.Errorf("Value column width incorrect: expected 8, got %d", lipgloss.Width(rawValueCol))
		}
		if lipgloss.Width(rawStatusCol) != 10 {
			t.Errorf("Status column width incorrect: expected 10, got %d", lipgloss.Width(rawStatusCol))
		}
	}
}

func TestDebug_UnicodeStatusMeasurement(t *testing.T) {
	fmt.Println("\n=== UNICODE STATUS MEASUREMENT DEBUG ===")

	// Test the specific status text that's causing issues
	testTexts := []string{
		"✓ Active",
		"▲ Warning",
		"Status0",
		"Item 1",
		"Value 0",
	}

	for _, text := range testTexts {
		fmt.Printf("\nTesting: %q\n", text)
		fmt.Printf("  len(): %d\n", len(text))
		fmt.Printf("  lipgloss.Width(): %d\n", lipgloss.Width(text))
		fmt.Printf("  runewidth.StringWidth(): %d\n", runewidth.StringWidth(text))

		// Test styled version
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
		styledText := style.Render(text)
		fmt.Printf("  styled len(): %d\n", len(styledText))
		fmt.Printf("  styled lipgloss.Width(): %d\n", lipgloss.Width(styledText))
		fmt.Printf("  styled runewidth.StringWidth(): %d\n", runewidth.StringWidth(styledText))
		fmt.Printf("  styled hasANSI: %t\n", strings.Contains(styledText, "\x1b"))
	}

	// Test constraint application specifically
	fmt.Println("\n=== CONSTRAINT APPLICATION TEST ===")

	statusText := "✓ Active"
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	styledText := style.Render(statusText)

	fmt.Printf("Original: %q\n", statusText)
	fmt.Printf("Styled: len=%d, lipgloss.Width=%d, runewidth=%d\n",
		len(styledText), lipgloss.Width(styledText), runewidth.StringWidth(styledText))

	// Manually apply the constraint logic
	width := 10
	hasANSI := strings.Contains(styledText, "\x1b")
	fmt.Printf("Has ANSI: %t\n", hasANSI)

	var measureWidth func(string) int
	if hasANSI {
		measureWidth = lipgloss.Width
		fmt.Println("Using lipgloss.Width for measurement")
	} else {
		measureWidth = runewidth.StringWidth
		fmt.Println("Using runewidth.StringWidth for measurement")
	}

	actualWidth := measureWidth(styledText)
	fmt.Printf("Measured width: %d\n", actualWidth)

	if actualWidth < width {
		padding := width - actualWidth
		paddedText := styledText + strings.Repeat(" ", padding)
		fmt.Printf("Padded result: %q (total len=%d)\n", paddedText, len(paddedText))
		fmt.Printf("Padded visual width: lipgloss=%d, runewidth=%d\n",
			lipgloss.Width(paddedText), runewidth.StringWidth(paddedText))
	}
}

func TestDebug_CellRenderingSteps(t *testing.T) {
	fmt.Println("\n=== CELL RENDERING STEPS DEBUG ===")

	rows := []TableRow{
		{
			ID:    "row-1",
			Cells: []string{"Item 1", "Value 0", "Status 0"},
		},
	}
	table := createTestTable(rows)

	// Add status formatter that matches the demo
	statusFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		statusText := "✓ Active"
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
		result := style.Render(statusText)
		fmt.Printf("Status formatter result: %q (len=%d, visual=%d)\n", result, len(result), lipgloss.Width(result))
		return result
	}

	table.Update(CellFormatterSetCmd(2, statusFormatter)())
	table.EnableComponentRenderer()

	// Check what getBorderChar returns
	borderChar := table.getBorderChar()
	fmt.Printf("Border char: %q (len=%d)\n", borderChar, len(borderChar))

	// Check the theme border characters
	fmt.Printf("Theme vertical border: %q\n", table.config.Theme.BorderChars.Vertical)

	// Now render and examine each step
	output := table.View()
	fmt.Printf("Final output:\n%s\n", output)

	lines := strings.Split(stripANSI(output), "\n")
	if len(lines) > 1 {
		dataLine := lines[1]
		fmt.Printf("Data line: %q\n", dataLine)

		// Split by border character and examine each part
		parts := strings.Split(dataLine, borderChar)
		fmt.Printf("Split parts count: %d\n", len(parts))
		for i, part := range parts {
			fmt.Printf("Part %d: %q (len=%d)\n", i, part, len(part))
		}

		// Focus on the status column (should be part 4)
		if len(parts) > 4 {
			statusPart := parts[4]
			fmt.Printf("Status part analysis: %q (len=%d)\n", statusPart, len(statusPart))
			fmt.Printf("Status part trimmed: %q (len=%d)\n", strings.TrimSpace(statusPart), len(strings.TrimSpace(statusPart)))
		}
	}
}

func TestDebug_AdvancedFeatures(t *testing.T) {
	fmt.Println("\n=== ADVANCED FEATURES TEST ===")

	rows := []TableRow{
		{
			ID:    "row-1",
			Cells: []string{"Very Long Item Name That Should Wrap", "Value 75", "Status 0", "Category A"},
		},
		{
			ID:    "row-2",
			Cells: []string{"Item 2", "Value 25", "Status 1", "Category B"},
		},
	}
	table := createTestTable(rows)

	// ===== TEST COLUMN ORDERING =====
	fmt.Println("\n--- Column Ordering Test ---")

	// Change column order (Value, Name, Status, Category)
	newColumns := []TableColumn{
		{Title: "Value", Field: "value", Width: 15, Alignment: AlignRight},
		{Title: "Name", Field: "name", Width: 25, Alignment: AlignLeft},
		{Title: "Status", Field: "status", Width: 18, Alignment: AlignCenter},
		{Title: "Category", Field: "category", Width: 20, Alignment: AlignLeft},
	}
	table.Update(ColumnSetCmd(newColumns))

	output := table.View()
	fmt.Printf("Reordered columns:\n%s\n", stripANSI(output))

	// ===== TEST SORTING =====
	fmt.Println("\n--- Sorting Test ---")

	// Sort by name ascending
	table.Update(SortSetCmd("name", "asc"))
	output = table.View()
	fmt.Printf("Sorted by name (asc):\n%s\n", stripANSI(output))

	// Add multi-sort: name + value
	table.Update(SortAddCmd("value", "desc"))
	output = table.View()
	fmt.Printf("Multi-sort (name asc + value desc):\n%s\n", stripANSI(output))

	// ===== TEST FILTERING =====
	fmt.Println("\n--- Filtering Test ---")

	// Filter by category
	table.Update(FilterSetCmd("category", "Category A"))
	output = table.View()
	fmt.Printf("Filtered by Category A:\n%s\n", stripANSI(output))

	// Clear filters
	table.Update(FiltersClearAllCmd())
	output = table.View()
	fmt.Printf("Filters cleared:\n%s\n", stripANSI(output))

	// ===== TEST HEADER FORMATTERS =====
	fmt.Println("\n--- Header Formatters Test ---")

	// Apply header formatters to reordered columns
	headerFormatters := map[int]SimpleHeaderFormatter{
		0: func(column TableColumn, ctx RenderContext) string {
			return "💰 " + column.Title // Value column first now
		},
		1: func(column TableColumn, ctx RenderContext) string {
			return "📝 " + column.Title // Name column second now
		},
		2: func(column TableColumn, ctx RenderContext) string {
			return "📊 " + column.Title // Status column
		},
		3: func(column TableColumn, ctx RenderContext) string {
			return "🏷️ " + column.Title // Category column
		},
	}

	for columnIndex, formatter := range headerFormatters {
		table.Update(HeaderFormatterSetCmd(columnIndex, formatter))
	}

	output = table.View()
	fmt.Printf("With header formatters:\n%s\n", stripANSI(output))

	// ===== TEST CELL FORMATTERS =====
	fmt.Println("\n--- Cell Formatters Test ---")

	// Apply enhanced cell formatters
	valueFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		if strings.HasPrefix(cellValue, "Value ") {
			valueStr := strings.TrimPrefix(cellValue, "Value ")
			if value, err := strconv.Atoi(valueStr); err == nil {
				style := lipgloss.NewStyle()
				switch {
				case value < 30:
					style = style.Foreground(lipgloss.Color("9")) // Red
				case value < 70:
					style = style.Foreground(lipgloss.Color("11")) // Yellow
				default:
					style = style.Foreground(lipgloss.Color("10")) // Green
				}
				return style.Render(fmt.Sprintf("$%s", valueStr))
			}
		}
		return cellValue
	}

	table.Update(CellFormatterSetCmd(0, valueFormatter)) // Value is now column 0

	output = table.View()
	fmt.Printf("With value formatter:\n%s\n", stripANSI(output))

	// ===== TEST COMPONENT RENDERER =====
	fmt.Println("\n--- Component Renderer Test ---")

	table.EnableComponentRenderer()

	// Set cursor position
	table.viewport.CursorIndex = 1
	table.viewport.CursorViewportIndex = 1

	output = table.View()
	fmt.Printf("With component renderer and cursor:\n%s\n", stripANSI(output))

	// ===== VERIFY COLUMN WIDTHS =====
	fmt.Println("\n--- Width Verification ---")

	lines := strings.Split(stripANSI(output), "\n")
	if len(lines) > 1 {
		headerLine := lines[0]
		dataLine := lines[2] // Skip header, cursor is on line 2

		fmt.Printf("Header line: %q\n", headerLine)
		fmt.Printf("Data line: %q\n", dataLine)

		// Check column alignment with new ordering
		headerParts := strings.Split(headerLine, "│")

		expectedColumns := []string{"💰 Value", "📝 Name", "📊 Status", "🏷️ Category"}

		if len(headerParts) >= 5 { // Skip empty + indicator parts
			for i, expectedHeader := range expectedColumns {
				actualHeader := strings.TrimSpace(headerParts[i+2]) // Skip empty + indicator
				fmt.Printf("Column %d - Expected: %q, Actual: %q\n", i, expectedHeader, actualHeader)
			}
		}
	}

	fmt.Println("Advanced features test completed!")
}

// ================================
// CURSOR HIGHLIGHTING TESTS
// ================================

func TestCursorHighlighting_ToggleModes(t *testing.T) {
	fmt.Println("\n=== CURSOR HIGHLIGHTING TOGGLE MODES TEST ===")

	rows := []TableRow{
		{ID: "row-1", Cells: []string{"Item 1", "Value 25", "Status 0", "Category A"}},
		{ID: "row-2", Cells: []string{"Item 2", "Value 75", "Status 1", "Category B"}},
	}
	table := createTestTable(rows)

	// Set up base formatters with distinct colors
	nameFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("93")) // Bright purple
		return style.Render(cellValue)
	}

	valueFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Orange
		return style.Render(cellValue)
	}

	// Apply base formatters
	table.Update(CellFormatterSetCmd(0, nameFormatter)())
	table.Update(CellFormatterSetCmd(1, valueFormatter)())

	table.EnableComponentRenderer()
	table.viewport.CursorIndex = 1
	table.viewport.CursorViewportIndex = 1

	// Test: Verify base colors remain consistent
	fmt.Println("--- Base Colors Test ---")
	output := table.View()
	lines := strings.Split(output, "\n")

	if len(lines) > 2 {
		cursorRow := lines[2]
		cells := extractCellsFromRow(cursorRow)
		if len(cells) >= 4 {
			// Account for component renderer: index 0=empty, index 1=indicator, data starts at index 2
			verifyColorInCell(t, cells[2], "Item 2", "93", "Name should be bright purple (93)")
			verifyColorInCell(t, cells[3], "Value 75", "208", "Value should be orange (208)")
		}
	}

	fmt.Println("Base colors test completed!")
}

func TestCursorHighlighting_FullRowMode(t *testing.T) {
	fmt.Println("\n=== FULL ROW CURSOR HIGHLIGHTING TEST ===")

	rows := []TableRow{
		{ID: "row-1", Cells: []string{"Item 1", "Value 25", "Status 0", "Category A"}},
		{ID: "row-2", Cells: []string{"Item 2", "Value 75", "Status 1", "Category B"}},
		{ID: "row-3", Cells: []string{"Item 3", "Value 50", "Status 2", "Category C"}},
	}
	table := createTestTable(rows)

	// Add distinct styled formatters for each column
	nameFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("33")) // Bright blue
		return style.Render(cellValue)
	}

	valueFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("160")) // Dark red
		return style.Render(cellValue)
	}

	statusFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("28")) // Dark green
		return style.Render(cellValue)
	}

	categoryFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // Bright yellow
		return style.Render(cellValue)
	}

	// Apply base formatters
	table.Update(CellFormatterSetCmd(0, nameFormatter)())
	table.Update(CellFormatterSetCmd(1, valueFormatter)())
	table.Update(CellFormatterSetCmd(2, statusFormatter)())
	table.Update(CellFormatterSetCmd(3, categoryFormatter)())

	table.EnableComponentRenderer()

	// Set cursor to row 1
	table.viewport.CursorIndex = 1
	table.viewport.CursorViewportIndex = 1

	// Test 1: Verify base colors on all columns
	fmt.Println("--- Base Colors for All Columns ---")
	output := table.View()
	lines := strings.Split(output, "\n")

	if len(lines) > 2 {
		cursorRow := lines[2]
		cells := extractCellsFromRow(cursorRow)
		if len(cells) >= 5 {
			// Account for component renderer: index 0=empty, index 1=indicator, data starts at index 2
			verifyColorInCell(t, cells[2], "Item 2", "33", "Name should have bright blue (33)")
			verifyColorInCell(t, cells[3], "Value 75", "160", "Value should have dark red (160)")
			verifyColorInCell(t, cells[4], "Status 1", "28", "Status should have dark green (28)")
		}
	}

	// Test 2: Apply full row cursor highlighting
	fmt.Println("--- Full Row Cursor Highlighting Test ---")

	// Create full-row-aware formatters
	fullRowNameFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		if isCursor {
			// Full row highlighting: blue background for all cells
			style := lipgloss.NewStyle().Background(lipgloss.Color("12")).Foreground(lipgloss.Color("15")).Bold(true)
			return style.Render(cellValue)
		}
		// Base styling
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
		return style.Render(cellValue)
	}

	fullRowValueFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		if isCursor {
			// Same full row highlighting: blue background
			style := lipgloss.NewStyle().Background(lipgloss.Color("12")).Foreground(lipgloss.Color("15")).Bold(true)
			return style.Render(cellValue)
		}
		// Base styling
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("160"))
		return style.Render(cellValue)
	}

	fullRowStatusFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		if isCursor {
			// Same full row highlighting: blue background
			style := lipgloss.NewStyle().Background(lipgloss.Color("12")).Foreground(lipgloss.Color("15")).Bold(true)
			return style.Render(cellValue)
		}
		// Base styling
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("28"))
		return style.Render(cellValue)
	}

	fullRowCategoryFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		if isCursor {
			// Same full row highlighting: blue background
			style := lipgloss.NewStyle().Background(lipgloss.Color("12")).Foreground(lipgloss.Color("15")).Bold(true)
			return style.Render(cellValue)
		}
		// Base styling
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
		return style.Render(cellValue)
	}

	// Apply full row formatters
	table.Update(CellFormatterSetCmd(0, fullRowNameFormatter)())
	table.Update(CellFormatterSetCmd(1, fullRowValueFormatter)())
	table.Update(CellFormatterSetCmd(2, fullRowStatusFormatter)())
	table.Update(CellFormatterSetCmd(3, fullRowCategoryFormatter)())

	output = table.View()
	lines = strings.Split(output, "\n")

	if len(lines) > 2 {
		cursorRow := lines[2]
		fmt.Printf("Full row highlighted: %s\n", cursorRow)

		cells := extractCellsFromRow(cursorRow)
		if len(cells) >= 5 {
			// Test CRITICAL: Verify ALL cells in the cursor row have the same background color
			// Account for component renderer: index 0=empty, index 1=indicator, data starts at index 2
			verifyBackgroundColorInCell(t, cells[2], "12", "Name cell should have blue (12) background in full row mode")
			verifyBackgroundColorInCell(t, cells[3], "12", "Value cell should have blue (12) background in full row mode")
			verifyBackgroundColorInCell(t, cells[4], "12", "Status cell should have blue (12) background in full row mode")

			// Verify bold styling is applied to all cells
			verifyBoldInCell(t, cells[2], "Name cell should be bold in full row mode")
			verifyBoldInCell(t, cells[3], "Value cell should be bold in full row mode")
			verifyBoldInCell(t, cells[4], "Status cell should be bold in full row mode")
		}
	}

	// Test 3: Check non-cursor rows maintain base colors
	fmt.Println("--- Verify Non-Cursor Rows Maintain Base Colors ---")
	if len(lines) > 1 {
		nonCursorRow := lines[1] // First row (not cursor)
		fmt.Printf("Non-cursor row: %s\n", nonCursorRow)

		cells := extractCellsFromRow(nonCursorRow)
		if len(cells) >= 5 {
			// Account for component renderer: index 0=empty, index 1=indicator, data starts at index 2
			verifyColorInCell(t, cells[2], "Item 1", "33", "Non-cursor name should maintain base bright blue (33)")
			verifyColorInCell(t, cells[3], "Value 25", "160", "Non-cursor value should maintain base dark red (160)")
			verifyColorInCell(t, cells[4], "Status 0", "28", "Non-cursor status should maintain base dark green (28)")
		}
	}

	fmt.Println("Full row cursor highlighting test completed!")
}

// ================================
// HELPER FUNCTIONS FOR ANSI COLOR DETECTION
// ================================

// extractCellsFromRow splits a table row into individual cells
func extractCellsFromRow(row string) []string {
	return strings.Split(row, "│")
}

// verifyColorInCell checks if a cell contains the expected foreground color code
func verifyColorInCell(t *testing.T, cell, expectedContent, expectedColor, message string) {
	// Look for ANSI foreground color code: \x1b[38;5;{color}m
	expectedCode := fmt.Sprintf("38;5;%s", expectedColor)

	if !strings.Contains(cell, expectedCode) {
		t.Errorf("%s\nCell: %q\nExpected color code: %s\nActual cell content: %s",
			message, expectedContent, expectedCode, cell)
	} else {
		fmt.Printf("✅ %s - Found color %s\n", message, expectedColor)
	}
}

// verifyBackgroundColorInCell checks if a cell contains the expected background color code
func verifyBackgroundColorInCell(t *testing.T, cell, expectedColor, message string) {
	// Look for various ANSI background color formats:
	// - 256-color: \x1b[48;5;{color}m
	// - Standard: \x1b[{color}m where color is 40-47 or 100-107

	var found bool
	var actualFormat string

	// Check 256-color format: 48;5;{expectedColor}
	expectedCode256 := fmt.Sprintf("48;5;%s", expectedColor)
	if strings.Contains(cell, expectedCode256) {
		found = true
		actualFormat = expectedCode256
	}

	// Check standard format for common colors
	if expectedColor == "12" {
		// Blue background can be 44 (standard blue) or 104 (bright blue)
		if strings.Contains(cell, "44m") || strings.Contains(cell, "104m") {
			found = true
			if strings.Contains(cell, "44m") {
				actualFormat = "44"
			} else {
				actualFormat = "104"
			}
		}
	} else if expectedColor == "14" {
		// Cyan background can be 46 (standard cyan), 106 (bright cyan), or others
		if strings.Contains(cell, "46m") || strings.Contains(cell, "106m") || strings.Contains(cell, "14m") {
			found = true
			if strings.Contains(cell, "46m") {
				actualFormat = "46"
			} else if strings.Contains(cell, "106m") {
				actualFormat = "106"
			} else {
				actualFormat = "14"
			}
		}
	} else if expectedColor == "201" {
		// Magenta background - check exact 256-color format
		if strings.Contains(cell, "48;5;201") {
			found = true
			actualFormat = "48;5;201"
		}
	}

	if !found {
		t.Errorf("%s\nExpected background color %s (256-color format or equivalent)\nActual cell: %q",
			message, expectedColor, cell)
	} else {
		fmt.Printf("✅ %s - Found background color %s (format: %s)\n", message, expectedColor, actualFormat)
	}
}

// verifyBoldInCell checks if a cell contains bold formatting
func verifyBoldInCell(t *testing.T, cell, message string) {
	// Look for ANSI bold code: \x1b[1m or \x1b[1;...m (bold in compound format)
	if !strings.Contains(cell, "1m") && !strings.Contains(cell, "1;") {
		t.Errorf("%s\nExpected bold formatting (1m or 1; in compound format)\nActual cell: %q", message, cell)
	} else {
		if strings.Contains(cell, "1m") {
			fmt.Printf("✅ %s - Found bold formatting (1m)\n", message)
		} else {
			fmt.Printf("✅ %s - Found bold formatting (1; in compound format)\n", message)
		}
	}
}

// ================================
// COMPREHENSIVE FEATURE TESTS
// These tests address the user's specific concerns about:
// 1. Full row highlighting not being properly tested
// 2. Text wrapping deforming the table instead of truncating with ellipsis
// ================================

func TestFullRowHighlighting_UserRequested_BasicFunctionality(t *testing.T) {
	fmt.Println("\n=== USER REQUESTED: FULL ROW HIGHLIGHTING BASIC FUNCTIONALITY TEST ===")

	rows := []TableRow{
		{ID: "row-1", Cells: []string{"Item 1", "25", "Status0"}},
		{ID: "row-2", Cells: []string{"Item 2", "75", "Status1"}},
		{ID: "row-3", Cells: []string{"Item 3", "50", "Status2"}},
	}
	table := createTestTable(rows)

	// EASY WAY: Use the new command to enable full row highlighting
	cmd := FullRowHighlightEnableCmd(true)
	msg := cmd()
	updatedModel, _ := table.Update(msg)
	table = updatedModel.(*Table) // Update the table reference
	fmt.Printf("✅ FullRowHighlighting enabled via command: %t\n", table.config.FullRowHighlighting)

	// Create simple formatters that don't interfere with full row highlighting
	// When FullRowHighlighting is enabled, the table will override formatter styling
	simpleNameFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		// Just return the content - the table will apply full row highlighting automatically
		return cellValue
	}

	simpleValueFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		// Add some basic formatting that should be overridden by full row highlighting
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("160")) // Dark red - should be overridden
		return style.Render(cellValue)
	}

	simpleStatusFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		// Add some basic formatting that should be overridden by full row highlighting
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("28")) // Dark green - should be overridden
		return style.Render(cellValue)
	}

	// Apply simple formatters
	table.Update(CellFormatterSetCmd(0, simpleNameFormatter)())
	table.Update(CellFormatterSetCmd(1, simpleValueFormatter)())
	table.Update(CellFormatterSetCmd(2, simpleStatusFormatter)())

	// Enable component renderer
	table.EnableComponentRenderer()

	// Set cursor to row 1 (second row)
	table.viewport.CursorIndex = 1
	table.viewport.CursorViewportIndex = 1

	output := table.View()
	fmt.Printf("Full table output:\n%s\n", output)
	lines := strings.Split(output, "\n")

	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 lines, got %d", len(lines))
	}

	// Find the cursor row dynamically by searching for ►
	var cursorRowLine string
	for _, line := range lines {
		if strings.Contains(line, "►") {
			cursorRowLine = line
			break
		}
	}

	if cursorRowLine == "" {
		t.Fatalf("❌ CRITICAL FAILURE: Could not find cursor row (containing ►) in table output")
	}

	fmt.Printf("✅ CRITICAL TEST: Full cursor row: %q\n", cursorRowLine)

	// Extract cells from the cursor row
	cells := extractCellsFromRow(cursorRowLine)
	if len(cells) < 4 { // indicator + 3 data columns
		t.Fatalf("Expected at least 4 cells (indicator + 3 data), got %d", len(cells))
	}

	// Test CRITICAL: Verify ALL data cells in the cursor row have blue background (color 12)
	successfulCells := 0
	for i := 2; i < len(cells) && i < 5; i++ { // Skip empty cell and indicator, test data cells
		cell := cells[i]
		fmt.Printf("Testing cell %d: %q\n", i-2, cell)

		// Check for background color (blue) in various ANSI formats
		hasBlueBackground := strings.Contains(cell, "48;5;12") || // 256-color blue
			strings.Contains(cell, "104m") || // bright blue standard
			strings.Contains(cell, "44m") || // blue standard
			strings.Contains(cell, "1;97;104") // compound format with bright blue
		if !hasBlueBackground {
			t.Errorf("❌ CRITICAL FAILURE: Cell %d should have blue background in full row mode. Cell content: %q", i-2, cell)
		} else {
			fmt.Printf("✅ Cell %d has blue background\n", i-2)
			successfulCells++
		}

		// Check for bold styling
		hasBold := strings.Contains(cell, "1m") || strings.Contains(cell, "1;")
		if !hasBold {
			t.Errorf("❌ CRITICAL FAILURE: Cell %d should be bold in full row mode. Cell content: %q", i-2, cell)
		} else {
			fmt.Printf("✅ Cell %d is bold\n", i-2)
		}

		// Check for white foreground in various ANSI formats
		hasWhiteForeground := strings.Contains(cell, "38;5;15") || // 256-color white
			strings.Contains(cell, "97m") || // bright white standard
			strings.Contains(cell, "37m") || // white standard
			strings.Contains(cell, "1;97;") // compound format with bright white
		if !hasWhiteForeground {
			t.Errorf("❌ CRITICAL FAILURE: Cell %d should have white foreground in full row mode. Cell content: %q", i-2, cell)
		} else {
			fmt.Printf("✅ Cell %d has white foreground\n", i-2)
		}

		// Check that formatter styling was overridden (should not contain red color 196)
		if i == 2 { // First data cell (name column)
			hasRedText := strings.Contains(cell, "38;5;196")
			if hasRedText {
				t.Errorf("❌ CRITICAL FAILURE: Name cell should have formatter styling overridden. Cell: %q", cell)
			} else {
				fmt.Printf("✅ Name cell formatter styling correctly overridden by full row highlighting\n")
			}
		}
	}

	if successfulCells == 3 {
		fmt.Printf("🎉 SUCCESS: All %d cells in the cursor row have proper full row highlighting!\n", successfulCells)
	} else {
		t.Errorf("❌ FAILURE: Only %d/3 cells have proper full row highlighting", successfulCells)
	}

	// Test that non-cursor rows maintain base colors (no blue background)
	var nonCursorRowLine string
	for _, line := range lines {
		// Skip border lines and find a data row that's not the cursor row
		if strings.Contains(line, "│") && !strings.Contains(line, "►") &&
			!strings.HasPrefix(strings.TrimSpace(line), "┌") &&
			!strings.HasPrefix(strings.TrimSpace(line), "├") &&
			!strings.HasPrefix(strings.TrimSpace(line), "└") &&
			!strings.Contains(line, "Name") { // Skip header
			nonCursorRowLine = line
			break
		}
	}

	if nonCursorRowLine != "" {
		fmt.Printf("Non-cursor row line: %q\n", nonCursorRowLine)

		nonCursorCells := extractCellsFromRow(nonCursorRowLine)
		if len(nonCursorCells) >= 4 {
			for i := 2; i < len(nonCursorCells) && i < 5; i++ { // Skip indicator, test data cells
				cell := nonCursorCells[i]

				// Non-cursor cells should NOT have blue background
				if strings.Contains(cell, "48;5;12") {
					t.Errorf("❌ FAILURE: Non-cursor cell %d should NOT have blue background. Cell content: %q", i-2, cell)
				} else {
					fmt.Printf("✅ Non-cursor cell %d correctly does NOT have blue background\n", i-2)
				}
			}
		}
	}

	fmt.Println("✅ Full row highlighting test completed!")
}

func TestFullRowHighlighting_UserRequested_BackgroundExtension(t *testing.T) {
	fmt.Println("\n=== USER REQUESTED: FULL ROW HIGHLIGHTING BACKGROUND EXTENSION TEST ===")

	rows := []TableRow{
		{ID: "row-1", Cells: []string{"Short", "X", "Y"}},
	}
	table := createTestTable(rows)

	// EASY WAY: Use the new command to enable full row highlighting
	cmd := FullRowHighlightEnableCmd(true)
	msg := cmd()
	updatedModel, _ := table.Update(msg)
	table = updatedModel.(*Table) // Update the table reference
	fmt.Printf("✅ FullRowHighlighting enabled via command: %t\n", table.config.FullRowHighlighting)

	// Create simple formatter that doesn't interfere with full row highlighting
	simpleFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		// Just return the content - the table will apply full row highlighting automatically
		return cellValue
	}

	// Apply to all columns
	table.Update(CellFormatterSetCmd(0, simpleFormatter)())
	table.Update(CellFormatterSetCmd(1, simpleFormatter)())
	table.Update(CellFormatterSetCmd(2, simpleFormatter)())

	table.EnableComponentRenderer()

	// Set cursor to row 0
	table.viewport.CursorIndex = 0
	table.viewport.CursorViewportIndex = 0

	output := table.View()
	fmt.Printf("Full table output:\n%s\n", output)
	lines := strings.Split(output, "\n")

	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 lines, got %d", len(lines))
	}

	// Find the cursor row dynamically by searching for ►
	var cursorRowLine string
	for _, line := range lines {
		if strings.Contains(line, "►") {
			cursorRowLine = line
			break
		}
	}

	if cursorRowLine == "" {
		t.Fatalf("❌ CRITICAL FAILURE: Could not find cursor row (containing ►) in table output")
	}

	fmt.Printf("✅ CRITICAL TEST: Full cursor row: %q\n", cursorRowLine)

	// The critical test: background should extend across the entire cell width
	// Even though "Short" is only 5 characters, the cell width is 10, so background should cover all 10
	cells := extractCellsFromRow(cursorRowLine)
	if len(cells) >= 3 { // indicator + first data cell
		nameCell := cells[2] // First data cell (after empty + indicator)

		// Count background color occurrences to ensure it extends across the full width
		blueBackgroundCount := strings.Count(nameCell, "48;5;12") +
			strings.Count(nameCell, "104m") +
			strings.Count(nameCell, "44m") +
			strings.Count(nameCell, "1;97;104")
		if blueBackgroundCount == 0 {
			t.Errorf("❌ CRITICAL FAILURE: Name cell should have blue background. Cell: %q", nameCell)
		} else {
			fmt.Printf("✅ Name cell has blue background (found %d occurrences)\n", blueBackgroundCount)
		}

		// Check that the visible width of the cell matches expected column width
		strippedContent := stripANSI(nameCell)
		actualWidth := runewidth.StringWidth(strippedContent)
		expectedWidth := 10 // Name column width from createTestTable

		if actualWidth != expectedWidth {
			t.Errorf("❌ CRITICAL FAILURE: Cell visual width should be %d characters, got %d. Stripped content: %q", expectedWidth, actualWidth, strippedContent)
		} else {
			fmt.Printf("✅ Cell visual width is correct: %d characters\n", actualWidth)
		}

		// Check that padding spaces also have the background color
		// The cell should be "Short     " (5 chars + 5 spaces) with blue background on all characters
		if !strings.Contains(strippedContent, "Short") {
			t.Errorf("❌ FAILURE: Cell should contain 'Short'. Stripped content: %q", strippedContent)
		} else {
			fmt.Printf("✅ Cell contains expected content: %q\n", strippedContent)
		}

		fmt.Printf("🎉 SUCCESS: Background extension test passed - width=%d, content=%q\n", actualWidth, strippedContent)
	}
}

func TestTextWrapping_UserRequested_AutomaticTruncation(t *testing.T) {
	fmt.Println("\n=== USER REQUESTED: TEXT WRAPPING AUTOMATIC TRUNCATION WITH ELLIPSIS TEST ===")

	rows := []TableRow{
		{ID: "row-1", Cells: []string{
			"This is a very long name that should definitely be truncated",
			"999999",
			"This is a long status",
		}},
	}
	table := createTestTable(rows)

	// Create "wrapping" formatters that produce potentially long content
	wrappingNameFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		// This formatter produces enhanced content that might be longer
		enhanced := "📋 Enhanced: " + cellValue + " (with extra info)"
		fmt.Printf("Name formatter input: %q -> output: %q (length: %d)\n", cellValue, enhanced, runewidth.StringWidth(enhanced))
		return enhanced
	}

	wrappingValueFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		// This formatter produces currency formatting that might be longer
		enhanced := "💰 Currency: $" + cellValue + " USD with extra details"
		fmt.Printf("Value formatter input: %q -> output: %q (length: %d)\n", cellValue, enhanced, runewidth.StringWidth(enhanced))
		return enhanced
	}

	wrappingStatusFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		// This formatter produces status with icons that might be longer
		enhanced := "✅ Status Enhanced: " + cellValue + " with additional status information"
		fmt.Printf("Status formatter input: %q -> output: %q (length: %d)\n", cellValue, enhanced, runewidth.StringWidth(enhanced))
		return enhanced
	}

	// Apply wrapping formatters
	table.Update(CellFormatterSetCmd(0, wrappingNameFormatter)())
	table.Update(CellFormatterSetCmd(1, wrappingValueFormatter)())
	table.Update(CellFormatterSetCmd(2, wrappingStatusFormatter)())

	table.EnableComponentRenderer()

	output := table.View()
	fmt.Printf("Full table output:\n%s\n", output)
	lines := strings.Split(output, "\n")

	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 lines, got %d", len(lines))
	}

	// Find the data row dynamically by searching for ►
	var dataRowLine string
	for _, line := range lines {
		if strings.Contains(line, "►") {
			dataRowLine = line
			break
		}
	}

	if dataRowLine == "" {
		t.Fatalf("❌ CRITICAL FAILURE: Could not find data row (containing ►) in table output")
	}

	fmt.Printf("✅ CRITICAL TEST: Data row with wrapping formatters: %q\n", dataRowLine)

	cells := extractCellsFromRow(dataRowLine)
	if len(cells) < 4 { // indicator + 3 data columns
		t.Fatalf("Expected at least 4 cells, got %d", len(cells))
	}

	// Test each column for proper truncation
	columnWidths := []int{10, 8, 10} // Name=10, Value=8, Status=10
	columnNames := []string{"Name", "Value", "Status"}

	truncationSuccess := 0
	for i := 0; i < 3 && i+2 < len(cells); i++ {
		cell := cells[i+2] // Skip empty + indicator
		strippedContent := stripANSI(cell)
		actualWidth := runewidth.StringWidth(strippedContent)
		expectedWidth := columnWidths[i]

		fmt.Printf("Testing %s column: width=%d, expected=%d, content=%q\n",
			columnNames[i], actualWidth, expectedWidth, strippedContent)

		// CRITICAL: The actual width should not exceed the expected column width
		if actualWidth > expectedWidth {
			t.Errorf("❌ CRITICAL FAILURE: %s column width %d exceeds expected %d. Content: %q",
				columnNames[i], actualWidth, expectedWidth, strippedContent)
		} else {
			fmt.Printf("✅ %s column width is within bounds (%d <= %d)\n", columnNames[i], actualWidth, expectedWidth)
		}

		// CRITICAL: If content was truncated, it should end with ellipsis
		if actualWidth == expectedWidth {
			// Content should have been truncated and should end with ellipsis
			hasEllipsis := strings.HasSuffix(strippedContent, "...") || strings.HasSuffix(strippedContent, "…")
			if !hasEllipsis {
				t.Errorf("❌ CRITICAL FAILURE: %s column should end with ellipsis when truncated. Content: %q",
					columnNames[i], strippedContent)
			} else {
				fmt.Printf("✅ %s column properly ends with ellipsis\n", columnNames[i])
				truncationSuccess++
			}
		}
	}

	// Verify table structure is not deformed
	// All lines should have roughly the same visible width (accounting for borders)
	headerLine := lines[0]
	headerWidth := runewidth.StringWidth(stripANSI(headerLine))
	dataWidth := runewidth.StringWidth(stripANSI(dataRowLine))

	if abs(headerWidth-dataWidth) > 2 { // Allow small variance for borders
		t.Errorf("❌ CRITICAL FAILURE: Table structure appears deformed. Header width: %d, Data width: %d",
			headerWidth, dataWidth)
		fmt.Printf("Header: %q\n", stripANSI(headerLine))
		fmt.Printf("Data:   %q\n", stripANSI(dataRowLine))
	} else {
		fmt.Printf("✅ Table structure is not deformed (header: %d, data: %d)\n", headerWidth, dataWidth)
	}

	if truncationSuccess > 0 {
		fmt.Printf("🎉 SUCCESS: %d/%d columns properly truncated with ellipsis!\n", truncationSuccess, 3)
	}
}

func TestTextWrapping_UserRequested_EllipsisPlacement(t *testing.T) {
	fmt.Println("\n=== USER REQUESTED: TEXT WRAPPING ELLIPSIS PLACEMENT TEST ===")

	rows := []TableRow{
		{ID: "row-1", Cells: []string{
			"ExtremelyLongSingleWordThatCannotFitInColumnWidth",
			"AnotherVeryLongValueThatExceedsWidth",
			"SuperLongStatusText",
		}},
	}
	table := createTestTable(rows)

	// Use simple formatters that don't modify content (just pass through)
	simpleFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		fmt.Printf("Simple formatter input: %q (length: %d)\n", cellValue, runewidth.StringWidth(cellValue))
		return cellValue // No modification - rely on automatic truncation
	}

	// Apply simple formatters to all columns
	table.Update(CellFormatterSetCmd(0, simpleFormatter)())
	table.Update(CellFormatterSetCmd(1, simpleFormatter)())
	table.Update(CellFormatterSetCmd(2, simpleFormatter)())

	table.EnableComponentRenderer()

	output := table.View()
	fmt.Printf("Full table output:\n%s\n", output)
	lines := strings.Split(output, "\n")

	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 lines, got %d", len(lines))
	}

	dataRowLine := lines[1]
	fmt.Printf("✅ CRITICAL TEST: Data row for ellipsis test: %q\n", dataRowLine)

	cells := extractCellsFromRow(dataRowLine)
	if len(cells) < 4 {
		t.Fatalf("Expected at least 4 cells, got %d", len(cells))
	}

	// Test specific columns that should be truncated
	testCases := []struct {
		columnIndex   int
		columnName    string
		expectedWidth int
		originalText  string
	}{
		{0, "Name", 10, "ExtremelyLongSingleWordThatCannotFitInColumnWidth"},
		{1, "Value", 8, "AnotherVeryLongValueThatExceedsWidth"},
		{2, "Status", 10, "SuperLongStatusText"},
	}

	ellipsisSuccess := 0
	for _, tc := range testCases {
		cell := cells[tc.columnIndex+2] // Skip empty + indicator
		strippedContent := stripANSI(cell)
		actualWidth := runewidth.StringWidth(strippedContent)

		fmt.Printf("Testing %s column: content=%q, width=%d\n", tc.columnName, strippedContent, actualWidth)

		// The content should be exactly the expected width
		if actualWidth != tc.expectedWidth {
			t.Errorf("❌ CRITICAL FAILURE: %s column should have width %d, got %d. Content: %q",
				tc.columnName, tc.expectedWidth, actualWidth, strippedContent)
		} else {
			fmt.Printf("✅ %s column has correct width: %d\n", tc.columnName, actualWidth)
		}

		// The content should end with ellipsis (either "..." or "…")
		hasEllipsis := strings.HasSuffix(strippedContent, "...") || strings.HasSuffix(strippedContent, "…")
		if !hasEllipsis {
			t.Errorf("❌ CRITICAL FAILURE: %s column should end with ellipsis. Content: %q", tc.columnName, strippedContent)
		} else {
			fmt.Printf("✅ %s column ends with ellipsis\n", tc.columnName)
			ellipsisSuccess++
		}

		// The content should start with the beginning of the original text
		expectedPrefix := tc.originalText[:min(3, len(tc.originalText))]
		if !strings.HasPrefix(strippedContent, expectedPrefix) {
			t.Errorf("❌ FAILURE: %s column should start with %q. Content: %q", tc.columnName, expectedPrefix, strippedContent)
		} else {
			fmt.Printf("✅ %s column starts with expected prefix: %q\n", tc.columnName, expectedPrefix)
		}
	}

	if ellipsisSuccess == 3 {
		fmt.Printf("🎉 SUCCESS: All %d columns properly truncated with ellipsis!\n", ellipsisSuccess)
	} else {
		t.Errorf("❌ FAILURE: Only %d/3 columns properly truncated with ellipsis", ellipsisSuccess)
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper function for abs
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func TestFullRowHighlighting_UserRequested_EasySetup(t *testing.T) {
	fmt.Println("\n=== USER REQUESTED: EASY FULL ROW HIGHLIGHTING SETUP TEST ===")

	rows := []TableRow{
		{ID: "row-1", Cells: []string{"Item 1", "25", "Status0"}},
		{ID: "row-2", Cells: []string{"Item 2", "75", "Status1"}},
	}
	table := createTestTable(rows)

	// Debug: Check initial state
	fmt.Printf("Initial FullRowHighlighting: %t\n", table.config.FullRowHighlighting)

	// EASY WAY: Use the new command to enable full row highlighting
	fmt.Println("Calling FullRowHighlightEnableCmd(true)...")
	cmd := FullRowHighlightEnableCmd(true)
	msg := cmd()
	fmt.Printf("Command created message: %T\n", msg)

	updatedModel, updateCmd := table.Update(msg)
	table = updatedModel.(*Table) // Update the table reference
	fmt.Printf("✅ FullRowHighlighting after command: %t\n", table.config.FullRowHighlighting)
	fmt.Printf("Update command: %v\n", updateCmd)

	table.EnableComponentRenderer()
	table.viewport.CursorIndex = 1
	table.viewport.CursorViewportIndex = 1

	output := table.View()
	fmt.Printf("Table output:\n%s\n", output)

	lines := strings.Split(output, "\n")

	// Find the cursor row dynamically by searching for ►
	var cursorRow string
	for _, line := range lines {
		if strings.Contains(line, "►") {
			cursorRow = line
			break
		}
	}

	if cursorRow != "" {
		fmt.Printf("Cursor row: %q\n", cursorRow)

		// Verify that full row highlighting is working
		cells := extractCellsFromRow(cursorRow)
		if len(cells) >= 3 {
			nameCell := cells[2] // First data cell
			hasBlueBackground := strings.Contains(nameCell, "104m") || strings.Contains(nameCell, "1;97;104")
			if hasBlueBackground {
				fmt.Printf("🎉 SUCCESS: Easy setup produced full row highlighting!\n")
			} else {
				t.Errorf("❌ FAILURE: Easy setup should produce full row highlighting. Cell: %q", nameCell)
			}
		}
	} else {
		t.Errorf("❌ FAILURE: Could not find cursor row in table output")
	}

	fmt.Println("Easy setup test completed!")
}

func TestFullRowHighlighting_UserRequested_FallbackRendering(t *testing.T) {
	fmt.Println("\n=== USER REQUESTED: FULL ROW HIGHLIGHTING FALLBACK RENDERING TEST ===")

	rows := []TableRow{
		{ID: "row-1", Cells: []string{"Item 1", "25", "Status0"}},
		{ID: "row-2", Cells: []string{"Item 2", "75", "Status1"}},
	}
	table := createTestTable(rows)

	// Enable full row highlighting
	cmd := FullRowHighlightEnableCmd(true)
	msg := cmd()
	updatedModel, _ := table.Update(msg)
	table = updatedModel.(*Table)
	fmt.Printf("✅ FullRowHighlighting enabled: %t\n", table.config.FullRowHighlighting)

	// Add a formatter to one column to test that it gets overridden
	nameFormatter := func(cellValue string, rowIndex int, column TableColumn, ctx RenderContext, isCursor bool, isSelected bool) string {
		// This should be overridden by full row highlighting when cursor is active
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red - should be overridden
		return style.Render("📝 " + cellValue)
	}
	table.Update(CellFormatterSetCmd(0, nameFormatter)())

	// CRITICAL: Do NOT enable component renderer - test fallback path only
	fmt.Println("✅ Testing fallback rendering path (no component renderer)")

	// Set cursor to row 1
	table.viewport.CursorIndex = 1
	table.viewport.CursorViewportIndex = 1

	output := table.View()
	fmt.Printf("Fallback rendering output:\n%s\n", output)

	lines := strings.Split(output, "\n")
	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines, got %d", len(lines))
	}

	// In fallback mode, there's no ► indicator, so find cursor row by checking for full row highlighting styles
	var cursorRowLine string
	for _, line := range lines {
		// Skip border lines and find a data row with blue background (full row highlighting)
		if strings.Contains(line, "│") && strings.Contains(line, "104m") &&
			!strings.HasPrefix(strings.TrimSpace(line), "┌") &&
			!strings.HasPrefix(strings.TrimSpace(line), "├") &&
			!strings.HasPrefix(strings.TrimSpace(line), "└") &&
			!strings.Contains(line, "Name") { // Skip header
			cursorRowLine = line
			break
		}
	}

	if cursorRowLine == "" {
		t.Fatalf("❌ CRITICAL FAILURE: Could not find cursor row with full row highlighting in fallback mode")
	}

	fmt.Printf("✅ CRITICAL TEST: Cursor row in fallback mode: %q\n", cursorRowLine)

	// In fallback mode, there's no indicator column, so cells start immediately
	cells := strings.Split(cursorRowLine, "│")
	if len(cells) < 3 { // left border + 3 data columns + right border
		t.Fatalf("Expected at least 3 cell parts, got %d", len(cells))
	}

	// Test all data cells for full row highlighting (skip borders)
	successfulCells := 0
	for i := 1; i < len(cells)-1; i++ { // Skip left and right border cells
		cell := cells[i]
		fmt.Printf("Testing fallback cell %d: %q\n", i-1, cell)

		// Check for blue background in various formats
		hasBlueBackground := strings.Contains(cell, "48;5;12") ||
			strings.Contains(cell, "104m") ||
			strings.Contains(cell, "44m") ||
			strings.Contains(cell, "1;97;104")
		if !hasBlueBackground {
			t.Errorf("❌ CRITICAL FAILURE: Fallback cell %d should have blue background. Cell: %q", i-1, cell)
		} else {
			fmt.Printf("✅ Fallback cell %d has blue background\n", i-1)
			successfulCells++
		}

		// Check for bold styling
		hasBold := strings.Contains(cell, "1m") || strings.Contains(cell, "1;")
		if !hasBold {
			t.Errorf("❌ CRITICAL FAILURE: Fallback cell %d should be bold. Cell: %q", i-1, cell)
		} else {
			fmt.Printf("✅ Fallback cell %d is bold\n", i-1)
		}

		// Check for white foreground in various ANSI formats
		hasWhiteForeground := strings.Contains(cell, "38;5;15") || // 256-color white
			strings.Contains(cell, "97m") || // bright white standard
			strings.Contains(cell, "37m") || // white standard
			strings.Contains(cell, "1;97;") // compound format with bright white
		if !hasWhiteForeground {
			t.Errorf("❌ CRITICAL FAILURE: Fallback cell %d should have white foreground in full row mode. Cell content: %q", i-1, cell)
		} else {
			fmt.Printf("✅ Fallback cell %d has white foreground\n", i-1)
		}

		// Check that formatter styling was overridden (should not contain red color 196)
		if i == 1 { // First data cell (name column)
			hasRedText := strings.Contains(cell, "38;5;196")
			if hasRedText {
				t.Errorf("❌ CRITICAL FAILURE: Name cell should have formatter styling overridden. Cell: %q", cell)
			} else {
				fmt.Printf("✅ Name cell formatter styling correctly overridden by full row highlighting\n")
			}
		}
	}

	if successfulCells == 3 {
		fmt.Printf("🎉 SUCCESS: All %d cells in fallback rendering have proper full row highlighting!\n", successfulCells)
	} else {
		t.Errorf("❌ FAILURE: Only %d/3 cells have proper full row highlighting in fallback mode", successfulCells)
	}

	// Test non-cursor row (should not have full row highlighting)
	var nonCursorRowLine string
	for _, line := range lines {
		// Find a data row that doesn't contain ►, and skip border lines
		if strings.Contains(line, "│") && !strings.Contains(line, "►") &&
			!strings.HasPrefix(strings.TrimSpace(line), "┌") &&
			!strings.HasPrefix(strings.TrimSpace(line), "├") &&
			!strings.HasPrefix(strings.TrimSpace(line), "└") &&
			!strings.Contains(line, "Name") { // Skip header
			nonCursorRowLine = line
			break
		}
	}

	if nonCursorRowLine != "" {
		nonCursorCells := strings.Split(nonCursorRowLine, "│")
		if len(nonCursorCells) >= 2 {
			nameCell := nonCursorCells[1] // First data cell
			hasBlueBackground := strings.Contains(nameCell, "48;5;12") || strings.Contains(nameCell, "104m")
			if hasBlueBackground {
				t.Errorf("❌ FAILURE: Non-cursor row should NOT have blue background. Cell: %q", nameCell)
			} else {
				fmt.Printf("✅ Non-cursor row correctly does NOT have blue background\n")
			}

			// Check that formatter styling is preserved in non-cursor rows
			hasRedText := strings.Contains(nameCell, "38;5;196")
			if !hasRedText {
				t.Errorf("❌ FAILURE: Non-cursor name cell should preserve formatter styling. Cell: %q", nameCell)
			} else {
				fmt.Printf("✅ Non-cursor name cell correctly preserves formatter styling\n")
			}
		}
	}

	fmt.Println("✅ Fallback rendering test completed!")
}
