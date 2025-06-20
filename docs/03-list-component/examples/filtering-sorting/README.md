# VTable Filtering & Sorting Demo

This example demonstrates VTable's powerful filtering and sorting capabilities. It shows how to:

- **Filter data** by different criteria (job role, city, age, search terms)
- **Sort data** by multiple fields with different directions
- **Combine filters and sorts** for complex data exploration
- **Track active filters and sorts** in the UI
- **Show filtered counts** vs total counts

## Running the Example

```bash
cd docs/03-list-component/examples/filtering-sorting
go run main.go
```

## Sample Data

The demo uses a list of 20 people with the following fields:
- **Name**: Person's full name
- **Age**: Age in years (25-45)
- **Job**: Various tech roles (Engineer, Manager, Designer, etc.)
- **City**: Different US cities

## Controls

### Filtering
- **1** - Toggle "Engineer" job filter
- **2** - Toggle "Manager" job filter  
- **3** - Toggle "30+" age filter (people 30 or older)
- **4** - Toggle "San" city filter (cities containing "San")
- **5** - Toggle "Developer" search (search across all fields)

### Sorting
- **!** (Shift+1) - Toggle sort by Name (asc → desc → none)
- **@** (Shift+2) - Toggle sort by Age (asc → desc → none)
- **#** (Shift+3) - Toggle sort by Job (asc → desc → none)
- **$** (Shift+4) - Toggle sort by City (asc → desc → none)

### Clearing
- **r** - Clear ALL filters and sorts
- **R** (Shift+r) - Clear only filters
- **Ctrl+R** - Clear only sorts

### Navigation
- **↑/↓** or **j/k** - Move cursor up/down
- **PgUp/PgDn** or **h/l** - Page up/down
- **Home/End** or **g/G** - Jump to start/end
- **Space** - Select/deselect current item
- **Ctrl+A** - Select all (visible) items
- **Ctrl+D** - Clear all selections
- **q** - Quit

## UI Features

The interface shows:

```
🔍 VTable Filtering & Sorting Demo

Filters: job=Engineer, city=San
Sorts: age↑, name↑
Showing 3 of 20 people

►  Alice Johnson (28) - UX Designer in San Francisco
   Bob Chen (34) - Software Engineer in New York  
   Carol Rodriguez (45) - Product Manager in Austin
   [... more items ...]

Added job filter: Engineer

Filters: 1=Engineer 2=Manager 3=30+ 4=San 5=Developer • Sorts: !=Name @=Age #=Job $=City • Clear: r=All R=Filters Ctrl+R=Sorts • Quit: q
```

### Status Display
- **Active Filters**: Shows current filter criteria
- **Active Sorts**: Shows sort fields with direction arrows (↑/↓)
- **Item Count**: "Showing X of Y" - filtered vs total
- **Last Action**: Status message showing what you just did

## Try These Combinations

1. **Find Senior Engineers**:
   - Press `1` (Engineer filter)
   - Press `3` (30+ age filter)
   - Press `@` (sort by age)

2. **West Coast Managers**:
   - Press `2` (Manager filter)  
   - Press `4` (San filter - catches San Francisco, San Jose, San Antonio)
   - Press `!` (sort by name)

3. **Multi-Level Sort**:
   - Press `#` (sort by job first)
   - Press `@` (then sort by age within jobs)
   - Press `!` (then sort by name within age groups)

4. **Global Search**:
   - Press `5` (search for "Developer")
   - See how it finds "Frontend Developer", "Backend Developer", etc.

## Technical Highlights

### DataSource Integration
The filtering and sorting logic is implemented in the `PersonDataSource`:
- **`matchesFilters()`** - Applies all active filters
- **`applySorting()`** - Handles multi-field sorting  
- **`ensureFilteredData()`** - Caches results for performance

### Command Pattern
Uses VTable's command system:
- **`FilterSetCmd()`** / **`FilterClearCmd()`** - Filter management
- **`SortAddCmd()`** / **`SortRemoveCmd()`** - Sort management  
- **`FiltersClearAllCmd()`** / **`SortsClearAllCmd()`** - Bulk clearing

### Performance Optimization
- **Dirty tracking**: Only reprocesses when filters/sorts change
- **Request caching**: Avoids redundant work on navigation
- **Filtered totals**: Shows accurate counts after filtering

## Key Learning Points

1. **DataSource Responsibility**: Your DataSource does the actual filtering/sorting, VTable manages the UI state

2. **Filter Combination**: Multiple filters are AND-ed together (all must match)

3. **Sort Priority**: Earlier sorts have higher priority in multi-field sorting

4. **Performance**: Cache filtered results and use dirty flags for large datasets

5. **User Feedback**: Always show what filters/sorts are active and their effect on the data

This example demonstrates that VTable can handle complex data operations while maintaining excellent performance and user experience! 