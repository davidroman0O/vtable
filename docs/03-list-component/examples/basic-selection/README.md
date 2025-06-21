# Basic Selection Example

This example builds on the enhanced navigation by adding item selection capability.

## What it shows

- **Selection toggle**: Spacebar to select/deselect current item
- **Selection state**: DataSource properly tracks which items are selected
- **Visual feedback**: Selected items render differently (using default theme)
- **Selection persistence**: Selected items stay selected as you navigate
- **Multiple selection**: Can select multiple items independently

## Run it

```bash
go run main.go
```

## Controls

- `↑`/`↓` or `j`/`k` - Navigate up/down one item
- `PgUp`/`PgDn` or `h`/`l` - Jump up/down by viewport size
- `Home`/`End` or `g`/`G` - Jump to first/last item
- `Space` - Toggle selection of current item
- `q` or `Ctrl+C` - Quit

## Key learnings

- How to implement DataSource selection methods properly
- Enabling selection mode in list configuration  
- Handling spacebar for selection toggle
- How selection state is managed and persisted
- Visual feedback for selected items 