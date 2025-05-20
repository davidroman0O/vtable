package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davidroman0O/vtable"
)

// Person represents a data item
type Person struct {
	ID        int
	FirstName string
	LastName  string
	Age       int
	City      string
}

// PersonDataProvider implements the DataProvider interface for Person data
type PersonDataProvider struct {
	data           []Person
	filteredData   []Person
	sortFields     []string
	sortDirections []string
	filters        map[string]any
	dirty          bool
}

// NewPersonDataProvider creates a new data provider with sample data
func NewPersonDataProvider() *PersonDataProvider {
	return &PersonDataProvider{
		data: []Person{
			{ID: 1, FirstName: "John", LastName: "Smith", Age: 32, City: "New York"},
			{ID: 2, FirstName: "Emily", LastName: "Johnson", Age: 28, City: "Chicago"},
			{ID: 3, FirstName: "Michael", LastName: "Brown", Age: 41, City: "Los Angeles"},
			{ID: 4, FirstName: "Jessica", LastName: "Davis", Age: 24, City: "Seattle"},
			{ID: 5, FirstName: "David", LastName: "Wilson", Age: 37, City: "Boston"},
			{ID: 6, FirstName: "Sarah", LastName: "Taylor", Age: 29, City: "San Francisco"},
			{ID: 7, FirstName: "James", LastName: "Anderson", Age: 45, City: "Denver"},
			{ID: 8, FirstName: "Jennifer", LastName: "Thomas", Age: 31, City: "Austin"},
			{ID: 9, FirstName: "Robert", LastName: "Jackson", Age: 39, City: "Portland"},
			{ID: 10, FirstName: "Lisa", LastName: "White", Age: 26, City: "Miami"},
			{ID: 11, FirstName: "Daniel", LastName: "Harris", Age: 33, City: "Atlanta"},
			{ID: 12, FirstName: "Michelle", LastName: "Martin", Age: 27, City: "Dallas"},
			{ID: 13, FirstName: "William", LastName: "Thompson", Age: 42, City: "Phoenix"},
			{ID: 14, FirstName: "Elizabeth", LastName: "Garcia", Age: 35, City: "Philadelphia"},
			{ID: 15, FirstName: "Richard", LastName: "Martinez", Age: 38, City: "San Diego"},
			{ID: 16, FirstName: "Barbara", LastName: "Robinson", Age: 44, City: "Houston"},
			{ID: 17, FirstName: "Joseph", LastName: "Clark", Age: 30, City: "Las Vegas"},
			{ID: 18, FirstName: "Susan", LastName: "Rodriguez", Age: 25, City: "Nashville"},
			{ID: 19, FirstName: "Thomas", LastName: "Lewis", Age: 36, City: "Detroit"},
			{ID: 20, FirstName: "Margaret", LastName: "Lee", Age: 40, City: "Chicago"},
			{ID: 21, FirstName: "Andrew", LastName: "Walker", Age: 29, City: "Toronto"},
			{ID: 22, FirstName: "Amanda", LastName: "Allen", Age: 43, City: "Boston"},
			{ID: 23, FirstName: "Ryan", LastName: "Young", Age: 34, City: "Montreal"},
			{ID: 24, FirstName: "Stephanie", LastName: "Hernandez", Age: 31, City: "San Diego"},
			{ID: 25, FirstName: "Jason", LastName: "King", Age: 38, City: "Orlando"},
			{ID: 26, FirstName: "Nicole", LastName: "Wright", Age: 27, City: "Minneapolis"},
			{ID: 27, FirstName: "Brandon", LastName: "Lopez", Age: 39, City: "Chicago"},
			{ID: 28, FirstName: "Amy", LastName: "Hill", Age: 41, City: "San Francisco"},
			{ID: 29, FirstName: "Justin", LastName: "Scott", Age: 30, City: "Portland"},
			{ID: 30, FirstName: "Katherine", LastName: "Green", Age: 36, City: "Denver"},
			{ID: 31, FirstName: "Jack", LastName: "Adams", Age: 45, City: "New York"},
			{ID: 32, FirstName: "Rebecca", LastName: "Baker", Age: 33, City: "Philadelphia"},
			{ID: 33, FirstName: "Eric", LastName: "Gonzalez", Age: 28, City: "Miami"},
			{ID: 34, FirstName: "Laura", LastName: "Nelson", Age: 34, City: "Phoenix"},
			{ID: 35, FirstName: "Adam", LastName: "Carter", Age: 37, City: "Boston"},
			{ID: 36, FirstName: "Christine", LastName: "Mitchell", Age: 26, City: "Seattle"},
			{ID: 37, FirstName: "Stephen", LastName: "Perez", Age: 40, City: "Dallas"},
			{ID: 38, FirstName: "Monica", LastName: "Roberts", Age: 29, City: "Houston"},
			{ID: 39, FirstName: "Kevin", LastName: "Turner", Age: 32, City: "Chicago"},
			{ID: 40, FirstName: "Tiffany", LastName: "Phillips", Age: 43, City: "Atlanta"},
			{ID: 41, FirstName: "Alexander", LastName: "Campbell", Age: 31, City: "Detroit"},
			{ID: 42, FirstName: "Samantha", LastName: "Parker", Age: 38, City: "Los Angeles"},
			{ID: 43, FirstName: "Brian", LastName: "Evans", Age: 26, City: "Nashville"},
			{ID: 44, FirstName: "Melissa", LastName: "Edwards", Age: 39, City: "Austin"},
			{ID: 45, FirstName: "Christopher", LastName: "Collins", Age: 44, City: "Denver"},
			{ID: 46, FirstName: "Angela", LastName: "Stewart", Age: 28, City: "Chicago"},
			{ID: 47, FirstName: "Timothy", LastName: "Sanchez", Age: 37, City: "San Francisco"},
			{ID: 48, FirstName: "Danielle", LastName: "Morris", Age: 35, City: "Seattle"},
			{ID: 49, FirstName: "Joshua", LastName: "Rogers", Age: 30, City: "Portland"},
			{ID: 50, FirstName: "Kimberly", LastName: "Reed", Age: 42, City: "Phoenix"},
			{ID: 51, FirstName: "Jonathan", LastName: "Cook", Age: 33, City: "New York"},
			{ID: 52, FirstName: "Victoria", LastName: "Morgan", Age: 25, City: "Miami"},
			{ID: 53, FirstName: "Aaron", LastName: "Bell", Age: 38, City: "Boston"},
			{ID: 54, FirstName: "Rachel", LastName: "Murphy", Age: 41, City: "Chicago"},
			{ID: 55, FirstName: "Patrick", LastName: "Bailey", Age: 29, City: "San Diego"},
			{ID: 56, FirstName: "Lauren", LastName: "Rivera", Age: 36, City: "Detroit"},
			{ID: 57, FirstName: "Gregory", LastName: "Cooper", Age: 44, City: "Houston"},
			{ID: 58, FirstName: "Ashley", LastName: "Richardson", Age: 32, City: "Atlanta"},
			{ID: 59, FirstName: "Jesse", LastName: "Cox", Age: 27, City: "Dallas"},
			{ID: 60, FirstName: "Megan", LastName: "Howard", Age: 39, City: "Los Angeles"},
			{ID: 61, FirstName: "Charles", LastName: "Ward", Age: 28, City: "Chicago"},
			{ID: 62, FirstName: "Alexis", LastName: "Torres", Age: 43, City: "San Francisco"},
			{ID: 63, FirstName: "Scott", LastName: "Peterson", Age: 31, City: "Seattle"},
			{ID: 64, FirstName: "Kayla", LastName: "Gray", Age: 36, City: "Portland"},
			{ID: 65, FirstName: "Jeffrey", LastName: "Ramirez", Age: 40, City: "Phoenix"},
			{ID: 66, FirstName: "Alexandra", LastName: "James", Age: 26, City: "Austin"},
			{ID: 67, FirstName: "Kyle", LastName: "Watson", Age: 37, City: "Denver"},
			{ID: 68, FirstName: "Hannah", LastName: "Brooks", Age: 32, City: "Boston"},
			{ID: 69, FirstName: "Tyler", LastName: "Kelly", Age: 29, City: "Miami"},
			{ID: 70, FirstName: "Amber", LastName: "Sanders", Age: 35, City: "New York"},
			{ID: 71, FirstName: "Jose", LastName: "Price", Age: 44, City: "Chicago"},
			{ID: 72, FirstName: "Heather", LastName: "Bennett", Age: 30, City: "Philadelphia"},
			{ID: 73, FirstName: "Zachary", LastName: "Wood", Age: 38, City: "Dallas"},
			{ID: 74, FirstName: "Brittany", LastName: "Barnes", Age: 27, City: "Houston"},
			{ID: 75, FirstName: "Samuel", LastName: "Ross", Age: 42, City: "Atlanta"},
			{ID: 76, FirstName: "Taylor", LastName: "Henderson", Age: 31, City: "Detroit"},
			{ID: 77, FirstName: "Anthony", LastName: "Coleman", Age: 36, City: "San Diego"},
			{ID: 78, FirstName: "Alyssa", LastName: "Jenkins", Age: 26, City: "Los Angeles"},
			{ID: 79, FirstName: "Nathan", LastName: "Perry", Age: 39, City: "Chicago"},
			{ID: 80, FirstName: "Sophia", LastName: "Powell", Age: 33, City: "San Francisco"},
			{ID: 81, FirstName: "Olivia", LastName: "Hughes", Age: 30, City: "Portland"},
			{ID: 82, FirstName: "Ethan", LastName: "Russell", Age: 45, City: "Phoenix"},
			{ID: 83, FirstName: "Madison", LastName: "Mason", Age: 28, City: "New York"},
			{ID: 84, FirstName: "Benjamin", LastName: "Simmons", Age: 41, City: "Boston"},
			{ID: 85, FirstName: "Isabella", LastName: "Warren", Age: 32, City: "Chicago"},
			{ID: 86, FirstName: "Matthew", LastName: "Nichols", Age: 37, City: "Seattle"},
			{ID: 87, FirstName: "Natalie", LastName: "Grant", Age: 29, City: "Dallas"},
			{ID: 88, FirstName: "Jacob", LastName: "Gardner", Age: 43, City: "Houston"},
			{ID: 89, FirstName: "Emma", LastName: "Shaw", Age: 27, City: "Austin"},
			{ID: 90, FirstName: "Lucas", LastName: "Tran", Age: 36, City: "San Diego"},
			{ID: 91, FirstName: "Ava", LastName: "Olson", Age: 44, City: "Miami"},
			{ID: 92, FirstName: "Mason", LastName: "Kim", Age: 31, City: "Philadelphia"},
			{ID: 93, FirstName: "Chloe", LastName: "Nguyen", Age: 25, City: "Chicago"},
			{ID: 94, FirstName: "Michael", LastName: "Silva", Age: 38, City: "San Francisco"},
			{ID: 95, FirstName: "Mia", LastName: "Hudson", Age: 40, City: "Portland"},
			{ID: 96, FirstName: "Noah", LastName: "Snyder", Age: 33, City: "Phoenix"},
			{ID: 97, FirstName: "Zoey", LastName: "Anderson", Age: 29, City: "Denver"},
			{ID: 98, FirstName: "Elijah", LastName: "Gordon", Age: 42, City: "Detroit"},
			{ID: 99, FirstName: "Addison", LastName: "Hunter", Age: 26, City: "Los Angeles"},
			{ID: 100, FirstName: "Jayden", LastName: "Stone", Age: 37, City: "New York"},
		},
		filteredData:   nil,
		filters:        make(map[string]any),
		sortFields:     []string{},
		sortDirections: []string{},
		dirty:          true,
	}
}

