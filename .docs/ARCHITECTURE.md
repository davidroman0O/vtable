# Pure Tea VTable Architecture

## Overview

This document outlines the complete architecture for a **pure Bubble Tea** implementation of the vtable library. The goal is to eliminate all direct method calls and create a fully message-driven system that follows Tea's idiomatic patterns.

We will use the codebase at the root of the project as inspiration to be ported into that pure bubble tea implementation under the folder `./pure`.

## Core Design Principles

1. **Single Flat Models** - No nested model wrapping (TeaTable → Table → List)
2. **Message-Only State Changes** - All mutations via `Update()` method
3. **No Direct Methods** - No `table.MoveUp()`, only commands that return `tea.Cmd`
4. **Immutable Updates** - Return new state, don't mutate existing
5. **Command-Based Operations** - Everything returns `tea.Cmd`
6. **Async Data Loading** - No blocking operations in the UI thread

## Core Models

### List Model
The base virtualized list component that handles:
- Virtual scrolling with chunks
- Data management via DataSource
- Selection (single/multiple/none)
- Navigation and focus
- Animation support
- Filtering and sorting

```go
type List struct {
    // Core virtualization
    viewportConfig    ViewportConfig
    viewportState     ViewportState
    
    // Data management
    dataSource        DataSource[any]
    totalItems        int
    dataRequest       DataRequest
    
    // Chunk management
    loadedChunks      map[int]*Chunk
    loadingChunks     map[int]bool
    chunkRequests     map[int]DataRequest
    visibleItems      []Data[any]
    
    // Selection
    selectionMode     SelectionMode
    selectedItems     map[string]bool
    
    // UI state
    focused           bool
    cursorIndex       int
    
    // Formatting
    formatter         ItemFormatter[any]
    formatterAnimated ItemFormatterAnimated[any]
    
    // Animation
    animationEngine   *AnimationEngine
    animationConfig   AnimationConfig
    itemAnimations    map[string]ListAnimation
    
    // Theme
    styleConfig       StyleConfig
    
    // Configuration
    keyMap            NavigationKeyMap
    maxWidth          int
}
```

### Table Model
Table-specific component that composes with List functionality:

```go
type Table struct {
    // Table-specific configuration
    columns               []TableColumn
    showHeader            bool
    showBorders           bool
    
    // Core list functionality (embedded or composed)
    listCore              ListCore
    
    // Multi-level formatters
    cellFormatters        map[int]CellFormatter
    cellFormattersAnimated map[int]CellFormatterAnimated
    rowFormatter          RowFormatter
    headerFormatter       HeaderFormatter
    
    // Constraints
    columnConstraints     []CellConstraint
    
    // Animation (more complex than list)
    cellAnimations        map[string]CellAnimation    // "rowID:colIndex"
    rowAnimations         map[string]RowAnimation
    
    // Table-specific styling
    theme                 Theme
    borderChars           BorderChars
    
    // Calculated layout
    totalWidth            int
    horizontalBorders     map[string]string  // top, middle, bottom
}
```

## Complete Message System

### Shared Messages (Both List & Table)

#### Navigation Messages
```go
type CursorUpMsg struct{}
type CursorDownMsg struct{}
type PageUpMsg struct{}
type PageDownMsg struct{}
type JumpToStartMsg struct{}
type JumpToEndMsg struct{}
type JumpToMsg struct{ Index int }
```

#### Data Messages
```go
type DataRefreshMsg struct{}
type DataChunkLoadedMsg struct {
    StartIndex int
    Items      []Data[any]
    Request    DataRequest  // For validation
}
type DataChunkErrorMsg struct {
    StartIndex int
    Error      error
    Request    DataRequest
}
type DataTotalMsg struct {
    Total int
}
type DataLoadErrorMsg struct{ Error error }
```

#### Selection Messages
```go
type SelectCurrentMsg struct{}
type SelectToggleMsg struct{ Index int }
type SelectAllMsg struct{}
type SelectClearMsg struct{}
type SelectRangeMsg struct{ StartID, EndID string }
```

#### Filter Messages
```go
type FilterSetMsg struct {
    Field string
    Value any
}
type FilterClearMsg struct{ Field string }
type FiltersClearAllMsg struct{}
```

