# Table Selection Example

A large employee database demonstrating VTable's selection capabilities with 10,000 employees, building on the data virtualization example with selection modes and state management.

## Features Demonstrated

- **Multiple selection mode** allowing simultaneous selection of multiple employees
- **Selection state management** with persistent selection across chunk loading
- **Selection commands** (select current, select all, clear all, show info)
- **Visual selection feedback** with ✓ indicators for selected items
- **Selection count tracking** displayed in status bar
- **Performance optimization** for large dataset selection
- **Jump-to-index functionality** with selection preserved
- **Chunked loading** with selection state maintained

## Selection Features

### Selection Modes
- **Multiple Selection**: Select many employees simultaneously (default in this example)
- **Single Selection**: Change `SelectionMode: core.SelectionSingle` for one-at-a-time selection
- **No Selection**: Change `SelectionMode: core.SelectionNone` to disable selection

### Selection Performance
- **Efficient storage**: Only selected IDs stored, not full employee objects
- **Large dataset support**: Handles selection of thousands of items smoothly
- **Chunk preservation**: Selection state maintained across data loading
- **Bulk operations**: Select All handles 10,000 employees efficiently

## Running the Example

```bash
cd docs/05-table-component/examples/selection-table
go run .
```

## Controls

### Navigation (same as virtualized example)
| Key | Action |
|-----|--------|
| `↑` `k` | Move up one row |
| `↓` `j` | Move down one row |
| `g` | Jump to first employee |
| `G` | Jump to last employee |
| `h` `PgUp` | Jump up 10 rows |
| `l` `PgDn` | Jump down 10 rows |
| `J` | Open jump-to-index form |
| `q` | Quit |

### Selection Controls
| Key | Action |
|-----|--------|
| `Space` `Enter` | Toggle selection of current employee |
| `Ctrl+A` | Select all employees |
| `c` | Clear all selections |
| `s` | Show selection information |

## What You'll See

### Normal Operation with Selection
```
Employee 1/10000 | Selected: 2 | Use space/enter ctrl+a c s J, q to quit

│ ●  │Employee Name       │  Department   │   Status   │      Salary│
│ ►  │Employee 1          │  Engineering  │   Active   │     $67,000│
│    │Employee 2          │   Marketing   │   Remote   │     $58,000│  
│ ✓  │Employee 3          │      Sales    │   Active   │     $73,000│ ← Selected
│    │Employee 4          │        HR     │  On Leave  │     $51,000│
│ ✓  │Employee 5          │  Engineering  │   Active   │     $89,000│ ← Selected
│    │Employee 6          │   Marketing   │   Remote   │     $64,000│
│    │Employee 7          │     Finance   │   Active   │     $76,000│
│    │Employee 8          │  Operations   │   Active   │     $52,000│
│    │Employee 9          │      Sales    │   Active   │     $68,000│
│    │Employee 10         │  Engineering  │  On Leave  │     $91,000│
```

### Selection Information Display
**After pressing 's' key:**
```
✓ 2 employees selected | Use c to clear, space to toggle
```

**When no items selected:**
```
No employees selected | Use space to select, ctrl+a for all
```

### Jump Form with Selection Preserved
```
Jump to employee (1-10000): 5000_

Enter employee number (1-10000), Enter to jump, Esc to cancel

│ ●  │Employee Name       │  Department   │   Status   │      Salary│
│ ►  │Employee 5000       │  Engineering  │   Active   │     $67,000│
│ ✓  │Employee 5001       │   Marketing   │   Remote   │     $58,000│ ← Still selected
```

## Try These Experiments

### 1. Selection Modes
Modify the configuration in `createTableConfig()`:

```go
// Multiple selection (default)
SelectionMode: core.SelectionMultiple,

// Single selection only
SelectionMode: core.SelectionSingle,

// No selection
SelectionMode: core.SelectionNone,
```

### 2. Selection Workflow
1. **Select scattered employees**: Use space to select employees 1, 5, 10
2. **Jump around**: Press `J`, jump to employee 3000, select it
3. **Check count**: Press `s` to see "4 employees selected"
4. **Jump back**: Press `J`, jump to employee 1, see it's still selected
5. **Bulk select**: Press `Ctrl+A` to select all 10,000 employees
6. **Clear and restart**: Press `c` to clear all selections

### 3. Performance Testing
1. **Bulk select**: Press `Ctrl+A` to select all 10,000 employees
2. **Navigate around**: Use `h/l` to jump through pages quickly
3. **Check responsiveness**: Notice selection state is maintained smoothly
4. **Clear large selection**: Press `c` to clear 10,000 selections instantly

### 4. Selection with Large Jumps
1. Select employee 1 (press space)
2. Jump to employee 5000 (`J` → `5000` → Enter)
3. Select employee 5000 (press space) 
4. Jump to employee 9999 (`J` → `9999` → Enter)
5. Select employee 9999 (press space)
6. Press `s` to see "3 employees selected"
7. Jump back to employee 1 - notice it's still selected

## Implementation Notes

### Selection State Management
- **ID-based tracking**: Uses employee IDs (`emp-1`, `emp-2`, etc.) for reliable selection
- **Chunk integration**: Selection state applied when chunks load
- **Memory efficient**: Only stores selected IDs, not full employee data
- **Count tracking**: Maintains running count for instant status display

### Data Source Enhancements
- **`SetSelectedByID`**: Updates selection state by employee ID
- **`GetSelectionCount`**: Returns total selected count for status display
- **`SelectAll`**: Actually selects all 10,000 employees
- **`ClearSelection`**: Instantly clears all selections

### Visual Feedback
- **✓ indicator**: Shows in first column for selected employees
- **Status display**: Shows "Selected: X" count in status bar
- **Selection info**: Dedicated `s` key for detailed selection information

## Real-World Applications

This selection pattern works well for:

- **HR Management**: Select employees for bulk operations (salary updates, transfers)
- **Project Management**: Select team members for assignments
- **Data Export**: Select specific records for reporting
- **Bulk Operations**: Select items for deletion, archiving, or processing
- **Workflow Management**: Select items for approval or review

## Key Differences from Data Virtualization

1. **SelectionMode**: Added to table configuration
2. **Selection tracking**: Enhanced data source with selection state
3. **Selection commands**: Space, Ctrl+A, c, s keys for selection operations
4. **Visual indicators**: ✓ symbols for selected items
5. **Status display**: Selection count in status bar
6. **Selection messages**: Proper handling of SelectionResponseMsg

This example demonstrates how VTable's selection system scales to large datasets while maintaining smooth performance and intuitive user experience.
