# The Viewport System: Your Window Into Data

The viewport system is VTable's navigation engine. It manages the visible area, handles cursor movement, and coordinates with data loading to create smooth scrolling through datasets of any size.

## What is the Viewport?

The viewport is a moving window that shows a small slice of your data. Think of it as looking at a massive spreadsheet through a small rectangular hole - you can see some rows, but the viewport can slide up and down to show different parts.

```
Your Dataset:    Viewport Window:
Item 1           ┌─────────────┐
Item 2           │   Item 6    │ ← ViewportStartIndex (6)
Item 3           │   Item 7    │
Item 4           │   Item 8    │
Item 5           │   Item 9    │
Item 6   ←─────────► Item 10   │
Item 7           │   Item 11   │
Item 8           │ ► Item 12   │ ← Cursor (position 6 in viewport)
Item 9           │   Item 13   │
Item 10          │   Item 14   │
Item 11          │   Item 15   │ ← ViewportEndIndex (15)
Item 12          └─────────────┘
...              Height = 10 items
Item 1000
```

The viewport tracks three critical pieces of information:
- **Where it starts** in your dataset (ViewportStartIndex)
- **Where the cursor is** within the visible area (CursorViewportIndex)
- **Where the cursor is** in the entire dataset (CursorIndex)

## Threshold System: Smart Scrolling

Thresholds control when the viewport scrolls versus when the cursor just moves within the visible area.

### How Thresholds Work

```
Viewport with TopThreshold=2, BottomThreshold=2:

┌─────────────┐
│   Item 20   │ ← Position 0
│   Item 21   │ ← Position 1  
│   Item 22   │ ← Position 2 (TopThreshold)
│   Item 23   │ ← Position 3 }
│   Item 24   │ ← Position 4 } Safe zone
│   Item 25   │ ← Position 5 } (cursor moves freely)
│   Item 26   │ ← Position 6 }
│   Item 27   │ ← Position 7 (BottomThreshold = Height-2-1)
│   Item 28   │ ← Position 8
│   Item 29   │ ← Position 9
└─────────────┘
```

**Navigation rules:**
- **Safe zone (positions 3-6)**: Cursor moves, viewport stays put
- **Top threshold (position 2)**: Press ↑ → viewport scrolls up, cursor stays at position 2
- **Bottom threshold (position 7)**: Press ↓ → viewport scrolls down, cursor stays at position 7
- **Edges (positions 0, 8-9)**: Cursor moves until it hits a threshold

### Why Use Thresholds?

Without thresholds, you'd have two bad options:
1. **Cursor always moves**: Eventually hits the edge where you can't see what's beyond
2. **Viewport always scrolls**: Can't browse within the current view

Thresholds give you the best of both: **browse locally, scroll globally**.

## Bounding Area: Predictive Loading

The bounding area determines which data chunks stay loaded around your viewport:

```
                 BoundingAreaBefore (20 items)
                           ↓
    Chunk 2    |    Chunk 3    |    Chunk 4    |    Chunk 5
   Items       |   Items       |   Items       |   Items
   40-59       |   60-79       |   80-99       |   100-119
               |               |               |
               |    ┌─────────────┐            |
               |    │   Item 85   │            |
               |    │   Item 86   │            |
               |    │   Item 87   │            |
               |    │ ► Item 88   │ ← Cursor   |
               |    │   Item 89   │            |
               |    │   Item 90   │            |
               |    │   Item 91   │            |
               |    │   Item 92   │            |
               |    │   Item 93   │            |
               |    │   Item 94   │            |
               |    └─────────────┘            |
               |         Viewport              |
               |      (showing 85-94)          |
               |                               |
                        BoundingAreaAfter (20 items)
                                 ↓

Total loaded: Items 60-119 (chunks 3, 4, 5)
```

**Benefits:**
- **Smooth scrolling**: Next/previous chunks already loaded
- **Memory efficiency**: Only loads ~3 chunks instead of entire dataset  
- **Automatic cleanup**: Distant chunks get unloaded

## How Navigation Actually Works

Understanding navigation means understanding the relationship between your actions and how the viewport responds.

### The Three-Coordinate System

Every viewport operation manages three related positions:

```
Dataset: [Item 0][Item 1][Item 2]...[Item 95][Item 96][Item 97][Item 98][Item 99]...

                    ┌─────────────┐
                    │   Item 95   │ ← Position 0 in viewport
                    │   Item 96   │ ← Position 1
                    │   Item 97   │ ← Position 2 (top threshold)  
                    │   Item 98   │ ← Position 3
                    │ ► Item 99   │ ← Position 4 (cursor here)
                    │   Item 100  │ ← Position 5
                    │   Item 101  │ ← Position 6
                    │   Item 102  │ ← Position 7 (bottom threshold)
                    │   Item 103  │ ← Position 8
                    │   Item 104  │ ← Position 9 
                    └─────────────┘

ViewportStartIndex = 95   (first visible item in dataset)
CursorIndex = 99          (selected item in dataset) 
CursorViewportIndex = 4   (cursor position within viewport)
```

**The key insight**: The cursor has two addresses - its absolute position in your dataset (99) and its relative position in the viewport window (4).

