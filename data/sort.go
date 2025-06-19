// Package data provides the core data handling capabilities for the vtable component.
// It includes functionalities for managing data requests, chunking, sorting, and caching,
// forming the backbone of the data virtualization layer. This package is designed to
// efficiently handle large datasets by loading data in manageable chunks, only when needed.
package data

// SortState represents the current sorting configuration of a component.
// It maintains an ordered list of fields to sort by and their corresponding
// directions (e.g., "asc" or "desc").
type SortState struct {
	Fields     []string // The fields to sort by, in order of priority.
	Directions []string // The corresponding sort directions ("asc" or "desc").
}

// ToggleSortField cycles through sorting states for a given field: from "asc"
// to "desc", and then removes it from the sort configuration. If the field is
// not currently part of the sort, it is added with an "asc" direction. This
// function provides a user-friendly way to interact with column sorting in a UI.
func ToggleSortField(currentSort SortState, field string) SortState {
	// Find field in current sort
	for i, sortField := range currentSort.Fields {
		if sortField == field {
			// Toggle direction
			if currentSort.Directions[i] == "asc" {
				currentSort.Directions[i] = "desc"
				return currentSort
			} else {
				// Remove field from sort (desc -> remove)
				newFields := make([]string, 0, len(currentSort.Fields)-1)
				newDirections := make([]string, 0, len(currentSort.Directions)-1)

				for j, f := range currentSort.Fields {
					if j != i {
						newFields = append(newFields, f)
						newDirections = append(newDirections, currentSort.Directions[j])
					}
				}

				return SortState{
					Fields:     newFields,
					Directions: newDirections,
				}
			}
		}
	}

	// Field not found, add it with ascending direction
	return SortState{
		Fields:     append(currentSort.Fields, field),
		Directions: append(currentSort.Directions, "asc"),
	}
}

// SetSortField creates a new SortState that sorts by a single field in a
// specified direction, discarding any previous sort configuration.
func SetSortField(field, direction string) SortState {
	return SortState{
		Fields:     []string{field},
		Directions: []string{direction},
	}
}

// AddSortField adds a new field to the current sort configuration. If the field
// already exists in the sort, it is removed and then re-added at the end to
// give it the highest priority in a multi-level sort.
func AddSortField(currentSort SortState, field, direction string) SortState {
	// Remove field if it already exists
	newFields := make([]string, 0, len(currentSort.Fields))
	newDirections := make([]string, 0, len(currentSort.Directions))

	for i, sortField := range currentSort.Fields {
		if sortField != field {
			newFields = append(newFields, sortField)
			newDirections = append(newDirections, currentSort.Directions[i])
		}
	}

	// Add to end
	newFields = append(newFields, field)
	newDirections = append(newDirections, direction)

	return SortState{
		Fields:     newFields,
		Directions: newDirections,
	}
}

// RemoveSortField removes a field from the current sort configuration. If the
// field is not part of the sort, the state is returned unchanged.
func RemoveSortField(currentSort SortState, field string) SortState {
	newFields := make([]string, 0, len(currentSort.Fields))
	newDirections := make([]string, 0, len(currentSort.Directions))

	for i, sortField := range currentSort.Fields {
		if sortField != field {
			newFields = append(newFields, sortField)
			newDirections = append(newDirections, currentSort.Directions[i])
		}
	}

	return SortState{
		Fields:     newFields,
		Directions: newDirections,
	}
}

// ClearSort returns an empty SortState, effectively removing all sorting.
func ClearSort() SortState {
	return SortState{
		Fields:     nil,
		Directions: nil,
	}
}

// GetSortDirection returns the current sort direction ("asc" or "desc") for a
// given field. If the field is not being sorted, it returns an empty string.
func GetSortDirection(currentSort SortState, field string) string {
	for i, sortField := range currentSort.Fields {
		if sortField == field {
			return currentSort.Directions[i]
		}
	}
	return ""
}

// IsSorted checks if a given field is currently part of the sort configuration,
// regardless of its direction.
func IsSorted(currentSort SortState, field string) bool {
	for _, sortField := range currentSort.Fields {
		if sortField == field {
			return true
		}
	}
	return false
}

// GetSortPriority returns the priority of a field in the current sort
// configuration. The priority is its 0-based index in the `Fields` slice. A lower
// number indicates a higher priority. If the field is not being sorted, it returns -1.
func GetSortPriority(currentSort SortState, field string) int {
	for i, sortField := range currentSort.Fields {
		if sortField == field {
			return i
		}
	}
	return -1
}
