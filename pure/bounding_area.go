package vtable

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// ================================
// BOUNDING AREA COMPONENT
// ================================

// BoundingAreaManager manages chunk loading/unloading based on viewport position
type BoundingAreaManager struct {
	config       BoundingAreaConfig
	loadedChunks map[int]bool // Track loaded chunk indices

	// Callbacks for chunk operations
	onChunkLoad   func(startIndex, count int) tea.Cmd
	onChunkUnload func(startIndex, count int) tea.Cmd
}

// BoundingAreaConfig configures the bounding area behavior
type BoundingAreaConfig struct {
	ChunkSize           int
	ChunksBefore        int
	ChunksAfter         int
	MaxLoadedChunks     int  // Optional limit on total loaded chunks
	UnloadDistantChunks bool // Whether to unload chunks far from viewport
}

// ================================
// BOUNDING AREA MESSAGES
// ================================

// BoundingAreaUpdateMsg triggers bounding area recalculation
type BoundingAreaUpdateMsg struct {
	ViewportState ViewportState
	TotalItems    int
}

// ChunkLoadRequestMsg requests loading of a specific chunk
type ChunkLoadRequestMsg struct {
	StartIndex int
	Count      int
}

// ChunkUnloadRequestMsg requests unloading of a specific chunk
type ChunkUnloadRequestMsg struct {
	StartIndex int
	Count      int
}

// BoundingAreaChangedMsg notifies when bounding area changes
type BoundingAreaChangedMsg struct {
	BoundingArea   BoundingArea
	LoadRequests   []ChunkLoadRequestMsg
	UnloadRequests []ChunkUnloadRequestMsg
}

// ================================
// CONSTRUCTOR
// ================================

// NewBoundingAreaManager creates a new bounding area manager
func NewBoundingAreaManager(config BoundingAreaConfig) *BoundingAreaManager {
	// Set defaults
	if config.ChunkSize <= 0 {
		config.ChunkSize = 20
	}
	if config.ChunksBefore == 0 {
		config.ChunksBefore = 1
	}
	if config.ChunksAfter == 0 {
		config.ChunksAfter = 2
	}
	if config.MaxLoadedChunks == 0 {
		config.MaxLoadedChunks = 10 // Default limit
	}

	return &BoundingAreaManager{
		config:       config,
		loadedChunks: make(map[int]bool),
	}
}

// SetCallbacks sets chunk operation callbacks
func (b *BoundingAreaManager) SetCallbacks(
	onChunkLoad func(startIndex, count int) tea.Cmd,
	onChunkUnload func(startIndex, count int) tea.Cmd,
) {
	b.onChunkLoad = onChunkLoad
	b.onChunkUnload = onChunkUnload
}

// ================================
// BUBBLE TEA INTERFACE
// ================================

// Init initializes the bounding area manager
func (b *BoundingAreaManager) Init() tea.Cmd {
	return nil
}

// Update handles bounding area messages
func (b *BoundingAreaManager) Update(msg tea.Msg) (*BoundingAreaManager, tea.Cmd) {
	switch msg := msg.(type) {
	case BoundingAreaUpdateMsg:
		return b.handleUpdate(msg)
	case ChunkLoadRequestMsg:
		return b.handleChunkLoad(msg)
	case ChunkUnloadRequestMsg:
		return b.handleChunkUnload(msg)
	}

	return b, nil
}

// View returns debug information about loaded chunks
func (b *BoundingAreaManager) View() string {
	if len(b.loadedChunks) == 0 {
		return "BoundingArea: No chunks loaded"
	}

	return "BoundingArea: " + b.getLoadedChunksDebugString()
}

// ================================
// PUBLIC INTERFACE
// ================================

// CalculateBoundingArea calculates the bounding area for a viewport state
func (b *BoundingAreaManager) CalculateBoundingArea(viewport ViewportState, totalItems int) BoundingArea {
	return calculateBoundingArea(viewport, totalItems, b.config)
}

// GetLoadedChunks returns the list of currently loaded chunk indices
func (b *BoundingAreaManager) GetLoadedChunks() []int {
	chunks := make([]int, 0, len(b.loadedChunks))
	for chunkIndex := range b.loadedChunks {
		chunks = append(chunks, chunkIndex)
	}
	return chunks
}

