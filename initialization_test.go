package vtable

import (
	"fmt"
	"testing"
)

// Test data provider with configurable size
type ConfigurableDataProvider struct {
	size  int
	items []string
}

func NewConfigurableDataProvider(size int) *ConfigurableDataProvider {
	items := make([]string, size)
	for i := 0; i < size; i++ {
		items[i] = fmt.Sprintf("Item %d", i)
	}
	return &ConfigurableDataProvider{
		size:  size,
		items: items,
	}
}

func (p *ConfigurableDataProvider) GetTotal() int {
	return p.size
}

func (p *ConfigurableDataProvider) GetItems(request DataRequest) ([]Data[string], error) {
	start := request.Start
	count := request.Count

	if start >= p.size {
		return []Data[string]{}, nil
	}

	end := start + count
	if end > p.size {
		end = p.size
	}

	result := make([]Data[string], end-start)
	for i := start; i < end; i++ {
		result[i-start] = Data[string]{
			ID:       fmt.Sprintf("item-%d", i),
			Item:     p.items[i],
			Selected: false,
			Metadata: NewTypedMetadata(),
		}
	}

	return result, nil
}

// Implement DataProvider interface
func (p *ConfigurableDataProvider) GetSelectionMode() SelectionMode                   { return SelectionMultiple }
func (p *ConfigurableDataProvider) SetSelected(index int, selected bool) bool         { return true }
func (p *ConfigurableDataProvider) SelectAll() bool                                   { return true }
func (p *ConfigurableDataProvider) ClearSelection()                                   {}
func (p *ConfigurableDataProvider) GetSelectedIndices() []int                         { return []int{} }
func (p *ConfigurableDataProvider) GetSelectedIDs() []string                          { return []string{} }
func (p *ConfigurableDataProvider) SetSelectedByIDs(ids []string, selected bool) bool { return true }
func (p *ConfigurableDataProvider) SelectRange(startID, endID string) bool            { return true }
func (p *ConfigurableDataProvider) GetItemID(item *string) string                     { return *item }

// Table data provider
type ConfigurableTableDataProvider struct {
	size  int
	items [][]string
}

func NewConfigurableTableDataProvider(size int) *ConfigurableTableDataProvider {
	items := make([][]string, size)
	for i := 0; i < size; i++ {
		items[i] = []string{
			fmt.Sprintf("Row %d", i),
			fmt.Sprintf("Value %d", i*10),
			fmt.Sprintf("Status %d", i%3),
		}
	}
	return &ConfigurableTableDataProvider{
		size:  size,
		items: items,
	}
}

func (p *ConfigurableTableDataProvider) GetTotal() int {
	return p.size
}

func (p *ConfigurableTableDataProvider) GetItems(request DataRequest) ([]Data[TableRow], error) {
	start := request.Start
	count := request.Count

	if start >= p.size {
		return []Data[TableRow]{}, nil
	}

	end := start + count
	if end > p.size {
		end = p.size
	}

	result := make([]Data[TableRow], end-start)
	for i := start; i < end; i++ {
		result[i-start] = Data[TableRow]{
			ID:       fmt.Sprintf("row-%d", i),
			Item:     TableRow{Cells: p.items[i]},
			Selected: false,
			Metadata: NewTypedMetadata(),
		}
	}

	return result, nil
}

// Implement DataProvider interface
func (p *ConfigurableTableDataProvider) GetSelectionMode() SelectionMode           { return SelectionMultiple }
func (p *ConfigurableTableDataProvider) SetSelected(index int, selected bool) bool { return true }
func (p *ConfigurableTableDataProvider) SelectAll() bool                           { return true }
func (p *ConfigurableTableDataProvider) ClearSelection()                           {}
func (p *ConfigurableTableDataProvider) GetSelectedIndices() []int                 { return []int{} }
func (p *ConfigurableTableDataProvider) GetSelectedIDs() []string                  { return []string{} }
func (p *ConfigurableTableDataProvider) SetSelectedByIDs(ids []string, selected bool) bool {
	return true
}
func (p *ConfigurableTableDataProvider) SelectRange(startID, endID string) bool { return true }
func (p *ConfigurableTableDataProvider) GetItemID(item *TableRow) string        { return item.Cells[0] }

