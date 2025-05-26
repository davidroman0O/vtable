package viewport_test

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	vtable "github.com/davidroman0O/vtable/pure"
)

// TestSeparatedComponentsDemo demonstrates the new architecture with separated components
func TestSeparatedComponentsDemo(t *testing.T) {
	fmt.Println("=== Separated Components Architecture Demo ===")

	// 1. Create a viewport component
	viewportConfig := vtable.ViewportConfig{
		Height:             5,
		TopThreshold:       1, // 1 position from viewport start
		BottomThreshold:    3, // 3 positions from viewport end (position 1 in height-5 viewport)
		ChunkSize:          10,
		InitialIndex:       0,
		BoundingAreaBefore: 1,
		BoundingAreaAfter:  2,
	}

	viewport := vtable.NewViewport(viewportConfig)
	viewport.SetDebugMode(true)

	// 2. Create a bounding area manager
	boundingConfig := vtable.BoundingAreaConfig{
		ChunkSize:           10,
		ChunksBefore:        1,
		ChunksAfter:         2,
		MaxLoadedChunks:     10,
		UnloadDistantChunks: true,
	}

	boundingManager := vtable.NewBoundingAreaManager(boundingConfig)

	// Set up chunk loading callbacks
	boundingManager.SetCallbacks(
		func(startIndex, count int) tea.Cmd {
			fmt.Printf("  📦 Loading chunk: start=%d, count=%d\n", startIndex, count)
			return nil
		},
		func(startIndex, count int) tea.Cmd {
			fmt.Printf("  🗑️  Unloading chunk: start=%d, count=%d\n", startIndex, count)
			return nil
		},
	)

	// 3. Simulate dataset with 100 items
	totalItems := 100

	// Initialize viewport with total items
	viewport.Update(vtable.ViewportDataChangedMsg{TotalItems: totalItems})

	fmt.Printf("Initial State:\n")
	fmt.Printf("  Viewport: %s\n", viewport.View())
	fmt.Printf("  Bounding: %s\n", boundingManager.View())

	// 4. Demonstrate navigation and chunk management
	fmt.Printf("\n--- Navigation Demo ---\n")

	// Move down several times
	for i := 0; i < 8; i++ {
		fmt.Printf("\nStep %d: Moving down\n", i+1)

		// Update viewport
		viewport, _ = viewport.Update(vtable.ViewportDownMsg{})
		state := viewport.GetState()

		// Update bounding area
		boundingManager, _ = boundingManager.Update(vtable.BoundingAreaUpdateMsg{
			ViewportState: state,
			TotalItems:    totalItems,
		})

		fmt.Printf("  Viewport: %s\n", viewport.View())
		fmt.Printf("  Loaded chunks: %v\n", boundingManager.GetLoadedChunks())
	}

	// 5. Demonstrate type-safe navigation commands
	fmt.Printf("\n--- Type-Safe Commands Demo ---\n")

	// Page down
	fmt.Printf("\nPage Down:\n")
	viewport, _ = viewport.Update(vtable.ViewportPageDownMsg{})
	state := viewport.GetState()
	boundingManager, _ = boundingManager.Update(vtable.BoundingAreaUpdateMsg{
		ViewportState: state,
		TotalItems:    totalItems,
	})
	fmt.Printf("  Viewport: %s\n", viewport.View())

	// Jump to specific index
	fmt.Printf("\nJump to index 75:\n")
	viewport, _ = viewport.Update(vtable.ViewportJumpMsg{Index: 75})
	state = viewport.GetState()
	boundingManager, _ = boundingManager.Update(vtable.BoundingAreaUpdateMsg{
		ViewportState: state,
		TotalItems:    totalItems,
	})
	fmt.Printf("  Viewport: %s\n", viewport.View())
	fmt.Printf("  Loaded chunks: %v\n", boundingManager.GetLoadedChunks())

	// Jump to end
	fmt.Printf("\nJump to end:\n")
	viewport, _ = viewport.Update(vtable.ViewportEndMsg{})
	state = viewport.GetState()
	boundingManager, _ = boundingManager.Update(vtable.BoundingAreaUpdateMsg{
		ViewportState: state,
		TotalItems:    totalItems,
	})
	fmt.Printf("  Viewport: %s\n", viewport.View())

	// 6. Demonstrate viewport resize
	fmt.Printf("\n--- Resize Demo ---\n")
	fmt.Printf("\nResizing viewport to height 3:\n")
	viewport, _ = viewport.Update(vtable.ViewportResizedMsg{Height: 3})
	state = viewport.GetState()
	fmt.Printf("  Viewport: %s\n", viewport.View())

	fmt.Printf("\n=== Demo Complete ===")
}

