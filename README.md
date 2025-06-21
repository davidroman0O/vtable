<!-- <p align="center">
  <img src="https://github.com/davidroman/vtable/blob/main/assets/demo.gif?raw=true" alt="VTable Demo" width="800">
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/davidroman/vtable/pure"><img src="https://pkg.go.dev/badge/github.com/davidroman/vtable/pure.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/davidroman/vtable/pure"><img src="https://goreportcard.com/badge/github.com/davidroman/vtable/pure" alt="Go Report Card"></a>
  <a href="https://github.com/davidroman/vtable/blob/main/LICENSE"><img src="https://img.shields.io/github/license/davidroman/vtable" alt="License"></a>
  <img src="https://img.shields.io/badge/go-%3E%3D1.18-blue" alt="Go Version">
</p>

# VTable

A high-performance, feature-rich library of virtualized `List`, `Table`, and `Tree` components for [Bubble Tea](https://github.com/charmbracelet/bubbletea). It's designed to handle millions of items efficiently through intelligent virtualization and asynchronous, chunk-based data loading, while offering extensive customization options.

## ✨ Features

### 🚀 Core Engine
- **High Performance**: Built for speed, using viewport virtualization (`Chunking`) to handle massive datasets without breaking a sweat.
- **Asynchronous by Design**: All data operations (loading, sorting, filtering) are non-blocking and handled via `tea.Cmd`s, ensuring a perfectly responsive UI.
- **Pure Go & Bubble Tea Native**: Implemented as standard `tea.Model`s for seamless integration into any Bubble Tea application.
- **Stateful & Predictable**: Manages its own internal state, updated immutably through messages, making it easy to reason about.

### 📦 Components
- **`List`**: A powerful and customizable vertical list for homogenous items.
- **`Table`**: A multi-column table with headers, borders, advanced formatters, and multiple horizontal scrolling modes.
- **`TreeList`**: A hierarchical list for displaying tree-like data structures with node expansion and collapsing.

### 🎨 Rendering & Styling
- **Component-Based Rendering**: A highly flexible rendering pipeline. Build custom item/row layouts by assembling components like `Cursor`, `Enumerator`, `Content`, `Background`, and more.
- **Advanced Formatters**: Full control over item/cell rendering with simple (`ItemFormatter`) or animated (`ItemFormatterAnimated`) formatters.
- **Extensive Theming**: Easily configure `lipgloss` styles for every part of your component, from cursor and selection to borders and alternating rows.
- **Granular Border Control**: Independently control visibility of top, bottom, and header-separator borders in tables.
- **Advanced Horizontal Scrolling**: Sophisticated per-cell or global scrolling for tables with `character`, `word`, and `smart` modes.
- **Active Cell Indication**: Highlight the currently active cell in a table with a background color or custom formatter logic.

### 📊 Data Management
- **Asynchronous `DataSource`**: Your data source is completely decoupled from the UI. The components request data chunks as needed via commands.
- **Multi-Column Sorting**: Sort by multiple fields with priority.
- **Dynamic Filtering**: Apply complex filters to your data on the fly.
- **Selection Modes**: Supports `SelectionSingle`, `SelectionMultiple`, and `SelectionNone`.

## 📦 Installation

```bash
go get github.com/davidroman/vtable/pure
```

## 🚀 Quick Start

`vtable` components are `tea.Model`s. You embed one in your own model and delegate `Update` calls to it. Interaction is done by sending `tea.Msg`s, which are created by `vtable`'s command functions (e.g., `vtable.CursorUpCmd()`).

### Basic Table Example

```go
package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman/vtable/pure"
)

// 1. Define your data source
type MyDataSource struct {
	items [][]string
}

func NewMyDataSource() *MyDataSource {
	// Let's create a large dataset
	items := make([][]string, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = []string{fmt.Sprintf("Name %d", i+1), fmt.Sprintf("%d", 20+i%50), "Active"}
	}
	return &MyDataSource{items: items}
}

// GetTotal returns the total number of items
func (ds *MyDataSource) GetTotal() tea.Cmd {
	return func() tea.Msg {
		return vtable.DataTotalMsg{Total: len(ds.items)}
	}
}

// LoadChunk loads a slice of data asynchronously
func (ds *MyDataSource) LoadChunk(request vtable.DataRequest) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(50 * time.Millisecond) // Simulate latency
		end := request.Start + request.Count
		if end > len(ds.items) {
			end = len(ds.items)
		}
		if request.Start >= end {
			return vtable.DataChunkLoadedMsg{Items: []vtable.Data[any]{}}
		}
		
		chunkItems := make([]vtable.Data[any], end-request.Start)
		for i := request.Start; i < end; i++ {
			chunkItems[i-request.Start] = vtable.Data[any]{
				ID: fmt.Sprintf("row-%d", i),
				Item: vtable.TableRow{ // For tables, the item is a TableRow
					ID:    fmt.Sprintf("row-%d", i),
					Cells: ds.items[i],
				},
			}
		}
		return vtable.DataChunkLoadedMsg{
			StartIndex: request.Start,
			Items:      chunkItems,
			Request:    request,
		}
	}
}

// Implement other DataSource methods (stubs for this example)
func (ds *MyDataSource) RefreshTotal() tea.Cmd { return ds.GetTotal() }
func (ds *MyDataSource) SetSelected(index int, selected bool) tea.Cmd { return nil }
func (ds *MyDataSource) SetSelectedByID(id string, selected bool) tea.Cmd { return nil }
func (ds *MyDataSource) SelectAll() tea.Cmd { return nil }
func (ds *MyDataSource) ClearSelection() tea.Cmd { return nil }
func (ds *MyDataSource) SelectRange(startID, endID string) tea.Cmd { return nil }
func (ds *MyDataSource) GetItemID(item any) string {
	if row, ok := item.(vtable.TableRow); ok { return row.ID }
	return ""
}

// 2. Define your application model
type model struct {
	table *vtable.Table
}

func initialModel() model {
	ds := NewMyDataSource()
	
	// Define columns for the table
	columns := []vtable.TableColumn{
		{Title: "Name", Width: 20, Field: "name"},
		{Title: "Age", Width: 10, Field: "age"},
		{Title: "Status", Width: 15, Field: "status"},
	}

	config := vtable.DefaultTableConfig()
	config.Columns = columns
	config.ViewportConfig.Height = 20
	config.ShowHeader = true
	config.ShowBorders = true
	
	table := vtable.NewTable(config, ds)

	return model{table: table}
}

func (m model) Init() tea.Cmd {
	return m.table.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle top-level keys
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	// Delegate all other messages to the table component
	var tableCmd tea.Cmd
	newTableModel, tableCmd := m.table.Update(msg)
	m.table = newTableModel.(*vtable.Table)
	
	return m, tableCmd
}

func (m model) View() string {
	return m.table.View()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
```

## 📚 Core Concepts

### The `DataSource` Interface

The `DataSource` is the heart of `vtable`. It's an interface you implement to provide data to the components. This design decouples your data source from the UI, allowing `vtable` to request only the data it needs to display. All methods return a `tea.Cmd`, embracing Bubble Tea's asynchronous architecture.

```go
type DataSource[T any] interface {
	LoadChunk(request DataRequest) tea.Cmd
	GetTotal() tea.Cmd
	RefreshTotal() tea.Cmd
	SetSelected(index int, selected bool) tea.Cmd
	SetSelectedByID(id string, selected bool) tea.Cmd
	SelectAll() tea.Cmd
	ClearSelection() tea.Cmd
	SelectRange(startID, endID string) tea.Cmd
	GetItemID(item T) string
}
```

### Viewport and Virtualization

`vtable` only renders the items that are currently visible in the "viewport". When you scroll, it calculates which new items should become visible and requests the corresponding data chunks from your `DataSource` by sending a `DataRequest`. This process, known as virtualization or windowing, is the key to its performance.

### Command & Message-Based API

You don't call methods like `list.CursorUp()` directly. Instead, you use command constructors that create a `tea.Cmd`. This command, when executed by the Bubble Tea runtime, produces a message that the component's `Update` function will handle.

**Example Interaction Flow:**
1. User presses the "up" arrow (`tea.KeyMsg`).
2. Your `Update` function maps this key to `vtable.CursorUpCmd()`.
3. You return this `Cmd` to the Bubble Tea runtime.
4. The runtime executes the `Cmd`, which produces a `vtable.CursorUpMsg{}`.
5. This message is passed back to your `Update` function.
6. You delegate the message to `list.Update(msg)` or `table.Update(msg)`.
7. The component updates its internal state (e.g., changes `CursorIndex`) and returns.

## 🎨 Component-Based Rendering

For ultimate flexibility, `vtable` includes a powerful component-based rendering system for both `List` and `Table`. Instead of a single formatter function, you can define a pipeline of render components and control their order and configuration. This allows you to build complex and dynamic layouts.

**Available `List` Components:**
- `CursorComponent`: Renders the cursor indicator.
- `SpacingComponent`: Renders spacing.
- `EnumeratorComponent`: Renders a bullet, number, etc.
- `ContentComponent`: Renders the main item content.
- `BackgroundComponent`: Renders a background style across other components.

**Available `Table` Components:**
- `TableCursorComponent`: Renders the row cursor.
- `TableRowNumberComponent`: Renders a row number.
- `TableSelectionMarkerComponent`: Renders a selection checkbox/marker.
- `TableCellsComponent`: Renders the main row of cells.
- `TableBorderComponent`: Renders left/right borders for the row.
- `TableBackgroundComponent`: Renders a background for the row.

**Example: Creating a custom checklist-style list renderer:**

```go
// 1. Get a default render config
renderConfig := vtable.DefaultListRenderConfig()

// 2. Customize the enumerator to show a checkbox
renderConfig.EnumeratorConfig.Enumerator = vtable.CheckboxEnumerator
renderConfig.EnumeratorConfig.Enabled = true

// 3. Customize the cursor
renderConfig.CursorConfig.CursorIndicator = " 👉 "

// 4. Define the render order
renderConfig.ComponentOrder = []vtable.ListComponentType{
	vtable.ListComponentCursor,
	vtable.ListComponentEnumerator,
	vtable.ListComponentContent,
}

// 5. Use a component-based formatter in your list constructor
formatter := vtable.ComponentBasedListFormatter(renderConfig)
list := vtable.NewList(config, ds, formatter)
```
The `enhanced-list` example provides a deep dive into this system.

## ⌨️ API Reference

Interaction with components is primarily through commands and messages.

### Common Commands (`pure/commands.go`)

A selection of available commands. See the file for a complete list.

| Command Constructor | Returns `tea.Cmd` that produces... | Description |
|---|---|---|
| `CursorUpCmd()` | `CursorUpMsg` | Move cursor up by one. |
| `CursorDownCmd()` | `CursorDownMsg` | Move cursor down by one. |
| `PageUpCmd()` | `PageUpMsg` | Move one viewport page up. |
| `PageDownCmd()` | `PageDownMsg` | Move one viewport page down. |
| `JumpToStartCmd()` | `JumpToStartMsg` | Jump to the first item. |
| `JumpToEndCmd()` | `JumpToEndMsg` | Jump to the last item. |
| `JumpToCmd(index)` | `JumpToMsg` | Jump to a specific absolute index. |
| `SelectCurrentCmd()` | `SelectCurrentMsg` | Toggle selection for the item at the cursor. |
| `SelectAllCmd()` | `SelectAllMsg` | Select all items (requires `DataSource` support). |
| `SelectClearCmd()` | `SelectClearMsg` | Clear all selections. |
| `FilterSetCmd(field, value)`| `FilterSetMsg` | Set a filter, triggering a data refresh. |
| `SortToggleCmd(field)` | `SortToggleMsg` | Toggle sorting for a field. |
| `ViewportResizeCmd(w, h)`| `ViewportResizeMsg` | Inform the component of a terminal resize. |
| `HeaderVisibilityCmd(visible)`| `HeaderVisibilityMsg` | Show/hide the table header. |
| `BorderVisibilityCmd(visible)`| `BorderVisibilityMsg` | Toggle all table borders. |
| `TopBorderVisibilityCmd(v)`| `TopBorderVisibilityMsg` | Toggle top table border. |
| `BottomBorderVisibilityCmd(v)`| `BottomBorderVisibilityMsg` | Toggle bottom table border. |

### Configuration (`pure/config.go`, `pure/types.go`)

Configuration is handled by passing a `ListConfig`, `TableConfig`, or `TreeConfig` struct to the constructor. A fluent builder API is also available (`NewListConfigBuilder`, `NewTableConfigBuilder`).

#### `ViewportConfig`

```go
type ViewportConfig struct {
	Height          int // Number of items visible in the viewport
	TopThreshold    int // Offset from top where scrolling up triggers
	BottomThreshold int // Offset from bottom where scrolling down triggers
	ChunkSize       int // Number of items to load in each chunk
	InitialIndex    int // Starting cursor position
}
```

#### `TableConfig` Highlights

```go
type TableConfig struct {
	Columns                 []TableColumn
	ShowHeader              bool
	ShowBorders             bool // Global border control

	// Granular border configuration
	ShowTopBorder           bool // Control top border independently
	ShowBottomBorder        bool // Control bottom border independently
	ShowHeaderSeparator     bool // Control header separator border independently

	// Space removal for borders (when true, completely removes the line space)
	RemoveTopBorderSpace    bool
	RemoveBottomBorderSpace bool

	// Highlighting configuration
	FullRowHighlighting bool // Enable full row highlighting mode

	// Horizontal scrolling configuration
	ResetScrollOnNavigation bool // Reset scroll when navigating between rows

	// Active cell indication settings
	ActiveCellIndicationEnabled bool   // Enable/disable active cell background
	ActiveCellBackgroundColor   string // Background color for active cell
    
	// And much more...
	ViewportConfig ViewportConfig
	Theme          Theme
	SelectionMode  SelectionMode
	KeyMap         NavigationKeyMap
}
```

## 📁 Examples

The `pure/examples/` directory contains comprehensive examples that are the best resource for learning advanced usage.

- **`basic-list`**: A simple list with asynchronous data loading.
- **`basic-table`**: A comprehensive demo of the `Table` component's features, including formatters, themes, sorting, filtering, border controls, and active cell indication.
- **`basic-tree-list`**: An example of the `TreeList` component for hierarchical data.
- **`enhanced-list`**: A deep dive into the component-based rendering system for lists, showing how to build completely custom item layouts.

## 📄 License

[MIT License](LICENSE)

---

<p align="center">
  <a href="https://github.com/charmbracelet/bubbletea">Built with Bubble Tea</a>
</p>

# Cell Constraints Example

A comprehensive demonstration of VTable's cell constraint system, building on the selection table example with advanced column layout controls.

## Features Demonstrated

- **Dynamic column width control** - narrow/normal/wide modes
- **Flexible alignment options** - separate alignment for data vs headers
- **Padding configuration** - none/normal/extra padding modes  
- **Text truncation** - automatic ellipsis for long content
- **Header constraints** - different formatting rules for headers vs data
- **Interactive testing** - keyboard shortcuts to test all constraint options

## Running the Example

```bash
cd docs/05-table-component/examples/cell-constraints
go run .
```

## Controls

### Constraint Controls
| Key | Action |
|-----|--------|
| `w` | Cycle column widths (narrow → normal → wide) |
| `a` | Cycle data alignment (mixed → left → center → right) |
| `A` | Cycle header alignment (mixed → left → center → right) |
| `p` | Cycle padding (none → normal → extra) |
| `t` | Cycle description width (20 → 30 → 40 → 50 chars) |

### Selection Controls (inherited from selection example)
| Key | Action |
|-----|--------|
| `Space` `Enter` | Toggle selection of current employee |
| `Ctrl+A` | Select all employees |
| `c` | Clear all selections |
| `s` | Show selection information |
| `J` | Jump to specific employee |

### Navigation (inherited from previous examples)
| Key | Action |
|-----|--------|
| `↑` `k` | Move up one row |
| `↓` `j` | Move down one row |
| `g` | Jump to first employee |
| `G` | Jump to last employee |
| `h` `PgUp` | Jump up rows |
| `l` `PgDn` | Jump down rows |
| `q` | Quit |

## What You'll See

**Default view with mixed alignments:**
```
Employee 1/1000 | Selected: 0 | Use space/enter ctrl+a c s w a A p t J, q to quit

│ ●  │   Employee Name    │   Department   │    Status    │   Salary   │             Description              │
│ ►  │ Employee 1         │   Engineering  │    Active    │  $67,000   │ Experienced software engineer spe...│
│    │ Employee 2         │   Marketing    │   On Leave   │  $58,000   │ Creative marketing professional f...│

Constraints: Width=normal | Data=mixed | Header=mixed | Padding=normal | Desc=30ch
```

**After pressing 'w' (wide columns):**
```
│ ●  │      Employee Name       │      Department      │     Status     │    Salary    │                   Description                    │
│ ►  │ Employee 1               │     Engineering      │     Active     │   $67,000    │ Experienced software engineer specializing in...│

Constraints: Width=wide | Data=mixed | Header=mixed | Padding=normal | Desc=30ch
```

**After pressing 'a' (all center alignment):**
```
│ ●  │     Employee Name        │      Department      │     Status     │    Salary    │                   Description                    │
│ ►  │      Employee 1          │     Engineering      │     Active     │   $67,000    │ Experienced software engineer specializing in...│

Constraints: Width=wide | Data=center | Header=mixed | Padding=normal | Desc=30ch
```

## Key Features

### Column Width Control
- **Narrow**: Compact layout (15/12/10/10/20 chars)
- **Normal**: Balanced layout (20/15/12/12/30 chars)  
- **Wide**: Spacious layout (25/20/15/15/40 chars)

### Alignment Options
- **Mixed**: Different alignment per column type (name=left, numbers=right, etc.)
- **All Left**: Everything left-aligned
- **All Center**: Everything center-aligned  
- **All Right**: Everything right-aligned

### Separate Header Alignment
- Headers can have different alignment from data cells
- Demonstrates independent header constraint system

### Padding Modes
- **None**: Tight layout with no padding
- **Normal**: Comfortable 1-space padding
- **Extra**: Spacious 2-space padding

### Text Truncation
- Description column width cycles: 20 → 30 → 40 → 50 characters
- Automatic ellipsis ("...") for content exceeding width
- Demonstrates text constraint handling

## Implementation Notes

### Cell vs Header Constraints
- **Cell alignment**: Controlled by `Alignment` field
- **Header alignment**: Controlled by `HeaderAlignment` field  
- **Header constraints**: Defined in `HeaderConstraint` field with separate padding/alignment

### Dynamic Column Updates
- Uses `core.ColumnSetCmd()` to update table structure at runtime
- Maintains selection state across constraint changes
- Smooth transitions between different layouts

### Progressive Enhancement
- Builds on selection table example
- Adds constraint control layer
- Preserves all selection functionality

This example demonstrates how VTable's constraint system provides precise control over table layout while maintaining smooth performance and user experience.  -->