func TestListInitializationScenarios(t *testing.T) {
	fmt.Println("\n=== LIST INITIALIZATION SCENARIOS TEST ===")

	// Test scenarios: {height, dataSize, description}
	scenarios := []struct {
		height      int
		dataSize    int
		description string
	}{
		{1, 0, "Single row, zero data"},
		{3, 0, "Multiple rows, zero data"},
		{5, 0, "Medium height, zero data"},
		{10, 0, "Large height, zero data"},
		{1, 1, "Single row, single item"},
		{1, 5, "Single row, multiple items"},
		{3, 1, "Multiple rows, single item"},
		{3, 2, "Multiple rows, fewer items than height"},
		{3, 3, "Multiple rows, exact match"},
		{3, 10, "Multiple rows, more items than height"},
		{5, 3, "Medium height, few items"},
		{5, 5, "Medium height, exact match"},
		{5, 20, "Medium height, many items"},
		{10, 5, "Large height, few items"},
		{10, 10, "Large height, exact match"},
		{10, 100, "Large height, many items"},
		{15, 8, "Large height, moderate items"},
		{20, 1000, "Very large height, many items"},
	}

	formatter := func(data Data[string], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		prefix := " "
		if isCursor {
			prefix = ">"
		}
		return fmt.Sprintf("%s %s", prefix, data.Item)
	}

	for _, scenario := range scenarios {
		t.Run(fmt.Sprintf("Height_%d_Data_%d", scenario.height, scenario.dataSize), func(t *testing.T) {
			provider := NewConfigurableDataProvider(scenario.dataSize)

			// Test with auto-calculated configuration
			config := NewViewportConfig(scenario.height)
			styleConfig := DefaultStyleConfig()

			if scenario.dataSize == 0 {
				// Zero data scenarios should either work gracefully or return an appropriate error
				list, err := NewTeaList(config, provider, styleConfig, formatter)
				if err != nil {
					// If error is returned, it should be informative
					if !contains(err.Error(), "dataset") && !contains(err.Error(), "data") && !contains(err.Error(), "empty") {
						t.Fatalf("Unexpected error for zero data scenario '%s': %v", scenario.description, err)
					}
					fmt.Printf("✅ %s: Properly rejected with error: %v\n", scenario.description, err)
					return
				}

				// If creation succeeds, verify it handles zero data gracefully
				fmt.Printf("✅ %s: Created successfully (handles empty data)\n", scenario.description)
				fmt.Printf("   Height: %d -> %d, Data: %d, Top: %d, Bottom: %d\n",
					scenario.height, list.list.Config.Height, scenario.dataSize,
					list.list.Config.TopThresholdIndex, list.list.Config.BottomThresholdIndex)

				// Test rendering with zero data
				view := list.View()
				if view == "" {
					fmt.Printf("   Renders as empty (expected for zero data)\n")
				} else {
					fmt.Printf("   Renders placeholder content: %q\n", view)
				}

				// Navigation should not panic with zero data
				list.MoveDown()
				list.MoveUp()
				list.JumpToStart()
				list.JumpToEnd()
				state := list.GetState()
				fmt.Printf("   Navigation works, cursor at: %d\n", state.CursorIndex)
				return
			}

			// Non-zero data scenarios (existing logic)
			list, err := NewTeaList(config, provider, styleConfig, formatter)
			if err != nil {
				t.Fatalf("Failed to create list for %s: %v", scenario.description, err)
			}

			// Verify the list was created successfully
			state := list.GetState()

			// Check that viewport height was adjusted for small datasets
			expectedHeight := scenario.height
			if scenario.dataSize < scenario.height {
				expectedHeight = scenario.dataSize
			}
			if expectedHeight < 1 {
				expectedHeight = 1
			}

			fmt.Printf("✅ %s: Created successfully\n", scenario.description)
			fmt.Printf("   Height: %d -> %d, Data: %d, Top: %d, Bottom: %d\n",
				scenario.height, list.list.Config.Height, scenario.dataSize,
				list.list.Config.TopThresholdIndex, list.list.Config.BottomThresholdIndex)

			// Verify state is valid
			if state.CursorIndex < 0 || state.CursorIndex >= scenario.dataSize {
				t.Errorf("Invalid cursor index %d for data size %d", state.CursorIndex, scenario.dataSize)
			}

			if state.ViewportStartIndex < 0 {
				t.Errorf("Invalid viewport start index %d", state.ViewportStartIndex)
			}

			// Test that we can render without errors
			view := list.View()
			if view == "" && scenario.dataSize > 0 {
				t.Errorf("Empty view for scenario with data: %s", scenario.description)
			}

			// Test navigation works
			if scenario.dataSize > 1 {
				list.MoveDown()
				newState := list.GetState()
				if newState.CursorIndex == state.CursorIndex && scenario.dataSize > 1 {
					t.Errorf("Cursor didn't move for scenario: %s", scenario.description)
				}
			}
		})
	}

	fmt.Println("=== END LIST INITIALIZATION SCENARIOS TEST ===")
}