// IsChunkLoaded checks if a specific chunk is loaded
func (b *BoundingAreaManager) IsChunkLoaded(chunkIndex int) bool {
	return b.loadedChunks[chunkIndex]
}

// ================================
// MESSAGE HANDLERS
// ================================

// handleUpdate processes viewport state changes and manages chunks
func (b *BoundingAreaManager) handleUpdate(msg BoundingAreaUpdateMsg) (*BoundingAreaManager, tea.Cmd) {
	// Calculate new bounding area
	boundingArea := b.CalculateBoundingArea(msg.ViewportState, msg.TotalItems)

	// Determine which chunks need to be loaded/unloaded
	loadRequests, unloadRequests := b.calculateChunkOperations(boundingArea, msg.TotalItems)

	// Execute chunk operations
	var cmds []tea.Cmd

	// Process load requests
	for _, req := range loadRequests {
		chunkIndex := req.StartIndex / b.config.ChunkSize
		b.loadedChunks[chunkIndex] = true

		if b.onChunkLoad != nil {
			cmds = append(cmds, b.onChunkLoad(req.StartIndex, req.Count))
		}
	}

	// Process unload requests
	for _, req := range unloadRequests {
		chunkIndex := req.StartIndex / b.config.ChunkSize
		delete(b.loadedChunks, chunkIndex)

		if b.onChunkUnload != nil {
			cmds = append(cmds, b.onChunkUnload(req.StartIndex, req.Count))
		}
	}

	// Emit bounding area changed message
	if len(loadRequests) > 0 || len(unloadRequests) > 0 {
		cmds = append(cmds, BoundingAreaChangedCmd(BoundingAreaChangedMsg{
			BoundingArea:   boundingArea,
			LoadRequests:   loadRequests,
			UnloadRequests: unloadRequests,
		}))
	}

	return b, tea.Batch(cmds...)
}

// handleChunkLoad processes chunk load requests
func (b *BoundingAreaManager) handleChunkLoad(msg ChunkLoadRequestMsg) (*BoundingAreaManager, tea.Cmd) {
	chunkIndex := msg.StartIndex / b.config.ChunkSize
	b.loadedChunks[chunkIndex] = true

	if b.onChunkLoad != nil {
		return b, b.onChunkLoad(msg.StartIndex, msg.Count)
	}

	return b, nil
}

// handleChunkUnload processes chunk unload requests
func (b *BoundingAreaManager) handleChunkUnload(msg ChunkUnloadRequestMsg) (*BoundingAreaManager, tea.Cmd) {
	chunkIndex := msg.StartIndex / b.config.ChunkSize
	delete(b.loadedChunks, chunkIndex)

	if b.onChunkUnload != nil {
		return b, b.onChunkUnload(msg.StartIndex, msg.Count)
	}

	return b, nil
}

// ================================
// HELPER METHODS
// ================================

// calculateChunkOperations determines which chunks to load/unload
func (b *BoundingAreaManager) calculateChunkOperations(boundingArea BoundingArea, totalItems int) ([]ChunkLoadRequestMsg, []ChunkUnloadRequestMsg) {
	var loadRequests []ChunkLoadRequestMsg
	var unloadRequests []ChunkUnloadRequestMsg

	// Calculate required chunks for bounding area
	requiredChunks := b.getRequiredChunks(boundingArea, totalItems)

	// Find chunks to load (required but not loaded)
	for _, chunkIndex := range requiredChunks {
		if !b.loadedChunks[chunkIndex] {
			startIndex := chunkIndex * b.config.ChunkSize
			count := b.config.ChunkSize

			// Adjust count for last chunk
			if startIndex+count > totalItems {
				count = totalItems - startIndex
			}

			if count > 0 {
				loadRequests = append(loadRequests, ChunkLoadRequestMsg{
					StartIndex: startIndex,
					Count:      count,
				})
			}
		}
	}

	// Find chunks to unload (loaded but not required)
	if b.config.UnloadDistantChunks {
		requiredChunkMap := make(map[int]bool)
		for _, chunkIndex := range requiredChunks {
			requiredChunkMap[chunkIndex] = true
		}

		for chunkIndex := range b.loadedChunks {
			if !requiredChunkMap[chunkIndex] {
				startIndex := chunkIndex * b.config.ChunkSize
				count := b.config.ChunkSize

				unloadRequests = append(unloadRequests, ChunkUnloadRequestMsg{
					StartIndex: startIndex,
					Count:      count,
				})
			}
		}
	}

	return loadRequests, unloadRequests
}

