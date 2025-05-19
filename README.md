# VTable - A Virtualized Table and List Component for Bubble Tea

VTable is a Go library providing high-performance virtualized table and list components for terminal user interfaces built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- Efficient rendering of large datasets through virtualization
- Support for tables with customizable columns, headers, and borders
- Virtualized lists with flexible data binding
- Rich customization API - change appearance and behavior at runtime
- Event callback system for responding to user interactions
- Keyboard navigation with customizable keymaps
- Platform-specific key bindings
- Search and jump functionality
- Built-in pagination and scrolling
- Cursor thresholds for efficient scrolling

## Installation

```bash
go get github.com/davidroman0O/vtable
```

## Basic Usage

### Creating a List

```go
// Create a data provider
provider := &MyDataProvider{}

// Configure the viewport
config := vtable.ViewportConfig{
    Height:               10,
    TopThresholdIndex:    2,
    BottomThresholdIndex: 7,
    ChunkSize:            20,
    InitialIndex:         0,
}

// Define styling
theme := vtable.DefaultTheme()
styleConfig := vtable.ThemeToStyleConfig(theme)

// Define how items are formatted
formatter := func(item MyItem, index int, isCursor bool, isTopThreshold bool, isBottomThreshold bool) string {
    // Format your item here
    return fmt.Sprintf("%d: %s", index, item.Name)
}

// Create the Bubble Tea component
listModel, err := vtable.NewTeaList(config, provider, styleConfig, formatter)
if err != nil {
    // Handle error
}

// Set up event handlers
listModel.OnSelect(func(item MyItem, index int) {
    fmt.Printf("Selected item: %s at index %d\n", item.Name, index)
})

listModel.OnHighlight(func(item MyItem, index int) {
    fmt.Printf("Highlighted item: %s at index %d\n", item.Name, index)
})

// Use in your Bubble Tea application
program := tea.NewProgram(listModel)
if _, err := program.Run(); err != nil {
    // Handle error
}
```

### Creating a Table

```go
// Create a data provider
provider := &MyTableProvider{}

// Define columns
tableConfig := vtable.TableConfig{
    Columns: []vtable.TableColumn{
        {Title: "ID", Width: 10, Alignment: vtable.AlignLeft},
        {Title: "Name", Width: 20, Alignment: vtable.AlignLeft},
        {Title: "Value", Width: 15, Alignment: vtable.AlignRight},
    },
    ShowHeader:     true,
    ShowBorders:    true,
    ViewportConfig: viewportConfig,
}

// Use a theme
theme := vtable.ColorfulTheme()

// Create the Bubble Tea component
tableModel, err := vtable.NewTeaTable(tableConfig, provider, theme)
if err != nil {
    // Handle error
}

// Set up event handlers
tableModel.OnSelect(func(row vtable.TableRow, index int) {
    fmt.Printf("Selected row at index %d\n", index)
})

// Use in your Bubble Tea application
program := tea.NewProgram(tableModel)
if _, err := program.Run(); err != nil {
    // Handle error
}
```

## Customization

### Runtime Customization

VTable allows for dynamic customization at runtime without recreating components:

```go
// For both List and Table components:

// Update a list's styling (theme)
myList.SetStyle(newStyleConfig)

// Update a table's theme
myTable.SetTheme(newTheme)

// Change the data provider
myList.SetDataProvider(newProvider)
myTable.SetDataProvider(newTableProvider)

// Refresh data when the source has changed
myList.RefreshData()
myTable.RefreshData()

// Update the formatter function
myList.SetFormatter(newFormatter)

// Table-specific customization
myTable.SetColumns(newColumns)
myTable.SetHeaderVisibility(false) // Hide header
myTable.SetBorderVisibility(true)  // Show borders

// Programmatically control the components
myList.HandleKeypress("j") // Simulate pressing 'j' (down)
myTable.JumpToIndex(42)    // Jump to a specific row
```

### Event Callbacks

VTable provides a rich event system to respond to user interactions:

```go
// Respond to item selection (Enter key)
myList.OnSelect(func(item MyItem, index int) {
    // Handle selection
})

// Respond to cursor movement
myList.OnHighlight(func(item MyItem, index int) {
    // Update details panel, etc.
})

// Respond to scrolling
myList.OnScroll(func(state vtable.ViewportState) {
    // Update scrollbar or other UI elements
})

// Same callbacks available for tables
myTable.OnSelect(func(row vtable.TableRow, index int) {
    // Handle row selection
})
```

### Themes

VTable comes with several built-in themes:

