package vtable

import (
	"testing"
)

func TestCalculateThresholds(t *testing.T) {
	tests := []struct {
		height                      int
		expectedTop, expectedBottom int
		description                 string
	}{
		{1, 0, 0, "single row"},
		{2, 0, 1, "two rows"},
		{3, 0, 2, "three rows"},
		{5, 1, 3, "five rows"},
		{10, 2, 7, "ten rows"},      // 10/5=2, 10-1-2=7
		{15, 3, 11, "fifteen rows"}, // 15/5=3, 15-1-3=11
		{20, 4, 15, "twenty rows"},  // 20/5=4, 20-1-4=15
	}

	for _, test := range tests {
		top, bottom := CalculateThresholds(test.height)
		if top != test.expectedTop || bottom != test.expectedBottom {
			t.Errorf("%s: height=%d, expected top=%d bottom=%d, got top=%d bottom=%d",
				test.description, test.height, test.expectedTop, test.expectedBottom, top, bottom)
		}

		// Verify the calculated thresholds are valid
		if top < 0 || top >= test.height {
			t.Errorf("%s: top threshold %d is out of bounds for height %d", test.description, top, test.height)
		}
		if bottom < 0 || bottom >= test.height {
			t.Errorf("%s: bottom threshold %d is out of bounds for height %d", test.description, bottom, test.height)
		}

		// Special case: for height=1, thresholds can be equal
		if test.height > 1 && bottom <= top {
			t.Errorf("%s: bottom threshold %d must be greater than top threshold %d", test.description, bottom, top)
		}
	}
}

func TestValidateAndFixViewportConfig(t *testing.T) {
	t.Run("auto-fix invalid height", func(t *testing.T) {
		config := ViewportConfig{Height: -5}
		ValidateAndFixViewportConfig(&config)

		if config.Height <= 0 {
			t.Errorf("Expected height to be fixed to positive value, got %d", config.Height)
		}
	})

	t.Run("auto-fix invalid thresholds", func(t *testing.T) {
		config := ViewportConfig{
			Height:               10,
			TopThresholdIndex:    15, // Invalid: > height
			BottomThresholdIndex: 2,  // Invalid: < top
		}
		ValidateAndFixViewportConfig(&config)

		if config.TopThresholdIndex < 0 || config.TopThresholdIndex >= config.Height {
			t.Errorf("Top threshold %d is still invalid for height %d", config.TopThresholdIndex, config.Height)
		}
		if config.BottomThresholdIndex < 0 || config.BottomThresholdIndex >= config.Height {
			t.Errorf("Bottom threshold %d is still invalid for height %d", config.BottomThresholdIndex, config.Height)
		}
		if config.BottomThresholdIndex <= config.TopThresholdIndex {
			t.Errorf("Bottom threshold %d must be greater than top threshold %d",
				config.BottomThresholdIndex, config.TopThresholdIndex)
		}
	})

	t.Run("auto-fix invalid chunk size", func(t *testing.T) {
		config := ViewportConfig{Height: 10, ChunkSize: -5}
		ValidateAndFixViewportConfig(&config)

		if config.ChunkSize <= 0 {
			t.Errorf("Expected chunk size to be fixed to positive value, got %d", config.ChunkSize)
		}
	})

	t.Run("auto-fix negative initial index", func(t *testing.T) {
		config := ViewportConfig{Height: 10, InitialIndex: -3}
		ValidateAndFixViewportConfig(&config)

		if config.InitialIndex < 0 {
			t.Errorf("Expected initial index to be fixed to non-negative value, got %d", config.InitialIndex)
		}
	})
}

func TestNewViewportConfig(t *testing.T) {
	config := NewViewportConfig(15)

	if config.Height != 15 {
		t.Errorf("Expected height 15, got %d", config.Height)
	}

	// Verify thresholds are valid
	if config.TopThresholdIndex < 0 || config.TopThresholdIndex >= config.Height {
		t.Errorf("Invalid top threshold %d for height %d", config.TopThresholdIndex, config.Height)
	}
	if config.BottomThresholdIndex < 0 || config.BottomThresholdIndex >= config.Height {
		t.Errorf("Invalid bottom threshold %d for height %d", config.BottomThresholdIndex, config.Height)
	}
	if config.BottomThresholdIndex <= config.TopThresholdIndex {
		t.Errorf("Bottom threshold %d must be greater than top threshold %d",
			config.BottomThresholdIndex, config.TopThresholdIndex)
	}

	if config.ChunkSize <= 0 {
		t.Errorf("Expected positive chunk size, got %d", config.ChunkSize)
	}
}

func TestColumnHelpers(t *testing.T) {
	t.Run("NewColumn", func(t *testing.T) {
		col := NewColumn("Test", 20)
		if col.Title != "Test" || col.Width != 20 || col.Alignment != AlignLeft {
			t.Errorf("NewColumn failed: %+v", col)
		}
	})

	t.Run("NewRightColumn", func(t *testing.T) {
		col := NewRightColumn("Price", 10)
		if col.Title != "Price" || col.Width != 10 || col.Alignment != AlignRight {
			t.Errorf("NewRightColumn failed: %+v", col)
		}
	})

	t.Run("CreateColumnsFromTitles", func(t *testing.T) {
		columns := CreateColumnsFromTitles("Name", "Age", "City")
		if len(columns) != 3 {
			t.Errorf("Expected 3 columns, got %d", len(columns))
		}

		for i, col := range columns {
			if col.Width < 8 {
				t.Errorf("Column %d width %d is too small", i, col.Width)
			}
		}
	})
}

// TestConfigurationErrorPrevention demonstrates that the old problematic configuration
// now works thanks to auto-correction
func TestConfigurationErrorPrevention(t *testing.T) {
	// This would have failed before with "top threshold must be within viewport bounds"
	config := TableConfig{
		Columns: []TableColumn{
			{Title: "Name", Width: 40, Alignment: AlignLeft},
			{Title: "Size", Width: 12, Alignment: AlignRight},
			{Title: "Location", Width: 30, Alignment: AlignLeft},
		},
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: ViewportConfig{
			Height:               10,
			TopThresholdIndex:    1, // Previously could cause issues
			BottomThresholdIndex: 7, // Previously could cause issues
			ChunkSize:            100,
		},
	}

	// This should not return an error anymore
	err := ValidateAndFixTableConfig(&config)
	if err != nil {
		t.Errorf("Configuration should be auto-fixed, but got error: %v", err)
	}

	// Verify the configuration is now valid
	if config.ViewportConfig.TopThresholdIndex < 0 ||
		config.ViewportConfig.TopThresholdIndex >= config.ViewportConfig.Height {
		t.Errorf("Top threshold is still invalid after auto-fix")
	}

	if config.ViewportConfig.BottomThresholdIndex < 0 ||
		config.ViewportConfig.BottomThresholdIndex >= config.ViewportConfig.Height {
		t.Errorf("Bottom threshold is still invalid after auto-fix")
	}

	if config.ViewportConfig.BottomThresholdIndex <= config.ViewportConfig.TopThresholdIndex {
		t.Errorf("Bottom threshold must be greater than top threshold after auto-fix")
	}
}