// ensureFilteredData ensures the filtered data cache is up to date
func (p *PersonDataProvider) ensureFilteredData() {
	// If data is already filtered and not dirty, no need to refilter
	if !p.dirty && p.filteredData != nil {
		return
	}

	// Apply filters to the entire dataset at once
	filtered := make([]Person, 0, len(p.data))
	for _, person := range p.data {
		if p.matchesFilters(person) {
			filtered = append(filtered, person)
		}
	}

	// Apply sorting to the entire filtered dataset
	if len(p.sortFields) > 0 && len(filtered) > 0 {
		p.sortPersons(filtered)
	}

	// Store the filtered data for future requests
	p.filteredData = filtered
	p.dirty = false
}

// matchesFilters checks if a person matches all filters
func (p *PersonDataProvider) matchesFilters(person Person) bool {
	if len(p.filters) == 0 {
		return true
	}

	for key, value := range p.filters {
		switch key {
		case "id":
			if idVal, ok := value.(int); ok && person.ID != idVal {
				return false
			}
		case "firstName":
			if strVal, ok := value.(string); ok && strVal != "" {
				// Case-insensitive substring search
				pName := strings.ToLower(person.FirstName)
				searchVal := strings.ToLower(strVal)

				if !strings.Contains(pName, searchVal) {
					// Debug for first few records - COMMENTED OUT
					/*
						if person.ID <= 3 {
							fmt.Printf("No match: '%s' doesn't contain '%s'\n",
								pName, searchVal)
						}
					*/
					return false
				} else if person.ID <= 10 {
					// Show matches for the first few records - COMMENTED OUT
					/*
						fmt.Printf("Match found: '%s' contains '%s' (ID: %d)\n",
							pName, searchVal, person.ID)
					*/
				}
			}
		case "lastName":
			if strVal, ok := value.(string); ok && strVal != "" {
				pName := strings.ToLower(person.LastName)
				searchVal := strings.ToLower(strVal)
				if !strings.Contains(pName, searchVal) {
					return false
				}
			}
		case "minAge":
			if intVal, ok := value.(int); ok && person.Age < intVal {
				return false
			}
		case "maxAge":
			if intVal, ok := value.(int); ok && person.Age > intVal {
				return false
			}
		case "city":
			if strVal, ok := value.(string); ok && strVal != "" {
				pCity := strings.ToLower(person.City)
				searchVal := strings.ToLower(strVal)
				if !strings.Contains(pCity, searchVal) {
					return false
				}
			}
		}
	}

	return true
}

