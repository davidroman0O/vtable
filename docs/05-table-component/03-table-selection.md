# Table Selection

## What We're Adding

Taking our virtualized employee table and adding **selection functionality**. We'll demonstrate single and multiple selection modes, selection state management, and visual feedback - all while maintaining smooth performance with large datasets.

## Why Selection Matters

Selection lets users:
- **Choose specific employees** for actions (edit, delete, export)
- **Bulk operations** on multiple employees at once  
- **Track which items** they've processed or flagged
- **Build workflows** around user choices

## Step 1: Configure Selection Mode

Start with the same virtualized table and add selection configuration:

```go
func createTableConfig() core.TableConfig {
	return core.TableConfig{
		Columns:     createEmployeeColumns(),
		ShowHeader:  true,
		ShowBorders: true,
		SelectionMode: core.SelectionMultiple, // Enable multiple selection
		ViewportConfig: core.ViewportConfig{
			Height:             10,
			ChunkSize:          25,
			TopThreshold:       3,
			BottomThreshold:    3,
			BoundingAreaBefore: 50,
			BoundingAreaAfter:  50,
		},
		Theme: config.DefaultTheme(),
		KeyMap: core.NavigationKeyMap{
			Up:       []string{"up", "k"},
			Down:     []string{"down", "j"},
			PageUp:   []string{"pgup", "h"},
			PageDown: []string{"pgdown", "l"},
			Home:     []string{"home", "g"},
			End:      []string{"end", "G"},
			Select:   []string{"enter", " "}, // Space and Enter to select
			SelectAll: []string{"ctrl+a"},   // Ctrl+A to select all
			Quit:     []string{"q"},
		},
	}
}
```

## Step 2: Enhanced Data Source with Full Data Storage

Update the data source to store all data and track selections:

```go
type LargeEmployeeDataSource struct {
	totalEmployees int
	data           []core.TableRow  // Store all data like the full featured example
	selectedItems  map[string]bool  // Selection state
	recentActivity []string         // Track recent selection activity
}

func NewLargeEmployeeDataSource(totalCount int) *LargeEmployeeDataSource {
	// Generate ALL data upfront like the full featured example
	data := make([]core.TableRow, totalCount)
	
	departments := []string{"Engineering", "Marketing", "Sales", "HR", "Finance", "Operations"}
	statuses := []string{"Active", "On Leave", "Remote"}

	for i := 0; i < totalCount; i++ {
		data[i] = core.TableRow{
			ID: fmt.Sprintf("emp-%d", i+1),
			Cells: []string{
				fmt.Sprintf("Employee %d", i+1),
				departments[rand.Intn(len(departments))],
				statuses[rand.Intn(len(statuses))],
				fmt.Sprintf("$%d,000", 45+rand.Intn(100)), // $45k-$145k
			},
		}
	}

	return &LargeEmployeeDataSource{
		totalEmployees: totalCount,
		data:           data,
		selectedItems:  make(map[string]bool),
		recentActivity: make([]string, 0),
	}
}
```

## Step 3: Proper Selection Methods

Implement selection methods following the full featured example pattern:

```go
func (ds *LargeEmployeeDataSource) SetSelected(index int, selected bool) tea.Cmd {
	return func() tea.Msg {
		if index >= 0 && index < len(ds.data) {
			id := ds.data[index].ID

			// Actually update selection state
			if selected {
				ds.selectedItems[id] = true
				ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected: %s", ds.data[index].Cells[0]))
			} else {
				delete(ds.selectedItems, id)
				ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Deselected: %s", ds.data[index].Cells[0]))
			}

			// Keep only last 10 activities
			if len(ds.recentActivity) > 10 {
				ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
			}

			return core.SelectionResponseMsg{
				Success:   true,
				Index:     index,
				ID:        id,
				Selected:  selected,
				Operation: "toggle",
			}
		}

		return core.SelectionResponseMsg{
			Success:   false,
			Index:     index,
			ID:        "",
			Selected:  false,
			Operation: "toggle",
			Error:     fmt.Errorf("invalid index: %d", index),
		}
	}
}

func (ds *LargeEmployeeDataSource) SelectAll() tea.Cmd {
	return func() tea.Msg {
		// Select all items - EXACTLY like full featured example
		for _, row := range ds.data {
			ds.selectedItems[row.ID] = true
		}

		ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Selected all %d items", len(ds.data)))

		// Keep only last 10 activities
		if len(ds.recentActivity) > 10 {
			ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
		}

		return core.SelectionResponseMsg{
			Success:   true,
			Index:     -1,
			ID:        "",
			Selected:  true,
			Operation: "selectAll",
		}
	}
}

func (ds *LargeEmployeeDataSource) ClearSelection() tea.Cmd {
	return func() tea.Msg {
		count := len(ds.selectedItems)
		ds.selectedItems = make(map[string]bool) // Clear all selections
		ds.recentActivity = append(ds.recentActivity, fmt.Sprintf("Cleared %d selections", count))

		// Keep only last 10 activities
		if len(ds.recentActivity) > 10 {
			ds.recentActivity = ds.recentActivity[len(ds.recentActivity)-10:]
		}

		return core.SelectionResponseMsg{
			Success:   true,
			Index:     -1,
			ID:        "",
			Selected:  false,
			Operation: "clear",
		}
	}
}

func (ds *LargeEmployeeDataSource) GetSelectionCount() int {
	return len(ds.selectedItems)
}

func (ds *LargeEmployeeDataSource) GetRecentActivity() []string {
	return ds.recentActivity
}
```

