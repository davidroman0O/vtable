# Separated Components Architecture

This document describes the new separated architecture for the Pure Tea vtable implementation, which separates viewport management and chunk loading concerns into independent, reusable components.

## Overview

The previous architecture had viewport management, navigation logic, and chunk loading tightly coupled within the List component. The new architecture separates these concerns into:

1. **Viewport Component** - Handles viewport state, navigation, and boundary detection
2. **BoundingArea Manager** - Handles chunk loading/unloading based on viewport position  
3. **List/Table Components** - Focus on data presentation and user interaction

## Architecture Benefits

### 1. **Type Safety**
- Replaced hardcoded string navigation types with type-safe constants
- Dedicated message types for each navigation action
- Compile-time validation of navigation commands

### 2. **Separation of Concerns**
- Viewport logic is independent of data loading
- Chunk management is independent of viewport positioning
- Components can be tested and debugged in isolation

### 3. **Reusability**
- Viewport component can be used by both List and Table
- BoundingArea manager can be used with different data sources
- Components follow Bubble Tea patterns and can be embedded

### 4. **Maintainability**
- Clear boundaries between components
- Easier to understand and modify individual behaviors
- Reduced complexity in main List/Table components

## Components

### Viewport Component (`viewport.go`)

Manages viewport state, navigation, and boundary detection as a pure Bubble Tea component.

#### Key Features:
- **Navigation Types**: Type-safe constants instead of strings
- **Message Types**: Dedicated messages for each navigation action
- **State Management**: Tracks cursor position, viewport bounds, and thresholds
- **Event Callbacks**: Optional notifications for state changes

#### Usage:
```go
// Create viewport
config := ViewportConfig{
    Height: 10,
    TopThresholdIndex: 2,
    BottomThresholdIndex: 7,
    ChunkSize: 20,
}
viewport := NewViewport(config)

// Handle navigation
viewport, cmd := viewport.Update(ViewportDownMsg{})

// Type-safe commands
cmd := ViewportUpCmd()
cmd := ViewportJumpCmd(42)
```

#### Navigation Types:
```go
type NavigationType int

const (
    NavigationUp NavigationType = iota
    NavigationDown
    NavigationPageUp
    NavigationPageDown
    NavigationStart
    NavigationEnd
    NavigationJump
)
```

### BoundingArea Manager (`bounding_area.go`)

Manages chunk loading/unloading based on viewport position to ensure smooth scrolling.

#### Key Features:
- **Proactive Loading**: Loads chunks before they're needed
- **Smart Unloading**: Unloads distant chunks to manage memory
- **Configurable Strategy**: Adjustable chunks before/after viewport
- **Callback System**: Integrates with any data loading system

#### Usage:
```go
// Create bounding area manager
config := BoundingAreaConfig{
    ChunkSize: 20,
    ChunksBefore: 1,
    ChunksAfter: 2,
    UnloadDistantChunks: true,
}
manager := NewBoundingAreaManager(config)

// Set callbacks for actual chunk operations
manager.SetCallbacks(
    func(startIndex, count int) tea.Cmd {
        // Load chunk
        return dataSource.LoadChunk(startIndex, count)
    },
    func(startIndex, count int) tea.Cmd {
        // Unload chunk
        return dataSource.UnloadChunk(startIndex, count)
    },
)

// Update based on viewport changes
manager, cmd := manager.Update(BoundingAreaUpdateMsg{
    ViewportState: viewportState,
    TotalItems: 1000,
})
```

## Message Flow

### Navigation Flow:
1. User input creates navigation message (e.g., `ViewportDownMsg{}`)
2. Viewport component processes navigation and updates state
3. Viewport emits `ViewportStateChangedMsg` with state changes
4. BoundingArea manager receives viewport state update
5. BoundingArea manager calculates required chunks and triggers load/unload

### Type-Safe Commands:
```go
// Old way (error-prone)
cmd := ViewportNavigationCmd("down", 0) // String could be typo'd

// New way (type-safe)
cmd := ViewportDownCmd() // Compile-time validated
```

## Integration Example

Here's how a List component integrates both components:

```go
type List struct {
    // Core components
    viewport *Viewport
    boundingManager *BoundingAreaManager
    
    // ... other fields
}

func (l *List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Convert key to navigation message
        switch msg.String() {
        case "down":
            l.viewport, cmd := l.viewport.Update(ViewportDownMsg{})
            cmds = append(cmds, cmd)
        }
        
    case ViewportStateChangedMsg:
        // Update bounding area when viewport changes
        l.boundingManager, cmd := l.boundingManager.Update(BoundingAreaUpdateMsg{
            ViewportState: msg.State,
            TotalItems: l.totalItems,
        })
        cmds = append(cmds, cmd)
    }
    
    return l, tea.Batch(cmds...)
}
```

## Testing

The new architecture enables focused testing:

```go
// Test viewport navigation in isolation
func TestViewportNavigation(t *testing.T) {
    viewport := NewViewport(config)
    viewport, _ = viewport.Update(ViewportDownMsg{})
    
    state := viewport.GetState()
    assert.Equal(t, 1, state.CursorIndex)
}

// Test bounding area calculation
func TestBoundingAreaCalculation(t *testing.T) {
    manager := NewBoundingAreaManager(config)
    area := manager.CalculateBoundingArea(viewportState, totalItems)
    
    assert.Equal(t, expectedStart, area.StartIndex)
    assert.Equal(t, expectedEnd, area.EndIndex)
}
```

See `pure/test/viewport/demo_test.go` for comprehensive examples.

## Migration Guide

### From Old Architecture:
```go
// Old: Hardcoded strings
switch msg.Type {
case "up":
    l.moveUp()
case "down": 
    l.moveDown()
}

// Old: Mixed concerns
func (l *List) moveDown() {
    // Navigation logic
    l.cursor++
    // Chunk loading logic
    l.ensureChunkLoaded(l.cursor)
    // Viewport management
    l.updateViewport()
}
```

### To New Architecture:
```go
// New: Type-safe constants
switch msg.Type {
case NavigationUp:
    v.moveUp()
case NavigationDown:
    v.moveDown()
}

// New: Separated concerns
func (v *Viewport) moveDown() {
    // Only navigation logic
    v.state.CursorIndex++
    v.updateViewportBounds()
}

func (b *BoundingAreaManager) handleUpdate(msg BoundingAreaUpdateMsg) {
    // Only chunk loading logic
    loadRequests := b.calculateChunkOperations(boundingArea, totalItems)
    // ...
}
```

## Performance Benefits

1. **Reduced Complexity**: Each component has a single responsibility
2. **Better Caching**: BoundingArea manager can optimize chunk loading patterns
3. **Memory Management**: Intelligent unloading of distant chunks
4. **Predictable Behavior**: Clear separation makes performance characteristics easier to understand

## Future Enhancements

The separated architecture enables future improvements:

1. **Multiple Viewport Strategies**: Different navigation behaviors
2. **Adaptive Chunk Loading**: Dynamic chunk sizes based on scroll patterns
3. **Parallel Loading**: Independent chunk loading for different components
4. **Enhanced Testing**: Mock components for unit testing
5. **Reusable Components**: Use viewport/bounding logic in other contexts

## Conclusion

The separated components architecture provides:
- **Type Safety**: Compile-time validation of navigation operations
- **Maintainability**: Clear separation of concerns
- **Reusability**: Components can be used independently
- **Testability**: Each component can be tested in isolation
- **Performance**: Optimized chunk loading strategies

This architecture follows Bubble Tea idioms while providing the flexibility and robustness needed for complex data virtualization scenarios. 