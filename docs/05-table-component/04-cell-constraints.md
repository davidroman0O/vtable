# Cell Constraints

## What We're Adding

Taking our selection-enabled employee table and adding **cell constraint controls**. We'll demonstrate column width management, text alignment options, padding configuration, and different formatting rules for cells versus headers.

## Why Cell Constraints Matter

Cell constraints let you:
- **Control column widths** for optimal data presentation
- **Align content properly** (left, center, right) based on data type
- **Add padding** for better visual spacing and readability
- **Apply different rules** to headers vs data cells
- **Handle text overflow** with truncation and ellipsis

## Step 1: Enhanced Column Configuration

Start with the selection table and add detailed cell constraints:

```go
func createEmployeeColumns() []core.TableColumn {
	return []core.TableColumn{
		{
			Title:           "Employee Name",
			Field:           "name",
			Width:           25,
			Alignment:       core.AlignLeft,   // Data cells: left aligned
			HeaderAlignment: core.AlignCenter, // Header: center aligned (different!)
			// Cell constraints for data
			CellConstraint: core.CellConstraint{
				Width:     25,
				Alignment: core.AlignLeft,
				Padding:   core.PaddingConfig{Left: 1, Right: 1}, // Add padding
			},
			// Different constraints for header
			HeaderConstraint: core.CellConstraint{
				Width:     25,
				Alignment: core.AlignCenter, // Header centered
				Padding:   core.PaddingConfig{Left: 2, Right: 2}, // More header padding
			},
		},
		{
			Title:           "Department",
			Field:           "department",
			Width:           18,
			Alignment:       core.AlignCenter, // Data: center aligned
			HeaderAlignment: core.AlignLeft,   // Header: left aligned (different!)
			CellConstraint: core.CellConstraint{
				Width:     18,
				Alignment: core.AlignCenter,
				Padding:   core.PaddingConfig{Left: 0, Right: 0}, // No padding for tight fit
			},
			HeaderConstraint: core.CellConstraint{
				Width:     18,
				Alignment: core.AlignLeft,
				Padding:   core.PaddingConfig{Left: 1, Right: 0}, // Left padding only
			},
		},
		{
			Title:           "Status",
			Field:           "status",
			Width:           15,
			Alignment:       core.AlignCenter, // Data: center aligned
			HeaderAlignment: core.AlignRight,  // Header: right aligned (different!)
			CellConstraint: core.CellConstraint{
				Width:     15,
				Alignment: core.AlignCenter,
				Padding:   core.PaddingConfig{Left: 1, Right: 1},
			},
			HeaderConstraint: core.CellConstraint{
				Width:     15,
				Alignment: core.AlignRight,
				Padding:   core.PaddingConfig{Left: 0, Right: 1}, // Right padding only
			},
		},
		{
			Title:           "Salary",
			Field:           "salary",
			Width:           12,
			Alignment:       core.AlignRight,   // Data: right aligned (numbers)
			HeaderAlignment: core.AlignCenter,  // Header: center aligned
			CellConstraint: core.CellConstraint{
				Width:     12,
				Alignment: core.AlignRight, // Right align for currency
				Padding:   core.PaddingConfig{Left: 0, Right: 1}, // Right padding for spacing
			},
			HeaderConstraint: core.CellConstraint{
				Width:     12,
				Alignment: core.AlignCenter,
				Padding:   core.PaddingConfig{Left: 1, Right: 1},
			},
		},
		{
			Title:           "Description",
			Field:           "description", 
			Width:           30, // Wider column to demonstrate truncation
			Alignment:       core.AlignLeft,
			HeaderAlignment: core.AlignLeft,
			CellConstraint: core.CellConstraint{
				Width:     30,
				Alignment: core.AlignLeft,
				Padding:   core.PaddingConfig{Left: 1, Right: 1},
				// Truncation will happen automatically for long text
			},
			HeaderConstraint: core.CellConstraint{
				Width:     30,
				Alignment: core.AlignLeft,
				Padding:   core.PaddingConfig{Left: 1, Right: 1},
			},
		},
	}
}
```

## Step 2: Enhanced Data with Long Descriptions

Update the data source to include longer descriptions that demonstrate truncation:

```go
func NewLargeEmployeeDataSource(totalCount int) *LargeEmployeeDataSource {
	// Generate ALL data upfront
	data := make([]core.TableRow, totalCount)
	
	departments := []string{"Engineering", "Marketing", "Sales", "HR", "Finance", "Operations"}
	statuses := []string{"Active", "On Leave", "Remote"}
	
	// Long descriptions to demonstrate text constraints
	longDescriptions := []string{
		"Experienced software engineer specializing in backend systems and database optimization with 5+ years",
		"Creative marketing professional focused on digital campaigns and brand management across multiple channels",
		"Senior sales representative with expertise in B2B relationships and territory management nationwide",
		"Human resources specialist handling recruitment, employee relations, and organizational development programs",
		"Financial analyst responsible for budget planning, forecasting, and quarterly reporting to executive team",
		"Operations manager overseeing logistics, supply chain optimization, and process improvement initiatives",
		"Product manager driving feature development and cross-functional collaboration with engineering teams",
		"Customer success specialist ensuring client satisfaction and managing long-term partnership relationships",
	}

	for i := 0; i < totalCount; i++ {
		data[i] = core.TableRow{
			ID: fmt.Sprintf("emp-%d", i+1),
			Cells: []string{
				fmt.Sprintf("Employee %d", i+1),
				departments[rand.Intn(len(departments))],
				statuses[rand.Intn(len(statuses))],
				fmt.Sprintf("$%d,000", 45+rand.Intn(100)),
				longDescriptions[i%len(longDescriptions)], // Long description for truncation demo
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

## Step 3: Constraint Control Commands

Add keyboard shortcuts to dynamically change constraints:

```go
func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle jump form if it's open
		if app.showJumpForm {
			// ... same jump form handling ...
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

		// Selection commands (same as before)
		case " ", "enter":
			return app, core.SelectCurrentCmd()
		case "ctrl+a":
			return app, core.SelectAllCmd()
		case "c":
			return app, core.SelectClearCmd()
		case "s":
			app.showSelectionInfo()
			return app, nil

		// NEW: Cell constraint controls
		case "w":
			// Toggle column widths (narrow → normal → wide)
			app.cycleColumnWidths()
			return app, app.updateTableColumns()
		case "a":
			// Cycle alignment modes for data cells
			app.cycleAlignment()
			return app, app.updateTableColumns()
		case "A":
			// Cycle alignment modes for headers (uppercase A)
			app.cycleHeaderAlignment()
			return app, app.updateTableColumns()
		case "p":
			// Toggle padding modes (none → normal → extra)
			app.cyclePadding()
			return app, app.updateTableColumns()
		case "t":
			// Toggle text truncation width for description column
			app.cycleDescriptionWidth()
			return app, app.updateTableColumns()

		default:
			var cmd tea.Cmd
			_, cmd = app.table.Update(msg)
			app.updateStatus()
			return app, cmd
		}

	// ... same message handling as selection example ...
	}
}
```

## Step 4: Dynamic Constraint Management

Add methods to cycle through different constraint configurations:

```go
type App struct {
	table          *table.Table
	dataSource     *LargeEmployeeDataSource
	statusMessage  string
	totalEmployees int

	// Jump-to-index form
	showJumpForm bool
	jumpInput    string

	// Constraint state tracking
	widthMode      int // 0=narrow, 1=normal, 2=wide
	alignmentMode  int // 0=mixed, 1=all-left, 2=all-center, 3=all-right
	headerAlignMode int // 0=mixed, 1=all-left, 2=all-center, 3=all-right  
	paddingMode    int // 0=none, 1=normal, 2=extra
	descriptionWidth int // Cycle through different widths
}

func (app *App) cycleColumnWidths() {
	app.widthMode = (app.widthMode + 1) % 3
	switch app.widthMode {
	case 0:
		app.statusMessage = "Column widths: NARROW (compact layout)"
	case 1:
		app.statusMessage = "Column widths: NORMAL (balanced layout)"
	case 2:
		app.statusMessage = "Column widths: WIDE (spacious layout)"
	}
}

func (app *App) cycleAlignment() {
	app.alignmentMode = (app.alignmentMode + 1) % 4
	switch app.alignmentMode {
	case 0:
		app.statusMessage = "Data alignment: MIXED (name=left, dept=center, status=center, salary=right)"
	case 1:
		app.statusMessage = "Data alignment: ALL LEFT"
	case 2:
		app.statusMessage = "Data alignment: ALL CENTER"
	case 3:
		app.statusMessage = "Data alignment: ALL RIGHT"
	}
}