```go
// Use a built-in theme
theme := vtable.DefaultTheme()  // Default theme
theme := vtable.DarkTheme()     // Dark theme
theme := vtable.LightTheme()    // Light theme
theme := vtable.ColorfulTheme() // Colorful theme

// Customize border characters
theme.BorderChars = vtable.RoundedBorderCharacters()
theme.BorderChars = vtable.ThickBorderCharacters()
theme.BorderChars = vtable.DoubleBorderCharacters()
theme.BorderChars = vtable.AsciiBoxCharacters()

// Create a custom theme
myTheme := vtable.Theme{
    BorderStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("63")),
    HeaderBorderStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("63")),
    HeaderStyle:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252")).Background(lipgloss.Color("57")),
    RowStyle:             lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
    RowEvenStyle:         lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
    RowOddStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
    SelectedRowStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252")).Background(lipgloss.Color("63")),
    TopThresholdStyle:    lipgloss.NewStyle(), // Optional threshold styling
    BottomThresholdStyle: lipgloss.NewStyle(), // Optional threshold styling
    BorderChars:          vtable.RoundedBorderCharacters(),
}

// Apply a theme to an existing table component
myTable.SetTheme(theme)
```

### Key Bindings

VTable supports customizable key bindings:

```go
// Get current keymap
currentKeyMap := myTableModel.GetKeyMap()

// Create a custom keymap
customKeyMap := currentKeyMap

// Modify specific bindings
customKeyMap.PageUp = key.NewBinding(
    key.WithKeys("u", "b", "space"),
    key.WithHelp("space/u/b", "page up (customized)"),
)

// Apply the custom keymap
myTableModel.SetKeyMap(customKeyMap)

// Create a platform-specific keymap
platformKeyMap := vtable.PlatformKeyMap()

// Or create a keymap for a specific platform
macKeyMap := vtable.MacOSKeyMap()
linuxKeyMap := vtable.LinuxKeyMap()
windowsKeyMap := vtable.WindowsKeyMap()
```

### Navigation Methods

```go
// Navigation
model.MoveUp()
model.MoveDown()
model.PageUp()
model.PageDown()
model.JumpToStart()
model.JumpToEnd()
model.JumpToIndex(42)

// Search
found := model.JumpToItem("id", 42)

// Get current state
state := model.GetState()
row, found := model.GetCurrentRow()

// Focus handling
model.Focus()
model.Blur()
isFocused := model.IsFocused()

// Help text
helpText := model.GetHelpView()
```

## Implementing Data Providers

To use VTable, you need to implement the `DataProvider` interface:

```go
// Basic DataProvider interface
type DataProvider[T any] interface {
    // GetTotal returns the total number of items
    GetTotal() int
    
    // GetItems returns a slice of items in the specified range
    GetItems(start, count int) ([]T, error)
}

// Example list provider
type MyListProvider struct {
    items []MyItem
}

func (p *MyListProvider) GetTotal() int {
    return len(p.items)
}

func (p *MyListProvider) GetItems(start, count int) ([]MyItem, error) {
    if start >= len(p.items) {
        return []MyItem{}, nil
    }

    end := start + count
    if end > len(p.items) {
        end = len(p.items)
    }

    return p.items[start:end], nil
}

// Optional SearchableDataProvider for search capabilities
type SearchableDataProvider[T any] interface {
    DataProvider[T]
    
    // FindItemIndex searches for an item based on the given criteria
    FindItemIndex(key string, value any) (int, bool)
}

// Example implementation of search
func (p *MyListProvider) FindItemIndex(key string, value any) (int, bool) {
    if key == "id" {
        id, ok := value.(int)
        if !ok {
            return -1, false
        }
        
        for i, item := range p.items {
            if item.ID == id {
                return i, true
            }
        }
    }
    
    return -1, false
}
```

## Common Patterns

### Updating Live Data

```go
// When your data source changes
func UpdateData() {
    // Update the data source
    myDataSource.AddItem(newItem)
    
    // Tell the list to refresh
    myList.RefreshData()
    
    // Or completely swap out the provider
    myList.SetDataProvider(newProvider)
}
```

### Creating a Master/Detail View

```go
// Set up a detail view updater
myList.OnHighlight(func(item MyItem, index int) {
    // Update your detail view with the highlighted item
    detailView.SetContent(RenderItemDetails(item))
})
```

### Toggling UI Features at Runtime

```go
// Toggle table borders with a key press
if msg.String() == "b" {
    // Toggle borders
    myTable.SetBorderVisibility(!showingBorders)
    showingBorders = !showingBorders
}

// Toggle header with a key press
if msg.String() == "h" {
    // Toggle header
    myTable.SetHeaderVisibility(!showingHeader) 
    showingHeader = !showingHeader
}
```

See the examples directory for complete examples of implementing data providers and building applications with VTable.

## License

[MIT License](LICENSE) 