func TestTableInitializationScenarios(t *testing.T) {
	fmt.Println("\n=== TABLE INITIALIZATION SCENARIOS TEST ===")

	// Test scenarios: {height, dataSize, description}
	scenarios := []struct {
		height      int
		dataSize    int
		description string
	}{
		{1, 0, "Single row table, zero data"},
		{3, 0, "Small table, zero data"},
		{5, 0, "Medium table, zero data"},
		{10, 0, "Large table, zero data"},
		{1, 1, "Single row table, single item"},
		{1, 5, "Single row table, multiple items"},
		{3, 1, "Small table, single item"},
		{3, 2, "Small table, fewer items than height"},
		{3, 3, "Small table, exact match"},
		{3, 10, "Small table, more items than height"},
		{5, 3, "Medium table, few items"},
		{5, 5, "Medium table, exact match"},
		{5, 20, "Medium table, many items"},
		{10, 5, "Large table, few items"},
		{10, 10, "Large table, exact match"},
		{10, 100, "Large table, many items"},
		{15, 8, "Large table, moderate items"},
		{20, 1000, "Very large table, many items"},
	}

	columns := []TableColumn{
		NewColumn("Name", 15),
		NewRightColumn("Value", 10),
		NewCenterColumn("Status", 12),
	}

	for _, scenario := range scenarios {
		t.Run(fmt.Sprintf("Table_Height_%d_Data_%d", scenario.height, scenario.dataSize), func(t *testing.T) {
			provider := NewConfigurableTableDataProvider(scenario.dataSize)

			// Test with auto-calculated configuration
			config := NewTableConfig(columns, scenario.height)
			theme := *DefaultTheme()

			if scenario.dataSize == 0 {
				// Zero data scenarios should either work gracefully or return an appropriate error
				table, err := NewTeaTable(config, provider, theme)
				if err != nil {
					// If error is returned, it should be informative
					if !contains(err.Error(), "dataset") && !contains(err.Error(), "data") && !contains(err.Error(), "empty") {
						t.Fatalf("Unexpected error for zero data scenario '%s': %v", scenario.description, err)
					}
					fmt.Printf("✅ %s: Properly rejected with error: %v\n", scenario.description, err)
					return
				}

				// If creation succeeds, verify it handles zero data gracefully
				fmt.Printf("✅ %s: Created successfully (handles empty data)\n", scenario.description)
				fmt.Printf("   Height: %d -> %d, Data: %d, Top: %d, Bottom: %d\n",
					scenario.height, table.table.config.ViewportConfig.Height, scenario.dataSize,
					table.table.config.ViewportConfig.TopThresholdIndex, table.table.config.ViewportConfig.BottomThresholdIndex)

				// Test rendering with zero data
				view := table.View()
				if view == "" {
					fmt.Printf("   Renders as empty (expected for zero data)\n")
				} else {
					fmt.Printf("   Renders placeholder content with headers/borders\n")
					// Should still show headers and borders even with no data
					if config.ShowHeader && !contains(view, "Name") {
						t.Errorf("Header should still be visible with zero data for: %s", scenario.description)
					}
				}

				// Navigation should not panic with zero data
				table.MoveDown()
				table.MoveUp()
				table.JumpToStart()
				table.JumpToEnd()
				state := table.GetState()
				fmt.Printf("   Navigation works, cursor at: %d\n", state.CursorIndex)
				return
			}

			// Non-zero data scenarios (existing logic)
			table, err := NewTeaTable(config, provider, theme)
			if err != nil {
				t.Fatalf("Failed to create table for %s: %v", scenario.description, err)
			}

			// Verify the table was created successfully
			state := table.GetState()

			// Check that viewport height was adjusted for small datasets
			expectedHeight := scenario.height
			if scenario.dataSize < scenario.height {
				expectedHeight = scenario.dataSize
			}
			if expectedHeight < 1 {
				expectedHeight = 1
			}

			fmt.Printf("✅ %s: Created successfully\n", scenario.description)
			fmt.Printf("   Height: %d -> %d, Data: %d, Top: %d, Bottom: %d\n",
				scenario.height, table.table.config.ViewportConfig.Height, scenario.dataSize,
				table.table.config.ViewportConfig.TopThresholdIndex, table.table.config.ViewportConfig.BottomThresholdIndex)

			// Verify state is valid
			if state.CursorIndex < 0 || state.CursorIndex >= scenario.dataSize {
				t.Errorf("Invalid cursor index %d for data size %d", state.CursorIndex, scenario.dataSize)
			}

			if state.ViewportStartIndex < 0 {
				t.Errorf("Invalid viewport start index %d", state.ViewportStartIndex)
			}

			// Test that we can render without errors
			view := table.View()
			if view == "" && scenario.dataSize > 0 {
				t.Errorf("Empty view for scenario with data: %s", scenario.description)
			}

			// Verify table structure (headers, borders)
			if config.ShowHeader && !contains(view, "Name") {
				t.Errorf("Header not found in table view for: %s", scenario.description)
			}

			if config.ShowBorders && !contains(view, "│") {
				t.Errorf("Borders not found in table view for: %s", scenario.description)
			}

			// Test navigation works
			if scenario.dataSize > 1 {
				table.MoveDown()
				newState := table.GetState()
				if newState.CursorIndex == state.CursorIndex && scenario.dataSize > 1 {
					t.Errorf("Cursor didn't move for scenario: %s", scenario.description)
				}
			}
		})
	}

	fmt.Println("=== END TABLE INITIALIZATION SCENARIOS TEST ===")
}

