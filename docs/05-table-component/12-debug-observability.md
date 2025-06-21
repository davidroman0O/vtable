# Debug and Observability

## What We're Adding

Building on our filtering and sorting example, we're adding comprehensive debugging and monitoring capabilities to understand what's happening inside VTable:

- **Message interception**: Capture internal VTable communications without interfering
- **Chunk loading visualization**: Monitor data loading and memory management in real-time  
- **Activity tracking**: Log all user interactions and system responses with timestamps
- **Performance monitoring**: Track timing, memory usage, and identify bottlenecks
- **Multi-level debug modes**: From basic activity logs to verbose message tracing

This transforms your table into a fully observable system perfect for development, debugging, and performance optimization.

## Core Concept: Non-Invasive Observation

**Critical Rule**: Never interfere with normal table operation. Always observe first, then pass all messages through to the table.

```go
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // 1. Intercept messages for debugging (don't change behavior)
    switch msg := msg.(type) {
    case core.ChunkLoadingStartedMsg:
        m.logActivity("system", "chunk_load_start", fmt.Sprintf("Chunk %d", msg.ChunkStart))
    case core.DataTotalMsg:
        m.logActivity("system", "data_total", fmt.Sprintf("Total: %d items", msg.Total))
    // ... other message types
    }
    
    // 2. ALWAYS pass ALL messages to table (critical!)
    var cmd tea.Cmd
    _, cmd = m.table.Update(msg)
    return m, cmd
}
```

## Key VTable Messages to Intercept

### Chunk Management Messages
These give you insight into VTable's memory management:

```go
case core.ChunkLoadingStartedMsg:
    // Triggered when table needs new data
    // Contains: ChunkStart, Request (with Start/Count)
    
case core.ChunkLoadingCompletedMsg:  
    // Triggered when chunk loading finishes
    // Contains: ChunkStart, ItemCount
    
case core.ChunkUnloadedMsg:
    // Triggered when old chunks are freed from memory
    // Contains: ChunkStart
    
case core.DataChunkLoadedMsg:
    // Triggered when actual data arrives
    // Contains: StartIndex, Items[], Request
```

### Data Operation Messages
Track data lifecycle and operations:

```go
case core.DataRefreshMsg:
    // Triggered when table requests data refresh
    
case core.DataTotalMsg:
    // Triggered when total count is received
    // Contains: Total
    
case core.SelectionResponseMsg:
    // Triggered on selection operations
    // Contains: Success, Index, ID, Operation
```

### Navigation Messages
Monitor user navigation:

```go
case core.CursorUpMsg, core.CursorDownMsg:
    // Row navigation
    
case core.PageUpMsg, core.PageDownMsg:
    // Page-based navigation
    
case core.JumpToStartMsg, core.JumpToEndMsg:
    // Jump operations
```

## Debug System Architecture

### 1. Activity Logging System
Track everything that happens with timestamps:

```go
type ActivityLog struct {
    Timestamp time.Time
    Type      string        // "user", "system", "performance"  
    Action    string        // "navigation", "chunk_load", "sort"
    Details   string        // Specific details
    Duration  time.Duration // For timing operations
}

func (m *AppModel) logActivity(activityType, action, details string) {
    activity := ActivityLog{
        Timestamp: time.Now(),
        Type:      activityType,
        Action:    action,
        Details:   details,
    }
    
    m.activityLog = append(m.activityLog, activity)
    
    // Keep only recent activities (prevent memory bloat)
    if len(m.activityLog) > 50 {
        m.activityLog = m.activityLog[len(m.activityLog)-50:]
    }
}
```

### 2. Chunk State Tracking
Monitor memory usage and loading patterns:

