# Data Virtualization in VTable

Data virtualization is the core technology that makes VTable efficient with large datasets. Instead of loading and rendering thousands of items at once, VTable creates an illusion, it shows you a small window into your data and manages everything behind the scenes.

Think of it like looking through a periscope at a massive landscape. You see a small portion at any given time, but you can move the periscope to see different parts. The landscape doesn't move, your viewing window does.

## Why Data Virtualization Matters

Without virtualization, displaying 100,000 items would mean:
- Loading all 100,000 items into memory immediately
- Rendering all 100,000 DOM elements (in web contexts) or text lines
- Sluggish scrolling as your system struggles with massive datasets
- Memory usage that grows linearly with data size

With virtualization:
- Only visible items (plus a small buffer) are loaded
- Memory usage stays constant regardless of total data size
- Smooth scrolling because you're only managing ~50 items instead of 100,000
- Lazy loading - data is fetched only when needed

## The Viewport: A Moving Window

The viewport is your window into the data. It's defined by three key properties that work together to create the scrolling experience.

```
Your Data:  Item 1
            Item 2  
            Item 3
            Item 4
            Item 5
            Item 6   ← ViewportStart
            Item 7
            Item 8   
            Item 9
            Item 10
            Item 11
            Item 12  ← Cursor here
            Item 13
            Item 14
            Item 15  ← ViewportEnd
            Item 16
            Item 17
            ...
            Item 1000

Terminal    ┌─────────────┐
Screen:     │   Item 6    │
            │   Item 7    │
            │   Item 8    │
            │   Item 9    │
            │   Item 10   │
            │   Item 11   │
            │ ► Item 12   │ ← Cursor
            │   Item 13   │
            │   Item 14   │
            │   Item 15   │
            └─────────────┘
```

The viewport tracks three essential pieces of state:

- **ViewportStartIndex (6)**: The first item visible on screen. This is the "top" of your window into the data.
- **Height (10)**: How many items can fit vertically in your terminal. This is determined by your terminal size and the component's allocated space.
- **CursorIndex (12)**: Which item the user has selected. This can be anywhere within the viewport, and its position relative to the viewport edges determines when scrolling happens.

These three values completely define what the user sees and where they are in the dataset. When you press arrow keys, VTable updates these values to create smooth navigation.

## The Scrolling Illusion

Here's the key insight: when you scroll, the data doesn't move - the viewport does. It's like sliding a window across a stationary wall of data.

```
Before:     Item 4         After ↓:    Item 5
            Item 5                     Item 6
            ┌─────────────┐             ┌─────────────┐
            │   Item 6    │             │   Item 7    │
            │   Item 7    │             │   Item 8    │
            │   Item 8    │             │   Item 9    │
            │   Item 9    │             │   Item 10   │
            │   Item 10   │             │   Item 11   │
            │   Item 11   │             │   Item 12   │
            │ ► Item 12   │    →        │ ► Item 13   │
            │   Item 13   │             │   Item 14   │
            │   Item 14   │             │   Item 15   │
            │   Item 15   │             │   Item 16   │
            └─────────────┘             └─────────────┘
            Item 16                     Item 17
            Item 17                     Item 18
```

What actually happened here? VTable incremented the ViewportStartIndex from 6 to 7. The cursor's absolute position in the dataset went from 12 to 13, but its relative position within the viewport stayed the same (position 6 out of 10).

This is much more efficient than trying to move or reorder actual data. VTable simply changes which slice of the data it asks for and renders.

## Cursor Thresholds: When to Scroll vs When to Move

Not every cursor movement triggers scrolling. VTable uses a threshold system to provide a smooth navigation experience where the cursor can move freely in the middle of the viewport, and scrolling only happens when you approach the edges.

```
Viewport showing items 20-29:
┌─────────────┐
│   Item 20   │ 
│   Item 21   │ 
│   Item 22   │ ← TopThreshold (position 2)
│   Item 23   │
│   Item 24   │ ← Safe zone (cursor moves, no scroll)
│   Item 25   │
│   Item 26   │
│   Item 27   │ ← BottomThreshold (position 7)
│   Item 28   │
│   Item 29   │ 
└─────────────┘
```

**The threshold logic works like this:**

When you press the up arrow:
- If cursor is in positions 0-2 (top threshold zone), the viewport scrolls up instead of just moving the cursor
- If cursor is in positions 3-6 (safe zone), only the cursor moves up
- This keeps the cursor away from the very edge where it would be hard to see context

When you press the down arrow:
- If cursor is in positions 7-9 (bottom threshold zone), the viewport scrolls down
- If cursor is in positions 3-6 (safe zone), only the cursor moves down

**Why use thresholds at all?** Without thresholds, you'd have two bad options:
1. Cursor always moves without scrolling - but then it hits the edge and you can't see what's below/above
2. Viewport always scrolls with cursor movement - but then you can't browse within the current view

