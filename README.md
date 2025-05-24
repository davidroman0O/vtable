# vtable

<p align="center">
  <img src="./demo.gif" alt="VTable Demo" width="700">
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/davidroman0O/vtable"><img src="https://pkg.go.dev/badge/github.com/davidroman0O/vtable.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/davidroman0O/vtable"><img src="https://goreportcard.com/badge/github.com/davidroman0O/vtable" alt="Go Report Card"></a>
  <a href="https://github.com/davidroman0O/vtable/blob/main/LICENSE"><img src="https://img.shields.io/github/license/davidroman0O/vtable" alt="License"></a>
  <img src="https://img.shields.io/badge/go-%3E%3D1.18-blue" alt="Go Version">
</p>

A high-performance virtualized table and list component for [Bubble Tea](https://github.com/charmbracelet/bubbletea) terminal applications. Handle millions of items efficiently through intelligent virtualization and chunked loading.

## ✨ Features

### 🚀 Virtualization
- **Memory efficient** - Only loads visible items, handles millions of records
- **Chunk-based loading** - Loads data in configurable chunks (default 20-50 items)
- **Smart caching** - Automatically manages 2-3 chunks in memory
- **Threshold scrolling** - Configurable scroll trigger points for smooth navigation

### 📊 Data Management
- **Multi-column sorting** - Sort by multiple fields with priority (SortFields, SortDirections)
- **Real-time filtering** - Apply filters with automatic data refresh
- **Dynamic updates** - Handle changing datasets with RefreshData()
- **Chunk optimization** - Configurable chunk sizes for different dataset sizes

### 🎨 Theming & Styling
- **Built-in themes** - DefaultTheme(), DarkTheme(), HighContrastTheme()
- **Border styles** - Multiple character sets (default, rounded, thick, double, ASCII)
- **Custom formatters** - Full control over item rendering with ItemFormatter
- **Animated formatters** - Delta-time animations with ItemFormatterAnimated

### 🎮 Selection & Interaction
- **Selection modes** - SelectionSingle, SelectionMultiple, SelectionNone
- **Bulk operations** - SelectAll(), ClearSelection(), GetSelectedIndices()
- **Platform keybindings** - Auto-detection for macOS, Linux, Windows
- **Custom keymaps** - NavigationKeyMap with full customization

### 🔍 Navigation & Search
- **Jump methods** - JumpToIndex(), JumpToStart(), JumpToEnd()
- **Search support** - Optional SearchableDataProvider interface
- **Navigation controls** - MoveUp(), MoveDown(), PageUp(), PageDown()
- **Viewport state** - Complete state tracking with ViewportState

### 🎬 Animation System
- **Delta-time rendering** - Frame-rate independent animations
- **Global animation loop** - Efficient centralized animation management
- **Dynamic control** - Enable/disable animations on-the-fly for performance
- **Trigger-based updates** - TriggerTimer, TriggerEvent, TriggerConditional
- **Configurable refresh rates** - SetTickInterval() for performance tuning

### 🛠️ Extensibility
- **Generic data providers** - Type-safe DataProvider[T] interface
- **Metadata system** - Rich TypedMetadata with type safety
- **Event callbacks** - OnSelect, OnHighlight, OnScroll, OnFiltersChanged, OnSortChanged
- **Component composition** - TeaTable and TeaList[T] components

## 📦 Installation

```bash
go get github.com/davidroman0O/vtable
```

## 🚀 Quick Start

### Basic Table

```go
package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/davidroman0O/vtable"
)

// 1. Define your data and implement DataProvider[vtable.TableRow]
type MyProvider struct {
	data []Person
	selection map[int]bool
}

func (p *MyProvider) GetTotal() int { return len(p.data) }
func (p *MyProvider) GetSelectionMode() vtable.SelectionMode { return vtable.SelectionNone }
// ... implement other required DataProvider methods

func (p *MyProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[vtable.TableRow], error) {
	// Return data as TableRow format
	result := make([]vtable.Data[vtable.TableRow], len(p.data))
	for i, person := range p.data {
		result[i] = vtable.Data[vtable.TableRow]{
			ID: fmt.Sprintf("person-%d", i),
			Item: vtable.TableRow{
				Cells: []string{person.Name, fmt.Sprintf("%d", person.Age)},
			},
			Metadata: vtable.NewTypedMetadata(),
		}
	}
	return result, nil
}

func main() {
	// 2. Configure table columns
	config := vtable.TableConfig{
		Columns: []vtable.TableColumn{
			{Title: "Name", Width: 20, Alignment: vtable.AlignLeft, Field: "name"},
			{Title: "Age", Width: 5, Alignment: vtable.AlignRight, Field: "age"},
		},
		ShowHeader:  true,
		ShowBorders: true,
		ViewportConfig: vtable.ViewportConfig{
			Height:               10,
			TopThresholdIndex:    2,
			BottomThresholdIndex: 7,
			ChunkSize:            50,
		},
	}

	// 3. Create table with theme
	provider := &MyProvider{data: loadPeople()}
	table, _ := vtable.NewTeaTable(config, provider, *vtable.DefaultTheme())
	
	// 4. Run
	p := tea.NewProgram(table)
	p.Run()
}
```

### Basic List

```go
// 1. Implement DataProvider[YourType]
type StringProvider struct {
	items []string
	selection map[int]bool
}

func (p *StringProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[string], error) {
	result := make([]vtable.Data[string], len(p.items))
	for i, item := range p.items {
		result[i] = vtable.Data[string]{
			ID: fmt.Sprintf("item-%d", i),
			Item: item,
			Metadata: vtable.NewTypedMetadata(),
		}
	}
	return result, nil
}
// ... implement other DataProvider methods

// 2. Create formatter
formatter := func(data vtable.Data[string], index int, ctx vtable.RenderContext, 
	isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
	prefix := "  "
	if isCursor {
		prefix = "> "
	}
	return fmt.Sprintf("%s%s", prefix, data.Item)
}

// 3. Create list
config := vtable.DefaultViewportConfig()
provider := &StringProvider{items: []string{"Item 1", "Item 2", "Item 3"}}
list, _ := vtable.NewTeaList(config, provider, vtable.DefaultStyleConfig(), formatter)

p := tea.NewProgram(list)
p.Run()
```

## 🎮 Selection & Events

### Multi-Selection

```go
// Enable in your DataProvider
func (p *MyProvider) GetSelectionMode() vtable.SelectionMode {
	return vtable.SelectionMultiple // or SelectionSingle, SelectionNone
}

// Handle in Update()
switch msg.String() {
case " ":
	table.ToggleCurrentSelection()
case "ctrl+a":
	table.SelectAll()
case "escape":
	table.ClearSelection()
}

// Check selection
selectedIndices := table.GetSelectedIndices()
selectionCount := table.GetSelectionCount()
```

### Event Callbacks

```go
// Selection events
table.OnSelect(func(row vtable.TableRow, index int) {
	fmt.Printf("Selected row %d\n", index)
})

// Navigation events  
table.OnHighlight(func(row vtable.TableRow, index int) {
	// Update preview panel
})

// Scroll events
table.OnScroll(func(state vtable.ViewportState) {
	// Update scroll indicators
})

// Data change events
table.OnFiltersChanged(func(filters map[string]any) {
	// Update filter UI
})

table.OnSortChanged(func(field, direction string) {
	// Update sort indicators
})
```

## 🎨 Theming

### Built-in Themes

```go
// Available themes
table.SetTheme(*vtable.DefaultTheme())      // Light theme
table.SetTheme(*vtable.DarkTheme())         // Dark theme  
table.SetTheme(*vtable.HighContrastTheme()) // High contrast for accessibility
```

### Border Characters

```go
// Available border styles
theme.BorderChars = vtable.DefaultBorderCharacters()  // ┌─┐│└─┘
theme.BorderChars = vtable.RoundedBorderCharacters()  // ╭─╮│╰─╯
theme.BorderChars = vtable.ThickBorderCharacters()    // ┏━┓┃┗━┛
theme.BorderChars = vtable.DoubleBorderCharacters()   // ╔═╗║╚═╝
theme.BorderChars = vtable.AsciiBoxCharacters()       // +-+|+-+
```

## 🎬 Animations

### Real-time Animations

```go
// Create animated formatter
animatedFormatter := func(data vtable.Data[Task], index int, ctx vtable.RenderContext,
	animationState map[string]any, isCursor bool, isTopThreshold bool, isBottomThreshold bool) vtable.RenderResult {
	
	// Use delta time for smooth animations
	deltaMs := ctx.DeltaTime.Milliseconds()
	
	// Animated content
	counter, _ := animationState["counter"].(int)
	spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinner := spinnerFrames[counter%len(spinnerFrames)]
	
	return vtable.RenderResult{
		Content: fmt.Sprintf("%s %s", spinner, data.Item.Title),
		RefreshTriggers: []vtable.RefreshTrigger{{
			Type: vtable.TriggerTimer,
			Interval: 100 * time.Millisecond,
		}},
		AnimationState: map[string]any{
			"counter": counter + 1,
		},
	}
}

// Enable animations
list.SetAnimatedFormatter(animatedFormatter)
list.SetTickInterval(100 * time.Millisecond) // 10fps
```

### Dynamic Animation Control

Control animations on-the-fly for performance optimization:

> **📝 Note:** Animations are **enabled by default** (`config.Enabled = true`), but the animation loop only starts when you actually use `SetAnimatedFormatter()`. If you never set an animated formatter, there's zero performance overhead.

```go
// Toggle animations during runtime
func (m MyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "a":
            // Toggle animations
            if table.IsAnimationEnabled() {
                table.DisableAnimations()
            } else {
                if cmd := table.EnableAnimations(); cmd != nil {
                    return m, cmd
                }
            }
        }
    }
    return m, nil
}

// Check animation status
isEnabled := table.IsAnimationEnabled()
isRunning := table.IsAnimationLoopRunning()
```

#### Performance Modes

```go
// Disable animations for large datasets
if dataSize > 10000 {
    table.DisableAnimations()
}

// Enable animations for real-time data
if isRealTimeData {
    if cmd := table.EnableAnimations(); cmd != nil {
        cmds = append(cmds, cmd)
    }
}

// Battery-saving mode
if lowPowerMode {
    table.DisableAnimations()
} else {
    table.SetTickInterval(50 * time.Millisecond) // Reduce frequency
}
```

#### Global Animation Control

```go
// Control all animations globally
vtable.StopGlobalAnimationLoop()
running := vtable.IsGlobalAnimationLoopRunning()

// Update global animation settings
config := vtable.DefaultAnimationConfig()
config.Enabled = false
if cmd := vtable.SetGlobalAnimationConfig(config); cmd != nil {
    return m, cmd
}
```

## 🔍 Filtering & Sorting

### Multi-Column Sorting

```go
// Single sort (replaces existing)
table.SetSort("lastName", "asc")

// Multi-sort (adds to existing)
table.AddSort("age", "desc") 

// Manage sorts
table.RemoveSort("age")
table.ClearSort()

// Check current sort
request := table.GetDataRequest()
fields := request.SortFields      // []string
directions := request.SortDirections // []string
```

### Dynamic Filtering

```go
// Apply filters
table.SetFilter("status", "active")
table.SetFilter("minAge", 18)

// Remove filters
table.RemoveFilter("status")
table.ClearFilters()

// Check current filters
request := table.GetDataRequest()
filters := request.Filters // map[string]any
```

## 🗂️ Data Provider Implementation

### Required Interface

```go
type DataProvider[T any] interface {
	GetTotal() int
	GetItems(request DataRequest) ([]Data[T], error)
	GetSelectionMode() SelectionMode
	SetSelected(index int, selected bool) bool
	SetSelectedByIDs(ids []string, selected bool) bool
	SelectRange(startID, endID string) bool
	SelectAll() bool
	ClearSelection()
	GetSelectedIndices() []int
	GetSelectedIDs() []string
	GetItemID(item *T) string
}

// Optional: For search functionality
type SearchableDataProvider[T any] interface {
	DataProvider[T]
	FindItemIndex(key string, value any) (int, bool)
}
```

### Efficient Implementation

```go
type PersonProvider struct {
	rawData      []Person
	filteredData []Person
	filters      map[string]any
	sortFields   []string
	sortDirs     []string
	selection    map[int]bool
	dirty        bool
}

func (p *PersonProvider) GetItems(request vtable.DataRequest) ([]vtable.Data[vtable.TableRow], error) {
	// Update internal state from request
	if !reflect.DeepEqual(p.filters, request.Filters) {
		p.filters = request.Filters
		p.dirty = true
	}
	
	// Rebuild filtered data if needed
	if p.dirty {
		p.rebuildFilteredData()
		p.dirty = false
	}
	
	// Return requested chunk
	start := request.Start
	count := min(request.Count, len(p.filteredData)-start)
	
	result := make([]vtable.Data[vtable.TableRow], count)
	for i := 0; i < count; i++ {
		person := p.filteredData[start+i]
		result[i] = vtable.Data[vtable.TableRow]{
			ID: fmt.Sprintf("person-%d", person.ID),
			Item: vtable.TableRow{
				Cells: []string{person.Name, fmt.Sprintf("%d", person.Age)},
			},
			Selected: p.selection[person.ID],
		}
	}
	return result, nil
}
```

## ⌨️ Keybindings

### Platform Detection

```go
// Automatic platform detection
keyMap := vtable.PlatformKeyMap() // Auto-detects macOS/Linux/Windows

// Or specify manually
keyMap := vtable.MacOSKeyMap()    // macOS optimized
keyMap := vtable.LinuxKeyMap()    // Linux optimized  
keyMap := vtable.WindowsKeyMap()  // Windows optimized

// Set custom keymap
table.SetKeyMap(keyMap)
```

## 📚 Complete API Reference

### Core Components

| Component | Description |
|-----------|-------------|
| `TeaTable` | Full table with headers, borders, sorting |
| `TeaList[T]` | Generic virtualized list component |

### Navigation Methods

| Method | Description |
|--------|-------------|
| `MoveUp()`, `MoveDown()` | Move cursor one position |
| `PageUp()`, `PageDown()` | Move cursor one page |
| `JumpToStart()`, `JumpToEnd()` | Jump to dataset boundaries |
| `JumpToIndex(int)` | Jump to specific index |
| `JumpToItem(key, value)` | Search and jump (requires SearchableDataProvider) |

### Selection Methods

| Method | Description |
|--------|-------------|
| `ToggleCurrentSelection()` | Toggle current item selection |
| `ToggleSelection(index)` | Toggle specific item selection |
| `SelectAll()` | Select all items |
| `ClearSelection()` | Clear all selections |
| `GetSelectedIndices()` | Get selected item indices |
| `GetSelectionCount()` | Get selection count |

### Data Methods

| Method | Description |
|--------|-------------|
| `SetFilter(field, value)` | Apply filter |
| `RemoveFilter(field)` | Remove filter |
| `ClearFilters()` | Clear all filters |
| `SetSort(field, direction)` | Set primary sort |
| `AddSort(field, direction)` | Add secondary sort |
| `RemoveSort(field)` | Remove sort field |
| `ClearSort()` | Clear all sorting |
| `RefreshData()` | Force data reload |

### Animation Methods

| Method | Description |
|--------|-------------|
| `SetAnimatedFormatter(formatter)` | Enable animations |
| `ClearAnimatedFormatter()` | Disable animations |
| `SetTickInterval(duration)` | Set refresh rate |
| `SetAnimationConfig(config)` | Configure animation behavior |
| `EnableAnimations()` | Enable animation system and start loop |
| `DisableAnimations()` | Disable animation system and stop loop |
| `IsAnimationEnabled()` | Check if animations are enabled |
| `IsAnimationLoopRunning()` | Check if animation loop is running |

### Event Callbacks

| Method | Description |
|--------|-------------|
| `OnSelect(func(item, index))` | Item selection callback |
| `OnHighlight(func(item, index))` | Cursor movement callback |
| `OnScroll(func(state))` | Viewport scroll callback |
| `OnFiltersChanged(func(filters))` | Filter change callback |
| `OnSortChanged(func(field, dir))` | Sort change callback |

### State & Info

| Method | Description |
|--------|-------------|
| `GetState()` | Get current ViewportState |
| `GetDataRequest()` | Get current DataRequest |
| `GetVisibleItems()` | Get currently visible items |
| `GetCurrentItem()` | Get item at cursor |

## ⚙️ Configuration

### ViewportConfig

```go
config := vtable.ViewportConfig{
	Height:               10,  // Visible rows
	TopThresholdIndex:    2,   // Top scroll trigger (0-based)
	BottomThresholdIndex: 7,   // Bottom scroll trigger
	ChunkSize:            50,  // Items per chunk
	InitialIndex:         0,   // Starting cursor position
}
```

### Animation Settings

| Use Case | Tick Interval | Performance |
|----------|---------------|-------------|
| Smooth UI | 16ms (60fps) | High CPU |
| Balanced | 50-100ms (10-20fps) | Moderate |
| Background | 250ms (4fps) | Low CPU |

#### Default Animation Behavior

```go
// Default animation configuration (animations are enabled but not running)
config := vtable.DefaultAnimationConfig()
// config.Enabled = true          // ✅ Animations are enabled by default
// config.TickInterval = 100ms    // 10fps default refresh rate
// config.MaxAnimations = 50      // Limit active animations for performance

// The animation loop only starts when you actually use animations:
// 1. Create table/list (no animation loop running yet)
table, _ := vtable.NewTeaTable(config, provider, theme)

// 2. Set animated formatter (animation loop starts automatically)
table.SetAnimatedFormatter(myAnimatedFormatter)

// 3. Clear animated formatter (animation loop stops automatically)  
table.ClearAnimatedFormatter()
```

## 📁 Examples

The `examples/` directory contains 14+ comprehensive examples:

### 🌟 Getting Started
- **`01-hello-world/`** - Basic table and list setup
- **`basic/`** - Foundation examples with core functionality

### 📊 Data Features  
- **`02-large-datasets/`** - 1M+ item virtualization demo
- **`04-filtering-sorting/`** - Multi-column sorting and filtering
- **`enhanced/`** - Advanced filtering with complex criteria
- **`10-dynamic-data/`** - Real-time data updates

### 🎮 Interaction
- **`03-selection/`** - Single and multi-selection modes
- **`05-keybindings/`** - Platform-specific key handling
- **`06-callbacks/`** - Event system demonstration
- **`07-search-jump/`** - Search and navigation features

### 🎨 Visual & Animation
- **`09-custom-formatters/`** - Rich formatting techniques
- **`animated/`** - Real-time animations and delta-time rendering

### 🌍 Real-world
- **`11-real-world-navigate-file-system/`** - Complete file browser applications

## 📄 License

[MIT License](LICENSE)

---

<p align="center">
  <a href="https://github.com/charmbracelet/bubbletea">Powered by Bubble Tea</a>
</p>