// GetTotal returns the total number of items after filtering
func (p *PersonDataProvider) GetTotal() int {
	p.ensureFilteredData()
	return len(p.filteredData)
}

// GetItems returns a slice of items based on the provided request
func (p *PersonDataProvider) GetItems(request vtable.DataRequest) ([]vtable.TableRow, error) {
	// Update provider's filter and sort state from the request
	changed := false

	// Check if filters have changed
	if len(p.filters) != len(request.Filters) {
		changed = true
	} else {
		for k, v := range request.Filters {
			if oldVal, exists := p.filters[k]; !exists || oldVal != v {
				changed = true
				break
			}
		}
	}

	// Apply new filters if changed
	if changed {
		p.filters = make(map[string]any)
		for k, v := range request.Filters {
			p.filters[k] = v
		}
		p.dirty = true
	}

	// Check if sort has changed
	if len(p.sortFields) != len(request.SortFields) {
		changed = true
	} else {
		for i, field := range request.SortFields {
			if i >= len(p.sortFields) || p.sortFields[i] != field || p.sortDirections[i] != request.SortDirections[i] {
				changed = true
				break
			}
		}
	}

	// Apply new sort if changed
	if changed {
		p.sortFields = make([]string, len(request.SortFields))
		copy(p.sortFields, request.SortFields)

		p.sortDirections = make([]string, len(request.SortDirections))
		copy(p.sortDirections, request.SortDirections)
		p.dirty = true
	}

	// Ensure filtered data is up to date
	p.ensureFilteredData()

	// Apply pagination - CRITICAL for consistent chunking
	start := request.Start
	count := request.Count

	// Make sure we don't go beyond the available data
	if start >= len(p.filteredData) {
		return []vtable.TableRow{}, nil
	}

	// Calculate end index, ensuring we don't go beyond available data
	end := start + count
	if end > len(p.filteredData) {
		end = len(p.filteredData)
	}

	// Convert to table rows - exactly what was requested, no more, no less
	rows := make([]vtable.TableRow, end-start)
	for i := 0; i < end-start; i++ {
		person := p.filteredData[start+i]
		rows[i] = vtable.TableRow{
			Cells: []string{
				fmt.Sprintf("%d", person.ID),
				person.FirstName,
				person.LastName,
				fmt.Sprintf("%d", person.Age),
				person.City,
			},
		}
	}

	return rows, nil
}