func (app *App) cycleHeaderAlignment() {
	app.headerAlignMode = (app.headerAlignMode + 1) % 4
	switch app.headerAlignMode {
	case 0:
		app.statusMessage = "Header alignment: MIXED (different from data alignment)"
	case 1:
		app.statusMessage = "Header alignment: ALL LEFT"
	case 2:
		app.statusMessage = "Header alignment: ALL CENTER"
	case 3:
		app.statusMessage = "Header alignment: ALL RIGHT"
	}
}

func (app *App) cyclePadding() {
	app.paddingMode = (app.paddingMode + 1) % 3
	switch app.paddingMode {
	case 0:
		app.statusMessage = "Padding: NONE (tight layout)"
	case 1:
		app.statusMessage = "Padding: NORMAL (comfortable spacing)"
	case 2:
		app.statusMessage = "Padding: EXTRA (spacious layout)"
	}
}

func (app *App) cycleDescriptionWidth() {
	widths := []int{20, 30, 40, 50}
	app.descriptionWidth = (app.descriptionWidth + 1) % len(widths)
	app.statusMessage = fmt.Sprintf("Description width: %d characters (see truncation effect)", widths[app.descriptionWidth])
}
```

## Step 5: Dynamic Column Updates

Implement the method to apply constraint changes:

```go
func (app *App) updateTableColumns() tea.Cmd {
	columns := app.buildColumnsWithConstraints()
	return core.ColumnSetCmd(columns)
}

func (app *App) buildColumnsWithConstraints() []core.TableColumn {
	// Base widths for each mode
	var nameWidth, deptWidth, statusWidth, salaryWidth, descWidth int
	
	switch app.widthMode {
	case 0: // Narrow
		nameWidth, deptWidth, statusWidth, salaryWidth = 15, 12, 10, 10
	case 1: // Normal
		nameWidth, deptWidth, statusWidth, salaryWidth = 20, 15, 12, 12
	case 2: // Wide
		nameWidth, deptWidth, statusWidth, salaryWidth = 25, 20, 15, 15
	}
	
	// Description width based on cycle
	descWidths := []int{20, 30, 40, 50}
	descWidth = descWidths[app.descriptionWidth]

	// Alignment based on mode
	var dataAlign, headerAlign core.Alignment
	
	// Data alignment
	switch app.alignmentMode {
	case 0: // Mixed
		dataAlign = core.AlignLeft // Will be overridden per column
	case 1: // All left
		dataAlign = core.AlignLeft
	case 2: // All center
		dataAlign = core.AlignCenter
	case 3: // All right
		dataAlign = core.AlignRight
	}

	// Header alignment
	switch app.headerAlignMode {
	case 0: // Mixed
		headerAlign = core.AlignCenter // Will be overridden per column
	case 1: // All left
		headerAlign = core.AlignLeft
	case 2: // All center
		headerAlign = core.AlignCenter
	case 3: // All right
		headerAlign = core.AlignRight
	}

	// Padding based on mode
	var leftPad, rightPad int
	switch app.paddingMode {
	case 0: // None
		leftPad, rightPad = 0, 0
	case 1: // Normal
		leftPad, rightPad = 1, 1
	case 2: // Extra
		leftPad, rightPad = 2, 2
	}

	columns := []core.TableColumn{
		{
			Title:           "Employee Name",
			Field:           "name",
			Width:           nameWidth,
			Alignment:       getColumnAlignment(app.alignmentMode, core.AlignLeft),
			HeaderAlignment: getColumnAlignment(app.headerAlignMode, core.AlignCenter),
			CellConstraint: core.CellConstraint{
				Width:     nameWidth,
				Alignment: getColumnAlignment(app.alignmentMode, core.AlignLeft),
				Padding:   core.PaddingConfig{Left: leftPad, Right: rightPad},
			},
			HeaderConstraint: core.CellConstraint{
				Width:     nameWidth,
				Alignment: getColumnAlignment(app.headerAlignMode, core.AlignCenter),
				Padding:   core.PaddingConfig{Left: leftPad + 1, Right: rightPad + 1}, // Extra header padding
			},
		},
		{
			Title:           "Department",
			Field:           "department",
			Width:           deptWidth,
			Alignment:       getColumnAlignment(app.alignmentMode, core.AlignCenter),
			HeaderAlignment: getColumnAlignment(app.headerAlignMode, core.AlignLeft),
			CellConstraint: core.CellConstraint{
				Width:     deptWidth,
				Alignment: getColumnAlignment(app.alignmentMode, core.AlignCenter),
				Padding:   core.PaddingConfig{Left: leftPad, Right: rightPad},
			},
			HeaderConstraint: core.CellConstraint{
				Width:     deptWidth,
				Alignment: getColumnAlignment(app.headerAlignMode, core.AlignLeft),
				Padding:   core.PaddingConfig{Left: leftPad, Right: rightPad},
			},
		},
		{
			Title:           "Status",
			Field:           "status",
			Width:           statusWidth,
			Alignment:       getColumnAlignment(app.alignmentMode, core.AlignCenter),
			HeaderAlignment: getColumnAlignment(app.headerAlignMode, core.AlignRight),
			CellConstraint: core.CellConstraint{
				Width:     statusWidth,
				Alignment: getColumnAlignment(app.alignmentMode, core.AlignCenter),
				Padding:   core.PaddingConfig{Left: leftPad, Right: rightPad},
			},
			HeaderConstraint: core.CellConstraint{
				Width:     statusWidth,
				Alignment: getColumnAlignment(app.headerAlignMode, core.AlignRight),
				Padding:   core.PaddingConfig{Left: leftPad, Right: rightPad},
			},
		},
		{
			Title:           "Salary",
			Field:           "salary",
			Width:           salaryWidth,
			Alignment:       getColumnAlignment(app.alignmentMode, core.AlignRight),
			HeaderAlignment: getColumnAlignment(app.headerAlignMode, core.AlignCenter),
			CellConstraint: core.CellConstraint{
				Width:     salaryWidth,
				Alignment: getColumnAlignment(app.alignmentMode, core.AlignRight),
				Padding:   core.PaddingConfig{Left: leftPad, Right: rightPad},
			},
			HeaderConstraint: core.CellConstraint{
				Width:     salaryWidth,
				Alignment: getColumnAlignment(app.headerAlignMode, core.AlignCenter),
				Padding:   core.PaddingConfig{Left: leftPad, Right: rightPad},
			},
		},
		{
			Title:           "Description",
			Field:           "description",
			Width:           descWidth,
			Alignment:       getColumnAlignment(app.alignmentMode, core.AlignLeft),
			HeaderAlignment: getColumnAlignment(app.headerAlignMode, core.AlignLeft),
			CellConstraint: core.CellConstraint{
				Width:     descWidth,
				Alignment: getColumnAlignment(app.alignmentMode, core.AlignLeft),
				Padding:   core.PaddingConfig{Left: leftPad, Right: rightPad},
			},
			HeaderConstraint: core.CellConstraint{
				Width:     descWidth,
				Alignment: getColumnAlignment(app.headerAlignMode, core.AlignLeft),
				Padding:   core.PaddingConfig{Left: leftPad, Right: rightPad},
			},
		},
	}

	return columns
}