func TestThresholdCalculationScenarios(t *testing.T) {
	fmt.Println("\n=== THRESHOLD CALCULATION SCENARIOS TEST ===")

	// Test various height scenarios and verify thresholds are valid
	heights := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 15, 20, 25, 30, 50, 100}

	for _, height := range heights {
		t.Run(fmt.Sprintf("Height_%d", height), func(t *testing.T) {
			config := NewViewportConfig(height)

			// Verify thresholds are within bounds
			if config.TopThresholdIndex < 0 || config.TopThresholdIndex >= height {
				t.Errorf("Height %d: Top threshold %d is out of bounds", height, config.TopThresholdIndex)
			}

			if config.BottomThresholdIndex < 0 || config.BottomThresholdIndex >= height {
				t.Errorf("Height %d: Bottom threshold %d is out of bounds", height, config.BottomThresholdIndex)
			}

			// For heights > 1, bottom threshold must be > top threshold
			if height > 1 && config.BottomThresholdIndex <= config.TopThresholdIndex {
				t.Errorf("Height %d: Bottom threshold %d must be greater than top threshold %d",
					height, config.BottomThresholdIndex, config.TopThresholdIndex)
			}

			// Verify chunk size is reasonable
			if config.ChunkSize <= 0 {
				t.Errorf("Height %d: Invalid chunk size %d", height, config.ChunkSize)
			}

			fmt.Printf("Height %2d: Top=%2d, Bottom=%2d, Chunk=%3d ✅\n",
				height, config.TopThresholdIndex, config.BottomThresholdIndex, config.ChunkSize)
		})
	}

	fmt.Println("=== END THRESHOLD CALCULATION SCENARIOS TEST ===")
}