// FindItemIndex implements the SearchableDataProvider interface
func (p *PersonDataProvider) FindItemIndex(key string, value any) (int, bool) {
	p.ensureFilteredData()

	for i, person := range p.filteredData {
		switch key {
		case "id":
			if idVal, ok := value.(int); ok && person.ID == idVal {
				return i, true
			}
		case "firstName":
			if strVal, ok := value.(string); ok && person.FirstName == strVal {
				return i, true
			}
		case "lastName":
			if strVal, ok := value.(string); ok && person.LastName == strVal {
				return i, true
			}
		case "age":
			if intVal, ok := value.(int); ok && person.Age == intVal {
				return i, true
			}
		case "city":
			if strVal, ok := value.(string); ok && person.City == strVal {
				return i, true
			}
		}
	}

	return -1, false
}

// sortPersons sorts the persons slice based on sort fields
func (p *PersonDataProvider) sortPersons(data []Person) {
	if len(p.sortFields) == 0 {
		return
	}

	// Simple bubble sort
	for i := 0; i < len(data)-1; i++ {
		for j := 0; j < len(data)-i-1; j++ {
			if p.comparePersons(data[j], data[j+1]) > 0 {
				data[j], data[j+1] = data[j+1], data[j]
			}
		}
	}
}

// comparePersons compares two persons based on the sort fields
func (p *PersonDataProvider) comparePersons(a, b Person) int {
	for i, field := range p.sortFields {
		ascending := p.sortDirections[i] != "desc"

		var comparison int
		switch field {
		case "0", "id": // Column 0 (ID)
			if a.ID < b.ID {
				comparison = -1
			} else if a.ID > b.ID {
				comparison = 1
			} else {
				comparison = 0
			}
		case "1", "firstName": // Column 1 (First Name)
			comparison = strings.Compare(a.FirstName, b.FirstName)
		case "2", "lastName": // Column 2 (Last Name)
			comparison = strings.Compare(a.LastName, b.LastName)
		case "3", "age": // Column 3 (Age)
			if a.Age < b.Age {
				comparison = -1
			} else if a.Age > b.Age {
				comparison = 1
			} else {
				comparison = 0
			}
		case "4", "city": // Column 4 (City)
			comparison = strings.Compare(a.City, b.City)
		default:
			comparison = 0
		}

		if comparison != 0 {
			if !ascending {
				comparison = -comparison
			}
			return comparison
		}
	}

	return 0
}

// KeyReleasedMsg is sent when a key should no longer be highlighted
type KeyReleasedMsg struct{}

// SendKeyReleasedMsg creates a command that will reset the active key
func SendKeyReleasedMsg() tea.Cmd {
	return func() tea.Msg {
		return KeyReleasedMsg{}
	}
}

// FilteredTableDemo is the Bubble Tea model for the demo
type FilteredTableDemo struct {
	table           *vtable.TeaTable
	provider        *PersonDataProvider
	status          string
	debug           bool
	activeFilters   map[string]any
	activeSortField string
	activeSortDir   string

	// Add mode flags for multi-sort and multi-filter
	multiSortEnabled   bool
	multiFilterEnabled bool

	// Track currently pressed key for highlighting
	activeKey string
}

// NewFilteredTableDemo creates a new demo
func NewFilteredTableDemo() (*FilteredTableDemo, error) {
	// Create columns
	columns := []vtable.TableColumn{
		{Title: "ID", Width: 5, Alignment: vtable.AlignRight, Field: "id"},
		{Title: "First Name", Width: 12, Alignment: vtable.AlignLeft, Field: "firstName"},
		{Title: "Last Name", Width: 12, Alignment: vtable.AlignLeft, Field: "lastName"},
		{Title: "Age", Width: 5, Alignment: vtable.AlignRight, Field: "age"},
		{Title: "City", Width: 15, Alignment: vtable.AlignLeft, Field: "city"},
	}

	// Create table config
	config := vtable.DefaultTableConfig()
	config.Columns = columns
	config.ViewportConfig.Height = 10

	// Create data provider
	provider := NewPersonDataProvider()

	// Create the table
	table, err := vtable.NewTeaTable(config, provider, vtable.ColorfulTheme())
	if err != nil {
		return nil, err
	}

	return &FilteredTableDemo{
		table:              table,
		provider:           provider,
		status:             "Use number keys (1-5) to sort columns, letter keys to filter",
		debug:              false,
		activeFilters:      make(map[string]any),
		activeSortField:    "",
		activeSortDir:      "",
		multiSortEnabled:   false,
		multiFilterEnabled: false,
	}, nil
}