Thresholds give you the best of both worlds: browse freely in the middle, auto-scroll at the edges.

Let's see this in action:

```
Cursor at bottom threshold:
┌─────────────┐
│   Item 20   │
│   Item 21   │
│   Item 22   │
│   Item 23   │
│   Item 24   │
│   Item 25   │
│   Item 26   │
│ ► Item 27   │ ← Bottom threshold
│   Item 28   │
│   Item 29   │
└─────────────┘

Press ↓ → Viewport scrolls:
┌─────────────┐
│   Item 21   │
│   Item 22   │
│   Item 23   │
│   Item 24   │
│   Item 25   │
│   Item 26   │
│   Item 27   │
│ ► Item 28   │ ← Cursor stayed at same relative position
│   Item 29   │
│   Item 30   │
└─────────────┘
```

Notice how the cursor stayed at the same relative position (7th slot) but now points to item 28 instead of 27. The viewport scrolled, not the cursor.

## Chunks: Loading Data in Blocks

Individual item loading would be incredibly inefficient. Imagine calling your database for each item as the user scrolls - you'd make thousands of tiny requests. Instead, VTable loads data in chunks (blocks of consecutive items).

```
Your Data:    Chunk 0:      Chunk 1:      Chunk 2:
              Item 1        Item 21       Item 41
              Item 2        Item 22       Item 42
              Item 3        Item 23       Item 43
              Item 4        Item 24       Item 44
              Item 5        Item 25       Item 45
              Item 6        Item 26       Item 46
              Item 7        Item 27       Item 47
              Item 8        Item 28       Item 48
              ...           ...           ...
              Item 20       Item 40       Item 60
```

**Chunk size is configurable** but typically ranges from 20-100 items. Larger chunks mean fewer network requests but more memory usage. Smaller chunks mean more responsive loading but more overhead.

**How chunk loading works:**
When you need item 35, VTable calculates which chunk contains it:

```go
Item 35, chunk size 20:
chunkNumber = 35 ÷ 20 = 1      // Integer division
chunkStart = 1 × 20 = 20       // Start of chunk 1  
chunkEnd = chunkStart + 20 = 40 // End of chunk 1
// Load chunk 1: items 20-39 (which includes item 35)
```

This means when you scroll to item 35, you automatically get items 20-39 loaded into memory. If you then scroll to item 30, no new loading is needed - it's already in memory from the same chunk.

**Chunk boundaries and viewport interaction:**
Your viewport might span multiple chunks. For example, if your viewport shows items 18-27 and your chunk size is 20:
- Item 18-19 come from chunk 0 (items 1-20)  
- Items 20-27 come from chunk 1 (items 21-40)

VTable automatically loads both chunks to fulfill the viewport requirements.

## The Bounding Area: Smart Buffering

Loading just the visible chunks would cause stuttering - every time you scroll past a chunk boundary, you'd wait for the next chunk to load. The bounding area solves this by loading extra chunks around your current viewport.

```
              Chunk 3        Chunk 4         Chunk 5
              Item 60        Item 80         Item 100
              Item 61        Item 81         Item 101
              Item 62        Item 82         Item 102
              Item 63        Item 83         Item 103
              Item 64        Item 84         Item 104
              Item 65      ┌─────────────┐   Item 105
              Item 66      │   Item 85   │   Item 106
              Item 67      │   Item 86   │   Item 107
              Item 68      │   Item 87   │   Item 108
              Item 69      │   Item 88   │   Item 109
              ...          │   Item 89   │   ...
              Item 79      │   Item 90   │   Item 119
                           │   Item 91   │
                           │ ► Item 92   │
                           │   Item 93   │
                           │   Item 94   │
                           └─────────────┘
                              Viewport
                           (showing 85-94)

VTable loads chunks 3, 4, and 5 (items 60-119)
```

**The bounding area calculation:**
- Viewport needs items 85-94 (chunk 4)
- BoundingAreaBefore: 20 items → load chunk 3 (items 60-79)  
- BoundingAreaAfter: 20 items → load chunk 5 (items 100-119)
- Total loaded: items 60-119 (60 items for a 10-item viewport)

**Why this works so well:**
When you press ↓ and scroll to item 95, VTable already has items 95-119 in memory. No loading delay, smooth scrolling. When you press ↑ and scroll to item 75, it's already loaded too.

**Memory management:**
VTable doesn't keep everything forever. As you scroll far away from chunks, they get unloaded to prevent memory bloat. Typically, chunks outside the bounding area are candidates for cleanup.

**Bounding area configuration:**
```go
BoundingAreaBefore: 20  // Keep 20 items before viewport
BoundingAreaAfter: 20   // Keep 20 items after viewport
ChunkSize: 20          // Load 20 items per chunk
```