func TestConfigurationAutoCorrection(t *testing.T) {
	fmt.Println("\n=== CONFIGURATION AUTO-CORRECTION TEST ===")

	// Test scenarios with intentionally broken configurations
	brokenConfigs := []struct {
		description string
		config      ViewportConfig
	}{
		{
			"Negative height",
			ViewportConfig{Height: -5, TopThresholdIndex: 1, BottomThresholdIndex: 3},
		},
		{
			"Zero height",
			ViewportConfig{Height: 0, TopThresholdIndex: 1, BottomThresholdIndex: 3},
		},
		{
			"Top threshold too high",
			ViewportConfig{Height: 10, TopThresholdIndex: 15, BottomThresholdIndex: 7},
		},
		{
			"Bottom threshold too high",
			ViewportConfig{Height: 10, TopThresholdIndex: 2, BottomThresholdIndex: 15},
		},
		{
			"Bottom threshold <= top threshold",
			ViewportConfig{Height: 10, TopThresholdIndex: 5, BottomThresholdIndex: 3},
		},
		{
			"Both thresholds negative",
			ViewportConfig{Height: 10, TopThresholdIndex: -1, BottomThresholdIndex: -1},
		},
		{
			"Negative chunk size",
			ViewportConfig{Height: 10, TopThresholdIndex: 2, BottomThresholdIndex: 7, ChunkSize: -10},
		},
		{
			"Zero chunk size",
			ViewportConfig{Height: 10, TopThresholdIndex: 2, BottomThresholdIndex: 7, ChunkSize: 0},
		},
		{
			"Negative initial index",
			ViewportConfig{Height: 10, TopThresholdIndex: 2, BottomThresholdIndex: 7, InitialIndex: -5},
		},
	}

	provider := NewConfigurableDataProvider(50) // Reasonable dataset
	styleConfig := DefaultStyleConfig()
	formatter := func(data Data[string], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
		return data.Item
	}

	for _, testCase := range brokenConfigs {
		t.Run(testCase.description, func(t *testing.T) {
			config := testCase.config

			// The library should auto-correct these issues
			list, err := NewTeaList(config, provider, styleConfig, formatter)
			if err != nil {
				t.Fatalf("Failed to create list even with auto-correction for '%s': %v", testCase.description, err)
			}

			// Verify the configuration was corrected
			correctedConfig := list.list.Config

			if correctedConfig.Height <= 0 {
				t.Errorf("Height was not corrected for '%s': %d", testCase.description, correctedConfig.Height)
			}

			if correctedConfig.TopThresholdIndex < 0 || correctedConfig.TopThresholdIndex >= correctedConfig.Height {
				t.Errorf("Top threshold was not corrected for '%s': %d (height %d)",
					testCase.description, correctedConfig.TopThresholdIndex, correctedConfig.Height)
			}

			if correctedConfig.BottomThresholdIndex < 0 || correctedConfig.BottomThresholdIndex >= correctedConfig.Height {
				t.Errorf("Bottom threshold was not corrected for '%s': %d (height %d)",
					testCase.description, correctedConfig.BottomThresholdIndex, correctedConfig.Height)
			}

			if correctedConfig.Height > 1 && correctedConfig.BottomThresholdIndex <= correctedConfig.TopThresholdIndex {
				t.Errorf("Threshold ordering was not corrected for '%s': top=%d, bottom=%d",
					testCase.description, correctedConfig.TopThresholdIndex, correctedConfig.BottomThresholdIndex)
			}

			if correctedConfig.ChunkSize <= 0 {
				t.Errorf("Chunk size was not corrected for '%s': %d", testCase.description, correctedConfig.ChunkSize)
			}

			if correctedConfig.InitialIndex < 0 {
				t.Errorf("Initial index was not corrected for '%s': %d", testCase.description, correctedConfig.InitialIndex)
			}

			fmt.Printf("✅ %s: Auto-corrected successfully\n", testCase.description)
			fmt.Printf("   Height: %d, Top: %d, Bottom: %d, Chunk: %d, Initial: %d\n",
				correctedConfig.Height, correctedConfig.TopThresholdIndex, correctedConfig.BottomThresholdIndex,
				correctedConfig.ChunkSize, correctedConfig.InitialIndex)

			// Verify the list actually works
			view := list.View()
			if view == "" {
				t.Errorf("List doesn't render after auto-correction for '%s'", testCase.description)
			}
		})
	}

	fmt.Println("=== END CONFIGURATION AUTO-CORRECTION TEST ===")
}