// toggleFilter toggles a filter on/off
func (m *FilteredTableDemo) toggleFilter(field string, value any) {
	// Get readable field name
	fieldName := getReadableFieldName(field)

	// Determine filter action (add/update/remove)
	filterAction := "added"

	// Check if we already have active filters and multi-filter is disabled
	if !m.multiFilterEnabled && len(m.activeFilters) > 0 && !hasKey(m.activeFilters, field) {
		// If multi-filter is disabled and we're trying to add a new filter when one already exists,
		// clear existing filters first
		m.activeFilters = make(map[string]any)
		m.provider.filters = make(map[string]any)
		m.table.ClearFilters()
	}

	// Check if filter already exists with the same key
	if existingValue, exists := m.activeFilters[field]; exists {
		// If filter exists with the same value, remove it
		if existingValue == value {
			delete(m.activeFilters, field)
			delete(m.provider.filters, field)
			filterAction = "removed"
		} else {
			// If filter exists with different value, update it
			m.activeFilters[field] = value
			m.provider.filters[field] = value
			filterAction = "updated"
		}
	} else {
		// If filter doesn't exist, add it
		m.activeFilters[field] = value
		m.provider.filters[field] = value
	}

	// Mark data as dirty to force refresh
	m.provider.dirty = true

	// Ensure filtered data is up to date
	m.provider.ensureFilteredData()

	// Get the new filter count to update status
	filteredCount := m.provider.GetTotal()

	// Check if filtering resulted in no data
	if filteredCount == 0 {
		// Handle empty result set case
		m.status = fmt.Sprintf("WARNING: Filter returned no results. %s filter: %s=%v",
			strings.Title(filterAction), fieldName, value)
	} else {
		// Regular status update
		m.status = fmt.Sprintf("%s filter: %s=%v (%d results)",
			strings.Title(filterAction), fieldName, value, filteredCount)

		// Add note about multi-filter mode
		if m.multiFilterEnabled && len(m.activeFilters) > 1 {
			m.status += fmt.Sprintf(" [%d active filters]", len(m.activeFilters))
		}
	}

	// Keep table's filters in sync with our local state
	// Apply the entire filter set at once
	m.table.ClearFilters()
	for field, value := range m.activeFilters {
		m.table.SetFilter(field, value)
	}

	// Important: Reset the data provider to refresh everything
	// This ensures proper height adjustment and consistent display
	m.table.SetDataProvider(m.provider)

	// Jump to the beginning of the filtered data to avoid scrolling issues
	m.table.JumpToStart()
}

// hasKey checks if a map has a specific key
func hasKey(m map[string]any, key string) bool {
	_, ok := m[key]
	return ok
}

// getReadableFieldName converts a field identifier to a readable name
func getReadableFieldName(field string) string {
	switch field {
	case "0", "id":
		return "ID"
	case "1", "firstName":
		return "First Name"
	case "2", "lastName":
		return "Last Name"
	case "3", "age", "minAge", "maxAge":
		return "Age"
	case "4", "city":
		return "City"
	default:
		return field
	}
}

// toggleSort cycles through sort states: ascending -> descending -> off
func (m *FilteredTableDemo) toggleSort(field string) {
	// Get readable field name
	fieldName := getReadableFieldName(field)

	// Print current sort state at the beginning if debugging
	if m.debug {
		fmt.Println("=== BEFORE TOGGLE ===")
		fmt.Printf("Provider sorts: %v\n", m.provider.sortFields)
		fmt.Printf("Provider directions: %v\n", m.provider.sortDirections)

		// Show table's sort state too
		request := m.table.GetDataRequest()
		fmt.Printf("Table sorts: %v\n", request.SortFields)
		fmt.Printf("Table directions: %v\n", request.SortDirections)
	}

	// Check if multi-sort is disabled and we already have sorts
	if !m.multiSortEnabled && len(m.provider.sortFields) > 0 && !contains(m.provider.sortFields, field) {
		// Clear existing sorts if we're adding a new one and multi-sort is disabled
		m.provider.sortFields = []string{}
		m.provider.sortDirections = []string{}
		m.table.ClearSort()
	}

	// Check if we're already sorting by this field
	var foundAt int = -1
	for i, existingField := range m.provider.sortFields {
		if existingField == field {
			foundAt = i
			break
		}
	}

	if foundAt >= 0 {
		// Field is already in sort list, check direction
		if m.provider.sortDirections[foundAt] == "asc" {
			// Switch to descending
			m.provider.sortDirections[foundAt] = "desc"

			// Move to front if not already at front
			if foundAt > 0 {
				// Save field/direction
				tempField := m.provider.sortFields[foundAt]
				tempDir := m.provider.sortDirections[foundAt]

				// Remove from current position
				m.provider.sortFields = append(m.provider.sortFields[:foundAt], m.provider.sortFields[foundAt+1:]...)
				m.provider.sortDirections = append(m.provider.sortDirections[:foundAt], m.provider.sortDirections[foundAt+1:]...)

				// Add to front
				m.provider.sortFields = append([]string{tempField}, m.provider.sortFields...)
				m.provider.sortDirections = append([]string{tempDir}, m.provider.sortDirections...)
			}

			// Build status message with all active sorts
			m.status = fmt.Sprintf("Sorting by %s (descending)", fieldName)
			if m.multiSortEnabled && len(m.provider.sortFields) > 1 {
				m.status += sortStatusSuffix(m.provider.sortFields, m.provider.sortDirections)
			}

			// Apply sort directly without clearing - use RemoveSort + AddSort to update the field
			m.table.RemoveSort(field)
			m.table.AddSort(field, "desc")
		} else {
			// Remove sort for this field
			m.provider.sortFields = append(m.provider.sortFields[:foundAt], m.provider.sortFields[foundAt+1:]...)
			m.provider.sortDirections = append(m.provider.sortDirections[:foundAt], m.provider.sortDirections[foundAt+1:]...)

			// Build status message
			if len(m.provider.sortFields) == 0 {
				m.status = fmt.Sprintf("Removed sort for %s (no active sorts)", fieldName)
			} else {
				m.status = fmt.Sprintf("Removed sort for %s", fieldName)
				if m.multiSortEnabled && len(m.provider.sortFields) > 1 {
					m.status += sortStatusSuffix(m.provider.sortFields, m.provider.sortDirections)
				}
			}

			// Remove just this field's sort
			m.table.RemoveSort(field)
		}

		m.provider.dirty = true
	} else {
		// Not currently sorting by this field, add ascending sort
		// Add to front to make it highest priority
		m.provider.sortFields = append([]string{field}, m.provider.sortFields...)
		m.provider.sortDirections = append([]string{"asc"}, m.provider.sortDirections...)

		m.provider.dirty = true

		// Build status message with all active sorts
		m.status = fmt.Sprintf("Sorting by %s (ascending)", fieldName)
		if m.multiSortEnabled && len(m.provider.sortFields) > 1 {
			m.status += sortStatusSuffix(m.provider.sortFields, m.provider.sortDirections)
		}

		// If there are existing sorts, use AddSort to preserve them
		// Otherwise use SetSort
		if len(m.provider.sortFields) > 1 {
			m.table.AddSort(field, "asc")
		} else {
			m.table.SetSort(field, "asc")
		}
	}

	// Ensure data is fully refreshed
	m.provider.ensureFilteredData()

	// Print final sort state if debugging
	if m.debug {
		fmt.Println("=== AFTER TOGGLE ===")
		fmt.Printf("Provider sorts: %v\n", m.provider.sortFields)
		fmt.Printf("Provider directions: %v\n", m.provider.sortDirections)

		request := m.table.GetDataRequest()
		fmt.Printf("Table sorts: %v\n", request.SortFields)
		fmt.Printf("Table directions: %v\n", request.SortDirections)
		fmt.Println("===================")
	}
}