func getColumnAlignment(mode int, defaultAlign core.Alignment) core.Alignment {
	switch mode {
	case 0: // Mixed - use default
		return defaultAlign
	case 1: // All left
		return core.AlignLeft
	case 2: // All center
		return core.AlignCenter
	case 3: // All right
		return core.AlignRight
	default:
		return defaultAlign
	}
}
```

## What You'll See

**Default view with mixed alignments and normal widths:**
```
Employee 1/10000 | Selected: 0 | Use space/enter ctrl+a c s w a A p t J, q to quit

│ ●  │   Employee Name    │   Department   │    Status    │   Salary   │             Description              │
│ ►  │ Employee 1         │   Engineering  │    Active    │  $67,000   │ Experienced software engineer spe...│
│    │ Employee 2         │   Marketing    │   On Leave   │  $58,000   │ Creative marketing professional f...│
│ ✓  │ Employee 3         │     Sales      │    Remote    │  $73,000   │ Senior sales representative with...│
│    │ Employee 4         │      HR        │    Active    │  $51,000   │ Human resources specialist handl...│
```

**After pressing 'w' (wide columns):**
```
Column widths: WIDE (spacious layout)

│ ●  │      Employee Name       │      Department      │     Status     │    Salary    │                   Description                    │
│ ►  │ Employee 1               │     Engineering      │     Active     │   $67,000    │ Experienced software engineer specializing in...│
│    │ Employee 2               │     Marketing        │    On Leave    │   $58,000    │ Creative marketing professional focused on di...│
│ ✓  │ Employee 3               │       Sales          │     Remote     │   $73,000    │ Senior sales representative with expertise in...│
```

**After pressing 'a' (all center alignment):**
```
Data alignment: ALL CENTER