#### Sort Messages
```go
type SortToggleMsg struct{ Field string }
type SortSetMsg struct {
    Field     string
    Direction string
}
type SortAddMsg struct {
    Field     string
    Direction string
}
type SortRemoveMsg struct{ Field string }
type SortsClearAllMsg struct{}
```

#### Focus Messages
```go
type FocusMsg struct{}
type BlurMsg struct{}
```

#### Animation Messages
```go
type GlobalAnimationTickMsg struct{ Timestamp time.Time }
type AnimationUpdateMsg struct{ UpdatedAnimations []string }
type AnimationConfigMsg struct{ Config AnimationConfig }
type AnimationStartMsg struct{ AnimationID string }
type AnimationStopMsg struct{ AnimationID string }
```

#### Theme Messages
```go
type ThemeSetMsg struct{ Theme interface{} } // Theme or StyleConfig
```

#### Real-time Update Messages
```go
type RealTimeUpdateMsg struct{}
type RealTimeConfigMsg struct {
    Enabled  bool
    Interval time.Duration
}
```

### Table-Specific Messages
```go
type ColumnSetMsg struct{ Columns []TableColumn }
type HeaderVisibilityMsg struct{ Visible bool }
type BorderVisibilityMsg struct{ Visible bool }
type CellFormatterSetMsg struct {
    ColumnIndex int               // -1 for all columns
    Formatter   CellFormatter
}
type CellAnimatedFormatterSetMsg struct {
    ColumnIndex int
    Formatter   CellFormatterAnimated  
}
type RowFormatterSetMsg struct{ Formatter RowFormatter }
type HeaderFormatterSetMsg struct{ Formatter HeaderFormatter }
type ColumnConstraintsSetMsg struct {
    ColumnIndex int
    Constraints CellConstraint
}
type TableThemeSetMsg struct{ Theme Theme }
```

### List-Specific Messages
```go
type FormatterSetMsg struct{ Formatter ItemFormatter[any] }
type AnimatedFormatterSetMsg struct{ Formatter ItemFormatterAnimated[any] }
type ChunkSizeSetMsg struct{ Size int }
type MaxWidthSetMsg struct{ Width int }
```

### Animation Control Messages
```go
type CellAnimationStartMsg struct {
    RowID       string
    ColumnIndex int
    Animation   CellAnimation
}
type RowAnimationStartMsg struct {
    RowID     string
    Animation RowAnimation
}
type ItemAnimationStartMsg struct {
    ItemID    string
    Animation ListAnimation
}
```

## Data Management System

### DataSource Interface
Pure Tea compatible data source that returns commands instead of direct data:

```go
type DataSource[T any] interface {
    // Async data loading via commands
    LoadChunk(request DataRequest) tea.Cmd
    GetTotal() tea.Cmd
    RefreshTotal() tea.Cmd
    
    // Pure functions (no state mutation)
    GetItemID(item T) string
    GetSelectionMode() SelectionMode
}
```

### Viewport-Chunk Coordination

#### Chunk Management
- **Async Loading**: All chunk loading via commands/messages
- **Stale Data Handling**: Validate requests against current filters/sorts
- **Smart Buffering**: Load chunks ahead/behind viewport for smooth scrolling
- **Memory Management**: Unload distant chunks automatically

#### State Coordination
```go
type Chunk struct {
    StartIndex   int
    EndIndex     int  
    Items        []Data[any]
    LoadedAt     time.Time
    Request      DataRequest    // The request that created this chunk
}

// Chunk state tracking
loadedChunks      map[int]*Chunk        // StartIndex -> Chunk
loadingChunks     map[int]bool          // Track chunks being loaded
chunkRequests     map[int]DataRequest   // Track pending requests
```

#### Filtering/Sorting Impact
- **Chunk Invalidation**: Clear all chunks when filters/sorts change
- **Request Validation**: Reject stale chunk responses
- **Viewport Reset**: Return to start when data changes significantly

## Rendering Architecture

### Hierarchical Formatter System

#### List Formatters
```go
type ItemFormatter[T any] func(
    data Data[T],
    index int,
    ctx RenderContext,
    isCursor bool,
    isTopThreshold bool,
    isBottomThreshold bool,
) string

type ItemFormatterAnimated[T any] func(
    data Data[T], 
    index int,
    ctx RenderContext,
    animationState map[string]any,
    isCursor bool,
    isTopThreshold bool,
    isBottomThreshold bool,
) RenderResult
```

