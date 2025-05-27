package vtable

// ================================
// SELECTION QUERY FUNCTIONS
// ================================

// GetSelectionCount returns the number of selected items from chunks
func GetSelectionCount[T any](chunks map[int]Chunk[T]) int {
	count := 0
	// Read selection state from chunks (DataSource owns the state)
	for _, chunk := range chunks {
		for _, item := range chunk.Items {
			if item.Selected {
				count++
			}
		}
	}
	return count
}