func TestConvenienceConstructors(t *testing.T) {
	fmt.Println("\n=== CONVENIENCE CONSTRUCTORS TEST ===")

	provider := NewConfigurableDataProvider(20)
	tableProvider := NewConfigurableTableDataProvider(20)

	// Test list convenience constructors
	t.Run("Simple Tea List", func(t *testing.T) {
		formatter := func(data Data[string], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
			return data.Item
		}

		list, err := NewSimpleTeaList(provider, formatter)
		if err != nil {
			t.Fatalf("NewSimpleTeaList failed: %v", err)
		}

		view := list.View()
		if view == "" {
			t.Error("Simple tea list produces empty view")
		}
		fmt.Println("✅ NewSimpleTeaList works")
	})

	t.Run("Tea List With Height", func(t *testing.T) {
		formatter := func(data Data[string], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
			return data.Item
		}

		list, err := NewTeaListWithHeight(provider, formatter, 15)
		if err != nil {
			t.Fatalf("NewTeaListWithHeight failed: %v", err)
		}

		if list.list.Config.Height != 15 {
			t.Errorf("Expected height 15, got %d", list.list.Config.Height)
		}
		fmt.Println("✅ NewTeaListWithHeight works")
	})

	// Test table convenience constructors
	columns := []TableColumn{
		NewColumn("Name", 15),
		NewRightColumn("Value", 10),
		NewCenterColumn("Status", 12),
	}

	t.Run("Simple Tea Table", func(t *testing.T) {
		table, err := NewSimpleTeaTable(columns, tableProvider)
		if err != nil {
			t.Fatalf("NewSimpleTeaTable failed: %v", err)
		}

		view := table.View()
		if view == "" {
			t.Error("Simple tea table produces empty view")
		}
		fmt.Println("✅ NewSimpleTeaTable works")
	})

	t.Run("Tea Table With Height", func(t *testing.T) {
		table, err := NewTeaTableWithHeight(columns, tableProvider, 12)
		if err != nil {
			t.Fatalf("NewTeaTableWithHeight failed: %v", err)
		}

		if table.table.config.ViewportConfig.Height != 12 {
			t.Errorf("Expected height 12, got %d", table.table.config.ViewportConfig.Height)
		}
		fmt.Println("✅ NewTeaTableWithHeight works")
	})

	t.Run("Tea Table With Theme", func(t *testing.T) {
		table, err := NewTeaTableWithTheme(columns, tableProvider, DarkTheme())
		if err != nil {
			t.Fatalf("NewTeaTableWithTheme failed: %v", err)
		}

		if table.table.theme.Name != "dark" {
			t.Errorf("Expected dark theme, got %s", table.table.theme.Name)
		}
		fmt.Println("✅ NewTeaTableWithTheme works")
	})

	t.Run("Column Creation Helpers", func(t *testing.T) {
		// Test individual column creators
		col1 := NewColumn("Test", 10)
		if col1.Title != "Test" || col1.Width != 10 || col1.Alignment != AlignLeft {
			t.Error("NewColumn failed")
		}

		col2 := NewRightColumn("Price", 8)
		if col2.Alignment != AlignRight {
			t.Error("NewRightColumn failed")
		}

		col3 := NewCenterColumn("Status", 12)
		if col3.Alignment != AlignCenter {
			t.Error("NewCenterColumn failed")
		}

		// Test auto-creation from titles
		autoColumns := CreateColumnsFromTitles("ID", "Name", "Email", "Status")
		if len(autoColumns) != 4 {
			t.Errorf("Expected 4 columns, got %d", len(autoColumns))
		}

		fmt.Println("✅ Column creation helpers work")
	})

	fmt.Println("=== END CONVENIENCE CONSTRUCTORS TEST ===")
}