## Step 4: Enhanced LoadChunk with Selection State

Update chunk loading to apply selection state properly:

```go
func (ds *LargeEmployeeDataSource) LoadChunk(request core.DataRequest) tea.Cmd {
	return func() tea.Msg {
		// Simulate realistic database query time
		time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

		start := request.Start
		end := start + request.Count
		if end > ds.totalEmployees {
			end = ds.totalEmployees
		}

		var items []core.Data[any]
		for i := start; i < end; i++ {
			if i < len(ds.data) {
				items = append(items, core.Data[any]{
					ID:       ds.data[i].ID,
					Item:     ds.data[i],
					Selected: ds.selectedItems[ds.data[i].ID], // Apply selection state
					Metadata: core.NewTypedMetadata(),
				})
			}
		}

		return core.DataChunkLoadedMsg{
			StartIndex: start,
			Items:      items,
			Request:    request,
		}
	}
}
```

## Step 5: Selection Commands and Response Handling

Handle selection commands in the app Update method:

```go
func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle jump form if it's open
		if app.showJumpForm {
			// ... jump form handling ...
		}
		
		// Normal key handling when form is not open
		switch msg.String() {
		case "q", "ctrl+c":
			return app, tea.Quit
		case "J":
			// Open jump form
			app.showJumpForm = true
			app.jumpInput = ""
			return app, nil
		
		// Selection commands
		case " ", "enter":
			// Toggle selection of current item
			return app, core.SelectCurrentCmd()
		case "ctrl+a":
			// Select all items
			return app, core.SelectAllCmd()
		case "c":
			// Clear all selections
			return app, core.SelectClearCmd()
		case "s":
			// Show selection info
			app.showSelectionInfo()
			return app, nil
			
		default:
			var cmd tea.Cmd
			_, cmd = app.table.Update(msg)
			app.updateStatus()
			return app, cmd
		}

	// Handle selection responses - clean and simple
	case core.SelectionResponseMsg:
		app.updateStatus()
		// Pass to table without extra processing
		var cmd tea.Cmd
		_, cmd = app.table.Update(msg)
		return app, cmd

	// ... other message handling ...
	}
}

func (app *App) showSelectionInfo() {
	count := app.dataSource.GetSelectionCount()
	if count > 0 {
		app.statusMessage = fmt.Sprintf("✓ %d employees selected | Use c to clear, space to toggle", count)
	} else {
		app.statusMessage = "No employees selected | Use space to select, ctrl+a for all"
	}
}
```

## Step 6: Enhanced View with Activity Display

Show selection info and recent activity:

```go
func (app App) View() string {
	var sections []string

	// Show jump form if active
	if app.showJumpForm {
		sections = append(sections, fmt.Sprintf("Jump to employee (1-%d): %s_", app.totalEmployees, app.jumpInput))
		sections = append(sections, "")
	}

	// Status message
	sections = append(sections, app.statusMessage)
	sections = append(sections, "")

	// Table
	sections = append(sections, app.table.View())

	// Show selection info
	selectionCount := app.dataSource.GetSelectionCount()
	if selectionCount > 0 {
		sections = append(sections, "")
		sections = append(sections, fmt.Sprintf("Selected: %d items", selectionCount))
	}

	// Show recent activity
	recentActivity := app.dataSource.GetRecentActivity()
	if len(recentActivity) > 0 {
		sections = append(sections, "")
		sections = append(sections, "Recent Activity:")
		for i := len(recentActivity) - 1; i >= 0 && i >= len(recentActivity)-3; i-- {
			sections = append(sections, fmt.Sprintf("  • %s", recentActivity[i]))
		}
	}

	// Join all sections
	return strings.Join(sections, "\n")
}
```

