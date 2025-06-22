# Core Concepts: The Viewport System

The viewport system is VTable's navigation engine. It's responsible for managing the visible area of your component, handling all cursor movements, and coordinating with the `DataSource` to create a smooth, seamless scrolling experience through datasets of any size.

## What is the Viewport?

The viewport is a **moving window** that shows a small slice of your data. Think of it as looking at a massive spreadsheet through a small rectangular hole—you can only see a few rows at a time, but you can slide the opening up and down to view any part of the sheet.

```text
Your Dataset:    Viewport Window:
Item 1           ┌─────────────┐
...              │   Item 6    │ ← ViewportStartIndex
Item 5           │   ...       │
Item 6   <─────────► Item 12   │ ← CursorIndex
...              │   ...       │
Item 15          │   Item 15   │ ← ViewportEnd
...              └─────────────┘
Item 1000
```

The viewport's state is defined by three critical coordinates:

-   `ViewportStartIndex`: The absolute index in your dataset of the *first visible item*. This is the top of the window.
-   `CursorIndex`: The absolute index of the *currently selected item* in the entire dataset.
-   `CursorViewportIndex`: The *relative position* of the cursor within the visible viewport (e.g., the 3rd item on screen).

When you send a navigation command like `core.CursorDownCmd()`, the viewport system intelligently updates these three coordinates to produce fluid movement.

## The Scrolling Illusion

When you scroll, the data itself doesn't move—the viewport does. It's an efficient illusion created by changing the `ViewportStartIndex`.

-   **Before `↓`:** Viewport shows items 6-15.
-   **After `↓`:** Viewport shows items 7-16.

The component simply asks the `DataSource` for a different chunk of data. This is far more efficient than reordering or re-rendering a massive list.

## Thresholds: The Key to Smart Scrolling

To prevent the UI from scrolling on every single key press, VTable uses a **threshold system**. This creates a "safe zone" in the middle of the viewport where the cursor can move freely. Scrolling only occurs when the cursor approaches the top or bottom edges.

```text
Viewport (Height=10) with TopThreshold=2 and BottomThreshold=2:

┌─────────────┐
│   Item 20   │ --
│   Item 21   │   | Top Threshold Zone
│   Item 22   │ --
│   Item 23   │
│   Item 24   │   } Safe Zone (cursor moves, no scrolling)
│   ...       │
│   Item 27   │ --
│   Item 28   │   | Bottom Threshold Zone
│   Item 29   │ --
└─────────────┘
```

**This creates two distinct navigation modes:**

1.  **Local Navigation (in the Safe Zone):**
    -   Pressing `↓` only changes the `CursorIndex` and `CursorViewportIndex`.
    -   The `ViewportStartIndex` remains the same. The view is stable.

2.  **Global Scrolling (at a Threshold):**
    -   When the cursor hits a threshold (e.g., position 2) and you press `↑`, the `ViewportStartIndex` changes.
    -   The entire view scrolls up, but the cursor appears to "stick" to the threshold position, providing a smooth transition.

This system provides a superior user experience by balancing local browsing with seamless global navigation.

## Bounding Area: Predictive Loading for Smoothness

The viewport system works hand-in-hand with the **bounding area** to ensure scrolling is always lag-free. It proactively loads data chunks just outside the visible area, so they are already in memory when the user scrolls to them.

```text
                 BoundingAreaBefore (e.g., 20 items)
                           ↓
    Chunk 3    |    Chunk 4    |    Chunk 5
   Items 60-79 |   Items 80-99 |   Items 100-119
               |               |
               |    ┌─────────────┐
               |    │   Item 85   │
               |    │   ...       │
               |    │ ► Item 92   │ ← Cursor
               |    │   ...       │
               |    └─────────────┘
               |      Viewport
               |
               └───── Bounding Area (Chunks 3, 4, 5 are loaded) ─────┘
                                 ↑
                        BoundingAreaAfter (e.g., 20 items)
```

As the viewport moves, the bounding area moves with it, loading new chunks and discarding old ones. This is the essence of VTable's performance: it keeps a small, relevant portion of your data in memory at all times.

## How It All Works Together

When you send a navigation command:
1.  **Calculate New Position:** The viewport system determines the new values for the three core coordinates based on your input and the threshold rules.
2.  **Check Data Needs:** It calculates the new bounding area.
3.  **Request Data:** It tells the `DataSource` to load any chunks needed for the new bounding area that aren't already in memory.
4.  **Update Display:** It renders the items for the new `ViewportStartIndex`.

This entire cycle happens on every navigation command, creating a system that is both highly responsive and incredibly scalable.

## What's Next?

Now that you understand the "engine" of VTable, let's look at the "controls." The next section covers the specific commands and messages you'll use to interact with VTable components.

**Next:** [Commands and Messages: Controlling VTable →](04-commands-and-messages.md) 