### Basic Movement: Cursor in Safe Zone

When the cursor is in the safe zone (positions 3-6), pressing arrow keys just moves the cursor:

```
Before ↓:
│   Item 97   │ ← Position 2 (top threshold)  
│   Item 98   │ ← Position 3
│ ► Item 99   │ ← Position 4 (cursor)
│   Item 100  │ ← Position 5
│   Item 101  │ ← Position 6

After ↓:
│   Item 97   │ ← Position 2 (top threshold)  
│   Item 98   │ ← Position 3
│   Item 99   │ ← Position 4 
│ ► Item 100  │ ← Position 5 (cursor moved)
│   Item 101  │ ← Position 6
```

**What happened**: CursorIndex changed from 99 to 100. CursorViewportIndex changed from 4 to 5. ViewportStartIndex stayed at 95. The viewport window didn't move at all.

### Threshold Movement: Viewport Scrolls

When the cursor hits a threshold, the behavior changes:

```
Before ↓ (cursor at bottom threshold):
│   Item 98   │ ← Position 3
│   Item 99   │ ← Position 4 
│   Item 100  │ ← Position 5
│   Item 101  │ ← Position 6
│ ► Item 102  │ ← Position 7 (bottom threshold)
│   Item 103  │ ← Position 8
│   Item 104  │ ← Position 9

After ↓ (viewport scrolls):
│   Item 99   │ ← Position 3
│   Item 100  │ ← Position 4 
│   Item 101  │ ← Position 5
│   Item 102  │ ← Position 6
│ ► Item 103  │ ← Position 7 (cursor stayed at threshold!)
│   Item 104  │ ← Position 8
│   Item 105  │ ← Position 9
```

**What happened**: The viewport scrolled down by 1. ViewportStartIndex changed from 95 to 96. CursorIndex changed from 102 to 103. But CursorViewportIndex stayed at 7 - the cursor "stuck" to the threshold position.

### Page Movement: Big Jumps

Page Up/Down moves by the full viewport height but tries to maintain cursor position:

```
Before Page Down:
ViewportStartIndex: 50
┌─────────────┐
│   Item 50   │ ← Position 0
│   Item 51   │ ← Position 1
│   Item 52   │ ← Position 2
│ ► Item 53   │ ← Position 3 (cursor)
│   Item 54   │ ← Position 4
│   ...       │
│   Item 59   │ ← Position 9
└─────────────┘

After Page Down:
ViewportStartIndex: 60
┌─────────────┐
│   Item 60   │ ← Position 0
│   Item 61   │ ← Position 1
│   Item 62   │ ← Position 2
│ ► Item 63   │ ← Position 3 (cursor maintained relative position!)
│   Item 64   │ ← Position 4
│   ...       │
│   Item 69   │ ← Position 9
└─────────────┘
```

The viewport jumped by 10 items (height), and the cursor stayed at the same relative position (3) within the new viewport.

## The Coordination Dance

The viewport system coordinates three things simultaneously:

1. **Your input** (press ↓)
2. **Data loading** (what chunks need to be loaded)
3. **Visual updates** (what the user sees)

### When You Press ↓

1. **Check current state**: Where is the cursor? Is it at a threshold?
2. **Calculate new position**: Should cursor move or viewport scroll?
3. **Update positions**: Change ViewportStartIndex and/or CursorIndex
4. **Check data needs**: What chunks are needed for the new viewport?
5. **Load missing chunks**: Request data from DataSource if needed
6. **Update display**: Show the new viewport content

This all happens in a single update cycle, creating the illusion of smooth navigation.

### Data Loading Coordination

The viewport system tells the data virtualization system what's needed:

```
Current viewport shows items 85-94
↓
Bounding area needs items 65-115  
↓
Chunks needed: 60-79, 80-99, 100-119
↓
DataSource.LoadChunk() calls for any missing chunks
↓
When chunks arrive, viewport can display them
```

The viewport never waits for data - it shows what's available and requests what's missing.

## Key Mental Models

### Model 1: Sliding Window
The viewport is a window that slides over your data. The data never moves - only your viewing window does.

### Model 2: Two Modes
Navigation has two modes:
- **Local mode**: Cursor moves within viewport (safe zone)
- **Global mode**: Viewport moves through dataset (threshold zones)

### Model 3: Predictive Loading
The viewport system is always thinking ahead, loading data for where you might go next.

### Model 4: Three-Layer Coordination
- **UI layer**: What you see
- **Viewport layer**: Position calculations  
- **Data layer**: Chunk loading

These layers work together but can operate independently (viewport can move while data loads).

## Why This Design Works

1. **Smooth navigation**: Thresholds prevent jarring viewport jumps
2. **Predictable behavior**: Same action always produces same result
3. **Efficient data use**: Only loads what's needed plus small buffer
4. **Scalable**: Works the same for 100 items or 100 million items
5. **Responsive**: UI updates immediately, data loads in background

The viewport system creates the illusion that you're smoothly scrolling through your entire dataset, when you're actually viewing a small, carefully-managed window that coordinates with efficient data loading.

**Next:** [Commands and Messages →](04-commands-and-messages.md) 