#### Table Formatters
```go
type CellFormatter func(
    cellValue string,
    rowIndex int,
    columnIndex int,
    column TableColumn,
    ctx RenderContext,
    isCursor bool,
    isSelected bool,
) string

type CellFormatterAnimated func(
    cellValue string,
    rowIndex int,
    columnIndex int, 
    column TableColumn,
    ctx RenderContext,
    animationState map[string]any,
    isCursor bool,
    isSelected bool,
    isTopThreshold bool,
    isBottomThreshold bool,
) CellRenderResult

type RowFormatter func(
    row TableRow,
    columns []TableColumn,
    cellResults []CellRenderResult,
    ctx RenderContext,
    isCursor bool,
    isSelected bool,
) string

type HeaderFormatter func(
    columns []TableColumn,
    ctx RenderContext,
) string
```

### Constraint Management
```go
type CellConstraint struct {
    Width     int           // Exact width (enforced)
    Height    int           // Usually 1 for tables  
    Alignment int           // Left/Right/Center
    Padding   PaddingConfig
    MaxLines  int           // For future multi-line support
}

type CellRenderResult struct {
    Content         string            // Final rendered content
    ActualWidth     int               // Actual width used
    ActualHeight    int               // Actual height used
    Overflow        bool              // Whether content was truncated
    AnimationState  map[string]any    // Animation state for next render
    RefreshTriggers []RefreshTrigger  // When to re-render
    Error           error             // Rendering error
    Fallback        string            // Fallback content
}
```

### Table Rendering Pipeline
1. **Cell Rendering**: Apply cell formatter with constraints
2. **Animation Processing**: Handle cell-level animations
3. **Constraint Enforcement**: Ensure exact width/height compliance
4. **Row Assembly**: Combine cells with borders/padding
5. **Header Rendering**: Format header if enabled
6. **Border Management**: Add table borders if enabled

## Animation System

### Animation Engine
Global animation management with pure Tea messages:

```go
type AnimationEngine struct {
    animations       map[string]*AnimationState
    config           AnimationConfig
    needsUpdate      bool
    lastGlobalUpdate time.Time
    isRunning        bool
}
```

### Animation Types

#### List Animations
```go
type ListAnimation struct {
    ItemID        string
    AnimationType string          // "fade", "slide", "highlight"
    State         map[string]any  // Animation-specific state
    Triggers      []RefreshTrigger
}
```

#### Table Animations
```go
type CellAnimation struct {
    RowID         string
    ColumnIndex   int
    AnimationType string
    State         map[string]any
    Triggers      []RefreshTrigger
}

type RowAnimation struct {
    RowID         string
    AnimationType string
    State         map[string]any
    Triggers      []RefreshTrigger
}
```

### Animation Lifecycle
1. **Registration**: Via messages (`CellAnimationStartMsg`, etc.)
2. **Global Ticker**: `GlobalAnimationTickMsg` drives all animations
3. **State Updates**: `AnimationUpdateMsg` when animations change
4. **Rendering Integration**: Formatters receive animation state
5. **Cleanup**: Automatic cleanup of completed animations

## Selection System

### Selection Modes
- **SelectionNone**: No selection allowed
- **SelectionSingle**: Only one item can be selected
- **SelectionMultiple**: Multiple items can be selected

### Selection State Management
- **ID-based Tracking**: Use item IDs for stable selection
- **Range Selection**: Support for selecting ranges of items
- **Persistence**: Selection survives filtering/sorting when possible

### Selection Messages
All selection changes via messages:
- `SelectCurrentMsg`: Select item at cursor
- `SelectToggleMsg{Index}`: Toggle specific item
- `SelectAllMsg`: Select all visible items
- `SelectClearMsg`: Clear all selections
- `SelectRangeMsg{StartID, EndID}`: Select range

## Theme System

### List Theming
```go
type StyleConfig struct {
    CursorStyle      lipgloss.Style
    SelectedStyle    lipgloss.Style
    DefaultStyle     lipgloss.Style
    ThresholdStyle   lipgloss.Style
    // ... other styles
}
```