// getRequiredChunks returns chunk indices needed for the bounding area
func (b *BoundingAreaManager) getRequiredChunks(boundingArea BoundingArea, totalItems int) []int {
	if totalItems == 0 {
		return nil
	}

	startChunk := boundingArea.StartIndex / b.config.ChunkSize
	endChunk := boundingArea.EndIndex / b.config.ChunkSize

	// Ensure we don't go beyond available data
	maxChunk := (totalItems - 1) / b.config.ChunkSize
	if endChunk > maxChunk {
		endChunk = maxChunk
	}

	chunks := make([]int, 0, endChunk-startChunk+1)
	for i := startChunk; i <= endChunk; i++ {
		chunks = append(chunks, i)
	}

	return chunks
}

// getLoadedChunksDebugString returns debug info about loaded chunks
func (b *BoundingAreaManager) getLoadedChunksDebugString() string {
	if len(b.loadedChunks) == 0 {
		return "none"
	}

	chunks := b.GetLoadedChunks()
	return fmt.Sprintf("%d chunks loaded", len(chunks))
}

// ================================
// UTILITY FUNCTIONS
// ================================

// calculateBoundingArea calculates the area around viewport where chunks should be loaded
func calculateBoundingArea(viewport ViewportState, totalItems int, config BoundingAreaConfig) BoundingArea {
	if totalItems == 0 {
		return BoundingArea{StartIndex: 0, EndIndex: 0, ChunkStart: 0, ChunkEnd: 0}
	}

	chunkSize := config.ChunkSize

	// Calculate viewport chunk boundaries
	viewportStart := viewport.ViewportStartIndex
	viewportEnd := viewport.ViewportStartIndex + chunkSize // Use config height equivalent
	if viewportEnd > totalItems {
		viewportEnd = totalItems
	}

	// Expand area by configured chunks before/after
	areaStart := viewportStart - (config.ChunksBefore * chunkSize)
	if areaStart < 0 {
		areaStart = 0
	}

	areaEnd := viewportEnd + (config.ChunksAfter * chunkSize)
	if areaEnd > totalItems {
		areaEnd = totalItems
	}

	// Calculate chunk boundaries
	chunkStart := (areaStart / chunkSize) * chunkSize
	chunkEnd := ((areaEnd / chunkSize) + 1) * chunkSize

	return BoundingArea{
		StartIndex: areaStart,
		EndIndex:   areaEnd,
		ChunkStart: chunkStart,
		ChunkEnd:   chunkEnd,
	}
}

// ================================
// COMMAND HELPERS
// ================================

// BoundingAreaUpdateCmd creates update command
func BoundingAreaUpdateCmd(viewport ViewportState, totalItems int) tea.Cmd {
	return func() tea.Msg {
		return BoundingAreaUpdateMsg{
			ViewportState: viewport,
			TotalItems:    totalItems,
		}
	}
}

// ChunkLoadRequestCmd creates chunk load command
func ChunkLoadRequestCmd(startIndex, count int) tea.Cmd {
	return func() tea.Msg {
		return ChunkLoadRequestMsg{
			StartIndex: startIndex,
			Count:      count,
		}
	}
}

// ChunkUnloadRequestCmd creates chunk unload command
func ChunkUnloadRequestCmd(startIndex, count int) tea.Cmd {
	return func() tea.Msg {
		return ChunkUnloadRequestMsg{
			StartIndex: startIndex,
			Count:      count,
		}
	}
}

// BoundingAreaChangedCmd creates bounding area changed command
func BoundingAreaChangedCmd(msg BoundingAreaChangedMsg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
