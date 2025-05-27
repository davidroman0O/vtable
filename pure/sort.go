package vtable

// ================================
// SORT MANAGEMENT FUNCTIONS
// ================================

// SortState represents the current sorting configuration
type SortState struct {
	Fields     []string
	Directions []string
}

// ToggleSortField toggles sorting on a field (asc -> desc -> remove)
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

// SetSortField sets sorting to a single field with specified direction
func SetSortField(field, direction string) SortState {
	return SortState{
		Fields:     []string{field},
		Directions: []string{direction},
	}
}

// AddSortField adds a sort field (removes if already exists, then adds to end)
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

// RemoveSortField removes a sort field
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

// ClearSort clears all sorting
func ClearSort() SortState {
	return SortState{
		Fields:     nil,
		Directions: nil,
	}
}

// GetSortDirection returns the direction for a field, or empty string if not sorted
func GetSortDirection(currentSort SortState, field string) string {
	for i, sortField := range currentSort.Fields {
		if sortField == field {
			return currentSort.Directions[i]
		}
	}
	return ""
}

// IsSorted returns true if the field is currently being sorted
func IsSorted(currentSort SortState, field string) bool {
	for _, sortField := range currentSort.Fields {
		if sortField == field {
			return true
		}
	}
	return false
}

// GetSortPriority returns the priority (0-based index) of a field in the sort order, or -1 if not sorted
func GetSortPriority(currentSort SortState, field string) int {
	for i, sortField := range currentSort.Fields {
		if sortField == field {
			return i
		}
	}
	return -1
}