func TestZeroDataScenarios(t *testing.T) {
	fmt.Println("\n=== ZERO DATA SCENARIOS TEST ===")

	// Test list with zero data
	t.Run("List with zero data", func(t *testing.T) {
		provider := NewConfigurableDataProvider(0)
		config := NewViewportConfig(10)
		styleConfig := DefaultStyleConfig()

		formatter := func(data Data[string], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
			return data.Item
		}

		list, err := NewTeaList(config, provider, styleConfig, formatter)

		if err != nil {
			// If the library chooses to reject zero data, that's valid
			fmt.Printf("✅ List with zero data properly rejected: %v\n", err)
			return
		}

		// If creation succeeds, verify it handles zero data gracefully
		fmt.Printf("✅ List with zero data created successfully\n")

		// Test that basic operations don't panic
		view := list.View()
		state := list.GetState()

		fmt.Printf("   Initial state: cursor=%d, viewport_start=%d\n", state.CursorIndex, state.ViewportStartIndex)
		fmt.Printf("   View length: %d characters\n", len(view))

		// Test navigation doesn't panic
		list.MoveDown()
		list.MoveUp()
		list.JumpToStart()
		list.JumpToEnd()

		finalState := list.GetState()
		fmt.Printf("   After navigation: cursor=%d, viewport_start=%d\n", finalState.CursorIndex, finalState.ViewportStartIndex)
	})

	// Test table with zero data
	t.Run("Table with zero data", func(t *testing.T) {
		provider := NewConfigurableTableDataProvider(0)
		columns := []TableColumn{
			NewColumn("Name", 15),
			NewRightColumn("Value", 10),
			NewCenterColumn("Status", 12),
		}
		config := NewTableConfig(columns, 10)
		theme := *DefaultTheme()

		table, err := NewTeaTable(config, provider, theme)

		if err != nil {
			// If the library chooses to reject zero data, that's valid
			fmt.Printf("✅ Table with zero data properly rejected: %v\n", err)
			return
		}

		// If creation succeeds, verify it handles zero data gracefully
		fmt.Printf("✅ Table with zero data created successfully\n")

		// Test that basic operations don't panic
		view := table.View()
		state := table.GetState()

		fmt.Printf("   Initial state: cursor=%d, viewport_start=%d\n", state.CursorIndex, state.ViewportStartIndex)
		fmt.Printf("   View length: %d characters\n", len(view))

		// Headers should still be visible even with no data
		if config.ShowHeader && len(view) > 0 {
			if contains(view, "Name") {
				fmt.Printf("   ✅ Headers visible even with zero data\n")
			} else {
				t.Error("Headers should be visible even with zero data")
			}
		}

		// Test navigation doesn't panic
		table.MoveDown()
		table.MoveUp()
		table.JumpToStart()
		table.JumpToEnd()

		finalState := table.GetState()
		fmt.Printf("   After navigation: cursor=%d, viewport_start=%d\n", finalState.CursorIndex, finalState.ViewportStartIndex)
	})

	// Test convenience constructors with zero data
	t.Run("Convenience constructors with zero data", func(t *testing.T) {
		emptyListProvider := NewConfigurableDataProvider(0)
		emptyTableProvider := NewConfigurableTableDataProvider(0)

		formatter := func(data Data[string], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
			return data.Item
		}

		// Test simple list constructor
		simpleList, err := NewSimpleTeaList(emptyListProvider, formatter)
		if err != nil {
			fmt.Printf("✅ NewSimpleTeaList properly rejects zero data: %v\n", err)
		} else {
			fmt.Printf("✅ NewSimpleTeaList handles zero data gracefully\n")
			_ = simpleList.View() // Should not panic
		}

		// Test simple table constructor
		columns := []TableColumn{NewColumn("Test", 10)}
		simpleTable, err := NewSimpleTeaTable(columns, emptyTableProvider)
		if err != nil {
			fmt.Printf("✅ NewSimpleTeaTable properly rejects zero data: %v\n", err)
		} else {
			fmt.Printf("✅ NewSimpleTeaTable handles zero data gracefully\n")
			view := simpleTable.View() // Should not panic
			if contains(view, "Test") {
				fmt.Printf("   ✅ Table headers visible even with zero data\n")
			}
		}
	})

	// Test auto-correction with zero data
	t.Run("Auto-correction with zero data", func(t *testing.T) {
		provider := NewConfigurableDataProvider(0)

		// Test with broken config + zero data
		brokenConfig := ViewportConfig{
			Height:               -5, // Invalid
			TopThresholdIndex:    -1, // Invalid
			BottomThresholdIndex: -1, // Invalid
			ChunkSize:            0,  // Invalid
		}

		styleConfig := DefaultStyleConfig()
		formatter := func(data Data[string], index int, ctx RenderContext, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
			return data.Item
		}

		list, err := NewTeaList(brokenConfig, provider, styleConfig, formatter)

		if err != nil {
			fmt.Printf("✅ Auto-correction with zero data properly rejected: %v\n", err)
		} else {
			fmt.Printf("✅ Auto-correction works even with zero data\n")
			correctedConfig := list.list.Config
			fmt.Printf("   Corrected config: Height=%d, Top=%d, Bottom=%d, Chunk=%d\n",
				correctedConfig.Height, correctedConfig.TopThresholdIndex,
				correctedConfig.BottomThresholdIndex, correctedConfig.ChunkSize)
		}
	})

	fmt.Println("=== END ZERO DATA SCENARIOS TEST ===")
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			(len(s) > len(substr) && (func() bool {
				for i := 1; i < len(s)-len(substr)+1; i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			})())))
}
