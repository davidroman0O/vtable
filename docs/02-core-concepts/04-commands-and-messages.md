# Commands and Messages: Controlling VTable

VTable follows Bubble Tea's command/message pattern. You control VTable components by sending commands, and they respond with messages. This section covers what commands you can use and which packages they come from.

## Basic Integration Pattern

Here's the essential pattern for working with VTable:

```go
func (m MyApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Map keys to VTable commands
        switch msg.String() {
        case "j":
            return m, core.CursorDownCmd()
        case "k":  
            return m, core.CursorUpCmd()
        case " ":
            return m, core.SelectCurrentCmd()
        }
    default:
        // Let VTable handle its messages
        newModel, cmd := m.list.Update(msg)
        m.list = newModel.(*list.List)
        return m, cmd
    }
    return m, nil
}
```

**Key point**: You import commands from `core` package and send them. VTable handles the rest.

## Navigation Commands

Import from `"github.com/davidroman0O/vtable/core"`:

```go
// Basic movement
core.CursorUpCmd()        // Move cursor up one item
core.CursorDownCmd()      // Move cursor down one item

// Page movement  
core.PageUpCmd()          // Move up one viewport height
core.PageDownCmd()        // Move down one viewport height

// Jump movement
core.JumpToStartCmd()     // Go to first item
core.JumpToEndCmd()       // Go to last item
core.JumpToCmd(index)     // Go to specific item
```

**Usage:**
```go
// In your key handler
switch key {
case "h":
    return m, core.PageUpCmd()
case "l":
    return m, core.PageDownCmd()  
case "g":
    return m, core.JumpToStartCmd()
case "G":
    return m, core.JumpToEndCmd()
}
```

## Selection Commands

Also from `core` package:

```go
core.SelectCurrentCmd()   // Toggle selection of current item
core.SelectAllCmd()       // Select all items  
core.SelectClearCmd()     // Clear all selections
```

**Usage:**
```go
switch key {
case " ":  // Spacebar
    return m, core.SelectCurrentCmd()
case "a":
    return m, core.SelectAllCmd()
case "c":
    return m, core.SelectClearCmd()
}
```

## Data Commands

For refreshing data, from `core` package:

```go
core.DataRefreshCmd()        // Reload all data from DataSource
core.DataChunksRefreshCmd()  // Refresh current chunks only
```

**Usage:**
```go
case "r":
    return m, core.DataRefreshCmd()
```

## Response Messages You Receive

VTable sends these messages back to your app. Handle them to update your UI:

### Selection Responses
```go
// From core package
case core.SelectionResponseMsg:
    if msg.Success {
        // Update your status display
        m.statusMessage = fmt.Sprintf("Selected item %s", msg.ID)
    }
    // Always pass to VTable too
    newModel, cmd := m.list.Update(msg)
    m.list = newModel.(*list.List) 
    return m, cmd
```

### Data Load Responses
```go
case core.DataTotalMsg:
    // VTable loaded total count
    m.totalItems = msg.Total
    
case core.DataChunkLoadedMsg:
    // VTable loaded a data chunk
    // Usually you just pass this through
    newModel, cmd := m.list.Update(msg)
    m.list = newModel.(*list.List)
    return m, cmd
```

## Tree-Specific Commands

For tree components, import from tree package:

```go
import "github.com/davidroman0O/vtable/tree"

// Tree navigation with expansion
core.TreeJumpToIndexCmd(index, expandParents)
```

## Combining Commands

Use `tea.Batch()` to send multiple commands:

```go
// Select current item and move down
return m, tea.Batch(
    core.SelectCurrentCmd(),
    core.CursorDownCmd(),
)
```

## Essential Integration Example

Here's a complete minimal example:

```go
package main

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/davidroman0O/vtable/core"
    "github.com/davidroman0O/vtable/list"
)

type App struct {
    list *list.List
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q":
            return a, tea.Quit
        case "j":
            return a, core.CursorDownCmd()
        case "k":
            return a, core.CursorUpCmd()
        case " ":
            return a, core.SelectCurrentCmd()
        }
    default:
        // Always pass other messages to VTable
        newModel, cmd := a.list.Update(msg)
        a.list = newModel.(*list.List)
        return a, cmd
    }
    return a, nil
}
```

## Quick Reference

**Navigation:**
- `core.CursorUpCmd()` / `core.CursorDownCmd()` - Basic movement
- `core.PageUpCmd()` / `core.PageDownCmd()` - Page movement  
- `core.JumpToStartCmd()` / `core.JumpToEndCmd()` - Jump to ends
- `core.JumpToCmd(index)` - Jump to specific item

**Selection:**
- `core.SelectCurrentCmd()` - Toggle current item
- `core.SelectAllCmd()` - Select all
- `core.SelectClearCmd()` - Clear selections

**Data:**
- `core.DataRefreshCmd()` - Reload data

**Important:** Always pass unhandled messages to your VTable component's `Update()` method.

## What You Need to Know

1. **Import from `core` package** for most commands
2. **Map your keys** to VTable commands in your `Update()` method
3. **Always pass through** messages you don't handle to VTable
4. **Handle response messages** to update your UI status
5. **Use `tea.Batch()`** to combine multiple commands

VTable handles all the complex viewport calculations, data loading, and state management. You just send simple commands and handle the responses.

**Next:** [Component Rendering →](05-component-rendering.md) 