## What You'll See

```
Employee 1/10000 | Selected: 0 | Use space/enter ctrl+a c s J, q to quit

│ ●  │Employee Name       │  Department   │   Status   │      Salary│
│ ►  │Employee 1          │      HR       │   Remote   │     $91,000│
│    │Employee 2          │   Marketing   │  On Leave  │     $62,000│  
│ ✓  │Employee 3          │    Finance    │  On Leave  │     $70,000│ ← Selected
│    │Employee 4          │      HR       │   Remote   │     $46,000│
│ ✓  │Employee 5          │  Engineering  │   Remote   │    $123,000│ ← Selected
│    │Employee 6          │   Marketing   │   Remote   │     $47,000│
│    │Employee 7          │  Operations   │  On Leave  │     $55,000│
│    │Employee 8          │   Marketing   │   Remote   │     $47,000│
│    │Employee 9          │  Engineering   │   Remote   │     $81,000│

Selected: 2 items

Recent Activity:
  • Selected: Employee 5
  • Selected: Employee 3
```

**After pressing Ctrl+A (select all):**
```
Employee 1/10000 | Selected: 10000 | Use space/enter ctrl+a c s J, q to quit

Selected: 10000 items

Recent Activity:
  • Selected all 10000 items
```

**After pressing 's' for selection info:**
```
✓ 10000 employees selected | Use c to clear, space to toggle
```

## Selection Modes

### Single Selection
```go
SelectionMode: core.SelectionSingle, // Only one item at a time
```
- Only one employee can be selected
- Selecting another automatically deselects the previous
- Good for "edit current employee" workflows

### Multiple Selection  
```go
SelectionMode: core.SelectionMultiple, // Multiple items allowed
```
- Multiple employees can be selected simultaneously  
- Each selection is tracked independently
- Good for bulk operations

### No Selection
```go
SelectionMode: core.SelectionNone, // Disable selection
```
- Selection commands do nothing
- Visual selection indicators hidden
- Good for read-only displays

## Key Controls

| Key | Action |
|-----|--------|
| `Space` `Enter` | Toggle selection of current employee |
| `Ctrl+A` | Select all employees |
| `c` | Clear all selections |
| `s` | Show selection information |
| `J` | Jump to specific employee |
| `j/k` `↑↓` | Navigate |
| `h/l` `PgUp/PgDn` | Page navigation |
| `g/G` | Jump to start/end |

## Key Improvements from Data Virtualization

1. **Enhanced data structure** - stores all data upfront for reliable selection
2. **Robust selection tracking** - proper state management with activity logging
3. **Selection commands** in Update method (space, ctrl+a, c, s)
4. **Clean selection response handling** - no unnecessary chunk refreshes
5. **Activity display** - shows recent selection operations
6. **Visual indicators** (✓) for selected items
7. **Selection count tracking** - displayed in status and separate section

## Try It

1. **Change selection mode**: Try `SelectionSingle` vs `SelectionMultiple`
2. **Select many items**: Use space to select, then `s` to see count
3. **Bulk select**: Press `Ctrl+A` to select all 10,000 employees
4. **Clear and repeat**: Press `c` to clear, then select different items
5. **Jump and select**: Use `J` to jump to employee 5000, then select it
6. **Watch activity**: Notice the "Recent Activity" section tracking your selections

## Selection Performance

- **Large datasets**: Selection works smoothly with 10,000+ items
- **Memory efficient**: Uses optimized data structure from full featured example
- **Activity tracking**: Shows last few selection operations
- **Visual feedback**: Selected items show ✓ indicator and count display
- **Chunk loading**: Selection state preserved across chunk loads
- **Bulk operations**: SelectAll handles 10,000 employees efficiently

## What's Next

The [cell constraints](04-cell-constraints.md) section shows how to control column width, alignment, and padding for better data presentation.

## Key Point

**Selection now works flawlessly with data virtualization** - you can select specific employees, jump to employee 5000, select it, then jump to employee 8000 and select it too. VTable maintains selection state perfectly across all chunk loading, with activity tracking showing exactly what's happening. 