// contains checks if a string slice contains a specific value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// sortStatusSuffix builds a suffix describing additional sorts
func sortStatusSuffix(fields []string, directions []string) string {
	if len(fields) <= 1 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(" (also sorted by: ")

	for i := 1; i < len(fields); i++ {
		if i > 1 {
			sb.WriteString(", ")
		}
		fieldName := getReadableFieldName(fields[i])
		dir := "asc"
		if directions[i] == "desc" {
			dir = "desc"
		}
		sb.WriteString(fmt.Sprintf("%s (%s)", fieldName, dir))
	}

	sb.WriteString(")")
	return sb.String()
}

// clearAll clears all filters and sorting
func (m *FilteredTableDemo) clearAll() {
	// Clear local state
	m.activeFilters = make(map[string]any)

	// Clear provider state directly to ensure clean state
	m.provider.filters = make(map[string]any)
	m.provider.sortFields = []string{}
	m.provider.sortDirections = []string{}
	m.provider.dirty = true // Force data refresh

	// Reset mode flags
	m.multiSortEnabled = false
	m.multiFilterEnabled = false

	// Clear all cached filtered data
	m.provider.filteredData = nil

	// Force data refresh
	m.provider.ensureFilteredData()

	// Clear table state
	m.table.ClearFilters()
	m.table.ClearSort()

	// Fully reset the table with the full dataset - this updates the height
	m.table.SetDataProvider(m.provider)

	// Jump to start with clean state
	m.table.JumpToStart()

	m.status = "Reset: cleared all filters and sorting"
}

// Init initializes the demo
func (m *FilteredTableDemo) Init() tea.Cmd {
	return nil
}