With these settings:
- Viewport shows 10 items
- Bounding area loads ~40 extra items (20 before + 20 after)  
- Total memory footprint: ~50 items regardless of dataset size
- Smooth scrolling in both directions

## Complete Example: Scrolling Through Data

Let's trace through a complete scrolling session to see how all these concepts work together in practice.

### Initial State
```
Data in memory: Chunks 0-2 (items 1-60)
┌─────────────┐
│   Item 25   │
│   Item 26   │
│   Item 27   │
│   Item 28   │
│   Item 29   │
│   Item 30   │
│ ► Item 31   │ ← Cursor here
│   Item 32   │
│   Item 33   │
│   Item 34   │
└─────────────┘
```

**What VTable knows right now:**
- ViewportStartIndex: 25
- CursorIndex: 31 (absolute position in dataset)
- Cursor relative position: 6 (within the 10-item viewport)
- Loaded chunks: 0, 1, 2 (items 1-60)

**Why chunks 0-2?** Viewport shows items 25-34. Bounding area before (20 items) needs items 5-24, so chunk 0 (1-20) and chunk 1 (21-40) are needed. Bounding area after (20 items) needs items 35-54, so chunk 2 (41-60) is loaded too.

### Press ↓ (cursor moves within viewport)
```
Data in memory: Same chunks (items 1-60)
┌─────────────┐
│   Item 25   │
│   Item 26   │
│   Item 27   │
│   Item 28   │
│   Item 29   │
│   Item 30   │
│   Item 31   │
│ ► Item 32   │ ← Cursor moved
│   Item 33   │
│   Item 34   │
└─────────────┘
```

**What happened:**
- CursorIndex: 31 → 32
- Cursor relative position: 6 → 7
- ViewportStartIndex: unchanged (25)
- No chunk loading needed

The cursor is now at relative position 7, which is the bottom threshold. The next ↓ press will trigger scrolling.

### Press ↓ until cursor hits bottom threshold
```
Data in memory: Same chunks (items 1-60)
┌─────────────┐
│   Item 25   │
│   Item 26   │
│   Item 27   │
│   Item 28   │
│   Item 29   │
│   Item 30   │
│   Item 31   │
│ ► Item 32   │ ← Bottom threshold (position 7)
│   Item 33   │
│   Item 34   │
└─────────────┘

Press ↓ → Viewport scrolls:
┌─────────────┐
│   Item 26   │ ← Viewport scrolled
│   Item 27   │
│   Item 28   │
│   Item 29   │
│   Item 30   │
│   Item 31   │
│   Item 32   │
│ ► Item 33   │ ← Cursor stayed at same relative position
│   Item 34   │
│   Item 35   │
└─────────────┘
```

**What happened:**
- ViewportStartIndex: 25 → 26 (viewport scrolled down by 1)
- CursorIndex: 32 → 33 (cursor moved to next item)
- Cursor relative position: stayed at 7 (same position within viewport)
- No chunk loading needed (item 35 is already in chunk 2)

This is the key insight: when scrolling is triggered, both the viewport and cursor move, but the cursor maintains its relative position within the viewport.

### Keep pressing ↓ until item 75
```
Data in memory: Chunks 2-4 (items 41-100)
                (VTable unloaded chunks 0&1, loaded chunks 3&4)
┌─────────────┐
│   Item 71   │
│   Item 72   │
│   Item 73   │
│   Item 74   │
│ ► Item 75   │ ← Cursor here
│   Item 76   │
│   Item 77   │
│   Item 78   │
│   Item 79   │
│   Item 80   │
└─────────────┘
```

**What happened during this longer scroll:**
- ViewportStartIndex: 26 → 71 (many scroll operations)
- CursorIndex: 33 → 75 (cursor followed along)
- Chunks 0&1 were unloaded (items 1-40) because they're now outside the bounding area
- Chunks 3&4 were loaded (items 61-100) to maintain the bounding area around the new viewport
- Memory usage stayed constant (still ~60 items total)

**The chunk loading decision:** When the viewport moved to show items 71-80, VTable calculated:
- Bounding before: need items 51-70 → chunk 2 (41-60) and chunk 3 (61-80)
- Bounding after: need items 81-100 → chunk 4 (81-100)  
- Old chunks 0&1 (items 1-40) are now more than 30 items away, so they get unloaded

## Key Point

VTable creates the **illusion** of scrolling through massive data by cleverly managing a small moving window. The core mechanisms work together:

1. **Viewport**: A small window (10-50 items) that moves through your dataset
2. **Thresholds**: Smart cursor positioning that triggers scrolling only when needed  
3. **Chunks**: Efficient bulk loading instead of individual item requests
4. **Bounding area**: Predictive loading for smooth scrolling
5. **Memory management**: Automatic cleanup of distant chunks

The result: constant memory usage, smooth scrolling, and lazy loading - regardless of whether your dataset has 100 items or 100 million items.

**Next:** [DataSources →](02-datasources.md) 