# Core Concepts: Data Virtualization

Data virtualization is the core concept that makes VTable powerful and efficient. Instead of loading and rendering thousands of items at once, VTable creates an illusion: it shows you a small, manageable window into your data and handles all the complex logic behind the scenes.

Think of it like looking at a massive landscape through a small, movable periscope. You only see a small portion at any given time, but you can smoothly move the periscope to view any part of the landscape you want. The landscape itself doesn't move—only your viewing window does.

## Why It Matters: Performance and Scale

Without virtualization, displaying 100,000 items would mean:
- Loading all 100,000 items into memory immediately.
- Attempting to render all 100,000 text lines.
- Sluggish, unresponsive scrolling.
- Memory usage that grows linearly with your data size.

With data virtualization, VTable does the opposite:
- Only visible items (plus a small buffer) are loaded into memory.
- Memory usage stays constant, regardless of the total data size.
- Scrolling is always fast and smooth because the component only manages a small number of items at a time.
- Data is loaded "lazily" from your `DataSource` only when needed.

## The Viewport: Your Moving Window

The **viewport** is your window into the data. It's defined by a few key properties that work together to create the seamless scrolling experience.

```text
Your Data:  Item 1
            Item 2
            Item 3
            Item 4
            Item 5
            Item 6   ← ViewportStartIndex
            Item 7
            Item 8
            Item 9
            Item 10
            Item 11
            Item 12  ← CursorIndex
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
            │ ► Item 12   │ ← Cursor (at CursorViewportIndex 6)
            │   Item 13   │
            │   Item 14   │
            │   Item 15   │
            └─────────────┘
```

The viewport tracks three essential pieces of information:

-   `ViewportStartIndex`: The absolute index of the first item visible on screen.
-   `CursorIndex`: The absolute index of the selected item in the *entire dataset*.
-   `CursorViewportIndex`: The cursor's *relative* position within the visible viewport (from `0` to `Height-1`).

When you navigate, VTable updates these three values to create the illusion of smooth movement.

## Scroll Thresholds: Smart Navigation

Not every cursor movement needs to trigger a full-screen scroll. VTable uses a **threshold system** to create a more natural navigation experience. The cursor can move freely within a "safe zone" in the middle of the viewport, and scrolling only happens when you approach the top or bottom edges.

```text
Viewport (Height=10) with TopThreshold=2 and BottomThreshold=2:

┌─────────────┐
│   Item 20   │ --
│   Item 21   │   | Top Threshold Zone (Positions 0-2)
│   Item 22   │ --
│   Item 23   │
│   Item 24   │
│   Item 25   │   } Safe Zone (Cursor moves, no scrolling)
│   Item 26   │
│   Item 27   │ --
│   Item 28   │   | Bottom Threshold Zone (Positions 7-9)
│   Item 29   │ --
└─────────────┘
```

**The logic:**
-   **In the Safe Zone:** When you press `↓`, only the `CursorIndex` and `CursorViewportIndex` change. The viewport itself doesn't move.
-   **At a Threshold:** When the cursor is at the `TopThreshold` and you press `↑`, the `ViewportStartIndex` changes (the whole view scrolls up), but the `CursorViewportIndex` remains "stuck" at the threshold.

This provides the best of both worlds: you can browse the currently visible items without the view jumping around, but the view scrolls automatically when you need to see more.

## Chunks: Loading Data in Blocks

Loading items one-by-one would be inefficient, especially over a network. Instead, VTable requests data from your `DataSource` in **chunks**—blocks of consecutive items.

```text
Your Data:    Chunk 0 (Items 1-20)     Chunk 1 (Items 21-40)    Chunk 2 (Items 41-60)
            ┌──────────────────┐       ┌───────────────────┐      ┌───────────────────┐
            │ Item 1           │       │ Item 21           │      │ Item 41           │
            │ ...              │       │ ...               │      │ ...               │
            │ Item 20          │       │ Item 40           │      │ Item 60           │
            └──────────────────┘       └───────────────────┘      └───────────────────┘
```

