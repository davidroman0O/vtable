# Multiple Selection Example

This example builds on basic selection by adding advanced multiple selection features and status feedback.

## What it shows

- **Select All**: Ctrl+A to select all 50 items instantly
- **Clear Selection**: Ctrl+D to deselect all items instantly
- **Selection Status**: Real-time count of selected items at bottom
- **Message Handling**: Proper handling of SelectionResponseMsg for UI updates
- **Mixed Operations**: Combine individual spacebar selection with bulk operations

## Run it

```bash
go run main.go
```

## Controls

- `↑`/`↓` or `j`/`k` - Navigate up/down one item
- `PgUp`/`PgDn` or `h`/`l` - Jump up/down by viewport size
- `Home`/`End` or `g`/`G` - Jump to first/last item
- `Space` - Toggle selection of current item
- `Ctrl+A` - Select all items
- `Ctrl+D` - Clear all selections
- `q` or `Ctrl+C` - Quit

## Key learnings

- How to add multiple selection key bindings
- Handling SelectionResponseMsg for UI feedback
- Tracking selection state in your app
- Efficient bulk selection operations
- Providing user feedback with status messages
- Keeping a reference to your DataSource for queries 