```go
type ChunkState struct {
    StartIndex    int
    Size          int
    LoadStartTime time.Time
    LoadEndTime   time.Time
    Status        string // "loading", "loaded", "unloaded"
    AccessCount   int
}

// In your update method:
case core.ChunkLoadingStartedMsg:
    m.chunkStates[msg.ChunkStart] = ChunkState{
        StartIndex:    msg.ChunkStart,
        Size:          msg.Request.Count,
        LoadStartTime: time.Now(),
        Status:        "loading",
    }
    
case core.ChunkLoadingCompletedMsg:
    if state, exists := m.chunkStates[msg.ChunkStart]; exists {
        state.LoadEndTime = time.Now()
        state.Status = "loaded"
        duration := state.LoadEndTime.Sub(state.LoadStartTime)
        m.logActivity("system", "chunk_complete", 
            fmt.Sprintf("Chunk %d loaded in %v", msg.ChunkStart, duration))
    }
```

### 3. Four-Level Debug System
Different levels provide different insights:

```go
const (
    DebugOff      = 0  // Clean UI, no debug info
    DebugBasic    = 1  // Activity log only  
    DebugDetailed = 2  // + Chunk states + Performance
    DebugVerbose  = 3  // + Full message logging
)

// Toggle with 'd' key
case "d":
    m.debugMode = (m.debugMode + 1) % 4
    m.statusMessage = fmt.Sprintf("Debug mode: %s", m.getDebugModeName())
```

## Practical Debug Workflows

### Performance Troubleshooting
1. **Enable Detailed mode**: Shows chunk loading times
2. **Navigate extensively**: Trigger chunk loading/unloading
3. **Watch for patterns**: Slow chunks, excessive unloading
4. **Check memory usage**: Monitor chunk count

```go
// Example: Track chunk loading performance
case core.ChunkLoadingCompletedMsg:
    duration := time.Since(state.LoadStartTime)
    if duration > 100*time.Millisecond {
        m.logActivity("performance", "slow_chunk", 
            fmt.Sprintf("Chunk %d took %v (>100ms)", msg.ChunkStart, duration))
    }
```

### User Interaction Analysis
1. **Enable Basic mode**: See user actions
2. **Perform operations**: Sort, filter, navigate
3. **Check sequence**: Verify expected message flow
4. **Time operations**: Identify slow user interactions

```go
// Track user input with context
func (m *AppModel) trackUserInput(key string) {
    var action string
    switch key {
    case "s":
        action = "sort_toggle" 
    case "1", "2", "3", "4", "5":
        action = "filter_toggle"
    case "up", "down", "j", "k":
        action = "row_navigation"
    default:
        action = "other_input"
    }
    
    m.logActivity("user", action, fmt.Sprintf("Key: %s", key))
}
```

### Data Source Debugging
1. **Enable Verbose mode**: See all messages
2. **Trigger data operations**: Refresh, sort, filter
3. **Verify message flow**: DataRefresh → DataTotal → ChunkLoaded
4. **Check data integrity**: Ensure proper item counts

## Real-Time Debug Display

### Activity Log (Basic Mode)
Shows recent system activity:

```go
func (m *AppModel) renderActivityLog() string {
    if len(m.activityLog) == 0 {
        return "No recent activity"
    }
    
    var activities []string
    count := len(m.activityLog)
    start := 0
    if count > 8 {
        start = count - 8  // Show last 8 activities
    }
    
    for i := start; i < count; i++ {
        activity := m.activityLog[i]
        timeStr := activity.Timestamp.Format("15:04:05")
        activities = append(activities,
            fmt.Sprintf("%s [%s] %s: %s",
                timeStr, activity.Type, activity.Action, activity.Details))
    }
    
    return strings.Join(activities, "\n")
}
```

### Chunk States (Detailed Mode)  
Monitor memory and loading:

```go
func (m *AppModel) renderChunkStates() string {
    if len(m.chunkStates) == 0 {
        return "No chunks loaded"
    }
    
    var chunks []string
    for start, state := range m.chunkStates {
        status := state.Status
        if state.Status == "loaded" && !state.LoadEndTime.IsZero() {
            duration := state.LoadEndTime.Sub(state.LoadStartTime)
            status = fmt.Sprintf("loaded (%v)", duration)
        }
        
        chunks = append(chunks,
            fmt.Sprintf("Chunk %d: %s (size: %d, accessed: %d)",
                start, status, state.Size, state.AccessCount))
    }
    
    return strings.Join(chunks, "\n")
}
```