│ ●  │     Employee Name        │      Department      │     Status     │    Salary    │                   Description                    │
│ ►  │      Employee 1          │     Engineering      │     Active     │   $67,000    │ Experienced software engineer specializing in...│
│    │      Employee 2          │     Marketing        │    On Leave    │   $58,000    │ Creative marketing professional focused on di...│
│ ✓  │      Employee 3          │       Sales          │     Remote     │   $73,000    │ Senior sales representative with expertise in...│
```

**After pressing 'p' twice (extra padding):**
```
Padding: EXTRA (spacious layout)

│ ●  │       Employee Name        │        Department        │       Status       │      Salary      │                     Description                       │
│ ►  │   Employee 1               │     Engineering          │       Active       │     $67,000      │   Experienced software engineer specializing in...   │
```

**After pressing 't' (narrow description):**
```
Description width: 20 characters (see truncation effect)

│ ●  │     Employee Name        │      Department      │     Status     │    Salary    │    Description       │
│ ►  │      Employee 1          │     Engineering      │     Active     │   $67,000    │ Experienced soft...  │
│    │      Employee 2          │     Marketing        │    On Leave    │   $58,000    │ Creative marketi...  │
```

## Alignment Options

### Cell Alignment
```go
Alignment: core.AlignLeft,   // Text starts from left edge
Alignment: core.AlignCenter, // Text centered in column
Alignment: core.AlignRight,  // Text aligned to right edge
```

### Independent Header Alignment
```go
HeaderAlignment: core.AlignCenter, // Header can be different from data
```

### Practical Alignment Guidelines
- **Names/Text**: Usually `AlignLeft`
- **Numbers/Currency**: Usually `AlignRight` 
- **Status/Categories**: Usually `AlignCenter`
- **Headers**: Often `AlignCenter` regardless of data alignment

## Padding Configuration

### Basic Padding
```go
Padding: core.PaddingConfig{Left: 1, Right: 1}, // 1 space on each side
```

### Asymmetric Padding
```go
Padding: core.PaddingConfig{Left: 2, Right: 0}, // More space on left only
```

### No Padding (Tight Layout)
```go
Padding: core.PaddingConfig{Left: 0, Right: 0}, // No extra spacing
```

## Width and Truncation

### Fixed Width with Auto-Truncation
```go
Width: 20, // Text longer than 20 chars gets "..." truncation
```

### Dynamic Width Control
- Change column widths at runtime using `core.ColumnSetCmd(newColumns)`
- Truncation happens automatically when text exceeds width
- Ellipsis ("...") added to truncated text

## Key Controls

| Key | Action |
|-----|--------|
| `w` | Cycle column widths (narrow → normal → wide) |
| `a` | Cycle data alignment (mixed → left → center → right) |
| `A` | Cycle header alignment (mixed → left → center → right) |
| `p` | Cycle padding (none → normal → extra) |
| `t` | Cycle description width (20 → 30 → 40 → 50 chars) |
| `Space` `Enter` | Toggle selection (same as before) |
| `Ctrl+A` | Select all (same as before) |
| `c` | Clear selections (same as before) |
| `s` | Show selection info (same as before) |
| `J` | Jump to employee (same as before) |

## Key Improvements from Table Selection

1. **Column width control** - dynamic width adjustment at runtime
2. **Flexible alignment** - different alignment for data vs headers
3. **Padding management** - configurable spacing within cells
4. **Text truncation** - automatic ellipsis for long content
5. **Constraint separation** - different rules for cells vs headers
6. **Interactive testing** - keyboard shortcuts to try all options

## Try It

1. **Test width modes**: Press `w` repeatedly to see narrow → normal → wide
2. **Try alignment**: Press `a` to cycle data alignment, `A` for headers
3. **Experiment with padding**: Press `p` to see tight → normal → spacious
4. **Watch truncation**: Press `t` to see description column width changes
5. **Mix and match**: Try wide columns + center alignment + extra padding
6. **Compare modes**: Notice how headers can align differently from data

## Constraint Best Practices

- **Numbers**: Right-align for easy comparison
- **Text**: Left-align for natural reading
- **Status**: Center-align for visual balance
- **Headers**: Often center-aligned regardless of data
- **Padding**: Add space for readability, but not too much
- **Width**: Size columns based on typical content length

## What's Next

The [column formatting](05-column-formatting.md) section shows how to add custom cell formatters, color coding, and advanced text processing.

## Key Point

**Cell constraints give you precise control over table layout** - you can make columns exactly the width you need, align content appropriately for the data type, add padding for visual breathing room, and apply different formatting rules to headers versus data cells. The table automatically handles text truncation when content exceeds column width. 