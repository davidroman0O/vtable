# Core Concepts: Commands and Messages

VTable components are controlled using the standard Bubble Tea **command/message pattern**. You don't call methods directly on VTable components. Instead, you send commands, and the components respond with messages. This keeps your application's architecture clean, asynchronous, and easy to test.

## The Basic Integration Pattern

Here is the essential pattern for using VTable in your Bubble Tea application's `Update` method:

```go
func (m MyApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // 1. Map user key presses to VTable commands.
        switch msg.String() {
        case "j", "down":
            return m, core.CursorDownCmd()
        case "k", "up":
            return m, core.CursorUpCmd()
        case " ":
            return m, core.SelectCurrentCmd()
        // ... other key mappings ...
        }

    // 2. Let VTable handle its own internal messages.
    default:
        var cmd tea.Cmd
        // This line is crucial!
        _, cmd = m.list.Update(msg)
        return m, cmd
    }
    return m, nil
}
```

-   **You send commands:** `core.CursorDownCmd()` tells VTable to move the cursor down.
-   **VTable handles the logic:** It calculates the new viewport state, determines if new data is needed, and issues its own internal commands (like loading a chunk).
-   **You pass messages through:** You must pass any unhandled messages to your VTable component's `Update` method so it can process data loading responses, etc.

## Navigation Commands

All core navigation commands are in the `github.com/davidroman0O/vtable/core` package.

#### Basic Movement
-   `core.CursorUpCmd()`: Move the cursor up one item.
-   `core.CursorDownCmd()`: Move the cursor down one item.

#### Page Movement
-   `core.PageUpCmd()`: Move the cursor up by one viewport height.
-   `core.PageDownCmd()`: Move the cursor down by one viewport height.

#### Jump Movement
-   `core.JumpToStartCmd()`: Move the cursor to the first item in the dataset.
-   `core.JumpToEndCmd()`: Move the cursor to the last item.
-   `core.JumpToCmd(index)`: Move the cursor to a specific absolute index.

## Selection Commands

Selection commands are also in the `core` package.

-   `core.SelectCurrentCmd()`: Toggle the selection state of the item currently under the cursor.
-   `core.SelectAllCmd()`: Select all items in the dataset (requires `DataSource` support).
-   `core.SelectClearCmd()`: Clear all selections (requires `DataSource` support).

## Data Commands

Use these commands to manage the component's data.

-   `core.DataRefreshCmd()`: Forces a full data reload. This clears all cached chunks and re-requests the total item count from the `DataSource`. Use this when your underlying data has changed significantly.
-   `core.DataChunksRefreshCmd()`: Refreshes only the currently loaded chunks. This is useful for reflecting minor state changes (like an updated selection) without a full reload.

## Important Response Messages

While VTable handles most of its internal messages, you might want to listen for a few key responses to update your application's UI.

#### `core.SelectionResponseMsg`
Sent by a `DataSource` after a selection operation completes. You can listen for this to update a status bar with the number of selected items.

```go
case core.SelectionResponseMsg:
    if msg.Success {
        // Update your app's state, e.g., a selection counter.
        m.selectionCount = m.list.GetSelectionCount()
        m.statusMessage = fmt.Sprintf("%d items selected", m.selectionCount)
    }
    // ALWAYS pass the message on to the VTable component.
    var cmd tea.Cmd
    _, cmd = m.list.Update(msg)
    return m, cmd
```

#### `core.DataTotalMsg`
Sent after the total number of items has been fetched. Useful for displaying total counts in your UI.

```go
case core.DataTotalMsg:
    m.totalItems = msg.Total
    // Let the VTable component process it too.
    var cmd tea.Cmd
    _, cmd = m.list.Update(msg)
    return m, cmd
```

## Tree-Specific Commands

The Tree component has a special navigation command.

-   `core.TreeJumpToIndexCmd(index, expandParents bool)`: Jumps to an item by its index in the "fully expanded" view of the tree. If `expandParents` is true, it will automatically expand all parent nodes to make the target item visible.

## What's Next?

You now know how to control VTable components using commands and how to listen for important messages. Next, we'll explore how to customize the visual appearance of your components.

**Next:** [Component Rendering: Making It Look Good →](05-component-rendering.md) 