### Performance Metrics
Track system resource usage:

```go
func (m *AppModel) updateMemoryUsage() {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    m.performanceMetrics.MemoryUsage = int64(memStats.Alloc)
}

// Display in debug UI
debug.WriteString(fmt.Sprintf("Mode: %s | Operations: %d | Memory: %s | Active Chunks: %d\n",
    m.getDebugModeName(),
    m.performanceMetrics.TotalOperations,
    m.formatBytes(m.performanceMetrics.MemoryUsage),
    len(m.chunkStates)))
```

## Debug Controls

### Primary Controls
| Key | Action | Purpose |
|-----|--------|---------|
| `d` | Cycle debug mode | Off→Basic→Detailed→Verbose |
| `D` | Toggle overlay | Persistent debug display |
| `Ctrl+R` | Reset debug data | Clear logs and metrics |

### What Each Mode Shows
- **Off**: Clean interface, no debug info
- **Basic**: Recent activity log (last 8 operations)
- **Detailed**: + Chunk states + Performance metrics  
- **Verbose**: + Full message logging (last 20 messages)

## Common Debug Scenarios

### "Table shows 'No data available'"
1. **Check activity log**: Look for `data_total` and `chunk_data_loaded` messages
2. **Verify data source**: Ensure `GetTotal()` and `LoadChunk()` are called
3. **Check message flow**: DataRefresh → DataTotal → DataChunkLoaded

### "Slow scrolling/navigation"
1. **Enable Detailed mode**: Monitor chunk loading times
2. **Check chunk count**: Too many active chunks indicate memory issues
3. **Look for patterns**: Excessive chunk unloading/reloading

### "Memory usage growing"
1. **Monitor chunk states**: Look for chunks that never unload
2. **Check access patterns**: High access counts indicate hot spots
3. **Watch for leaks**: Growing chunk count over time

### "Sorting/filtering not working"
1. **Enable Basic mode**: Verify user input is detected
2. **Check data operations**: Look for data_refresh messages
3. **Monitor chunk reloading**: Ensure chunks reload after data changes

## Implementation Pattern

Here's the complete pattern for adding debug observability:

```go
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Log verbose messages first
    if m.debugMode >= DebugVerbose {
        m.logMessage(msg)
    }

    // Intercept specific messages for debug tracking
    switch msg := msg.(type) {
    case core.ChunkLoadingStartedMsg:
        // Track chunk loading start
    case core.ChunkLoadingCompletedMsg:
        // Track chunk loading completion  
    case core.DataTotalMsg:
        // Track data total updates
    case tea.KeyMsg:
        // Track user input
        if !m.searchMode {  // Handle special modes
            m.trackUserInput(msg.String())
        }
    }

    // ALWAYS pass ALL messages to the table
    var cmd tea.Cmd
    _, cmd = m.table.Update(msg)
    
    // Update performance metrics after table processes message
    m.updateMemoryUsage()
    m.performanceMetrics.ActiveChunks = len(m.chunkStates)

    return m, cmd
}
```

## Key Takeaways for Developers

1. **Non-invasive observation**: Never block or modify table messages
2. **Message interception**: Powerful way to understand VTable internals
3. **Chunk tracking**: Critical for understanding performance patterns
4. **Layered debugging**: Different detail levels for different needs
5. **Real-time feedback**: Essential for debugging complex interactions
6. **Performance focus**: Memory and timing metrics reveal bottlenecks

## Running the Example

```bash
cd docs/05-table-component/examples/debug-observability  
go run .
```

**Try this workflow:**
1. Start with debug mode Off (clean interface)
2. Press `d` to enable Basic mode and see activity
3. Navigate, sort, filter while watching the activity log
4. Press `d` again for Detailed mode to see chunk loading
5. Press `d` once more for Verbose mode to see all messages
6. Use `D` for persistent overlay, `Ctrl+R` to reset data

This example demonstrates how to build production-ready debug and monitoring capabilities that provide deep insights into VTable behavior without affecting performance or functionality. 