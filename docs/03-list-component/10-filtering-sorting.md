# Filtering and Sorting

## What This Really Is

Filtering and sorting aren't "features we're adding" to VTable - **they're built into the core DataSource interface**. VTable has always sent filter and sort parameters in every data request. We're just implementing a DataSource that actually uses them.

## The Core Truth

Every time VTable requests data, it sends a `DataRequest` with filters and sorts already included:

```go
type DataRequest struct {
    Start          int
    Count          int
    SortFields     []string        // ← Always sent
    SortDirections []string        // ← Always sent  
    Filters        map[string]any  // ← Always sent
}
```

**Your DataSource receives these parameters automatically**. You just need to use them.

## Understanding the Flow

```
User presses "1" → FilterSetMsg → VTable updates internal state → 
VTable calls LoadChunk(DataRequest{Filters: {"job": "Engineer"}}) →
Your DataSource handles the filtering and returns results
```

VTable manages the UI state, your DataSource handles the data logic.

## Implementing a Filtering DataSource

Here's how to implement the DataSource interface to handle the filters VTable sends:

```go
type PersonDataSource struct {
    people []Person
    selected map[int]bool
}

func (ds *PersonDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
    return func() tea.Msg {
        // VTable automatically provides filters in the request
        filtered := ds.applyFilters(ds.people, request.Filters)
        sorted := ds.applySorts(filtered, request.SortFields, request.SortDirections)
        
        // Return the requested chunk from filtered/sorted data
        start := request.Start
        count := request.Count
        
        var items []core.Data[any]
        for i := start; i < start+count && i < len(sorted); i++ {
            items = append(items, core.Data[any]{
                ID:       fmt.Sprintf("person-%d", i),
                Item:     sorted[i],
                Selected: ds.selected[i],
            })
        }

        return core.DataChunkLoadedMsg{
            StartIndex: request.Start,
            Items:      items,
            Request:    request,
        }
    }
}
```

**Key insight**: `request.Filters` contains whatever filters VTable has active. You just apply them to your data.

## Implementing the Filter Logic

```go
func (ds *PersonDataSource) applyFilters(people []Person, filters map[string]any) []Person {
    if len(filters) == 0 {
        return people // No filters, return all data
    }
    
    var filtered []Person
    for _, person := range people {
        if ds.matchesAllFilters(person, filters) {
            filtered = append(filtered, person)
        }
    }
    return filtered
}

func (ds *PersonDataSource) matchesAllFilters(person Person, filters map[string]any) bool {
    for field, value := range filters {
        switch field {
        case "job":
            if jobFilter, ok := value.(string); ok {
                if !strings.Contains(strings.ToLower(person.Job), strings.ToLower(jobFilter)) {
                    return false
                }
            }
        case "city":
            if cityFilter, ok := value.(string); ok {
                if !strings.Contains(strings.ToLower(person.City), strings.ToLower(cityFilter)) {
                    return false
                }
            }
        case "minAge":
            if ageFilter, ok := value.(int); ok {
                if person.Age < ageFilter {
                    return false
                }
            }
        }
    }
    return true // Passes all filters
}
```

## Implementing the Sort Logic

```go
func (ds *PersonDataSource) applySorts(people []Person, fields []string, directions []string) []Person {
    if len(fields) == 0 {
        return people // No sorting requested
    }
    
    // Make a copy to avoid modifying the original
    sorted := make([]Person, len(people))
    copy(sorted, people)
    
    sort.Slice(sorted, func(i, j int) bool {
        for idx, field := range fields {
            direction := "asc"
            if idx < len(directions) {
                direction = directions[idx]
            }
            
            var comparison int
            switch field {
            case "name":
                comparison = strings.Compare(sorted[i].Name, sorted[j].Name)
            case "age":
                comparison = sorted[i].Age - sorted[j].Age
            case "job":
                comparison = strings.Compare(sorted[i].Job, sorted[j].Job)
            case "city":
                comparison = strings.Compare(sorted[i].City, sorted[j].City)
            }
            
            if comparison != 0 {
                if direction == "desc" {
                    return comparison > 0
                }
                return comparison < 0
            }
        }
        return false
    })
    
    return sorted
}
```

## Returning Filtered Totals

Your `GetTotal()` must return the count after filtering:

```go
func (ds *PersonDataSource) GetTotal() tea.Cmd {
    return func() tea.Msg {
        // Apply current filters to get accurate count
        filtered := ds.applyFilters(ds.people, ds.lastFilters)
        return core.DataTotalMsg{Total: len(filtered)}
    }
}
```

**Critical**: VTable needs to know the filtered total for proper scrollbar sizing and navigation.

## Application Setup

Your application just uses VTable's built-in filtering commands:

```go
func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "1":
            // Use VTable's built-in filter command
            return app, core.FilterSetCmd("job", "Engineer")
        case "2":
            return app, core.FilterSetCmd("job", "Manager")
        case "3":
            return app, core.FilterSetCmd("minAge", 30)
        case "!":
            // Use VTable's built-in sort command  
            return app, core.SortToggleCmd("name")
        case "@":
            return app, core.SortToggleCmd("age")
        case "r":
            // Clear all filters and sorts
            return app, tea.Batch(
                core.FiltersClearAllCmd(),
                core.SortsClearAllCmd(),
            )
        }
    }
    
    // VTable handles the filtering automatically
    var cmd tea.Cmd
    _, cmd = app.list.Update(msg)
    return app, cmd
}
```

**The magic**: You don't manage filter state in your app. VTable does that automatically and passes the current state to your DataSource in every request.

## What You'll See

```
🔍 VTable Filtering & Sorting Demo

Filters: job=Engineer
Sorts: age↑
Showing 6 of 20 people

►  David Kim (29) - DevOps Engineer in Seattle
   Bob Chen (34) - Software Engineer in New York
   Peter Anderson (36) - DevOps Engineer in San Jose
   Henry Garcia (38) - Backend Developer in Chicago
   Tara Kim (31) - Software Engineer in San Francisco
   Ivy Martinez (31) - QA Engineer in Miami

Press 1=Engineer 2=Manager 3=30+ • !=Name @=Age • r=Reset • q=Quit
```

## Key Architectural Points

### 1. **DataSource Interface is Filter-Ready**
The `DataRequest` struct has always included filters and sorts. No "upgrading" needed.

### 2. **VTable Manages UI State**
You don't track active filters in your app. VTable does that and sends them to your DataSource.

### 3. **Your DataSource Applies Logic**
Whether you filter in-memory, query a database, or call an API - that's your choice. VTable just provides the parameters.

### 4. **Commands Drive Everything**
Use `core.FilterSetCmd()`, `core.SortToggleCmd()`, etc. These are VTable's built-in commands for managing filter/sort state.

## The Core Pattern

```go
// 1. VTable receives command
core.FilterSetCmd("job", "Engineer")

// 2. VTable updates internal state and calls your DataSource
dataSource.LoadChunk(DataRequest{
    Start: 0,
    Count: 10,
    Filters: {"job": "Engineer"},  // ← VTable provides this
})

// 3. Your DataSource handles the filtering
filtered := applyFilters(data, request.Filters)

// 4. Return results
return DataChunkLoadedMsg{Items: filtered}
```

VTable's job: UI state management  
Your job: Data logic implementation