### Table Theming
```go
type Theme struct {
    HeaderStyle      lipgloss.Style
    CellStyle        lipgloss.Style
    CursorStyle      lipgloss.Style
    SelectedStyle    lipgloss.Style
    BorderChars      BorderChars
    BorderColor      string
    // ... other theme elements
}
```

## Command System

Every operation returns a `tea.Cmd`:

### Navigation Commands
```go
func CursorUpCmd() tea.Cmd
func CursorDownCmd() tea.Cmd
func PageUpCmd() tea.Cmd
func PageDownCmd() tea.Cmd
func JumpToCmd(index int) tea.Cmd
func JumpToStartCmd() tea.Cmd
func JumpToEndCmd() tea.Cmd
```

### Data Commands
```go
func LoadDataCmd(rows []any) tea.Cmd
func RefreshDataCmd() tea.Cmd
func SetFilterCmd(field string, value any) tea.Cmd
func ClearFiltersCmd() tea.Cmd
func SetSortCmd(field, direction string) tea.Cmd
func ClearSortCmd() tea.Cmd
```

### Selection Commands
```go
func SelectCurrentCmd() tea.Cmd
func SelectAllCmd() tea.Cmd
func ClearSelectionCmd() tea.Cmd
func ToggleSelectionCmd(index int) tea.Cmd
```

### Animation Commands
```go
func StartAnimationCmd(id string, animation Animation) tea.Cmd
func StopAnimationCmd(id string) tea.Cmd
func SetAnimationConfigCmd(config AnimationConfig) tea.Cmd
```

## Implementation Plan

### Phase 1: Core Infrastructure
**Files to create:**
- `pure/types.go` - All data types (TableRow, Column, Data, etc.)
- `pure/interfaces.go` - DataSource, formatters, etc.
- `pure/messages.go` - Complete message system
- `pure/commands.go` - All command functions
- `pure/config.go` - Configuration types

### Phase 2: Core Models
**Files to create:**
- `pure/list_core.go` - Core virtualization logic (shared)
- `pure/list.go` - List model with Update/View
- `pure/table.go` - Table model with Update/View
- `pure/viewport.go` - Viewport calculations

### Phase 3: Subsystems
**Files to create:**
- `pure/animation.go` - Animation engine + messages
- `pure/selection.go` - Selection management
- `pure/filtering.go` - Filter logic
- `pure/sorting.go` - Sort logic
- `pure/theme.go` - Theming system
- `pure/navigation.go` - Navigation logic
- `pure/keybindings.go` - Key handling
- `pure/chunk.go` - Chunk management
- `pure/datasource.go` - DataSource implementations

### Phase 4: Testing & Examples
**Directories to create:**
- `pure/examples/basic-list/` - Simple list example
- `pure/examples/basic-table/` - Simple table example  
- `pure/examples/advanced/` - All features demo
- `pure/examples/animated/` - Animation showcase
- `pure/examples/selection/` - Selection demo
- `pure/examples/realtime/` - Real-time updates

### Phase 5: Migration & Documentation
**Files to create:**
- `pure/MIGRATION.md` - Migration guide from old API
- `pure/EXAMPLES.md` - Usage examples
- `pure/API.md` - Complete API reference

## Error Handling

### Data Loading Errors
- `DataChunkErrorMsg` for chunk loading failures
- Fallback content in render results
- Retry mechanisms via commands

### Animation Errors
- Error fields in `CellRenderResult`
- Fallback content when animations fail
- Graceful degradation to non-animated content

### Validation Errors
- Configuration validation at startup
- Runtime validation of message parameters
- Clear error messages with suggested fixes

## Performance Considerations

### Memory Management
- Automatic chunk unloading for distant data
- Animation state cleanup
- Selection state optimization

### Rendering Optimization
- Only re-render changed content
- Animation batching
- Efficient constraint calculation

### Data Loading Optimization
- Smart chunk prefetching
- Request deduplication
- Stale data rejection

## Testing Strategy

### Unit Tests
- Individual component testing
- Message handling verification
- State transition validation

### Integration Tests
- Viewport-chunk coordination
- Animation system integration
- Selection persistence across operations

### Performance Tests
- Large dataset handling
- Memory usage validation
- Animation performance

This architecture provides a complete foundation for a pure Bubble Tea implementation that eliminates all direct method calls while preserving all functionality of the original vtable library. 