-   **Chunk size is configurable** but is typically between 20-100 items.
-   When you need `Item 35`, VTable calculates which chunk it belongs to (Chunk 1) and requests the entire chunk (`Items 21-40`).
-   If you then scroll to `Item 30`, no new data is loaded because it's already in memory.

## The Bounding Area: Smart Buffering

Loading only the currently visible chunks would cause a noticeable lag every time you scroll past a chunk boundary. To prevent this, VTable uses a **bounding area** to proactively load chunks *around* your current viewport.

```text
              Chunk 3        Chunk 4         Chunk 5
              Items 60-79    Items 80-99     Items 100-119
             ┌──────────┐   ┌─────────────┐   ┌───────────┐
             │ ...      │   │   Item 85   │   │ ...       │
             │          │   │   ...       │   │           │
             │          │   │ ► Item 92   │   │           │
             │          │   │   ...       │   │           │
             │          │   │   Item 94   │   │           │
             └──────────┘   └─────────────┘   └───────────┘
               ^              ^               ^
               |           Viewport           |
               └───── Bounding Area ─────┘
               (Chunks 3, 4, and 5 are loaded)
```

**How it works:**
-   The viewport needs items from **Chunk 4**.
-   The `BoundingAreaBefore` configuration tells VTable to also load **Chunk 3**.
-   The `BoundingAreaAfter` configuration tells VTable to also load **Chunk 5**.

As you scroll, the bounding area moves with the viewport. New chunks are loaded ahead of time, and distant chunks that are no longer in the bounding area are unloaded to conserve memory.

## Putting It All Together: A Scrolling Session

Let's trace a complete scrolling session to see how these concepts work together.

**Initial State:**
-   **Viewport shows:** Items 25-34.
-   **Cursor is on:** Item 31.
-   **Bounding area needs:** Items 5-54.
-   **VTable loads:** Chunks 0, 1, and 2 (items 1-60).

```text
Data in memory: Chunks 0-2 (items 1-60)
┌─────────────┐
│   Item 25   │
│   ...       │
│ ► Item 31   │ ← Cursor here
│   ...       │
│   Item 34   │
└─────────────┘
```

**1. Press `↓` (Cursor moves within viewport):**
-   The cursor moves to Item 32.
-   `CursorIndex` becomes `32`.
-   `CursorViewportIndex` becomes `7` (now at the bottom threshold).
-   `ViewportStartIndex` remains `25`.
-   No new data is loaded.

**2. Press `↓` again (Viewport scrolls):**
-   The cursor was at the bottom threshold, so the viewport scrolls.
-   `ViewportStartIndex` becomes `26`.
-   `CursorIndex` becomes `33`.
-   `CursorViewportIndex` remains `7`. The cursor "sticks" to the threshold.
-   No new data is loaded yet, as Item 35 is still in the loaded chunks.

**3. Keep pressing `↓` until the cursor is on Item 75:**
-   `ViewportStartIndex` is now `69`.
-   `CursorIndex` is `75`.
-   As you scrolled, the bounding area moved. Chunks 0 and 1 were unloaded from memory. Chunks 3 and 4 were loaded.
-   **Memory usage remains constant**, even though you've scrolled through dozens of items.

## The VTable Philosophy

VTable creates the **illusion** of a massive, instantly accessible dataset by intelligently managing a small, moving window.

1.  **Viewport**: Defines the small window your user sees.
2.  **Thresholds**: Ensure smooth, predictable navigation.
3.  **Chunks**: Make data loading efficient.
4.  **Bounding Area**: Provides a seamless, lag-free scrolling experience.

This architecture is what allows VTable to deliver high performance at any scale.

## What's Next?

Now that you understand *how* VTable works, let's look at the `DataSource`—the component that you will implement to feed your data into this system.

**Next:** [DataSources: Your Data Provider →](02-datasources.md) 