// TestNavigationTypeConstants tests the type safety of navigation types
func TestNavigationTypeConstants(t *testing.T) {
	fmt.Println("\n=== Navigation Type Safety Demo ===")

	// Test all navigation types
	navTypes := []vtable.NavigationType{
		vtable.NavigationUp,
		vtable.NavigationDown,
		vtable.NavigationPageUp,
		vtable.NavigationPageDown,
		vtable.NavigationStart,
		vtable.NavigationEnd,
		vtable.NavigationJump,
	}

	for _, navType := range navTypes {
		fmt.Printf("Navigation type: %s\n", navType.String())

		// Test command creation
		cmd := vtable.ViewportNavigationCmd(navType, 0)
		if cmd == nil {
			t.Errorf("Failed to create command for navigation type %s", navType.String())
		}
	}

	// Test specific command constructors
	commands := map[string]tea.Cmd{
		"Up":       vtable.ViewportUpCmd(),
		"Down":     vtable.ViewportDownCmd(),
		"PageUp":   vtable.ViewportPageUpCmd(),
		"PageDown": vtable.ViewportPageDownCmd(),
		"Start":    vtable.ViewportStartCmd(),
		"End":      vtable.ViewportEndCmd(),
		"Jump":     vtable.ViewportJumpCmd(42),
	}

	for name, cmd := range commands {
		if cmd == nil {
			t.Errorf("Failed to create %s command", name)
		} else {
			fmt.Printf("✓ %s command created successfully\n", name)
		}
	}

	fmt.Printf("=== Type Safety Demo Complete ===")
}

// TestBoundingAreaCalculation tests the bounding area calculation logic
func TestBoundingAreaCalculation(t *testing.T) {
	fmt.Println("\n=== Bounding Area Calculation Demo ===")

	config := vtable.BoundingAreaConfig{
		ChunkSize:    10,
		ChunksBefore: 1,
		ChunksAfter:  2,
	}

	// Test various viewport positions
	testCases := []struct {
		name       string
		viewport   vtable.ViewportState
		totalItems int
	}{
		{
			name: "Start of dataset",
			viewport: vtable.ViewportState{
				ViewportStartIndex:  0,
				CursorIndex:         2,
				CursorViewportIndex: 2,
			},
			totalItems: 100,
		},
		{
			name: "Middle of dataset",
			viewport: vtable.ViewportState{
				ViewportStartIndex:  45,
				CursorIndex:         47,
				CursorViewportIndex: 2,
			},
			totalItems: 100,
		},
		{
			name: "End of dataset",
			viewport: vtable.ViewportState{
				ViewportStartIndex:  95,
				CursorIndex:         99,
				CursorViewportIndex: 4,
			},
			totalItems: 100,
		},
	}

	for _, tc := range testCases {
		fmt.Printf("\nTest case: %s\n", tc.name)
		fmt.Printf("  Viewport: start=%d, cursor=%d\n", tc.viewport.ViewportStartIndex, tc.viewport.CursorIndex)

		// Note: calculateBoundingArea is not exported, so we'll use the BoundingAreaManager
		manager := vtable.NewBoundingAreaManager(config)
		boundingArea := manager.CalculateBoundingArea(tc.viewport, tc.totalItems)
		fmt.Printf("  Bounding area: start=%d, end=%d, chunkStart=%d, chunkEnd=%d\n",
			boundingArea.StartIndex, boundingArea.EndIndex,
			boundingArea.ChunkStart, boundingArea.ChunkEnd)
	}

	fmt.Printf("=== Bounding Area Demo Complete ===")
}