// Update handles events and updates the model
func (m *FilteredTableDemo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Store the pressed key for highlighting
		m.activeKey = msg.String()
		cmds = append(cmds, tea.Tick(200*time.Millisecond, func(_ time.Time) tea.Msg {
			return KeyReleasedMsg{}
		}))

		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit

		// Numeric keys for sorting
		case "1":
			// Toggle sort for ID (column 0)
			m.toggleSort("id")
			return m, nil

		case "2":
			// Toggle sort for first name (column 1)
			m.toggleSort("firstName")
			return m, nil

		case "3":
			// Toggle sort for last name (column 2)
			m.toggleSort("lastName")
			return m, nil

		case "4":
			// Toggle sort for age (column 3)
			m.toggleSort("age")
			return m, nil

		case "5":
			// Toggle sort for city (column 4)
			m.toggleSort("city")
			return m, nil

		// Special mode toggles
		case "S":
			// Toggle multi-sort mode
			m.multiSortEnabled = !m.multiSortEnabled
			if m.multiSortEnabled {
				m.status = "✅ MULTI-SORT MODE ENABLED - add multiple sorts with keys 1-5!"
			} else {
				// When disabling, clear all but the primary sort
				if len(m.provider.sortFields) > 1 {
					// Preserve only the first (primary) sort
					primaryField := m.provider.sortFields[0]
					primaryDir := m.provider.sortDirections[0]

					// Clear all sorts
					m.table.ClearSort()

					// Re-apply just the primary sort
					m.table.SetSort(primaryField, primaryDir)

					// Update provider to match
					m.provider.sortFields = []string{primaryField}
					m.provider.sortDirections = []string{primaryDir}
					m.provider.dirty = true
					m.provider.ensureFilteredData()
				}
				m.status = "❌ MULTI-SORT MODE DISABLED - only one sort field allowed"
			}
			return m, nil

		case "F":
			// Toggle multi-filter mode
			m.multiFilterEnabled = !m.multiFilterEnabled
			if m.multiFilterEnabled {
				m.status = "✅ MULTI-FILTER MODE ENABLED - add multiple filters with f/l/a/c keys!"
			} else {
				// When disabling, clear all but one filter if any exist
				if len(m.activeFilters) > 1 {
					// Get the first filter (arbitrary)
					var keepField string
					var keepValue any

					for k, v := range m.activeFilters {
						keepField = k
						keepValue = v
						break
					}

					// Clear all filters
					m.table.ClearFilters()
					m.activeFilters = make(map[string]any)

					// Re-apply just one filter
					m.activeFilters[keepField] = keepValue
					m.table.SetFilter(keepField, keepValue)

					// Update provider to match
					m.provider.filters = make(map[string]any)
					m.provider.filters[keepField] = keepValue
					m.provider.dirty = true
					m.provider.ensureFilteredData()

					m.status = fmt.Sprintf("❌ MULTI-FILTER MODE DISABLED - kept only filter: %s=%v",
						keepField, keepValue)
				} else {
					m.status = "❌ MULTI-FILTER MODE DISABLED - only one filter allowed"
				}
			}
			return m, nil

		// Filters
		case "f":
			// Toggle filter by first name with 'a'
			m.toggleFilter("firstName", "a")
			return m, nil

		case "l":
			// Toggle filter by last name with 'on'
			m.toggleFilter("lastName", "on")
			return m, nil

		case "a":
			// Toggle filter by age > 35
			m.toggleFilter("minAge", 35)
			return m, nil

		case "c":
			// Toggle filter by city with 'o'
			m.toggleFilter("city", "o")
			return m, nil

		// Multi-example
		case "m":
			// Remove the 'm' case for the demo
			return m, nil

		// Utility keys
		case "D":
			// Toggle debug mode
			m.debug = !m.debug
			if m.debug {
				// Enable more verbose internal debug output
				fmt.Println("\n==== DEBUG MODE ENABLED ====")
				fmt.Println("Current states:")
				fmt.Printf("Multi-sort: %v\n", m.multiSortEnabled)
				fmt.Printf("Multi-filter: %v\n", m.multiFilterEnabled)
				fmt.Printf("Number of active filters: %d\n", len(m.activeFilters))
				fmt.Printf("Number of active sorts: %d\n", len(m.provider.sortFields))

				fmt.Println("\nSort details:")
				fmt.Printf("Provider sorts: %v\n", m.provider.sortFields)
				fmt.Printf("Provider directions: %v\n", m.provider.sortDirections)

				request := m.table.GetDataRequest()
				fmt.Printf("Table sorts: %v\n", request.SortFields)
				fmt.Printf("Table directions: %v\n", request.SortDirections)
				fmt.Println("===========================")

				m.status = "Debug mode enabled - check console for detailed information"
			} else {
				fmt.Println("\n==== DEBUG MODE DISABLED ====")
				m.status = "Debug mode disabled"
			}

		case "r", "backspace":
			// Clear all filters and sorting
			m.clearAll()
			return m, nil
		}
	case KeyReleasedMsg:
		// Reset the active key
		m.activeKey = ""
		return m, nil
	}

	// Update the table
	_, tableCmd := m.table.Update(msg)
	cmds = append(cmds, tableCmd)

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m *FilteredTableDemo) View() string {
	var sb strings.Builder

	// Table (at the top)
	sb.WriteString(m.table.View())
	sb.WriteString("\n\n")

	// Status message
	sb.WriteString(m.status)
	sb.WriteString("\n\n")

	// Show mode status
	sb.WriteString("Modes: ")
	if m.multiSortEnabled {
		sb.WriteString("MULTI-SORT [ON] ")
	} else {
		sb.WriteString("MULTI-SORT [OFF] ")
	}

	if m.multiFilterEnabled {
		sb.WriteString("MULTI-FILTER [ON]")
	} else {
		sb.WriteString("MULTI-FILTER [OFF]")
	}
	sb.WriteString("\n\n")

	// Show active filters
	if len(m.activeFilters) > 0 {
		sb.WriteString("Active Filters: ")
		for k, v := range m.activeFilters {
			sb.WriteString(fmt.Sprintf("%s=%v ", k, v))
		}
		sb.WriteString("\n\n")
	}

	// Add help text
	sb.WriteString(m.renderHelpText())

	// Debug info with improved aesthetics
	if m.debug {
		// Add extra spacing before debug section
		sb.WriteString("\n\n")

		debugHeaderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#FF5F87")).
			Padding(0, 1)

		sectionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5F87FF")).
			Bold(true)

		valueStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

		labelStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA"))

		// Format a label-value pair
		formatItem := func(label, value string) string {
			return fmt.Sprintf("%s %s",
				labelStyle.Render(label+":"),
				valueStyle.Render(value))
		}

		// Start debug section
		sb.WriteString(debugHeaderStyle.Render(" Debug Information "))
		sb.WriteString("\n\n")

		// Table state section
		sb.WriteString(sectionStyle.Render("Table State"))
		sb.WriteString("\n")

		debugInfo := m.table.RenderDebugInfo()
		sb.WriteString(valueStyle.Render(debugInfo))
		sb.WriteString("\n\n")

		// Dataset information
		sb.WriteString(sectionStyle.Render("Dataset Info"))
		sb.WriteString("\n")
		sb.WriteString(formatItem("Filtered count", fmt.Sprintf("%d", m.provider.GetTotal())))
		sb.WriteString("\n\n")

		// Sort information
		sb.WriteString(sectionStyle.Render("Sort State"))
		sb.WriteString("\n")

		if len(m.provider.sortFields) == 0 {
			sb.WriteString(formatItem("Sort fields", "none"))
		} else {
			for i, field := range m.provider.sortFields {
				direction := m.provider.sortDirections[i]
				dirSymbol := "↑"
				if direction == "desc" {
					dirSymbol = "↓"
				}
				fieldName := getReadableFieldName(field)
				sb.WriteString(formatItem(
					fmt.Sprintf("Sort %d", i+1),
					fmt.Sprintf("%s %s", fieldName, dirSymbol)))
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")

		// Filter information
		sb.WriteString(sectionStyle.Render("Filter State"))
		sb.WriteString("\n")

		if len(m.provider.filters) == 0 {
			sb.WriteString(formatItem("Filters", "none"))
		} else {
			i := 0
			for k, v := range m.provider.filters {
				fieldName := getReadableFieldName(k)
				sb.WriteString(formatItem(
					fmt.Sprintf("Filter %d", i+1),
					fmt.Sprintf("%s=%v", fieldName, v)))
				sb.WriteString("\n")
				i++
			}
		}
	}

	return sb.String()
}

// renderHelpText creates help text with currently pressed key highlighted
func (m *FilteredTableDemo) renderHelpText() string {
	// Regular style for help text
	regularStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))

	// Key style for keys (not pressed)
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF9900")).
		Bold(true)

	// Active style for currently pressed key
	activeKeyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Background(lipgloss.Color("#FFFF00")).
		Bold(true)

	// Group label style
	groupStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		Bold(true)

	// Create a map to match the displayed key with what keypress actually generates
	keyMap := map[string][]string{
		"↑/↓": {"up", "down"},
		"j/k": {"j", "k"},
		"u/d": {"u", "d", "pgup", "pgdown"},
		"g/G": {"g", "G", "home", "end"},
		"1-5": {"1", "2", "3", "4", "5"},
		"S":   {"S"},
		"f":   {"f"},
		"l":   {"l"},
		"a":   {"a"},
		"c":   {"c"},
		"F":   {"F"},
		"r":   {"r"},
		"D":   {"D"},
		"q":   {"q"},
	}

	// Style a key based on whether it matches the active key
	k := func(displayKey string) string {
		// Check if this display key matches the currently pressed key
		if possibleKeys, exists := keyMap[displayKey]; exists {
			for _, possibleKey := range possibleKeys {
				if possibleKey == m.activeKey {
					return activeKeyStyle.Render(displayKey)
				}
			}
		}
		return keyStyle.Render(displayKey)
	}

	// Style a group label
	g := func(label string) string {
		return groupStyle.Render(label + ":")
	}

	// Navigation group
	nav := []string{
		fmt.Sprintf("%s/%s navigate", k("↑/↓"), k("j/k")),
		fmt.Sprintf("%s page up/down", k("u/d")),
		fmt.Sprintf("%s jump to start/end", k("g/G")),
	}

	// Sorting group
	sorting := []string{
		fmt.Sprintf("%s toggle column sort", k("1-5")),
		fmt.Sprintf("%s multi-sort mode", k("S")),
	}

	// Filtering group
	filtering := []string{
		fmt.Sprintf("%s first name", k("f")),
		fmt.Sprintf("%s last name", k("l")),
		fmt.Sprintf("%s age > 35", k("a")),
		fmt.Sprintf("%s city", k("c")),
		fmt.Sprintf("%s multi-filter mode", k("F")),
	}

	// Actions group
	actions := []string{
		fmt.Sprintf("%s reset all", k("r")),
		fmt.Sprintf("%s debug", k("D")),
		fmt.Sprintf("%s quit", k("q")),
	}

	// Format each group on its own line with bullet separators
	var sb strings.Builder

	sb.WriteString(g("Navigation") + " ")
	sb.WriteString(strings.Join(nav, " • "))
	sb.WriteString("\n")

	sb.WriteString(g("Sorting") + " ")
	sb.WriteString(strings.Join(sorting, " • "))
	sb.WriteString("\n")

	sb.WriteString(g("Filtering") + " ")
	sb.WriteString(strings.Join(filtering, " • "))
	sb.WriteString("\n")

	sb.WriteString(g("Actions") + " ")
	sb.WriteString(strings.Join(actions, " • "))

	return regularStyle.Render(sb.String())
}

func main() {
	// Create the model
	model, err := NewFilteredTableDemo()
	if err != nil {
		fmt.Printf("Error creating demo: %v\n", err)
		os.Exit(1)
	}

	// Create and run the program WITHOUT fullscreen mode
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
