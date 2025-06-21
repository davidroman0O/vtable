// Package data provides the core data handling capabilities for the vtable component.
// It includes functionalities for managing data requests, chunking, sorting, and caching,
// forming the backbone of the data virtualization layer. This package is designed to
// efficiently handle large datasets by loading data in manageable chunks, only when needed.
package data

import (
	"github.com/davidroman0O/vtable/core"
)

// GetSelectionCount iterates through all loaded chunks and counts the number of
// items that are marked as selected. This provides a snapshot of the current
// selection state based on the data available in memory. The definitive source of
// selection state is the DataSource, but this function offers a quick way to get
// a count from the client-side cache.
func GetSelectionCount[T any](chunks map[int]core.Chunk[T]) int {
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
